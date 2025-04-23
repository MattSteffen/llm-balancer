package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"llm-balancer/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = Describe("Server Integration", func() {
	var (
		serverExitChan chan struct{}
		mockServer     *mock.MockOpenAIServer
	)

	BeforeEach(func() {
		// Initialize mock server first
		mockServer = mock.NewMockOpenAIServer()

		// Create test config using mock server URL
		testConfig := `
general:
  listen_address: "127.0.0.1"
  listen_port: 8081
  log_level: "debug"
llm_apis:
  - provider: "openai"
    model: "gpt-4"
    base_url: "` + mockServer.URL() + `"
    rate_limit: 60
    token_limit: 1000
    context_length: 2048
    api_key_name: "OPENAI_API_KEY"
    price: 0.03
    quality: 5
`
		err := os.WriteFile("test_config.yaml", []byte(testConfig), 0644)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("OPENAI_API_KEY", "test-api-key")

		serverExitChan = make(chan struct{})
		go func() {
			defer close(serverExitChan)
			os.Args = []string{"cmd", "-config", "test_config.yaml"}
			main()
		}()

		// Wait for server to start
		time.Sleep(2 * time.Second)
	})

	AfterEach(func() {
		mockServer.Close()
		os.Remove("test_config.yaml")
		os.Unsetenv("OPENAI_API_KEY")
	})

	It("should handle LLM requests", func() {
		requestBody := map[string]interface{}{
			"model": "gpt-4",
			"messages": []map[string]string{
				{
					"role":    "user",
					"content": "Hello, how are you?",
				},
			},
		}
		bodyBytes, err := json.Marshal(requestBody)
		Expect(err).NotTo(HaveOccurred())

		resp, err := http.Post("http://127.0.0.1:8081/v1/chat/completions",
			"application/json",
			bytes.NewBuffer(bodyBytes))
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		var response map[string]interface{}
		err = json.Unmarshal(body, &response)
		Expect(err).NotTo(HaveOccurred())

		Expect(response).To(HaveKey("choices"))
		choices, ok := response["choices"].([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(choices)).To(BeNumerically(">", 0))
	})
})
