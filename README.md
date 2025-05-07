# LLM Load Balancer

- [ ] Instead of implementing the api for each client, I'll just assume each client is openai compatible.
  - I don't want to have to implement the api for each client, because there is a lot of types required. Just doing chat completions is not enough, but structured outputs and tools.
  - Maybe I can translate though.
- [ ] I'll import the openai-go package and use the types defined there?

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
    base_url: "[https://generativelanguage.googleapis.com/v1beta](https://generativelanguage.googleapis.com/v1beta)" # Base URL for this API type
    rate_limit: 1000000 # Tokens per minute
    requests_per_min: 60
    context_length: 128000
    api_key_name: "GOOGLE_API_KEY" # Environment variable name for the API key
    api_type: "google" # Identifier for the specific API implementation (e.g., google, ollama, groq)
    price: 0.007 # Example price per 1k input tokens (for future optimization)
    quality: 9 # Example quality score (1-10) (for future optimization)

  - name: "Ollama-Local"
    model: "llama3" # Or appropriate model name
    base_url: "http://localhost:11434/api" # Base URL for Ollama
    rate_limit: 5000000 # Ollama might not have strict token limits, set based on local capacity
    requests_per_min: 500 # Set based on local capacity
    context_length: 8192 # Model specific context length
    api_key_name: "" # Ollama often doesn't use API keys
    api_type: "ollama"
    price: 0.0 # Local APIs are often free
    quality: 7 # Example quality score

  - name: "Groq-Llama3"
    model: "llama3-8b-8192" # Or appropriate model name
    base_url: "[https://api.groq.com/openai/v1](https://api.groq.com/openai/v1)" # Groq's OpenAI compatible base URL
    rate_limit: 270000000 # Tokens per minute (example value, check Groq docs)
    requests_per_min: 15000 # Requests per minute (example value, check Groq docs)
    context_length: 8192
    api_key_name: "GROQ_API_KEY"
    api_type: "groq"
    price: 0.0005 # Example price (check Groq docs)
    quality: 8 # Example quality score
```

### API Types

The `api_type` field in the `llms` configuration specifies which internal API integration the load balancer should use. For the MVP, supported types will include:

- `google`: For Google AI models via their REST API.
- `ollama`: For local Ollama instances.
- `groq`: For Groq cloud API.

Each `api_type` implementation is responsible for translating the incoming request format (which the load balancer aims to standardize internally) into the provider's specific API call and handling its response.

### Environment Variables

Set your API keys as environment variables corresponding to the `api_key_name` in your config file. For example:

```bash
export GOOGLE_API_KEY=your_google_api_key_here
export GROQ_API_KEY=your_groq_api_key_here
# Ollama typically doesn't require an API key
```

---

## Usage

1. **Send a Request**
   Point your application's LLM calls to the load balancer endpoint (e.g., `http://localhost:8080`). Ensure your request body is in a format the load balancer understands (initially targeting standard chat completion JSON structures). The server will handle routing the request to the appropriate LLM API and will queue the request if necessary until an API is available.

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
- [ ] Configure into packages and folders (`config`, `llm`, `queue`, `api`, `balancer`)
- [ ] Implement Request Queueing (single queue, waits for any ready LLM)
- [ ] Implement MVP LLM Selection Logic (based on estimated tokens, requests, context)
- [ ] Implement API Integrations for MVP (Google, Ollama, Groq - text chat/completion)
  - [ ] Google REST API integration
  - [ ] Ollama API integration
  - [ ] Groq API integration
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

This project is licensed under the MIT License. See the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.

---

## Community and Support

- **Discussions:** Join the conversation on [GitHub Discussions](https://github.com/your-username/llm-load-balancer/discussions)
- **Issues:** Report bugs or request features via [GitHub Issues](https://github.com/your-username/llm-load-balancer/issues)

---

### Let's Build Together\! 🚀

If this project resonates with you, consider giving it a star ⭐ and sharing it with your community. Together, we can make this tool even better\!

---

## MVP_PLAN.md

```markdown
# MVP Plan: LLM Load Balancer

This document outlines the Minimum Viable Product (MVP) for the LLM Load Balancer project, detailing the necessary components, structs, and functions to achieve a functional core that can load balance text-based chat completion requests across configured Google, Ollama, and Groq APIs, respecting rate limits via a request queue.

**Goal:** Create a server that accepts HTTP requests (mimicking LLM API calls), estimates request tokens using byte count, selects an available LLM based on simple rate limits and context window, queues the request if no LLM is ready, forwards the request once an LLM is available (modifying the body for the specific API), and returns the response.

## Project Structure (Planned Packages)
```

llm-load-balancer/
├── main.go \# Entry point, initialization, HTTP server setup
├── config/ \# Configuration loading and structures
│ └── config.go
├── llm/ \# LLM representation and state management
│ └── llm.go
│ ├── google.go \# Google API implementation
│ ├── ollama.go \# Ollama API implementation
│ └── groq.go \# Groq API implementation
├── queue/ \# Request queue implementation
│ └── queue.go
├── api/ \# API-specific integrations and handling
│ ├── api.go \# Interface/common structures
└── balancer/ \# Core load balancing logic, request handling
└── balancer.go \# Selects LLM, interacts with queue and api packages

## Core Structs

### `config.Config`

- Creates a struct for general config items
  - listen address `localhost`
  - port `8080`
  - log level `INFO`
- Creates a list of structs for LLM configuration
  - provider `ollama, google-genai, openai`
  - model `llama 3.3 70b`
  - base url `localhost:11434`
  - modes (what kind of data does the model support)
    - text
    - vision
    - audio
    - video
    - other
  - rate limit (requsts per min) `10`
  - token limit (tokens per min) `128000`
  - context length `128000`
  - api key name (environment variable to get the api key from) `OPENAI_API_KEY`
  - cost input in dollars `.1`
  - cost output in dollars `1`
  - quality (general vibe of how the model performs, scale 0-1) `.8`

### `llm.LLM`

- Each LLM will have the responsibility of determining when it is available for a new request. They will have their own counters keeping track of rate limits.
- Each LLM will also have an api client that is responsible for constructing, sending, recieving, returning requests.
- The client that populates this will depend on the provider (ollama vs openai vs etc)
- They also keep track of cost accrued
- By keeping track of rate limits they'll also refill counters/buckets periodically (every minute as a go func)

### `queue.RequestQueue`

- Responsible for waiting until an LLM becomes ready (finding them when ready) and submitting the query
- functions include enqueue, dequeue to add and remove requests from the queue

### `api.Client` (Interface)

- An interface that implements the functions required to build, send, recieve, process requests
- These will be implemented uniquely to each provider.
- Will unify the input/outputs of models

## Core Functions

### `balancer.Balancer`

- A central struct to hold the state of the load balancer.

## Main Entry Point (`main.go`)

## Testing (MVP Focus)

- **Unit Tests:**
  - `llm` package: Test `DecrementCounters` and `RefillCounters` for correct mutex usage and state changes. Test a theoretical `IsAvailable` method.
  - `queue` package: Test `Enqueue` and `Dequeue` (including empty queue case).
  - `balancer` package: Test `SelectLLM` with different LLM states. Mock LLMs/API clients if necessary to isolate logic.
- **Integration Tests:**
  - HTTP Handler: Send requests to the load balancer and verify responses. Test cases should include:
    - Successful request when an LLM is available.
    - Requests that cause an LLM's rate limits to be hit, verifying subsequent requests get queued.
    - Requests sent when all LLMs are rate-limited, verifying they are enqueued and processed after a refill.
    - Requests with bodies of different sizes to check token estimation.
    - Test forwarding to different API types (requires mock or actual API endpoints).
  - Refill Logic: Test that LLM counters reset after the refill interval and queued requests are processed.
