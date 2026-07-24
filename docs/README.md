# Documentation

The repository `README.md` is the installation and basic usage guide. This
directory provides detailed command references and cross-cutting design
documents.

## Command references

- [`init`](COMMANDS/init.md)
- [`refs`](COMMANDS/refs.md)

Command references describe a subcommand's public interface, matching rules,
output, and non-goals.

## Design documents

- [Architecture](DESIGNS/architecture.md): goals, boundaries, artifacts, and indexing
  data flow
- [Index lifecycle](DESIGNS/lifecycle.md): initialization, rebuilds, updates, locking,
  compatibility, status, and operation logs
- [Configuration](DESIGNS/configuration.md): root and database resolution,
  configuration scopes, defaults, and command-line overrides
- [Storage](DESIGNS/storage.md): SQLite schema, metadata, optional FTS, and data
  sensitivity
- [Symbols](DESIGNS/symbols.md): definition extraction, symbol kinds, and the roles of
  `defs`, `outline`, and `refs`
- [CLI contract](DESIGNS/cli.md): output formats, errors, stable results, and command
  families
- [Source text encoding](DESIGNS/text-encoding.md)

The main body of each document describes the current implementation. An
`Open Questions` section records intentionally deferred decisions; it is not a
promise that the described feature exists.
