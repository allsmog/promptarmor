package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client sends prompt payloads to a target HTTP endpoint.
type Client struct {
	target        string
	promptField   string
	responseField string
	httpClient    *http.Client
}

// New creates a Client that POSTs prompts to target.
func New(target, promptField, responseField string, timeout time.Duration) *Client {
	return &Client{
		target:        target,
		promptField:   promptField,
		responseField: responseField,
		httpClient:    &http.Client{Timeout: timeout},
	}
}

// Send POSTs the prompt to the target and returns the response text.
func (c *Client) Send(ctx context.Context, prompt string) (string, error) {
	payload := map[string]string{c.promptField: prompt}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.target, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Try to extract the response field from JSON.
	var parsed map[string]any
	if err := json.Unmarshal(respBody, &parsed); err == nil {
		if val, ok := parsed[c.responseField]; ok {
			if s, ok := val.(string); ok {
				return s, nil
			}
		}
	}

	// Fall back to raw body.
	return string(respBody), nil
}
