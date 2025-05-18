package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"llm-balancer/balancer"
	"llm-balancer/config"
	"llm-balancer/handlers"
)

// BytesPerToken remains here or move to a utility package? Keep simple for MVP.
const (
	BytesPerToken = 16 // Approximately 16 bytes per token
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath) // Use config package
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

	balancer, err := balancer.NewPool(balancer.Config{
		Models:         cfg.LLMAPIs,
		SortStrategy:   &balancer.QualitySortStrategy{},
		ContextTimeout: time.Duration(cfg.General.ContextTimeout) * time.Second,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create balancer pool")
	}

	handler := handlers.NewHandler(balancer) // Use handlers package

	// Make 2 groups /llm/v1 or /v1/llm and /api/v1 etc
	http.HandleFunc("/v1/chat/completions", handler.HandleChatCompletion) // Use handler's method
	// TODO: Add handler for new llm like `add this llm to the list of available ones`
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
