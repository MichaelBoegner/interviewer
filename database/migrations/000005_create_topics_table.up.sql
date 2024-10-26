CREATE TABLE IF NOT EXISTS topics (
    id SERIAL PRIMARY KEY,
    conversation_id INT REFERENCES conversations(id),
    name VARCHAR(255) NOT NULL
);
