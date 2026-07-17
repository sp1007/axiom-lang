---
name: project-next-step-13
description: Trạng thái next-step-13 (done) và next-step-14 in-progress — regalloc root cause fix
metadata: 
  node_type: memory
  type: project
  originSessionId: 9720703b-c141-45a7-9477-748fe53e183d
---

## next-step-13: DONE (2026-06-06)
- axc_stage2_native.exe (1,606,144 bytes, PE32+) built by patched axc_stage1.exe
- SHA-256 determinism confirmed: Build1==Build2 = `9b9babb6f78d140d974e2e6afb45169e9792aba17f71bbb8eb68403c9c626451`
- axc_stage1.exe binary patch: NOP at file_off=0x67066 (skip duplicate compile_native_asm)

## next-step-14: IN PROGRESS (2026-06-06)

### Root cause found and fixed (commit 937c375)
**Bug**: x86_regalloc.ax line 774: when spilling a 16-byte str vreg, `MACH_LOAD(R10, [rbp-X])` was used for ALL ops. This loaded str.ptr (8 bytes from spill slot = first field), not the struct ADDRESS. Any subsequent `[R10+8]` then double-dereferenced through str.ptr into string content.

**Fix**: Added `elif is_16: MACH_LEA(R10, [rbp-X])` before `else: MACH_LOAD`. For non-MACH_LEA ops with 16-byte src1, use LEA to give struct address so `[R10+8]` correctly reads str.len.

**Also fixed** (commit b68ed14): OP_CAST str→ptr in x86_selector.ax line 669.

### Rebuild status
- axc_stage2_native_fixed2.exe building with corrected regalloc (PID 4089, started 15:45 Jun 6)
- ETA: ~18:00 Jun 6

### Next steps after rebuild
1. Test: `axc_stage2_native_fixed2.exe` no-args → should print Usage
2. Compile test_mut_str.ax → success  
3. Stage2→Stage3 SHA-256 test (~4h total)

## Why (root cause context)
axc_stage1.exe was compiled from old AXIOM source WITH the regalloc bug. So binaries it produces all have wrong 16-byte str access. Fix is in source → rebuild → fixed binary.
