/**
 * ax_string_ops.c — Real C implementations for AXIOM string operations.
 */

#include "ax_stdlib.h"
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
    ax_u8* buf = (ax_u8*)ax_alloc(total + 1);
    if (!buf) ax_panic("out of memory in string concat");
    memcpy(buf, a.ptr, a.len);
    memcpy(buf + a.len, b.ptr, b.len);
    buf[total] = 0;
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
    ax_i64 val = value;
    int is_neg = 0;
    if (val < 0) {
        is_neg = 1;
        val = -val;
    }
    int idx = 0;
    do {
        buf[idx++] = '0' + (val % 10);
        val /= 10;
    } while (val > 0 && idx < 24);
    if (is_neg && idx < 24) {
        buf[idx++] = '-';
    }
    // Reverse
    for (int i = 0; i < idx / 2; ++i) {
        char temp = buf[i];
        buf[i] = buf[idx - 1 - i];
        buf[idx - 1 - i] = temp;
    }
    
    ax_u8* result = (ax_u8*)ax_alloc((size_t)idx);
    if (!result) ax_panic("out of memory in i64_to_str");
    memcpy(result, buf, (size_t)idx);
    return (ax_string){ .ptr = result, .len = (ax_u64)idx };
}

ax_string ax_f64_to_str(ax_f64 value) {
    char buf[64];
    ax_f64 val = value;
    int is_neg = 0;
    if (val < 0) {
        is_neg = 1;
        val = -val;
    }
    
    ax_i64 integer_part = (ax_i64)val;
    ax_f64 fraction_part = val - (ax_f64)integer_part;
    
    // Convert integer part
    int idx = 0;
    ax_i64 val_int = integer_part;
    do {
        buf[idx++] = '0' + (val_int % 10);
        val_int /= 10;
    } while (val_int > 0 && idx < 64);
    if (is_neg && idx < 64) {
        buf[idx++] = '-';
    }
    // Reverse integer part
    for (int i = 0; i < idx / 2; ++i) {
        char temp = buf[i];
        buf[i] = buf[idx - 1 - i];
        buf[idx - 1 - i] = temp;
    }
    
    // Add dot
    buf[idx++] = '.';
    
    // Convert fractional part (6 digits precision)
    ax_i64 val_frac = (ax_i64)(fraction_part * 1000000.0 + 0.5);
    int frac_start = idx;
    for (int f = 0; f < 6; ++f) {
        buf[idx++] = '0' + (val_frac % 10);
        val_frac /= 10;
    }
    // Reverse fractional part
    for (int i = 0; i < 3; ++i) {
        char temp = buf[frac_start + i];
        buf[frac_start + i] = buf[idx - 1 - i];
        buf[idx - 1 - i] = temp;
    }
    
    // Trim trailing zeros in fractional part
    while (idx > frac_start && buf[idx - 1] == '0') {
        idx--;
    }
    if (idx > 0 && buf[idx - 1] == '.') {
        idx--; // remove dot if no decimals left
    }
    
    ax_u8* result = (ax_u8*)ax_alloc((size_t)idx);
    if (!result) ax_panic("out of memory in f64_to_str");
    memcpy(result, buf, (size_t)idx);
    return (ax_string){ .ptr = result, .len = (ax_u64)idx };
}

ax_string ax_bool_to_str(ax_bool value) {
    return value ? AX_STR("true") : AX_STR("false");
}

ax_string ax_str_replace(ax_string s, ax_string old, ax_string new_val) {
    if (old.len == 0) return s;
    if (s.len < old.len) return s;
    
    // Count occurrences
    ax_u64 count = 0;
    for (ax_u64 i = 0; i <= s.len - old.len; ) {
        if (memcmp(s.ptr + i, old.ptr, old.len) == 0) {
            count++;
            i += old.len;
        } else {
            i++;
        }
    }
    
    if (count == 0) return s;
    
    // Calculate new length
    ax_i64 new_len = (ax_i64)s.len + (ax_i64)count * ((ax_i64)new_val.len - (ax_i64)old.len);
    if (new_len < 0) ax_panic("negative length or overflow in string replace");
    
    ax_u8* buf = (ax_u8*)ax_alloc((size_t)new_len + 1);
    if (!buf) ax_panic("out of memory in string replace");
    
    ax_u64 dest_idx = 0;
    for (ax_u64 i = 0; i < s.len; ) {
        if (i <= s.len - old.len && memcmp(s.ptr + i, old.ptr, old.len) == 0) {
            if (new_val.len > 0) {
                memcpy(buf + dest_idx, new_val.ptr, new_val.len);
                dest_idx += new_val.len;
            }
            i += old.len;
        } else {
            buf[dest_idx++] = s.ptr[i++];
        }
    }
    buf[new_len] = 0;
    
    return (ax_string){ .ptr = buf, .len = (ax_u64)new_len };
}

