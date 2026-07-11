package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/irisflair/api/db"
	"github.com/irisflair/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---- Track rate limiting (in-memory, per client IP, fixed window) ----

type rateWindow struct {
	count       int
	windowStart time.Time
}

var (
	trackWindows  = make(map[string]*rateWindow)
	trackWindowMu sync.Mutex
)

const (
	trackMaxRequests   = 120
	trackWindowLength  = 10 * time.Minute
	trackMaxMapEntries = 10000
	maxEnquiryProducts = 40
)

// trackAllowed reports whether this IP may record another event, counting the call.
func trackAllowed(ip string) bool {
	trackWindowMu.Lock()
	defer trackWindowMu.Unlock()

	now := time.Now()

	// Lazy cleanup so the map can't grow unbounded under IP churn.
	if len(trackWindows) > trackMaxMapEntries {
		for k, w := range trackWindows {
			if now.Sub(w.windowStart) > trackWindowLength {
				delete(trackWindows, k)
			}
		}
	}

	w, ok := trackWindows[ip]
	if !ok || now.Sub(w.windowStart) > trackWindowLength {
		trackWindows[ip] = &rateWindow{count: 1, windowStart: now}
		return true
	}
	w.count++
	return w.count <= trackMaxRequests
}

type trackViewRequest struct {
	ProductID string `json:"productId"`
}

type trackEnquiryRequest struct {
	ProductIDs []string `json:"productIds"`
}

// TrackView records one product view: increments the product's view counter and
// logs a TTL'd event for windowed "trending" computation. Always responds 204 for
// well-formed requests so callers can't probe which product IDs exist.
func TrackView(w http.ResponseWriter, r *http.Request) {
	if db.Client == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	if !trackAllowed(clientIP(r)) {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var req trackViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	objID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	database := db.GetDB()
	res, err := database.Collection("products").UpdateOne(ctx,
		bson.M{"_id": objID, "active": true},
		bson.M{"$inc": bson.M{"viewCount": 1}},
	)
	if err != nil {
		log.Printf("TrackView: counter update error: %v", err)
		http.Error(w, "Failed to record event", http.StatusInternalServerError)
		return
	}

	// Only log events for products that exist and are active — no junk rows.
	if res.MatchedCount > 0 {
		_, err = database.Collection("product_events").InsertOne(ctx, models.ProductEvent{
			ProductID: objID,
			Type:      "view",
			CreatedAt: time.Now(),
		})
		if err != nil {
			log.Printf("TrackView: event insert error: %v", err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// TrackEnquiry records one WhatsApp enquiry hand-off for 1–40 products:
// increments each product's click counter, logs per-product click events, and
// always logs one enquiry_events doc (the enquiry-volume log; multi-product
// docs are the co-enquiry pair data).
func TrackEnquiry(w http.ResponseWriter, r *http.Request) {
	if db.Client == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	if !trackAllowed(clientIP(r)) {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req trackEnquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.ProductIDs) == 0 || len(req.ProductIDs) > maxEnquiryProducts {
		http.Error(w, "productIds must contain 1–40 entries", http.StatusBadRequest)
		return
	}

	seen := make(map[primitive.ObjectID]bool, len(req.ProductIDs))
	ids := make([]primitive.ObjectID, 0, len(req.ProductIDs))
	for _, raw := range req.ProductIDs {
		objID, err := primitive.ObjectIDFromHex(raw)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}
		if !seen[objID] {
			seen[objID] = true
			ids = append(ids, objID)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	database := db.GetDB()
	now := time.Now()

	if _, err := database.Collection("products").UpdateMany(ctx,
		bson.M{"_id": bson.M{"$in": ids}},
		bson.M{"$inc": bson.M{"whatsappClickCount": 1}},
	); err != nil {
		log.Printf("TrackEnquiry: counter update error: %v", err)
		http.Error(w, "Failed to record event", http.StatusInternalServerError)
		return
	}

	events := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		events = append(events, models.ProductEvent{
			ProductID: id,
			Type:      "whatsapp_click",
			CreatedAt: now,
		})
	}
	if _, err := database.Collection("product_events").InsertMany(ctx, events); err != nil {
		log.Printf("TrackEnquiry: event insert error: %v", err)
	}

	if _, err := database.Collection("enquiry_events").InsertOne(ctx, models.EnquiryEvent{
		ProductIDs: ids,
		CreatedAt:  now,
	}); err != nil {
		log.Printf("TrackEnquiry: enquiry insert error: %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
