# AXIOM — Sổ tay lỗi (Known Bugs & Fix Playbook)

Tài liệu này ghi lại các lỗi đã gặp + cách nhận diện + cách khắc phục, để lần sau không phải mò lại từ đầu.
Trọng tâm: **native self-hosting** (stage1 → stage2 → stage3, `-self-link`).

Bối cảnh kiến trúc cần nhớ:
- **str = giá trị 16-byte** `{ptr: i64, len: i64}` (type_id 12, size 16). Struct = con trỏ heap 8-byte.
- **ABI giá trị của AXIOM** (native backend): tham số 16-byte (str) truyền **bằng con trỏ** (LEA &val vào thanh ghi arg); trả 16-byte qua **RAX:RDX**. Khác hẳn ABI C/Win64 (`ax_runtime.dll`: arg by-ref, trả 16-byte qua hidden-pointer RCX) → KHÔNG tương thích.
- **2 backend**: C backend (stage0 Go `bin/axc.exe` → build stage1 qua C) và **native backend** (logic trong `bootstrap/stage1/x86_*.ax`, chạy khi `-self-link`). **Một bug chỉ ở native backend sẽ KHÔNG lộ khi build stage1** (vì stage1 build bằng C backend) — chỉ lộ khi stage2 chạy/tự-compile.
- Build stage1: `scripts/rebuild_stage1.ps1` (~4 phút, regen `tmp_concatenated_air.ax` từ các file `bootstrap/stage1/*.ax` rồi build qua C backend). Build stage2/3 + SHA: `scripts/build_stage2_3_sha.ps1` (~10 phút).

---

## BUG CLASS #1 — `@compiler_intrinsic("...")` KHÔNG được native backend lower ⚠️ (LẶP LẠI NHIỀU LẦN)

**Mức độ:** Cao. Đã cắn 3 lần (concat, replace, rồi slice/contains/starts_with/ends_with).

### Mô tả / Triệu chứng
- Hàm stdlib có thân `return @compiler_intrinsic("str_xxx")` chạy **đúng dưới C backend** nhưng **trả rác dưới native** (str có `.ptr = NULL`, `.len` rác; bool/i64 rác).
- Hệ quả điển hình: `@memcpy(dst, s.ptr, n)` với `s.ptr == NULL` → **crash 0xC0000005** trong `ucrtbase!memmove`.
- Stage2 tự-compile chết tại các hàm dùng những intrinsic này (vd `concatenate_stdlib` dùng `std.string.replace` → `res_src.ptr == NULL` → memcpy crash).

### Vì sao xảy ra
Native `air_builder.ax` (`lower_call_expr`, ~line 860) **chỉ** lower vài intrinsic: `size_of`, `is_windows`, `is_linux`, `is_macos`, `os_name`, `arch_name`, `path_separator`. **Mọi intrinsic khác** (`str_replace`, `str_slice`, `str_contains`, `str_starts_with`, `str_ends_with`, `str_index_of`, `str_trim`, `atomic_*`, `simd_*`, `gpu_*`, `sha*`, `json_*`, `channel_*`, …) **rơi xuống** thành lời gọi hàm stub `compiler_intrinsic(name)` (std/runtime.ax) → trả giá trị mặc định/rác. C backend (cgen.ax) thì có map đầy đủ các intrinsic này nên KHÔNG lộ.

### Cách nhận diện nhanh
1. Crash `0xC0000005` trong `ucrtbase!memmove`/`memcpy` với **rdx=0 (src=NULL)** (dùng gdb, xem mục Playbook).
2. Truy ngược: con trỏ NULL đó là `.ptr` của một `str` trả về từ hàm stdlib.
3. Mở hàm đó trong `std/*.ax` → thấy `return @compiler_intrinsic("...")`.
4. Xác nhận bằng test cô lập build bằng stage1 `-self-link`: gọi hàm rồi check `(result as ptr[u8]) as i64 == 0`.

### Cách khắc phục
**Viết thân hàm AXIOM value-ABI thật**, dùng các primitive ĐÃ chạy đúng native: truy cập field `s.ptr` / `s.len`, `malloc`, `@memcpy`, index byte `p[i]`, ép `buf as str`. Mẫu (concat đã fix, std/string.ax):
```axiom
pub fn concat(a: str, b: str) -> str:
    let total = a.len + b.len
    let buf = malloc((total + 1) as u64) as ptr[u8]
    @memcpy(buf, a.ptr, a.len)
    @memcpy(((buf as i64) + a.len) as ptr[u8], b.ptr, b.len)
    buf[total] = 0 as u8
    return buf as str
```
Đã viết thân thật cho: `concat, replace, slice, index_of, contains, starts_with, ends_with, trim, to_upper, to_lower, repeat` trong `std/string.ax`.

**Quan trọng:** một intrinsic CHƯA fix chỉ gây lỗi nếu **được GỌI** lúc stage2 chạy. Hàm không gọi → stub vô hại. Khi build path đổi/dùng thêm hàm mới, kiểm tra lại.

**KHÔNG thể fix bằng AXIOM thuần:** `atomic_load/store/cas`, `simd_*`, `gpu_*`, `sha*`, `json_*`, `channel_*`, `quantum` → CẦN native backend lower thật. Hiện compiler KHÔNG gọi chúng lúc build nên để stub. Nếu sau này build path cần → phải thêm lowering trong `air_builder.ax`/`x86_selector.ax`, KHÔNG phải sửa source stdlib.

### Phòng ngừa
- Khi thêm/sửa hàm stdlib mà compiler sẽ gọi: KHÔNG để thân là `@compiler_intrinsic` trừ 6 cái OS đã được lower. Viết thân AXIOM thật.
- Audit nhanh: `grep -rn '@compiler_intrinsic' std/ bootstrap/stage1/*.ax` → đối chiếu với danh sách native lower được trong `air_builder.ax`.

---

## BUG CLASS #2 — `regalloc_is_16byte` đệ quy sai trên OP_CAST

**File:** `bootstrap/stage1/x86_selector.ax` (`regalloc_is_16byte`, ~line 422).

### Triệu chứng
- Crash 0xC0000005 hoặc giá trị không ổn định. Test cô lập: `(sp as i64) != (sp as i64)` trả **true** (cùng biểu thức cho 2 giá trị khác nhau).
- `sp[0]` (OP_INDEX) tình cờ đọc đúng slot thấp → test index đơn giản qua được, che giấu bug.

### Root cause
Code cũ gộp `OP_COPY | OP_MOVE | OP_CAST`: khi dest-type ≠ 16-byte thì **đệ quy xuống src1**. Với cast `str→ptr` (`s as ptr[u8]`): dest=`ptr`(8B) nhưng đệ quy xuống src=`str`(16B) → return true → con trỏ 8-byte bị phân loại **NHẦM 16-byte** → cấp 2 spill slot → mọi lần dùng sau đọc lệch slot.

### Cách fix
Tách `OP_CAST` riêng: cast **ĐỔI kiểu** nên tính size **CHỈ theo dest type_id** (`== 12` hoặc `entries[type_id].size == 16` → true, else false), **KHÔNG đệ quy src1**. COPY/MOVE giữ đệ quy (bảo toàn kiểu).

---

## BUG CLASS #3 — Spill-all fallback cấp 1 slot cho giá trị 16-byte

**File:** `bootstrap/stage1/x86_regalloc.ax` (nhánh fallback, ~line 407-415).

### Triệu chứng
- Hàm LỚN (str-heavy) crash, hàm nhỏ tương tự thì chạy đúng → khó reproduce cô lập.

### Root cause
Hàm lớn (`intervals > 50000` hoặc `insts_len > 5000`, vd `concatenate_stdlib`, `main`, nhiều hàm compiler) rơi vào nhánh **spill-all fallback** cấp **mỗi vreg đúng 1 slot 8-byte** (`spill_count_s + 1`) — kể cả str 16-byte → mỗi str spill (16B) **đè slot vreg kế tiếp**. Nhánh graph-coloring (~line 586-588) thì đã cấp đúng 2 slot cho 16-byte, nên test nhỏ qua được.

### Cách fix
Trong fallback, cấp 2 slot cho vreg 16-byte (khớp graph-coloring):
```axiom
if regalloc_is_16byte(sv, fn_ptr, table, symbols):
    spill_count_s = spill_count_s + 2
else:
    spill_count_s = spill_count_s + 1
```

---

## BUG CLASS #4 — str param 16-byte aliasing (đã fix trước session này)

**File:** `x86_selector.ax` (`emit_param_prologue`).
Hai tham số str (16-byte) liền nhau: đọc tham số sau ra giá trị tham số trước. Gốc: param 16-byte vật hoá bằng LOAD/LEA/STORE với scratch temp KHÔNG được tag bảo vệ arg-register → temp giành thanh ghi của param sau. Fix: `emit_param_prologue` snapshot TẤT CẢ arg-register int bằng MOV (được bảo vệ) TRƯỚC, rồi mới nạp giá trị; set `param_idx_processed = nparams`.

---

## BUG CLASS #5 — stage1 OOM khi build stage2 (`ax_alloc: out of memory`)

**File liên quan:** allocator `runtime/axalloc/axalloc.c` (`ax_alloc` wrap `malloc`, panic line 65); regalloc `bootstrap/stage1/x86_regalloc.ax`.

### Triệu chứng
- Build stage2 dừng giữa chừng (vd codegen func ~483/748), log có:
  ```
  AXIOM PANIC in 'bin\axc_stage1.exe': ax_alloc: out of memory
  Stack trace: #0 ... #9 ...
  ```
- KHÔNG phải bug stage2 — là **chính stage1** (compiler) hết bộ nhớ khi sinh mã native (`-self-link`).

### Phân tích
- `ax_alloc` chỉ wrap `malloc` → NULL nghĩa là process xài hết RAM khả dụng. Kiểm tra: cả axc.exe và axc_stage1.exe đều **PE32+ (64-bit)** → KHÔNG bị giới hạn 2GB. Free RAM lúc kiểm tra ~7.7/16GB.
- Allocator là `axalloc.c` (debug, canary) — **+48 byte overhead mỗi alloc** (header 32B + footer 16B) + danh sách liên kết, KHÔNG nén. Compiler tự-compile (~1.2MB source, ~480+ hàm giữ AIR+mach đồng thời) → đỉnh bộ nhớ cao; source lớn hơn (vd thêm thân hàm string thật) đẩy qua mức khả dụng.
- `graph_size = max_vreg + 1` (x86_regalloc.ax:275), `adj = @alloc(graph_size*24)` + cạnh O(V²). Nhánh spill-all (`intervals>50000 || insts_len>5000`) né đồ thị O(V²); `adj` ĐƯỢC free mỗi hàm (line 628) nên không leak per-func. → OOM là **tích luỹ toàn cục**, không phải 1 alloc khổng lồ.

### Cách nhận diện
1. Log `ax_alloc: out of memory` + stack trace trong `bin\axc_stage1.exe` (KHÔNG phải stage2/stage3).
2. `Get-CimInstance Win32_OperatingSystem` xem FreePhysicalMemory; `Get-Process *axc*` tìm tiến trình mồ côi giữ RAM.
3. Kiểm tra có nhiều build nền song song / git op nặng cùng lúc gây spike không.

### Cách khắc phục (thứ tự ưu tiên)
1. **Kill tiến trình mồ côi + đóng app khác, build lại** (`Get-Process *axc* | Stop-Process -Force`). Nếu là spike tạm thời (cleanup/git/build chạy song song) → retry là đủ.
2. Nếu **deterministic** (OOM cùng 1 func mỗi lần): giảm footprint stage1:
   - Hạ ngưỡng spill-all để nhiều hàm dùng đường nhẹ-bộ-nhớ: `REGALLOC_INSTS_LIMIT` (5000) / `REGALLOC_GRAPH_LIMIT` (50000) trong x86_regalloc.ax — đánh đổi chất lượng mã.
   - Dùng allocator release (bỏ canary 48B/alloc) cho stage1.
   - Free cấu trúc trung gian (AIR sau khi đã sinh mã) — cần sửa pipeline.
3. KHÔNG để source phình vô cớ: hàm stdlib KHÔNG cần trên build path thì không nhất thiết viết thân thật (giữ stub intrinsic nhỏ hơn) — cân nhắc nếu sát ngưỡng OOM.

### Phòng ngừa
- Theo dõi đỉnh RAM khi self-compile; compiler càng lớn càng sát trần. Cân nhắc allocator nhẹ + giải phóng AIR sớm trước khi source phình thêm.

---

## BUG CLASS #6 — `mut self` value-struct receiver qua con trỏ KHÔNG truyền by-reference ⚠️ (BLOCKER lexer)

**Mức độ:** Cao — chặn stage2 ngay khi vào lexer (`tokenize`).

### Triệu chứng
- Stage2 qua được `concatenate_stdlib`, in `[Debug] Stage 1: Starting compilation...` rồi **crash 0xC0000005 trong lexer** (trước `[Debug] Finished Lexing.`).
- gdb: crash tại `mov %r12d,(%rcx)` = `IntVec.push` dòng `self.data[self.len] = val`, với **`self` (rdi) là địa chỉ STACK** và `self.data` (rcx) trỏ vào **.text (mã)** — tức `self` là BẢN SAO stack đọc sai/rác.
- Bản rút gọn (1 push): kết quả SAI (len=0 thay vì 1) — mutation KHÔNG được ghi lại. Bản đầy đủ (vòng lặp + grow/memcpy): CRASH.

### Root cause
Phương thức nhận receiver **by value** nhưng khai báo `mut` (vd `pub fn push(mut self: IntVec, val: i32)` trong lexer.ax) khi gọi trên một **field value-struct qua con trỏ** (`self.newline_offsets.push(...)` với `self: ptr[Lexer]`; field `newline_offsets: IntVec` là struct giá trị inline): native backend truyền receiver **by-value (copy lên stack)** thay vì **by-reference (&field)**. Hậu quả: (1) mutation (grow đặt `self.data`, append `self.len`) mất khi return; (2) bản copy đọc/ghi lệch → `self.data` thành địa chỉ rác (mã) → ghi vào trang read-only → crash. C backend truyền đúng by-reference cho lvalue receiver nên KHÔNG lộ.

### Cách nhận diện / reproduce nhanh (stage1 `-self-link`)
- `bin/t_min.ax` (đã lưu trong `tests/codegen/`): struct `H { x: i64, v: MyVec }`; `fn fill(h: ptr[H]): h.v.push(7)`; main alloc H, gọi fill, trả `h.v.len`. **Đúng = 1; bug = 0 (mutation mất)**.
- Bản loop + grow đầy đủ (`t_lexlike`/`t_off2`) → CRASH 0xC0000005.
- Điều kiện kích hoạt: gọi method `mut self: <ValueStruct>` trên **field qua con trỏ** hoặc **field ở offset ≥ ~24** / trong hàm helper nhận `ptr[Struct]`. Gọi inline trong main trên local đôi khi may mắn đúng → đừng kết luận từ 1 test.

### Cơ chế chính xác (ĐÃ XÁC MINH BẰNG DUMP ASM — root cause thật)
⚠️ Giả thuyết cũ "register_params block-copy" là SAI (đã bác bỏ). Selection + arg-pass đều ĐÚNG.
Bằng chứng quyết định: wire `compile_native_asm("tmp_dump.s",...,"nasm")` vào `main_air.ax` (Stage 3, chỉ self-link) → dump asm thật của `ax_H_fill`:
```asm
ax_H_fill:
    mov rbx, rcx              ; rbx = h (ptr[H])
    mov [rbp - 24], r11       ; SPILL h vào slot [rbp-24]
    lea rax, [rbp - 16]       ; <<< BUG: self = &(home_của_h + 8) = ĐỊA CHỈ STACK, KHÔNG phải [value(h)+8]
    mov rcx, rax              ; rcx = self = con trỏ vào stack frame của fill
    call ax_MyVec_push
```
- `getfld(h, field v)` (size 24 > 16) → selector phát `MACH_LEA dest, [src1=h + disp=8]` (đúng: muốn `value(h)+8` = địa chỉ heap của h.v).
- NHƯNG khi `h` bị **spill**, pass `insert_spill_code` (x86_regalloc.ax ~775) coi **MỌI `LEA` của vreg spilled = "địa chỉ home của slot"** → rewrite thành `[RBP + slot_offset + disp]` = `[rbp-24+8] = [rbp-16]`. Đó là địa chỉ STACK HOME của h **cộng disp**, KHÔNG phải giá trị-con-trỏ-trong-slot cộng disp.
- ⇒ self trỏ vào stack frame của `fill`. push ghi `self.data/len/cap` vào stack (t_min: heap h.v.len vẫn 0 → trả 0). Với grow/memcpy (t_off): ghi đè saved-regs/return-addr → CRASH 0xC0000005. Khớp gdb "self.data = địa chỉ .text/rác".
- Receiver LOCAL chạy được vì local đúng là 1 object có home trên stack — home-address chính là đối tượng (không có indirection con trỏ).
- `LEA` home-address (đúng) vs `LEA` value+disp (GET_FIELD) phân biệt được bằng **src2**: GET_FIELD size>16 là `MACH_LEA` DUY NHẤT có `src2 = OPND_IMM(disp)`; mọi LEA address-of-home (MAKE_REF, materialize 16-byte) dùng `src2 = OPND_NONE`.

### Cách khắc phục (ĐÃ ÁP DỤNG — x86_regalloc.ax insert_spill_code, nhánh src1)
Tách nhánh `MACH_LEA` của vreg spilled trong `insert_spill_code`:
- `if inst.op == MACH_LEA and inst.src2.kind == OPND_IMM:` (GET_FIELD trên con trỏ/by-ref): slot chứa **giá trị con trỏ** → `LOAD R10, [RBP+offset]` rồi giữ nguyên `src1=R10` + `src2=disp` (không cộng disp vào offset). ⇒ `LEA dest, [R10+disp]` = `value(ptr)+disp` (heap, đúng).
- `elif inst.op == MACH_LEA:` (MAKE_REF / &self / 16-byte address-of, src2=NONE): giữ rewrite cũ `src1=RBP, src2=IMM(offset)` (địa chỉ home).
Không đụng selection/arg-pass/register_params (chúng vốn đúng). Fix tổng quát mọi `getfld qua con trỏ bị spill`, không chỉ `mut self`.

### Phòng ngừa
- Test codegen `tests/codegen/t_min.ax` (đúng=1), `t_off.ax`/`t_off2.ax` (đúng=99). Đưa vào regression.
- Khi thêm MACH_LEA mới ở selector: nhớ quy ước **src2=IMM ⇔ value+disp (deref pointer)**, **src2=NONE ⇔ home address**. Đừng phá quy ước này.

---

## BUG #7 — GET_FIELD trên field là STRUCT LỒNG NHỎ (≤15B) load value rồi deref như con trỏ (crash stage2→stage3)

### Bối cảnh
Sau khi fix BUG #6, **stage2 self-compile THÀNH CÔNG** (`bin/axc_stage2_native.exe` 1.65MB). Nhưng stage2 build stage3 thì **SIGSEGV** ngay sau "Finished Lexing".

### Triệu chứng (gdb)
```asm
mov rdx, [arr_base + i*0x18]   ; phần tử mảng 24-byte
mov rax, [rdx]                 ; arr[i].field0 (offset 0, 8 byte) = GIÁ TRỊ INLINE
=> movzwq 0x2(%rax), %rcx      ; deref [rax+2] như con trỏ → CRASH; rax=0x500000007 (={7,5} hai i32)
   or %r12, %rax
   mov %ax, 0x2(%r13)          ; ghi u16 tại [r13+2]
```
`rax` = giá trị inline của một **struct con ≤8 byte**, bị dùng làm con trỏ để truy cập sub-field u16 tại offset 2.

### Root cause
`OP_GET_FIELD` (x86_selector.ax) phân nhánh theo **SIZE**: `>16`→LEA(địa chỉ, đúng cho struct lớn); `==16`→copy 2 nửa (str); **`else` (≤15)→`LOAD [src1+disp]`** coi field là **scalar** và load giá trị. NHƯNG nếu field đó **bản thân là struct (aggregate) nhỏ ≤15B** (vd `Inner` lồng trong `Outer`), thì truy cập sub-field tiếp theo (`outer.inner.x`) sẽ deref giá trị inline như con trỏ → crash. Trong backend này **mọi struct value (kể cả nhỏ) được giữ BY ADDRESS** (literal→OP_ALLOC→con trỏ), nên getfld của field-aggregate phải trả **ĐỊA CHỈ field**, không phải load bytes. (Struct 9–15B cũng hỏng vì LOAD size 12 không hợp lệ.)

### Reproduce nhanh
`tests/codegen/t_nested.ax`: `struct Inner{a:i16,b:i16,c:i32}` (8B) lồng trong `struct Outer{inner:Inner,p:i64,q:i64}` (24B); helper `touch(arr: ptr[Outer], i, val): arr[i].inner.b = arr[i].inner.b | val`. **Đúng=99; bug=CRASH 0xC0000005**. (`t_small.ax` struct nhỏ trả-by-value/truyền-by-value KHÔNG lộ — phải có **field aggregate lồng + truy cập sub-field**.)

### Cách khắc phục (ĐÃ ÁP DỤNG)
- Thêm `field_is_aggregate(table, struct_type_id, field_idx) -> bool` (x86_selector.ax, cạnh `field_size`): điều hướng giống field_size, trả true nếu **kiểu của field là aggregate**: STRUCT(1)/ARRAY(3)/TUPLE(5)/SUM(6)/GENERIC_INST(8)/OPTION(11)/RESULT(12). (Ban đầu chỉ STRUCT/GENERIC_INST → stage3 vẫn crash y hệt vì field là **SUM/OPTION/RESULT** (8-byte tagged, pattern `or tag`); phải mở rộng đủ kind. Tra `TYPE_KIND_*` ở typetable.ax.)
- Trong `OP_GET_FIELD`, thêm nhánh trước `else`: `elif field_is_aggregate(...)`: phát `MACH_LEA dest, [src1 + disp]` với **src2=OPND_IMM(disp)** (trả địa chỉ field; src2=IMM để spill-rewrite BUG#6 xử lý đúng khi src1 spill). `else` (scalar) giữ LOAD.
- Không cần sửa OP_SET_FIELD: base của `setfld` lồng đến từ getfld (giờ trả địa chỉ) nên ghi cũng đúng.

### Phòng ngừa
- Quy ước: **mọi aggregate value giữ by-address**; getfld field-aggregate → địa chỉ (LEA src2=IMM), getfld scalar → LOAD. Đừng phân nhánh getfld chỉ theo size.
- Regression: `tests/codegen/t_nested.ax` (đúng=99).

---

## BUG #9 — OP_INDEX phần tử aggregate LOAD value thay vì trả ĐỊA CHỈ (crash Parser.set_flags)

### Triệu chứng
Cùng họ BUG #7/#8 nhưng ở OP_INDEX. Crash `Parser.set_flags` (parser.ax:140) khi stage2 build stage3:
```
self.tree.nodes.data[node].flags = self.tree.nodes.data[node].flags | flags
```
`nodes.data: ptr[AstNode]` (phần tử 24B), `.flags` là u16 @offset 2. gdb: `mov rax, [&elem]` (LOAD 8 byte field0=0x500000007) rồi `movzwq [rax+2]` (deref) → crash.

### Cách tìm (công cụ vàng — asm dump)
Wire `compile_native_asm("tmp_dump.s",...,"nasm")` vào `main_air.ax` Stage 3, rebuild stage1, build full compiler `-self-link` → tmp_dump.s = asm thật của toàn compiler (~65k dòng, label `ax_Struct_method`). Grep pattern crash (`or rax, r12` + `mov word [..+2], ax`) → `awk` tìm label `:` gần nhất → ra **`ax_Parser_set_flags`** → đọc source. (Map RVA→hàm rất khó vì PE strip; asm dump theo tên hàm là cách nhanh nhất.)

### Root cause
`OP_INDEX` (x86_selector.ax) nhánh `else` (size != 16) phát `MACH_LOAD dest = [base + i*size]` — coi phần tử là scalar, LOAD bytes. Với phần tử **aggregate** (struct/sum 24B), phải trả **ĐỊA CHỈ** `base+i*size` (quy ước aggregate by-address) để `.field` sau đó deref đúng. (t_nested may mắn qua được nhờ **O1 optimizer fold** redundant load ở case đơn giản; chuỗi sâu `self.tree.nodes.data[..]` của parser không fold → lộ bug. Đừng dựa optimizer.)

### Cách khắc phục (ĐÃ ÁP DỤNG)
- Thêm `type_is_aggregate(table, type_id) -> bool` (x86_selector.ax): true nếu kind ∈ {1,3,5,6,8,11,12}.
- `OP_INDEX`: `if size==16: <str copy>; elif type_is_aggregate(inst.type_id): MOV dest = tmp_addr (địa chỉ phần tử); else: LOAD + mask`. Chuyển mask size 1/2/4 vào TRONG nhánh scalar `else` (không mask địa chỉ).
- Lợi ích phụ: `arr[i].field = x` (ghi field qua index) và `dst[j] = src[i]` (copy struct giữa mảng) cũng đúng (block-copy từ địa chỉ).

### Phòng ngừa
- Regression: repro `data[i].directfield` (u16 @offset 2) qua pointer-chain — phải ĐÚNG, không crash. Cùng quy ước "aggregate by-address" với getfld.

---

## BUG #10 — OP_SET_FIELD/OP_STORE ghi ĐỊA CHỈ aggregate thay vì block-copy NỘI DUNG (≤8B) → regress t_nested + parser đọc Token sai

### Bối cảnh — MÔ HÌNH NHẤT QUÁN "aggregate by-address"
Sau BUG #9, t_nested **regress 99→1** và parser stage3 đọc Token sai (lỗi parse rác `error: e2��`, kind/offset rỗng). Token = struct 8 byte `{kind:u8, padding:u8, len:u16@2, offset:u32@4}`. Đây là lỗi nhất quán hoá: backend giữ **mọi aggregate BY ADDRESS** (vreg chứa con trỏ tới dữ liệu), nhưng các opcode GHI lại không đồng bộ:
- **ĐỌC/giữ/truyền/trả** aggregate → ĐỊA CHỈ: OP_INDEX (BUG#9), OP_GET_FIELD (BUG#7/#8), OP_ALLOC, OP_COPY/MOV (MOV địa chỉ = alias), param (caller MOV địa chỉ, callee MOV vào vreg), OP_RETURN/call-result (MOV RAX=địa chỉ). ✓ (đã đúng/đã sửa)
- **GHI vào bộ nhớ** aggregate (setfld/store) → phải **block-copy NỘI DUNG** (size byte từ [src_addr]).

### Root cause
`OP_SET_FIELD` và `OP_STORE` chỉ block-copy khi `size > 8`. Với aggregate **≤8 byte** (Inner 8B, Token 8B): size==8 → nhánh `else` phát `MACH_STORE` ghi **8-byte ĐỊA CHỈ** của src vào field/slot, thay vì copy nội dung. ⇒ field chứa con trỏ trong khi reader coi là inline (getfld/index trả địa chỉ) → đọc lệch.
- t_nested qua được TRƯỚC BUG#9 do **trùng khớp ngẫu nhiên**: OP_INDEX cũ LOAD field0 (= con trỏ mà setfld đã ghi) → vô tình deref đúng. BUG#9 (OP_INDEX trả địa chỉ) bỏ "nạng" này → lộ.
- Parser: lexer `TokenVec.push(t: Token)` lower `data[len] = t` → OP_STORE ghi **địa chỉ** của t (stack temp caller) vào mảng → token rác sau khi push return → parser đọc kind/offset rác → lỗi parse giả.

### Cách khắc phục (ĐÃ ÁP DỤNG)
- `OP_SET_FIELD`: `if size > 8 or field_is_aggregate(table, struct_type, src2): emit_block_copy(...) else MACH_STORE`.
- `OP_STORE`: thêm `store_is_agg = type_is_aggregate(table, inst.type_id)`; cả 2 nhánh (indexed có src2-offset + direct) đổi `if size > 8` → `if size > 8 or store_is_agg` (block-copy nội dung).
- Scalar (≤8B, value trong reg) giữ STORE. str (16B) đã có nhánh riêng.

### Kiểm chứng
t_nested=99 (từ 1), `tests/codegen/t_tok.ax`=99 (repro Token 8B: index field + return-by-value peek), không regress (t_min=1, t_off2=99, test_sum_native=0, opt_test=103, t_castbug=0, t_strfns=0).

### Quy ước CHỐT (đừng phá)
**Aggregate (struct/array/tuple/sum/option/result, mọi size) = BY ADDRESS.** Đọc → trả địa chỉ; ghi → block-copy nội dung. Mọi opcode mới phải tuân: dùng `type_is_aggregate`/`field_is_aggregate` để phân biệt aggregate vs scalar, KHÔNG chỉ dựa size.

---

## BUG #11 — OP_MAKE_REF (`&aggregate`) trả địa chỉ SLOT thay vì giá trị (địa chỉ aggregate) khi spill → hỏng `print_raw_ptr`/reinterpret → mọi %d/%s in rỗng

### Triệu chứng
Diagnostic của stage2 in `%d`/`%s` RỖNG (vd `total_len=` rỗng, parser `error: ??? at offset` + `tokens[]: kind= offset=` rỗng). PRE-EXISTING (chỉ lộ khi native chạy print path nhiều lần ở stage2), KHÔNG do BUG#6–10.

### Cô lập (build NHỎ, nhanh)
`bin/t_pstr.ax` 3 cách in str dựng runtime: m1 `buf as str` ✓, **m2 `let s=AxiomString(ptr,len); print((&s as ptr[str])[0])` RỖNG ✗**, m3 `slice` ✓. m2 = đúng pattern `print_raw_ptr` (print_helpers.ax) dùng. Dump asm `main`:
```asm
mov [rbp-176], r11    ; s = địa chỉ HEAP của AxiomString {buf,2} (aggregate by-address)
lea rax, [rbp-176]    ; &s = địa chỉ SLOT (SAI)
mov rcx, [rbx]        ; [&slot] = con trỏ heap (KHÔNG phải {buf,2})
```

### Root cause
`OP_MAKE_REF` phát `MACH_LEA dest, [src1]` (src2=NONE). Khi src1 ở REGISTER → LEA[reg]=giá trị (đúng cho aggregate). Khi src1 **SPILL** → spill-rewrite (nhánh home-address src2=NONE) thành `LEA [RBP+slot]` = địa chỉ slot. Với **aggregate** (vreg giữ ĐỊA CHỈ heap, by-address), `&agg` phải = giá trị (địa chỉ heap = nơi data sống), KHÔNG phải &slot. Scalar thì &x = &slot là đúng (value sống trong slot).

### Cách khắc phục (ĐÃ ÁP DỤNG)
`OP_MAKE_REF`: `if type_is_aggregate(get_register_type(src1)): MACH_MOV dest=src1 (copy giá trị — spill-rewrite tự LOAD [slot]) else: MACH_LEA dest=[src1] (địa chỉ slot)`. Kiểm chứng: t_pstr m2 in "hi", không regress (t_tok=99,t_nested=99,t_min=1,t_off2=99,t_printf OK).

### Lưu ý còn lại
`bin/t_p64.ax` (replica print_i64_raw đầy đủ) vẫn in sai số nhiều chữ số — KHÁC bug (reverse-loop/print_raw_ptr replica của riêng test, hoặc spill trong hàm ~7 locals). Digit-loop cơ bản ĐÚNG (t_printf big=12345 ✓). print_raw_ptr THẬT (AxiomString reinterpret) đã đúng sau BUG#11.

---

## BUG #12 — Result/Option.unwrap() làm hỏng payload → stage2 crash 0xC0000409 khi đọc file nguồn (REGRESSION do BUG#7–11) ⚠️ BLOCKER stage3

### Triệu chứng
- stage2 build OK (1.66MB) nhưng **crash ngay** khi compile BẤT KỲ file nào (kể cả `t_min.ax` nhỏ nhất): `0xC0000409` (STATUS_STACK_BUFFER_OVERRUN / invalid-param fast-fail), **log 0 byte**.
- gdb: crash trong `ucrtbase!fread` → `_invalid_parameter` (RCX/stream = rác hoặc 0). stage2 KHÔNG hề gọi fread trong code người dùng → là `std.io.read` (`fread(buf,1,count,self.handle)`) với `self.handle` RÁC.
- Truy ngược: `read_file_content` → `std.io.open()` trả `Result[File,str]`; `r.unwrap()` trả về **File có handle=0/rác** → `fread(handle rác)` → crash.

### Cô lập (build NHỎ bằng stage1 `-self-link`, ~30s)
- `bin/t_io3.ax`: `fn mk()->Result[H,str]: return Ok(H(handle:123456,...))` ; `main: let h = mk().unwrap(); p(h.handle)` → in **rác/crash** (đáng lẽ 123456). Repro TỐI THIỂU.
- `bin/t_io6.ax` (replicate unwrap thủ công): `raw=(&r) as ptr[u64]; boxptr=raw[0]; masked=boxptr&~1; stored=(masked as ptr[u64])[0]` → cho thấy `boxptr` và `[box]` để soi tầng indirection.
- `bin/t_io2.ax` / `t_io5.ax`: dùng `std.io.open`+`read_all` THẬT để xác nhận end-to-end.
- LƯU Ý BẪY: ĐỪNG đặt tên hàm/struct trùng stdlib (`read_all`/`read`/`File`) trong repro — `-self-link` kéo std.io vào, lời gọi nhảy nhầm vào std.io (collision) → false alarm.

### Root cause (2 phần, cùng gốc: pointer-layout sum bị coi là aggregate by-address)
Constructor `Ok(x)`/`Err(x)`/`Some(x)` (air_builder.ax ~line 1070) **luôn box trên heap** (pointer-layout): `box=@alloc(16); OP_STORE box, payload (8 byte); return box`. Tag (Err) = bit 0. `std/result.ax` `unwrap` đọc layout này: `(&self) as ptr[u64]; box=raw[0]&~1; typed=box as ptr[T]; return typed.*`.
- **(a) MAKE_REF**: `Result/Option/Sum` (kind 6/11/12) là **giá trị 8-byte (con trỏ box)**, KHÔNG phải aggregate-blob nhiều byte. BUG#11 thêm chúng vào nhánh aggregate→`MOV` của OP_MAKE_REF, làm `&self` = giá trị box (không phải địa chỉ slot). → `raw[0]` đọc nhầm `[box]` (PAYLOAD) thay vì con trỏ box. unwrap hỏng.
- **(b) OP_DEREF `typed.*`**: deref `ptr[T]` với T=struct phải **LOAD 8 byte** (con trỏ heap struct lưu trong box — "struct value = 8-byte heap reference"), KHÔNG load `entry.size` (32B, sai width) cũng KHÔNG trả về con trỏ. type_id của deref thường **còn generic** (từ `masked as ptr[T]`) → phải recover kiểu pointee từ operand.

### Cách khắc phục (ĐÃ ÁP DỤNG — 3 chỗ)
1. **x86_selector.ax `OP_MAKE_REF`**: chỉ `MOV` (giá trị) cho aggregate-blob THẬT (struct/array/tuple/generic-inst/str); với **sum/option/result (kind 6/11/12) → `LEA` slot** (như scalar). `&self` = địa chỉ slot → `raw[0]` = con trỏ box. ✓
2. **x86_selector.ax `OP_DEREF`**: nếu deref ra aggregate (kind 1/3/5/6/8/11/12, ≠str) → `type_size = 8` (LOAD con trỏ 8-byte), thay vì `entry.size`. Phát hiện aggregate qua **cả** `type_id` dest (đã recover) **và** pointee của `src_type` (operand `ptr[T]`).
3. **air_builder.ax `lower_deref_expr`**: recover `type_id` của deref = pointee cụ thể khi operand là `ptr[aggregate]` (mirror đúng recovery của str type 12) — để selector thấy kiểu thật thay vì generic T.

### Kiểm chứng
`t_io3`=123456, `t_io6` box+42, `t_io5` unwrap i64=42 + `read_all` đọc 589 byte, `t_io2` read=16 byte, `t_io` struct by-value=123456 — TẤT CẢ exit 0. (trước fix: crash/rác).

### Phòng ngừa / quy ước CHỐT
- **Pointer-layout sum (Result/Option/Sum đã box) = SCALAR 8-byte con trỏ**, KHÔNG phải aggregate by-address. MAKE_REF của chúng = `LEA` slot. Nhưng **field_is_aggregate/type_is_aggregate VẪN coi 6/11/12 là aggregate** cho GET_FIELD/INDEX/STORE (sub-field tag/payload cần địa chỉ — xem BUG#7/#9 comment). Chỉ MAKE_REF là ngoại lệ.
- **Aggregate (struct) VALUE = con trỏ heap 8-byte.** `ptr[struct].*` (OP_DEREF) = LOAD 8 byte (con trỏ), KHÔNG load entry.size, KHÔNG trả &ptr.
- Deref kiểu generic: LUÔN recover pointee từ operand (`masked as ptr[T]` để T generic).

---

## BUG #13 — regalloc gán vreg SỐNG vào RDX/RAX qua lệnh `idiv` (cqo clobber) → kết quả `/`,`%` sai / vòng lặp vô hạn ⚠️⚠️ ROOT CAUSE parser đọc sai token (BLOCKER stage3 chính)

**Mức độ:** RẤT CAO. Là gốc của: parser stage2 đọc sai token, `print_i64_raw` in số rỗng/rác, reverse-loop treo. Heisenbug (thêm/bớt lệnh đổi cấp phát → khi chạy khi sai).

### Triệu chứng
- stage2 chạy full pipeline nhưng **parser báo lỗi token sai** khi build stage3 (`unexpected token`, `expected expression nud`, …); diagnostic `tokens[%d]: kind=%d` in RỖNG.
- Test nhỏ: `print_i64_raw(12345)` in rác/rỗng (chỉ `0` đúng); reverse-loop (`while i < len/2: swap`) **treo vô hạn** trong hàm nhiều local.
- Heisenbug: thêm `putchar` marker giữa vòng lặp → HẾT lỗi (đổi register allocation).

### Cô lập (vàng) — đọc ASM
1. Wire tạm asm dump trong `main_air.ax` (sau Stage 4): `let _ = compile_native_asm("tmp_dump.s", air_builder.module, resolver.symtable, parser.pool, typetable, "nasm")`.
2. Build test nhỏ `-self-link` → `tmp_dump.s`. Tìm label `main:` (hoặc `ax_<fn>`), đọc vòng lặp.
3. Thấy:
```asm
.L_loop:
    mov rax, r12     ; len
    cqo              ; ⚠️ CLOBBER RDX — nhưng RDX đang giữ i (loop counter)!
    idiv rcx         ; len/2
    cmp rdx, rax     ; so sánh rdx(=len%2 rác) thay vì i  → điều kiện sai → vô hạn
```

### Root cause
`idiv` (x86) chia `RDX:RAX / src` → **luôn ghi đè RAX (thương) và RDX (dư)**; `cqo` trước đó cũng ghi đè RDX. Regalloc (`x86_regalloc.ax`) chỉ cấm RAX/RDX cho **dividend** (`src1` của MACH_IDIV, ~line 337 `forbid_rax_rdx[src1]=true`), **KHÔNG** cấm cho vreg có live-range **CẮT QUA** idiv. Nên một biến sống (vd `i`) được cấp RDX và bị division huỷ. Mọi hàm có `/` hoặc `%` (lexer offset, parser, hash, print số…) đều dính nếu allocator lỡ chọn RDX/RAX cho biến sống.

### Cách khắc phục (ĐÃ ÁP DỤNG)
`x86_regalloc.ax`: thêm tính **`spans_idiv`** y hệt `spans_call`: thu thập vị trí mọi `MACH_IDIV` (`idiv_pos[]`), với mỗi interval, nếu `istart < idiv_pos < iend` → `forbid_rax_rdx[vreg] = true` (cơ chế forbid RAX+RDX có sẵn ở ~line 573). Đặt ngay sau khối `spans_call`/`@free(call_pos)`.

### Kiểm chứng
`t_p64` in `0/12345/-678` ĐÚNG (trước: rác); `t_p64e` reverse-loop hết treo "54321/12345/12345"; t_io3/t_min/t_tok/t_nested OK.

### Phòng ngừa / quy ước CHỐT
- **Bất kỳ lệnh ghi đè thanh ghi cố định (idiv→RAX/RDX, shl/shr biến→RCX, call→caller-saved)** PHẢI cấm thanh ghi đó cho MỌI vreg sống-xuyên-qua, không chỉ toán hạng tại điểm đó. Mẫu chuẩn: `spans_call` (đã có), `spans_idiv` (BUG#13), `forbid_rcx` cho shift.
- Khi nghi codegen sai trong vòng lặp/hàm: **dump asm và đọc** — nhanh hơn đoán. cqo/idiv quanh biến vòng lặp là cờ đỏ.

---

## BUG #14 — str `==` (ax_str_eq) trả bool 1-byte (AL) nhưng native đọc full RAX (rác) → keyword lookup sai → parser đọc sai token ⚠️⚠️ ROOT CAUSE THẬT của stage3 token-misread

**Mức độ:** RẤT CAO. `str ==` dùng KHẮP NƠI (lexer keyword lookup, name resolution, intern compare). Sai → toàn bộ phân loại keyword sai → parser đọc token rác → stage3 fail. Là blocker CUỐI sau BUG#12/#13.

### Triệu chứng
- stage3 (và stage2 compile file nhỏ — vì stdlib được nối vào trước) báo `unexpected token`, `expected expression nud`… tại offset trong `print_helpers.ax`/`ax_printf_local`. Token kind đọc ra là KEYWORD (34=STRUCT, 31=PUB, 24=LET…) cho IDENTIFIER (`val`, `arg_idx`).
- Diagnostic đọc được (sau BUG#13) mới lộ ra.

### Cô lập (vàng) — test str== trực tiếp
`bin/t_streq.ax`: `check(a,b,expect)` in `.`/`X`. Kết quả `.....XX.`:
- ĐÚNG: khác độ dài (`"val"=="struct"`→false), giống hệt (`"if"=="if"`→true).
- SAI: **cùng độ dài, khác nội dung** (`"a1"=="as"`→true, `"elif"=="else"`→true).
→ Pattern: chỉ case đi qua `memcmp` (con trỏ khác + cùng len) mới sai. `bin/t_se2.ax` (1 so sánh) in `T` cho `"a1"=="as"`.

### Phân biệt nhanh
- `bin/t_memcmp.ax` gọi `memcmp` TRỰC TIẾP → ĐÚNG ("DS"). Nên memcmp không lỗi.
- Asm dump `ax_check`: args truyền ĐÚNG (`lea rcx,&a; lea rdx,&b; call ax_str_eq`), data str literal đúng (label khác nhau). → lỗi ở XỬ LÝ KẾT QUẢ.

### Root cause
`runtime/ax_runtime.h`: `typedef _Bool ax_bool;` → **1 BYTE**. `ax_str_eq` (ax_runtime.dll) trả `ax_bool` trong **AL**; upper 56 bit của RAX là RÁC (nhánh `memcmp()==0 ? :` để lại giá trị memcmp ở RAX). Nhánh len-mismatch dùng `xor eax,eax` (sạch) → vì sao chỉ memcmp-path sai. `x86_selector.ax` OP_EQ str path làm `MOV dest, RAX` (full 64-bit) và OP_NE làm `CMP RAX, 0` — đều đọc cả rác → `==` trả true sai cho chuỗi cùng-độ-dài-khác-nội-dung.

### Cách khắc phục (ĐÃ ÁP DỤNG)
`x86_selector.ax` OP_EQ/OP_NE str (src_type==12): ngay sau `call ax_str_eq` (+ `add rsp,32` win64), chèn `MACH_MOVZX_B RAX, RAX` (zero-extend AL→RAX) TRƯỚC khi xử lý OP_NE / `MOV dest`. Đảm bảo chỉ đọc byte bool.

### Kiểm chứng
`t_streq` → `........` (8/8 đúng); `t_se2` → `F`. (trước: `.....XX.` / `T`).

### Phòng ngừa / quy ước CHỐT
- **Mọi hàm runtime/C trả type < 8 byte (bool/_Bool, u8/u16/u32, char)** → chỉ tin AL/AX/EAX, PHẢI zero/sign-extend trước khi dùng như i64. RAX upper bits là rác. Kiểm các call runtime khác trả bool/i32 (vd is_* trả bool, ax_str_contains/starts_with/ends_with cũng trả ax_bool → cùng pattern, đã/đang dùng MOVZX chưa?).
- C `_Bool`/`bool` = 1 byte, KHÔNG phải int.

---

## BUG #15 — `copy_prop_func` (SSA optimizer -O1) rewrite NHẦM field-index của GET_FIELD/SET_FIELD → đọc/ghi SAI field ⚠️⚠️ ROOT CAUSE crash NameResolver `pre_define_top_levels` ở stage3

**Mức độ:** RẤT CAO. Chỉ xảy ra ở `-O1` (optimizer bật). stage2/stage3 build bằng `-self-link -O1` → mọi struct có field-index trùng số với một vreg-id đang nằm trong copy-chain đều bị hỏng. Là blocker stage3 SAU BUG#12/#13/#14 (sau khi lex+parse đã sạch).

### Triệu chứng
- stage3 SIGSEGV trong `NameResolver.pre_define_top_levels` (`mov (%r14),%eax` deref con trỏ rác) tại define thứ ~11 (đúng lúc `Scope` hashmap grow lần đầu, count 49).
- Repro nhỏ KHÔNG tái hiện ở `-O0`; chỉ sai ở `-O1`.

### Cô lập (vàng) — `bin/t_sym.ax`
Mô phỏng `define()`: `Scope` (8-byte entries) trong `Vec`, gọi qua hàm `zdefine(mut syms, mut scopes: ptr[Vec], …)` → `let sc = &scopes.data[0]; sc.zput(...)`; `zput` gọi `zinsert(self,...)` và `self.count = self.count + 1`.
- Bằng chứng QUYẾT ĐỊNH: đọc count qua helper param-pointer `rc(&scopes)` trả **đúng (R=1)** nhưng main đọc TRỰC TIẾP `(&scopes.data[0]).count` trả **stale (n=0)**; địa chỉ KHỚP (cùng `&scopes.data[0]`). → bộ nhớ đúng, lệnh đọc field bị hỏng.
- `-O0` → ĐÚNG (N=50, F=50); `-O1` → SAI (N=0). → pass SSA optimizer.
- Bisect: disable từng pass → **`copy_prop_func`** là thủ phạm.

### Root cause
`OP_GET_FIELD`/`OP_SET_FIELD` mã hoá **FIELD INDEX (immediate)** trong `src2` (air_builder: `src2: field_idx` từ `node.extra_idx`), KHÔNG phải vreg dữ liệu. `copy_prop_func` (ssa_opt.ax ~line 293) rewrite `src2` của MỌI inst non-control theo copy-chain (`copy_map`). Khi field-index (vd 6 = `count`) TRÙNG SỐ với một vreg-id (vreg #6) tình cờ là copy của vreg khác (vd #1), copy_prop đổi `getfld base, 6` → `getfld base, 1` → đọc field 1 (`padding1`, offset 1, =0) thay vì field 6 (`count`, offset 24) → trả 0. (Diff dump-air `-O1`: `getfld %61, %6` → `getfld %60, %1`.)
- CALL không dính vì `opcode_is_control` đã gồm `OP_CALL` (src2=arg_start được loại trừ).
- INDEX/binary-ALU giữ nguyên (src2 = vreg/value thật).

### Cách khắc phục (ĐÃ ÁP DỤNG)
`ssa_opt.ax` `copy_prop_func`: thêm điều kiện loại trừ vào khối rewrite `src2`:
`if inst.src2 != 0 and not opcode_is_control(inst.opcode) and inst.opcode != OP_GET_FIELD and inst.opcode != OP_SET_FIELD:`
→ copy_prop VẪN BẬT (giữ tối ưu), chỉ không đụng field-index.

### Kiểm chứng
`t_sym` full (grow 64→128, 50 entries, prelude str==) `-O1`: `C=128 N=50 S=50 F=50` (trước: `C=128 N=0 ...`). Reduced: n=1,2,3 N=3.

### Phòng ngừa / quy ước CHỐT
- **`src2` của OP_GET_FIELD/OP_SET_FIELD là FIELD-INDEX immediate, KHÔNG phải vreg.** Mọi pass duyệt/rewrite toán hạng (copy_prop, và CẦN KIỂM licm/strength_reduction/cse/dce ở -O2+) PHẢI coi nó là structural, không phải dataflow. Tương tự: src2 của CALL = arg_start (immediate) — đã được che bởi opcode_is_control.
- Bất kỳ toán hạng "immediate đội lốt vreg-id" nào (field index, arg_start, block id) đều dễ bị các pass dataflow ăn nhầm khi giá trị trùng số vreg. Khi thêm pass mới: whitelist toán hạng nào là vreg thật.

---

## BUG #16 — stage2 (NATIVE BACKEND, cả -O0 LẪN -O1) sinh name_id SAI cho 1 identifier stdlib → resolve thất bại → `symbols.data[name_id]` OOB ⚠️⚠️ (ĐANG ĐIỀU TRA)

**Mức độ:** RẤT CAO. Blocker stage3 SAU BUG#15. ⚠️ KHÔNG phải -O1-only — **PIVOT -O0 ĐÃ THẤT BẠI: stage2-O0 crash Y HỆT**. Vậy là **bug NATIVE CODEGEN** (selector/regalloc, họ BUG#6-15), KHÔNG phải optimizer. Mọi build `-self-link` (kể cả chương trình 2 dòng) đều crash vì stdlib luôn được nối vào → crash CÙNG node 2215 (trong stdlib), CÙNG registers (rbx=2215, rdi=1075).

**MANH MỐI ban đầu:** `main_air.ax:393 total_len=%d` in RỖNG ở stage2 (giá trị âm/rác).

### ✅ ROOT CAUSE TÌM RA (repro nhanh `bin/t_intern2.ax`)
- Repro string-intern (mirror intern.ax: arena+fnv1a+byte-compare+grow): intern N chuỗi rồi RE-intern → ở stage1 đúng (tìm lại), ở native-codegen (stage1 -self-link -O1 VÀ -O0) **mọi re-intern THẤT BẠI** (B=600/600, count phình gấp đôi). Tái hiện cả khi KHÔNG grow (3 chuỗi).
- Instrument `zintern`: tại slot đã chiếm, in `entry.hash` vs `h` (hash mới tính): `entry.hash=1706517079` (u32 đúng) NHƯNG `h as i64 = 7836038537563038295` = **0x6CC..._65C55017 — RÁC ở 32 bit CAO**. Lower-32 khớp, upper-32 rác → `entry.hash == h` so ở **64-bit** → KHÔNG bằng → false-negative → chèn trùng → name_id phình (176→1075) → resolve lệch.
- **Bản chất:** `fnv1a` làm `h = h * 16777619 as u32` (u32 multiply). Native backend emit **lệnh 64-bit** (MACH_IMUL width-less, assembler luôn 64-bit) → tích tràn vào upper-32 (rác). x86-64: lệnh **32-bit** tự zero upper-32, lệnh 64-bit thì KHÔNG. `entry.hash` LOAD từ field u32 (zero-extended, sạch) nhưng `h` COMPUTED qua u32-mul (rác upper) → so sánh lệch. Mọi u32-arith-rồi-so-sánh-với-u32-loaded đều rủi ro. CÙNG HỌ BUG#14 (giá trị <64-bit upper rác).

### Hướng fix (đang làm)
- Toán học/so sánh kiểu 32-bit (u32/i32) phải dùng op 32-bit (tự zero upper) HOẶC mask 0xFFFFFFFF. Ưu tiên fix tại OP_EQ/OP_NE (so sánh) width-aware: nếu operand ≤32-bit → mask/so 32-bit. Hoặc mask kết quả OP_MUL/arith khi type 32-bit. Cần xem assembler MACH_CMP/MACH_IMUL có hỗ trợ 32-bit không (MachInst hiện width-less).
- Repro xác nhận fix: `bin/t_intern2.ax` phải in `D=600 B=0` (hiện `D=1200 B=600`).

### Triệu chứng
- stage3 SIGSEGV trong `resolve_node` NODE_FIELD_EXPR: `movzbq 0x4(%r12)` đọc `Symbol.kind` (offset 4) với con trỏ rác; gdb: r12 = symbols.data + idx*24, **idx=1075 >> symlen(507)** → OOB.
- Tái hiện NHANH: chỉ cần stage2 compile chương trình 6 dòng dùng `std.string.concat` (`bin/t_std.ax`). Crash ngay "Resolving root...".

### Cô lập (markers IDENT/FE770/FE798 trong resolver.ax) — so stage1 vs stage2 CÙNG node 2215
- **stage1 (đúng):** `IDENT node=2215 name_id=176 sym_idx=506 symlen=507` → resolve OK.
- **stage2 -O1 (sai):** `IDENT node=2215 name_id=1075 sym_idx=0 symlen=507` → resolve(1075)=0 (KHÔNG tìm thấy).
- **name_id của CÙNG identifier (cùng node AST) KHÁC NHAU: stage1=176, stage2=1075.** symlen giống nhau (507) → intern pool numbering/dedup khác.

### Phân tích
- name_id do parser/intern gán lúc parse (node.payload = intern_string(text)). stage2 -O1 gán 1075, stage1 gán 176 cho cùng chuỗi → **intern_string của stage2 không nhất quán** (cùng chuỗi → id khác giữa parse-time và define-time, hoặc tạo entry trùng). → resolve(name_id) không khớp symbol đã define → trả 0.
- LATENT LOGIC BUG kèm theo (resolver.ax ~line 796-798): `sym_idx = updated_lhs_node.payload` rồi `if sym_idx != 0: symbols.data[sym_idx]` — KHÔNG kiểm payload có phải symbol-index thật không. Khi IDENT chưa resolve (resolve=0), payload GIỮ NGUYÊN name_id (nonzero) → `symbols.data[name_id]` OOB. Chỉ lộ khi intern/resolve sai (BUG#16). Cân nhắc thêm bound-check `sym_idx < symbols.len` (nhưng sẽ MASK bug -O1 thành mis-compile thầm lặng → KHÔNG thêm khi đang debug -O1).

### Trạng thái: PIVOT -O0
- Nhiều bug -O1 codegen trong stage2 (BUG#15 + intern). Săn từng cái qua chu kỳ build stage2 ~2h quá chậm.
- QUYẾT ĐỊNH: build stage2/stage3 ở `-O0` (optimizer TẮT hoàn toàn — `if optimize:` ở main_air.ax bỏ qua `optimizer.run`). Native codegen/regalloc (BUG#6-13 đã fix) chạy ở cả -O0. Nếu stage2-O0 compile stage3-O0 và SHA khớp → self-hosting đạt (bản chưa tối ưu). Bug -O1 để fix sau.
- TODO -O1: tìm bug intern_string ở -O1 (nghi cùng họ field-index/spill trong hàm lexer/parser LỚN). Dùng repro hash-intern pool ở -O1 (chưa tái hiện được — bug context-sensitive trong hàm lớn).

### Công cụ đã dùng
- gdb lấy fault: r12=symbols.data+idx*24, idx=payload. disassemble `$pc-220,$pc+12` trace ngược index.
- markers IDENT (resolver.ax line ~955) + FE770/FE798 (NODE_FIELD_EXPR) so stage1 baseline vs stage2 cùng node index (AST giống nhau → node index giống nhau → so trực tiếp).

---

## BUG #17 — stage2 NATIVE CODEGEN crash trong REGISTER ALLOCATOR: `<u32-field> as i64` rác upper-32 → loop-bound khổng lồ → `allocs[]` OOB ✅ ĐÃ FIX

**Bối cảnh:** Sau khi fix BUG#15+#16, stage2 qua sạch lex/parse/resolve/typecheck/optimizer khi compile `bin/t_std.ax`, rồi crash trong NATIVE CODEGEN.

### Định vị (markers AXCG SEL/LIVE/RA/DONE trong x86_coff.ax codegen driver)
Chạy `axc_stage2_cg.exe build bin/t_std.ax -self-link -O1` (stage2 có markers) → log dừng tại:
```
AXCG fn=0 SEL
AXCG fn=0 mach=38 LIVE
AXCG fn=0 intervals=21 RA      <- in xong "RA" rồi SEGFAULT (không có "DONE")
```
→ crash **bên trong `allocate_registers_orchestrator`** (x86_regalloc.ax), func 0 NHỎ (21 intervals, mach=38) — KHÔNG phải spill-all path như nghi ban đầu.

### Root cause
`allocate_registers_orchestrator` (x86_regalloc.ax:964) merge sub-result GPR/XMM:
```axiom
while idx <= gpr_alloc.max_vreg as i64:        // <- THỦ PHẠM
    let al = gpr_alloc.allocs[idx]              // 12-byte RegAllocation copy
    if al.vreg < graph_size as u32: allocs[al.vreg] = al
```
`RegAllocResult = {allocs:ptr@0, max_vreg:u32@8, spill_count:i32@12}` (16 byte). Native backend mis-compile **`gpr_alloc.max_vreg as i64`** (zero-extend u32 struct-field): load 64-bit ở offset 8 → kéo luôn `spill_count` (offset 12) vào upper-32 → bound **rác khổng lồ**. `gpr_alloc.allocs` chỉ có graph_size≈22 phần tử (264 byte); loop đọc tuần tự `allocs[idx]` đến khi vượt page (~70KB ⇒ idx≈5943) → SEGFAULT. **Họ BUG#14/#16** (giá trị <64-bit upper-bits rác) nhưng ở phía **ĐỌC struct field + `as i64`**.
- gdb khớp 100%: copy 12-byte `{ptr@0,len@8}`, `i<=[vec+8]`, body `if elem.field0<limit: BASE[elem.field0*12]=elem`. "len@8" = `max_vreg` field, *12 = sizeof(RegAllocation). "i=5943" chỉ là chỗ chạm page unmapped, KHÔNG phải len thật.
- Chỉ lộ ở stage2 (native), không ở stage1 (C backend zero-extend u32→i64 đúng).

### Fix (bootstrap/stage1/x86_regalloc.ax)
3 site dùng `<= max_vreg as i64`:
- **L966 / L975** (orchestrator gpr/xmm merge): đổi `while idx <= gpr_alloc.max_vreg as i64` → `while idx < graph_size`. `graph_size` là local i64 sạch (== max_vreg+1, cùng `insts`); `gpr_alloc.allocs`/`xmm_alloc.allocs` đều sized `graph_size`, index 0..graph_size-1 đã memset+init đủ → tương đương về tập phần tử nhưng tránh đọc field rác.
- **L674** (`get_used_callee_saved`, không có graph_size): mask `let mv_bound = (max_vreg as i64) & (4294967295 as i64)` rồi `while i <= mv_bound`. Mask low-32 phục hồi đúng giá trị (low-32 của field LUÔN đúng; chỉ upper-32 rác) — đúng mẫu fnv1a BUG#16.

### Bài học / cảnh giác
- **Mọi `<u32 struct-field> as i64`** dùng làm loop-bound/size đều rủi ro native: backend có thể load 64-bit kéo field kề bên vào upper-32. Ưu tiên bound bằng local i64 sạch, hoặc mask `& 4294967295`.
- Markers per-function (SEL/LIVE/RA/DONE) trong codegen driver là cách định vị nhanh nhất stage/func crash — đáng giữ pattern này.

---

## BUG #18 — struct **16-byte** `{ptr, u32, i32}` trả-by-value bị native backend đọc SAI field (đường str-like RAX:RDX) ✅ ĐÃ FIX — ROOT CAUSE CHUNG với BUG#17

**Mức độ:** Cao, tinh vi. Là **root cause thật** của cả BUG#17 (workaround graph_size chỉ che được ở orchestrator; `get_used_callee_saved`/`compute_frame` vẫn dính).

### Triệu chứng / định vị
- Sau khi fix BUG#17, stage2 qua func 0 register-alloc (AXCG fn=0 DONE) rồi **segfault NGAY** trước "AXCG fn=1 SEL" → crash ở xử lý sau-RA của func 0 (x86_coff.ax driver: `get_used_callee_saved` / `compute_frame` / `insert_spill_code`).
- gdb tại fault `movzbq 0x4(%rcx)` (= `allocs[i].phys`): **r13 (mv_bound, đã mask) == rsi (allocs base) == 0x8b907e8** ⇒ `alloc_res.max_vreg` (offset 8) đọc ra **giá trị của `alloc_res.allocs` (offset 0)** ⇒ bound = con trỏ allocs (~146 triệu) ⇒ `allocs[i]` quét tràn page (i=5975) ⇒ segfault.

### Root cause
`RegAllocResult = {allocs:ptr@0, max_vreg:u32@8, spill_count:i32@12}` = **đúng 16 byte**. Native backend trả struct 16-byte **by-value qua RAX:RDX theo đường "str-like"** (giả định layout `{ptr@0, i64@8}` như `str`). Đường này đọc/ghi SAI khi 8 byte thứ hai là **hai field 4-byte** (`u32@8`,`i32@12`) thay vì một `i64` → đọc `max_vreg`/`spill_count` trả rác (cụ thể trả về luôn giá trị field offset-0). Struct >16 byte đi đường **by-address (sret)** — đường tổng quát đã test kỹ (mọi heap-struct, `LiveIntervalVec` 24-byte by-value chạy đúng).

### Repro CÔ LẬP (vàng — `bin/t_ret16.ax`, build stage1 `-self-link`, ~vài giây)
Struct 16-byte `R16{p:ptr,mv:u32,sc:i32}` trả by-value vs 24-byte `R24{...,pad:i64}`:
```
A=3787587592 B=604 C=21 D=3
```
- A/B (R16 16-byte): RÁC (A = low-32 con trỏ, đáng lẽ 21; B=604 đáng lẽ 3) → BUG.
- C/D (R24 24-byte): ĐÚNG (21, 3) → FIX. **Repro này xác minh fix mà KHÔNG cần build stage2 2h.**

### Fix (bootstrap/stage1/x86_regalloc.ax)
Thêm field padding `pub pad0: i64` vào `RegAllocResult` → 24 byte → trả by-address (sret). Cập nhật 4 constructor (`pad0: 0 as i64`). Không đổi kiểu field, không đổi call-site. Workaround BUG#17 (graph_size ở L966/975, mask ở get_used_callee_saved) GIỮ NGUYÊN — đúng & phòng thủ thêm, vô hại.

### Cảnh giác / việc tương lai
- **MỌI struct đúng 16 byte mà KHÔNG phải layout `{i64,i64}`/`{ptr,i64}`** (vd `{ptr,u32,i32}`, `{i64,u32,u32}`, `{u32×4}`...) trả-by-value đều RỦI RO → đọc field sau offset 0 sai. Workaround: pad >16 byte HOẶC trả qua con trỏ heap.
- Fix gốc đúng đắn (chưa làm, rủi ro ABI): backend xử lý đúng struct 16-byte by-value có sub-8-byte field (không giả định str-layout). Hiện workaround per-struct.

---

## BUG #19 — (CÙNG HỌ BUG#18) struct 16-byte split-word truyền **literal by-value làm ARG** bị mangle — `Fixup` trong `push_fixup` ✅ ĐÃ FIX

**Là cùng root cause BUG#18**, nhưng manifest ở **truyền tham số** (không chỉ return). Sau khi fix BUG#18 (RegAllocResult), stage2 qua func 0,1,2 register-alloc rồi segfault ở `emitter_resolve_fixups` (x86_emitter.ax:442 `e.code.data[fix.offset]=...`).

### Định vị (gdb)
- Fault byte-store `e.code.data[fix.offset+k] = (target-fix.offset-4)>>... & 0xff` (patch rel32). `fix.offset = 0x15e918` = **địa chỉ STACK** (rác), `fix.label_id=0`, đáng lẽ offset nhỏ < code.len(258).
- Dump heap `e.fixups.data[]`: TẤT CẢ fixup giống hệt rác (`off=0x15e918, lbl=0, sz=0`) dù push với buf.len/label khác nhau ⇒ **push_fixup nhận `f: Fixup` rác**, không phải đọc sai.

### Root cause
`Fixup = {offset:i64@0, label_id:u32@8, inst_size:i32@12}` = **16 byte** split-word (như RegAllocResult). `push_fixup(&e.fixups, Fixup(offset: buf.len+1, label_id:.., inst_size:5))` — **literal `Fixup(...)` mới tạo truyền BY-VALUE làm arg** → native backend marshal qua str-like RAX:RDX → `f` rác. (Lưu ý: struct 16-byte **từ MEMORY** truyền by-value thì ĐÚNG — chỉ **literal mới tạo / giá trị trong thanh ghi** mới hỏng.)

### Repro CÔ LẬP (`bin/t_ret16c.ax`) — phân biệt 3 trường hợp
- `bin/t_ret16.ax`: 16-byte return by-value → RÁC (BUG#18). 24-byte → đúng.
- `bin/t_ret16b.ax`: 16-byte từ ARRAY truyền by-value → **ĐÚNG** (không trigger).
- `bin/t_ret16c.ax`: 16-byte **literal** `Fx(...)` truyền vào `pushit(self, f)` → RÁC `O=615613725640 L=0 S=0`; pad 24-byte → đúng `O=100 L=7 S=5`.
- `bin/t_ret12.ax`: **12-byte** `{u32,i32,i32}` literal pass → **ĐÚNG**. ⇒ bug CHỈ ở **đúng 16-byte** (2 full register RAX:RDX) với word-2 chia sub-8-field.

### Fix + AUDIT TOÀN BỘ struct
Pad `Fixup` thêm `pub pad0: i64` → 24 byte; sửa `push_fixup` (@alloc/@memcpy 16→24) + 4 constructor (`pad0: 0 as i64`).
**Audit mọi struct trong bootstrap/stage1**: chỉ **2 struct** đúng-16-byte-split-word: `RegAllocResult{ptr,u32,i32}` (BUG#18, fixed) + `Fixup{i64,u32,i32}` (fixed). Các struct khác: ≤12 byte (an toàn — đã test), >16 byte (by-address, an toàn), hoặc 16-byte nhưng word-2 là 1 field đơn (`ConstVal{bool,u64}` an toàn như str). ⇒ lớp lỗi này ĐÃ ĐÓNG cho codebase hiện tại.

### Cảnh giác / fix gốc tương lai
- **Quy tắc:** mọi struct **đúng 16 byte** mà word thứ 2 (offset 8-15) chứa **>1 field** (vd `{i64,u32,u32}`, `{ptr,u32,i32}`, `{u32×4}`) → KHÔNG được trả/truyền-literal by-value. Pad >16 byte. Struct `{i64,i64}`/`{ptr,i64}`/`{bool,u64}` (word-2 đơn field) an toàn.
- Fix gốc (chưa làm, rủi ro ABI): backend phải marshal struct 16-byte by-value đúng cho layout có sub-8 field ở word-2 (cả return RAX:RDX lẫn arg). Hiện workaround per-struct padding.

---

## BUG #20 — native backend miscompile **OR-chain ≥3 toán hạng có shift** (`b0|(b1<<8)|(b2<<16)|(b3<<24)`) → read_u32_le/read_u64_le trả rác → self-link "Unresolved external symbol ''" ✅ ĐÃ FIX (workaround)

**Bối cảnh:** Sau khi fix #17/#18/#19, stage2 codegen TRỌN VẸN 750 func + qua self-link, nhưng linker báo **80× `Linker Error: Unresolved external symbol ''`** (tên RỖNG) → "Self-linking failed". stage1 (C backend) self-link CÙNG t_std thì OK.

### Định vị (không cần build stage2 2h — repro trực tiếp trên axiom_temp.obj)
- `main_air.ax:812` self-link ghi `axiom_temp.obj` rồi linker parse lại; **dòng remove BỊ COMMENT (L828)** nên file CÒN LẠI để soi.
- `objdump -t axiom_temp.obj` (tool độc lập) đọc symbol names ĐÚNG → **COFF serialize ĐÚNG**; bug ở **parse phía stage2 (native)**.
- linker.ax:1570 `target_name = obj.sym_names.data[r.sym_idx]` rỗng. parse_object (L295+): `strtab_size = data_len - strtab_off`; name dài (>8) đọc nếu `str_off < strtab_size`.
- **Repro `bin/t_coffparse.ax`** (parse chính axiom_temp.obj bằng stage1 `-self-link`): byte thô `buffer[8..15]` ĐÚNG (174,32,1,0,246,0,0,0 khớp od) NHƯNG `read_u32_le(buffer,8)`=65552 (đáng lẽ 73902), `read_u32_le(buffer,12)`=16 (đáng lẽ 246). → **read_u32_le miscompile**.
- **Repro `bin/t_u32le*.ax`** (bisect): 2-operand `b0|(b1<<8)` ĐÚNG; **3-operand `b0|(b1<<8)|(b2<<16)` SAI** (65552, cả -O0 LẪN -O1); **dạng accumulator `mut r; r=r|((d[o+k])<<s)` ĐÚNG (73902)**. Bit: b2<<16 đúng nhưng b0→0x10, b1→0 (b0/b1 bị clobber).

### Root cause
Native selector/regalloc dựng cây OR **≥3 toán hạng có shift** sai (clobber/spill các operand đầu) dưới áp lực thanh ghi. KHÔNG phải u32 (i64 cũng sai), KHÔNG phải optimizer (-O0 cũng sai). read_u32_le/read_u64_le của linker dùng đúng pattern này → MỌI field COFF parse ra rác → tên symbol rỗng + offset sai.

### Fix (linker.ax) — workaround
Viết lại `read_u32_le`, `read_u64_le` dạng **single-accumulator fold** (`mut r := d[off]; r = r | ((d[off+k] as uX) << s)` lặp) — áp lực thanh ghi thấp, compile đúng (đã verify `four_acc`). `read_u16_le` (2-operand) GIỮ NGUYÊN (an toàn). Các `(x<<32)|y` khác (cgen/ssa_opt/elf64) đều 2-operand → an toàn.

### Cảnh giác / fix gốc tương lai
- **Mọi biểu thức ≥3 toán hạng kết hợp (OR/cộng) có shift, dưới áp lực thanh ghi** có thể sai native → dùng accumulator/tách `mut`. Repro bộ `bin/t_u32le*.ax`.
- Fix gốc (chưa làm): selector/regalloc xử lý đúng cây nhị phân OR rộng (giữ live đủ operand, không clobber). Hiện workaround per-site.

---

## BUG #21 — native `lower_struct_constructor_call`/`lower_struct_lit` gán field POSITIONAL (bỏ qua TÊN) → constructor thiếu field/sai thứ tự bị gán lệch ✅ FIX (targeted; root = RFC)

**Bối cảnh:** Sau #20, stage2 link t_std đọc đúng tên symbol nhưng **181 unresolved với tên THẬT** (`ax_File_close`, `ax__AX_std_Result__..._unwrap__...`, toàn method/generic). objdump xác nhận các symbol này ĐỊNH NGHĨA ĐÚNG (sec 1) trong obj; stage1 (C backend) link CÙNG t_std SẠCH (0 unresolved).

### Định vị
- Symbol defined trong obj nhưng stage2 linker không match → `func_names` (build từ `if sym.defined`) thiếu chúng → `sym.defined` bị false.
- **Repro `bin/t_omitfield.ax`**: struct `LSym{name:str, section:i64, offset:u64, size:u64, defined:bool}` construct **bỏ field `size`** (named): `LSym(name:.., section:.., offset:.., defined:true)` → `S=1 O=99 Z=1 D=N` (SAI: `defined:true` lọt vào field `size`→Z=1, `defined`→0=N). Thêm đủ `size:0` đúng thứ tự → `S=1 O=99 Z=0 D=Y` (ĐÚNG).
- Root: `air_builder.ax` `lower_struct_constructor_call` (L1530-1556) & `lower_struct_lit` (L1453-1467) gán mỗi arg vào `field_idx` TĂNG DẦN, **không dùng tên field** (`cn.payload`). Thuần positional. Constructor `LinkerSymbol(...)` (linker.ax) bỏ `size` (field thứ 4/5) → `defined` lọt vào `size`, field `defined` để 0 → mọi symbol parse ra `defined=false` → không vào func_names → unresolved.
- stage1 OK vì **Go stage0 lower struct theo TÊN** (đúng); chỉ native (self-hosted air_builder) positional → chỉ lộ ở stage2.

### Fix (targeted, linker.ax)
Thêm `size: 0 as u64` (đúng thứ tự khai báo) vào CẢ 2 `LinkerSymbol(...)` constructor → đủ 5 field theo thứ tự → positional == đúng. Verify bằng `t_omitfield2.ax`.

### ⚠️ Root fix (RFC, CHƯA làm — rủi ro chạm mọi construction)
`lower_struct_constructor_call`/`lower_struct_lit` phải tra **field index theo TÊN** (`cn.payload` = field name id → tìm trong `struct_info.fields` như typecheck.ax L2058-2068) thay vì counter. Testable nhanh (rebuild stage1 25s + t_omitfield). Cho đến khi làm: **MỌI struct construction PHẢI liệt kê ĐỦ field theo ĐÚNG thứ tự khai báo** — bỏ field hoặc đảo thứ tự sẽ gán lệch ở native. Audit khi build SHA (compile full compiler).

---

## BUG #22 — native backend lower `~` (bitwise NOT) thành LOGICAL not (`x==0`) → `~N`=0 → `& ~(align-1)`=0 → SizeOfImage=0 → PE INVALID ✅ FIX (CHƯA build được: axc.exe bị WDAC chặn)

**Bối cảnh:** Sau #21, stage2 compile+link t_std **0 unresolved** + ra t_std_x.exe (57856B) — nhưng exe **KHÔNG phải PE hợp lệ** ("not a valid application for this OS platform").

### Định vị (so byte PE header stage1-valid vs stage2-invalid)
- DOS header giống. PE optional header: stage2 **SizeOfImage=0x00000000** (stage1=0x13000) + SizeOfInitializedData=0 → Windows từ chối.
- size_of_image (linker.ax:609) = `(idata_rva + idata_raw_size + 0xFFF) & 0xFFFFF000`. idata_rva (L929) = `(...) & ~4095`.
- **Repro `bin/t_not.ax`**: `~4095`=**0** (đáng lẽ -4096); `v & ~4095`=0; `(0x1000+0+4095) & ~4095`=0 (đáng lẽ 4096). → `~` trả 0. (3-operand add `(a+b+c)&mask` ĐÚNG — `t_add3.ax` A=B=57344 — nên KHÔNG phải họ #20.)

### Root cause (x86_selector.ax:863)
`OP_NOT` (air_builder map CẢ logical `not`/`!` LẪN bitwise `~` → OP_NOT, phân biệt bằng type) chỉ được native lower **MỘT dạng = logical** (`CMP src,0; SETcc E; MOVZX` ⇒ `src==0`). cgen (C backend) phân biệt: `type_id==11(bool)→ !`, else `~`. Native bỏ nhánh bitwise → `~<int>` = `(int==0)` = 0. ⇒ mọi `& ~(align-1)` (struct layout air_builder/selector; PE section RVA/size linker) cho 0/sai.

### Fix (x86_selector.ax) — đã áp dụng vào source
Thêm điều kiện như cgen: `if inst.type_id == 11 as u32:` giữ logical (CMP/SETcc/MOVZX); `else:` emit bitwise = `MACH_MOV dest,src1` + `MACH_NOT dest`. (MACH_NOT=9 đã có; is_two_operand_read đã gồm.)

### ⚠️ CHẶN: không rebuild được — `bin/axc.exe` (Go stage0 C-backend) bị **Application Control/WDAC policy chặn** (2026-06-22). Các exe AXIOM-native chạy bình thường; chỉ axc.exe bị chặn → `rebuild_stage1.ps1` fail ("blocked by Device Guard policy"). CẦN: org whitelist axc.exe; HOẶC rebuild stage0 từ Go source. Verify fix #22 bằng t_not sau khi rebuild được (kỳ vọng N=-4096 A=4096 B=4096).

---

## BUG #23 — native backend ZERO-extend giá trị SIGNED nhỏ hơn 64-bit khi LOAD từ memory (đáng lẽ SIGN-extend) → `coff_sym_map[k] == -1 as i32` luôn FALSE → KHÔNG tạo extern symbol → binary native ra **0 imports** (extern call → `*ABS*`) ✅ FIX (verified)

**Bối cảnh:** Sau #22, stage2 compile t_std ra PE hợp lệ nhưng binary do **stage2 (native)** tạo có **0 imports**: mọi lời gọi extern (putchar/fflush/malloc…) reloc về `*ABS*` thay vì external symbol sec-0. stage1 (C-backend) luôn đúng (putchar→sym_idx 285, fflush→263, tạo external sec-0) → bug CHỈ ở logic native codegen, chỉ hiện khi stage2 chạy.

### Định vị (repro nhanh, KHÔNG cần build stage2 2h)
- `x86_coff.ax` reloc loop: `coff_sym_map` được `@alloc + @memset(…, 0xff, …)` (mỗi i32 = -1 = "chưa có"). Khi tra `target_sym_idx = coff_sym_map[target_id]`; `if target_sym_idx == -1 as i32:` thì mới `push_coff_symbol(section_num: 0)` (external chưa định nghĩa → import).
- **Repro `bin/t_memset.ax`** (build `-self-link -O1` = backend native): `@memset(p,0xff,n*4)` rồi đọc `p[0] as i64` → ra **4294967295** (đáng lẽ -1); `p[0] == -1 as i32` → **false**. ⇒ memset ĐÚNG (byte = 0xFFFFFFFF) nhưng LOAD i32 từ memory **zero-extend** thay vì sign-extend.
- **Repro `bin/t_sext.ax`** cô lập: literal i32 -1→i64 (`C=-1`) OK; i32 từ phép tính (`D=-1`) OK; **i32 LOAD từ memory `M=4294967295` SAI**; so sánh `c==-1`(`E=1`) OK nhưng `p[0]==-1`(`P=0`) SAI. ⇒ bug đúng ở đường LOAD-từ-memory của type signed.
- **Repro `bin/t_field.ax`**: field `i32=-1`→`A=-1` (sau fix), field `u32=0xFFFFFFFF`→`B=4294967295` (PHẢI giữ zero-extend), field `i16=-1`→`C=-1`, `s.a==-1`→`Q=1`.

### Root cause (x86_selector.ax — 4 đường load)
`mov r32,[mem]` của x86 zero-extend 32 bit cao. Backend sau khi MACH_LOAD (size 1/2/4) **luôn** zero-extend: OP_INDEX/OP_DEREF dùng `AND mask` (255/65535/0xFFFFFFFF); OP_LOAD/OP_GET_FIELD không mask (dựa luôn vào zero-extend của CPU). Với type SIGNED (i8/i16/i32) giá trị ÂM mất dấu. Latent vì i32 hầu hết KHÔNG âm — chỉ lộ ở sentinel -1 trong `coff_sym_map`. (cgen C backend đúng vì C tự sign-extend `(int64_t)i32`.) Cast i32→i64 (OP_CAST) chỉ MOV 64-bit nên DỰA VÀO việc giá trị đã sign-extend sẵn trong reg ⇒ phải sửa ở LOAD.

### Fix (x86_selector.ax) — đã áp dụng + verified
Thêm helper `emit_load_extend(sel, dest, size, type_id, out_insts)`: nếu `type_id ∈ {TYPE_I8,I16,I32}` → SIGN-extend bằng `MACH_SHL dest, (64-bits); MACH_SAR dest, (64-bits)` (shift imm: size4→32, size2→48, size1→56; MACH_SHL/SAR hỗ trợ OPND_IMM trong x86_emitter.ax:278-289); ngược lại (u8/u16/u32/bool/char) giữ `MOV_IMM mask + AND` zero-extend như cũ. Gọi ở CẢ 4 site: OP_LOAD, OP_INDEX, OP_GET_FIELD (nhánh scalar else), OP_DEREF — cho size 1/2/4. (i64/u64/isize/usize = 8 byte: không đụng.)
Verified (stage1 rebuilt 1823124B, repro qua backend native): t_sext `C=-1 D=-1 M=-1 E=1 P=1`; t_memset `A=-1 B=-1 F=1`; t_field `A=-1 B=4294967295 C=-1 Q=1`; regression t_not `N=-4096 A=4096 B=4096`, t_dec `R=4294967274 H=4294967296 D=-22` (u32 vẫn zero-extend đúng).

### ⚠️ Ghi chú
- Fix này đổi codegen của MỌI load scalar nhỏ hơn 8 byte trong toàn compiler ⇒ phải verify bằng full stage2→stage3 SHA (i32 field âm/sentinel ở data structure khác có thể từng latent).
- Trước build SHA cuối: GỠ diagnostic `RZ sym_idx=…` (x86_coff.ax reloc loop) + mọi marker AXCG/[codegen].

---

## BUG #24 — stage2 đọc SAI `inst.dst.phys` (nested MachOperand field) → mọi prologue `mov %rsp,%rbp` thành `mov %rsp,%rbx`, regalloc loạn → binary do stage2 sinh ra SEGFAULT + stage2 crash typecheck source lớn 🔍 ĐANG ĐIỀU TRA

**Bối cảnh:** Sau #23, stage2 BUILD THÀNH CÔNG (1769984B) + CHẠY ĐƯỢC (lần đầu — PE hợp lệ, có imports). Nhưng: (a) stage3 fail — stage2 crash trong `checker.run_type_checker()` (log dừng ở "Finished Resolving", trước "Finished Typechecking", hard crash không error); (b) binary nhỏ do stage2 sinh (t_field_s2.exe) **SEGFAULT**.

### Định vị (so disasm obj do stage1 vs stage2 sinh — KHÔNG cần build 2h)
- `cp axiom_temp.obj` sau khi build t_field bằng stage1 (obj_s1) và stage2 (obj_s2); `objdump -d` so sánh `main`.
- **100% SYSTEMATIC**: stage2 sinh `mov %rsp,%rbx` (48 89 **e3**) ở CẢ 278 hàm; stage1 sinh `mov %rsp,%rbp` (48 89 **e5**) ở 148 hàm. Toàn bộ register allocation của stage2 loạn (mọi dst dồn về %rbx, có `neg` thừa, giá trị sai).
- **Tách biệt nguyên nhân:** `push %rbp` (0x55) stage2 sinh ĐÚNG ⇒ const REG_RBP=5 OK. `mov_imm` dst (VREG) ĐÚNG. CHỈ `mov_rr` với dst = OPND_PHYS REG_RBP đọc `inst.dst.phys` ra **3 (RBX)** thay vì 5 (RBP). PUSH đọc `inst.src1.phys` → ĐÚNG (5). ⇒ Lỗi ở việc đọc field `dst` (MachOperand lồng, offset trong MachInst) — `inst.dst.phys` sai, `inst.src1.phys` đúng.
- MachOperand chứa i64 → align 8, size 24. MachInst: op(u16)+cc+padding=4, +pad → dst@8, src1@32, src2@56. emitter_resolve_reg nhận `op: MachOperand` BY VALUE (24B).

### Quan trọng: KHÔNG repro được bằng hàm nhỏ
`bin/t_nestfield.ax` (mô phỏng Opnd có i64 lồng trong Inst, đọc `inst.dst.phys` + `read_phys(inst.dst)` by-value) build bằng stage1 → ĐÚNG `D=5 S=4 R=5 T=4`. ⇒ native backend của stage1 xử lý nested-field/by-value ĐÚNG ở hàm nhỏ. Bug CHỈ hiện ở **hàm LỚN** (emit_mach_inst/emitter_resolve_reg — register pressure cao, SPILLING). Khớp ghi chú cũ "bug parser hàm LỚN (spilling)". ⇒ stage1 mis-compile truy cập `inst.dst` (spill/reload sai offset) trong hàm codegen lớn → stage2 đọc dst.phys sai.

### HƯỚNG ĐIỀU TRA TIẾP
1. Cô lập: tạo repro hàm LỚN (nhiều biến/spill) đọc `.dst.phys` từ struct lồng, build bằng stage1, xem có ra sai như stage2 không.
2. Hoặc dump spill code của emitter_resolve_reg / emit_mach_inst (insert_spill_code có debug print bị comment ~L787) để xem reload `inst.dst` sai offset.
3. Nghi can: spill/reload của aggregate-by-value 24B, hoặc field_offset của MachOperand lồng khi vreg bị spill. So x86_regalloc.ax insert_spill_code + cách load 16/24-byte spilled (regalloc_is_16byte chỉ check 16, KHÔNG check 24?).

---

## BUG #24 — register allocator gán biến SỐNG vào RCX, bị `SHL/SAR/SHR` (dùng CL) clobber → x86_encode_modrm_rr sinh modrm/REX sai → mọi prologue `mov %rsp,%rbp`→`mov %rsp,%rbx` + `mov rax,r12/r13` mất REX.B → stage2 regalloc loạn, binary segfault 🔧 FIX v2 (reserve RCX) — đang verify

**Bối cảnh:** Sau #23, stage2 BUILD OK + CHẠY (lần đầu) nhưng stage3 crash trong run_type_checker (sau "Finished Resolving") + binary nhỏ do stage2 sinh SEGFAULT. Lỗi CHỈ ở native backend (stage1 C-built sinh đúng).

### Định vị (vàng — diff disasm, KHÔNG cần build lại)
- `cp axiom_temp.obj` sau build t_field bằng stage1 (obj_s1c, ĐÚNG) và stage2 (obj_s2c, LỖI); `objdump -d` so `main`.
- Triệu chứng: stage2 sinh `mov %rsp,%rbx` (48 89 **e3**) ở 100% hàm thay vì `mov %rsp,%rbp` (48 89 **e5**).
- **Build full obj bằng stage1 (`-self-link`, có symbol) → disasm `ax_x86_encode_modrm_rr`** thấy chính xác: biến `rm_field`(=5=RBP) bị gán RCX; phép `out_modrm = (3<<6)|((reg&7)<<3)|(rm&7)` lower `<<6`/`<<3` thành `MOV RCX,count; SHL dst,%cl` → RCX bị ghi đè bằng shift count (3) → `(rm_field&7)` đọc 3 → modrm rm=3=RBX. CL form của shift clobber RCX mà allocator KHÔNG model.
- Tương tự `mov rax,r12` (cần REX.B vì r12≥8): giá trị `rm` cho `reg_needs_rex(rm)` bị clobber → REX.B mất → `49 89 c4`→`48 89 c4` = `mov rax,rsp` → phá stack (gdb: `mov $const,%rsp`+`mov $const,%rbp`, rbp=4 rsp nhỏ → crash trong allocator path của t_field).

### Fix v1 (span-forbid_rcx) — ĐÃ BỎ vì regression
Thêm logic "vreg span qua SHL/SAR/SHR → forbid_rcx" (sao chép pattern IDIV span forbid_rax_rdx). Sửa được 150/153 prologue NHƯNG thay đổi allocation làm hỏng `rm_need` (REX.B) → r12/r13→rsp/rbp → vẫn crash. Cách forbid theo span quá mong manh.

### Fix v2 (reserve RCX khỏi pool) — ĐÃ BỎ vì regression khác
Loại RCX khỏi get_allocatable_gprs (12→11 reg). Prologue đúng (main `e5`) NHƯNG tăng register pressure → **lộ bug spill** trong frontend của stage2 → stage2 build t_field chỉ codegen 2 hàm (drop hết lazy-module/stdlib functions) → segfault. ⇒ tăng spill là nguy hiểm (có bug spill latent riêng).

### Fix v3 (immediate-form shift cho hằng số) — ❌ KHÔNG ĐỦ (CHỈ fix được 1 phần)
Sửa **selector** (x86_selector.ax OP_SHL/OP_SHR): helper `const_shift_amount(fn_ptr, vreg)` quét defining-inst của count vreg; nếu là `OP_ICONST` (0x0201, value=src1|src2<<32) với value 0..63 → emit `MACH_SHL dst, OPND_IMM`. Verified `1<<32`→`shl $0x20` (vì count i64 không có cast). NHƯNG **stage2-v3 VẪN sinh `mov %rsp,%rbx` (e3) 277/277, t_field crash** ⇒ v3 chưa fix. Lý do ở v4.

### Fix v4 (const_shift_amount NHÌN XUYÊN OP_CAST/COPY/MOVE) — ✅ ROOT CAUSE THẬT
**Tại sao v3 trượt:** `x86_encode_modrm_rr` viết `(3 as u8 << 6 as u8)` và `((reg_field & 7 as u8) << 3 as u8)`. Literal `6 as u8`/`3 as u8` lower thành `ICONST → OP_CAST(u8)`, nên `inst.src2` của OP_SHL trỏ tới **OP_CAST**, KHÔNG phải OP_ICONST. v3 chỉ khớp ICONST trực tiếp → trả -1 → rơi vào CL-form `MOV RCX,count; SHL dst,%cl` → clobber RCX (giữ rm_field) → modrm rm 5→3 → mọi `mov %rsp,%rbp`→`%rbx` ở stage2.
**Bằng chứng định vị (vàng, KHÔNG cần build 2h):** repro `bin/t_modrm.ax` = `enc_modrm(reg,rm) = (3<<6)|((reg&7)<<3)|(rm&7)`. Build bằng stage1, `cp axiom_temp.obj`, `objdump -d ax_enc_modrm`: v3 sinh `mov $0x6,%rcx; shl %cl,%rbx` + `shl %cl,%rax` (CL-form ⚠️). Sau v4: `shl $0x6,%rbx` + `shl $0x3,%rax` (immediate ✅).
**Fix:** const_shift_amount (x86_selector.ax ~L654) đổi từ "khớp 1 lần ICONST" sang **vòng chase** (depth≤16): nếu defining-inst là OP_ICONST(0x0201) value∈[0,64) → trả value; nếu là OP_CAST(0x0223)/OP_COPY(0x0106)/OP_MOVE(0x0107) → `cur = inst.src1`, lặp; else -1. AN TOÀN: chỉ chấp nhận khi ICONST nguồn đã ∈[0,64) nên cast thu hẹp không bao giờ làm sai. Verified: t_modrm `M=229 N=227` + immediate-form; 5/5 repro (t_field/t_sext/t_memset/t_not/t_bignest) PASS không regression. Rebuild stage1 1821654B. Đang build stage2-v4 verify (prologue e5 + stage3 SHA).
**Repro spill nested-struct (loại trừ giả thuyết sai):** `bin/t_bignest.ax` (Op 24B lồng trong Inst 80B, pass by-value, hàm lớn). stage1 compile ĐÚNG: `inst` spill `-0xa0(%rbp)`, reload `mov -0xa0(%rbp),%r10` rồi `lea 0x8(%r10)`/`lea 0x20(%r10)` (offset dst@8, src1@32 ĐÚNG). ⇒ bug KHÔNG phải spill aggregate như từng nghi; là RCX clobber bởi cast-wrapped const shift.

### ⚠️ Lưu ý lớp lỗi "implicit fixed-register clobber"
Các lệnh dùng register cố định ngầm phải được allocator model: IDIV/CQO→RAX/RDX (đã có span forbid), SHL/SAR/SHR→RCX, CALL→caller-saved (spans_call). Với SHL/SAR/SHR: cách hiện tại (v4) là **né CL-form cho shift HẰNG** (const_shift_amount nhìn xuyên cast → immediate `shl $imm`). ⚠️ Shift VARIABLE (count runtime) VẪN dùng CL-form và allocator VẪN KHÔNG model RCX-clobber → nếu compiler có shift biến mà allocator đặt giá trị sống vào RCX qua đó thì sẽ lỗi tương tự. Hiện compiler hầu hết shift hằng nên an toàn; nếu sau này lỗi lại → cần model RCX-clobber cho CL-form (forbid_rcx theo span, cẩn thận regression REX.B như v1).

---

## BUG #25 — SSA optimizer (-O1) const-fold biến ADDRESS-TAKEN → bỏ `if flag` sau `f(&flag)` → mất REX.B ở mov_rr → stage2 segfault (rbp=4) ✅ FIXED
**Triệu chứng:** sau khi v4 sửa modrm shift, stage2-v4 BUILD ok nhưng binary nó sinh vẫn segfault (gdb: `rbp=0x4`, `rsp` tí xíu, crash tại `call`). Prologue đã đúng (e5). 
**Định vị (vàng, KHÔNG cần build 2h):** crash ở `ax_std_mem_alloc_get_slab` (allocator, gọi bởi `@alloc`). Diff disasm hàm này stage1(đúng) vs stage2-v4(lỗi): stage1 `mov %rax,%r12; mov %rax,%r13; mov %r12,%r8; mov %r13,%r9`; stage2-v4 `mov %rax,%rsp; mov %rax,%rbp; mov %rsp,%rax; mov %rbp,%rcx`. → **mất REX.R/REX.B** (giữ W): `mov %r12,%r8`(4d 89 e0)→`mov %rsp,%rax`(48 89 e0), `mov %r13,%r9`→`mov %rbp,%rcx`. `mov %rax,%rsp` phá stack → crash.
**Truy nguồn:** allocator gán r12/r13 ĐÚNG (get_allocatable_gprs, reg_hw_reg, reg_needs_rex, emitter_resolve_reg trong stage2-v4 đều disasm ĐÚNG). Bug ở `x86_encode_mov_rr`: `mut need_rex:=false; x86_encode_modrm_rr(...,&need_rex); if need_rex: push(rex|W) else push(0x48)`. Disasm stage2-v4: sau call modrm_rr là **`jmp` vô điều kiện thẳng nhánh else** — `if need_rex` BỊ XÓA. Optimizer const-fold `need_rex=false` (giá trị init), KHÔNG biết `&need_rex` escape vào call (callee ghi qua con trỏ).
**Repro vàng:** `bin/t_alias.ax` — `mut flag:=false; set_outs(...,&flag); if flag: return 1; return 0`. stage1 build: **-O0 → A=1 (đúng), -O1 → A=0 (SAI)**. `dump-air -O1`: `%9=iconst(false); %13=mkref %9; call; jump`(branch bị fold). Ở -O0 đúng vì regalloc spill vreg address-taken ra slot (x86_selector graph_coloring `elif address_taken[vreg]: spilled`) → reload đúng; -O1 fold TRƯỚC codegen nên bỏ qua reload.
**FIX (ssa_opt.ax):** helper `mark_addr_taken_regs(f, max_reg) -> ptr[bool]` quét OP_MAKE_REF(0x0108), đánh dấu `src1`. Trong `fold_func` + `alg_simp_func`: khi ghi `vals[dest]=known` từ ICONST (và binary/unary trong fold_func) → thêm điều kiện `and not addr_taken[dest]`. ⇒ register address-taken KHÔNG bao giờ bị coi là const → branch-fold/const-fold bỏ qua. An toàn (chỉ giảm tối ưu, đã có backend spill). Verified t_alias -O1 → A=1, branch `branch %9` giữ nguyên; 7/7 repro pass.
**⚠️ Lớp lỗi tổng quát:** bất kỳ pass tối ưu nào dùng giá trị const của biến có `&` (OP_MAKE_REF) hoặc bị STORE qua con trỏ đều phải coi là escaped/unknown. Đây là alias-analysis còn thiếu; nếu thêm pass forward giá trị (GVN/store-forwarding) phải tôn trọng addr_taken.

---

## BUG #26 — `strip_package_prefixes` TỰ XÓA literal prefix của chính nó → stage2 KHÔNG strip `std.*` → call thành field-expr unresolved → emitter self-call → STACK_OVERFLOW ✅ FIXED (2026-06-25)

### ROOT CAUSE THẬT (không phải codegen!)
Driver self-link (`main_air.ax`) gọi `strip_package_prefixes(src)` để xóa prefix module (`std.string.len(s)` → `len(s)`) ở MỨC TEXT trước khi lex/parse, để qualified call resolve về hàm top-level đã concat. Hàm này dùng `std.string.replace(s, "std.string.", "")` v.v. với 9 literal prefix.
**Vòng tự tham chiếu:** khi stage1 build stage2, driver stage1 chạy `strip_package_prefixes` TRÊN CHÍNH source compiler (`tmp_concatenated_air.ax`, áp ở main_air.ax:378 cho user src). Nó xóa luôn các LITERAL `"std.string."`, `"std.os."`, ... NẰM TRONG thân `strip_package_prefixes`. Kết quả stage2 thấy `r6 = replace(r5, "", "")` (old="") → no-op (replace return s khi old.len==0). → stage2's strip KHÔNG làm gì → mọi `std.*` qualified call sống sót → parser tạo FIELD_EXPR → resolver KHÔNG resolve (flags 0) → air_builder else-branch (1199): callee_reg=lower_expr(field), type_id=return-type → selector OP_CALL (1128) `if src1==0` FALSE (src1=callee_reg≠0) → sym_imm GIỮ 0 → MACH_CALL imm=0 → emitter self-call.
**Bằng chứng:** mô phỏng strip trên tmp_concatenated_air.ax (perl s///): tất cả 8 dòng `r1..r8 = replace(rX, "", "")` (rỗng). stage1 (chưa bị strip vì PowerShell concat chỉ strip import) → callee là IDENT 'len' resolve type_id=237=ax_len (đúng). stage2-diag2 `[B26 main 30 0. 0 4 4 182]` (FIELD_EXPR, type_id=0/4, self-call) vs stage1 `[B26 main 42 0. 0 0 237 237]` (IDENT, ax_len). `std.string.replace/concat/len` ĐỀU OK natively (t_replace/t_concat) — KHÔNG phải codegen bug.

### FIX (main_air.ax strip_package_prefixes)
Build pattern lúc RUNTIME bằng concat để literal prefix KHÔNG xuất hiện liền mạch trong source: `let p = "std."`; `replace(r, std.string.concat(p, "string."), "")`. Mảnh `"std."`, `"string."`, `":"` không khớp pattern nào nên SỐNG SÓT qua self-strip; concat dựng lại pattern thật ("std.string.") lúc chạy. (commit 6a31427)
**VERIFIED ✅:** rebuild stage2 (stage1+fix) → compile t_strlen: **reloc 518 = 518 (stage1), KHÔNG còn self-call, strip ĐẦY ĐỦ**. BUG#26 GIẢI QUYẾT TRIỆT ĐỂ.
**Lớp lỗi:** self-referential source transform — bất kỳ pass nào biến đổi text dựa trên literal mà CŨNG chạy trên chính nó sẽ tự phá. Tương tự rủi ro cho mọi text-rewrite trong self-hosting.

---

## BUG #27 — byte-store encoder BỎ REX cho src byte-reg 4-7 (sil/dil) → `mov %sil,m` thành `mov %dh,m` (lưu rác) ✅ FIXED (2026-06-25)

### ROOT CAUSE (REX-loss, BUG#25 family — 8-bit store)
`x86_encode_mov_store_sized` (x86_encoding.ax:445): `needs_rex_prefix = w_bit or reg_needs_rex(src) or reg_needs_rex(base)`. `reg_needs_rex(r) = (r&15) >= 8`. Với store 8-bit (size==1) mà src là reg 4-7 (rsp/rbp/rsi/rdi → byte spl/bpl/sil/dil), CẦN REX prefix bắt buộc; thiếu REX thì modrm reg field 4-7 giải mã thành high-byte LEGACY ah/ch/dh/bh. ⇒ `mov %sil,m` (40 88..) bị emit thành `mov %dh,m` (88..) → lưu RÁC (dh = byte cao của con trỏ heap).
**Cách lộ:** `emit_param_prologue` build `MachOperand(kind:OPND_PHYS, phys: phys, ...)`; `phys`(u8) sống qua call `next_vreg` nên regalloc đặt vào callee-saved rsi(sil); store `q.phys = phys` → `88 72 01 mov %dh,0x1(%rdx)` (thiếu REX) → src1.phys = rác (13/14 = byte cao con trỏ alloc). ⇒ stage2's emit_param_prologue set phys param-reg = rác → param load `mov %r13/%r14,...` thay rcx/rdx/r8/r9 → đọc tham số rác → crash. (Chỉ lộ khi regalloc đặt u8 vào rsi/rdi/rbp/rsp — nên hiếm/khó thấy.)
**Chứng minh:** obj_stage2_diag.obj ax_emit_param_prologue @1414cc: `88 72 01 mov %dh,0x1(%rdx)` (đáng lẽ `40 88 72 01 mov %sil`). t_param5 (5-param fn) repro: stage1→A38, stage2→crash. arg_reg/local-across-call compile đúng natively (loại trừ).

### FIX (x86_encoding.ax:445)
Trong x86_encode_mov_store_sized: nếu `size==1` và `reg_hw_reg(src) ∈ [4,7]` → `needs_rex_prefix = true` (encode_rex trả 0x40-based REX sẵn). VERIFIED: stage1+fix compile t_bytestore → `40 88 71 01 mov %sil,0x1(%rcx)` (có REX), chạy in 006 007 013 ✓; t_param5/t_strlen không regress. Đang build stage2→stage3 verify + SHA.
**Lớp lỗi:** REX-loss cho 8-bit access reg 4-7 — kiểm tra các encoder byte khác (mov_rr 8-bit, setcc, movzx src reg) có cùng thiếu sót không.

---

## BUG #28 — stage2 PE stack reserve 1MB quá nhỏ → stack overflow khi typecheck compiler đầy đủ ✅ FIXED (2026-06-25)
Sau fix BUG#27, stage2 compile+chạy chương trình nhỏ ĐÚNG (t_param5→A38, t_strlen→5) NHƯNG build stage3 (typecheck compiler 1.2MB) crash exit 127, log dừng ở "Finished Resolving" (vào typecheck). gdb: rsp=0x63400, rbp=0x6a400 (RẤT thấp), rip rác, bt gãy = STACK OVERFLOW (đệ quy sâu infer_node trên input lớn). So PE: stage1 SizeOfStackReserve=**0x200000 (2MB)** (C-linked) vs stage2=**0x100000 (1MB)** (self-linker). Self-linker đặt 1MB cứng → đệ quy typecheck compiler cần >1MB → overflow (stage1 2MB sống). KHÔNG phải đệ quy vô hạn (stage1 2MB typecheck xong). FIX: linker.ax:615 `SizeOfStackReserve = 0x1000000` (16MB, headroom rộng; commit vẫn 0x1000, Windows auto-grow). VERIFIED: stage1+fix self-link → PE reserve 0x1000000. Đang build stage2→stage3 verify+SHA.

### (lịch sử định vị BUG#27 — đã sửa nhãn)
## BUG #27 (cũ nhãn) — ax_actor_send param-load miscompile
**Lộ ra SAU khi fix BUG#26** (self-call hết): stage2 compile t_strlen → obj reloc ĐÚNG (518=518) NHƯNG binary crash exit 139 (rip rác 0x..40a4, rax=rcx=0, bt rác) + in byte rác; stage2 build stage3 crash ở typecheck (log dừng "Finished Resolving", exit 127).
**Định vị (THẬT — không phải prologue):** prologue ax_actor_send ĐÚNG cả 2 (`382c 48 89 e5`). Cái `48 89 e3` mình grep được là 1 lệnh PARAM-MOVE bị hỏng, không phải prologue. Diff vùng load tham số (sau push callee-saved + sub rsp):
- **stage1 (ĐÚNG):** `mov %rcx,%r11`/`mov %rdx,%r11`/`mov %r8,%rbx`/`mov %r9,%rsi` — load param từ ĐÚNG win64 param-regs (rcx,rdx,r8,r9), frame `sub $0x58`, SPILL param ra stack (-0x68/-0x50(rbp)).
- **stage2 (SAI):** `mov %rbx,%rax`/`mov %rbx,%rcx`/`mov %rsp,%rdx`/`mov %rsp,%rbx` — load từ **rbx/rsp (rác, KHÔNG phải param-reg)**, frame `sub $0x38`, KHÔNG spill. ⇒ ax_actor_send đọc tham số rác → crash.
**Bản chất:** stage2's backend miscompile PHẦN LOAD THAM SỐ của hàm nhiều param (sai source-reg + regalloc/spill khác hẳn) — KHÔNG phải REX-loss đơn giản, KHÔNG phải BUG#24 rm-clobber. stage1 native-miscompile 1 hàm backend (selector OP_CALL param-setup / regalloc move param-phys→vreg / abi_int_arg_reg) → stage2 codegen tham số sai. Nghi: hàm gán param physical reg vào vreg ở đầu hàm (selector ~param lowering) hoặc regalloc move-coalescing.
**REPRO NHANH (vàng):** `bin/axc_stage2.exe` ĐÃ build (1775616B) — compile chương trình nhỏ trong vài giây rồi `cp axiom_temp.obj` → objdump đếm `48 89 e3`. So với stage1 obj. Khả năng: viết hàm nhỏ giống ax_actor_send (có shift + nhiều biến sống) build bằng stage2 → tái hiện e3.
**HƯỚNG TIẾP:** (1) objdump ax_actor_send đầy đủ trong stage1 vs stage2 obj, tìm chỗ shift dùng CL-form (MOV RCX+SHL %cl) ở stage2 mà stage1 dùng immediate (SHL $imm) → xác nhận const_shift_amount native fail. (2) Disasm const_shift_amount trong obj_stage2 (axc_stage2 obj nếu lưu) tìm miscompile. (3) Cân nhắc fix gốc class: regalloc reserve/avoid RCX cho rm-field khi có shift, HOẶC spill-safe CL-form. Xem BUG#24 lịch sử (v2 reserve RCX bị revert do tăng spill).

### (lịch sử điều tra — sai hướng codegen)
**Triệu chứng:** sau khi fix BUG#24+#25, stage2-v5 BUILD ok, prologue đúng (148 e5/0 e3, không mất REX) NHƯNG binary nó sinh crash khi chạy: `EXITCODE=0xC00000FD` (STACK_OVERFLOW), không in gì. (Crash này lộ ra SAU khi BUG#25 — vốn crash sớm hơn ở get_slab — đã được sửa.)
**Định vị (gdb + diff disasm):** gdb `t_field_s2v5.exe`: crash tại `call` đầu trong `ax_ax_os_alloc_report_error`, rbp chain đệ quy hoàn hảo (mọi frame 0xC0 byte, cùng return addr). `x/120a` cho thấy report_error tự gọi chính nó vô hạn. Disasm `-dr obj_s2v5.obj` (obj stage2-v5 sinh khi build t_field):
- report_error: call `std.string.len(msg)` tại 0x126d ra `e8 50 ff ff ff` = **direct rel -0xb0 = đầu report_error (self)**, KHÔNG có relocation.
- So stage1 (obj_s1e): cùng chỗ là `e8 00000000` + **IMAGE_REL_AMD64_REL32 ax_len** (đúng).
**Cơ chế:** emitter `x86_emitter.ax` MACH_CALL: `if inst.src1.imm == 0: disp = -(buf.len+5)` (coi là self-call); else push_fixup theo symbol index. Selector `x86_selector.ax` OP_CALL (~L1128): `sym_imm = callee_sym_idx = inst.type_id` (cho direct call inst.src1==0), push MACH_CALL imm=sym_imm (L1275). ⇒ stage2-v5 có `inst.type_id == 0` cho các call này (symbol CHƯA resolve) → emitter self-call → đệ quy.
**Phạm vi (đếm reloc obj_s1e vs obj_s2v5):** stage1 527 reloc / stage2-v5 511 (thiếu 16). Symbol stage2-v5 KHÔNG resolve được (đều `std.string.*`): **ax_len×7, ax_slice×8, ax_starts_with×1, ax_concat×1, ax_str_slice×1**. Các call nội-module khác (ax_print_hex_freestanding...) vẫn reloc ĐÚNG.
**Bản chất:** bug FRONTEND/RESOLVER của stage2 (lazy-module) — resolve cross-module `std.string.*` ra symbol index 0. KHÔNG repro được bằng stage1 build (stage1 C-built resolve đúng); phải diff hàm resolver trong obj_stage2_v5.obj (có symbol) hoặc dump symbol table stage2 sinh. ⚠️ Emitter heuristic `imm==0 → self-call` quá nguy hiểm: biến symbol-unresolved thành đệ quy âm thầm thay vì lỗi link — cân nhắc đổi thành sentinel/asserts.
**HƯỚNG TIẾP:** (1) tìm trong resolver/symbol-table cách `std.string.*` được gán index khi gọi từ module khác (os/alloc gọi std.string.len); so logic stage1 vs cách nó compile trong stage2. (2) Hoặc check air_builder lower OP_CALL: callee sym_idx vào type_id — có thể u16 type_id tràn nếu symbol index của std.string.* > 65535? Kiểm tra symbol index thực của ax_len. (3) Repro vàng: cp axiom_temp.obj khi stage2-v5 build chương trình NHỎ có dùng `std.string.len` → objdump -dr xem self-call.

---

## BUG #29 — `std.string.replace` redirect → C-ABI `ax_str_replace` bị stage2 miscompile (3 str-arg by-ref setup) → stage3 crash, phá fixpoint stage2==stage3 ✅ FIXED (2026-06-25)

**Bối cảnh:** Sau fix BUG#27+#28, stage2 BUILD ok, chạy chương trình nhỏ ĐÚNG (t_param5→A38, t_strlen→5), và **build stage3 đầy đủ exit=0** (753 funcs, qua typecheck nhờ 16MB stack). NHƯNG `SHA(stage2) != SHA(stage3)` (1776128 vs 1815040, lệch 38912 byte). Build stage4 = stage3 compile S → **stage3 SEGFAULT (exit 139) ngay**. ⇒ stage2 miscompile stage3 (stage3 là binary hỏng).

**Định vị (fast repro qua bin/axc_stage2.exe + gdb):**
- gdb stage3: crash trong `ax_str_replace` (ax_runtime.dll), gọi từ `strip_package_prefixes` (replace đầu tiên), ngay sau "Stripping imports from result.ax...". Stack có nhiều int nhỏ (0xf,0x14,0x11...) = length args spill.
- Repro nhỏ `bin/t_strip.ax` (chained replace + concat): **stage1 build → chạy ĐÚNG (`a.b len exit print`); stage2 build → segfault.** ⇒ lỗi native codegen, KHÔNG phải logic.
- Bisect: `concat` đơn lẻ OK (tcA); `replace` literal → OOM `ax_alloc` (len rác); `replace` var/nested-arg → segfault (ptr rác). ⇒ **`replace` nhận args hỏng.**
- `myreplace` (thân value-ABI Y HỆT std/string.ax nhưng TÊN KHÁC, không bị redirect) build bằng stage2 → chạy ĐÚNG `a.b.c`. ⇒ khác biệt DUY NHẤT = cái redirect tên.
- `strings t_strip_s2.exe` import `ax_str_replace` (crash); `tmyrep_s2.exe` KHÔNG (chạy ok).

### ROOT CAUSE
[x86_coff.ax](../bootstrap/stage1/x86_coff.ax) `x86_resolve_call_target` redirect `std.string.replace` → symbol C-runtime `ax_str_replace` (giống concat CŨ). `ax_str_replace` dùng **C/Win64 ABI** (str args by-ref, 16-byte return qua hidden sret pointer) — KHÔNG tương thích AXIOM value ABI. Với **3 str args**, call lấp đủ 4 thanh ghi int (sret + &s + &old + &new); phần setup by-ref str-arg bị backend (do stage1 sinh) miscompile → stage2 phát call `ax_str_replace` truyền ptr/len rác → OOM (len sai)/segfault (ptr sai). Comment ngay trên đó đã ghi rõ concat KHÔNG redirect vì chính lý do ABI này — replace bị bỏ sót.

**Bản chất chính xác (lớp lỗi):** không phải số lượng str-arg, mà là **C-ABI call TRẢ VỀ str (struct 16-byte qua hidden sret pointer)**. stage1's backend miscompile phần SETUP sret return-pointer của C-ABI call → stage2 phát call với sret/ptr rác. `len`→i64 và `eq`→bool KHÔNG có sret nên ĐÚNG (vì vậy stage2 tự chạy ok); chỉ các call stage2 PHÁT RA vào stage3 bị hỏng. `replace` và `slice` đều trả str → đều hỏng.

⚠️ **Lần fix đầu chỉ bỏ replace, GIỮ slice (sai)** → stage2 chạy ok nhưng stage3 vẫn segfault, lần này trong `ax_str_slice` (gdb). Repro: `bin/tslice.ax` build bằng stage2(fixed-replace) → segfault; `myslice` value-ABI → ok. ⇒ slice CÙNG lớp lỗi sret. Vì sao tưởng slice ok: stage2 TỰ chạy slice (call do stage1 phát, đúng) nhưng call stage2 phát vào stage3 thì hỏng — y hệt replace.

### FIX (x86_coff.ax `x86_resolve_call_target`)
Bỏ redirect CẢ `std.string.replace` LẪN `std.string.slice` → rơi xuống hàm value-ABI in-program (std/string.ax:192 replace, :38 slice), Y HỆT concat. Giữ nguyên path `-30` trong `resolve_binary_sym_name` + alias `ax_block_size`→`ax_str_slice` (machinery riêng, không phải call std.string.slice).

**VERIFIED (stage1-level):** rebuild_stage1 + fix → stage1 compile t_strip (`a.b len exit print`) + tslice (`hello`/`world`) → ĐÚNG, **0 thunk `jmp ax_str_replace`, 0 `jmp ax_str_slice`** (cả hai = value-ABI). Đang chạy full verify stage2→stage3→stage4 xác nhận fixpoint SHA.

---

## BUG #30 — `copy` của struct VALUE 16-byte (by-address) bị lower thành block-copy INLINE → đọc con trỏ thành dữ liệu → str/struct 16B hỏng trong stage3 🔬 ROOT-CAUSED (chưa fix, cần thiết kế cẩn thận)

**Đây là ROOT CAUSE THẬT đứng sau BUG#29.** replace/slice/print_raw_ptr chỉ là triệu chứng vì `str` = struct 16-byte.

**Triệu chứng:** Sau khi fix BUG#29 (bỏ redirect replace+slice), stage2 build stage3 OK, nhưng stage3 chạy:
- printf in số `%d` ra RỖNG (literal in được, số nguyên ≠0 mất) — `print_i64_raw` hỏng.
- crash ở pha "building reloc table".

**Repro nhanh (vàng):**
- `bin/tpi.ax` (chép y `print_i64_raw`): stage1→`0/753/1213284`, stage2→`0` rồi 2 dòng rỗng.
- `bin/tprp.ax` (chép `print_raw_ptr`): stage1→`[ABC]`, stage2→`[]`.
- `bin/tf3.ax`: `struct S2{a,b:i64}; let x=S2(7,9); return x.a*10+x.b` → stage1=**79**, stage2=**80**.
- **`S2` 2×i64 = 16 byte → HỎNG; `S3` 3×i64 = 24 byte → ĐÚNG (123).** ⇒ lỗi đặc thù struct ĐÚNG 16 byte (= cùng size với `str`).

**Cơ chế (objdump `main` của tf3/tff_a, obj = `axiom_temp.obj` còn lại sau -self-link):**
`let x = S2(...)`: struct literal heap-alloc → kết quả là CON TRỎ (trong slot -0x18). Khi materialize `let x`:
- stage1 (đúng): KHÔNG copy, x = con trỏ heap, `x.a` = `mov (heapptr)` → 7.
- stage2 (sai): block-copy 16 byte từ `lea -0x18` (ĐỊA CHỈ của slot chứa con trỏ) → `(-0x18)` = con trỏ, không phải `*(-0x18)` = struct. ⇒ x[0] = con trỏ (rác), x.a = con trỏ.

**Phân tầng (dump-air tff_a):**
- `dump-air -O0`: stage1 == stage2, CẢ HAI có `%6: t22 = copy %1` rồi `getfld %6` (air_builder luôn phát copy).
- `dump-air -O1`: stage1 XÓA copy → `getfld %1` (đúng); stage2 GIỮ copy → `copy %1; getfld %6` (sai).
- ⇒ **(1) LATENT bug ở selector**: `select_inst` OP_COPY/OP_MOVE (x86_selector.ax:764) khi `regalloc_is_16byte(dest)` true → `emit_block_copy(LEA dest, LEA src1, 16)`. Đúng cho `str` (INLINE 16B) nhưng SAI cho struct by-address (src1 chứa con trỏ → phải MOV con trỏ / deep-copy, không block-copy slot). `regalloc_is_16byte` (x86_selector.ax:435) trả true cho MỌI size==16, gộp nhầm `str` (inline) với struct (by-address).
- ⇒ **(2)** stage1's optimizer (O1) ELIDE được copy thừa nên che lỗi (1); stage2's optimizer KHÔNG elide (vì optimizer của stage2 bị stage1 miscompile) → lộ lỗi (1).

**Bản chất SÂU HƠN (đã xác nhận bằng disasm O0):** đây là MISMATCH REPRESENTATION.
- `%1 = alloc` (struct literal) → `regalloc_is_16byte(alloc)` = FALSE → %1 là POINTER-repr (8-byte, vreg chứa địa chỉ heap). `getfld %1` (khi %1 ở thanh ghi) deref đúng.
- `%6 = copy %1` (type struct size16) → `regalloc_is_16byte` (case OP_COPY) trả TRUE chỉ vì size==16 → %6 bị coi là INLINE-16 (home = 16 byte dữ liệu, "giá trị" = &home). 
- ⇒ copy 8-byte-pointer (%1) vào ô 16-byte-inline (%6) → lệch repr. `getfld %6` đọc `home[disp]` (= con trỏ) thay vì deref. (`str` thì home ĐÚNG là 16-byte data nên ok.)
- stage1-O1 sống vì optimizer ELIDE copy → %6 ≡ %1 (pointer-repr) → getfld deref đúng. stage2 không elide → lộ mismatch.

### FIX (x86_selector.ax `regalloc_is_16byte`, case OP_COPY/OP_MOVE ~485)
`if size == 16 and not type_is_aggregate(table, type_id): return true` — chỉ `str` (PRIMITIVE size16) mới inline-16; aggregate by-address size16 rơi xuống `regalloc_is_16byte(src1)` ⇒ khớp repr của nguồn (alloc → pointer 8-byte). Khi đó copy = MOV con trỏ (alias, đúng ngữ nghĩa tham chiếu của struct — verified `let x=y; x.a=9` ⇒ y.a==9), getfld deref đúng. `type_is_aggregate` (theo KIND) tách `str`(primitive) khỏi struct/sum/option/result/array/tuple. KHÔNG đụng case param/return/call (giữ nguyên).

**VERIFIED (stage1, CẢ -O0 lẫn -O1 — O0 = proxy cho stage2 KHÔNG elide):**
tf3 80→**79**, tff_a 127→**7**, tpi/`print_i64_raw` rỗng→**0/753/1213284**, tprp/`print_raw_ptr` `[]`→**[ABC]**, t_strip/tslice ĐÚNG. Đang full rebuild stage2→stage3→stage4 so SHA.

**Trạng thái trước fix:** stage2 build stage3 (exit0) nhưng stage3 crash (print %d rỗng + segfault pha reloc); stage2 1775104 ≠ stage3 1814528.

**✅ KẾT QUẢ sau fix (verify đầy đủ blga6v6wi):** stage3 KHÔNG còn crash — chạy TOÀN BỘ pipeline, **print %d ĐÚNG** (total_len=1214955, func 660/753...), build được stage4 (exit 0). **Cột mốc lớn: stage3 từ crash → chạy.** Nhưng CHƯA fixpoint: stage2(1775616) ≠ stage3(1814528), và stage4 chỉ 656384 byte → lộ BUG#31.

---

## BUG #31 — stage3 sinh PROLOGUE hỏng: `mov %rsp,%rbp` (48 89 e5) → `48 89 00` (mov %rax,(%rax)) + MẤT push callee-saved → mọi chương trình stage3-compiled hỏng 🔬 ROOT-CAUSED (chưa fix)

**Triệu chứng:** Sau BUG#30, stage3 CHẠY nhưng compile chương trình ra binary HỎNG: exe nhỏ ~60% (t_param5: stage2=71168 vs stage3=43520), chạy KHÔNG in gì (exit 0 nhưng rỗng). stage4 (656384) segfault. Differential stage2-vs-stage3 trên t_param5/t_strip/tf3/tpi: stage2 ĐÚNG, stage3 ra rỗng hết.

**Định vị (objdump main của t_param5, obj=axiom_temp.obj):**
- stage2 prologue (ĐÚNG): `55 push rbp` / `48 89 e5 mov rsp,rbp` / `push rbx,rsi,rdi,r12` / `sub rsp`.
- stage3 prologue (HỎNG): `55 push rbp` / **`48 89 00` = mov %rax,(%rax)** (đáng lẽ `48 89 e5`) / THIẾU HẲN push rbx/rsi/rdi/r12 / `sub rsp`.
- ⇒ ModRM của `mov rsp,rbp` bị ZERO (reg=rsp(4)/rm=rbp(5) → 0/0, mod 11→00) VÀ các push callee-saved bị bỏ. Frame hỏng → crash/không chạy.

**Bản chất (REFINED — KHÔNG chỉ prologue):** đếm byte `48 89 XX` (MOV r/m64,r64) trong obj stage3-compiled t_param5: **149× `48 89 00`** (modrm=0 = hỏng) vs stage2 (đúng) có modrm đa dạng (01,02,10,11,41,45,4a,c8...). ⇒ stage3 sinh **modrm byte = 0 cho HẦU HẾT lệnh MOV**, chỉ vài cái (48 89 c8) đúng. modrm = `(mod<<6)|(reg<<3)|rm` ra 0 ⇒ field reg/rm/mod đọc thành 0 — lỗi SPILL/đọc-thanh-ghi trong hàm mã hoá modrm (x86_modrm.ax/x86_encoding.ax x86_encode_modrm_rr) dưới áp lực thanh ghi stage2 sinh ra. **HỌ BUG#24/#27** (REX/modrm-loss under spill) tái diễn trong ENCODER của stage3. emit_prologue chỉ là 1 nạn nhân (mov rsp,rbp + push callee-saved cũng dùng cùng encoder).
Cần soi: `x86_encode_modrm_rr` (x86_modrm.ax:82) — `out_modrm[0] = (MOD_REG_DIRECT<<6)|((reg_field&7)<<3)|(rm_field&7)`. `MOD_REG_DIRECT=3`, `3<<6=0xC0` ⇒ modrm KHÔNG BAO GIỜ được =0; stage3 ra 0 ⇒ `<<6` collapse (chữ ký BUG#24: `6 as u8` qua OP_CAST → CL-form shift clobber).

⚠️ **KHÔNG phải shift!** repro nhỏ `bin/tshift.ax` → stage2 ĐÚNG = 229. Và disasm `ax_x86_encode_modrm_rr` trong obj stage3-compiled (rebuild stage2→compiler, giữ /tmp/stage3_modrm.obj có symbol) cho thấy **modrm_rr ĐÚNG** (`shl $0x6` immediate, OR chain đúng, store dl=0xE5 vào out_modrm[0]). modrm byte 0xC0+ không bao giờ =0 từ rr-path.

### ROOT CAUSE THẬT (disasm `ax_x86_encode_mov_rr` obj stage3-compiled): addr-taken const-fold (HỌ BUG#25)
`mov rsp,rbp` (48 89 00) đi qua **x86_encode_mov_rr** (x86_encoding.ax:23), KHÔNG phải rr-path trực tiếp:
```
mut modrm := 0; x86_encode_modrm_rr(src, dst, &modrm, ...); push 0x89; push_byte(buf, modrm)
```
Disasm stage3's mov_rr: `modrm` init 0 giữ ở **r15** (17170f xor→r15=0); `&modrm` (lea -0x40) truyền cho modrm_rr (callee GHI 0xE5 vào [-0x40] — ĐÚNG); NHƯNG `push_byte(buf, modrm)` cuối (1717ac/af) đẩy **r15 (=0 initial)**, KHÔNG reload -0x40. ⇒ modrm byte = 0 cho MỌI mov_rr/add_rr/... (cả prologue mov rsp,rbp + push callee-saved bị bỏ vì cùng path hỏng).
**Bản chất:** stage2's OPTIMIZER const-fold biến `modrm` (init 0) thành hằng, BỎ QUA `&modrm` escape vào call (callee ghi đè). = **đúng BUG#25** (addr-taken const-fold làm mất giá trị ghi qua con trỏ). stage1's optimizer xử lý ĐÚNG (nên stage2.exe modrm đúng) NHƯNG stage1 MISCOMPILE chính mã optimizer của stage2 → stage2's optimizer mất khả năng tôn trọng addr-taken → const-fold `modrm` khi compile mov_rr/modrm encoder.

### HƯỚNG FIX
Tìm vì sao stage1 miscompile addr-taken tracking của stage2 (ssa_opt.ax `mark_addr_taken_regs` + fold/alg_simp chỉ ghi vals khi `not addr_taken`). Có thể: (a) mark_addr_taken_regs CHÍNH NÓ bị stage1 miscompile (CLASS#3 trong optimizer); (b) gap: `&local` truyền làm CALL ARG (OP_MAKE_REF rồi store vào arg slot) không được mark khi optimizer của stage2 chạy. → disasm `ax_mark_addr_taken_regs` / fold trong obj stage3-compiled so stage1; HOẶC kiểm tra mark_addr_taken_regs có quét hết OP_MAKE_REF + nhận diện modrm-style local. Repro nhanh: chính `ax_x86_encode_mov_rr` trong /tmp/stage3_modrm.obj (push r15=0 thay vì reload -0x40). Đây là 4-5 lớp sâu (chain #25 tái diễn ở meta-level: optimizer tự miscompile optimizer). XÁC NHẬN CLASS#3: `bin/t_alias.ax` (repro gốc BUG#25, cùng pattern &local→call-arg) chạy ĐÚNG qua stage2 (A=1); chỉ `modrm` trong x86_encode_mov_rr (hàm lớn, áp lực thanh ghi) bị fold → KHÔNG repro được ở test nhỏ, phải disasm stage1-obj của mark_addr_taken_regs/fold hoặc marker-pin build thật.

⚠️ **Cũng phát hiện latent bug (KHÔNG phải blocker)**: struct 16-byte truyền BY-VALUE làm param + return (`bin/tsp.ax`) sai ngay ở stage1 (O1→16, O0→127, want 7). Có ở stage1 (reference) nên là latent cũ; compiler tránh truyền struct by-value (dùng ptr) nên không chặn self-host. Cùng họ size==16 với BUG#30 nhưng ở case param/return của regalloc_is_16byte — để sau.

**Trạng thái:** stage2≠stage3≠stage4, chưa fixpoint. Chain bug: #24→#25→#26→#27→#28→#29→#30(✅ stage3 chạy)→#31(prologue). Mỗi fix lộ lớp tiếp.

---

## AUDIT lỗi đã biết trên toàn source (2026-06-25, trước khi fix BUG#31)

Quét theo HỌ bug (vì #24-31 tái diễn theo lớp). Kết quả:

### Họ A — REX/ModRM mất cho byte-reg 4-7 (spl/bpl/sil/dil) [#24/#27/#31] → ✅ SẠCH
- `x86_encode_setcc` (x86_encoding.ax:233): `if reg_needs_rex(dst) or reg_hw_reg(dst) >= 4` → force REX. ĐÚNG.
- `x86_encode_movzx_br` (244): `encode_rex(true,...)` REX.W=1 luôn có REX. ĐÚNG.
- `x86_encode_mov_store_sized` (445): đã có fix BUG#27 (size==1, src_hw 4-7 → force REX). ĐÚNG.
- KHÔNG còn instance latent. (Op 64-bit luôn có REX.W nên reg 4-7 an toàn.)

### Họ B — addr-taken const-fold [#25/#31] → ✅ SOURCE SẠCH
- `mark_addr_taken_regs` (ssa_opt.ax:161) quét OP_MAKE_REF → mark src1. ĐÚNG & đầy đủ.
- Chỉ `fold_func` (195) + `alg_simp_func` (1176) fold giá trị → CẢ HAI đã guard `not addr_taken[dest]`. ĐÚNG.
- `copy_prop_func` (263) chỉ đổi tên vreg (dest→src1, def_count==1), KHÔNG fold hằng → an toàn. Verified `bin/t_cpaddr.ax` O0/O1 = 7.
- `cse_func` (1388) — verified `bin/t_cse.ax` O0/O1 = 98 (KHÔNG CSE nhầm qua mutation). An toàn.
- ⇒ BUG#31 KHÔNG phải lỗ hổng source; là stage2's optimizer bị stage1 miscompile dưới register pressure (CLASS#3). Fix ở meta-level (xem BUG#31).

### Họ C — str(inline-16) vs struct by-address(pointer-8) lẫn lộn [#29/#30] → ⚠️ CÒN LATENT
`regalloc_is_16byte` (x86_selector.ax:435) còn NHIỀU case `size==16` CHƯA guard `not type_is_aggregate` (chỉ case OP_COPY:496 đã fix ở BUG#30):
- **param (442)**, **call-return direct (469) + dynamic (477)**, **cast (511/525)**, **case ~547** — đều `if size==16: return true` cho MỌI aggregate.
- **CONFIRMED broken (latent):** struct 16-byte BY-VALUE làm param+return → `bin/tsp.ax` sai (stage1 O1→16, O0→127, want 7). KHÔNG block self-host vì compiler truyền struct qua ptr, hiếm by-value.
- ⚠️ FIX RỦI RO (chạm ABI 16-byte gồm cả `str` truyền/return hợp lệ inline-16): phải thêm `not type_is_aggregate` từng case + TEST tsp=7 VÀ str param/return vẫn đúng, rồi verify 2.5h. Nên gộp vào RFC representation 16-byte (cùng BUG#30), KHÔNG sửa vội trước BUG#31.
- Site OP_GET_FIELD/SET_FIELD/MAKE_REF (1389/1467/1526/1583...) ĐÃ dùng `type_is_aggregate` đúng (lịch sử fix nested struct).
- **THỬ FIX 1-DÒNG (2026-06-26) → THẤT BẠI, ĐÃ REVERT:** thêm `not type_is_aggregate` vào param(442)+call-return(469/477) làm tsp ĐỔI từ wrong-value 16 → **SEGFAULT** (str regression vẫn OK). Lý do: return struct by-value KHÔNG phải 8-byte pointer — pointer trỏ stack đã chết của mk → deref crash. Cần ABI sret (caller cấp chỗ, callee copy 16 byte) HOẶC inline-register copy NHẤT QUÁN cả 3 phía (call-arg lowering + param receipt + return). ⇒ ĐÚNG là việc của RFC ABI 16-byte (CLAUDE.md: ABI change cần RFC). KHÔNG fix bằng guard regalloc đơn lẻ.
- **→ ĐÃ VIẾT RFC:** `rfcs/0001-16byte-aggregate-byvalue-abi.md` (Phương án B = by-address nhất quán).
- **✅ ĐÃ IMPLEMENT + VERIFIED (2026-06-26):** guard `not type_is_aggregate` 4 điểm trong x86_selector.ax: `regalloc_is_16byte` param + call-return direct/dynamic; `emit_param_prologue` register + stack (aggregate param = lưu con trỏ 8-byte, KHÔNG copy 16-byte inline). Root tìm bằng gdb: thủ phạm là `emit_param_prologue` (không phải OP_COPY lazy path — đã chết). Kết quả: **tsp=7, tsp2=9, tsp3=12** (16+24 byte), str + toàn bộ regression OK, **fixpoint stage3==stage4 = d7f14c2c GIỮ NGUYÊN**. Family C ĐÓNG.
- **Semantics đã chốt (không còn TODO treo):** struct = REFERENCE nhất quán (truyền con trỏ, không memcpy). Verified `bin/tstruct_abi.ax`: A=7(16B) B=12(24B) C=15(str-field) D=6(nested) E=99(mut param ⇒ caller đổi, GIỐNG assignment-alias BUG#30). Khớp spec Single-Ownership/heap. Đổi sang value-copy sẽ bất nhất + phá fixpoint ⇒ BÁC; nếu muốn value-type rõ ràng là feature riêng (RFC khác). Chi tiết RFC 0001 §11. tstrfield "16" trước đó là STALE exe (build timeout 124 do Defender) — rebuild sạch = 15.

**Kết luận:** A & B sạch. C có 1 latent thật (struct by-value param/return, repro tsp) nhưng KHÔNG block self-host → defer. Blocker self-host vẫn là BUG#31 (meta-level optimizer miscompile).

---

## BUG#31 — ROOT CAUSE THẬT + FIX (2026-06-25): copy_prop thiếu guard addr_taken

**Repro NHANH (KHÔNG cần build 2.5h):** `bin/t_movrr.ax` = copy nguyên văn `x86_encode_mov_rr` + call chain (modrm encoder, encode_rex, push_byte, ByteVec). Emit `mov %rsp,%rbp` phải = `48 89 e5` → in `072 137 229`.
- `axc_stage1 build t_movrr.ax -O1` → **072 137 229** ✓
- `axc_stage2 build t_movrr.ax -O1` → **072 137 000** ✗ (modrm byte = 0)
- `axc_stage2 build t_movrr.ax -O0` → 072 137 229 ✓ (optimizer tắt → đúng)
→ Lỗi ở OPTIMIZER (-O1), và stage1 đúng / stage2 SAI trên CÙNG source ⇒ META miscompile.

**Cô lập bằng dump-air (vàng):** `axc_stage1 dump-air t_movrr.ax -O1` vs `axc_stage2 dump-air ...`, diff hàm `x86_encode_mov_rr`:
```
stage1 (đúng):            stage2 (sai):
  %5 = cast %4 (modrm)      %5 = cast %4 (modrm)
                            %6 = copy %5        <-- COPY thừa của biến addr-taken
  %12 = mkref %5            %12 = mkref %6      <-- &modrm trỏ vào BẢN COPY %6
```
`%6 = copy %5` là CELL của biến `modrm` (addr-taken). callee ghi slot %6 qua con trỏ. Nhưng read modrm bị copy_prop đổi tên %6→%5 (init 0) trong khi mkref/store vẫn ở %6 → read đọc giá trị init cũ = STALE.

**ROOT CAUSE:** `copy_prop_func` (ssa_opt.ax) THIẾU guard `addr_taken` mà `fold_func` đã có. Nó lập `copy_map[%6]=%5` cho `%6=copy %5` dù %6 addr-taken, rồi propagate read trong khi địa chỉ vẫn ở %6. stage1 (gcc-compiled) tình cờ propagate NHẤT QUÁN (cả mkref) nên đúng; stage2 (native, bị stage1 build) propagate KHÔNG nhất quán (chỉ read) → sai. Backend ĐÃ force-spill addr-taken vreg (x86_regalloc address_taken→spill); guard này chỉ chặn optimizer short-circuit reload (đúng intent ghi ở ssa_opt.ax:155-159).

**FIX (ssa_opt.ax `copy_prop_func`):** thêm `let addr_taken = mark_addr_taken_regs(f, max_reg)`; khi lập copy_map: `... and not addr_taken[inst.dest] and not addr_taken[inst.src1]`; free cuối hàm. Guard CẢ 2 đầu: dest addr-taken (memory đổi) + src addr-taken (snapshot phân kỳ sau ghi qua con trỏ). Robust theo construction: copy_prop KHÔNG đụng addr-taken vreg → không thể propagate sai dù stage2's optimizer có bị miscompile.

**Validate stage1 (no regression):** t_movrr=072 137 229, t_cp2=7, t_cpaddr=7, t_cse=98, t_modrm=229 — tất cả ĐÚNG.

**✅ VERIFIED — SELF-HOST FIXPOINT (2026-06-26, verify_bug29_selfhost.sh):**
```
stage2: bef3522d… (1776128, build by stage1/gcc)
stage3: cf5a7c6a… (1814528, build by stage2/native)
stage4: cf5a7c6a… (1814528, build by stage3/native)
SELF-HOST OK: stage3 == stage4 (CONVERGED FIXPOINT)
```
Trước fix: stage4=656384 segfault. Sau fix: stage4 byte-identical stage3 (SHA khớp). Compiler native compile chính nó tái tạo bit-for-bit. Xác nhận trực tiếp: stage2 MỚI build t_movrr → **072 137 229** (cũ: 072 137 000); stage3 build t_movrr → 072 137 229. **BUG#31 ĐÓNG.** Backup golden binaries: bin/axc_stage{2,3}_selfhost_fixpoint.exe.

(stage2≠stage3 là BÌNH THƯỜNG: builder của stage2 là stage1/gcc khác codegen native; tiêu chí self-host đúng là native→native fixpoint stage3==stage4.)

**Lưu ý môi trường:** Windows Defender realtime ON → mỗi build self-link scan exe/obj ~51s (build tiny program tốn ~51s wall, CPU ~0). KHÔNG phải hang; đừng đặt timeout < 90s. Build compiler đầy đủ thì 51s này là nhiễu nhỏ so với ~3h compile stage1→stage2.

---

## PLAYBOOK — Quy trình debug stage2 self-compile crash

Đây là cách làm hiệu quả nhất đã rút ra (tránh mò mẫm):

### 1. gdb để lấy ĐÚNG lệnh gây lỗi (công cụ vàng)
`C:\msys64\ucrt64\bin\gdb.exe` chạy được trên PE self-link (không symbol vẫn lấy được fault address + thanh ghi):
```bash
gdb --batch \
  -ex "run build <src.ax> -o out.exe -self-link -O1" \
  -ex "info registers rip rax rbx rcx rdx rsi rdi rbp rsp r8 r9" \
  -ex "x/4i \$rip" -ex "bt" \
  bin/axc_stage2_native.exe 2>&1 | grep -v "^\[Debug\]" | tail -40
```
- Win64 ABI: `memmove(rcx=dst, rdx=src, r8=count)`. `rdx=0` ⇒ memcpy với src NULL ⇒ một `str.ptr` bị NULL ⇒ thường là **BUG CLASS #1**.

### 2. Marker in + fflush để pin dòng crash (khi log 0 byte / không rõ)
- Chèn `ax_puts_local("AXDBG-Xn")` + `fflush(null as ptr[void])` quanh vùng nghi. **Prefix KHÔNG bắt đầu bằng `[`** để tránh bị `is_verbose_debug` (print_helpers.ax) lọc. Marker `[Debug] ...` không whitelist sẽ bị NUỐT — đừng dùng.
- Sửa trong `bootstrap/stage1/main_air.ax` (hoặc file gốc), rồi `rebuild_stage1.ps1` (regen concat tự xoá marker khi build sạch). KHÔNG sửa trực tiếp `tmp_concatenated_air.ax` (bị regen đè).

### 3. Test cô lập (nhanh, ~30s, không cần build stage2)
- Viết `bin/t_xxx.ax`, build bằng stage1 `-self-link -O1`, chạy, đọc exit code. Tái hiện bug ở quy mô nhỏ → fix → verify mà không tốn ~10 phút build stage2.
- LƯU Ý: bug phụ thuộc áp lực thanh ghi/hàm lớn (CLASS #3) **không** reproduce được ở test nhỏ — phải pin bằng marker trong build stage2 thật.

### 4. Phân biệt lỗi native vs logic
Nếu hàm chạy đúng dưới C backend (stage1 build OK / `emit-c`) nhưng sai khi `-self-link` → **lỗi native codegen**, KHÔNG phải lỗi logic. Đừng sửa thuật toán; soi `x86_selector.ax` / `x86_regalloc.ax` / `air_builder.ax`.

### 5. Lưu ý vận hành
- Tiến trình `axc_stage1.exe`/`axc_stage2_native.exe` mồ côi giữ khóa file exe → "process cannot access the file". Kiểm tra `Get-Process *axc*`, kill trước khi build lại.
- Determinism: `tmp_concatenated_air.ax` regen từ source nên markers tự biến mất ở build sạch. Đảm bảo source sạch trước khi so SHA-256(stage2)==SHA-256(stage3).

---

## BUG#32 (GAP, phát hiện 2026-06-26) — User-defined sum type / enum: KHÔNG có codegen native (construct + match)

**Triệu chứng:** `type Color = Red|Green|Blue` (và `enum` RFC 0003 desugar về cùng AST) build được nhưng SAI runtime. Variant no-payload + match → exit 0 (mọi arm bị bỏ). Variant có payload `Circle(i64)` → linker error `Unresolved external symbol 'ax_Circle'`. dump-air main = `iconst; ret` RỖNG.

**Root cause (kép, đều ở native pipeline):**
1. **match KHÔNG được lower:** `lower_stmt` (air_builder.ax:1686) THIẾU nhánh `NODE_MATCH_STMT` (21) → rơi vào `else: lower_expr` → lower_expr cũng không có NODE_MATCH → match sinh 0 AIR. (NODE_MATCH_STMT/ARM CÓ trong parser+resolver+typecheck, nhưng air_builder bỏ sót.)
2. **Variant constructor user-defined KHÔNG được lower:** air_builder.ax:1033 CHỈ hard-code Ok/Err/Some/None (pointer-layout box, tag bit0). Mọi variant khác (`Red`/`Circle`) rơi vào struct-ctor/func-call → `ax_<Variant>` unresolved hoặc bỏ.

**Vì sao self-host vẫn OK:** compiler tự host KHÔNG dùng `match` / sum-type-do-người-dùng (chỉ Result/Option + if/elif). Nên gap này không phá fixpoint — nhưng khiến enum/sum type vô dụng lúc runtime.

**Để dùng enum/sum type thật cần (RFC backend, lớn):** (a) biểu diễn runtime tagged-union tổng quát (tag + payload), (b) lower constructor mọi variant (gán tag, box payload), (c) lower NODE_MATCH_STMT (đọc tag, nhánh theo arm, bind payload, exhaustiveness), (d) typetable lưu tag/layout variant. Rủi ro fixpoint: trung bình (đụng air_builder + typetable). KHÔNG phải parser sugar.

**RFC 0003 (enum) trạng thái:** parser sugar ĐÚNG (enum desugar == sum type, exit byte-identical). Nhưng giá trị thực tế bị chặn bởi BUG#32. Enum chỉ usable sau khi có ADT codegen.

**BUG#32 — ✅ FIXED v1 (2026-06-26, RFC 0004):** thêm ADT codegen native. Representation: sum value = con trỏ 8-byte tới heap box [tag i64 @0, payload 8B @8] (size 16, kind6 aggregate → value vẫn 8-byte by-address). Dùng thẳng sum type id cho OP_ALLOC/SET_FIELD/GET_FIELD: get_register_type trả sum id, field_offset(sum,idx) rơi fallback idx*8 → field0@0/field1@8 ĐÚNG, field_size(sum)=8; KHÔNG cần sửa selector. Lowering ở air_builder.ax: lower_variant_construct (no-payload qua lower_ident, payload qua lower_call_expr — đặt SAU path tên Ok/Err/Some/None + guard kind==SUM), lower_match (OP_GET_FIELD tag@0 + chuỗi OP_EQ/OP_BRANCH; bind payload field1; bare-ident-là-variant resolve theo tên; catch-all binding + wildcard). **Cạm bẫy:** Option/Result CŨNG là `type Option=Some(T)|None` (Some/Ok là SYM_VARIANT kind SUM) — phân biệt CHỈ bằng TÊN; path tên phải chạy TRƯỚC user-variant path, KHÔNG reorder (reorder làm is_some/unwrap đọc sai layout → bug). Test: t_enum_np=6, t_enum=42, t_adt2=104, t_adt3=19, t_builtin_opt=15; regression 19/19; **fixpoint stage3==stage4=692ba8e9 GIỮ** (compiler tự host không dùng match/user-sum). Hạn chế v1 (follow-up): multi-field variant, payload str/>8B, generic user sum + match trên Option/Result. Chi tiết: rfcs/0004-adt-codegen.md, docs/next-step-15-sub-1.md.

---

## BUG#33 (GAP, phát hiện 2026-06-26) — Float arithmetic KHÔNG được sinh trong native backend (compiler integer-only)

**Triệu chứng:** `let c = (b + 1.0) / 2` (b: u32) → typecheck cho `c: f64` ĐÚNG (1.0 float→f64; promotion u32+f64→f64; rồi /2 với 2 int-literal adopt f64 theo RFC 0005 → f64). NHƯNG AIR: `iadd`/`idiv`/`iconst` (số nguyên) dù kiểu f64 → **mis-compute lúc chạy**.

**Root cause (kép):**
1. `map_binary_op` (air_builder.ax:171) trả OP_IADD/ISUB/IMUL/IDIV theo TOÁN TỬ, KHÔNG theo kiểu. OP_FADD/FSUB/FMUL/FDIV CÓ định nghĩa (air.ax) và x86 selector XỬ LÝ được (x86_selector.ax:899-911) nhưng **AIR builder KHÔNG BAO GIỜ phát chúng** (grep OP_FADD air_builder = rỗng). → kể cả `1.0 + 2.0` thuần float cũng ra iadd.
2. Mix int-var + float (vd `b(u32) + 1.0`): KHÔNG chèn cvt int→float cho biến `b` (AXIOM cố ý không implicit int→float cho biến; RFC 0005 chỉ ép literal). typecheck promotion ra f64 nhưng codegen không convert.

**Vì sao chưa lộ:** compiler tự host integer-only (không dùng float math) → fixpoint không ảnh hưởng. Đây là gap aspirational.

**Để fix (RFC float-arithmetic, follow-up):** (a) lower_binary_expr chọn OP_FADD/FSUB/FMUL/FDIV khi node_type là f32/f64 (thay vì luôn OP_IADD…); (b) chèn OP_CAST int→float (cvtsi2sd) khi một toán hạng là int-var còn vế kia float (hoặc YÊU CẦU `as f64` tường minh + chỉ chọn float-op); (c) literal float/int hỗn hợp đã do RFC 0005 + promotion lo phần kiểu. Cần test + giữ fixpoint (compiler không dùng float nên an toàn).

**✅ FIXED cho f64 (2026-06-26, RFC 0006 phần 1; f32 deferred).** Matrix 216/232 PASS (16 fail còn lại = ĐÚNG f32, deferred). Hoá ra "float arith never emitted" chỉ là phần NỔI; đào sâu lộ 3 bug encoder/regalloc nghiêm trọng hơn (float backend CÓ sẵn nhưng CHƯA TỪNG chạy đúng):
1. **lower_float_lit truncate float→int:** `src1: val as u32` → hằng số 3.5 lưu thành 3, mất bit pattern; OP_FCONST còn emit MACH_MOV_IMM vào vreg XMM (vô nghĩa). FIX: lower_float_lit lưu IEEE-754 bits (reinterpret `(&fv as ptr[u64])[0]`) tách low32→src1/high32→src2; OP_FCONST selector dựng lại bits trong GPR scratch (dest+60000) rồi MACH_MOVDQ (movq xmm,r64) vào XMM.
2. **🔑 movq xmm↔gpr THIẾU prefix 0x66** (x86_encoding.ax movdq/movqd): `4d 0f 7e c3` decode = `movq r11,mm0` (MMX! mm0-7 alias x87 stack, KHÔNG phải xmm). Mọi round-trip gpr↔xmm đọc/ghi sai register file → int→float ra rác. FIX: thêm `push_byte 0x66` trước REX ở cả movdq (66 REX.W 0F 6E) và movqd (66 REX.W 0F 7E).
3. **🔑 spill reload float dùng GP scratch:** insert_spill_code (x86_regalloc.ax) reload toán hạng spill vào REG_R10/R11 bằng MACH_LOAD nguyên. Với vreg float, lệnh float (cvttsd2si/addsd) encode operand là XMM; R10 (hw-idx 10) ALIAS XMM10 → CPU đọc xmm10 trong khi giá trị ở GP r10 → rác (FTOI ra 0/7). FIX: nếu `is_float_vreg(v)` → reload vào XMM scratch (src1→XMM0, src2→XMM1) bằng movsd. MACH_MOV đã reg-class-aware (movsd/movdq/movqd) nên FADD path cũ vẫn đúng.

**Code path chạy đúng f64:** lower_binary_expr remap OP_F* khi result f32/f64 + chèn OP_ITOF promote toán hạng int; lower_cast_expr int→float=OP_ITOF, float→int=OP_FTOI (truncate).

**✅ FIXED f32 (2026-06-27, RFC 0006 part 4) — MATRIX 232/232 (F=0!).** Thêm họ encoder `ss` (x86_encoding.ax: addss/subss/mulss/divss F3; comiss 0F2F không 66; cvtsi2ss/cvttss2si F3+REX.W; cvtss2sd F3 0F5A; cvtsd2ss F2 0F5A; movss F3 0F10/11; movd 66 0F6E KHÔNG REX.W). Width 4-byte threaded qua field `padding` của MachInst: emitter chọn ss khi padding==4, sd khi khác; movss vs movsd cho xmm load/store; movd vs movq cho MACH_MOVDQ (FCONST). Selector set padding theo type_id (9→4): FADD/FSUB/FMUL/FDIV/ITOF (theo dest), FTOI/FCMP (theo SOURCE float). f32↔f64 cast: OP_CAST selector phát MACH_CVTSS2SD (f32→f64)/MACH_CVTSD2SS (f64→f32) mới. lower_float_lit: f32 lưu 32-bit pattern (`fv as f32` — dead path self-host, an toàn fixpoint). **Lưu ý:** ss ops để f32 ở low-32 của xmm; spill/store dùng movsd 8-byte vẫn OK vì slot 8-byte & đọc offset 0 lấy đúng low-32. Repro bin/t_f32.ax (exit 5). **Float dst-spill** (DST_UNUSED cho MACH_FADD/ITOF/MOVDQ) vẫn chưa xử lý — chưa lộ. **Float compare** (OP_LT… trên f32/f64) đã có sẵn qua select_comparison (FCMP + unsf_cc), matrix chưa test float-compare value nhưng path có.

---

## BUG#34 (GAP, phát hiện 2026-06-26) — Chia/mod unsigned dùng IDIV signed + narrowing thầm lặng

**✅ FIXED part 2 (2026-06-27, RFC 0006 phần 2): unsigned div/mod (#34.1) + signed-aware shift (#34.2).** Commit sau.
- **OP_IDIV/OP_IMOD** giờ chọn theo `sel_type_is_unsigned(get_register_type src1)`: unsigned → zero rdx (`mov rdx,0`) + **MACH_DIV** (F7 /6, encoder mới x86_encode_div_r); signed → cqo + MACH_IDIV (cũ). Trước: LUÔN cqo+IDIV → u32/u64 high-bit sai (vd 0x80000000/2 ra 3221225472 thay vì 1073741824; u64 max/2 ra 0).
- **OP_SHR** chọn MACH_SAR (arithmetic, signed) vs MACH_SHR (logical, unsigned). Trước luôn SHR → i32 `-16>>2` ra số dương khổng lồ thay vì -4.
- **🔑 BẪY: divisor của DIV phải FORBID rax/rdx** (x86_regalloc.ax:344). MACH_IDIV đã được đăng ký `forbid_rax_rdx[divisor]` (vì thế signed chạy); MACH_DIV mới QUÊN → divisor rơi vào rdx → `mov rdx,0` đè divisor → **div by 0 → crash (exit 127)**. FIX: thêm MACH_DIV vào điều kiện forbid_rax_rdx. Bài học: mọi lệnh div implicit-operand cần ràng buộc regalloc.
- Test: tests/arith/_t2 (6 ca: u32/u64 udiv, umod, SAR, SHR, signed div) PASS; matrix 216/232 (16 f32 deferred, không đổi); regression giữ.
- **✅ FIXED part 3 (2026-06-27): unsigned COMPARE.** `select_comparison` (x86_selector.ax) trước LUÔN dùng signed CC (CC_L/G/LE/GE) cho mọi integer compare → sai cho u32/u64 high-bit (vd `0x80000000 < 1` ra true vì so signed). FIX: nhánh integer chọn `icc = unsf_cc` nếu `sel_type_is_unsigned(src_type)`. May mắn: CC unsigned (B/BE/A/AE) TRÙNG CC float (comisd set CF/ZF) → tham số thứ 2 (đổi tên float_cc→unsf_cc) phục vụ cả hai. Branch (`if`/`while`) test bool từ setcc nên KHÔNG có fused signed-cmp-branch → fix ở select_comparison là đủ. Repro bin/t_ucmp.ax (exit 6). Fixpoint commit sau.
- **CÒN LẠI BUG#34:** (#34.4) lan kiểu 2 chiều, narrowing/đổi-dấu thầm lặng (copy reinterpret i32↔u32) + diagnostics conversion (BUG#35), và BUG#36 mask-về-width. Phần sau RFC 0006.

**Phát hiện qua ví dụ `let a: i32 = -2; let b: u32 = (a-4)/2` → b = 4294967293 (2^32-3).** AIR: isub/idiv SIGNED trên i32 → -3, rồi `copy` i32→u32 reg (reinterpret thầm lặng, không cast/cảnh báo).

**Các gap số học (gộp xử lý ở RFC 0006):**
1. **Unsigned div/mod sai:** `map_binary_op("/")→OP_IDIV` luôn; x86_selector OP_IDIV/OP_IMOD chỉ `CQO+IDIV` (signed), KHÔNG có OP_UDIV/MACH_DIV. → `u32`/`u64` chia số lớn (>2^31/2^63) ra sai. Cần: chọn IDIV vs DIV theo signedness của kiểu toán hạng.
2. **Float arithmetic không phát** (BUG#33): cần lower_binary chọn OP_FADD/FSUB/FMUL/FDIV theo f32/f64.
3. **Không lan kiểu 2 chiều** từ kiểu khai báo (`b: u32 = ...`) vào toán hạng literal trong biểu thức số học (RFC 0005 chỉ lan giữa hai toán hạng cùng biểu thức nhị phân, chưa lan từ expected của let-binding xuống sâu).
4. **Narrowing/đổi dấu thầm lặng** khi gán khác kiểu (i32↔u32, i64→i32): hiện OP_COPY reinterpret, không cast tường minh, không cảnh báo → dễ giấu bug sign/width.

**RFC 0006 (numeric & arithmetic semantics) cần định nghĩa:** signed vs unsigned cho mọi op (div/mod/shift/compare), quy tắc int↔float (chèn cvt hay yêu cầu `as`), chính sách narrowing (cho phép thầm lặng / cảnh báo / cấm), overflow/wrap. Giữ fixpoint: compiler tự host hiện chỉ dùng i32/i64/u32/u64 với div nhỏ → phải kiểm tra kỹ thay đổi div signedness KHÔNG đổi codegen các div hiện có (hoặc đổi nhưng vẫn đúng).

---

## BUG#35 (GAP, phát hiện 2026-06-26) — Gán int↔float ngầm KHÔNG báo lỗi (phải cấm)

**Kiểm chứng (dump-air):**
- `let a: u64 = (b + 2.0) * 3` (b: f64) → KHÔNG lỗi; AIR: `imul` ra t10(f64) rồi `copy` t10→t8(u64) — gán f64 vào u64 thầm lặng (reinterpret bit, không convert). **PHẢI báo lỗi.**
- `let a: u64 = ((b + 2.0) as u32) * 3` → đúng: `cast` f64→u32, `imul` u32, `copy` u32→u64 (widening). KHÔNG lỗi. ✓

**Quy tắc RFC 0006 (conversion policy) — BẤT ĐỐI XỨNG (user chốt 2026-06-26):**
- **float → int: CẤM ngầm, phải `as`** (lossy/truncation). `let a: i32 = 3.0` → LỖI; `let a: u64 = (b+2.0)*3` (b:f64) → LỖI (case 1). Có `as` → OK (case 2).
- **int → float: NGẦM OK** (widening, không cần cast). `let b: f64 = a/3 + 20` (a:i32) → hợp lệ, b là float; chèn cvt int→float tại chỗ gán.
  - Giá trị `a/3+20` (user CHỐT 2026-06-26): operands quyết định op → `a/3` là **int division** (10/3=3), +20=23 (i32), rồi convert int→float tại chỗ gán → **b = 23.0** (KHÔNG phải 23.333). expected float KHÔNG lan xuống làm a/3 thành float-div (giống C/Go/Rust, nhất quán RFC 0005). Muốn float-div: viết `a/3.0` hoặc `(a as f64)/3`.
- Float literal luôn f64 (NODE_FLOAT_LIT); `let x: f64 = 3` → 3 adopt f64 (RFC 0005) → OK.
- Rule cài: let/assign/arg/return — target int-family & expr float-family → **LỖI**; target float-family & expr int-family → chèn cvt int→float (implicit, KHÔNG lỗi).

**Promote ngầm int→float TRONG biểu thức nhị phân (mixed operand):** một toán hạng float làm cả phép thành float, toán hạng int được ngầm convert int→float. VD tương phản (a:i32):
- `a/3 + 20` → `3` int literal → **int division** 10/3=3, +20=23, cvt → **b=23.0**.
- `a/3.0 + 20` → `3.0` float literal → **float division**, a ngầm→f64 → 10.0/3.0=3.333, +20.0 → **b=23.333…**.
AIR hiện tại SAI cả hai phần float: `a/3.0` ra `idiv` trên i32+f64 (không cvt a, không FDIV), `20` ra `iconst` (BUG#33). RFC 0006 phải: typecheck mixed→float + đánh dấu toán hạng int cần cvt; codegen chèn cvtsi2sd + dùng OP_FADD/FDIV.
**Harness:** mode `mixed` hiện dùng cast tường minh; mode `imix` phủ ca **promote ngầm** (int-var + float-literal không cast).

**Ca `let a:f32 = 10/3; let b:i64 = a+10` (user 2026-06-26):** PHẢI lỗi tại `let b:i64 = a+10` (float→int ngầm). Nếu sửa `(a+10) as i64`: `10/3`=int div=3 → a=int→float→**3.0** → `a+10`=13.0 (10 promote) → `as i64`=**13**. AIR hiện tại SAI KÉP: (1) `let a:f32 = 3` ra `copy` t3→t9 (reinterpret bit, a=denormal rác) thay vì cvtsi2ss — **int→float assignment cũng phải chèn cvt, không phải copy**; (2) `let b:i64 = <f32>` ra `copy` không báo lỗi. → RFC 0006: int→float (cả assignment lẫn promote) PHẢI chèn cvt int→float thật (cvtsi2ss/sd); float→int ngầm PHẢI lỗi.

**Liên quan:** phần (d)+(e) RFC 0006. int→int khác kiểu (i32→u32, i64→u32) hiện COPY reinterpret thầm lặng — RFC 0006 quyết định (widening OK / cảnh báo narrowing). float↔int: float→int LỖI, int→float NGẦM.

**An toàn self-host:** compiler tự host integer-only, không gán int↔float → thêm rule này KHÔNG sinh lỗi mới trong source compiler → fixpoint giữ. (Verify lại khi implement.)

---

## BUG#36 (GAP, phát hiện 2026-06-26 qua tests/arith matrix) — Phép toán int < 64-bit KHÔNG mask về bề rộng

**Triệu chứng (matrix run):** `(255 as u8)+(1 as u8)` → got 256 (không wrap 0); `(200 as u8)*2` → 400 (không 144); `(1234 as i16)*(56 as i16)` → 69104 (không wrap 3568). Backend tính trong thanh ghi 64-bit, KHÔNG truncate/mask về bề rộng kiểu sau op.

**Đúng (RFC 0006 §6):** phép toán giữ kiểu toán hạng; runtime tràn → wrap (mask về W bit, sign-interpret nếu signed). Cần chèn mask (and reg, (1<<W)-1) + sign-extend cho signed sau add/sub/mul/shl trên kiểu <64-bit. (Tràn toàn-hằng → lỗi compile, checker riêng.)

**✅ FIXED (2026-06-28).** x86_selector.ax: helper `emit_wrap_to_width(sel, dest, type_id, out_insts)` — nếu type_id là narrow int (i8/u8 size1, i16/u16 size2, i32/u32 size4) gọi lại `emit_load_extend` (đã có sẵn, dùng cho load/cast): signed→SHL/SAR (sign-extend), unsigned→MOV_IMM mask + AND. Gọi cuối các nhánh **OP_IADD/ISUB/IMUL/SHL + OP_NEG(int)** (các op tạo tràn). KHÔNG mask AND/OR/XOR/SHR/DIV/MOD (range-preserving) — **bất biến: mọi producer giá trị narrow (load/cast/arith) đều normalize → toán hạng luôn trong range**. i64/u64/usize/isize (size8) không mask. Probe trước fix: u8/u16/u32/i32 đều KHÔNG wrap (backend dùng 64-bit cho mọi phép). Sau fix: t_mask.ax exit 15 (4 check: 200u8+100u8==44, 60000u16+10000u16==4464, 4e9u32+1e9==705032704, 2e9i32+2e9==-294967296); baseline không-tràn vẫn đúng; regression 29. **RỦI RO fixpoint:** compiler dùng i32/u32 nhiều → mask thêm SHL/SAR/AND sau mỗi narrow arith (bloat); nhưng giá trị compiler luôn trong range nên kết quả KHÔNG đổi → fixpoint phải hội tụ lại SHA mới. [verify đang chạy — nếu vỡ thì revert].

**Lưu ý test:** matrix run 2026-06-26 = 184 PASS / 57 FAIL: nhóm FAIL = (1) MỌI float cast/op/mixed (BUG#33: int→float & float→int & f64→f32 dùng copy-reinterpret, float arith dùng iadd/idiv), (2) wrap-hằng (giờ là EXPECT-ERROR theo policy, đã dời sang diag). int in-range + int↔int cast PASS hết. Unsigned-div (BUG#34) CHƯA lộ ở matrix (operand nhỏ 7/3) — cần chạy fuzz `int` operand lớn.

---

## BUG#37 (GAP, phát hiện 2026-06-27) — `match` trên số nguyên / LiteralPat KHÔNG được codegen (không có "C switch")

**Triệu chứng:** Grammar ([GRAMMAR.ebnf:236](docs/GRAMMAR.ebnf)) định nghĩa `LiteralPat = INT_LIT | FLOAT_LIT | STRING_LIT | 'true' | 'false' | 'nil'`, nên cú pháp cho phép `match x:\n  1: ...\n  2: ...\n  _: ...` (giống `switch` C). NHƯNG codegen KHÔNG implement:
- `lower_match` (air_builder.ax:1855) có guard `if not is_sum: return` → scrutinee KHÔNG phải sum type ⇒ **không sinh code nào** (im lặng no-op, sai).
- Vòng lặp arm chỉ phân loại WildcardPat/VariantPat/BindingPat — **không có nhánh NODE_LITERAL_PAT**, kể cả trong match sum-type.

**Hệ quả:** KHÔNG có lệnh tương đương `switch` C cho số nguyên. Rẽ nhánh theo giá trị int hiện phải dùng `if/elif/else`. `match` chỉ chạy cho sum type / enum / Option / Result (tag dispatch). Đây là lỗi grammar-vs-implementation (CLAUDE.md: grammar là authoritative).

**Để fix (moderate):** mở rộng lower_match: (a) bỏ guard cho scrutinee số nguyên (dùng thẳng scrut_reg thay vì tag), (b) thêm nhánh NODE_LITERAL_PAT → `OP_EQ`(scrut, literal) + `OP_BRANCH`, `_` làm default. Cộng typecheck cho literal pattern. (Tối ưu jump-table = task #7 next-step-15, RFC riêng.) Liên quan [[next-step-15]] task #7.

---

## BUG#38 (AUDIT 2026-06-27) — Nhiều cấu trúc grammar parse/định nghĩa node nhưng codegen SAI/THIẾU (latent vì compiler không tự dùng)

Audit toàn bộ GRAMMAR.ebnf vs parser/typecheck/air_builder/selector. Tất cả LATENT (compiler bootstrap không dùng → self-host xanh; sẽ cắn code người dùng). Mức ưu tiên + cách fix:

- **#38.1 Compound assign `+= -= *= /= %=` (CAO, sai âm thầm):** parser (parser.ax:619) giữ NODE_ASSIGN_STMT với token op nhưng `lower_assign` (air_builder.ax:2077) KHÔNG đọc op → `a += b` chạy thành `a = b` (mất phép). FIX: desugar ở parser `a op= b` → `a = a op b` (build binary node), hoặc lower_assign đọc token → emit binary trước store.
- **#38.2 `match` int/literal = BUG#37 (CAO):** không "C switch". lower_match guard `not is_sum: return` + thiếu NODE_LITERAL_PAT.

**✅ #38.1 + #38.2 FIXED (2026-06-27, commit sau).** #38.1: helper assign_value (air_builder.ax) — lower rhs + nếu token `op=` thì lower_expr(lhs hiện tại) + emit binary (float→OP_F*, unsigned div theo selector); 4 nhánh lower_assign route qua. Hạn chế: lhs side-effecting eval 2 lần (v1). Repro bin/t_cassign.ax exit12. #38.2 "C SWITCH": lower_match thêm is_int (scrut 1-8/11/13/15/16)→match_reg=scrut_reg (không tag); NODE_LITERAL_PAT→parse_int_from_str+ICONST(typed scrut_type,split64)+OP_EQ(match_reg)+branch; typecheck NODE_MATCH_ARM thêm else cho binding non-sum. Repro bin/t_switch.ax exit4.
- **#38.3 Lũy thừa `**` — ✅ FIXED (2026-06-27).** TK_STAR_STAR lex+parse OK nhưng map_binary_op KHÔNG map → OP_NOP (kết quả rác). FIX (air_builder.ax): lower_binary_expr nhận diện op `**` ở đầu → lower_power. Result type từ typecheck (nhánh else numeric-promotion) chọn int OP_IMUL vs float OP_FMUL; base int promote OP_ITOF khi result float. **Số mũ hằng nguyên không âm ≤64 → unroll chuỗi nhân** (n=0→emit_one ICONST/FCONST 1.0; n=1→COPY; n≥2→chuỗi nhân). **Số mũ biến → loop đếm-lùi** (acc/cnt loop-carried qua OP_COPY vào vreg cố định — đúng kiểu register-machine IR mà front-end dùng cho user `mut` local, KHÔNG cần phi). Số mũ âm/phân số KHÔNG hỗ trợ (giới hạn đã biết, không miscompile âm thầm). Base eval đúng 1 lần. Runtime-verified bin/t_pow.ax (exit 59 = unroll 2^10 + 5^0 + 7^1 + loop 3^5 + float 2.0^3); regression 27.
- **#38.4 `defer` — ✅ FIXED (2026-06-28).** Trước: parse_stmt KHÔNG có nhánh TK_DEFER (rơi vào parse_expr_stmt → rác) VÀ lower_defer là STUB SAI (lower_expr NGAY tại chỗ). FIX: (1) parser.ax parse_defer_stmt (TK_DEFER + Expr → NODE_DEFER_STMT). (2) air_builder: FuncLowering thêm `defers: U32Vec`; lower_defer giờ **push** node expr (không chạy); flush_defers emit LIFO vào block hiện tại (KHÔNG clear — mỗi exit path là block riêng); gọi flush ở lower_return (sau khi tính ret_val, trước OP_RETURN) + ensure_return (cuối hàm). (3) typecheck infer_node nhánh NODE_DEFER_STMT type expr con; resolver lo qua fallback resolve_children. **Ngữ nghĩa: function-level LIFO (Go-style)** — defer vô điều kiện ở đầu hàm chạy trên MỌI return path (đúng case chuẩn `defer file.close()`). **GIỚI HẠN:** defer trong nhánh điều kiện không-chọn vẫn bị emit ở exit sau (không có runtime registration) → cần defer-list runtime để đúng 100%. Runtime-verified bin/t_defer.ax: stdout `start/early/B/A/end` (không chạy ngay + LIFO + early-return kích hoạt) + exit 7 (return value giữ); regression 28.
- **#38.5 `unsafe:` block — ✅ FIXED (2026-06-27).** parse_stmt nhánh TK_UNSAFE → parse_unsafe_block (NODE_UNSAFE_BLOCK ôm block); lower_stmt lower block con trong suốt; typecheck recurse vào body; resolver đã có sẵn (resolver.ax:926). unsafe = no-op ngữ nghĩa codegen (gen-check chưa enforce).
- **#38.6 `in [arena]:` block — ✅ FIXED parse+exec (2026-06-27).** parse_stmt nhánh TK_IN → parse_arena_block (NODE_ARENA_BLOCK, lưu tên arena vào payload); lower block con trong suốt. **Routing cấp phát theo arena vẫn DEFERRED** (chỉ chạy block như thường). Repro chung bin/t_unsafe.ax (exit 20: unsafe+arena+compound).
- **#38.7 Closures `|x|` (TB):** resolver biết NODE_CLOSURE_EXPR, không typecheck/codegen. Feature lớn → RFC riêng.
- **#38.8 Tuple patterns — KHÔNG phải fix rẻ, là FEATURE LỚN (RFC).** Audit lại 2026-06-28: tuple KHÔNG phải feature thật. Chỉ có NODE_TUPLE_PAT (parse pattern) + enum TYPE_KIND_TUPLE. KHÔNG có: tuple type parsing (`(i32,i32)`), tuple literal/construction (`(1,2)` — không có NODE_TUPLE_EXPR), field access `.0/.1`, ABI layout. → codegen tuple-pattern một mình là VÔ NGHĨA (không tạo được tuple để match). Cần implement tuple như 1 aggregate type hoàn chỉnh (RFC) trước. Hoãn.
- **#38.9 spawn/await/async, Isolated/Future (Thấp, phase 9):** dispatch/node có nhưng runtime stub.

**Kế hoạch:** nhóm rẻ+an-toàn-fixpoint (#38.1 compound, #38.2 match-int, #38.4 defer, #38.5 unsafe) gộp 1 verify (AST compiler không đổi → fixpoint giữ). #38.3 power + #38.7 closures = lớn hơn, sau. Làm sau khi RFC 0006 part 4 (f32) commit.

**LƯU Ý audit này KHÔNG vét cạn** (pass tĩnh, tầng cú pháp). Chưa audit: type-expr codegen (slice [T], array [T;N], func-type value, Isolated/Future), effect annotation {.raises.}, comptime, monomorphization generic ca khó, runtime/allocator/stdlib, float-compare value-correctness. Audit chuẩn cần test-driven (compile+run từng feature) — cần build. TODO sau.

---

## BUG#39 (AUDIT 2026-06-27) — KHÔNG có operator overloading → bignum/scientific số tùy-biến KHÔNG khả thi với cú pháp `a + b`

**Bối cảnh:** AXIOM định hướng hỗ trợ số siêu lớn / siêu chính xác (scientific). User yêu cầu dùng GIỐNG phép toán số hiện tại (`a + b`, `a * b`). Hiện KHÔNG được vì thiếu 2 trụ cột:

1. **Operator overloading KHÔNG tồn tại.** Có method/function overloading (resolve_method_overload typecheck.ax:373, theo chữ ký) nhưng binary op hardcode qua map_binary_op → OP_IADD/etc (lệnh máy). Toán hạng struct → `a+b` emit OP_IADD trên con trỏ → rác. Không có dispatch `+`→method.
2. **Không có kiểu arbitrary-precision** (chỉ fixed-width i8..i64/u8..u64/f32/f64; KHÔNG cả i128/u128). Không có thư viện BigInt/BigDecimal. Không có add-with-carry/u128 cho limb-arithmetic.

**Để hỗ trợ (RFC + nhiều phase):** (a) operator overloading: typecheck phát hiện toán hạng non-primitive → resolve operator-method (interface Add/Mul hoặc method add/mul), air_builder emit OP_CALL thay OP_IADD; (b) primitive carry (add-with-carry intrinsic hoặc u128); (c) stdlib BigInt/BigDecimal (limb u64 trên heap); (d) arbitrary-precision float (MPFR-style) riêng. Khối xây sẵn CÓ: heap @alloc, u64 arith đúng (sau BUG#34 fix), struct/generics/method. Phase 9+ / aspirational. Operator overloading là bước MỞ KHÓA, làm trước.

**✅ (a) operator overloading FIXED (2026-06-28, RFC 0007, convention-based).** Toàn bộ trong air_builder lower_binary_expr (KHÔNG đổi grammar/parser/typecheck): nếu **kiểu toán hạng TRÁI (node_types[child]) là user type** (STRUCT/SUM/GENERIC_INST) → map op→tên method quy ước (op_to_method_name: + add, - sub, * mul, / div, % rem, == != eq, < lt, <= le, > gt, >= ge) → resolve_op_method (scan SYM_FUNC, match_mangled_method_raw_bytes theo tên + first-param-type==rec sau unwrap ptr/ref — đúng cơ chế lower_call_expr non-generic method) → lower_op_overload emit OP_CALL [self,rhs] (type_id=msym, src1=0, giống method call); `!=` gọi `eq` rồi OP_NOT. **BÀI HỌC:** resolve_method_sym (typecheck) CHỈ tìm generic method ở fallback (guard SYM_FLAG_GENERIC) → KHÔNG dùng được cho struct method non-generic; phải resolve trong air_builder bằng match_mangled_method_raw_bytes. Result type: arithmetic = lhs type (numeric-promotion fallback `result_type=t1`), comparison = bool — KHÔNG cần đổi typecheck. Builtin numeric bỏ qua hoàn toàn → fixpoint-safe. Repro bin/t_opover.ax (exit 44: + add, < lt, == eq, != eq+neg trên struct Num). CÒN LẠI cho bignum: (b) carry/u128, (c) stdlib BigInt limb-u64 — giờ KHẢ THI vì đã có operator overloading.

**GIỚI HẠN MVP v1 (biểu thức trộn bignum + primitive — user nêu 2026-06-28):** (1) dispatch CHỈ theo toán hạng TRÁI → `5 + big` (primitive lhs) đi numeric path → OP_IADD(int, ptr) = RÁC; phải để user-type bên trái. (2) KHÔNG kiểm kiểu rhs (resolve_op_method chỉ khớp param đầu=self); `big + 5` resolve trúng add(self,o:BigInt) rồi truyền int → mismatch; phải convert `big + BigInt.from(5)`. (3) Không ép ngầm primitive→user-type. → **Follow-up nên làm trước khi dùng bignum thật:** dispatch đối xứng (lhs primitive + rhs user-type) + so kiểu rhs với param thứ 2 (lỗi nếu lệch, cần error-infra BUG#35). Primitive-trộn-primitive (`(a>>8) as u8`) KHÔNG bị giới hạn này — chỉ cần `as` tường minh (RFC 0006); ví dụ user `let b:u8=(a>>8) as u8` với a:u32=65555 → b=0 ĐÚNG (65555>>8=256, 256 as u8=0).

**✅ MULTIPLY bignum KHẢ THI THUẦN LIBRARY (2026-06-28) — mulhi/u128 KHÔNG bắt buộc.** 64×64→128 widening multiply viết bằng AXIOM: tách mỗi u64 thành 2 nửa u32, 4 tích u32×u32→u64 (vừa u64), recombine + carry. examples/mul128.ax + repro bin/t_mul128.ax (exit7: 0xFFFFFFFFFFFFFFFF²=...FE_00..01, 1e6×1e6=1e12). Cùng với add (limb+carry, U128), **cả add lẫn mul bignum đều làm được không cần đổi compiler**. → BUG#39 (b) "cần u128/mulhi" KHÔNG còn là blocker; chỉ là tối ưu tốc độ (intrinsic MUL→RDX:RAX nhanh hơn half-limb ~4 lần). Full N-limb BigInt giờ là PURE LIBRARY work.

**✅ NÂNG resolve theo rhs type (2026-06-28):** resolve_op_method giờ nhận rhs_type, khớp param0(self)==lhs VÀ param1==rhs (khi rhs biết) — hỗ trợ operator method mixed-type đơn `add(self,o:i64)` (repro bin/t_opmix.ax exit13), tránh gọi nhầm method khác kiểu rhs. **GAP CÒN LẠI (phát hiện qua dump-air):** method overloading (2 method CÙNG TÊN `add(o:BigInt)` + `add(o:i64)`) collision về 1 symbol @4 (mangling theo TÊN không theo chữ ký) → segfault. Muốn big+big VÀ big+int đồng thời cần **signature-based mangling** (resolver+symtab+codegen, RFC riêng lớn). Giờ: 1 operator method/tên/type (mixed-type được).

**Tham-số-hóa kích thước (`bignum[256]`) cần CONST GENERICS — hiện KHÔNG có.** parse_generic_param (parser.ax:734) chỉ nhận `IDENT [: TypeExpr]` = tham số KIỂU, không có value/const generic; generic arg parse bằng parse_type_expr (chỉ kiểu) → int-literal `256` không hợp lệ; array `[T;N]` N bắt buộc INT_LIT (không generic-hóa). (Phụ: AXIOM dùng `[]` không `<>`.) → Hai hướng: (A) **BigInt động runtime-sized** (arbitrary precision, chỉ cần operator overloading + lib — KHUYẾN NGHỊ trước); (B) fixed-width `bignum[N]` cần const generics (RFC type-system nặng kiểu Rust/Zig, phase rất sau). Const generics cũng mở khóa array-size generic.

---

## BUG#40..44 (2026-06-28) — Float (f32/f64) chuỗi pipeline native: nhiều lỗi chí mạng chặn pure-AXIOM math (sin/cos/exp/sqrt thay libm)

**Bối cảnh:** User muốn thay các hàm extern libm (sin/cos/pow...) bằng AXIOM thuần. Khi viết thử (Newton sqrt, Taylor exp) phát hiện CHUỖI lỗi float ở native backend. RFC 0006 thêm float compare/arith nhưng chỉ test ở pattern đơn giản; matrix 232/232 KHÔNG bắt được vì dùng toán hạng biến (có COPY lót đường) + ít float đồng thời + không gọi hàm.

- **BUG#40 — eval_binary fold so sánh float như integer (ssa_opt.ax).** OP_EQ/NE/LT/LE/GT/GE dùng CHUNG opcode cho int và float; const-fold cast cả 2 toán hạng `as i64` → `1.41 > 10.0` sai ở -O1 (fold), đúng -O0. **FIX:** mảng `is_flt` song song trong fold_func (đánh dấu dest của FCONST), guard `if known and known and not has_flt` → bỏ qua fold khi toán hạng float. (ssa_opt.ax ~201/237).

- **BUG#41 — FCMP dst KHÔNG được tính là READ trong liveness (x86_regalloc.ax).** `comisd dst, src` ĐỌC dst (toán hạng float lhs) nhưng MACH_FCMP thiếu trong `is_two_operand_read` → live-interval của lhs KHÔNG kéo dài tới lệnh so sánh → lhs "chết" ngay sau FCONST → FCONST của rhs được cấp phát TRÙNG XMM → `comisd` so 1 giá trị với CHÍNH NÓ → mọi so sánh strict (`>`,`<`) đều false. -O0 có COPY lót che giấu; -O1 copy-prop bỏ COPY → lộ. **FIX:** thêm MACH_FCMP vào is_two_operand_read.

- **BUG#42 — Float RMW (FADD/FSUB/FMUL/FDIV) dst KHÔNG tính READ (x86_regalloc.ax).** Lower thành `MOV dst,src1; addsd dst,src2` — addsd/subsd/mulsd/divsd ĐỌC+GHI dst. Cùng họ BUG#41: thiếu trong is_two_operand_read → chuỗi `(a+b)*0.5` bị coi chết → XMM tái dùng → sai. **FIX:** thêm FADD/FSUB/FMUL/FDIV vào is_two_operand_read.

- **BUG#43 — Tham số float ĐỌC từ thanh ghi INTEGER thay vì XMM (x86_selector.ax emit_param_prologue).** ABI win64: arg float nằm ở XMM0-3 (positional). emit_param_prologue (nơi THẬT SỰ nạp param; lazy-path trong select_inst là CODE CHẾT vì param_idx_processed bị set = params.len ở select_all) snapshot + nạp MỌI param qua `abi_int_arg_reg` (RCX/RDX) → param f64 đọc RCX = rác → mọi đối số float = 0. **FIX:** trong emit_param_prologue phase 1 (snapshot) + phase 2 (store), nếu `params.data[p]` là 9/10 dùng `abi_float_arg_reg(p)`. Caller-side (đặt arg vào XMM) + return (XMM0) ĐÃ đúng từ trước. (Disasm xác nhận: caller `movsd xmm0,xmm8` đúng; callee `mov rax,rcx` sai → sau fix đọc XMM0). Repro: fa(arg float vào)=75, fb(return float)=75, farg(2 arg)=65.

- **BUG#44 — Thanh ghi callee-saved XMM (win64 XMM6-15) KHÔNG được prologue/epilogue lưu → float bị hỏng QUA lời gọi hàm. 🔴 CHƯA FIX (bị chặn bởi trần stage0).** Allocator cấp XMM8-15 cho float (callee-saved theo ABI win64) NHƯNG codegen không save/restore chúng → callee clobber XMM8 của caller. Test: `let a=3.0; noop(9.0); return (a*100) as i32` → 132 thay vì 300&255=44. Đây là lý do `sqrt_newton` (gọi fabs trong vòng lặp) + fdbg2 (giữ a,b qua call idf) sai. **FIX ĐÚNG (đã thiết kế, chưa land):** thu mask XMM callee-saved đã dùng → reserve slot trong frame → movsd save sau SUB RSP / restore trước ADD RSP (push/pop KHÔNG dùng được cho XMM); offset spill + str-lit dịch thêm xmm_saved_len*8; chèn vùng XMM giữa GPR-pushes và spill. (SysV: XMM toàn caller-saved → cần spill phía caller, chưa làm; win64 là target self-host.)

**🔴 TRẦN STAGE0 (phát hiện 2026-06-28, RẤT QUAN TRỌNG):** stage0 (bin/axc.exe, C-backend bootstrap) FAIL khi compile stage1 source nếu THÊM ĐỦ code — biểu hiện LỖI Ở CHỖ KHÁC: `read_file_content` (main_air.ax) → `r2.unwrap()` suy luận generic Result<str,str> ra type 0 → "expected 12 found 0". Bisect: baseline + chỉ thêm field struct (4 dòng tầm thường) = PASS; thêm 1 HÀM RỖNG mới = FAIL; thêm ~60 dòng code (không hàm mới) = FAIL. → Trần là TỔNG (symbol/IR size), KHÔNG chỉ số hàm. **HỆ QUẢ:** không thể thêm hàm top-level mới / nhiều code vào stage1 qua stage0. Khi land BUG#44 phải: (a) viết INLINE (không hàm mới) — đã thử bitmask inline NHƯNG ~60 dòng vẫn vượt trần; (b) hoặc nâng giới hạn trong stage0 trước; (c) hoặc nén code cực gọn / bỏ bớt path (vd x86_asm_emitter text-asm không nằm trên đường self-host). Đây là blocker cứng cho BUG#44 và mọi mở rộng stage1 tương lai.

**TRẠNG THÁI:** BUG#40/41/42/43 ĐÃ FIX + build OK (1891441) + float test xanh (fa=75 fb=75 fret=75 farg=65 flit=15, chuỗi arith đúng) + regression. → float compare, float arithmetic chains, float args/returns ĐÚNG ở cả -O0/-O1. CÒN BUG#44 (float qua call) + trần stage0. Pure-AXIOM math KHẢ THI nếu hàm tự-chứa (inline helper, tránh giữ float qua call nội bộ); user gọi sin(x) (1 arg vào, 1 kết quả ra) ĐÃ chạy đúng.

---

## std.math + bignum (2026-06-28) — pure-AXIOM math, import-bundling limit, bignum scaling

**std.math viết lại THUẦN AXIOM (bỏ extern libm):** std/math.ax giờ implement sin/cos/tan/exp/ln/log2/log10/pow/sqrt/floor/ceil/round/fmod + abs/min/max/gcd/lcm/pow_i64 bằng AXIOM (Newton, range-reduction, Taylor, atanh series). Parse OK (dump-air thấy đủ hàm). Pattern self-contained để né BUG#44 (float-qua-call): hàm-lá OK; tail-compose `ln(x)/LN2` OK (x tiêu thụ trước call); pow/tan INLINE (param float cần sống qua call). Repro examples/math_pure.ax + bin/t_math.ax exit127 (sqrt/exp/ln/pow/sin/cos/floor) cả -O0/-O1.

**🔴 std.math CHƯA dùng được qua `import` — import bundler hardcode (main_air.ax concatenate_stdlib).** `import std.X` của user KHÔNG load file tùy ý; concatenate_stdlib chỉ đọc CỐ ĐỊNH 7 file: result/mem.alloc/scheduler/runtime/os/string/io. `import std.math` bị strip → math.ax KHÔNG compile vào chương trình. `std.math.sqrt` "chạy" chỉ vì rơi xuống extern libm symbol (runtime/ax_math.c link lúc gcc); hàm non-libm (gcd, abs_i64, axiom_marker...) → unresolved → SEGFAULT. → Muốn `import std.math` thật: thêm math.ax vào concatenate_stdlib (sửa main_air.ax → rebuild stage1 → **đụng TRẦN STAGE0** + đổi fixpoint + bloat mọi binary). Cùng keystone với BUG#44. TẠM THỜI: dùng hàm math IN-FILE (như math_pure.ax) — chạy đúng ngay.

**Bignum = PURE LIBRARY, scale tùy ý (KHÔNG đụng compiler/fixpoint):** limb u64 + carry + operator overloading (RFC 0007). **BUG#44 KHÔNG ảnh hưởng bignum** (struct/int dùng GPR; GPR callee-saved ĐƯỢC lưu qua call) → bignum math COMPOSE tự do qua call. Đã demo:
- examples/bignum_math.ax (bin/t_bignum exit63): U128 — fib(100) (hi=19), factorial(20/25), 2^100, gcd. fib/fact/pow2 đúng cả -O0/-O1.
- examples/bignum256.ax (bin/t_bn256 exit31): **U256 (4 limb)** add ripple-carry + mul_u64 + factorial(40) vượt 128-bit (limb cao ≠0) + **Q64.64 fixed-point** (0.5+0.25=0.75) — cơ sở cho big-float.
- → u512/u1024 = thêm limb; **f128/f256 bigfloat = bignum mantissa + exponent** (hoặc fixed-point Q.n) — cùng pattern, là module lib lớn hơn (std.bignum/std.bigfloat), pure-library; bundling cần giải keystone stage0.

**Bug số nguyên TANGENTIAL phát hiện (pre-existing, KHÔNG từ float work):** (1) **u64 `%`/`/` SAI trong vài ngữ cảnh:** `48%36` đúng ở -O1 khi inline trong main NHƯNG = 0/1 sai ở -O0, và SAI trong vòng lặp gcd ở cả -O0/-O1 (gcd_u64 trả 0). i64 gcd thì đúng. → unsigned div/mod (BUG#34) còn lỗi residual ở -O0 + trong loop. (2) **so sánh u64 literal lớn lệch:** computed fib.lo (đúng, %1000==371) != literal `3737010778780434371 as u64` dù `2432902008176640000` so đúng — parse/compare literal u64 ~19 chữ số có quirk. Cả 2 latent (compiler không dùng), chưa fix; né bằng i64 / modulo-check trong demo.

---

## Mở rộng std.math + dynamic-bignum + codegen footguns (2026-06-28, phiên 2)

**⚠️ COMPILER THỰC HÀNH = bin/axc_stage1.exe, KHÔNG phải bin/axc.exe.** bin/axc.exe (stage0, Go/C-backend) là bản CŨ (Jun 22) VÀ — đã xác nhận bằng `go build` mới tinh — Go stage0 THỰC SỰ thiếu method (`fn m(self, o)`), operator-overload, và có bug type literal u64 (`let mask = 4294967295 as u64` → "found 8 and 3"). Các tính năng đó CHỈ có ở stage1 (compiler AXIOM tự-host). ⇒ build mọi test/example bằng **bin/axc_stage1.exe**. (Đây là "trần stage0" nhìn từ phía tính năng: source stage1 phải nằm trong subset stage0 build được; nhưng compiler để DÙNG là stage1.)

**🟢 dynamic-width bignum (std/bignum.ax + std/xmath.ax, COMMIT 99cd609):** thay vì struct field cố định (U128/U256), số = heap limb array `ptr[u64]` + limb count `n` (width = n*64, chọn lúc RUNTIME). Default 128 (n=2) → u4096 (n=64) → tùy ý. Unsigned + signed (two's complement, bi_*) + fixed-point (bf_*). **Fixed array `[u64; N]` KHÔNG dùng được** — stage0/stage1 codegen sinh `ax_slice_void` undeclared cho array-literal field → phải heap ptr. examples/bignum_dynamic.ax + bin/t_bndyn.ax = 10 check (width động, 2^100@4096-bit, 50!, borrow-sub, shift, binary-gcd, isqrt, fib, signed neg/mul, fixed-mul) → exit 99 @ -O0/-O1, trong regression.

**🟢 std.math +41 hàm (COMMIT a56cc42):** sign/clamp/recip/trunc/fract/square/cube/cbrt/hypot/exp2/exp10/log1p/atan/asin/acos/atan2/sinh/cosh/tanh/asinh/acosh/atanh/deg_to_rad/rad_to_deg/lerp/inverse_lerp/smoothstep/saturate/approx_equal/sigmoid/relu/leaky_relu/softplus/swish/popcount/clz/ctz/bit_width/rotate_left/has_single_bit/mul_hi + const LOG2E/LOG10E/SQRT1_2/EPSILON. bin/t_mathx.ax exit 28 @ -O0/-O1, trong regression.

**🔴 BUG#44 — PATTERN CHÍNH XÁC (mới làm rõ):** float giữ trong thanh ghi QUA một call bị hỏng. Cụ thể:
- `CONST <op> f(call)` → HỎNG (const/biến nạp vào XMM TRƯỚC call → clobber). VD `acos = HALF_PI - asin(x)` trả 0; `atanh = 0.5 * ln(...)` sai.
- `f(call) <op> CONST` / `let r = f(call); ... r` → AN TOÀN (const materialize SAU call). VD `sinh: let e=exp(x); (e-1.0/e)*0.5` đúng; `atan2: return atan(y/x)+PI` đúng.
- ⇒ QUY TẮC LIB: không bao giờ viết `CONST op f(call)`; luôn `let r=f(call); return CONST op r`. acos viết lại theo thứ tự đó; atanh dùng Taylor series thuần (hàm-lá, 0 call). Header std/math.ax ghi rõ quy tắc. Mong manh theo reg-alloc (đổi thứ tự làm flip pass/fail) → lý do PHẢI fix BUG#44 mới mở rộng float-compose diện rộng an toàn.

**🔴 CODEGEN BUG MỚI — struct value trong loop (chặn lớp number-theory bignum):**
1. **In-place accumulate field qua loop KHÔNG persist:** `let q = bu_alloc(n); ...loop... q.limbs[idx] = q.limbs[idx] + X` → chỉ lần ghi CUỐI còn (các vòng trước mất). divmod ban đầu cho quotient sai (q=1 thay 3 cho 10/3). WORKAROUND: build kết quả bằng REASSIGN struct mỗi vòng (`q = bu_shl1(q)` rồi `q.limbs[0]=q.limbs[0]+1`) — chỉ chạm limbs[0], giống pattern bu_from_u64/r đang chạy đúng. divmod sau fix ĐÚNG standalone (1000000/7=142857 r1).
2. **Struct-return field-bind rồi mutate / struct aliasing trong loop → SEGFAULT hoặc kết quả 0:** `mut x := bu_divmod(a,m).r` trong HÀM (không phải main) rồi mutate x trong loop → segfault; Euclid `let t=y; let d=bu_divmod(x,y); y=d.r; x=t` → trả 0 (aliasing vreg OP_COPY). divmod gọi 1 lần trong hàm rồi return field thì OK (e.ax). ⇒ modpow/lcm-Euclid/binomial (divmod-in-loop) CHƯA ship được — pending fix codegen struct-copy/aliasing. Binary-gcd (shift, không divmod) thì OK → đã ship trong xmath. divmod prototype giữ ở scratch.

**Footgun tên hàm:** đặt hàm tên `close` (hoặc tên trùng libc) → LINK ngầm vào libc `close(int fd)`, không phải hàm user → truyền bit f64 làm fd → SEGFAULT. Không phải bug AXIOM nhưng compiler nên cảnh báo shadow extern. Tránh: đừng đặt tên trùng symbol C (close/open/read/write/...).

---

## 🔓 KEYSTONE GIẢI: "trần stage0" KHÔNG phải size — là bug inference let-bound generic-call (2026-06-29)

**Tái hiện chính xác:** concat stage1 source (25585 dòng) build OK qua stage0; +1 hàm tầm thường OK; +40 hàm (~321 dòng) OK → **KHÔNG phải giới hạn kích thước**. NHƯNG **+1 struct mới** HOẶC **+1 hàm dùng Result/.unwrap()** → FAIL ngay tại `read_file_content` (main_air.ax) dòng `return content`: `error[E3005] return type mismatch: expected 12, found 0`. Bug có ở CẢ stage0 cũ (Jun22) lẫn `go build` mới → bug Go source hiện hành.

**Root-cause (đào bằng debug printf trong compiler/sema/inference.go):** `read_file_content` có `let content = r2.unwrap(); return content` với r2: Result[str,str]. Inference pass tính `r2.unwrap()` ĐÚNG = str(12) (DBG: objType=252 gargs=[string,string] subRet=12). NHƯNG ở **CHECK pass** (check_stmt.go:347 `exprType := tc.infer.TypeOf(content-ident)`) ident `content` đọc type=**0**. Tức symbol-type của biến `content` (let-bound từ generic method call) bị 0 ở check pass khi type-table đủ lớn (thêm struct/Result-inst làm dịch/đụng type-id → multi-pass inference cache/ordering cho ra 0). Bug = **multi-pass type-inference instability cho biến let-bound = kết quả gọi generic method**, KHÔNG phải tổng-size.

**FIX (workaround tại stage1 source, commit ...):** thay `let content = r2.unwrap(); return content` → **`return r2.unwrap()`** (trả call trực tiếp, không qua biến trung gian). NodeCallExpr ở vị trí return được cache nodeTypes đúng (=12) nên check pass đọc đúng. Đã CHỨNG MINH: concat-đã-fix + struct + Result-fn build OK qua stage0 (trước đó fail). ⇒ **TRẦN NÂNG: stage1 lại có thể thêm struct/hàm/code** → mở khóa BUG#44 fix + import-bundler extensibility + mọi mở rộng stage1.

**QUY TẮC source stage1 (đến khi fix root Go sema):** tránh `let x = <generic-method-call như .unwrap()/.read_all()>` rồi dùng x ở chỗ cần type chính xác (return/so sánh); trả/dùng call TRỰC TIẾP. Nếu cần biến, cân nhắc annotation kiểu (chưa test kỹ). Root fix đúng = ổn định multi-pass inference cho let-bound generic-call trong Go sema (inference.go NodeVarDecl/NodeIdent + check pass TypeOf) — để phiên riêng; workaround này đủ mở khóa.

---

## 🟢 BUG#44 FIXED (2026-06-29) — callee-saved XMM6-15 save/restore

Sau khi nâng trần stage0 (commit 23188e5), thêm được code vào stage1 → land fix BUG#44.

**Cơ chế bug:** allocator cấp XMM8-15 (get_allocatable_xmms, phys 24-31) cho float vreg — đây là callee-saved theo win64 ABI — NHƯNG `get_used_callee_saved`/`reg_is_win64_callee_saved` chỉ xử lý GPR → prologue KHÔNG lưu XMM8-15 → callee clobber chúng → float giữ qua call hỏng. Repro `let a=3.0; noop(9.0); return (a*100) as i32` = 132 (a hỏng) thay vì 300&255=44.

**FIX (3 chỗ trong x86_regalloc.ax, design "treat XMM như GPR push"):**
1. `get_used_callee_saved`: mở rộng `used[]` 16→32 (cover XMM phys 16-31); thêm `if reg_is_xmm(phys) and phys >= REG_XMM6: used[phys]=true` cho win64; loop 0→32 → XMM callee-saved trả về SAU GPR trong cùng mảng `callee_saved`.
2. `emit_prologue`: trong loop callee_saved, nếu `reg_is_xmm(reg)` → `sub rsp,8` + `movsd [rbp-(i+1)*8], reg` (MACH_STORE padding=8, dst=RBP, src1=xmm, src2=IMM off) thay vì PUSH. Dùng rbp-relative (reuse encoding spill đã chạy, né rsp-base SIB).
3. `emit_epilogue`: reverse loop, nếu xmm → `movsd reg,[rbp-(i+1)*8]` (MACH_LOAD padding=8) + `add rsp,8` thay vì POP.

**Vì sao low-risk:** XMM được đếm trong `callee_saved_len` y như GPR push (mỗi slot 8 byte) → MỌI offset math (spill `-(callee_saved_len+...)*8`, str-lit `callee_saved_len*8+...`, `pushed_bytes=(callee_saved_len+2)*8`) tự đúng, KHÔNG cần đổi total_size hay thêm field StackFrame hay sửa insert_spill_code/emitter. Call site (x86_coff.ax/asm_emitter) không đổi.

**Verified:** repro canonical=44 (trước 132); 2 float=54; 3 float qua 2 call=155 — tất cả O1 đúng. Regression 43/43 (+ bin/t_b44.ax exit54). Fixpoint: stage3==stage4=fca6d684 (BẢO TOÀN).

**HỆ QUẢ:** float COMPOSE qua call giờ an toàn → pattern `CONST op f(call)` (acos=HALF_PI-asin(x), atanh=0.5*ln(...)) hết hỏng → mở khóa thư viện float diện rộng (trig/hyperbolic/special/complex/quaternion) cho task toán lớn. Quy tắc "tránh CONST op f(call)" trong std/math.ax có thể nới (nhưng leaf impl hiện tại vẫn đúng, giữ nguyên).

---

## ⚠️ NAMESPACING / symbol mangling (2026-06-29) — user fn trùng tên libc → collision

**Triệu chứng:** đặt hàm user tên `close` → exit 139 (segfault); tên `ok` → exit 127. Đổi tên (close→myclose, ok→mok) là hết.

**Root-cause:** user function emit symbol = TÊN THÔ. `resolve_binary_sym_name` (x86_coff.ax) chỉ redirect `alloc`→ax_alloc, `free`→ax_free, `std.string.len`→ax_str_len; mọi user fn khác `return name` (dòng 348) = tên thô KHÔNG mangle. → user fn dùng chung namespace symbol C toàn cục với runtime (ax_*.c) + CRT/libc. Trùng tên libc (close/open/read/write/...) → self-linker gộp/chọn nhầm → call vào libc (vd close(int fd) nhận bit f64 làm fd) → crash.

**Trả lời "cùng tên khác lib/alias có khác nhau không":** ĐÚNG VỀ NGUYÊN TẮC — với mangling module-qualified thì user `A.close` ≠ libc `close` ≠ user `B.close`. NHƯNG hiện CHƯA mangle → KHÔNG khác nhau (cùng symbol thô). Alias cũng không cứu (vẫn trỏ cùng symbol).

**FIX đúng (RFC-level, RỦI RO CAO — để phiên riêng):** mangle symbol user theo module-qualified (vd `ax_u__<module>__<fn>` hoặc hash signature) + whitelist `extern "C"` giữ tên thô. Đụng MỌI symbol → ảnh hưởng self-host fixpoint + linker + ranh giới ABI runtime → cần RFC + verify đầy đủ. **Workaround hiện tại:** đừng đặt hàm trùng tên libc/runtime (close/open/read/write/ok/...).

---

## 🔴 BUG#45 (2026-06-29) — struct 32-byte (4×f64) codegen sai nhiều kịch bản

Khi viết std.quaternion (Quat = 4×f64 = 32 byte) phát hiện 32-byte struct CÓ FIELD FLOAT bị miscompile ở nhiều pattern, dù **U256 (4×i64, 32 byte) HOẠT ĐỘNG** (t_bn256=31) và **Complex/Vec (≤24 byte, f64) HOẠT ĐỘNG** (t_complex/t_vec=63). → đặc thù **>16-byte struct + field f64**.

**Kịch bản ĐÚNG (đã kiểm):** return-only copy (`fn mk()->Quat`); 1 struct-arg→scalar (`sumq(a:Quat)->f64`=10); FREE-fn 2 struct-arg→struct (`q_mul(a,b:Quat)->Quat` với constructed args = 30 ĐÚNG).

**Kịch bản SAI:** 
1. METHOD `fn mul(self, o: Quat)->Quat` đọc field của `o` (32B f64 arg sau self) → ra 0 (q2/q7, cả O0/O1). (U256.add cùng shape với i64 thì ĐÚNG → float-specific.)
2. struct call-RESULT làm arg cho call khác (`qp=q_mul(q,p); q_mul(qp,...)`) → sai.
3. constructor field = call-result + sret (`q_from_axis_angle` trả Quat(w: cosf(h),...)) → q.w/q.z sai (tách cosf ra local KHÔNG cứu).

→ Quaternion/Mat3/Mat4 (cần 32B+ f64 struct) BỊ CHẶN. Complex(16B)/Vec3(24B) OK nên ship được. **FIX:** debug backend struct-ABI cho >16-byte f64 (param materialization của struct-arg-sau-self + float field load/store qua by-address ptr + sret tương tác float register). Cùng họ với Family C (struct by-value) nhưng cho FLOAT fields + size 32. Cần phiên backend riêng + fixpoint verify. **Workaround tạm:** giữ struct toán học ≤24 byte (≤3 f64) hoặc dùng free-fn với args constructed tươi (không call-result), tránh method-với-struct-arg cho 32B f64.

---

## ✅ BUG#46 (2026-06-29; FIXED — verified 2026-07-03) — float call-result đọc qua loop sau đó bị hỏng

`let m = f(...)` (m: f64, kết quả 1 lời gọi) rồi đọc `m` trong một VÒNG LẶP tiếp theo (loop KHÔNG có call) → m sai. VD st_variance: `let m = st_mean(a,n); while: d = a[i]-m; acc+=d*d` → variance sai; nhưng tính mean INLINE (sum loop tại chỗ, không call) → ĐÚNG. Nguyên nhân cùng họ BUG#48 (float-call-result bị spill; đường lưu dest-spill dùng R11/GPR thay vì XMM/movsd).

**FIXED bởi commit `476b352`** ("float call-result vregs mis-allocated to GPR pool"). **Verified 2026-07-03:** tái hiện pattern GỐC verbatim (`st_mean` là call riêng, rồi đọc `m` trong loop variance của dữ liệu [2,4,4,4,5,5,7,9]) → variance=4 ĐÚNG ở **cả -O0 và -O1**. Entry cũ (⚠️ open) đã STALE. Ghi chú: std/stats.ax vẫn tính mean inline (an toàn, không cần đổi).

---

## ✅ BUG#47 (2026-06-29) — unsigned `%`/`/` trong loop làm hỏng giá trị sống qua phép chia (FIXED)

**Triệu chứng:** Vòng trial-division `while i*i<=n: if n%i==0: return …; i=i+2` với u64 cho kết quả SAI khi loop chạy >1 vòng VÀ thân loop trả về HẰNG (`return 0`/`return false`/`return 99`). Trả về biến `i` thì ĐÚNG. → số hợp (91=7·13, 25=5·5) bị coi là nguyên tố. Lặp ở CẢ -O0 lẫn -O1 (codegen tất định, KHÔNG phải lỗi tối ưu).

**Root cause:** `x86_regalloc.ax` có 2 cơ chế cấm RAX/RDX quanh phép chia:
1. dòng ~357: cấm chính **toán hạng chia** (src1 của MACH_IDIV/MACH_DIV) — đã xử cả 2.
2. dòng ~428 (span pass): cấm RAX/RDX cho **mọi vreg có live-range bắc qua** phép chia (vd số bị chia `n`, hay biến vòng lặp) — vì `cqo`/`mov rdx,0` + `div` ghi đè RAX:RDX. **CHỖ NÀY CHỈ QUÉT `MACH_IDIV`** (đường ký hiệu/signed). Đường **unsigned** (u8/u16/u32/u64) phát `MACH_DIV`, nên span pass bỏ sót → một giá trị sống (vd `n`) nằm trong RDX bị `mov rdx,0; div` xoá → vòng sau đọc `n` sai. Khi thân trả về `i`, allocator giữ `i`/`n` ở thanh ghi an toàn khác nên không lộ; trả hằng thì `n` rơi vào RDX.

**Fix:** span pass quét cả `MACH_DIV`:
`if insts[di_scan].op == MACH_IDIV or insts[di_scan].op == MACH_DIV:` (x86_regalloc.ax ~L428). Chỉ THÊM thanh ghi bị cấm → bảo toàn/ngặt hơn, không đổi ngữ nghĩa hợp lệ. Verify repro: rc0/rc7 (return const) trong loop trial-division u64 đúng ở O0+O1; numtheory (gcd/lcm/is_prime/next_prime/mod_exp/totient/divisor_count) = 127.

**Ảnh hưởng:** rất phổ biến — mọi `%`/`/` unsigned có giá trị sống qua phép chia trong loop (hash, đổi cơ số, gcd-bằng-%, primality...). Là lỗi tất định ⇒ phải fixpoint verify (đổi codegen của chính compiler). Cùng họ BUG#34.1 (signextend div) và phần "spans idiv" cũ.

---

## ✅ BUG#48 (2026-06-29; FIXED — verified 2026-07-03) — float spill khi áp lực thanh ghi cao (≥8 float sống đồng thời)

**FIXED bởi commit `476b352`** ("float call-result vregs mis-allocated to GPR pool") — đúng root cause dest-spill-dùng-GPR mô tả bên dưới. **Verified 2026-07-03:** stress repro giữ **10** f64 call-result (`sq(c_i)`) sống ĐỒNG THỜI (vượt 8 XMM → ép ≥2 spill) rồi cộng tất cả → tổng ĐÚNG (110) ở **cả -O0 và -O1**; bản 8-live cũng đúng (204). ⚠️ Chưa tìm lại được FILE repro gốc "254/255" verbatim, nhưng cơ chế (float spill, giá trị đầu `a` được dùng lại sau spill) đã bị stress nhiều cách và đều đúng. Entry cũ (🔴 CHƯA fix) đã STALE.

**Triệu chứng (lịch sử):** ≥8 giá trị f64 (kết quả lời gọi) sống ĐỒNG THỜI vượt 8 thanh ghi XMM cấp phát (XMM8-15) → giá trị bị spill (vd `a` đầu tiên) bị HỎNG. Repro `bin/`-style: 8 `let x_i = sq(c_i)` rồi check → exit thiếu bit của `a` (254 thay 255); N≤5 ĐÚNG. Lặp O0+O1 (tất định). Cùng họ BUG#46 (float-call-result qua loop = cũng do spill).

**Root cause (một phần, xác định qua đọc x86_regalloc.ax):** đường lưu **DEST bị spill** (~L929-958) LUÔN dùng **R11 (GPR) + mov 8-byte** cho mọi dest, kể cả float → `mov r11, xmm` sai register-file → giá trị float hỏng. Đường RELOAD operand spill (~L887-915) ĐÃ xử đúng float (XMM0/XMM1 + movsd) nhưng đường STORE dest thì CHƯA.

**Fix thử (CHƯA đủ, đã revert):** thêm nhánh float ở dest-spill dùng XMM2 + movsd (padding 8). Build lại nhưng repro VẪN 254 → còn nguyên nhân nữa: nghi `is_float_vreg(dst_alloc.vreg)` trả false cho vreg bắt return-value (MOV bắt XMM0 có type_id=0 không phải 10), HOẶC reload tại điểm dùng cũng cần sửa. Cần điều tra thêm + sửa CẢ store lẫn nhận diện float-vreg, rồi fixpoint verify (~3h). Đã revert để giữ stage1 ở mốc xanh 5c398609.

**Ảnh hưởng/độ ưu tiên:** HẸP — cần ≥8 float-local sống đồng thời (hiếm; thư viện toán hiện tại KHÔNG dính: hàm dùng đối-số-trực-tiếp / ít local). Là KEYSTONE backend float (gộp với BUG#45 struct>16B-f64, BUG#46 float-qua-loop) cho phiên chuyên sâu. **std.math đã KIỂM CHỨNG CHÍNH XÁC** (tests/mathlib, sai số <1e-5) — bug này là codegen áp-lực-thanh-ghi, không phải sai công thức.

**Quy tắc tạm cho thư viện/test toán:** tránh giữ ≥8 float-call-result trong local đồng thời; dùng giá trị ở vị trí đối số hoặc tính lại; verify bằng probe đối-số-trực-tiếp (tests/mathlib).

---

## 🟢 BUG#49 (2026-07-02) — Function pointers / higher-order functions (native) + 2 lỗi optimizer -O1

**Bối cảnh:** mở khóa #25 (numerical analysis generic: truyền hàm `f` bất kỳ vào quadrature/root-find) và higher-order tổng quát. Con trỏ hàm = giá trị địa chỉ hàm, gọi gián tiếp `MACH_CALL_INDIRECT`.

### Phần A — Front-end đã có, thiếu 1 mắt xích `lower_ident`
- Parser/typecheck/AIR đã hiểu type `fn(...)->T` và tham số con-trỏ-hàm. Nhưng `lower_ident` (air_builder.ax) khi gặp `SYM_FUNC` (tên hàm trần dùng như GIÁ TRỊ, vd `let f = add` hoặc `apply(add, ...)`) trả 0 → gọi rác → exit 127.
- **Fix:** thêm nhánh `SYM_FUNC` → emit **`OP_FUNC_ADDR` (0x030D)** với `src1 = sym_idx` (chỉ số symbol hàm — **là IMMEDIATE, không phải vreg**), `type_id=4` (i64), `dest = faddr`.

### Phần B — Native backend cho OP_FUNC_ADDR + gọi gián tiếp (keystone = ASLR)
- **cgen.ax (C backend):** `r_%d = (void*)&%s;` qua `get_mangled_name_by_sym(inst.src1)`; thêm scan `max_reg` bỏ qua src1/src2 cho OP_FUNC_ADDR.
- **x86_selector.ax:** (a) OP_FUNC_ADDR → `MACH_MOV_IMM dst, vreg=3 imm=sym_idx` (địa chỉ hàm RIP-relative). (b) Ở nhánh call thường: nếu `inst.src1 != 0` (gọi gián tiếp) → **bắt target vào R11 TRƯỚC khi marshaling arg** (nếu không, arg-setup ghi đè vreg target → segfault). (c) Khi phát call: `src1 != 0` → `MACH_CALL_INDIRECT` (FF/2) qua R11, ngược lại `MACH_CALL`.
- **x86_emitter.ax:** (a) `MACH_CALL_INDIRECT` → `x86_encode_call_r(treg)`. (b) `MACH_MOV_IMM` vreg==3 → **lea RIP-relative + RELOC_PC32** (sym_name = sym_idx, addend −4). **ASLR keystone:** PE có `DllCharacteristics=0x0160` (DYNAMICBASE). `movabs` tuyệt đối vào địa chỉ hàm bị SAI khi loader rebase ASLR → segfault 139. Phải dùng **lea RIP-relative (RELOC_PC32)** — an toàn ASLR — thay cho movabs tuyệt đối.

### Phần C — 2 lỗi optimizer -O1 làm hỏng `let f = add; f(...)` (biến con-trỏ-hàm local)
Triệu chứng: `let f=add; return f(2,3)` chạy đúng ở **-O0 (=5)** nhưng **crash 139 ở -O1** (higher-order dạng THAM SỐ `apply(add,...)` thì -O1 vẫn OK — vì hàm là ARG). Data-dependent. Hai nguyên nhân trong `ssa_opt.ax`:

1. **Copy-prop làm hỏng IMMEDIATE (giống BUG#15).** `OP_FUNC_ADDR.src1` là *symbol index* (immediate), KHÔNG phải vreg. Vòng copy-prop (~L330) rewrite `inst.src1` qua `copy_map` cho MỌI opcode (trừ JUMP). Khi sym_idx **trùng số** một vreg đang trong copy-chain (`let f=add` làm đổi cách đánh số vreg → dễ trùng) → symbol index bị ghi đè → **sai địa chỉ hàm**. Fix: thêm `and inst.opcode != OP_FUNC_ADDR` vào guard — y hệt cách GET_FIELD/SET_FIELD.src2 được loại trừ.

2. **DCE xóa nhầm định nghĩa func_addr.** Trong pass đếm use (~L517), block `OP_CALL` chỉ đếm **args** (từ `extras`) rồi `continue` — **KHÔNG đếm `inst.src1`** (target vreg của gọi gián tiếp). `OP_FUNC_ADDR` không có side-effect ⇒ nếu use_count[target]=0, DCE biến nó thành NOP → call qua vreg rác → **crash 139**. (Vì sao `apply(add,...)` không dính: hàm là ARG nên được đếm ở nhánh arg; target của indirect-call bên trong `apply` là THAM SỐ, không do OP_FUNC_ADDR định nghĩa.) Fix: trong block OP_CALL, thêm `if inst.src1 != 0: use_count[inst.src1] += 1`.

**Cả 2 fix đều strictly conservative:** #1 chỉ *chặn* một propagation (không thêm), #2 chỉ *tăng* use-count (không xóa) → không thể phá code đúng. CSE (`cse_func`, ~L1450) chỉ xử binary-ALU + NEG/NOT nên bỏ qua OP_FUNC_ADDR — an toàn.

### Kết quả verify (binary rebuild với cả 2 fix)
| test | O0 | O1 |
|---|---|---|
| fp0i (`let f=add; f(2,3)`) | 5 ✓ | **5 ✓** (trước = 139 crash) |
| fpBi (`apply(add,2,3)` higher-order param) | 5 ✓ | 5 ✓ |
| fpf (`sum3(sq)` higher-order) | 14 ✓ | 14 ✓ |

Đã đăng ký `fp0i`/`fpBi`/`fpf` vào `scripts/regression_repros.sh`. Full regression (native -O1) + fixpoint verify (`verify_bug29_selfhost.sh`) đang chạy trước khi commit.

### ⚠️ Bẫy chẩn đoán — STATUS_ENTRYPOINT_NOT_FOUND KHÔNG phải bug "no-import" (ĐÍNH CHÍNH 2026-07-03)
Ban đầu tưởng: `plain5` (`pub fn main()->i32: return 5`, KHÔNG `extern`) self-link native → **STATUS_ENTRYPOINT_NOT_FOUND (0xC0000139 = −1073741511)** là "quirk program không-import". **SAI.** Đã kiểm lại trên máy KHỎE: no-import program (plain return / direct call / local fn-ptr) đều CHẠY ĐÚNG ở O0+O1 qua bash/cmd/PowerShell (=5). Triệu chứng cũ chỉ xuất hiện khi máy **quá tải nặng** (build trivial 84-112s do browser/thrash) → self-linker sinh PE hỏng dưới áp lực bộ nhớ = **artifact resource-starvation**, KHÔNG phải lỗi xử-lý-no-import. ⇒ Thấy STATUS_ENTRYPOINT_NOT_FOUND thì **kiểm tải hệ thống trước**, đừng đuổi "no-import bug". (Follow-up robustness khả dĩ: self-linker nên fail-loud thay vì phát PE hỏng khi alloc/write thất bại.)
- **git-bash `$?` là 8-bit:** 0xC0000139 & 0xFF = **57**, không phải giá trị return thật. Lấy exit code thật của Windows qua PowerShell `$LASTEXITCODE` hoặc `cmd //c`.

**Files:** air.ax (OP_FUNC_ADDR const), air_builder.ax (lower_ident SYM_FUNC), cgen.ax, x86_selector.ax, x86_emitter.ax, ssa_opt.ax (2 fix). **Blast radius:** `lower_ident SYM_FUNC` fire cho MỌI tên hàm trần trong toàn codebase (kể cả stdlib) → bắt buộc regression + fixpoint.

---

## 🟡 KNOWN-GAP (2026-07-03) — Selective/capability import `import X { a, b }` parse-được-nhưng-trơ

**Câu hỏi khơi nguồn:** "2 thư viện cùng có hàm `close()`, import cả hai, dùng chung một chỗ — có gây lỗi không?"

**Trả lời:** KHÔNG va chạm, nhưng vì một lý do đáng lưu ý: **danh sách trong ngoặc `{...}` hiện là cú pháp chết.**

### Mô hình import thực tế của AXIOM (module-qualified)
- `import lib_a` chỉ tạo **một** symbol `SYM_MODULE` tên `lib_a` (resolver.ax:716-753). KHÔNG đổ hàm của thư viện vào namespace người dùng dưới tên trần.
- Truy cập thành viên qua `lib_a.close()` (`NODE_FIELD_EXPR`): resolver phân giải `lib_a`→module rồi gọi `lazy_resolver_resolve_field(lib_a, close)` — **chỉ duyệt bảng export của đúng module đó** (resolver.ax:1023-1065). `lib_a.close` và `lib_b.close` đi qua hai bảng export khác nhau → hai symbol riêng biệt → không va chạm.
- **Không có glob-import** (`import *`) trong ngôn ngữ. Nên tên trần của thư viện không bao giờ bị trộn vào scope người dùng.
- Va chạm THẬT chỉ khi hai hàm cùng tên + cùng chữ ký định nghĩa trong **cùng một file/scope**: `define()` (resolver.ax:517-538) dựng **overload chain** (`next_overload`), phân giải theo kiểu tham số ở call-site; `resolve()` trả đầu chuỗi.

### Cú pháp `import X { a, b, c }` — spec vs implementation
- **Spec:** grammar chính `ImportDecl ::= "import" Identifier ["as" Identifier]`. Ngoài ra spec §permissions định nghĩa `import std.fs { read, write }` = **khai báo quyền tĩnh (capability)** để compiler theo dõi trên AST và chặn build nếu module bên thứ 3 dùng quyền chưa khai báo. ĐÂY KHÔNG PHẢI selective-name-import kiểu Python.
- **Parser (parser.ax:1445-1460):** CÓ đọc `{ a, b, c }`, gắn mỗi tên thành `NODE_IDENT` con của node import.
- **Nhưng toàn pipeline chỉ có ĐÚNG một chỗ chạm `NODE_IMPORT_DECL` = resolver.ax:716**, và handler đó **không duyệt các con**. typecheck/air_builder không tham chiếu. Không có capability/permission pass nào tồn tại (grep=rỗng).
- ⟹ Danh sách `{...}` bị **parse rồi vứt (inert)**. Không bind tên trần, không kiểm quyền. Trơ hoàn toàn.

### Hệ quả & việc cần làm sau
- Hiện tại an toàn (không va chạm) nhưng là **bất nhất tiềm ẩn**: tính năng bảo mật capability của spec chưa được nối dây. Nếu sau này hiện thực (capability-check HOẶC selective-bind tên trần), lúc đó **mới** cần chính sách va chạm rõ ràng (báo lỗi ambiguous vs overload). Việc này đụng semantic import → cần **RFC** trước khi làm.
- **Thực nghiệm cô lập chưa hoàn tất:** build thử `lib_a.close(0)+lib_b.close(0)` (kỳ vọng 101) từng cho kết quả rác (2/99/NO-EXE) NHƯNG bị nhiễu do regression chạy song song giẫm lên file trung gian tên-cố-định `axiom_temp.obj` (build đồng thời không reentrant). Cần build cô lập (không concurrent) để xác nhận `lib_a.close` vs `lib_b.close` có map đúng hai symbol khác nhau qua concatenation không. **TODO: chạy lại khi máy rảnh.**
- **Phụ đề — non-reentrant temp:** compiler ghi obj trung gian ra tên cố định `axiom_temp.obj` ở cwd → **không thể build song song nhiều tiến trình cùng cwd**. Ảnh hưởng CI/parallel build. Đáng cân nhắc đặt tên tạm theo output/PID.

---

## 🟢 BUG#50 (2026-07-03) — "First import wins": hàm cùng tên ở 2 module gộp về module import đầu

**Repro:** `lib_a.close(x)=x+1`, `lib_b.close(x)=x+100`; `lib_a.close(0)+lib_b.close(0)` → 2 (cả hai chạy lib_a); đảo thứ tự import → 200. Thậm chí `lib_b.close(0)` một mình → 1. Silent, không diagnostic. Repro: bin/t_modcollide.ax (+ lib_a.ax, lib_b.ax ở root; oracle=101).

**Root cause (2 tầng):**
1. **Resolver dedup sai:** `define()` (resolver.ax) dedup overload bằng `decl_node == decl_node` — nhưng decl_node là node index CỤC BỘ THEO CÂY của từng module. Hai module cấu trúc giống nhau → close của lib_b trùng index với lib_a → bị coi là cùng declaration → trả symbol lib_a. **Fix:** so thêm `symbol_trees.data[curr_idx] == current_tree` (bảng cây-của-symbol đã có sẵn).
2. **Mangling trùng:** cả hai backend (cgen `get_mangled_name_by_sym`, native `x86_resolve_sym_name`) sinh tên từ bare `sym.name_id` → cả hai hàm = `ax_close` → link-time gộp về định nghĩa đầu. **Fix:** flag `SYM_FLAG_MODDUP=2048` set tại define() khi chain-head thuộc cây khác; mangler thấy flag → emit `ax_<name>__m<sym_idx>` (deterministic; cả 2 backend cùng quy tắc).

**Verify:** oracle 4/4 (1/100/101/101 hai chiều import) trên native-built; C-backend inspect thấy `ax_close` + `ax_close__m187` với call-site khớp; full regression 92/93 (fail duy nhất = t_mathx = BUG#51 pre-existing, chứng minh có/không fix đều crash); fast fixpoint a3==a4 converged. Registered `t_modcollide|exit|101` (⚠️ flaky vì BUG#52 cho tới khi BUG#52 fixed).

**Còn lại (không thuộc scope này):** bare-name call `close()` khi đã import module có hàm cùng tên → vẫn resolve về chain head (cần diagnostic ambiguity); intra-module bare call trong module bị trùng tên chưa được bảo vệ.

---

## 🔴 BUG#51 (2026-07-03) — Native-built compiler MISCOMPILED: segfault deterministic khi compile t_mathx

**Phát hiện chấn động:** mọi binary compiler build bằng NATIVE backend (stage2/stage3/converged chains) **segfault 4/4 deterministic** khi `build bin/t_mathx.ax -O1`; binary build bằng stage0 C-backend/gcc compile t_mathx OK (=28). Cùng source (fixpoint!). ⟹ native backend miscompile ≥1 hàm của chính compiler; chỉ lộ khi compile t_mathx (pattern đặc thù).

**Hệ quả cực quan trọng:** **SHA fixpoint convergence ≠ correctness** — stage3==stage4 hội tụ trên code SAI đồng nhất. Runtime check hiện tại (t_param5/t_strip) quá yếu. **Quy trình mới bắt buộc: chạy FULL regression suite bằng chính converged binary** (không chỉ SHA + 2 test).

**Ý nghĩa hiệu năng:** native-built compiler nhanh hơn stage0-built ~1200x (8s vs 2h50m cho 790 hàm self-build) — là daily-driver lý tưởng NHƯNG bị block bởi bug này. Fix BUG#51 = mở khóa vòng dev 8 giây.

**Trạng thái:** OPEN — cần binary-search hàm bị miscompile (build từng phần / chèn probe, so stage1-built vs stage3-built hành vi từng function).

---

## 🔴 BUG#52 (2026-07-03) — Lazy module loading: memory corruption PHI-TẤT-ĐỊNH (works/crash/hang ngẫu nhiên)

**Repro:** cùng binary stage1, cùng input `tcollide.ax` (import 2 module local), chạy 4 lần liên tiếp: OK(2), OK(2), crash(127), **HANG** (>2min). ASLR/allocator-layout dependent.

**Nghi phạm:** cleanup cuối `ax_driver_load_module` (main_air.ax ~1085-1094) free lexer arrays trong khi ModuleInfo giữ con trỏ (`tokens_leak`/`tokens_data`, node_types của checker...); hoặc temp_stack swap/restore; hoặc export-loop đọc str đã free. Use-after-free biểu hiện tùy layout.

**Hệ quả:** MỌI test có import module local = flaky (t_modcollide sẽ chập chờn trong regression cho tới khi fix). Vi phạm nguyên tắc deterministic compilation (CLAUDE.md §3). Đã đốt ~1h debug BUG#50 vì crash giả từ bug này. **Bài học chẩn đoán: crash không tái lập ổn định → chạy lặp ≥5 lần TRƯỚC khi đổ cho thay đổi mới.**

**Trạng thái:** OPEN — ưu tiên cao nhất cùng BUG#51.

---

## 🟢 BUG#51 + BUG#52 (2026-07-03) — Stale aggregate-alias qua vec-grow: native-built compiler segfault deterministic (t_mathx) + module-load flaky

**Triệu chứng:**
- #51: MỌI binary compiler build bằng native backend (stage2/3, a2/a3) segfault **deterministic** khi compile `bin/t_mathx.ax` (cả -O0 lẫn -O1); binary build bằng stage0/gcc compile OK. SHA fixpoint vẫn hội tụ (bug tự tái tạo y hệt) ⟹ **fixpoint ≠ correctness**.
- #52: test import local module (t_modcollide) **phi-tất-định** — cùng binary+input lúc OK lúc crash/hang.

**Hunt (kỹ thuật đáng tái dùng):**
1. Bisect input: crash chỉ khi main() đủ lớn; **non-monotonic** ở -O1, sạch ở -O0; thêm 1 hàm đệm cũng crash ⟹ **size-threshold, không phải construct** — dấu hiệu grow-boundary.
2. gdb: crash `movzbq (%r10)` với r10 = pointer heap-thấp "hợp lệ-nhìn-như-thật" (0x535b488) load từ spill slot; frame 18KB = mega-fn `infer_node` (161KB code, spill-all fallback).
3. **Instrument-probe loop <1 phút**: chèn `ax_printf_local + fflush` vào typecheck.ax → rebuild concat → build bằng converged-binary (8s) → chạy repro. Build instrument ở **-O0** để thứ tự trung thực (probe -O1 bị optimizer reorder qua → bracket sai).
4. Probe chốt: crash ĐÚNG tại lệnh đọc `callee_node.kind` (BUG#45 guard, NODE_CALL handler).

**Root cause:** `let callee_node = self.tree.nodes.data[callee]` — native backend giữ aggregate **BY-ADDRESS** (x86_selector OP_INDEX, comment "Aggregates are held by-address in this backend") ⟹ `callee_node` = ALIAS vào nodes buffer. Generic instantiation (mono, ~L1612) GROW nodes vec → realloc + **@free buffer cũ** → alias dangling. Vì sao gcc-built sống: malloc giữ trang cũ mapped + nội dung nguyên (memcpy khi grow) → đọc stale trả **giá trị đúng y hệt** → chạy "đúng". Native: block lớn (~512KB) → ax_free → **VirtualFree(MEM_RELEASE)** unmap → segfault. #52 cùng cơ chế (module load grow vecs; block size dao động quanh ngưỡng free-list vs VirtualFree → phi-tất-định).

**Fix (targeted, conservative):** typecheck.ax NODE_CALL handler — thay các use `callee_node` SAU điểm instantiation bằng fresh read `self.tree.nodes.data[callee].kind/payload/first_child` (2 vùng: else-branch syscall/panic + BUG#45 guard). Các site instantiate khác (L1957/L2282) đã dùng fresh reads sẵn.

**Verify:** t_mathx = build OK + run 28/28 ×3 deterministic; t_modcollide 10/10 (hết flaky — #52 đóng cùng); fast-fixpoint n2==n3 OK; full regression bằng **binary native** (chuẩn mới).

**⚠️ Landmine class còn lại:** MỌI `let s = arr[i]` (aggregate) giữ qua bất kỳ grow nào của vec đó = bom nổ chậm. Backend divergence: cgen COPY (C `r = arr[i]`) vs native ALIAS — hai ngữ nghĩa khác nhau cho cùng AIR! Cần fix hệ thống: **RFC 0010 — value semantics cho aggregate index-load** (emit block-copy vào stack slot). Cân nhắc: code hiện tại có thể VÔ TÌNH dựa alias (mut local ghi xuyên về array) — RFC phải audit trước khi đổi.

**Hệ quả chiến lược:** mở khóa **daily-driver native** (self-build 790 hàm = 8s vs 2h50m stage0-built = 1200x) — perf root-cause thật của "compile chậm" là binary stage0-built chậm, không phải thuật toán compiler.

---

## PERF #1 — scratch vreg `inst.dest + 60000` thổi phồng `graph_size` → regalloc chiếm ~90% codegen ✅ ĐÃ FIX (ee139a2)

**Triệu chứng:** codegen chậm bất thường. `--time` cho thấy t_mathx -O1: codegen 2074ms/2.5s tổng (~80%). Probe sâu: `allocate_registers_orchestrator` = ~2008ms; trong đó `graph_coloring_alloc` build interference + mọi vòng O(graph_size). `--time` là bạn — luôn đo trước khi đoán.

**Root cause:** `x86_selector.ax` lower **OP_FCONST** (float const) và **OP_NEG** (float negate) cấp GPR-scratch = `inst.dest + 60000` — hack để đảm bảo vreg "không bao giờ là AIR dest" (⇒ `is_float_vreg()`=false ⇒ cấp GPR; còn `inst.dest` f-typed ⇒ XMM). Hệ quả: **mọi hàm đụng hằng float** (≈ mọi hàm toán) có 1 vreg ≈ 60000 ⇒ trong `graph_coloring_alloc`, `max_vreg` scan từ insts thấy 60000 ⇒ `graph_size = max_vreg+1 ≈ 60000` DÙ chỉ vài chục giá trị sống. Mọi `@alloc(graph_size*…)` + init-loop + free-loop + `new_u32_vec()×graph_size` chạy 60000 vòng/hàm. Đo: GS=60244 cho hàm chỉ V=250; GS=60056 cho hàm V=33.

**Fix:** thay `inst.dest + 60000` → `next_vreg(sel)` (2 site). Cho vreg **dense** (max_vreg+1) cũng không phải AIR dest ⇒ phân loại GPR/XMM/16-byte **y hệt** (byte-for-byte), nhưng graph_size xẹp về số vreg thật.

**Kết quả (t_mathx -O1, binary native đã promote):** codegen **2074ms → 126ms (~16×)**; tổng build **~2.5s → ~0.34s (~7×)**; graph_size **~60000 → 13-95**/hàm.

**Verify (backend ⇒ fixpoint BẮT BUỘC):** self-build ×2 SHA giống hệt (72c9aefc…) = fixpoint; full regression **93/93** bằng binary mới; t_mathx output không đổi (exit=28).

**Bài học tái dùng:** (1) hack "unique-vreg-bằng-offset-lớn" là **anti-pattern** — mọi array/loop keyed theo max_vreg trả giá; luôn dùng `next_vreg`/counter dense. (2) ssa_opt cũng dùng `max_reg+1` arrays — nếu tương lai lại có vreg thưa-cao thì cùng dính. (3) quy trình probe: `--time` (pha) → clock() bọc từng sub-phase trong vòng codegen → in min/max/graph_size để phát hiện thưa → truy ngược tới nơi sinh vreg lạc (quét result MachInsts theo op).

---

## PERF #2 — `is_float_vreg` linear-scan trong split-loop = O(V×N) trên mega-function ✅ ĐÃ FIX (bb58f18)

**Triệu chứng:** sau PERF#1, compiler tự-build vẫn codegen 9560ms. Probe per-func (`fn-cost` in hàm >80ms): tập trung ở vài mega-function THẬT (không phải artifact): `select_inst` insts=27071/spills=12882 = **3045ms** (select_all 1341 + regalloc 2076), `infer_node` 660ms, `translate_inst` 524ms, `emit_inst` 174ms.

**Root cause:** `allocate_registers_orchestrator` split từng live-interval thành bucket GPR/XMM bằng `is_float_vreg()` — hàm này **linear-scan `fn_ptr.insts`** tìm inst định nghĩa vreg. Gọi 1 lần/interval ⇒ **O(intervals × insts)**. select_inst (~13k vreg × ~13k inst) = ~2s chỉ để tra cùng 1 thứ lặp lại. (Lưu ý: hàm insts>5000 đi **spill-all fallback** trong graph_coloring, return sớm TRƯỚC interference — nên chi phí KHÔNG ở graph coloring mà ở split-loop + prescan.)

**Fix:** `build_vreg_def_idx(fn_ptr, size)` = 1 lượt O(insts) map `vreg → index inst def đầu tiên` (khớp ngữ nghĩa first-match của is_float_vreg), + `is_float_vreg_cached(...def_idx...)` = **cùng logic phân loại per-inst** sau lookup O(1) (KHÔNG nhân đôi logic ⇒ byte-identical). Split-loop: O(intervals×insts) → O(insts).

**Kết quả (full compiler self-build -O1, native đã promote):** codegen **9560ms → 4452ms (~2.1×)**; select_inst regalloc-phase **2076ms → 689ms**. Fixpoint SHA=c5316008 + regression 93/93.

---

## PERF #3 — `regalloc_is_16byte` linear-scan (spill-all + insert_spill_code) = O(V×N) ✅ ĐÃ FIX (cc9acf9)

**Root cause:** giống PERF#2 nhưng cho `regalloc_is_16byte` (x86_selector.ax, ~120 dòng) — linear-scan `fn_ptr.insts` tìm def, gọi per-interval trong spill-all fallback (x86_regalloc.ax:482) + per-spilled-operand trong insert_spill_code (886/929/956). ~689ms residual của select_inst.

**Fix (KEY INSIGHT — an toàn hơn PERF#2):** KHÔNG nhân đôi 120 dòng (có đệ quy COPY/MOVE→src1, inner-scan-2 OP_DEREF, control-flow-skip). Thay bằng **wrapper mỏng giữ signature cũ** (16 caller nguyên, truyền null→linear) + `regalloc_is_16byte_cached(...def_idx, def_size)` = **cùng body**, chỉ đổi 1 thứ: **vòng scan START ở `def_idx[vreg]`** thay vì 0. Mọi index [0, def_idx[vreg]) có dest≠vreg ⇒ bỏ chúng là **chứng minh được identical** — thân vòng (kể cả cf-skip) KHÔNG đụng, chỉ bỏ no-op prefix. `def_idx[vreg]==-1` ⇒ return false luôn. Đệ quy + inner-scan-2 cũng jump-start cùng def_idx. **Đây là pattern chuẩn để cache hàm scan-tìm-def phức tạp mà không rủi ro phân kỳ: đừng replicate logic — chỉ nhảy tới index bắt đầu.**

**Kết quả:** codegen **4452ms → 2396ms** (riêng change này); **9560ms → 2396ms (~4×)** cộng dồn PERF#2. Fixpoint SHA=13ac5d05 + regression 93/93.

## PERF #4 — `get_register_type` linear-scan trong selection = O(insts × N) ✅ ĐÃ FIX (1411787)

**Root cause:** `get_register_type` (x86_selector.ax) linear-scan `sel.fn_ptr.insts` tìm def của reg; gọi ~25 lần trong `select_inst`, nhiều lần/inst lowered ⇒ quadratic (~747ms select_inst trong select_all).

**Fix:** cache def_idx **trên struct `InstructionSelector`** (`def_idx: ptr[i32]`, `def_size: i64`) — đã threaded khắp selection qua `sel`. Build 1 lần trong `select_all` (sized max AIR dest+1), get_register_type jump-start tại `def_idx[reg]` (cùng pattern PERF#3). Chỉ 1 nơi khởi tạo InstructionSelector nên thêm field an toàn.

**Kết quả:** codegen **2396→1442ms**. Fixpoint 966dfb09 + 93/93.

**Tổng kết perf session (2026-07-04) — 4 fix:** codegen compiler-self-build **9560ms → 1442ms (~6.6×)**; float-heavy t_mathx **2074→126ms (~16×)**; tổng self-build wall-clock **~14.8s → ~4.1s (~3.6×)**. Kỹ thuật xuyên suốt: `--time` → probe per-func/per-sub-phase (clock bracket) → phát hiện quadratic (in V/GS/spills/min-max) → memoize bằng def_idx (1 lượt O(N), jump-to-def-index) thay linear-scan-per-vreg. **Nguyên tắc: hàm scan-tìm-def gọi trong vòng O(V) = quadratic ẩn; luôn nghĩ tới def_idx.** Sau PERF#1-4: 2 phase lớn nhất = typecheck 1233ms + codegen 1336ms (residual lành mạnh); typecheck chưa đụng (frontend, hotspot kế tiềm năng).
