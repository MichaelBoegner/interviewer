CREATE TABLE IF NOT EXISTS interviews (
    id SERIAL PRIMARY KEY,  
    conversation_id int,                
    user_id INT REFERENCES users(id),         
    length INT NOT NULL,                   
    number_questions INT NOT NULL,         
    number_questions_answered INT NOT NULL DEFAULT 0,
    score_numerator INT NOT NULL DEFAULT 0,
    score INT,                             
    difficulty VARCHAR(50) NOT NULL,       
    status VARCHAR(50) NOT NULL,           
    language VARCHAR(50) NOT NULL,
    prompt TEXT,         
    jd_summary TEXT,
    first_question TEXT, 
    subtopic VARCHAR(255),                       
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);