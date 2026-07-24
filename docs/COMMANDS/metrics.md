# `metrics` command

## Purpose

`metrics` reports stored code/comment/blank line and symbol counts by language
or by file.

## Usage

```text
code-index metrics [--root ROOT|--db DB] [--language LANG]
                   [--limit N] [--format text|json] [PATH_QUERY]
```

- Without `PATH_QUERY`, rows aggregate by language.
- With one query, matching is a case-insensitive path substring and rows are
  per file.
- `--language` is an exact filter.
- `--limit` defaults to 100.
- More than one positional argument is rejected.

## Ordering and output

Summary rows sort by descending total lines and then language. Missing language
is shown as `(unknown)`. JSON summary rows contain:

```json
{"language":"go","files":20,"lines":2000,"code":1600,"comments":100,"blank":300,"symbols":80}
```

File rows sort by path and contain:

```json
{"path":"cmd.go","language":"go","lines":200,"code":160,"comments":10,"blank":30,"symbols":12}
```

All counts are numbers. File `language` is nullable. Text is tabular and empty
JSON results are `[]`.

Metrics exist only for successfully indexed files; encoding-skipped and
pre-decoding exclusions do not contribute.

## Examples

```sh
code-index metrics
code-index metrics --language go --format json
code-index metrics internal/ --limit 20
```

## Related commands

Use `stats` for index-wide totals and `files` for encoding/status metadata.

## Open Questions

- `--limit` is not validated as positive; zero returns no rows and a negative
  SQLite limit is unbounded.
- Comment counting remains lightweight and language-dependent rather than a
  parser-backed comment-region contract.
