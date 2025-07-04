-- Fix Supabase Database Schema
-- Add unique constraint for ticker + snapshot_date combination

-- First, remove any duplicate records (if they exist)
DELETE FROM assets a 
USING assets b 
WHERE a.id > b.id 
AND a.ticker = b.ticker 
AND a.snapshot_date = b.snapshot_date;

-- Add the unique constraint
ALTER TABLE assets 
ADD CONSTRAINT unique_ticker_snapshot_date 
UNIQUE (ticker, snapshot_date);

-- Verify the constraint was added
SELECT conname, contype, conkey 
FROM pg_constraint 
WHERE conname = 'unique_ticker_snapshot_date'; 