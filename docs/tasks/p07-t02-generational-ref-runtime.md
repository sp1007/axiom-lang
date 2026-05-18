# p07-t02: Generational Reference Runtime

## Purpose
Implement the generational reference checking system in the C runtime. This provides use-after-free detection by comparing the generation ID stored in an `AxRef` against the generation ID in the allocation's `AxHeader`. When they differ, the object has been freed since the reference was captured, and a panic is triggered immediately.

## Context
AXIOM's ownership model prevents use-after-free at compile time in safe code, but unsafe blocks and C-interop still benefit from runtime validation. The generational reference system is the last line of defense: every heap pointer dereference in generated C code goes through `ax_deref(ref)`, which validates the reference before allowing access. This is a lightweight alternative to reference counting or a tracing GC — the overhead is a single integer comparison per dereference.

The inline function in the header ensures the check is inlined at every call site, and the `__builtin_expect` hint tells the CPU branch predictor that mismatches are rare (the slow path is the error path).

## Inputs
- `runtime/axalloc/axalloc.h` and `axalloc.c` from p07-t01
- C11 compiler with `__builtin_expect` support (GCC or Clang)

## Outputs
- `runtime/axalloc/genref.h` — header with `ax_deref`, `ax_make_ref`, `ax_invalidate` inline functions and C unit test declarations
- `runtime/axalloc/test_genref.c` — C unit tests for the generational reference system

## Dependencies
- p07-t01 (axalloc MVP — provides `AxHeader`, `AxRef`, `ax_alloc`, `ax_free`, `ax_get_header`)

## Subsystems Affected
- Runtime (generational safety layer)
- C-Backend (p08-t06 emits `ax_deref` and `ax_make_ref` calls)
- Ownership system (every heap deref in generated code goes through this)
- Panic handler (p07-t03 is called on gen_id mismatch)

## Detailed Requirements

### `ax_deref` — inline hot-path validation
```c
static inline void* ax_deref(AxRef ref) {
    if (__builtin_expect(ref.ptr == NULL, 0))
        ax_panic("null pointer dereference");
    AxHeader* h = ((AxHeader*)ref.ptr) - 1;
    if (__builtin_expect(h->gen_id != ref.gen_id, 0))
        ax_panic("GenerationalID mismatch: use-after-free detected");
    return ref.ptr;
}
```

Key properties:
- Must be `static inline` to eliminate call overhead at every dereference site
- NULL check first (cheapest guard)
- Arithmetic `((AxHeader*)ref.ptr) - 1` gets to the header immediately before user data
- `__builtin_expect(..., 0)` marks both error branches as unlikely
- Returns the raw `void*` on success so the caller can cast to the desired type

### `ax_make_ref` — construct an AxRef from an ax_alloc pointer
```c
static inline AxRef ax_make_ref(void* ptr) {
    if (__builtin_expect(ptr == NULL, 0))
        ax_panic("ax_make_ref: cannot make ref from NULL");
    AxHeader* h = ((AxHeader*)ptr) - 1;
    AxRef ref;
    ref.ptr    = ptr;
    ref.gen_id = h->gen_id;
    return ref;
}
```

Called immediately after `ax_alloc` to capture the current generation ID. Every allocation in generated code follows the pattern:
```c
void* raw = ax_alloc(sizeof(SomeType));
AxRef ref = ax_make_ref(raw);
// from this point, use ax_deref(ref) to access the data
```

### `ax_invalidate` — explicit invalidation without freeing
```c
static inline void ax_invalidate(void* ptr) {
    if (!ptr) return;
    AxHeader* h = ((AxHeader*)ptr) - 1;
    h->gen_id = 0;  // 0 is the explicit invalidation sentinel
}
```

Used when ownership is transferred to another context and the local reference must be immediately invalid. Setting `gen_id = 0` means any existing `AxRef` with `gen_id >= 1` will fail the check in `ax_deref`.

### `ax_ref_valid` — non-panicking validity check
```c
static inline int ax_ref_valid(AxRef ref) {
    if (ref.ptr == NULL) return 0;
    AxHeader* h = ((AxHeader*)ref.ptr) - 1;
    return h->gen_id == ref.gen_id;
}
```

Used by optional checks (e.g., in assertions and debug tools) without triggering a panic.

### `AX_NULL_REF` — the null AxRef constant
```c
#define AX_NULL_REF ((AxRef){.ptr = NULL, .gen_id = 0})
```

### gen_id Semantics Table
| State | gen_id value | ax_deref result |
|-------|-------------|----------------|
| Live allocation | 1 | success |
| Freed once | 2 | mismatch → panic |
| Freed twice | 3 | mismatch → panic |
| Explicitly invalidated | 0 | mismatch → panic (any ref has gen_id >= 1) |

## Implementation Steps

### Step 1: Add `genref.h` to `runtime/axalloc/`
Create `runtime/axalloc/genref.h` containing:
- Include guard `#pragma once`
- `#include "axalloc.h"` (for `AxHeader`, `AxRef`)
- Forward declaration of `ax_panic` (defined in `runtime/panic/panic.c`)
- All four inline functions: `ax_deref`, `ax_make_ref`, `ax_invalidate`, `ax_ref_valid`
- The `AX_NULL_REF` macro

### Step 2: Update `axalloc.h` to include `genref.h`
At the bottom of `axalloc.h`, add:
```c
#include "genref.h"
```
This ensures any file that includes `axalloc.h` automatically gets the reference functions.

### Step 3: Write `test_genref.c`
```c
#include <stdio.h>
#include <setjmp.h>
#include "axalloc.h"

// --- Test harness ---
static int   test_count = 0;
static int   pass_count = 0;
static jmp_buf panic_jmp;
static int   panic_triggered = 0;

// Override ax_panic for testing: longjmp instead of abort
void ax_panic(const char* msg) {
    (void)msg;
    panic_triggered = 1;
    longjmp(panic_jmp, 1);
}

#define ASSERT(cond, name) do { \
    test_count++; \
    if (cond) { pass_count++; printf("[PASS] %s\n", name); } \
    else { printf("[FAIL] %s\n", name); } \
} while(0)

#define ASSERT_PANIC(expr, name) do { \
    test_count++; \
    panic_triggered = 0; \
    if (setjmp(panic_jmp) == 0) { expr; } \
    if (panic_triggered) { pass_count++; printf("[PASS] %s\n", name); } \
    else { printf("[FAIL] %s (no panic)\n", name); } \
} while(0)

// --- Tests ---
static void test_make_ref_valid(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref = ax_make_ref(p);
    ASSERT(ref.ptr == p, "make_ref: ptr matches");
    ASSERT(ref.gen_id == 1, "make_ref: gen_id is 1");
    ax_free(p);
}

static void test_deref_live(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    *p = 42;
    AxRef ref = ax_make_ref(p);
    int* result = (int*)ax_deref(ref);
    ASSERT(result == p, "deref live: returns same pointer");
    ASSERT(*result == 42, "deref live: can read value");
    ax_free(p);
}

static void test_deref_after_free_panics(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref = ax_make_ref(p);
    ax_free(p);
    ASSERT_PANIC(ax_deref(ref), "deref after free triggers panic");
}

static void test_null_deref_panics(void) {
    AxRef null_ref = AX_NULL_REF;
    ASSERT_PANIC(ax_deref(null_ref), "null deref triggers panic");
}

static void test_invalidate(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref = ax_make_ref(p);
    ax_invalidate(p);
    ASSERT_PANIC(ax_deref(ref), "deref after invalidate triggers panic");
    // Must manually free (gen_id != standard live value, but memory is not freed yet)
    ax_get_header(p)->gen_id = 1; // restore to free cleanly
    ax_free(p);
}

static void test_ref_valid(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref = ax_make_ref(p);
    ASSERT(ax_ref_valid(ref), "ref_valid: live ref is valid");
    ax_free(p);
    ASSERT(!ax_ref_valid(ref), "ref_valid: freed ref is invalid");
}

int main(void) {
    test_make_ref_valid();
    test_deref_live();
    test_deref_after_free_panics();
    test_null_deref_panics();
    test_invalidate();
    test_ref_valid();
    printf("\nResults: %d/%d passed\n", pass_count, test_count);
    return (pass_count == test_count) ? 0 : 1;
}
```

### Step 4: Update the `Makefile` in `runtime/axalloc/`
Add a `test_genref` target:
```makefile
test_genref: test_genref.c axalloc.o
	$(CC) $(CFLAGS) test_genref.c axalloc.o -o test_genref
	./test_genref

all: test_alloc test_genref
```

## Test Plan
1. `ax_make_ref` captures correct ptr and gen_id == 1
2. `ax_deref` on a live reference returns the original pointer
3. `ax_deref` after `ax_free` triggers panic (use `setjmp`/`longjmp` in test harness)
4. `ax_deref` with NULL ref triggers panic
5. `ax_invalidate` sets gen_id to 0; subsequent `ax_deref` panics
6. `ax_ref_valid` returns 1 for live ref, 0 after free
7. `ax_make_ref(NULL)` triggers panic
8. Multiple refs to same allocation: all valid while live, all invalid after free
9. `AX_NULL_REF` macro produces `{.ptr=NULL, .gen_id=0}`

## Validation Checklist
- [ ] `genref.h` compiles standalone with `-Wall -Wextra -Werror`
- [ ] All inline functions are `static inline` (no linker symbol conflicts)
- [ ] `ax_deref` uses `__builtin_expect` on both error branches
- [ ] `test_genref.c` compiles and all tests pass
- [ ] AddressSanitizer reports no errors on test suite
- [ ] `ax_make_ref(NULL)` panics (tested)
- [ ] `ax_invalidate(NULL)` is a no-op (tested)

## Acceptance Criteria
- `genref.h` provides all four functions as `static inline`
- All 9 tests in `test_genref.c` pass
- The `setjmp`/`longjmp` test harness correctly captures panics
- Compilation clean with GCC and Clang under `-Wall -Wextra -Werror`

## Definition of Done
- `runtime/axalloc/genref.h` exists with all inline functions
- `runtime/axalloc/test_genref.c` exists and passes
- `runtime/axalloc/Makefile` updated to include `test_genref` target
- `axalloc.h` updated to include `genref.h`

## Risks & Mitigations
- **Risk**: `ax_panic` is called from an inline function; if the panic implementation is in a separate TU, the linker must find it. **Mitigation**: Forward-declare `ax_panic` in `genref.h` with `extern void ax_panic(const char* msg);`. The test harness provides its own definition via override.
- **Risk**: Compiler may not inline the function if the TU is compiled with `-O0`. **Mitigation**: `static inline` always inlines with `__attribute__((always_inline))` as a fallback in performance-critical builds.
- **Risk**: gen_id comparison with 0 (invalidated state) needs careful documentation. **Mitigation**: Add a comment in the header explaining the gen_id value semantics table.

## Future Follow-up Tasks
- p07-t03: Panic handler implementation (used by both axalloc and genref)
- p07-t04: Unified runtime header includes genref.h
- p08-t06: C-Backend emits `ax_make_ref`/`ax_deref` at every heap allocation/access point
- p10-t05: CTGC pass decides when to emit `ax_invalidate` vs `ax_free`
