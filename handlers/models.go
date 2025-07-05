package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandleModels(w http.ResponseWriter, r *http.Request) {
	response := []map[string]string{
		{
			"id":               "balancer",
			"name":             "llm-balancer",
			"object":           "model",
			"created":          "1700000000",
			"owned_by":         "openai",
			"permission":       "all",
			"root":             "gpt-4",
			"parent":           "",
			"capabilities":     "chat, completion, edit",
			"max_tokens":       "8192",
			"max_input_tokens": "4096",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
