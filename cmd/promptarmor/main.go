package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			output, _ := cmd.Flags().GetString("output")
			suite, _ := cmd.Flags().GetString("suite")

			if target == "" {
				return fmt.Errorf("--target is required")
			}

			color.Cyan("Scanning target: %s", target)
			color.Cyan("Output format:   %s", output)
			color.Cyan("Test suite:      %s", suite)
			color.Yellow("\nScan engine not yet implemented.")
			return nil
		},
	}

	scanCmd.Flags().StringP("target", "t", "", "Target endpoint or config file to scan")
	scanCmd.Flags().StringP("output", "o", "text", "Output format: text, json")
	scanCmd.Flags().StringP("suite", "s", "all", "Test suite to run: jailbreak, tool-abuse, all")

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
