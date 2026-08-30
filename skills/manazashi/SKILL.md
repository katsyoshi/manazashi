---
name: manazashi
description: Build and use a local SQLite index for code navigation instead of ad hoc grep/rg/find searches. Use when an AI coding agent needs to locate files, classes, functions, methods, symbols, definitions, references, source lines, code metrics, or index status across an unfamiliar repository; when repeated codebase lookup would otherwise use shell text search; or when the user asks for SQL-backed, database-backed, or indexed code search.
---

# Manazashi

Use Manazashi as the first search surface for codebase navigation. Re-query the
index as needed instead of retaining a deep codebase context. Treat all indexed
results, including `show`, as snapshot candidates. Read the live checkout
before editing or making behavioral claims.

When modifying Manazashi itself, also read the `Design` section in the
checkout's root `README.md`. Its absence in other repositories is expected.

## Select the executable

Resolve the target repository root to an absolute path and bind it to
`REPO_ROOT`, then select `mzci` using this precedence:

| Condition | Tool |
| --- | --- |
| `MANAZASHI_EXECUTABLE` is an absolute path | That exact path |
| `MANAZASHI_EXECUTABLE` is relative | Stop and request an absolute path |
| `MANAZASHI_EXECUTABLE` is unset | `exec/mzci` beside this `SKILL.md` |
| The selected file is missing or not executable | Stop and request installation or configuration |

Never discover a binary from `PATH` or the target repository as a fallback.
Bind the selected absolute path to `TOOL`; use `version` only for compatibility
diagnosis, not as a preflight.

Do not rely on the process working directory to select the repository. Pass
`--root "$REPO_ROOT"` to commands that accept `--root`; when the user supplied
a DB path, pass `--db` instead. Commands such as `rebuild` and `update` take the
repository root as a positional argument, so pass `"$REPO_ROOT"` to them.

`MANAZASHI_EXECUTABLE` selects the tool; `MANAZASHI_CACHE_DIR` selects the
SQLite index location. They are independent, and a value from one must never
be assigned to or inferred as the other. If the executable variable is unset,
use the bundled `exec/mzci`.

When checking both variables, print each name and value with labels; never
infer ownership from unlabeled `printenv` output.

## Maintain the index

Unless the user supplied a DB path, use a writable cache directory in
sandboxed sessions, preserving any existing value:

```bash
export MANAZASHI_CACHE_DIR="${MANAZASHI_CACHE_DIR:-/tmp/manazashi}"
```

Do not add a per-command cache assignment when a value is already set. Resolve
relative cache paths from `REPO_ROOT`.

Do not preflight freshness: query the existing index first. If a search is
empty, run `update --format json "$REPO_ROOT"` once and retry it, except for
bare-filename lookups. Request initialization if no index exists; rebuild when
diagnostics require it, and never use `update --adopt` without explicit intent.
Use `status` or `logs` for diagnostics. Index Git-tracked files only.

Untracked files are outside the index. Read them from the live checkout and do
not update or rebuild solely to make them searchable.

## Navigate

Prefer dedicated commands over raw SQL:

```bash
"$TOOL" defs --root "$REPO_ROOT" --format json parse_config
"$TOOL" refs --root "$REPO_ROOT" --format json parse_config
"$TOOL" outline --root "$REPO_ROOT" --format json path/to/file.go
"$TOOL" files --root "$REPO_ROOT" --format json config
"$TOOL" show --root "$REPO_ROOT" --line 42 --context 20 --format json path/to/file.go
```

Use `defs`/`outline` for definitions, `refs` for references, `files` for paths,
and `show` for bounded context. Use `schema`, `stats`, or `metrics` for index
structure and counts; use `help --format json COMMAND` for exact options.

Read explicitly supplied repository-relative paths from the live checkout.
Treat a bare filename as repository-root-relative first; if absent, search for
that basename under the repository root without broadening to content search.

If dedicated commands cannot express a query, use read-only SQL after reading
[`references/query-patterns.md`](references/query-patterns.md). Keep results
bounded and join `lines` to `files` for paths.

Do not use `grep`, `rg`, or `find` for code navigation until indexed commands
and read-only SQL are insufficient. Ordinary shell commands remain appropriate
for non-navigation work such as tests and Git status.
