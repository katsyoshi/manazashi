# Context-size benchmark

Compares how much text an agent receives when navigating a repository with
manazashi queries versus text search and whole-file reads. Each task runs a
fixed command sequence for both conditions and measures the stdout of every
step. It does not run an agent, call an API, or measure time.

## Usage

```sh
WORK=/tmp/manazashi-context
git init -q "$WORK/rails"
git -C "$WORK/rails" remote add origin https://github.com/rails/rails.git
git -C "$WORK/rails" fetch -q --depth 1 origin b6cb9ea7657ea46eb35fdd73b9b2174f7a93a8c0
git -C "$WORK/rails" checkout -q FETCH_HEAD
go build -o "$WORK/mzci" ./cmd/mzci
"$WORK/mzci" rebuild --db "$WORK/rails.sqlite" "$WORK/rails"
bench/context/run.sh --steps --repo "$WORK/rails" --db "$WORK/rails.sqlite" \
  --mzci "$WORK/mzci" bench/context/rails-tasks.sh
```

`--steps` adds a per-step table. The script exits non-zero if a condition
does not surface its task's target path. `bench/context/test.sh` verifies the
harness against a small fixture repository.

## Procedure rules

- Characters are Unicode code points (`wc -m` under `en_US.UTF-8`); bytes and
  lines are also recorded. The ratio is with/without characters.
- `mzci` runs with `--format json` and default limits, matching the manazashi
  skill.
- A whole-file read emulates an agent Read: the first 2,000 lines with
  `cat -n` style line-number prefixes.
- Both conditions search the same scope. When a search returns several files,
  the without condition reads the file holding the intended hit, as an agent
  choosing from the listing would; it does not read every hit.
- The without condition always reads whole files, even where `rg -n` already
  reports the line; it does not use a ranged read. The with condition reads a
  bounded range with `show`.
- Command lines are used for the target check but are not measured. A step
  with empty output fails the target check.

## Targets

Targets are commonly visited Rails core APIs, chosen before measuring rather
than for their output size.

| Task | Target | Why |
| --- | --- | --- |
| `def_body` | `ActiveRecord::Persistence#save` | Common method name with many same-named definitions in Active Record. |
| `refs` | `assert_valid_keys` | Moderately used helper with a single definition. |
| `find_file` | `routing/route_set.rb` | File lookup by name fragment. |
| `outline` | `core_ext/string/inflections.rb` | Mid-sized file whose structure is the question. |
| `method_in_class` | `ActionController::Metal#dispatch` | Known file, one method of interest. |
| `def_body_ambiguous` | `ActiveModel::Errors#initialize` | Extremely common name; `defs` hits its default limit of 50. |

In `def_body`, the first `rg` hit in path order is a comment in
`attribute_methods.rb`, not the definition; the without condition still reads
`persistence.rb` under the rule above.
