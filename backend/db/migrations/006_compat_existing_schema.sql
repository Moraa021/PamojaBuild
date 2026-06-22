-- Bring older local development databases up to the current application schema.
ALTER TABLE tasks ADD COLUMN category TEXT NOT NULL DEFAULT '';
ALTER TABLE donations ADD COLUMN payment_request TEXT NOT NULL DEFAULT '';
ALTER TABLE payout_requests ADD COLUMN total_sats INTEGER NOT NULL DEFAULT 0;
ALTER TABLE payout_signatures ADD COLUMN action TEXT NOT NULL DEFAULT 'sign';
ALTER TABLE payout_signatures ADD COLUMN created_at DATETIME;
