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

void ax_print_str_native(const char* ptr) {
    if (!ptr) return;
    ax_string s = { (const ax_u8*)ptr, strlen(ptr) };
    ax_print_str(s);
}

void ax_println_str_native(const char* ptr) {
    if (!ptr) return;
    ax_string s = { (const ax_u8*)ptr, strlen(ptr) };
    ax_println_str(s);
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

/* ================================================================
 * Freestanding C-Independent File I/O Implementations
 * ================================================================ */

#if defined(_WIN32)

void* ax_fopen(const char* filename, const char* mode) {
    DWORD access = 0;
    DWORD creation = 0;
    
    if (strchr(mode, 'w')) {
        access = GENERIC_WRITE;
        creation = CREATE_ALWAYS;
    } else if (strchr(mode, 'r')) {
        access = GENERIC_READ;
        creation = OPEN_EXISTING;
    } else if (strchr(mode, 'a')) {
        access = GENERIC_WRITE;
        creation = OPEN_ALWAYS;
    } else {
        access = GENERIC_READ;
        creation = OPEN_EXISTING;
    }
    
    HANDLE h = CreateFileA(filename, access, FILE_SHARE_READ, NULL, creation, FILE_ATTRIBUTE_NORMAL, NULL);
    if (h == INVALID_HANDLE_VALUE) {
        return NULL;
    }
    return (void*)h;
}

int ax_fclose(void* stream) {
    if (!stream) return -1;
    return CloseHandle((HANDLE)stream) ? 0 : -1;
}

size_t ax_fread(void* buffer, size_t size, size_t count, void* stream) {
    if (!stream || !buffer || size == 0 || count == 0) return 0;
    size_t bytes_to_read = size * count;
    DWORD read_bytes = 0;
    if (ReadFile((HANDLE)stream, buffer, (DWORD)bytes_to_read, &read_bytes, NULL)) {
        return (size_t)read_bytes / size;
    }
    return 0;
}

size_t ax_fwrite(const void* buffer, size_t size, size_t count, void* stream) {
    if (!stream || !buffer || size == 0 || count == 0) return 0;
    size_t bytes_to_write = size * count;
    DWORD written = 0;
    if (WriteFile((HANDLE)stream, buffer, (DWORD)bytes_to_write, &written, NULL)) {
        return (size_t)written / size;
    }
    return 0;
}

int ax_fputs_custom(const char* s, void* stream) {
    if (!s || !stream) return -1;
    size_t len = strlen(s);
    size_t written = ax_fwrite(s, 1, len, stream);
    return written == len ? 0 : -1;
}

int ax_fseek(void* stream, long offset, int origin) {
    if (!stream) return -1;
    DWORD method = 0;
    if (origin == 0) method = FILE_BEGIN;      // SEEK_SET = 0
    else if (origin == 1) method = FILE_CURRENT; // SEEK_CUR = 1
    else if (origin == 2) method = FILE_END;     // SEEK_END = 2
    
    DWORD res = SetFilePointer((HANDLE)stream, (LONG)offset, NULL, method);
    if (res == INVALID_SET_FILE_POINTER && GetLastError() != NO_ERROR) {
        return -1;
    }
    return 0;
}

long ax_ftell(void* stream) {
    if (!stream) return -1;
    DWORD res = SetFilePointer((HANDLE)stream, 0, NULL, FILE_CURRENT);
    if (res == INVALID_SET_FILE_POINTER && GetLastError() != NO_ERROR) {
        return -1;
    }
    return (long)res;
}

void ax_rewind(void* stream) {
    ax_fseek(stream, 0, 0); // SEEK_SET = 0
}

#else

/* Linux Freestanding Syscall Implementations */
#include <fcntl.h>

void* ax_fopen(const char* filename, const char* mode) {
    int flags = 0;
    if (strchr(mode, 'w')) {
        flags = O_WRONLY | O_CREAT | O_TRUNC;
    } else if (strchr(mode, 'r')) {
        flags = O_RDONLY;
    } else if (strchr(mode, 'a')) {
        flags = O_WRONLY | O_CREAT | O_APPEND;
    } else {
        flags = O_RDONLY;
    }
    
    long fd = syscall(SYS_open, filename, flags, 0666);
    if (fd < 0) {
        return NULL;
    }
    return (void*)fd;
}

int ax_fclose(void* stream) {
    if (!stream) return -1;
    return syscall(SYS_close, (long)stream) == 0 ? 0 : -1;
}

size_t ax_fread(void* buffer, size_t size, size_t count, void* stream) {
    if (!stream || !buffer || size == 0 || count == 0) return 0;
    size_t bytes_to_read = size * count;
    long n = syscall(SYS_read, (long)stream, buffer, bytes_to_read);
    if (n < 0) return 0;
    return (size_t)n / size;
}

size_t ax_fwrite(const void* buffer, size_t size, size_t count, void* stream) {
    if (!stream || !buffer || size == 0 || count == 0) return 0;
    size_t bytes_to_write = size * count;
    long n = syscall(SYS_write, (long)stream, buffer, bytes_to_write);
    if (n < 0) return 0;
    return (size_t)n / size;
}

int ax_fputs_custom(const char* s, void* stream) {
    if (!s || !stream) return -1;
    size_t len = strlen(s);
    size_t written = ax_fwrite(s, 1, len, stream);
    return written == len ? 0 : -1;
}

int ax_fseek(void* stream, long offset, int origin) {
    if (!stream) return -1;
    long res = syscall(SYS_lseek, (long)stream, offset, origin);
    return res < 0 ? -1 : 0;
}

long ax_ftell(void* stream) {
    if (!stream) return -1;
    return (long)syscall(SYS_lseek, (long)stream, 0, 1); // SEEK_CUR = 1
}

void ax_rewind(void* stream) {
    ax_fseek(stream, 0, 0); // SEEK_SET = 0
}

#endif


