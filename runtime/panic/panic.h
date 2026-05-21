/**
 * panic.h — AXIOM Runtime Panic Handler
 *
 * The panic handler is the terminal error reporting mechanism.
 * It prints a diagnostic message to stderr with a stack trace
 * (where available), then calls abort(). It is called by the
 * allocator, generational reference checker, bounds checking,
 * and user-visible assert calls.
 */
#pragma once

#include <stddef.h>
#include <stdio.h>

/* Portability macro for no-return functions */
#if defined(_MSC_VER)
  #define AX_NORETURN __declspec(noreturn)
#else
  #define AX_NORETURN __attribute__((noreturn))
#endif

/* Portability macro for branch prediction hints */
#if defined(__GNUC__) || defined(__clang__)
  #define AX_UNLIKELY(x) __builtin_expect(!!(x), 0)
#else
  #define AX_UNLIKELY(x) (x)
#endif

/**
 * ax_set_program_name — Register the program name for panic messages.
 * Call once from ax_main before anything else.
 */
void ax_set_program_name(const char* name);

/**
 * ax_panic — Print msg to stderr with stack trace, then abort().
 * Never returns.
 */
AX_NORETURN void ax_panic(const char* msg);

/**
 * ax_bounds_check — Panic if idx >= len.
 * Used for array bounds checks in generated code.
 */
static inline void ax_bounds_check(size_t idx, size_t len) {
    if (AX_UNLIKELY(idx >= len)) {
        char buf[128];
        snprintf(buf, sizeof(buf),
                 "index out of bounds: index %zu, length %zu", idx, len);
        ax_panic(buf);
    }
}

/**
 * ax_assert — Panic if cond is false.
 * Used for runtime assertions in generated code.
 */
static inline void ax_assert(int cond, const char* msg) {
    if (AX_UNLIKELY(!cond))
        ax_panic(msg);
}
