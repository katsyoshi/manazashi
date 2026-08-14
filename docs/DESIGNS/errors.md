# Structured CLI errors

## Motivation

Commands already support machine-readable success output through
`--format json`, but failures are currently written as plain text. A caller
cannot reliably distinguish an absent index from an incompatible index or a
general command failure without parsing human-readable messages.

## Output contract

When a command accepts `--format json` or `--format=json` and subsequently
fails, `mzci` writes one JSON document to stderr and exits non-zero:

```json
{
  "error": {
    "code": "index_not_found",
    "message": "index not found: /tmp/index.sqlite; run init or rebuild first, or pass --db",
    "details": {
      "db": "/tmp/index.sqlite"
    }
  }
}
```

`code` is a stable machine-readable category. `message` preserves the useful
human-readable error. `details` is optional and contains typed context that is
safe for callers to branch on. Unknown or not-yet-classified failures use
`command_failed`; callers must not parse `message` to create additional error
categories.

Text-format failures retain their existing one-line stderr output. Successful
stdout shapes and process exit statuses do not change. A JSON failure is
written to stderr so stdout remains reserved for successful command results.

The top-level error renderer recognizes the last `--format` selection in the
argument vector up to `--`. Flag-package errors that terminate parsing before
the format is accepted retain the flag package's existing behavior.

## Initial error codes

- `index_not_found`: the resolved index database does not exist. Details
  include `db`.
- `command_failed`: the failure has no more specific stable classification.

Additional codes should be added only when callers have a concrete recovery
decision to make. Adding a code is an output-contract change and requires
tests for both its JSON representation and unchanged text message.

## Non-goals

- Assigning a unique code to every internal error
- Returning a successful exit status for structured errors
- Moving warnings or successful results from their existing streams
- Encoding stack traces or nested Go error chains in public output
- Replacing command-specific diagnostic results such as `status` and
  `files --explain`
