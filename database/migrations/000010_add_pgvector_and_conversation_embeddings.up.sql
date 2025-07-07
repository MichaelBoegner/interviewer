CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE conversation_embeddings (
  id SERIAL PRIMARY KEY,
  interview_id     INT NOT NULL,
  conversation_id  INT NOT NULL,
  topic_id         INT NOT NULL,
  question_number  INT NOT NULL,
  message_id       INT NOT NULL,
  summary          TEXT NOT NULL,
  embedding        VECTOR(1536) NOT NULL,
  created_at       TIMESTAMP DEFAULT now()
);

CREATE INDEX conversation_embeddings_embedding_idx
  ON conversation_embeddings USING ivfflat (embedding vector_cosine_ops)
  WITH (lists = 100); -- adjust based on size

CREATE INDEX conversation_embeddings_lookup_idx
  ON conversation_embeddings (interview_id, message_id);
