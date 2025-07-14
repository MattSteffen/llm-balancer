package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandleModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.LLMs)
}
