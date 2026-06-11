package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// EnsureIndexes creates indexes used by the common query patterns.
// Index creation is idempotent — re-running is safe.
func EnsureIndexes() error {
	if Client == nil {
		return fmt.Errorf("cannot create indexes: MongoDB not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database := GetDB()

	indexes := map[string][]mongo.IndexModel{
		"products": {
			{Keys: bson.D{{Key: "active", Value: 1}, {Key: "createdAt", Value: -1}}},
			{Keys: bson.D{{Key: "category", Value: 1}}},
			{Keys: bson.D{{Key: "featured", Value: 1}}},
		},
		"categories": {
			{Keys: bson.D{{Key: "active", Value: 1}, {Key: "order", Value: 1}}},
		},
		"testimonials": {
			{Keys: bson.D{{Key: "active", Value: 1}, {Key: "order", Value: 1}}},
		},
		"faqs": {
			{Keys: bson.D{{Key: "active", Value: 1}, {Key: "order", Value: 1}}},
		},
		"settings": {
			{Keys: bson.D{{Key: "name", Value: 1}}},
		},
	}

	for collection, models := range indexes {
		if _, err := database.Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
			return fmt.Errorf("failed to create indexes on %s: %w", collection, err)
		}
	}

	log.Println("✓ Database indexes ensured")
	return nil
}
