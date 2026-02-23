package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "promptarmor.yaml")

	content := `target: http://localhost:8080/chat
suite: jailbreak
concurrency: 10
timeout: 60s
prompt_field: message
response_field: reply
api_key: sk-test-123
provider: anthropic
model: claude-haiku-4-5-20251001
output: json
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Target != "http://localhost:8080/chat" {
		t.Errorf("Target = %q, want %q", cfg.Target, "http://localhost:8080/chat")
	}
	if cfg.Suite != "jailbreak" {
		t.Errorf("Suite = %q, want %q", cfg.Suite, "jailbreak")
	}
	if cfg.Concurrency != 10 {
		t.Errorf("Concurrency = %d, want 10", cfg.Concurrency)
	}
	if cfg.Timeout != "60s" {
		t.Errorf("Timeout = %q, want %q", cfg.Timeout, "60s")
	}
	if cfg.PromptField != "message" {
		t.Errorf("PromptField = %q, want %q", cfg.PromptField, "message")
	}
	if cfg.ResponseField != "reply" {
		t.Errorf("ResponseField = %q, want %q", cfg.ResponseField, "reply")
	}
	if cfg.APIKey != "sk-test-123" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "sk-test-123")
	}
	if cfg.Provider != "anthropic" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "anthropic")
	}
	if cfg.Model != "claude-haiku-4-5-20251001" {
		t.Errorf("Model = %q, want %q", cfg.Model, "claude-haiku-4-5-20251001")
	}
	if cfg.Output != "json" {
		t.Errorf("Output = %q, want %q", cfg.Output, "json")
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/promptarmor.yaml")
	if err != nil {
		t.Fatalf("Load() error: %v, want nil for missing file", err)
	}
	if cfg != (FileConfig{}) {
		t.Errorf("Load() = %+v, want zero struct", cfg)
	}
}

func TestLoadPartialFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "promptarmor.yaml")

	content := `target: http://localhost:8080/chat
concurrency: 3
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Target != "http://localhost:8080/chat" {
		t.Errorf("Target = %q, want %q", cfg.Target, "http://localhost:8080/chat")
	}
	if cfg.Concurrency != 3 {
		t.Errorf("Concurrency = %d, want 3", cfg.Concurrency)
	}
	// Unset fields should be zero values.
	if cfg.Suite != "" {
		t.Errorf("Suite = %q, want empty", cfg.Suite)
	}
	if cfg.Timeout != "" {
		t.Errorf("Timeout = %q, want empty", cfg.Timeout)
	}
	if cfg.Provider != "" {
		t.Errorf("Provider = %q, want empty", cfg.Provider)
	}
}

func TestLoadEnvVarNotExpanded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "promptarmor.yaml")

	content := `api_key: ${ANTHROPIC_API_KEY}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Raw string, not expanded.
	if cfg.APIKey != "${ANTHROPIC_API_KEY}" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "${ANTHROPIC_API_KEY}")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")

	if err := os.WriteFile(path, []byte(":::invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want error for invalid YAML")
	}
}
