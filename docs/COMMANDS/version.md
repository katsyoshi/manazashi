# `version` command

## Purpose

`version` identifies the binary build and its index compatibility constants.

## Usage

```text
mzci version [--format text|json]
```

`--format` defaults to `text`. Positional arguments are rejected.

## Output

Text is a `key<TAB>value` table. JSON returns:

```json
{
  "commit": "0123456789abcdef",
  "modified": false,
  "schema_version": 4,
  "file_source": "git-tracked"
}
```

| Field | Type | Meaning |
| --- | --- | --- |
| `commit` | string or null | VCS revision embedded by Go build information or linker flags |
| `modified` | boolean or null | Whether the build source tree was modified |
| `schema_version` | number | Schema version written by this binary |
| `file_source` | string | Source selection contract, currently `git-tracked` |

`commit` and `modified` are null when build metadata is unavailable. A commit
hash is an identity, not an ordered version number. Text uses `unknown` when no
commit is available and omits `modified` when it is unknown.

## Examples

```sh
mzci version
mzci version --format json
```

## Related commands

Use `status` to compare a database with the current tool and checkout.

## Open Questions

- Whether a future `features` command should expose capabilities more directly
  than commit allowlists and schema constants.
