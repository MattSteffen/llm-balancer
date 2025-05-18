package api_test

import (
	"context"
	"llm-balancer/api"
	"llm-balancer/openai"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

const (
	defaultResponseMessage = "This is a test response"
)

var _ = Describe("OpenAIClient", func() {
	var (
		server    *ghttp.Server
		client    *api.OpenAIClient
		ctx       context.Context
		apiKey    string
		testModel string
		mockResp  openai.ChatCompletionResponse
	)

	BeforeEach(func() {
		// Set up a new test server
		server = ghttp.NewServer()
		responseMessage := defaultResponseMessage

		ctx = context.Background()
		apiKey = "test-api-key"
		testModel = "gpt-4"

		// Set up a default mock response
		mockResp = openai.ChatCompletionResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   testModel,
			Choices: []openai.Choice{
				{
					Message: openai.CompletionMessage{
						Role:    "assistant",
						Content: &responseMessage,
					},
					FinishReason: "stop",
					Index:        0,
				},
			},
		}

		// Create the client with the test server URL
		client = api.NewOpenAIClient(server.URL(), apiKey)
	})

	AfterEach(func() {
		// Clean up the server
		server.Close()
	})

	Describe("NewOpenAIClient", func() {
		It("should create a new client with the correct base URL and API key", func() {
			baseURL := "https://api.openai.com/v1"
			apiKey := "test-key"

			client := api.NewOpenAIClient(baseURL, apiKey)
			Expect(client).NotTo(BeNil())

			// Note: We can't directly access BaseURL and APIKey since they're private,
			// but we can test the client behavior indirectly in other tests
		})
	})

	Describe("POSTChatCompletion", func() {
		var request *api.Request

		BeforeEach(func() {
			request = &api.Request{
				Request: &openai.ChatCompletionRequest{
					Model: "", // Will be set by the client
					Messages: []openai.Message{
						{
							Role:    "user",
							Content: "Hello, world!",
						},
					},
					ResponseFormat: &openai.ResponseFormat{
						Type: "text",
					},
				},
			}
		})

		Context("when the request is successful", func() {
			BeforeEach(func() {
				// Configure the mock server to expect a specific request and return a successful response
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/chat/completions"),
						ghttp.VerifyHeaderKV("Content-Type", "application/json"),
						ghttp.VerifyHeaderKV("Authorization", "Bearer "+apiKey),
						ghttp.VerifyJSONRepresenting(func(req *openai.ChatCompletionRequest) bool {
							return req.Model == testModel && len(req.Messages) == 1 && req.Messages[0].Role == "user"
						}),
						ghttp.RespondWithJSONEncoded(http.StatusOK, mockResp),
					),
				)
			})

			It("should return a valid response", func() {
				response, err := client.POSTChatCompletion(ctx, request, testModel)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Response).NotTo(BeNil())
				Expect(response.Response.ID).To(Equal(mockResp.ID))
				Expect(response.Response.Model).To(Equal(testModel))
				Expect(response.Response.Choices).To(HaveLen(1))
				Expect(response.Response.Choices[0].Message.Content).To(Equal("This is a test response"))

				// Verify the server received exactly one request
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the request returns a non-200 status code", func() {
			BeforeEach(func() {
				// Configure the mock server to return a rate limit error
				errorResp := map[string]interface{}{
					"error": map[string]interface{}{
						"message": "Rate limit exceeded",
						"type":    "rate_limit_error",
					},
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/chat/completions"),
						ghttp.RespondWithJSONEncoded(http.StatusTooManyRequests, errorResp),
					),
				)
			})

			It("should return an error", func() {
				response, err := client.POSTChatCompletion(ctx, request, testModel)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code 429"))
				Expect(response).To(BeNil())

				// Verify the server received exactly one request
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the response body cannot be unmarshaled", func() {
			BeforeEach(func() {
				// Configure the mock server to return invalid JSON
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/chat/completions"),
						ghttp.RespondWith(http.StatusOK, `{invalid json}`),
					),
				)
			})

			It("should return an error", func() {
				response, err := client.POSTChatCompletion(ctx, request, testModel)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error unmarshaling response"))
				Expect(response).To(BeNil())

				// Verify the server received exactly one request
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when using different response formats", func() {
			BeforeEach(func() {
				// Update the request with JSON response format
				request.Request.ResponseFormat = &openai.ResponseFormat{
					Type: "json_object",
				}

				// Update the mock response with JSON content
				jsonContent := `{"result": "test result"}`
				mockResp.Choices[0].Message.Content = &jsonContent

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/chat/completions"),
						ghttp.VerifyJSONRepresenting(func(req *openai.ChatCompletionRequest) bool {
							return req.ResponseFormat != nil && req.ResponseFormat.Type == "json_object"
						}),
						ghttp.RespondWithJSONEncoded(http.StatusOK, mockResp),
					),
				)
			})

			It("should handle the JSON response format correctly", func() {
				response, err := client.POSTChatCompletion(ctx, request, testModel)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Response.Choices[0].Message.Content).To(Equal(`{"result": "test result"}`))

				// Verify the server received exactly one request
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		Context("when the server returns an error", func() {
			BeforeEach(func() {
				// Close the server to force a connection error
				server.Close()
			})

			It("should return an error", func() {
				response, err := client.POSTChatCompletion(ctx, request, testModel)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
			})
		})

		Context("when the context is canceled", func() {
			var cancelFunc context.CancelFunc

			BeforeEach(func() {
				// Create a cancellable context
				ctx, cancelFunc = context.WithCancel(ctx)

				// Set up server to delay response to test context cancellation
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/chat/completions"),
						func(w http.ResponseWriter, r *http.Request) {
							// This handler will never respond, allowing the context cancellation to take effect
							<-r.Context().Done() // Wait for context cancellation
						},
					),
				)
			})

			It("should return an error when context is canceled", func() {
				// Start a goroutine to perform the request
				errChan := make(chan error)
				respChan := make(chan *api.Response)

				go func() {
					resp, err := client.POSTChatCompletion(ctx, request, testModel)
					errChan <- err
					respChan <- resp
				}()

				// Cancel the context after starting the request
				cancelFunc()

				// Check the results
				err := <-errChan
				resp := <-respChan

				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})
		})
	})
})
