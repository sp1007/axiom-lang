#!/usr/bin/env bash
# RFC 0030 / [[feedback-no-exe-bloat]] — executable SIZE guard.
#
# The regression suite compares exit codes, so it is blind to output growing. This suite
# asserts the opposite property: that emitted executables stay small. It exists because a
# one-line defect in x86_coff.ax (padding each global to a multiple of its own SIZE rather
# than its ALIGNMENT) silently doubled every executable carrying a large global — 1.6 MB of
# storage produced a 3.2 MB file — and every functional test still passed.
#
# Each row: name | source | max bytes over the no-global baseline.
# A row FAILS when the executable is larger than baseline + budget.
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_native.exe}"
# Deliberately NOT $REGTMP: this suite must not share a scratch directory with
# regression_repros.sh. When it did, running the two concurrently produced a spurious
# t_strsplit failure — the documented contention trap in [[infra-defender-build-throttle]],
# reintroduced by this script's own default. Keep it on its own directory.
TMP="${SIZETMP:-bin/_sizetmp}"
mkdir -p "$TMP"
pass=0; fail=0; failed=""

# Baseline: a program with no globals at all.
cat > "$TMP/sz_baseline.ax" <<'EOF'
fn main() -> i64:
    return 42 as i64
EOF
"$AXC" build "$TMP/sz_baseline.ax" -o "$TMP/sz_baseline.exe" -O1 >/dev/null 2>&1
if [ ! -f "$TMP/sz_baseline.exe" ]; then
  echo "FATAL: baseline build failed"; exit 1
fi
BASE=$(stat -c%s "$TMP/sz_baseline.exe")
echo "baseline (no globals): $BASE bytes"

# name | array elements | budget over baseline
# Budget today is "the declared storage plus slack" — the 1x regime after the alignment
# fix. RFC 0030 P3/P4 (.bss) should drop these budgets to near zero; TIGHTEN THEM THEN, so
# the oracle keeps proving the win instead of silently tolerating a regression back to 1x.
#
# CALIBRATION (do this for every row — a budget that cannot fail is not a test). Under the
# pre-fix `% gsize` padding both rows land at ~2x storage, because the bundled stdlib's own
# globals leave data_buf non-empty, which is what forces the pad up to a full multiple:
#   sz_small: 32,768 storage -> ~65,536 delta ; budget 40,000 => FAILS pre-fix, passes at 33,280
#   sz_big: 1,600,000 storage -> ~3,200,512 delta ; budget 1,700,000 => FAILS pre-fix
# An earlier draft used 65,536 for sz_small, which the pre-fix value would have exactly
# matched (the check is `>`), so that row would have passed the very defect it exists for.
# Budgets TIGHTENED for RFC 0030 P3: zero-initialized globals now occupy no file bytes at
# all, so the executable barely grows regardless of how much storage is declared. These
# budgets are deliberately near-zero -- if .bss ever stops engaging, the delta jumps back to
# the storage size (1x) or the padding size (2x) and both rows fail loudly.
rows=(
  "sz_small|4096|8192"
  "sz_big|200000|8192"
)

for row in "${rows[@]}"; do
  name="${row%%|*}"; rest="${row#*|}"
  n="${rest%%|*}"; budget="${rest#*|}"
  src="$TMP/$name.ax"
  printf 'mut big: [i64; %s]\nfn main() -> i64:\n    big[0 as i64] = 42 as i64\n    return big[0 as i64]\n' "$n" > "$src"
  "$AXC" build "$src" -o "$TMP/$name.exe" -O1 >/dev/null 2>&1
  if [ ! -f "$TMP/$name.exe" ]; then
    echo "FAIL $name (build failed)"; fail=$((fail+1)); failed="$failed $name"; continue
  fi
  # The program must still RUN correctly -- a size win that breaks the global is not a win.
  "./$TMP/$name.exe" >/dev/null 2>&1; rc=$?
  sz=$(stat -c%s "$TMP/$name.exe")
  delta=$((sz - BASE))
  if [ "$rc" != "42" ]; then
    echo "FAIL $name (ran with exit $rc, want 42)"; fail=$((fail+1)); failed="$failed $name"
  elif [ "$delta" -gt "$budget" ]; then
    echo "FAIL $name (delta=$delta > budget=$budget; storage=$((n*8)))"
    fail=$((fail+1)); failed="$failed $name"
  else
    echo "PASS $name (delta=$delta <= $budget, storage=$((n*8)), exit=$rc)"
    pass=$((pass+1))
  fi
done

# --- RFC 0031: pin the dead-function-elimination win --------------------------
#
# §5 requires a baseline row so the win cannot silently regress. Two properties,
# because either alone is a bad test:
#
#   SIZE  -- the pruned build must be far below the unpruned one. The cap is
#            absolute (30,000) and calibrated against the measured pair
#            77,824 unpruned / 22,016 pruned. If pruning stops engaging for any
#            reason the value jumps back to ~77,824 and the row fails loudly. A
#            relative budget would have quietly passed a no-op.
#
#   OUTPUT -- and the program must still PRINT THE RIGHT THING. Exit code alone
#            is far too weak here: the whole hazard of this optimization is a
#            call into an address that was never emitted, and the print / string
#            path is exactly where that first bit us (the ELF SIGSEGV, see
#            [[dfe-elf-runtime-is-in-program]]). A program that only returns an
#            integer never calls into that machinery, so it cannot detect the
#            failure this row exists to catch.
cat > "$TMP/sz_dfe.ax" <<'EOF'
fn main() -> i64:
    println("dfe alive")
    let s = std.string.concat("a", "b")
    println(s)
    return 42 as i64
EOF
"$AXC" build "$TMP/sz_dfe.ax" -o "$TMP/sz_dfe_plain.exe" -O1 >/dev/null 2>&1
"$AXC" build "$TMP/sz_dfe.ax" -o "$TMP/sz_dfe.exe" -O1 -dfe >/dev/null 2>&1
if [ ! -f "$TMP/sz_dfe.exe" ] || [ ! -f "$TMP/sz_dfe_plain.exe" ]; then
  echo "FAIL sz_dfe (build failed)"; fail=$((fail+1)); failed="$failed sz_dfe"
else
  plain=$(stat -c%s "$TMP/sz_dfe_plain.exe")
  pruned=$(stat -c%s "$TMP/sz_dfe.exe")
  got=$("./$TMP/sz_dfe.exe" 2>/dev/null); rc=$?
  want=$'dfe alive\nab'
  if [ "$got" != "$want" ]; then
    echo "FAIL sz_dfe (output '$got', want 'dfe alive/ab') -- a pruned root in the print path"
    fail=$((fail+1)); failed="$failed sz_dfe"
  elif [ "$rc" != "42" ]; then
    echo "FAIL sz_dfe (exit $rc, want 42)"; fail=$((fail+1)); failed="$failed sz_dfe"
  elif [ "$pruned" -gt 30000 ]; then
    echo "FAIL sz_dfe (pruned=$pruned > 30000; plain=$plain) -- pruning stopped engaging"
    fail=$((fail+1)); failed="$failed sz_dfe"
  else
    echo "PASS sz_dfe (pruned=$pruned vs plain=$plain, output ok, exit=$rc)"
    pass=$((pass+1))
  fi
fi

echo "=== exe-size check: $pass passed, $fail failed ==="
if [ "$fail" != "0" ]; then
  echo "FAILED:$failed"; exit 1
fi
echo "EXE_SIZE_OK"
