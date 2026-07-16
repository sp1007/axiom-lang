#!/usr/bin/env bash
# RFC 0015 P3 flag-on regression: for each program, building with -ctgc-free
# (compile-time free active) must produce the SAME result as the default build
# and must match the oracle. A mismatch or crash under -ctgc-free means the escape
# analysis is unsound (freed something still live -> UAF) — this is the gate that
# guards the opt-in free path (the normal regression only exercises the flag OFF).
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_native.exe}"
TIMEOUT="${TIMEOUT:-150}"
pass=0; fail=0; failed=""

# name | expected-exit
rows=(
  "t_ctgcfree|42"
  "t_escape|41"
  "t_escapeloop|42"
  "t_optmethod|42"
  "t_forvec|60"
  "t_gentree|15"
  "t_structoptfield|42"
  "t_vecstructopt|42"
)

for row in "${rows[@]}"; do
  IFS='|' read -r name want <<< "$row"
  src="bin/${name}.ax"
  [ -f "$src" ] || { echo "SKIP $name (no $src)"; continue; }
  off="/tmp/ctgc_off_${name}.exe"; on="/tmp/ctgc_on_${name}.exe"
  rm -f "$off" "$on"
  timeout "$TIMEOUT" "$AXC" build "$src" -o "$off" -O1 >/dev/null 2>&1
  timeout "$TIMEOUT" "$AXC" build "$src" -o "$on" -ctgc-free -O1 >/dev/null 2>&1
  if [ ! -f "$off" ] || [ ! -f "$on" ]; then
    echo "FAIL $name (build failed: off=$([ -f "$off" ]&&echo y||echo n) on=$([ -f "$on" ]&&echo y||echo n))"
    fail=$((fail+1)); failed="$failed $name"; continue
  fi
  "$off" >/dev/null 2>&1; o=$?
  "$on"  >/dev/null 2>&1; n=$?
  if [ "$o" = "$n" ] && [ "$n" = "$want" ]; then
    echo "PASS $name (off=$o on=$n)"; pass=$((pass+1))
  else
    echo "FAIL $name (off=$o on=$n want=$want)"; fail=$((fail+1)); failed="$failed $name"
  fi
done

echo "=== ctgc-free check: $pass passed, $fail failed ==="
[ -n "$failed" ] && echo "FAILED:$failed"
[ "$fail" -eq 0 ] && echo "CTGC_FREE_OK" || echo "CTGC_FREE_FAIL"
