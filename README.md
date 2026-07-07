# Mock OpenAI Server

A high-performance, lightweight mock of the OpenAI API built in **Go** using the **Gin** framework. This server is designed for local development and integration testing, allowing you to build and test AI-powered applications without spending a dime on API credits.

Response shapes are kept in sync with the official OpenAI OpenAPI spec (`openai.yml`, vendored in this repo). See `memory.md` for the current fidelity status of each API area and known deviations from the real API.

Covers Chat Completions (including streaming, tool/function calling, and multimodal content parts), Embeddings, Images, Audio, Models, Files, and Batch — see [Usage Examples](#usage-examples) below.

---

## Project Structure

```text
mock-openai/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── config/               # .env loading + API_KEY config
│   ├── database/             # SQLite request logging
│   ├── errs/                 # Centralized OpenAI-shaped error envelope
│   ├── handlers/             # HTTP logic (Chat, Images, Audio, Files, Batch, etc.)
│   ├── models/                # OpenAI-compatible structs
│   ├── mockdata/               # Mock response generators
│   ├── router/                 # Gin route definitions + auth middleware
│   └── store/                  # In-memory stores (chat, files, batches)
├── docs/                     # Generated Swagger/OpenAPI docs (see below)
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


3. (Optional) Configure an API key:
```bash
cp .env.example .env
# then set API_KEY=sk-your-mock-key in .env

```

If `API_KEY` is left unset, the server runs in permissive mode: any non-empty
Bearer token is accepted on `/v1/*` routes. If set, requests must present that
exact token or get a `401` with the real OpenAI error envelope.

### Running the Server

```bash
make run

```

The server will start on `http://localhost:8080`.

---

## API Docs (Swagger)

Interactive API docs for every route this mock actually implements (as opposed
to the full upstream spec in `openai.yml`) are generated from annotations on
the handler functions using [swaggo/swag](https://github.com/swaggo/swag),
and served at runtime via `gin-swagger`:

* Swagger UI: `http://localhost:8080/swagger/index.html`
* Raw OpenAPI 2.0 spec: `http://localhost:8080/swagger/doc.json` (also checked
  into `docs/swagger.json` / `docs/swagger.yaml`)

If you change a handler's request/response shape or add a route, regenerate
the docs:

```bash
make swagger
```

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

All `/v1/*` routes require an `Authorization: Bearer <token>` header. With no
`API_KEY` configured (the default), any non-empty token is accepted — this mock
doesn't validate real API keys, it just enforces the header shape like the real
API does. Set `API_KEY` (see [Getting Started](#getting-started)) to require an
exact match instead.

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

### Chat Completion (tool/function calling)

`tool_choice: "auto"` (the default once `tools` is set) triggers a call unless
the last message already has role `tool` — so the standard ask → `tool_calls` →
append `tool` result → final answer round trip works with zero configuration.

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-mock" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "What is the weather in Paris?"}],
    "tools": [{
      "type": "function",
      "function": {
        "name": "get_weather",
        "parameters": {
          "type": "object",
          "properties": {"location": {"type": "string"}}
        }
      }
    }]
  }'

```

### Chat Completion (multimodal content)

`content` also accepts an array of text/image_url/input_audio/file parts:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-mock" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{
      "role": "user",
      "content": [
        {"type": "text", "text": "What is in this image?"},
        {"type": "image_url", "image_url": {"url": "https://example.com/cat.png"}}
      ]
    }]
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

### Files

Upload a file, then reference it (e.g. as a Batch input file):

```bash
curl http://localhost:8080/v1/files \
  -H "Authorization: Bearer sk-mock" \
  -F "purpose=batch" \
  -F "file=@requests.jsonl"

```

### Batch

Batches complete synchronously and instantly in this mock — each JSONL line is
executed against the server's own chat/embeddings generators, so the response
is ready as soon as creation returns:

```bash
curl http://localhost:8080/v1/batches \
  -H "Authorization: Bearer sk-mock" \
  -H "Content-Type: application/json" \
  -d '{
    "input_file_id": "file-abc123",
    "endpoint": "/v1/chat/completions",
    "completion_window": "24h"
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
| `make swagger` | Regenerates Swagger/OpenAPI docs in `docs/` from handler annotations. |
