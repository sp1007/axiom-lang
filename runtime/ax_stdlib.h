/**
 * ax_stdlib.h — AXIOM Standard Library C Runtime
 *
 * Real C implementations for all @compiler_intrinsic functions.
 * These are linked into every compiled AXIOM program.
 *
 * Uses ax_panic(const char*) from panic.h for error handling.
 */
#pragma once

#include "ax_runtime.h"
#include <math.h>

#ifdef __cplusplus
extern "C" {
#endif

/* ================================================================
 * String Operations (std::string)
 * ================================================================ */

#ifdef AX_NATIVE_LINK
ax_string* ax_str_concat(ax_string* a, ax_string* b);
ax_string* ax_str_slice(ax_string* s, ax_i64 start, ax_i64 end);
ax_string* ax_str_trim(ax_string* s);
ax_string* ax_str_replace(ax_string* s, ax_string* old, ax_string* new_val);
ax_string* ax_i64_to_str(ax_i64 value);
ax_string* ax_f64_to_str(ax_f64 value);
ax_string* ax_bool_to_str(ax_bool value);
#else
ax_string ax_str_concat(ax_string a, ax_string b);
ax_string ax_str_slice(ax_string s, ax_i64 start, ax_i64 end);
ax_string ax_str_trim(ax_string s);
ax_string ax_str_replace(ax_string s, ax_string old, ax_string new_val);
ax_string ax_i64_to_str(ax_i64 value);
ax_string ax_f64_to_str(ax_f64 value);
ax_string ax_bool_to_str(ax_bool value);
#endif

ax_i64    ax_str_len(ax_string s);
ax_i64    ax_str_char_count(ax_string s);
ax_bool   ax_str_contains(ax_string s, ax_string sub);
ax_bool   ax_str_starts_with(ax_string s, ax_string prefix);
ax_bool   ax_str_ends_with(ax_string s, ax_string suffix);
ax_i64    ax_str_index_of(ax_string s, ax_string sub);
ax_bool   ax_str_eq(ax_string a, ax_string b);
ax_string ax_str_to_upper(ax_string s);
ax_string ax_str_to_lower(ax_string s);
ax_string ax_str_repeat(ax_string s, ax_i64 count);
ax_bool   ax_str_parse_i64(const char* s, ax_i64* out_val);
ax_bool   ax_str_parse_f64(const char* s, ax_f64* out_val);
ax_bool   ax_str_is_valid_utf8(ax_string s);
void*     ax_str_split(ax_string s, ax_string sep);
ax_u8* ax_string_get_char_ptr(void* s);
ax_u8* ax_string_get_ptr(ax_string s);
struct ax_AxiomString;
ax_u8* ax_string_get_ptr_val(struct ax_AxiomString s);

/* ================================================================
 * Print / Format (std::fmt) — type-dispatched by codegen
 * ================================================================ */

void ax_print_str(ax_string s);
void ax_println_str(ax_string s);
void ax_print_str_native(const char* ptr);
void ax_println_str_native(const char* ptr);
void ax_print_i64(ax_i64 value);
void ax_println_i64(ax_i64 value);
void ax_print_f64(ax_f64 value);
void ax_println_f64(ax_f64 value);
void ax_print_bool(ax_bool value);
void ax_println_bool(ax_bool value);
void ax_eprint_str(ax_string s);
void ax_eprintln_str(ax_string s);

/* ================================================================
 * Assertions (std::testing)
 * Use ax_assert_axiom to avoid conflict with panic.h's ax_assert
 * ================================================================ */

void ax_assert_axiom(ax_bool condition, ax_string message);
void ax_assert_eq_i64(ax_i64 actual, ax_i64 expected);
void ax_assert_eq_str(ax_string actual, ax_string expected);
void ax_assert_eq_bool(ax_bool actual, ax_bool expected);

/* ================================================================
 * Math (std::math)
 * ================================================================ */

ax_i64 ax_abs_i64(ax_i64 x);
ax_i32 ax_abs_i32(ax_i32 x);
ax_i64 ax_min_i64(ax_i64 a, ax_i64 b);
ax_i64 ax_max_i64(ax_i64 a, ax_i64 b);
ax_f64 ax_min_f64(ax_f64 a, ax_f64 b);
ax_f64 ax_max_f64(ax_f64 a, ax_f64 b);
ax_i64 ax_clamp_i64(ax_i64 x, ax_i64 lo, ax_i64 hi);
ax_i64 ax_pow_i64(ax_i64 base, ax_i64 exp);
ax_i64 ax_gcd(ax_i64 a, ax_i64 b);
ax_i64 ax_lcm(ax_i64 a, ax_i64 b);
ax_f64 ax_pow(ax_f64 base, ax_f64 exp);

/* ================================================================
 * Vec[T] — Dynamic Array (generic via void*)
 * ================================================================ */

typedef struct {
    void*   data;
    ax_i64  len;
    ax_i64  cap;
    ax_i64  elem_size;
} ax_vec;

ax_vec  ax_vec_new(ax_i64 elem_size);
ax_vec  ax_vec_with_capacity(ax_i64 elem_size, ax_i64 cap);
void    ax_vec_push(ax_vec* v, const void* elem);
ax_bool ax_vec_pop(ax_vec* v, void* out_elem);
void*   ax_vec_get(ax_vec* v, ax_i64 index);
void    ax_vec_set(ax_vec* v, ax_i64 index, const void* elem);
void    ax_vec_clear(ax_vec* v);
void    ax_vec_free(ax_vec* v);

/* ================================================================
 * Arena Allocator (std::mem)
 * ================================================================ */

typedef struct {
    ax_u8*  base;
    ax_u8*  bump;
    ax_u8*  limit;
    ax_i64  size;
} ax_arena;

ax_arena ax_arena_new(ax_i64 capacity);
void*    ax_arena_alloc(ax_arena* a, ax_i64 size, ax_i64 align);
void     ax_arena_reset(ax_arena* a);
void     ax_arena_destroy(ax_arena* a);
ax_i64   ax_arena_remaining(const ax_arena* a);
ax_i64   ax_arena_used(const ax_arena* a);

#ifdef __cplusplus
}
#endif
