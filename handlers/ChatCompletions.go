package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"llm-balancer/api"
	"llm-balancer/balancer"
	"llm-balancer/openai"
	"net/http"
	"slices"
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

	fmt.Printf("Received chat completion request messages: %s\n", string(bodyBytes))

	tokensNeeded, err := countTokens(string(bodyBytes))
	if err != nil {
		tokensNeeded = int(1.1 * float64(len(bodyBytes)) / BytesPerToken)
	}

	apiReq := &api.Request{
		Request:      &reqBody,
		TokensNeeded: tokensNeeded,
	}

	ctx := r.Context()

	// Route to the correct model
	var ml *balancer.ModelLimiter
	model := reqBody.Model
	if slices.Contains(h.Pool.Models, model) {
		ml = h.Pool.Assign(apiReq)
	} else if group, ok := h.Pool.Groups[model]; ok {
		ml = h.Pool.PickGroup(apiReq.TokensNeeded, group)
	} else {
		ml = h.Pool.PickAny(apiReq.TokensNeeded)
	}

	resp, err := h.Pool.DoAssigned(ctx, ml, apiReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("balancer Do failed: %v", err), http.StatusInternalServerError)
		return
	}
	if resp.Error != nil {
		http.Error(w, fmt.Sprintf("API request failed: %v", resp.Error), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp.Response); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
