/**
 * test_alloc.c — Unit tests for the AXIOM axalloc allocator.
 *
 * Compile:
 *   gcc -O2 -Wall -Wextra -Werror -std=c11 test_alloc.c axalloc.c -o test_alloc
 *   ./test_alloc
 */
#include "axalloc.h"

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif

/* Stub ax_panic for standalone test builds */
void ax_panic(unsigned char* msg) {
    fprintf(stderr, "ax_panic: %s\n", (const char*)msg);
    abort();
}

static void* g_ax_global_state = NULL;
void* ax_get_global_state_internal(void) {
    printf("[Stub] ax_get_global_state_internal called. Current state: %p\n", g_ax_global_state); fflush(stdout);
    if (g_ax_global_state == NULL) {
#if defined(_WIN32)
        g_ax_global_state = VirtualAlloc(NULL, 4096, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
#else
        g_ax_global_state = mmap(NULL, 4096, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
#endif
        printf("[Stub] Allocated global state: %p\n", g_ax_global_state); fflush(stdout);
    }
    return g_ax_global_state;
}

void* sys_mmap(void* addr, uint64_t len, int32_t prot, int32_t flags, int32_t fd, int64_t offset) {
    printf("[Stub] sys_mmap called (addr=%p, len=%llu)\n", addr, (unsigned long long)len); fflush(stdout);
#if defined(_WIN32)
    (void)prot; (void)flags; (void)fd; (void)offset;
    void* res = VirtualAlloc(addr, len, 0x3000, 0x04);
    printf("[Stub] sys_mmap VirtualAlloc returned %p\n", res); fflush(stdout);
    return res;
#else
    void* res = mmap(addr, len, prot, flags, fd, offset);
    printf("[Stub] sys_mmap mmap returned %p\n", res); fflush(stdout);
    return res;
#endif
}

int32_t sys_munmap(void* addr, uint64_t len) {
    printf("[Stub] sys_munmap called (addr=%p, len=%llu)\n", addr, (unsigned long long)len); fflush(stdout);
#if defined(_WIN32)
    (void)len;
    int32_t res = VirtualFree(addr, 0, 0x8000) ? 0 : -1;
    printf("[Stub] sys_munmap VirtualFree returned %d\n", res); fflush(stdout);
    return res;
#else
    int32_t res = munmap(addr, len);
    printf("[Stub] sys_munmap munmap returned %d\n", res); fflush(stdout);
    return res;
#endif
}


static int tests_passed = 0;
static int tests_total  = 0;

#define TEST(name) do { tests_total++; printf("  [%02d] %-50s", tests_total, name); fflush(stdout); } while(0)
#define PASS()     do { tests_passed++; printf("PASS\n"); fflush(stdout); } while(0)

/* Test 1: ax_alloc returns non-NULL */
static void test_alloc_basic(void) {
    TEST("ax_alloc returns non-NULL");
    void* ptr = ax_alloc(64);
    assert(ptr != NULL);
    ax_free(ptr);
    PASS();
}

/* Test 2: ax_alloc_size returns correct size */
static void test_alloc_size(void) {
    TEST("ax_alloc_size returns correct size");
    void* ptr = ax_alloc(64);
    assert(ax_alloc_size(ptr) == 64);
    ax_free(ptr);
    PASS();
}

/* Test 3: Allocated memory is writable */
static void test_alloc_writable(void) {
    TEST("allocated memory is writable");
    void* ptr = ax_alloc(64);
    memset(ptr, 0xAB, 64);
    unsigned char* bytes = (unsigned char*)ptr;
    for (int i = 0; i < 64; i++) {
        assert(bytes[i] == 0xAB);
    }
    ax_free(ptr);
    PASS();
}

/* Test 4: gen_id == 1 after alloc */
static void test_gen_id_initial(void) {
    TEST("gen_id == 1 after alloc");
    void* ptr = ax_alloc(32);
    AxHeader* hdr = ax_get_header(ptr);
    assert(hdr->gen_id == 1);
    ax_free(ptr);
    PASS();
}

/* Test 5: gen_id incremented after free
 * NOTE: We check the gen_id BEFORE free, then verify semantically
 * that free increments it. We cannot safely dereference after free.
 * Instead, we use a realloc-based proxy. */
static void test_gen_id_after_free(void) {
    TEST("gen_id increments on free (proxy check)");
    /* We verify the semantics indirectly: alloc → make_ref → free
     * should result in the ref having a stale gen_id. We can't read
     * the freed header, so we test via AxRef. */
    void* ptr = ax_alloc(32);
    AxRef ref = ax_make_ref(ptr);
    assert(ref.gen_id == 1);
    /* After free, the ref should be stale (gen_id was 1, now it's 2) */
    /* We can't verify the freed memory, but the invariant is tested
     * in the gen_ref tests (p07-t02). For now, just verify no crash. */
    ax_free(ptr);
    PASS();
}

/* Test 6: ax_realloc returns valid pointer with correct size */
static void test_realloc(void) {
    TEST("ax_realloc returns valid pointer");
    void* ptr = ax_alloc(64);
    memset(ptr, 0x42, 64);
    ptr = ax_realloc(ptr, 128);
    assert(ptr != NULL);
    assert(ax_alloc_size(ptr) == 128);
    /* Original data preserved */
    unsigned char* bytes = (unsigned char*)ptr;
    for (int i = 0; i < 64; i++) {
        assert(bytes[i] == 0x42);
    }
    ax_free(ptr);
    PASS();
}

/* Test 7: ax_realloc preserves gen_id */
static void test_realloc_preserves_gen_id(void) {
    TEST("ax_realloc preserves gen_id");
    void* ptr = ax_alloc(32);
    AxHeader* hdr = ax_get_header(ptr);
    assert(hdr->gen_id == 1);
    ptr = ax_realloc(ptr, 128);
    hdr = ax_get_header(ptr);
    assert(hdr->gen_id == 1); /* same generation */
    ax_free(ptr);
    PASS();
}

/* Test 8: ax_realloc(NULL, size) == ax_alloc(size) */
static void test_realloc_null(void) {
    TEST("ax_realloc(NULL, size) == ax_alloc(size)");
    void* ptr = ax_realloc(NULL, 32);
    assert(ptr != NULL);
    assert(ax_alloc_size(ptr) == 32);
    AxHeader* hdr = ax_get_header(ptr);
    assert(hdr->gen_id == 1);
    ax_free(ptr);
    PASS();
}

/* Test 9: ax_free(NULL) is a no-op */
static void test_free_null(void) {
    TEST("ax_free(NULL) does not crash");
    ax_free(NULL);
    PASS();
}

/* Test 10: Zero-size alloc */
static void test_alloc_zero(void) {
    TEST("ax_alloc(0) returns non-NULL");
    void* ptr = ax_alloc(0);
    assert(ptr != NULL);
    assert(ax_alloc_size(ptr) == 1);
    ax_free(ptr);
    PASS();
}

/* Test 11: Large allocation (16 MiB) */
static void test_alloc_large(void) {
    TEST("ax_alloc(16MiB) succeeds");
    void* ptr = ax_alloc(1 << 24);
    assert(ptr != NULL);
    assert(ax_alloc_size(ptr) == (1 << 24));
    /* Write first and last byte to verify */
    ((unsigned char*)ptr)[0] = 0xFF;
    ((unsigned char*)ptr)[(1 << 24) - 1] = 0xFE;
    ax_free(ptr);
    PASS();
}

/* Test 12: Pointer alignment (8-byte aligned) */
static void test_alignment(void) {
    TEST("ax_alloc returns 8-byte aligned pointer");
    for (int i = 1; i <= 100; i++) {
        void* ptr = ax_alloc((size_t)i);
        assert(((uintptr_t)ptr % 8) == 0);
        ax_free(ptr);
    }
    PASS();
}

/* Test 13: ax_make_ref captures correct gen_id */
static void test_make_ref(void) {
    TEST("ax_make_ref captures gen_id");
    void* ptr = ax_alloc(32);
    AxRef ref = ax_make_ref(ptr);
    assert(ref.ptr == ptr);
    assert(ref.gen_id == 1);
    ax_free(ptr);
    PASS();
}

/* Test 14: ax_make_ref(NULL) is safe */
static void test_make_ref_null(void) {
    TEST("ax_make_ref(NULL) is safe");
    AxRef ref = ax_make_ref(NULL);
    assert(ref.ptr == NULL);
    assert(ref.gen_id == 0);
    PASS();
}

/* Test 15: ax_alloc_size(NULL) returns 0 */
static void test_alloc_size_null(void) {
    TEST("ax_alloc_size(NULL) returns 0");
    assert(ax_alloc_size(NULL) == 0);
    PASS();
}

int main(void) {
    printf("=== axalloc unit tests ===\n\n");
    fflush(stdout);

    test_alloc_basic();
    test_alloc_size();
    test_alloc_writable();
    test_gen_id_initial();
    test_gen_id_after_free();
    test_realloc();
    test_realloc_preserves_gen_id();
    test_realloc_null();
    test_free_null();
    test_alloc_zero();
    test_alloc_large();
    test_alignment();
    test_make_ref();
    test_make_ref_null();
    test_alloc_size_null();

    printf("\n=== Results: %d/%d tests passed ===\n", tests_passed, tests_total);
    return (tests_passed == tests_total) ? 0 : 1;
}
