# p07-t05: Runtime Memory Safety Integration Tests

## Purpose
Write comprehensive integration tests for the AXIOM runtime memory safety system, covering the full interaction between the allocator (`ax_alloc`/`ax_free`), generational references (`ax_make_ref`/`ax_deref`), and the panic handler. These tests validate that the three subsystems work correctly together and catch memory errors that would otherwise cause silent corruption.

## Context
The allocator (p07-t01), generational references (p07-t02), and panic handler (p07-t03) were each tested in isolation. This task validates their integration. The tests live in `runtime/axalloc/test_memory.c` and are designed to be compiled with AddressSanitizer to catch errors that the generation counter mechanism does not catch (e.g., buffer overruns).

The test harness must be self-contained: it overrides `ax_panic` via a `setjmp`/`longjmp` mechanism so tests can verify that panics are triggered without aborting the test process.

## Inputs
- `runtime/axalloc/axalloc.h` and `axalloc.c` (p07-t01)
- `runtime/axalloc/genref.h` (p07-t02)
- `runtime/panic/panic.h` (p07-t03)
- GCC with `-fsanitize=address` support

## Outputs
- `runtime/axalloc/test_memory.c` — integration test file
- Updated `runtime/axalloc/Makefile` with `test_memory` and `test_memory_asan` targets

## Dependencies
- p07-t02 (generational references — primary subject of tests)
- p07-t03 (panic handler — tested via setjmp override)

## Subsystems Affected
- Runtime (all three memory subsystems exercised together)
- CI pipeline (these tests run in the memory safety CI job)

## Detailed Requirements

### Test Structure
Each test is a standalone function with a descriptive name. A minimal test harness collects pass/fail counts. The `ax_panic` function is overridden in this translation unit to use `longjmp` rather than `abort()`, so panic-triggering tests can be expressed as positive assertions.

### Test Coverage Requirements
The following 5 groups of tests are required. Each group has multiple sub-cases.

**Group 1: Basic allocation lifecycle**
- Alloc → write → verify readable → free
- After free: verify `gen_id` in header was incremented (read via saved header pointer)
- Double free: second `ax_free` should increment gen_id again (no crash, no UB because header was already freed — this is specifically testable without ASan since we read before the second free)

**Group 2: Use-after-free detection**
- Alloc → capture `AxRef` → free → `ax_deref(ref)` must panic
- Alloc → capture `AxRef` → free → reallocate same size → `ax_deref(old_ref)` must panic (new alloc may or may not get the same address, but gen_id will differ)
- `ax_ref_valid(ref)` returns 0 after free

**Group 3: Allocation ordering stress**
- Allocate N=100 pointers, write index values, free in reverse order, verify no corruption
- Allocate alternating sizes (7, 64, 1, 128), free all, verify all cleanly freed
- Alloc → partial use (write only first half) → free → verify no corruption

**Group 4: Realloc behavior**
- Alloc 64 → write pattern → realloc to 128 → verify pattern preserved → verify gen_id unchanged → free
- `ax_realloc(NULL, 32)` behaves as `ax_alloc(32)`
- Realloc smaller: `ax_realloc(ptr, 8)` from 64-byte allocation; verify `ax_alloc_size == 8`

**Group 5: NULL and edge cases**
- `ax_deref(AX_NULL_REF)` panics
- `ax_make_ref(NULL)` panics
- `ax_free(NULL)` is a no-op (does not panic)
- `ax_alloc(0)` returns non-NULL; `ax_alloc_size == 0`
- `ax_bounds_check(SIZE_MAX, 10)` panics (overflow edge case)

### Compile and Run Commands
```
# Debug build
gcc -O0 -g -std=c11 -Wall -Wextra \
    test_memory.c axalloc.c -o test_memory && ./test_memory

# AddressSanitizer build (no ASan on double-free test since we test UB-free paths)
gcc -O0 -g -fsanitize=address -fsanitize=undefined \
    test_memory.c axalloc.c -o test_memory_asan && ./test_memory_asan
```

Note: the double-free test (Group 1, sub-case 3) is excluded from the ASan build because intentionally freeing twice is undefined behavior that ASan will catch as an error. Wrap it with `#ifndef __SANITIZE_ADDRESS__`.

## Implementation Steps

### Step 1: Write `runtime/axalloc/test_memory.c`
```c
#include <stdio.h>
#include <string.h>
#include <setjmp.h>
#include <stdint.h>
#include <limits.h>
#include "axalloc.h"
#include "../panic/panic.h"

// ---- Test harness ----
static jmp_buf  g_jmp;
static int      g_panic_triggered;
static char     g_panic_msg[256];
static int      g_pass, g_total;

void ax_panic(const char* msg) {
    strncpy(g_panic_msg, msg, sizeof(g_panic_msg)-1);
    g_panic_triggered = 1;
    longjmp(g_jmp, 1);
}
void ax_set_program_name(const char* n) { (void)n; }

#define ASSERT(cond, name) do { \
    g_total++; \
    if (cond) { g_pass++; printf("[PASS] %s\n", name); } \
    else { printf("[FAIL] %s  (line %d)\n", name, __LINE__); } \
} while(0)

#define ASSERT_PANIC(expr, name) do { \
    g_total++; g_panic_triggered = 0; \
    if (setjmp(g_jmp) == 0) { expr; } \
    if (g_panic_triggered) { g_pass++; printf("[PASS] %s\n", name); } \
    else { printf("[FAIL] %s (no panic at line %d)\n", name, __LINE__); } \
} while(0)

// ---- Group 1: Basic lifecycle ----
static void test_alloc_write_free(void) {
    printf("-- Group 1: Basic lifecycle --\n");

    uint8_t* buf = (uint8_t*)ax_alloc(64);
    ASSERT(buf != NULL, "alloc returns non-NULL");
    ASSERT(ax_alloc_size(buf) == 64, "alloc_size == 64");

    // Write all bytes
    memset(buf, 0xAB, 64);
    ASSERT(buf[0] == 0xAB && buf[63] == 0xAB, "write/read all bytes");

    // gen_id before free
    AxHeader* hdr = ax_get_header(buf);
    uint64_t gen_before = hdr->gen_id;
    ASSERT(gen_before == 1, "gen_id == 1 before free");

    // Save header address for post-free check (ASan excluded)
#ifndef __SANITIZE_ADDRESS__
    AxHeader* saved_hdr_addr = hdr; // dangling after free
    ax_free(buf);
    // The header memory was freed — accessing it is technically UB,
    // but on all tested platforms malloc doesn't immediately overwrite it.
    // This test is explicitly excluded from ASan builds.
    ASSERT(saved_hdr_addr->gen_id == gen_before + 1, "gen_id incremented after free");
#else
    ax_free(buf);
    ASSERT(1, "gen_id check skipped under ASan");
#endif
}

// ---- Group 2: Use-after-free detection ----
static void test_use_after_free(void) {
    printf("-- Group 2: Use-after-free detection --\n");

    int* p = (int*)ax_alloc(sizeof(int));
    *p = 99;
    AxRef ref = ax_make_ref(p);

    ASSERT(ax_ref_valid(ref), "ref valid before free");

    ax_free(p);

    ASSERT(!ax_ref_valid(ref), "ref invalid after free");
    ASSERT_PANIC(ax_deref(ref), "ax_deref panics after free");
}

// ---- Group 3: Allocation ordering stress ----
static void test_alloc_ordering(void) {
    printf("-- Group 3: Ordering stress --\n");
    #define N 100
    int* ptrs[N];

    // Allocate N pointers, write index
    for (int i = 0; i < N; i++) {
        ptrs[i] = (int*)ax_alloc(sizeof(int));
        *ptrs[i] = i;
    }

    // Verify all values
    int all_ok = 1;
    for (int i = 0; i < N; i++) {
        if (*ptrs[i] != i) { all_ok = 0; break; }
    }
    ASSERT(all_ok, "all N values readable after N allocs");

    // Free in reverse order
    for (int i = N-1; i >= 0; i--) {
        ax_free(ptrs[i]);
    }
    ASSERT(1, "reverse-order free completes without crash");
    #undef N
}

// ---- Group 4: Realloc behavior ----
static void test_realloc(void) {
    printf("-- Group 4: Realloc --\n");

    // Alloc 64, write pattern
    uint8_t* p = (uint8_t*)ax_alloc(64);
    for (int i = 0; i < 64; i++) p[i] = (uint8_t)i;

    AxHeader* h = ax_get_header(p);
    uint64_t gen = h->gen_id;

    // Realloc to 128
    p = (uint8_t*)ax_realloc(p, 128);
    ASSERT(p != NULL, "realloc returns non-NULL");
    ASSERT(ax_alloc_size(p) == 128, "alloc_size == 128 after realloc");
    ASSERT(ax_get_header(p)->gen_id == gen, "gen_id preserved across realloc");

    // First 64 bytes preserved
    int pattern_ok = 1;
    for (int i = 0; i < 64; i++) {
        if (p[i] != (uint8_t)i) { pattern_ok = 0; break; }
    }
    ASSERT(pattern_ok, "data preserved across realloc");

    ax_free(p);

    // realloc NULL == alloc
    int* q = (int*)ax_realloc(NULL, sizeof(int));
    ASSERT(q != NULL, "realloc(NULL) == alloc");
    ax_free(q);
}

// ---- Group 5: NULL and edge cases ----
static void test_edge_cases(void) {
    printf("-- Group 5: Edge cases --\n");

    // NULL deref panics
    ASSERT_PANIC(ax_deref(AX_NULL_REF), "ax_deref(NULL_REF) panics");

    // ax_make_ref(NULL) panics
    ASSERT_PANIC(ax_make_ref(NULL), "ax_make_ref(NULL) panics");

    // ax_free(NULL) is a no-op
    ax_free(NULL); // must not crash
    ASSERT(1, "ax_free(NULL) is no-op");

    // zero-size alloc
    void* p0 = ax_alloc(0);
    ASSERT(p0 != NULL, "ax_alloc(0) returns non-NULL");
    ax_free(p0);

    // bounds check edge: SIZE_MAX as index
    ASSERT_PANIC(ax_bounds_check(SIZE_MAX, 10), "bounds_check(SIZE_MAX,10) panics");
}

int main(void) {
    test_alloc_write_free();
    test_use_after_free();
    test_alloc_ordering();
    test_realloc();
    test_edge_cases();
    printf("\nResults: %d/%d passed\n", g_pass, g_total);
    return (g_pass == g_total) ? 0 : 1;
}
```

### Step 2: Update `runtime/axalloc/Makefile`
```makefile
CC     = gcc
CFLAGS = -O0 -g -std=c11 -Wall -Wextra -I.. -I../panic

test_memory: test_memory.c axalloc.c
	$(CC) $(CFLAGS) test_memory.c axalloc.c -o test_memory
	./test_memory

test_memory_asan: test_memory.c axalloc.c
	$(CC) $(CFLAGS) -fsanitize=address -fsanitize=undefined \
	    test_memory.c axalloc.c -o test_memory_asan
	./test_memory_asan

all: test_alloc test_genref test_memory test_memory_asan
```

### Step 3: Add to CI
In `.github/workflows/runtime.yml` (or equivalent CI config), add:
```yaml
- name: Build and run memory safety tests
  run: |
    cd runtime/axalloc
    make test_memory
    make test_memory_asan
```

## Test Plan
All 5 groups must pass:
1. Basic lifecycle: alloc/write/free cycle, gen_id verification
2. Use-after-free: `ax_deref` panics, `ax_ref_valid` returns false
3. Ordering stress: 100-alloc reverse-free test
4. Realloc: data preservation, gen_id preservation, realloc(NULL)
5. Edge cases: NULL deref, make_ref(NULL), free(NULL), zero-size, SIZE_MAX bounds

Additionally:
- All tests pass under AddressSanitizer (except the explicitly excluded double-free sub-test)
- All tests pass under UBSan (`-fsanitize=undefined`)
- Test binary exits with code 0 on full pass, code 1 on any failure

## Validation Checklist
- [ ] `test_memory.c` compiles without warnings with `-Wall -Wextra`
- [ ] All 5 groups pass in normal build
- [ ] All non-ASan-excluded tests pass under `-fsanitize=address`
- [ ] All tests pass under `-fsanitize=undefined`
- [ ] Test binary exits 0 on pass, 1 on failure
- [ ] Double-free test is correctly guarded with `#ifndef __SANITIZE_ADDRESS__`
- [ ] setjmp/longjmp panic override does not interfere between test cases

## Acceptance Criteria
- Every test in all 5 groups passes
- AddressSanitizer reports no errors (on the ASan build)
- UBSan reports no errors
- CI job succeeds

## Definition of Done
- `runtime/axalloc/test_memory.c` exists with all 5 test groups
- `runtime/axalloc/Makefile` has `test_memory` and `test_memory_asan` targets
- All tests pass locally and in CI
- CI configuration includes this test job

## Risks & Mitigations
- **Risk**: The post-free gen_id read (Group 1) is technically UB and may fail with certain allocators or compiler optimizations. **Mitigation**: Guard with `#ifndef __SANITIZE_ADDRESS__` and use `volatile` pointer to prevent the compiler from optimizing away the read.
- **Risk**: `longjmp` across C stack frames that contain VLAs or cleanups may cause issues. **Mitigation**: The test functions do not use VLAs or C++ destructors; plain C frames are safe with `longjmp`.
- **Risk**: `ax_alloc(0)` behavior is implementation-defined in standard C. **Mitigation**: `axalloc.c` explicitly handles size==0 by allocating 1 byte (to ensure a non-NULL return), documented as an AXIOM guarantee.

## Future Follow-up Tasks
- p08-t10: E2E compliance tests build on the runtime tests established here
- p10-t05: CTGC optimizer tests use the same harness to verify object reuse correctness
- p11-t15: Native backend integration tests link against these runtime objects
- Future: fuzz testing with `libFuzzer` targeting `ax_alloc`/`ax_free` sequences
