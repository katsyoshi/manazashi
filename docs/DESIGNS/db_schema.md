# Database schema intent

## Purpose and authority

This document defines why the main index and operation-log sidecar have their
current shape, which representations are canonical within the index, and which
properties consumers may rely on. It is the design contract for schema version
4, not a transcription of every SQL type, constraint, or index name.

The embedded SQL under `sql/` is the exact executable definition of that
contract. A design change starts here when it changes a table's responsibility,
data ownership, stable identity, or a consumer-visible invariant; the SQL and
tests then implement and verify it. Purely mechanical SQL details do not need
prose duplication.

[`storage.md`](storage.md) separately defines storage lifecycle, compatibility,
and confidentiality policy.

## Design goals

The schema supports a rebuildable navigation memory for coding agents. It is
optimized for progressively retrieving files, definitions, references, and
selected source without requiring a deep codebase context to remain in the
model throughout a task.

The design favors:

- repository-relative paths and source positions as stable navigation facts;
- normalized source that can be queried without rereading the live checkout;
- direct, bounded queries over strict normalization where duplication makes
  agent retrieval simpler;
- atomic replacement and cheap rebuilds instead of in-place schema migration;
- explicit separation between canonical indexed data and search caches.

It does not make extracted symbols or lexical matches authoritative program
semantics.

## Ownership and identity

The following conceptual map distinguishes canonical indexed data, derived
facts, search projections, and operational history. It is not an ER diagram or
a catalog of relational joins.

```text
files
  ├──< lines
  ├──< symbols
  └─── file_metrics

meta             index identity and completed-build metadata
components       completed component readiness
files_fts        optional agent-oriented file-content search cache
symbols_fts      optional symbol search cache

<index-db>.logs.sqlite
  └── build_runs operation history without indexed source
```

`files` owns the lifecycle of line, symbol, and metric rows. Deleting a file
removes those dependent facts. FTS rows are maintained explicitly because the
FTS tables are search projections rather than relational children.

Internal integer IDs exist to join rows efficiently inside one completed
database. Rebuild may assign different IDs, so consumers must not persist them.
Repository-relative `path` plus an appropriate source position is the stable
navigation identity exposed across rebuilds.

## File facts and indexing state

`files` records one parent row for each successfully indexed file and each
tracked file retained as encoding-skipped metadata. Its columns serve four
design roles:

| Role | Columns | Intent |
| --- | --- | --- |
| Internal identity | `id` | Efficient joins within one rebuilt database; not stable externally |
| Navigation identity | `path` | Unique repository-relative identity used by consumers |
| Classification | `language`, `extension` | Query filtering and extractor selection |
| Source observation | `size`, `mtime`, `content_hash` | Describe the indexed snapshot; `content_hash` is the incremental-update content identity, while size and modification time remain metadata |
| Indexing outcome | `index_status` | Distinguish searchable files from retained skipped metadata |
| Decoding outcome | `source_encoding`, `encoding_source`, `transcoded`, `skip_reason` | Explain how stored UTF-8 was obtained or why source was omitted |

An `indexed` row may own source, symbol, metric, and optional FTS data. A
`skipped` row preserves enough metadata to diagnose an encoding failure but
must not appear to have searchable source. Files rejected before decoding are
outside this schema rather than represented as encoding-skipped rows.

## Canonical indexed source

`lines(file_id, line, text)` is the canonical source representation inside the
index. The checkout remains the authoritative project source; `lines` is the
authoritative representation for commands reading a completed index snapshot.

Source is stored as normalized UTF-8, one row per one-based source line. This
shape intentionally supports:

- bounded `show` results without reconstructing or slicing whole files;
- stable path-and-line navigation;
- lexical reference candidates and source-oriented SQL;
- deletion and replacement at file granularity through the parent relation.

Original source bytes are not preserved in SQLite. Encoding policy is defined
in [`text-encoding.md`](text-encoding.md).

## Extracted definitions

`symbols` stores lightweight navigation candidates rather than resolved
semantic entities. A symbol identifies a likely definition and where to inspect
it; it does not replace reading source or provide language-server resolution.
The extraction and navigation contract is defined in
[`symbols.md`](symbols.md). Exact stored fields are part of the executable
schema.

Symbol extraction may miss dynamic definitions or report approximate spans.
Consumers must inspect source before making behavioral claims.

## Precomputed metrics

`file_metrics` contains derived per-file summaries used by metrics and status
queries. It does not add a second source of program meaning; exact stored
fields and aggregation mechanics are implementation details of the executable
schema.

## Index identity and component state

`meta` records the identity of a completed index: schema and file-source
contracts, hashing and effective indexing configuration, root, timestamps,
operation, Git state, and FTS availability. These values let `status` and
`update` decide whether an index can be refreshed safely or must be rebuilt.

`components` records readiness for files, lines, symbols, metrics, and FTS.
Only completed state belongs in the main database. In-progress ownership and
timing belong to the external lock so readers never interpret partial build
progress as a usable component state.

## Agent-oriented FTS projections

When FTS5 is available, the schema adds:

- `files_fts(path, language, content)`
- `symbols_fts(name, kind, language, path, signature, context)`

`files_fts.content` deliberately duplicates normalized file source. It is not
the canonical indexed representation and is not primarily a human-written SQL
interface. It is a denormalized agent search cache that keeps the completed
index snapshot self-contained and enables bounded full-text queries without
rereading the checkout or reconstructing files from line rows.

Humans searching the live checkout should normally use `grep`. Agents should
prefer dedicated commands such as `defs`, `refs`, `files`, and `show`, then use
read-only FTS SQL when those commands do not express the required search. FTS
queries should return candidates or limited context rather than placing an
entire matched file into model context by default.

`symbols_fts` similarly projects definition fields for broad search. Both FTS
tables may be rebuilt entirely from canonical relational data. Their
availability is part of index compatibility because changing it changes the
completed query surface.

The intentional source duplication may be reconsidered if measured database
size or update cost outweighs snapshot and query benefits. External-content or
contentless FTS are alternatives, not the current contract.

## Operation-log isolation

`<index-db>.logs.sqlite` is a separate database so operation history survives
atomic replacement or failed construction of the main index. `build_runs`
records operation identity, target paths, status, timing, and error information
but no indexed source corpus.

Keeping logs separate prevents operational history from becoming part of the
replaceable navigation snapshot and avoids copying it during rebuild.

## Exact implementation and evolution

The exact executable structures are:

- `sql/schema.sql` for core tables, constraints, and indexes;
- `sql/schema_fts.sql` for optional FTS5 projections;
- `sql/operation_logs_schema.sql` for the sidecar.

An incompatible change to the responsibilities, ownership, identity, or
invariants described here requires a schema-version change and rebuild
behavior. When only an internal SQL detail changes without altering this design
contract, the SQL and focused tests may change without expanding this document
into a complete DDL catalog.
