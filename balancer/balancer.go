package balancer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"llm-balancer/api"
	"llm-balancer/config"
	"llm-balancer/llm"
	"llm-balancer/queue"
)

var (
	ErrNoAPIClient     = errors.New("no API client available")
	ErrNoLLMsAvailable = errors.New("no LLM available for request")
)

const BytesPerToken = 16

type Balancer struct {
	llms       map[int]*llm.LLM // Index -> LLM instance
	queue      *queue.RequestQueue
	apiClients map[string]api.Client // Map API type string to Client instance
	// Need a way to signal the balancer/workers when LLMs become ready
	// or just poll the queue/check LLM status before dequeueing.
	// Let's start with polling/checking status.
	refillInterval time.Duration // e.g., 1 * time.Minute
	mu             sync.Mutex
}

// NewBalancer creates and initializes the balancer.
func NewBalancer(cfg *config.Config) (*Balancer, error) {
	// Initialize LLMs from config.LLMAPIs
	llmInstances := make(map[int]*llm.LLM)
	apiClients := make(map[string]api.Client)

	for _, llmCfg := range cfg.LLMAPIs {
		// Initialize API clients based on type, ensure only one client per type if stateless
		if _, exists := apiClients[llmCfg.Provider]; !exists {
			switch llmCfg.Provider {
			case "google":
				apiClients[llmCfg.Provider] = api.NewGoogleClient(llmCfg.BaseURL) // BaseURL might need adjustment per API type
			// case "ollama":
			// 	apiClients[llmCfg.APIType] = api.NewOllamaClient(llmCfg.BaseURL)
			// case "groq":
			// 	apiClients[llmCfg.APIType] = api.NewGroqClient(llmCfg.BaseURL)
			default:
				// Log warning or error for unsupported API type
			}
		}
	}

	b := &Balancer{
		llms:           llmInstances,
		queue:          queue.NewRequestQueue(),
		refillInterval: 1 * time.Minute, // Hardcoded for MVP, could be config
	}

	go b.startRefiller() // Start the background refill process
	// TODO: Potentially start workers to process the queue?
	// Or process queue in the main handler loop? Let's start with processing in handler.

	return b, nil
}

// startRefiller runs in a goroutine to periodically refill LLM rate limits.
func (b *Balancer) startRefiller() {
	ticker := time.NewTicker(b.refillInterval)
	defer ticker.Stop()

	for range ticker.C {
		for _, llm := range b.llms {
			llm.RefillCounters()
		}
	}
}

// processQueue attempts to handle requests waiting in the queue.
// This needs careful synchronization if called from multiple places (refiller, main handler).
// For MVP, maybe only call from refiller or after a successful request?
func (b *Balancer) processQueue() {
	// Dequeue requests and try to forward them if an LLM is available.
	// This function needs to be non-blocking or use a worker pool.
	// Let's refine this structure later - for MVP, maybe the main handler
	// checks the queue before exiting if no LLM was found initially.
}

// HandleRequest is the main HTTP handler for the load balancer.
func (b *Balancer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 1. Read request body and estimate tokens (byte count)
	// 2. Attempt to select an available LLM using SelectLLM
	// 3. If LLM found:
	//    - Decrement LLM counters
	//    - Forward request using ForwardRequestToAPI
	//    - Copy response back to client
	//    - Log success
	// 4. If no LLM found:
	//    - Enqueue the request
	//    - Log that request is queued
	//    - (MVP Simplicity): Maybe immediately try to process the queue once? Or rely on refiller?
	//      Let's just enqueue and return HTTP 200 OK with a message indicating it's queued,
	//      or maybe 202 Accepted. The actual response will be sent later by a worker (Post-MVP)
	//      or upon refilling (MVP simple approach, but client won't get real response easily).
	//      Okay, new plan for MVP: if no LLM is ready, the handler waits briefly or fails?
	//      User wants it to wait. This implies the handler goroutine itself blocks or
	//      it hands off to something that blocks and writes the response later.
	//      Sending the response back via a channel in QueuedRequest is the right pattern.
	//      So, the handler will enqueue and wait on the response channel.

	log.Info().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Msg("Received request")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading request body")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	estimatedTokens := int(math.Ceil(float64(len(bodyBytes)) / BytesPerToken)) // BytesPerToken const from main

	// Channel to receive the response asynchronously
	responseChan := make(chan queue.QueuedResponse, 1)
	queuedReq := &queue.QueuedRequest{
		OriginalRequestBytes: bodyBytes,
		OriginalRequest:      r,
		ResponseChan:         responseChan,
		EstimatedTokens:      estimatedTokens,
	}

	// Attempt to process immediately or enqueue
	if b.tryProcessRequest(queuedReq) {
		log.Debug().Msg("Request processed immediately")
	} else {
		log.Info().Msg("Request enqueued")
		b.queue.Enqueue(queuedReq)
		// Don't return here, wait for the response on the channel
		// Maybe return 202 Accepted immediately and response comes via webhook?
		// No, user wants the direct response. So, block and wait.
	}

	// Wait for the response (either immediate or from queue processing)
	resp := <-responseChan

	if resp.Error != nil {
		log.Error().Err(resp.Error).Msg("Error processing request after selection/queue")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError) // Or more specific error
		return
	}

	// Copy response headers and body back to the original client
	for header, values := range resp.Response.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	w.WriteHeader(resp.Response.StatusCode)
	_, err = io.Copy(w, resp.Response.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error copying response body")
		// Note: Headers and status code already sent, difficult to recover here.
	}
	resp.Response.Body.Close() // Close the upstream response body

	log.Info().Msgf("Request completed. Estimated tokens: %d", estimatedTokens)
}

// tryProcessRequest attempts to select an LLM and process a single request.
// Returns true if processed immediately, false if enqueued (should be done by caller).
func (b *Balancer) tryProcessRequest(req *queue.QueuedRequest) bool {
	b.mu.Lock() // Protect access to LLM states
	defer b.mu.Unlock()

	selectedLLM, err := b.selectLLM(req.EstimatedTokens)
	if err != nil {
		log.Debug().Err(err).Msg("No LLM available, cannot process immediately")
		// Signal that it couldn't be processed immediately
		return false
	}

	// Decrement counters immediately upon selection
	selectedLLM.Decrement(req.EstimatedTokens) // Assume 1 request per forward

	// Process the request (this part shouldn't hold the balancer's main mutex)
	// Need a way to run this concurrently or outside the lock.
	// A worker pool is ideal here. For MVP, let's call a helper function
	// that handles the API call and sending the response via the channel.
	go b.executeRequest(selectedLLM, req) // Execute concurrently

	return true // Successfully selected an LLM and passed for execution
}

// selectLLM finds an available LLM based on the criteria.
// Assumes caller holds the balancer's mutex.
func (b *Balancer) selectLLM(numTokens int) (*llm.LLM, error) {
	var selected *llm.LLM

	// Simple availability check for MVP
	for _, llm := range b.llms {
		// Check LLM's internal state safely if selectLLM doesn't hold the LLM's mutex
		// Currently selectLLM holds the *balancer* mutex, which should be enough
		// if LLM counter methods use their own mutex. Yes, they do.
		llm.Mu.Lock() // Need to lock LLM to check counters safely
		isAvailable := llm.IsAvailable(numTokens)
		llm.Mu.Unlock()

		if isAvailable {
			selected = llm // Pick the first one found
			break          // For MVP, simple first fit is okay
		}
	}

	if selected == nil {
		return nil, ErrNoLLMsAvailable
	}

	return selected, nil
}

// executeRequest handles the API call and sends the result back via the channel.
// This takes the exact same HTTP request that was originally provided and forwards it to the LLM.
// Runs in a goroutine.
func (b *Balancer) executeRequest(llm *llm.LLM, req *queue.QueuedRequest) {
	// Get the specific API client instance
	apiClient, ok := b.apiClients[llm.Provider]
	if !ok {
		err := fmt.Errorf("unsupported API type: %s", llm.Provider)
		req.ResponseChan <- queue.QueuedResponse{Response: nil, Error: err}
		return
	}

	// Modify the request URL to use the LLM's base URL
	apiReq, err := http.NewRequest(http.MethodPost, llm.BaseURL, io.NopCloser(bytes.NewReader(req.OriginalRequestBytes)))
	if err != nil {
		err = fmt.Errorf("failed to create API request for %s: %w", llm.Model, err)
		req.ResponseChan <- queue.QueuedResponse{Response: nil, Error: err}
		return
	}

	// Copy headers from the original request
	apiReq.Header = http.Header{}
	for key, values := range req.OriginalRequest.Header {
		for _, value := range values {
			apiReq.Header.Add(key, value)
		}
	}

	// Send the request
	log.Info().Str("llm_model", llm.Model).Str("url", apiReq.URL.String()).Msg("Forwarding request to LLM")
	resp, err := apiClient.SendRequest(apiReq)
	if err != nil {
		err = fmt.Errorf("failed to send request to %s: %w", llm.Model, err)
		req.ResponseChan <- queue.QueuedResponse{Response: nil, Error: err}
		return
	}

	req.ResponseChan <- queue.QueuedResponse{Response: resp, Error: nil}
}

// TODO: Need a background process or trigger to check the queue
// when an LLM becomes available (e.g., after refill).
// A simple approach for MVP is that the refiller calls a function
// that attempts to process a few items from the queue.
// This needs careful synchronization with the main handler.
