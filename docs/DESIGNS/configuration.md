# Configuration and path resolution

## Configuration selection

At most one TOML configuration file is loaded. Candidates are checked in this
order:

1. `REPOSITORY_ROOT/.manazashi.toml`
2. `$XDG_CONFIG_HOME/manazashi/config.toml`, or
   `~/.config/manazashi/config.toml`
3. `/etc/manazashi/config.toml`

The first existing file completely shadows less-specific files. Values are not
merged across scopes. Fields omitted from the selected file use built-in
defaults. Unknown keys, invalid TOML, and invalid values are errors.

Supported values are:

```toml
max_bytes = 1000000
ignore_dirs = ["generated", "scratch"]
db = ".manazashi/index.sqlite"

[encoding]
fallbacks = ["Windows-31J", "EUC-JP"]
```

`db` and `[encoding]` are project-only because their meaning belongs to a
specific repository. `db` must be relative and remain inside the repository.
`max_bytes` must be positive. Encoding fallback names must be non-empty and
case-insensitively unique.

`ignore_dirs` supplies additional directory names alongside built-in ignored
names. Matching is by directory name, not by path glob.

## Command-line precedence

Explicit command-line values override the selected configuration for one
operation:

- `--db` selects the database path directly.
- `--max-bytes` replaces configured `max_bytes`.
- each `--ignore-dir` adds another ignored directory.

Changing an indexing value can make an existing database incompatible with
`update`; use `rebuild` to construct it with the new identity.

`init --db` is intentionally different from the general `--db`: it accepts
only a repository-relative path, writes the normalized value to the generated
project configuration, and never replaces an existing database.

## Repository root

With no root argument, commands start from the current directory and use the
containing Git root when one can be discovered. Otherwise they use the current
directory as an absolute path; commands that require Git later reject a
non-work-tree.

An explicit root used by most commands is converted to an absolute directory
but is not promoted to its containing Git root. `init` is stricter: both an
explicit path and the current directory must resolve to a Git repository, and
the generated configuration is always placed at that Git root.

Query commands accept `--root` for resolving configuration and the default
database. An explicit `--db` bypasses root/config database selection.

## Default database path

When neither CLI nor project configuration selects a database, the path is:

1. `$MANAZASHI_CACHE_DIR/<root-hash>.sqlite`
2. `$XDG_CACHE_HOME/manazashi/<root-hash>.sqlite`
3. `~/.cache/manazashi/<root-hash>.sqlite`
4. the platform temporary directory as a final fallback

The filename uses the first 16 hexadecimal characters of SHA-256 over the
absolute root path. Moving a checkout therefore selects a different default
database.

## Open Questions

- Whether project configuration needs a `cache_dir` distinct from the exact
  `db` path, and where it belongs in database path precedence.
- Whether explicit nested roots should consistently resolve to the containing
  Git root, as `init` already does. Current behavior differs across commands.
- Whether user/system configuration should eventually permit encoding policy;
  it is currently rejected outside project configuration.
- Whether configuration introspection needs a command that reports the
  selected file, scope, defaults, and effective values.
