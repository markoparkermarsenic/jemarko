# Jemarko Wedding API

Go backend for the Svelte-based wedding website with guest validation, email notifications, and RSVP management.

**🚀 Deployment Architecture:** This API is structured as Vercel Serverless Functions for easy deployment. See [DEPLOYMENT.md](../DEPLOYMENT.md) for full deployment guide.

## Features

### ✅ Guest Validation
- Validates guest names against a configurable guest list
- Case-insensitive name matching
- Loads guest list from Supabase or falls back to in-memory list for development

### 📧 Email Notifications
- **Guest Confirmation Emails**: Automatic RSVP confirmation via Resend
- **Admin Alerts**: Receive notifications when unlisted guests attempt to RSVP
- Includes request details (timestamp, IP address, user agent) for context

### 🗄️ Database Integration
- Supabase integration for guest list and RSVP storage
- Automatic fallback to in-memory storage for development
- Validates all RSVP guests against the master guest list

### 🔒 Security
- CORS middleware for cross-origin requests
- Email validation
- Guest list verification before accepting RSVPs
- Request logging for audit trails

## API Endpoints

### `GET /api/health`
Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1704223200,
  "guests": 5
}
```

### `POST /api/verify-name`
Validates if a name exists on the guest list.

**Request:**
```json
{
  "name": "John Smith"
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Guest found"
}
```

**Response (Not Found):**
```json
{
  "success": false,
  "message": "Name not found on the guest list. Please check the spelling or contact us."
}
```

**Note:** When a guest is not found, an email notification is automatically sent to the admin email address configured in `ADMIN_EMAIL`.

### `POST /api/submit-rsvp`
Submits an RSVP with validation.

**Request:**
```json
{
  "name": "John Smith",
  "email": "john@example.com",
  "attendingGuests": ["John Smith", "Jane Smith"]
}
```

**Response:**
```json
{
  "success": true,
  "message": "RSVP submitted successfully"
}
```

**Validation Rules:**
- Email must be valid format
- At least one guest must be attending
- All attending guests must exist in the guest list

## Setup Instructions

### 1. Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Edit `.env` with your values:

```bash
# Server Configuration
PORT=8080

# Supabase Database
SUPABASE_URL=https://your-project-id.supabase.co
SUPABASE_API_KEY=your-anon-key

# Resend Email Service
RESEND_API_KEY=re_xxxxxxxxxxxxxxxxxxxxxxxxxx
FROM_EMAIL=wedding@yourdomain.com
FROM_NAME=Jemima & Marko Wedding

# Admin Notifications
ADMIN_EMAIL=markoparkermarsenic@gmail.com
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up Database (Optional)

#### Database Connection Options

The current implementation uses **Supabase REST API** which requires:
- `SUPABASE_URL` - Your Supabase project URL
- `SUPABASE_API_KEY` - Your Supabase anon/public key

**Alternative:** If you have `PSQL_CONNECTION_STRING` configured in Vercel, note that the current code doesn't use it yet. The REST API approach is simpler and works well for this use case.

#### Create Required Tables

If using Supabase, create these tables in your Supabase SQL Editor:

**Guests Table:**
```sql
CREATE TABLE guests (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**RSVPs Table:**
```sql
CREATE TABLE rsvps (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  attending_guests TEXT[] NOT NULL,
  submitted_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Note:** Both the REST API and direct PostgreSQL connection access the same database tables.

### 4. Import Guest List

Create a `guests.csv` file with your guest list:

```csv
name
John Smith
Jane Smith
Bob Johnson
```

Then import using Supabase's CSV import feature or insert directly via SQL.

### 5. Run the Server

```bash
# Development
go run .

# Production
go build -o wedding-api
./wedding-api
```

The server will start on `http://localhost:8080` (or the port specified in `.env`).

## Admin Notifications Feature

### How It Works

When someone enters a name that's **not on the guest list**, the system automatically:

1. Logs the attempt to the console
2. Sends an email notification to the `ADMIN_EMAIL` address
3. Includes useful context:
   - The name that was entered
   - Timestamp of the attempt
   - IP address of the requester
   - User agent (browser/device info)

### Email Example

**Subject:** `⚠️  Unlisted Guest Attempt: Jennifer Unknown`

**Body:**
```
Hello,

Someone not on the guest list attempted to RSVP for your wedding.

Name Entered: Jennifer Unknown
Time: Sun, 01 Dec 2026 19:30:00 GMT
IP Address: 192.168.1.100
User Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...

This person was not found in your guest list. You may want to:
1. Check if this is a misspelling of an existing guest
2. Add them to the guest list if they should be invited
3. Contact them directly if needed

---
This is an automated notification from your wedding RSVP system.
```

### Configuration

Set `ADMIN_EMAIL` in your `.env` file:

```bash
ADMIN_EMAIL=markoparkermarsenic@gmail.com
```

**Note:** If `ADMIN_EMAIL` is not configured, the notification feature will be skipped (no emails sent, but attempts are still logged).

## Development Mode

Without Supabase/Resend configured:
- **Database**: Uses in-memory guest list with sample data
- **Emails**: Logs to console instead of sending actual emails
- **Admin Notifications**: Logs to console if `ADMIN_EMAIL` not set

This allows you to develop and test without external services.

## Guest List Management

### Simple Approach (Current)

The guest list is a flat list of names. When someone RSVPs:
1. They enter their name
2. System validates it exists in the guest list
3. They can RSVP for any other guests on the list (system validates each)

**Benefits:**
- Simple and flexible
- Users can RSVP for anyone invited
- No complex "allowed guests" mapping needed
- Validation ensures only invited guests can be added to RSVPs

### Example Flow

1. **John Smith** enters his name → ✅ Found on guest list
2. He proceeds to RSVP
3. He selects: "John Smith" and "Jane Smith" as attending
4. System validates both names exist → ✅ Both on guest list
5. RSVP is saved and confirmation email sent

## Tech Stack

- **Language:** Go 1.25.5
- **Email Service:** Resend API
- **Database:** Supabase (PostgreSQL)
- **Deployment:** Vercel Functions
- **Frontend:** Svelte (separate project)

## Deployment

### Vercel

1. Install Vercel CLI: `npm i -g vercel`
2. Run: `vercel`
3. Add environment variables in Vercel dashboard
4. Deploy: `vercel --prod`

### Environment Variables in Vercel

Add these in your Vercel project settings:
- `SUPABASE_URL`
- `SUPABASE_API_KEY`
- `RESEND_API_KEY`
- `FROM_EMAIL`
- `FROM_NAME`
- `ADMIN_EMAIL`

## Logging

The API logs important events:

- ✅ Guest found/verified
- ❌ Guest not found (with admin notification)
- 📧 Emails sent (confirmation and admin alerts)
- 💾 RSVPs saved to database
- ⚠️ Errors (email failures, database issues)

## Error Handling

- Email failures don't block RSVP submissions
- Database failures fall back gracefully
- Admin notifications are fire-and-forget (non-blocking)
- All errors are logged for debugging

## Support

For issues or questions, contact: markoparkermarsenic@gmail.com
