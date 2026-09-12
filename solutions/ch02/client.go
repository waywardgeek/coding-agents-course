package main

// The transport. Vendor types live here and nowhere else: responses are
// parsed into events at this boundary and the wire structs never leak inward.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultModel is a snapshot of a moving target. Never pick a model id from
// memory — ask the API which ones exist:
//
//	curl -s https://api.anthropic.com/v1/models?limit=40 \
//	  -H "x-api-key: $ANTHROPIC_API_KEY" -H "anthropic-version: 2023-06-01"
//
// Mid-turn system messages — the first-class way to carry a hint — need
// claude-opus-4-8 or newer. On a model without them you will see a one-round
// lag and conclude your code is broken when what you are seeing is the model.
const DefaultModel = "claude-sonnet-5"

// APIResponse is the subset of a Messages API response this exercise reads.
type APIResponse struct {
	ID      string `json:"id"`
	Content []struct {
		Type      string          `json:"type"`
		Text      string          `json:"text"`
		ID        string          `json:"id"`
		Name      string          `json:"name"`
		Input     json.RawMessage `json:"input"`
		Thinking  string          `json:"thinking"`
		Signature string          `json:"signature"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Client talks to the Messages API.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

// NewClient reads its configuration from the environment. ANTHROPIC_BASE_URL
// is the seam that lets the grader — and the course proxy — stand in for the
// real API without changing a line of this program.
func NewClient() *Client {
	base := os.Getenv("ANTHROPIC_BASE_URL")
	if base == "" {
		base = os.Getenv("ANTHROPIC_API_URL")
	}
	if base == "" {
		base = "https://api.anthropic.com"
	}
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		BaseURL: strings.TrimRight(base, "/"),
		APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
		Model:   model,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

// Send posts already-rendered request bytes and parses the reply.
func (c *Client) Send(body []byte) (*APIResponse, error) {
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	var out APIResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("response was not JSON: %w", err)
	}
	return &out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
