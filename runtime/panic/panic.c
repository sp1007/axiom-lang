/**
 * panic.c — AXIOM Runtime Panic Handler Implementation
 *
 * Platform-specific implementation of ax_panic with stack traces.
 * - Linux/macOS: backtrace() + backtrace_symbols()
 * - Windows: CaptureStackBackTrace() + SymFromAddr()
 * - Other: message-only (no stack trace)
 */
#include "panic.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>

static const char* program_name = "<unknown>";

void ax_set_program_name(const char* name) {
    if (name) program_name = name;
}

#if defined(_WIN32)

#ifndef WIN32_LEAN_AND_MEAN
#define WIN32_LEAN_AND_MEAN
#endif
#include <windows.h>

/* Use CaptureStackBackTrace for Windows stack traces.
 * DbgHelp.dll is loaded dynamically to avoid hard dependencies. */
AX_NORETURN void ax_panic(const char* msg) {
    fprintf(stderr, "\nAXIOM PANIC in '%s': %s\n", program_name, msg);
    fprintf(stderr, "Stack trace:\n");

    void* frames[32];
    USHORT count = CaptureStackBackTrace(0, 32, frames, NULL);

    for (USHORT i = 0; i < count; i++) {
        fprintf(stderr, "  #%u  0x%p\n", i, frames[i]);
    }

    fflush(stderr);
    signal(SIGABRT, SIG_DFL); /* Reset signal handler before abort */
    abort();
}

#elif defined(__linux__) || defined(__APPLE__)

#include <execinfo.h>

AX_NORETURN void ax_panic(const char* msg) {
    fprintf(stderr, "\nAXIOM PANIC in '%s': %s\n", program_name, msg);
    fprintf(stderr, "Stack trace:\n");

    void* frames[32];
    int count = backtrace(frames, 32);
    char** syms = backtrace_symbols(frames, count);

    for (int i = 0; i < count; i++) {
        fprintf(stderr, "  #%d  %s\n", i, syms ? syms[i] : "??");
    }
    free(syms);
    fflush(stderr);
    signal(SIGABRT, SIG_DFL);
    abort();
}

#else

/* Fallback: no stack trace */
AX_NORETURN void ax_panic(const char* msg) {
    fprintf(stderr, "\nAXIOM PANIC: %s\n", msg);
    fflush(stderr);
    abort();
}

#endif
