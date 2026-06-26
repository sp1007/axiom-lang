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

**KẾT LUẬN QUAN TRỌNG (sau khi probe lại bằng syntax grammar-conformant):** GRAMMAR
ĐÃ ĐƯỢC IMPLEMENT ĐẦY ĐỦ. Mọi "gap" ban đầu là do tôi probe bằng syntax KHÔNG theo
grammar (untyped `self`, `;`, braces `{}`, `enum`, Go-method `fn (recv)`).

| Feature (grammar-conformant) | Status | Bằng chứng |
|---|---|---|
| Sum type `type X=A\|B` | ✅ works | probe 0 lỗi |
| `match` (block-form arms) | ✅ works | arm = Block hoặc bare-Expr+NEWLINE |
| Generics `fn f[T]`, `T[U]` | ✅ works | |
| Struct + by-value ABI (Family C) | ✅ works | |
| **Method = typed `self` + UFCS** `fn m(self: ptr[T])`, gọi `x.m()` | ✅ works | u1.exe/u2.exe → exit 7 |
| **Inline method** (typed self trong struct body) | ✅ works | parser.ax:918 + exit 7 |
| **Interface** (method-sig typed/no-param) | ✅ works | if1/if2 0 lỗi; std/io.ax 0 lỗi |
| `.slice/.len/.ptr` str | ✅ works | |
| File grammar-đúng | ✅ 0 lỗi: collections, result, process, io | |

**KHÔNG có trong grammar (⇒ KHÔNG phải bug compiler):** untyped `self`, `Self` type,
`enum {}`, braces block, Go-method `fn (recv)`, associated `type Item`. Các file
iter/json/log/net/fmt + tests/*_suite* viết bằng dialect này (19-55 lần `{`/`enum`/
untyped-self mỗi file) → 160-713 lỗi parse. Đây là **nợ dialect**, không phải thiếu feature.

## Nghĩa lại của "mở rộng feature": THÊM grammar mới (RFC-gated)

Vì grammar hiện tại đã chạy đủ, "mở rộng" = thêm feature MỚI vào grammar. CLAUDE.md:
syntax change PHẢI có RFC + cập nhật `docs/GRAMMAR.ebnf` TRƯỚC khi implement.

### T1 — RFC 0002: Untyped `self`/`mut self` sugar  ✅ DONE (fixpoint fc5e6f48 giữ)
Hiện mọi method phải viết `mut self: ptr[Parser]` (boilerplate khắp compiler). Sugar:
trong thân struct, `fn m(self)` / `fn m(mut self)` ⇒ tự suy type = `ptr[StructCha]`.
- [ ] RFC 0002 + sửa GRAMMAR.ebnf: `Param = [Mod] IDENT [':' TypeExpr]` (type optional khi IDENT=self trong method).
- [ ] Parser: `parse_func_decl` nhận first-param `self`/`mut self` không type khi gọi từ parse_struct_decl (truyền struct type).
- [ ] Resolver/typecheck: gán self type = ptr[struct cha].
- [ ] Test `bin/t_method.ax` + regression gate + fixpoint giữ.

### T2 — `enum` + ADT codegen ✅ DONE (RFC 0003+0004, BUG#32 FIXED, fixpoint 692ba8e9)
- [x] RFC 0003: `enum Name:` indent-style sugar, desugar PARSER → sum type AST.
- [x] RFC 0004: codegen native cho user sum type (construct + match). Phát hiện BUG#32:
      trước đó match KHÔNG được lower + chỉ Ok/Err/Some/None hard-code → enum/sum vô dụng
      runtime. Fix: box [tag@0, payload@8], lower_variant_construct + lower_match (air_builder).
- [x] Test 5 ADT + built-in Option/Result; regression 19/19; fixpoint giữ.
- Chi tiết theo dõi: docs/next-step-15-sub-1.md.
- **Còn lại của T2:** `Self` type (RFC riêng); hạn chế v1 ADT (multi-field variant, str
  payload, generic user sum + match Option/Result) — follow-up.

### T3 — Triage nợ dialect (iter/json/log/net/fmt + suites)  [dọn nợ] ✅ DONE (triage+document)
- [x] Đo toàn bộ std/: build chỉ import std.string/io/collections (đều 0 lỗi). ~13 file
      (net/iter/log/json/mem/math/fmt/gpu/compiler/ffi/crypto/cli/arch) là stub aspirational
      dialect-ngoài-grammar (70-713 lỗi), KHÔNG import, KHÔNG compile được, KHÔNG block gì.
- [x] Tạo `std/MODULE_STATUS.md`: phân loại working vs aspirational + policy kích hoạt
      (rewrite-to-grammar HOẶC RFC mở rộng grammar; KHÔNG bịa braces/enum). → hết "masquerade".
- [ ] (Backlog, không gấp) rewrite từng module aspirational sang grammar khi cần.

## TASK QUEUE (thực hiện TUẦN TỰ — mỗi task: implement → regression → fixpoint → commit)

1. **RFC 0005 — int literal type inference** ⏳ ĐANG VERIFY FIXPOINT.
   `b + 1`, `x < 10` không cần `as`. Implement xong, regression 20/20, fixpoint đang chạy.
2. **ADT v2 — multi-field variant** (`Rect(i64,i64)`): mỗi field ≤8B (scalar/ptr), box
   size = 8 + max_fields*8, OP_SET_FIELD/GET_FIELD field idx 1..n. typecheck gán type cho
   mọi binding sub-pattern. (đã thiết kế ở docs/next-step-15-sub-1.md; đã revert checkpoint).
3. **ADT v2 — str / >8B payload**: cần máy 16-byte field (offset+size theo type thật).
4. **RFC 0006 — numeric & arithmetic semantics (NGHIÊM TÚC, gộp BUG#33+#34)**: phép toán
   là nền tảng → xử lý kỹ. Gồm: (a) **unsigned div/mod** dùng DIV thay vì IDIV signed cho
   u32/u64 (BUG#34.1); (b) **float arithmetic** phát OP_FADD/FSUB/FMUL/FDIV theo f32/f64 +
   cvt int→float (BUG#33); (c) **lan kiểu 2 chiều** từ kiểu khai báo xuống toán hạng
   (BUG#34.3); (d) **chính sách narrowing/đổi dấu** khi gán khác kiểu — cast tường minh /
   cảnh báo thay vì reinterpret thầm lặng (BUG#34.4); (e) shift/compare signed-aware;
   overflow policy. Mỗi phần: test + giữ fixpoint. knowledge/bugs.md BUG#33+#34.
   - **Test harness:** `tests/arith/` (gen + oracle Python) — `MODE=int|float|mixed N=100000`.
     Chạy `int` NGAY sau fixpoint RFC0005 (kỳ vọng lộ BUG#34 unsigned div); `float`+`mixed`
     chạy SAU khi fix float (BUG#33). So sánh bit-exact, oracle fixed-width chuẩn.
5. **`Self` type** (RFC riêng): cho return type / interface (`fn clone(self) -> Self`).
6. **Rewrite module aspirational** (iter/json/log/net/fmt…) sang grammar — cần enum/Self.

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
