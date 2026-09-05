---
name: stdlib-reachability-option-c-stages
description: The option-C 4-stage plan for stdlib reachability — stage 1 SHIPPED; the ORIGINAL stage order is REFUTED (stage 4 is a precondition of stage 2, stage 3 depends on stage 2); also records why option B (bundling more modules) was rejected with measurements
metadata:
  type: project
---

# Stdlib reachability — option C, 4 stage (stage 1 đã ship; **thứ tự gốc ĐÃ BỊ BÁC BỎ**)

> Tách khỏi `BACKLOG.md` ngày 2026-09-05 (§24: file con trỏ không giữ sự thật).
> ⛔ **Thứ tự stage ghi trong bản gốc dưới đây là SAI** — phán quyết đầy đủ ở
> [[bug-b3b-cross-module-index-and-loader-defects]] §5–§7. Giữ nguyên văn để đối chiếu.
> Mục **option B bị bác bỏ** (kèm số đo DFE) chỉ tồn tại ở đây — đừng xây lại nó.

## 🧭 HƯỚNG ĐÃ ĐỊNH GIÁ — option C, làm 4 stage, mỗi stage A==B
**Stage 1 ✅ XONG 2026-08-07** (A==B `9C6726C1…`, 709/709 ở cả hai mức, breakage audit 870 file).
`strip_imports` nay bôi trắng **chỉ khi tên module đúng bằng** một mục trong **MỘT bảng duy nhất**
`preprocessed_module_name()` (11 mục: 8 bundled + `std.os.win32`/`std.os.linux_sys`/`std.net` —
ba tên chỉ bị `strip_package_prefixes` viết lại, không bundled, nhưng import của chúng vẫn phải bôi
trắng vì call site đã mất tiền tố). `concatenate_stdlib` đọc đường dẫn qua `bundled_module_path(i)`
suy ra **từ chính bảng đó** ⇒ hết "hai danh sách".
⇒ `import std.math` + `std.math.sqrt(16.0)` **chạy end-to-end** (`bin/t_stdmath.ax` = 12.000000).
✅ **B2 rơi ra miễn phí** (so khớp CHÍNH XÁC ⇒ có biên phân cách theo cấu trúc); **B4** và **B5**
đã sửa cùng stage (xem §BUG PHỤ).
⚠️ **Hệ quả đã đo:** `import std.{sync,thread,process,iter,cli}` nay **tới được loader ⇒ compiler
SEGV** (B3/B6) thay vì bị nuốt im lặng như trước. Không file nào trong corpus dính (audit 0
collateral), nhưng đây là **bug tiếp theo phải sửa** — crash không chẩn đoán là tệ hơn cả im lặng.
⛔⛔ **THỨ TỰ DƯỚI ĐÂY LÀ SAI — đã đo 2026-09-05, xem
`knowledge/bug-b3b-cross-module-index-and-loader-defects.md` §5-§7.** Giữ nguyên văn để đối chiếu,
nhưng **đừng thực hiện theo thứ tự này**:
- **Stage 4 KHÔNG "tuỳ chọn" — nó là TIỀN ĐỀ của stage 2.** Không có danh tính theo module thì stage 2
  **buộc phải khớp theo CHÍNH TẢ** ⇒ **vi phạm quyết định D1-3**.
- **Stage 3 KHÔNG rẻ hơn stage 2 — nó PHỤ THUỘC stage 2** (hôm nay chính lời gọi `std.string.replace`
  của compiler chỉ chạy nhờ viết lại văn bản). Và **B1 không cần stage 3**: cho
  `strip_package_prefixes` **bỏ qua vùng `"…"`** là một hàm, không RFC, và mở khoá luôn số CỘT.
- **Thứ tự đúng (rủi ro thấp trước):** B3b-3 → B3b-2 → B3b (chỉ số xuyên cây) →
  `strip_package_prefixes` biết-literal → *rồi mới* định giá lại stage 2 trên nền stage 4.

*(Nguyên văn cũ:)*
**Stage 2:** đăng ký 8 tên bundled thành `SYM_MODULE` có cờ *bundled*; `lazy_resolver_resolve_field`
tra cứu **scope toàn cục của unit hiện tại** thay vì nạp file ⇒ `std.collections.new_vec` chạy, và
`std.string.len` bind **cùng một symbol** với `len` trần ⇒ **delta codegen = 0**.
⚠️ Lời hứa "delta codegen = 0" **chỉ đúng cho những tên tình cờ đang là duy nhất** — vỡ trên
overload thật (`map` ở cả Option lẫn Vec; `unwrap` ở cả Option lẫn Result), mà `resolver.ax:653/662`
đã ghi rõ là chọn **TÙY TIỆN theo thứ tự hash-slot**. **BẮT BUỘC RFC** (§13: đổi `std.X.f` biểu thị gì).
**Stage 3:** **XOÁ HẲN `strip_package_prefixes`** ⇒ hết B1, và **mở khoá CỘT trong chẩn đoán** (§3b).
**Stage 4 (~~tuỳ chọn~~ → **TIỀN ĐỀ**):** module scoping thật cho symbol spliced, dùng
`Symbol.decl_node` → offset → `srcmap_find` (hạ tầng đã có: `main_air.ax:682-689` ghi sẵn vùng srcmap
cho **từng** module bundled).
⛔ **KHÔNG làm option B** (nhồi thêm module vào bundle): chỉ 2/12 module thêm được an toàn
(`math`, `sort`), trả **+0,15 s mỗi lần biên dịch VĨNH VIỄN** cho mọi user kể cả người không dùng,
và **không sửa gì về cấu trúc**.
📏 Đo được (bác bỏ lo ngại của tôi): DFE **bật mặc định** (`main_air.ax:928`) ⇒ bundle thêm module
không dùng cho ra **binary BYTE-IDENTICAL**. Giá thật là **thời gian biên dịch**, không phải kích thước.

