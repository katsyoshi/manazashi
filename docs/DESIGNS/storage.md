# Storage

The schema responsibilities, data ownership, stable identity, invariants, and
intentional FTS duplication are documented in
[`db_schema.md`](db_schema.md). Exact executable columns and constraints live
in the embedded SQL. This document covers storage policy around that schema.

## Main database

The current schema version is 4. The main database is one completed,
replaceable navigation snapshot. It contains canonical indexed source,
extracted navigation facts, derived metrics, compatibility metadata, and
optional search projections. Rebuild may assign new internal IDs; consumers
must use stable fields such as path and source position instead of persisting
row IDs.

## Optional FTS

If the installed `sqlite3` supports FTS5, construction also creates:

- `files_fts(path, language, content)`
- `symbols_fts(name, kind, language, path, signature, context)`

`files_fts.content` is an intentional agent-oriented search cache rather than
the canonical source representation. See [`db_schema.md`](db_schema.md) for
the rationale and query-boundary guidance.

FTS availability is stored as part of the index identity. An environment whose
FTS5 support differs from the environment that built the database must rebuild
before update. The `schema` command hides FTS shadow tables but exposes the
user-facing virtual tables.

## Indexed and skipped files

`files.index_status` is `indexed` or `skipped`. Successfully indexed files have
line, symbol, metric, and optional FTS rows. Encoding-skipped files retain path,
language, raw size/hash, and diagnostic metadata but have no searchable source
rows.

Files rejected before decoding because they are untracked, non-regular,
ignored, too large, binary, or otherwise unsupported are not represented as
encoding-skipped rows.

Source encoding behavior is specified in
[`text-encoding.md`](text-encoding.md).

## Metadata and components

Build identity includes schema version, `git-tracked` file source, SHA-256
content hashing, FTS availability, and effective indexing configuration.
Completed operations also record root, timestamps, operation name, and
available Git revision/branch/dirty metadata.

Components describe the last completed database state. In-progress state lives
in the external lock, so readers never interpret a partially written component
row as progress reporting.

## Data sensitivity

The database contains source text in `lines`, `files_fts`, symbol signatures,
and symbol context. It has the same confidentiality requirements as the
repository and should not be published as an innocuous cache artifact.

The database is plain SQLite. Encryption, access control, backup, and secure
deletion are delegated to the host environment.

## Schema evolution

There is no in-place schema migration command. A tool/database schema mismatch
causes update to fail and instructs the caller to rebuild. The schema version
must change whenever an incompatible stored contract changes.

## Open Questions

- Whether future non-symbol observations belong in a generic `signals` table
  or in feature-specific tables. No signals table exists today.
- Whether optional components will ever need a disabled state selected by
  configuration; current builds attempt every component.
- Whether storing source context in both symbols and FTS remains worthwhile as
  extractors and query patterns evolve.
