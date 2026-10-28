-- Legacy accounts remain unverified; do not claim ownership on their behalf.
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMPTZ;
CREATE TABLE pending_registrations (
 email VARCHAR(320) PRIMARY KEY,
 token_hash VARCHAR(64) NOT NULL UNIQUE,
 password_hash VARCHAR(255) NOT NULL,
 first_name VARCHAR(100) NOT NULL,
 last_name VARCHAR(100) NOT NULL,
 phone VARCHAR(30) NOT NULL DEFAULT '',
 document_number VARCHAR(40) NOT NULL DEFAULT '',
 nationality VARCHAR(80) NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL,
 expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX pending_registrations_expiry_idx ON pending_registrations(expires_at);
