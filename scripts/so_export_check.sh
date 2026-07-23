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

echo ""
echo "=== RFC 0032 P2: shared object with a MODULE GLOBAL ==="
SOG="bin/soglobal.so"
SOG_WSL="$ROOT_WSL/bin/soglobal.so"
rm -f "$SOG"
"$AXC" build tests/ffi/soglobal.ax --shared -o "$SOG" --target linux -self-link -O1 > /tmp/so_global.log 2>&1
if [ ! -f "$SOG" ]; then echo "FAIL soglobal build"; cat /tmp/so_global.log; fail=1; else
  echo "PASS soglobal build ($(stat -c %s "$SOG" 2>/dev/null) bytes)"
  # Global addressing must survive relocation: bump 3x -> 1,2,3; then addg(10) -> 13.
  # Proves OP_GLOBAL_ADDR is RIP-relative (a link-time absolute would fault or read the
  # wrong address once the .so is mapped at a random base under ASLR).
  wsl python3 - "$SOG_WSL" <<'PY'
import sys, ctypes
lib = ctypes.CDLL(sys.argv[1])
lib.ax_bump.restype = ctypes.c_int64
lib.ax_addg.restype = ctypes.c_int64
lib.ax_addg.argtypes = [ctypes.c_int64]
seq = [lib.ax_bump(), lib.ax_bump(), lib.ax_bump()]
g = lib.ax_addg(10)
ok = (seq == [1, 2, 3] and g == 13)
print(f"bump->{seq} addg(10)->{g} -> {'PASS' if ok else 'FAIL'}")
sys.exit(0 if ok else 1)
PY
  if [ $? -ne 0 ]; then echo "FAIL soglobal dlopen/global"; fail=1; fi
fi

echo ""
echo "=== RFC 0032 P2: intra-module call + string-returning export ==="
SOP="bin/soprobe.so"
SOP_WSL="$ROOT_WSL/bin/soprobe.so"
rm -f "$SOP"
"$AXC" build tests/ffi/soprobe.ax --shared -o "$SOP" --target linux -self-link -O1 > /tmp/so_probe.log 2>&1
if [ ! -f "$SOP" ]; then echo "FAIL soprobe build"; cat /tmp/so_probe.log; fail=1; else
  echo "PASS soprobe build ($(stat -c %s "$SOP" 2>/dev/null) bytes)"
  wsl python3 - "$SOP_WSL" <<'PY'
import sys, ctypes
lib = ctypes.CDLL(sys.argv[1])
lib.ax_calc.restype = ctypes.c_int64
lib.ax_calc.argtypes = [ctypes.c_int64]
lib.ax_greet.restype = ctypes.c_char_p
c = lib.ax_calc(4)   # sum_{i=1..4}(i*i+1) = 2+5+10+17 = 34
g = lib.ax_greet()
ok = (c == 34 and g == b'hi-from-so')
print(f"ax_calc(4)={c} (want 34)  ax_greet()={g!r} -> {'PASS' if ok else 'FAIL'}")
sys.exit(0 if ok else 1)
PY
  if [ $? -ne 0 ]; then echo "FAIL soprobe dlopen/call"; fail=1; fi
fi

echo "======================================"
if [ "$fail" -ne 0 ]; then echo "SO_EXPORT_FAILED"; exit 1; fi
echo "SO_EXPORT_OK"
