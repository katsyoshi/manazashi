# `refs` identifier-index experiment

This experiment estimates whether a derived identifier-per-line table is worth
adding to the production index. It compares the current `refs` command with a
temporary SQLite overlay containing one row per distinct identifier on each
source line. The overlay is created only in the benchmark database; production
schema and query behavior are unchanged.

The fixed workload matches the existing context benchmark: Rails at revision
`b6cb9ea7657ea46eb35fdd73b9b2174f7a93a8c0`, query `assert_valid_keys`, and a
single-file update to
`activesupport/lib/active_support/core_ext/hash/keys.rb`.

## Run

```sh
WORK=/tmp/manazashi-refs-index
git init -q "$WORK/rails"
git -C "$WORK/rails" remote add origin https://github.com/rails/rails.git
git -C "$WORK/rails" fetch -q --depth 1 origin b6cb9ea7657ea46eb35fdd73b9b2174f7a93a8c0
git -C "$WORK/rails" checkout -q FETCH_HEAD
go build -o "$WORK/mzci" ./cmd/mzci
ruby bench/refs-index/compare.rb --repo "$WORK/rails" --mzci "$WORK/mzci"
```

The script needs Ruby's standard library, Git, and the `sqlite3` executable
used by `mzci`. It creates a temporary clone for the single-file update
measurement and does not modify the supplied repository. Set `--work DIR` to
retain benchmark outputs for inspection; without it, temporary files are
removed when the script ends. Use `--repeats N` to change the query timing
sample count.

## Reported measurements

- Baseline `mzci rebuild` wall time.
- Additional wall time to tokenize existing indexed lines and materialize the
  experimental table, plus the projected sum of the two.
- Main database file size before and after the overlay, including its index.
- Median wall time for the current `refs` command and the indexed prototype
  query, along with whether both returned the same candidate path and line
  pairs for the fixed workload. It reports both end-to-end CLI time and a
  baseline scan versus prototype query through the same sqlite3 CLI/CSV path.
- Baseline one-file `mzci update` wall time and the additional time to refresh
  the overlay rows for that file, plus their projected sum.

The baseline scan preserves the current `refs` SQL substring prefilter and
applies the same Unicode identifier-boundary rule in Ruby. The baseline scan
and prototype both invoke `sqlite3 -csv -noheader`, parse results with Ruby's
CSV library, and use the same timer. This makes their execution path
comparable; the end-to-end CLI measurement is reported separately because it
includes `mzci`'s orchestration and JSON formatting overhead.

The overlay tokenizer intentionally scans all source text, so identifiers in
comments and strings remain searchable like they are in the current `refs`
contract. It stores a distinct identifier once per line because `refs` returns
candidate lines rather than individual token positions.

This is a sizing experiment, not a production implementation. The overlay is
materialized from the completed line index, so the projected rebuild/update
times are additive estimates. Run multiple times on an otherwise idle machine.
The database size reports the main SQLite file, excluding transient sidecar
files.

## Preliminary Rails snapshot

One run on revision `b6cb9ea7657ea46eb35fdd73b9b2174f7a93a8c0` found:

| Measurement | Current index | With identifier-line overlay |
| --- | ---: | ---: |
| Rebuild / overlay materialization | 5.13 s | +5.08 s additional |
| Main DB size | 133,947,392 bytes | +87,719,936 bytes (+65.49%) |
| `assert_valid_keys` query, same sqlite3 CSV path | 71.30 ms scan | 2.14 ms indexed |
| End-to-end `refs assert_valid_keys` CLI | 54.00 ms | Not implemented |
| One-file update | 80.72 ms via `mzci` | +9.46 ms overlay refresh |

The 31 candidate path/line pairs matched. The scan and indexed queries both
use the sqlite3 CSV subprocess and Ruby CSV parser; the scan also applies the
current exact identifier-boundary check. The end-to-end CLI result is listed
separately. These are one-machine directional results, not an adoption
decision.
