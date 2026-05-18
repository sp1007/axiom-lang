# p08-t08: C-Backend Unsafe Blocks and Arena Allocation

## Purpose
Implement C code generation for unsafe blocks and arena allocation in `codegen/cgen/arena.go` and extend the unsafe handling in `codegen/cgen/stmts.go`. Unsafe blocks remove safety checks from generated code; arena blocks redirect allocations to a bump allocator that can be freed in O(1) by destroying the arena.

## Context
AXIOM's arena allocation (`in [arena]: block`) is a performance-critical feature for scenarios where many objects are created and freed together. Instead of calling `ax_alloc`/`ax_free` per object, the arena allocator uses a bump pointer to allocate sequentially, then frees all objects at once when the arena is destroyed. This gives near-zero allocation overhead and zero per-object free overhead.

The arena must also be implemented in C as `runtime/axalloc/arena.c`. The C-Backend generates calls to the arena API when it encounters `in [arena]: block` syntax.

## Inputs
- `codegen/cgen/stmts.go` (p08-t03) — `UnsafeStmt` and `ArenaStmt` handling
- `codegen/cgen/gencheck.go` (p08-t06) — unsafe mode suppresses gencheck
- Typed AST arena block node
- p06-t07 (arena allocation analysis — identifies allocations inside arena blocks)

## Outputs
- `runtime/axalloc/arena.c` and `runtime/axalloc/arena.h` — the arena allocator C implementation
- `codegen/cgen/arena.go` — arena code generation helpers
- `codegen/cgen/arena_test.go` — unit tests

## Dependencies
- p08-t06 (generational check emission — arena allocations bypass per-object checks)
- p08-t05 (ownership code gen — arena blocks affect ownership of allocated objects)

## Subsystems Affected
- Runtime (new arena allocator C code)
- C-Backend (arena allocation sites in generated code)
- Memory safety (arena allocations skip per-object generational checks)

## Detailed Requirements

### Arena C API
```c
// runtime/axalloc/arena.h

typedef struct AxArena AxArena;

// Create a new arena with the given initial capacity.
// If the capacity is exceeded, the arena grows by allocating new slabs.
AxArena* ax_arena_create(size_t capacity);

// Allocate `size` bytes from the arena, aligned to 8 bytes.
// Returns NULL only if the arena cannot grow (OOM).
void* ax_arena_alloc(AxArena* arena, size_t size);

// Destroy the arena and free all memory in O(1).
void ax_arena_destroy(AxArena* arena);

// Reset the arena without freeing (reuse same memory for a new batch).
void ax_arena_reset(AxArena* arena);

// Return total bytes allocated from this arena (for profiling).
size_t ax_arena_used(AxArena* arena);
```

### Arena Implementation Strategy
The arena uses a slab list:
```c
typedef struct AxArenaSlab {
    struct AxArenaSlab* next;
    size_t capacity;
    size_t used;
    uint8_t data[];  // flexible array member
} AxArenaSlab;

struct AxArena {
    AxArenaSlab* current;  // current active slab
    AxArenaSlab* first;    // first slab (for destroy/reset)
    size_t       slab_size; // default slab size for growth
};
```

When a slab is full, allocate a new slab of `max(slab_size, requested_size + sizeof(AxArenaSlab))` and link it.

On `ax_arena_destroy`: walk the slab list, call `free()` on each slab, then `free(arena)`.

On `ax_arena_reset`: walk the slab list, set `slab->used = 0` on each, set `arena->current = arena->first`.

### Generated C for `in [arena]: block`
AXIOM:
```
let arena = Arena.new(4096)
in [arena]:
    let nodes = Node{...}
    let edges = Edge{...}
    // ... process ...
// arena is destroyed here (CTGC or explicit)
```

Generated C:
```c
AxArena* arena = ax_arena_create(4096);
{
    /* arena block */
    struct ax_Node* nodes = (struct ax_Node*)ax_arena_alloc(arena, sizeof(struct ax_Node));
    *nodes = (struct ax_Node){...};
    struct ax_Edge* edges = (struct ax_Edge*)ax_arena_alloc(arena, sizeof(struct ax_Edge));
    *edges = (struct ax_Edge){...};
    /* ... process ... */
}
ax_arena_destroy(arena);
```

Note: arena-allocated variables are plain pointers (not `AxRef`), because arena lifetime is managed collectively. The C-Backend must track that `nodes` and `edges` are arena-allocated and emit plain pointer accesses for them.

### Generated C for `unsafe: block`
AXIOM:
```
unsafe:
    let raw: *u8 = some_ptr
    raw[0] = 255
```

Generated C:
```c
{
    /* unsafe block */
    ax_u8* raw = some_ptr;
    raw[0] = 255;  // no bounds check, no ax_deref
}
```

The `/* unsafe block */` comment is required for auditability. The `ExprGen.unsafe` flag is set to `true` for the duration of the block.

### Arena Allocation Mode
In the `AllocMode` enum (p08-t06), `AllocArena` indicates the variable was allocated in an arena. The `StmtGen` must know the current arena variable name to emit `ax_arena_alloc(arena_var, ...)` calls.

Extend `StmtGen` with:
```go
type StmtGen struct {
    // ... existing fields ...
    currentArena string  // name of the active AxArena* variable, or "" if none
}
```

When entering an `in [arena]: block`, set `currentArena` to the arena's C variable name. On exit, restore to the previous value (for nested arenas).

### Unsafe Block Nesting
Unsafe blocks can be nested inside arena blocks and vice versa. The `unsafe` flag and `currentArena` are independent.

### Arena Safety Guarantees
Although arena allocations bypass per-object generational checks:
1. The arena pointer itself (`AxArena*`) is a regular heap pointer with an `AxRef`
2. Arena-allocated pointers cannot outlive the arena (enforced at compile time by the ownership checker)
3. After `ax_arena_destroy`, all arena pointers become dangling — the ownership checker must ensure they are not used

## Implementation Steps

### Step 1: Create `runtime/axalloc/arena.h` and `arena.c`
Implement the slab-based arena allocator as described above.

```c
// arena.c skeleton
#include "arena.h"
#include "axalloc.h"
#include <string.h>

static AxArenaSlab* new_slab(size_t capacity) {
    AxArenaSlab* s = (AxArenaSlab*)malloc(sizeof(AxArenaSlab) + capacity);
    if (!s) ax_panic("ax_arena: out of memory");
    s->next = NULL;
    s->capacity = capacity;
    s->used = 0;
    return s;
}

AxArena* ax_arena_create(size_t capacity) {
    AxArena* a = (AxArena*)malloc(sizeof(AxArena));
    if (!a) ax_panic("ax_arena_create: out of memory");
    a->slab_size = capacity > 0 ? capacity : 4096;
    a->first = a->current = new_slab(a->slab_size);
    return a;
}

void* ax_arena_alloc(AxArena* arena, size_t size) {
    // Align size to 8 bytes
    size = (size + 7) & ~(size_t)7;
    if (arena->current->used + size > arena->current->capacity) {
        size_t new_cap = arena->slab_size > size ? arena->slab_size : size * 2;
        AxArenaSlab* s = new_slab(new_cap);
        arena->current->next = s;
        arena->current = s;
    }
    void* ptr = arena->current->data + arena->current->used;
    arena->current->used += size;
    return ptr;
}

void ax_arena_destroy(AxArena* arena) {
    AxArenaSlab* s = arena->first;
    while (s) {
        AxArenaSlab* next = s->next;
        free(s);
        s = next;
    }
    free(arena);
}
```

### Step 2: Create `codegen/cgen/arena.go`
```go
package cgen

import "fmt"

// emitArenaBlock generates C code for an AXIOM 'in [arena]: block'
func (g *StmtGen) emitArenaBlock(s *ast.ArenaStmt) {
    arenaVar := s.ArenaVarName  // the AXIOM variable holding the arena
    prev := g.currentArena
    g.currentArena = arenaVar

    g.w.Line("{")
    g.w.Line("    /* arena block */")
    g.w.Indent()
    g.EmitBlock(s.Body)
    g.w.Dedent()
    g.w.Line("}")
    g.w.Line(fmt.Sprintf("ax_arena_destroy(%s);", arenaVar))

    g.currentArena = prev
}

// emitArenaDecl generates C code for a variable allocated in the current arena.
func (g *StmtGen) emitArenaDecl(s *ast.VarDeclStmt) {
    ctype := CTypeName(s.TypeID, g.table, g.queue)
    g.w.Line(fmt.Sprintf("%s* %s = (%s*)ax_arena_alloc(%s, sizeof(%s));",
        ctype, s.Name, ctype, g.currentArena, ctype))
    if s.Init != nil {
        g.w.Line(fmt.Sprintf("*%s = %s;", s.Name, g.exprGen.EmitExpr(s.Init)))
    }
}
```

### Step 3: Implement `emitUnsafe` in `stmts.go`
```go
func (g *StmtGen) emitUnsafe(s *ast.UnsafeStmt) {
    g.w.Line("{")
    g.w.Line("    /* unsafe block */")
    g.w.Indent()
    oldUnsafe := g.exprGen.unsafe
    g.exprGen.unsafe = true
    g.EmitBlock(s.Body)
    g.exprGen.unsafe = oldUnsafe
    g.w.Dedent()
    g.w.Line("}")
}
```

### Step 4: Write `runtime/axalloc/test_arena.c`
Test: create arena, allocate N objects, verify data, destroy arena.

### Step 5: Add arena to `runtime/axalloc/Makefile`
```makefile
arena.o: arena.c arena.h
    $(CC) $(CFLAGS) -c arena.c -o arena.o

test_arena: test_arena.c arena.o axalloc.o
    $(CC) $(CFLAGS) test_arena.c arena.o axalloc.o -o test_arena
    ./test_arena
```

## Test Plan
1. `ax_arena_create(4096)` → non-NULL arena
2. `ax_arena_alloc(arena, 64)` → non-NULL pointer, 8-byte aligned
3. Allocate until slab is full → new slab created automatically
4. `ax_arena_destroy` → no memory leak (verify with Valgrind or ASan)
5. `ax_arena_reset` → can reallocate from the same arena
6. `ax_arena_used` tracks total bytes allocated
7. Generated `unsafe: block` has no bounds checks in output
8. Generated `in [arena]: block` uses `ax_arena_alloc`, not `ax_alloc`
9. Arena destroy is emitted after the arena block closes

## Validation Checklist
- [ ] `arena.c` compiles without warnings
- [ ] Slab growth works correctly (tested by allocating > initial capacity)
- [ ] `ax_arena_destroy` frees all slabs (ASan verified)
- [ ] `unsafe` flag is correctly scoped (not leaked after unsafe block)
- [ ] Arena variable name is correctly restored after nested arena blocks
- [ ] All tests pass

## Acceptance Criteria
- Arena allocator passes all unit tests including ASan
- Generated C for `in [arena]: block` compiles and runs correctly
- `unsafe:` blocks suppress all safety checks

## Definition of Done
- `runtime/axalloc/arena.h` and `arena.c` exist
- `runtime/axalloc/test_arena.c` exists and passes
- `codegen/cgen/arena.go` exists with arena block emission
- Updated `runtime/axalloc/Makefile` includes arena targets

## Risks & Mitigations
- **Risk**: Arena-allocated pointers used after `ax_arena_destroy` cause silent corruption. **Mitigation**: The ownership checker (Phase 06) enforces that arena pointers cannot outlive the arena block. In debug mode, `ax_arena_destroy` can `memset` each slab to 0 before freeing.
- **Risk**: Nested arenas: inner arena allocations must go to the inner arena. **Mitigation**: `currentArena` is saved and restored on `emitArenaBlock`, correctly implementing lexical scoping.

## Future Follow-up Tasks
- p10-t05: CTGC pass may convert heap allocations to arena allocations when lifetime analysis allows
- p11-t15: Native backend implements arena allocation as a bump pointer in a register
- Future: Thread-local arenas for zero-contention allocation in parallel code
