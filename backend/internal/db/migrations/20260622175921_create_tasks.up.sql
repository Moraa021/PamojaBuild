-- STEP 1: Create dropdown lists (Enums) so the database only accepts valid states.
CREATE TYPE task_status AS ENUM ('open', 'in_progress', 'pending_verification', 'completed');
CREATE TYPE volunteer_mode AS ENUM ('open', 'approval_required');
CREATE TYPE task_financial_state AS ENUM ('ACTIVE', 'LIQUIDATING', 'READY_FOR_PAYOUT', 'SYSTEM_LOCKDOWN', 'ARCHIVED');

-- STEP 2: Create the main unified Tasks table.
CREATE TABLE tasks (
    -- --- ORIGINAL VOLUNTEER FIELDS ---
    id BIGSERIAL PRIMARY KEY,                          -- Automatic counting number (1, 2, 3...)
    creator_id BIGINT NOT NULL,                        -- ID of the person who created it
    title VARCHAR(255) NOT NULL,                       -- Title of the project
    description TEXT NOT NULL,                         -- Deep description of the project
    -- TODO: Review how to store category, region and location_detail
    category VARCHAR(100) NOT NULL,                    -- Example: "Water", "Education"
    region VARCHAR(100) NOT NULL,                      -- Example: "Kibera"
    location_detail TEXT,                              -- Specific coordinates or landmarks
    status task_status NOT NULL DEFAULT 'open',        -- What the volunteers are doing on the ground
    goal_sats BIGINT,                                  -- The fundraising goal in satoshis
    max_volunteers BIGINT NOT NULL DEFAULT 0,          -- Max headcount for volunteers
    volunteer_mode volunteer_mode NOT NULL DEFAULT 'open', -- Signup rules
    image_path TEXT,                                   -- Web path to the campaign cover photo
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Exact date & time the campaign started

    -- --- THE NEW CRYPTOGRAPHIC VAULT FIELDS ---
    slug VARCHAR(64) UNIQUE NOT NULL,                  -- Clean URL/ID string (e.g., "kibera-clean-water") [cite: 59, 66]
    financial_state task_financial_state NOT NULL DEFAULT 'ACTIVE', -- Controls what the money can do [cite: 37, 38]
    l2_balance_sats BIGINT NOT NULL DEFAULT 0,         -- Loose change on the active node [cite: 6]
    l1_balance_sats BIGINT NOT NULL DEFAULT 0,         -- Safe money in the offline vault [cite: 6]
    current_index INT NOT NULL DEFAULT 0,              -- Used to generate new unique Bitcoin addresses

    -- Safety Guards: The database will reject updates if money or step counters drop below zero
    CONSTRAINT chk_positive_l2 CHECK (l2_balance_sats >= 0),
    CONSTRAINT chk_positive_l1 CHECK (l1_balance_sats >= 0),
    CONSTRAINT chk_positive_idx CHECK (current_index >= 0)
);

-- STEP 3: Create the table that holds the 5 community leaders' public keys.
-- Each campaign ("slug") gets exactly 5 keys attached to it[cite: 7].
CREATE TABLE task_keys (
    task_slug VARCHAR(64) REFERENCES tasks(slug) ON DELETE RESTRICT,
    trustee_index INT NOT NULL,                  -- Number 0 through 4 (for the 5 leaders) [cite: 21]
    xpub TEXT NOT NULL,                          -- Online wallet master public key [cite: 21]
    web_crypto_pubkey_hex VARCHAR(64) NOT NULL,  -- Phone/Browser backup public key [cite: 21]
    PRIMARY KEY (task_slug, trustee_index),      -- Prevents duplicate slot entries

    -- Safety Guard: Ensure leader index slots are only 0, 1, 2, 3, or 4
    CONSTRAINT chk_trustee_index_bounds CHECK (trustee_index >= 0 AND trustee_index < 5)
);

-- Indexing: Speeds up database lookups when scanning by the slug text
CREATE INDEX idx_tasks_slug ON tasks(slug);