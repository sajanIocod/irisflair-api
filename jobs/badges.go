// Package jobs holds background compute work. The only precedent in this
// codebase is the keep-alive ticker in main.go; Start follows the same
// ticker-in-a-goroutine pattern.
package jobs

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/irisflair/api/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Tunables for badge selection. All selection uses absolute thresholds and
// top-N cuts — no ratios, so empty data degrades to "no badge", never an error.
const (
	newWindow        = 30 * 24 * time.Hour  // "new" = created in the last 30 days
	trendWindow      = 14 * 24 * time.Hour  // trending looks at the last 14 days
	clickWeight      = 5                    // one WhatsApp click is worth 5 views
	minTrendScore    = 10                   // below this, nothing is "trending"
	minBestsellerAll = 5                    // min all-time clicks for "bestseller"
	topN             = 8                    // max products per badge
	pairWindow       = 180 * 24 * time.Hour // co-enquiry pairs from the last 180 days
	pairMaxDocs      = 5000                 // cap the slice the job reads, keeps cost flat
	minPairCount     = 2                    // noise floor for "often enquired together"
	maxCoProducts    = 4                    // FBT list length per product
	recomputeTimeout = 30 * time.Second
)

// Start runs Recompute immediately and then on every tick. Call only when
// MongoDB is connected. Panics are recovered so the job can never take the
// server down.
func Start(interval time.Duration) {
	go func() {
		run := func() {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("badge job: recovered from panic: %v", rec)
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), recomputeTimeout)
			defer cancel()
			if err := Recompute(ctx); err != nil {
				log.Printf("badge job: recompute failed: %v", err)
			}
		}
		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
	log.Printf("✓ Badge recompute job started (every %s)", interval)
}

// productSignals is the projection Recompute works with.
type productSignals struct {
	ID                 primitive.ObjectID `bson:"_id"`
	Active             bool               `bson:"active"`
	CreatedAt          time.Time          `bson:"createdAt"`
	ViewCount          int64              `bson:"viewCount"`
	WhatsappClickCount int64              `bson:"whatsappClickCount"`
	Badges             []string           `bson:"badges"`
	OftenEnquiredWith  []string           `bson:"oftenEnquiredWith"`
}

type pairKey struct{ a, b primitive.ObjectID } // a < b (hex order), unordered pair

// Recompute recalculates badges and often-enquired-with lists for all products
// and bulk-writes only the products whose values changed.
func Recompute(ctx context.Context) error {
	started := time.Now()
	database := db.GetDB()
	products := database.Collection("products")

	// 1. Load all products (projection only — the catalog is small).
	cur, err := products.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{
		"_id": 1, "active": 1, "createdAt": 1,
		"viewCount": 1, "whatsappClickCount": 1,
		"badges": 1, "oftenEnquiredWith": 1,
	}))
	if err != nil {
		return err
	}
	var all []productSignals
	if err := cur.All(ctx, &all); err != nil {
		return err
	}

	now := time.Now()

	// 2. Windowed event counts: one aggregation over the trending window.
	views14 := map[primitive.ObjectID]int64{}
	clicks14 := map[primitive.ObjectID]int64{}
	agg, err := database.Collection("product_events").Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"createdAt": bson.M{"$gte": now.Add(-trendWindow)}}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{"productId": "$productId", "type": "$type"},
			"n":   bson.M{"$sum": 1},
		}}},
	})
	if err != nil {
		return err
	}
	var grouped []struct {
		ID struct {
			ProductID primitive.ObjectID `bson:"productId"`
			Type      string             `bson:"type"`
		} `bson:"_id"`
		N int64 `bson:"n"`
	}
	if err := agg.All(ctx, &grouped); err != nil {
		return err
	}
	for _, g := range grouped {
		switch g.ID.Type {
		case "view":
			views14[g.ID.ProductID] = g.N
		case "whatsapp_click":
			clicks14[g.ID.ProductID] = g.N
		}
	}

	// 3. Badge selection.
	trendScore := func(p productSignals) int64 {
		return views14[p.ID] + clickWeight*clicks14[p.ID]
	}

	// Deterministic ordering helper: primary desc, then secondary desc,
	// then createdAt desc, then _id asc.
	rank := func(list []productSignals, primary, secondary func(productSignals) int64) {
		sort.Slice(list, func(i, j int) bool {
			a, b := list[i], list[j]
			if pa, pb := primary(a), primary(b); pa != pb {
				return pa > pb
			}
			if sa, sb := secondary(a), secondary(b); sa != sb {
				return sa > sb
			}
			if !a.CreatedAt.Equal(b.CreatedAt) {
				return a.CreatedAt.After(b.CreatedAt)
			}
			return a.ID.Hex() < b.ID.Hex()
		})
	}

	var trendCandidates, sellerCandidates []productSignals
	for _, p := range all {
		if !p.Active {
			continue
		}
		if trendScore(p) >= minTrendScore {
			trendCandidates = append(trendCandidates, p)
		}
		if p.WhatsappClickCount >= minBestsellerAll {
			sellerCandidates = append(sellerCandidates, p)
		}
	}
	rank(trendCandidates, trendScore, func(p productSignals) int64 { return clicks14[p.ID] })
	rank(sellerCandidates,
		func(p productSignals) int64 { return p.WhatsappClickCount },
		func(p productSignals) int64 { return p.ViewCount })
	if len(trendCandidates) > topN {
		trendCandidates = trendCandidates[:topN]
	}
	if len(sellerCandidates) > topN {
		sellerCandidates = sellerCandidates[:topN]
	}
	trending := map[primitive.ObjectID]bool{}
	for _, p := range trendCandidates {
		trending[p.ID] = true
	}
	bestseller := map[primitive.ObjectID]bool{}
	for _, p := range sellerCandidates {
		bestseller[p.ID] = true
	}

	// 4. Co-enquiry pairs from the recent slice of the enquiry log.
	pairCounts := map[pairKey]int{}
	enqCur, err := database.Collection("enquiry_events").Find(ctx,
		bson.M{"createdAt": bson.M{"$gte": now.Add(-pairWindow)}},
		options.Find().SetSort(bson.M{"createdAt": -1}).SetLimit(pairMaxDocs),
	)
	if err != nil {
		return err
	}
	var enquiries []struct {
		ProductIDs []primitive.ObjectID `bson:"productIds"`
	}
	if err := enqCur.All(ctx, &enquiries); err != nil {
		return err
	}
	for _, e := range enquiries {
		ids := e.ProductIDs
		for i := 0; i < len(ids); i++ {
			for j := i + 1; j < len(ids); j++ {
				a, b := ids[i], ids[j]
				if a == b {
					continue
				}
				if b.Hex() < a.Hex() {
					a, b = b, a
				}
				pairCounts[pairKey{a, b}]++
			}
		}
	}

	activeByID := map[primitive.ObjectID]productSignals{}
	for _, p := range all {
		if p.Active {
			activeByID[p.ID] = p
		}
	}

	type coProduct struct {
		id    primitive.ObjectID
		count int
	}
	coByProduct := map[primitive.ObjectID][]coProduct{}
	for pk, n := range pairCounts {
		if n < minPairCount {
			continue
		}
		// Only pairs where both sides still exist and are active.
		if _, ok := activeByID[pk.a]; !ok {
			continue
		}
		if _, ok := activeByID[pk.b]; !ok {
			continue
		}
		coByProduct[pk.a] = append(coByProduct[pk.a], coProduct{pk.b, n})
		coByProduct[pk.b] = append(coByProduct[pk.b], coProduct{pk.a, n})
	}
	fbt := map[primitive.ObjectID][]string{}
	for id, cos := range coByProduct {
		sort.Slice(cos, func(i, j int) bool {
			if cos[i].count != cos[j].count {
				return cos[i].count > cos[j].count
			}
			ca := activeByID[cos[i].id].WhatsappClickCount
			cb := activeByID[cos[j].id].WhatsappClickCount
			if ca != cb {
				return ca > cb
			}
			return cos[i].id.Hex() < cos[j].id.Hex()
		})
		if len(cos) > maxCoProducts {
			cos = cos[:maxCoProducts]
		}
		hexes := make([]string, len(cos))
		for i, c := range cos {
			hexes[i] = c.id.Hex()
		}
		fbt[id] = hexes
	}

	// 5. Diff and bulk-write only changed products. updatedAt is intentionally
	// untouched — it stays an admin-edit signal.
	var writes []mongo.WriteModel
	counts := map[string]int{}
	for _, p := range all {
		badges := make([]string, 0, 3)
		if p.Active && now.Sub(p.CreatedAt) <= newWindow {
			badges = append(badges, "new")
		}
		if trending[p.ID] {
			badges = append(badges, "trending")
		}
		if bestseller[p.ID] {
			badges = append(badges, "bestseller")
		}
		for _, b := range badges {
			counts[b]++
		}
		enquiredWith := fbt[p.ID]
		if enquiredWith == nil {
			enquiredWith = []string{}
		}
		if !equalStrings(badges, p.Badges) || !equalStrings(enquiredWith, p.OftenEnquiredWith) {
			writes = append(writes, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": p.ID}).
				SetUpdate(bson.M{"$set": bson.M{
					"badges":            badges,
					"oftenEnquiredWith": enquiredWith,
				}}))
		}
	}
	if len(writes) > 0 {
		if _, err := products.BulkWrite(ctx, writes); err != nil {
			return err
		}
	}

	log.Printf("badges recomputed: %d new, %d trending, %d bestseller, %d with FBT, %d updated (took %s)",
		counts["new"], counts["trending"], counts["bestseller"], len(fbt), len(writes), time.Since(started).Round(time.Millisecond))
	return nil
}

// equalStrings compares two slices treating nil and empty as equal.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
