CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    billing_customer_id TEXT,
    billing_subscription_id TEXT, 
    subscription_tier TEXT DEFAULT 'free',
    subscription_start_date TIMESTAMP,
    credits INT,
    billing_status TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);