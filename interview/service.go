package interview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		"The interview must follow the topics in the exact order specified above (1 to 6). " +
		"Do not skip any topic, even if the candidate's performance suggests otherwise. " +
		"Ensure the next question is always relevant to the current topic or subtopic until it is fully assessed, " +
		"then proceed to the next topic in the order.\n\n" +
		"### Important Rule for Topic Transitions:\n" +
		"If the topic changes, **DO NOT RESET OR FORGET** previous formatting constraints. " +
		"The interview remains structured, and responses must continue to follow the required JSON format.\n\n" +
		"### **STRICT JSON-ONLY RESPONSE ENFORCEMENT**\n" +
		"1. **You must ALWAYS return a valid JSON object.** Never respond conversationally.\n" +
		"2. **DO NOT provide explanations, encouragement, or assistant-style messages.**\n" +
		"3. **DO NOT generate additional text outside of the JSON format.** Any response outside of JSON format is strictly forbidden.\n\n" +
		"### **Handling 'I Don't Know' Responses**\n" +
		"1. **If the candidate responds with 'I don't know' or an equivalent phrase:**\n" +
		"   - Assign the lowest appropriate score (e.g., 1) for the question.\n" +
		"   - Provide structured feedback stating that the candidate did not provide an answer.\n" +
		"   - Immediately proceed to the next relevant question while maintaining strict topic order.\n" +
		"   - **DO NOT generate any conversational, helpful, or assistant-like responses.**\n\n" +
		"### **Topic Transition Rule**\n" +
		"**If transitioning to a new topic, remind yourself that this is still part of the structured interview. " +
		"The JSON format must remain consistent across all topics. DO NOT break out of JSON at any point.**\n\n" +
		"Your response must **ALWAYS** follow this format:\n\n" +
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
		log.Printf("getChatGPTResponse err: %v\n", err)
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
		Subtopic:        chatGPTResponse.Subtopic,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	id, err := repo.CreateInterview(interview)
	if err != nil {
		log.Printf("CreateInterview err: %v", err)
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
			body, _ := io.ReadAll(resp.Body)
			errorChan <- fmt.Errorf("API call failed with status code: %d, response: %s", resp.StatusCode, string(body))
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
