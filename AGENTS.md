# AGENTS.md

This file is the shared repository guidance for `manazashi`. It should contain
project conventions that are useful to every contributor.

If `.local/AGENTS.md` exists at the repository root, read it after this file.
Use it for personal workflow preferences, local paths, machine-specific
commands, or other per-user notes. Do not commit files under `.local/`.

This repository contains `manazashi`, a small Go CLI that builds a local
SQLite index for source-code navigation. Keep changes focused on preserving a
lightweight, rebuildable index for agents rather than turning the project into
a language server or full code-intelligence system.

## Development

- Prefer small, behavior-preserving commits.
- Keep implementation style close to the existing Go code.
- Use `gofmt` on edited Go files.
- Do not rewrite unrelated code while making narrow changes.
- The checked-in SQL under `sql/` is embedded into the binary with `go:embed`.
  Prefer adding or changing SQL files over growing large SQL string literals in
  Go source.
- Keep dynamic SQL concerns in Go when values, filters, limits, or feature
  flags are assembled at runtime. Quote values explicitly with the existing
  helpers.
- Before implementing a new CLI command or a materially new output contract,
  add its interface design under `docs/DESIGNS/` so behavior and non-goals can
  be reviewed independently from the implementation.
- `rebuild` is the atomic full rebuild path. `update` is the incremental
  refresh path. Preserve that division.
- Indexing should continue to use Git-tracked files only unless a task
  explicitly changes that behavior.

## Verification

Run the focused checks that match the change. For normal code changes, use:

```sh
gofmt -w <edited-go-files>
go test ./...
go vet ./...
```

For CLI behavior changes, also run a small smoke test with a temporary DB, for
example:

```sh
go run ./cmd/mzci rebuild --db /tmp/manazashi-smoke.sqlite .
go run ./cmd/mzci stats --db /tmp/manazashi-smoke.sqlite
go run ./cmd/mzci metrics --db /tmp/manazashi-smoke.sqlite
```

When changing `update`, include a smoke path that exercises incremental refresh
or rely on the existing tests that cover changed, added, and deleted tracked
files.

## Repository Notes

- `README.md` is the user-facing documentation.
- `skills/manazashi/SKILL.md` is the reusable agent skill source.
- The root `mzci` binary is ignored and may be stale. Prefer
  `go run ./cmd/mzci` or rebuild it explicitly with
  `go build -o mzci ./cmd/mzci`.
- Generated SQLite databases and sidecar files are ignored.
