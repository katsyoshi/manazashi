# `schema` command

## Purpose

`schema` lists user-facing index tables and columns so callers can compose
read-only SQL without inspecting SQLite internals.

## Usage

```text
mzci schema [--root ROOT|--db DB] [--format text|json]
```

Positional arguments are rejected. `--format` defaults to `text`.

## Behavior and ordering

The command includes ordinary user tables and FTS5 virtual tables. SQLite
system tables and FTS5 shadow tables are omitted. Rows sort by table name and
column ordinal.

The reported schema is the selected database's actual schema, which may differ
from the current binary before compatibility is checked.

## Output

Text is tabular. JSON rows contain:

| Field | Type | Meaning |
| --- | --- | --- |
| `table_name` | string | User-facing table or virtual table |
| `ordinal` | number | One-based column position |
| `column_name` | string | Column name |
| `type` | string | Declared SQLite type, or `-` when unavailable |
| `nullable` | boolean | Whether null values are allowed |
| `key` | string or null | `primary(N)` for primary-key columns |

No rows returns `[]`.

## Examples

```sh
mzci schema
mzci schema --format json --db /tmp/index.sqlite
```

## Related commands

Use `sql` to query the described schema and the
[storage design](../DESIGNS/storage.md) for table responsibilities.

## Open Questions

- Whether indexes, foreign keys, check constraints, and schema-version metadata
  should be exposed without leaking SQLite implementation tables.
