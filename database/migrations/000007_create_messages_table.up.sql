CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    question_id INT REFERENCES questions(id),
    author VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
