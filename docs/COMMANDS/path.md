# `path` command

## Purpose

`path` prints the database path that normal commands would select without
creating the database.

## Usage

```text
mzci path [--format text|json] [ROOT]
```

- `ROOT` is optional. Without it, the current directory is promoted to its
  containing Git root when possible.
- An explicit root is resolved as the exact absolute directory; it is not
  promoted to a containing Git root.
- `--format` defaults to `text`.

## Resolution

The command loads effective configuration for the resolved root. A project
`db` value wins; otherwise it prints the root-keyed cache path described in
[configuration design](../DESIGNS/configuration.md).

Text prints the path followed by a newline. JSON returns:

```json
{"path": "/cache/manazashi/0123456789abcdef.sqlite"}
```

The command rejects more than one root, an invalid directory, invalid
configuration, or an unsupported format.

## Examples

```sh
mzci path
mzci path /path/to/repo
mzci path --format json
```

## Related commands

`init` creates configuration and an empty database. Query commands also accept
`--db` as a one-run override.

## Open Questions

- Whether an explicit nested root should resolve to its containing Git root,
  matching `init`.
