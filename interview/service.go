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
	questionContext, err := getQuestionContext()
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
		QuestionContext: questionContext,
		FirstQuestion:   questionContext.Question,
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

func getQuestionContext() (*QuestionContext, error) {
	ctx := context.Background()
	prompt := "You are conducting a technical interview for a backend development position. " +
		"The interview is divided into six main topics:\n\n" +
		"1. **Introduction**\n" +
		"2. **Coding**\n" +
		"3. **System Design**\n" +
		"4. **Databases and Data Management**\n" +
		"5. **Behavioral**\n" +
		"6. **General Backend Knowledge**\n\n" +
		"Each main topic may contain multiple subtopics. For each subtopic:\n" +
		"- Ask as many subtopic-specific questions as needed until the candidate has " +
		"sufficiently or insufficiently proven their understanding of the subtopic.\n" +
		"- Provide a score (1-10) for their performance on each subtopic question, with feedback " +
		"explaining the score.\n" +
		"- When the subtopic is sufficiently assessed, decide if the current topic needs " +
		"further exploration or if it's time to move to the next topic.\n\n" +
		"After each question:\n" +
		"1. Wait for the candidate's response before evaluating their answer.\n" +
		"2. If the candidate has not yet responded, return blank values for \"score\", \"feedback\", and \"next_question\".\n" +
		"3. Evaluate the candidate's response and decide if additional questions about the " +
		"current subtopic are necessary.\n" +
		"4. If transitioning to a new subtopic or topic, include a flag (move_to_new_subtopic) " +
		"and/or (move_to_new_topic) in the response. Otherwise, set these flags to false.\n" +
		"5. Return your response in the following JSON format:\n\n" +
		"{\n" +
		"    \"topic\": \"System Design\",\n" +
		"    \"subtopic\": \"Scalability\",\n" +
		"    \"question\": \"How would you design a system to handle a high number of concurrent users?\",\n" +
		"    \"score\": null,\n" +
		"    \"feedback\": \"\",\n" +
		"    \"next_question\": \"\",\n" +
		"    \"move_to_new_subtopic\": false,\n" +
		"    \"move_to_new_topic\": false\n" +
		"}"

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "gpt-4",
		"messages":    []map[string]string{{"role": "system", "content": prompt}},
		"max_tokens":  150,
		"temperature": 0.7,
	})
	if err != nil {
		return nil, err
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	responseChan := make(chan *QuestionContext)
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

		questionContextResponse := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

		var questionContext QuestionContext
		if err := json.Unmarshal([]byte(questionContextResponse), &questionContext); err != nil {
			errorChan <- fmt.Errorf("failed to parse question context: %v", err)
			return
		}

		responseChan <- &questionContext
	}()

	select {
	case questionContext := <-responseChan:
		return questionContext, nil
	case err := <-errorChan:
		return nil, err
	}
}
