/**
 * test_ax_stdlib.c — Tests for AXIOM C Runtime Standard Library.
 *
 * Compile: gcc -o test_stdlib test_ax_stdlib.c ax_string_ops.c ax_print.c
 *          ax_assert.c ax_math.c ax_collections.c panic/panic.c
 *          -I. -lm
 */

#include "ax_stdlib.h"
#include <stdio.h>
#include <string.h>

static int tests_passed = 0;
static int tests_failed = 0;

#define TEST(name) do { printf("  %-40s", name); } while(0)
#define PASS() do { printf("PASS\n"); tests_passed++; } while(0)
#define FAIL(msg) do { printf("FAIL: %s\n", msg); tests_failed++; } while(0)
#define CHECK(cond, msg) do { if (!(cond)) { FAIL(msg); return; } } while(0)

/* ================================================================
 * String Tests
 * ================================================================ */

static void test_str_len(void) {
    TEST("str_len");
    ax_string s = AX_STR("hello");
    CHECK(ax_str_len(s) == 5, "expected 5");
    CHECK(ax_str_len(AX_STR("")) == 0, "expected 0 for empty");
    CHECK(ax_str_len(AX_STR("a")) == 1, "expected 1");
    PASS();
}

static void test_str_char_count(void) {
    TEST("str_char_count (ASCII)");
    CHECK(ax_str_char_count(AX_STR("hello")) == 5, "expected 5 ASCII chars");
    CHECK(ax_str_char_count(AX_STR("")) == 0, "expected 0");
    PASS();
}

static void test_str_concat(void) {
    TEST("str_concat");
    ax_string a = AX_STR("hello ");
    ax_string b = AX_STR("world");
    ax_string result = ax_str_concat(a, b);
    CHECK(result.len == 11, "expected len 11");
    CHECK(memcmp(result.ptr, "hello world", 11) == 0, "content mismatch");
    free((void*)result.ptr);
    PASS();
}

static void test_str_slice(void) {
    TEST("str_slice");
    ax_string s = AX_STR("hello world");
    ax_string sub = ax_str_slice(s, 6, 11);
    CHECK(sub.len == 5, "expected len 5");
    CHECK(memcmp(sub.ptr, "world", 5) == 0, "content mismatch");

    ax_string empty = ax_str_slice(s, 5, 5);
    CHECK(empty.len == 0, "expected empty slice");
    PASS();
}

static void test_str_contains(void) {
    TEST("str_contains");
    ax_string s = AX_STR("hello world");
    CHECK(ax_str_contains(s, AX_STR("world")) == AX_TRUE, "expected true");
    CHECK(ax_str_contains(s, AX_STR("xyz")) == AX_FALSE, "expected false");
    CHECK(ax_str_contains(s, AX_STR("")) == AX_TRUE, "empty is contained");
    PASS();
}

static void test_str_starts_with(void) {
    TEST("str_starts_with");
    ax_string s = AX_STR("hello world");
    CHECK(ax_str_starts_with(s, AX_STR("hello")) == AX_TRUE, "expected true");
    CHECK(ax_str_starts_with(s, AX_STR("world")) == AX_FALSE, "expected false");
    CHECK(ax_str_starts_with(s, AX_STR("")) == AX_TRUE, "empty prefix");
    PASS();
}

static void test_str_ends_with(void) {
    TEST("str_ends_with");
    ax_string s = AX_STR("hello world");
    CHECK(ax_str_ends_with(s, AX_STR("world")) == AX_TRUE, "expected true");
    CHECK(ax_str_ends_with(s, AX_STR("hello")) == AX_FALSE, "expected false");
    PASS();
}

static void test_str_index_of(void) {
    TEST("str_index_of");
    ax_string s = AX_STR("hello world hello");
    CHECK(ax_str_index_of(s, AX_STR("world")) == 6, "expected 6");
    CHECK(ax_str_index_of(s, AX_STR("xyz")) == -1, "expected -1");
    CHECK(ax_str_index_of(s, AX_STR("hello")) == 0, "expected 0");
    PASS();
}

static void test_str_trim(void) {
    TEST("str_trim");
    ax_string s = AX_STR("  hello  ");
    ax_string t = ax_str_trim(s);
    CHECK(t.len == 5, "expected len 5");
    CHECK(memcmp(t.ptr, "hello", 5) == 0, "content mismatch");

    ax_string no_ws = ax_str_trim(AX_STR("abc"));
    CHECK(no_ws.len == 3, "no trim needed");
    PASS();
}

static void test_str_eq(void) {
    TEST("str_eq");
    CHECK(ax_str_eq(AX_STR("hello"), AX_STR("hello")) == AX_TRUE, "same");
    CHECK(ax_str_eq(AX_STR("hello"), AX_STR("world")) == AX_FALSE, "diff");
    CHECK(ax_str_eq(AX_STR(""), AX_STR("")) == AX_TRUE, "both empty");
    CHECK(ax_str_eq(AX_STR("hi"), AX_STR("hi!")) == AX_FALSE, "diff len");
    PASS();
}

static void test_i64_to_str(void) {
    TEST("i64_to_str");
    ax_string s = ax_i64_to_str(42);
    CHECK(s.len == 2, "expected len 2");
    CHECK(memcmp(s.ptr, "42", 2) == 0, "expected '42'");
    free((void*)s.ptr);

    ax_string neg = ax_i64_to_str(-123);
    CHECK(neg.len == 4, "expected len 4");
    CHECK(memcmp(neg.ptr, "-123", 4) == 0, "expected '-123'");
    free((void*)neg.ptr);

    ax_string zero = ax_i64_to_str(0);
    CHECK(zero.len == 1, "expected len 1");
    CHECK(zero.ptr[0] == '0', "expected '0'");
    free((void*)zero.ptr);
    PASS();
}

static void test_bool_to_str(void) {
    TEST("bool_to_str");
    ax_string t = ax_bool_to_str(AX_TRUE);
    CHECK(ax_str_eq(t, AX_STR("true")), "expected 'true'");
    ax_string f = ax_bool_to_str(AX_FALSE);
    CHECK(ax_str_eq(f, AX_STR("false")), "expected 'false'");
    PASS();
}

/* ================================================================
 * Math Tests
 * ================================================================ */

static void test_abs(void) {
    TEST("abs_i64");
    CHECK(ax_abs_i64(5) == 5, "pos");
    CHECK(ax_abs_i64(-5) == 5, "neg");
    CHECK(ax_abs_i64(0) == 0, "zero");
    PASS();
}

static void test_min_max(void) {
    TEST("min_max_i64");
    CHECK(ax_min_i64(3, 7) == 3, "min");
    CHECK(ax_max_i64(3, 7) == 7, "max");
    CHECK(ax_min_i64(-1, -5) == -5, "min neg");
    CHECK(ax_max_i64(-1, -5) == -1, "max neg");
    PASS();
}

static void test_clamp(void) {
    TEST("clamp_i64");
    CHECK(ax_clamp_i64(5, 0, 10) == 5, "in range");
    CHECK(ax_clamp_i64(-1, 0, 10) == 0, "below");
    CHECK(ax_clamp_i64(15, 0, 10) == 10, "above");
    PASS();
}

static void test_pow_i64(void) {
    TEST("pow_i64");
    CHECK(ax_pow_i64(2, 10) == 1024, "2^10");
    CHECK(ax_pow_i64(3, 5) == 243, "3^5");
    CHECK(ax_pow_i64(7, 0) == 1, "x^0");
    CHECK(ax_pow_i64(1, 100) == 1, "1^100");
    PASS();
}

static void test_gcd_lcm(void) {
    TEST("gcd / lcm");
    CHECK(ax_gcd(12, 8) == 4, "gcd(12,8)");
    CHECK(ax_gcd(17, 5) == 1, "coprime");
    CHECK(ax_gcd(0, 5) == 5, "gcd(0,5)");
    CHECK(ax_lcm(4, 6) == 12, "lcm(4,6)");
    CHECK(ax_lcm(0, 5) == 0, "lcm(0,5)");
    PASS();
}

/* ================================================================
 * Vec Tests
 * ================================================================ */

static void test_vec_basic(void) {
    TEST("vec push/pop/get");
    ax_vec v = ax_vec_new(sizeof(ax_i64));
    CHECK(v.len == 0, "empty initially");

    ax_i64 vals[] = {10, 20, 30, 40, 50};
    for (int i = 0; i < 5; i++) {
        ax_vec_push(&v, &vals[i]);
    }
    CHECK(v.len == 5, "len after 5 pushes");

    ax_i64* got = (ax_i64*)ax_vec_get(&v, 2);
    CHECK(*got == 30, "v[2] == 30");

    ax_i64 popped;
    CHECK(ax_vec_pop(&v, &popped) == AX_TRUE, "pop succeeds");
    CHECK(popped == 50, "popped 50");
    CHECK(v.len == 4, "len after pop");

    ax_vec_free(&v);
    PASS();
}

static void test_vec_grow(void) {
    TEST("vec grow");
    ax_vec v = ax_vec_new(sizeof(ax_i32));
    for (ax_i32 i = 0; i < 100; i++) {
        ax_vec_push(&v, &i);
    }
    CHECK(v.len == 100, "100 elements");
    CHECK(v.cap >= 100, "cap >= 100");

    ax_i32* val50 = (ax_i32*)ax_vec_get(&v, 50);
    CHECK(*val50 == 50, "v[50] == 50");

    ax_vec_clear(&v);
    CHECK(v.len == 0, "cleared");
    CHECK(v.cap >= 100, "cap preserved after clear");

    ax_vec_free(&v);
    PASS();
}

static void test_vec_set(void) {
    TEST("vec set");
    ax_vec v = ax_vec_with_capacity(sizeof(ax_i64), 10);
    ax_i64 val = 42;
    ax_vec_push(&v, &val);
    val = 99;
    ax_vec_set(&v, 0, &val);

    ax_i64* got = (ax_i64*)ax_vec_get(&v, 0);
    CHECK(*got == 99, "set to 99");
    ax_vec_free(&v);
    PASS();
}

/* ================================================================
 * Arena Tests
 * ================================================================ */

static void test_arena_basic(void) {
    TEST("arena alloc/reset");
    ax_arena a = ax_arena_new(4096);
    CHECK(ax_arena_remaining(&a) == 4096, "4096 remaining");

    int* p = (int*)ax_arena_alloc(&a, sizeof(int), 4);
    CHECK(p != NULL, "alloc succeeded");
    *p = 42;
    CHECK(*p == 42, "value stored");
    CHECK(ax_arena_used(&a) > 0, "some used");

    ax_arena_reset(&a);
    CHECK(ax_arena_used(&a) == 0, "reset to 0");
    CHECK(ax_arena_remaining(&a) == 4096, "fully reclaimed");

    ax_arena_destroy(&a);
    PASS();
}

static void test_arena_alignment(void) {
    TEST("arena alignment");
    ax_arena a = ax_arena_new(4096);

    /* Allocate with different alignments */
    void* p1 = ax_arena_alloc(&a, 1, 1);   /* 1-byte aligned */
    void* p2 = ax_arena_alloc(&a, 4, 8);   /* 8-byte aligned */
    void* p3 = ax_arena_alloc(&a, 16, 16); /* 16-byte aligned */

    CHECK(((uintptr_t)p1 % 1) == 0, "1-byte aligned");
    CHECK(((uintptr_t)p2 % 8) == 0, "8-byte aligned");
    CHECK(((uintptr_t)p3 % 16) == 0, "16-byte aligned");

    ax_arena_destroy(&a);
    PASS();
}

/* ================================================================
 * Print Tests (just verify they don't crash)
 * ================================================================ */

static void test_print_functions(void) {
    TEST("print functions (no crash)");
    ax_print_str(AX_STR(""));           /* empty */
    ax_print_i64(12345);
    ax_print_f64(3.14159);
    ax_print_bool(AX_TRUE);
    ax_print_str(AX_STR(" "));          /* separator */
    ax_println_str(AX_STR("end"));
    PASS();
}

/* ================================================================
 * Main
 * ================================================================ */

int main(void) {
    printf("=== AXIOM C Runtime Stdlib Tests ===\n\n");

    printf("[String Operations]\n");
    test_str_len();
    test_str_char_count();
    test_str_concat();
    test_str_slice();
    test_str_contains();
    test_str_starts_with();
    test_str_ends_with();
    test_str_index_of();
    test_str_trim();
    test_str_eq();
    test_i64_to_str();
    test_bool_to_str();

    printf("\n[Math Operations]\n");
    test_abs();
    test_min_max();
    test_clamp();
    test_pow_i64();
    test_gcd_lcm();

    printf("\n[Vec Operations]\n");
    test_vec_basic();
    test_vec_grow();
    test_vec_set();

    printf("\n[Arena Operations]\n");
    test_arena_basic();
    test_arena_alignment();

    printf("\n[Print Functions]\n");
    test_print_functions();

    printf("\n=== Results: %d passed, %d failed ===\n",
           tests_passed, tests_failed);
    return tests_failed > 0 ? 1 : 0;
}
