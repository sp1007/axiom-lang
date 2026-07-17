---
name: project-next-step-14
description: "Trạng thái next-step-14 — fix str codegen bug, rebuild axc_stage2, SHA-256 test"
metadata: 
  node_type: memory
  type: project
  originSessionId: 9720703b-c141-45a7-9477-748fe53e183d
---

## Trạng thái next-step-14 (cập nhật 2026-06-07)

**Mục tiêu**: axc_stage2_native compiles tmp_concatenated_air.ax → axc_stage3, SHA-256 reproducibility test.

### Đã hoàn thành
- axc_stage1_fixed2.exe: 2 code cave patches (cave1: MACH_LOAD→MACH_LEA src1 spill, cave2: regalloc_is_16byte OP_GET_FIELD)
- axc_stage2_native_fixed5.exe: compiled từ axc_stage1_fixed2 (1,598,976 bytes, hoàn thành 23:36)
- 518 MACH_LEA occurrences trong fixed5 vs 0 trong fixed3 (patch có tác dụng)

### Vấn đề hiện tại (BLOCKER)
1. **axc_stage2_native_fixed5.exe** crash ở startup (trước khi compile bất cứ thứ gì)
   - Crash ở get_freestanding_args: str.ptr = "bin/triv" bytes thay vì valid heap address
   - args array overlaps với cmdline buffer do populate_str_in_place viết sai ptr value
   - Root cause: cave1 patch (MACH_LEA) có thể đang convert 8-byte ptr[u8] params sang 16-byte hidden pointer convention
   
2. **axc_stage2_patched_v4.exe** (from fixed3): startup OK nhưng crash khi compile bất kỳ program nào có str operations
   - Crash pattern: `mov (%rcx), %rax` với rcx = huge value (arithmetic overflow)
   - Không thể compile test_mut_str.ax

### Phân tích kỹ thuật
- Crash cả hai: str.ptr = filename bytes thay vì valid heap address
- Pattern `populate_str_in_place` viết byte-by-byte từ `s_ptr as i64`
- Nếu `s_ptr` bị classify sai là 16-byte → hidden pointer convention → [rcx] = string bytes
- src2 và dst paths trong insert_spill_code chưa được patch

### Files đã modified
- bin/axc_stage1_fixed2.exe: cave1 + cave2 patches
- bin/axc_stage2_native_fixed5.exe: compiled bởi fixed2
- bin/axc_stage2_patched_v4/v5/v6.exe: binary patches

### Bước tiếp theo cần làm
1. Fix THÊM bugs trong axc_stage1_fixed2.exe:
   - src2 path tại 14003f741 (file 0x03ED41)
   - dst path tại 14003EED8
   - OR: Fix AXIOM source (air_builder.ax OP_GET_FIELD type_id) + rebuild

**Why**: Cần axc_stage2 không crash ở startup để bắt đầu compile tmp_concatenated_air.ax

[[feedback_compact]]
