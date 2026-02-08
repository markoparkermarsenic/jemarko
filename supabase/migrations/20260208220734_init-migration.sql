-- Create guests table
CREATE TABLE IF NOT EXISTS guests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  address TEXT,
  dietary TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes on guests table
CREATE INDEX IF NOT EXISTS idx_guests_name ON guests(name);
CREATE INDEX IF NOT EXISTS idx_guests_address ON guests(address);

-- Create rsvps table
CREATE TABLE IF NOT EXISTS rsvps (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  is_attending BOOLEAN NOT NULL DEFAULT false,
  attending_guests TEXT[] DEFAULT '{}',
  diet TEXT,
  submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  avatar_data JSONB DEFAULT '[]'::jsonb
);

-- Create indexes on rsvps table
CREATE INDEX IF NOT EXISTS idx_rsvps_email ON rsvps(email);
CREATE INDEX IF NOT EXISTS idx_rsvps_submitted_at ON rsvps(submitted_at DESC);
CREATE INDEX IF NOT EXISTS idx_rsvps_avatar_data ON rsvps USING gin(avatar_data);

-- Enable Row Level Security
ALTER TABLE guests ENABLE ROW LEVEL SECURITY;
ALTER TABLE rsvps ENABLE ROW LEVEL SECURITY;

-- Create policies for guests table (allow all operations)
CREATE POLICY "Allow all operations on guests" ON guests
  FOR ALL
  USING (true)
  WITH CHECK (true);

-- Create policies for rsvps table (allow all operations)
CREATE POLICY "Allow all operations on rsvps" ON rsvps
  FOR ALL
  USING (true)
  WITH CHECK (true);
