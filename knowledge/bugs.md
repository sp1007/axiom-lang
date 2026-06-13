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
