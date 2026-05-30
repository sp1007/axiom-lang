#include "panic.h"

static const char* program_name = "<unknown>";

void ax_set_program_name(const char* name) {
    if (name) program_name = name;
}

static size_t my_strlen(const char* s) {
    size_t len = 0;
    while (s && s[len]) len++;
    return len;
}

#if defined(_WIN32)

#ifndef WIN32_LEAN_AND_MEAN
#define WIN32_LEAN_AND_MEAN
#endif
#ifndef NOMINMAX
#define NOMINMAX
#endif
#include <windows.h>

// Stub for extern "C" fn syscall on Windows to satisfy the linker.
// This is never called at runtime on Windows because memory allocations are routed via VirtualAlloc/VirtualFree.
long long syscall(long long num, long long a1, long long a2, long long a3, long long a4, long long a5, long long a6) {
    (void)num; (void)a1; (void)a2; (void)a3; (void)a4; (void)a5; (void)a6;
    return 0;
}

AX_NORETURN void ax_panic(const char* msg) {
    HANDLE hErr = GetStdHandle(STD_ERROR_HANDLE);
    if (hErr != INVALID_HANDLE_VALUE) {
        DWORD written;
        const char* prefix = "\nAXIOM PANIC in '";
        WriteFile(hErr, prefix, (DWORD)my_strlen(prefix), &written, NULL);
        WriteFile(hErr, program_name, (DWORD)my_strlen(program_name), &written, NULL);
        const char* mid = "': ";
        WriteFile(hErr, mid, (DWORD)my_strlen(mid), &written, NULL);
        WriteFile(hErr, msg, (DWORD)my_strlen(msg), &written, NULL);
        WriteFile(hErr, "\n", 1, &written, NULL);

        const char* trace_title = "Stack trace:\n";
        WriteFile(hErr, trace_title, (DWORD)my_strlen(trace_title), &written, NULL);
        void* frames[32];
        USHORT count = CaptureStackBackTrace(0, 32, frames, NULL);
        for (USHORT i = 0; i < count; i++) {
            char buf[64];
            int len = 0;
            buf[len++] = ' '; buf[len++] = ' '; buf[len++] = '#';
            if (i < 10) {
                buf[len++] = '0' + i;
            } else {
                buf[len++] = '0' + (i / 10);
                buf[len++] = '0' + (i % 10);
            }
            buf[len++] = ' '; buf[len++] = ' '; buf[len++] = '0'; buf[len++] = 'x';
            uintptr_t val = (uintptr_t)frames[i];
            for (int shift = 60; shift >= 0; shift -= 4) {
                int digit = (val >> shift) & 0xF;
                buf[len++] = (digit < 10) ? ('0' + digit) : ('a' + digit - 10);
            }
            buf[len++] = '\n';
            WriteFile(hErr, buf, (DWORD)len, &written, NULL);
        }
    }
    ExitProcess(101);
}

#elif defined(__linux__) || defined(__APPLE__)

#include <unistd.h>
#include <sys/syscall.h>

AX_NORETURN void ax_panic(const char* msg) {
    const char* prefix = "\nAXIOM PANIC in '";
    syscall(SYS_write, 2, prefix, my_strlen(prefix));
    syscall(SYS_write, 2, program_name, my_strlen(program_name));
    const char* mid = "': ";
    syscall(SYS_write, 2, mid, my_strlen(mid));
    syscall(SYS_write, 2, msg, my_strlen(msg));
    syscall(SYS_write, 2, "\n", 1);
    syscall(SYS_exit, 101);
    while (1) {} // Unreachable
}

#else

/* Fallback: direct exit */
AX_NORETURN void ax_panic(const char* msg) {
    ExitProcess(101);
}

#endif

#if !defined(_WIN32)
#include <sys/mman.h>
#endif

static void* g_ax_global_state = NULL;
void* ax_get_global_state_internal(void) {
    if (g_ax_global_state == NULL) {
#if defined(_WIN32)
        g_ax_global_state = VirtualAlloc(NULL, 4096, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
#else
        g_ax_global_state = mmap(NULL, 4096, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
#endif
    }
    return g_ax_global_state;
}

static int my_strcmp(const char* s1, const char* s2) {
    while (*s1 && (*s1 == *s2)) {
        s1++;
        s2++;
    }
    return *(const unsigned char*)s1 - *(const unsigned char*)s2;
}

long long ax_compiler_intrinsic(const char* name, void* p1, void* p2, void* p3) {
    if (my_strcmp(name, "atomic_load") == 0) {
        return (long long)__atomic_load_n((long long*)p1, 5 /* __ATOMIC_SEQ_CST */);
    }
    if (my_strcmp(name, "atomic_store") == 0) {
        __atomic_store_n((long long*)p1, (long long)p2, 5 /* __ATOMIC_SEQ_CST */);
        return 0;
    }
    if (my_strcmp(name, "atomic_cas") == 0) {
        long long expected = (long long)p2;
        return __atomic_compare_exchange_n((long long*)p1, &expected, (long long)p3, 0, 5 /* __ATOMIC_SEQ_CST */, 5 /* __ATOMIC_SEQ_CST */);
    }
    if (my_strcmp(name, "is_windows") == 0) {
#if defined(_WIN32)
        return 1;
#else
        return 0;
#endif
    }
    return 0;
}


