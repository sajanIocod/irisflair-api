# Deployment Guide - Production Setup

Deploy the IrisFlair Go API to production using Railway, Render, or Heroku.

## Pre-Deployment Checklist

- [ ] `.env` file configured with secure values
- [ ] All environment variables set correctly
- [ ] MongoDB Atlas cluster created
- [ ] Local testing completed
- [ ] Git repository initialized and pushed
- [ ] No sensitive data in code or .env file
- [ ] API endpoints tested and working

## Option 1: Railway.app (Recommended)

Railway is the easiest option for Go apps with free tier and instant deployments.

### Setup Railway

1. **Create Railway account**
   - Go to https://railway.app
   - Sign up with GitHub

2. **Connect GitHub repository**
   - Click "Create New Project"
   - Select "Deploy from GitHub"
   - Connect your GitHub account
   - Select the `irisflair` repository

3. **Create MongoDB service**
   - In Railway project, click "Add Service"
   - Select "Database"
   - Choose "MongoDB"
   - Railway creates MongoDB instance automatically

4. **Get MongoDB connection string**
   - Click MongoDB service
   - Copy the `MONGO_URL` variable
   - Or construct: `mongodb+srv://username:password@cluster.mongodb.net/irisflair?retryWrites=true`

5. **Deploy Go API**
   - In Railway, click "Add Service" again
   - Select "GitHub Repo"
   - Railway auto-detects Go project
   - Click "Deploy"

6. **Configure environment variables**
   - Click the Go API service
   - Go to "Variables" tab
   - Add these variables:
     ```
     PORT=8080
     MONGODB_URI=<MongoDB connection string from step 4>
     DB_NAME=irisflair
     JWT_SECRET=<generate 32+ random characters>
     ADMIN_USERNAME=admin
     ADMIN_PASSWORD=<secure password>
     CLOUDINARY_CLOUD_NAME=<your cloud name>
     ```

7. **Get production API URL**
   - Railway generates automatic domain like: `api-production-xxx.up.railway.app`
   - Copy this URL for frontend configuration

### Update Frontend for Railway

In `irisflair-app/.env.production`:

```env
NEXT_PUBLIC_API_URL=https://api-production-xxx.up.railway.app/api
```

---

## Option 2: Render.com

Render offers free tier with GitHub integration.

### Setup Render

1. **Create Render account**
   - Go to https://render.com
   - Sign up with GitHub

2. **Create new Web Service**
   - Click "New +" → "Web Service"
   - Connect GitHub repo
   - Select the `irisflair` repository

3. **Configure service**
   - Name: `irisflair-api`
   - Runtime: `Go`
   - Build command: `go build -o app`
   - Start command: `./app`
   - Instance type: Free (for testing) or Paid

4. **Create MongoDB Database**
   - Click "New +" → "MongoDB"
   - Name: `irisflair`
   - Region: Same as API
   - Tier: Free (for testing)

5. **Configure environment variables**
   - In Web Service settings, go to "Environment"
   - Add these:
     ```
     PORT=8080
     MONGODB_URI=<MongoDB connection string from step 4>
     DB_NAME=irisflair
     JWT_SECRET=<generate 32+ random characters>
     ADMIN_USERNAME=admin
     ADMIN_PASSWORD=<secure password>
     CLOUDINARY_CLOUD_NAME=<your cloud name>
     ```

6. **Deploy**
   - Click "Create Web Service"
   - Render auto-deploys from GitHub

7. **Get production URL**
   - Render generates domain: `irisflair-api.onrender.com`

### Auto-Deploy on Push

Render auto-deploys when you push to GitHub:

```bash
git push origin main
# Render automatically rebuilds and deploys
```

---

## Option 3: Heroku

Heroku is popular but has removed free tier. Use paid plan or alternatives.

### Setup Heroku (Paid)

1. **Create Heroku account**
   - Go to https://heroku.com
   - Sign up

2. **Install Heroku CLI**

   ```bash
   brew install heroku
   heroku login
   ```

3. **Create app**

   ```bash
   cd irisflair-api
   heroku create irisflair-api
   ```

4. **Add MongoDB Atlas**
   - Go to https://www.mongodb.com/cloud/atlas
   - Create free cluster
   - Get connection string

5. **Set environment variables**

   ```bash
   heroku config:set PORT=8080
   heroku config:set MONGODB_URI="mongodb+srv://user:pass@cluster.mongodb.net/irisflair?retryWrites=true"
   heroku config:set DB_NAME=irisflair
   heroku config:set JWT_SECRET="your-32-char-secret-key"
   heroku config:set ADMIN_USERNAME=admin
   heroku config:set ADMIN_PASSWORD="secure-password"
   heroku config:set CLOUDINARY_CLOUD_NAME="your-cloud"
   ```

6. **Deploy**

   ```bash
   git push heroku main
   ```

7. **View logs**
   ```bash
   heroku logs --tail
   ```

---

## MongoDB Atlas Setup (For All Platforms)

MongoDB Atlas is required for production databases.

### Create MongoDB Atlas Cluster

1. **Create Atlas account**
   - Go to https://www.mongodb.com/cloud/atlas
   - Sign up (free tier available)

2. **Create organization and project**
   - Create organization: `irisflair`
   - Create project: `production`

3. **Create cluster**
   - Click "Build a Database"
   - Choose Free Tier
   - Cloud Provider: AWS
   - Region: Closest to your API server
   - Cluster Name: `irisflair-cluster`
   - Click "Create Cluster"

4. **Create database user**
   - Go to "Database Access"
   - Click "Add New Database User"
   - Username: `irisflair_user`
   - Password: Use strong password (save it!)
   - Click "Create User"

5. **Whitelist IP addresses**
   - Go to "Network Access"
   - Click "Add IP Address"
   - Add your IP or use `0.0.0.0/0` (allows all - less secure)
   - For production: add specific IPs of your API servers
   - Railway/Render IPs auto-allowed

6. **Get connection string**
   - Go to "Clusters"
   - Click "Connect"
   - Choose "Connect your application"
   - Copy connection string:
     ```
     mongodb+srv://irisflair_user:password@irisflair-cluster.mongodb.net/?retryWrites=true&w=majority
     ```
   - Replace `<password>` with user password
   - Set `database=irisflair` at end if not already set

---

## Database Initialization

### Create Indexes (Optional but Recommended)

After first deployment, create database indexes for faster queries.

Run this in MongoDB Atlas Web Shell or Compass:

```javascript
// Connect to MongoDB Atlas
// Go to Clusters → Connect → Connect with MongoDB Compass or Mongo Shell

// Create indexes for better performance
db.products.createIndex({ active: 1 });
db.products.createIndex({ category: 1 });
db.products.createIndex({ featured: 1 });
db.products.createIndex({ createdAt: -1 });

db.categories.createIndex({ order: 1 });
db.categories.createIndex({ active: 1 });

db.testimonials.createIndex({ order: 1 });
db.testimonials.createIndex({ active: 1 });

db.settings.createIndex({ name: 1 }, { unique: true });
```

---

## Domain Setup (Optional)

If you want a custom domain like `api.irisflair.com`:

### Railway Custom Domain

1. Service Settings → "Domains"
2. Add custom domain
3. Update DNS records (CNAME) in your domain registrar

### Render Custom Domain

1. Dashboard → Web Service → "Settings"
2. Custom Domain
3. Add domain and update DNS

### Heroku Custom Domain

```bash
heroku domains:add api.irisflair.com
# Then update DNS CNAME to xxx.herokuapp.com
```

---

## Environment Variables Reference

| Variable              | Production Value  | Notes                          |
| --------------------- | ----------------- | ------------------------------ |
| PORT                  | 8080              | Don't change, platform assigns |
| MONGODB_URI           | mongodb+srv://... | MongoDB Atlas connection       |
| DB_NAME               | irisflair         | Database name                  |
| JWT_SECRET            | 32+ random chars  | Use `openssl rand -base64 32`  |
| ADMIN_USERNAME        | admin             | Change in production           |
| ADMIN_PASSWORD        | Strong password   | 12+ chars, mixed case          |
| CLOUDINARY_CLOUD_NAME | your-cloud-name   | Your Cloudinary account        |

### Generate Secure JWT Secret

```bash
# macOS/Linux
openssl rand -base64 32

# Windows PowerShell
[System.Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes((1..32 | % {[char](Get-Random -Min 33 -Max 127))} | Join-String)))
```

---

## Testing Production Deployment

### Health Check

```bash
curl https://your-api-domain.com/health
# Expected: {"status":"ok"}
```

### Test Authentication

```bash
curl -X POST https://your-api-domain.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'
```

### Test Product Read (No Auth)

```bash
curl https://your-api-domain.com/api/products
# Expected: [] or products array
```

### Test Product Create (With Auth)

```bash
TOKEN="<token from login>"
curl -X POST https://your-api-domain.com/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Test","code":"TEST","price":99.99,"active":true}'
```

---

## Update Frontend for Production

Update `irisflair-app/.env.production`:

```env
NEXT_PUBLIC_API_URL=https://your-api-domain.com/api
```

Redeploy frontend to Vercel:

```bash
cd irisflair-app
git push origin main
# Vercel auto-deploys
```

---

## Monitoring & Logs

### Railway Logs

- Dashboard → Service → "Logs" tab
- Real-time request logs

### Render Logs

- Dashboard → Web Service → "Logs"
- Searchable log history

### Heroku Logs

```bash
heroku logs --tail
heroku logs --tail --app irisflair-api
```

### MongoDB Atlas Logs

- Cluster → "Monitoring"
- View query performance
- Check resource usage

---

## Troubleshooting Deployment

### API Returns 404 Error

- Check base URL is correct
- Verify NEXT_PUBLIC_API_URL in frontend
- Ensure routes are `/api/...` not `/...`

### Authentication Fails

- Check JWT_SECRET is set correctly
- Verify ADMIN_USERNAME and ADMIN_PASSWORD
- Ensure token format is `Bearer <token>`

### Database Connection Error

```
Failed to connect to MongoDB
```

- Check MONGODB_URI is correct
- Verify IP whitelist includes platform IPs
- Test connection with MongoDB Compass
- Check database user password

### Slow Responses

- Verify MongoDB indexes are created
- Check MongoDB resource usage
- Consider upgrading cluster tier
- View slow queries in Atlas monitoring

### Memory or Resource Issues

- Check logs for memory leaks
- Verify no large bulk operations
- Consider upgrading instance type
- Implement pagination for large queries

---

## Backup & Recovery

### Backup MongoDB Atlas

**Automatic Backups:**

- Atlas provides daily backups automatically
- Go to "Backup" tab in cluster
- View backup history

**Manual Backup:**

```bash
# Backup to file
mongodump --uri "mongodb+srv://user:pass@cluster.mongodb.net/irisflair" \
  --out ./backup

# Restore from backup
mongorestore --uri "mongodb+srv://user:pass@cluster.mongodb.net" \
  ./backup/irisflair
```

---

## Production Checklist

Before going live:

- [ ] API deployed and running
- [ ] MongoDB Atlas set up with backups
- [ ] Environment variables all configured
- [ ] Custom domain set up (optional)
- [ ] SSL/TLS enabled (automatic on Railway/Render)
- [ ] All API endpoints tested
- [ ] Admin authentication working
- [ ] Frontend connected to production API
- [ ] Logged in to admin panel
- [ ] Can create/read/update/delete products
- [ ] Can create/read/update/delete categories
- [ ] Can create/read/update/delete testimonials
- [ ] Settings can be updated
- [ ] Performance acceptable (< 500ms per request)
- [ ] No errors in logs
- [ ] Database backups enabled
- [ ] Team notified of launch

---

## Support

### Platform Support

- Railway: https://help.railway.app
- Render: https://render.com/docs
- Heroku: https://devcenter.heroku.com
- MongoDB Atlas: https://docs.mongodb.com/atlas

### Common Issues

1. Check application logs first
2. Verify all environment variables set
3. Test API endpoints with curl
4. Check MongoDB connection and data
5. Review deployment platform docs

## Next Steps After Deployment

1. **Monitor Performance**
   - Track response times
   - Monitor resource usage
   - Check error rates

2. **User Testing**
   - Have team test admin panel
   - Test on mobile/different browsers
   - Verify all features working

3. **Optimization**
   - Add database indexes if needed
   - Implement caching if needed
   - Monitor and optimize slow queries

4. **Maintenance**
   - Regular backups
   - Update dependencies
   - Monitor security patches
   - Keep logs for debugging

Congratulations on launching IrisFlair! 🚀
