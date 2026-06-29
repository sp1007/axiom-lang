# AXIOM Math Library — Coverage Tracker (36 categories)

Authoritative checklist for the scientific-computing math library the user
requested (~300 functions across 36 categories). This file is the single source
of truth for scope + status so it is never lost; every batch updates it.

Legend: ✅ implemented · 🔶 partial (gaps listed) · ⬜ todo (BUG-immune, shippable) ·
🔒 blocked (needs a backend keystone or hardware).

Build/verify rule: library modules do NOT affect the self-host fixpoint (the
compiler does not import `std.*`). Ship pattern per module: self-contained repro
→ build O0+O1 with `bin/axc_stage1.exe` (exit must match) → `std/<m>.ax` +
`bin/t_<m>.ax` + regression row → commit/push. No fixpoint verify needed.

Blockers that gate whole categories:
- **BUG#45** — struct >16 bytes containing f64 miscompiles (Vec3=24B OK, but
  4×f64=32B fails). Blocks Mat2/3/4, Quaternion, Vec4. Backend keystone.
- **BUG#46** — float call-result read across a following loop. Work around by
  inlining; avoid holding a float call result across a loop.
- **closures / function pointers** — not implemented (#38.7). Blocks generic
  numerical-analysis (root-find/integrate of an arbitrary f); array-sampled
  variants are still doable.
- **SIMD / true intrinsics** — need hardware vector ABI; aspirational.

---

## Status by category

| # | Category | Status | Module(s) | Gaps / notes |
|---|----------|--------|-----------|--------------|
| 1 | Constants | ✅ | math | PI/TAU/HALF_PI/E/SQRT2/LN2/LN10/LOG2E/LOG10E/SQRT1_2/EPSILON. todo: PHI |
| 2 | Arithmetic | ✅ | math | abs/sign/min/max/clamp/square/cube/recip |
| 3 | Rounding | ✅ | math | floor/ceil/round/trunc/fract/round_to |
| 4 | Integer | ✅ | math,xmath | abs_i64/min/max/clamp_i64/gcd/lcm/pow_i64/isqrt |
| 5 | Power | ✅ | math | pow/square/cube/cbrt/exp2/exp10/pow_i64 |
| 6 | Exp | ✅ | math | exp/exp2/exp10/expm1 |
| 7 | Log | ✅ | math | ln/log2/log10/log1p/log_base |
| 8 | Trig | ✅ | math | sin/cos/tan/asin/acos/atan/atan2. todo: sec/csc/cot |
| 9 | Hyperbolic | ✅ | math | sinh/cosh/tanh/asinh/acosh/atanh |
| 10 | Angle | ✅ | math | deg_to_rad/rad_to_deg/normalize_angle |
| 11 | FP-utils | ✅ | math | is_nan/is_inf/is_finite/approx_equal/fract/copysign |
| 12 | Remainder | 🔶 | math | fmod. todo: ieee_remainder |
| 13 | Comparison | ✅ | math | min/max/clamp/approx_equal/saturate |
| 14 | Random | ✅ | rng | working xorshift64: next_u64/next_f64/range_i64/next_range_f64/gaussian (oracle-verified). random.ax=aspirational |
| 15 | Statistics | ✅ | stats | sum/mean/variance/stddev/min/max/median/percentile/covariance/correlation/skewness/kurtosis (oracle-verified) |
| 16 | Combinatorics | ✅ | combinatorics | factorial/perm/binom/catalan/multichoose/double_factorial |
| 17 | Number theory | ✅ | numtheory | gcd/lcm/is_prime/next_prime/mod_exp/totient/divisor*/is_perfect/is_coprime/mod_inverse/int_log/sum_digits |
| 18 | Bit math | ✅ | math | popcount/clz/ctz/bit_width/rotate_left/rotate_right/reverse_bits/bit_floor/bit_ceil/parity/has_single_bit/mul_hi |
| 19 | Interpolation | ✅ | interpolation,math | lerp/unlerp/remap/smoothstep/smootherstep/catmull/bezier3/bilerp/easing |
| 20 | Geometry | ✅ | geometry,vec | 2D dist/cross/dot/triangle/orient/point-in-tri/circle/polygon-area/segment-dist + Vec3 (3D in vec) |
| 21 | Matrix | 🔒 | — | needs Mat2/3/4 (≥32B f64) → **BUG#45**. Workaround: flat `ptr[f64]`+dims |
| 22 | Complex | ✅ | complex | add/sub/mul/div/conj/abs/arg/exp/ln/sqrt/from_polar |
| 23 | Quaternion | 🔒 | — | 4×f64=32B → **BUG#45** |
| 24 | Special functions | ✅ | math | erf/erfc/gamma/lgamma/beta/factorial_real/digamma/bessel_j0 (oracle-verified). todo: zeta |
| 25 | Numerical analysis | 🔶 | numerical | sampled trapezoid/simpson/central+forward diff (oracle-verified). generic root-find/quadrature need closures 🔒 |
| 26 | Signal processing | 🔶 | signal | window fns hann/hamming/blackman/bartlett/welch (oracle-verified). todo: FFT (complex arrays) |
| 27 | Machine learning | ✅ | math,ml | sigmoid/relu/leaky_relu/softplus/swish/tanh (math) + gelu/mse/mae/bce/softmax (ml, oracle-verified) |
| 28 | Probability | ✅ | probability | normal/logistic/exp pdf+cdf |
| 29 | Coordinate systems | ✅ | coordinates | polar/cylindrical/spherical ↔ cartesian (per-component, scalar) |
| 30 | Color | ✅ | color | RGB↔HSV, RGB↔HSL, luminance, lerp (ptr out-param) |
| 31 | Utility | ✅ | math | clamp/saturate/approx_equal/sign/step |
| 32 | SIMD | 🔒 | arch (stub) | hardware vector ABI |
| 33 | Big-int / decimal | 🔶 | bignum | u/i/f dynamic to 4096+. todo: decimal helpers |
| 34 | Crypto math | ✅ | numtheory | mod_exp/mod_inverse/gcd + Miller-Rabin is_probable_prime (det. < 2^32) |
| 35 | Intrinsics | ✅ | math | popcount/clz/ctz/parity/bit_* as fns (true HW intrinsics = aspirational) |
| 36 | Vector types | 🔶 | vec | Vec2/Vec3. Vec4 (32B) → **BUG#45** |

---

## Remaining BUG-immune work queue (drives the autonomous batch)

1. ~~coordinates (#29)~~ ✅ done (std.coordinates).
2. ~~color (#30)~~ 🔶 done (std.color rgb↔hsv+luminance; HSL todo).
3. ~~stats extensions (#15)~~ ✅ done (median/percentile/covariance/correlation/skew/kurtosis, oracle-verified; correlation uses helper-calls to stay robust under -O1 / BUG#48).
4. ~~signal windows (#26)~~ 🔶 done (hann/hamming/blackman/bartlett/welch; FFT todo).
5. ~~ml extensions (#27)~~ ✅ done (gelu/mse/mae/bce/softmax, oracle-verified).
6. ~~math fill-ins~~ ✅ done (expm1/log_base/normalize_angle/copysign/round_to).
7. ~~bit math fill-ins (#18)~~ ✅ done (rotate_right/reverse_bits/bit_floor/bit_ceil/parity).
8. ~~special fns (#24)~~ ✅ done (factorial_real/digamma/bessel_j0, oracle-verified vs scipy).
9. ~~crypto (#34)~~ ✅ done (Miller-Rabin nt_is_probable_prime; detects Carmichael).
10. ~~numerical analysis (#25)~~ 🔶 done (sampled trapezoid/simpson/diff; generic-f forms need closures 🔒).

## Blocked until backend keystone
- **BUG#45 fix** unlocks: Matrix (#21), Quaternion (#23), Vec4 (#36). Needs x86
  struct-ABI work for >16B f64 + ~3h fixpoint verify.
- **closures (#38.7)** unlock generic numerical analysis / function-of-f.
- **SIMD/HW intrinsics** (#32/#35) aspirational.
