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
	General GeneralConfig       `yaml:"general"`
	LLMAPIs []*llm.LLM          `yaml:"llms"`
	Groups  map[string][]string `yaml:"-"` // maybe not a map[string][]string, but a struct with fields like free, fast, local, provider, etc.
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

	cfg.Groups = make(map[string][]string)
	cfg.Groups["free"] = []string{}
	cfg.Groups["fast"] = []string{}
	cfg.Groups["local"] = []string{}
	cfg.Groups["provider"] = []string{} // should be a map[string][]string like google, openai, etc. -> []string{model1, model2, ...}
	for _, llm := range cfg.LLMAPIs {
		cfg.Groups[llm.Provider] = append(cfg.Groups[llm.Provider], llm.Model)
		if llm.CostInput+llm.CostOutput == 0 {
			cfg.Groups["free"] = append(cfg.Groups["free"], llm.Model)
		}
		// TODO: add more groups based on local, quality, etc.
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
