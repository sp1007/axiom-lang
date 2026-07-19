#!/usr/bin/env bash
# RFC 0014/0015 `-ctgc-free` regression (the normal suite only exercises the flag OFF).
# With -ctgc-free, CTGC calls `drop(self)` then frees the block for every non-escaping,
# owned local. As of RFC 0015 P3 activation (18db268) this ALSO frees non-drop aggregates
# (a plain OP_DESTROY, no drop call) -- so the flag must still be behaviour-preserving for
# correct programs. So:
#   * a program with NO drop type must be UNCHANGED by the flag (on == off), and
#   * a drop program must run its drop(s) exactly when the flag is on.
# A crash or wrong result under -ctgc-free means the escape analysis freed/dropped
# something still live (UAF / early-drop). Row: name | flag-off exit | flag-on exit.
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_native.exe}"
TIMEOUT="${TIMEOUT:-150}"
pass=0; fail=0; failed=""

# name | off-exit | on-exit
rows=(
  # drop-typed: drop fires only with the flag
  "t_drop|0|42"
  # non-drop programs: -ctgc-free must not change behavior (on == off)
  "t_ctgcfree|42|42"
  "t_escape|41|41"
  "t_escapeloop|42|42"
  "t_optmethod|42|42"
  "t_forvec|60|60"
  "t_gentree|15|15"
  "t_structoptfield|42|42"
  "t_vecstructopt|42|42"
  # NEGATIVE escape oracle: a ctor local pushed into a Vec must NOT be freed under
  # -ctgc-free (container-store escape f873948 + reassign-to-borrow 68d2c78). If either
  # escape fix regresses, `it` becomes wrongly freeable -> freed -> the Vec dangles ->
  # on != off / crash. on==off==33 proves the escaped local is retained.
  "t_ctgcescape|33|33"
  # general-free (activated) escape oracle: a returned non-drop aggregate must NOT be freed
  "t_ctgcfreeesc|16|16"
  # RFC 0027 container free-glue: scratch container buffers freed, escaping/aliased survive
  "t_ctgccont|42|42"
)

for row in "${rows[@]}"; do
  IFS='|' read -r name woff won <<< "$row"
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
  if [ "$o" = "$woff" ] && [ "$n" = "$won" ]; then
    echo "PASS $name (off=$o on=$n)"; pass=$((pass+1))
  else
    echo "FAIL $name (off=$o want=$woff | on=$n want=$won)"; fail=$((fail+1)); failed="$failed $name"
  fi
done

echo "=== ctgc-free check: $pass passed, $fail failed ==="
[ -n "$failed" ] && echo "FAILED:$failed"
[ "$fail" -eq 0 ] && echo "CTGC_FREE_OK" || echo "CTGC_FREE_FAIL"
