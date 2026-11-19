#!/usr/bin/env bash
# Verifies bench/context/run.sh against a small fixture repository.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$here/../.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

[ -x "$here/run.sh" ] || fail "bench/context/run.sh is missing"

repo="$tmp/repo"
mkdir -p "$repo/lib"
cat >"$repo/lib/a.rb" <<'EOF'
class Account
  # café
  def save
    true
  end
end
EOF
{
  echo "def first; end"
  for i in $(seq 2 2500); do echo "x$i = $i"; done
} >"$repo/lib/long.rb"
git -C "$repo" init -q
git -C "$repo" add lib
git -C "$repo" -c user.name=t -c user.email=t@example.com commit -qm fixture

(cd "$root" && go build -o "$tmp/mzci" ./cmd/mzci)
"$tmp/mzci" rebuild --db "$tmp/db.sqlite" "$repo" >/dev/null

status=0
"$here/run.sh" --steps --repo "$repo" --db "$tmp/db.sqlite" --mzci "$tmp/mzci" \
  "$here/testdata/fixture-tasks.sh" >"$tmp/out.md" 2>"$tmp/err" || status=$?

[ "$status" -eq 1 ] || fail "expected exit 1, got $status: $(cat "$tmp/err")"

# row TASK CONDITION prints the summary row's cells separated by "|".
row() {
  awk -F'|' -v t="$1" -v c="$2" '
    /^\| Task \| Condition \| Step / { steps = 1 }
    !steps && $2 ~ "^ " t " $" && $3 ~ "^ " c " $" { print; exit }
  ' "$tmp/out.md"
}
cell() { echo "$1" | awk -F'|' -v n="$2" '{ gsub(/ /, "", $n); print $n }'; }

for t in def_save long_read; do
  for c in with without; do
    r="$(row "$t" "$c")"
    [ -n "$r" ] || fail "missing row $t $c"
    [ "$(cell "$r" 9)" = OK ] || fail "$t $c target not OK: $r"
  done
done

for c in with without; do
  r="$(row broken "$c")"
  [ "$(cell "$r" 9)" = NG ] || fail "broken $c should be NG: $r"
  [ "$(cell "$r" 5)" = 0 ] || fail "broken $c should have 0 chars: $r"
done

r="$(row missing_read with)"
[ "$(cell "$r" 9)" = NG ] || fail "empty step output should be NG: $r"

r="$(row def_save without)"
[ "$(cell "$r" 5)" -lt "$(cell "$r" 6)" ] || fail "chars should be fewer than bytes: $r"

grep -q '^| long_read | without | read | [0-9]* | [0-9]* | 2000 |$' "$tmp/out.md" ||
  fail "long file read should be truncated to 2000 lines"

# Relative --db and --mzci must resolve against the caller's directory.
(cd "$tmp" && "$here/run.sh" --repo repo --db db.sqlite --mzci ./mzci \
  "$here/testdata/fixture-tasks.sh" >"$tmp/rel.md" 2>/dev/null) || true
grep -q '^| def_save | with | 2 | [1-9][0-9]* | .* | OK |$' "$tmp/rel.md" ||
  fail "relative paths should work: $(grep def_save "$tmp/rel.md")"

out="$("$here/run.sh" --repo "$repo" --db /nonexistent.sqlite --mzci "$tmp/mzci" \
  "$here/testdata/fixture-tasks.sh" 2>/dev/null)" && fail "missing DB should fail"
[ -z "$out" ] || fail "missing DB should print nothing on stdout"

echo PASS
