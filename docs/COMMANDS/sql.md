# `sql` command

## Purpose

`sql` runs one caller-supplied query against the index for search shapes not
covered by dedicated commands.

## Usage

```text
mzci sql [--root ROOT|--db DB] [--format text|json] [SQL]
```

If `SQL` is omitted, the entire standard input is read as the query. More than
one positional argument is rejected.

## Validation

After whitespace, the query must start with `SELECT`, `WITH`, or `PRAGMA`.
After trimming one optional trailing semicolon, another semicolon is rejected.
A case-insensitive keyword check rejects `insert`, `update`, `delete`, `drop`,
`alter`, `create`, `replace`, `vacuum`, `attach`, and `detach`.

This is an accidental-write guard, not a security boundary or a complete SQL
parser. Only trusted queries and database paths should be used.

## Output

Text enables SQLite headers and tab-separated mode. JSON returns an array of
objects whose fields and types come from the query:

```json
[{"path":"cmd.go","line":51,"name":"Run"}]
```

SQLite integers and real values remain JSON numbers; null remains null. Use
unique column names or explicit aliases because duplicate names are ambiguous
as object keys. No rows returns `[]`.

An active mutation lock causes a warning on stderr while stdout remains a valid
result document from the previous database.

## Examples

```sh
mzci sql \
  "select path, line, name from symbols where kind = 'function' limit 20"

printf '%s\n' "select count(*) as files from files" |
  mzci sql --format json
```

## Related commands

Use `schema` to discover tables and columns. Prefer dedicated commands when
their stable result shape fits the task.

## Open Questions

- Writable PRAGMA forms can pass the current validator. Strong read-only
  enforcement requires a SQLite-level mechanism.
- Whether query plans, parameter binding, or saved query templates are useful
  without turning the CLI into a general database shell.
