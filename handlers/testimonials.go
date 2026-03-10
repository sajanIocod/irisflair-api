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

// GetTestimonials returns all testimonials
func GetTestimonials(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("testimonials")
	opts := options.Find().SetSort(bson.M{"order": 1})
	
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		http.Error(w, "Failed to fetch testimonials", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	testimonials := make([]models.Testimonial, 0)
	if err := cursor.All(ctx, &testimonials); err != nil {
		http.Error(w, "Failed to decode testimonials", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testimonials)
}

// GetActiveTestimonials returns only active testimonials
func GetActiveTestimonials(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("testimonials")
	opts := options.Find().SetSort(bson.M{"order": 1})
	
	cursor, err := collection.Find(ctx, bson.M{"active": true}, opts)
	if err != nil {
		http.Error(w, "Failed to fetch testimonials", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	testimonials := make([]models.Testimonial, 0)
	if err := cursor.All(ctx, &testimonials); err != nil {
		http.Error(w, "Failed to decode testimonials", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testimonials)
}

// CreateTestimonial creates a new testimonial
func CreateTestimonial(w http.ResponseWriter, r *http.Request) {
	var testimonial models.Testimonial
	if err := json.NewDecoder(r.Body).Decode(&testimonial); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	testimonial.ID = primitive.NewObjectID()
	if testimonial.Order == 0 {
		testimonial.Order = 9999 // Default order
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("testimonials")
	result, err := collection.InsertOne(ctx, testimonial)
	if err != nil {
		http.Error(w, "Failed to create testimonial", http.StatusInternalServerError)
		return
	}

	testimonial.ID = result.InsertedID.(primitive.ObjectID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(testimonial)
}

// UpdateTestimonial updates an existing testimonial
func UpdateTestimonial(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid testimonial ID", http.StatusBadRequest)
		return
	}

	var updates bson.M
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("testimonials")
	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updates})
	if err != nil {
		http.Error(w, "Failed to update testimonial", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Testimonial not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Testimonial updated"})
}

// DeleteTestimonial deletes a testimonial
func DeleteTestimonial(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid testimonial ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.GetDB().Collection("testimonials")
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Failed to delete testimonial", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Testimonial not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Testimonial deleted"})
}
