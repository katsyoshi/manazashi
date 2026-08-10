# Architecture

## Purpose

`manazashi` is a small local retrieval layer for source navigation. It turns
Git-tracked source files into a rebuildable SQLite database so repository
navigation state does not have to remain in a deep model-side context
throughout a task.

Queries expose progressively deeper information: compact file or symbol
candidates first, structural and lexical relationships next, and selected
source only when needed. Small results are a means to this end rather than the
entire design goal.

The index is a navigation aid, not authoritative program analysis. Callers
must inspect source before making behavioral claims.

## Data flow

Full and incremental indexing use the same per-file pipeline:

```text
Git tracked files
  -> size/type filtering
  -> source encoding normalization
  -> language detection
  -> lines + metrics + symbol extraction
  -> SQLite tables and optional FTS5 tables
  -> defs/files/outline/refs/show/sql queries
```

Only regular files reported by Git are candidates. Source files are read-only
inputs. Text accepted for indexing is normalized to UTF-8; the original bytes
remain in the checkout and are used for content hashes.

Query commands return compact candidates or selected indexed source. They do
not invoke an LSP, resolve types, or execute project code.

## Runtime artifacts

A repository normally has three independent runtime artifacts:

- the main SQLite index, either in the cache or at an explicitly selected path
- `<index-db>.lock`, present while an index-changing operation owns the index
- `<index-db>.logs.sqlite`, a persistent operation-log sidecar

The main database contains indexed source text and must be protected like the
source repository. The lock contains process and operation metadata. The log
sidecar contains paths, timestamps, statuses, and error messages but not the
indexed source corpus.

The project configuration, when present, is `.manazashi.toml` at the Git
repository root. It is input to index construction, not part of the generated
database.

## Boundaries

The project deliberately does not provide:

- language-server semantics, type checking, or exact name resolution
- a call, inheritance, dependency, or reference graph
- network services or remote storage
- automatic indexing of untracked or ignored files
- a general document knowledge base
- automatic execution of source code, except the bounded Ruby parser dump
  used for Ruby symbol extraction

Raw read-only SQL remains available so an agent can compose searches without
requiring every query shape to become a dedicated command.

## Open Questions

- Whether a future bounded workspace abstraction would improve multi-step
  agent search without turning the project into a retrieval service.
- Which missing navigation tasks justify new stored facts or signals rather
  than SQL over the existing index.
