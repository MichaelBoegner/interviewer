CREATE TABLE IF NOT EXISTS interviews (
    id SERIAL PRIMARY KEY,                 
    user_id INT REFERENCES users(id),         
    length INT NOT NULL,                   
    number_questions INT NOT NULL,         
    difficulty VARCHAR(50) NOT NULL,       
    status VARCHAR(50) NOT NULL,           
    score INT,                             
    language VARCHAR(50) NOT NULL,         
    first_question VARCHAR(255),                       
    created_at TIMESTAMP, 
    updated_at TIMESTAMP  
);