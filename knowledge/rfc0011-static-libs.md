---
name: rfc0011-static-libs
description: "RFC 0011 static libraries (.lib) + precompiled-stdlib cache — P1+P2 shipped, P4 stdlib-cache là đích lớn"
metadata: 
  node_type: memory
  type: project
  originSessionId: ed12f2e7-f4ab-41c0-9bb3-940acfa7aaec
---

**RFC 0011** (`rfcs/0011-static-libraries.md`, Draft) — user nêu 2026-07-04: xuất/dùng
thư viện liên kết TĨNH; tối ưu lớn = precompile stdlib vào `std.lib` để build user KHÔNG
recompile toàn stdlib mỗi lần (chi phí cố định lớn nhất của build nhỏ); + cơ chế staleness
(source đổi → rebuild `.lib`).

**Đã SHIP (commit 390d3df):**
- **P1 producer:** `axc build --staticlib x.ax -o x.lib` → COFF `!<arch>` (magic +
  first-linker-member symbol index BE + 1 member "axiom.o") + manifest `.manifest`
  (djb2 object bytes). `axiom_write_static_lib` trong linker.ax. Verify `ar t`/`nm`.
- **P2 consumer:** `-l <path>` (main_air gom extra_libs → linker inputs); linker input-loop
  nhận `!<arch>`, pull mọi member không-đặc-biệt (bỏ "/" index + "//" longnames), parse COFF,
  dùng lại reloc name-based. Test app_static.exe exit 84, KHÔNG phụ thuộc .lib lúc chạy.

**✅ P3 SHIPPED (commit 0b148f3):** `axc build --staticlib` idempotent — djb2-hash src
(post-concat), nếu `<out>.manifest` khớp + `.lib` tồn tại → in "[lib] up to date" + exit
TRƯỚC lex (bỏ toàn pipeline). Edit src → hash đổi → rebuild. Đáp thẳng yêu cầu "rebuild
khi source đổi". `axc lib` verb hoãn (gộp vào --staticlib); multi-file hash = P4.
⟹ **user-library story XONG**: build lib 1 lần → .lib → app link (-l) KHÔNG recompile lib.

**CÒN LẠI (đích lớn = tiết kiệm thời gian build):**
- **P4 stdlib-cache (headline benefit, KHÓ):** cần **SEPARATE COMPILATION** — typecheck user
  code theo *chữ ký* stdlib (interface `.axi`) KHÔNG cần body, rồi link `std.lib`. Đây là hạ
  tầng module-interface mới (lớn nhất roadmap sau codegen). Hiện AXIOM whole-program-concat
  (main_air `concatenate_stdlib` splice source stdlib vào MỌI TU → recompile mỗi build).
- **P5 generic separate-comp:** monomorphize generic trong archive theo yêu cầu.

**Why:** giải quyết trực tiếp "compile chậm" — stdlib được lex/parse/typecheck/codegen lại
từ đầu mỗi build. **How to apply:** P1+P2 (archive I/O) xong = nền; P4 mới cho lợi ích tốc
độ nhưng cần separate-compilation trước. Liên quan [[ffi-dynamic-linking-priority]] (P2 dùng
chung mangling `ax_` cho intra-AXIOM link), [[project_next_step_13]] (perf).
