package mockdata

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"mock-openai/internal/models"
)

// batchLine is the shape of one line of a batch input JSONL file, per the
// (undocumented-in-schema, but well-known) batch request format.
type batchLine struct {
	CustomID string          `json:"custom_id"`
	Method   string          `json:"method,omitempty"`
	URL      string          `json:"url,omitempty"`
	Body     json.RawMessage `json:"body,omitempty"`
}

type batchResponseLine struct {
	ID       string       `json:"id"`
	CustomID string       `json:"custom_id"`
	Response *batchResult `json:"response"`
	Error    *batchError  `json:"error"`
}

type batchResult struct {
	StatusCode int         `json:"status_code"`
	RequestID  string      `json:"request_id"`
	Body       interface{} `json:"body"`
}

type batchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// mockResponseBody builds the mock response body for one batch request line,
// reusing the same generators as the live endpoints where the endpoint is
// one this mock server implements. Endpoints without a live handler here
// (moderations, videos, legacy completions) get a minimal placeholder body
// so the batch still completes rather than failing every line.
func mockResponseBody(endpoint string, rawBody json.RawMessage) (interface{}, error) {
	switch endpoint {
	case "/v1/chat/completions":
		var req models.ChatRequest
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return nil, err
		}
		return BuildChatResponse(req), nil
	case "/v1/embeddings":
		var req models.EmbeddingRequest
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return nil, err
		}
		return GenerateMockEmbeddingResponse(req), nil
	default:
		return map[string]interface{}{"mock": true}, nil
	}
}

// BuildBatch parses a batch input file (JSONL) and synchronously executes it
// against this mock server's response generators, returning the completed
// Batch plus the assembled output file's JSONL bytes and a suggested
// filename. Real batches run asynchronously over up to 24h; this mock
// completes them immediately so SDKs polling for completion see a result
// right away.
func BuildBatch(req models.CreateBatchRequest, inputContent []byte) (models.Batch, []byte, string) {
	id := "batch_" + randomSuffix()
	now := time.Now().Unix()

	var out bytes.Buffer
	total, completed, failed := 0, 0, 0

	scanner := bufio.NewScanner(bytes.NewReader(inputContent))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		total++

		var in batchLine
		if err := json.Unmarshal([]byte(line), &in); err != nil {
			failed++
			writeOutputLine(&out, batchResponseLine{
				ID:       "batch_req_" + randomSuffix(),
				CustomID: "",
				Error:    &batchError{Code: "invalid_json_line", Message: err.Error()},
			})
			continue
		}

		endpoint := in.URL
		if endpoint == "" {
			endpoint = req.Endpoint
		}

		body, err := mockResponseBody(endpoint, in.Body)
		if err != nil {
			failed++
			writeOutputLine(&out, batchResponseLine{
				ID:       "batch_req_" + randomSuffix(),
				CustomID: in.CustomID,
				Error:    &batchError{Code: "invalid_request_body", Message: err.Error()},
			})
			continue
		}

		completed++
		writeOutputLine(&out, batchResponseLine{
			ID:       "batch_req_" + randomSuffix(),
			CustomID: in.CustomID,
			Response: &batchResult{
				StatusCode: 200,
				RequestID:  "req_" + randomSuffix(),
				Body:       body,
			},
		})
	}

	batch := models.Batch{
		ID:               id,
		Object:           "batch",
		Endpoint:         req.Endpoint,
		Errors:           nil,
		InputFileID:      req.InputFileID,
		CompletionWindow: req.CompletionWindow,
		Status:           "completed",
		CreatedAt:        now,
		InProgressAt:     &now,
		CompletedAt:      &now,
		RequestCounts: models.BatchRequestCounts{
			Total:     total,
			Completed: completed,
			Failed:    failed,
		},
		Metadata: req.Metadata,
	}

	return batch, out.Bytes(), fmt.Sprintf("%s_output.jsonl", id)
}

func writeOutputLine(out *bytes.Buffer, line batchResponseLine) {
	data, _ := json.Marshal(line)
	out.Write(data)
	out.WriteByte('\n')
}
