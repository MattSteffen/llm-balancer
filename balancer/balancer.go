package balancer

import (
	"context"
	"errors"
	"llm-balancer/api"
	"llm-balancer/llm"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

var (
	cancel context.CancelFunc
)

// ModelLimiter wraps an LLM with both request and token limiters.
type ModelLimiter struct {
	LLM          *llm.LLM
	ReqLimiter   *rate.Limiter // limits requests per second
	TokenLimiter *rate.Limiter // limits tokens per second
}

// TODO: I need to make every family of LLMs have the same rate limiter, they must share accross source or name or api key

// SortStrategy defines how to order limiters when picking.
type SortStrategy interface {
	Sort(limiters []*ModelLimiter)
}

// QualitySortStrategy sorts by descending Quality
type QualitySortStrategy struct{}

// Sort implements SortStrategy; higher Quality first
func (s *QualitySortStrategy) Sort(limiters []*ModelLimiter) {
	sort.SliceStable(limiters, func(i, j int) bool {
		return limiters[i].LLM.Quality > limiters[j].LLM.Quality
	})
}

// Config holds pool initialization settings
type Config struct {
	Models         []*llm.LLM
	SortStrategy   SortStrategy
	ContextTimeout time.Duration // optional default timeout when waiting
}

// Pool manages multiple ModelLimiters
// and dispatches requests based on limiter availability.
type Pool struct {
	limiters       []*ModelLimiter
	sorter         SortStrategy
	mu             sync.Mutex
	next           int
	defaultTimeout time.Duration
}

// New creates a Pool given Config; error if no valid models.
// Initializes both request and token limiters for each model.
func NewPool(cfg Config) (*Pool, error) {
	if len(cfg.Models) == 0 {
		return nil, errors.New("no models provided")
	}
	if cfg.SortStrategy == nil {
		cfg.SortStrategy = &QualitySortStrategy{}
	}
	pool := &Pool{
		sorter:         cfg.SortStrategy,
		defaultTimeout: cfg.ContextTimeout,
	}
	for _, llm := range cfg.Models {
		if !llm.Validate() {
			return nil, errors.New("invalid model configuration: " + llm.String())
		}
		// compute request rate per second
		ratePerSec := rate.Limit(float64(llm.RequestsPerMin) / 60.0)
		if ratePerSec <= 0 {
			return nil, errors.New("invalid rate limit for model " + llm.Model)
		}
		// compute token rate per second
		tokenRate := rate.Limit(float64(llm.TokensPerMin) / 60.0)
		if tokenRate <= 0 {
			return nil, errors.New("invalid tokens per minute for model " + llm.Model)
		}

		ml := &ModelLimiter{
			LLM:          llm,
			ReqLimiter:   rate.NewLimiter(ratePerSec, llm.RequestsPerMin),
			TokenLimiter: rate.NewLimiter(tokenRate, llm.TokensPerMin),
		}
		pool.limiters = append(pool.limiters, ml)
	}
	pool.sorter.Sort(pool.limiters)
	return pool, nil
}

// Pick chooses the next available ModelLimiter.
// It only checks availability via Allow() (snon-blocking).
// Blocking for quota happens in Do(), so Pick never waits.
func (p *Pool) Pick(tokensNeeded int) *ModelLimiter {
	p.mu.Lock()
	defer p.mu.Unlock()

	// round-robin check
	n := len(p.limiters)
	for i := range n {
		idx := (p.next + i) % n
		ml := p.limiters[idx]
		if ml.ReqLimiter.Allow() && tokensNeeded < ml.LLM.ContextLength && float64(tokensNeeded) <= ml.TokenLimiter.Tokens() {
			p.next = (idx + 1) % n
			return ml
		}
	}

	// fallback to highest quality
	best := p.limiters[0]
	for _, ml := range p.limiters {
		if ml.LLM.Quality > best.LLM.Quality {
			best = ml
		}
	}
	return best
}

// Do executes a request using an available ModelLimiter,
// blocking until both a request token and the needed tokens are reserved.
// Returns api.Response or error (including context.DeadlineExceeded).
func (p *Pool) Do(ctx context.Context, req *api.Request) (*api.Response, error) {
	// apply optional default timeout
	if p.defaultTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, p.defaultTimeout)
		defer cancel()
	}
	// select model
	ml := p.Pick(req.TokensNeeded)
	if ml == nil {
		log.Fatal().Msg("No model selected by pool.Pick()")
	}

	log.Debug().Str("Selected model", ml.LLM.String()).Int("Tokens", req.TokensNeeded).Msg("Dispatching request")

	// reserve one request slot
	if err := ml.ReqLimiter.WaitN(ctx, 1); err != nil {
		return nil, err
	}
	// reserve token budget
	if err := ml.TokenLimiter.WaitN(ctx, req.TokensNeeded); err != nil {
		return nil, err
	}
	// execute the call
	return ml.LLM.Client.POSTChatCompletion(ctx, req, ml.LLM.Model)
}
