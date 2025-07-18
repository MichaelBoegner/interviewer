package embedding

import (
	"context"
	"fmt"
	"log"
	"strings"
)

func (s *Service) ProcessAndRetrieve(ctx context.Context, input EmbedInput, limit int) ([]string, error) {
	// DEBUG
	fmt.Printf("ProcessAndRetrive firing\n")

	summaryResp, err := s.Summarizer.ExtractResponseSummary(input.Question, input.UserResponse)
	if err != nil {
		log.Printf("s.Summarizer.ExtractResponseSummary failed: %v", err)
		return nil, err
	}

	// DEBUG
	fmt.Printf("SummaryResp: %v\n", summaryResp)

	for _, point := range summaryResp.UserRespSummary {
		vector, err := s.Embedder.EmbedText(ctx, point)
		if err != nil {
			log.Printf("s.Embedder.EmbedText failed: %v", err)
			return nil, err
		}
		vectorString := formatVector(vector)

		// DEBUG
		fmt.Printf("vector: %v\n", vector)

		err = s.Repo.StoreEmbedding(ctx, Embedding{
			InterviewID:    input.InterviewID,
			ConversationID: input.ConversationID,
			MessageID:      input.MessageID,
			TopicID:        input.TopicID,
			QuestionNumber: input.QuestionNumber,
			Summary:        point,
			Vector:         vectorString,
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

	// DEBUG
	fmt.Printf("responseVector: %v\n", responseVector)

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

	// DEBUG
	fmt.Printf("relevant: %v\n", relevant)

	return relevant, nil
}

func formatVector(vec []float32) string {
	strs := make([]string, len(vec))
	for i, v := range vec {
		strs[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(strs, ", ") + "]"
}
