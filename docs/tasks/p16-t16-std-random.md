# p16-t16: std.random — Random Number Generation

## Purpose
Implement cryptographically-secure and fast pseudo-random number generation for AXIOM programs, supporting both deterministic seeded RNG (for testing) and OS-entropy-based secure RNG.

## Context
`std.random` provides two RNG variants: `Rng` (fast, Xoshiro256** PRNG — deterministic from seed) and `SecureRng` (wraps OS `/dev/urandom` or `getrandom()` — non-deterministic). All collection shuffle operations use `Rng`.

## Inputs
- OS entropy: `/dev/urandom` (Linux/macOS), `BCryptGenRandom` (Windows), `getrandom(2)` syscall
- Seed API for reproducible testing

## Outputs
- `stdlib/random/rng.ax` — Rng (Xoshiro256**) + SecureRng

## Dependencies
- p16-t03: std-collections — `Array.shuffle()` uses Rng
- p16-t10: std-time — timestamp-based default seed

## Detailed Requirements

```axiom
# stdlib/random/rng.ax

# Fast PRNG: Xoshiro256** (non-cryptographic)
type Rng:
    var state: [u64; 4]

    fn new() -> Rng            # seeded from SystemTime
    fn with_seed(seed: u64) -> Rng
    fn next_u64(mut self) -> u64
    fn next_u32(mut self) -> u32
    fn next_f64(mut self) -> f64   # [0.0, 1.0)
    fn next_f32(mut self) -> f32
    fn next_bool(mut self) -> bool
    fn next_range_i64(mut self, lo: i64, hi: i64) -> i64   # [lo, hi)
    fn next_range_u64(mut self, lo: u64, hi: u64) -> u64

# Cryptographically secure RNG
type SecureRng:
    fn new() -> SecureRng
    fn fill_bytes(mut self, buf: []u8)
    fn next_u64(mut self) -> u64
    fn next_u32(mut self) -> u32

# Extension methods on Array
fn shuffle[T](mut arr: Array[T], rng: mut Rng)
fn choose[T](arr: Array[T], rng: mut Rng) -> Option[T]

# Convenience functions (global Rng, thread-local)
fn random_u64() -> u64
fn random_f64() -> f64
fn random_range(lo: i64, hi: i64) -> i64
fn random_bool() -> bool
```

Xoshiro256** algorithm:
```c
uint64_t xoshiro256ss_next(uint64_t s[4]) {
    const uint64_t result = rotl(s[1] * 5, 7) * 9;
    const uint64_t t = s[1] << 17;
    s[2] ^= s[0]; s[3] ^= s[1]; s[1] ^= s[2]; s[0] ^= s[3];
    s[2] ^= t; s[3] = rotl(s[3], 45);
    return result;
}
```

Thread-local global Rng: seeded from `SystemTime::now().unix_timestamp() ^ thread_id`.

SecureRng: use `getrandom(2)` syscall (Linux 3.17+) or `/dev/urandom` fallback.

## Implementation Steps

1. Create `stdlib/random/rng.ax`.
2. Implement Xoshiro256** in C shim (`runtime/rng.c`).
3. Implement `next_range` using rejection sampling (avoid modulo bias).
4. Implement SecureRng wrapping `getrandom()`.
5. Implement thread-local global Rng.
6. Implement `Array.shuffle()` using Fisher-Yates algorithm.
7. Write statistical tests (chi-squared for uniformity).

## Test Plan
- `TestRngSeed`: same seed → same sequence
- `TestRngRange`: 10K values in [0,100) → all within range
- `TestRngUniform`: chi-squared test on 100K values → p-value > 0.05
- `TestSecureRng`: fill_bytes → no repeating patterns (basic)
- `TestShuffle`: shuffle [0..N], sort, verify all N elements present

## Validation Checklist
- [ ] Deterministic: same seed same sequence
- [ ] next_range uses rejection sampling (no modulo bias)
- [ ] SecureRng never uses PRNG state
- [ ] Thread-local Rng seeded differently per thread

## Acceptance Criteria
- `chi_squared(rng.next_f64, 1M samples)` shows uniform distribution

## Definition of Done
- [ ] `stdlib/random/rng.ax` implemented
- [ ] Uniformity test passes

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| getrandom() not available on older kernels | Fallback to /dev/urandom open + read |

## Future Follow-up Tasks
- Gaussian distribution sampling
- UUID generation using SecureRng
