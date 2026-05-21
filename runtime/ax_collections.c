/**
 * ax_collections.c — Real C implementations for AXIOM Vec and Arena.
 */

#include "ax_stdlib.h"
#include <stdlib.h>
#include <string.h>

/* ================================================================
 * Vec — Dynamic Array
 * ================================================================ */

ax_vec ax_vec_new(ax_i64 elem_size) {
    ax_vec v;
    v.data = NULL;
    v.len = 0;
    v.cap = 0;
    v.elem_size = elem_size;
    return v;
}

ax_vec ax_vec_with_capacity(ax_i64 elem_size, ax_i64 cap) {
    ax_vec v;
    v.elem_size = elem_size;
    v.len = 0;
    v.cap = cap;
    if (cap > 0) {
        v.data = malloc((size_t)(cap * elem_size));
        if (!v.data) ax_panic("out of memory in Vec::with_capacity");
    } else {
        v.data = NULL;
    }
    return v;
}

static void ax_vec_grow(ax_vec* v) {
    ax_i64 new_cap = v->cap == 0 ? 8 : v->cap * 2;
    void* new_data = realloc(v->data, (size_t)(new_cap * v->elem_size));
    if (!new_data) ax_panic("out of memory in Vec grow");
    v->data = new_data;
    v->cap = new_cap;
}

void ax_vec_push(ax_vec* v, const void* elem) {
    if (v->len == v->cap) ax_vec_grow(v);
    memcpy((char*)v->data + v->len * v->elem_size, elem, (size_t)v->elem_size);
    v->len++;
}

ax_bool ax_vec_pop(ax_vec* v, void* out_elem) {
    if (v->len == 0) return AX_FALSE;
    v->len--;
    if (out_elem) {
        memcpy(out_elem, (char*)v->data + v->len * v->elem_size,
               (size_t)v->elem_size);
    }
    return AX_TRUE;
}

void* ax_vec_get(ax_vec* v, ax_i64 index) {
    ax_bounds_check((size_t)index, (size_t)v->len);
    return (char*)v->data + index * v->elem_size;
}

void ax_vec_set(ax_vec* v, ax_i64 index, const void* elem) {
    ax_bounds_check((size_t)index, (size_t)v->len);
    memcpy((char*)v->data + index * v->elem_size, elem, (size_t)v->elem_size);
}

void ax_vec_clear(ax_vec* v) {
    v->len = 0;
}

void ax_vec_free(ax_vec* v) {
    free(v->data);
    v->data = NULL;
    v->len = 0;
    v->cap = 0;
}

/* ================================================================
 * Arena — Bump-Pointer Allocator
 * ================================================================ */

ax_arena ax_arena_new(ax_i64 capacity) {
    ax_arena a;
    a.base = (ax_u8*)malloc((size_t)capacity);
    if (!a.base) ax_panic("out of memory in Arena::new");
    a.bump = a.base;
    a.limit = a.base + capacity;
    a.size = capacity;
    return a;
}

static ax_u8* align_up_ptr(ax_u8* ptr, ax_i64 align) {
    uintptr_t addr = (uintptr_t)ptr;
    uintptr_t aligned = (addr + (uintptr_t)(align - 1)) & ~((uintptr_t)(align - 1));
    return (ax_u8*)aligned;
}

void* ax_arena_alloc(ax_arena* a, ax_i64 size, ax_i64 align) {
    ax_u8* aligned = align_up_ptr(a->bump, align);
    if (aligned + size > a->limit) ax_panic("arena out of memory");
    a->bump = aligned + size;
    return aligned;
}

void ax_arena_reset(ax_arena* a) {
    a->bump = a->base;
}

void ax_arena_destroy(ax_arena* a) {
    free(a->base);
    a->base = NULL;
    a->bump = NULL;
    a->limit = NULL;
    a->size = 0;
}

ax_i64 ax_arena_remaining(const ax_arena* a) {
    return (ax_i64)(a->limit - a->bump);
}

ax_i64 ax_arena_used(const ax_arena* a) {
    return (ax_i64)(a->bump - a->base);
}
