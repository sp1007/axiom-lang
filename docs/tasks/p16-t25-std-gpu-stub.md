# p16-t25: `std.gpu` — Stub (CPU Fallback)

## Purpose
Implement the `std.gpu` stub module providing a CPU-based fallback for GPU dispatch operations. This allows code using `dispatch` and GPU annotations to compile and run on any machine without a GPU. Per plan §12.7: _"std.gpu — CPU fallback [Stub]"_.

## Context
Plan §Phase 9 lists: _"`std/gpu.ax` — stub (CPU fallback)"_. AIR includes `dispatch <kernel>, %grid, %block` opcode. The stub executes kernels sequentially on CPU, enabling development and testing without GPU hardware. Real GPU backends (CUDA/ROCm/Metal) are Phase 10+ [Future].

## Inputs
- AIR `dispatch` opcode from p09-t01
- Kernel function signatures

## Outputs
- `std/gpu.ax` — stub GPU module with CPU fallback
- Tests

## Dependencies
- p16-t01: std-testing-assert — test framework
- p09-t01: air-instruction-set — `dispatch` opcode definition

## Detailed Requirements

### API Surface

```axiom
pub struct Grid:
    x: u32
    y: u32
    z: u32

pub struct Block:
    x: u32
    y: u32
    z: u32

pub struct ThreadIdx:
    x: u32
    y: u32
    z: u32

// CPU fallback: executes kernel sequentially for each thread in grid
pub fn dispatch(kernel: fn(ThreadIdx), grid: Grid, block: Block):
    for gz in 0..grid.z:
        for gy in 0..grid.y:
            for gx in 0..grid.x:
                for bz in 0..block.z:
                    for by in 0..block.y:
                        for bx in 0..block.x:
                            let idx = ThreadIdx {
                                x: gx * block.x + bx,
                                y: gy * block.y + by,
                                z: gz * block.z + bz,
                            }
                            kernel(idx)

// GPU memory stubs (just heap allocations on CPU)
pub fn alloc_device[T](count: usize) -> *mut T:
    return std.mem.alloc[T](count)

pub fn copy_to_device[T](dst: *mut T, src: *T, count: usize):
    std.mem.copy(dst as *mut u8, src as *u8, count * std.mem.size_of[T]())

pub fn copy_from_device[T](dst: *mut T, src: *T, count: usize):
    std.mem.copy(dst as *mut u8, src as *u8, count * std.mem.size_of[T]())

pub fn free_device[T](ptr: *mut T):
    std.mem.free(ptr)
```

### C-Backend Mapping
- AIR `dispatch` → nested for-loops calling kernel function
- `alloc_device` → `malloc`
- `copy_to_device` / `copy_from_device` → `memcpy`

## Implementation Steps

1. Create `std/gpu.ax` with Grid/Block/ThreadIdx types.
2. Implement `dispatch()` as nested CPU loops.
3. Implement device memory stubs (heap wrappers).
4. Write tests.

## Test Plan

- `TestDispatch1D`: dispatch 256 threads → all indices visited
- `TestDispatch2D`: dispatch 16x16 grid → 256 threads executed
- `TestDeviceMemory`: alloc → copy_to → copy_from → values match
- `TestKernelOutput`: vector add kernel → correct results

## Acceptance Criteria

- `dispatch` executes all grid×block threads sequentially
- Results match expected kernel output
- No GPU hardware required

## Definition of Done

- [ ] `std/gpu.ax` implemented
- [ ] Tests pass on CPU-only machines
- [ ] AIR `dispatch` opcode maps to stub

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| CPU fallback extremely slow for large grids | Document as stub; warn if grid > 1M threads |
| API divergence from future real GPU backend | Design API around CUDA/Metal common subset |
