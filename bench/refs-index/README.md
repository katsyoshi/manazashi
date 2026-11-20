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
  pairs for the fixed workload.
- Baseline one-file `mzci update` wall time and the additional time to refresh
  the overlay rows for that file, plus their projected sum.

The overlay tokenizer intentionally scans all source text, so identifiers in
comments and strings remain searchable like they are in the current `refs`
contract. It stores a distinct identifier once per line because `refs` returns
candidate lines rather than individual token positions.

This is a sizing experiment, not a production implementation. The overlay is
materialized from the completed line index, so the projected rebuild/update
times are additive estimates. The prototype query uses Python's SQLite API,
while `mzci refs` invokes the CLI-based query path; their wall times include
different process and JSON overhead. Run multiple times on an otherwise idle
machine and treat query latency as directional. The database size reports the
main SQLite file, excluding transient sidecar files.
