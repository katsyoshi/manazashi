# Reference benchmarks

These measurements provide points of reference for real-world repositories.
They are not performance guarantees; results depend on the repository,
hardware, and software environment.

## Rails

Measured on July 24, 2026 (UTC; July 25 in JST):

- Rails `main`: `b6cb9ea7657ea46eb35fdd73b9b2174f7a93a8c0`
- code-index: `8af82fc792c514d78ab038b6bd3879929bb101f8`
- Git-tracked files: 4,971
- Indexed files: 4,854
- Indexed lines: 684,668
- Symbols: 52,304
- Full rebuild: 92.66 seconds
- No-change incremental update: 0.04 seconds
- SQLite database size: 128 MB
