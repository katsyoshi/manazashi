# `status` command

## Purpose

`status` combines completed index metadata, current checkout state, update
compatibility, component state, and the current lock.

## Usage

```text
code-index status [--root ROOT|--db DB] [--format text|json]
```

- `--db` selects a database directly.
- `--root` selects configuration and the default database and is also used for
  current Git state.
- `--format` defaults to `text`.
- Positional arguments are rejected.

If neither the database nor its lock exists, status fails with initialization
guidance. A lock without a database is a valid status result.

## Text output

Text is a `key<TAB>value` table. It always includes `db`, `exists`, and
`locked`. Available database, current-checkout, compatibility, component, and
lock fields follow. Missing values are omitted; unknown staleness is printed as
`unknown`.

## JSON output

JSON always contains `db`, `exists`, and `locked`. All other fields are
nullable.

| Group | Fields |
| --- | --- |
| Stored identity | `root`, `schema_version`, `file_source`, `hash_algorithm`, `fts5` |
| Stored configuration | `config_max_bytes`, `config_ignore_dirs`, `config_encoding_fallbacks` |
| Stored operation | `indexed_at`, `updated_at`, `last_operation`, `components` |
| Stored VCS | `vcs_kind`, `vcs_head`, `vcs_branch`, `vcs_dirty`, `vcs_dirty_hash`, `vcs_revision`, `vcs_ref` |
| Current VCS | `current_vcs_kind`, `current_vcs_head`, `current_vcs_branch`, `current_vcs_dirty` |
| Compatibility | `update_compatible`, `update_requires_adopt`, `update_rebuild_required`, `update_blocker`, `index_stale` |
| Lock | `lock`, `lock_operation`, `lock_pid`, `lock_stale`, `lock_started_at`, `lock_root` |

Numbers and booleans use native JSON types. `components` is an array of
`{"name", "status", "updated_at"}` objects or null when unavailable.

`update_blocker` may identify schema, file-source, hash, FTS, indexing-config,
different-checkout, or unknown-history incompatibility. `index_stale` compares
the stored Git head and dirty snapshot with the current checkout; it is null
when that comparison cannot be made.

`status` reports a stale lock but does not remove it. Mutation/query paths
perform stale-lock cleanup under their own rules.

## Examples

```sh
code-index status
code-index status --root /path/to/repo --format json
code-index status --db /tmp/index.sqlite
```

## Related commands

Use `logs` for operation history, `stats` for corpus counts, and the reported
compatibility fields to choose `update`, `update --adopt`, or `rebuild`.

## Open Questions

- Whether blocker values and the status JSON shape should become a separately
  versioned public schema.
- Whether status should explain malformed or damaged databases more precisely.
