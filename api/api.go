package api

import "net/http"

/*
TODO: Implement clients for openai, ollama, google, groq
*/

// Represents an API-specific client capable of sending requests
// and potentially handling API-specific details like token usage in response.
// For MVP, focus primarily on BuildRequest and sending.
type Client interface {
	// BuildRequest creates a net/http.Request for the specific API.
	// It takes the original request body (potentially modified) and the target model.
	BuildRequest(originalBody []byte, model string, apiKey string) (*http.Request, error)

	// SendRequest executes the prepared request.
	// It could potentially handle API-specific error parsing (e.g., rate limit errors).
	SendRequest(req *http.Request) (*http.Response, error)

	ProcessRequest(req *http.Request) (*http.Response, error)

	// ParseResponseForTokens (Post-MVP) parses the response body to find token usage.
	// ParseResponseForTokens(respBody []byte) (inputTokens, outputTokens int, err error)
}

type Request struct {
	Prompt       string
	TokensNeeded int
	// Add other fields as needed (e.g., user ID, metadata)
}

type Response struct {
	Output string
	Error  error
	// Add other fields as needed (e.g., usage stats)
}
