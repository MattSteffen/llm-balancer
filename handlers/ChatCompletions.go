package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"llm-balancer/api"
	"llm-balancer/openai"
	"net/http"
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

	apiReq := &api.Request{
		Request:      &reqBody,
		TokensNeeded: int(1.1 * float64(len(bodyBytes)) / BytesPerToken),
	}

	ctx := r.Context()
	resp, err := h.Pool.Do(ctx, apiReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("balancer Do failed: %v", err), http.StatusInternalServerError)
		return
	}
	if resp.Error != nil {
		http.Error(w, fmt.Sprintf("API request failed: %v", resp.Error), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Response)
}
