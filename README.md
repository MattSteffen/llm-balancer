# LLM Load Balancer

The **LLM Load Balancer** is an open-source project that helps developers manage multiple LLM APIs with rate limits efficiently. By balancing load based on token limits, request rates, and context lengths, this server provides seamless routing for API requests, allowing developers to maximize their use of available free-tier APIs during development.

---

## Key Features

- **Rate-Limit Awareness:** Automatically balances load across APIs to avoid exceeding token and request rate limits.
- **Dynamic Forwarding:** Routes HTTP calls to the most appropriate LLM API based on:
  - Token availability
  - Requests per minute limit
  - Context length
- **Configurable Settings:** Easily set up multiple APIs via a YAML configuration file.
- **Automatic Rate Limit Reset:** Tokens and request counters are replenished periodically based on API limits.

---

## Getting Started

### Prerequisites

- **Go 1.23.3 or higher**
- API keys for the LLM APIs you intend to use
- A YAML configuration file (example provided below)

### Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/mattsteffen/llm-load-balancer.git
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
  log_level: "debug"

llms:
  - name: "API1"
    model: "gpt-4"
    url: "https://api.example1.com"
    rate_limit: 100000
    requests_per_min: 60
    context_length: 4096
    api_key_name: "API1_KEY"

  - name: "API2"
    model: "gpt-3.5-turbo"
    url: "https://api.example2.com"
    rate_limit: 50000
    requests_per_min: 30
    context_length: 2048
    api_key_name: "API2_KEY"
```

### Environment Variables

Set your API keys as environment variables corresponding to the `api_key_name` in your config file. For example:

```bash
export API1_KEY=your_api1_key_here
export API2_KEY=your_api2_key_here
```

---

## Usage

1. **Send a Request**  
   Point your application to the load balancer endpoint (e.g., `http://localhost:8080`).
   The server will handle forwarding requests to the appropriate LLM API.

2. **Monitor Logs**  
   Logs provide insights into:

   - Request routing
   - Token usage
   - API rate limits

   To adjust the log verbosity, update the `log_level` in your config (`debug`, `info`, `warn`, `error`).

---

## Contributing

Contributions are welcome! To get started:

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

### Ideas for Contribution

- Add support for more API providers
- Enhance load-balancing algorithms
- Add metrics and monitoring (e.g., Prometheus, Grafana)
- Build a UI for configuration and monitoring

---

## Roadmap

- [ ] Add queue functionality
  - [ ] channel based queue
  - [ ] priority queue for advanced prioritization (when money is involved)
- [ ] Configure into packages and folders
- [ ] Unit tests
- [ ] Add metrics and performance monitoring
- [ ] Provide support for additional LLM APIs
- [ ] Implement persistent storage for request logs
- [ ] Introduce retry mechanisms for failed requests

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Community and Support

- **Discussions:** Join the conversation on [GitHub Discussions](https://github.com/your-username/llm-load-balancer/discussions)
- **Issues:** Report bugs or request features via [GitHub Issues](https://github.com/your-username/llm-load-balancer/issues)

---

### Let's Build Together! 🚀

If this project resonates with you, consider giving it a star ⭐ and sharing it with your community. Together, we can make this tool even better!
