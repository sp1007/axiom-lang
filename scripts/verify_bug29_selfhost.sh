#!/usr/bin/env bash
# BUG#29 end-to-end self-host verify: stage1(+fix) -> stage2 -> stage3 -> stage4,
# check fixpoint SHA(stage2)==SHA(stage3) and SHA(stage3)==SHA(stage4).
set -u
cd /d/projects/compiler/Axiom
S=bootstrap/stage1/tmp_concatenated_air.ax

echo "=== [1] stage1 -> stage2  $(date +%H:%M:%S) ==="
./bin/axc_stage1.exe build $S -o bin/axc_stage2.exe -self-link -O1 > compiler_stage2_v7.log 2>&1
echo "stage2 build exit=$? $(date +%H:%M:%S) size=$(stat -c %s bin/axc_stage2.exe 2>/dev/null)"
[ -f bin/axc_stage2.exe ] || { echo "STAGE2 BUILD FAILED"; exit 1; }

echo "=== [2] stage2 runtime checks ==="
./bin/axc_stage2.exe build bin/t_param5.ax -o bin/t_param5_s2c.exe -self-link -O1 > /dev/null 2>&1
echo "t_param5 (expect A38): $(./bin/t_param5_s2c.exe 2>&1); exit=$?"
./bin/axc_stage2.exe build bin/t_strip.ax -o bin/t_strip_s2c.exe -self-link -O1 > /dev/null 2>&1
echo "t_strip  (expect a.b len exit print): $(./bin/t_strip_s2c.exe 2>&1); exit=$?"

echo "=== [3] stage2 -> stage3  $(date +%H:%M:%S) ==="
./bin/axc_stage2.exe build $S -o bin/axc_stage3.exe -self-link -O1 > compiler_stage3_v7.log 2>&1
echo "stage3 build exit=$? $(date +%H:%M:%S) size=$(stat -c %s bin/axc_stage3.exe 2>/dev/null)"

echo "=== [4] stage3 -> stage4  $(date +%H:%M:%S) ==="
./bin/axc_stage3.exe build $S -o bin/axc_stage4.exe -self-link -O1 > compiler_stage4_v7.log 2>&1
echo "stage4 build exit=$? $(date +%H:%M:%S) size=$(stat -c %s bin/axc_stage4.exe 2>/dev/null)"

echo "=== [5] SHA-256 ==="
s2=$(sha256sum bin/axc_stage2.exe 2>/dev/null | cut -d' ' -f1)
s3=$(sha256sum bin/axc_stage3.exe 2>/dev/null | cut -d' ' -f1)
s4=$(sha256sum bin/axc_stage4.exe 2>/dev/null | cut -d' ' -f1)
echo "stage2: $s2"
echo "stage3: $s3"
echo "stage4: $s4"
if [ -n "$s2" ] && [ "$s2" = "$s3" ]; then echo "SELF-HOST OK: stage2 == stage3 (true fixpoint)"; \
elif [ -n "$s3" ] && [ "$s3" = "$s4" ]; then echo "SELF-HOST OK: stage3 == stage4 (converged fixpoint)"; \
else echo "NO FIXPOINT (stage2!=stage3 and stage3!=stage4)"; fi
echo "=== DONE $(date +%H:%M:%S) ==="
