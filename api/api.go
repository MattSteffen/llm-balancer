package api

import (
	"context"
	"llm-balancer/openai"
)

/*
TODO: Implement clients for openai (includes groq and openrouter), ollama, google
*/

// Represents an API-specific client capable of sending requests
// and potentially handling API-specific details like token usage in response.
// For MVP, focus primarily on BuildRequest and sending.
type Client interface {
	POSTChatCompletion(ctx context.Context, request *Request, model string) (*Response, error)
	// httpPOST a shared post method for all clients
}

type Request struct {
	Request      *openai.ChatCompletionRequest
	TokensNeeded int
}

type Response struct {
	Response *openai.ChatCompletionResponse
	Error    error
}

/*
Handle Errors with Headers:

Rate Limit Headers

In addition to viewing your limits on your account's limits page, you can also view rate limit information such as remaining requests and tokens in HTTP response headers as follows:
The following headers are set (values are illustrative):
HEADER	VALUE	NOTES
retry-after	2	In seconds
x-ratelimit-limit-requests	14400	Always refers to Requests Per Day (RPD)
x-ratelimit-limit-tokens	18000	Always refers to Tokens Per Minute (TPM)
x-ratelimit-remaining-requests	14370	Always refers to Requests Per Day (RPD)
x-ratelimit-remaining-tokens	17997	Always refers to Tokens Per Minute (TPM)
x-ratelimit-reset-requests	2m59.56s	Always refers to Requests Per Day (RPD)
x-ratelimit-reset-tokens	7.66s	Always refers to Tokens Per Minute (TPM)
Handling Rate Limits

When you exceed rate limits, our API returns a 429 Too Many Requests HTTP status code.
Note: retry-after is only set if you hit the rate limit and status code 429 is returned. The other headers are always included.
*/
