/**
 * ax_assert.c — Real C implementations for AXIOM assert functions.
 */

#include "ax_stdlib.h"

#ifdef _WIN32
#include <windows.h>
static void ax_write_stderr(const char* buf, int len) {
    HANDLE hStderr = GetStdHandle(STD_ERROR_HANDLE);
    DWORD written;
    WriteFile(hStderr, buf, len, &written, NULL);
}
#else
#include <unistd.h>
#include <sys/syscall.h>
static void ax_write_stderr(const char* buf, int len) {
    syscall(SYS_write, 2, buf, len);
}
#endif

static int helper_i64_to_str(char* buf, int buf_size, int64_t val) {
    if (buf_size < 24) return 0;
    int is_neg = 0;
    if (val < 0) {
        is_neg = 1;
        val = -val;
    }
    int idx = 0;
    do {
        buf[idx++] = '0' + (val % 10);
        val /= 10;
    } while (val > 0 && idx < buf_size - 1);
    if (is_neg && idx < buf_size - 1) {
        buf[idx++] = '-';
    }
    for (int i = 0; i < idx / 2; ++i) {
        char temp = buf[i];
        buf[i] = buf[idx - 1 - i];
        buf[idx - 1 - i] = temp;
    }
    buf[idx] = '\0';
    return idx;
}

void ax_assert_axiom(ax_bool condition, ax_string message) {
    if (!condition) {
        ax_write_stderr("assertion failed: ", 18);
        if (message.len > 0 && message.ptr) {
            ax_write_stderr((const char*)message.ptr, message.len);
        }
        ax_write_stderr("\n", 1);
        ax_panic("assertion failed");
    }
}

void ax_assert_eq_i64(ax_i64 actual, ax_i64 expected) {
    if (actual != expected) {
        char buf1[32];
        char buf2[32];
        int len1 = helper_i64_to_str(buf1, sizeof(buf1), actual);
        int len2 = helper_i64_to_str(buf2, sizeof(buf2), expected);
        
        ax_write_stderr("assert_eq failed: ", 18);
        ax_write_stderr(buf1, len1);
        ax_write_stderr(" != ", 4);
        ax_write_stderr(buf2, len2);
        ax_write_stderr("\n", 1);
        ax_panic("assert_eq failed");
    }
}

void ax_assert_eq_str(ax_string actual, ax_string expected) {
    if (!ax_str_eq(actual, expected)) {
        ax_write_stderr("assert_eq failed: \"", 19);
        if (actual.ptr) ax_write_stderr((const char*)actual.ptr, actual.len);
        ax_write_stderr("\" != \"", 6);
        if (expected.ptr) ax_write_stderr((const char*)expected.ptr, expected.len);
        ax_write_stderr("\"\n", 2);
        ax_panic("assert_eq failed: strings not equal");
    }
}

void ax_assert_eq_bool(ax_bool actual, ax_bool expected) {
    if (actual != expected) {
        ax_write_stderr("assert_eq failed: ", 18);
        if (actual) ax_write_stderr("true", 4);
        else ax_write_stderr("false", 5);
        ax_write_stderr(" != ", 4);
        if (expected) ax_write_stderr("true\n", 5);
        else ax_write_stderr("false\n", 6);
        ax_panic("assert_eq failed");
    }
}
