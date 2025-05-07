package handlers

import (
	"llm-balancer/balancer"

	"github.com/rs/zerolog/log"
)

const (
	BytesPerToken = 4 // Average bytes per token for OpenAI models
)

type Handler struct {
	Pool *balancer.Pool
}

func NewHandler(pool *balancer.Pool) *Handler {
	log.Debug().Msg("Creating new handler")
	return &Handler{
		Pool: pool,
	}
}
