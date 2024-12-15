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

	"github.com/michaelboegner/interviewer/models"
)

func StartInterview(repo InterviewRepo, userId, length, numberQuestions int, difficulty string) (*Interview, error) {
	now := time.Now()
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
		"2. Include the next question in the `next_question` field, even if transitioning to a new topic or subtopic.\n" +
		"3. If the subtopic or topic is complete but the interview is not finished, always populate `next_question` with the first question for the next topic or subtopic.\n" +
		"4. If the interview has ended, set `next_question` to \"finished\".\n" +
		"5. Always include a summary of the candidate's performance in the `feedback` field when the interview ends.\n" +
		"6. Always include all fields in the JSON response, even if some fields have empty or null values.\n\n" +
		"Flags (`move_to_new_subtopic` and `move_to_new_topic`) must follow these rules:\n" +
		"1. Set `move_to_new_topic` to `true` only for the **first question** in a new topic.\n" +
		"2. Set `move_to_new_subtopic` to `true` only for the **first question** in a new subtopic.\n" +
		"3. For all subsequent questions within the same topic or subtopic, set both flags to `false`.\n\n" +
		"Your response must always follow this format:\n\n" +
		"{\n" +
		"    \"topic\": \"System Design\",\n" +
		"    \"subtopic\": \"Scalability\",\n" +
		"    \"question\": \"How would you design a system to handle a high number of concurrent users?\",\n" +
		"    \"score\": null,\n" +
		"    \"feedback\": \"\",\n" +
		"    \"next_question\": \"What challenges might arise when scaling such a system?\",\n" +
		"    \"move_to_new_subtopic\": false,\n" +
		"    \"move_to_new_topic\": false\n" +
		"}"

	chatGPTResponse, err := getChatGPTResponse(prompt)
	if err != nil {
		return nil, err
	}
	chatGPTResponse.CreatedAt = now

	interview := &Interview{
		UserId:          userId,
		Length:          length,
		NumberQuestions: numberQuestions,
		Difficulty:      difficulty,
		Status:          "Running",
		Score:           100,
		Language:        "Python",
		Prompt:          prompt,
		ChatGPTResponse: chatGPTResponse,
		FirstQuestion:   chatGPTResponse.Question,
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

	fmt.Printf("interview in GetInterview: %v\n", interview)

	return interview, nil
}

func getChatGPTResponse(prompt string) (*models.ChatGPTResponse, error) {
	ctx := context.Background()

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "gpt-4",
		"messages":    []map[string]string{{"role": "system", "content": prompt}},
		"max_tokens":  150,
		"temperature": 0.2,
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

	responseChan := make(chan *models.ChatGPTResponse)
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

		chatGPTResponseResponse := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

		var chatGPTResponse models.ChatGPTResponse
		if err := json.Unmarshal([]byte(chatGPTResponseResponse), &chatGPTResponse); err != nil {
			errorChan <- fmt.Errorf("failed to parse question context: %v", err)
			return
		}

		responseChan <- &chatGPTResponse
	}()

	select {
	case chatGPTResponse := <-responseChan:
		return chatGPTResponse, nil
	case err := <-errorChan:
		return nil, err
	}
}
