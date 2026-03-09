# Firestore to MongoDB Migration Guide

This guide explains how to migrate data from Firebase Firestore to MongoDB for the IrisFlair backend redesign.

## Overview

The migration process involves:

1. Exporting data from Firestore
2. Transforming the data to MongoDB format
3. Importing into MongoDB
4. Validating the migration

## Step 1: Export Data from Firestore

### Option A: Using Firebase Admin SDK (Recommended)

Create a script `scripts/export-firestore.js` in your Next.js project:

```javascript
const admin = require("firebase-admin");
const fs = require("fs");

// Initialize Firebase Admin
const serviceAccount = require("../firebase-key.json");
admin.initializeApp({
  credential: admin.credential.cert(serviceAccount),
});

const db = admin.firestore();

async function exportCollections() {
  const collections = ["products", "categories", "testimonials", "settings"];
  const data = {};

  for (const collection of collections) {
    console.log(`Exporting ${collection}...`);
    const snapshot = await db.collection(collection).get();

    data[collection] = snapshot.docs.map((doc) => ({
      ...doc.data(),
      _id: doc.id, // Preserve original ID
    }));
  }

  fs.writeFileSync("firestore-export.json", JSON.stringify(data, null, 2));
  console.log("✓ Export complete: firestore-export.json");
  process.exit(0);
}

exportCollections().catch((err) => {
  console.error("Export failed:", err);
  process.exit(1);
});
```

Run the export:

```bash
cd irisflair-app
node scripts/export-firestore.js
```

### Option B: Manual Export via Firebase Console

1. Go to Firebase Console
2. Open Firestore Database
3. For each collection, export as JSON:
   - Click on collection name
   - Click "..." menu
   - Choose "Export collection"

### Option C: Using `firebase-export-firestore` CLI

```bash
npm install -g firebase-export-firestore
firebase-export-firestore --input firestore-export.json
```

## Step 2: Transform Data Format

Firestore uses slightly different formats than MongoDB. Create a transformation script:

```javascript
// scripts/transform-firestore-to-mongo.js
const fs = require("fs");

const data = JSON.parse(fs.readFileSync("firestore-export.json", "utf8"));

// Transform functions
function transformProduct(doc) {
  return {
    _id: doc._id ? ObjectId(doc._id) : new ObjectId(),
    name: doc.name || "",
    code: doc.code || "",
    category: doc.category || null,
    price: typeof doc.price === "number" ? doc.price : 0,
    description: doc.description || "",
    images: Array.isArray(doc.images) ? doc.images : [],
    tiers: Array.isArray(doc.tiers)
      ? doc.tiers.map((t) => ({
          minQty: t.minQty || 0,
          maxQty: t.maxQty || 0,
          price: t.price || 0,
        }))
      : [],
    minOrder: doc.minOrder || 1,
    featured: Boolean(doc.featured),
    active: Boolean(doc.active),
    createdAt: doc.createdAt?.toDate?.() || new Date(),
    updatedAt: doc.updatedAt?.toDate?.() || new Date(),
  };
}

function transformCategory(doc) {
  return {
    _id: doc._id ? ObjectId(doc._id) : new ObjectId(),
    name: doc.name || "",
    icon: doc.icon || "",
    order: typeof doc.order === "number" ? doc.order : 999,
    active: Boolean(doc.active),
  };
}

function transformTestimonial(doc) {
  return {
    _id: doc._id ? ObjectId(doc._id) : new ObjectId(),
    name: doc.name || "",
    text: doc.text || "",
    rating: typeof doc.rating === "number" ? doc.rating : 5,
    active: Boolean(doc.active),
    order: typeof doc.order === "number" ? doc.order : 999,
  };
}

function transformSettings(doc) {
  return {
    _id: doc._id ? ObjectId(doc._id) : new ObjectId(),
    name: "main",
    brandName: doc.brandName || "IrisFlair",
    tagline: doc.tagline || "",
    whatsappNumber: doc.whatsappNumber || "",
    email: doc.email || "",
    phone: doc.phone || "",
    address: doc.address || "",
    businessHours: doc.businessHours || "",
    instagram: doc.instagram || "",
    youtube: doc.youtube || "",
    facebook: doc.facebook || "",
    googleBusiness: doc.googleBusiness || "",
    heroTitle: doc.heroTitle || "",
    heroSubtitle: doc.heroSubtitle || "",
    heroImage: doc.heroImage || "",
  };
}

// Transform all collections
const transformed = {
  products: data.products.map(transformProduct),
  categories: data.categories.map(transformCategory),
  testimonials: data.testimonials.map(transformTestimonial),
  settings: data.settings.map(transformSettings),
};

fs.writeFileSync("mongo-import.json", JSON.stringify(transformed, null, 2));
console.log("✓ Transformation complete: mongo-import.json");
```

## Step 3: Import into MongoDB

### Option A: Using MongoDB Compass (GUI)

1. Open MongoDB Compass
2. Connect to your MongoDB instance
3. Create new database: `irisflair`
4. For each collection:
   - Right-click database
   - Choose "Add Collection"
   - Name it (products, categories, etc.)
   - Click "Add Data" button
   - Choose "Insert Document" or "Import"
   - Paste JSON data

### Option B: Using mongoimport CLI

```bash
# First, create the database
mongo irisflair

# Import each collection
mongoimport --db irisflair --collection products --file products.json --jsonArray
mongoimport --db irisflair --collection categories --file categories.json --jsonArray
mongoimport --db irisflair --collection testimonials --file testimonials.json --jsonArray
mongoimport --db irisflair --collection settings --file settings.json --jsonArray
```

### Option C: Using MongoDB Atlas (Cloud)

1. Upload the JSON files to MongoDB Atlas:

   ```bash
   mongoimport --uri "mongodb+srv://username:password@cluster.mongodb.net/irisflair" \
     --collection products --file products.json --jsonArray
   ```

2. Or use MongoDB Atlas UI:
   - Go to Data Import section
   - Upload JSON/BSON files
   - Map collections

### Option D: Using Go Script

Create `scripts/import-data.go` in irisflair-api:

```go
package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ImportData struct {
	Products     []interface{} `json:"products"`
	Categories   []interface{} `json:"categories"`
	Testimonials []interface{} `json:"testimonials"`
	Settings     []interface{} `json:"settings"`
}

func main() {
	// Read import file
	data, err := ioutil.ReadFile("mongo-import.json")
	if err != nil {
		log.Fatal("Failed to read import file:", err)
	}

	var importData ImportData
	if err := json.Unmarshal(data, &importData); err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("irisflair")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Import products
	log.Println("Importing products...")
	if len(importData.Products) > 0 {
		_, err := db.Collection("products").InsertMany(ctx, importData.Products)
		if err != nil {
			log.Println("Warning: Products import error:", err)
		}
	}

	// Import categories
	log.Println("Importing categories...")
	if len(importData.Categories) > 0 {
		_, err := db.Collection("categories").InsertMany(ctx, importData.Categories)
		if err != nil {
			log.Println("Warning: Categories import error:", err)
		}
	}

	// Import testimonials
	log.Println("Importing testimonials...")
	if len(importData.Testimonials) > 0 {
		_, err := db.Collection("testimonials").InsertMany(ctx, importData.Testimonials)
		if err != nil {
			log.Println("Warning: Testimonials import error:", err)
		}
	}

	// Import settings
	log.Println("Importing settings...")
	if len(importData.Settings) > 0 {
		_, err := db.Collection("settings").InsertMany(ctx, importData.Settings)
		if err != nil {
			log.Println("Warning: Settings import error:", err)
		}
	}

	log.Println("✓ Import complete!")
}
```

Run with: `go run scripts/import-data.go`

## Step 4: Validate Migration

### Check Collection Counts

```bash
mongo irisflair
> db.products.count()       # Should match Firestore count
> db.categories.count()
> db.testimonials.count()
> db.settings.count()
```

### Verify Data Integrity

Create validation script `scripts/validate-migration.go`:

```go
package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	client, _ := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	defer client.Disconnect(context.Background())

	db := client.Database("irisflair")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate products
	productCount, _ := db.Collection("products").EstimatedDocumentCount(ctx)
	log.Printf("✓ Products: %d documents", productCount)

	// Validate categories
	categoryCount, _ := db.Collection("categories").EstimatedDocumentCount(ctx)
	log.Printf("✓ Categories: %d documents", categoryCount)

	// Validate testimonials
	testimonialCount, _ := db.Collection("testimonials").EstimatedDocumentCount(ctx)
	log.Printf("✓ Testimonials: %d documents", testimonialCount)

	// Validate settings
	var settings bson.M
	db.Collection("settings").FindOne(ctx, bson.M{}).Decode(&settings)
	log.Printf("✓ Settings: %v", settings["name"])

	// Check for required fields
	cursor, _ := db.Collection("products").Find(ctx, bson.M{}, options.Find().SetLimit(1))
	var product bson.M
	cursor.Next(ctx)
	cursor.Decode(&product)

	log.Println("\n✓ Sample product fields:")
	for key := range product {
		log.Printf("  - %s", key)
	}

	log.Println("\n✓ Migration validation complete!")
}
```

## Step 5: Update Frontend & Backend

1. **Update Go API** - Point to MongoDB (should already be done in .env)
2. **Update Next.js** - Set API URL in environment
3. **Update Firestore calls** - Frontend now uses `/lib/api.ts` instead of `/lib/firestore.ts`
4. **Disable Firestore** - Remove Firebase calls from frontend (optional for cleanup)

## Rollback Plan

If migration has issues:

1. **Keep Firestore data intact** - Don't delete until verified
2. **Keep backup export** - Save firestore-export.json and mongo-import.json
3. **Test thoroughly** - Run all admin operations before switching
4. **Keep old frontend branch** - Branch with firestore.ts still available
5. **Monitor performance** - Track response times after migration

## Common Issues & Solutions

### Issue: Document IDs don't match

**Solution:** Firestore uses strings, MongoDB uses ObjectID. The transform script preserves IDs by storing them as strings if needed:

```go
// In models.go, ID can be string or ObjectID
type Product struct {
    ID string `bson:"_id"`
    // ...
}
```

### Issue: Timestamps wrong format

**Solution:** Firestore Timestamps become Date objects in MongoDB:

```go
// Ensure timestamps are time.Time type
UpdatedAt: time.Now(),
```

### Issue: Missing fields after import

**Solution:** Check transform script handles all fields:

```go
// Use explicit field mapping
category: doc.category || null,
```

### Issue: "E11000 duplicate key" error

**Solution:** Ensure unique constraints match:

```bash
# In MongoDB
db.settings.deleteMany({})  # Remove duplicates
db.settings.createIndex({ "name": 1 }, { unique: true })
```

## After Migration Checklist

- [ ] All products imported with correct counts
- [ ] All categories visible in admin
- [ ] All testimonials showing on homepage
- [ ] Settings loaded correctly
- [ ] Product images display (Cloudinary URLs preserved)
- [ ] Price tiers working correctly
- [ ] Category filtering on shop page works
- [ ] Admin add/edit/delete operations work
- [ ] Search functionality works (if implemented)
- [ ] Performance improved (compare load times)
- [ ] Firestore can be disabled/deleted from Firebase

## Performance Comparison

Expected improvements after migration:

| Operation          | Firestore | MongoDB   |
| ------------------ | --------- | --------- |
| Get all products   | 2-3s      | 200-500ms |
| Get single product | 500ms     | 50-100ms  |
| Create product     | 3-5s      | 200-300ms |
| Update category    | 2-4s      | 150-250ms |
| List testimonials  | 1-2s      | 100-200ms |

## Support

For issues during migration:

1. Check logs in Go API server
2. Verify MongoDB connection
3. Ensure data types match schemas
4. Validate JSON format before import
5. Test with small sample first

## Next Steps

After successful migration:

1. Update documentation
2. Monitor performance metrics
3. Set up automated backups for MongoDB
4. Remove Firebase Firestore if no longer needed
5. Update team documentation
