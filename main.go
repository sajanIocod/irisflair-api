package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/irisflair/api/db"
	"github.com/irisflair/api/handlers"
	"github.com/irisflair/api/middleware"
)

func main() {
	// Load environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Log config for debugging
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("WARNING: MONGODB_URI not set, using localhost")
	} else {
		log.Println("MONGODB_URI is set (Atlas)")
	}
	log.Printf("PORT=%s", port)

	// Connect to MongoDB
	log.Println("Connecting to MongoDB...")
	if err := db.Connect(); err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		log.Println("Server will start anyway - MongoDB connection will retry on requests")
	} else {
		defer db.Disconnect()
	}

	// Initialize router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.ErrorRecoveryMiddleware)

	// Public routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes
		r.Post("/auth/login", handlers.Login)

		// Products routes (public read, protected write)
		r.Get("/products", handlers.GetProducts)
		r.Get("/products/active", handlers.GetActiveProducts)
		r.Get("/products/{id}", handlers.GetProduct)

		r.With(middleware.AuthMiddleware).Post("/products", handlers.CreateProduct)
		r.With(middleware.AuthMiddleware).Put("/products/{id}", handlers.UpdateProduct)
		r.With(middleware.AuthMiddleware).Delete("/products/{id}", handlers.DeleteProduct)

		// Categories routes (public read, protected write)
		r.Get("/categories", handlers.GetCategories)
		r.Get("/categories/active", handlers.GetActiveCategories)

		r.With(middleware.AuthMiddleware).Post("/categories", handlers.CreateCategory)
		r.With(middleware.AuthMiddleware).Put("/categories/{id}", handlers.UpdateCategory)
		r.With(middleware.AuthMiddleware).Delete("/categories/{id}", handlers.DeleteCategory)

		// Testimonials routes (public read, protected write)
		r.Get("/testimonials", handlers.GetTestimonials)
		r.Get("/testimonials/active", handlers.GetActiveTestimonials)

		r.With(middleware.AuthMiddleware).Post("/testimonials", handlers.CreateTestimonial)
		r.With(middleware.AuthMiddleware).Put("/testimonials/{id}", handlers.UpdateTestimonial)
		r.With(middleware.AuthMiddleware).Delete("/testimonials/{id}", handlers.DeleteTestimonial)

		// Settings routes
		r.Get("/settings", handlers.GetSettings)
		r.With(middleware.AuthMiddleware).Put("/settings", handlers.UpdateSettings)
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Start server
	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
