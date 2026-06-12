package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fatih/color"
	"github.com/allsmog/promptarmor/internal/config"
	"github.com/allsmog/promptarmor/internal/report"
	"github.com/allsmog/promptarmor/internal/scanner"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "promptarmor",
		Short: "LLM Prompt Injection Scanner",
		Long:  "promptarmor scans LLM applications for prompt injection vulnerabilities including jailbreaks and tool-abuse attacks.",
	}

	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Run prompt injection test suites against a target",
		RunE:  runScan,
	}

	scanCmd.Flags().StringP("target", "t", "", "Target endpoint to scan")
	scanCmd.Flags().StringP("output", "o", "text", "Output format: text, json")
	scanCmd.Flags().StringP("suite", "s", "all", "Test suite to run: jailbreak, tool-abuse, all")
	scanCmd.Flags().IntP("concurrency", "c", 5, "Number of concurrent requests")
	scanCmd.Flags().Duration("timeout", 30*time.Second, "HTTP request timeout")
	scanCmd.Flags().String("prompt-field", "prompt", "JSON field name for the prompt in requests")
	scanCmd.Flags().String("response-field", "response", "JSON field name for the response in replies")
	scanCmd.Flags().String("api-key", "", "LLM API key for judge detection (or set ANTHROPIC_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY)")
	scanCmd.Flags().String("provider", "", "LLM judge provider: anthropic, openai, gemini (auto-detected if omitted)")
	scanCmd.Flags().String("model", "", "Override default model for the chosen provider")
	scanCmd.Flags().String("config", "promptarmor.yaml", "Path to configuration file")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of promptarmor",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("promptarmor %s\n", version)
		},
	}

	rootCmd.AddCommand(scanCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runScan(cmd *cobra.Command, args []string) error {
	// Load configuration file.
	configPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config %s: %w", configPath, err)
	}

	// Read flag values.
	target, _ := cmd.Flags().GetString("target")
	output, _ := cmd.Flags().GetString("output")
	suite, _ := cmd.Flags().GetString("suite")
	concurrency, _ := cmd.Flags().GetInt("concurrency")
	timeout, _ := cmd.Flags().GetDuration("timeout")
	promptField, _ := cmd.Flags().GetString("prompt-field")
	responseField, _ := cmd.Flags().GetString("response-field")
	apiKey, _ := cmd.Flags().GetString("api-key")
	provider, _ := cmd.Flags().GetString("provider")
	model, _ := cmd.Flags().GetString("model")

	// Merge: flag (if explicitly set) > config file > cobra default.
	mergeConfig(cmd, &cfg, map[string]func(){
		"target":         func() { target = cfg.Target },
		"output":         func() { output = cfg.Output },
		"suite":          func() { suite = cfg.Suite },
		"concurrency":    func() { concurrency = cfg.Concurrency },
		"prompt-field":   func() { promptField = cfg.PromptField },
		"response-field": func() { responseField = cfg.ResponseField },
		"api-key":        func() { apiKey = cfg.APIKey },
		"provider":       func() { provider = cfg.Provider },
		"model":          func() { model = cfg.Model },
	})

	// Timeout needs special handling since it's a string in the config.
	if !cmd.Flags().Changed("timeout") && cfg.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Timeout); err == nil {
			timeout = d
		}
	}

	if target == "" {
		return fmt.Errorf("--target is required")
	}

	// API key: flag > config file > provider-specific env var > env vars in order.
	if apiKey == "" {
		apiKey = resolveAPIKeyFromEnv(provider)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	color.Cyan("promptarmor v%s", version)
	color.Cyan("Target:      %s", target)
	color.Cyan("Suite:       %s", suite)
	color.Cyan("Concurrency: %d", concurrency)

	s, err := scanner.New(scanner.Config{
		Target:        target,
		Suite:         suite,
		Concurrency:   concurrency,
		Timeout:       timeout,
		PromptField:   promptField,
		ResponseField: responseField,
		APIKey:        apiKey,
		Provider:      provider,
		Model:         model,
	})
	if err != nil {
		return err
	}

	if s.UsingJudge() {
		color.Cyan("Detection:   LLM judge (%s)", s.JudgeName())
	} else {
		color.Yellow("Detection:   pattern matching (set ANTHROPIC_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY for LLM judge)")
	}
	fmt.Println()

	results, err := s.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println()
	switch output {
	case "json":
		if err := report.WriteJSON(os.Stdout, results); err != nil {
			return fmt.Errorf("write JSON report: %w", err)
		}
	default:
		report.WriteText(os.Stdout, results)
	}

	if report.HasFailures(results) {
		os.Exit(1)
	}
	return nil
}

// mergeConfig applies config file values for flags that were not explicitly set
// by the user. For each flag name, if the flag was not changed and the config
// has a non-zero value, the apply function overwrites the local variable.
func mergeConfig(cmd *cobra.Command, cfg *config.FileConfig, fields map[string]func()) {
	for flag, apply := range fields {
		if cmd.Flags().Changed(flag) {
			continue
		}
		// Check if the config value is non-zero before applying.
		switch flag {
		case "target":
			if cfg.Target == "" {
				continue
			}
		case "output":
			if cfg.Output == "" {
				continue
			}
		case "suite":
			if cfg.Suite == "" {
				continue
			}
		case "concurrency":
			if cfg.Concurrency == 0 {
				continue
			}
		case "prompt-field":
			if cfg.PromptField == "" {
				continue
			}
		case "response-field":
			if cfg.ResponseField == "" {
				continue
			}
		case "api-key":
			if cfg.APIKey == "" {
				continue
			}
		case "provider":
			if cfg.Provider == "" {
				continue
			}
		case "model":
			if cfg.Model == "" {
				continue
			}
		default:
			continue
		}
		apply()
	}
}

// resolveAPIKeyFromEnv checks environment variables for an API key.
// If provider is specified, it checks only the matching env var.
// Otherwise it tries all three in order.
func resolveAPIKeyFromEnv(provider string) string {
	envVars := map[string]string{
		"anthropic": "ANTHROPIC_API_KEY",
		"openai":    "OPENAI_API_KEY",
		"gemini":    "GEMINI_API_KEY",
	}

	if provider != "" {
		if envVar, ok := envVars[provider]; ok {
			return os.Getenv(envVar)
		}
		return ""
	}

	// Try all in order.
	for _, envVar := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY"} {
		if key := os.Getenv(envVar); key != "" {
			return key
		}
	}
	return ""
}
