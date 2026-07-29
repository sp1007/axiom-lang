---
name: session-handoff-2026-07-29f
description: "HANDOFF 2026-07-29f — HEAD sạch, driver B==C 3A245462, 558/558. 4 commit đã ship. Việc TIẾP THEO đã xác định và đã ĐỊNH GIÁ: coalescing cặp copy dạng 2-toán-hạng, định giá trên xorshift (89,8 ms) chứ KHÔNG phải fib."
metadata:
  type: project
---

# HANDOFF 2026-07-29f — **ĐỌC ĐẦU TIÊN**

## Trạng thái
- HEAD `7ebbfb9`, cây sạch (chỉ còn `.claude/settings.json` untracked — **của user, đừng đụng**).
- Daily driver `bin/axc_native.exe` = binary **B==C `3A245462`**. Regression **558/558**,
  ELF 12/12, ctgc 16/16, exe_size 4/4, lib_collision 6/6, so_export ✓.
- Đã ship trong phiên: `7f3aee3` RFC 0035 P2 (symbol thư viện mang tên module),
  `bc9bf35` guard flag-day cho `.lib`, `6c34374` LEA fold + xoá hằng chết,
  `7ebbfb9` định giá xorshift.

## VIỆC TIẾP THEO (đã chọn, đã định giá, chưa bắt đầu code)
**Coalescing cặp copy của dạng 2-toán-hạng phá huỷ**, định giá trên **xorshift**.

Trần lợi ích đã ĐO: **89,8 ms / 30% chương trình** (W1 213,7 → W2 303,4 —
xem [[m6-perf-baseline]] mục 2026-07-29f). Đây là 95% khe hở 1,42x của xorshift.

Mục tiêu cụ thể, mỗi bước xorshift AXIOM phát:
```
mov %rax,%rdx ; shl $0xd,%rdx ; mov %rax,%rbx ; xor %rdx,%rbx ; mov %rbx,%rax
```
NASM floor chỉ cần: `mov rdx,rax ; shl rdx,13 ; xor rax,rdx` (dst == src1).
⇒ Phải bỏ được **`mov %rax,%rbx`** (copy vào dest) và **`mov %rbx,%rax`** (copy trả về).

### Điều tra đang dở (dừng ở đây, chưa kết luận)
Bias coalescing **đã tồn tại** trong `x86_regalloc.ax`, hai nửa:
- `move_partner[]` (`:577` khởi tạo, `:621` ghi khi gặp `MACH_MOV` vreg←vreg) —
  **first-move-wins, MỘT đối tác duy nhất**;
- bỏ cạnh interference cho cặp partner khi live range chỉ **chạm biên** (`:805-818`);
- ưu tiên lấy màu của partner khi tô màu (`:948-970`).

**Câu hỏi chưa trả lời** (bắt đầu từ đây): vì sao 2 copy trên KHÔNG được coalesce?
Giả thuyết cần kiểm chứng bằng probe, ĐỪNG suy luận:
1. `dest` và `src1` có thực sự **không** interfere không? (x cũ chết ngay sau `xor` ⇒ về lý
   thuyết hợp lệ). Nếu interference bị ghi nhầm thì lỗi ở mô hình interval `iv1.end <= iv2.start`.
2. `move_partner` **first-wins**: rax có thể đã bị "đặt chỗ" bởi một cặp khác trước đó ⇒
   cặp cần thiết không bao giờ được ghi. Đây là hạn chế ĐÃ BIẾT ("chỉ bias 1 đối tác").
3. Chuỗi `x = u` ở tầng AIR có thể sinh copy thứ hai mà bias không nối được.

### ⛔ Cảnh báo trước khi làm
- **ĐỪNG định giá trên fib.** Precolored-bias từng cài rồi REVERT vì fib chậm hơn 1,5–2% —
  nhưng fib **latency-bound** nên copy vô hình; xorshift **throughput-bound** mới là shape đúng.
  Đây là đính chính trung tâm của phiên này.
- **ĐỪNG cài loop rotation / bottom-test** — đã đo **+0,1%**, vô giá trị.
- Mọi số perf phải đo **GHÉP CẶP xen kẽ** trong cùng phiên, các vòng **không chồng lấp**.
  Một lần chạy `perf_fib` đơn lẻ ở máy này lệch 8–10% và đã suýt làm ship một regression.
- Backend ⇒ **B==C bắt buộc** trước commit (A!=B là bình thường khi codegen đổi).

## Backlog còn lại (sau việc trên)
- RFC 0035: method/global/ctor vẫn dùng scheme cũ (`axS_`/`axG_`/`axC_` chưa làm) ⇒ fn-vs-struct
  vẫn dựa reject ở typecheck. P3 (E0501 → error) vẫn bị chặn bởi shim runtime trùng lặp hợp lệ.
- `mod_name` rỗng ở `register_module_from_lib` (binding `mod.NAME` là đồ chết) — vô hại, nhưng
  là bẫy cho người đọc.
- Module path nhiều đoạn không resolve ở call site (`bin.libcol.liba.helper()`) — có sẵn từ trước.
- Nếu hết việc: `axiom-bug-probe` để nạp lại backlog.
