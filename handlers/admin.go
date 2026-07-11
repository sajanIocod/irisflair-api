package handlers

import (
	"context"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/irisflair/api/db"
	"github.com/irisflair/api/jobs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RecomputeBadges runs the badge/FBT compute synchronously. Auth-protected;
// used by admins (and tests) to refresh badges without waiting for the ticker.
func RecomputeBadges(w http.ResponseWriter, r *http.Request) {
	if db.Client == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	started := time.Now()
	if err := jobs.Recompute(ctx); err != nil {
		log.Printf("RecomputeBadges: %v", err)
		http.Error(w, "Recompute failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "recomputed",
		"durationMs": time.Since(started).Milliseconds(),
	})
}

type productAnalytics struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Code               string   `json:"code"`
	Category           string   `json:"category"`
	Active             bool     `json:"active"`
	ViewCount          int64    `json:"viewCount"`
	WhatsappClickCount int64    `json:"whatsappClickCount"`
	RecentViews        int64    `json:"recentViews"`
	RecentClicks       int64    `json:"recentClicks"`
	Badges             []string `json:"badges"`
}

type analyticsSummary struct {
	TotalViews      int64 `json:"totalViews"`
	TotalClicks     int64 `json:"totalClicks"`
	RecentViews     int64 `json:"recentViews"`
	RecentClicks    int64 `json:"recentClicks"`
	RecentEnquiries int64 `json:"recentEnquiries"`
	EnquiriesTotal  int64 `json:"enquiriesTotal"`
	WindowDays      int   `json:"windowDays"`
}

// dailyActivity is one day of the timeseries (UTC day boundaries).
type dailyActivity struct {
	Date      string `json:"date"` // YYYY-MM-DD
	Views     int64  `json:"views"`
	Clicks    int64  `json:"clicks"`
	Enquiries int64  `json:"enquiries"`
}

// GetAnalytics returns the tracking data (admin-only): per-product view/click
// counters plus a recent window (?days=N, default 14, clamped 1–90) and
// enquiry volume totals.
func GetAnalytics(w http.ResponseWriter, r *http.Request) {
	if db.Client == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	days := 14
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			days = n
		}
	}
	if days < 1 {
		days = 1
	}
	if days > 90 {
		days = 90
	}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	database := db.GetDB()
	now := time.Now()

	// Products with raw counters (hidden from the public API, exposed here).
	cur, err := database.Collection("products").Find(ctx, bson.M{},
		options.Find().SetProjection(bson.M{
			"_id": 1, "name": 1, "code": 1, "category": 1, "active": 1,
			"viewCount": 1, "whatsappClickCount": 1, "badges": 1,
		}))
	if err != nil {
		log.Printf("GetAnalytics: find error: %v", err)
		http.Error(w, "Failed to fetch analytics", http.StatusInternalServerError)
		return
	}
	var raw []struct {
		ID                 primitive.ObjectID `bson:"_id"`
		Name               string             `bson:"name"`
		Code               string             `bson:"code"`
		Category           string             `bson:"category"`
		Active             bool               `bson:"active"`
		ViewCount          int64              `bson:"viewCount"`
		WhatsappClickCount int64              `bson:"whatsappClickCount"`
		Badges             []string           `bson:"badges"`
	}
	if err := cur.All(ctx, &raw); err != nil {
		log.Printf("GetAnalytics: decode error: %v", err)
		http.Error(w, "Failed to decode analytics", http.StatusInternalServerError)
		return
	}

	// Windowed counts over the requested period (same aggregation the badge
	// job uses for its fixed 14-day trending window).
	recentViews := map[primitive.ObjectID]int64{}
	recentClicks := map[primitive.ObjectID]int64{}
	windowStart := now.AddDate(0, 0, -days)
	agg, err := database.Collection("product_events").Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"createdAt": bson.M{"$gte": windowStart}}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{"productId": "$productId", "type": "$type"},
			"n":   bson.M{"$sum": 1},
		}}},
	})
	if err == nil {
		var grouped []struct {
			ID struct {
				ProductID primitive.ObjectID `bson:"productId"`
				Type      string             `bson:"type"`
			} `bson:"_id"`
			N int64 `bson:"n"`
		}
		if err := agg.All(ctx, &grouped); err == nil {
			for _, g := range grouped {
				switch g.ID.Type {
				case "view":
					recentViews[g.ID.ProductID] = g.N
				case "whatsapp_click":
					recentClicks[g.ID.ProductID] = g.N
				}
			}
		}
	}

	// Enquiry volume.
	enquiries := database.Collection("enquiry_events")
	enquiriesTotal, _ := enquiries.CountDocuments(ctx, bson.M{})
	recentEnquiries, _ := enquiries.CountDocuments(ctx,
		bson.M{"createdAt": bson.M{"$gte": windowStart}})

	// Daily timeseries over the window (UTC day buckets), zero-filled.
	dayViews := map[string]int64{}
	dayClicks := map[string]int64{}
	dayEnquiries := map[string]int64{}
	dayAgg, err := database.Collection("product_events").Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"createdAt": bson.M{"$gte": windowStart}}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"day": bson.M{"$dateToString": bson.M{
					"format": "%Y-%m-%d", "date": "$createdAt"}},
				"type": "$type",
			},
			"n": bson.M{"$sum": 1},
		}}},
	})
	if err == nil {
		var rows []struct {
			ID struct {
				Day  string `bson:"day"`
				Type string `bson:"type"`
			} `bson:"_id"`
			N int64 `bson:"n"`
		}
		if err := dayAgg.All(ctx, &rows); err == nil {
			for _, row := range rows {
				switch row.ID.Type {
				case "view":
					dayViews[row.ID.Day] = row.N
				case "whatsapp_click":
					dayClicks[row.ID.Day] = row.N
				}
			}
		}
	}
	enqAgg, err := enquiries.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"createdAt": bson.M{"$gte": windowStart}}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{"$dateToString": bson.M{
				"format": "%Y-%m-%d", "date": "$createdAt"}},
			"n": bson.M{"$sum": 1},
		}}},
	})
	if err == nil {
		var rows []struct {
			ID string `bson:"_id"`
			N  int64  `bson:"n"`
		}
		if err := enqAgg.All(ctx, &rows); err == nil {
			for _, row := range rows {
				dayEnquiries[row.ID] = row.N
			}
		}
	}
	daily := make([]dailyActivity, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := now.UTC().AddDate(0, 0, -i).Format("2006-01-02")
		daily = append(daily, dailyActivity{
			Date:      d,
			Views:     dayViews[d],
			Clicks:    dayClicks[d],
			Enquiries: dayEnquiries[d],
		})
	}

	summary := analyticsSummary{
		EnquiriesTotal:  enquiriesTotal,
		RecentEnquiries: recentEnquiries,
		WindowDays:      days,
	}
	products := make([]productAnalytics, 0, len(raw))
	for _, p := range raw {
		badges := p.Badges
		if badges == nil {
			badges = []string{}
		}
		pa := productAnalytics{
			ID:                 p.ID.Hex(),
			Name:               p.Name,
			Code:               p.Code,
			Category:           p.Category,
			Active:             p.Active,
			ViewCount:          p.ViewCount,
			WhatsappClickCount: p.WhatsappClickCount,
			RecentViews:        recentViews[p.ID],
			RecentClicks:       recentClicks[p.ID],
			Badges:             badges,
		}
		summary.TotalViews += pa.ViewCount
		summary.TotalClicks += pa.WhatsappClickCount
		summary.RecentViews += pa.RecentViews
		summary.RecentClicks += pa.RecentClicks
		products = append(products, pa)
	}

	// Most-enquired first, then most-viewed.
	sort.Slice(products, func(i, j int) bool {
		if products[i].WhatsappClickCount != products[j].WhatsappClickCount {
			return products[i].WhatsappClickCount > products[j].WhatsappClickCount
		}
		if products[i].ViewCount != products[j].ViewCount {
			return products[i].ViewCount > products[j].ViewCount
		}
		return products[i].ID < products[j].ID
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"summary":  summary,
		"products": products,
		"daily":    daily,
	})
}
