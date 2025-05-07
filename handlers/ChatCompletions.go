package handlers

import (
	"encoding/json"
	"io"
	"llm-balancer/api"
	"llm-balancer/openai"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (h *Handler) HandleChatCompletion(w http.ResponseWriter, r *http.Request) {
	var reqBody openai.ChatCompletionRequest
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	log.Debug().Msgf("Received request: %s", string(bodyBytes))

	// 2) Build your API request (your api.Request type)
	apiReq := &api.Request{
		Request:      &reqBody,
		TokensNeeded: int(1.1 * float64(len(bodyBytes)) / BytesPerToken),
	}

	// 3) Dispatch via balancer pool
	ctx := r.Context() // TODO: Why is the context coming from the request?

	resp, err := h.Pool.Do(ctx, apiReq)
	if err != nil {
		log.Err(err).Msg("balancer Do failed")
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}
	if resp.Error != nil {
		log.Err(resp.Error).Msg("API request failed")
		http.Error(w, resp.Error.Error(), http.StatusInternalServerError)
		return
	}

	// 4) Send the response back to the client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Response)
}
