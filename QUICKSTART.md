# Quick Start Guide - Local Development

Get the IrisFlair API running on your machine in 5 minutes.

## Prerequisites

- Go 1.21+ ([download](https://golang.org/dl))
- MongoDB Community Edition ([install](https://docs.mongodb.com/manual/installation/))
- Postman or curl (for testing)

## Setup (First Time)

### 1. Start MongoDB

**macOS with Homebrew:**

```bash
brew services start mongodb-community
```

**Linux (Ubuntu):**

```bash
sudo systemctl start mongod
```

**Windows:**

```bash
# MongoDB should auto-start after installation
# Or manually run: "C:\Program Files\MongoDB\Server\5.0\bin\mongod.exe"
```

**Verify MongoDB is running:**

```bash
mongo
> db.adminCommand("ping")
```

### 2. Setup API Environment

```bash
cd irisflair-api

# Copy example env file
cp .env.example .env

# Edit .env with your values (optional - defaults work for local dev)
# nano .env
```

Your `.env` should look like:

```env
PORT=8080
MONGODB_URI=mongodb://localhost:27017
DB_NAME=irisflair
JWT_SECRET=your-secret-key-min-32-chars-recommended
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
CLOUDINARY_CLOUD_NAME=your_cloud_name
```

### 3. Download Dependencies

```bash
go mod download
```

### 4. Run the Server

```bash
go run main.go
```

You should see:

```
Connecting to MongoDB...
Starting server on port 8080...
```

Server is now running at `http://localhost:8080`

## Testing the API

### Option A: Using curl (Terminal)

**1. Check server is alive:**

```bash
curl http://localhost:8080/health
# Expected: {"status":"ok"}
```

**2. Admin login:**

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

# Expected response:
# {
#   "token": "eyJhbGc...",
#   "expiresAt": 1234567890
# }
```

**3. Get all products:**

```bash
curl http://localhost:8080/api/products
# Expected: [] (empty array initially)
```

**4. Create a product (use token from login):**

```bash
TOKEN="your-token-from-login-response"

curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Test Product",
    "code": "TEST001",
    "price": 99.99,
    "description": "A test product",
    "images": ["https://example.com/image.jpg"],
    "minOrder": 1,
    "active": true
  }'
```

**5. Get all products again:**

```bash
curl http://localhost:8080/api/products
# Expected: [{ your product }]
```

### Option B: Using Postman (GUI)

**1. Import collection:**

- Open Postman
- Create new request
- Set method to `GET`
- URL: `http://localhost:8080/health`
- Click Send
- Should see: `{"status":"ok"}`

**2. Create login request:**

- New request: `POST`
- URL: `http://localhost:8080/api/auth/login`
- Body (raw JSON):
  ```json
  {
    "username": "admin",
    "password": "admin123"
  }
  ```
- Click Send
- Copy the `token` from response

**3. Create product request:**

- New request: `POST`
- URL: `http://localhost:8080/api/products`
- Headers: Add `Authorization: Bearer <your-token>`
- Body (raw JSON):
  ```json
  {
    "name": "Test Product",
    "code": "TEST001",
    "price": 99.99,
    "description": "A test product",
    "images": ["https://example.com/image.jpg"],
    "minOrder": 1,
    "active": true
  }
  ```
- Click Send

**4. Get products:**

- New request: `GET`
- URL: `http://localhost:8080/api/products`
- Click Send
- Should see your created product

## Using Makefile (Convenience)

```bash
# Show all available commands
make help

# Download dependencies
make install-deps

# Build the app
make build

# Run the app
make run

# Code quality
make fmt    # Format code
make vet    # Lint code
make test   # Run tests

# Clean up
make clean
```

## Common Development Tasks

### View MongoDB Data

```bash
# Connect to MongoDB
mongo

# List databases
show dbs

# Use irisflair database
use irisflair

# See collections
show collections

# View products
db.products.find()

# Count products
db.products.countDocuments()

# Find specific product
db.products.findOne({ "name": "Test Product" })

# Update product
db.products.updateOne(
  { "_id": ObjectId("...") },
  { $set: { "active": false } }
)

# Delete product
db.products.deleteOne({ "_id": ObjectId("...") })

# Clear all products
db.products.deleteMany({})
```

### Check Server Logs

The server prints logs to terminal showing all requests and errors:

```
GET /api/products 200
POST /api/products 201 (requires auth)
DELETE /api/products/{id} 200
```

### Debug Mode

Add debug logging to your curl requests:

```bash
curl -v http://localhost:8080/health
# Shows request/response headers
```

## Connecting the Frontend

Update your Next.js frontend to use the local API:

**Create `.env.local` in `irisflair-app`:**

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

Then the frontend will call your local API instead of Firestore!

## Troubleshooting

### "Connection refused" error

```
Problem: API won't start
Solution:
  - Check MongoDB is running: `brew services list`
  - Check port 8080 is free: `lsof -i :8080`
  - Try different port: PORT=3001 go run main.go
```

### "cannot find module" error

```
Problem: Missing dependencies
Solution:
  - Run: go mod download
  - Or: go mod tidy
```

### Database "irisflair" not found

```
Problem: MongoDB created but no data
Solution: This is normal! Collections are auto-created on first insert
  - Create a product first
  - Or manually: use irisflair (in mongo shell)
```

### Port 8080 already in use

```
Problem: Another process using port 8080
Solution:
  - List processes: lsof -i :8080
  - Kill: kill -9 <PID>
  - Or use different port: PORT=8081 go run main.go
```

### Token expired or invalid

```
Problem: 401 Unauthorized
Solution:
  - Get fresh token: curl -X POST .../auth/login
  - Ensure JWT_SECRET matches in .env
  - Check token format: "Bearer <token>"
```

## Next Steps

- [ ] API running locally
- [ ] Login endpoint working
- [ ] Can create products
- [ ] Frontend connected (update .env.local)
- [ ] All CRUD operations tested
- [ ] Ready to deploy!

## Commands Cheat Sheet

```bash
# Terminal 1: Start MongoDB
brew services start mongodb-community

# Terminal 2: Run API
cd irisflair-api
go run main.go

# Terminal 3: Test endpoints
curl http://localhost:8080/health

# View MongoDB data
mongo
> use irisflair
> db.products.find()
```

## Performance Notes

Your local development setup:

- **Database:** MongoDB (local)
- **API Server:** Go (http://localhost:8080)
- **Expected Response Time:** 50-200ms per request
- **Data Limit:** Only limited by your machine's storage

This is 10-100x faster than Firebase Firestore!

## Ready for Production?

When moving to production:

1. Update `MONGODB_URI` to MongoDB Atlas connection string
2. Update `NEXT_PUBLIC_API_URL` to production API domain
3. Use strong `JWT_SECRET` (min 32 random chars)
4. Use secure `ADMIN_PASSWORD`
5. Set `CLOUDINARY_CLOUD_NAME` for image uploads
6. Deploy on Railway, Render, Heroku, or VPS

See `README.md` for detailed deployment instructions.

## Get Help

- Check logs in terminal running `go run main.go`
- Verify .env file has required values
- Test with curl first before using frontend
- Check MongoDB is connected: `mongo` command
- Review README.md for more details

Happy developing! 🚀
