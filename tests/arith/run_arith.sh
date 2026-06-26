#!/usr/bin/env bash
# Differential arithmetic test runner (RFC 0006).
# Generates N expressions with a trusted fixed-width oracle, builds the
# self-checking AXIOM program with the native compiler, runs it, and reports
# PASS/FAIL. Each "F <id> <got> <exp>" line is a mismatch = an arithmetic bug.
#
# Usage: AXC=bin/axc_stage1.exe N=100000 SEED=1 bash tests/arith/run_arith.sh
# NOTE: builds via self-link (writes axiom_temp.obj) — do NOT run while a
# self-host verify build is in progress.
set -u
cd "$(dirname "$0")/../.."
AXC="${AXC:-bin/axc_stage1.exe}"
N="${N:-100000}"
SEED="${SEED:-1}"
MODE="${MODE:-int}"   # int | float | mixed
TIMEOUT="${TIMEOUT:-1200}"
PY=python; command -v python >/dev/null 2>&1 || PY=python3

src="tests/arith/arith_test.ax"
exe="/tmp/arith_test.exe"
echo "[gen] N=$N seed=$SEED mode=$MODE -> $src"
"$PY" tests/arith/arith_gen.py "$N" "$SEED" "$MODE" > "$src" || { echo "GEN FAIL"; exit 2; }
echo "[gen] $(wc -l < "$src") lines"

rm -f "$exe"
echo "[build] $AXC build $src -O1 (timeout ${TIMEOUT}s)"
t0=$(date +%s)
timeout "$TIMEOUT" "$AXC" build "$src" -o "$exe" -O1 >/tmp/arith_build.log 2>&1
t1=$(date +%s)
if [ ! -f "$exe" ]; then
  echo "BUILD FAILED (>${TIMEOUT}s or error) — tail log:"; tail -5 /tmp/arith_build.log; exit 3
fi
echo "[build] ok in $((t1-t0))s"

out=$("$exe" 2>/dev/null); code=$?
echo "$out" | grep -E '^F ' | head -40
summary=$(echo "$out" | grep -E '^P=')
echo "[result] $summary  (exit=$code)"
fails=$(echo "$summary" | sed -n 's/.*F=\([0-9]*\).*/\1/p')
if [ "${fails:-1}" = "0" ]; then echo "ARITH_OK"; else echo "ARITH_FAIL ($fails mismatches)"; exit 1; fi
