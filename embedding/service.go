package embedding

import (
	"context"
	"fmt"
	"log"

	"github.com/pgvector/pgvector-go"
)

func (s *Service) ProcessAndRetrieve(ctx context.Context, input EmbedInput) ([]string, error) {
	fmt.Printf("ProcessAndRetrieve firing\n")

	summaryResp, err := s.Summarizer.ExtractResponseSummary(input.Question, input.UserResponse)
	if err != nil {
		log.Printf("s.Summarizer.ExtractResponseSummary failed: %v", err)
		return nil, err
	}

	fmt.Printf("SummaryResp: %v\n", summaryResp)

	allRelevant := []string{}
	seen := map[string]struct{}{}
	limit := 1

	for _, point := range summaryResp.UserRespSummary {
		rawVec, err := s.Embedder.EmbedText(ctx, point)
		if err != nil {
			log.Printf("s.Embedder.EmbedText failed: %v", err)
			return nil, err
		}
		vector := pgvector.NewVector(rawVec)

		fmt.Printf("vector: %v\n", vector)

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

		relevantEmbeddings, err := s.Repo.GetSimilarEmbeddings(
			ctx,
			input.InterviewID,
			input.TopicID,
			input.QuestionNumber,
			input.MessageID,
			vector,
			limit,
		)
		if err != nil {
			log.Printf("s.Repo.GetSimilarEmbeddings failed: %v", err)
			return nil, err
		}

		for _, r := range relevantEmbeddings {
			if _, exists := seen[r]; !exists {
				seen[r] = struct{}{}
				allRelevant = append(allRelevant, r)
			}
		}

		fmt.Printf("relevant: %v\n", allRelevant)
	}

	return allRelevant, nil
}
