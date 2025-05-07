# LLM Load Balancer

The **LLM Load Balancer** is an open-source project designed to help developers and small-scale users efficiently manage and route requests to multiple Large Language Model (LLM) APIs. It intelligently balances load based on critical constraints like token limits, request rates, and context lengths. By acting as a proxy, this server provides seamless routing for API requests, allowing users to optimize usage across various providers, including free tiers during development and paid tiers in production, potentially balancing cost and quality based on configuration.
Important note, this assumes that any api used, will also be compatible with any api client provided. You might as well just use openai compatible clients.

---

## Key Features

- **Rate-Limit & Context Awareness:** Automatically balances load across configured APIs to prevent exceeding token and request rate limits and respects context window limitations.
- **Dynamic Forwarding:** Routes incoming HTTP requests to the most appropriate LLM API based on current availability, estimated token usage, and context length.
- **Request Queueing:** Holds requests in a queue when no suitable API is immediately available due to rate limits, releasing them as APIs become ready. This ensures requests are eventually processed without being immediately rejected due to temporary limits.
- **Configurable Settings:** Easily set up multiple APIs and general server parameters via a YAML configuration file, including potential future optimization preferences (like cost vs. quality).
- **Automatic Rate Limit Reset:** Tokens and request counters for each API are replenished periodically based on their defined limits.
- **API Abstraction (MVP focus: Google, Ollama, Groq text generation):** Handles specific API request/response formats for supported providers, abstracting away some differences for the client.

---

## Getting Started

### Prerequisites

- **Go 1.24 or higher**
- API keys for the LLM APIs you intend to use
- A YAML configuration file (example provided below)

### Installation

1. **Clone the repository**:

   ```bash
   git clone [https://github.com/mattsteffen/llm-load-balancer.git](https://github.com/mattsteffen/llm-load-balancer.git)
   cd llm-load-balancer
   ```

2. **Build the project**:

   ```bash
   go build -o llm-load-balancer
   ```

3. **Run the server**:

   ```bash
   ./llm-load-balancer
   ```

---

## Configuration

The configuration file (`config.yaml`) defines the server settings and API parameters. Below is an example:

```yaml
general:
  listen_address: "0.0.0.0"
  listen_port: 8080
  log_level: "debug" # debug, info, warn, error
  # Future: optimization_weights: { cost: 0.5, quality: 0.5, speed: 0.0 } # Example weights

llms:
  - name: "Google-Gemini"
    model: "gemini-1.5-pro" # Or appropriate model name
    base_url: "https://generativelanguage.googleapis.com/v1beta" # Base URL for this API type
    rate_limit: 1000000 # Tokens per minute
    requests_per_min: 60
    context_length: 128000
    api_key_name: "GOOGLE_API_KEY" # Environment variable name for the API key
    api_type: "google" # Identifier for the specific API implementation (e.g., google, ollama, groq)
    price: 0.007 # Example price per 1k input tokens (for future optimization)
    quality: 9 # Example quality score (1-10) (for future optimization)

```

### API Compatibility

This library assumes you use an openai client. It will recieve requests as an openai-compatible server and will respond in kind. Configured LLM Clients that do not meet those standards take the openai inputs, translate to their API, and translate the responses back to openai outputs. This is only really necessary if the provider does not implement the appropriate endpoints.

Currently we support only openai-compatible clients.
In progress:

- [x] Google API (/chat/completions works in simple cases)
- [ ] OpenRouter (should be quick)
- [x] Groq (Tested and works)
- [ ] Ollama (should be quick)

### Environment Variables

Set your API keys as environment variables corresponding to the `api_key_name` in your config file or strait into the config file. For example:

```bash
export GOOGLE_API_KEY=your_google_api_key_here
export GROQ_API_KEY=your_groq_api_key_here
# Ollama typically doesn't require an API key
```

---

## Usage

1. **Send a Request**
   Point your application's LLM calls to the load balancer endpoint (e.g., `http://localhost:8080`). Ensure your request body is in a format the load balancer understands (I implement all important features of the openai /chat/completions endpoint). The server will handle routing the request to the appropriate LLM API and will wait with the request if necessary until an API is available.

2. **Monitor Logs**
   Logs provide insights into:

   - Request reception and queueing
   - LLM selection and forwarding
   - Token usage (estimated for MVP requests)
   - API rate limit status
   - Request completion or errors

   To adjust the log verbosity, update the `log_level` in your config (`debug`, `info`, `warn`, `error`).

---

## Contributing

Contributions are welcome\! To get started:

1. Fork the repository.

2. Create a feature branch:

   ```bash
   git checkout -b feature-name
   ```

3. Commit your changes:

   ```bash
   git commit -m "Add new feature"
   ```

4. Push the branch:

   ```bash
   git push origin feature-name
   ```

5. Open a pull request.

### Ideas for Contribution (Roadmap)

- Implement accurate token counting using provider-specific libraries or endpoints.
- Refine and enhance load-balancing algorithms, including implementing the cost/quality/speed optimization function using configuration weights.
- Add support for more API providers and different request types (e.g., image generation, embeddings, file uploads).
- Add metrics and monitoring (e.g., Prometheus, Grafana).
- Build a UI for configuration and monitoring.
- Implement persistent storage for request logs and usage statistics.
- Introduce robust retry mechanisms for failed API requests (especially non-rate-limit errors).
- Improve error handling and user feedback when requests fail or time out in the queue.

---

## Roadmap (MVP and Beyond)

- [x] Basic server structure and configuration loading
- [x] Basic LLM representation with rate limit counters
- [x] Configure into packages and folders (`config`, `llm`, `api`, `balancer`)
- [x] Implement Request Queueing (single queue, waits for any ready LLM) *Done with go's time/rate package*
- [x] Implement MVP LLM Selection Logic (based on estimated tokens, requests, context)
- [ ] Implement API Integrations for MVP (Google, Ollama, Groq - text chat/completion)
  - [ ] Google REST API integration
  - [ ] Ollama API integration
  - [x] Groq API integration
  - [ ] OpenRouter integration
- [ ] Refine Request Handling (read body, estimate tokens (byte count MVP), modify body, forward, copy response)
- [x] Implement Automatic Rate Limit Refill
- [ ] Implement Unit tests (for core logic like selection, queueing)
- [ ] Implement Integration tests (HTTP handler, end-to-end flow through balancer)
- [ ] **Post-MVP:** Implement accurate token counting (using libraries/APIs)
- [ ] **Post-MVP:** Implement advanced load balancing (cost/quality/speed optimization)
- [ ] **Post-MVP:** Add support for more LLMs and request types (images, etc.)
- [ ] **Post-MVP:** Implement persistent storage for logs/stats
- [ ] **Post-MVP:** Introduce retry mechanisms
- [ ] **Post-MVP:** Add metrics and monitoring
- [ ] **Post-MVP:** Develop a UI

---

## License

Do what you want.

---

## Community and Support

- Leave a comment or contact me.

---

### Let's Build Together\! 🚀

If this project resonates with you, consider giving it a star ⭐ and sharing it with your community. Together, we can make this tool even better\!
