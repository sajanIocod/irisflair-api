package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/irisflair/api/models"
	"go.mongodb.org/mongo-driver/bson"
)

// Shared timeouts for database operations
const (
	queryTimeout = 5 * time.Second
	writeTimeout = 15 * time.Second
)

// writeJSON encodes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// normalizeProduct ensures nested slices are never nil so JSON returns [] instead of null.
func normalizeProduct(p *models.Product) {
	if p.Images == nil {
		p.Images = make([]string, 0)
	}
	if p.Tiers == nil {
		p.Tiers = make([]models.PriceTier, 0)
	}
	if p.PaperTypes == nil {
		p.PaperTypes = make([]models.PaperType, 0)
	}
	if p.ColorVariants == nil {
		p.ColorVariants = make([]models.ColorVariant, 0)
	}
	for i := range p.ColorVariants {
		if p.ColorVariants[i].Images == nil {
			p.ColorVariants[i].Images = make([]string, 0)
		}
	}
	if p.Tags == nil {
		p.Tags = make([]string, 0)
	}
	if p.Badges == nil {
		p.Badges = make([]string, 0)
	}
	if p.OftenEnquiredWith == nil {
		p.OftenEnquiredWith = make([]string, 0)
	}
}

// sanitizeUpdates removes dangerous keys from a client-supplied update document:
// MongoDB operators ($-prefixed), dotted paths, and immutable fields.
func sanitizeUpdates(updates bson.M, immutable ...string) bson.M {
	clean := bson.M{}
	for k, v := range updates {
		if strings.HasPrefix(k, "$") || strings.Contains(k, ".") {
			continue
		}
		blocked := false
		for _, im := range immutable {
			if k == im {
				blocked = true
				break
			}
		}
		if !blocked {
			clean[k] = v
		}
	}
	return clean
}

// validateProduct validates required product fields and value ranges.
func validateProduct(p *models.Product) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(p.Name) > 255 {
		return fmt.Errorf("name must be at most 255 characters")
	}
	if p.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	if p.DiscountPercent < 0 || p.DiscountPercent > 90 {
		return fmt.Errorf("discountPercent must be between 0 and 90")
	}
	if len(p.Description) > 10000 {
		return fmt.Errorf("description must be at most 10000 characters")
	}
	if p.MinOrder < 0 {
		return fmt.Errorf("minOrder cannot be negative")
	}
	for _, t := range p.Tiers {
		if t.Price < 0 {
			return fmt.Errorf("tier price cannot be negative")
		}
		if t.MinQty < 0 || t.MaxQty < 0 {
			return fmt.Errorf("tier quantities cannot be negative")
		}
	}
	return nil
}

// validateCategory validates required category fields.
func validateCategory(c *models.Category) error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.Name) > 100 {
		return fmt.Errorf("name must be at most 100 characters")
	}
	return nil
}

// validateTestimonial validates required testimonial fields.
func validateTestimonial(t *models.Testimonial) error {
	t.Name = strings.TrimSpace(t.Name)
	t.Text = strings.TrimSpace(t.Text)
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if t.Text == "" {
		return fmt.Errorf("text is required")
	}
	if t.Rating < 1 || t.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	return nil
}

// validateFAQ validates required FAQ fields.
func validateFAQ(f *models.FAQ) error {
	f.Question = strings.TrimSpace(f.Question)
	f.Answer = strings.TrimSpace(f.Answer)
	if f.Question == "" {
		return fmt.Errorf("question is required")
	}
	if f.Answer == "" {
		return fmt.Errorf("answer is required")
	}
	return nil
}
