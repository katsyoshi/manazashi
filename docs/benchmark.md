# Reference benchmarks

These measurements provide points of reference for real-world repositories.
They are not performance guarantees; results depend on the repository,
hardware, and software environment.

## Rails

Measured on 2026-07-24 (UTC):

- Rails `main`: `b6cb9ea7657ea46eb35fdd73b9b2174f7a93a8c0`
- code-index: `8af82fc792c514d78ab038b6bd3879929bb101f8`
- Git-tracked files: 4,971
- Indexed files: 4,854
- Indexed lines: 684,668
- Symbols: 52,304
- Full rebuild: 92.66 seconds
- No-change incremental update: 0.04 seconds
- SQLite database size: 128 MB

## Go

Measured on 2026-07-24 (UTC):

- Go `master`: `a961f702a48edbfc044639775f4ffae692b7f0dc`
- code-index: `039d23b24bc5b58e9b1446ff53ad1871f4ff34b8`
- Git-tracked files: 15,671
- Indexed files: 13,660
- Indexed lines: 2,832,488
- Symbols: 218,928
- Full rebuild: 10.43 seconds
- No-change incremental update: 0.05 seconds
- SQLite database size: 468 MB
