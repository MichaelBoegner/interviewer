package embedding

import (
	"context"
	"log"
)

func (s *Service) ProcessAndRetrieve(ctx context.Context, input EmbedInput, limit int) ([]string, error) {
	summaryResp, err := s.Summarizer.ExtractResponseSummary(input.Question, input.UserResponse)
	if err != nil {
		log.Printf("s.Summarizer.ExtractResponseSummary failed: %v", err)
		return nil, err
	}

	for _, point := range summaryResp.UserRespSummary {
		vector, err := s.Embedder.EmbedText(ctx, point)
		if err != nil {
			log.Printf("s.Embedder.EmbedText failed: %v", err)
			return nil, err
		}

		err = s.Repo.StoreEmbedding(ctx, Embedding{
			InterviewID:    input.InterviewID,
			ConversationID: input.ConversationID,
			MessageID:      input.MessageID,
			TopicID:        input.TopicID,
			QuestionNumber: input.QuestionNumber,
			Summary:        point,
			Vector:         vector,
			CreatedAt:      input.CreatedAt,
		})
		if err != nil {
			log.Printf("s.Repo.StoreEmbedding failed: %v", err)
			return nil, err
		}
	}

	responseVector, err := s.Embedder.EmbedText(ctx, input.UserResponse)
	if err != nil {
		log.Printf("s.Embedder.EmbedText failed: %v", err)
		return nil, err
	}

	relevant, err := s.Repo.GetSimilarEmbeddings(
		ctx,
		input.InterviewID,
		input.TopicID,
		input.QuestionNumber,
		input.MessageID,
		responseVector,
		limit,
	)
	if err != nil {
		log.Printf("s.Repo.GetSimilarEmbeddings failed: %v", err)
		return nil, err
	}

	return relevant, nil
}
