# `logs` command

## Purpose

`logs` reads recent index-operation history from `<index-db>.logs.sqlite`.

## Usage

```text
code-index logs [--root ROOT|--db DB] [--limit N] [--format text|json]
```

- `--limit` defaults to 20 and must be positive.
- `--root`/`--db` select the main index path; logs then use its sidecar path.
- Positional arguments are rejected.

An absent log sidecar is successful: text prints only headers and JSON returns
`[]`. Rows are newest first.

## Output

Text columns and JSON fields are:

| Field | Type | Meaning |
| --- | --- | --- |
| `id` | number | Sidecar-local increasing row ID |
| `operation` | string | `init`, `rebuild`, or `update` |
| `status` | string | `succeeded`, `failed`, or `skipped` |
| `root` | string | Resolved operation root |
| `db` | string | Main database path |
| `started_at` | string | UTC RFC3339 timestamp |
| `finished_at` | string | UTC RFC3339 timestamp |
| `error` | string or null | Error message for failed runs |

The sidecar retains the newest 1,000 rows. Logging failure during mutation is a
warning and does not change the mutation result.

## Examples

```sh
code-index logs
code-index logs --limit 50 --format json
code-index logs --db /tmp/index.sqlite
```

## Related commands

`status` reports the current database/lock state. `rebuild` and `update` create
most log rows.

## Open Questions

- Whether pruning, filtering by operation/status, or explicit sidecar
  diagnostics are needed beyond the current fixed retention and limit.
