package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/irisflair/api/db"
	"github.com/irisflair/api/models"
	"go.mongodb.org/mongo-driver/bson"
)

// GetSettings returns the main site settings
func GetSettings(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("settings")
	var settings models.SiteSettings
	
	if err := collection.FindOne(ctx, bson.M{"name": "main"}).Decode(&settings); err != nil {
		// Return default/empty settings if not found
		settings = models.SiteSettings{
			BrandName: "IrisFlair",
		}
	}

	// Ensure slices are never nil (return [] not null in JSON)
	if settings.Announcements == nil {
		settings.Announcements = make([]string, 0)
	}
	if settings.ShowcaseBoxes == nil {
		settings.ShowcaseBoxes = make([]models.ShowcaseBox, 0)
	}
	if settings.SocialProofs == nil {
		settings.SocialProofs = make([]models.SocialProof, 0)
	}
	if settings.HeroImages == nil {
		settings.HeroImages = make([]string, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSettings updates the main site settings
func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var updates bson.M
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("settings")
	
	// Upsert: update if exists, create if doesn't
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"name": "main"},
		bson.M{"$set": updates},
	)
	if err != nil {
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	// If no document was matched, insert a new one
	if result.MatchedCount == 0 {
		updates["name"] = "main"
		_, err := collection.InsertOne(ctx, updates)
		if err != nil {
			http.Error(w, "Failed to create settings", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settings updated"})
}
