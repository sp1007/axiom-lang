#!/usr/bin/env bash
# RFC 0032 P1 gate: build an AXIOM `--shared --target linux` library and verify it is a real,
# dlopen-able ELF shared object. Runs the .so under WSL via python3 ctypes (no gcc needed).
# Requires: WSL + python3. Run from repo root.
#
# Checks, in order:
#   1. builds without error
#   2. readelf -h => Type: DYN (ET_DYN, value 3) -- NOT EXEC
#   3. readelf --dyn-syms lists the exported ax_add / ax_mul
#   4. python3 ctypes: dlopen the .so, call ax_add(2,3)==5 and ax_mul(4,5)==20
#
# Until PIC codegen (P2/P3) lands, only leaf functions with no global/func-addr refs are
# guaranteed correct after relocation; axmath's two functions are pure arithmetic leaves.
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_native.exe}"
export MSYS2_ARG_CONV_EXCL="*"
ROOT_WSL="/mnt/d/projects/compiler/Axiom"
SO="bin/axmath.so"
SO_WSL="$ROOT_WSL/bin/axmath.so"
fail=0

echo "=== build ==="
rm -f "$SO"
"$AXC" build tests/ffi/axmath.ax --shared -o "$SO" --target linux -self-link -O1 > /tmp/so_export.log 2>&1
if [ ! -f "$SO" ]; then echo "FAIL build"; cat /tmp/so_export.log; exit 1; fi
echo "PASS build ($(stat -c %s "$SO" 2>/dev/null) bytes)"

echo "=== e_type == DYN ==="
etype=$(wsl readelf -h "$SO_WSL" 2>&1 | grep -iE "^\s*Type:" | head -1)
echo "$etype"
if echo "$etype" | grep -qiE "DYN"; then echo "PASS ET_DYN"; else echo "FAIL want ET_DYN"; fail=1; fi

echo "=== .dynsym exports (nm -D reads the dynamic segment; we emit no section headers, so"
echo "    readelf --dyn-syms shows nothing here — that is a display quirk, not a defect) ==="
dsyms=$(wsl nm -D "$SO_WSL" 2>&1)
for s in ax_add ax_mul; do
  # `T <name>` = defined (text) global export
  if echo "$dsyms" | grep -qE "\bT ${s}\b"; then echo "PASS export $s (defined)"; else echo "FAIL missing export $s"; fail=1; fi
done

echo "=== dlopen + call (python3 ctypes) ==="
wsl python3 - "$SO_WSL" <<'PY'
import sys, ctypes
lib = ctypes.CDLL(sys.argv[1])
lib.ax_add.restype = ctypes.c_int32
lib.ax_add.argtypes = [ctypes.c_int32, ctypes.c_int32]
lib.ax_mul.restype = ctypes.c_int32
lib.ax_mul.argtypes = [ctypes.c_int32, ctypes.c_int32]
a = lib.ax_add(2, 3)
m = lib.ax_mul(4, 5)
ok = (a == 5 and m == 20)
print(f"ax_add(2,3)={a} ax_mul(4,5)={m} -> {'PASS' if ok else 'FAIL'}")
sys.exit(0 if ok else 1)
PY
if [ $? -ne 0 ]; then echo "FAIL dlopen/call"; fail=1; fi

echo "======================================"
if [ "$fail" -ne 0 ]; then echo "SO_EXPORT_FAILED"; exit 1; fi
echo "SO_EXPORT_OK"
