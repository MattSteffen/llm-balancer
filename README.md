# llm-balancer

A load balancer for multiple LLM apis allowing development to exceed rate limits.

#### Overview

This design outlines the architecture for a load balancer that distributes requests to Large Language Models (LLMs) based on various metrics, starting with a single OpenAI-compatible API for the MVP. The system is designed to be scalable and modular, allowing for future enhancements.

#### Components

1. **Configuration Management**

   - **Source**: YAML configuration file and environment variables for API keys.
   - **Structures**:

     ```go
     type Config struct {
         General  GeneralConfig  `yaml:"general"`
         LLMAPIs []LLMApiConfig `yaml:"LLM_APIs"`
     }

     type GeneralConfig struct {
         ListenAddress string `yaml:"listen_address"`
         LogLevel      string `yaml:"log_level"`
     }

     type LLMApiConfig struct {
         Name          string `yaml:"name"`
         URL           string `yaml:"url"`
         RateLimit     int    `yaml:"rate_limit"`
         ContextLength int    `yaml:"context_length"`
         ApiKeyName    string `yaml:"api_key_name"`
     }
     ```

2. **Logging**

   - **Library**: `zerolog` for structured logging.
   - **Setup**:

     ```go
     import "github.com/rs/zerolog/log"

     func init() {
         level, err := zerolog.ParseLevel(config.General.LogLevel)
         if err != nil {
             log.Error().Err(err).Msg("Invalid log level, using INFO")
             level = zerolog.InfoLevel
         }
         zerolog.SetGlobalLevel(level)
     }
     ```

3. **HTTP Server**

   - **Handler**: Handles incoming requests and forwards them to the appropriate LLM API.
   - **Forwarding Logic**:

     ```go
     func llmHandler(w http.ResponseWriter, r *http.Request) {
         // Load balancing logic to select LLM
         selectedLLM := selectLLM()
         if selectedLLM == nil {
             http.Error(w, "No LLM available", http.StatusServiceUnavailable)
             return
         }

         // Call LLM API
         response, err := callLLM(selectedLLM, r)
         if err != nil {
             log.Error().Err(err).Msg("Failed to call LLM API")
             http.Error(w, "Internal server error", http.StatusInternalServerError)
             return
         }

         // Write response back to client
         w.WriteHeader(response.StatusCode)
         _, err = io.Copy(w, response.Body)
         if err != nil {
             log.Error().Err(err).Msg("Failed to copy response body")
         }
     }
     ```

4. **Metrics Collection**

   - **Library**: Prometheus for metrics collection and exposure.
   - **Metrics**:
     ```go
     var (
         RequestDuration = prometheus.NewHistogramVec(
             prometheus.HistogramOpts{
                 Name:    "llm_request_duration_seconds",
                 Help:    "Duration of requests to the LLM API",
                 Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
             },
             []string{"llm"},
         )
         TotalTokensUsed = prometheus.NewCounterVec(
             prometheus.CounterOpts{
                 Name: "llm_total_tokens_used",
                 Help: "Total tokens used in API requests",
             },
             []string{"llm"},
         )
         // Additional metrics as needed
     )
     ```

5. **LLM Management**

   - **Struct**:
     ```go
     type LLM struct {
         Name          string
         URL           string
         RateLimit     int
         ContextLength int
         ApiKeyName    string
         TokensLeft    int
         mu            sync.Mutex
     }
     ```
   - **Token Refill**:
     ```go
     func refillTokens(llms []*LLM) {
         ticker := time.NewTicker(time.Minute)
         for _ = range ticker.C {
             for _, llm := range llms {
                 llm.mu.Lock()
                 llm.TokensLeft += llm.RateLimit
                 llm.mu.Unlock()
             }
         }
     }
     ```

6. **Load Balancing**

   - **Logic**:
     ```go
     func selectLLM() *LLM {
         for _, llm := range llms {
             llm.mu.Lock()
             if llm.TokensLeft > requiredTokens {
                 llm.mu.Unlock()
                 return llm
             }
             llm.mu.Unlock()
         }
         return nil
     }
     ```

#### Potential Pitfalls and Considerations

- **Concurrent Access**: Use mutexes to prevent race conditions on `TokensLeft`.
- **Rate Limiting Accuracy**: Ensure precise token refilling using `time.Ticker`.
- **API Response Parsing**: Handle differences in response formats between LLMs.
- **Error Handling**: Implement retry mechanisms and circuit breakers for transient errors.
- **Scalability**: Design components to be modular and easily extendable.

#### Testing

- **Unit Tests**: For configuration parsing, metrics collection, and load balancing logic.
- **Integration Tests**: Simulate requests to ensure the entire flow works correctly.
- **Stress Tests**: Evaluate system performance under high load.

#### Documentation

- **Configuration Format**: Detailed explanation of the YAML configuration and environment variables.
- **Request/Response Formats**: Expected formats for HTTP requests and responses.
- **Environment Variables**: List of required environment variables for API keys.
