package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/irisflair/api/db"
	"github.com/irisflair/api/jobs"
	"github.com/irisflair/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetProducts returns all products
func GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	collection := db.GetDB().Collection("products")
	opts := options.Find().SetSort(bson.M{"createdAt": -1})
	
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		log.Printf("GetProducts: find error: %v", err)
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	products := make([]models.Product, 0)
	if err := cursor.All(ctx, &products); err != nil {
		log.Printf("GetProducts: decode error: %v", err)
		http.Error(w, "Failed to decode products", http.StatusInternalServerError)
		return
	}

	for i := range products {
		normalizeProduct(&products[i])
	}

	writeJSON(w, http.StatusOK, products)
}

// GetActiveProducts returns only active products
func GetActiveProducts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	collection := db.GetDB().Collection("products")
	opts := options.Find().SetSort(bson.M{"createdAt": -1})
	
	cursor, err := collection.Find(ctx, bson.M{"active": true}, opts)
	if err != nil {
		log.Printf("GetActiveProducts: find error: %v", err)
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	products := make([]models.Product, 0)
	if err := cursor.All(ctx, &products); err != nil {
		log.Printf("GetActiveProducts: decode error: %v", err)
		http.Error(w, "Failed to decode products", http.StatusInternalServerError)
		return
	}

	for i := range products {
		normalizeProduct(&products[i])
	}

	writeJSON(w, http.StatusOK, products)
}

// GetProduct returns a single product by hex ObjectID or by slug.
func GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	filter := bson.M{"slug": id}
	if objID, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": objID}
	}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	collection := db.GetDB().Collection("products")
	var product models.Product

	if err := collection.FindOne(ctx, filter).Decode(&product); err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	normalizeProduct(&product)

	writeJSON(w, http.StatusOK, product)
}

// CreateProduct creates a new product
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		log.Printf("CreateProduct decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateProduct(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	product.ID = primitive.NewObjectID()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	// Computed fields are server-owned; a fresh product starts with none.
	// (Counters can't arrive via JSON — they're json:"-" — so zero values stand.)
	product.Badges = nil
	product.OftenEnquiredWith = nil

	// Slug is server-generated from the name and immutable afterwards.
	slugCtx, slugCancel := context.WithTimeout(context.Background(), queryTimeout)
	slug, err := jobs.UniqueSlug(slugCtx, product.Name, product.ID)
	slugCancel()
	if err != nil {
		log.Printf("CreateProduct: slug generation error: %v", err)
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}
	product.Slug = slug

	normalizeProduct(&product)

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	collection := db.GetDB().Collection("products")
	result, err := collection.InsertOne(ctx, product)
	if err != nil {
		log.Printf("CreateProduct insert error: %v", err)
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	product.ID = result.InsertedID.(primitive.ObjectID)

	writeJSON(w, http.StatusCreated, product)
}

// UpdateProduct updates an existing product
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var updates bson.M
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("UpdateProduct decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updates = sanitizeUpdates(updates, "_id", "id", "createdAt", "slug",
		"badges", "oftenEnquiredWith", "viewCount", "whatsappClickCount")
	if v, ok := updates["discountPercent"]; ok {
		f, isNum := v.(float64) // encoding/json decodes all JSON numbers as float64
		if !isNum || f < 0 || f > 90 {
			http.Error(w, "discountPercent must be between 0 and 90", http.StatusBadRequest)
			return
		}
	}
	updates["updatedAt"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	collection := db.GetDB().Collection("products")
	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updates})
	if err != nil {
		log.Printf("UpdateProduct update error: %v", err)
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Product updated"})
}

// DeleteProduct deletes a product
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	collection := db.GetDB().Collection("products")
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		log.Printf("DeleteProduct delete error: %v", err)
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Product deleted"})
}
