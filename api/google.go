package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"llm-balancer/openai"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type (
	GoogleClient struct {
		BaseURL string
		APIKey  string
	}

	// Request represents a request to the Google API.
	GeminiRequest struct {
		SystemInstructions GeminiSystemInstruction `json:"system_instruction"`
		Contents           []GeminiMessage         `json:"contents"`
		Tools              []GeminiTool            `json:"tools,omitempty"`
		GenerationConfig   *GenerationConfig       `json:"generationConfig,omitempty"`
	}

	GeminiMessage struct {
		Role  string       `json:"role"`
		Parts []GeminiPart `json:"parts"`
	}

	GeminiSystemInstruction struct {
		Parts []GeminiPart `json:"parts"`
	}

	GeminiPart struct {
		Text       string            `json:"text,omitempty"`
		InlineData *GeminiPartInline `json:"inline_data,omitempty"`
	}

	GeminiPartInline struct {
		MimeType string `json:"mime_type"`
		Data     string `json:"data"`
	}

	// GenerationConfig represents the generation configuration for the Google API.
	GenerationConfig struct {
		ResponseMimeType string          `json:"responseMimeType,omitempty"`
		ResponseSchema   map[string]any  `json:"responseSchema,omitempty"`
		ThinkingConfig   *ThinkingConfig `json:"thinkingConfig,omitempty"`
		StopSequences    []string        `json:"stopSequences,omitempty"`
		Temperature      *float64        `json:"temperature,omitempty"`
		MaxOutputTokens  *int            `json:"maxOutputTokens,omitempty"`
		TopP             *float64        `json:"topP,omitempty"`
		TopK             *int            `json:"topK,omitempty"`
	}

	// ThinkingConfig represents the thinking configuration for the Google API.
	ThinkingConfig struct {
		ThinkingBudget int `json:"thinkingBudget,omitempty"`
	}

	// GeminiTool represents a tool for the Google API.
	GeminiTool struct {
		Functions []GeminiFunction `json:"functionDeclarations"`
	}
	GeminiFunction struct {
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		Parameters  map[string]any `json:"parameters,omitempty"`
	}

	// Define the Gemini response structure
	GeminiResponse struct {
		Candidates    []GeminiCandidate   `json:"candidates"`
		ModelVersion  string              `json:"modelVersion"`
		UsageMetadata GeminiUsageMetadata `json:"usageMetadata"`
		Error         *GeminiError        `json:"error,omitempty"`
	}

	GeminiContent struct {
		Role         string       `json:"role"`
		FinishReason string       `json:"finishReason"`
		Index        int          `json:"index"`
		Parts        []GeminiPart `json:"parts"`
	}

	GeminiCandidate struct {
		Content GeminiContent `json:"content"`
	}

	GeminiUsageMetadata struct {
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		PromptTokenCount     int `json:"promptTokenCount"`
		ThoughtsTokenCount   int `json:"thoughtsTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
		PromptTokenDetails   []struct {
			Modality   string `json:"modality"`
			TokenCount int    `json:"tokenCount"`
		}
	}

	GeminiError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	}
)

// NewGoogleClient creates a new Google API client.
func NewGoogleClient(baseURL string, apiKey string) *GoogleClient {
	return &GoogleClient{BaseURL: baseURL, APIKey: apiKey}
}

// POSTChatCompletion sends a chat completion request to the Google API.
func (c *GoogleClient) POSTChatCompletion(ctx context.Context, request *Request, model string) (*Response, error) {
	// Prepare the request URL
	url := fmt.Sprintf("%s/models/%s:generateContent", c.BaseURL, model)

	geminiRequest, err := geminiRequestFromOpenAIRequest(request.Request)
	if err != nil {
		return nil, fmt.Errorf("error converting OpenAI request to Gemini request: %v", err)
	}

	// Set the request body to the modified request
	jsonBody, err := json.Marshal(geminiRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// For Gemini, the API key is usually included as a query parameter
	q := req.URL.Query()
	q.Add("key", c.APIKey)
	req.URL.RawQuery = q.Encode()
	req = req.WithContext(ctx)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making Gemini request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading Gemini response body: %w", err)
	}

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling Gemini response: %w", err)
	}

	// Check if there's an error in the response
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API returned error: %s", geminiResp.Error.Message)
	}

	// Check if we have candidates
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("Gemini API returned no candidates")
	}

	// Convert the Gemini response to OpenAI response
	response, err := openAIResponseFromGeminiResponse(&geminiResp)
	if err != nil {
		return nil, fmt.Errorf("error converting Gemini response to OpenAI response: %v", err)
	}
	return &Response{Response: response, Error: nil}, nil
}

// Convert OpenAI request to Gemini request
func geminiRequestFromOpenAIRequest(request *openai.ChatCompletionRequest) (*GeminiRequest, error) {
	// Iterate through messages and convert them to Gemini format
	var systemInstructions GeminiPart
	var contents []GeminiMessage
	for _, message := range request.Messages {
		if message.Role == "system" {
			systemInstructions = GeminiPart{Text: message.Content.(string)}
		} else {
			contents = append(contents, GeminiMessage{
				Role:  message.Role,
				Parts: []GeminiPart{{Text: message.Content.(string)}},
			})
		}
	}

	// Set the generation config if provided
	stops := make([]string, 0)
	if request.Stop != nil {
		switch v := request.Stop.(type) {
		case string:
			stops = append(stops, v)
		case []string:
			stops = append(stops, v...)
		default:
			return nil, fmt.Errorf("unsupported stop type: %T", v)
		}
	}
	schema := request.ResponseFormat.JSONSchema.Schema
	if schema != nil {
		fmt.Println("Schema: ", schema)
		// delete the additional properties from the schema
		if _, ok := schema["additionalProperties"]; ok {
			delete(schema, "additionalProperties")
		}
	}
	config := &GenerationConfig{
		ResponseMimeType: "application/json",
		ResponseSchema:   schema,
		StopSequences:    stops,
		Temperature:      request.Temperature,
		MaxOutputTokens:  request.MaxCompletionTokens,
		TopP:             request.TopP,
	}

	geminiReq := &GeminiRequest{
		SystemInstructions: GeminiSystemInstruction{Parts: []GeminiPart{systemInstructions}},
		Contents:           contents,
		GenerationConfig:   config,
	}

	return geminiReq, nil
}

func openAIResponseFromGeminiResponse(geminiResp *GeminiResponse) (*openai.ChatCompletionResponse, error) {
	resp := &openai.ChatCompletionResponse{
		ID:                uuid.New().String(),
		Created:           int(time.Now().Unix()),
		Model:             geminiResp.ModelVersion,
		Object:            "chat.completion",
		ServiceTier:       nil,
		SystemFingerprint: geminiResp.ModelVersion,
		Usage: openai.Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}

	// Convert Gemini response to OpenAI response
	var choices []openai.Choice
	for _, candidate := range geminiResp.Candidates {
		choices = append(choices, openai.Choice{
			FinishReason: candidate.Content.FinishReason,
			Index:        candidate.Content.Index,
			Message: openai.CompletionMessage{
				Content: &candidate.Content.Parts[0].Text,
				Role:    "assistant",
			},
		})
	}

	// Set the choices in the response
	resp.Choices = choices
	return resp, nil

}

// potentially do: toOpenAIResponse and fromOpenAIResponse as functions that take an openai.ChatCompletionRequest
// and return a google compatible request and similary for the response
