# Deployment Guide - Jemima & Marko Wedding Website

This guide will help you deploy your wedding website to Vercel with Supabase backend.

## Architecture Overview

```
┌─────────────────────────────────────────┐
│         Vercel Project                  │
├─────────────────────────────────────────┤
│  Frontend (SvelteKit)                   │
│  • Static pages + SSR                   │
│  • Served from global CDN               │
├─────────────────────────────────────────┤
│  Backend (Go Serverless Functions)      │
│  • /api/verify-name                     │
│  • /api/submit-rsvp                     │
│  • /api/health                          │
└─────────────────────────────────────────┘
            ↓
    ┌───────────────┐
    │   Supabase    │
    │  PostgreSQL   │
    └───────────────┘
```

## Prerequisites

1. **Vercel Account**: Sign up at https://vercel.com
2. **Supabase Project**: Set up at https://supabase.com
3. **Resend Account** (Optional): For email functionality at https://resend.com

## Supabase Setup

### 1. Create Database Tables

Run these SQL commands in Supabase SQL Editor:

```sql
-- Create guests table
CREATE TABLE guests (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create rsvps table
CREATE TABLE rsvps (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  is_attending BOOLEAN NOT NULL,
  attending_guests TEXT[] DEFAULT '{}',
  diet TEXT,
  submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_guests_name ON guests(name);
CREATE INDEX idx_rsvps_email ON rsvps(email);
CREATE INDEX idx_rsvps_submitted_at ON rsvps(submitted_at DESC);
```

### 2. Populate Guest List

You can import your guest list from CSV or add manually:

```sql
INSERT INTO guests (name) VALUES
  ('John Smith'),
  ('Jane Smith'),
  ('Bob Johnson');
```

### 3. Get Supabase Credentials

1. Go to Project Settings → API
2. Copy your:
   - Project URL (e.g., `https://xxxxx.supabase.co`)
   - `anon` or `service_role` API key

## Vercel Deployment

### Option 1: Deploy via GitHub (Recommended)

1. **Push to GitHub**
   ```bash
   git add .
   git commit -m "Prepare for Vercel deployment"
   git push origin main
   ```

2. **Connect to Vercel**
   - Go to https://vercel.com/new
   - Import your GitHub repository
   - Vercel will auto-detect SvelteKit and Go

3. **Configure Environment Variables**
   
   In Vercel Project Settings → Environment Variables, add:

   **Required:**
   - `SUPABASE_URL` = Your Supabase project URL
   - `SUPABASE_API_KEY` = Your Supabase API key

   **Optional (for email functionality):**
   - `RESEND_API_KEY` = Your Resend API key
   - `FROM_EMAIL` = Sender email (e.g., `wedding@yourdomain.com`)
   - `FROM_NAME` = Sender name (e.g., `Jemima & Marko Wedding`)
   - `ADMIN_EMAIL` = Your email to receive unlisted guest notifications

4. **Deploy**
   - Click "Deploy"
   - Vercel will build and deploy automatically
   - Future git pushes will auto-deploy

### Option 2: Deploy via Vercel CLI

1. **Install Vercel CLI**
   ```bash
   npm i -g vercel
   ```

2. **Login**
   ```bash
   vercel login
   ```

3. **Deploy**
   ```bash
   vercel --prod
   ```

4. **Add Environment Variables**
   ```bash
   vercel env add SUPABASE_URL
   vercel env add SUPABASE_API_KEY
   vercel env add RESEND_API_KEY
   vercel env add FROM_EMAIL
   vercel env add FROM_NAME
   vercel env add ADMIN_EMAIL
   ```

## Local Development

### Setup

1. **Install Dependencies**
   ```bash
   # Frontend
   cd src
   npm install
   
   # Backend
   cd ../api
   go mod download
   ```

2. **Configure Environment Variables**
   
   Create `api/.env`:
   ```bash
   SUPABASE_URL=https://xxxxx.supabase.co
   SUPABASE_API_KEY=your_api_key
   RESEND_API_KEY=your_resend_key
   FROM_EMAIL=wedding@yourdomain.com
   FROM_NAME=Jemima & Marko Wedding
   ADMIN_EMAIL=your@email.com
   ```

### Run Locally with Vercel Dev

```bash
# From project root
vercel dev
```

This will:
- Start SvelteKit dev server
- Run Go functions locally
- Simulate Vercel environment

### Run Separately (Alternative)

**Terminal 1 - Frontend:**
```bash
cd src
npm run dev
```

**Terminal 2 - Backend:**
```bash
cd api
go run main.go  # Note: You'll need to create a local dev server for this
```

## API Endpoints

Once deployed, your API will be available at:

- `https://yourdomain.vercel.app/api/health` - Health check
- `https://yourdomain.vercel.app/api/verify-name` - Verify guest name
- `https://yourdomain.vercel.app/api/submit-rsvp` - Submit RSVP

## Updating the Frontend API URL

Update the API base URL in your frontend code to point to your Vercel deployment:

```javascript
// In your Svelte component
const API_BASE = import.meta.env.PROD 
  ? 'https://yourdomain.vercel.app/api'
  : 'http://localhost:3000/api';
```

Or set it in `src/.env`:
```
PUBLIC_API_URL=https://yourdomain.vercel.app/api
```

## Custom Domain (Optional)

1. Go to Vercel Project Settings → Domains
2. Add your custom domain (e.g., `jemarko.com`)
3. Update DNS records as instructed by Vercel
4. SSL certificate is automatically provisioned

## Monitoring & Logs

- **View Logs**: Vercel Dashboard → Your Project → Functions
- **Monitor Performance**: Vercel Analytics (enable in project settings)
- **Database Logs**: Supabase Dashboard → Logs

## Troubleshooting

### Build Fails

**Issue**: Go build fails
**Solution**: Ensure `go.mod` has correct module path:
```go
module github.com/markoparkermarsenic/jemarko/api
```

**Issue**: SvelteKit build fails
**Solution**: Check `src/svelte.config.js` has `@sveltejs/adapter-vercel`

### Functions Timeout

Vercel serverless functions have a 10s timeout on Hobby plan.
- Optimize database queries
- Use connection pooling
- Upgrade to Pro plan for 60s timeout

### CORS Errors

Ensure CORS headers are set in all handler functions (already configured).

### Database Connection Issues

- Verify `SUPABASE_URL` and `SUPABASE_API_KEY` are set correctly
- Check Supabase project is not paused (auto-pauses after 1 week of inactivity on free tier)
- Verify database tables exist

## Production Checklist

- [ ] Supabase database tables created
- [ ] Guest list imported to Supabase
- [ ] Environment variables configured in Vercel
- [ ] Custom domain configured (optional)
- [ ] Email service (Resend) configured and tested
- [ ] Test all RSVP flows (attending, not attending, dietary requirements)
- [ ] Verify email confirmations are sent
- [ ] Check unlisted guest notifications work
- [ ] Monitor logs for first few days

## Cost Estimates

**Vercel (Hobby Plan - FREE)**
- Unlimited deployments
- 100GB bandwidth/month
- Serverless function executions included

**Supabase (Free Tier)**
- 500MB database
- Unlimited API requests
- 2GB file storage

**Resend (Free Tier)**
- 100 emails/day
- 3,000 emails/month

Perfect for a wedding website! 🎉

## Support

For issues:
1. Check Vercel deployment logs
2. Check Supabase logs
3. Review function logs in Vercel dashboard
4. Test locally with `vercel dev`

## Next Steps

1. Deploy to Vercel
2. Test thoroughly
3. Share your wedding website URL! 💒
