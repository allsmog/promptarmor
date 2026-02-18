package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/shayaun-nejad/promptarmor/internal/scanner"
)

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
}

// WriteJSON writes a JSON report to w.
func WriteJSON(w io.Writer, results []scanner.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
