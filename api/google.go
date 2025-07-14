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
		Text         string              `json:"text,omitempty"`
		InlineData   *GeminiPartInline   `json:"inline_data,omitempty"`
		FunctionCall *GeminiFunctionCall `json:"functionCall,omitempty"`
		Thought      string              `json:"thought,omitempty"`
	}
	GeminiPartInline struct {
		MimeType string `json:"mime_type"`
		Data     string `json:"data"`
	}

	// GenerationConfig represents the generation configuration for the Google API.
	GenerationConfig struct {
		ResponseMimeType string           `json:"responseMimeType,omitempty"`
		ResponseSchema   GeminiJSONSchema `json:"responseSchema,omitempty"`
		ThinkingConfig   *ThinkingConfig  `json:"thinkingConfig,omitempty"`
		StopSequences    []string         `json:"stopSequences,omitempty"`
		Temperature      *float64         `json:"temperature,omitempty"`
		MaxOutputTokens  *int             `json:"maxOutputTokens,omitempty"`
		TopP             *float64         `json:"topP,omitempty"`
		TopK             *int             `json:"topK,omitempty"`
	}

	// ThinkingConfig represents the thinking configuration for the Google API.
	ThinkingConfig struct {
		ThinkingBudget int `json:"thinkingBudget,omitempty"` // 0 => off, -1 => dynamic
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

	GeminiFunctionCall struct {
		Name string                 `json:"name"`
		Args map[string]interface{} `json:"args"`
	}

	GeminiUsageMetadata struct {
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		PromptTokenCount     int `json:"promptTokenCount"`
		ThoughtsTokenCount   int `json:"thoughtsTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
		PromptTokensDetails  []struct {
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
	defer func() { _ = resp.Body.Close() }()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading Gemini response body: %w", err)
	}

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("Gemini response: ", string(body))

	// Parse the response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling Gemini response: %w", err)
	}

	// Check if there's an error in the response
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("gemini API returned error: %s", geminiResp.Error.Message)
	}

	b, err := json.MarshalIndent(geminiResp, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling Gemini response: %w", err)
	}
	fmt.Println("Gemini response: ", string(b))

	// Check if we have candidates
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini API returned no candidates")
	}

	fmt.Println("Gemini response: ", geminiResp.Candidates)

	// Convert the Gemini response to OpenAI response
	response, err := openAIResponseFromGeminiResponse(&geminiResp)
	if err != nil {
		return nil, fmt.Errorf("error converting Gemini response to OpenAI response: %v", err)
	}
	return &Response{Response: response, Error: nil}, nil
}

// Convert OpenAI request to Gemini request
func geminiRequestFromOpenAIRequest(request *openai.ChatCompletionRequest) (*GeminiRequest, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}

	// Iterate through messages and convert them to Gemini format
	var systemInstructions GeminiPart
	var contents []GeminiMessage
	var tools []GeminiTool

	for _, message := range request.Messages {
		if message.Role == "system" {
			if content, ok := message.Content.(string); ok {
				systemInstructions = GeminiPart{Text: content}
			}
		} else {
			if content, ok := message.Content.(string); ok {
				contents = append(contents, GeminiMessage{
					Role:  message.Role,
					Parts: []GeminiPart{{Text: content}},
				})
			}
		}
	}

	if request.Tools != nil {
		tools = geminiToolsFromOpenAIRequest(request.Tools)
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

	config := &GenerationConfig{
		StopSequences:   stops,
		Temperature:     request.Temperature,
		MaxOutputTokens: request.MaxCompletionTokens,
		TopP:            request.TopP,
	}

	// Only set JSON response format if we have a schema
	if request.ResponseFormat != nil && request.ResponseFormat.JSONSchema != nil {
		schema := request.ResponseFormat.JSONSchema.Schema
		if schema != nil {
			// delete(schema, "additionalProperties")
			config.ResponseMimeType = "application/json"
			geminiSchema := NewGeminiJSONSchema(schema)
			if geminiSchema != nil {
				config.ResponseSchema = *geminiSchema
			}
		}
	}

	geminiReq := &GeminiRequest{
		SystemInstructions: GeminiSystemInstruction{Parts: []GeminiPart{systemInstructions}},
		Contents:           contents,
		GenerationConfig:   config,
		Tools:              tools,
	}

	return geminiReq, nil
}

func geminiToolsFromOpenAIRequest(tools []openai.Tool) []GeminiTool {
	geminiTools := make([]GeminiTool, 0)
	for _, tool := range tools {
		geminiTools = append(geminiTools, GeminiTool{
			Functions: []GeminiFunction{
				{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			},
		})
	}
	return geminiTools
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
		var content *string
		var toolCalls []openai.ToolCall

		// Process all parts to collect content and tool calls
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				content = &part.Text
			}
			if part.FunctionCall != nil {
				toolCalls = append(toolCalls, openai.ToolCall{
					Function: openai.FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: part.FunctionCall.Args,
					},
				})
			}
		}

		choices = append(choices, openai.Choice{
			FinishReason: candidate.Content.FinishReason,
			Index:        candidate.Content.Index,
			Message: openai.CompletionMessage{
				Content:   content,
				ToolCalls: toolCalls,
				Role:      candidate.Content.Role,
			},
		})
	}

	// Set the choices in the response
	resp.Choices = choices
	return resp, nil

}

// potentially do: toOpenAIResponse and fromOpenAIResponse as functions that take an openai.ChatCompletionRequest
// and return a google compatible request and similary for the response

// Type represents the OpenAPI data types
type GeminiJSONSchemaType string

const (
	TypeString  GeminiJSONSchemaType = "string"
	TypeInteger GeminiJSONSchemaType = "integer"
	TypeNumber  GeminiJSONSchemaType = "number"
	TypeBoolean GeminiJSONSchemaType = "boolean"
	TypeArray   GeminiJSONSchemaType = "array"
	TypeObject  GeminiJSONSchemaType = "object"
)

// Schema represents a JSON schema structure compatible with Google API
type GeminiJSONSchema struct {
	// Core fields
	Type        GeminiJSONSchemaType `json:"type,omitempty"`
	Format      string               `json:"format,omitempty"`
	Description string               `json:"description,omitempty"`
	Nullable    *bool                `json:"nullable,omitempty"`

	// Enum values (valid for string, integer, number)
	Enum []string `json:"enum,omitempty"`

	// Array-specific fields
	MaxItems *int              `json:"maxItems,omitempty"`
	MinItems *int              `json:"minItems,omitempty"`
	Items    *GeminiJSONSchema `json:"items,omitempty"`

	// Object-specific fields
	Properties       map[string]*GeminiJSONSchema `json:"properties,omitempty"`
	Required         []string                     `json:"required,omitempty"`
	PropertyOrdering []string                     `json:"propertyOrdering,omitempty"`

	// Number/Integer-specific fields (mentioned in comments but not in original schema)
	Minimum *float64 `json:"minimum,omitempty"`
	Maximum *float64 `json:"maximum,omitempty"`
}

// NewGeminiJSONSchema creates a new GeminiJSONSchema from a standard JSON schema
func NewGeminiJSONSchema(schema map[string]any) *GeminiJSONSchema {
	result := &GeminiJSONSchema{}

	// Handle type conversion
	if typeVal, ok := schema["type"]; ok {
		switch typeStr := typeVal.(string); typeStr {
		case "string":
			result.Type = TypeString
		case "integer":
			result.Type = TypeInteger
		case "number":
			result.Type = TypeNumber
		case "boolean":
			result.Type = TypeBoolean
		case "array":
			result.Type = TypeArray
		case "object":
			result.Type = TypeObject
		default:
			// Default to string for unknown types
			result.Type = TypeString
		}
	} else {
		// Default to object if no type specified
		result.Type = TypeObject
	}

	// Handle description
	if desc, ok := schema["description"].(string); ok {
		result.Description = desc
	}

	// Handle format
	if format, ok := schema["format"].(string); ok {
		result.Format = format
	}

	// Handle nullable (could be direct boolean or from anyOf/oneOf patterns)
	if nullable, ok := schema["nullable"].(bool); ok {
		result.Nullable = &nullable
	}

	// Handle enum values
	if enumVal, ok := schema["enum"]; ok {
		if enumSlice, ok := enumVal.([]interface{}); ok {
			enumStrings := make([]string, 0, len(enumSlice))
			for _, v := range enumSlice {
				enumStrings = append(enumStrings, fmt.Sprintf("%v", v))
			}
			result.Enum = enumStrings
		}
	}

	// Handle array-specific fields
	if result.Type == TypeArray {
		if minItems, ok := schema["minItems"]; ok {
			if minInt, ok := minItems.(float64); ok {
				min := int(minInt)
				result.MinItems = &min
			}
		}

		if maxItems, ok := schema["maxItems"]; ok {
			if maxInt, ok := maxItems.(float64); ok {
				max := int(maxInt)
				result.MaxItems = &max
			}
		}

		if items, ok := schema["items"]; ok {
			if itemsMap, ok := items.(map[string]any); ok {
				result.Items = NewGeminiJSONSchema(itemsMap)
			}
		}
	}

	// Handle object-specific fields
	if result.Type == TypeObject {
		if properties, ok := schema["properties"]; ok {
			if propsMap, ok := properties.(map[string]any); ok {
				result.Properties = make(map[string]*GeminiJSONSchema)
				for propName, propSchema := range propsMap {
					if propMap, ok := propSchema.(map[string]any); ok {
						result.Properties[propName] = NewGeminiJSONSchema(propMap)
					}
				}
			}
		}

		if required, ok := schema["required"]; ok {
			if reqSlice, ok := required.([]interface{}); ok {
				requiredStrings := make([]string, 0, len(reqSlice))
				for _, v := range reqSlice {
					if str, ok := v.(string); ok {
						requiredStrings = append(requiredStrings, str)
					}
				}
				result.Required = requiredStrings
			}
		}

		// Handle property ordering if present (non-standard but supported by your schema)
		if ordering, ok := schema["propertyOrdering"]; ok {
			if orderSlice, ok := ordering.([]interface{}); ok {
				orderStrings := make([]string, 0, len(orderSlice))
				for _, v := range orderSlice {
					if str, ok := v.(string); ok {
						orderStrings = append(orderStrings, str)
					}
				}
				result.PropertyOrdering = orderStrings
			}
		}
	}

	// Handle number/integer-specific fields
	if result.Type == TypeNumber || result.Type == TypeInteger {
		if minimum, ok := schema["minimum"]; ok {
			if minFloat, ok := minimum.(float64); ok {
				result.Minimum = &minFloat
			}
		}

		if maximum, ok := schema["maximum"]; ok {
			if maxFloat, ok := maximum.(float64); ok {
				result.Maximum = &maxFloat
			}
		}
	}

	// Handle common JSON Schema patterns that might indicate nullable
	if result.Nullable == nil {
		// Check for anyOf/oneOf patterns that include null
		if anyOf, ok := schema["anyOf"]; ok {
			if anyOfSlice, ok := anyOf.([]interface{}); ok {
				hasNull := false
				for _, item := range anyOfSlice {
					if itemMap, ok := item.(map[string]any); ok {
						if typeVal, ok := itemMap["type"]; ok {
							if typeVal == "null" {
								hasNull = true
								break
							}
						}
					}
				}
				if hasNull {
					nullable := true
					result.Nullable = &nullable
				}
			}
		}

		if oneOf, ok := schema["oneOf"]; ok {
			if oneOfSlice, ok := oneOf.([]interface{}); ok {
				hasNull := false
				for _, item := range oneOfSlice {
					if itemMap, ok := item.(map[string]any); ok {
						if typeVal, ok := itemMap["type"]; ok {
							if typeVal == "null" {
								hasNull = true
								break
							}
						}
					}
				}
				if hasNull {
					nullable := true
					result.Nullable = &nullable
				}
			}
		}
	}

	return result
}

// Helper function to convert from JSON string to GeminiJSONSchema
func NewGeminiJSONSchemaFromJSON(jsonStr string) (*GeminiJSONSchema, error) {
	var schema map[string]any
	err := json.Unmarshal([]byte(jsonStr), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON schema: %w", err)
	}
	return NewGeminiJSONSchema(schema), nil
}

// SetNullable sets the nullable field (helper method)
func (s *GeminiJSONSchema) SetNullable(nullable bool) *GeminiJSONSchema {
	s.Nullable = &nullable
	return s
}

// SetFormat sets the format field
func (s *GeminiJSONSchema) SetFormat(format string) *GeminiJSONSchema {
	s.Format = format
	return s
}

// SetDescription sets the description field
func (s *GeminiJSONSchema) SetDescription(description string) *GeminiJSONSchema {
	s.Description = description
	return s
}

// SetEnum sets the enum values
func (s *GeminiJSONSchema) SetEnum(values []string) *GeminiJSONSchema {
	s.Enum = values
	return s
}

// SetMinItems sets the minimum items for arrays
func (s *GeminiJSONSchema) SetMinItems(min int) *GeminiJSONSchema {
	s.MinItems = &min
	return s
}

// SetMaxItems sets the maximum items for arrays
func (s *GeminiJSONSchema) SetMaxItems(max int) *GeminiJSONSchema {
	s.MaxItems = &max
	return s
}

// SetItems sets the items schema for arrays
func (s *GeminiJSONSchema) SetItems(items *GeminiJSONSchema) *GeminiJSONSchema {
	s.Items = items
	return s
}

// SetProperties sets the properties for objects
func (s *GeminiJSONSchema) SetProperties(properties map[string]*GeminiJSONSchema) *GeminiJSONSchema {
	s.Properties = properties
	return s
}

// AddProperty adds a single property to the schema
func (s *GeminiJSONSchema) AddProperty(name string, property *GeminiJSONSchema) *GeminiJSONSchema {
	if s.Properties == nil {
		s.Properties = make(map[string]*GeminiJSONSchema)
	}
	s.Properties[name] = property
	return s
}

// SetRequired sets the required fields for objects
func (s *GeminiJSONSchema) SetRequired(required []string) *GeminiJSONSchema {
	s.Required = required
	return s
}

// SetPropertyOrdering sets the property ordering for objects
func (s *GeminiJSONSchema) SetPropertyOrdering(ordering []string) *GeminiJSONSchema {
	s.PropertyOrdering = ordering
	return s
}

// SetMinimum sets the minimum value for numbers/integers
func (s *GeminiJSONSchema) SetMinimum(min float64) *GeminiJSONSchema {
	s.Minimum = &min
	return s
}

// SetMaximum sets the maximum value for numbers/integers
func (s *GeminiJSONSchema) SetMaximum(max float64) *GeminiJSONSchema {
	s.Maximum = &max
	return s
}

// Validate checks if the schema is valid according to the type-specific field restrictions
func (s *GeminiJSONSchema) Validate() error {
	switch s.Type {
	case TypeString:
		// Valid fields: enum, format, nullable
		if s.MaxItems != nil || s.MinItems != nil || s.Items != nil ||
			s.Properties != nil || s.Required != nil || s.PropertyOrdering != nil ||
			s.Minimum != nil || s.Maximum != nil {
			return fmt.Errorf("string type only supports enum, format, and nullable fields")
		}
	case TypeInteger, TypeNumber:
		// Valid fields: format, minimum, maximum, enum, nullable
		if s.MaxItems != nil || s.MinItems != nil || s.Items != nil ||
			s.Properties != nil || s.Required != nil || s.PropertyOrdering != nil {
			return fmt.Errorf("integer/number type only supports format, minimum, maximum, enum, and nullable fields")
		}
	case TypeBoolean:
		// Valid fields: nullable
		if s.Format != "" || s.Enum != nil || s.MaxItems != nil || s.MinItems != nil ||
			s.Items != nil || s.Properties != nil || s.Required != nil ||
			s.PropertyOrdering != nil || s.Minimum != nil || s.Maximum != nil {
			return fmt.Errorf("boolean type only supports nullable field")
		}
	case TypeArray:
		// Valid fields: minItems, maxItems, items, nullable
		if s.Format != "" || s.Enum != nil || s.Properties != nil ||
			s.Required != nil || s.PropertyOrdering != nil ||
			s.Minimum != nil || s.Maximum != nil {
			return fmt.Errorf("array type only supports minItems, maxItems, items, and nullable fields")
		}
	case TypeObject:
		// Valid fields: properties, required, propertyOrdering, nullable
		if s.Format != "" || s.Enum != nil || s.MaxItems != nil ||
			s.MinItems != nil || s.Items != nil || s.Minimum != nil || s.Maximum != nil {
			return fmt.Errorf("object type only supports properties, required, propertyOrdering, and nullable fields")
		}
	default:
		return fmt.Errorf("unsupported type: %s", s.Type)
	}
	return nil
}
