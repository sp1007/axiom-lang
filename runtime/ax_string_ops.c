/**
 * ax_string_ops.c — Real C implementations for AXIOM string operations.
 */

#include "ax_stdlib.h"
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

#ifdef AX_NATIVE_LINK
#define ax_str_concat ax_str_concat_real
#define ax_str_slice ax_str_slice_real
#define ax_str_trim ax_str_trim_real
#define ax_i64_to_str ax_i64_to_str_real
#define ax_f64_to_str ax_f64_to_str_real
#define ax_bool_to_str ax_bool_to_str_real
#define ax_str_replace ax_str_replace_real
#endif

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

ax_u8* ax_string_get_char_ptr(void* s) {
    return (ax_u8*)((ax_string*)s)->ptr;
}

ax_u8* ax_string_get_ptr(ax_string s) {
    return (ax_u8*)s.ptr;
}

struct ax_AxiomString {
    ax_u8* ptr;
    ax_i64 len;
};

ax_u8* ax_string_get_ptr_val(struct ax_AxiomString s) {
    return (ax_u8*)s.ptr;
}

#ifdef AX_NATIVE_LINK
#undef ax_str_concat
#undef ax_str_slice
#undef ax_str_trim
#undef ax_i64_to_str
#undef ax_f64_to_str
#undef ax_bool_to_str
#undef ax_str_replace
static ax_string* get_ring_slot(void) {
    ax_string* slot = (ax_string*)ax_alloc(sizeof(ax_string));
    if (!slot) ax_panic("out of memory in get_ring_slot");
    return slot;
}

ax_string* ax_str_concat(ax_string* a, ax_string* b) {
    ax_string* slot = get_ring_slot();
    *slot = ax_str_concat_real(*a, *b);
    return slot;
}

ax_string* ax_str_slice(ax_string* s, ax_i64 start, ax_i64 end) {
    ax_string* slot = get_ring_slot();
    *slot = ax_str_slice_real(*s, start, end);
    return slot;
}

ax_string* ax_str_trim(ax_string* s) {
    ax_string* slot = get_ring_slot();
    *slot = ax_str_trim_real(*s);
    return slot;
}

ax_string* ax_i64_to_str(ax_i64 value) {
    ax_string* slot = get_ring_slot();
    *slot = ax_i64_to_str_real(value);
    return slot;
}

ax_string* ax_f64_to_str(ax_f64 value) {
    ax_string* slot = get_ring_slot();
    *slot = ax_f64_to_str_real(value);
    return slot;
}

ax_string* ax_bool_to_str(ax_bool value) {
    ax_string* slot = get_ring_slot();
    *slot = ax_bool_to_str_real(value);
    return slot;
}

ax_string* ax_str_replace(ax_string* s, ax_string* old, ax_string* new_val) {
    ax_string* slot = get_ring_slot();
    *slot = ax_str_replace_real(*s, *old, *new_val);
    return slot;
}
#endif


/* Compatibility wrappers for standard library mangled names */
__attribute__((weak)) ax_bool ax_std_string_starts_with(ax_string s, ax_string prefix) {
    return ax_str_starts_with(s, prefix);
}

__attribute__((weak)) ax_string ax_std_string_replace(ax_string s, ax_string old, ax_string new_val) {
#ifdef AX_NATIVE_LINK
    return ax_str_replace_real(s, old, new_val);
#else
    return ax_str_replace(s, old, new_val);
#endif
}

__attribute__((weak)) ax_i64 ax_std_string_len(ax_string s) {
    return ax_str_len(s);
}

__attribute__((weak)) ax_string ax_std_string_concat(ax_string a, ax_string b) {
#ifdef AX_NATIVE_LINK
    return ax_str_concat_real(a, b);
#else
    return ax_str_concat(a, b);
#endif
}

__attribute__((weak)) ax_string ax_std_string_slice(ax_string s, ax_i64 start, ax_i64 end) {
#ifdef AX_NATIVE_LINK
    return ax_str_slice_real(s, start, end);
#else
    return ax_str_slice(s, start, end);
#endif
}

__attribute__((weak)) ax_string ax_std_string_trim(ax_string s) {
#ifdef AX_NATIVE_LINK
    return ax_str_trim_real(s);
#else
    return ax_str_trim(s);
#endif
}



ax_string ax_str_to_upper(ax_string s) {
    ax_u8* buf = s.len > 0 ? (ax_u8*)ax_alloc(s.len + 1) : NULL;
    if (buf) {
        for (ax_u64 i = 0; i < s.len; i++) {
            ax_u8 c = s.ptr[i];
            if (c >= 'a' && c <= 'z') {
                buf[i] = c - 'a' + 'A';
            } else {
                buf[i] = c;
            }
        }
        buf[s.len] = '\0';
    }
    ax_string res = { .ptr = buf ? buf : (const ax_u8*)"", .len = s.len };
    return res;
}

ax_string ax_str_to_lower(ax_string s) {
    ax_u8* buf = s.len > 0 ? (ax_u8*)ax_alloc(s.len + 1) : NULL;
    if (buf) {
        for (ax_u64 i = 0; i < s.len; i++) {
            ax_u8 c = s.ptr[i];
            if (c >= 'A' && c <= 'Z') {
                buf[i] = c - 'A' + 'a';
            } else {
                buf[i] = c;
            }
        }
        buf[s.len] = '\0';
    }
    ax_string res = { .ptr = buf ? buf : (const ax_u8*)"", .len = s.len };
    return res;
}

ax_string ax_str_repeat(ax_string s, ax_i64 count) {
    if (count < 0) count = 0;
    ax_u64 total_len = s.len * count;
    ax_u8* buf = total_len > 0 ? (ax_u8*)ax_alloc(total_len + 1) : NULL;
    if (buf) {
        for (ax_i64 i = 0; i < count; i++) {
            memcpy(buf + i * s.len, s.ptr, (size_t)s.len);
        }
        buf[total_len] = '\0';
    }
    ax_string res = { .ptr = buf ? buf : (const ax_u8*)"", .len = total_len };
    return res;
}

ax_bool ax_str_parse_i64(const char* s, ax_i64* out_val) {
    if (!s) return AX_FALSE;
    char* endptr;
    ax_i64 val = strtoll(s, &endptr, 10);
    if (endptr != s) {
        *out_val = val;
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_bool ax_str_parse_f64(const char* s, ax_f64* out_val) {
    if (!s) return AX_FALSE;
    char* endptr;
    ax_f64 val = strtod(s, &endptr);
    if (endptr != s) {
        *out_val = val;
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_bool ax_str_is_valid_utf8(ax_string s) {
    ax_bool valid = AX_TRUE;
    ax_u64 i = 0;
    while (i < s.len) {
        ax_u8 b = s.ptr[i];
        if (b <= 0x7F) {
            i++;
        } else if ((b & 0xE0) == 0xC0) {
            if (i + 1 >= s.len || (s.ptr[i+1] & 0xC0) != 0x80) {
                valid = AX_FALSE;
                break;
            }
            i += 2;
        } else if ((b & 0xF0) == 0xE0) {
            if (i + 2 >= s.len || (s.ptr[i+1] & 0xC0) != 0x80 || (s.ptr[i+2] & 0xC0) != 0x80) {
                valid = AX_FALSE;
                break;
            }
            i += 3;
        } else if ((b & 0xF8) == 0xF0) {
            if (i + 3 >= s.len || (s.ptr[i+1] & 0xC0) != 0x80 || (s.ptr[i+2] & 0xC0) != 0x80 || (s.ptr[i+3] & 0xC0) != 0x80) {
                valid = AX_FALSE;
                break;
            }
            i += 4;
        } else {
            valid = AX_FALSE;
            break;
        }
    }
    return valid;
}

void* ax_str_split(ax_string s, ax_string sep) {
    ax_vec* v = (ax_vec*)ax_alloc(sizeof(ax_vec));
    *v = ax_vec_new(sizeof(ax_string));
    if (sep.len == 0) {
        for (ax_u64 i = 0; i < s.len; i++) {
            ax_string sub = { .ptr = s.ptr + i, .len = 1 };
            ax_vec_push(v, &sub);
        }
    } else {
        ax_u64 last = 0;
        for (ax_u64 i = 0; i <= s.len - sep.len; ) {
            ax_bool match = AX_TRUE;
            for (ax_u64 j = 0; j < sep.len; j++) {
                if (s.ptr[i+j] != sep.ptr[j]) {
                    match = AX_FALSE;
                    break;
                }
            }
            if (match) {
                ax_string sub = { .ptr = s.ptr + last, .len = i - last };
                ax_vec_push(v, &sub);
                i += sep.len;
                last = i;
            } else {
                i++;
            }
        }
        if (last <= s.len) {
            ax_string sub = { .ptr = s.ptr + last, .len = s.len - last };
            ax_vec_push(v, &sub);
        }
    }
    return v;
}

__attribute__((weak)) ax_string ax_std_string_to_upper(ax_string s) {
    return ax_str_to_upper(s);
}

__attribute__((weak)) ax_string ax_std_string_to_lower(ax_string s) {
    return ax_str_to_lower(s);
}

__attribute__((weak)) ax_string ax_std_string_repeat(ax_string s, ax_i64 count) {
    return ax_str_repeat(s, count);
}


__attribute__((weak)) ax_bool ax_std_string_is_valid_utf8(ax_string s) {
    return ax_str_is_valid_utf8(s);
}

__attribute__((weak)) void* ax_std_string_split(ax_string s, ax_string sep) {
    return ax_str_split(s, sep);
}


