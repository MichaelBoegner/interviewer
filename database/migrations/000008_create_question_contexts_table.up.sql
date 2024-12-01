CREATE TABLE IF NOT EXISTS question_contexts (
    id SERIAL PRIMARY KEY,                
    topic TEXT NOT NULL,                  
    subtopic TEXT NOT NULL,               
    question TEXT NOT NULL,               
    score INT NOT NULL,                   
    feedback TEXT NOT NULL,               
    next_question TEXT NOT NULL,          
    move_to_new_subtopic BOOLEAN NOT NULL,
    move_to_new_topic BOOLEAN NOT NULL,   
    created_at TIMESTAMP NOT NULL         
);
