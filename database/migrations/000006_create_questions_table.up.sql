CREATE TABLE IF NOT EXISTS questions (
    id SERIAL PRIMARY KEY,
    topic_id INT REFERENCES topics(id),
    question_number INT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
