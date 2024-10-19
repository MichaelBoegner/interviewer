CREATE TABLE IF NOT EXISTS conversations (
    id SERIAL PRIMARY KEY,
    interview_id INT REFERENCES interviews(id),
    messages JSONB NOT NULL, 
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
