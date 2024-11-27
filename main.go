package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

/*
- [ ] Configure into packages and folders
- [ ] Unit tests
*/

const (
	BytesPerToken = 16 // Approximately 16 bytes per token
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
	Name           string  `yaml:"name"`
	Model          string  `yaml:"model"`
	URL            string  `yaml:"url"`
	RateLimit      int     `yaml:"rate_limit"`
	RequestsPerMin int     `yaml:"requests_per_minute"`
	ContextLength  int     `yaml:"context_length"`
	ApiKeyName     string  `yaml:"api_key_name"`
	Price          float64 `yaml:"price"`
	Quality        int     `yaml:"quality"`
}

// LLM represents a configured LLM instance
type LLM struct {
	Name           string
	Model          string
	URL            string
	RateLimit      int
	RequestsPerMin int
	ContextLength  int
	ApiKeyName     string
	Price          float64
	Quality        int

	mu           sync.Mutex
	TokensLeft   int
	RequestsLeft int
	Ready        bool
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
		Name:           config.Name,
		Model:          config.Model,
		URL:            config.URL,
		RateLimit:      config.RateLimit,
		RequestsPerMin: config.RequestsPerMin,
		ContextLength:  config.ContextLength,
		ApiKeyName:     config.ApiKeyName,
		Price:          config.Price,
		Quality:        config.Quality,

		TokensLeft:   config.RateLimit,
		RequestsLeft: config.RequestsPerMin,
		Ready:        true,
	}
}

func refillRateLimits(llms map[int]*LLM) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		for _, llm := range llms {
			llm.mu.Lock()
			llm.TokensLeft = llm.RateLimit
			llm.RequestsLeft = llm.RequestsPerMin
			llm.Ready = true
			llm.mu.Unlock()
			log.Info().Str("llm_name", llm.Name).Msg("Rate limits refilled")
		}
	}
}

func rankLLMs(llms []LLMApiConfig) map[int]*LLM {
	rankedllms := make(map[int]*LLM)
	for ind, llm := range llms {
		rankedllms[ind] = createLLM(llm)
	}
	return rankedllms
}

func selectLLM(llms map[int]*LLM, numTokens int) (*LLM, error) {
	var selected, potential *LLM

	for _, llm := range llms {
		llm.mu.Lock()
		fmt.Printf("llm.TokensLeft: %d, numTokens: %d\n", llm.TokensLeft, numTokens)
		fmt.Printf("llm.RequestsLeft: %d\n", llm.RequestsLeft)
		fmt.Printf("llm.requestsPerMin: %d\n", llm.RequestsPerMin)
		if llm.TokensLeft >= numTokens &&
			llm.RequestsLeft > 0 &&
			llm.ContextLength >= numTokens {
			selected = llm
		}
		if potential == nil && llm.ContextLength >= numTokens*2 { // TODO: numTokens*1.5
			potential = llm
		}
		llm.mu.Unlock()
	}

	if selected == nil {
		if potential != nil {
			selected = potential
			selected.Ready = false
		} else {
			return nil, fmt.Errorf("no LLM available for %d tokens", numTokens)
		}
	}

	// Update counters
	selected.mu.Lock()
	selected.TokensLeft -= numTokens
	selected.RequestsLeft--
	if selected.RequestsLeft == 0 || selected.TokensLeft < 1000 {
		selected.Ready = false
	}
	selected.mu.Unlock()

	return selected, nil
}

func forwardRequest(llms map[int]*LLM, w http.ResponseWriter, r *http.Request) error {
	// Log the incoming request
	log.Info().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Msg("Received request")

	// Read the incoming body
	var numTokens int
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
		numTokens = int(math.Ceil(float64(len(bodyBytes)) / BytesPerToken))
		log.Debug().Msgf("Size of body: %d -- using %d tokens", len(bodyBytes), numTokens)
	}

	// select the appropriate LLM
	llm, err := selectLLM(llms, numTokens)
	if err != nil {
		return fmt.Errorf("no suitable LLM: %w", err)
	}

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

	// Wait until the LLM is ready
	// TODO: Determine how to know if the LLM is ready
	// wait until llm.Ready is true (maybe check every 1s)
	// maybe something like sync.Cond

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

	if len(config.LLMAPIs) == 0 {
		log.Fatal().Msg("No LLM APIs configured")
	}

	// For MVP, we'll just use the first configured API
	llms := rankLLMs(config.LLMAPIs)

	go refillRateLimits(llms)

	// Create handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := forwardRequest(llms, w, r); err != nil {
			log.Error().Err(err).Msg("Error handling request")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", config.General.ListenAddress, config.General.ListenPort)
	log.Info().
		Str("address", addr).
		Msg("Starting server")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
