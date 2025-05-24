package chatgpt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

func (c *OpenAIClient) GetChatGPTResponseInterview(prompt string) (*ChatGPTResponse, error) {
	ctx := context.Background()

	var messagesArray []map[string]string
	messagesArray = append(messagesArray, map[string]string{
		"role":    "system",
		"content": prompt,
	})

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "gpt-4",
		"messages":    messagesArray,
		"max_tokens":  150,
		"temperature": 0.7,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("NewRequestWithContext failed: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, &OpenAIError{StatusCode: resp.StatusCode, Message: errResp.Error.Message}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Unmarshal result err: %v", err)
		return nil, err
	}

	choices := result["choices"].([]interface{})
	if len(choices) == 0 {
		err := errors.New("no question generated")
		return nil, err
	}

	chatGPTResponseRaw := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	var chatGPTResponse ChatGPTResponse
	if err := json.Unmarshal([]byte(chatGPTResponseRaw), &chatGPTResponse); err != nil {
		log.Printf("Unmarshal chatGPTResponse err: %v", err)
		return nil, err
	}

	return &chatGPTResponse, nil
}

func (c *OpenAIClient) GetChatGPTResponseConversation(conversationHistory []map[string]string) (*ChatGPTResponse, error) {
	ctx := context.Background()

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":       "gpt-4",
		"messages":    conversationHistory,
		"max_tokens":  150,
		"temperature": 0.7,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("NewRequestWithContext failed: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, &OpenAIError{StatusCode: resp.StatusCode, Message: errResp.Error.Message}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Unmarshal result err: %v", err)
		return nil, err
	}

	choices := result["choices"].([]interface{})
	if len(choices) == 0 {
		err := errors.New("no question generated")
		return nil, err
	}

	chatGPTResponseRaw := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	var chatGPTResponse ChatGPTResponse
	if err := json.Unmarshal([]byte(chatGPTResponseRaw), &chatGPTResponse); err != nil {
		log.Printf("Unmarshal chatGPTResponse err: %v", err)
		return nil, err
	}

	return &chatGPTResponse, nil
}
