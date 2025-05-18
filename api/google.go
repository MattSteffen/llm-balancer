package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GoogleClient struct {
	BaseURL string
	APIKey  string
}

// NewGoogleClient creates a new Google API client.
func NewGoogleClient(baseURL string, apiKey string) *GoogleClient {
	return &GoogleClient{BaseURL: baseURL, APIKey: apiKey}
}

// POSTChatCompletion sends a chat completion request to the Google API.
func (c *GoogleClient) POSTChatCompletion(request *Request, model string) (*Response, error) {
	url := fmt.Sprintf("%s/chat/completions", c.BaseURL)
	rf, err := json.Marshal(request.Request.ResponseFormat)
	fmt.Printf("response format: %+v\n", rf)
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	// log.Debug().Msgf("Response: %s", string(bodyBytes))
	var response Response
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}
	if response.Error != nil {
		return nil, fmt.Errorf("API error: %s", response.Error.Error())
	}
	return &response, nil
}

// potentially do: toOpenAIResponse and fromOpenAIResponse as functions that take an openai.ChatCompletionRequest
// and return a google compatible request and similary for the response
