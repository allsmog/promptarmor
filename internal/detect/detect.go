package detect

import "strings"

// Check tests whether the response contains any of the given indicators
// (case-insensitive). It returns true and the matched indicator if injection
// was detected, or false and "" if the response is clean.
func Check(response string, indicators []string) (bool, string) {
	lower := strings.ToLower(response)
	for _, ind := range indicators {
		if strings.Contains(lower, strings.ToLower(ind)) {
			return true, ind
		}
	}
	return false, ""
}
