/**
 * ax_string_ops.c — Real C implementations for AXIOM string operations.
 */

#include "ax_stdlib.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

ax_i64 ax_str_len(ax_string s) {
    return (ax_i64)s.len;
}

ax_i64 ax_str_char_count(ax_string s) {
    ax_i64 count = 0;
    ax_u64 i = 0;
    while (i < s.len) {
        ax_u8 byte = s.ptr[i];
        if (byte < 0x80)      i += 1;
        else if (byte < 0xE0) i += 2;
        else if (byte < 0xF0) i += 3;
        else                  i += 4;
        count++;
    }
    return count;
}

ax_string ax_str_concat(ax_string a, ax_string b) {
    ax_u64 total = a.len + b.len;
    ax_u8* buf = (ax_u8*)malloc(total);
    if (!buf) ax_panic("out of memory in string concat");
    memcpy(buf, a.ptr, a.len);
    memcpy(buf + a.len, b.ptr, b.len);
    return (ax_string){ .ptr = buf, .len = total };
}

ax_string ax_str_slice(ax_string s, ax_i64 start, ax_i64 end) {
    if (start < 0) start = 0;
    if (end > (ax_i64)s.len) end = (ax_i64)s.len;
    if (start >= end) return (ax_string){ .ptr = s.ptr, .len = 0 };
    return (ax_string){ .ptr = s.ptr + start, .len = (ax_u64)(end - start) };
}

ax_bool ax_str_contains(ax_string s, ax_string sub) {
    return ax_str_index_of(s, sub) >= 0 ? AX_TRUE : AX_FALSE;
}

ax_bool ax_str_starts_with(ax_string s, ax_string prefix) {
    if (prefix.len > s.len) return AX_FALSE;
    return memcmp(s.ptr, prefix.ptr, prefix.len) == 0 ? AX_TRUE : AX_FALSE;
}

ax_bool ax_str_ends_with(ax_string s, ax_string suffix) {
    if (suffix.len > s.len) return AX_FALSE;
    return memcmp(s.ptr + s.len - suffix.len, suffix.ptr, suffix.len) == 0
           ? AX_TRUE : AX_FALSE;
}

ax_i64 ax_str_index_of(ax_string s, ax_string sub) {
    if (sub.len == 0) return 0;
    if (sub.len > s.len) return -1;
    ax_u64 limit = s.len - sub.len;
    for (ax_u64 i = 0; i <= limit; i++) {
        if (memcmp(s.ptr + i, sub.ptr, sub.len) == 0) return (ax_i64)i;
    }
    return -1;
}

ax_string ax_str_trim(ax_string s) {
    ax_u64 start = 0;
    while (start < s.len && isspace(s.ptr[start])) start++;
    ax_u64 end = s.len;
    while (end > start && isspace(s.ptr[end - 1])) end--;
    return (ax_string){ .ptr = s.ptr + start, .len = end - start };
}

ax_bool ax_str_eq(ax_string a, ax_string b) {
    if (a.len != b.len) return AX_FALSE;
    if (a.len == 0) return AX_TRUE;
    return memcmp(a.ptr, b.ptr, a.len) == 0 ? AX_TRUE : AX_FALSE;
}

ax_string ax_i64_to_str(ax_i64 value) {
    char buf[24];
    int n = snprintf(buf, sizeof(buf), "%lld", (long long)value);
    if (n < 0) n = 0;
    ax_u8* result = (ax_u8*)malloc((size_t)n);
    if (!result) ax_panic("out of memory in i64_to_str");
    memcpy(result, buf, (size_t)n);
    return (ax_string){ .ptr = result, .len = (ax_u64)n };
}

ax_string ax_f64_to_str(ax_f64 value) {
    char buf[64];
    int n = snprintf(buf, sizeof(buf), "%.6g", value);
    if (n < 0) n = 0;
    ax_u8* result = (ax_u8*)malloc((size_t)n);
    if (!result) ax_panic("out of memory in f64_to_str");
    memcpy(result, buf, (size_t)n);
    return (ax_string){ .ptr = result, .len = (ax_u64)n };
}

ax_string ax_bool_to_str(ax_bool value) {
    return value ? AX_STR("true") : AX_STR("false");
}
