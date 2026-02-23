package judge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockOpenAIServer(verdictText string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openaiResponse{
			Choices: []openaiChoice{
				{Message: openaiMessage{Role: "assistant", Content: verdictText}},
			},
		})
	}))
}

func TestOpenAIProvider_Call(t *testing.T) {
	srv := mockOpenAIServer("VERDICT: SAFE\nRefused the injection.")
	defer srv.Close()

	p := NewOpenAIProvider("test-key", "", srv.URL)
	text, err := p.Call(context.Background(), "system", "user msg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text == "" {
		t.Fatal("expected non-empty response")
	}
}

func TestOpenAIProvider_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	p := NewOpenAIProvider("test-key", "", srv.URL)
	_, err := p.Call(context.Background(), "system", "user msg")
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}

func TestOpenAIProvider_Name(t *testing.T) {
	p := NewOpenAIProvider("key", "", "")
	if p.Name() != "OpenAI (gpt-4o-mini)" {
		t.Fatalf("unexpected name: %s", p.Name())
	}
}
