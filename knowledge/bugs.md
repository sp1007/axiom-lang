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
