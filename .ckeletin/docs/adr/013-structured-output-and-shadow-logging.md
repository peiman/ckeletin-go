# ADR-012: Structured Output and Shadow Logging

## Status
Accepted (Updated: 2026-11-22 - Command path, output flag, execute entry point, exit codes, text writers)

## Context

CLI applications operate with three distinct audiences, often conflated into two streams (stdout/stderr):
1.  **The User (Interactive)**: Needs concise, formatted, often colored output (tables, success messages) on `stdout`.
2.  **The Operator (Status)**: Needs progress updates ("downloading...", "connecting...") on `stderr` to avoid polluting piped data.
3.  **The Auditor (File/Debug)**: Needs raw, structured data (JSON) of what occurred and what was returned, regardless of display format.

**The Problem:**
- Writing output directly to `stdout` (e.g., `fmt.Println`) leaves the log file "blind" to the final result.
- Writing output to logs makes it hard for users to read (JSON pollution in console) or breaks shell pipelines (text pollution in `stdout`).
- There is often no link between what the user saw and what the logs recorded.

## Decision

We adopt a **3-Stream Output Model** using a **UI Adapter Pattern** with **Shadow Logging**.

### 1. The Three Streams

| Stream | Description | Destination | Format | Use Case |
| :--- | :--- | :--- | :--- | :--- |
| **Data** | The "Product" | `stdout` | Table, Text, JSON (User pref) | The actual result of the command. |
| **Status** | The "Process" | `stderr` | Human-readable Log | "Doing X...", Warnings, Errors. |
| **Audit** | The "Trace" | `Log File` | JSON (Structured) | Full request/response data, invisible to user. |

### 2. The UI Adapter (Shadow Logging)

Business logic (`internal/*`) MUST NOT print directly to `os.Stdout` or `fmt.Println`, and it MUST NOT import `.ckeletin/pkg/output`.
It reports progress through `progress.Reporter` (`Step`, `Status`, `Warn`, `Done`). The command layer implements that port with `output.NewReporter` and renders the returned data.

The `ui.Renderer` is responsible for **Shadow Logging**:
1.  **Render**: Formats the data for the User (Stream 1).
2.  **Record**: Logs the raw data struct to the Audit Log (Stream 3).

**Example:**

```go
// Business logic reports. It does not choose a stream.
user := User{ID: 1, Name: "Alice"}
rep.Step("creating user")
return user, nil

// The command layer renders.
return ui.RenderSuccess(cmd.OutOrStdout(), "User created", user)
```

**Result:**
- **Stdout**: `✅ User created: Alice`
- **Stderr**: (Empty, or previous status logs)
- **Log File**: `{"level":"info", "msg":"User created", "data":{"ID":1, "Name":"Alice"}}`

## Consequences

### Positive
- **Pipeline Safety**: `stdout` is never polluted by logs; `stderr` is used for status/logs.
- **Full Observability**: The log file contains the *actual data* returned to the user, not just "success".
- **Separation of Concerns**: Business logic generates data; UI layer formats it.
- **Consistency**: All commands output in a standard way.

### Negative
- **Boilerplate**: The command passes a `progress.Reporter` and renders the returned value. Business logic does not call `fmt.Println`.

## Enforcement

Output patterns are validated automatically via static analysis:

**1. Validation Script**
- `scripts/validate-output-patterns.sh` checks source code patterns
- Detects direct `fmt.Print*` and `os.Stdout` usage in `internal/*`
- Allows the rendering layer (`internal/ui`, `internal/progress`, `internal/logger`, `internal/testutil`) and `*_test.go`
- Requires `.ckeletin-output-allowlist` at the repository root. A listed path, with a reason, may write directly. A missing or malformed file fails on its own. A new `fmt.Println` that is not listed fails. `task check` runs the script, including in a derived project whose check binary does not.

**2. Task Command**
```bash
task validate:output  # Checks output patterns
task check           # Includes output validation
```

**3. What Gets Validated**
- Business logic reports through `progress.Reporter` and does not import `.ckeletin/pkg/output`
- The command layer and presentation packages render the Data Stream
- Business logic packages use the logger for the Status Stream's audit copy
- Fails on direct `fmt.Print*` in `internal/*` (except ui/)
- Fails on direct `os.Stdout` writes in `internal/*` (except ui/)

**4. Integration**
- **Local**: Part of `task check` (before every commit)
- **CI**: Runs in quality gate pipeline
- **Pre-commit**: Via lefthook hooks

## Implementation

- `internal/logger` defaults console output to `os.Stderr`.
- `internal/ui` provides `RenderSuccess`, `RenderTable`, `RenderError`.
- Commands delegate final output to `internal/ui` or to `.ckeletin/pkg/output`.

## Framework output API (0.2.12)

The reference `main` and `cmd/root.go` call this package. A derived application does the same instead of copying the wiring.

### Command identifier

The envelope field `command` is the full path without the program name, words separated by one space: `check model`. A top-level command is its name. `--version` uses `version`. `SetCommandName` still overrides the path. A failure before any command resolves marshals `command` as JSON null. A failure after a parent resolved uses that parent's path (`check` when `check nope` does not resolve).

### One mode resolver

`RegisterOutputFlag(root, WithEnvPrefix(prefix), WithAliasFlag("format"))` registers persistent `--output text|json` and resolves the mode in `PersistentPreRunE` before the application's own pre-run, then again after the config file is read and before `logger.Init`.

Precedence is the changed `--output` flag, then a changed alias flag, then the environment variable, then the config file, then the default `text`. `WithAliasFlag` does not register the alias. It recognises a flag the command already has. An invalid value (`yaml`) returns one error. The exit code is `output.ExitInvalidOutput` (2) for every command.

### One envelope per run

`Execute(root) int` runs the command and returns the process exit code. Call `SetAfterRun(logger.Cleanup)` so the audit file stays open until the error document is written.

- Success exits 0.
- JSON mode, when the command succeeded and wrote no envelope, writes one success envelope with `data` null and `error` null. A command annotated `ckeletin:out-002-exempt` (the value is the reason) prints its native output instead. `EnvelopeExemptions` lists those commands. An empty reason does not exempt anything.
- JSON mode, on an error the command did not already render, writes one error envelope to stdout and nothing to stderr.
- Text mode writes `Error: …` to stderr and leaves stdout empty.
- `ErrRendered` and `Rendered(code)` mean the command already wrote the document. Execute does not write a second one.
- `Rendered(code)` wraps `ErrRendered`. `errors.Is(Rendered(code), ErrRendered)` is true.
- `ExitCoder` (`ExitCode() int`) is honoured for any error. A plain error exits 1.
- Exit code 0 is not returned together with `status: "error"`. A non-zero code is not returned for a success. `Rendered(0)` and an `ExitCoder` of 0 both exit 1.

The failure is logged at debug, so the audit file receives it, before cleanup.

### Text writers

`Result(cmd)` writes stdout. `Status(cmd)` writes stderr. Both honour `SetOut` and `SetErr`. `Resultf`, `Statusf`, and `Success` are the helpers. `Progressf` writes a status line and strips ANSI control codes when stderr is not a terminal.

Every successful write is shadow-logged to the audit stream afterwards, with `command` and `stream` (`stdout` or `stderr`). A failed write is not logged.

JSON mode: `Result` and `Success` succeed without writing, so stdout stays the envelope. `Status` is silenced, matching the console logger.

Business logic still must not call `fmt.Print*` (see Enforcement). Commands in `cmd/` use these writers.
