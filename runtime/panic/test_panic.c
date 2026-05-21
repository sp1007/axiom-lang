/**
 * test_panic.c — Unit tests for AXIOM panic handler.
 *
 * Uses setjmp/longjmp to intercept ax_panic for testing.
 * This file defines its own ax_panic to override the real one.
 *
 * Compile:
 *   gcc -O2 -Wall -Wextra -Werror -std=c11 test_panic.c -o test_panic
 *   ./test_panic
 */
#include <stdio.h>
#include <setjmp.h>
#include <string.h>

/* We need panic.h for ax_bounds_check and ax_assert, but we override ax_panic. */
#include "panic.h"

/* Stub for ax_set_program_name since we can't link panic.c (conflicting ax_panic) */
static const char* test_program_name = "<test>";
void ax_set_program_name(const char* name) {
    if (name) test_program_name = name;
}

/* Test harness state */
static jmp_buf  test_jmp;
static char     last_panic_msg[256];
static int      panic_triggered;
static int      pass_count = 0;
static int      test_count = 0;

/* Override ax_panic for testing: longjmp instead of abort. */
AX_NORETURN void ax_panic(const char* msg) {
    strncpy(last_panic_msg, msg, sizeof(last_panic_msg) - 1);
    last_panic_msg[sizeof(last_panic_msg) - 1] = '\0';
    panic_triggered = 1;
    longjmp(test_jmp, 1);
    /* longjmp never returns, satisfying noreturn */
}

#define ASSERT(cond, name) do { \
    test_count++; \
    if (cond) { pass_count++; printf("  [PASS] %s\n", name); } \
    else { printf("  [FAIL] %s\n", name); } \
} while(0)

#define ASSERT_PANIC(expr, name) do { \
    test_count++; \
    panic_triggered = 0; \
    last_panic_msg[0] = '\0'; \
    if (setjmp(test_jmp) == 0) { expr; } \
    if (panic_triggered) { pass_count++; printf("  [PASS] %s\n", name); } \
    else { printf("  [FAIL] %s (no panic)\n", name); } \
} while(0)

/* Test: bounds check with valid index doesn't panic */
static void test_bounds_valid(void) {
    ax_bounds_check(0, 10);
    ASSERT(1, "bounds_check(0, 10) no panic");

    ax_bounds_check(9, 10);
    ASSERT(1, "bounds_check(9, 10) no panic");
}

/* Test: bounds check with invalid index panics */
static void test_bounds_invalid(void) {
    ASSERT_PANIC(ax_bounds_check(10, 10), "bounds_check(10, 10) panics");
    ASSERT(strstr(last_panic_msg, "index out of bounds") != NULL,
           "bounds_check panic message contains 'index out of bounds'");

    ASSERT_PANIC(ax_bounds_check(100, 10), "bounds_check(100, 10) panics");
}

/* Test: assert true doesn't panic */
static void test_assert_true(void) {
    ax_assert(1, "should not fail");
    ASSERT(1, "ax_assert(true) no panic");
}

/* Test: assert false panics */
static void test_assert_false(void) {
    ASSERT_PANIC(ax_assert(0, "test assertion failed"), "ax_assert(false) panics");
    ASSERT(strstr(last_panic_msg, "test assertion failed") != NULL,
           "ax_assert passes message to panic");
}

/* Test: bounds check edge cases */
static void test_bounds_edge(void) {
    ASSERT_PANIC(ax_bounds_check(0, 0), "bounds_check(0, 0) panics (empty array)");
    ax_bounds_check(0, 1);
    ASSERT(1, "bounds_check(0, 1) no panic");
}

/* Test: program name can be set */
static void test_set_program_name(void) {
    ax_set_program_name("test_program");
    ASSERT(1, "ax_set_program_name doesn't crash");
    ax_set_program_name(NULL); /* null is handled */
    ASSERT(1, "ax_set_program_name(NULL) doesn't crash");
}

int main(void) {
    printf("=== panic handler unit tests ===\n\n");

    test_bounds_valid();
    test_bounds_invalid();
    test_assert_true();
    test_assert_false();
    test_bounds_edge();
    test_set_program_name();

    printf("\n=== Results: %d/%d tests passed ===\n", pass_count, test_count);
    return (pass_count == test_count) ? 0 : 1;
}
