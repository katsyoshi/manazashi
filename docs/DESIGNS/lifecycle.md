# Index lifecycle

## Operations

The lifecycle separates project setup, full replacement, and incremental
refresh:

- `init` creates project configuration and an empty schema/metadata database.
  It does not index source. See [`init`](../COMMANDS/init.md).
- `rebuild` constructs a complete database from current Git-tracked files in a
  temporary file and renames it over the target only after success.
- `update` requires an existing compatible database and changes it inside one
  SQLite `begin immediate` transaction.

`rebuild` is the recovery path after schema changes and indexing-configuration
changes. `update` is the normal refresh path after source changes. Renames are
represented as deletion plus addition.

## Update compatibility

Before modifying an existing database, `update` compares stored metadata with
the current request. A rebuild is required when any of these identities differ:

- schema version
- file source
- content hash algorithm
- FTS5 availability
- maximum source size
- ignored directory names
- source-encoding fallback list

A database from another checkout or an unknown Git history is rejected by
default. `update --adopt` permits those ownership/history cases only; it does
not bypass schema or indexing-configuration incompatibility.

## Locking and readers

`init`, `rebuild`, and `update` coordinate through `<index-db>.lock`. The lock
records operation, repository root, process ID, and start time.

When `rebuild` or `update` finds an active lock, it skips successfully. JSON
reports `skipped: true` and `reason: "locked"`; text output writes a notice to
stderr. `init` treats failure to acquire its lock as an error.

Query commands continue reading the previous database while a lock is active
and emit a warning to stderr. If no previous database exists, queries fail.
A lock is removed as stale only when it has a valid local PID that can be
confirmed as no longer running. Malformed or unverifiable locks are retained.

## Metadata, status, and logs

Every successful database construction or refresh updates `meta` and the five
component rows: `files`, `lines`, `symbols`, `metrics`, and `fts`. Component
statuses are `ready`, `disabled`, or `unavailable`; the current implementation
uses `unavailable` when FTS5 cannot be created.

`status` combines completed database metadata, current Git state, and lock
state. It reports whether normal update, adoption, or rebuild is required.

Database-creating `init` operations and `rebuild`/`update` attempts that reach
resolved root and database paths are written to `<index-db>.logs.sqlite` as
`succeeded`, `failed`, or `skipped`.
Config-only `init --force` against an existing database is not logged. Log
failures produce warnings but do not change the indexing result. The newest
1,000 rows are retained.

## Failure behavior

- A failed rebuild leaves the previously installed database in place.
- A failed update rolls back its SQLite transaction.
- A failed initial database creation restores or removes the configuration as
  specified by the `init` contract.
- Temporary database files and SQLite sidecars are cleaned up on the normal
  error paths.

## Open Questions

- Whether all index-changing commands should use identical lock-conflict exit
  behavior; `init` currently differs from `rebuild` and `update`.
- Whether status should expose more detail about an incompatible or partially
  damaged database than the current blocker fields.
