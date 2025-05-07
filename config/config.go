package config

import (
	"fmt"
	"llm-balancer/llm"
	"os"

	"gopkg.in/yaml.v3"
)

// GeneralConfig holds the general server settings.
type GeneralConfig struct {
	ListenAddress  string `yaml:"listen_address"`
	ListenPort     int    `yaml:"listen_port"`
	LogLevel       string `yaml:"log_level"`
	ContextTimeout int    `yaml:"context_timeout"` // in seconds
	// OptimizationWeights map[string]float64 `yaml:"optimization_weights"` // For future use
}

// Config is the root configuration struct.
type Config struct {
	General GeneralConfig `yaml:"general"`
	LLMAPIs []*llm.LLM    `yaml:"llms"`
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
	if !cfg.Validate() {
		return nil, fmt.Errorf("invalid configuration")
	}
	return &cfg, nil
}

func (c *Config) Validate() bool {
	// Check if all required fields are set
	if c.General.ListenAddress == "" || c.General.ListenPort <= 0 {
		return false
	}
	if c.General.ContextTimeout <= 0 {
		c.General.ContextTimeout = 90 // default to 30 seconds
	}
	if len(c.LLMAPIs) == 0 {
		return false
	}
	for _, llm := range c.LLMAPIs {
		if !llm.Validate() {
			return false
		}
	}
	return true
}
