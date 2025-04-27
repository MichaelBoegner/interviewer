CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    refresh_token VARCHAR(255),
    expires_at TIMESTAMP,
     created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);