package judge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockGeminiServer(verdictText string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "key=test-key") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(geminiResponse{
			Candidates: []geminiCandidate{
				{
					Content: geminiContent{
						Parts: []geminiPart{{Text: verdictText}},
					},
				},
			},
		})
	}))
}

func TestGeminiProvider_Call(t *testing.T) {
	srv := mockGeminiServer("VERDICT: VULNERABLE\nComplied with injection.")
	defer srv.Close()

	p := NewGeminiProvider("test-key", "", srv.URL)
	text, err := p.Call(context.Background(), "system", "user msg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text == "" {
		t.Fatal("expected non-empty response")
	}
}

func TestGeminiProvider_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	p := NewGeminiProvider("test-key", "", srv.URL)
	_, err := p.Call(context.Background(), "system", "user msg")
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}

func TestGeminiProvider_Name(t *testing.T) {
	p := NewGeminiProvider("key", "", "")
	if p.Name() != "Gemini (gemini-2.0-flash)" {
		t.Fatalf("unexpected name: %s", p.Name())
	}
}
