# ADR-012: Structured Output and Shadow Logging

## Status
Accepted

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

Business logic (`internal/*`) MUST NOT print directly to `os.Stdout` or `fmt.Println`.
Instead, it must pass data to a `ui.Renderer`.

The `ui.Renderer` is responsible for **Shadow Logging**:
1.  **Render**: Formats the data for the User (Stream 1).
2.  **Record**: Logs the raw data struct to the Audit Log (Stream 3).

**Example:**

```go
// Internal Logic
user := User{ID: 1, Name: "Alice"}
ui.RenderSuccess(cmd.OutOrStdout(), "User created", user)
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
- **Boilerplate**: Requires passing a `Renderer` or using `internal/ui` helper functions instead of simple `fmt.Println`.

## Implementation

- `internal/logger` defaults console output to `os.Stderr`.
- `internal/ui` provides `RenderSuccess`, `RenderTable`, `RenderError`.
- Commands delegate final output to `internal/ui`.
