# `stats` command

## Purpose

`stats` reports index-wide metadata and corpus counts from one database.

## Usage

```text
mzci stats [--root ROOT|--db DB] [--format text|json]
```

Positional arguments are rejected. `--format` defaults to `text`.

## Output

Text is a `key<TAB>value` sequence. JSON is one object:

| Field | Type |
| --- | --- |
| `root`, `file_source`, `indexed_at`, `updated_at`, `last_operation` | string or null |
| `vcs_kind`, `vcs_revision`, `vcs_ref`, `vcs_head`, `vcs_branch` | string or null |
| `schema_version` | number or null |
| `vcs_dirty`, `fts5` | boolean or null |
| `files`, `skipped_files`, `symbols`, `lines` | number |
| `code_lines`, `comment_lines`, `blank_lines` | number |
| `hash_algorithm` | string or null |

`files` counts `index_status = indexed`; `skipped_files` counts retained
encoding failures. Line and symbol totals cover searchable rows. Code/comment/
blank counts are sums of `file_metrics`.

Stats describes the last completed database snapshot. It does not inspect the
current checkout or lock compatibility.

## Examples

```sh
mzci stats
mzci stats --root /path/to/repo --format json
```

## Related commands

Use `status` for freshness and compatibility, `metrics` for language/file
breakdowns, and `schema` for storage shape.

## Open Questions

- Whether stats should include component state, database byte size, or FTS row
  counts without duplicating status/schema responsibilities.
