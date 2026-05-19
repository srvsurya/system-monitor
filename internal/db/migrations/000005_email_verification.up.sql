ALTER TABLE users ADD COLUMN verified BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE verification_tokens (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);