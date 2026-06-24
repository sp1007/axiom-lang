#!/usr/bin/env bash
# BUG#26 end-to-end verification + self-host SHA.
# 1) stage1(fixed) -> stage2 ; 2) stage2 compiles t_strlen (must be ax_len + "5");
# 3) stage2 -> stage3 ; 4) SHA-256(stage2)==SHA-256(stage3).
set -u
cd /d/projects/compiler/Axiom
OBJDUMP=C:/msys64/ucrt64/bin/objdump.exe
echo "=== [1] stage1 -> stage2  $(date +%H:%M:%S) ==="
./bin/axc_stage1.exe build bootstrap/stage1/tmp_concatenated_air.ax -o bin/axc_stage2.exe -self-link -O1 > compiler_stage2_v6.log 2>&1
echo "stage2 build exit=$? $(date +%H:%M:%S) size=$(stat -c %s bin/axc_stage2.exe 2>/dev/null)"
[ -f bin/axc_stage2.exe ] || { echo "STAGE2 BUILD FAILED"; exit 1; }

echo "=== [2] stage2 compiles t_strlen (BUG#26 check) ==="
./bin/axc_stage2.exe build bin/t_strlen.ax -o bin/t_strlen_s2.exe -self-link -O1 > /dev/null 2>&1
cp axiom_temp.obj bin/obj_strlen_s2_fixed.obj
echo "ax_len reloc count in t_strlen obj: $("$OBJDUMP" -dr bin/obj_strlen_s2_fixed.obj 2>/dev/null | grep -c ax_len)"
echo "t_strlen run output (expect 5): $(./bin/t_strlen_s2.exe; echo " exit=$?")"

echo "=== [3] stage2 -> stage3  $(date +%H:%M:%S) ==="
./bin/axc_stage2.exe build bootstrap/stage1/tmp_concatenated_air.ax -o bin/axc_stage3.exe -self-link -O1 > compiler_stage3_v6.log 2>&1
echo "stage3 build exit=$? $(date +%H:%M:%S) size=$(stat -c %s bin/axc_stage3.exe 2>/dev/null)"

echo "=== [4] SHA-256 stage2 vs stage3 ==="
s2=$(sha256sum bin/axc_stage2.exe 2>/dev/null | cut -d' ' -f1)
s3=$(sha256sum bin/axc_stage3.exe 2>/dev/null | cut -d' ' -f1)
echo "stage2: $s2"
echo "stage3: $s3"
if [ -n "$s2" ] && [ "$s2" = "$s3" ]; then echo "SELF-HOST OK: SHA MATCH"; else echo "SHA MISMATCH (or stage3 missing)"; fi
echo "=== DONE $(date +%H:%M:%S) ==="
