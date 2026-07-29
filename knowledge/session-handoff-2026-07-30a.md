---
name: session-handoff-2026-07-30a
description: "HANDOFF 2026-07-30a — HEAD 1172648 pushed, driver seed==A==B==C E72FB62E, 564/564. M6-codegen ĐO ĐẦY ĐỦ lần đầu: 3/4 shape ĐẠT, callloop 1.22x MISS. Peephole 1d (fold hằng vào toán hạng ALU) đã ship."
metadata:
  type: project
---

# HANDOFF 2026-07-30a — **ĐỌC ĐẦU TIÊN**

## Trạng thái
- HEAD `1172648`, đã **push lên `origin/main`**, cây sạch (chỉ `.claude/settings.json`
  untracked — **của user, đừng đụng**).
- Daily driver `bin/axc_native.exe` = **`E72FB62E`**, tự tái tạo: **seed == A == B == C**.
- Gate đầy đủ XANH: regression **564/564** (+3 dòng `t_aluimmfold`), ELF 12/12, ctgc 16/16,
  exe_size 4/4, lib_collision 6/6, so_export ✓.

## Đã ship phiên này (2 commit)
1. `ae8516a` — **NASM floor cho arrwalk + callloop** (chỉ `scripts/perf_suite.ps1`).
2. `1172648` — **peephole 1d `fold_alu_immediate`** + 4 encoder mới (`and_ri` /4, `or_ri` /1,
   `xor_ri` /6, `imul_ri` 0x69) + oracle `t_aluimmfold`(42).

## ⭐ KẾT QUẢ LỚN NHẤT: mốc M6-codegen nay ĐO ĐƯỢC ĐẦY ĐỦ — và **CHƯA ĐẠT**
Trước phiên này chỉ fib/xorshift có floor nên **không thể kết luận**. Nay cả 4 shape đều có
floor NASM cùng hình dạng. Đo ghép cặp, 2 vòng không chồng lấp, **sau** khi ship 1d:

| shape | vs asm floor (2 vòng) | ≤15%? |
|---|---|---|
| fib | 1,13x / 1,13x | ✅ |
| xorshift | 1,00x / 1,01x | ✅ |
| arrwalk | 1,15x / 1,14x | ✅ (sát mép) |
| **callloop** | **1,21x / 1,23x** | ❌ **MISS** |

⇒ **callloop là shape DUY NHẤT còn chặn mốc M6-codegen.** (Trước 1d: 1,28x/1,29x.)

⚠️ Máy chạy chậm hơn ~5% ở vòng đo thứ hai — **mọi cột floor cố định đều tăng** (fib floor
519→548, xorshift 216→222, arrwalk 338→362). Vì vậy **chỉ tin TỶ LỆ, đừng tin ms tuyệt đối**
giữa các phiên. callloop giảm 79,2→77,4 ms tuyệt đối TRONG KHI máy chậm đi ⇒ win là thật.

## VIỆC TIẾP THEO — đóng nốt callloop (đã có disassembly, chưa bắt đầu code)
Thân vòng lặp hiện tại (10 lệnh, `-O3`, driver `E72FB62E`):
```
mov  $0x7,%rdx          <- hằng BẤT BIẾN trong vòng, nạp lại mỗi vòng
mov  %rax,%rbx          <- copy acc (destructive form)
mov  %rcx,%rsi          <- copy i
imul $0x2,%rsi,%rsi     <- nên là add/lea
add  %rsi,%rbx
imul $0x3,%rdx,%rdx     <- 7*3 là hằng HOÀN TOÀN, vẫn tính lại mỗi vòng
add  %rdx,%rbx
and  $0xfffff,%rbx
mov  %rbx,%rax          <- copy ra
lea  0x1(%rcx),%rcx
```
Floor tương ứng chỉ 5 lệnh: `lea rax,[rax+rcx*2] ; add rax,21 ; and rax,1048575 ; inc rcx ; cmp/jb`.

Ba mục còn lại, **mỗi mục đo RIÊNG (đừng batch)**:
1. **Constant folding `7*3` → `21`** sau inline. `work(acc,i,7)` được inline nên compiler ĐÃ có
   thông tin; hai lệnh `mov $7` + `imul $3` là thuần lãng phí. ⚠️ Đây là **AIR-level const
   propagation**, ranh giới M6-codegen / M6-opt cần cân nhắc — nhưng floor đã tính nó vào.
2. `IMUL vD, imm(2^k)` → `SHL vD,k`; `imm(2)` → `ADD vD,vD`. Rẻ, cùng thành ngữ peephole.
3. Hai copy quanh dạng destructive mà `coalesce_dest_copy` chưa với tới (nó cần 3 lệnh LIỀN KỀ).
4. (arrwalk, riêng) địa chỉ có scale `(%rbx,%rax,8)` + hoist địa chỉ bảng bất biến — `imul` nằm
   TRONG chuỗi phụ thuộc pointer-chasing nên đắt hơn số lệnh gợi ý.

## ⭐⭐⭐ BÀI HỌC PHƯƠNG PHÁP PHIÊN NÀY (quan trọng hơn cả bản vá)
**Bản ĐẦU của peephole 1d KHÔNG LÀM GÌ CẢ — và qua sạch mọi gate trong khi không làm gì.**
Viết với cửa sổ 2 lệnh liền kề (`MOV_IMM vC,k ; ALU vD,vC`), nó khớp **0/4** hằng của callloop.
Lý do: selection **materialise hằng TRƯỚC** khi phát copy `MOV vT,lhs` của dạng destructive, nên
MOV_IMM và bên tiêu thụ cách nhau HAI lệnh. Shape thật là:
`MOV_IMM vC,k ; MOV vT,vD ; ALU vT,vC`.

⛔ **Disassembly KHÔNG THỂ phát hiện điều này** — ở đó cặp lệnh trông LIỀN KỀ
(`mov $0x2,%rbx; imul %rbx,%rdx`), vì các tầng sau không giữ nguyên thứ tự này. **Disassembly là
bằng chứng về BINARY ĐÃ XONG, không phải về mảng mà một peephole PRE-ALLOCATION khớp trên.**
Cách bắt: in thẳng instruction stream mà pass nhìn thấy — 1 lần chạy là ra.
⇒ **Luật mới: một peephole phải được chứng minh là CÓ NỔ, không chỉ là AN TOÀN.**
Cùng họ thất bại với "module cursor" RFC 0035 (`A==B` xanh chứng minh SỐ KHÔNG).

⭐ **Và lần thứ HAI cùng một sai lầm định giá**: chính ứng viên này đã bị NASM định giá
2026-07-24e ra −0,5/−13 ms (= nhiễu) rồi **loại bỏ** — vì định giá trên **fib**, vốn
latency-bound, nơi lệnh nạp hằng nằm NGOÀI chuỗi phụ thuộc nên vô hình. Y hệt chuyện register
coalescing: đo 0 trên fib, rồi −30% trên xorshift. **Định giá một thay đổi codegen trên MỘT
shape sẽ liên tục cho ra số 0 tự tin cho những thứ đáng vài phần trăm ở chỗ khác.**

## ⭐ Oracle `t_aluimmfold` — ĐÃ HIỆU CHUẨN bằng 3 lần phá có chủ ý
- `and_ri` viết nhầm digit /1 (thành `or`) → **crash** (nó phá luôn `and rsp,-16` căn stack).
- Bỏ guard imm32 → **exit 8**, đúng như thiết kế (0xFFFFFFFF sign-extend thành −1 ⇒ giá trị
  không đổi).
- Bỏ REX.R của `imul_ri` → **PASS 42** lần đầu! ⇒ **lộ lỗ hổng THẬT**: không check nào chạm
  r8–r15, mà `imul_ri` là encoder DUY NHẤT trong 4 cái mới cần REX.R (dst nằm ở CẢ reg lẫn rm).
  Đã thêm khối 12 biến sống đồng thời để ép allocator dùng thanh ghi cao; nay phá là crash.
  **Bài học: encoder mới phải có test chạm r8–r15, nếu không nửa trường REX không có coverage.**

## 🔎 ĐÃ ĐIỀU TRA (read-only, cuối phiên) — **MACH_LEA KHÔNG có index/scale**, và đó là nút thắt chung
Đã đọc thẳng nguồn, không suy đoán:
- `x86_encode_lea(dst, base, disp)` (`x86_encoding.ax:322`) chỉ nhận **base + displacement**.
- `x86_encode_modrm_rm` (`x86_modrm.ax:102`) chỉ phát **SIB** cho đúng ca bắt buộc
  (`base & 7 == 4`, tức RSP/R12), và khi đó **hard-code index = 0x04 = "không có index",
  scale = 0** (`x86_modrm.ax:134`). **REX.X không bao giờ được set** ở đường này.
⇒ `lea (%rdx,%rdx,2)` hay `mov (%rbx,%rax,8),%rsi` **hiện KHÔNG diễn đạt được**.

**Đây là nút thắt CHUNG của 3 mục backlog còn lại**, nên làm nó TRƯỚC sẽ mở khoá cả ba:
1. `imul $2/$3` → LEA scale (callloop). LEA scale phủ k ∈ {2,3,4,5,8,9}: `x*3` =
   `lea (%r,%r,2)`. Latency 1 thay vì 3, mà `imul` đang nằm TRONG chuỗi tích luỹ.
   ⚠️ **ĐÍNH CHÍNH ghi chú trước đó trong chính file này**: tôi đã viết rằng `IMUL → SHL/ADD`
   "phải chứng minh không JCC/SETCC nào đọc cờ ở giữa". **Rào cản đó KHÔNG tồn tại.** `IMUL`
   vốn đã phá cờ (ISA để SF/ZF/AF/PF **undefined**), nên **không có consumer ĐÚNG ĐẮN nào có
   thể đọc cờ xuyên qua nó** ⇒ thay `IMUL` bằng `ADD`/`SHL` không thể phá code đúng, xét về cờ.
   Đã kiểm chứng thêm: compiler **không** phát nhánh đọc overflow sau phép toán — wrap theo bề
   rộng làm bằng mask/sign-extend (`emit_wrap_to_width` → `emit_load_extend`, `x86_selector.ax:887`),
   không đọc OF; `CC_O` có khai báo nhưng không có site nào phát `JO` sau IMUL.
   ⇒ **`imul $2` → `ADD vD,vD` là một thay đổi ~5 dòng, an toàn, KHÔNG cần chờ SIB.** Vẫn nên
   ưu tiên LEA cho bức tranh chung (nó phủ cả `$3`, và không ghi cờ nên không ràng buộc thứ tự),
   nhưng nếu muốn một win nhỏ, độc lập, đo được trước khi làm SIB thì đây là mục rẻ nhất.
2. Địa chỉ có scale cho index mảng (arrwalk) — mục ĐẮT NHẤT của arrwalk vì `imul` nằm trong
   chuỗi pointer-chasing.
3. `base + idx*k` tổng quát.

**Việc cần làm (ước lượng, chưa code):** thêm cách mang index-vreg + scale trong `MachInst`
(dùng `src2` cho index vreg — regalloc đã đếm `src2` là một operand nên liveness tự đúng; scale
nhét vào `padding` hoặc `imm`), một encoder SIB thật (set cả **REX.X** cho index r8–r15 — xem
bài học REX.R của `imul_ri` ở trên: nửa trường REX không có test thì không có coverage), và mở
rộng `format_operand` cho đường asm-text. Backend ⇒ **B==C bắt buộc**.

## Backlog còn lại (sau callloop)
- RFC 0035: method/global/ctor vẫn scheme cũ (`axS_`/`axG_`/`axC_`); P3 (E0501 → error) vẫn bị
  chặn bởi shim runtime trùng lặp hợp lệ.
- `mod_name` rỗng ở `register_module_from_lib` (binding `mod.NAME` là đồ chết) — vô hại.
- Module path nhiều đoạn không resolve ở call site (`bin.libcol.liba.helper()`) — có sẵn từ trước.
- M6-opt (accumulator/tail-rec) là milestone RIÊNG với M6-codegen.
- Nếu hết việc: `axiom-bug-probe`.

## ⛔ Cảnh báo còn hiệu lực
- **ĐỪNG cài loop rotation / bottom-test** — đã đo **+0,1%**.
- **ĐỪNG định giá copy/coalescing/hằng số trên fib** — fib latency-bound.
- Backend/ABI/linker ⇒ **B==C bắt buộc** trước commit (A!=B là bình thường khi codegen đổi).
- Một lần chạy `perf_suite` KHÔNG đáng tin (phương sai 8–10%): đo ghép cặp xen kẽ, ≥2 vòng, và
  so TỶ LỆ chứ không so ms giữa các phiên.
