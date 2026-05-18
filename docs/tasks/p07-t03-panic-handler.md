# p07-t03: Panic Handler

## Purpose
Implement the AXIOM runtime panic handler in `runtime/panic/panic.c` and `runtime/panic/panic.h`. The panic handler is the terminal error reporting mechanism: it prints a diagnostic message to stderr (with a stack trace where available), then calls `abort()`. It is called by the allocator, the generational reference checker, bounds checking, and user-visible `assert` calls.

## Context
In AXIOM's safety model, a panic represents an unrecoverable error — something that should have been prevented at compile time by the ownership checker, but was caught at runtime as a last resort. Panics must be loud, informative, and fast to terminate. They must never be silently swallowed.

On Linux, the `backtrace()` POSIX API provides a stack trace. On Windows, `CaptureStackBackTrace` from `DbgHelp.dll` provides equivalent functionality. On macOS, `backtrace()` is also available. The implementation must compile on all three platforms.

## Inputs
- C11 compiler with platform-specific extensions (GCC/Clang on Linux/macOS, MSVC or Clang-CL on Windows)
- `runtime/axalloc/axalloc.h` (for the `AxHeader` type; used in allocator-originated panics)
- System headers: `<execinfo.h>` on POSIX, `<windows.h>` + `<DbgHelp.h>` on Windows

## Outputs
- `runtime/panic/panic.h` — public header declaring the panic API
- `runtime/panic/panic.c` — platform-specific implementation
- `runtime/panic/test_panic.c` — unit tests using setjmp/longjmp override

## Dependencies
- p07-t01 (axalloc MVP; panic.c must compile alongside axalloc.c without circular deps)

## Subsystems Affected
- All runtime subsystems (panic is the universal error exit)
- C-Backend: generates `ax_bounds_check` and `ax_assert` calls in bounds-checked array accesses
- Allocator (p07-t01): calls `ax_panic` on OOM
- Generational ref (p07-t02): calls `ax_panic` on gen_id mismatch

## Detailed Requirements

### Public API
```c
// Print msg to stderr with stack trace, then abort(). Never returns.
__attribute__((noreturn)) void ax_panic(const char* msg);

// Panic if idx >= len. Used for array bounds checks in generated code.
static inline void ax_bounds_check(size_t idx, size_t len) {
    if (__builtin_expect(idx >= len, 0)) {
        char buf[128];
        snprintf(buf, sizeof(buf),
                 "index out of bounds: index %zu, length %zu", idx, len);
        ax_panic(buf);
    }
}

// Panic if cond is false. Used for runtime assertions.
static inline void ax_assert(int cond, const char* msg) {
    if (__builtin_expect(!cond, 0))
        ax_panic(msg);
}
```

### Panic Output Format
```
AXIOM PANIC: <msg>
Stack trace:
  #0  ax_panic (panic.c:42)
  #1  ax_deref (genref.h:18)
  #2  ax_module_main (main.c:7)
  #3  main (main.c:1)
Aborted (core dumped)
```

- Always write to `stderr` (fd 2), not stdout
- Include the program name if available via `argv[0]` stored in a global
- Print stack trace depth up to 32 frames
- Symbol names: use `backtrace_symbols` on POSIX (requires linking with `-rdynamic`)

### Platform Conditionals
```c
#ifdef __linux__
  // Use execinfo.h: backtrace() + backtrace_symbols()
#elif defined(__APPLE__)
  // Use execinfo.h (same API as Linux)
#elif defined(_WIN32)
  // Use DbgHelp: CaptureStackBackTrace + SymFromAddr
#else
  // Fallback: print message only, no stack trace
#endif
```

### Program Name Registration
```c
// Call once from ax_main (the entry point wrapper) to store argv[0]
void ax_set_program_name(const char* name);
```

### `__attribute__((noreturn))` / `__declspec(noreturn)`
On MSVC: use `__declspec(noreturn)`. On GCC/Clang: use `__attribute__((noreturn))`. Provide a portability macro:
```c
#if defined(_MSC_VER)
  #define AX_NORETURN __declspec(noreturn)
#else
  #define AX_NORETURN __attribute__((noreturn))
#endif
```

## Implementation Steps

### Step 1: Create `runtime/panic/` directory
```
runtime/
  panic/
    panic.h
    panic.c
    test_panic.c
    Makefile
```

### Step 2: Write `panic.h`
```c
#pragma once
#include <stddef.h>

// Portability macro for no-return functions
#if defined(_MSC_VER)
  #define AX_NORETURN __declspec(noreturn)
#else
  #define AX_NORETURN __attribute__((noreturn))
#endif

// Register the program name (call from ax_main before anything else)
void ax_set_program_name(const char* name);

// Core panic: print message + stack trace, then abort. Never returns.
AX_NORETURN void ax_panic(const char* msg);

// Inline bounds check (emitted by C-Backend at every array index)
#include <stdio.h>
static inline void ax_bounds_check(size_t idx, size_t len) {
    if (__builtin_expect(idx >= len, 0)) {
        char buf[128];
        snprintf(buf, sizeof(buf),
                 "index out of bounds: index %zu, length %zu", idx, len);
        ax_panic(buf);
    }
}

// Inline assertion (emitted for debug builds)
static inline void ax_assert(int cond, const char* msg) {
    if (__builtin_expect(!cond, 0))
        ax_panic(msg);
}
```

### Step 3: Write `panic.c` (Linux/macOS path shown; Windows guarded)
```c
#include "panic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const char* program_name = "<unknown>";

void ax_set_program_name(const char* name) {
    program_name = name;
}

#if defined(__linux__) || defined(__APPLE__)
#include <execinfo.h>

AX_NORETURN void ax_panic(const char* msg) {
    fprintf(stderr, "\nAXIOM PANIC in '%s': %s\n", program_name, msg);
    fprintf(stderr, "Stack trace:\n");

    void* frames[32];
    int   count = backtrace(frames, 32);
    char** syms = backtrace_symbols(frames, count);

    for (int i = 0; i < count; i++) {
        fprintf(stderr, "  #%d  %s\n", i, syms ? syms[i] : "??");
    }
    free(syms);
    fflush(stderr);
    abort();
}

#elif defined(_WIN32)
#include <windows.h>
#include <DbgHelp.h>
#pragma comment(lib, "DbgHelp.lib")

AX_NORETURN void ax_panic(const char* msg) {
    fprintf(stderr, "\nAXIOM PANIC in '%s': %s\n", program_name, msg);
    fprintf(stderr, "Stack trace:\n");

    void* frames[32];
    USHORT count = CaptureStackBackTrace(0, 32, frames, NULL);

    HANDLE process = GetCurrentProcess();
    SymInitialize(process, NULL, TRUE);

    char sym_buf[sizeof(SYMBOL_INFO) + MAX_SYM_NAME];
    SYMBOL_INFO* sym = (SYMBOL_INFO*)sym_buf;
    sym->SizeOfStruct = sizeof(SYMBOL_INFO);
    sym->MaxNameLen   = MAX_SYM_NAME;

    for (USHORT i = 0; i < count; i++) {
        DWORD64 disp = 0;
        if (SymFromAddr(process, (DWORD64)frames[i], &disp, sym)) {
            fprintf(stderr, "  #%u  %s + 0x%llx\n", i, sym->Name, (unsigned long long)disp);
        } else {
            fprintf(stderr, "  #%u  0x%p\n", i, frames[i]);
        }
    }
    fflush(stderr);
    abort();
}

#else
// Fallback: no stack trace
AX_NORETURN void ax_panic(const char* msg) {
    fprintf(stderr, "\nAXIOM PANIC: %s\n", msg);
    fflush(stderr);
    abort();
}
#endif
```

### Step 4: Write `test_panic.c`
Use function pointer override to intercept `ax_panic` in tests:
```c
#include <stdio.h>
#include <setjmp.h>
#include <string.h>
#include "panic.h"

// For testing: replace ax_panic with a longjmp version
static jmp_buf  test_jmp;
static char     last_panic_msg[256];
static int      panic_triggered;

// Weak override: in the test binary, link this definition after panic.o
// Alternatively, use a function pointer indirection in panic.h (test mode)
// Here we use a simple compile-time swap: test_panic.c defines its own ax_panic
void ax_panic(const char* msg) {
    strncpy(last_panic_msg, msg, sizeof(last_panic_msg)-1);
    panic_triggered = 1;
    longjmp(test_jmp, 1);
}

// --- helpers ---
static int pass_count = 0, test_count = 0;
#define ASSERT(c, name) do { test_count++; if(c){pass_count++;printf("[PASS] %s\n",name);} else printf("[FAIL] %s\n",name); } while(0)
#define ASSERT_PANIC(expr, name) do { \
    test_count++; panic_triggered = 0; \
    if (setjmp(test_jmp)==0) { expr; } \
    if(panic_triggered){pass_count++;printf("[PASS] %s\n",name);} else printf("[FAIL] %s (no panic)\n",name); \
} while(0)

int main(void) {
    // bounds check: valid
    ax_bounds_check(0, 10); // should not panic
    ASSERT(1, "bounds_check valid index");

    // bounds check: invalid
    ASSERT_PANIC(ax_bounds_check(10, 10), "bounds_check idx==len panics");
    ASSERT_PANIC(ax_bounds_check(100, 10), "bounds_check idx>len panics");

    // assert: cond true
    ax_assert(1, "this should not panic");
    ASSERT(1, "assert true: no panic");

    // assert: cond false
    ASSERT_PANIC(ax_assert(0, "test assertion failed"), "assert false panics");
    ASSERT(strstr(last_panic_msg, "test assertion failed") != NULL,
           "assert passes message to panic");

    printf("\nResults: %d/%d passed\n", pass_count, test_count);
    return (pass_count == test_count) ? 0 : 1;
}
```

### Step 5: Write `Makefile`
```makefile
CC     = gcc
CFLAGS = -O2 -Wall -Wextra -Werror -std=c11 -rdynamic

all: test_panic

panic.o: panic.c panic.h
	$(CC) $(CFLAGS) -c panic.c -o panic.o

# test_panic defines its own ax_panic (override), so don't link panic.o
test_panic: test_panic.c panic.h
	$(CC) $(CFLAGS) test_panic.c -o test_panic
	./test_panic

clean:
	rm -f panic.o test_panic
```

## Test Plan
1. `ax_bounds_check(0, 10)` does not panic
2. `ax_bounds_check(10, 10)` panics (idx == len is out-of-bounds)
3. `ax_bounds_check(100, 10)` panics
4. `ax_assert(1, "msg")` does not panic
5. `ax_assert(0, "msg")` panics and the message is propagated
6. `ax_panic` writes to stderr (use `dup2` to capture stderr in test)
7. On Linux: running a real panic binary produces a non-zero exit code (from `abort()`)
8. `ax_set_program_name` is reflected in the panic output

## Validation Checklist
- [ ] `panic.h` compiles standalone with `-Wall -Wextra -Werror`
- [ ] `panic.c` compiles on Linux, Windows, and macOS (CI matrix)
- [ ] `AX_NORETURN` macro defined correctly for GCC/Clang/MSVC
- [ ] All tests in `test_panic.c` pass
- [ ] `ax_panic` calls `abort()` — process exits with non-zero status
- [ ] Stack trace printed on POSIX with `-rdynamic`
- [ ] `ax_bounds_check` and `ax_assert` are `static inline` (no linker symbol)

## Acceptance Criteria
- `ax_panic` never returns (verified by `AX_NORETURN` and compiler warning if path exists)
- Stack trace is printed to stderr on Linux and macOS
- All unit tests pass
- No warnings under `-Wall -Wextra -Werror`

## Definition of Done
- `runtime/panic/panic.h` exists
- `runtime/panic/panic.c` exists with Linux/macOS/Windows platform branches
- `runtime/panic/test_panic.c` exists and all tests pass
- `runtime/panic/Makefile` exists and `make` succeeds

## Risks & Mitigations
- **Risk**: `backtrace_symbols` returns mangled C++ names or unhelpful addresses without debug symbols. **Mitigation**: Require `-rdynamic` link flag and document in `Makefile`. Generated binaries from `axc build` will link with `-rdynamic` by default in debug mode.
- **Risk**: On Windows, `DbgHelp.dll` may not be present on minimal systems. **Mitigation**: `LoadLibrary` `DbgHelp.dll` at runtime with a fallback to address-only printing.
- **Risk**: `abort()` generates SIGABRT which might be caught by a signal handler. **Mitigation**: Before `abort()`, call `signal(SIGABRT, SIG_DFL)` to reset the handler.

## Future Follow-up Tasks
- p07-t04: `ax_runtime.h` includes `panic.h`
- p09-t01: AIR `OpPanic` instruction lowers to a call to `ax_panic`
- p11-t12: ELF emitter links `panic.o` into every generated binary
- Future: structured panic payloads with source location for better diagnostics
