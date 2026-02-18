package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fatih/color"
	"github.com/shayaun-nejad/promptarmor/internal/report"
	"github.com/shayaun-nejad/promptarmor/internal/scanner"
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
	scanCmd.Flags().String("api-key", "", "Anthropic API key for LLM judge detection (or set ANTHROPIC_API_KEY)")

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
	target, _ := cmd.Flags().GetString("target")
	output, _ := cmd.Flags().GetString("output")
	suite, _ := cmd.Flags().GetString("suite")
	concurrency, _ := cmd.Flags().GetInt("concurrency")
	timeout, _ := cmd.Flags().GetDuration("timeout")
	promptField, _ := cmd.Flags().GetString("prompt-field")
	responseField, _ := cmd.Flags().GetString("response-field")
	apiKey, _ := cmd.Flags().GetString("api-key")

	if target == "" {
		return fmt.Errorf("--target is required")
	}

	// API key: flag > env var.
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	color.Cyan("promptarmor v%s", version)
	color.Cyan("Target:      %s", target)
	color.Cyan("Suite:       %s", suite)
	color.Cyan("Concurrency: %d", concurrency)

	s := scanner.New(scanner.Config{
		Target:        target,
		Suite:         suite,
		Concurrency:   concurrency,
		Timeout:       timeout,
		PromptField:   promptField,
		ResponseField: responseField,
		APIKey:        apiKey,
	})

	if s.UsingJudge() {
		color.Cyan("Detection:   LLM judge (Claude Haiku)")
	} else {
		color.Yellow("Detection:   pattern matching (set ANTHROPIC_API_KEY for LLM judge)")
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
