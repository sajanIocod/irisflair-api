package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/irisflair/api/db"
	"github.com/irisflair/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetFAQs returns all FAQs
func GetFAQs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("faqs")
	opts := options.Find().SetSort(bson.M{"order": 1})

	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		http.Error(w, "Failed to fetch FAQs", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	faqs := make([]models.FAQ, 0)
	if err := cursor.All(ctx, &faqs); err != nil {
		http.Error(w, "Failed to decode FAQs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(faqs)
}

// GetActiveFAQs returns only active FAQs
func GetActiveFAQs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("faqs")
	opts := options.Find().SetSort(bson.M{"order": 1})

	cursor, err := collection.Find(ctx, bson.M{"active": true}, opts)
	if err != nil {
		http.Error(w, "Failed to fetch FAQs", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	faqs := make([]models.FAQ, 0)
	if err := cursor.All(ctx, &faqs); err != nil {
		http.Error(w, "Failed to decode FAQs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(faqs)
}

// CreateFAQ creates a new FAQ
func CreateFAQ(w http.ResponseWriter, r *http.Request) {
	var faq models.FAQ
	if err := json.NewDecoder(r.Body).Decode(&faq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	faq.ID = primitive.NewObjectID()
	if faq.Order == 0 {
		faq.Order = 9999
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("faqs")
	result, err := collection.InsertOne(ctx, faq)
	if err != nil {
		http.Error(w, "Failed to create FAQ", http.StatusInternalServerError)
		return
	}

	faq.ID = result.InsertedID.(primitive.ObjectID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(faq)
}

// UpdateFAQ updates an existing FAQ
func UpdateFAQ(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid FAQ ID", http.StatusBadRequest)
		return
	}

	var updates bson.M
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("faqs")
	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updates})
	if err != nil {
		http.Error(w, "Failed to update FAQ", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "FAQ not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "FAQ updated"})
}

// DeleteFAQ deletes a FAQ
func DeleteFAQ(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid FAQ ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("faqs")
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Failed to delete FAQ", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "FAQ not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "FAQ deleted"})
}
