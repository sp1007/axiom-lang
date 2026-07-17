---
name: ffi-dynamic-linking-priority
description: Ưu tiên chiến lược — FFI/dynamic-linking để tiêu thụ & sản xuất DLL (RFC 0009)
metadata: 
  node_type: memory
  type: project
  originSessionId: ed12f2e7-f4ab-41c0-9bb3-940acfa7aaec
---

Ngày 2026-07-03 user nêu ưu tiên chiến lược: **chương trình AXIOM không chạy độc lập** (phụ thuộc ax_runtime.dll/ucrtbase.dll, DLL-binding cứng-hóa), app lớn ⇒ codebase khổng lồ không thể mãi whole-program concatenation, **chính AXIOM sẽ tự sinh DLL** nên phải nhập/gọi được chúng + tiêu thụ được DLL C bên thứ ba (`anylib.dll`).

**Hiện trạng (khảo sát code):**
- `extern "C" fn` là cơ chế FFI duy nhất; parser vứt ABI-string, không có trường DLL.
- `import` chỉ nhận dotted-ident (`import "x.dll"` = parse error). KHÔNG có `import "dll"` như user tưởng.
- DLL gán bằng whitelist tên cứng `linker.ax:get_dll_for_symbol` (kernel32/ucrtbase/ax_runtime); symbol lạ → mặc định ucrtbase → unresolved.
- Xuất DLL (`#[export]`/`--shared`) + `.idata` đa-DLL: spec+p12-t04 có ý định, **chưa implement**; `std/ffi/` rỗng.
- Đường vòng hôm nay: chỉ C-backend `emit-c` + gcc `-lanylib`.

**Giải pháp: [[RFC 0009 — FFI dynamic linking]]** (`rfcs/0009-ffi-dynamic-linking.md`).
- Nhập: `extern "C" from "anylib.dll" fn foo(...)` (bác bỏ `import "dll"` — lẫn module-import với FFI, phá layering).
- Xuất: `#[export] pub fn` + `axc build --shared x.ax -o x.dll`.
- Phân pha: P1 nhập-PE (`.idata` đa-DLL, per-symbol DLL attribution thay whitelist) → P2 xuất-PE (export table + `.lib`) → P3 ELF/Mach-O (PLT/GOT, DT_NEEDED, LC_LOAD_DYLIB) → P4 runtime dyn-load (`std/ffi/dyn.ax` LoadLibrary/dlopen).
- Blast radius: nới `Symbol` struct mang `dll_name_id` → resolver/typecheck/linker → **bắt buộc regression + fixpoint**.

**Why:** mắt xích còn thiếu để AXIOM thành hệ sinh thái thư viện thực thụ (không phải single-binary toy), phục vụ self-hosting + app lớn.
**How to apply:** khi rảnh gate closures, bắt đầu P1. Đây là RFC-gated (thay đổi cú pháp, CLAUDE.md §13) — phải duyệt RFC trước khi code. Liên quan [[next-step-16-fnptr-shipped]] (BUG#49 fn-ptr = nền cho runtime dyn-load qua MACH_CALL_INDIRECT).

**🎉 P2 SHIPPED (2026-07-04, commit 475566d):** xuất DLL `#[export] pub fn` +
`axc build --shared x.ax -o x.dll` → `IMAGE_EXPORT_DIRECTORY`/`.edata`. AXIOM giờ
**vừa sản xuất vừa tiêu thụ** DLL. Mangling: export ghi tên SẠCH (`ax_add`), EAT trỏ
symbol `ax_`+tên (KHÔNG đụng codegen); khớp `extern "C" from` của bên nhập → AXIOM↔AXIOM.
`linker_build_pe_headers` guarded (edata=0 ⇒ EXE byte-identical). Test: axmath.dll gọi
từ C (LoadLibrary+GetProcAddress 42/42) VÀ từ AXIOM (`extern from` exit 84). fixpoint
265606A1, 0 hồi quy. Chi tiết RFC 0009 §13. Còn: DllMain-init cho export dùng heap; P3 ELF.

**🎉 STATIC LIB [[rfc0011-static-libs]] P1+P2 SHIPPED (2026-07-04, commit 390d3df):**
`--staticlib` xuất COFF `!<arch>` (.lib) + `-l x.lib` link tĩnh (code copy vào EXE, KHÔNG
phụ thuộc runtime). Test: app_static.exe chạy từ temp-dir trống → exit 84.

**🎉 P1 SHIPPED (2026-07-04, commit 488c283):** `extern "C" from "x.dll"` nhập-PE
HOẠT ĐỘNG end-to-end. Frontend §11 re-applied + linker N-DLL `.idata` **additive**
(3 bucket built-in giữ nguyên từng dòng; user DLL qua `UserDllBucket` vec + vòng lặp
chèn sau mỗi section — 0 user DLL ⇒ dormant ⇒ byte-identical ⇒ fixpoint tự giữ).
Luồng: `resolver.SymbolTable.dll_bind_syms/dlls` → main_air `pool.get` →
`AxiomLinker.user_dll_syms/names` → routing `user_dll_for_symbol` → `user_bucket_find_or_add`.
Test `tests/ffi/` (gcc DLL ffi_add/ffi_mul → exit 84; objdump IDT#4; xoá DLL →
0xC0000135). Verify tăng-dần A/B/C, regression y hệt baseline (0 hồi quy). **Daily
driver axc_native.exe promoted (1926656 bytes).** Chi tiết impl ở RFC 0009 §12.
**Còn P2 (xuất `--shared`/`#[export]` + export table + `.lib`) + P3 (ELF/Mach-O).**

**🔬 FRONTEND VALIDATED + reverted (2026-07-04, RFC 0009 §11) [lịch sử, đã dùng lại ở P1]:** đã implement CẢ 4 lớp frontend (parser `from`, FLAG_FROM_DLL=8192 + AstNode.extra_idx, SymbolTable.dll_bind_syms/dlls parallel vecs, resolver pre_define recording), compiler **tự-build sạch 2 vòng** với chúng → rồi **revert giữ cây sạch** (feature atomic, vô giá trị+footgun tới khi linker xong; tiền lệ RFC 0010). Cơ chế chính xác + tái dùng ghi ở **RFC 0009 §11**. **KHẲNG ĐỊNH LUỒNG:** `-self-link` → compile_native_binary chỉ GHI object (`axiom_temp.obj`), rồi `main_air.ax:851` gọi **`axiom_linker_link` (linker.ax:793, ~870 dòng)** build `.idata` → plan §10 nhắm ĐÚNG file. **Còn ĐÚNG 1 việc: N-DLL `.idata` trong hàm 870 dòng đó** (§10 7-điểm-chèn + AxiomLinker struct:443 thêm field cặp-chuỗi, populate ở main_air). = focused effort riêng, byte-perfect PE + fixpoint, KHÔNG rush cuối session (CLAUDE.md §16).
