/**
 * genref.h — AXIOM Generational Reference Checking
 *
 * Provides use-after-free detection by comparing the generation ID
 * stored in an AxRef against the generation ID in the allocation's
 * AxHeader. When they differ, the object has been freed since the
 * reference was captured, and a panic is triggered immediately.
 *
 * All functions are static inline for zero call overhead.
 */
#pragma once

#include "axalloc.h"

/* Forward declaration of panic handler */
extern void ax_panic(const char* msg);

/* Portability macro for branch prediction hints */
#ifndef AX_UNLIKELY
  #if defined(__GNUC__) || defined(__clang__)
    #define AX_UNLIKELY(x) __builtin_expect(!!(x), 0)
  #else
    #define AX_UNLIKELY(x) (x)
  #endif
#endif

/**
 * AX_NULL_REF — the null AxRef constant.
 */
#define AX_NULL_REF ((AxRef){.ptr = NULL, .gen_id = 0})

/**
 * ax_deref — Validate and dereference an AxRef.
 * Panics on NULL or generation mismatch (use-after-free).
 * Returns the raw void* on success.
 *
 * This is the hot-path for every heap pointer dereference in generated code.
 */
static inline void* ax_deref(AxRef ref) {
    if (AX_UNLIKELY(ref.ptr == NULL))
        ax_panic("null pointer dereference");
    AxHeader* h = ((AxHeader*)ref.ptr) - 1;
    if (AX_UNLIKELY(h->gen_id != ref.gen_id))
        ax_panic("GenerationalID mismatch: use-after-free detected");
    return ref.ptr;
}

/**
 * ax_invalidate — Explicitly invalidate a pointer without freeing.
 * Sets gen_id to 0, making all existing AxRef values stale.
 * ax_invalidate(NULL) is a no-op.
 */
static inline void ax_invalidate(void* ptr) {
    if (!ptr) return;
    AxHeader* h = ((AxHeader*)ptr) - 1;
    h->gen_id = 0; /* 0 is the explicit invalidation sentinel */
}

/**
 * ax_ref_valid — Non-panicking validity check.
 * Returns 1 if the reference is still valid, 0 otherwise.
 * Used by debug assertions and optional checking.
 */
static inline int ax_ref_valid(AxRef ref) {
    if (ref.ptr == NULL) return 0;
    AxHeader* h = ((AxHeader*)ref.ptr) - 1;
    return h->gen_id == ref.gen_id;
}
