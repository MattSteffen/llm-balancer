package llm

import (
	"fmt"
	"llm-balancer/api"
	"llm-balancer/config"
	"sync"
)

type LLM struct {
	Provider       string
	Model          string
	BaseURL        string
	RateLimit      int
	RequestsPerMin int
	ContextLength  int
	ApiKeyName     string
	APIType        string
	Price          float64
	Quality        int

	client api.Client

	Mu           sync.Mutex
	TokensLeft   int
	RequestsLeft int
}

func NewLLM(config config.LLMApiConfig) (*LLM, error) {
	return nil, nil
}

// determines if it has the token and request budget for the request
func (llm *LLM) IsAvailable(tokensNeeded int) bool {
	if llm.TokensLeft < 2*tokensNeeded && llm.RequestsLeft > 1 {
		return true
	}
	return false
}

// decreases requests left and tokens left
func (llm *LLM) Decrement(tokensUsed int) error {
	llm.TokensLeft -= tokensUsed
	llm.RequestsLeft -= 1
	if llm.TokensLeft < 0 || llm.RequestsLeft < 0 {
		return fmt.Errorf("Invalid use of LLM %s, not enough tokens or requests", llm.Model)
	}
	return nil
}

func (llm *LLM) RefillCounters() {}
