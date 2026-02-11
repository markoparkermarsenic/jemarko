-- Add verified field to rsvps table
ALTER TABLE rsvps ADD COLUMN IF NOT EXISTS verified BOOLEAN NOT NULL DEFAULT false;

-- Create index on verified field for faster queries
CREATE INDEX IF NOT EXISTS idx_rsvps_verified ON rsvps(verified);
