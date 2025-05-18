package handlers

import (
	"llm-balancer/balancer"
)

const (
	BytesPerToken = 4 // Average bytes per token for OpenAI models
)

type Handler struct {
	Pool *balancer.Pool
}

func NewHandler(pool *balancer.Pool) *Handler {
	return &Handler{
		Pool: pool,
	}
}
