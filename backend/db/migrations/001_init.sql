-- Users table
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	phone TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	role TEXT NOT NULL DEFAULT 'user',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	creator_id INTEGER NOT NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	category TEXT NOT NULL DEFAULT '',
	region TEXT NOT NULL,
	location_detail TEXT,
	status TEXT NOT NULL DEFAULT 'open',
	goal_sats INTEGER,
	max_volunteers INTEGER NOT NULL,
	volunteer_mode TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	image_path TEXT,
	FOREIGN KEY (creator_id) REFERENCES users(id)
);

-- Volunteers table
CREATE TABLE IF NOT EXISTS volunteers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id),
	FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Donations table
CREATE TABLE IF NOT EXISTS donations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL,
	donor_id INTEGER NOT NULL,
	amount_sats INTEGER NOT NULL,
	payment_hash TEXT NOT NULL,
	payment_request TEXT NOT NULL,
	is_anonymous INTEGER NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'pending',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id),
	FOREIGN KEY (donor_id) REFERENCES users(id)
);

-- Keyholders table
CREATE TABLE IF NOT EXISTS keyholders (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL UNIQUE,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Payout requests table
CREATE TABLE IF NOT EXISTS payout_requests (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL,
	total_sats INTEGER NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id)
);

-- Payout signatures table
CREATE TABLE IF NOT EXISTS payout_signatures (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	payout_request_id INTEGER NOT NULL,
	keyholder_id INTEGER NOT NULL,
	action TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (payout_request_id) REFERENCES payout_requests(id),
	FOREIGN KEY (keyholder_id) REFERENCES keyholders(id),
	UNIQUE(payout_request_id, keyholder_id)
);

-- Seed test keyholders for development (user IDs 2, 3, 4, 5, 6)
INSERT OR IGNORE INTO keyholders (user_id) VALUES (2), (3), (4), (5), (6);