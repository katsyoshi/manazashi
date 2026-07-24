# Symbols and definition navigation

## Symbol contract

A symbol is a statically recognizable definition. Stored rows contain:

- repository-relative path and language
- normalized kind and name
- one-based definition line, end line, and column
- a bounded signature
- bounded nearby source context for search

Symbols are navigation candidates rather than resolved semantic entities.
Names are not globally unique, ownership is not stored as a graph, and row IDs
are not stable across rebuilds.

Comments, string literals, call sites, fields inferred from assignment, and
runtime-generated definitions are not symbols. Comment counts belong to
metrics. Comments and strings may still appear as lexical candidates in
`refs`.

## Extraction

Go uses the standard Go parser. It distinguishes receiver methods from
functions and extracts types, variables, and constants. End lines come from
the parsed declaration.

Ruby invokes the available `ruby` command with source on stdin and a five-second
timeout. It prefers a Prism parse-tree dump, falls back to the traditional
parse-tree dump, and extracts classes, modules, methods, singleton method names,
and constant writes. If Ruby parsing is unavailable or fails, extraction falls
back to the language's regular-expression patterns.

Other supported languages use intentionally lightweight regular-expression
patterns. These patterns identify common declaration forms and may miss valid
syntax or report unusual syntax imperfectly. Their end line defaults to the
definition line.

Extraction never executes the indexed source as a program.

## Kinds

The shared vocabulary currently includes kinds such as:

`class`, `module`, `type`, `interface`, `trait`, `enum`, `function`, `method`,
`macro`, `constant`, and `variable`.

A function is not owned by a receiver/type construct; a method is. A macro
remains distinct because source expansion is not an ordinary call. Languages
use only distinctions that their extractor can recognize reliably.

Kinds are stored without collapsing language-specific definitions into one
generic callable category. Callers may query several kinds separately when
they need a broader concept.

## Command responsibilities

- `defs` lists or searches definition rows. Its query mode is deliberately
  broader than exact-name matching and may match name prefixes, signatures, or
  paths.
- `outline` lists stored definitions for one file in source order.
- `refs` reports exact-identifier lexical occurrences and matching definitions;
  it does not resolve occurrences to definitions. See
  [`refs`](../COMMANDS/refs.md).
- `show` retrieves indexed source around a position selected from another
  command.

Structural hierarchy is not part of `refs`. If a caller needs file structure,
it should query `outline`; if it needs local context, it should query `show`.

## Metaprogramming

Macros are explicit definitions and may be stored as `symbols.kind = macro`.
Metaprogramming operations such as Ruby `define_method` are not normal static
definitions and are not currently stored. The design reserves the conceptual
term `metaprogramming` for a future signal, not a symbol kind that merges with
macros.

## Open Questions

- Which languages justify a parser-backed extractor beyond Go and Ruby.
- Whether the cross-language kind vocabulary needs tighter rules for constructs
  such as Java constructors, TypeScript members, Rust `impl` methods, and Swift
  protocols.
- Whether qualified names or explicit ownership are needed by `defs` or
  `outline`; they should not be added solely to enrich `refs`.
- Whether comments, strings, and executable source ever need lexical-region
  classification outside the symbol table.
- How language-specific metaprogramming signals should be represented if a
  concrete navigation use case appears.
