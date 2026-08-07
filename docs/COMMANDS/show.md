# `show` command

## Purpose

`show` returns indexed source around a one-based line number. It reads the
database snapshot, not the live checkout.

## Usage

```text
mzci show [--root ROOT|--db DB] --line N [--context N]
                [--format text|json] PATH
```

- `PATH` is required.
- `--line` is required and must be positive.
- `--context` defaults to 3 and selects that many lines before and after the
  target.

## Path and range selection

An exact normalized repository-relative path wins. Otherwise a stored path
ending with the supplied text may match; the shortest path is selected.
The start is clamped to line 1. Rows sort by line.

An unknown path or a requested range outside stored source is successful and
returns no rows.

## Output

Text is tabular. JSON is an array of:

```json
{"path":"cmd.go","line":51,"text":"func Run(args []string) error {"}
```

`line` is numeric; empty results are `[]`. Lock warnings use stderr.

## Examples

```sh
mzci show --line 42 cmd.go
mzci show --line 42 --context 8 --format json cmd.go
```

## Related commands

Select positions with `defs`, `outline`, or `refs`. Use `files` to resolve
paths.

## Open Questions

- `--context` does not currently reject negative values.
- Equally short suffix matches have no final deterministic tie-breaker, and
  `show` suffix matching is broader than `outline` matching.
