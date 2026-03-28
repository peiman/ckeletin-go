# Task Commands Are Mandatory

ALWAYS use `task` commands for standard workflows. NEVER use raw go/lint/format commands directly.

| Instead of (NEVER) | Use (ALWAYS) | Why |
|--------------------|--------------|-----|
| `go test ./...` | `task test` | Runs coverage, gotestsum, correct flags |
| `go build ./...` or `go build` | `task build` | Correct build tags and flags |
| `golangci-lint run` | `task lint` | Correct timeout and config |
| `goimports -w .` or `gofmt` | `task format` | Handles all formatting consistently |
| `go vet ./...` | `task lint` | Included in lint task |
| `go mod tidy` | `task tidy` | Ensures consistency |
| Running multiple checks manually | `task check` | Runs ALL checks in correct order |

**The ONLY acceptable raw `go` command:** `go test -v -run TestName ./path/...` for debugging a specific test.

Always return to `task` commands for final validation. Direct commands are fine for exploration, but finish with `task check`.

Task commands ensure all flags, coverage settings, and checks are applied consistently. This is the single most common rule violation in this project — pay attention.
