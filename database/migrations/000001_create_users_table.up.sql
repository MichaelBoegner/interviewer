CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    stripe_customer_id TEXT,
    subscription_tier TEXT DEFAULT 'free',
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);