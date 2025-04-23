package mock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

type MockOpenAIServer struct {
	Server *httptest.Server
}

func NewMockOpenAIServer() *MockOpenAIServer {
	mock := &MockOpenAIServer{}
	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"id":      "mock-response",
			"object":  "chat.completion",
			"created": 1234567890,
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "This is a mock response",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 10,
				"total_tokens":      20,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	return mock
}

func (m *MockOpenAIServer) Close() {
	m.Server.Close()
}

func (m *MockOpenAIServer) URL() string {
	return m.Server.URL
}
