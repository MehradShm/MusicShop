-- Enable case-insensitive text type
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     username CITEXT NOT NULL UNIQUE,  -- case-insensitive
                                     name TEXT NOT NULL,
                                     email TEXT NOT NULL UNIQUE,
                                     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
-- No extra idx on username needed: the UNIQUE constraint already creates an index.