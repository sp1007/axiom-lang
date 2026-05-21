/**
 * ax_print.c — Real C implementations for AXIOM print/println/eprint.
 *
 * These are the functions called by the C backend when user code
 * calls print() or println().
 */

#include "ax_stdlib.h"
#include <stdio.h>

/* ================================================================
 * Print to stdout
 * ================================================================ */

void ax_print_str(ax_string s) {
    if (s.len > 0 && s.ptr != NULL) {
        fwrite(s.ptr, 1, s.len, stdout);
    }
}

void ax_println_str(ax_string s) {
    ax_print_str(s);
    putchar('\n');
    fflush(stdout);
}

void ax_print_i64(ax_i64 value) {
    printf("%lld", (long long)value);
}

void ax_println_i64(ax_i64 value) {
    printf("%lld\n", (long long)value);
    fflush(stdout);
}

void ax_print_f64(ax_f64 value) {
    printf("%.6g", value);
}

void ax_println_f64(ax_f64 value) {
    printf("%.6g\n", value);
    fflush(stdout);
}

void ax_print_bool(ax_bool value) {
    fputs(value ? "true" : "false", stdout);
}

void ax_println_bool(ax_bool value) {
    puts(value ? "true" : "false");
    fflush(stdout);
}

/* ================================================================
 * Print to stderr
 * ================================================================ */

void ax_eprint_str(ax_string s) {
    if (s.len > 0 && s.ptr != NULL) {
        fwrite(s.ptr, 1, s.len, stderr);
    }
}

void ax_eprintln_str(ax_string s) {
    ax_eprint_str(s);
    fputc('\n', stderr);
    fflush(stderr);
}
