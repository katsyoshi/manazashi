---
name: manazashi
description: Build and use a local SQLite index for code navigation instead of ad hoc grep/rg/find searches. Use when an AI coding agent needs to locate files, classes, functions, methods, symbols, definitions, references, source lines, code metrics, or index status across an unfamiliar repository; when repeated codebase lookup would otherwise use shell text search; or when the user asks for SQL-backed, database-backed, or indexed code search.
---

# Manazashi

Use Manazashi as the first search surface for codebase navigation. Re-query the
index as needed instead of retaining a deep codebase context. Treat matches as
navigation candidates and inspect source before making behavioral claims.

When modifying Manazashi itself, also read the `Design` section in the
checkout's root `README.md`. Its absence in other repositories is expected.

## Select the executable

Resolve the target repository root, then select `mzci` using this precedence:

| Condition | Tool |
| --- | --- |
| `MANAZASHI_BIN` is an absolute path | That exact path |
| `MANAZASHI_BIN` is relative | Stop and request an absolute path |
| `MANAZASHI_BIN` is unset | `exec/mzci` beside this `SKILL.md` |
| The selected file is missing or not executable | Stop and request installation or configuration |

Never discover a binary from `PATH` or the target repository as a fallback.
Bind the selected absolute path to `TOOL` for subsequent commands.

Check the selected build before using it:

```bash
"$TOOL" version --format json
```

Treat a build commit as an identity, not an ordered version. Use known
compatibility information or explicit feature checks; if the required command
is unavailable, request a compatible binary.

## Maintain the index

Unless the user supplied a DB path, use a writable cache directory in
sandboxed sessions:

```bash
export MANAZASHI_CACHE_DIR="${MANAZASHI_CACHE_DIR:-/tmp/manazashi}"
```

Refresh the index before substantial navigation work:

```bash
"$TOOL" update --format json
```

If the repository has not been initialized, run `init` and then `update`. Use
`status --format json` when freshness, locks, or compatibility matter. Inspect
`components`, `update_compatible`, `update_requires_adopt`,
`update_rebuild_required`, and `update_blocker` when present.

- Run `rebuild` when status reports incompatible schema, file source, hashing,
  or indexing configuration.
- Run `rebuild` for another checkout path or unknown Git history unless the
  user explicitly wants the index adopted; only then use `update --adopt`.
- If `status` is unsupported, continue with query commands and rely on rebuild
  output.
- Use `logs --format json` to diagnose failed or skipped index operations.

Index Git-tracked files only unless the task explicitly changes that contract.
Run `update` after editing tracked files. Run `rebuild` after schema, tool, or
indexing-option changes that require every file to be refreshed.

## Navigate

Prefer dedicated commands over raw SQL:

```bash
"$TOOL" defs --format json parse_config
"$TOOL" refs --format json parse_config
"$TOOL" outline --format json path/to/file.go
"$TOOL" files --format json config
"$TOOL" show --line 42 --context 20 --format json path/to/file.go
```

Use `defs` and `outline` for definitions, `refs` for likely references, `files`
for paths and indexing diagnostics, and `show` for bounded source context. Use
`schema`, `stats`, or `metrics` when the task needs index structure or counts.
Use `help --format json COMMAND` for exact command options instead of relying
on a copied command catalog.

If dedicated commands cannot express a query, use read-only SQL. Read
[`references/query-patterns.md`](references/query-patterns.md) before composing
raw SQL. Prefer bounded results and join `lines` to `files` for paths.

Do not use `grep`, `rg`, or `find` for code navigation until indexed commands
and read-only SQL are insufficient. Ordinary shell commands remain appropriate
for non-navigation work such as tests and Git status.
