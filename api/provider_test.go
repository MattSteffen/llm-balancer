package api_test

// import (
// 	"context"
// 	"llm-balancer/api"
// 	"llm-balancer/openai"

// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"
// )

// var _ = Describe("Provider Integration Tests", func() {
// 	var (
// 		ctx       context.Context
// 		providers map[string]*providerConfig
// 	)

// 	type providerConfig struct {
// 		baseURL string
// 		apiKey  string
// 		model   string
// 	}

// 	BeforeEach(func() {
// 		ctx = context.Background()

// 		// Configure providers for testing
// 		providers = map[string]*providerConfig{
// 			"google": {
// 				baseURL: "https://generativelanguage.googleapis.com/v1beta/openai",
// 				apiKey:  "YOUR_GOOGLE_API_KEY",
// 				model:   "gemini-2.5-pro-preview-05-06",
// 			},
// 			"groq": {
// 				baseURL: "https://api.groq.com/openai/v1",
// 				apiKey:  "YOUR_GROQ_API_KEY",
// 				model:   "llama-3.3-70b-versatile",
// 			},
// 			"openrouter": {
// 				baseURL: "https://openrouter.ai/api/v1",
// 				apiKey:  "YOUR_OPENROUTER_API_KEY",
// 				model:   "google/gemini-2.5-pro-exp-03-25",
// 			},
// 		}
// 	})

// 	for providerName, config := range providers {
// 		// Use closure to capture provider config
// 		providerName, config := providerName, config

// 		Describe(providerName+" Provider", func() {
// 			var client *api.OpenAIClient

// 			BeforeEach(func() {
// 				client = api.NewOpenAIClient(config.baseURL, config.apiKey)
// 			})

// 			Context("Basic Chat Completion", func() {
// 				It("should return a valid chat completion response", func() {
// 					request := &api.Request{
// 						Request: &openai.ChatCompletionRequest{
// 							Model: config.model,
// 							Messages: []openai.ChatCompletionMessage{
// 								{
// 									Role:    "user",
// 									Content: "What is the capital of France?",
// 								},
// 							},
// 						},
// 					}

// 					response, err := client.POSTChatCompletion(ctx, request, config.model)
// 					Expect(err).NotTo(HaveOccurred())
// 					Expect(response).NotTo(BeNil())
// 					Expect(response.Response).NotTo(BeNil())
// 					Expect(response.Response.Choices).To(HaveLen(1))
// 					Expect(response.Response.Choices[0].Message.Content).NotTo(BeEmpty())
// 				})
// 			})

// 			Context("JSON Response Format", func() {
// 				It("should return a valid JSON-formatted response", func() {
// 					request := &api.Request{
// 						Request: &openai.ChatCompletionRequest{
// 							Model: config.model,
// 							Messages: []openai.ChatCompletionMessage{
// 								{
// 									Role:    "user",
// 									Content: "How do I solve 8x + 7 = -23?",
// 								},
// 							},
// 							ResponseFormat: &openai.ResponseFormat{
// 								Type: "json_object",
// 							},
// 						},
// 					}

// 					response, err := client.POSTChatCompletion(ctx, request, config.model)
// 					Expect(err).NotTo(HaveOccurred())
// 					Expect(response).NotTo(BeNil())
// 					Expect(response.Response).NotTo(BeNil())
// 					Expect(response.Response.Choices).To(HaveLen(1))
// 					Expect(response.Response.Choices[0].Message.Content).To(ContainSubstring("{"))
// 					Expect(response.Response.Choices[0].Message.Content).To(ContainSubstring("}"))
// 				})
// 			})

// 			Context("Function Calling", func() {
// 				It("should handle function calls correctly", func() {
// 					request := &api.Request{
// 						Request: &openai.ChatCompletionRequest{
// 							Model: config.model,
// 							Messages: []openai.ChatCompletionMessage{
// 								{
// 									Role:    "user",
// 									Content: "What is 23 plus 45?",
// 								},
// 							},
// 							Tools: []openai.Tool{
// 								{
// 									Type: "function",
// 									Function: &openai.FunctionDefinition{
// 										Name:        "calculate",
// 										Description: "Calculate a mathematical expression",
// 										Parameters: map[string]interface{}{
// 											"type": "object",
// 											"properties": map[string]interface{}{
// 												"expression": map[string]interface{}{
// 													"type":        "string",
// 													"description": "The mathematical expression to calculate",
// 												},
// 											},
// 											"required": []string{"expression"},
// 										},
// 									},
// 								},
// 							},
// 							ToolChoice: "auto",
// 						},
// 					}

// 					response, err := client.POSTChatCompletion(ctx, request, config.model)
// 					Expect(err).NotTo(HaveOccurred())
// 					Expect(response).NotTo(BeNil())
// 					Expect(response.Response).NotTo(BeNil())
// 					Expect(response.Response.Choices).To(HaveLen(1))

// 					// Check for either tool calls or direct answer
// 					choice := response.Response.Choices[0]
// 					if choice.Message.ToolCalls != nil {
// 						Expect(choice.Message.ToolCalls).To(HaveLen(1))
// 						Expect(choice.Message.ToolCalls[0].Function.Name).To(Equal("calculate"))
// 					} else {
// 						Expect(choice.Message.Content).To(ContainSubstring("68"))
// 					}
// 				})
// 			})
// 		})
// 	}
// })
