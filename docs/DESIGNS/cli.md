# CLI contract

## Command families

The public commands fall into four groups:

- setup and mutation: `init`, `rebuild`, `update`
- inspection and health: `version`, `path`, `status`, `logs`, `stats`, `schema`
- navigation: `defs`, `outline`, `refs`, `files`, `show`, `metrics`
- composition and discovery: `sql`, `help`

All commands are non-interactive. They write their result to stdout, warnings
to stderr, and return an error/non-zero process status on failure. Lock-related
successful skips are the documented exception for `rebuild` and `update`.

## Output formats

Commands accept `--format text|json`; text is the default. Unsupported format
names are errors.

JSON output is exactly one document followed by a newline. Native JSON numbers
and booleans are not stringified. Unknown scalar values use `null`, while
row-set results use empty arrays rather than `null`.

Result shapes are command-specific:

- arrays: `defs`, `outline`, `files`, `schema`, `metrics`, and `sql`
- objects: `version`, `path`, `init`, `rebuild`, `update`, `refs`, `status`,
  `stats`, and `show`; `files --explain` is also an object
- `help` returns a command-list object or one command metadata object
- `logs` returns an array of operation objects

`refs` is an object because query semantics, matching definitions, and lexical
candidates are separate arrays. Build/update objects use nullable fields when
a lock skip makes normal result values unavailable.

## Ordering and limits

Built-in navigation results use deterministic source/path ordering unless a
command documents another order. Limits must be positive and cap the relevant
row or candidate set. No result is successful.

Commands do not attach statistical relevance scores to lexical or definition
results. Dedicated commands narrow candidates; callers use `show` to retrieve
only the source context they choose.

Raw `sql` is intended for read-only composition. It accepts one statement that
starts with `SELECT`, `WITH`, or `PRAGMA` and applies lightweight
mutation-keyword checks.
This is an accidental-write guard, not a security sandbox.

## Database and lock behavior

Query commands resolve a database from explicit `--db`, effective project
configuration, or the root-keyed cache. They do not create a missing database.

While an index lock is active, queries use the previously completed database
and warn on stderr. JSON on stdout remains a valid single document. If there is
no previous database, the query fails.

## Help as the interface inventory

`mzci help --format json` is the machine-readable inventory of public
commands, usages, and summaries. It is intended to be the source used to check
README examples and design documents, though the current manually maintained
usage strings do not expose every accepted flag.

## Open Questions

- Whether a future `features` or richer help schema should expose capability
  detection instead of requiring skill-side build allowlists.
- How to keep help metadata synchronized with actual `flag.FlagSet`
  definitions. Current usage strings omit `--ignore-dir` for build commands and
  `--limit` for `defs` and `files`; the `update` summary also says "create"
  although update requires an existing index.
- Whether JSON schemas should become versioned artifacts. They are currently
  stabilized by command-specific tests rather than published schema files.
- Whether query commands should report truncation explicitly when a limit is
  reached; current results do not distinguish exact exhaustion from truncation.
- Whether `sql` should enforce read-only access with a stronger SQLite
  mechanism. The current prefix/keyword validation can admit writable PRAGMA
  forms and must not be treated as an authorization boundary.
