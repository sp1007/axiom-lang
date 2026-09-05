---
name: lesson-comment-protects-one-line-only
description: Four measured instances in one day of the same meta-defect — the rule was written down correctly and violated anyway, because prose protects one line and only a check protects every line
metadata:
  type: feedback
---

# Một comment bảo vệ ĐÚNG MỘT DÒNG. Chỉ một CHECK mới bảo vệ mọi dòng.

Câu này **đã có sẵn** trong [[lesson-exit-code-8bit-masking]]. Ngày 2026-09-05 nó tái xuất **bốn lần
trong một phiên**, mỗi lần ở một hệ thống con khác nhau. Ghi lại ở đây như một **lớp defect có tên**,
kèm *check nào lẽ ra đã bắt được* — vì sản phẩm bền của bài học này không phải thêm một ghi chú, mà là
những cái check đó.

## Hình dạng chung
> Một kỹ sư **đụng phải** bẫy → **mô tả đúng** cơ chế bằng prose → **né tại chỗ** →
> **bản sao khác của cùng luật vẫn vi phạm.**

Prose gắn với **vị trí**, không gắn với **luật**. Ai đọc dòng đó thì được bảo vệ; ai viết dòng thứ hai
ở file khác thì không. Và cả bốn ca dưới đây đều **thất bại IM LẶNG** — không có test nào đỏ.

## Bốn ca đo được (2026-09-05)

| # | Luật đã viết ở đâu | Bị vi phạm ở đâu | Hậu quả |
|---|---|---|---|
| 1 | `typecheck.ax:1985` — có **canh sẵn** + comment nêu đúng cơ chế "`decl_node` có thể trỏ cây module KHÁC … *regression: t_mathx*" | `typecheck.ax:1626`, `:5898`, `:7081` — **không canh, không kiểm biên** | **SEGV 139** compiler (B3/B6); và biến thể **im lặng** đóng dấu chữ ký sai khi chỉ số rơi trúng vùng hợp lệ |
| 2 | `BACKLOG.md` §print/println — ghi rõ **"P1+P2+P3 đóng bằng RFC 0038"** | cùng file, ~300 dòng sau: **"BỊ CHẶN BỞI P1 ⇒ phải sửa P1 trước"**; và `CLAUDE.md §7.1` | Chặn oan **quy tắc oracle §7.1** suốt nhiều phiên |
| 3 | `print_helpers.ax:114` `is_verbose_debug` — whitelist **19 chuỗi** quyết định "dòng nào là debug" | mọi call site: phân loại theo **CÁCH VIẾT**; `parser.ax:171` đã ghi một ca rò rỉ | 19 dòng in **vô điều kiện**, không cờ nào tắt được; **đổi câu chữ = âm thầm đổi việc in hay không** |
| 4 | `main_air.ax:869-873` — prose mô tả **chính xác** bẫy aliasing của `replace`, kèm *"which segfaulted the resolver on the first attempt here"* | `main_air.ax:1957` **vẫn free alias** (dù `resolver.ax:803`/`:809` đã comment bỏ free, và chuỗi `r0…r8` không free trung gian) | **UAF** tên module ⇒ chẩn đoán in **rác khác nhau theo mức tối ưu** (`\xef\xbf\xbd\xef\xbf\xbd` ở -O0, `Sj` ở -O1) |
| **5** | **CHÍNH TÔI**, cùng phiên: viết `lesson-comment-protects-one-line-only.md` (file này) + luật §24 "commit sau đóng TODO thì sửa file chứa TODO **trong cùng commit**" | **`session-handoff-2026-09-05a.md`** — mục "VIỆC TIẾP THEO" vẫn kể **cơ chế B3b ĐÃ BỊ BÁC BỎ** và vẫn nói investigator "đang chạy" dù nó đã báo cáo xong | File có nhiệm vụ **DUY NHẤT** là làm điểm resume ⇒ phiên mới đọc nó **đầu tiên** và tin theo. Sửa ở `4e1426b` |

## Vì sao nó cứ lọt (đừng đổ cho bất cẩn)
- **Prose vô hình ở nơi cần nó.** Người viết `typecheck.ax:5898` không có lý do gì để đọc `:1985`.
- **Né tại chỗ trông giống như đã sửa.** Ca 4: comment bỏ hai dòng free ở `resolver.ax` **có tác dụng**
  — bug biến mất khỏi tầm mắt, nên không ai đi tìm bản sao thứ ba.
- **Cả bốn ca đều im lặng.** Không có ca nào làm đỏ một test đang có. Ca 1 và 4 còn **phụ thuộc dữ
  liệu** (kích thước cây / có dấu chấm trong tên) nên tái hiện thất thường.
- **"Một luật, N bản sao" là hình dạng, không phải sự trùng hợp.** Cùng họ với P6
  (`ownership.ax:138,162`), RFC 0037 rank 2/3, và P4 — tất cả đều là **khớp theo chính tả / lặp lại
  luật**, đúng lớp mà user đã phán quyết ở **D1 quyết định 3: khoá theo DANH TÍNH, không theo CÁCH VIẾT.**

## ⭐⭐⭐ Ca 5 là ca quan trọng nhất: luật này áp cho **CHÍNH BẢN VIẾT CỦA MÌNH**
Bốn ca đầu là defect **tìm thấy trong repo**. Ca 5 do **chính tôi tạo ra, trong cùng phiên mà tôi
đang viết file bài học này** — và nằm ở **file resume**, chỗ đắt nhất có thể.
⇒ Kết luận không dễ chịu nhưng đúng: **prose của tôi cũng gắn với VỊ TRÍ y như prose của người
khác.** Viết ra luật **không** miễn nhiễm cho người viết. Nó cũng bác bỏ cách đọc dễ chịu về bốn ca
kia ("ai đó bất cẩn"). Cụ thể:
- Khi một phép đo **BÁC BỎ** một mô tả cũ, phải sửa **mọi chỗ đang kể mô tả đó** trong **CÙNG
  commit** — nhất là handoff/BACKLOG, vốn được đọc **TRƯỚC** file bằng chứng.
- **Trạng thái động ("đang chạy nền", "chờ báo cáo") hết hạn nhanh nhất.** Đã ghi vào file bền thì
  phải sửa ngay khi nó hết đúng, hoặc **đừng ghi vào file bền**.
- Kiểm rẻ trước khi commit tài liệu: đọc lại **chính đoạn mình vừa dán** và hỏi *"câu này còn đúng
  sau những gì tôi vừa đo không?"* Ca 5 lẽ ra chỉ tốn một lần đọc lại.

## Luật hành động
1. ⭐ **Khi phát hiện một bẫy, đừng chỉ ghi comment — hãy đi tìm CÁC BẢN SAO KHÁC NGAY.**
   `grep` theo *cái được dùng* (`decl_node`, `@free(get_str_ptr(`, `replace(`), không theo lời văn.
   Ca 1 lẽ ra chỉ tốn một lần `grep pre_infer_func_signature`.
2. ⭐⭐ **Biến luật thành MỘT CHỖ Ở DUY NHẤT, rồi bắt mọi người đi qua đó.** Đây là điều đã làm cho
   ca 1 (`sym_decl_tree` + `pre_infer_symbol_signature`) và là lý do nó sẽ không tái diễn: không còn
   "bản sao cần nhớ canh" nữa. Với ca 4, tương đương là cho `replace` **luôn copy** (đóng cả lớp) —
   nhưng đó là đổi hợp đồng nóng **24 call site** ⇒ phải đo A==B riêng, **đừng gộp với bản vá call site**.
3. **Kiểm đầu vào ngay trong hàm bị lạm dụng** (CLAUDE.md §9), đừng trông vào việc caller cư xử đúng.
   Ca 1: `pre_infer_func_signature` nay tự kiểm `node_idx`/`sym_idx` ⇒ giết luôn **biến thể im lặng**
   mà không call site nào bắt được.
4. **Khi một commit sau đóng một TODO, sửa file chứa TODO đó TRONG CÙNG commit** (đã có ở §24; ca 2 là
   ví dụ nó bị bỏ qua). Hệ quả cho người đọc: **`git log` file mình sắp tin**.
5. **Nghi ngờ mọi "workaround tại chỗ" trong repo.** Một dòng `// @free(...)` bị comment, một vòng lặp
   tường minh "viết tay thay vì dùng helper", một tham số thừa — mỗi cái là **dấu vết của một bẫy chưa
   được đóng ở tầng luật**. Đọc lý do, rồi đi tìm bản sao.

## Cách phát hiện rẻ nhất (dùng được ngay)
Bốn ca đều lộ ra bằng **một** trong ba động tác, không cái nào cần build:
- `grep` tên hàm/biểu thức nguy hiểm rồi **đếm bản sao có canh vs không canh**;
- đọc **comment cạnh chỗ vừa sửa** và hỏi *"luật này còn ở đâu nữa?"*;
- với contract cấp phát: `grep '@free(get_str_ptr('` rồi đối chiếu từng cái với hàm sinh ra chuỗi đó
  (có bao giờ trả về **chính đầu vào** không?).

Liên quan: [[lesson-exit-code-8bit-masking]] · [[bug-cross-tree-decl-node-segv]] ·
[[bug-b3b-cross-module-index-and-loader-defects]] · [[lesson-taskstop-leaves-suite-running]]
