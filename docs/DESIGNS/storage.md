# Storage

## Main database

The current schema version is 4. The main tables are:

- `meta`: string key/value metadata describing schema, root, build identity,
  timestamps, operation, and Git state
- `components`: completed state and update time for `files`, `lines`,
  `symbols`, `metrics`, and `fts`
- `files`: path, language, raw size/hash, indexing status, and encoding
  diagnostics
- `lines`: normalized UTF-8 source, keyed by file and one-based line number
- `symbols`: extracted definitions with source position, signature, and nearby
  context
- `file_metrics`: per-file code, comment, blank, line, and symbol counts

Foreign keys connect line, symbol, and metric rows to `files`. Rebuild assigns
new internal IDs; consumers must use stable fields such as path and source
position rather than persisting row IDs across rebuilds.

## Optional FTS

If the installed `sqlite3` supports FTS5, construction also creates:

- `files_fts(path, language, content)`
- `symbols_fts(name, kind, language, path, signature, context)`

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
