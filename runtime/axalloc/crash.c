/*
 * p14-t07: AxAlloc Crash Cleanup — Implementation
 */

#include "crash.h"
#include <string.h>

#ifdef _WIN32
  #include <windows.h>
#else
  #include <signal.h>
  #include <unistd.h>
  #include <sys/mman.h>
#endif

/* --------------------------------------------------------------------------
 * Global Registry (async-signal-safe: only atomic loads in handler)
 * -------------------------------------------------------------------------- */

static AxAllocRegistry g_registry[AX_MAX_REGISTERED_ALLOCATORS];
static int g_registry_count = 0;
static volatile int g_in_crash_handler = 0;

/* --------------------------------------------------------------------------
 * Registration
 * -------------------------------------------------------------------------- */

int ax_register_allocator(void* base, size_t size, AxAllocType type) {
    if (g_registry_count >= AX_MAX_REGISTERED_ALLOCATORS) return -1;

    int handle = g_registry_count;
    g_registry[handle].base = base;
    g_registry[handle].size = size;
    g_registry[handle].type = type;
    g_registry[handle].active = 1;
    g_registry_count++;
    return handle;
}

void ax_unregister_allocator(int handle) {
    if (handle < 0 || handle >= AX_MAX_REGISTERED_ALLOCATORS) return;
    g_registry[handle].active = 0;
    g_registry[handle].base = NULL;
}

/* --------------------------------------------------------------------------
 * Crash Cleanup (async-signal-safe)
 * -------------------------------------------------------------------------- */

static void safe_write(const char* msg) {
#ifdef _WIN32
    HANDLE h = GetStdHandle(STD_ERROR_HANDLE);
    DWORD written;
    WriteFile(h, msg, (DWORD)strlen(msg), &written, NULL);
#else
    size_t len = 0;
    const char* p = msg;
    while (*p++) len++;
    (void)write(STDERR_FILENO, msg, len);
#endif
}

void ax_crash_cleanup(void) {
    /* Prevent re-entry */
    if (g_in_crash_handler) return;
    g_in_crash_handler = 1;

    safe_write("[axiom] crash cleanup: releasing resources...\n");

    for (int i = 0; i < g_registry_count; i++) {
        if (!g_registry[i].active) continue;
        if (!g_registry[i].base) continue;

        switch (g_registry[i].type) {
        case AX_ALLOC_TYPE_GPU_PINNED:
        case AX_ALLOC_TYPE_SHARED_MEM:
        case AX_ALLOC_TYPE_FILE_BACKED:
            /* Release non-anonymous mappings */
#ifdef _WIN32
            VirtualFree(g_registry[i].base, 0, MEM_RELEASE);
#else
            munmap(g_registry[i].base, g_registry[i].size);
#endif
            break;
        case AX_ALLOC_TYPE_HEAP:
            /* Regular heap: let OS reclaim on exit */
            break;
        }

        g_registry[i].active = 0;
        g_registry[i].base = NULL;
    }

    safe_write("[axiom] crash cleanup complete.\n");
}

/* --------------------------------------------------------------------------
 * Signal / SEH Handler Registration
 * -------------------------------------------------------------------------- */

#ifdef _WIN32

static LONG WINAPI ax_seh_handler(EXCEPTION_POINTERS* ep) {
    (void)ep;
    ax_crash_cleanup();
    return EXCEPTION_CONTINUE_SEARCH;
}

void ax_register_crash_cleanup(void) {
    SetUnhandledExceptionFilter(ax_seh_handler);
}

#else

static void ax_signal_handler(int sig) {
    (void)sig;
    ax_crash_cleanup();
    /* Re-raise to get default behavior (core dump etc.) */
    signal(sig, SIG_DFL);
    raise(sig);
}

void ax_register_crash_cleanup(void) {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = ax_signal_handler;
    sa.sa_flags = SA_RESETHAND; /* one-shot */

    sigaction(SIGSEGV, &sa, NULL);
    sigaction(SIGABRT, &sa, NULL);
    sigaction(SIGBUS,  &sa, NULL);
    sigaction(SIGTERM, &sa, NULL);
}

#endif
