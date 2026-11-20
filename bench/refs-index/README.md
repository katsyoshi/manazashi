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
python3 bench/refs-index/compare.py --repo "$WORK/rails" --mzci "$WORK/mzci"
```

The script needs Python 3, Git, and the `sqlite3` executable used by `mzci`.
It creates a temporary clone for the single-file update measurement and does
not modify the supplied repository. Set `--work DIR` to retain benchmark DBs
for inspection; without it, temporary files are removed when the script ends.
Use `--repeats N` to change the query timing sample count.

## Reported measurements

- Baseline `mzci rebuild` wall time.
- Additional wall time to tokenize existing indexed lines and materialize the
  experimental table, plus the projected sum of the two.
- Main database file size before and after the overlay, including its index.
- Median wall time for the current `refs` command and the indexed prototype
  query, along with whether both returned the same candidate path and line
  pairs for the fixed workload. It reports both end-to-end CLI time and an
  in-process baseline scan versus prototype query using the same Python process
  and SQLite connection.
- Baseline one-file `mzci update` wall time and the additional time to refresh
  the overlay rows for that file, plus their projected sum.

The in-process baseline preserves the current `refs` SQL substring prefilter
and applies the same Unicode identifier-boundary rule in Python. Its timing
can be compared directly with the prototype timing. The CLI measurement is
reported separately because it includes subprocess and JSON overhead.

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
| Rebuild / overlay materialization | 5.02 s | +5.60 s additional |
| Main DB size | 133,947,392 bytes | +93,859,840 bytes (+70.07%) |
| `assert_valid_keys` query, same-process scan vs prototype | 58.00 ms | 0.16 ms |
| End-to-end `refs assert_valid_keys` CLI | 57.56 ms | Not implemented |
| One-file update | 81.66 ms via `mzci` | +2.57 ms overlay refresh |

The 31 candidate path/line pairs matched. The in-process baseline and prototype
use the same Python process and SQLite connection; the baseline applies the
current substring prefilter and exact identifier-boundary check. The CLI
measurement is separate, and the prototype is not wired into `mzci`, so the
0.16 ms result is a query-path comparison rather than an end-to-end command
measurement. These are one-machine directional results, not an adoption
decision.
