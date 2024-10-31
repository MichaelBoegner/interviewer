package interview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func StartInterview(repo InterviewRepo, userId, length, numberQuestions int, difficulty string) (*Interview, error) {
	firstQuestion, err := getFirstQuestion()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	interview := &Interview{
		UserId:          userId,
		Length:          length,
		NumberQuestions: numberQuestions,
		Difficulty:      difficulty,
		Status:          "Running",
		Score:           100,
		Language:        "Python",
		FirstQuestion:   firstQuestion,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	id, err := repo.CreateInterview(interview)
	if err != nil {
		return nil, err
	}

	interview.Id = id

	return interview, nil
}

func GetInterview(repo InterviewRepo, interviewID int) (*Interview, error) {
	interview, err := repo.GetInterview(interviewID)
	if err != nil {
		return nil, err
	}

	return interview, nil
}

func getFirstQuestion() (string, error) {
	ctx := context.Background()
	prompt := "You are conducting a technical interview for a backend development position. The candidate is at a junior to mid-level skill level. Start by asking a technical interview question that assesses the candidate's understanding of core backend development concepts such as RESTful APIs, databases, server architecture, or programming best practices. Provide the first question in a clear and concise manner."

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "gpt-4",
		"messages":    []map[string]string{{"role": "system", "content": prompt}},
		"max_tokens":  150,
		"temperature": 0.7,
	})
	if err != nil {
		return "", err
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	responseChan := make(chan string)
	errorChan := make(chan error)
	go func() {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errorChan <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorChan <- fmt.Errorf("API call failed with status code: %d", resp.StatusCode)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errorChan <- err
			return
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			errorChan <- err
			return
		}

		choices := result["choices"].([]interface{})
		if len(choices) == 0 {
			errorChan <- fmt.Errorf("no question generated")
			return
		}

		firstQuestion := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
		responseChan <- firstQuestion
	}()

	select {
	case firstQuestion := <-responseChan:
		response := firstQuestion
		return response, nil

	case err := <-errorChan:
		return "", err
	}
}
