/**
 * ax_assert.c — Real C implementations for AXIOM assert functions.
 */

#include "ax_stdlib.h"
#include <stdio.h>
#include <stdlib.h>

void ax_assert_axiom(ax_bool condition, ax_string message) {
    if (!condition) {
        fprintf(stderr, "assertion failed: ");
        if (message.len > 0 && message.ptr) {
            fwrite(message.ptr, 1, message.len, stderr);
        }
        fputc('\n', stderr);
        fflush(stderr);
        ax_panic("assertion failed");
    }
}

void ax_assert_eq_i64(ax_i64 actual, ax_i64 expected) {
    if (actual != expected) {
        fprintf(stderr, "assert_eq failed: %lld != %lld\n",
            (long long)actual, (long long)expected);
        fflush(stderr);
        ax_panic("assert_eq failed");
    }
}

void ax_assert_eq_str(ax_string actual, ax_string expected) {
    if (!ax_str_eq(actual, expected)) {
        fprintf(stderr, "assert_eq failed: \"");
        if (actual.ptr) fwrite(actual.ptr, 1, actual.len, stderr);
        fprintf(stderr, "\" != \"");
        if (expected.ptr) fwrite(expected.ptr, 1, expected.len, stderr);
        fprintf(stderr, "\"\n");
        fflush(stderr);
        ax_panic("assert_eq failed: strings not equal");
    }
}

void ax_assert_eq_bool(ax_bool actual, ax_bool expected) {
    if (actual != expected) {
        fprintf(stderr, "assert_eq failed: %s != %s\n",
            actual ? "true" : "false", expected ? "true" : "false");
        fflush(stderr);
        ax_panic("assert_eq failed");
    }
}
