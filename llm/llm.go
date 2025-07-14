package llm

import (
	"fmt"
	"llm-balancer/api"
	"os"

	"github.com/rs/zerolog/log"
)

// LLMApiConfig holds the configuration for each LLM API.
type LLM struct {
	Name           string   `yaml:"name" json:"name"`
	Provider       string   `yaml:"provider" json:"provider"`
	Model          string   `yaml:"model" json:"model"`
	BaseURL        string   `yaml:"base_url" json:"base_url"`
	TokensPerMin   int      `yaml:"tokens_per_minute" json:"tokens_per_minute"`     // tokens per minute
	RequestsPerMin int      `yaml:"requests_per_minute" json:"requests_per_minute"` // requests per minute
	ContextLength  int      `yaml:"context_length" json:"context_length"`
	CostInput      float64  `yaml:"cost_input" json:"cost_input"`   // in dollars
	CostOutput     float64  `yaml:"cost_output" json:"cost_output"` // in dollars
	Quality        int      `yaml:"quality" json:"quality"`
	APIKey         string   `yaml:"api_key" json:"-"`                 // API key for the provider
	APIKeyName     string   `yaml:"api_key_name" json:"api_key_name"` // API key name for the provider
	Modalities     []string `yaml:"modalities" json:"modalities"`     // text, vision, audio, etc
	Groups         []string `yaml:"groups" json:"groups"`

	Client api.Client `yaml:"-"` // API client for the provider
}

func (llm *LLM) String() string {
	return fmt.Sprintf("%s-%s", llm.Provider, llm.Model)
}

func (llm *LLM) Validate() bool {
	// Check if all required fields are set
	if llm.Provider == "" || llm.Model == "" || llm.BaseURL == "" || llm.RequestsPerMin <= 0 || llm.TokensPerMin <= 0 {
		return false
	}

	if len(llm.Modalities) == 0 {
		llm.Modalities = []string{"text"} // default to text mode if none specified
	}

	if llm.ContextLength <= 0 {
		llm.ContextLength = 4096 * 8 // default context length
	}

	if llm.APIKey == "" {
		apiKey := os.Getenv(llm.APIKeyName) // use environment variable if API key is not provided
		if apiKey == "" {
			log.Warn().Msgf("API key for %s is not set and not provided in environment variable %s\n", llm.Provider, llm.APIKeyName)
			return false
		}
		llm.APIKey = apiKey
	}

	if err := llm.SetClient(); err != nil {
		return false
	}
	return true
}

func (llm *LLM) SetClient() error {
	// Initialize API client based on API provider
	switch llm.Provider {
	case "openai":
		llm.Client = api.NewOpenAIClient(llm.BaseURL, llm.APIKey)
	case "ollama":
		llm.Client = api.NewOpenAIClient(llm.BaseURL, llm.APIKey) // should be ollama client
	case "groq":
		llm.Client = api.NewOpenAIClient(llm.BaseURL, llm.APIKey)
	case "google":
		llm.Client = api.NewGoogleClient(llm.BaseURL, llm.APIKey)
	case "openrouter":
		llm.Client = api.NewOpenAIClient(llm.BaseURL, llm.APIKey)
	default:
		return fmt.Errorf("unsupported provider: %s", llm.Provider)
	}

	return nil
}
