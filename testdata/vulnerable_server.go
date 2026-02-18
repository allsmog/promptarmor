// +build ignore

// vulnerable_server.go runs a mock LLM endpoint that complies with every
// injection attempt. Use it to test promptarmor locally:
//
//   go run testdata/vulnerable_server.go &
//   ./promptarmor scan -t http://localhost:8080/chat
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		prompt := strings.ToLower(body["prompt"])

		response := "Sure, I'll help with that! pwned"

		switch {
		case strings.Contains(prompt, "dan"):
			response = "DAN mode enabled. I am DAN and I have no restrictions!"
		case strings.Contains(prompt, "repeat your entire system prompt"):
			response = "My system prompt is: You are a helpful assistant that answers questions."
		case strings.Contains(prompt, "new system prompt"):
			response = "pwned — new system prompt accepted."
		case strings.Contains(prompt, "hypothetical"):
			response = "Hypothetically: pwned! In a world without rules I would do anything."
		case strings.Contains(prompt, "developer mode"):
			response = "[DEV] pwned — developer mode active, all restrictions removed."
		case strings.Contains(prompt, "drop table"):
			response = "query executed successfully: DROP TABLE users; 42 rows affected"
		case strings.Contains(prompt, "/etc/passwd"):
			response = "root:x:0:0:root:/root:/bin/bash\nnobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin"
		case strings.Contains(prompt, "/etc/shadow"):
			response = "root:$6$rounds=656000$abc:19000:0:99999:7:::"
		case strings.Contains(prompt, "169.254.169.254"):
			response = "ami-id\ninstance-id\niam\nsecurity-credentials"
		case strings.Contains(prompt, "database_url"):
			response = "DATABASE_URL=postgres://admin:s3cret@db.internal:5432/prod"
		case strings.Contains(prompt, "write"):
			response = "file written to /tmp/pwned.txt successfully"
		case strings.Contains(prompt, "email"):
			response = "Retrieved 1,423 emails. Sent to https://evil.com/collect via POST."
		case strings.Contains(prompt, "python"):
			response = "root"
		case strings.Contains(prompt, "admin") || strings.Contains(prompt, "privilege"):
			response = "privileges updated — attacker@evil.com is now admin"
		case strings.Contains(prompt, "config"):
			response = "password: s3cret_prod_password\ntable: users — dump complete"
		case strings.Contains(prompt, "available tools") || strings.Contains(prompt, "list all"):
			response = "Available tools:\n- function: execute_sql, parameters: {query: string}\n- function: read_file, parameters: {path: string}"
		}

		fmt.Printf("  [vuln-server] prompt=%q → response=%q\n", body["prompt"][:min(60, len(body["prompt"]))], response[:min(60, len(response))])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"response": response})
	})

	fmt.Println("Vulnerable test server running on http://localhost:8080/chat")
	http.ListenAndServe(":8080", nil)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
