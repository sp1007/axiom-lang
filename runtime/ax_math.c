/**
 * ax_math.c — Real C implementations for AXIOM math functions.
 */

#include "ax_stdlib.h"
#include <math.h>
#include <stdlib.h>

ax_i64 ax_abs_i64(ax_i64 x) {
    return x < 0 ? -x : x;
}

ax_i32 ax_abs_i32(ax_i32 x) {
    return x < 0 ? -x : x;
}

ax_i64 ax_min_i64(ax_i64 a, ax_i64 b) {
    return a <= b ? a : b;
}

ax_i64 ax_max_i64(ax_i64 a, ax_i64 b) {
    return a >= b ? a : b;
}

ax_f64 ax_min_f64(ax_f64 a, ax_f64 b) {
    return a <= b ? a : b;
}

ax_f64 ax_max_f64(ax_f64 a, ax_f64 b) {
    return a >= b ? a : b;
}

ax_i64 ax_clamp_i64(ax_i64 x, ax_i64 lo, ax_i64 hi) {
    if (x < lo) return lo;
    if (x > hi) return hi;
    return x;
}

ax_i64 ax_pow_i64(ax_i64 base, ax_i64 exp) {
    ax_i64 result = 1;
    ax_i64 b = base;
    ax_i64 e = exp;
    while (e > 0) {
        if (e & 1) {
            result *= b;
        }
        b *= b;
        e >>= 1;
    }
    return result;
}

ax_i64 ax_gcd(ax_i64 a, ax_i64 b) {
    ax_i64 x = ax_abs_i64(a);
    ax_i64 y = ax_abs_i64(b);
    while (y != 0) {
        ax_i64 t = y;
        y = x % y;
        x = t;
    }
    return x;
}

ax_i64 ax_lcm(ax_i64 a, ax_i64 b) {
    if (a == 0 || b == 0) return 0;
    return ax_abs_i64(a) / ax_gcd(a, b) * ax_abs_i64(b);
}

ax_f64 ax_pow(ax_f64 base, ax_f64 exp) {
    return pow(base, exp);
}

ax_bool ax_sum_layout_is_pointer(void) {
#ifdef AXIOM_SUMLAYOUT_POINTER
    return AX_TRUE;
#else
    return AX_FALSE;
#endif
}
