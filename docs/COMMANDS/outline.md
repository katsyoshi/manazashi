# `outline` command

## Purpose

`outline` returns stored symbol definitions for one indexed file in source
order. Ruby `_spec.rb` files also include RSpec `test_group` and `test_example`
rows for recognized block calls. See [RSpec outline entries](../DESIGNS/rspec-outline.md).

## Usage

```text
mzci outline [--root ROOT|--db DB] [--details] [--format text|json] PATH
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

Text is tabular with the same fields as JSON. Every row belongs to the selected
file, so `path` is omitted by default:

```json
[
  {
    "line": 10,
    "kind": "type",
    "name": "Config",
    "signature": "type Config struct {"
  }
]
```

`--details` returns the full `defs` definition object, adding `path` and
`language`. See [Navigation output fields](../DESIGNS/output-fields.md).
`language` is nullable; empty results are `[]`.

## Examples

```sh
mzci outline internal/config.go
mzci outline --format json config.go
```

## Related commands

Use `defs` to search across files, `refs` for occurrences, and `show` to read
source.

## Open Questions

- When equally short suffix matches exist, SQL does not define a final
  tie-breaker. Whether ambiguous suffixes should error or list candidates is
  undecided.
- Whether qualified names or explicit ownership are needed in outlines.
