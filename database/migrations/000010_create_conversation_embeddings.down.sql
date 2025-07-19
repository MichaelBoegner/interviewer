DROP INDEX IF EXISTS conversation_embeddings_lookup_idx;
DROP INDEX IF EXISTS convo_embeddings_by_question_idx;
DROP INDEX IF EXISTS conversation_embeddings_embedding_idx;

DROP TABLE IF EXISTS conversation_embeddings;

DROP EXTENSION IF EXISTS vector;
