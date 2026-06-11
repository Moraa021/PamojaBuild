-- Add columns for PSBT/Multisig implementation and volunteer payment requests
ALTER TABLE keyholders ADD COLUMN public_key TEXT;
ALTER TABLE payout_requests ADD COLUMN psbt TEXT;
ALTER TABLE payout_requests ADD COLUMN tx_id TEXT;
ALTER TABLE volunteers ADD COLUMN payment_request TEXT;
