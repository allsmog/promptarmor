package judge

import "testing"

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		want    string
		wantErr bool
	}{
		{"anthropic key", "sk-ant-abc123", "anthropic", false},
		{"openai key", "sk-proj-abc123", "openai", false},
		{"gemini key", "AIzaSyABC123", "gemini", false},
		{"unknown key", "xyz-unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detectProvider(tt.apiKey)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveProvider_Explicit(t *testing.T) {
	tests := []struct {
		provider string
		wantType string
	}{
		{"anthropic", "*judge.AnthropicProvider"},
		{"openai", "*judge.OpenAIProvider"},
		{"gemini", "*judge.GeminiProvider"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			p, err := ResolveProvider(ProviderConfig{
				ProviderName: tt.provider,
				APIKey:       "test-key",
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Check it's non-nil and has a name.
			if p.Name() == "" {
				t.Fatal("expected non-empty provider name")
			}
		})
	}
}

func TestResolveProvider_AutoDetect(t *testing.T) {
	p, err := ResolveProvider(ProviderConfig{APIKey: "sk-ant-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := p.(*AnthropicProvider); !ok {
		t.Fatalf("expected AnthropicProvider, got %T", p)
	}

	p, err = ResolveProvider(ProviderConfig{APIKey: "sk-proj-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := p.(*OpenAIProvider); !ok {
		t.Fatalf("expected OpenAIProvider, got %T", p)
	}
}

func TestResolveProvider_UnknownProvider(t *testing.T) {
	_, err := ResolveProvider(ProviderConfig{ProviderName: "azure", APIKey: "key"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestResolveProvider_ModelOverride(t *testing.T) {
	p, err := ResolveProvider(ProviderConfig{
		ProviderName: "openai",
		Model:        "gpt-4o",
		APIKey:       "test-key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "OpenAI (gpt-4o)" {
		t.Fatalf("expected model override in name, got %s", p.Name())
	}
}
