# `files` command

## Purpose

`files` lists indexed file metadata or searches repository-relative paths.

## Usage

```text
mzci files [--root ROOT|--db DB] [--language LANG]
                 [--status indexed|skipped|all] [--limit N] [--explain] [--list]
                 [--format text|json] [QUERY]
```

- Query mode requires one path substring; matching is case-insensitive.
- `--list` forbids a query and lists all rows passing filters.
- `--status` defaults to `indexed`.
- `--language` is an exact filter.
- `--limit` defaults to 100.

Rows sort by path. No result is successful.

## Status

- `indexed`: searchable source with line/symbol/metric rows
- `skipped`: a tracked source file retained only as metadata because encoding
  could not be decoded
- `all`: both

Files excluded before encoding, such as oversized or binary files, are not
reported as encoding-skipped rows.

## Output

Text is tabular. JSON rows contain:

| Field | Type | Meaning |
| --- | --- | --- |
| `path` | string | Repository-relative path |
| `language` | string or null | Detected language |
| `size` | number | Original byte size |
| `status` | string | `indexed` or `skipped` |
| `source_encoding` | string or null | Selected source encoding |
| `encoding_source` | string or null | `utf8`, `bom`, `declaration`, or `fallback` |
| `transcoded` | boolean | Whether conversion to UTF-8 occurred |
| `skip_reason` | string or null | Stable encoding failure reason |

Empty JSON results are `[]`.

`--explain` requires JSON output and a path query. It wraps results as
`{"query":"...","found":false,"reason":"untracked","matches":[]}` so a
caller can distinguish missing, untracked, skipped, excluded, and unknown
paths. Without `--explain`, the established array output is unchanged.

## Examples

```sh
mzci files --list
mzci files config --format json
mzci files --status skipped --list --format json
mzci files --explain --format json path/to/file.rb
```

## Related commands

Use `show` for indexed source, `metrics` for per-file counts, and `status` for
index-wide state.

## Open Questions

- Public help omits `--limit`.
- Non-positive limits are not validated consistently; zero returns no rows and
  a negative SQLite limit is unbounded.
- Whether non-encoding exclusions should be retained as file metadata.
