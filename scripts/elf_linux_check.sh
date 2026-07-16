#!/usr/bin/env bash
# RFC 0009 P3: Linux ELF target smoke test. Builds a set of pure-compute AXIOM
# programs with `--target linux`, runs each under WSL, and compares the exit code
# to an oracle. Requires WSL with a Linux distro. AXC defaults to the ELF-capable
# daily driver. Run from repo root.
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_native.exe}"
export MSYS2_ARG_CONV_EXCL="*"
pass=0; fail=0; failed=""
# name | expected-exit
rows=(
  "elf42|42"
  "elfloop|7"
  "elfglob|55"
)
for row in "${rows[@]}"; do
  name="${row%%|*}"; exp="${row##*|}"
  src="bin/${name}.ax"; out="bin/t_${name}.elf"
  rm -f "$out"
  "$AXC" build "$src" -o "$out" --target linux -self-link -O1 > /tmp/elfchk_$name.log 2>&1
  if [ ! -f "$out" ]; then echo "FAIL $name (build)"; fail=$((fail+1)); failed="$failed $name"; continue; fi
  winabs="/mnt/d/projects/compiler/Axiom/$out"
  wsl "$winabs"; got=$?
  if [ "$got" = "$exp" ]; then echo "PASS $name (exit=$got)"; pass=$((pass+1));
  else echo "FAIL $name (exit=$got want=$exp)"; fail=$((fail+1)); failed="$failed $name"; fi
done
echo "=== ELF linux: $pass passed, $fail failed ==="
if [ "$fail" -gt 0 ]; then echo "FAILED:$failed"; exit 1; fi
echo "ELF_LINUX_OK"
