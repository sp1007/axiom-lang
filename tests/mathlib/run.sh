#!/usr/bin/env bash
# Careful accuracy verification for std.math: bundle std/math.ax with the curated
# accuracy probe (oracle-computed reference bands), build with the real compiler
# (bin/axc_stage1.exe), and assert exit 127 (all functions accurate to ~1e-5).
#
# Run from repo root: bash tests/mathlib/run.sh
set -u
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
AXC="$ROOT/bin/axc_stage1.exe"
TMP="$(mktemp -d)"
bundle="$TMP/probe.ax"; exe="$TMP/probe.exe"
cat "$ROOT/std/math.ax" "$ROOT/tests/mathlib/accuracy_probe.ax" > "$bundle"
"$AXC" build "$bundle" -O1 -o "$exe" >/dev/null 2>&1 || { echo "FAIL: build error"; exit 1; }
"$exe"; code=$?
if [ "$code" -eq 127 ]; then
  echo "PASS: std.math accuracy verified (exit 127, all functions within oracle band)"
  exit 0
fi
echo "FAIL: accuracy probe exit=$code (expected 127; a bit-N=0 means that function drifted)"
echo "  bits: 1=sqrt 2=sin 4=exp 8=erf 16=asin 32=cbrt 64=tgamma"
exit 1
