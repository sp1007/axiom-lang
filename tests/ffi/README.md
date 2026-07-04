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
