---
name: rfc0014-drop-glue-blocked
description: "RFC 0014 drop-glue ✅ SHIPPED c149872 2026-07-16 (opt-in -ctgc-free): resolve_drop_method + lower_destroy gọi Type.drop(self) trước OP_DESTROY cho non-escaping owned local có drop method. Fix 2 bug: (1) unblocked bởi RFC 0015 P2/P3 CTGC activation; (2) lower_destroy reg-lookup bug (sym.name_id vs sym_idx key→reg=0→free bị bỏ, tức P3 038c2ea INERT ko free gì). Free SCOPE = drop-typed only; general free DEFER (unsound self-host UAF, RFC 0010 §9). Oracle bin/t_drop.ax, scripts/ctgc_free_check.sh. Mở khóa bignum leak."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**RFC 0014 — `drop(self)` hook tối giản cho CTGC**, viết để giải quyết
[[bignum-ctgc-conflict]] (leak thật trong `std.bignum.BigUint.limbs`). Thiết kế: struct
định nghĩa `fn drop(self):` tùy chọn, CTGC gọi nó ngay trước khi free khối instance riêng
— tái dùng NGUYÊN VẸN cơ chế method-resolution của RFC 0007 (operator overload,
`match_mangled_method_raw_bytes`) để tránh lặp lại lỗi BUG#68 (diagnostic lệch cơ chế thật).
Không cần owned-field annotation, không topo-order, không auto field-walking — người dùng
viết đúng những gì struct sở hữu (giống Rust Drop / C++ destructor, tối giản hơn).

Đầy đủ thiết kế + rationale (vì sao không chọn owned-field-annotation, rủi ro double-free
khi copy struct-có-drop, kế hoạch P0-P4) nằm trong `rfcs/0014-drop-glue.md`.

**Implementation (P2) đã thử ngay trong session viết RFC này** — `resolve_drop_method` +
wiring vào `lower_destroy` (air_builder.ax) — code ĐÚNG THIẾT KẾ nhưng **REVERT hoàn
toàn** sau khi phát hiện oracle test `bin/t_drop.ax` không bao giờ thấy `drop` chạy, dẫn
tới phát hiện [[bug69-ctgc-ownership-escape-noop]]: CtgcInjector (cùng OwnershipChecker,
EscapeAnalyser) là NO-OP hoàn toàn từ trước tới giờ (bug độc lập, nghiêm trọng hơn nhiều).
Không có destroy node nào từng được tạo ra để gọi tới `drop` — không thể verify code đúng
hay sai, chỉ chứng minh được nó "an toàn vì không bao giờ chạy".

**Trạng thái: BLOCKED.** RFC 0014 CHỈ nên tiếp tục triển khai thật SAU KHI BUG#69 được xử
lý (CTGC activation — dự kiến là một nỗ lực riêng, lớn, xem chi tiết trong
[[bug69-ctgc-ownership-escape-noop]]). Giữ lại trong repo: `rfcs/0014-drop-glue.md` (thiết
kế đầy đủ, sẵn dùng), `bin/t_drop.ax` (oracle tài liệu — chưa đăng ký regression, giống tinh
thần `bin/t_aggcopy.ax` của RFC 0010).
