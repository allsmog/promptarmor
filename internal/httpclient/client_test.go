package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSend_JSONExtraction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request format.
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["prompt"] == "" {
			t.Error("expected prompt field in body")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response": "I cannot help with that.",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "prompt", "response", 5*time.Second)
	resp, err := c.Send(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "I cannot help with that." {
		t.Fatalf("unexpected response: %q", resp)
	}
}

func TestSend_RawFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "plain text response")
	}))
	defer srv.Close()

	c := New(srv.URL, "prompt", "response", 5*time.Second)
	resp, err := c.Send(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "plain text response" {
		t.Fatalf("unexpected response: %q", resp)
	}
}

func TestSend_MissingField_FallsBackToRawJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"other_field": "some value",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "prompt", "response", 5*time.Second)
	resp, err := c.Send(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp, "other_field") {
		t.Fatalf("expected raw JSON fallback, got: %q", resp)
	}
}

func TestSend_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(srv.URL, "prompt", "response", 5*time.Second)
	_, err := c.Send(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("expected HTTP 500 in error, got: %v", err)
	}
}

func TestSend_ConnectionError(t *testing.T) {
	c := New("http://127.0.0.1:1", "prompt", "response", 1*time.Second)
	_, err := c.Send(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
