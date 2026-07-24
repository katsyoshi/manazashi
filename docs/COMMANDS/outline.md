# `outline` command

## Purpose

`outline` returns stored symbol definitions for one indexed file in source
order.

## Usage

```text
code-index outline [--root ROOT|--db DB] [--format text|json] PATH
```

`PATH` is required. A leading slash is removed and separators are normalized
for repository-style matching.

## Path selection and ordering

An exact repository-relative path wins. Otherwise a path ending in
`/PATH` may match; the shortest matching stored path is selected. Only one
file is returned. Symbols sort by line, column, and name.

An unknown file or a file with no stored symbols is successful and produces an
empty row set.

## Output

Text is tabular. JSON uses the same definition object as `defs`:

```json
[
  {
    "path": "internal/config.go",
    "line": 10,
    "kind": "type",
    "name": "Config",
    "language": "go",
    "signature": "type Config struct {"
  }
]
```

`language` is nullable; empty results are `[]`.

## Examples

```sh
code-index outline internal/config.go
code-index outline --format json config.go
```

## Related commands

Use `defs` to search across files, `refs` for occurrences, and `show` to read
source.

## Open Questions

- When equally short suffix matches exist, SQL does not define a final
  tie-breaker. Whether ambiguous suffixes should error or list candidates is
  undecided.
- Whether qualified names or explicit ownership are needed in outlines.
