#!/usr/bin/env bash
# Regression gate for the known-bug repros (chain #24..#31 + Family C / RFC 0001).
# Run after ANY codegen/optimizer/ABI change before trusting the next step.
# Each case: build with the given compiler (default bin/axc_stage1.exe) at -O1,
# run, and compare stdout (CMP=out) or exit code (CMP=exit) to the expected value.
#
# NOTE: on this Windows box, Windows Defender realtime adds ~50s per self-link
# build (CPU ~0, pure scan). That is NOT a hang — keep the per-build timeout >=120s.
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_stage1.exe}"
TIMEOUT="${TIMEOUT:-150}"
pass=0; fail=0; failed=""

# rows: name | cmp(out|exit) | expected
rows=(
  "tsp|exit|7"
  "tsp2|exit|9"
  "tsp3|exit|12"
  "tstruct_abi|out|A=7 B=12 C=15 D=6 E=99 "
  "t_movrr|out|072 137 229 "
  "t_modrm|out|229"
  "t_cp2|out|7"
  "t_cpaddr|exit|7"
  "t_cse|exit|98"
  "t_param5|out|A38"
  "t_strip|out|a.b len exit print"
  "t_alias|out|A=1 B=0 "
  "t_mutstr|exit|9"
  "t_method|exit|24"
  "t_enum_np|exit|6"
  "t_enum|exit|42"
  "t_adt2|exit|104"
  "t_adt3|exit|19"
  "t_builtin_opt|exit|15"
  "t_litinfer|exit|111"
)

for row in "${rows[@]}"; do
  IFS='|' read -r name cmp want <<< "$row"
  src="bin/${name}.ax"
  [ -f "$src" ] || { echo "SKIP $name (no $src)"; continue; }
  out_exe="/tmp/reg_${name}.exe"
  rm -f "$out_exe"
  timeout "$TIMEOUT" "$AXC" build "$src" -o "$out_exe" -O1 >/dev/null 2>&1
  if [ ! -f "$out_exe" ]; then echo "FAIL $name (build produced no exe)"; fail=$((fail+1)); failed="$failed $name"; continue; fi
  got_out=$("$out_exe" 2>/dev/null); got_exit=$?
  if [ "$cmp" = "exit" ]; then got="$got_exit"; else got="$got_out"; fi
  if [ "$got" = "$want" ]; then
    echo "PASS $name ($cmp=$got)"; pass=$((pass+1))
  else
    echo "FAIL $name ($cmp: got '$got' want '$want')"; fail=$((fail+1)); failed="$failed $name"
  fi
done

echo "=== regression: $pass passed, $fail failed ==="
[ -n "$failed" ] && echo "FAILED:$failed"
[ "$fail" -eq 0 ] && echo "REGRESSION_OK" || echo "REGRESSION_FAIL"
