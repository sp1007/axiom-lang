#!/usr/bin/env bash
# M6 perf harness — builds each benchmark with AXIOM (-O1 by default) and clang
# -O2, times best-of-N wall-clock, prints the AXIOM/clang ratio. The M6 gate is
# ratio <= 1.05 on fib(40). Deterministic, isolated, reversible (§10).
#
# Usage:  bash scripts/bench_perf.sh [name ...]      (default: fib collatz)
#   AXC=bin/axc_native.exe   AXOPT=-O1   RUNS=3   CLANG=/c/msys64/ucrt64/bin/clang.exe
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_native.exe}"
AXOPT="${AXOPT:--O1}"
CLANG="${CLANG:-/c/msys64/ucrt64/bin/clang.exe}"
RUNS="${RUNS:-3}"
OUT="${OUT:-bin/_regtmp}"
mkdir -p "$OUT" 2>/dev/null || true
names=("$@"); [ ${#names[@]} -eq 0 ] && names=(fib collatz)
PY="$(command -v python || command -v python3)"

# best-of-N run time (ms) of an executable, via python perf_counter subprocess.
# Native-Windows python needs a Windows path, so convert with cygpath.
time_exe() {
  local win; win="$(cygpath -w "$1" 2>/dev/null || echo "$1")"
  "$PY" - "$win" "$RUNS" <<'PYEOF'
import subprocess, sys, time
exe, runs = sys.argv[1], int(sys.argv[2])
best = None
for _ in range(runs):
    t0 = time.perf_counter()
    subprocess.run([exe], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    dt = (time.perf_counter() - t0) * 1000.0
    best = dt if best is None else min(best, dt)
print(f"{best:.1f}")
PYEOF
}

printf "%-12s %12s %12s %8s\n" "bench" "axiom(ms)" "clang(ms)" "ratio"
printf "%-12s %12s %12s %8s\n" "-----" "---------" "---------" "-----"
for n in "${names[@]}"; do
  axsrc="benchmarks/${n}.ax"; csrc="benchmarks/${n}.c"
  [ -f "$axsrc" ] || { echo "SKIP $n (no $axsrc)"; continue; }
  axexe="$OUT/bench_${n}_ax.exe"; cexe="$OUT/bench_${n}_c.exe"
  rm -f "$axexe" "$cexe"
  "$AXC" build "$axsrc" -o "$axexe" $AXOPT >/dev/null 2>&1
  [ -f "$axexe" ] || { echo "FAIL $n (axiom build)"; continue; }
  if [ -f "$csrc" ]; then "$CLANG" -O2 "$csrc" -o "$cexe" >/dev/null 2>&1; fi
  axms=$(time_exe "$axexe")
  if [ -f "$cexe" ]; then
    cms=$(time_exe "$cexe")
    ratio=$("$PY" -c "print(f'{$axms/$cms:.2f}x')" 2>/dev/null)
  else
    cms="n/a"; ratio="n/a"
  fi
  printf "%-12s %12s %12s %8s\n" "$n" "$axms" "$cms" "$ratio"
done
