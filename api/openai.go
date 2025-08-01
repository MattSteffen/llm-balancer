package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"llm-balancer/openai"
	"net/http"

	"github.com/rs/zerolog/log"
)

var (
	canStream = false // Set to false if streaming is not supported
)

type OpenAIClient struct {
	BaseURL string
	APIKey  string
}

// NewOpenAIClient creates a new Google API client.
func NewOpenAIClient(baseURL string, apiKey string) *OpenAIClient {
	return &OpenAIClient{BaseURL: baseURL, APIKey: apiKey}
}

// POSTChatCompletion sends a chat completion request to the OpenAI API.
func (c *OpenAIClient) POSTChatCompletion(ctx context.Context, request *Request, model string) (*Response, error) {
	url := fmt.Sprintf("%s/chat/completions", c.BaseURL)
	log.Info().Str("provider", "openai").Str("model", model).Msg("POSTChatCompletion")
	// Set the model in the request body
	request.Request.Model = model
	request.Request.Stream = &canStream

	// Set the request body to the modified request
	jsonBody, err := json.Marshal(request.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	fmt.Println("--------------------------------")
	fmt.Println("Response body:", string(bodyBytes))
	fmt.Println("--------------------------------")

	// handle non-200 status codes, if rate limit related, put back onto queue
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	var response openai.ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	FullResponse := &Response{
		Response: &response,
		Error:    nil,
	}

	// fmt.Printf("Response: %+v\n", FullResponse)

	return FullResponse, FullResponse.Error
}
