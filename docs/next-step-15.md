# AXIOM Language Project — next-step-15: Mở rộng feature ngôn ngữ

## Bối cảnh

Sau next-step-14: AXIOM đã **tự host xác định** (fixpoint stage3==stage4 = d7f14c2c),
chuỗi bug #24→#31 + Family C đã đóng. Hướng tiếp theo (do user chọn): **mở rộng độ
phủ feature ngôn ngữ** — implement các feature CÓ TRONG grammar (`docs/GRAMMAR.ebnf`)
nhưng compiler chưa hỗ trợ.

**Nguyên tắc (CLAUDE.md):** KHÔNG bịa syntax ngoài spec. Grammar là authoritative.
AXIOM là **indent-based** (`:` + INDENT), sum type `type X = A | B`, interface
duck-typing. KHÔNG thêm braces `{}` / `enum {}` / Go-style `fn (recv)` / associated
`type Item` — đó là dialect SAI (xem §Triage).

## Gap matrix (đo bằng `dump-air` trên syntax ĐÚNG + file std thật, 2026-06-26)

| Feature | Grammar | Status | Ghi chú |
|---|---|---|---|
| Sum type `type X=A\|B` | ✅ | ✅ works | result.ax/option payload OK |
| `match` (block-form arms) | ✅ §207 | ✅ works | arm phải là Block hoặc bare-Expr+NEWLINE (KHÔNG `return` cùng dòng pattern) |
| Generics `fn f[T]`, `T[U]` | ✅ | ✅ parse OK | |
| Struct (fields) + by-value ABI | ✅ | ✅ works | Family C đã fix |
| `.slice`, `.len`, `.ptr` trên str | ✅ | ✅ works | |
| std files đúng syntax | — | ✅ 0 lỗi: collections.ax, result.ax, process.ax | |
| **Inline struct method** `fn m(self)` | ✅ §83-85 | ❌ **GAP** | parse_struct_decl ĐÃ gọi parse_func_decl cho `fn` trong body (parser.ax:918), nhưng `parse_func_decl` đòi `name: type` → **param `self`/`mut self` (không type) gãy** ("expected type expression" tại `self`) |
| **Interface** (colon-form) | ✅ §94 | ❌ **GAP (partial)** | `interface X:\n  fn m(self)->T` còn 2 lỗi — parse_interface_decl/parse_method_sig chưa nuốt `self` + method-sig |
| `enum {}`, braces, `fn (recv)`, `type Item` | ❌ KHÔNG có trong grammar | n/a | iter.ax/json.ax/log.ax/net.ax + tests/*suite* viết SAI dialect → 247-713 lỗi |

## Nhiệm vụ (ưu tiên)

### T1 — Inline struct methods (`self` receiver)  [ƯU TIÊN 1, giá trị cao nhất]
Root: `parse_func_decl` không nhận param `self`/`mut self` (implicit-typed = struct bao quanh).
- [ ] Parser: `parse_func_decl` (hoặc parse_param) chấp nhận first-param `self` / `mut self`
      không cần `: type`; gắn type = struct cha (khi gọi từ parse_struct_decl).
- [ ] Resolver: bind `self` trong thân method về type struct cha; đăng ký method vào
      namespace của struct (để `p.getx()` resolve được).
- [ ] Typecheck: method call `recv.method(args)` → tìm method trong struct của recv,
      truyền recv làm `self` (by-address pointer — khớp struct reference semantics).
- [ ] AIR/codegen: lower `recv.method(args)` thành call(method, recv, args). Vì struct =
      con trỏ 8-byte (BUG#30), self = con trỏ → khớp param ABI sẵn có. Khả năng KHÔNG cần
      sửa codegen, chỉ frontend.
- [ ] Test: `bin/t_method.ax` (`Pt{x}.getx()==7`, `mut self` mutate, method gọi method).
- [ ] Regression gate + verify fixpoint stage3==stage4 KHÔNG vỡ.

### T2 — Interface declarations (structural/duck typing)  [ƯU TIÊN 2]
- [ ] Parser: sửa `parse_interface_decl`/`parse_method_sig` cho colon-form + `self` (2 lỗi còn lại).
- [ ] Typecheck: structural check — struct nào có đủ method-sig thì thỏa interface (không cần `implements`).
- [ ] (Tùy phạm vi) dynamic dispatch: defer nếu lớn; làm static/duck-typed trước.
- [ ] Test + fixpoint.

### T3 — Triage file dialect-sai  [ƯU TIÊN 3, dọn nợ]
iter.ax, json.ax, log.ax, net.ax + tests/*_suite*.ax dùng braces/enum/Go-method/assoc-type.
- [ ] Quyết định policy: (a) viết lại theo grammar AXIOM, hoặc (b) đánh dấu out-of-scope/quarantine.
- [ ] KHÔNG thêm syntax braces/enum vào compiler (vi phạm spec). Nếu muốn `enum`/assoc-type
      là feature thật → cần RFC riêng cập nhật GRAMMAR.ebnf trước.

## Definition of done mỗi task
- [ ] Probe syntax đúng → 0 lỗi parse/typecheck.
- [ ] Compile + run test cho ra giá trị đúng (thêm vào `scripts/regression_repros.sh`).
- [ ] `scripts/verify_bug29_selfhost.sh`: fixpoint stage3==stage4 GIỮ NGUYÊN.
- [ ] Commit + push; cập nhật doc này + knowledge/bugs.md nếu có root-cause đáng lưu.

## Lưu ý môi trường (kế thừa)
- Windows Defender realtime ON → mỗi self-link build ~50s wall (CPU~0), KHÔNG phải hang; timeout ≥120s.
- `dump-air` (parse+resolve+typecheck+AIR) KHÔNG ghi exe → nhanh, dùng để probe feature.
- Rebuild stage1: `& "scripts\rebuild_stage1.ps1"` (~25s). Verify fixpoint ~2-2.5h.
- Kill orphan `axc*` trước rebuild; không chạy 2 self-link build song song (đè axiom_temp.obj).
