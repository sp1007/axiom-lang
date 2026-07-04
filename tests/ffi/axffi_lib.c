/* tests/ffi/axffi_lib.c — third-party C DLL for RFC 0009 P1 FFI import test.
 * Build (MSYS2 ucrt64 gcc): gcc -shared -o axffi_lib.dll axffi_lib.c
 * Exports two plain C symbols consumed by t_ffi.ax via `extern "C" from`. */
#include <stdint.h>
__declspec(dllexport) int32_t ffi_add(int32_t a, int32_t b) { return a + b; }
__declspec(dllexport) int32_t ffi_mul(int32_t a, int32_t b) { return a * b; }
