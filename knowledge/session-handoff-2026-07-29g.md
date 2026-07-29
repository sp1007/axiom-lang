---
name: session-handoff-2026-07-29g
description: "HANDOFF 2026-07-29g — HEAD 6141531 sạch+pushed, driver seed==A==B 2077495B, 561/561. Coalescing ĐÃ SHIP: xorshift −30,2%, bằng asm floor. Việc kế tiếp: đo lại toàn bộ perf_suite để xem mốc M6-codegen (≤15% so ASM floor) đã ĐẠT chưa."
metadata:
  type: project
---

# HANDOFF 2026-07-29g — **ĐỌC ĐẦU TIÊN**

## Trạng thái
- HEAD `6141531`, đã **push lên `origin/main`**, cây sạch (chỉ `.claude/settings.json`
  untracked — **của user, đừng đụng**).
- Daily driver `bin/axc_native.exe` = **`2077495B`**, và nó **tự tái tạo**: fast_fixpoint cho
  **seed == A == B**, cộng thêm **B == C** đã chạy riêng trước đó.
- Gate đầy đủ XANH: regression **561/561** (+3 dòng `t_coalescedest`), ELF 12/12, ctgc 16/16,
  exe_size 4/4, lib_collision 6/6, so_export ✓.

## Đã ship phiên này (một commit)
`6141531` **coalesce_dest_copy** — fold cặp copy của dạng 2-toán-hạng phá huỷ.
Chi tiết đầy đủ + bài học ở [[m6-perf-baseline]] mục 2026-07-29g. Tóm tắt:
- xorshift **313,7 → 218,8 ms (−30,2%)**, asm floor 217,4 ⇒ **1,006x**. Đo ghép cặp xen kẽ,
  **2 vòng độc lập** cho cùng kết quả. Trần dự báo 89,8 ms, đo được 94,9 ms — **dự báo ĐÚNG**.
- fib −2,6%, arrwalk −1,8%, callloop +0,5%.
- Trả lời câu hỏi bỏ ngỏ của handoff 07-29f: **giả thuyết 1 đúng** — `vD` loop-carried nên
  interval gộp-def phủ cả vòng lặp, temp nằm hẳn bên trong ⇒ interfere THẬT ⇒ bias trong
  allocator **về nguyên tắc** không với tới. Không phải hạn chế first-move-wins.

## ĐÃ ĐO XONG SAU KHI SHIP — trạng thái mốc M6-codegen
Mốc (quyết định D1 của user): **≤15% so với ASM floor** viết tay cùng hình dạng.

| shape | axiom | asm floor | vs asm | vs clang |
|---|---|---|---|---|
| fib | 586,0 | 528,4 | **1,11x** ✅ | 1,70x |
| xorshift | 217,6 | 218,2 | **1,00x** ✅ | 1,01x |
| arrwalk | 390,5 | *chưa có floor* | ? | 1,13x |
| callloop | 81,9 | *chưa có floor* | ? | 0,83x |

fib đã xác nhận GHÉP CẶP 2 vòng, **đảo thứ tự xen kẽ**: 1,127x rồi 1,114x ⇒ nằm TRONG mốc,
không phải nhiễu. ⇒ **Cả hai shape CÓ floor đều ĐẠT mốc M6-codegen.** Không thể tuyên bố mốc
ĐẠT toàn phần vì arrwalk/callloop chưa có NASM floor để so.

## VIỆC TIẾP THEO (đã chọn, đã có bằng chứng disassembly, chưa bắt đầu code)
**Fold hằng số vào TOÁN HẠNG IMMEDIATE của ALU, thay vì nạp `mov $C, reg` mỗi vòng lặp.**
Đây là "thuế codegen #2" đã đặt tên từ 2026-07-24e, và nay có bằng chứng đo được ở CẢ HAI shape
còn lại. Đọc thẳng từ disassembly (`-O3`, driver `2077495B`):

**callloop** — thân vòng lặp 13 lệnh, trong đó **4 lệnh chỉ để nạp hằng**:
```
mov $0x7,%rax ; mov %rsi,%rcx ; mov %rdi,%rdx ; mov $0x2,%rbx ; imul %rbx,%rdx ; add %rdx,%rcx
mov $0x3,%rdx ; imul %rdx,%rax ; add %rax,%rcx ; mov $0xfffff,%rax ; and %rax,%rcx
mov %rcx,%rsi ; lea 0x1(%rdi),%rdi
```
`and $0xfffff` vừa imm32; `imul $2` nên là `add`/`lea`; và `7*3` là biểu thức HOÀN TOÀN hằng
mà vẫn tính lại mỗi vòng.

**arrwalk** — thân vòng lặp 9 lệnh, và **chuỗi phụ thuộc mới là thứ đáng tiền** (`idx = tbl[idx]`
là pointer-chasing tuần tự ⇒ latency-bound):
```
lea 0xf42(%rip),%rbx   <- địa chỉ bảng, BẤT BIẾN trong vòng lặp mà vẫn nạp lại mỗi vòng
mov %rax,%rsi ; mov $0x8,%rdi ; imul %rdi,%rsi ; add %rsi,%rbx   <- đáng ra là (%rbx,%rax,8)
mov (%rbx),%rsi ; mov %rsi,%rax ; add %rax,%rcx ; lea 0x1(%rdx),%rdx
```
`imul` nằm TRONG chuỗi phụ thuộc (idx → imul → add → load) ⇒ ~3 chu kỳ latency/vòng, đây là
mục đắt nhất của arrwalk chứ không phải số lệnh.

**Thứ tự đề xuất, mỗi bước đo riêng (ĐỪNG batch — quy delta cho từng thay đổi):**
1. `MOV_IMM vC,k ; ALU vD,vC` → `ALU vD,imm(k)` khi k vừa imm32 và `counts[vC]==2`. Dùng ĐÚNG
   thành ngữ reference-count của `drop_dead_mov_imm`/`fuse_cmp_immediate`/`coalesce_dest_copy`
   (đã chứng minh an toàn 3 lần). Blast radius rộng nhất, rủi ro thấp nhất.
2. `IMUL vD, imm(2^k)` → `SHL vD,k` (và `imm(2)` → `ADD vD,vD`).
3. Địa chỉ có scale: `mov (%base,%idx,8),%dst` cho array index — lớn nhất, làm sau cùng.
4. Mở rộng shape B của `coalesce_dest_copy` sang `MACH_LOAD` (nó cũng ghi dst mà không đọc dst,
   đúng lập luận như LEA) ⇒ xoá `mov %rsi,%rax` của arrwalk, cũng nằm trong chuỗi phụ thuộc.

⚠️ **Cần NASM floor cho arrwalk + callloop** trước khi tuyên bố mốc M6-codegen ĐẠT; thêm vào
`$srcs[...].asm` trong `scripts/perf_suite.ps1` (fib/xorshift đã có mẫu ngay tại đó).
⚠️ **Một lần chạy `perf_suite`/`perf_fib` KHÔNG đáng tin** (phương sai 8–10%/lần chạy trên máy
này). Mọi con số phải đo GHÉP CẶP xen kẽ, các vòng KHÔNG chồng lấp, và lặp lại ít nhất 2 vòng.

## Backlog còn lại (sau việc trên)
- RFC 0035: method/global/ctor vẫn dùng scheme cũ (`axS_`/`axG_`/`axC_` chưa làm) ⇒ fn-vs-struct
  vẫn dựa reject ở typecheck. P3 (E0501 → error) vẫn bị chặn bởi shim runtime trùng lặp hợp lệ.
- `mod_name` rỗng ở `register_module_from_lib` (binding `mod.NAME` là đồ chết) — vô hại, nhưng
  là bẫy cho người đọc.
- Module path nhiều đoạn không resolve ở call site (`bin.libcol.liba.helper()`) — có sẵn từ trước.
- M6-opt (accumulator/tail-rec) là milestone RIÊNG với M6-codegen — ROI cao hơn allocator và
  KHÔNG đụng code self-host-critical.
- Nếu hết việc: `axiom-bug-probe` để nạp lại backlog.

## ⛔ Cảnh báo còn hiệu lực
- **ĐỪNG cài loop rotation / bottom-test** — đã đo **+0,1%**.
- **ĐỪNG định giá coalescing/copy trên fib** — fib latency-bound, copy vô hình ở đó.
- Backend/ABI/linker ⇒ **B==C bắt buộc** trước commit (A!=B là bình thường khi codegen đổi).
