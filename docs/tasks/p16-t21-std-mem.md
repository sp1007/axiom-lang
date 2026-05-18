# p16-t21: `std.mem` (Arena, size_of, align_of)

## Purpose
Implement the standard library memory utilities module providing Arena allocators, `size_of[T]()`, `align_of[T]()`, `byte_swap()`, and other low-level memory operations. This module is the AXIOM-level interface to the runtime memory system.

## Context
Plan §Phase 9 lists: _"`std/mem.ax` — Arena, addr, size_of, align_of, byte_swap"_. The `in [Arena]` block syntax requires a stdlib Arena type. `size_of` and `align_of` are compiler built-ins exposed through the stdlib.

## Inputs
- Runtime AxAlloc API from p07/p14
- Arena block syntax from p06-t07
- Compiler built-ins for `size_of`, `align_of`

## Outputs
- `std/mem.ax` — Arena type, memory utilities
- Tests

## Dependencies
- p16-t01: std-testing-assert — test framework
- p08-t08: cgen-unsafe-arena — Arena C-backend codegen
- p06-t07: arena-block-handling — Arena semantics

## Subsystems Affected
- Standard library: mem module
- Arena blocks: `in [Arena]` uses `std.mem.Arena`

## Detailed Requirements

### API

```axiom
// Compiler built-in intrinsics exposed as stdlib functions
pub fn size_of[T]() -> usize     // compile-time constant
pub fn align_of[T]() -> usize    // compile-time constant

// Arena allocator
pub struct Arena:
    fn new(capacity: usize) -> Arena
    fn alloc[T](mut self) -> lent T
    fn alloc_array[T](mut self, count: usize) -> lent Seq[T]
    fn reset(mut self)       // reset bump pointer, keep memory
    fn destroy(mut self)     // free all memory

// Raw memory utilities
pub fn byte_swap[T](value: T) -> T   // endian swap
pub fn zeroed[T]() -> T              // zero-initialized value
pub fn copy(dst: *mut u8, src: *u8, len: usize)  // memcpy
pub fn set(dst: *mut u8, val: u8, len: usize)     // memset
```

### Arena Implementation Notes
- Uses bump-pointer allocation within a fixed-size buffer
- O(1) allocation (increment pointer, check bounds)
- O(1) reset (reset pointer to start)
- O(1) destroy (single free call)
- No generational IDs within arena (per plan §Phase 3)
- `lent` references: cannot escape the `in [Arena]` block (enforced by type checker)

### `size_of` and `align_of` Implementation
These are compiler intrinsics, not runtime functions. The type checker (p05) resolves them to constant values based on TypeTable layout information. The C-backend emits `sizeof()` and `_Alignof()`.

## Implementation Steps

1. Create `std/mem.ax`.
2. Implement `size_of[T]` and `align_of[T]` as compiler built-in intrinsics.
3. Implement `Arena` struct with bump-pointer allocation.
4. Implement `byte_swap` using bitwise operations.
5. Implement `zeroed[T]`, `copy`, `set` as thin wrappers around runtime calls.
6. Write tests.

## Test Plan

- `TestSizeOfI32`: `size_of[i32]()` == 4
- `TestSizeOfStruct`: `size_of[MyStruct]()` matches expected
- `TestAlignOfF64`: `align_of[f64]()` == 8
- `TestArenaAlloc`: allocate 100 objects → no heap allocations
- `TestArenaReset`: reset then reallocate → same memory reused
- `TestArenaOverflow`: allocate beyond capacity → panic
- `TestByteSwap`: `byte_swap(0x01020304_u32)` == `0x04030201`
- `TestZeroed`: `zeroed[i32]()` == 0

## Acceptance Criteria

- Arena-based programs use zero heap allocations for arena-scoped objects
- `size_of` and `align_of` evaluate at compile time

## Definition of Done

- [ ] `std/mem.ax` implemented
- [ ] Tests pass
- [ ] Arena integrates with `in [Arena]` block syntax

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Arena fragmentation | Bump-pointer has zero fragmentation by design |
| `size_of` for generic types | Resolve after monomorphization |

## Future Follow-up Tasks

- p14-t01: Production AxAlloc uses similar segment-based design
- p18-t06: Runtime self-hosting uses std.mem for bootstrap memory management
