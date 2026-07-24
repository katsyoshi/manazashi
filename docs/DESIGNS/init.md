# `init` Design

`code-index init` initializes a Git repository for later indexing. It creates
the project configuration and, when absent, an empty index database. It does
not index source files; the first content load is `code-index update`.

## Interface

```text
code-index init [--db RELATIVE_DB] [--force] [--format text|json] [ROOT]
```

- `ROOT` defaults to the current directory. A nested path resolves to the
  containing Git repository root. A path outside a Git work tree is an error.
- `.code-index.toml` is created at the repository root.
- An existing configuration is preserved and causes an error unless `--force`
  is supplied. `--force` replaces only the configuration.
- An existing database is always preserved. Initialization succeeds with
  `db_created: false`; `--force` never resets database contents.
- `--db` is repository-relative and must remain inside the repository. The
  normalized path is written as an active `db` setting.
- Without `--db`, the generated configuration does not set `db`, and the
  normal cache database path is used.

The generated configuration is intentionally small:

```toml
# max_bytes = 1000000
# ignore_dirs = ["generated", "scratch"]
# db = ".code-index/index.sqlite"

[encoding]
fallbacks = []
```

When `--db` is present, an active `db = "..."` line precedes the comments.
Project configuration continues to shadow user and machine configuration as a
whole; settings omitted here use built-in defaults.

## Atomicity

Database creation uses the existing atomic database installation path. If a
new database cannot be created after the configuration was written, `init`
removes a newly created configuration or restores the exact previous
configuration replaced by `--force`.

Only creation of a new database records an `init` build operation. Replacing
configuration while preserving an existing database does not add a build log.

## Output

Text output reports `root`, `config`, `config_created`, `config_replaced`,
`db`, `db_created`, and `next`. JSON emits the equivalent native values:

```json
{
  "operation": "init",
  "root": "/repo",
  "config": "/repo/.code-index.toml",
  "config_created": true,
  "config_replaced": false,
  "db": "/cache/code-index.sqlite",
  "db_created": true,
  "next_command": "code-index update"
}
```

`init` does not stage files, modify `.gitignore`, prompt interactively, or
select partial index components.
