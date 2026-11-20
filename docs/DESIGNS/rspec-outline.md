# RSpec outline entries

## Goal

Make `outline` useful for navigating RSpec files, where the existing Ruby
symbol extractor does not report the `describe` and `it` blocks that organize
the examples.

## Contract

- Apply RSpec extraction only to Ruby paths ending in `_spec.rb`.
- Store `describe`, `context`, and `feature` blocks as `test_group` symbols.
- Store `it`, `specify`, `scenario`, and `example` blocks as `test_example`
  symbols.
- Use the first argument's source text as the symbol name (without surrounding
  quotes for string literals) and the call line as its line. Store the block's
  end line when present.
- Keep these rows in the existing `symbols` table so `outline` and `show` can
  use the same navigation flow as other symbols.
- Continue extracting ordinary Ruby classes, modules, methods, and constants
  from the same file.

## Limits

This is static syntax-based navigation. It does not execute RSpec, infer
metadata, expand shared examples, or attempt to interpret Ruby expressions in
descriptions. Calls without a first argument are omitted.

The extractor uses RSpec's conventional method names in `_spec.rb` paths; it
does not verify that RSpec is loaded or that a matched call is actually an
RSpec DSL call.
