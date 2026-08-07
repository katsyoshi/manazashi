# `help` command

## Purpose

`help` lists public commands or describes one command. Its JSON form is the
machine-readable CLI inventory.

## Usage

```text
mzci help [--format text|json] [COMMAND]
```

- `COMMAND` is optional and must be a known command name.
- `--format` defaults to `text`.

## Behavior and output

Without `COMMAND`, text output prints top-level usage and one summary line per
command. JSON returns:

```json
{
  "usage": "mzci <help|version|...> [options]",
  "commands": [
    {"name": "status", "usage": "mzci status ...", "summary": "..."}
  ]
}
```

With `COMMAND`, text prints its usage and summary. JSON returns one object with
the string fields `name`, `usage`, and `summary`.

An unknown command, more than one positional argument, or an unsupported
format is an error.

## Examples

```sh
mzci help
mzci help status
mzci help --format json
mzci help --format json status
```

## Related commands

`version` reports build compatibility information rather than command usage.

## Open Questions

- Usage strings are manually maintained and currently omit some accepted
  flags: build commands accept `--ignore-dir`, while `defs` and `files` accept
  `--limit`.
- The `update` summary says it may create an index although the implementation
  requires an existing database.
