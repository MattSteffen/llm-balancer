package openai

import (
	"encoding/json"
	"fmt"
)

type ChatCompletionRequest struct {
	Messages            []Message         `json:"messages"`
	Model               string            `json:"model"`
	Audio               *AudioOptions     `json:"audio,omitempty"`
	FrequencyPenalty    *float64          `json:"frequency_penalty,omitempty"`
	LogitBias           map[string]int    `json:"logit_bias,omitempty"`
	LogProbs            *bool             `json:"logprobs,omitempty"`
	MaxCompletionTokens *int              `json:"max_completion_tokens,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	Modalities          []string          `json:"modalities,omitempty"`
	N                   *int              `json:"n,omitempty"`
	ParallelToolCalls   *bool             `json:"parallel_tool_calls,omitempty"`
	Prediction          *Prediction       `json:"prediction,omitempty"`
	PresencePenalty     *float64          `json:"presence_penalty,omitempty"`
	ReasoningEffort     *string           `json:"reasoning_effort,omitempty"`
	ResponseFormat      *ResponseFormat   `json:"response_format,omitempty"`
	Seed                *int              `json:"seed,omitempty"`
	ServiceTier         *string           `json:"service_tier,omitempty"`
	Stop                any               `json:"stop,omitempty"` // can be string or []string
	Store               *bool             `json:"store,omitempty"`
	Stream              *bool             `json:"stream,omitempty"`
	StreamOptions       *StreamOptions    `json:"stream_options,omitempty"`
	Temperature         *float64          `json:"temperature,omitempty"`
	ToolChoice          any               `json:"tool_choice,omitempty"` // string or ToolChoice
	Tools               []Tool            `json:"tools,omitempty"`
	TopLogprobs         *int              `json:"top_logprobs,omitempty"`
	TopP                *float64          `json:"top_p,omitempty"`
	User                string            `json:"user,omitempty"`
	WebSearchOptions    *WebSearchOptions `json:"web_search_options,omitempty"`
}

// TODO: Replace the message with an interface and message types of developer, assistant, system, tool, and user
// TODO: This will also imply I need to create a custom marshaller for the message

type Message struct {
	Role       string `json:"role"`
	Content    any    `json:"content"` // string or []ContentPart
	Name       string `json:"name,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

type ContentPart struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Refusal string `json:"refusal,omitempty"`
}

type AudioOptions struct {
	Format string `json:"format"`
	Voice  string `json:"voice"`
}

type Prediction struct {
	Type    string `json:"type"`
	Content any    `json:"content"` // string or []ContentPart
}

type ResponseFormat struct {
	Type       string      `json:"type"`
	JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

type JSONSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"` // The actual json schema
	Strict      *bool          `json:"strict,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Strict      *bool          `json:"strict,omitempty"`
}

// type ToolChoice struct {
// 	Type     string             `json:"type"`
// 	Function ChoiceFunctionCall `json:"function"`
// }

// type ChoiceFunctionCall struct {
// 	Name string `json:"name"`
// }

type WebSearchOptions struct {
	// Add web search specific fields here when they become available
}

// ChatCompletionResponse represents the chat completion response
type ChatCompletionResponse struct {
	Choices           []Choice `json:"choices"`                // List of chat completion choices
	Created           int      `json:"created"`                // Unix timestamp of creation
	ID                string   `json:"id"`                     // Unique identifier for completion
	Model             string   `json:"model"`                  // Model used for completion
	Object            string   `json:"object"`                 // Always "chat.completion"
	ServiceTier       *string  `json:"service_tier,omitempty"` // Service tier used for processing
	SystemFingerprint string   `json:"system_fingerprint"`     // Backend configuration fingerprint
	Usage             Usage    `json:"usage"`                  // Usage statistics
}

// UnmarshalJSON implements custom unmarshaling for ChatCompletionResponse
func (c *ChatCompletionResponse) UnmarshalJSON(data []byte) error {
	// Create auxiliary type to avoid recursion
	type Aux ChatCompletionResponse
	var aux Aux

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("error unmarshaling response: %v", err)
	}

	// potentially unmarshal the tool choice function call arguments into map[string]any

	// Copy data from aux to c
	*c = ChatCompletionResponse(aux)
	return nil
}

type Choice struct {
	FinishReason string            `json:"finish_reason"` // Reason for stopping generation
	Index        int               `json:"index"`         // Index of the choice
	Logprobs     *LogProbs         `json:"logprobs"`      // Log probability information
	Message      CompletionMessage `json:"message"`       // Generated message
}

type LogProbs struct {
	Content any `json:"content"` // Log probabilities for each token
	Refusal any `json:"refusal"` // Refusal information if any
}

type CompletionMessage struct {
	Content     *string      `json:"content"`               // Message content
	Refusal     *string      `json:"refusal"`               // Refusal message if any
	Role        string       `json:"role"`                  // Role of message author
	Annotations []Annotation `json:"annotations,omitempty"` // Optional annotations
	Audio       *AudioOutput `json:"audio,omitempty"`       // Audio output if requested
	ToolCalls   []ToolCall   `json:"tool_calls,omitempty"`  // Tool calls made by assistant
}

type ToolCall struct {
	Function FunctionCall `json:"function"` // Function details
	ID       string       `json:"id"`       // Unique identifier for the tool call
	Type     string       `json:"type"`     // Type of tool call
}

type FunctionCall struct {
	Name      string `json:"name"`      // Name of the function
	Arguments any    `json:"arguments"` // Arguments for the function
}

type Annotation struct {
	Type        string       `json:"type"`         // Type of annotation
	URLCitation *URLCitation `json:"url_citation"` // URL citation details
}

type URLCitation struct {
	URL        string `json:"url"`         // The cited URL
	Title      string `json:"title"`       // Title of the cited content
	EndIndex   int    `json:"end_index"`   // End index of the citation
	StartIndex int    `json:"start_index"` // Start index of the citation
}

type AudioOutput struct {
	Data       string `json:"data"`       // Base64 encoded audio bytes
	ExpiresAt  int    `json:"expires_at"` // Expiration timestamp
	ID         string `json:"id"`         // Unique audio response ID
	Transcript string `json:"transcript"` // Audio transcript
}

type Usage struct {
	CompletionTokens        int           `json:"completion_tokens"`                   // Completion tokens count
	PromptTokens            int           `json:"prompt_tokens"`                       // Prompt tokens count
	TotalTokens             int           `json:"total_tokens"`                        // Total tokens used
	CompletionTokensDetails *TokenDetails `json:"completion_tokens_details,omitempty"` // Detailed completion tokens
	PromptTokensDetails     *TokenDetails `json:"prompt_tokens_details,omitempty"`     // Detailed prompt tokens
}

type TokenDetails struct {
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens,omitempty"` // Accepted prediction tokens
	AudioTokens              int `json:"audio_tokens,omitempty"`               // Audio tokens
	ReasoningTokens          int `json:"reasoning_tokens,omitempty"`           // Reasoning tokens
	RejectedPredictionTokens int `json:"rejected_prediction_tokens,omitempty"` // Rejected prediction tokens
	CachedTokens             int `json:"cached_tokens,omitempty"`              // Cached tokens
}
