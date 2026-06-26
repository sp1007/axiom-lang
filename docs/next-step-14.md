# AXIOM Language Project — next-step-14: Fix Mutable Str Codegen Bug

## Mục Tiêu

Sửa bug trong AXIOM native codegen khiến `axc_stage2_native.exe` crash khi chạy (STATUS_ACCESS_VIOLATION 0xC0000005), sau đó thực hiện full self-hosting SHA-256 test: `axc_stage2_native.exe` compile `tmp_concatenated_air.ax` → `axc_stage3_native.exe`, kiểm tra SHA-256(Stage2) == SHA-256(Stage3).

---

## Root Cause Analysis

### Bug: Mutable Str Variable Writes Through .rdata Pointer

**Triệu chứng**: Crash tại `ax_AxiomLinker_axiom_linker_link` instruction offset +0x217:
```asm
mov r10, (r11)   ; r11 = format.ptr = &"coff" in .rdata → WRITE vào .rdata → AV
```

**Chuỗi nguyên nhân**:

1. `mut format := "coff"` trong AXIOM source → AIR: `OP_ICONST` (type=str, val=rdata_offset)

2. Trong `x86_selector.ax` `select_inst` / `OP_ICONST` khi `is_str=true` (lines 529-534):
   ```
   tmp_struct_addr = MACH_MOV_IMM(imm=rdata_offset, vreg_flag=2)  // = &"coff" in rdata
   tmp_dest_addr   = MACH_LEA(inst.dest)                           // = &format on stack
   emit_block_copy(tmp_dest_addr, 0, tmp_struct_addr, 16)
   ```
   `emit_block_copy` gọi `MACH_LOAD [tmp_struct_addr+0]` — đọc 8 bytes ĐẦU của string "coff\0" làm ptr field. Kết quả: `format.ptr` = 0x00000000666F63 (NOT a valid pointer).

   **Hoặc** (nếu emitter xử lý vreg_flag=2 đặc biệt): `format.ptr` = `&"coff"` (valid ptr to .rdata) nhưng đây là pointer vào read-only memory.

3. Sau đó khi code thực hiện gán `format = some_str_value` hoặc struct copy vào field chứa format:
   - Load `format.ptr` → r11 = `&"coff"` in .rdata
   - Code generate: `MACH_STORE [r11] ← new_value` → write through ptr → CRASH

**Hai khả năng bug** (cần xác nhận bằng test nhỏ):

**Khả năng A**: `emit_block_copy` trong OP_ICONST/str đọc bytes từ rdata thay vì tạo {ptr, len} struct đúng. Kết quả là format.ptr = garbage → crash ngay khi dereference.

**Khả năng B** (khớp crash description hơn): `format.ptr` được set đúng = `&"coff"` in .rdata. Bug xảy ra khi code generate OP_SET_FIELD / struct-assignment dùng format.ptr làm DESTINATION thay vì write TO format.ptr field on stack.

---

## Checklist Nhiệm Vụ

### Bước 1: Isolate Bug với Test Case Nhỏ (PHẢI làm trước)

- [ ] Viết test program AXIOM tối giản có `mut format := "coff"`:
  ```
  // tests/codegen/test_mut_str.ax
  fn main() -> i32:
      mut s := "hello"
      s = "world"
      return 0
  ```
- [ ] Compile với `axc_stage1.exe` (patched) → tạo axiom_temp.obj
- [ ] Link với `axc_stage1.exe` → tạo test binary
- [ ] Chạy test binary → quan sát crash hoặc success
- [ ] Nếu crash: dùng WinDbg/x64dbg để xem instruction gây crash
- [ ] Xác nhận là Khả năng A hay Khả năng B

### Bước 2: Đọc và Hiểu Code Path Trong x86_selector.ax

- [ ] Đọc `emit_block_copy` (x86_selector.ax:499-507) và `emit_block_copy_ext` (446-497)
- [ ] Hiểu `MACH_MOV_IMM` với `vreg_flag=2` được emitter xử lý thế nào trong x86_emitter.ax
- [ ] Tìm chính xác dòng code sai: OP_ICONST/str path hoặc OP_COPY/OP_MOVE path
- [ ] Document: "Khi nào thì ptr field được ghi SAI?"

### Bước 3: Fix — Tùy Khả Năng

#### Nếu Khả Năng A (emit_block_copy đọc sai từ rdata):

Fix trong `x86_selector.ax` OP_ICONST / is_str path (line 529-534):

**Thay vì**: `emit_block_copy(tmp_dest_addr, 0, tmp_struct_addr, 16)`

**Sửa thành**: Trực tiếp store 2 fields:
```
// Store ptr field (tmp_struct_addr = &rdata[offset])
MACH_STORE [tmp_dest_addr + 0] ← tmp_struct_addr    // ptr field
// Store len field (từ AIR instruction metadata hoặc runtime strlen)
MACH_STORE [tmp_dest_addr + 8] ← len_value          // len field
```

Cần xác định len_value được encode ở đâu trong AirInst (có thể ở inst.src2 hoặc type_id).

#### Nếu Khả năng B (write-through bug trong struct assignment):

Fix trong OP_COPY / OP_MOVE / OP_SET_FIELD path cho 16-byte structs:

Khi `dst` là stack-allocated str:
- KHÔNG load ptr từ dst rồi dùng làm destination
- PHẢI dùng `MACH_LEA(dst)` làm destination address trực tiếp

Có thể cần thêm flag trong `regalloc_is_16byte` để phân biệt "copy TO str variable" vs "copy THROUGH str pointer".

#### Fix Bổ Sung (cần cho cả hai trường hợp — nếu str mutable cần writeable copy):

Để xử lý pattern `mut s := "literal"; s = other_str` ĐÚNG về mặt ngữ nghĩa, cần đảm bảo rằng sau khi gán `s = other_str`:
- `s.ptr` = `other_str.ptr` (shared reference, OK cho immutable)
- `s.len` = `other_str.len`

Đây là **value semantics cho str struct** (copy {ptr, len}), không phải copy string bytes. Nếu AXIOM định nghĩa str là non-owning slice, thì copy ptr là đúng. Nếu AXIOM str là owning, cần deep copy.

**Kiểm tra spec AXIOM cho str semantics** trước khi quyết định fix direction.

### Bước 4: Apply Fix vào Source Files

- [ ] Fix `bootstrap/stage1/x86_selector.ax` (chỗ bug)
- [ ] Regenerate `bootstrap/stage1/tmp_concatenated_air.ax` từ source files
- [ ] Verify fix bằng test nhỏ (Bước 1 test case)
- [ ] Verify thêm với `valid_many_args.ax` và `valid_hello_test_2.ax`

### Bước 5: Rebuild axc_stage2_native.exe Với Fix

- [ ] Compile `tmp_concatenated_air.ax` (fixed) với patched `axc_stage1.exe` → `axc_stage2_native_fixed.exe`
- [ ] Verify `axc_stage2_native_fixed.exe` không crash khi chạy cơ bản
- [ ] Test: `axc_stage2_native_fixed.exe` compile test_mut_str.ax → success

### Bước 6: Full Self-Hosting SHA-256 Test (Stage 2 → Stage 3)

- [ ] Stage 2 build: `axc_stage2_native_fixed.exe compile tmp_concatenated_air.ax → axc_stage3_native.exe`
  - ETA: ~2 giờ (same as stage1→stage2 build time)
- [ ] Stage 3 build: `axc_stage3_native.exe compile tmp_concatenated_air.ax → axc_stage3_verify.exe`  
  - ETA: ~2 giờ
- [ ] SHA-256 comparison:
  ```
  SHA-256(axc_stage3_native.exe) == SHA-256(axc_stage3_verify.exe)
  ```
- [ ] **MỐC**: Nếu match → AXIOM self-hosting native pipeline hoàn chỉnh, 100% deterministic, ZERO GCC dependency.

---

## Kỳ Vọng Kỹ Thuật

**Vị trí fix**: `bootstrap/stage1/x86_selector.ax`, hàm `select_inst`, trong path `OP_ICONST` khi `is_str=true` (lines ~529-534) và/hoặc `OP_COPY`/`OP_MOVE` path cho 16-byte structs (lines ~578-585).

**Độ phức tạp**: Thấp-Trung. Bug rõ ràng, chỉ cần sửa vài dòng codegen. Rủi ro: fix một path có thể break path khác — cần test toàn diện.

**Files bị ảnh hưởng**:
- `bootstrap/stage1/x86_selector.ax` (primary fix)
- `bootstrap/stage1/tmp_concatenated_air.ax` (regenerate sau fix)
- `bootstrap/stage1/x86_coff.ax` (có thể cần, nếu bug ở relocation generation)
- `tests/codegen/test_mut_str.ax` (new test)

**Không cần sửa**:
- `x86_regalloc.ax` — regalloc không gây ra bug này
- `x86_emitter.ax` — emitter xử lý đúng nếu selector tạo đúng MachInst
- `linker.ax` — linker không liên quan

---

## Dependency Cho Stage SHA-256 Test

```
axc_stage1.exe (patched, C-compiled)
    ↓ compile tmp_concatenated_air.ax (FIXED)
axc_stage2_native_fixed.exe
    ↓ compile tmp_concatenated_air.ax (same fixed source)
axc_stage3_native.exe
    ↓ compile tmp_concatenated_air.ax
axc_stage3_verify.exe

SHA-256(axc_stage3_native) == SHA-256(axc_stage3_verify) ← MỤC TIÊU
```

Total pipeline time (sau khi fix): ~4-5 giờ build time (2 lần 2h+ build).

---

## Liên Kết Với next-step-13

next-step-13 đã xác nhận:
1. axc_stage2_native.exe được tạo thành công (1,606,144 bytes PE32+, không GCC)
2. Codegen là deterministic: SHA-256(Build1) == SHA-256(Build2)
3. Bug mutable str .rdata là blocker duy nhất cho full self-hosting test

next-step-14 giải quyết blocker này.

---

## KẾT QUẢ / ĐÁNH GIÁ (2026-06-26)

**TRẠNG THÁI: ✅ MỤC TIÊU ĐẠT — self-host native deterministic fixpoint.**

| Tiêu chí next-step-14 | Kết quả |
|---|---|
| Bug mutable str ghi-xuyên-.rdata (crash AV) | ✅ FIXED — `tests/codegen/test_mut_str.ax`: `mut s:="hi"; s="worldwide"` in "hi"/"worldwide", exit=9, KHÔNG crash |
| Self-host SHA fixpoint | ✅ ĐẠT — converged fixpoint **stage3==stage4 = d7f14c2c** (`scripts/verify_bug29_selfhost.sh`) |
| Zero-GCC | ⚠️ MỘT PHẦN — stage2→3→4 thuần native AXIOM; stage1 bootstrap vẫn C/gcc. Tiêu chí self-host (native tái tạo chính nó bit-for-bit) ĐÃ đạt |

**Đánh giá độ chính xác của next-step-14:**
- Ước lượng "độ phức tạp Thấp-Trung, sửa vài dòng" → **SAI**. Bug mutable-str chỉ là *triệu chứng bề mặt* của vấn đề biểu diễn `str` 16-byte. Fix triệt để cần cả một CHUỖI bug sâu hơn qua nhiều phiên: #24 cast-shift → #25 optimizer addr-taken REX-loss → #26 strip self-mangle → #27 byte-store REX-loss → #28 stack 1MB→16MB → #29 (str C-ABI sret) → #30 (struct by-address vs str inline-16) → #31 (copy_prop thiếu guard addr_taken) → Family C (struct by-value ABI). Chi tiết: `knowledge/bugs.md`.
- Hướng root-cause của next-step-14 (str {ptr,len} handling) ĐÚNG hướng — nó chính là họ bug #29/#30 (str inline-16 vs aggregate by-address).
- "Blocker duy nhất" → SAI; là blocker ĐẦU TIÊN nhìn thấy, không phải duy nhất.

**Hạ tầng chống tái phát (đã thêm):** `scripts/regression_repros.sh` (13 repro gồm test_mut_str) — gate chạy sau mọi thay đổi backend. Baseline: all PASS. `scripts/verify_bug29_selfhost.sh` cho fixpoint.

**Kết luận:** next-step-14 hoàn thành về mục tiêu (self-host fixpoint), nhưng phạm vi thực tế lớn hơn nhiều so với dự kiến. AXIOM nay tự host xác định + struct/str ABI nhất quán.
