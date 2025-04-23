package llm

import (
	"fmt"
	"llm-balancer/api"
	"llm-balancer/config"
	"sync"
)

// TODO: This should match the config file
type LLM struct {
	Provider       string
	Model          string
	BaseURL        string
	RateLimit      int
	RequestsPerMin int
	ContextLength  int
	Price          float64
	Quality        int
	APIType        string // openai, ollama, etc. Not used yet

	client api.Client

	Mu           sync.Mutex
	TokensLeft   int
	RequestsLeft int
}

func NewLLM(config config.LLMApiConfig) (*LLM, error) {
	llm := &LLM{
		Provider:       config.Provider,
		Model:          config.Model,
		BaseURL:        config.BaseURL,
		RateLimit:      config.RateLimit,
		RequestsPerMin: config.RateLimit,
		ContextLength:  config.ContextLength,
		Price:          config.Price,
		Quality:        config.Quality,
		TokensLeft:     config.ContextLength,
		RequestsLeft:   config.RateLimit,
	}

	return llm, nil
}

// determines if it has the token and request budget for the request
func (llm *LLM) IsAvailable(tokensNeeded int) bool {
	return llm.TokensLeft >= tokensNeeded && llm.RequestsLeft > 0
}

// decreases requests left and tokens left
func (llm *LLM) Decrement(tokensUsed int) error {
	llm.TokensLeft -= tokensUsed
	llm.RequestsLeft -= 1
	if llm.TokensLeft < 0 || llm.RequestsLeft < 0 {
		return fmt.Errorf("invalid use of LLM %s, not enough tokens or requests", llm.Model)
	}
	return nil
}

func (llm *LLM) RefillCounters() {
	llm.Mu.Lock()
	defer llm.Mu.Unlock()

	llm.TokensLeft = llm.ContextLength
	llm.RequestsLeft = llm.RequestsPerMin
}
