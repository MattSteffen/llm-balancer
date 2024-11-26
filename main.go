package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Config structures as defined in the README
type Config struct {
	General GeneralConfig  `yaml:"general"`
	LLMAPIs []LLMApiConfig `yaml:"llms"`
}

type GeneralConfig struct {
	ListenAddress string `yaml:"listen_address"`
	ListenPort    int    `yaml:"listen_port"`
	LogLevel      string `yaml:"log_level"`
}

type LLMApiConfig struct {
	Name           string `yaml:"name"`
	Model          string `yaml:"model"`
	URL            string `yaml:"url"`
	RateLimit      int    `yaml:"rate_limit"`
	RequestsPerMin int    `yaml:"requests_per_min"`
	ContextLength  int    `yaml:"context_length"`
	ApiKeyName     string `yaml:"api_key_name"`
}

// LLM represents a configured LLM instance
type LLM struct {
	Name          string
	Model         string
	URL           string
	RateLimit     int
	ContextLength int
	ApiKeyName    string
	TokensLeft    int
	// mu            sync.Mutex
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

func createLLM(config LLMApiConfig) *LLM {
	return &LLM{
		Name:          config.Name,
		Model:         config.Model,
		URL:           config.URL,
		RateLimit:     config.RateLimit,
		ContextLength: config.ContextLength,
		ApiKeyName:    config.ApiKeyName,
		TokensLeft:    config.RateLimit,
	}
}

func forwardRequest(llms []*LLM, w http.ResponseWriter, r *http.Request) error {
	// Log the incoming request
	log.Info().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Msg("Received request")

	// Read the incoming body
	var originalBody map[string]interface{}
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error reading request body: %w", err)
		}
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &originalBody); err != nil {
				return fmt.Errorf("error unmarshaling request body: %w", err)
			}
		}
	}

	// select the appropriate LLM
	llm := llms[0] // For now, just use the first configured API

	// Add the model to the body
	if originalBody == nil {
		originalBody = make(map[string]interface{})
	}
	originalBody["model"] = llm.Model

	// Marshal the updated body to JSON
	updatedBodyBytes, err := json.Marshal(originalBody)
	if err != nil {
		return fmt.Errorf("error marshaling updated request body: %w", err)
	}

	// Create a new request to forward
	proxyReq, err := http.NewRequest(r.Method, llm.URL, io.NopCloser(bytes.NewReader(updatedBodyBytes)))
	if err != nil {
		return fmt.Errorf("error creating proxy request: %w", err)
	}

	// Copy original headers
	for header, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	// Replace or add the API key header
	apiKey := os.Getenv(llm.ApiKeyName)
	if apiKey == "" {
		return fmt.Errorf("API key not found for %s", llm.Name)
	}
	proxyReq.Header.Set("Authorization", "Bearer "+apiKey)
	proxyReq.Header.Set("Content-Type", "application/json") // Ensure the content type is correct

	// Log where the request is being sent
	log.Info().
		Str("llm_name", llm.Name).
		Str("llm_url", llm.URL).
		Msg("Forwarding request to LLM")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("error forwarding request: %w", err)
	}
	defer resp.Body.Close()

	// Copy response headers
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	// Copy status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("error copying response: %w", err)
	}

	return nil
}

func main() {
	// Set up zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	if len(config.LLMAPIs) == 0 {
		log.Fatal().Msg("No LLM APIs configured")
	}

	// Set log level
	switch config.General.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// For MVP, we'll just use the first configured API
	llms := make([]*LLM, len(config.LLMAPIs))
	for ind, llm := range config.LLMAPIs {
		llms[ind] = createLLM(llm)
	}

	// Create handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := forwardRequest(llms, w, r); err != nil {
			log.Error().Err(err).Msg("Error handling request")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	// Start server
	log.Info().
		Str("address", config.General.ListenAddress).
		Int("port", config.General.ListenPort).
		Msg("Starting server")
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.General.ListenAddress, config.General.ListenPort), nil); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
