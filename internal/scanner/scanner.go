package scanner

// Result holds the outcome of a single test case.
type Result struct {
	TestName string `json:"test_name"`
	Suite    string `json:"suite"`
	Passed   bool   `json:"passed"`
	Detail   string `json:"detail,omitempty"`
}

// Scanner runs prompt injection test suites against a target.
type Scanner struct {
	Target string
	Suite  string
}

// New creates a Scanner for the given target and suite.
func New(target, suite string) *Scanner {
	return &Scanner{Target: target, Suite: suite}
}

// Run executes the selected test suites and returns results.
func (s *Scanner) Run() ([]Result, error) {
	// TODO: implement scanning engine
	return nil, nil
}
