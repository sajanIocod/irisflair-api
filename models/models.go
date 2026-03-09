package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a wedding card product
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Code        string             `bson:"code" json:"code"`
	Category    string             `bson:"category" json:"category"`
	Price       int                `bson:"price" json:"price"`
	Description string             `bson:"description" json:"description"`
	Images      []string           `bson:"images" json:"images"`
	Tiers       []PriceTier        `bson:"tiers" json:"tiers"`
	MinOrder    int                `bson:"minOrder" json:"minOrder"`
	Featured    bool               `bson:"featured" json:"featured"`
	Active      bool               `bson:"active" json:"active"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type PriceTier struct {
	MinQty int `bson:"minQty" json:"minQty"`
	MaxQty int `bson:"maxQty" json:"maxQty"`
	Price  int `bson:"price" json:"price"`
}

// Category represents a product category
type Category struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name   string             `bson:"name" json:"name"`
	Icon   string             `bson:"icon" json:"icon"`
	Order  int                `bson:"order" json:"order"`
	Active bool               `bson:"active" json:"active"`
}

// Testimonial represents a customer review
type Testimonial struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name   string             `bson:"name" json:"name"`
	Text   string             `bson:"text" json:"text"`
	Rating int                `bson:"rating" json:"rating"`
	Active bool               `bson:"active" json:"active"`
	Order  int                `bson:"order" json:"order"`
}

// SiteSettings represents site configuration
type SiteSettings struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BrandName      string             `bson:"brandName" json:"brandName"`
	Tagline        string             `bson:"tagline" json:"tagline"`
	WhatsappNumber string             `bson:"whatsappNumber" json:"whatsappNumber"`
	Email          string             `bson:"email" json:"email"`
	Phone          string             `bson:"phone" json:"phone"`
	Address        string             `bson:"address" json:"address"`
	BusinessHours  string             `bson:"businessHours" json:"businessHours"`
	Instagram      string             `bson:"instagram" json:"instagram"`
	YouTube        string             `bson:"youtube" json:"youtube"`
	Facebook       string             `bson:"facebook" json:"facebook"`
	GoogleBusiness string             `bson:"googleBusiness" json:"googleBusiness"`
	HeroTitle      string             `bson:"heroTitle" json:"heroTitle"`
	HeroSubtitle   string             `bson:"heroSubtitle" json:"heroSubtitle"`
	HeroImage      string             `bson:"heroImage" json:"heroImage"`
	// Announcement bar messages (scrolling marquee)
	Announcements  []string           `bson:"announcements" json:"announcements"`
	// Welcome popup
	PopupEnabled   bool               `bson:"popupEnabled" json:"popupEnabled"`
	PopupImage     string             `bson:"popupImage" json:"popupImage"`
	PopupTitle     string             `bson:"popupTitle" json:"popupTitle"`
	PopupText      string             `bson:"popupText" json:"popupText"`
	// Showcase boxes (Best Sellers / Latest Designs / Hot Picks)
	ShowcaseBoxes  []ShowcaseBox      `bson:"showcaseBoxes" json:"showcaseBoxes"`
	// Social proof notifications
	SocialProofs   []SocialProof      `bson:"socialProofs" json:"socialProofs"`
}

// ShowcaseBox represents a featured collection box on homepage
type ShowcaseBox struct {
	Title string `bson:"title" json:"title"`
	Image string `bson:"image" json:"image"`
	Link  string `bson:"link" json:"link"`
}

// SocialProof represents a social proof notification
type SocialProof struct {
	Name     string `bson:"name" json:"name"`
	City     string `bson:"city" json:"city"`
	Quantity int    `bson:"quantity" json:"quantity"`
	Product  string `bson:"product" json:"product"`
}

// Admin represents admin user
type Admin struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"password"` // hashed
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
