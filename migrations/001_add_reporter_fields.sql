-- Migration: Add reporter_id and reporter_name to reports table
-- Date: 2025-11-20
-- Purpose: Track which user submitted each report

-- Add reporter_id column (nullable initially for existing records)
ALTER TABLE reports ADD COLUMN IF NOT EXISTS reporter_id INTEGER;

-- Add reporter_name column (nullable initially for existing records)
ALTER TABLE reports ADD COLUMN IF NOT EXISTS reporter_name VARCHAR(100);

-- Optional: Add index on reporter_id for faster queries
CREATE INDEX IF NOT EXISTS idx_reports_reporter_id ON reports(reporter_id);

-- Optional: Add foreign key constraint (uncomment if you want referential integrity)
-- ALTER TABLE reports ADD CONSTRAINT fk_reports_reporter FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE SET NULL;

-- For existing records without a reporter, you could set a default or leave NULL
-- UPDATE reports SET reporter_name = 'Unknown' WHERE reporter_name IS NULL;
