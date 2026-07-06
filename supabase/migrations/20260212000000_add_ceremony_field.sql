-- Add ceremony field to guests table
ALTER TABLE guests ADD COLUMN IF NOT EXISTS ceremony BOOLEAN NOT NULL DEFAULT false;

-- Create index on ceremony field for faster filtering in the admin dashboard
CREATE INDEX IF NOT EXISTS idx_guests_ceremony ON guests(ceremony);
