package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/allsmog/promptarmor/internal/scanner"
)

// JSONReport wraps results with summary statistics.
type JSONReport struct {
	Total   int              `json:"total"`
	Passed  int              `json:"passed"`
	Failed  int              `json:"failed"`
	Results []scanner.Result `json:"results"`
}

// WriteText writes a human-readable report to w.
func WriteText(w io.Writer, results []scanner.Result) {
	if len(results) == 0 {
		color.New(color.FgYellow).Fprintln(w, "No results.")
		return
	}
	for _, r := range results {
		status := color.GreenString("PASS")
		if !r.Passed {
			status = color.RedString("FAIL")
		}
		fmt.Fprintf(w, "[%s] [%s] %s\n", status, r.Suite, r.TestName)
		if r.Detail != "" {
			fmt.Fprintf(w, "       %s\n", r.Detail)
		}
	}

	passed, failed := tally(results)
	fmt.Fprintf(w, "\nResults: %d/%d passed, %d/%d failed\n", passed, len(results), failed, len(results))
}

// WriteJSON writes a JSON report with summary envelope to w.
func WriteJSON(w io.Writer, results []scanner.Result) error {
	passed, failed := tally(results)
	report := JSONReport{
		Total:   len(results),
		Passed:  passed,
		Failed:  failed,
		Results: results,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// HasFailures returns true if any result failed.
func HasFailures(results []scanner.Result) bool {
	for _, r := range results {
		if !r.Passed {
			return true
		}
	}
	return false
}

func tally(results []scanner.Result) (passed, failed int) {
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}
	return
}
