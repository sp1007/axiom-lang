# FFI import test (RFC 0009 P1)

End-to-end verification that an AXIOM native binary can import and call symbols
from a third-party DLL declared with `extern "C" from "x.dll"`, resolved through
the compiler's own N-DLL `.idata` emission (no C-backend / external linker).

Not wired into the auto-harness because it needs a gcc-built DLL. Manual steps
(from this directory, MSYS2 ucrt64 gcc + the daily-driver compiler):

```sh
gcc -shared -o axffi_lib.dll axffi_lib.c
../../bin/axc_native.exe build t_ffi.ax -o t_ffi.exe -self-link -O1
cp ../../bin/ax_runtime.dll .
./t_ffi.exe ; echo "exit=$?"        # expect exit=84
```

Checks:
- `exit=84` — the call `(40+2)*2` returned the DLL's real result.
- `objdump -p t_ffi.exe` lists a 4th import descriptor `axffi_lib.dll` with
  hint-name entries `ffi_add` / `ffi_mul`.
- Negative: rename `axffi_lib.dll` away → the process fails to load with
  `0xC0000135` (STATUS_DLL_NOT_FOUND), proving genuine dynamic binding.

## Export test (RFC 0009 P2)

`axmath.ax` exports `ax_add`/`ax_mul` into a DLL; consumed two ways:

```sh
# Build the AXIOM DLL (IMAGE_EXPORT_DIRECTORY)
../../bin/axc_native.exe build axmath.ax -o axmath.dll --shared -self-link -O1
cp ../../bin/ax_runtime.dll .

# (a) Call from C via LoadLibrary/GetProcAddress
gcc -o host.exe host.c
./host.exe ; echo "exit=$?"        # prints 42/42, expect exit=0

# (b) Call from another AXIOM binary via `extern "C" from` (P1+P2 interop)
../../bin/axc_native.exe build t_useaxmath.ax -o t_useaxmath.exe -self-link -O1
./t_useaxmath.exe ; echo "exit=$?" # expect exit=84

objdump -p axmath.dll | grep -A12 "Export Tables"   # ax_add / ax_mul, base 1
```

## Static library test (RFC 0011 P1+P2)

`--staticlib` emits a COFF `!<arch>` archive; `-l x.lib` statically links it (code
copied into the EXE, no runtime dependency).

```sh
# Produce a stdlib-free static lib (pure fns → symbols ax_ax_add / ax_ax_mul)
../../bin/axc_native.exe build axmath.ax -o axmath_ns.lib --staticlib --no-stdlib -self-link -O1
ar t axmath_ns.lib          # lists member axiom.o
nm axmath_ns.lib | grep ax_ax_   # symbol index has ax_ax_add / ax_ax_mul

# Statically link it into an app that only declares those symbols
../../bin/axc_native.exe build app_static.ax -o app_static.exe -l axmath_ns.lib -self-link -O1
objdump -p app_static.exe | grep "DLL Name"   # only kernel32/ax_runtime/ucrtbase — NO axmath
./app_static.exe ; echo "exit=$?"             # expect exit=84, runs with no .lib present
```

## Import-driven auto-library test (RFC 0011 P4 inc3b)

With `--auto-lib`, `import x` finds-or-builds a fresh `x.lib` (no `-l`). The compiler
registers the library's public functions from its `__axiom_iface` member (typecheck
against the interface, no source recompile) and the linker pulls the code from the `.lib`.
The flag is opt-in: without it, imports stay on the source path (AXIOM can't yet tell an
app-module from a standalone library — an app-module may collide by name or need stdlib).

```sh
# One command: --auto-lib builds imp_mymath.lib from source on first sight, caches it
# (manifest hash), reuses it while unchanged, and rebuilds it when imp_mymath.ax changes.
../../bin/axc_native.exe build imp_app.ax -o imp_app.exe --auto-lib -self-link -O1
cp ../../bin/ax_runtime.dll .
./imp_app.exe ; echo "exit=$?"     # expect exit=42  (14*3, from the precompiled lib)
../../bin/axc_native.exe iface imp_mymath.lib      # raw + parsed round-trip

# Or precompile the library by hand and let --auto-lib just reuse it:
../../bin/axc_native.exe build imp_mymath.ax -o imp_mymath.lib --staticlib --no-stdlib -self-link -O1
```

Libraries must be stdlib-free today (`--no-stdlib`, which `--auto-lib` uses); stdlib-using
libraries await the separate-compilation stdlib cache (RFC 0011 §6). Build from a directory
where both `std/*.ax` (stdlib concat) and the `.lib` resolve on the CWD-relative path the
module name maps to (`import a.b` → `a/b.lib`).

## Both `.lib` and `.dll` from one library (RFC 0011 P4 inc3e)

`--staticlib --shared` together emit BOTH artifacts for the same source, base-named:

```sh
../../bin/axc_native.exe build axmath.ax -o axmath.lib --staticlib --shared --no-stdlib -self-link -O1
# → axmath.lib (static archive, __axiom_iface: F ax_add / F ax_mul)
# → axmath.dll (PE export table: ax_add / ax_mul)
../../bin/axc_native.exe iface axmath.lib          # interface of the static lib
objdump -p axmath.dll | grep -A3 "Ordinal/Name"    # exported DLL names
```

Consumers then choose: static-link the `.lib` (`-l`, code embedded) or bind the DLL
dynamically (`extern "C" from "axmath.dll"`). The DLL exports follow `#[export]`; the
`.lib` exposes every defined symbol.

## `library`-marked DLL exports its full public API (RFC 0011 P4)

A module with the `library` marker at the top (e.g. `library libapi`) exports EVERY
`pub` function when built as a DLL — no `#[export]` needed — so the `.dll` surface
matches the `.lib` and the auto-generated FFI wrapper. Unmarked `--shared` keeps the
`#[export]`-only behavior (backward compatible).

```sh
# libapi.ax has `library libapi` + two plain `pub fn`s (no #[export])
../../bin/axc_native.exe build libapi.ax -o libapi.dll --shared --no-stdlib -self-link -O1
objdump -p libapi.dll | grep -A4 "Ordinal   Hint"   # BOTH lib_add and lib_sub exported
cat libapi_ffi.ax                                    # wrapper binds BOTH pub fns

# consumer imports the wrapper and calls the DLL by name
../../bin/axc_native.exe build libapi_use.ax -o libapi_use.exe -self-link -O1
cp ../../bin/ax_runtime.dll .
./libapi_use.exe ; echo "exit=$?"   # expect exit=34  ((40+2)-8), libapi.dll in .idata
```

## Auto-generated FFI wrapper for a DLL (RFC 0011 P4 inc3f)

Building a DLL also emits `<base>_ffi.ax` — `pub extern "C" from "<dll>"` bindings for every
`#[export]`ed function — so a consumer `import`s ONE file instead of hand-writing an extern
per function.

```sh
../../bin/axc_native.exe build axmath.ax -o axmath.dll --shared --no-stdlib -self-link -O1
cat axmath_ffi.ax
#   pub extern "C" from "axmath.dll" fn ax_add(p0: i32, p1: i32) -> i32
#   pub extern "C" from "axmath.dll" fn ax_mul(p0: i32, p1: i32) -> i32

# consumer: `import axmath_ffi` then call axmath_ffi.ax_add(...) — the linker adds
# axmath.dll to the import table automatically. (Needs axmath.dll + ax_runtime.dll present.)
```

## Public integer const through the interface (RFC 0011 P4 inc4a)

A `library`-marked module can export a `pub const` of integer type. Its value is written
into the `.lib` interface as `C <name> <type> <value>`; the consumer inlines it at each
use site with no access to the library source (no recompile).

```sh
# constlib.ax: `library constlib` + `pub const MAX_ITEMS: i32 = 42` + `pub fn base()`
../../bin/axc_native.exe build constuse.ax -o constuse.exe -self-link -O1   # auto-builds constlib.lib
../../bin/axc_native.exe iface constlib.lib | grep '^C '     # C MAX_ITEMS i32 42
cp ../../bin/ax_runtime.dll .
./constuse.exe ; echo "exit=$?"   # expect exit=34  (MAX_ITEMS - base() = 42 - 8)
```

Only INT_LIT initializers of a primitive-int type round-trip today (non-negative, <2^32);
float/bool/struct consts await a later increment.

## Struct through the interface (RFC 0011 P4 inc4b)

A `library`-marked module can export a `pub struct` (primitive fields today). The interface
carries `S <name> <nfields> <fname> <ftype> ...`; the consumer re-registers the struct and
recomputes its layout from the same fields, so size/align/offsets match exactly. Functions
that take or return the struct reference it by name (`F origin 0 -> Point`), enabling
struct-by-value ABI across separately-compiled units.

```sh
# geolib.ax: `library geolib` + `pub struct Point { x,y: i32 }` + origin()/add_pt()
../../bin/axc_native.exe build geouse.ax -o geouse.exe -self-link -O1   # auto-builds geolib.lib
../../bin/axc_native.exe iface geolib.lib | grep -E '^S |^F '   # S Point 2 x i32 y i32 ; F origin 0 -> Point
cp ../../bin/ax_runtime.dll .
./geouse.exe ; echo "exit=$?"   # expect exit=14  (add_pt(origin(), origin()) = (3+4)+(3+4))
```

Only structs with all-primitive fields round-trip today; pointer/nested-struct fields (which
need the pointee/struct in the token) are a later slice.

## Multi-DLL test (RFC 0009 P1, N=2)

```sh
gcc -shared -o libA.dll libA.c
gcc -shared -o libB.dll libB.c
../../bin/axc_native.exe build multi.ax -o multi.exe -self-link -O1
objdump -p multi.exe | grep "DLL Name"   # both libA.dll and libB.dll present
./multi.exe ; echo "exit=$?"             # expect exit=100
```
