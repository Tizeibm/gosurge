# GoSurge ⚡

**A lightweight concurrent HTTP load testing CLI built with Go.**

GoSurge is a command-line tool written in Go for sending concurrent HTTP requests to a target endpoint and measuring its response performance. It provides a terminal dashboard with request statistics, latency measurements, HTTP status distribution, and execution progress.

The project explores Go's concurrency model through goroutines, channels, worker pools, and synchronization with `sync.WaitGroup`.

> **Project status:** Early development — v1.0

## ✨ Features

- **Concurrent HTTP requests** — Execute multiple requests in parallel using a configurable worker pool.
- **Configurable workload** — Set the total number of requests and the concurrency level.
- **Request timeout** — Configure the maximum duration allowed for each HTTP request.
- **Live progress bar** — Track completed requests directly in the terminal.
- **Performance statistics** — Display total execution time, requests per second (RPS), and minimum, average, and maximum latency.
- **HTTP status distribution** — Summarize responses by HTTP status code.
- **Error tracking** — Count failed requests, including network errors and timeouts.
- **Terminal dashboard** — Present test results using ANSI colors and formatted tables.
- **Input validation** — Validate the target URL, request count, concurrency, and timeout configuration.

## 🛠️ Tech Stack

| Technology | Purpose |
|---|---|
| [Go](https://go.dev/) | Core language |
| [Cobra](https://github.com/spf13/cobra) | CLI commands and flags |
| `net/http` | HTTP client and request execution |
| `sync.WaitGroup` | Worker synchronization |
| Go channels | Job distribution and result collection |
| ANSI escape sequences | Terminal colors and progress display |

## 📋 Prerequisites

- Go installed on your machine.
- An accessible HTTP or HTTPS endpoint.
- A terminal that supports ANSI escape sequences for the colored dashboard.

Check your Go installation:

```bash
go version
```

## 🚀 Installation

Clone the repository:

```bash
git clone https://github.com/Tizeibm/gosurge.git
cd gosurge
```

Initialize the Go module if the repository does not already contain a `go.mod` file:

```bash
go mod init github.com/Tizeibm/gosurge
```

Install the Cobra dependency:

```bash
go get github.com/spf13/cobra
```

Download dependencies:

```bash
go mod tidy
```

Build the executable:

```bash
go build -o gosurge .
```

> Replace the repository URL and module path if your actual GitHub repository uses a different name or location.

## ⚙️ Usage

GoSurge accepts four command-line flags.

| Flag | Short | Default | Description |
|---|---|---:|---|
| `--url` | `-u` | Required | Target HTTP or HTTPS URL |
| `--requests` | `-n` | `100` | Total number of requests |
| `--concurrency` | `-c` | `10` | Number of concurrent workers |
| `--timeout` | `-t` | `5s` | Maximum duration per request |

### Basic example

```bash
./gosurge --url https://example.com
```

This sends 100 requests using 10 concurrent workers, with a timeout of 5 seconds per request.

### Custom load test

```bash
./gosurge \
  --url https://example.com \
  --requests 1000 \
  --concurrency 50 \
  --timeout 3s
```

This configures a workload of 1,000 requests with up to 50 workers and a 3-second timeout per request.

### Short flags

```bash
./gosurge -u https://example.com -n 500 -c 25 -t 5s
```

GoSurge validates its configuration before executing the test.

## 📊 Dashboard

During execution, GoSurge displays a live progress bar. Once the requests have completed, it prints a dashboard containing the following information:

- **Total execution time:** Duration of the test run.
- **Requests per second (RPS):** Number of completed requests divided by elapsed test time.
- **Average latency:** Mean duration of individual requests.
- **Minimum and maximum latency:** Fastest and slowest observed requests.
- **HTTP status distribution:** Number of responses grouped by status code.
- **Error count:** Number of requests that returned an error at the HTTP client level.

The dashboard is designed to make basic HTTP performance results easy to inspect directly from the terminal.

## 🧠 How It Works

GoSurge uses a worker-pool architecture to execute HTTP requests concurrently.

```text
                 ┌──────────────────┐
                 │   CLI Arguments  │
                 └────────┬─────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │ Configuration    │
                 │ Validation       │
                 └────────┬─────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │    Jobs Channel  │
                 └────────┬─────────┘
                          │
              ┌───────────┼───────────┐
              ▼           ▼           ▼
          ┌────────┐  ┌────────┐  ┌────────┐
          │ Worker │  │ Worker │  │ Worker │
          └───┬────┘  └───┬────┘  └───┬────┘
              │           │           │
              └───────────┼───────────┘
                          ▼
                 ┌──────────────────┐
                 │ Results Channel  │
                 └────────┬─────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │ Metrics & Report │
                 └──────────────────┘
```

### Execution flow

1. The CLI reads the user configuration.
2. The configuration is validated.
3. A buffered jobs channel receives the requests to execute.
4. A fixed number of workers consume jobs concurrently.
5. Each worker sends an HTTP request and records its result.
6. Results are collected and aggregated into statistics.
7. The terminal dashboard displays the final report.

The worker pool limits the number of requests executing concurrently, while the channels coordinate work distribution and result collection.

## 📐 Project Structure

The current implementation is organized around the following components:

```text
gosurge/
├── main.go       # CLI, worker pool, HTTP requests and dashboard
├── go.mod        # Module definition and dependencies
├── go.sum        # Dependency checksums
└── README.md     # Project documentation
```

As the project grows, the code can be separated into dedicated packages for configuration, execution, metrics, and terminal output.

## 🧪 Testing

Run the project's tests with:

```bash
go test ./...
```

Check for common concurrency problems with:

```bash
go test -race ./...
```

The race detector is useful for identifying unsynchronized access to shared memory. These commands should be run after tests have been implemented; they do not themselves guarantee that every concurrency issue has been covered.

## ⚠️ Current Limitations

GoSurge is an early-stage project and has several areas for improvement:

- Latency percentiles such as p95 and p99 are not yet implemented.
- The current dashboard highlights HTTP 200 and 500 responses separately rather than presenting every status code individually.
- HTTP error responses, such as 404 or 500, are not automatically classified as client errors by `net/http`; they must be evaluated using the status code.
- The current implementation buffers all jobs and results in memory, which may become expensive for very large request counts.
- The current execution timing should be measured through completion of all requests to obtain accurate throughput.
- Context-based cancellation, configurable retry policies, and rate limiting are potential future improvements.
- The tool does not yet provide distributed load generation or persistent historical reports.

These limitations are part of the project's ongoing development roadmap.

## 🗺️ Roadmap

- [ ] Improve execution timing and throughput accuracy.
- [ ] Add unit tests for request execution and metric calculations.
- [ ] Add integration tests using `net/http/httptest`.
- [ ] Support additional HTTP status codes in the dashboard.
- [ ] Implement p50, p95, and p99 latency percentiles.
- [ ] Add graceful cancellation with `context.Context`.
- [ ] Add configurable request methods and headers.
- [ ] Introduce configurable request rate limiting.
- [ ] Add JSON report export.
- [ ] Add CI with GitHub Actions and automated race detection.
- [ ] Publish cross-platform release binaries.

## 🔐 Responsible Usage

GoSurge is intended for authorized performance testing, development, and learning.

Only test endpoints you own or have explicit permission to test. High-concurrency workloads can overload servers, consume resources, or disrupt production services.

Start with a small request count and low concurrency, and increase them gradually while monitoring the target system.

## 🎯 Project Goals

This project is an opportunity to deepen practical knowledge of:

- Go concurrency with goroutines and channels.
- Worker-pool design and synchronization.
- HTTP client behavior and network timeouts.
- Performance measurement and latency statistics.
- Error handling and resource management.
- CLI design and terminal-based reporting.
- Automated testing and race detection.

## 👤 Author

**Tize Ibrahim Ahmed**

Software Engineering Student | Aspiring Cloud & DevOps Engineer

- GitHub: [@Tizeibm](https://github.com/Tizeibm)

## 📄 License

This project is intended to be released under the MIT License. Add a `LICENSE` file containing the MIT License text before claiming that the repository is officially distributed under that license.

---

*Built with Go to explore concurrent HTTP execution, performance measurement, and reliable command-line tooling.*
