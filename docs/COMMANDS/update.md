# `update` command

## Purpose

`update` incrementally refreshes an existing compatible index from current
Git-tracked files.

## Usage

```text
code-index update [--db DB] [--max-bytes N] [--ignore-dir NAME]...
                  [--adopt] [-v|--verbose] [--format text|json] [ROOT]
```

Build/configuration flags have the same meaning as `rebuild`.

- `--adopt` permits an index from another checkout path or unknown Git history
  to become owned by this checkout.
- `--adopt` does not bypass schema, file-source, hash, FTS, or indexing-config
  incompatibility.

## Behavior

The database must already exist, commonly after `init` or `rebuild`. Update
acquires the index lock, checks stored build identity, and applies additions,
changed files, encoding-status transitions, and removals in one SQLite `begin
immediate` transaction. Git renames are represented as deletion plus addition.

An active lock causes a successful skip with the same stderr/JSON convention
as `rebuild`. Attempts after root/database resolution are recorded in the log
sidecar.

## Compatibility

Update requires matching schema version, `git-tracked` file source, SHA-256
hash algorithm, FTS5 availability, maximum bytes, ignored directories, and
encoding fallback list. A mismatch returns an error directing the caller to
`rebuild`.

Different checkout roots and unknown Git histories require `--adopt`.

## Output

Successful text reports `added_files`, `updated_files`, `deleted_files`,
symbols extracted from added/updated files, encoding counts, hash algorithm,
and FTS5 support.

JSON fields are:

| Field | Type when updated | Locked skip |
| --- | --- | --- |
| `operation` | `"update"` | `"update"` |
| `db`, `root` | string | string |
| `skipped` | false | true |
| `reason` | null | `"locked"` |
| `added_files`, `updated_files`, `deleted_files` | number | null |
| `symbols` | number | null |
| `hash_algorithm` | string | null |
| `fts5` | boolean | null |
| `transcoded_files`, `encoding_skipped_files` | number | null |
| `diagnostics` | array when verbose diagnostics exist | omitted |

Each diagnostic contains `path`, `status`, `reason`, and nullable
`source_encoding`/`encoding_source`.

## Errors and examples

Missing database, incompatible metadata, unsafe ownership/history, invalid
configuration, Git failure, or SQLite failure is an error.

```sh
code-index update
code-index update --format json /path/to/repo
code-index update --adopt --db /tmp/shared.sqlite /path/to/repo
```

## Related commands

Use `status` before choosing normal update, adoption, or rebuild. Use `init` for
project setup and `logs` for history.

## Open Questions

- Help currently describes update as able to "create" an index and omits
  `--ignore-dir`; the implementation requires an existing database.
- Whether update should expose unchanged or candidate file counts.
