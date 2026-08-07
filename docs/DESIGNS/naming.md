# Project and CLI naming

## Public names

The project, repository, Go module, and reusable agent skill use the name
`manazashi`. The installed command uses the shorter name `mzci`.

The Go module path is:

```text
github.com/katsyoshi/manazashi
```

The command is installed from its dedicated main package:

```sh
go install github.com/katsyoshi/manazashi/cmd/mzci@latest
```

The repository root is an importable `manazashi` package containing the CLI
implementation and embedded SQL. `cmd/mzci` is a thin process entry point. This
keeps the checked-in `sql/` assets at the repository root while allowing the
installed executable name to differ from the project name.

## Configuration names

Project-owned configuration and cache locations consistently use the project
name:

- project configuration: `.manazashi.toml`
- user configuration: `$XDG_CONFIG_HOME/manazashi/config.toml`
- system configuration: `/etc/manazashi/config.toml`
- cache environment: `MANAZASHI_CACHE_DIR`
- optional binary override used by the skill: `MANAZASHI_BIN`

The agent skill directory and frontmatter name are both `manazashi`. Its
default binary is `exec/mzci` within that directory. An absolute
`MANAZASHI_BIN` overrides that exact path when set. The skill never searches
`PATH` or selects a binary from the target repository.

## Migration policy

This rename is a clean break. The old command, module path, skill name,
configuration filename, environment variables, and cache directory are not
accepted as aliases. Existing indexes remain rebuildable data rather than a
compatibility boundary; users should rebuild after moving to the new names.

GitHub repository renaming and redirects are hosting operations outside the CLI
contract and are performed separately from the source change.

## Non-goals

- Supporting two executable names in one release.
- Reading legacy configuration or environment variables.
- Migrating or renaming existing cache databases automatically.
- Adding runtime branding fields to index metadata.
