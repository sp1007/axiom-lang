# p16-t05: std.math — Math Functions

## Purpose
Implement the AXIOM standard math library providing trigonometric, exponential, logarithmic, and utility math functions for both `f32` and `f64` types.

## Context
`std.math` wraps C libm functions via `extern` declarations, providing type-safe AXIOM interfaces. It also provides integer math utilities (min, max, abs, clamp, pow_int) and compile-time constants (PI, E, etc.).

## Inputs
- C libm functions via `extern` linkage
- AXIOM type system from p04

## Outputs
- `stdlib/math/math.ax` — math function library
- `stdlib/math/constants.ax` — mathematical constants

## Dependencies
- p04-t02: type-table — f32, f64 types
- p08 or p11: C or native backend for extern linkage

## Detailed Requirements

```axiom
# stdlib/math/constants.ax
const PI:    f64 = 3.141592653589793
const E:     f64 = 2.718281828459045
const TAU:   f64 = 6.283185307179586
const SQRT2: f64 = 1.4142135623730951
const LN2:   f64 = 0.6931471805599453
const INF:   f64 = 1.0 / 0.0
const NAN:   f64 = 0.0 / 0.0

# stdlib/math/math.ax
extern fn sin(x: f64) -> f64
extern fn cos(x: f64) -> f64
extern fn tan(x: f64) -> f64
extern fn asin(x: f64) -> f64
extern fn acos(x: f64) -> f64
extern fn atan(x: f64) -> f64
extern fn atan2(y: f64, x: f64) -> f64
extern fn exp(x: f64) -> f64
extern fn log(x: f64) -> f64
extern fn log2(x: f64) -> f64
extern fn log10(x: f64) -> f64
extern fn sqrt(x: f64) -> f64
extern fn cbrt(x: f64) -> f64
extern fn pow(base: f64, exp: f64) -> f64
extern fn floor(x: f64) -> f64
extern fn ceil(x: f64) -> f64
extern fn round(x: f64) -> f64
extern fn abs(x: f64) -> f64
extern fn fmod(x: f64, y: f64) -> f64

# Integer utilities (AXIOM-implemented)
fn min[T: Ord](a: T, b: T) -> T
fn max[T: Ord](a: T, b: T) -> T
fn clamp[T: Ord](val: T, lo: T, hi: T) -> T
fn abs_i32(n: i32) -> i32
fn pow_i32(base: i32, exp: u32) -> i32
fn gcd(a: u64, b: u64) -> u64
fn lcm(a: u64, b: u64) -> u64
fn is_nan(x: f64) -> bool
fn is_inf(x: f64) -> bool
fn is_finite(x: f64) -> bool
```

f32 variants:
```axiom
extern fn sinf(x: f32) -> f32
extern fn cosf(x: f32) -> f32
extern fn sqrtf(x: f32) -> f32
# ... etc
```

Link: `-lm` (on Linux/macOS).

## Implementation Steps

1. Create `stdlib/math/constants.ax` — compile-time constant definitions.
2. Create `stdlib/math/math.ax` — extern declarations wrapping libm.
3. Implement integer math utilities in pure AXIOM.
4. Implement `is_nan`, `is_inf`, `is_finite` using bit patterns.
5. Write tests comparing results to known values.

## Test Plan
- `TestSin`: sin(PI/2) ≈ 1.0 (within f64 epsilon)
- `TestSqrt`: sqrt(4.0) = 2.0 exactly
- `TestPow`: pow(2.0, 10.0) = 1024.0
- `TestMin`: min(3, 7) = 3
- `TestGCD`: gcd(12, 8) = 4
- `TestIsNan`: is_nan(NAN) = true; is_nan(1.0) = false

## Validation Checklist
- [ ] All libm functions linked at compile time
- [ ] f32 and f64 variants both available
- [ ] Integer overflow in pow_i32 documented (not checked for perf)
- [ ] Constants correct to full f64 precision

## Acceptance Criteria
- `std.math.sin(std.math.PI)` returns value within 1e-15 of 0.0

## Definition of Done
- [ ] `stdlib/math/math.ax` implemented
- [ ] All tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| libm not linked on some targets | Add -lm to linker flags for Linux/macOS; built-in on Windows |

## Future Follow-up Tasks
- SIMD-accelerated math (vectorized sin/cos via AVX2)
- Fixed-point math for embedded targets
