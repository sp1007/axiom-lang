# p06-t07: Arena Block Handling

## Purpose
Parse and implement the `in [arena]:` block syntax that routes all allocations within the block to an arena allocator, enabling O(1) bulk deallocation of all objects in the block at once. This is critical for performance-sensitive code (game loops, math kernels) that needs C-like allocation speed.

## Context
Arena allocators are dramatically faster than general-purpose allocators because they bump a pointer for each allocation and free everything at once. AXIOM's `in [arena]:` block creates a lexical region where all allocations use the provided arena. This completely bypasses generational reference checks (the arena owns all memory; use-after-free is the programmer's responsibility, similar to `unsafe`). Arena blocks are the primary escape hatch from automatic memory management.

## Inputs
- Parser — `in [arena]: block` syntax (NodeKind.ArenaBlock)
- Type checker — `arena` variable must be `mem.Arena` type
- SymbolTable — `UsesArena` flag on VarDecl nodes

## Outputs
- `NodeKind.ArenaBlock` AST node
- `Flags |= UsesArena` on all VarDecl nodes inside the block
- No DestroyStmt injection for arena-allocated values (CTGC skips them)
- Arena destructor injected at block exit: `=destroy(arena)`

## Dependencies
- p06-t05: ctgc-destroy-injection — must be aware of arena blocks to skip
- p03-t06: parser-indentation — block parsing used here
- p04-t02: type-table-primitives — mem.Arena type must exist

## Subsystems Affected
- Parser: new ArenaBlock node kind
- Ownership checker: UsesArena values exempt from ownership rules
- CTGC: skips destroy injection for UsesArena values
- C-backend: uses arena allocator API

## Detailed Requirements

1. Parser: `in [exprList]: block` → `NodeKind.ArenaBlock{arenas:[]exprIdx, body:blockIdx}`.
   - `exprList` can be `[arena1, arena2]` — multiple arenas (round-robin or first-fit).
2. Type checker: verify each expr in arenaList has type `mem.Arena`.
3. Mark all `VarDecl` nodes inside the ArenaBlock body with `Flags |= FlagUsesArena`.
4. CTGC pass: skip `DestroyStmt` injection for `FlagUsesArena` values.
5. At ArenaBlock exit: inject single `DestroyStmt{target: arena}` → `ax_arena_destroy(arena)`.
6. No generational reference checks for arena-allocated values (no `ax_make_ref`, no `ax_deref`).
7. C-backend: `ax_arena_alloc(arena, sizeof(T))` instead of `ax_alloc(sizeof(T))`.
8. Nested arenas: inner `in [inner_arena]:` allocates into inner_arena; outer arena still accessible.

## Implementation Steps

1. Add `NodeKind.ArenaBlock` to `ast/node.go`.
2. In parser: parse `in [expr_list]: block` into ArenaBlock node.
3. In type checker: verify arena types.
4. In ownership pass: mark all VarDecl in ArenaBlock body with `FlagUsesArena`.
5. In CTGC pass: add check `if flags & FlagUsesArena { skip }`.
6. Add arena destroy at ArenaBlock exit.
7. In cgen: check `FlagUsesArena` → use `ax_arena_alloc`.
8. Write arena C runtime: `ax_arena_create(size)`, `ax_arena_alloc(arena, size)`, `ax_arena_destroy(arena)`.
9. Write tests: `TestArenaBasic`, `TestArenaNoGenCheck`, `TestArenaDestroy`, `TestArenaNested`.

## Test Plan

- `TestArenaBasic`: `let arena = mem.Arena(1024); in [arena]: { let x = Foo{} }` — x allocated in arena
- `TestArenaNoGenCheck`: within arena block, no ax_deref calls emitted (verified in C output)
- `TestArenaDestroy`: at block exit, ax_arena_destroy called
- `TestArenaNested`: nested arena blocks each use their own allocator
- `TestArenaCompliance`: compliance tests sys_025 (custom arena allocator) passes

## Validation Checklist

- [ ] in [arena]: block parses correctly
- [ ] Arena type verified in type checker
- [ ] VarDecl inside marked with FlagUsesArena
- [ ] CTGC skips UsesArena values
- [ ] Arena destroy injected at block exit
- [ ] C-backend uses ax_arena_alloc for arena values
- [ ] No gen_id checks for arena values

## Acceptance Criteria

- Low-level compliance test `sys_025_custom_arena_allocator` passes
- Arena-allocated code has no ax_alloc or ax_deref calls (verified in emitted C)

## Definition of Done

- [ ] ArenaBlock node kind added and parsed
- [ ] Arena type checking implemented
- [ ] CTGC integration complete
- [ ] C runtime arena API implemented
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Use-after-arena-destroy (UAF) not caught | Document as unsafe behavior; generational refs don't apply; programmer's responsibility |
| Arena overflow (arena too small) | ax_arena_alloc panics with clear message on overflow |

## Future Follow-up Tasks

- p10-t05: opt-ctgc-air handles arena at AIR level
- p16-t14: std.mem implements the mem.Arena type
