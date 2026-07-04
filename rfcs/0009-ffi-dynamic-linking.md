# RFC 0009 — FFI dynamic linking (tiêu thụ & sản xuất DLL/.so/.dylib)

- **Status:** Accepted (2026-07-03, user-approved) — P1 frontend prototyped; linker
  integration scoped (§10, large careful change deferred to a focused effort)
- **Author:** self-host team
- **Tracking:** #FFI — follows BUG#49 (bare function pointers), p12-t04 (dynamic-linking task)
- **Liên quan:** parser (`parse_extern_decl`), AST/symbol metadata, resolver
  (`NODE_IMPORT_DECL`), linker.ax (`get_dll_for_symbol`, `.idata`/IAT emit),
  x86_coff.ax, x86_elf64.ax, cgen.ax
- **Blocks:** ứng dụng lớn (codebase phân mảnh thành nhiều .dll), tiêu thụ thư viện
  C bên thứ ba (OpenSSL, SQLite, libm…), AXIOM-as-a-library cho ngôn ngữ khác,
  phân phối binary standalone.

---

## 1. Motivation

Vấn đề chiến lược (user-raised, 2026-07-03):

1. **Chương trình AXIOM không chạy độc lập.** Binary native hiện phụ thuộc
   `ax_runtime.dll` + `ucrtbase.dll` + `kernel32.dll` — nhưng việc gán DLL là
   **cứng-hóa** bằng whitelist tên symbol trong `linker.ax:get_dll_for_symbol`.
2. **App lớn ⇒ codebase khổng lồ.** Không thể mãi nối-toàn-chương-trình
   (whole-program concatenation) vào một translation unit. Phải tách thành nhiều
   thư viện động, biên dịch riêng, link động.
3. **AXIOM sẽ tự sinh DLL.** Khi chia app thành module `.dll`, chính output của
   `axc` phải: (a) **xuất** symbol (export table / `.dynsym`), và (b) một module
   AXIOM khác phải **nhập** và gọi được các symbol đó.
4. **Tiêu thụ thư viện C bên thứ ba** (`anylib.dll`) — hiện **không làm được** qua
   native self-link (mọi symbol lạ bị mặc định gán về `ucrtbase.dll` → unresolved).

Đây là mắt xích còn thiếu để AXIOM từ "đồ chơi single-binary" thành **hệ sinh thái
thư viện thực thụ**.

## 2. Hiện trạng (đã khảo sát code, 2026-07-03)

### 2.1 Nhập (import) — không có DLL-binding trong source
- `extern "C" fn foo(...) -> T` (`parse_extern_decl`, parser.ax:1370): parser đọc
  chuỗi ABI `"C"` rồi **vứt đi**, chỉ giữ tên + chữ ký. **Không có trường DLL.**
- `import` (`parse_import_decl`, parser.ax:1418) chỉ nhận **dotted identifier**
  (`expect(TK_IDENT)`) — `import "anylib.dll"` là **parse error**. Danh sách
  `import X { a, b }` bị parse-nhưng-trơ (xem KNOWN-GAP bugs.md 2026-07-03).
- Linker gán DLL bằng **whitelist tên cứng** (`get_dll_for_symbol`, linker.ax:496):
  Win32-whitelist→`kernel32.dll`, `ax_*` whitelist→`ax_runtime.dll`, còn lại→
  `ucrtbase.dll`. Không đọc tên DLL từ source; symbol lạ = sai DLL.

### 2.2 Xuất (export) — spec có ý định, chưa nối dây
- Spec "08. Linker riêng.md": hàm `extern "C"` giữ raw symbol trong Export Table.
- Design task p12-t04 mô tả `DynLinkSpec{Library, Symbols}` + PLT/GOT (ELF) / IAT
  (PE) / `LC_LOAD_DYLIB` (Mach-O) + `#[export]` annotation — **chưa implement**
  (path Go `codegen/native/dynlink.go` chưa tồn tại). `std/ffi/` rỗng.

### 2.3 Đường vòng khả dụng hôm nay
Chỉ có **C-backend**: `axc emit-c prog.ax -o out.c` rồi `gcc out.c -L. -lanylib`.
Native self-link **không** tiêu thụ được DLL bên thứ ba.

## 3. Design

### 3.1 Cú pháp nhập — `extern` mang tên thư viện

Phương án chọn (khớp `extern "C"` sẵn có, đối xứng ABI-string):

```axiom
extern "C" from "anylib.dll" fn foo(a: i32, b: ptr[u8]) -> i32
extern "C" from "libm.so.6"  fn cos(x: f64) -> f64
```

- `from "<lib>"` **tùy chọn**; vắng mặt ⇒ giữ hành vi whitelist hiện tại (tương
  thích ngược). Tên lib là **string literal** (giữ nguyên, không mangle).
- Gom nhiều extern cùng lib qua block (sugar, giai đoạn 2):
  ```axiom
  extern "C" from "sqlite3.dll":
      fn sqlite3_open(path: ptr[u8], db: ptr[ptr[void]]) -> i32
      fn sqlite3_close(db: ptr[void]) -> i32
  ```

**Bác bỏ** `import "anylib.dll"`: `import` mang ngữ nghĩa *module `.ax` nội bộ*
(namespace, resolution theo path thư mục). Trộn DLL vào đó phá layering
frontend/linker. FFI thuộc về `extern`, không phải `import`.

### 3.2 Cú pháp xuất — `#[export]`

```axiom
#[export] pub fn ax_add(a: i64, b: i64) -> i64:
    return a + b
```
Xuất raw symbol `ax_add` vào Export Table (PE) / `.dynsym STB_GLOBAL` (ELF).
Sinh kèm import library `.lib` (Windows) để module AXIOM khác link.

Build shared: `axc build --shared math.ax -o math.dll` (⇒ không cần `main`, sinh
export table thay vì entrypoint EXE).

### 3.3 Luồng dữ liệu qua pipeline

1. **Parser:** `from "<lib>"` → lưu lib-name-id vào node extern (dùng `extra_idx`
   hoặc child NODE_STRING_LIT). `#[export]` → set `SYM_FLAG_EXPORT`.
2. **Symbol metadata:** mỗi `Symbol` extern mang `dll_name_id: u32` (0 = whitelist
   mặc định). Cần nới `Symbol` struct — **blast radius resolver + typecheck**.
3. **Linker (`get_dll_for_symbol`):** nếu symbol có `dll_name_id != 0` → dùng tên
   đó; ngược lại rơi về whitelist cũ. Thay bảng cứng bằng **per-symbol attribution**.
4. **`.idata`/IAT (PE):** hiện chỉ sinh cho kernel32/ucrtbase/ax_runtime cố định.
   Phải tổng quát: **một `IMAGE_IMPORT_DESCRIPTOR` cho mỗi DLL** xuất hiện trong tập
   extern, mỗi hàm một IAT slot, `CALL [RIP + func@IAT]`.
5. **ELF:** `.plt`/`.got.plt`/`.dynamic` với `DT_NEEDED` cho mỗi `.so` (theo p12-t04).
6. **Export (PE):** `IMAGE_EXPORT_DIRECTORY` + `.edata`; sinh `.lib` import library.

### 3.4 Runtime dynamic loading (giai đoạn 3, tùy chọn)
`std/ffi/dyn.ax`: wrapper `LoadLibraryA`/`GetProcAddress` (Win) & `dlopen`/`dlsym`
(POSIX) → trả con-trỏ-hàm gọi qua BUG#49 `MACH_CALL_INDIRECT`. Cần thêm
`LoadLibraryA`/`GetProcAddress` vào kernel32-whitelist (hoặc dùng 3.1). Cho phép
nạp plugin động không cần import-time binding.

## 4. Alternatives

- **`import "anylib.dll"`** (user gợi ý): bác bỏ — lẫn lộn module-import với FFI,
  phá layering (xem 3.1).
- **Chỉ dựa C-backend + gcc:** không self-hosting-friendly, phụ thuộc toolchain
  ngoài, mất kiểm soát deterministic linking. Giữ làm escape hatch, không phải giải.
- **`@link("anylib")` annotation trên hàm:** tương đương `from`, nhưng kém rõ ràng
  về ABI; `from` gắn liền `extern "C"` mạch lạc hơn.
- **Static archive (.a/.lib) linking:** bổ trợ, không thay dynamic — RFC riêng.

## 5. Drawbacks

- Nới `Symbol` struct + luồng lib-name → blast radius resolver/typecheck/linker;
  bắt buộc regression + fixpoint.
- Sinh `.idata` đa-DLL + export table + `.lib` là khối lượng lớn (PE/ELF/Mach-O).
- ABI stability: spec (RFC-12) cảnh báo truyền struct AXIOM giữa 2 `.dll` do 2 bản
  `axc` khác nhau là **không an toàn** — biên FFI phải qua `extern "C"` + type C.

## 6. Migration & Compatibility

- `from` là **opt-in**; extern không có `from` giữ nguyên hành vi whitelist ⇒
  không phá code hiện tại (stdlib, runtime).
- `#[export]`/`--shared` là tính năng mới, không đụng đường EXE cũ.
- Whitelist cứng trong `get_dll_for_symbol` được giữ làm **fallback**, gỡ dần sau.

## 7. Kế hoạch thực hiện (phân pha, mỗi pha có gate)

- **P0 — Khảo sát & RFC** (xong): tài liệu này + KNOWN-GAP bugs.md.
- **P1 — Nhập tối thiểu (PE):** parse `extern "C" from "x.dll"` → symbol.dll_name →
  `get_dll_for_symbol` per-symbol → `.idata` đa-DLL. Test: gọi 1 hàm từ `.dll` tự
  tạo bằng gcc. Gate: regression + fixpoint (đụng parser/symbol/linker).
- **P2 — Xuất (PE):** `#[export]` + `--shared` → export table + `.lib`. Test:
  module A xuất `ax_add`, module B `extern from "A.dll"` gọi được.
- **P3 — ELF/Mach-O:** PLT/GOT + `DT_NEEDED` (Linux), `LC_LOAD_DYLIB` (macOS).
- **P4 — Runtime dyn-load:** `std/ffi/dyn.ax` (LoadLibrary/dlopen) + plugin demo.

## 8. Open questions

- Phân giải đường dẫn DLL: chỉ tên (loader tự tìm) hay cho path tuyệt đối/tương đối?
- Versioned `.so` (`libm.so.6`) & symlink resolution trên Linux.
- Calling convention không-C (stdcall/fastcall Win32 legacy) — cần không? (mặc định
  chỉ hỗ trợ System V / MS x64 C ABI.)
- Struct-by-value qua biên FFI: giới hạn theo RFC-0001 (16-byte aggregate) hay cấm,
  bắt buộc `ptr`?

## 9. Success criteria

- `axc build app.ax` sinh EXE + `axc build --shared lib.ax -o lib.dll` sinh DLL có
  export table hợp lệ (`dumpbin /exports` / `nm -D`).
- Module AXIOM thứ hai `extern "C" from "lib.dll"` gọi được symbol xuất.
- Gọi được 1 hàm C bên thứ ba (vd `sqlite3_open` từ `sqlite3.dll`) qua native
  self-link, **không** cần C-backend/gcc.
- Toàn bộ deterministic, qua regression + self-host fixpoint.

## 10. Implementation notes — P1 COFF integration (reverse-engineered 2026-07-03)

Khảo sát `linker.ax` cho thấy phần `.idata`/thunks là hàm **~600 dòng** hardcode
cứng đúng 3 DLL (kernel32/ax_runtime/ucrtbase) qua `has_X` booleans + offset-math
tuần tự, luồn qua CẢ COFF `.idata` LẪN ELF `.dynamic`/PLT. Đây là hàm **chịu tải
nặng nhất toàn toolchain** (mọi build, kể cả tự-build compiler, phụ thuộc) → theo
CLAUDE.md §16 phải sửa **có kiểm chứng tăng dần**, KHÔNG blind-rewrite.

**Hợp đồng đã xác nhận (điểm mấu chốt để tích hợp an toàn):**
- **Reloc resolution là name-based** (`dyn_sym_names`/`dyn_sym_rvas`, ~L1612): mọi
  symbol có thunk trong `dyn_sym_names` được patch tự động. ⟹ user-DLL thunk chỉ
  cần push vào `dyn_sym_names` là hoạt động — không cần logic reloc riêng.
- **`thunk_count` (~L896)** = tổng imports 3 bucket, điều khiển `dyn_sym_rvas` alloc
  + `thunks_offset`/`thunks_rva`/`idata_rva` (~L928-932). ⟹ user imports phải được
  cộng vào count TRƯỚC L896 để layout-math tự chảy qua.
- **IDT phải là MỘT mảng null-terminated liền** (loader đọc tới entry null) → mọi
  DLL (built-in + user) chung một IDT; `K` (số entry, ~L957) phải += số user DLL.
- Extern `from` symbol: `x86_resolve_sym_name` trả tên RAW (không prefix `ax_`), nên
  `target_name` khớp trực tiếp `intern.get(name_id)` của binding.

**Điểm chèn (7 nơi) cho path COFF:** (1) collection loop L859-895 route symbol có
binding vào user-bucket thay vì ucrtbase; (2) thunk_count += user_total; (3) `K` +=
#user_dll; (4) IAT/ILT/hint-name/DLL-name/IDT emit cho mỗi user bucket (offset qua
mảng vì tên DLL dài biến thiên); (5) thunks + push dyn_sym.

**Truyền dữ liệu:** resolve `SymbolTable.dll_binds` (name_id→dll_id) thành cặp CHUỖI
`[sym_name, dll_name, ...]` tại call-site (main_air, nơi có symtable+pool) rồi truyền
vào hàm linker — tránh cấp pool cho linker.

**Fixpoint:** KHÔNG cần byte-identical với binary cũ; chỉ cần self-reproduce
(stage_n==stage_n+1) — nên cho phép reorder bucket. ELF `from`-DLL out-of-scope P1
(giữ mapping built-in libax_runtime.so/libc.so.6).

**Ước lượng:** ~250 dòng mới, cần 3-4 vòng verify (8s build + fixpoint + t_ffi với
DLL gcc-built). Là công việc tập trung riêng, không ghép cùng task khác.
