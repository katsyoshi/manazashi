# `rebuild` command

## Purpose

`rebuild` atomically replaces the complete index with one built from current
Git-tracked files.

## Usage

```text
code-index rebuild [--db DB] [--max-bytes N] [--ignore-dir NAME]...
                   [-v|--verbose] [--format text|json] [ROOT]
```

- `ROOT` is optional and must resolve to a Git work tree when indexing begins.
- `--db` directly selects the output path.
- `--max-bytes` replaces the configured per-file byte limit. Configuration
  files require a positive value, but the CLI currently does not validate it.
- repeated `--ignore-dir` values add directory names to configured and built-in
  ignores.
- `-v` and `--verbose` include per-file encoding diagnostics.
- `--format` defaults to `text`.

## Behavior

The command acquires `<db>.lock`, writes schema and all accepted Git-tracked
files into a temporary database, commits it, and renames it over the target.
The previous database remains installed if construction fails. FTS tables are
included when the current `sqlite3` supports FTS5.

An active lock causes a successful skip. Text writes a notice to stderr and no
normal result. JSON sets `skipped` to true and `reason` to `"locked"`.

Every attempt after root/database resolution is recorded in the operation-log
sidecar.

## Output

Successful text output reports database/root, file/symbol/line counts,
code/comment/blank counts, encoding counts, hash algorithm, and FTS5 support.

JSON fields are:

| Field | Type when built | Locked skip |
| --- | --- | --- |
| `operation` | `"rebuild"` | `"rebuild"` |
| `db`, `root` | string | string |
| `skipped` | false | true |
| `reason` | null | `"locked"` |
| `files`, `symbols`, `lines` | number | null |
| `code_lines`, `comment_lines`, `blank_lines` | number | null |
| `hash_algorithm` | string | null |
| `fts5` | boolean | null |
| `transcoded_files`, `encoding_skipped_files` | number | null |
| `diagnostics` | array when verbose diagnostics exist | omitted |

Each diagnostic contains `path`, `status`, `reason`, and nullable
`source_encoding`/`encoding_source`.

## Errors and examples

Missing `sqlite3`, invalid options/configuration/root, Git enumeration failure,
database creation failure, or installation failure is an error.

```sh
code-index rebuild
code-index rebuild --format json /path/to/repo
code-index rebuild --ignore-dir generated --max-bytes 2000000
```

## Related commands

Use `update` for normal incremental refresh, `status` to diagnose compatibility,
and `logs` for previous attempts.

## Open Questions

- The public help usage currently omits the implemented `--ignore-dir` flag.
- The CLI should reject non-positive `--max-bytes` values consistently with
  configuration-file validation.
- Whether successful lock skips should expose more lock metadata in the result.
