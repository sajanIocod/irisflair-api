# IrisFlair Go API Backend

A high-performance REST API backend built with Go, Chi router, and MongoDB for the IrisFlair e-commerce platform. Designed to replace the Firebase Firestore backend for significantly faster performance (10-100x improvement expected).

## Architecture

```
irisflair-api/
├── handlers/          # HTTP request handlers for all endpoints
│   ├── products.go    # Product CRUD operations
│   ├── categories.go  # Category CRUD operations
│   ├── testimonials.go # Testimonial CRUD operations
│   ├── settings.go    # Site settings operations
│   └── auth.go        # Authentication (login, token generation)
├── middleware/        # HTTP middleware
│   └── auth.go        # JWT validation, CORS, error recovery
├── db/                # Database operations
│   └── mongodb.go     # MongoDB connection and lifecycle
├── models/            # Data structures
│   └── models.go      # MongoDB schema definitions
├── main.go            # Server setup and route configuration
├── go.mod             # Go module and dependencies
├── .env.example       # Environment variables template
└── README.md          # This file
```

## Features

- **Fast REST API** - Built with Chi v5 for minimal overhead
- **MongoDB Integration** - Flexible document storage with proper indexing
- **JWT Authentication** - Secure admin endpoints with token-based auth
- **CORS Support** - Allow cross-origin requests from frontend
- **Error Recovery** - Panic recovery middleware for stability
- **Timeout Protection** - All database queries have 5-second timeout
- **Type-Safe** - Full Go type safety with proper error handling

## Getting Started

### Prerequisites

- Go 1.21+ installed
- MongoDB 4.4+ running locally or connection string to MongoDB Atlas
- Environment variables configured

### Installation

1. **Clone and navigate to API directory**

```bash
cd irisflair-api
```

2. **Install dependencies**

```bash
go mod download
```

3. **Create .env file** (copy from .env.example)

```bash
cp .env.example .env
```

4. **Update .env with your values**

```env
PORT=8080
MONGODB_URI=mongodb://localhost:27017  # or your MongoDB Atlas connection string
DB_NAME=irisflair
JWT_SECRET=your-secret-key-min-32-chars-recommended
ADMIN_USERNAME=admin
ADMIN_PASSWORD=secure-password-here
CLOUDINARY_CLOUD_NAME=your-cloud-name  # for image uploads
```

5. **Run the server**

```bash
go run main.go
```

Server will start on `http://localhost:8080`

## API Endpoints

### Authentication

#### Login

```
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password"
}

Response:
{
  "token": "eyJhbGc...",
  "expiresAt": 1234567890
}
```

### Products

#### Get All Products

```
GET /api/products

Response: [
  {
    "_id": "507f1f77bcf86cd799439011",
    "name": "Product Name",
    "code": "PROD001",
    "category": "Category ID",
    "price": 99.99,
    "description": "Product description",
    "images": ["url1", "url2"],
    "tiers": [
      {
        "minQty": 1,
        "maxQty": 10,
        "price": 99.99
      }
    ],
    "minOrder": 1,
    "featured": false,
    "active": true,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Active Products Only

```
GET /api/products/active
```

#### Get Single Product

```
GET /api/products/{id}
```

#### Create Product (Admin Only)

```
POST /api/products
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "New Product",
  "code": "NEW001",
  "category": "507f1f77bcf86cd799439011",
  "price": 149.99,
  "description": "Product description",
  "images": ["url1"],
  "minOrder": 5,
  "active": true
}
```

#### Update Product (Admin Only)

```
PUT /api/products/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "price": 199.99,
  "featured": true
}
```

#### Delete Product (Admin Only)

```
DELETE /api/products/{id}
Authorization: Bearer {token}
```

### Categories

#### Get All Categories

```
GET /api/categories

Response: [
  {
    "_id": "507f1f77bcf86cd799439011",
    "name": "Category Name",
    "icon": "emoji-or-icon-code",
    "order": 1,
    "active": true
  }
]
```

#### Get Active Categories Only

```
GET /api/categories/active
```

#### Create Category (Admin Only)

```
POST /api/categories
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "New Category",
  "icon": "🎁",
  "order": 1,
  "active": true
}
```

#### Update Category (Admin Only)

```
PUT /api/categories/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Updated Name",
  "order": 2
}
```

#### Delete Category (Admin Only)

```
DELETE /api/categories/{id}
Authorization: Bearer {token}
```

### Testimonials

#### Get All Testimonials

```
GET /api/testimonials

Response: [
  {
    "_id": "507f1f77bcf86cd799439011",
    "name": "Customer Name",
    "text": "Great product!",
    "rating": 5,
    "active": true,
    "order": 1
  }
]
```

#### Get Active Testimonials Only

```
GET /api/testimonials/active
```

#### Create Testimonial (Admin Only)

```
POST /api/testimonials
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "New Customer",
  "text": "Excellent quality",
  "rating": 4,
  "active": true,
  "order": 1
}
```

#### Update Testimonial (Admin Only)

```
PUT /api/testimonials/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "active": false
}
```

#### Delete Testimonial (Admin Only)

```
DELETE /api/testimonials/{id}
Authorization: Bearer {token}
```

### Settings

#### Get Settings

```
GET /api/settings

Response: {
  "_id": "507f1f77bcf86cd799439011",
  "name": "main",
  "brandName": "IrisFlair",
  "tagline": "Your tagline",
  "whatsappNumber": "+1234567890",
  "email": "contact@irisflair.com",
  "phone": "+1234567890",
  "address": "123 Main St",
  "businessHours": "9AM-6PM",
  "instagram": "@irisflair",
  "youtube": "irisflair",
  "facebook": "irisflair",
  "googleBusiness": "irisflair-link",
  "heroTitle": "Welcome to IrisFlair",
  "heroSubtitle": "Premium products",
  "heroImage": "image-url"
}
```

#### Update Settings (Admin Only)

```
PUT /api/settings
Authorization: Bearer {token}
Content-Type: application/json

{
  "tagline": "New tagline",
  "whatsappNumber": "+1987654321"
}
```

### Health Check

#### Check API Health

```
GET /health

Response: {"status":"ok"}
```

## Environment Variables

| Variable              | Required | Default   | Description                                       |
| --------------------- | -------- | --------- | ------------------------------------------------- |
| PORT                  | No       | 8080      | Server port                                       |
| MONGODB_URI           | Yes      | -         | MongoDB connection string                         |
| DB_NAME               | Yes      | irisflair | Database name                                     |
| JWT_SECRET            | Yes      | -         | Secret for JWT signing (min 32 chars recommended) |
| ADMIN_USERNAME        | Yes      | -         | Admin username for login                          |
| ADMIN_PASSWORD        | Yes      | -         | Admin password for login                          |
| CLOUDINARY_CLOUD_NAME | No       | -         | Cloudinary integration (for future image uploads) |

## Deployment

### Option 1: Railway.app (Recommended)

1. Push code to GitHub
2. Connect Railway to GitHub repo
3. Add MongoDB Atlas connection string as variable
4. Set all environment variables
5. Deploy

### Option 2: Render.com

1. Create new Web Service
2. Connect GitHub repo
3. Build command: `go build -o app`
4. Start command: `./app`
5. Add environment variables
6. Deploy

### Option 3: Heroku

```bash
heroku create irisflair-api
heroku config:set MONGODB_URI=mongodb+srv://...
heroku config:set JWT_SECRET=...
heroku config:set ADMIN_USERNAME=admin
heroku config:set ADMIN_PASSWORD=...
git push heroku main
```

### Option 4: VPS (DigitalOcean, AWS EC2, Linode)

1. SSH into server
2. Install Go 1.21+
3. Clone repository
4. Create `.env` file with production values
5. Run `go build -o app && ./app` or use systemd service

## MongoDB Setup

### Local MongoDB

```bash
# Install MongoDB Community Edition
# macOS with Homebrew
brew tap mongodb/brew
brew install mongodb-community

# Start MongoDB
brew services start mongodb-community

# Stop MongoDB
brew services stop mongodb-community
```

### MongoDB Atlas (Cloud)

1. Create account at https://www.mongodb.com/cloud/atlas
2. Create cluster (free tier available)
3. Get connection string
4. Set `MONGODB_URI` in .env
5. Whitelist IP addresses in network access

## Development

### Run Tests

```bash
go test ./...
```

### Format Code

```bash
go fmt ./...
```

### Check Errors

```bash
go vet ./...
```

### Build Binary

```bash
go build -o app
```

### Run Binary

```bash
./app
```

## Database Indexing

MongoDB automatically uses ObjectID for fast lookups. For production, consider adding indexes:

```javascript
// Create indexes in MongoDB
db.products.createIndex({ active: 1 });
db.products.createIndex({ category: 1 });
db.products.createIndex({ createdAt: -1 });

db.categories.createIndex({ order: 1 });
db.categories.createIndex({ active: 1 });

db.testimonials.createIndex({ order: 1 });
db.testimonials.createIndex({ active: 1 });

db.settings.createIndex({ name: 1 }, { unique: true });
```

## Performance Optimization Tips

1. **Connection Pooling** - Go driver handles this automatically
2. **Query Optimization** - Ensure proper indexes exist
3. **Caching** - Next.js frontend handles caching (10-min TTL)
4. **Compression** - Add gzip middleware if needed
5. **Rate Limiting** - Add rate limiter for production

## Troubleshooting

### MongoDB Connection Error

```
Failed to connect to MongoDB
```

- Check MONGODB_URI is correct
- Ensure MongoDB is running locally or connection string is valid
- Verify network access in MongoDB Atlas

### JWT Token Invalid

```
Invalid or expired token
```

- Ensure JWT_SECRET is the same in auth and middleware
- Check token hasn't expired
- Verify Authorization header format: `Bearer {token}`

### 5-Second Timeout on Queries

```
context deadline exceeded
```

- Database query is too slow
- Add proper indexes
- Check database performance
- Consider MongoDB Atlas tier upgrade

## Frontend Integration

The Next.js frontend uses `/src/lib/api.ts` to call this API. Set the API URL:

```typescript
// .env.local in irisflair-app
NEXT_PUBLIC_API_URL=http://localhost:8080/api  # Development
NEXT_PUBLIC_API_URL=https://api.irisflair.com/api  # Production
```

## Migration from Firestore

See `MIGRATION.md` for steps to migrate data from Firestore to MongoDB.

## Contributing

1. Follow Go conventions and style
2. Add error handling for all operations
3. Write comments for exported functions
4. Test all endpoints before committing
5. Keep timeout values at 5 seconds

## License

MIT License - See LICENSE file for details

## Support

For issues, questions, or suggestions:

- Create an issue in the GitHub repository
- Email: support@irisflair.com
- WhatsApp: Check settings endpoint for business number
