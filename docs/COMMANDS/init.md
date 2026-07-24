# `init` command

## Purpose

`code-index init` initializes a Git repository for later indexing. It creates
the project configuration and, when absent, an empty index database. It does
not index source files; the first content load is `code-index update`.

## Usage

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
- `--format` defaults to `text`.
- More than one positional root is rejected.

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

`operation`, `root`, `config`, `db`, and `next_command` are strings.
`config_created`, `config_replaced`, and `db_created` are booleans.

`init` does not stage files, modify `.gitignore`, prompt interactively, or
select partial index components.

## Errors

Missing `sqlite3`, a non-Git root, invalid or escaping `--db`, an existing
configuration without `--force`, lock acquisition failure, or database/config
write failure is an error. If database creation fails after configuration was
written, the configuration rollback described above is attempted.

## Examples

```sh
code-index init
code-index init --db .code-index/index.sqlite /path/to/repo
code-index init --force --format json
```

## Related commands

Run `update` next to load content. Use `rebuild` to replace an existing index
from current Git-tracked files and `path` to inspect default database
resolution.

## Open Questions

- Whether lock conflicts should eventually use the same successful-skip
  contract as `rebuild` and `update`; `init` currently returns an error.
- Whether future configuration fields should be added to the minimal generated
  template or remain discoverable only through documentation.
