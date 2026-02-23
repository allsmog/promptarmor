package judge

import (
	"context"
	"fmt"
)

// Judge uses an LLM to evaluate whether prompt injection succeeded.
type Judge struct {
	provider Provider
}

// New creates a Judge that delegates to the given provider.
func New(provider Provider) *Judge {
	return &Judge{provider: provider}
}

// Evaluate determines if the LLM response indicates injection succeeded.
// Returns (injectionDetected, reasoning, error).
func (j *Judge) Evaluate(ctx context.Context, prompt, response string) (bool, string, error) {
	userMsg := fmt.Sprintf("INJECTION ATTEMPT:\n%s\n\nAPPLICATION RESPONSE:\n%s", prompt, response)

	text, err := j.provider.Call(ctx, systemPrompt, userMsg)
	if err != nil {
		return false, "", err
	}

	return parseVerdict(text)
}

// ProviderName returns the display name of the underlying provider.
func (j *Judge) ProviderName() string {
	return j.provider.Name()
}
