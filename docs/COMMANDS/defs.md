# `defs` command

## Purpose

`defs` lists stored symbol definitions or searches for likely definitions.

## Usage

```text
mzci defs [--root ROOT|--db DB] [--kind KIND]
                [--language LANG] [--path QUERY] [--limit N] [--list]
                [--details] [--format text|json] [QUERY]
```

- Without `--list`, exactly one `QUERY` is required.
- With `--list`, `QUERY` is forbidden.
- `--kind` and `--language` are exact stored-value filters.
- `--path` restricts rows to repository-relative paths containing the query,
  case-insensitively.
- `--limit` defaults to 50.
- `--details` adds `language` to text and JSON output.
- `--format` defaults to `text`.

`QUERY` accepts one query per invocation. To search several symbols, invoke
`defs` separately for each query or use `sql` for a custom bounded search.

## Matching and ordering

Query mode matches the stored symbol name exactly, ignoring case. It does not
search prefixes, signatures, or paths. Use `files` to search paths, `refs` to
find lexical occurrences, or `sql` for a custom bounded symbol search.

List mode does not search. It sorts by path, line, column, and name.

No results is successful. Definitions are static navigation candidates and may
share a name or be missed by a lightweight extractor.

The result is a snapshot of the index. Use `status` to check freshness and
read the live checkout before editing or relying on source details. For
same-named definitions, narrow the result with `--kind`, `--language`, or
`--path`.

## Output

Text is a tabular SQLite result with the same fields as JSON. JSON is an array
of:

```json
{
  "path": "internal/config.go",
  "line": 42,
  "kind": "function",
  "name": "parseConfig",
  "signature": "func parseConfig(...) ..."
}
```

With `--details`, each row also has `language` between `name` and
`signature`. See [Navigation output fields](../DESIGNS/output-fields.md).

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
