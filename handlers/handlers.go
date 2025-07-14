package handlers

import (
	"llm-balancer/balancer"
	"llm-balancer/llm"
)

const (
	BytesPerToken = 4 // Average bytes per token for OpenAI models
)

type Handler struct {
	Pool *balancer.Pool
	LLMs []*llm.LLM
}

func NewHandler(pool *balancer.Pool, llms []*llm.LLM) *Handler {
	return &Handler{
		Pool: pool,
		LLMs: llms,
	}
}
