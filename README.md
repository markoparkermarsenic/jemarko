# Jemarko Wedding 
this is a wedding website that allows users to rsvp; once they have rsvpd the will be able to select an avatar and leave an optional message, this avatar will join the other guests avatars as they walk around the screen; the avatar plaza will be visible in the backround of the rsvp process. on the 1st of agaust 2026 the rsvp for will close and only the plaza will be visible 


## Requirements:
- the website will be mobile first
- the assets will be hand drawn pngs

### rsvp requirements
- organiser can input the guest list as a csv
- the user will input their name, if in the guest list they will be allowed to input their email
- a confirmation email of who has rsvp'd will be sent to the user
- a user can rsvp on behalf of multiple people

### avatar plaza requirements
- on completion of rsvp the user will be given a selection screen of different avatars, for each user rvpd there should be a selection 
- a user can leave an optionally leave a message, with a 140 character limit 
- useres will be able to see other users avatars with thier messages periodically displaying 
- the plaza will be drawn as the backround to the rsvp process
- users avatars will move around randomly 



## Tech stack 
- Frontend: Svelte (Vercel)
- Backend: Go (Vercel Serverless Functions) 
- Database: Supabase 
- Email: Resend 
- Domain: Cloudflare

## Project Structure

```
jemarko/
├── src/                    # SvelteKit frontend
│   ├── routes/            # Pages and routes
│   └── lib/               # Components and utilities
├── api/                   # Go serverless functions
│   ├── health.go          # Health check endpoint
│   ├── verify-name.go     # Guest name verification
│   ├── submit-rsvp.go     # RSVP submission
│   └── shared/            # Shared utilities
│       ├── types.go       # Type definitions
│       ├── database.go    # Supabase integration
│       ├── email.go       # Email service (Resend)
│       └── utils.go       # Helper functions
├── vercel.json            # Vercel deployment config
└── DEPLOYMENT.md          # Comprehensive deployment guide
```

## Deployment

This project is optimized for Vercel deployment with:
- **No Docker required** - Uses Vercel's native Go and Node.js runtimes
- **Serverless architecture** - Auto-scales and costs nothing when idle
- **Global CDN** - Frontend served from edge locations worldwide
- **Supabase integration** - PostgreSQL database with REST API

📚 **See [DEPLOYMENT.md](./DEPLOYMENT.md) for complete deployment instructions**

### Quick Deploy

1. Push to GitHub
2. Import to Vercel from GitHub
3. Add environment variables in Vercel dashboard
4. Deploy automatically

## Development

### Local Development with Vercel

```bash
# Install Vercel CLI
npm i -g vercel

# Run locally (simulates Vercel environment)
vercel dev
```

### Frontend Only

```bash
cd src
npm install
npm run dev
```

## Environment Variables

Required for production:
- `SUPABASE_URL` - Your Supabase project URL
- `SUPABASE_API_KEY` - Your Supabase API key
- `RESEND_API_KEY` - Your Resend API key
- `FROM_EMAIL` - Sender email address
- `FROM_NAME` - Sender name
- `ADMIN_EMAIL` - Admin email for notifications

See `.env.example` files for details.
