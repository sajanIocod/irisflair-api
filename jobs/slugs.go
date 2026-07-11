package jobs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/irisflair/api/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Slugify converts a product name to a URL-safe slug: lowercase ASCII
// alphanumerics with single hyphens ("Rose Gold Invite!" → "rose-gold-invite").
func Slugify(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prevDash := false
	for _, r := range name {
		isAlnum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlnum {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// UniqueSlug returns Slugify(name), appending -2, -3, … until no other
// product (excluding excludeID, if non-zero) already holds it.
func UniqueSlug(ctx context.Context, name string, excludeID primitive.ObjectID) (string, error) {
	base := Slugify(name)
	if base == "" {
		base = "card"
	}
	products := db.GetDB().Collection("products")
	for i := 1; i <= 100; i++ {
		candidate := base
		if i > 1 {
			candidate = fmt.Sprintf("%s-%d", base, i)
		}
		filter := bson.M{"slug": candidate}
		if !excludeID.IsZero() {
			filter["_id"] = bson.M{"$ne": excludeID}
		}
		n, err := products.CountDocuments(ctx, filter)
		if err != nil {
			return "", err
		}
		if n == 0 {
			return candidate, nil
		}
	}
	// Pathological collision count — fall back to the object id for uniqueness.
	return fmt.Sprintf("%s-%s", base, primitive.NewObjectID().Hex()[18:]), nil
}

// EnsureSlugs backfills slugs for products created before the field existed.
// Idempotent; run once at startup.
func EnsureSlugs() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	products := db.GetDB().Collection("products")
	cur, err := products.Find(ctx, bson.M{"$or": []bson.M{
		{"slug": bson.M{"$exists": false}},
		{"slug": ""},
	}})
	if err != nil {
		log.Printf("slug backfill: find error: %v", err)
		return
	}
	var missing []struct {
		ID   primitive.ObjectID `bson:"_id"`
		Name string             `bson:"name"`
	}
	if err := cur.All(ctx, &missing); err != nil {
		log.Printf("slug backfill: decode error: %v", err)
		return
	}
	if len(missing) == 0 {
		return
	}
	done := 0
	for _, p := range missing {
		slug, err := UniqueSlug(ctx, p.Name, p.ID)
		if err != nil {
			log.Printf("slug backfill: %q: %v", p.Name, err)
			continue
		}
		if _, err := products.UpdateOne(ctx,
			bson.M{"_id": p.ID}, bson.M{"$set": bson.M{"slug": slug}}); err != nil {
			log.Printf("slug backfill: update %q: %v", p.Name, err)
			continue
		}
		done++
	}
	log.Printf("✓ slug backfill: %d/%d products", done, len(missing))
}
