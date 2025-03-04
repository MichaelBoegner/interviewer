CREATE TABLE IF NOT EXISTS interviews (
    id SERIAL PRIMARY KEY,                 
    user_id INT REFERENCES users(id),         
    length INT NOT NULL,                   
    number_questions INT NOT NULL,         
    difficulty VARCHAR(50) NOT NULL,       
    status VARCHAR(50) NOT NULL,           
    score INT,                             
    language VARCHAR(50) NOT NULL,
    prompt TEXT,         
    first_question TEXT, 
    subtopic VARCHAR(255) NOT NULL,                       
    created_at TIMESTAMP, 
    updated_at TIMESTAMP  
);