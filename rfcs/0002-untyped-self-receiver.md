# RFC 0002 — Untyped `self` receiver sugar

- **Status:** Implemented (2026-06-26)
- **Author:** self-host team
- **Tracking:** next-step-15 T1
- **Liên quan:** GRAMMAR.ebnf §Param/SelfParam

## 1. Motivation

Method hiện phải khai báo receiver có type tường minh:
```axiom
struct Pt:
    x: i64
    fn getx(self: ptr[Pt]) -> i64:   # boilerplate: phải lặp lại ptr[Pt]
        return self.x
```
Compiler tự host viết `mut self: ptr[Parser]` hàng trăm lần. Sugar giảm boilerplate:
```axiom
    fn getx(self) -> i64:            # type suy ra = ptr[Pt]
        return self.x
    fn set(mut self, v: i64):        # mut self -> mut ptr[Pt]
        self.x = v
```

## 2. Design

`self` / `mut self` không type, **chỉ trong thân struct**, desugar thành
`self: ptr[EnclosingStruct]` (giữ `mut` nếu có). Mọi thứ sau đó (resolver,
typecheck, codegen, UFCS call `p.getx()`) y hệt typed-self ĐÃ hoạt động — nên
chỉ cần đổi PARSER, không đụng tầng sau.

**Cài đặt (parser.ax):**
1. Thêm field `Parser.current_struct: u32` (intern id struct đang parse; 0 = none).
2. `parse_struct_decl`: set `current_struct = <tên struct>` trước khi parse body,
   reset 0 sau DEDENT (lồng struct không hỗ trợ — đủ dùng).
3. `parse_func_decl` vòng param: nếu IDENT == `self` và token kế KHÔNG phải `:` →
   tổng hợp node type `ptr[current_struct]` = `NODE_GENERIC_TYPE[ TYPE_EXPR("ptr"),
   TYPE_EXPR(current_struct) ]` (đúng dạng node mà `self: ptr[Pt]` sinh ra), gắn vào param.
   Ngược lại giữ nguyên đường `expect(':') + parse_type_expr`.

## 3. Tại sao an toàn cho self-host

Node tổng hợp GIỐNG HỆT node của `self: ptr[Pt]` (đã verify chạy: u2.exe exit 7).
Resolver/typecheck/codegen không phân biệt được → không có code path mới ở backend.
Compiler hiện viết typed-self khắp nơi → KHÔNG dùng sugar này, nên stage1→stage2
không đổi hành vi; fixpoint phải giữ. (Có thể dần refactor compiler dùng sugar sau,
nhưng KHÔNG bắt buộc.)

## 4. Drawbacks / giới hạn

- Chỉ áp dụng method trong struct body (cần `current_struct`). Top-level `fn f(self)`
  không có struct cha → vẫn lỗi (đúng: `self` ngoài method vô nghĩa).
- Interface method-sig (`parse_method_sig`) CHƯA áp dụng (separate fn) — follow-up nếu cần.
- Nested struct (struct trong struct) không hỗ trợ current_struct lồng — hiện grammar
  không có nested struct decl nên không phải vấn đề.

## 5. Alternatives

- `Self` type keyword: tổng quát hơn (cả return type, interface) nhưng cần thêm
  type-resolution cho `Self` → lớn hơn, RFC riêng. Sugar `self` đủ cho 90% nhu cầu.
- Giữ nguyên (typed self): không boilerplate-saving.

## 6. Test plan / DoD

- [x] `bin/t_method.ax`: `fn getx(self)`, `fn bump(mut self)` → giá trị đúng + mut ảnh hưởng caller.
- [x] typed-self cũ vẫn chạy (không regression).
- [x] regression_repros.sh PASS (14/14).
- [x] verify_bug29_selfhost.sh: fixpoint stage3==stage4 GIỮ (fc5e6f48, 2026-06-26).
