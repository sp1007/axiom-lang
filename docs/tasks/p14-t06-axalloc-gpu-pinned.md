# p14-t06: GPU-Pinned Memory Allocator

## Purpose
Implement a GPU-pinned memory allocator variant that allocates host memory accessible by both CPU and GPU, enabling zero-copy data transfer for AXIOM GPU compute workloads.

## Context
GPU compute (CUDA/OpenCL/Vulkan) requires pinned (page-locked) host memory for efficient DMA transfers. AxAlloc's segment model is well-suited for this: each 64KB segment can be pinned as a unit, and the size-class structure minimizes wasted pinned memory.

## Inputs
- `AxAlloc` base from p14-t01 through p14-t03
- CUDA: `cudaMallocHost()` / `cudaFreeHost()`
- OpenCL: `clCreateBuffer(CL_MEM_ALLOC_HOST_PTR)`
- Vulkan: VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT

## Outputs
- `runtime/axalloc_gpu.c` — GPU-pinned segment allocator
- `ax_gpu_alloc(size)` / `ax_gpu_free(ptr)` for GPU-accessible allocations

## Dependencies
- p14-t01: axalloc-size-classes — reuse size class structure
- p14-t02: axalloc-segment-manager — override segment allocation hook

## Subsystems Affected
- GPU compute API (future phase): uses pinned allocator for transfer buffers
- Standard `ax_alloc` unaffected — GPU pinned is opt-in

## Detailed Requirements

```c
typedef enum GPUBackend {
    GPU_BACKEND_NONE   = 0,
    GPU_BACKEND_CUDA   = 1,
    GPU_BACKEND_OPENCL = 2,
    GPU_BACKEND_VULKAN = 3,
} GPUBackend;

typedef struct AxGPUAlloc {
    AxAlloc      base;
    GPUBackend   backend;
    void*        device;     // CUdevice, cl_context, or VkDevice
} AxGPUAlloc;

// Initialize GPU-pinned allocator
int ax_gpu_alloc_init(AxGPUAlloc* alloc, GPUBackend backend, void* device);

// Allocate pinned host memory
void* ax_gpu_alloc(AxGPUAlloc* alloc, size_t size);

// Free pinned host memory
void ax_gpu_free(AxGPUAlloc* alloc, void* ptr);

// Destroy GPU allocator (unpin all segments)
void ax_gpu_alloc_destroy(AxGPUAlloc* alloc);
```

Segment allocation override:
- CUDA: `cudaMallocHost(&seg, 64*1024)` instead of `mmap`
- OpenCL: `clCreateBuffer(ctx, CL_MEM_ALLOC_HOST_PTR, 64*1024, ...)`
- Vulkan: `vkAllocateMemory` with host-visible flags

Fallback: if no GPU backend, use regular `mmap` (pinning is best-effort).

Compile-time guards:
```c
#ifdef AX_GPU_CUDA
#include <cuda_runtime.h>
#endif
#ifdef AX_GPU_OPENCL
#include <CL/cl.h>
#endif
```

## Implementation Steps

1. Create `runtime/axalloc_gpu.c`.
2. Define `GPUBackend` enum and `AxGPUAlloc` struct.
3. Implement segment allocation hooks per backend.
4. Override `ax_alloc_segment()` and `ax_free_segment()` in base allocator.
5. Implement `ax_gpu_alloc()` / `ax_gpu_free()` using AxAlloc size classes.
6. Build with conditional compilation for each backend.
7. Test with CUDA mock (no real GPU required for unit tests).

## Test Plan
- `TestGPUAllocInit`: initialize with CUDA backend → no error
- `TestGPUAllocRoundtrip`: alloc + free, no leak
- `TestGPUAllocPinned`: allocated memory has cudaPointerAttributes pinned flag
- `TestGPUAllocFallback`: no GPU → falls back to regular mmap

## Validation Checklist
- [ ] Segments pinned via cudaMallocHost or equivalent
- [ ] All pinned segments unpinned on ax_gpu_alloc_destroy
- [ ] Compile without CUDA/OpenCL/Vulkan (all guards work)

## Acceptance Criteria
- AXIOM GPU program can write to ax_gpu_alloc buffer; GPU reads same data

## Definition of Done
- [ ] `runtime/axalloc_gpu.c` implemented
- [ ] Builds with all backends disabled (no GPU headers required)

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Pinned memory is a scarce resource (limited by GPU) | Cap total pinned allocation; fall back to regular alloc when cap exceeded |
| GPU API not available at build time | All GPU code behind compile-time flags |

## Future Follow-up Tasks
- AXIOM GPU compute language extensions (phase post-18)
- Unified memory (cudaMallocManaged) variant
