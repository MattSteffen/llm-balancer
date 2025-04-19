package api

import (
	"bytes"
	"fmt"
	"net/http"
)

type GoogleClient struct {
	BaseURL string
}

// NewGoogleClient creates a new Google API client.
func NewGoogleClient(baseURL string) *GoogleClient {
	return &GoogleClient{BaseURL: baseURL}
}

// BuildRequest prepares a request specifically for the Google API.
// Needs to handle paths like /models/{model}:generateContent or /models/{model}:streamGenerateContent
func (c *GoogleClient) BuildRequest(originalBody []byte, model string, apiKey string) (*http.Request, error) {
	// Assume originalBody is the standard chat completion format
	// Need to convert to Google's format if necessary, or expect client to send Google format?
	// For MVP, let's assume the client sends a format that maps reasonably well,
	// but we need to set the correct URL path and add the API key.
	url := fmt.Sprintf("%s/models/%s:generateContent", c.BaseURL, model) // Example path for chat completions
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(originalBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create google request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Google uses ?key=YOUR_API_KEY in the URL or x-goog-api-key header
	// Let's use header for consistency with Bearer, although their docs often show URL param
	req.Header.Set("X-Goog-Api-Key", apiKey)

	return req, nil
}

// SendRequest executes the request.
func (c *GoogleClient) SendRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	// TODO: Add context with timeout?
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send google request: %w", err)
	}
	// TODO: Check response status for API-specific errors, including rate limits?
	return resp, nil
}

func (c *GoogleClient) ProcessRequest(req *http.Request) (*http.Response, error) {
	return nil, nil
}
