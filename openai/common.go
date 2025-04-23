package openai

/*

## Most Commonly Used OpenAI API Endpoints

**Chat Completions Endpoint**
- The `/v1/chat/completions` endpoint is the most widely used API endpoint for OpenAI. It powers conversational applications and supports advanced models like GPT-3.5 and GPT-4, allowing for both single-prompt and multi-turn chat interactions[1][4][7].

**Completions Endpoint**
- The `/v1/completions` endpoint generates text completions from a single prompt. It is commonly used for tasks that require straightforward text generation, such as summarization or answering questions, and is often used with older models like text-davinci-003[4][5][7].

**Embeddings Endpoint**
- The `/v1/embeddings` endpoint returns vector representations of input text, which are useful for tasks like semantic search, clustering, and recommendation systems[3][4][7].

**Images Endpoint**
- The `/v1/images` endpoint is used for generating images from text prompts, modifying existing images, or creating image variations, leveraging models like DALL-E[4][5][7].


**Models Endpoint**
- The `/v1/models` endpoint lists available models and provides metadata about them, such as ownership and permissions[4][7].


---

### Summary Table

| Endpoint             | Primary Use Case                          |
|----------------------|-------------------------------------------|
| `/v1/chat/completions` | Conversational AI, chatbots              |
| `/v1/completions`      | Text generation, Q&A, summarization      |
| `/v1/embeddings`       | Semantic search, recommendations         |
| `/v1/images`           | Image generation and editing             |
| `/v1/models`           | Model listing and metadata               |


*/

// Common response formats
type ErrorResponse struct {
	Error struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Param   interface{} `json:"param"`
		Code    interface{} `json:"code"`
	} `json:"error"`
}

// Chat Completions Types
type ChatCompletionRequest struct {
	Model            string                  `json:"model"`
	Messages         []ChatCompletionMessage `json:"messages"`
	Temperature      float32                 `json:"temperature,omitempty"`
	TopP             float32                 `json:"top_p,omitempty"`
	N                int                     `json:"n,omitempty"`
	Stream           bool                    `json:"stream,omitempty"`
	Stop             interface{}             `json:"stop,omitempty"`
	MaxTokens        int                     `json:"max_tokens,omitempty"`
	PresencePenalty  float32                 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32                 `json:"frequency_penalty,omitempty"`
	User             string                  `json:"user,omitempty"`
}

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   CompletionUsage        `json:"usage"`
}

type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

// Completions Types
type CompletionRequest struct {
	Model            string      `json:"model"`
	Prompt           interface{} `json:"prompt"`
	MaxTokens        int         `json:"max_tokens,omitempty"`
	Temperature      float32     `json:"temperature,omitempty"`
	TopP             float32     `json:"top_p,omitempty"`
	N                int         `json:"n,omitempty"`
	Stream           bool        `json:"stream,omitempty"`
	Stop             interface{} `json:"stop,omitempty"`
	PresencePenalty  float32     `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32     `json:"frequency_penalty,omitempty"`
	User             string      `json:"user,omitempty"`
}

type CompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   CompletionUsage    `json:"usage"`
}

type CompletionChoice struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

type CompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Embeddings Types
type EmbeddingRequest struct {
	Model string      `json:"model"`
	Input interface{} `json:"input"`
	User  string      `json:"user,omitempty"`
}

type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// Image Types
type ImageGenerationRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
	User   string `json:"user,omitempty"`
}

type ImageResponse struct {
	Created int64       `json:"created"`
	Data    []ImageData `json:"data"`
}

type ImageData struct {
	URL string `json:"url"`
}

// Models Types
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ModelList struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}
