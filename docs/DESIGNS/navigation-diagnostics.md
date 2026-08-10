# Navigation diagnostics and narrowing

## Motivation

Navigation commands should let an agent recover the next relevant fact without
keeping a deep codebase context active throughout a task. They return compact
results so the agent can progressively retrieve detail rather than load broad
source context up front.

An empty indexed row set alone cannot distinguish a typo from an untracked,
excluded, or encoding-skipped file. Short symbol names also need a lightweight
way to avoid unrelated namespaces in other paths.

## Path narrowing

`defs` and `refs` accept `--path QUERY`. The filter is a case-insensitive
substring match against repository-relative indexed paths. It combines with
the existing kind and language filters.

For `refs`, the filter applies to both definitions and lexical candidates and
is recorded as nullable `query.path`. It remains lexical filtering rather than
symbol resolution.

## File diagnostics

`files --explain QUERY` preserves the normal path search but wraps JSON output
in an object containing `query`, `found`, `reason`, and `matches`. It is only
valid with JSON output and without `--list`.

When matches exist, `found` is true and `reason` is null. Otherwise `reason`
is one of:

- `missing`: the requested repository-relative path does not exist;
- `untracked`: it exists under the indexed root but Git does not track it;
- `skipped`: it has retained file metadata but no indexed source;
- `excluded`: Git tracks it, but the effective indexing configuration omitted
  it;
- `unknown`: the database does not provide enough root metadata to classify
  it safely.

Diagnostics classify a normalized exact repository-relative path. A general
substring query that happens not to match may therefore report `missing`.
Callers should use `--explain` with a path when the distinction matters.

## Compact `show` JSON

JSON `show` returns one object rather than one object per line:

```json
{
  "path": "lib/config.rb",
  "found": true,
  "reason": null,
  "start_line": 40,
  "lines": ["def load", "  parse", "end"]
}
```

The selected stored path and first returned line are emitted once. `lines` is
always an array. An indexed path with a range outside stored source has
`found: true` and an empty array. An unresolved path uses the same diagnostic
reasons as `files --explain`, with `path` set to null.

Text output remains tabular for compatibility.

## Build identity naming

`version` names the build-tree dirty boolean `build_modified` in both text and
JSON. The old ambiguous `modified` field is removed. This value describes the
source tree used to build `mzci`, never the repository currently being
indexed.

## Non-goals

- Resolving a reference to a namespace or definition
- Indexing untracked files
- Turning file diagnostics into a general filesystem search
- Returning source that is not already in the completed index

## Skill self-containment

The reusable skill carries its required operating principles directly in
`SKILL.md`. Normal use must not depend on locating the Manazashi repository
README or fetching it over the network. The root README's `Design` section is
additional contributor context only when the target checkout is Manazashi
itself.
