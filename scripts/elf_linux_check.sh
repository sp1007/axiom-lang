#!/usr/bin/env bash
# RFC 0009 P3: Linux ELF target smoke test. Builds pure-compute + print AXIOM
# programs with `--target linux`, runs each under WSL, and checks exit code
# (and, for elfhello, exact stdout). Requires WSL + a Linux distro. Run from repo root.
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_native.exe}"
export MSYS2_ARG_CONV_EXCL="*"
ROOT_WSL="/mnt/d/projects/compiler/Axiom"
pass=0; fail=0; failed=""
# Extra flags for every build, e.g. AXEXTRA=-dfe. Unquoted so empty adds nothing.
# The ELF path shares compile_native_binary with COFF, so RFC 0031 pruning applies
# to it too -- and a pruned ELF is the one thing regression_repros.sh cannot check,
# since it only ever produces PE.
AXEXTRA="${AXEXTRA:-}"
build() { rm -f "bin/t_$1.elf"; "$AXC" build "bin/$1.ax" -o "bin/t_$1.elf" --target linux -self-link -O1 $AXEXTRA > "/tmp/elfchk_$1.log" 2>&1; }
# exit-code oracles: name|expected
# `t_bssglobal` pins RFC 0030 P4: a large zero-initialized global is reserved via SHT_NOBITS
# and p_memsz > p_filesz. It checks BEHAVIOUR under Linux (reads zero, both ends of the
# extent addressable) — the size win is meaningless if the kernel does not actually map the
# tail, and that can only be observed by running it. It deliberately reuses the SAME source
# the Windows regression gates (bin/t_bssglobal.ax) rather than duplicating the program, so
# the two formats can never drift apart; the `t_t_` output filename is the cosmetic cost.
for row in "elf42|42" "elfloop|7" "elfglob|55" "elfglobuninit|92" "elfvec|42" "elfmap|42" "elfsmap|42" "elfparse|40" "t_bssglobal|42"; do
  name="${row%%|*}"; exp="${row##*|}"; build "$name"
  if [ ! -f "bin/t_$name.elf" ]; then echo "FAIL $name (build)"; fail=$((fail+1)); failed="$failed $name"; continue; fi
  wsl "$ROOT_WSL/bin/t_$name.elf"; got=$?
  if [ "$got" = "$exp" ]; then echo "PASS $name (exit=$got)"; pass=$((pass+1)); else echo "FAIL $name (exit=$got want=$exp)"; fail=$((fail+1)); failed="$failed $name"; fi
done
# stdout oracle: elfhello
build elfhello
expected=$'Hello from AXIOM on Linux!\n2 + 3 = 5\nfib(10) = 55\ntrue'
if [ -f bin/t_elfhello.elf ]; then
  got=$(wsl "$ROOT_WSL/bin/t_elfhello.elf" 2>/dev/null)
  if [ "$got" = "$expected" ]; then echo "PASS elfhello (stdout)"; pass=$((pass+1)); else echo "FAIL elfhello (stdout)"; echo "--- got ---"; echo "$got"; fail=$((fail+1)); failed="$failed elfhello"; fi
else echo "FAIL elfhello (build)"; fail=$((fail+1)); failed="$failed elfhello"; fi
# stdout oracle: elffloat (f64 print)
build elffloat
expf=$'3.141590
0.500000
42.000000
0.285714'
if [ -f bin/t_elffloat.elf ]; then
  gotf=$(wsl "$ROOT_WSL/bin/t_elffloat.elf" 2>/dev/null)
  if [ "$gotf" = "$expf" ]; then echo "PASS elffloat (stdout)"; pass=$((pass+1)); else echo "FAIL elffloat (stdout)"; echo "--- got ---"; echo "$gotf"; fail=$((fail+1)); failed="$failed elffloat"; fi
else echo "FAIL elffloat (build)"; fail=$((fail+1)); failed="$failed elffloat"; fi
# stdout oracle: elfstr (String concat on the heap)
build elfstr
exps=$'Hello, Linux!'
if [ -f bin/t_elfstr.elf ]; then
  gots=$(wsl "$ROOT_WSL/bin/t_elfstr.elf" 2>/dev/null)
  if [ "$gots" = "$exps" ]; then echo "PASS elfstr (stdout)"; pass=$((pass+1)); else echo "FAIL elfstr (stdout): $gots"; fail=$((fail+1)); failed="$failed elfstr"; fi
else echo "FAIL elfstr (build)"; fail=$((fail+1)); failed="$failed elfstr"; fi
echo "=== ELF linux: $pass passed, $fail failed ==="
[ "$fail" -gt 0 ] && { echo "FAILED:$failed"; exit 1; }
echo "ELF_LINUX_OK"
