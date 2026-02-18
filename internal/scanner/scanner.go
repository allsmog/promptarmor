package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shayaun-nejad/promptarmor/internal/detect"
	"github.com/shayaun-nejad/promptarmor/internal/httpclient"
	"github.com/shayaun-nejad/promptarmor/internal/judge"
	"github.com/shayaun-nejad/promptarmor/internal/suites"
)

// Result holds the outcome of a single test case.
type Result struct {
	TestName string `json:"test_name"`
	Suite    string `json:"suite"`
	Passed   bool   `json:"passed"`
	Detail   string `json:"detail,omitempty"`
}

// Config holds scanner configuration.
type Config struct {
	Target        string
	Suite         string
	Concurrency   int
	Timeout       time.Duration
	PromptField   string
	ResponseField string
	APIKey        string
}

// Scanner runs prompt injection test suites against a target.
type Scanner struct {
	cfg    Config
	client *httpclient.Client
	judge  *judge.Judge
}

// New creates a Scanner with the given configuration.
func New(cfg Config) *Scanner {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.PromptField == "" {
		cfg.PromptField = "prompt"
	}
	if cfg.ResponseField == "" {
		cfg.ResponseField = "response"
	}

	s := &Scanner{
		cfg:    cfg,
		client: httpclient.New(cfg.Target, cfg.PromptField, cfg.ResponseField, cfg.Timeout),
	}
	if cfg.APIKey != "" {
		s.judge = judge.New(cfg.APIKey)
	}
	return s
}

// UsingJudge reports whether the LLM judge is enabled.
func (s *Scanner) UsingJudge() bool {
	return s.judge != nil
}

// Run executes the selected test suites and returns results.
func (s *Scanner) Run(ctx context.Context) ([]Result, error) {
	cases := loadCases(s.cfg.Suite)
	if len(cases) == 0 {
		return nil, fmt.Errorf("no test cases found for suite %q", s.cfg.Suite)
	}

	results := make([]Result, len(cases))
	sem := make(chan struct{}, s.cfg.Concurrency)
	var wg sync.WaitGroup

	for i, tc := range cases {
		wg.Add(1)
		go func(idx int, tc suites.TestCase) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = s.runOne(ctx, tc)
		}(i, tc)
	}

	wg.Wait()
	return results, nil
}

func (s *Scanner) runOne(ctx context.Context, tc suites.TestCase) Result {
	resp, err := s.client.Send(ctx, tc.Prompt)
	if err != nil {
		return Result{
			TestName: tc.Name,
			Suite:    tc.Suite,
			Passed:   false,
			Detail:   fmt.Sprintf("request error: %v", err),
		}
	}

	// Try LLM judge first if available.
	if s.judge != nil {
		injected, reasoning, err := s.judge.Evaluate(ctx, tc.Prompt, resp)
		if err == nil {
			if injected {
				return Result{
					TestName: tc.Name,
					Suite:    tc.Suite,
					Passed:   false,
					Detail:   fmt.Sprintf("injection detected (judge): %s", reasoning),
				}
			}
			return Result{
				TestName: tc.Name,
				Suite:    tc.Suite,
				Passed:   true,
				Detail:   reasoning,
			}
		}
		// Judge failed — fall through to substring detection.
	}

	// Fallback: substring indicator matching.
	matched, indicator := detect.Check(resp, tc.Indicators)
	if matched {
		return Result{
			TestName: tc.Name,
			Suite:    tc.Suite,
			Passed:   false,
			Detail:   fmt.Sprintf("injection detected: matched indicator %q", indicator),
		}
	}

	return Result{
		TestName: tc.Name,
		Suite:    tc.Suite,
		Passed:   true,
	}
}

func loadCases(suite string) []suites.TestCase {
	switch suite {
	case "jailbreak":
		return suites.Jailbreak()
	case "tool-abuse":
		return suites.ToolAbuse()
	default:
		return suites.All()
	}
}
