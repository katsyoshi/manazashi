# Reference benchmarks

These measurements provide points of reference for real-world repositories.
They are not performance guarantees; results depend on the repository,
hardware, and software environment.

## Environment

- CPU: AMD Ryzen 9 7950X3D, 16 cores and 32 threads
- Memory: 124 GiB
- Storage: KIOXIA EXCERIA PRO NVMe SSD with Btrfs
- OS: Linux x86_64, kernel 7.1.4-gentoo
- SQLite: 3.53.1
- Go: 1.26.4
- Ruby: 4.0.6 with Prism
- hyperfine: 1.20.0

Query measurements run after three explicit warm-up iterations. The repository
checkouts had also been read and indexed before the recorded rebuilds, so the
rebuild measurements may benefit from a warm filesystem cache. They should not
be interpreted as cold-cache disk benchmarks.

Query timings use three warm-up runs followed by 30 measured runs. The tables
report the median and nearest-rank p95 elapsed time. Command output is discarded
after generation, so serialization is included while terminal rendering is not.
Search commands use a limit of 100,000 to include all results in these indexes.

Targets are selected from each index rather than fixed across repositories:

- `defs` uses the name with the most exact definitions.
- `refs` tests the 100 most-defined simple identifiers of at least three
  characters, then uses the one with the most reference candidates.
- `files` uses the path component present in the most indexed files.
- `outline` uses the file with the most symbols.
- `show` uses the symbol nearest the middle of that file.

## Rails

Measured on 2026-07-24 (UTC):

- Rails `main`: `b6cb9ea7657ea46eb35fdd73b9b2174f7a93a8c0`
- manazashi: `0a55e39fa7366beca0272457547b05f987eeab8b`
- Git-tracked files: 4,971
- Indexed files: 4,854
- Indexed lines: 684,668
- Symbols: 52,304
- Full rebuild: 92.52 seconds
- No-change incremental update: 0.04 seconds
- SQLite database size: 128 MB

| Command | Selected target | Results | Median | p95 |
| --- | --- | ---: | ---: | ---: |
| `defs` | `initialize` (1,169 exact definitions) | 1,419 rows | 23.06 ms | 24.41 ms |
| `refs` | `name` | 34 definitions, 17,888 candidates | 4.439 s | 4.477 s |
| `files` | path component `test` (2,728 files) | 2,857 rows | 15.36 ms | 16.17 ms |
| `outline` | `guides/assets/javascripts/@hotwired--turbo.js` | 618 symbols | 11.94 ms | 12.79 ms |
| `show` | `scrollToAnchor` at line 1,969 | 7 lines | 4.41 ms | 4.97 ms |
| `stats` | index summary | 1 object | 8.21 ms | 9.04 ms |
| `metrics` | default output | 7 rows | 5.43 ms | 5.98 ms |
| `status` | index status | 1 object | 16.21 ms | 16.87 ms |

### Rails merge-history updates

Measured on 2026-08-10 (UTC) with Manazashi
`5053ef8ef4d3da3840dccef31bdd7b07338e6131`, schema 4, and FTS5 enabled.
The Rails history ended at the same `b6cb9ea7` revision used above.

This benchmark replays first-parent merge commits to approximate changes as
they reached the main branch. For each window, it rebuilds at the first parent
of the oldest selected merge, then checks out and incrementally updates through
each merge in chronological order. Checkout time is excluded from update time.
Direct main-branch commits between selected merges are included in the next
update interval.

The early-history window contains the first 105 first-parent merges. Later
windows contain the 100 merges ending near each named year boundary; the 2026
window ends at `b6cb9ea7`. Measurements used a local shared clone and warm
filesystem caches.

| Window | Merges | Baseline rebuild | Update total | Update median | Update p95 | Update max | Changed files median | Changed files p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 2008–2010 early history | 105 | 5.12 s | 104.42 s | 572 ms | 3.202 s | 4.469 s | 43 | 422 |
| 2011 year-end | 100 | 9.17 s | 11.62 s | 80 ms | 261 ms | 878 ms | 3 | 18 |
| 2014 year-end | 100 | 11.00 s | 10.78 s | 71 ms | 311 ms | 728 ms | 2 | 19 |
| 2017 year-end | 100 | 13.46 s | 13.23 s | 95 ms | 308 ms | 899 ms | 2 | 13 |
| 2020 year-end | 100 | 16.74 s | 12.07 s | 97 ms | 235 ms | 913 ms | 2 | 11 |
| 2023 year-end | 100 | 19.06 s | 13.67 s | 90 ms | 267 ms | 2.014 s | 2 | 10 |
| 2026 latest | 100 | 21.06 s | 15.03 s | 119 ms | 407 ms | 608 ms | 2 | 13 |

Across the windows, Pearson correlation between changed-file count and update
time ranges from 0.953 to 0.985. Normal merge updates remain small as the
repository grows; the early-history outliers are large branch integrations of
hundreds of files rather than current-style small pull requests.

At the end of the 2026 window, SQLite `dbstat` accounted for 132.4 MiB of
pages. The `files_fts` tables used 36.4 MiB and `symbols_fts` used 25.5 MiB,
about 47% combined. This establishes storage share, not component update cost.
The history results do not justify table-selective updates: update time is
already strongly explained by the number of changed files, while splitting
dependent projections would require separate freshness tracking. Measure
internal build phases before considering component-specific update behavior.

## Go

Measured on 2026-07-24 (UTC):

- Go `master`: `a961f702a48edbfc044639775f4ffae692b7f0dc`
- manazashi: `0a55e39fa7366beca0272457547b05f987eeab8b`
- Git-tracked files: 15,671
- Indexed files: 13,660
- Indexed lines: 2,832,488
- Symbols: 218,928
- Full rebuild: 10.29 seconds
- No-change incremental update: 0.06 seconds
- SQLite database size: 468 MB

| Command | Selected target | Results | Median | p95 |
| --- | --- | ---: | ---: | ---: |
| `defs` | `main` (2,128 exact definitions) | 3,960 rows | 59.58 ms | 64.17 ms |
| `refs` | `test` | 69 definitions, 18,373 candidates | 31.694 s | 32.100 s |
| `files` | path component `src` (10,036 files) | 10,036 rows | 37.75 ms | 40.04 ms |
| `outline` | `src/syscall/zerrors_linux_loong64.go` | 1,915 symbols | 31.83 ms | 32.64 ms |
| `show` | `PR_SET_MM_MAP` at line 1,050 | 7 lines | 5.37 ms | 5.61 ms |
| `stats` | index summary | 1 object | 17.46 ms | 18.89 ms |
| `metrics` | default output | 11 rows | 8.21 ms | 8.82 ms |
| `status` | index status | 1 object | 18.29 ms | 19.09 ms |

### Go commit-history updates

Measured on 2026-08-10 (UTC) with Manazashi
`5053ef8ef4d3da3840dccef31bdd7b07338e6131`, schema 4, and FTS5 enabled.
The Go history ended at the same `a961f702` revision used above.

Go has only 40 first-parent merge commits in this history, so the Rails merge
methodology is not representative. Each Go window instead contains 100
consecutive first-parent commits ending near the named year boundary. The
benchmark rebuilds at the parent of the oldest selected commit, then checks out
and incrementally updates through the window. Checkout time is excluded from
update time. Measurements used a local shared clone and warm filesystem caches.

| Window | Commits | Baseline rebuild | Update total | Update median | Update p95 | Update max | Indexed files median | Indexed files p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 2009 year-end | 100 | 0.56 s | 4.84 s | 41 ms | 87 ms | 150 ms | 1 | 6 |
| 2011 year-end | 100 | 0.89 s | 4.88 s | 42 ms | 79 ms | 179 ms | 0 | 5 |
| 2014 year-end | 100 | 3.11 s | 13.13 s | 77 ms | 390 ms | 1.736 s | 1 | 12 |
| 2017 year-end | 100 | 5.57 s | 13.83 s | 102 ms | 317 ms | 731 ms | 1 | 7 |
| 2020 year-end | 100 | 6.84 s | 24.40 s | 130 ms | 890 ms | 2.150 s | 1 | 18 |
| 2023 year-end | 100 | 8.26 s | 22.33 s | 135 ms | 631 ms | 1.923 s | 1 | 10 |
| 2026 latest | 100 | 10.25 s | 39.42 s | 222 ms | 1.155 s | 5.894 s | 2 | 16 |

The 2009 and 2011 trees keep canonical Go source under `src/pkg`. Manazashi's
built-in ignored directory names include `pkg`, so many source changes in those
windows are absent from the index. These timings describe the current default
configuration against historical trees, not complete indexing of early Go
source. The configuration currently permits adding ignored names but not
removing built-in names.

Correlation between files actually added, updated, or deleted in the index and
update time is 0.850 for 2009 and ranges from 0.961 to 0.997 for the later
windows. The 2026 maximum is a compiler refactor that changes 73 indexed files.

At the end of the 2026 window, SQLite `dbstat` accounted for 477.7 MiB of
pages. The `files_fts` tables used 124.4 MiB and `symbols_fts` used 98.1 MiB,
about 47% combined. As with Rails, this does not isolate FTS update time. The
history measurements show that normal updates remain small and that large
source changes, rather than total repository size alone, dominate the outliers.
