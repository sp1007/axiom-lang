---
name: session-handoff-2026-08-07a
description: Resume point after RFC 0038, libc steps 0-2, the P6 payload-discriminator fix, and RFC 0039; diagnostics rework is the next task
metadata:
  type: project
---

# Handoff 2026-08-07a

## Trạng thái

| | |
|---|---|
| HEAD | `74eab1d` fix(typecheck): callee payload đọc như symbol chỉ khi cờ nói vậy (P6) |
| Driver `bin/axc_native.exe` | A==B `A58F762AACDB79C1164481FFED09AD2DD4B0A357B662A4C08420C06BABB5FE97`, 2.307.584 byte |
| ⏳ **CHƯA COMMIT** | RFC 0039 (struct literal suy diễn) — A==B `84A13E958B59D2A1022C860C8E4637E81716BA65A66BFDDFD096034E4DB3FF68`, 2.313.728 byte, đang ở `bin/axc_cand.exe`/`axc_fpB.exe` |
| Baseline | **685/685** (cả default lẫn `-O0`) với RFC 0039; 682 nếu chưa có nó |
| Mốc B==C gần nhất | `c3eae77` / `52D1ABD4…` — **backend kế tiếp phải dựng lại B==C từ driver MỚI** |
| libc | `ucrtbase.dll` **16 → 13** |

## Đã ship
1. **`ca7a98d` RFC 0038** — `print`/`println` variadic. Mọi đối số sau đối số đầu **bị nuốt im
   lặng**. Desugar ở **`air_builder`, KHÔNG ở selector** ⇒ cả ba backend được sửa miễn phí.
   Kèm `println()` trần hết segfault + `error[E3033]` cho kiểu không in được.
2. **`50d92ce`** — audit libc → [audit-libc-dependencies](audit-libc-dependencies-2026-08-05.md).
3. **`3f3ea74`** — dọn libc bước 0-2: bỏ `clock`, `exit`, `fflush`.
4. **`74eab1d` P6** — xem dưới, bài học lớn nhất phiên này.
5. **`2ffa953`** — RFC 0039 (proposed → nay đã implement, chờ commit).

## ⭐ P6 — bài học đắt nhất, đọc trước khi đụng `payload`
`NODE_FIELD_EXPR.payload` **QUÁ TẢI**: `parser.ax:521-524` đặt nó = `pool.intern(field_text)`
(**id InternPool**); mono hoá generic **GHI ĐÈ** bằng **chỉ số symbol** và bật cờ **2048**
(`typecheck.ax:5451`). **Cờ 2048 là discriminator.** Mọi consumer khác đều tôn trọng nó; riêng
`:5765` thì không ⇒ đọc `symbols[intern("count")]` = **symbol stdlib không liên quan**, rồi dán kiểu
trả về aggregate của nó lên một lời gọi trả `i64`.
⚠️ **Đây là bug NGƯỜI DÙNG CHẠM ĐƯỢC trên cây sạch**, không phải vật cản dọn dẹp: quét định danh
stdlib làm tên method trong struct ⇒ **23/297 (7,7%) tên thường gặp fail** — `count`, `alloc`,
`actor`, `attempts`, `child_count`, `data_size`. Không cần sửa `std/` gì cả.
Sửa bằng accessor **`callee_symbol`** (payload cho `NODE_IDENT`; cho `NODE_FIELD_EXPR` **chỉ khi**
`flags & 2048`; ngược lại 0) + **`kind == SYM_FUNC`**, và bản sao open-code cùng luật ở `:5470` nay
đi qua accessor ⇒ **một luật, MỘT bản**.
⭐ **Kiểm tra `SYM_FUNC` KHÔNG phải "cho chắc"** — nó là load-bearing: `pre_infer_func_signature`
ở `:5772` đang **dán chữ ký FUNC 0 tham số lên symbol SYM_VARIANT**, gây **134 lỗi giả** ở
`stage2_preprocessed.ax`, 67 ở `self_linked_concatenated.ax`, 7 ở `std/result.ax`. Cờ 2048 một
mình **không** bắt được, vì payload của callee `NODE_IDENT` kiểu variant **đúng là** chỉ số symbol —
chỉ không phải của một hàm.
❌ **Đính chính bản ghi cũ của tôi:** "thêm extern thì vô hại, chỉ xoá mới gây lỗi" — **SAI**, chỉ là
sampling artifact. Thêm cũng gây lỗi (Δ=+2,+5,+6). Và nguyên nhân **không** phải type-table trôi.

## 🔜 VIỆC TIẾP THEO
0. **Commit RFC 0039** khi breakage audit xong (đang chạy nền: 818 file × 2 compiler,
   `/tmp/BEFORE.txt` vs `/tmp/AFTER.txt`). Lập luận giải tích: chuỗi `(ident:` **trước đây là lỗi
   parse cứng** ⇒ không file nào đang được nhận có thể chứa nó ⇒ thay đổi parser **thuần cộng
   thêm**; trong cây chỉ có trong COMMENT. Closure dùng `|params|` nên **không đụng** `(x: T)`.
1. 🔴 **CHẨN ĐOÁN — vi phạm §8 nặng nhất, xem BACKLOG §3b.** Repro `bin/probe11/s1.ax` (9 dòng):
   `error: unexpected token at offset 142042` — **offset byte vào buffer ĐÃ NỐI stdlib** (~142 KB),
   không file/dòng/cột; **9 lỗi dây chuyền**; **dump token thô** `kind=67`; thuật ngữ Pratt `nud`.
   Kèm: `main_air.ax:489` in `total_len=<số>` **vô điều kiện, không nhãn** trên mọi build; và **mọi**
   chẩn đoán thiếu dòng `--> file:line:col` + số dòng trong gutter (kiểm cả E3030 lẫn E3033).
   Frontend ⇒ A==B.
2. **libc bước 3** — viết lại `std/io.ax` sang handle native. Bỏ thêm 4 ký hiệu.
   ⚠️ `main_air.ax:163 read_file_content` đọc TOÀN BỘ input compiler qua đây.
3. **libc bước 4-9**, `atof` cần RFC. Xem file audit.
4. **P4** (RFC 0038): hàm `println` do user định nghĩa bị selector cướp vì khớp theo **chuỗi tên**
   chứ không theo **danh tính symbol** (`x86_selector.ax:1730`→`:1737`) ⇒ **B==C**.
5. **Di trú oracle sang §7.1** (stdout thay exit code) — nay khả thi, làm **dần**.

## Bài học đo lường (đắt, đừng học lại)
1. ⛔ **`strings | grep <tên libc>` KHÔNG trả lời được "có phụ thuộc libc không"** — nó báo
   `printf`/`puts`/`malloc` "có", nhưng đó là **hằng chuỗi** trong whitelist `cgen.ax:140`.
   **Phải parse bảng import PE.**
2. ⭐ **Compiler bị CẮT CỤT (25 KB) xuất hiện KHÔNG cần tải đồng thời** — log thành công đầy đủ,
   exit 0. ⇒ **`ls -l` sau MỌI lần build compiler** (đúng ~2,3 MB). Một compiler cụt đọc y hệt
   "regression thảm hoạ" (mọi test fail cùng lúc) ⇒ **kiểm binary trước khi nghi thay đổi vừa làm**.
3. **Một test fail không bao giờ là "flake"** — `t_ifaceconsumer` fail ⇒ bisect ra P6 trong ~6 lệnh.
4. **Comment trong repo trôi lệch và đã làm lạc hướng cả một cuộc điều tra** (`std/io.ax:4-5` đổ lỗi
   cho resolver; thủ phạm là danh sách bundle cứng `main_air.ax:401-426`, và cái đúng đã có từ
   **hai tuần trước** khi comment được viết). ⇒ luôn `git log` file mình sắp tin.
5. **Vứt lượt chạy BỊ GIẾT, đừng union với lượt sạch** (một audit hết giờ 10 phút hôm nay ⇒ vứt,
   chạy lại nền). Union chỉ báo THỪA reject, nhưng vẫn làm mất thời gian truy vết.
6. `grep -c` **exit 1 khi đếm 0** ⇒ `cmd | grep -c x && next` sẽ nuốt mất `next`.
