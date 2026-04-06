# ckeletin-go — Cursor Rules

Read [AGENTS.md](AGENTS.md) for the full project guide. This file extracts the critical rules.

## Non-Negotiable Rules

1. **TDD**: Write failing tests FIRST, then implement. Test + implementation go in one atomic commit.
2. **`task check` before every commit** — runs all quality checks.
3. **Commands ≤30 lines** — `cmd/*.go` files wire things together; logic goes in `internal/`.
4. **Use `config.Key*` constants** — never hardcode config strings.
5. **Never reduce test coverage** — 85% minimum, use `testify/assert`.
6. **Check licenses after `go get`** — run `task check:license:source` immediately.

## Command Translation (MANDATORY)

| Instead of | Use |
|-----------|-----|
| `go test ./...` | `task test` |
| `go build` | `task build` |
| `golangci-lint run` | `task lint` |
| `goimports -w .` | `task format` |
| `go mod tidy` | `task tidy` |
| Multiple checks | `task check` |

Only exception: `go test -v -run TestName ./path/...` for debugging a specific test.

## Architecture

```
cmd/           → Ultra-thin command wiring (≤30 lines)
internal/      → Business logic (private)
pkg/           → Public reusable libraries
.ckeletin/     → Framework (config, logging, scripts, ADRs)
```

## JSON Output Mode

Every command supports `--output json` for machine-readable output. Commands using `ui.RenderSuccess` get JSON automatically. For custom shapes, implement `ui.JSONResponder`.

## Full Reference

See [AGENTS.md](AGENTS.md) for: ADRs, config system, logging conventions, testing patterns, commit conventions, and troubleshooting.
