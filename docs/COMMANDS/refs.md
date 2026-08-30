# `refs` command

## Purpose

`refs NAME` finds likely occurrences of an identifier in indexed source code.
It is intended for both interactive human use and agent/script use.

The command is a lightweight navigation aid. It does not resolve names to a
specific definition and does not build a reference, call, inheritance, or type
graph.

## Usage

```text
mzci refs [--root ROOT|--db DB] [--kind KIND]... [--language LANG]
                [--path QUERY] [--ignore-case] [--limit N] [--format text|json] NAME
```

- `NAME` is required.
- Matching is case-sensitive by default.
- `--ignore-case` enables case-insensitive matching.
- `--kind` accepts an existing symbol kind such as `class`, `module`, `method`,
  `function`, `macro`, `type`, `interface`, `trait`, `enum`, or `constant`.
- `--kind` may be repeated to search several kinds without collapsing their
  stored kinds, for example `--kind function --kind method`.
- With no `--kind`, definitions of every kind are considered.
- `--kind` filters matching definitions and records the caller's search intent.
  It does not prove that every occurrence resolves to a definition of one of
  those kinds. This matters when the definition belongs to a dependency that
  is not indexed.
- `--language` restricts definitions and candidates to one indexed language.
- `--path` restricts definitions and candidates to repository-relative paths
  containing the query, case-insensitively.
- `--limit` limits reference candidates, not matching definitions. Its default
  is 100 and it must be positive.
- `--format` defaults to `text`.

`NAME` accepts one identifier per invocation. Extra positional arguments are
rejected; run separate queries when inspecting several identifiers.

No results is a successful query. Text output shows empty sections and JSON
uses empty arrays.

Invalid/missing `NAME`, non-positive `--limit`, an empty `--kind`, extra
positional arguments, invalid database/configuration, or unsupported format is
an error.

## Matching contract

The default query matches `NAME` as a complete identifier, not as a substring.

For example, `refs List` matches the `List` token in:

```java
List<File> fileList = new ArrayList<>();
java.util.List<File> files;
```

It does not match `fileList`, `ArrayList`, `Listing`, or `list`. With
`--ignore-case`, the standalone identifier `list` also matches, while
`fileList` and `ArrayList` still do not.

Identifier boundaries should follow the indexed language when practical. A
fallback matcher must at least treat Unicode letters and digits, `_`, and `$`
as identifier characters so a partial identifier is not reported as an exact
match.

Lines containing an exact matching symbol definition are excluded from
reference candidates. Comments and string literals remain searchable and are
returned when they contain the requested identifier. This is intentional: the
command reports lexical candidates, and documentation or examples can be useful
navigation results even when they are not executable references.

Multiple definitions with the same name are not disambiguated. All matching
definitions and all lexical candidates are returned subject to filters and the
candidate limit.

The result is a snapshot of the index. Use `status` to check freshness and
read the live checkout before editing or relying on source details. Use
`--kind`, `--language`, and `--path` to reduce same-name results.

## Symbol kinds

Symbol kinds retain distinctions that affect their meaning and lexical
hierarchy. They are not normalized into one broad callable kind:

- `function` is a callable definition that is not owned by a class, type,
  receiver, or equivalent language construct.
- `method` is a callable definition owned by a class, type, receiver, `impl`,
  or equivalent language construct.
- `macro` is a source-expansion definition and is separate from both
  `function` and `method`.

Languages map their native constructs to these common kinds only where the
distinction is meaningful. For example, Go distinguishes functions from
receiver methods, Java definitions are methods, and C definitions are
functions or macros.

A macro remains `macro` regardless of what its expansion may produce. A C or
C++ macro that expands to a class declaration, method declaration, function,
constant, or arbitrary statements is not reclassified as any of those kinds.
`refs --kind macro NAME` searches occurrences of the macro name; it does not
run a preprocessor or infer symbols created by expansion.

Metaprogramming is conceptually separate from macro symbols. An explicitly
defined macro is a symbol with `symbols.kind = macro`. A metaprogramming
construct that cannot be treated as a normal static definition may eventually
be represented as a signal such as `signals.kind = metaprogramming`.

```text
symbols.kind = macro
signals.kind = metaprogramming
```

No `signals` table exists today. These concepts must not be collapsed into a
stored `macro_or_metaprogramming` symbol kind if signals are implemented.

Language-specific metaprogramming details are outside this command design and
should be defined when support for a language is implemented. For example,
Ruby may record an operation such as `define_method`, but `refs` does not infer
the generated method as a normal static symbol.

## Text output

Text output is optimized for interactive navigation. It separates matching
definitions from reference candidates and groups candidates by file.

```text
query: List (kinds: class, case-sensitive)

definitions:
  lib/list.rb:3  class List

references:
  app/models/item.rb
    12  List.new

  app/services/load.rb
    8  parser.call(List)
```

If no indexed definition exists, the definitions section remains empty while
lexical candidates are still shown. This supports references to standard
libraries and external dependencies.

## JSON output

JSON returns one object rather than a bare row array so query semantics,
definitions, and candidates remain distinct.

```json
{
  "query": {
    "name": "List",
    "kinds": ["class"],
    "language": "java",
    "path": null,
    "case_sensitive": true,
    "limit": 100
  },
  "definitions": [
    {
      "path": "lib/list.java",
      "line": 3,
      "kind": "class",
      "name": "List",
      "language": "java",
      "signature": "class List {"
    }
  ],
  "candidates": [
    {
      "path": "app/models/item.java",
      "line": 12,
      "language": "java",
      "text": "List<Item> items = new ArrayList<>();"
    }
  ]
}
```

Nullable query filters use `null`. Row collections always use arrays, including
when empty. Candidate ordering is stable by path and line. Definition ordering
is stable by path, line, column, and name.

When no kind filter is supplied, `query.kinds` is an empty array. An empty
array means all kinds; it is distinct from an unknown or unavailable value.

## Related commands

- `defs NAME` discovers definitions and may remain broader or fuzzier than
  `refs`.
- `outline PATH` shows definitions and structure in one file.
- `show PATH --line N` retrieves context around a selected candidate.
- `sql` and FTS remain available for substring and custom searches such as
  finding `fileList` or `ArrayList` from the text `List`.

## Examples

```sh
mzci refs UserRepository
mzci refs --kind class --language ruby UserRepository
mzci refs --ignore-case --limit 20 --format json run_task
```

## Non-goals

- Resolving a candidate to one definition
- Caller/callee edges
- Inheritance or implementation graphs
- Dynamic dispatch analysis
- Expanding macros or inferring classes, methods, functions, or other symbols
  produced by a macro
- Defining one cross-language schema for language-specific metaprogramming
  operations
- Inferring a reference from names such as `fileList` when `List` is absent
- Classifying comments or string literals separately from other lexical
  candidates
- Reporting scores, relevance rankings, occurrence statistics, or per-file
  reference counts
- Replacing a language server

## Open Questions

The command interface does not require a particular storage design. Any future
addition of structural context should be based on demonstrated navigation
needs and must preserve the lightweight, rebuildable index design.
