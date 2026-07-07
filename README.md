# Mock OpenAI Server

A high-performance, lightweight mock of the OpenAI API built in **Go** using the **Gin** framework. This server is designed for local development and integration testing, allowing you to build and test AI-powered applications without spending a dime on API credits.

Response shapes are kept in sync with the official OpenAI OpenAPI spec (`openai.yml`, vendored in this repo). See `memory.md` for the current fidelity status of each API area and known deviations from the real API.



---

## Project Structure

```text
mock-openai/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── handlers/            # HTTP logic (Chat, Images, Audio, etc.)
│   ├── models/              # OpenAI-compatible structs
│   ├── mockdata/            # Mock response generators
│   └── router/              # Gin route definitions
├── go.mod                   # Dependencies
└── Makefile                 # Automation scripts

```

---

## Getting Started

### Prerequisites

* [Go 1.21+](https://golang.org/dl/)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/youruser/mock-openai.git
cd mock-openai

```


2. Sync dependencies:
```bash
make tidy

```



### Running the Server

```bash
make run

```

The server will start on `http://localhost:8080`.

---

## Testing

We use Go's built-in testing tool with a focus on concurrency and race conditions.

```bash
# Run all tests
make test

# Run specifically for the handlers package
go test ./internal/handlers/... -v

```

---

## Usage Examples

All `/v1/*` routes require an `Authorization: Bearer <token>` header (any non-empty token is accepted — this mock doesn't validate real API keys, it just enforces the header shape like the real API does).

### Chat Completion

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-mock" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

```

### Chat Completion (streaming)

```bash
curl -N http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-mock" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'

```

### Embeddings (1536 Dimensions)

```bash
curl http://localhost:8080/v1/embeddings \
  -H "Authorization: Bearer sk-mock" \
  -H "Content-Type: application/json" \
  -d '{
    "input": "The food was delicious",
    "model": "text-embedding-3-small"
  }'

```

---

## Makefile Commands

| Command | Description |
| --- | --- |
| `make run` | Runs the server locally. |
| `make tidy` | Formats code (`go fmt`) and cleans `go.mod`. |
| `make test` | Runs tests with the `-race` detector. |
| `make build` | Compiles a binary for production. |

