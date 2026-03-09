# Next Steps - Frontend Integration & Deployment

This file outlines what needs to be done to complete the backend redesign and launch the new Go + MongoDB architecture.

## Current Status

✅ **Completed (Tasks 1-4):**

- Go API server with all endpoints
- MongoDB models and connection
- JWT authentication
- CORS and middleware
- TypeScript API client (lib/api.ts)
- Comprehensive documentation

⏳ **Pending (Tasks 5-7):**

- Update frontend to use new API
- Deploy to production
- Migrate Firestore data

---

## Task 5: Update Frontend to Use Go API

### Step 1: Configure API URL

**In `irisflair-app/.env.local`:**

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

### Step 2: Replace firestore.ts Imports

Files to update:

1. `src/app/admin/login/page.tsx`
   - Replace `firebase.ts` auth calls with `api.ts` login
2. `src/app/admin/products/page.tsx`
   - Replace `getProducts()` with `api.getProducts()`
   - Replace `createProduct()` with `api.createProduct()`
   - Replace `updateProduct()` with `api.updateProduct()`
   - Replace `deleteProduct()` with `api.deleteProduct()`

3. `src/app/admin/categories/page.tsx`
   - Replace category Firestore calls with api.ts equivalents
4. `src/app/admin/testimonials/page.tsx`
   - Replace testimonial Firestore calls with api.ts equivalents
5. `src/app/admin/settings/page.tsx`
   - Replace settings Firestore calls with api.ts equivalents
6. `src/app/(store)/page.tsx`
   - Replace featured products and testimonials calls
7. `src/app/(store)/shop/page.tsx`
   - Replace product list and category calls

8. `src/app/(store)/product/[id]/page.tsx`
   - Replace single product fetch

9. `src/components/CartSidebar.tsx`
   - Update if using any API calls

### Step 3: Update Authentication Context

`src/lib/auth-context.tsx`:

```typescript
// Replace Firebase auth with JWT token management
import { login, logout, getAuthToken, setAuthToken } from "./api";

// Update login handler:
async function handleLogin(username: string, password: string) {
  const response = await login(username, password);
  setAuthToken(response.token);
  // Redirect to admin dashboard
}

// Update logout handler:
function handleLogout() {
  clearAuthToken();
  // Redirect to login
}
```

### Step 4: Test Locally

```bash
# Terminal 1: Start Go API
cd irisflair-api
go run main.go

# Terminal 2: Start frontend
cd irisflair-app
npm run dev

# Browser: Test flows
# 1. Admin login
# 2. View products/categories/testimonials
# 3. Create new product
# 4. Update product
# 5. Delete product
# 6. View homepage (products should load)
```

### Step 5: Verify All Operations

**Admin Panel Tests:**

- [ ] Login works (uses api.login)
- [ ] Product list loads
- [ ] Can create product
- [ ] Can edit product
- [ ] Can delete product
- [ ] Category management works
- [ ] Testimonial management works
- [ ] Settings save correctly

**Store Tests:**

- [ ] Homepage loads
- [ ] Featured products display
- [ ] Testimonials section works
- [ ] Shop page loads products
- [ ] Category filter works
- [ ] Product detail page loads
- [ ] Add to cart works

### Step 6: Remove Firestore References

Once everything works:

1. Delete unused `src/lib/firestore.ts`
2. Remove Firebase from `src/lib/firebase.ts` (optional)
3. Update imports in all files
4. Run eslint to check for missed imports

---

## Task 6: Deploy API & MongoDB

### Step 1: Choose Deployment Platform

**Option A: Railway.app (Recommended)**

- Easiest setup
- Free tier available
- Auto-deploys from GitHub
- Integrated MongoDB option

**Option B: Render.com**

- Similar to Railway
- Good free tier
- GitHub integration

**Option C: Heroku**

- Paid option (free tier removed)
- Most popular
- Good documentation

### Step 2: Prepare for Deployment

**Create `.env` for production:**

```bash
cd irisflair-api
cp .env.example .env

# Update with real values:
# - MONGODB_URI from MongoDB Atlas
# - Strong JWT_SECRET
# - Secure ADMIN_PASSWORD
```

**Push to GitHub:**

```bash
git add .
git commit -m "Add Go API backend with MongoDB integration"
git push origin main
```

### Step 3: Deploy Go API

**If using Railway:**

1. Go to https://railway.app
2. Click "New Project"
3. Select "Deploy from GitHub"
4. Select your repository
5. Configure environment variables
6. Deploy!
7. Get API URL (e.g., `api-production-xxx.up.railway.app`)

**If using Render:**

1. Go to https://render.com
2. Click "New +" → "Web Service"
3. Connect GitHub
4. Build command: `go build -o app`
5. Start command: `./app`
6. Add environment variables
7. Deploy!

**If using Heroku:**

1. `heroku create irisflair-api`
2. Set environment variables: `heroku config:set KEY=value`
3. `git push heroku main`
4. Check logs: `heroku logs --tail`

### Step 4: Create MongoDB Atlas Database

1. Go to https://www.mongodb.com/cloud/atlas
2. Create free cluster
3. Create database user (save password!)
4. Whitelist IP addresses (0.0.0.0/0 for testing)
5. Get connection string
6. Copy connection string to deployment platform

### Step 5: Test Production Deployment

```bash
# Get your API domain from deployment platform
PROD_API="https://your-api-domain.com"

# Test health check
curl $PROD_API/health

# Test login
curl -X POST $PROD_API/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'

# Test product creation
curl -X POST $PROD_API/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"name":"Test","code":"T001","price":99.99,"active":true}'
```

### Step 6: Update Frontend for Production

**In `irisflair-app/.env.production`:**

```env
NEXT_PUBLIC_API_URL=https://your-api-domain.com/api
```

**Redeploy frontend to Vercel:**

```bash
git push origin main
# Vercel auto-detects and deploys
```

### Step 7: Verify Production Setup

- [ ] API health check responds
- [ ] Login endpoint works
- [ ] Can create products
- [ ] Can view products on homepage
- [ ] Admin panel works
- [ ] Database queries fast (< 500ms)
- [ ] No errors in logs

---

## Task 7: Migrate Firestore Data to MongoDB

### Step 1: Export Firestore Data

**Option A: Using admin script (easiest)**

```bash
# Create export script in irisflair-app/scripts/export-firestore.js
# (See MIGRATION.md for full script)

cd irisflair-app
node scripts/export-firestore.js
# Creates firestore-export.json
```

**Option B: Manual export**

- Firebase Console → Firestore → Collections
- Right-click each collection → Export
- Save as JSON

### Step 2: Transform Data Format

```bash
# Create and run transformation script
cd irisflair-api
node scripts/transform-firestore-to-mongo.js
# Creates mongo-import.json
```

### Step 3: Import into MongoDB

**Using mongoimport (command line):**

```bash
mongoimport --uri "mongodb+srv://user:pass@cluster.mongodb.net/irisflair" \
  --collection products \
  --file products.json \
  --jsonArray
```

**Using MongoDB Compass (GUI):**

1. Connect to MongoDB Atlas
2. Create collection
3. Upload JSON file
4. Import data

**Using MongoDB Atlas UI:**

1. Cluster → Data Import
2. Upload JSON files
3. Map to collections

### Step 4: Validate Migration

```bash
# Check document counts
mongo irisflair
> db.products.countDocuments()
> db.categories.countDocuments()
> db.testimonials.countDocuments()
> db.settings.countDocuments()

# Verify data integrity
> db.products.findOne()
> db.categories.findOne()
```

### Step 5: Test in Production

1. Visit admin panel
2. Check all products loaded
3. Verify all categories
4. Check testimonials
5. View homepage
6. Test add to cart
7. Verify product details

### Step 6: Monitor Performance

- Response times (should be < 500ms)
- Database connection status
- Error logs
- Feature functionality

---

## Deployment Checklist

### Pre-Deployment

- [ ] All local tests passing
- [ ] Frontend .env.local configured
- [ ] Go API running locally without errors
- [ ] MongoDB local connection working
- [ ] Git repository up to date

### Platform Setup

- [ ] Hosting platform chosen (Railway/Render/Heroku)
- [ ] Account created and project set up
- [ ] GitHub repository connected
- [ ] Environment variables configured
- [ ] MongoDB Atlas cluster created

### Deployment

- [ ] Go API deployed and accessible
- [ ] API health check responding
- [ ] MongoDB connection working
- [ ] Authentication endpoint tested
- [ ] API responding in < 500ms

### Frontend Integration

- [ ] Frontend .env.production updated with API URL
- [ ] Frontend redeployed to Vercel
- [ ] Admin login works
- [ ] Product list loads
- [ ] Can create/edit/delete products
- [ ] Homepage loads with data
- [ ] All features working

### Data Migration

- [ ] Firestore data exported
- [ ] Data transformed to MongoDB format
- [ ] Data imported to MongoDB Atlas
- [ ] Document counts verified
- [ ] Data consistency checked
- [ ] Backup of Firestore created

### Final Verification

- [ ] All admin operations working
- [ ] All store pages working
- [ ] Performance acceptable
- [ ] No errors in logs
- [ ] Backups configured
- [ ] Monitoring set up

---

## Quick Command Reference

```bash
# Local development
go run main.go              # Start API
npm run dev                # Start frontend

# Testing
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/auth/login ...

# Deployment
git push origin main        # Triggers railway/render auto-deploy
heroku logs --tail         # View heroku logs
mongo irisflair            # Access MongoDB locally

# Data migration
node export-firestore.js   # Export from Firestore
node transform-to-mongo.js # Transform format
mongoimport ...            # Import to MongoDB
```

---

## Expected Timeline

**Task 5: Frontend Integration**

- Estimate: 2-3 hours
- Update imports: 30 min
- Test all pages: 1 hour
- Fix any issues: 30 min - 1 hour

**Task 6: Deploy API & MongoDB**

- Estimate: 1-2 hours
- Platform setup: 15 min
- Deploy API: 15 min
- Configure MongoDB: 30 min
- Test production: 30 min

**Task 7: Data Migration**

- Estimate: 1-2 hours
- Export data: 15 min
- Transform format: 15 min
- Import to MongoDB: 15 min
- Validate data: 30 min

**Total Time: 4-7 hours**

---

## Common Issues & Solutions

### API Connection Errors

**Problem:** Frontend can't reach API
**Solution:**

- Check NEXT_PUBLIC_API_URL environment variable
- Verify API is running and accessible
- Check CORS headers if on different domain

### Authentication Failures

**Problem:** Login endpoint returns 401
**Solution:**

- Verify JWT_SECRET matches in .env
- Check ADMIN_USERNAME and ADMIN_PASSWORD
- Ensure token format is `Bearer <token>`

### Data Not Showing

**Problem:** Products/categories not loading
**Solution:**

- Check MongoDB connection
- Verify data was imported
- Check database name is correct
- Review API logs for errors

### Performance Issues

**Problem:** Slow responses in production
**Solution:**

- Create MongoDB indexes (see README.md)
- Check MongoDB resource usage
- Verify API logs for timeouts
- Consider upgrading MongoDB tier

---

## Success Criteria

✅ All admin operations working
✅ Homepage displaying data correctly
✅ Shop page functional
✅ Product detail pages loading
✅ Cart operations working
✅ API responses < 500ms
✅ No errors in logs
✅ Testimonials showing on homepage
✅ Settings displaying correctly
✅ Authentication working

---

## Post-Launch Tasks

After everything is deployed:

1. **Monitor Performance**
   - Track response times
   - Monitor database usage
   - Check error rates

2. **User Feedback**
   - Have team test
   - Gather feedback
   - Fix any issues

3. **Optimization**
   - Add database indexes if needed
   - Implement caching if needed
   - Optimize slow queries

4. **Maintenance**
   - Set up automated backups
   - Monitor for errors
   - Keep dependencies updated
   - Document API for future developers

---

## Support Resources

- **Go API README:** `irisflair-api/README.md`
- **Quick Start:** `irisflair-api/QUICKSTART.md`
- **Deployment Guide:** `irisflair-api/DEPLOYMENT.md`
- **Data Migration:** `irisflair-api/MIGRATION.md`
- **API Completeness:** `irisflair-api/COMPLETE.md`

---

## Next Immediate Action

1. **Start Task 5:**
   - Begin updating admin pages to use api.ts
   - Test each page locally with Go API running
   - Verify authentication flow

2. **Once Task 5 Complete:**
   - Choose deployment platform
   - Set up MongoDB Atlas
   - Deploy and test in production

3. **Once Deployed:**
   - Export and migrate Firestore data
   - Verify all data in MongoDB
   - Monitor production performance

---

You've got this! The backend redesign is complete. Now let's integrate it with the frontend and launch! 🚀
