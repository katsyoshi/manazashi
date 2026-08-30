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

Ruby uses the embedded Tree-sitter Ruby grammar. It extracts classes, modules,
methods, singleton method names, and unqualified constant assignments. Positions
and end lines come from syntax-tree nodes. The embedded parser is authoritative
for Ruby; extraction does not require an installed Ruby command and does not
fall back to regular-expression symbol extraction.

Rust uses the embedded Tree-sitter Rust grammar. It extracts modules, structs,
unions, enums, traits, type aliases, constants, statics, free functions, trait
and implementation methods, and `macro_rules!` definitions. Positions and end
lines come from syntax-tree nodes. Functions nested below a trait or `impl` are
methods; other functions are free functions. The extractor records explicit
source definitions only: it does not expand declarative macros, execute
procedural macros or build scripts, or evaluate Cargo features and `cfg`
expressions. The embedded parser is authoritative for Rust; Rust does not fall
back to regular-expression symbol extraction.

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

- Whether the cross-language kind vocabulary needs tighter rules for constructs
  such as Java constructors, TypeScript members, and Swift protocols.
- Whether qualified names or explicit ownership are needed by `defs` or
  `outline`; they should not be added solely to enrich `refs`.
- Whether comments, strings, and executable source ever need lexical-region
  classification outside the symbol table.
- How language-specific metaprogramming signals should be represented if a
  concrete navigation use case appears.
