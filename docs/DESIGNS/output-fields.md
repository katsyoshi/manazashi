# Navigation output fields

## Motivation

Navigation results are read by agents as context, and every row repeats its
fields. The [Rails context-size benchmark](../benchmark.md#rails-context-size)
showed that fields a caller already knows or rarely needs dominate several
results: `outline` repeats the requested `path` on every row, and `files`
spends most of its output on encoding metadata. A `jq` projection of the same
results reduced `outline` from 3,280 to 1,278 characters and `files` from 790
to 201.

The README design states that agents retrieve compact candidates first and
read deeper only when a task needs it. Field selection follows that rule.

## Default and `--details`

Navigation commands return a minimal field set by default. `--details` returns
the full field set that the command returned before this change.

Field selection is independent of `--format`. The output format does not
imply whether a person or a program is reading it: agents read text, and
people pipe JSON into other tools. Text and JSON therefore select the same
fields, and `--details` applies to both.

| Command | Default fields | Added by `--details` |
| --- | --- | --- |
| `defs` | `path`, `line`, `kind`, `name`, `signature` | `language` |
| `outline` | `line`, `kind`, `name`, `signature` | `path`, `language` |
| `refs` definitions | `path`, `line`, `kind`, `name`, `signature` | `language` |
| `refs` candidates | grouped by `path`; `[line, text]` pairs | `language` (flat rows) |
| `files` | `path` | `language`, `size`, `status`, `source_encoding`, `encoding_source`, `transcoded`, `skip_reason` |

The `refs` `query` object stays in the default output. It records the query
semantics that make `refs` an object rather than an array, as described in the
[CLI contract](cli.md).

`files --status skipped` and `files --status all` return the full field set
without `--details`. Their purpose is to explain why a file was omitted, which
requires `skip_reason` and the encoding fields. `files --explain` is unchanged.

`show`, which already returns compact JSON, and the inspection commands are
unchanged.

## Shapes

`defs`, `outline`, and `files` remain arrays of objects, and `refs` remains an
object. Omitted fields are absent from JSON objects, not set to `null`, so a
present `null` still means an unknown value. Text output drops the same
columns.

### Compact `refs` candidates

Reference candidates dominate `refs` output, and the flat rows repeat each
path once per line. By default JSON `candidates` groups lines by file:

```json
"candidates": [
  {
    "path": "actionview/lib/action_view/helpers/output_safety_helper.rb",
    "lines": [[43, "options.assert_valid_keys(:words_connector, ...)"]]
  }
]
```

- Groups follow path order and lines follow line order, as before. An array
  is used rather than an object keyed by path so the order is explicit.
- Each `lines` entry is a two-element array: the one-based line number, then
  the line text with leading and trailing whitespace removed, matching the
  text output. Use `show` when indentation matters.
- `--details` returns the previous flat candidate objects with `path`, `line`,
  `language`, and untrimmed `text`.
- Text output already groups candidates by file and is unchanged.

For the Rails `assert_valid_keys` query (31 candidates in 14 files) this
reduces the JSON from 5,621 to an estimated 4,190 characters, below the 4,526
characters of `rg -n -w`.

## Compatibility

This changes the default output contract. Callers that read the removed
fields must add `--details`. The bundled skill reads only default fields.

## Non-goals

- Choosing defaults from `--format`, the environment, or configuration
- Changing which rows are returned, ordering, or limits
- Reducing `show` output

## Open Questions

- Selecting individual fields with a `+FIELD` notation that adds one field to
  the default set, for example `+size`. The flag form and whether a `-FIELD`
  counterpart is needed are undecided; this design reserves the notation only.
- Reference candidates remain the largest part of `refs` output even when
  grouped. A representation that holds references more compactly, for example
  by summarizing per file or per enclosing definition before returning lines,
  needs its own design.
- The default `defs` limit of 50 now dominates `defs` output size; lowering it
  or ranking exact matches first is a separate decision.
