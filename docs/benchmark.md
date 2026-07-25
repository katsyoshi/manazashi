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
- code-index: `0a55e39fa7366beca0272457547b05f987eeab8b`
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

## Go

Measured on 2026-07-24 (UTC):

- Go `master`: `a961f702a48edbfc044639775f4ffae692b7f0dc`
- code-index: `0a55e39fa7366beca0272457547b05f987eeab8b`
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
