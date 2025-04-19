package main

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"llm-balancer/balancer"
	"llm-balancer/config"
)

// BytesPerToken remains here or move to a utility package? Keep simple for MVP.
const (
	BytesPerToken = 16 // Approximately 16 bytes per token
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml") // Use config package
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Set log level
	switch cfg.General.LogLevel {
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

	if len(cfg.LLMAPIs) == 0 {
		log.Fatal().Msg("No LLM APIs configured")
	}

	// Create and initialize the balancer
	b, err := balancer.NewBalancer(cfg) // Use balancer package
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create balancer")
	}

	// Set up HTTP handler using the balancer instance
	// Use GIN to run the server
	// Make 2 groups /llm/v1 or /v1/llm and /api/v1 etc
	http.HandleFunc("/", b.HandleRequest) // Use balancer's method
	// TODO: Add handler for new llm like `add this llm to the list of available ones`
	// TODO: Add handler to modify the config like a patch
	// TODO: Add a catch all the rest and give a 404

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.General.ListenAddress, cfg.General.ListenPort)
	log.Info().
		Str("address", addr).
		Msg("Starting server")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
