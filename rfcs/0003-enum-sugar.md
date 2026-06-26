# RFC 0003 — `enum` indent-style sugar

- **Status:** Implemented (2026-06-26)
- **Author:** self-host team
- **Tracking:** next-step-15 T2
- **Liên quan:** GRAMMAR.ebnf §EnumDecl, RFC 0002 (cùng pattern parser-only sugar)

## 1. Motivation

Sum type hiện chỉ viết được dạng một dòng:
```axiom
type Color = Red | Green | Blue
type Shape = Circle(i64) | Rect(i64, i64)
```
Với nhiều variant (vd token kinds, AST node kinds, JSON value) một dòng dài khó đọc.
Nhiều module aspirational (`std.json`, `std.iter`, `std.log`) viết `enum Name:` indent-
style. Sugar này cho phép viết variant theo từng dòng thụt lề, đồng nhất với
struct/interface (đều indent-based):
```axiom
enum Color:
    Red
    Green
    Blue

enum Shape:
    Circle(i64)
    Rect(i64, i64)
```

## 2. Design

`enum Name:` + INDENT + danh sách `TypeVariant` mỗi dòng, desugar **trong PARSER**
thành đúng AST của sum type:
`NODE_TYPE_ALIAS_DECL(payload=Name)` → con `NODE_SUM_TYPE` → các con `NODE_VARIANT_DECL`.

Vì AST sinh ra GIỐNG HỆT `type Name = A | B(T)` (đã chạy đủ: construct + `match`),
resolver / typecheck / codegen / match **không thấy node kind mới** → không có code
path mới ở backend. Giống RFC 0002, đây là thay đổi parser-only, rủi ro fixpoint thấp.

**Cài đặt:**
1. `token.ax`: thêm `TK_ENUM = 86`.
2. `lexer.ax`: keyword `"enum" -> TK_ENUM`.
3. `parser.ax`: `parse_enum_decl(is_pub)` — mirror `parse_type_alias_decl` cho header
   (tên + optional `GenericParams`) nhưng đọc variant từ INDENT block bằng
   `parse_type_variant` + `expect_newline`, build cùng node như sum type.
4. Dispatch: thêm nhánh `TK_ENUM` trong `parse_program` (cả `pub` và non-`pub`).

Generics được hỗ trợ (`enum Opt[T]:`) y như `type` generic (FLAG_IS_GENERIC).

## 3. Tại sao an toàn cho self-host

Node desugar GIỐNG HỆT sum type đang chạy. Compiler tự host hiện KHÔNG dùng `enum`
(vẫn viết `type X = ...`) → stage1→stage2 không đổi hành vi → fixpoint phải giữ.
`enum` chỉ là cú pháp mới cho người dùng / module tương lai.

## 4. Drawbacks / giới hạn

- Chỉ top-level (như sum type hiện tại). Không có discriminant value tường minh
  (`Red = 1`) — variant không-payload map về tag tuần tự, giống sum type. Nếu cần
  giá trị cố định → RFC riêng.
- Không nested. Không method trong enum body (chỉ variant) — method dùng UFCS như
  với sum type bình thường.

## 5. Alternatives

- Giữ nguyên sum type một dòng: kém đọc với nhiều variant.
- `enum {}` braces: đi ngược triết lý indent-based; bị CLAUDE.md cấm bịa syntax.

## 6. Test plan / DoD

- [x] `bin/t_enum.ax`: `enum` không-payload + có-payload, construct + `match` → giá trị đúng.
- [x] sum type `type X = A | B` cũ vẫn chạy (không regression).
- [ ] `scripts/regression_repros.sh` PASS.
- [ ] `scripts/verify_bug29_selfhost.sh`: fixpoint stage3==stage4 GIỮ.
