package config_test

import (
	"os"

	"llm-balancer/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var tempFile *os.File

	BeforeEach(func() {
		var err error
		tempFile, err = os.CreateTemp("", "config-*.yaml")
		Expect(err).NotTo(HaveOccurred())

		yamlContent := `
    general:
      listen_address: "127.0.0.1"
      listen_port: 8080
      log_level: "debug"
    llms:
      - provider: "openai"
        model: "gpt-4"
        base_url: "https://api.openai.com"
        modes: ["text"]
        rate_limit: 60
        token_limit: 1000
        context_length: 2048
        api_key_name: "OPENAI_API_KEY"
        cost_input: 0.01
        cost_output: 0.02
        price: 0.03
        quality: 5
    `
		_, err = tempFile.Write([]byte(yamlContent))
		Expect(err).NotTo(HaveOccurred())
		tempFile.Close()
	})

	AfterEach(func() {
		os.Remove(tempFile.Name())
	})

	It("should load and parse the configuration correctly", func() {
		cfg, err := config.LoadConfig(tempFile.Name())
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).NotTo(BeNil())

		Expect(cfg.General.ListenAddress).To(Equal("127.0.0.1"))
		Expect(cfg.General.ListenPort).To(Equal(8080))
		Expect(cfg.General.LogLevel).To(Equal("debug"))

		Expect(cfg.LLMAPIs).To(HaveLen(1))
		llm := cfg.LLMAPIs[0]
		Expect(llm.Provider).To(Equal("openai"))
		Expect(llm.Model).To(Equal("gpt-4"))
		Expect(llm.BaseURL).To(Equal("https://api.openai.com"))
		Expect(llm.Modes).To(ConsistOf("text"))
		Expect(llm.RequestsPerMin).To(Equal(60))
		Expect(llm.TokensPerMin).To(Equal(1000))
		Expect(llm.ContextLength).To(Equal(2048))
		Expect(llm.CostInput).To(Equal(0.01))
		Expect(llm.CostOutput).To(Equal(0.02))
		Expect(llm.CostInput).To(Equal(0.03))
		Expect(llm.Quality).To(Equal(5))
	})
})
