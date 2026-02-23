package judge

import (
	"fmt"
	"strings"
)

// ProviderConfig holds the parameters needed to resolve a provider.
type ProviderConfig struct {
	ProviderName string // "anthropic", "openai", "gemini", or "" for auto-detect
	Model        string // override default model; empty means use provider default
	APIKey       string
	BaseURL      string // override base URL (mainly for testing)
}

// ResolveProvider creates the appropriate Provider from the given config.
func ResolveProvider(cfg ProviderConfig) (Provider, error) {
	name := cfg.ProviderName
	if name == "" {
		var err error
		name, err = detectProvider(cfg.APIKey)
		if err != nil {
			return nil, err
		}
	}

	switch strings.ToLower(name) {
	case "anthropic":
		return NewAnthropicProvider(cfg.APIKey, cfg.Model, cfg.BaseURL), nil
	case "openai":
		return NewOpenAIProvider(cfg.APIKey, cfg.Model, cfg.BaseURL), nil
	case "gemini":
		return NewGeminiProvider(cfg.APIKey, cfg.Model, cfg.BaseURL), nil
	default:
		return nil, fmt.Errorf("unknown provider %q (supported: anthropic, openai, gemini)", name)
	}
}

// detectProvider guesses the provider from the API key prefix.
func detectProvider(apiKey string) (string, error) {
	switch {
	case strings.HasPrefix(apiKey, "sk-ant-"):
		return "anthropic", nil
	case strings.HasPrefix(apiKey, "sk-"):
		return "openai", nil
	case strings.HasPrefix(apiKey, "AIza"):
		return "gemini", nil
	default:
		return "", fmt.Errorf("cannot auto-detect provider from API key prefix; use --provider flag")
	}
}
