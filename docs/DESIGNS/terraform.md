# Terraform support

## Scope

Terraform support keeps the index lightweight: it recognizes Terraform source,
counts its comments, and uses HashiCorp's HCL parser to extract navigation
candidates without executing Terraform, loading providers, or evaluating
expressions.

The `terraform` language includes native `.tf` and `.tfvars` files and the JSON
variants `.tf.json` and `.tfvars.json`. Generic `.hcl` files are not classified
as Terraform because HCL is also used by other products.

## Metrics

Native Terraform source supports `#` and `//` line comments and `/* ... */`
block comments. Metrics recognize all three. JSON variants share the Terraform
language classification; their quoted JSON content does not look like a
comment-only line and is therefore counted as code.

## Symbols

The dedicated parser-based extractor records these top-level named blocks from
both native and JSON configuration:

| Block | Symbol kind | Symbol name |
| --- | --- | --- |
| `resource "TYPE" "NAME"` | `resource` | `NAME` |
| `data "TYPE" "NAME"` | `data` | `NAME` |
| `module "NAME"` | `module` | `NAME` |
| `variable "NAME"` | `variable` | `NAME` |
| `output "NAME"` | `output` | `NAME` |
| `provider "NAME"` | `provider` | `NAME` |
| `check "NAME"` | `check` | `NAME` |

For resources and data sources, the provider-defined type remains searchable in
the stored signature while the local label is the symbol name used by `defs`,
`outline`, and definition filtering.

Parser diagnostics do not prevent extraction from the valid portions of a
partially-written file. Symbol positions and inclusive end lines come from HCL
source ranges rather than line matching.

The extractor intentionally does not treat arbitrary assignments as symbols.
That avoids reporting nested resource arguments as definitions, but also means
that local values and `.tfvars` assignments do not currently produce symbols.

## Non-goals

- Terraform validation, formatting, dependency resolution, or provider schema
  loading
- expression evaluation, reference resolution, or a Terraform dependency graph
- generic HCL support
- nested block hierarchy

## Open Questions

- Whether local values and `.tfvars` assignments should be symbols without
  conflating arbitrary block arguments with definitions.
- Whether additional named block types have enough navigation value to store.
