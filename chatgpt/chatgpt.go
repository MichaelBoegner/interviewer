package chatgpt

import (
	"fmt"
	"log"
	"net/http"
)

func GetFirstQuestion(username string) (string, error) {
	client := &http.Client{}
	resp, err := client.Get("https://example.com")
	if err != nil {
		log.Printf("Error getting first question from ChatGPT: %v", err)
		return "", err
	}

	return fmt.Sprintf("%v, this is the first questino: %v", username, resp), nil
}
