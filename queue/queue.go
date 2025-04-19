package queue

import (
	"net/http"
	"sync"
)

type RequestQueue struct {
	mu    sync.Mutex
	items []*QueuedRequest
	// Notification mechanism for when LLMs become available?
	// Or the balancer just checks the queue periodically/when LLMs refill?
	// Let's start simple: balancer checks the queue when an LLM refills.
}

func NewRequestQueue() *RequestQueue {
	return &RequestQueue{}
}

func (q *RequestQueue) Enqueue(req *QueuedRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, req)
}

func (q *RequestQueue) Dequeue() *QueuedRequest {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil
	}
	req := q.items[0]
	q.items = q.items[1:]
	return req
}

type QueuedRequest struct {
	OriginalRequestBytes []byte
	ResponseChan         chan<- QueuedResponse
	EstimatedTokens      int
}

type QueuedResponse struct {
	Response *http.Response
	Error    error
}
