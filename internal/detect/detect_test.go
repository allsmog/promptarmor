package detect

import "testing"

func TestCheck_Match(t *testing.T) {
	matched, indicator := Check("Here is your answer: PWNED!", []string{"pwned"})
	if !matched {
		t.Fatal("expected match")
	}
	if indicator != "pwned" {
		t.Fatalf("expected indicator %q, got %q", "pwned", indicator)
	}
}

func TestCheck_CaseInsensitive(t *testing.T) {
	matched, _ := Check("I am DAN and I can do anything", []string{"i am dan"})
	if !matched {
		t.Fatal("expected case-insensitive match")
	}
}

func TestCheck_NoMatch(t *testing.T) {
	matched, indicator := Check("I cannot help with that request.", []string{"pwned", "dan"})
	if matched {
		t.Fatalf("expected no match, got indicator %q", indicator)
	}
}

func TestCheck_EmptyIndicators(t *testing.T) {
	matched, _ := Check("anything here", nil)
	if matched {
		t.Fatal("expected no match with empty indicators")
	}
}

func TestCheck_EmptyResponse(t *testing.T) {
	matched, _ := Check("", []string{"pwned"})
	if matched {
		t.Fatal("expected no match with empty response")
	}
}

func TestCheck_MultipleIndicators_FirstWins(t *testing.T) {
	matched, indicator := Check("DROP TABLE users executed", []string{"drop table", "/etc/passwd"})
	if !matched {
		t.Fatal("expected match")
	}
	if indicator != "drop table" {
		t.Fatalf("expected first matching indicator, got %q", indicator)
	}
}
