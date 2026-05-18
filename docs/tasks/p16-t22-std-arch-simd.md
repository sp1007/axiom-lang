# p16-t22: `std.arch.x86` — SIMD Intrinsics

## Purpose
Implement the platform-specific SIMD intrinsics module exposing x86-64 SSE2/AVX2/AVX-512 operations as type-safe AXIOM functions. This enables high-performance numerical code without dropping to `unsafe` or `extern "C"`.

## Context
Plan §Phase 9 lists: _"`std/arch/x86.ax` — SIMD intrinsics (`_mm256_*`)"_. The loop vectorizer (p10-t08) auto-generates SIMD via AIR `v_add`/`v_mul`, but developers also need explicit intrinsics for hand-tuned kernels.

## Inputs
- AIR SIMD opcodes from p09-t01 (`v_load`, `v_store`, `v_add`, `v_mul`, `v_fma`)
- Native x86 backend from p11 (instruction encoding)
- Loop vectorization pass from p10-t08

## Outputs
- `std/arch/x86.ax` — SSE2/AVX2/AVX-512 intrinsic wrappers
- `std/arch/detect.ax` — runtime CPU feature detection (CPUID)
- Tests

## Dependencies
- p16-t01: std-testing-assert — test framework
- p11-t02: x86-instruction-set — x86 instruction definitions
- p10-t08: opt-vectorization — SIMD AIR opcodes

## Detailed Requirements

### API Surface (AVX2 subset)

```axiom
pub type V256[T] = @simd(T, 32)  // 256-bit vector

// Load / Store
pub fn v256_load[T](ptr: *T) -> V256[T]
pub fn v256_store[T](ptr: *mut T, v: V256[T])
pub fn v256_loadu[T](ptr: *T) -> V256[T]   // unaligned

// Arithmetic (f32x8, f64x4, i32x8)
pub fn v256_add[T](a: V256[T], b: V256[T]) -> V256[T]
pub fn v256_sub[T](a: V256[T], b: V256[T]) -> V256[T]
pub fn v256_mul[T](a: V256[T], b: V256[T]) -> V256[T]
pub fn v256_fma[T](a: V256[T], b: V256[T], c: V256[T]) -> V256[T]

// Shuffle / Permute
pub fn v256_shuffle_f32(a: V256[f32], b: V256[f32], imm: u8) -> V256[f32]

// Comparison → mask
pub fn v256_cmp_gt[T](a: V256[T], b: V256[T]) -> V256[T]

// CPU feature detection
pub fn has_avx2() -> bool
pub fn has_avx512f() -> bool
```

### Implementation Strategy
- Each intrinsic maps 1:1 to an AIR SIMD opcode or a native instruction
- C-backend: emit GCC/Clang `__builtin_ia32_*` or `_mm256_*` intrinsics
- Native backend: emit raw VEX-encoded instructions
- Feature detection: CPUID instruction via inline assembly or runtime call

### Compile-Time Guards
```axiom
if #target.has_feature("avx2"):
    let result = v256_add(a, b)
else:
    let result = scalar_fallback(a, b)
```

## Implementation Steps

1. Create `std/arch/x86.ax` with type definitions.
2. Implement AVX2 arithmetic intrinsics (add, sub, mul, fma).
3. Implement load/store (aligned + unaligned).
4. Implement `std/arch/detect.ax` with CPUID-based detection.
5. Add C-backend mappings to `__builtin_ia32_*`.
6. Write tests.

## Test Plan

- `TestV256AddF32`: 8 floats added correctly
- `TestV256FmaF64`: fused multiply-add matches scalar
- `TestV256LoadStore`: round-trip load→store→load
- `TestCPUIDDetect`: `has_avx2()` returns correct value on test machine
- `TestScalarFallback`: non-AVX2 path produces identical results

## Acceptance Criteria

- SIMD intrinsics produce correct results matching scalar equivalents
- C-backend emits correct `_mm256_*` calls
- Feature detection works at runtime

## Definition of Done

- [ ] `std/arch/x86.ax` implemented with AVX2 subset
- [ ] `std/arch/detect.ax` implemented
- [ ] Tests pass on AVX2-capable hardware

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| CI machines may lack AVX2 | Feature detection guards + scalar fallback tests |
| VEX encoding complexity | Reuse ModRM/SIB library from p11-t17 |
