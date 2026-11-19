#!/usr/bin/env bash
# Measures the context size of fixed navigation procedures with and without
# manazashi. See bench/context/README.md.
set -euo pipefail

usage() {
  echo "usage: $0 [--steps] --repo REPO --db DB --mzci MZCI TASKFILE" >&2
  exit 2
}

show_steps=0
REPO= DB= MZCI= taskfile=
while [ $# -gt 0 ]; do
  case "$1" in
    --steps) show_steps=1; shift ;;
    --repo) REPO="$2"; shift 2 ;;
    --db) DB="$2"; shift 2 ;;
    --mzci) MZCI="$2"; shift 2 ;;
    -*) usage ;;
    *) taskfile="$1"; shift ;;
  esac
done
[ -n "$REPO" ] && [ -n "$DB" ] && [ -n "$MZCI" ] && [ -n "$taskfile" ] || usage

# Steps run inside $REPO, so resolve caller-relative paths first.
abspath() {
  case "$1" in
    /*) echo "$1" ;;
    */*) echo "$PWD/$1" ;;
    *) if [ -e "$1" ]; then echo "$PWD/$1"; else echo "$1"; fi ;;
  esac
}
REPO="$(abspath "$REPO")"
DB="$(abspath "$DB")"
MZCI="$(abspath "$MZCI")"

if ! "$MZCI" status --db "$DB" --format json >/dev/null 2>&1; then
  echo "error: index is not usable: $DB" >&2
  exit 1
fi

export LC_ALL=en_US.UTF-8
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# mz SUBCOMMAND ARGS... runs mzci against the benchmark DB with JSON output.
mz() {
  local sub="$1"
  shift
  "$MZCI" "$sub" --db "$DB" --format json "$@"
}

# read_file PATH emulates an agent Read: the first 2,000 lines, numbered.
read_file() {
  head -n 2000 "$1" | awk '{ printf "%6d\t%s\n", NR, $0 }'
}

step_rows=()
cond_out=
cond_steps=0 cond_chars=0 cond_bytes=0 cond_lines=0 cond_empty=0

# step LABEL CMD... runs CMD in the repository and measures its stdout.
step() {
  local label="$1"
  shift
  local out="$work/step"
  (cd "$REPO" && "$@") </dev/null >"$out" 2>/dev/null || true
  local chars bytes lines
  chars=$(wc -m <"$out" | tr -d ' ')
  bytes=$(wc -c <"$out" | tr -d ' ')
  lines=$(wc -l <"$out" | tr -d ' ')
  cat "$out" >>"$cond_out"
  printf '%s\n' "$*" >>"$cond_out.cmd"
  cond_steps=$((cond_steps + 1))
  [ "$chars" -gt 0 ] || cond_empty=1
  cond_chars=$((cond_chars + chars))
  cond_bytes=$((cond_bytes + bytes))
  cond_lines=$((cond_lines + lines))
  step_rows+=("| $task | $cond | $label | $chars | $bytes | $lines |")
}

# shellcheck source=/dev/null
. "$taskfile"

failed=0
echo "| Task | Condition | Steps | Chars | Bytes | Lines | Chars ratio | Target |"
echo "| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |"
for task in "${TASKS[@]}"; do
  target_var="${task}_target"
  target="${!target_var}"
  with_chars=0
  for cond in with without; do
    cond_out="$work/$cond"
    : >"$cond_out"
    : >"$cond_out.cmd"
    cond_steps=0 cond_chars=0 cond_bytes=0 cond_lines=0 cond_empty=0
    "${task}_${cond}"
    # The command line is checked but not measured: an agent reading a file
    # knows its path from the call, not from the Read output. A step with no
    # output is a failed lookup even when its command names the target.
    if [ "$cond_empty" -eq 0 ] && grep -qF -- "$target" "$cond_out" "$cond_out.cmd"; then
      check=OK
    else
      check=NG
      failed=1
    fi
    ratio=
    if [ "$cond" = with ]; then
      with_chars=$cond_chars
    elif [ "$cond_chars" -gt 0 ]; then
      ratio=$(awk -v a="$with_chars" -v b="$cond_chars" 'BEGIN { printf "%.3f", a / b }')
    fi
    echo "| $task | $cond | $cond_steps | $cond_chars | $cond_bytes | $cond_lines | $ratio | $check |"
  done
done

if [ "$show_steps" -eq 1 ]; then
  echo
  echo "| Task | Condition | Step | Chars | Bytes | Lines |"
  echo "| --- | --- | --- | ---: | ---: | ---: |"
  printf '%s\n' "${step_rows[@]}"
fi

exit "$failed"
