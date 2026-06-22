CREATE TYPE entry_type AS ENUM ('INBOUND_DONATION', 'SUBMARINE_SWAP', 'TAIL_PAYOUT');

CREATE TABLE donation_ledger (
    id BIGSERIAL PRIMARY KEY,
    task_id VARCHAR(64) REFERENCES tasks(id) ON DELETE RESTRICT NOT NULL,
    type entry_type NOT NULL,
    amount_sats BIGINT NOT NULL,
    reference_id VARCHAR(255) NOT NULL, -- LND Payment Hash or Bitcoin TxID
    previous_row_hash BYTEA NOT NULL,    -- SHA256 output (32 bytes) linking previous row
    row_hmac BYTEA NOT NULL,             -- Calculated engine verification token
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT chk_positive_amount CHECK (amount_sats > 0)
);

-- Performance indices for sequential validation and lookups
CREATE INDEX idx_ledger_task ON donation_ledger(task_id);
CREATE UNIQUE INDEX idx_ledger_reference ON donation_ledger(reference_id);
CREATE INDEX idx_ledger_task_id_desc ON donation_ledger(task_id, id DESC);