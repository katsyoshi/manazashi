# `defs` command

## Purpose

`defs` lists stored symbol definitions or searches for likely definitions.

## Usage

```text
mzci defs [--root ROOT|--db DB] [--kind KIND]
                [--language LANG] [--limit N] [--list]
                [--format text|json] [QUERY]
```

- Without `--list`, exactly one `QUERY` is required.
- With `--list`, `QUERY` is forbidden.
- `--kind` and `--language` are exact stored-value filters.
- `--limit` defaults to 50.
- `--format` defaults to `text`.

## Matching and ordering

Query mode matches case-insensitively when the stored name equals the query,
the name starts with the query, the signature contains it, or the path contains
it. Exact-name rows sort first, then name-prefix rows, then other matches;
ties sort by path and line.

List mode does not search. It sorts by path, line, column, and name.

No results is successful. Definitions are static navigation candidates and may
share a name or be missed by a lightweight extractor.

## Output

Text is a tabular SQLite result. JSON is an array of:

```json
{
  "path": "internal/config.go",
  "line": 42,
  "kind": "function",
  "name": "parseConfig",
  "language": "go",
  "signature": "func parseConfig(...) ..."
}
```

`line` is numeric and one-based. `language` is string or null; the other fields
are strings. Empty results are `[]`.

## Examples

```sh
mzci defs parseConfig
mzci defs --kind method --language ruby --format json save
mzci defs --list --limit 100 --format json
```

Combining list and query, omitting a required query, invalid database/config,
or unsupported format is an error.

## Related commands

Use `outline` for definitions in one file, `refs` for lexical occurrences, and
`show` for surrounding source.

## Open Questions

- Public help omits `--limit`.
- Unlike `refs` and `logs`, `defs` does not require a positive limit; zero
  returns no rows and SQLite treats a negative limit as unbounded.
