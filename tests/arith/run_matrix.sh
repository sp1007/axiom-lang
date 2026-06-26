#!/usr/bin/env bash
# Structured numeric matrix test runner (RFC 0006). Self-checking; F lines =
# mismatches with their (type/op/cast) label in the generated source.
# NOTE: self-link build — do NOT run during a self-host verify.
set -u
cd "$(dirname "$0")/../.."
AXC="${AXC:-bin/axc_stage1.exe}"
TIMEOUT="${TIMEOUT:-600}"
PY=python; command -v python >/dev/null 2>&1 || PY=python3
src="tests/arith/matrix_test.ax"; exe="/tmp/matrix_test.exe"
"$PY" tests/arith/matrix_gen.py > "$src" || { echo "GEN FAIL"; exit 2; }
echo "[gen] $(grep -c '// ' "$src") lines, $(sed -n '2p' "$src")"
rm -f "$exe"
timeout "$TIMEOUT" "$AXC" build "$src" -o "$exe" -O1 >/tmp/matrix_build.log 2>&1
[ -f "$exe" ] || { echo "BUILD FAILED"; tail -5 /tmp/matrix_build.log; exit 3; }
out=$("$exe" 2>/dev/null); code=$?
# map failing ids back to labels
echo "$out" | grep -E '^F ' | while read -r _ id got exp; do
  label=$(grep -oE "// [a-z0-9_]+\$" "$src" | sed -n "$((id+1))p")
  echo "FAIL id=$id got=$got exp=$exp $label"
done | head -60
echo "[result] $(echo "$out" | grep -E '^P=')  (exit=$code)"
echo "$out" | grep -qE '^P=.* F=0$' && echo "MATRIX_OK" || { echo "MATRIX_FAIL"; exit 1; }
