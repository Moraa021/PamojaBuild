CREATE TYPE task_state AS ENUM ('ACTIVE', 'LIQUIDATING', 'READY_FOR_PAYOUT', 'SYSTEM_LOCKDOWN', 'ARCHIVED');

CREATE TABLE tasks (
    id VARCHAR(64) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    state task_state NOT NULL DEFAULT 'ACTIVE',
    l2_balance_sats BIGINT NOT NULL DEFAULT 0,
    l1_balance_sats BIGINT NOT NULL DEFAULT 0,
    current_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Production safety constraint: preventing accidental negative ledger state
    CONSTRAINT chk_positive_l2_balance CHECK (l2_balance_sats >= 0),
    CONSTRAINT chk_positive_l1_balance CHECK (l1_balance_sats >= 0),
    CONSTRAINT chk_positive_index CHECK (current_index >= 0)
);

CREATE TABLE task_keys (
    task_id VARCHAR(64) REFERENCES tasks(id) ON DELETE RESTRICT,
    trustee_index INT NOT NULL,
    xpub TEXT NOT NULL,
    web_crypto_pubkey_hex VARCHAR(64) NOT NULL,
    PRIMARY KEY (task_id, trustee_index),
    
    CONSTRAINT chk_trustee_index_bounds CHECK (trustee_index >= 0 AND trustee_index < 5)
);

-- Indices for performance optimization
CREATE INDEX idx_tasks_state ON tasks(state);