package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"llm-balancer/openai"
	"net/http"

	"github.com/rs/zerolog/log"
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
	log.Debug().Msgf("URL: %s", url)
	// Set the model in the request body
	request.Request.Model = model
	// Set the request body to the modified request
	jsonBody, err := json.Marshal(request.Request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Request: %s", string(jsonBody))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// handle non-200 status codes, if rate limit related, put back onto queue
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	var response openai.ChatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	FullResponse := &Response{
		Response: &response,
		Error:    nil,
	}

	fmt.Printf("Response: %+v\n", FullResponse)

	return FullResponse, FullResponse.Error
}
