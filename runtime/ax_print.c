#include "ax_stdlib.h"

#if defined(_WIN32)
#ifndef WIN32_LEAN_AND_MEAN
#define WIN32_LEAN_AND_MEAN
#endif
#ifndef NOMINMAX
#define NOMINMAX
#endif
#include <windows.h>

static void win32_write_str(HANDLE h, ax_string s) {
    if (s.len > 0 && s.ptr != NULL) {
        DWORD written;
        WriteFile(h, s.ptr, (DWORD)s.len, &written, NULL);
    }
}

static void win32_write_raw(HANDLE h, const char* ptr, DWORD len) {
    DWORD written;
    WriteFile(h, ptr, len, &written, NULL);
}

static void win32_print_i64(HANDLE h, ax_i64 val) {
    char buf[32];
    int len = 0;
    if (val == 0) {
        buf[len++] = '0';
    } else {
        int is_neg = 0;
        ax_u64 utemp;
        if (val < 0) {
            is_neg = 1;
            utemp = (ax_u64)(-val);
        } else {
            utemp = (ax_u64)val;
        }
        while (utemp > 0) {
            buf[len++] = '0' + (utemp % 10);
            utemp /= 10;
        }
        if (is_neg) {
            buf[len++] = '-';
        }
        for (int i = 0; i < len / 2; i++) {
            char t = buf[i];
            buf[i] = buf[len - 1 - i];
            buf[len - 1 - i] = t;
        }
    }
    DWORD written;
    WriteFile(h, buf, (DWORD)len, &written, NULL);
}

static void win32_print_f64(HANDLE h, ax_f64 val) {
    if (val != val) {
        DWORD written;
        WriteFile(h, "nan", 3, &written, NULL);
        return;
    }
    if (val < 0) {
        DWORD written;
        WriteFile(h, "-", 1, &written, NULL);
        val = -val;
    }
    ax_i64 int_part = (ax_i64)val;
    ax_f64 frac_part = val - (ax_f64)int_part;
    win32_print_i64(h, int_part);
    DWORD written;
    WriteFile(h, ".", 1, &written, NULL);
    for (int i = 0; i < 6; i++) {
        frac_part *= 10.0;
        int digit = (int)frac_part;
        if (digit < 0) digit = 0;
        if (digit > 9) digit = 9;
        char c = '0' + digit;
        WriteFile(h, &c, 1, &written, NULL);
        frac_part -= digit;
    }
}

void ax_print_str(ax_string s) {
    win32_write_str(GetStdHandle(STD_OUTPUT_HANDLE), s);
}

void ax_println_str(ax_string s) {
    HANDLE h = GetStdHandle(STD_OUTPUT_HANDLE);
    win32_write_str(h, s);
    win32_write_raw(h, "\n", 1);
}

void ax_print_i64(ax_i64 value) {
    win32_print_i64(GetStdHandle(STD_OUTPUT_HANDLE), value);
}

void ax_println_i64(ax_i64 value) {
    HANDLE h = GetStdHandle(STD_OUTPUT_HANDLE);
    win32_print_i64(h, value);
    win32_write_raw(h, "\n", 1);
}

void ax_print_f64(ax_f64 value) {
    win32_print_f64(GetStdHandle(STD_OUTPUT_HANDLE), value);
}

void ax_println_f64(ax_f64 value) {
    HANDLE h = GetStdHandle(STD_OUTPUT_HANDLE);
    win32_print_f64(h, value);
    win32_write_raw(h, "\n", 1);
}

void ax_print_bool(ax_bool value) {
    win32_write_raw(GetStdHandle(STD_OUTPUT_HANDLE), value ? "true" : "false", value ? 4 : 5);
}

void ax_println_bool(ax_bool value) {
    HANDLE h = GetStdHandle(STD_OUTPUT_HANDLE);
    win32_write_raw(h, value ? "true\n" : "false\n", value ? 5 : 6);
}

void ax_eprint_str(ax_string s) {
    win32_write_str(GetStdHandle(STD_ERROR_HANDLE), s);
}

void ax_eprintln_str(ax_string s) {
    HANDLE h = GetStdHandle(STD_ERROR_HANDLE);
    win32_write_str(h, s);
    win32_write_raw(h, "\n", 1);
}

#else
// Linux Direct Syscalls / Assembly implementation
#include <unistd.h>
#include <sys/syscall.h>

static void posix_write_str(int fd, ax_string s) {
    if (s.len > 0 && s.ptr != NULL) {
        syscall(SYS_write, fd, s.ptr, s.len);
    }
}

static void posix_write_raw(int fd, const char* ptr, size_t len) {
    syscall(SYS_write, fd, ptr, len);
}

static void posix_print_i64(int fd, ax_i64 val) {
    char buf[32];
    int len = 0;
    if (val == 0) {
        buf[len++] = '0';
    } else {
        int is_neg = 0;
        ax_u64 utemp;
        if (val < 0) {
            is_neg = 1;
            utemp = (ax_u64)(-val);
        } else {
            utemp = (ax_u64)val;
        }
        while (utemp > 0) {
            buf[len++] = '0' + (utemp % 10);
            utemp /= 10;
        }
        if (is_neg) {
            buf[len++] = '-';
        }
        for (int i = 0; i < len / 2; i++) {
            char t = buf[i];
            buf[i] = buf[len - 1 - i];
            buf[len - 1 - i] = t;
        }
    }
    syscall(SYS_write, fd, buf, len);
}

static void posix_print_f64(int fd, ax_f64 val) {
    if (val != val) {
        syscall(SYS_write, fd, "nan", 3);
        return;
    }
    if (val < 0) {
        syscall(SYS_write, fd, "-", 1);
        val = -val;
    }
    ax_i64 int_part = (ax_i64)val;
    ax_f64 frac_part = val - (ax_f64)int_part;
    posix_print_i64(fd, int_part);
    syscall(SYS_write, fd, ".", 1);
    for (int i = 0; i < 6; i++) {
        frac_part *= 10.0;
        int digit = (int)frac_part;
        if (digit < 0) digit = 0;
        if (digit > 9) digit = 9;
        char c = '0' + digit;
        syscall(SYS_write, fd, &c, 1);
        frac_part -= digit;
    }
}

void ax_print_str(ax_string s) {
    posix_write_str(1, s);
}

void ax_println_str(ax_string s) {
    posix_write_str(1, s);
    posix_write_raw(1, "\n", 1);
}

void ax_print_i64(ax_i64 value) {
    posix_print_i64(1, value);
}

void ax_println_i64(ax_i64 value) {
    posix_print_i64(1, value);
    posix_write_raw(1, "\n", 1);
}

void ax_print_f64(ax_f64 value) {
    posix_print_f64(1, value);
}

void ax_println_f64(ax_f64 value) {
    posix_print_f64(1, value);
    posix_write_raw(1, "\n", 1);
}

void ax_print_bool(ax_bool value) {
    posix_write_raw(1, value ? "true" : "false", value ? 4 : 5);
}

void ax_println_bool(ax_bool value) {
    posix_write_raw(1, value ? "true\n" : "false\n", value ? 5 : 6);
}

void ax_eprint_str(ax_string s) {
    posix_write_str(2, s);
}

void ax_eprintln_str(ax_string s) {
    posix_write_str(2, s);
    posix_write_raw(2, "\n", 1);
}

#endif

