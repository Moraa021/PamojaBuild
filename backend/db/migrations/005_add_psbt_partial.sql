-- Add psbt_partial to payout_signatures table to store the hex-encoded partially signed PSBT
ALTER TABLE payout_signatures ADD COLUMN psbt_partial TEXT;
