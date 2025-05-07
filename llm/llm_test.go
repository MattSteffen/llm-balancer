package llm_test

// import (
// 	"llm-balancer/config"
// 	"llm-balancer/llm"

// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"
// )

// var _ = Describe("LLM", func() {
// 	var llmConfig config.LLMApiConfig

// 	BeforeEach(func() {
// 		llmConfig = config.LLMApiConfig{
// 			Provider:      "openai",
// 			Model:         "gpt-4",
// 			BaseURL:       "https://api.openai.com",
// 			RequestLimit:  60,
// 			TokenLimit:    1000,
// 			ContextLength: 2048,
// 			Price:         0.03,
// 			Quality:       5,
// 		}
// 	})

// 	Describe("NewLLM", func() {
// 		It("should create a new LLM instance", func() {
// 			llmInstance, err := llm.NewLLM(llmConfig)
// 			Expect(err).NotTo(HaveOccurred())
// 			Expect(llmInstance).NotTo(BeNil())
// 			Expect(llmInstance.Provider).To(Equal("openai"))
// 			Expect(llmInstance.Model).To(Equal("gpt-4"))
// 			Expect(llmInstance.TokensLeft).To(Equal(2048))
// 			Expect(llmInstance.RequestsLeft).To(Equal(60))
// 		})
// 	})

// 	Describe("IsAvailable", func() {
// 		var llmInstance *llm.LLM

// 		BeforeEach(func() {
// 			var err error
// 			llmInstance, err = llm.NewLLM(llmConfig)
// 			Expect(err).NotTo(HaveOccurred())
// 		})

// 		It("should return true when enough tokens and requests are available", func() {
// 			Expect(llmInstance.IsAvailable(100)).To(BeTrue())
// 		})

// 		It("should return false when not enough tokens are available", func() {
// 			llmInstance.TokensLeft = 50
// 			Expect(llmInstance.IsAvailable(100)).To(BeFalse())
// 		})
// 	})

// 	Describe("Decrement", func() {
// 		var llmInstance *llm.LLM

// 		BeforeEach(func() {
// 			var err error
// 			llmInstance, err = llm.NewLLM(llmConfig)
// 			Expect(err).NotTo(HaveOccurred())
// 		})

// 		It("should decrease tokens and requests correctly", func() {
// 			err := llmInstance.Decrement(100)
// 			Expect(err).NotTo(HaveOccurred())
// 			Expect(llmInstance.TokensLeft).To(Equal(1948))
// 			Expect(llmInstance.RequestsLeft).To(Equal(59))
// 		})

// 		It("should return error when not enough tokens", func() {
// 			llmInstance.TokensLeft = 50
// 			err := llmInstance.Decrement(100)
// 			Expect(err).To(HaveOccurred())
// 		})
// 	})

// 	Describe("RefillCounters", func() {
// 		var llmInstance *llm.LLM

// 		BeforeEach(func() {
// 			var err error
// 			llmInstance, err = llm.NewLLM(llmConfig)
// 			Expect(err).NotTo(HaveOccurred())
// 			llmInstance.TokensLeft = 100
// 			llmInstance.RequestsLeft = 5
// 		})

// 		It("should refill counters to initial values", func() {
// 			llmInstance.RefillCounters()
// 			Expect(llmInstance.TokensLeft).To(Equal(2048))
// 			Expect(llmInstance.RequestsLeft).To(Equal(60))
// 		})
// 	})
// })
