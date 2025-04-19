package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// GeneralConfig holds the general server settings.
type GeneralConfig struct {
	ListenAddress string `yaml:"listen_address"`
	ListenPort    int    `yaml:"listen_port"`
	LogLevel      string `yaml:"log_level"`
	// OptimizationWeights map[string]float64 `yaml:"optimization_weights"` // For future use
}

// LLMApiConfig holds the configuration for each LLM API.
type LLMApiConfig struct {
	Provider      string   `yaml:"provider"`
	Model         string   `yaml:"model"`
	BaseURL       string   `yaml:"base_url"`
	Modes         []string `yaml:"modes"`       // text, vision, audio, etc
	RateLimit     int      `yaml:"rate_limit"`  // requests per minute
	TokenLimit    int      `yaml:"token_limit"` // tokens per minute
	ContextLength int      `yaml:"context_length"`
	ApiKeyName    string   `yaml:"api_key_name"`
	CostInput     float64  `yaml:"cost_input"`  // in dollars
	CostOutput    float64  `yaml:"cost_output"` // in dollars
	Price         float64  `yaml:"price"`
	Quality       int      `yaml:"quality"`
}

// Config is the root configuration struct.
type Config struct {
	General GeneralConfig  `yaml:"general"`
	LLMAPIs []LLMApiConfig `yaml:"llms"`
}

// LoadConfig reads the YAML config file and unmarshals it into a Config struct.
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
