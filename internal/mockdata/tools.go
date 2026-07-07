package mockdata

import (
	"encoding/json"
	"fmt"

	"mock-openai/internal/models"
)

// ParseToolChoice interprets the oneOf(string|object) `tool_choice` field
// (ChatCompletionToolChoiceOption in openai.yml). mode is one of "none",
// "auto", "required", or "function" (with functionName set).
func ParseToolChoice(raw json.RawMessage, hasTools bool) (mode string, functionName string, err error) {
	if len(raw) == 0 {
		if hasTools {
			return "auto", "", nil
		}
		return "none", "", nil
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		switch asString {
		case "none", "auto", "required":
			return asString, "", nil
		}
	}

	var asObject struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(raw, &asObject); err == nil && asObject.Type == "function" && asObject.Function.Name != "" {
		return "function", asObject.Function.Name, nil
	}

	return "", "", fmt.Errorf("must be 'none', 'auto', 'required', or an object naming a function")
}

// BuildToolCalls decides whether, and which, tools the mock "calls" for this
// request. Since a mock can't infer intent from message content, the trigger
// is conversation-state-driven: a forced tool_choice (named function or
// "required") always calls; "auto" (the default once tools are present)
// calls unless the client's last message already carries a tool result,
// which lets the natural ask -> tool_calls -> tool-result -> final-answer
// round trip work without any content inspection. See memory.md for the
// full rationale.
func BuildToolCalls(req models.ChatRequest) []models.ToolCall {
	if len(req.Tools) == 0 {
		return nil
	}

	mode, functionName, err := ParseToolChoice(req.ToolChoice, true)
	if err != nil || mode == "none" {
		return nil
	}

	if mode == "function" {
		for _, t := range req.Tools {
			if t.Function.Name == functionName {
				return []models.ToolCall{buildToolCall(t)}
			}
		}
		return nil
	}

	if mode == "auto" {
		last := req.Messages[len(req.Messages)-1]
		if last.Role == "tool" {
			return nil
		}
	}

	toolsToCall := req.Tools[:1]
	parallel := req.ParallelToolCalls == nil || *req.ParallelToolCalls
	if parallel {
		toolsToCall = req.Tools
	}

	calls := make([]models.ToolCall, len(toolsToCall))
	for i, t := range toolsToCall {
		calls[i] = buildToolCall(t)
	}
	return calls
}

func buildToolCall(tool models.ChatTool) models.ToolCall {
	return models.ToolCall{
		ID:   "call_" + randomSuffix(),
		Type: "function",
		Function: models.ToolCallFunction{
			Name:      tool.Function.Name,
			Arguments: mockArguments(tool.Function.Parameters),
		},
	}
}

// mockArguments synthesizes a plausible JSON-encoded arguments string by
// walking the tool's JSON-Schema `parameters.properties`. Every listed
// property is filled (not just `required`), since Structured Outputs
// (`strict: true`) requires all properties anyway, and it gives callers a
// fuller example to exercise their tool-call handling with.
func mockArguments(parameters map[string]interface{}) string {
	properties, _ := parameters["properties"].(map[string]interface{})
	if len(properties) == 0 {
		return "{}"
	}

	args := make(map[string]interface{}, len(properties))
	for name, rawSchema := range properties {
		schema, _ := rawSchema.(map[string]interface{})
		args[name] = mockValueForSchema(name, schema)
	}

	encoded, err := json.Marshal(args)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func mockValueForSchema(name string, schema map[string]interface{}) interface{} {
	if schema == nil {
		return "mock_value"
	}
	if enumVals, ok := schema["enum"].([]interface{}); ok && len(enumVals) > 0 {
		return enumVals[0]
	}
	switch schema["type"] {
	case "string":
		return "mock_" + name
	case "integer", "number":
		return 0
	case "boolean":
		return true
	case "array":
		return []interface{}{}
	case "object":
		return map[string]interface{}{}
	default:
		return "mock_value"
	}
}
