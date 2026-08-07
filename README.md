# manazashi

`manazashi` builds a local SQLite index for source-code navigation. It is intended for LLM and agent workflows that should query code structure through SQL instead of repeatedly using grep-style text search.

It is not intended to replace language servers, VCS indexes, or full code-intelligence systems; it provides a small, local, queryable index that agents can rebuild cheaply.

The binary is written in Go and uses the `sqlite3` command for database creation and queries. The built binary does not require a Go runtime.

The reusable agent skill source lives at `skills/manazashi/SKILL.md`. It is written to be usable by coding agents such as Codex or Claude. The skill uses its bundled `exec/mzci` by default and accepts `MANAZASHI_BIN` as an explicit override. It never searches `PATH`, so it does not accidentally run a different `mzci` binary.

## Agent configuration

Repository-wide contributor guidance belongs in `AGENTS.md`, and the reusable
agent workflow belongs in `skills/manazashi/`. Configure agent runtimes such
as Codex or Claude individually for your own environment. Their repository-local
configuration directories (`.codex/` and `.claude/`) are ignored and should not
be committed. Machine-specific paths and commands can be kept in
`.local/AGENTS.md`, which is also ignored.

## Design

`manazashi` is a lightweight retrieval layer for agents, not a language server. It is designed to reduce how much source an LLM needs to read, not to expand context.

- Query the local index first, then read only the source needed for the task.
- Keep the index rebuildable and local.
- Treat Markdown as documentation or notes, not as the index database.

The [documentation index](docs/README.md) links detailed command references and
the implemented architecture, lifecycle, configuration, storage, symbol, and
CLI contracts. Deferred decisions are explicitly marked as open questions.

See the [reference benchmarks](docs/benchmark.md) for point-in-time measurements
against real-world repositories.

## Install

`manazashi` requires `git`, the `sqlite3` command, and Go for installation.

Choose the installed skill directory for your agent runtime. This is the directory that contains `SKILL.md`:

```sh
SKILL_DIR=/path/to/installed/skills/manazashi
```

Install the binary under that skill directory:

```sh
mkdir -p "$SKILL_DIR/exec"
GOBIN="$SKILL_DIR/exec" go install github.com/katsyoshi/manazashi/cmd/mzci@latest
```

The skill uses this bundled binary without requiring `MANAZASHI_BIN`. Set
`MANAZASHI_BIN` in the agent runtime only when the skill should use a different
trusted executable. The override must be an absolute path naming the executable
directly; the skill never searches `PATH`. In sandboxed agent sessions, set
`MANAZASHI_CACHE_DIR` to a writable location through the agent runtime rather
than relying on an interactive shell profile.

```sh
export MANAZASHI_CACHE_DIR=/tmp/manazashi
"$SKILL_DIR/exec/mzci" version
```

`version` prints the build commit hash when available, plus schema metadata useful for compatibility checks.

For agents and scripts, use JSON output:

```sh
"$SKILL_DIR/exec/mzci" version --format json
```

The JSON format emits one object. `modified` is a boolean and `schema_version` is a number; unavailable build information is `null`.

For local development in this repository, build the checked-out source into the
installed skill:

```sh
SKILL_DIR=/path/to/installed/skills/manazashi
mkdir -p "$SKILL_DIR/exec"
go build -o "$SKILL_DIR/exec/mzci" ./cmd/mzci
```

For Codex local skill development, you can symlink the checked-in skill into Codex's skill directory first:

```sh
SKILL_DIR="${CODEX_HOME:-$HOME/.codex}/skills/manazashi"
mkdir -p "${CODEX_HOME:-$HOME/.codex}/skills"
ln -s "$PWD/skills/manazashi" "$SKILL_DIR"
```

If the target already exists, remove or rename it first. For other agent
runtimes, set `SKILL_DIR` to that runtime's installed `skills/manazashi`
directory. Configure `MANAZASHI_BIN` through the runtime only when overriding
the bundled executable.

Host-specific metadata can live under `skills/manazashi/agents/`; agents that do not use those files can ignore them.

## Usage

See the [command reference index](docs/README.md#command-references) for the
complete interface, output fields, errors, and edge cases of every subcommand.

Run commands from anywhere inside a Git work tree to use its repository root automatically:

```sh
mzci rebuild
mzci update
mzci status --format json
```

Passing a root remains supported when operating on another checkout.

Update an existing index incrementally:

```sh
mzci update /path/to/repo
mzci update --format json /path/to/repo
```

`update` requires an existing index database and a Git work tree. It refreshes changed Git-tracked files and removes files that are no longer tracked. If the index does not exist yet, run `init` or `rebuild` first.

`update` refuses incompatible indexes instead of silently mixing metadata. It asks for `rebuild` when schema, file source, hash, or indexing config settings do not match. If the database was created from another checkout path or Git history and you intentionally want this checkout to take it over, run:

```sh
mzci update --adopt /path/to/repo
```

Its change counts are file counts: `added_files`, `updated_files`, and `deleted_files`. `symbols` reports symbols indexed from added or updated files during that update; use `stats` or `metrics` for index-wide totals.

Rebuild an index atomically from Git-tracked files:

```sh
mzci rebuild /path/to/repo
mzci rebuild --format json /path/to/repo
```

Source text is stored as UTF-8. UTF-8 and BOM-marked UTF-16/UTF-32 are handled
internally. Ruby and Python source-encoding declarations are honored. A project
with legacy source may provide ordered fallback encodings in `.manazashi.toml`:

```toml
[encoding]
fallbacks = ["Windows-31J", "EUC-JP"]
```

Fallback conversion uses the optional `iconv` command. Files with unknown or
unconvertible encodings are retained as skipped metadata and do not contribute
lines, symbols, metrics, or FTS content. Build output reports transcoded and
encoding-skipped counts; use `-v` or `--verbose` to include per-file details.
Changing the fallback list requires a rebuild before the next update. See
[`docs/DESIGNS/text-encoding.md`](docs/DESIGNS/text-encoding.md) for the full
contract.

`rebuild` requires a Git work tree and indexes files reported by `git ls-files`. If another `init`, `rebuild`, or `update` is already running for the same database, `rebuild` skips and exits successfully.

Initialize a Git project for indexing:

```sh
mzci init /path/to/repo
mzci init --db .manazashi/index.sqlite /path/to/repo
mzci init --format json /path/to/repo
```

`init` creates `.manazashi.toml` at the Git repository root and creates an
empty schema/metadata database when one does not already exist. It does not
index source files; run `mzci update` next. `--db` accepts a
repository-relative path and stores it in the generated configuration.
Existing configuration is preserved unless `--force` is given, while an
existing database is always preserved. See
[`docs/COMMANDS/init.md`](docs/COMMANDS/init.md) for the complete contract.

For `init`, `rebuild`, and `update`, the JSON format emits one operation-result
object with native values. If `rebuild` or `update` skips because another index
operation holds the lock, it exits successfully with `skipped: true`,
`reason: "locked"`, and unavailable result fields set to `null`; the warning
remains on stderr.

Show index status:

```sh
mzci status --root /path/to/repo
mzci status --root /path/to/repo --format json
```

`status` reports database metadata, current lock state, current Git head/branch/dirty state, and `index_stale`. A dirty work tree is fresh when it matches the dirty snapshot recorded by the last successful index operation.

The JSON format is intended for agents and scripts. It emits one object, uses native JSON booleans and numbers, and represents unavailable values as `null`.

It also reports whether `update` can safely proceed, whether `update --adopt` would be required, or whether `rebuild` is required.

Show recent build operation logs:

```sh
mzci logs
mzci logs --limit 50 --format json
```

Database-creating `init` operations and all `rebuild` and `update` operations
record `succeeded`, `failed`, and lock-related `skipped` runs after root and
database resolution succeeds. Config-only `init --force` operations against an
existing database are not logged. Logs are stored outside the replaceable
index DB in a SQLite sidecar named `<index-db>.logs.sqlite`, so a failed atomic
rebuild can still be recorded. The newest 1,000 runs are retained. Logging
failures are warnings and do not change the index operation result. The JSON
result is an array ordered newest first; an absent sidecar returns `[]`.

Show command help:

```sh
mzci help
mzci help update
mzci help --format json
mzci help --format json update
```

The JSON format returns the top-level usage and command metadata. Whole-program help contains a `commands` array; command-specific help contains one object with `name`, `usage`, and `summary`.

Find symbol definitions:

```sh
mzci defs --root /path/to/repo --list --format json
mzci defs --root /path/to/repo parse_config
mzci defs --root /path/to/repo --format json parse_config
```

Use `--list` without `QUERY` to list definitions ordered by path and source position. `--kind`, `--language`, and `--limit` apply to both listing and searching. Combining `--list` with `QUERY` is an error.

Show the symbols in one indexed file:

```sh
mzci outline --root /path/to/repo path/to/file.go
mzci outline --root /path/to/repo --format json path/to/file.go
```

`outline` returns symbols in source order. It prefers an exact repository-relative path and also accepts a path suffix such as `file.go`.

Find likely symbol references:

```sh
mzci refs --root /path/to/repo parse_config
mzci refs --root /path/to/repo --kind function --kind method parse_config
mzci refs --root /path/to/repo --language go --ignore-case --format json parseConfig
```

The default text output is intended for interactive CLI use. `--format json`
is available for agents and scripts. `refs` matches complete identifiers with
case sensitivity by default; `--ignore-case` changes only case handling and
does not enable substring matches. Exact definition lines are excluded while
comments and strings remain searchable. `--kind` may be repeated to filter the
definitions reported with the candidates, and `--limit` controls candidate
count. Use `outline` to inspect structure and `show` to read context around a
selected candidate. Results remain candidates rather than a resolved call or
reference graph.
See [`docs/COMMANDS/refs.md`](docs/COMMANDS/refs.md) for the full contract.

Find files:

```sh
mzci files --root /path/to/repo --list --format json
mzci files --root /path/to/repo config
mzci files --root /path/to/repo --format json config
mzci files --root /path/to/repo --status skipped --list
mzci files --root /path/to/repo --status all --list --format json
```

Use `--list` without `QUERY` to list files, ordered by path. The default
`--status indexed` returns searchable files. Use `--status skipped` for files
omitted because their encoding could not be determined or converted, and
`--status all` for both. `--language` and `--limit` apply to listing and
searching. Combining `--list` with `QUERY` is an error.

Run read-only SQL:

```sh
mzci sql --root /path/to/repo \
  "select path, line, kind, name, signature from symbols where name like '%parse%' order by path, line limit 50"

mzci sql --root /path/to/repo --format json \
  "select path, line, kind, name, signature from symbols where name like '%parse%' order by path, line limit 50"
```

The JSON format emits an array of objects with SQLite numbers and nulls preserved. Use unique column names or explicit aliases in JSON queries; duplicate column names are ambiguous when represented as object fields.

Show indexed source around a line:

```sh
mzci show --root /path/to/repo --line 42 lib/config.rb
mzci show --root /path/to/repo --line 42 --format json lib/config.rb
```

Show the current index tables and columns:

```sh
mzci schema --root /path/to/repo
mzci schema --root /path/to/repo --format json
```

`schema` reports user-facing tables, virtual tables, column types, nullability, and primary-key positions. SQLite and FTS5 internal tables are omitted.
The JSON format emits an array with native numbers and booleans; `key` is `null` for columns that are not part of a primary key.

Show indexed code metrics:

```sh
mzci metrics --root /path/to/repo
mzci metrics --root /path/to/repo lib/config
mzci metrics --root /path/to/repo --format json
```

For `defs`, `outline`, `files`, `show`, and `metrics`, the JSON format emits an array of objects, uses native JSON numbers, preserves nullable fields as `null`, and emits `[]` when there are no rows. `refs` instead emits one object containing `query`, `definitions`, and `candidates` arrays.

Show index-wide counts and build metadata:

```sh
mzci stats --root /path/to/repo
mzci stats --root /path/to/repo --format json
```

The JSON format emits one object with native counts and booleans. Unavailable metadata fields are `null`.

The default database is stored under `MANAZASHI_CACHE_DIR` when set. Otherwise it uses `$XDG_CACHE_HOME/manazashi` or `~/.cache/manazashi`, keyed by the absolute repository path. Use `--db` to provide an explicit database path.

Configuration is selected by taking the first existing file in this order:

1. `REPOSITORY_ROOT/.manazashi.toml`
2. `$XDG_CONFIG_HOME/manazashi/config.toml`, or `~/.config/manazashi/config.toml` when `XDG_CONFIG_HOME` is unset
3. `/etc/manazashi/config.toml`

Only that file is read. Missing fields use built-in defaults; values are not merged from less-specific files. Supported top-level fields are:

```toml
max_bytes = 1000000
ignore_dirs = ["generated", "scratch"]
db = ".manazashi/index.sqlite"
```

`ignore_dirs` adds directory names to the built-in ignore list. `db` is allowed only in the project configuration, must be relative to the repository root, and must remain inside it. `--db` and `--max-bytes` override configured values for one run; `--ignore-dir` adds another ignored directory. Unknown fields and invalid values are errors.

Print the default database path without creating it:

```sh
mzci path /path/to/repo
mzci path --format json /path/to/repo
mzci path
```

The JSON format emits one object with a `path` field.

`rebuild` and `update` index Git-tracked files only. Initialize Git and add files before indexing a directory.

## Git Hooks

You can refresh the local index automatically after branch checkouts and merges. These hooks are optional and only run when `MANAZASHI_BIN` points to an executable `mzci` binary. If Git runs hooks outside your shell environment, set `MANAZASHI_BIN` inside the hook or from the environment that launches Git.

Refresh the index after switching branches:

```sh
cat > .git/hooks/post-checkout <<'EOF'
#!/bin/sh
# Args: previous HEAD, new HEAD, branch checkout flag.
[ "$3" = "1" ] || exit 0

root="$(git rev-parse --show-toplevel)" || exit 0
[ -n "${MANAZASHI_BIN:-}" ] || exit 0
[ -x "$MANAZASHI_BIN" ] || exit 0

(
  "$MANAZASHI_BIN" update "$root" >/dev/null 2>&1 ||
    "$MANAZASHI_BIN" rebuild "$root" >/dev/null 2>&1
) &
EOF
chmod +x .git/hooks/post-checkout
```

Refresh the index after pulls or merges:

```sh
cat > .git/hooks/post-merge <<'EOF'
#!/bin/sh
root="$(git rev-parse --show-toplevel)" || exit 0
[ -n "${MANAZASHI_BIN:-}" ] || exit 0
[ -x "$MANAZASHI_BIN" ] || exit 0

(
  "$MANAZASHI_BIN" update "$root" >/dev/null 2>&1 ||
    "$MANAZASHI_BIN" rebuild "$root" >/dev/null 2>&1
) &
EOF
chmod +x .git/hooks/post-merge
```

The examples run in the background so Git commands do not wait for indexing. During `init`, `rebuild`, and `update`, `mzci` writes a `.lock` file next to the target database. Queries keep using a consistent SQLite snapshot and print a warning to stderr while the lock is present. If no previous index exists yet, queries fail with a message that indexing is still in progress.

If a lock file records a PID that is no longer running, build/update/query commands treat it as stale and remove it before continuing. `status` reports `lock_stale` for visibility without requiring a rebuild.

`status` combines the lock file with database metadata. Lock fields describe a currently running operation; metadata such as `indexed_at`, `last_operation`, `vcs_head`, and `vcs_branch` describes the last successful update. The JSON result also includes `components`, with the completed `ready`, `disabled`, or `unavailable` state and update time for each index component.

## Schema

Main tables:

- `meta`: schema version, file source, hash algorithm, last successful index time, operation, and VCS metadata such as head, branch, tracked-file dirty state, and dirty snapshot hash
- `components`: completed state and update time for the `files`, `lines`, `symbols`, `metrics`, and `fts` index components
- `files`: repository-relative paths and file metadata
- `symbols`: regex-extracted definitions such as functions, methods, classes, modules, interfaces, traits, and types
- `lines`: indexed source lines
- `file_metrics`: per-file line, blank line, comment line, code line, and symbol counts
- `files_fts` and `symbols_fts`: FTS5 tables when the installed `sqlite3` supports FTS5

See the [storage design](docs/DESIGNS/storage.md) for table responsibilities,
build identity, schema evolution, and data-sensitivity requirements.

## Notes

This is a lightweight regex-based index, not a language server. Treat matches as navigation candidates and inspect source before making behavioral claims.

## License

MIT
