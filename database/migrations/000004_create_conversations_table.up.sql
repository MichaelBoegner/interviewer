CREATE TABLE IF NOT EXISTS conversations (
    id SERIAL PRIMARY KEY,
    interview_id INT NOT NULL,
    current_topic INT NOT NULL,
    current_subtopic VARCHAR(255) NOT NULL,
    current_question_number INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);