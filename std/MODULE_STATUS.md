# AXIOM stdlib — Module Status (2026-06-26)

Phân loại các module trong `std/` theo việc compiler hiện tại có compile được không.
Đo bằng `axc_stage1 dump-air std/<file>.ax -O0` (đếm lỗi parse/typecheck).

## ✅ Grammar-conformant (compile sạch, 0 lỗi) — STDLIB THẬT

| Module | Dùng bởi self-host build? |
|---|---|
| `std.string` | ✅ (toàn bộ compiler) |
| `std.io` | ✅ (main_air, lsp) |
| `std.collections` | ✅ (lsp, air pipeline) |
| `std.os` | ✅ (qua concatenated) |
| `std.result`, `std.option` | ✅ (payload/unwrap) |
| `std.process`, `std.scheduler`, `std.runtime`, `std.alloc` | ✅ |
| `std.collections_test` | ✅ (test) |

Self-host build chỉ `import std.string / std.io / std.collections` (+ các file concatenated).

## ⚠️ ASPIRATIONAL — viết bằng DIALECT NGOÀI GRAMMAR (CHƯA compile được)

Các file dưới đây dùng cú pháp KHÔNG có trong `docs/GRAMMAR.ebnf`: braces `{}`,
`enum {...}`, Go-style `fn (recv)`, associated `type Item`, `Self` type. Chúng là
**stub nguyện vọng** (chưa từng compile), **KHÔNG được build/import**, và sẽ báo
hàng trăm lỗi parse nếu thử compile. ĐỪNG nhầm là stdlib hỏng.

| Module | Lỗi (dump-air -O0) | dòng braces/enum |
|---|---|---|
| `std.net` | 713 | 54 |
| `std.iter` | 571 | 55 |
| `std.log` | 254 | 25 |
| `std.json` | 247 | 19 |
| `std.mem` | 236 | 26 |
| `std.math` | 218 | 20 |
| `std.fmt` | 160 | 19 |
| `std.gpu` | 150 | 14 |
| `std.compiler` | 111 | 11 |
| `std.ffi` | 74 | 6 |
| `std.crypto` | 70 | 8 |
| `std.cli` | 291 | 23 |
| `std.arch` | 287 | 29 |

## Policy để "kích hoạt" một aspirational module

AXIOM là **indent-based** (`:` + INDENT), sum type `type X = A | B`, method =
`fn m(self: T)` / sugar `fn m(self)` (RFC 0002) + UFCS `x.m()`, interface `interface X:`.

Muốn dùng một module aspirational, CHỌN một trong hai:
1. **Viết lại theo grammar** (đổi braces→indent, `enum`→sum type, Go-method→`self`-method,
   bỏ `Self`/assoc-type hoặc thay bằng generics). Mỗi file là một rewrite lớn.
2. **RFC mở rộng grammar** TRƯỚC (vd RFC cho `enum`/`Self`/associated type), cập nhật
   `docs/GRAMMAR.ebnf`, implement, rồi mới giữ syntax đó. CLAUDE.md: KHÔNG bịa syntax
   ngoài spec; KHÔNG thêm braces/enum chỉ để "chiều" các file này.

KHÔNG có thời hạn ép buộc — đây là backlog tính năng, không phải bug chặn self-host.
Theo dõi tại `docs/next-step-15.md`.
