package judge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockAnthropicServer(verdictText string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			http.Error(w, "bad version", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: verdictText},
			},
		})
	}))
}

func TestAnthropicProvider_Call(t *testing.T) {
	srv := mockAnthropicServer("VERDICT: VULNERABLE\nComplied with injection.")
	defer srv.Close()

	p := NewAnthropicProvider("test-key", "", srv.URL)
	text, err := p.Call(context.Background(), "system", "user msg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text == "" {
		t.Fatal("expected non-empty response")
	}
}

func TestAnthropicProvider_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	p := NewAnthropicProvider("test-key", "", srv.URL)
	_, err := p.Call(context.Background(), "system", "user msg")
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}

func TestAnthropicProvider_Name(t *testing.T) {
	p := NewAnthropicProvider("key", "", "")
	if p.Name() != "Anthropic (claude-haiku-4-5-20251001)" {
		t.Fatalf("unexpected name: %s", p.Name())
	}

	p2 := NewAnthropicProvider("key", "claude-sonnet-4-6", "")
	if p2.Name() != "Anthropic (claude-sonnet-4-6)" {
		t.Fatalf("unexpected name: %s", p2.Name())
	}
}
