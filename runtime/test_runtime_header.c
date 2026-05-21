/**
 * test_runtime_header.c — Smoke test for ax_runtime.h
 *
 * Verifies that including ax_runtime.h in isolation compiles cleanly,
 * all type sizes are correct, and basic runtime operations work.
 *
 * Compile:
 *   gcc -O2 -Wall -Wextra -Werror -std=c11 -I. test_runtime_header.c axalloc/axalloc.c -o test_runtime
 *   ./test_runtime
 */

/* Override ax_panic for standalone test (must be before includes) */
#include <setjmp.h>
static jmp_buf test_panic_jmp;
static int test_panic_called = 0;
void ax_panic(const char* msg) {
    (void)msg;
    test_panic_called = 1;
    longjmp(test_panic_jmp, 1);
}
void ax_set_program_name(const char* name) { (void)name; }

#include "ax_runtime.h"

/* Static assertions for type sizes */
AX_STATIC_ASSERT(sizeof(ax_i8)  == 1,  "ax_i8 is 1 byte");
AX_STATIC_ASSERT(sizeof(ax_i16) == 2,  "ax_i16 is 2 bytes");
AX_STATIC_ASSERT(sizeof(ax_i32) == 4,  "ax_i32 is 4 bytes");
AX_STATIC_ASSERT(sizeof(ax_i64) == 8,  "ax_i64 is 8 bytes");
AX_STATIC_ASSERT(sizeof(ax_u8)  == 1,  "ax_u8 is 1 byte");
AX_STATIC_ASSERT(sizeof(ax_u16) == 2,  "ax_u16 is 2 bytes");
AX_STATIC_ASSERT(sizeof(ax_u32) == 4,  "ax_u32 is 4 bytes");
AX_STATIC_ASSERT(sizeof(ax_u64) == 8,  "ax_u64 is 8 bytes");
AX_STATIC_ASSERT(sizeof(ax_f32) == 4,  "ax_f32 is 4 bytes");
AX_STATIC_ASSERT(sizeof(ax_f64) == 8,  "ax_f64 is 8 bytes");
AX_STATIC_ASSERT(sizeof(ax_bool) == 1, "ax_bool is 1 byte");
AX_STATIC_ASSERT(sizeof(ax_byte) == 1, "ax_byte is 1 byte");
AX_STATIC_ASSERT(sizeof(ax_char) == 4, "ax_char is 4 bytes");

static int pass = 0, total = 0;
#define CHECK(cond, name) do { \
    total++; \
    if (cond) { pass++; printf("  [PASS] %s\n", name); } \
    else { printf("  [FAIL] %s\n", name); } \
} while(0)

int main(void) {
    printf("=== ax_runtime.h smoke tests ===\n\n");

    /* Type exercises */
    ax_i32 x = 42;
    CHECK(x == 42, "ax_i32 value");

    ax_f64 pi = 3.14159;
    CHECK(pi > 3.0, "ax_f64 value");

    ax_bool b = AX_TRUE;
    CHECK(b == 1, "AX_TRUE == 1");
    CHECK(AX_FALSE == 0, "AX_FALSE == 0");

    /* String */
    ax_string s = AX_STR("hello");
    CHECK(s.len == 5, "AX_STR length");
    CHECK(s.ptr != NULL, "AX_STR ptr not null");
    CHECK(memcmp(s.ptr, "hello", 5) == 0, "AX_STR content");

    /* Allocator round trip */
    void* p = ax_alloc(64);
    CHECK(p != NULL, "ax_alloc returns non-NULL");
    CHECK(ax_alloc_size(p) == 64, "ax_alloc_size correct");

    AxRef ref = ax_make_ref(p);
    CHECK(ref.ptr == p, "ax_make_ref ptr");
    CHECK(ref.gen_id == 1, "ax_make_ref gen_id");

    void* p2 = ax_deref(ref);
    CHECK(p2 == p, "ax_deref returns same ptr");

    ax_free(p);

    /* Bounds check (valid) */
    ax_bounds_check(0, 10);
    CHECK(1, "ax_bounds_check(0, 10) no panic");

    /* Assert (true) */
    ax_assert(1, "should not panic");
    CHECK(1, "ax_assert(true) no panic");

    /* Slice type smoke */
    ax_slice_i32 si;
    si.ptr = NULL;
    si.len = 0;
    si.cap = 0;
    CHECK(sizeof(si) > 0, "ax_slice_i32 exists");

    /* Utility macros */
    CHECK(AX_MIN(3, 5) == 3, "AX_MIN(3, 5) == 3");
    CHECK(AX_MAX(3, 5) == 5, "AX_MAX(3, 5) == 5");

    int arr[] = {1, 2, 3, 4, 5};
    CHECK(AX_ARRAY_LEN(arr) == 5, "AX_ARRAY_LEN works");

    AX_UNUSED(si);
    CHECK(1, "AX_UNUSED compiles");

    printf("\n=== Results: %d/%d tests passed ===\n", pass, total);
    return (pass == total) ? 0 : 1;
}
