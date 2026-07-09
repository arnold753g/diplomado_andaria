ALTER TABLE users DROP CONSTRAINT users_role_check;
UPDATE users SET role = 'turista' WHERE role = 'user';
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'turista';
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'turista', 'encargado_agencia', 'encargado_atraccion'));
ALTER TABLE users ADD COLUMN phone VARCHAR(30) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN document_number VARCHAR(40) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN nationality VARCHAR(80) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN google_subject VARCHAR(255) UNIQUE;
-- Google-only accounts have no local password. Password login cannot match an empty hash.
ALTER TABLE users ALTER COLUMN password_hash SET DEFAULT '';

CREATE TABLE oauth_attempts (
    state_hash VARCHAR(64) PRIMARY KEY,
    verifier VARCHAR(128) NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(64) REFERENCES auth_sessions(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX oauth_attempts_expiry ON oauth_attempts(expires_at);
