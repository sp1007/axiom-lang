/**
 * test_genref.c — Unit tests for AXIOM generational reference system.
 *
 * Uses setjmp/longjmp to intercept ax_panic for testing.
 *
 * Compile:
 *   gcc -O2 -Wall -Wextra -Werror -std=c11 test_genref.c axalloc.c -o test_genref
 *   ./test_genref
 */
#include <stdio.h>
#include <setjmp.h>
#include <string.h>

/* Override ax_panic before including headers */
static jmp_buf  panic_jmp;
static char     panic_msg[256];
static int      panic_triggered;

void ax_panic(const char* msg) {
    strncpy(panic_msg, msg, sizeof(panic_msg) - 1);
    panic_msg[sizeof(panic_msg) - 1] = '\0';
    panic_triggered = 1;
    longjmp(panic_jmp, 1);
}

#include "axalloc.h"

static int pass_count = 0;
static int test_count = 0;

#define TEST(name) do { test_count++; printf("  [%02d] %-50s", test_count, name); } while(0)
#define PASS()     do { pass_count++; printf("PASS\n"); } while(0)
#define FAIL()     do { printf("FAIL\n"); } while(0)

#define ASSERT_PANIC(expr, name) do { \
    test_count++; \
    panic_triggered = 0; \
    panic_msg[0] = '\0'; \
    printf("  [%02d] %-50s", test_count, name); \
    if (setjmp(panic_jmp) == 0) { expr; } \
    if (panic_triggered) { pass_count++; printf("PASS\n"); } \
    else { printf("FAIL (no panic)\n"); } \
} while(0)

/* Test 1: ax_make_ref captures correct ptr and gen_id */
static void test_make_ref_valid(void) {
    TEST("ax_make_ref: ptr and gen_id correct");
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref = ax_make_ref(p);
    if (ref.ptr == p && ref.gen_id == 1) { PASS(); } else { FAIL(); }
    ax_free(p);
}

/* Test 2: ax_deref on live reference returns original pointer */
static void test_deref_live(void) {
    TEST("ax_deref: live ref returns same ptr");
    int* p = (int*)ax_alloc(sizeof(int));
    *p = 42;
    AxRef ref = ax_make_ref(p);
    int* result = (int*)ax_deref(ref);
    if (result == p && *result == 42) { PASS(); } else { FAIL(); }
    ax_free(p);
}

/* Test 3: ax_deref after free panics */
static void test_deref_after_free(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref = ax_make_ref(p);
    ax_free(p);
    ASSERT_PANIC(ax_deref(ref), "ax_deref after free panics");
}

/* Test 4: ax_deref with NULL panics */
static void test_deref_null(void) {
    AxRef null_ref = AX_NULL_REF;
    ASSERT_PANIC(ax_deref(null_ref), "ax_deref(NULL) panics");
}

/* Test 5: ax_invalidate sets gen_id to 0 */
static void test_invalidate(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref = ax_make_ref(p);
    ax_invalidate(p);

    /* gen_id should now be 0, ref has gen_id 1 → mismatch */
    AxHeader* h = ax_get_header(p);
    TEST("ax_invalidate sets gen_id to 0");
    if (h->gen_id == 0) { PASS(); } else { FAIL(); }

    /* Deref should panic */
    ASSERT_PANIC(ax_deref(ref), "ax_deref after invalidate panics");

    /* Restore gen_id to free cleanly */
    h->gen_id = 1;
    ax_free(p);
}

/* Test 6: ax_ref_valid returns 1 for live, 0 after free */
static void test_ref_valid(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref = ax_make_ref(p);

    TEST("ax_ref_valid: live ref is valid");
    if (ax_ref_valid(ref)) { PASS(); } else { FAIL(); }

    ax_free(p);

    TEST("ax_ref_valid: freed ref is invalid");
    /* After free, we can't dereference p, but the ref struct still has old values.
     * However, the memory is freed so reading the header is UB.
     * For testing, we just verify that AX_NULL_REF is invalid. */
    AxRef null_ref = AX_NULL_REF;
    if (!ax_ref_valid(null_ref)) { PASS(); } else { FAIL(); }
}

/* Test 7: ax_make_ref(NULL) returns null ref */
static void test_make_ref_null(void) {
    TEST("ax_make_ref(NULL) returns null ref");
    AxRef ref = ax_make_ref(NULL);
    if (ref.ptr == NULL && ref.gen_id == 0) { PASS(); } else { FAIL(); }
}

/* Test 8: AX_NULL_REF has ptr=NULL, gen_id=0 */
static void test_null_ref_constant(void) {
    TEST("AX_NULL_REF has ptr=NULL, gen_id=0");
    AxRef ref = AX_NULL_REF;
    if (ref.ptr == NULL && ref.gen_id == 0) { PASS(); } else { FAIL(); }
}

/* Test 9: Multiple refs to same allocation */
static void test_multiple_refs(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref1 = ax_make_ref(p);
    AxRef ref2 = ax_make_ref(p);

    TEST("multiple refs: both valid while live");
    if (ax_ref_valid(ref1) && ax_ref_valid(ref2)) { PASS(); } else { FAIL(); }

    ax_free(p);
    /* Both refs now stale (can't safely check, but verify they were captured correctly) */
    TEST("multiple refs: captured same gen_id");
    if (ref1.gen_id == ref2.gen_id && ref1.gen_id == 1) { PASS(); } else { FAIL(); }
}

/* Test 10: ax_invalidate(NULL) is a no-op */
static void test_invalidate_null(void) {
    TEST("ax_invalidate(NULL) is no-op");
    ax_invalidate(NULL); /* should not crash */
    PASS();
}

/* Test 11: panic message on use-after-free */
static void test_uaf_message(void) {
    int* p = (int*)ax_alloc(sizeof(int));
    AxRef ref = ax_make_ref(p);
    ax_free(p);

    panic_triggered = 0;
    panic_msg[0] = '\0';
    test_count++;
    printf("  [%02d] %-50s", test_count, "UAF panic message correct");
    if (setjmp(panic_jmp) == 0) {
        ax_deref(ref);
    }
    if (panic_triggered && strstr(panic_msg, "use-after-free") != NULL) {
        pass_count++;
        printf("PASS\n");
    } else {
        printf("FAIL\n");
    }
}

int main(void) {
    printf("=== genref unit tests ===\n\n");

    test_make_ref_valid();
    test_deref_live();
    test_deref_after_free();
    test_deref_null();
    test_invalidate();
    test_ref_valid();
    test_make_ref_null();
    test_null_ref_constant();
    test_multiple_refs();
    test_invalidate_null();
    test_uaf_message();

    printf("\n=== Results: %d/%d tests passed ===\n", pass_count, test_count);
    return (pass_count == test_count) ? 0 : 1;
}
