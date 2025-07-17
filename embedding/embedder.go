package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type HTTPEmbedder struct {
	Endpoint string
	Timeout  time.Duration
}

func NewHTTPEmbedder(endpoint string) *HTTPEmbedder {
	return &HTTPEmbedder{
		Endpoint: endpoint,
		Timeout:  3 * time.Second,
	}
}

func (e *HTTPEmbedder) EmbedText(ctx context.Context, input string) ([]float32, error) {
	body, err := json.Marshal(map[string]string{"text": input})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: e.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Embedding, nil
}
