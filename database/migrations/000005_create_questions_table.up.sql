CREATE TABLE IF NOT EXISTS questions (
    id SERIAL PRIMARY KEY,
    conversation_id INT REFERENCES conversations(id),
    topic_id INT NOT NULL,
    question_number INT NOT NULL,
    prompt TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
