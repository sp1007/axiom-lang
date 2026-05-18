# p07-t01: Implement MVP Allocator (axalloc)

## Purpose
Implement the MVP allocator in `runtime/axalloc/axalloc.c` and `runtime/axalloc/axalloc.h`. This allocator wraps `malloc`/`free` but prepends an 8-byte `AxHeader{gen_id:uint64}` to every allocation, enabling the generational reference safety system used throughout the AXIOM runtime.

## Context
Every heap allocation in AXIOM carries a generation counter in its header. When a pointer is freed, the generation counter is incremented. Any held reference (`AxRef`) that still carries the old generation value will fail the validity check in `ax_deref`, catching use-after-free at runtime without a garbage collector. This is the foundational memory safety mechanism for AXIOM's Compile-Time GC (CTGC) system.

The C-Backend generates `#include "ax_runtime.h"` at the top of every emitted `.c` file. The `ax_runtime.h` header in turn includes `axalloc.h`, so the allocator is always available in generated code.

## Inputs
- C11 compiler (GCC or Clang) available in the build environment
- `runtime/` directory exists in the repository root
- Standard C library (`stdlib.h`, `stdint.h`, `stddef.h`)
- No external dependencies

## Outputs
- `runtime/axalloc/axalloc.h` — public header with type definitions and inline helpers
- `runtime/axalloc/axalloc.c` — implementation of the allocator functions
- `runtime/axalloc/axalloc.go` — Go wrapper with `//go:build ignore` for FFI testing
- `runtime/axalloc/Makefile` — build rules to compile the C code and run tests

## Dependencies
- p01-t01 (repository structure must exist so the `runtime/` directory is present)

## Subsystems Affected
- Runtime (primary)
- C-Backend (consumes the header)
- Generational reference system (p07-t02 builds on top of this)
- Panic handler (p07-t03 is called by the allocator on invalid states)

## Detailed Requirements

### Data Structures
```c
// AxHeader: prepended to every heap allocation (8 bytes, 8-byte aligned)
typedef struct {
    uint64_t gen_id;  // generation counter; 1 = live, 0 = invalid, >1 = freed N times
} AxHeader;

// AxRef: a fat pointer holding both the data pointer and the generation at capture time
typedef struct {
    void*    ptr;    // points to the byte immediately AFTER the AxHeader
    uint64_t gen_id; // generation at the time this reference was created
} AxRef;
```

### API
```c
// Allocate `size` bytes. Internally allocates sizeof(AxHeader)+size, sets gen_id=1,
// returns pointer to byte after the header.
void* ax_alloc(size_t size);

// Free a pointer previously returned by ax_alloc.
// Increments header->gen_id before calling free(), invalidating live AxRef values.
void  ax_free(void* ptr);

// Reallocate. Preserves the existing gen_id. Returns new pointer.
void* ax_realloc(void* ptr, size_t new_size);

// Return the user-visible size of the allocation (does NOT include the header).
// Requires that ptr was returned by ax_alloc / ax_realloc.
size_t ax_alloc_size(void* ptr);
```

### Invariants
- `ax_alloc` never returns NULL; on OOM it calls `ax_panic("out of memory")` and aborts.
- `ax_free(NULL)` is a no-op (matches POSIX `free` semantics).
- `ax_realloc(NULL, size)` is equivalent to `ax_alloc(size)`.
- The header is always 8-byte aligned; user data starts at an 8-byte-aligned address.
- `gen_id` starts at 1. After one free, it becomes 2. It must never be 0 for a live allocation (0 is the sentinel for "explicitly invalidated" — see p07-t02).

### Size Tracking
The header must also track the user allocation size so that `ax_alloc_size` works without platform-specific extensions (`malloc_usable_size` is not portable). Extend `AxHeader` to include a `size` field:

```c
typedef struct {
    uint64_t gen_id;
    uint64_t size;   // user-requested size in bytes
} AxHeader;          // 16 bytes total
```

Update all code and documentation to reflect the 16-byte header. The frozen spec says 8 bytes but this phase establishes the concrete C implementation; the extra 8-byte size field is an implementation detail not visible to AXIOM programs.

### Compilation Requirements
- Must compile cleanly with `gcc -O2 -Wall -Wextra -Werror -std=c11`
- Must compile cleanly with `clang -O2 -Wall -Wextra -Werror -std=c11`
- No compiler warnings permitted

## Implementation Steps

### Step 1: Create directory structure
```
runtime/
  axalloc/
    axalloc.h
    axalloc.c
    axalloc.go
    test_alloc.c
    Makefile
```

### Step 2: Write `axalloc.h`
```c
#pragma once
#include <stddef.h>
#include <stdint.h>

// AxHeader: 16 bytes prepended to every heap allocation.
typedef struct {
    uint64_t gen_id;  // generation counter; 1=live, incremented on free
    uint64_t size;    // user allocation size (not including header)
} AxHeader;

// AxRef: fat pointer with captured generation ID.
typedef struct {
    void*    ptr;
    uint64_t gen_id;
} AxRef;

// Core allocator API
void*  ax_alloc(size_t size);
void   ax_free(void* ptr);
void*  ax_realloc(void* ptr, size_t new_size);
size_t ax_alloc_size(void* ptr);

// Inline helper: get header from user pointer
static inline AxHeader* ax_get_header(void* ptr) {
    return ((AxHeader*)ptr) - 1;
}
```

### Step 3: Write `axalloc.c`
```c
#include "axalloc.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Declared in panic.h (included transitively via ax_runtime.h in full builds)
extern void ax_panic(const char* msg);

void* ax_alloc(size_t size) {
    AxHeader* hdr = (AxHeader*)malloc(sizeof(AxHeader) + size);
    if (!hdr) ax_panic("ax_alloc: out of memory");
    hdr->gen_id = 1;
    hdr->size   = size;
    return (void*)(hdr + 1);
}

void ax_free(void* ptr) {
    if (!ptr) return;
    AxHeader* hdr = ax_get_header(ptr);
    hdr->gen_id++;   // invalidate all live AxRef values
    free(hdr);
}

void* ax_realloc(void* ptr, size_t new_size) {
    if (!ptr) return ax_alloc(new_size);
    AxHeader* old_hdr = ax_get_header(ptr);
    uint64_t gen = old_hdr->gen_id;
    AxHeader* new_hdr = (AxHeader*)realloc(old_hdr, sizeof(AxHeader) + new_size);
    if (!new_hdr) ax_panic("ax_realloc: out of memory");
    new_hdr->gen_id = gen;   // preserve generation
    new_hdr->size   = new_size;
    return (void*)(new_hdr + 1);
}

size_t ax_alloc_size(void* ptr) {
    if (!ptr) return 0;
    return ax_get_header(ptr)->size;
}
```

### Step 4: Write `axalloc.go` (FFI stub for testing from Go)
```go
//go:build ignore

package axalloc

/*
#cgo CFLAGS: -O2 -Wall -Wextra
#include "axalloc.h"
#include "axalloc.c"

// Stub panic for standalone testing
void ax_panic(const char* msg) {
    fprintf(stderr, "ax_panic: %s\n", msg);
    abort();
}
*/
import "C"
import "unsafe"

func Alloc(size int) unsafe.Pointer {
    return C.ax_alloc(C.size_t(size))
}

func Free(ptr unsafe.Pointer) {
    C.ax_free(ptr)
}
```

### Step 5: Write `Makefile`
```makefile
CC     = gcc
CFLAGS = -O2 -Wall -Wextra -Werror -std=c11

all: test_alloc

axalloc.o: axalloc.c axalloc.h
	$(CC) $(CFLAGS) -c axalloc.c -o axalloc.o

test_alloc: test_alloc.c axalloc.o
	$(CC) $(CFLAGS) test_alloc.c axalloc.o -o test_alloc
	./test_alloc

clean:
	rm -f axalloc.o test_alloc
```

## Test Plan
- Unit tests in `runtime/axalloc/test_alloc.c`:
  1. `ax_alloc(64)` returns non-NULL pointer
  2. `ax_alloc_size(ptr) == 64` after allocating 64 bytes
  3. Write to all 64 bytes without segfault (memory is writable)
  4. `ax_get_header(ptr)->gen_id == 1` immediately after alloc
  5. `ax_free(ptr)` increments `gen_id` (read header from a copy of the pointer before freeing)
  6. `ax_realloc(ptr, 128)` returns valid pointer with `ax_alloc_size == 128`
  7. `ax_realloc(NULL, 32)` behaves as `ax_alloc(32)`
  8. `ax_free(NULL)` does not crash
  9. Zero-size alloc: `ax_alloc(0)` returns non-NULL, `ax_alloc_size == 0`
  10. Large alloc: `ax_alloc(1 << 24)` succeeds (16 MiB)

Run with AddressSanitizer: `gcc -fsanitize=address -g test_alloc.c axalloc.c -o test_alloc_asan && ./test_alloc_asan`

## Validation Checklist
- [ ] `axalloc.h` compiles standalone with `-Wall -Wextra -Werror`
- [ ] `axalloc.c` compiles with GCC and Clang without warnings
- [ ] `sizeof(AxHeader) == 16` verified via `_Static_assert`
- [ ] `ax_alloc` returns 8-byte-aligned pointer
- [ ] `gen_id == 1` immediately after alloc
- [ ] `gen_id == 2` immediately after first free
- [ ] `ax_realloc` preserves `gen_id`
- [ ] `ax_free(NULL)` is a no-op
- [ ] All unit tests pass
- [ ] AddressSanitizer reports no errors

## Acceptance Criteria
- All unit tests in `test_alloc.c` pass
- No compiler warnings with `-Wall -Wextra -Werror` on GCC and Clang
- AddressSanitizer clean
- `ax_alloc_size` returns exact user-requested size
- The `AxHeader` and `AxRef` types are correctly defined and usable from other C files via `#include "axalloc.h"`

## Definition of Done
- `runtime/axalloc/axalloc.h` exists and is correct
- `runtime/axalloc/axalloc.c` exists and implements all four functions
- `runtime/axalloc/axalloc.go` exists with `//go:build ignore` tag
- `runtime/axalloc/Makefile` exists and `make` succeeds
- All unit tests pass
- Code is reviewed against the invariants listed in Detailed Requirements

## Risks & Mitigations
- **Risk**: Platform-specific alignment requirements differ. **Mitigation**: Use `_Static_assert(sizeof(AxHeader) % 8 == 0, "AxHeader must be 8-byte aligned")` and ensure `malloc` always returns at least 8-byte-aligned memory (guaranteed by POSIX and Windows CRT).
- **Risk**: `ax_panic` is defined in another translation unit, causing linker issues during standalone testing. **Mitigation**: Provide a stub in `test_alloc.c` and use the `axalloc.go` wrapper only with the full build.
- **Risk**: `gen_id` overflow after 2^64 frees of the same pointer. **Mitigation**: This is astronomically unlikely in practice; document it as a known limitation. A `uint64_t` is sufficient for any realistic program.

## Future Follow-up Tasks
- p07-t02: Add `ax_deref`, `ax_make_ref`, `ax_invalidate` built on top of this allocator
- p07-t05: Full memory safety integration tests
- p10-t05: CTGC optimization pass uses `ax_realloc` for object reuse (`OpReuseAlloc`)
- p08-t08: Arena allocator will be a separate allocation path in `runtime/axalloc/arena.c`
