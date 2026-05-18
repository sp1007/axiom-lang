# p08-t13: C-Backend `defer` Statement

## Purpose
Implement C code generation for AXIOM's `defer` statement. `defer` schedules a call to be executed when the enclosing scope exits — similar to Go's `defer` but block-scoped. The C-backend emits deferred calls at every exit point (normal, return, break, continue) in LIFO order.

## Context
Plan §Issue 8 resolves `defer` as a reserved keyword: `DeferStmt ::= "defer" CallExpr NEWLINE`. C has no native `defer`, so the C-backend must collect deferred calls and emit them at block exits.

## Inputs
- `DeferStmt` AST node from parser (p03-t04)
- C-backend statement generator from p08-t03
- Scope tracking from block emission

## Outputs
- `codegen/cgen/defer.go` — defer collection and emission
- Updated `codegen/cgen/stmtgen.go` — integration with block/return
- Tests and golden output

## Dependencies
- p08-t03: cgen-statements — statement emission
- p08-t05: cgen-ownership — `=destroy` ordering with defer

## Subsystems Affected
- C-Backend: block and return emission
- Ownership: `=destroy` nodes run AFTER deferred calls

## Detailed Requirements

### Defer Stack

Each scope block maintains a LIFO stack of deferred calls:

```go
type DeferEntry struct {
    CallExpr   uint32  // AST node index
    ScopeDepth int     // nesting depth
}
```

### C Emission Strategy

**AXIOM:**
```axiom
fn process(path: string) -> i32:
    let fd = open(path)
    defer close(fd)
    let data = read(fd)
    defer free(data)
    if data.len == 0:
        return -1       # emits: free(data); close(fd); return -1;
    return 0            # emits: free(data); close(fd); return 0;
```

**Generated C:**
```c
int32_t _AX_process(ax_string path) {
    int32_t fd = _AX_open(path);
    ax_data data = _AX_read(fd);
    if (data.len == 0) {
        _AX_free(data);
        _AX_close(fd);
        return -1;
    }
    _AX_free(data);
    _AX_close(fd);
    return 0;
}
```

### Execution Order at Block Exit

1. **Deferred calls** (LIFO order)
2. **`=destroy` nodes** (ownership cleanup)

### Nested Blocks

Deferred calls are scoped to their declaring block:
```axiom
fn example():
    defer a()
    if cond:
        defer b()
        # block exit: b() only (a() is in outer scope)
    # function exit: a()
```

### Break/Continue

`break` and `continue` emit deferred calls for the exiting block before the jump.

### Error Handling

If deferred call itself can fail, error is silently ignored in MVP.

## Implementation Steps

1. Create `codegen/cgen/defer.go` with `DeferStack`.
2. In `emitBlock()`: push scope onto defer stack at entry.
3. On `DeferStmt`: push call to stack (don't emit yet).
4. In `emitReturn()`: emit all defers LIFO before `return`.
5. In `emitBlockEnd()`: emit block-scoped defers, pop scope.
6. In `emitBreak()/emitContinue()`: emit current block defers before jump.
7. Ensure `=destroy` nodes emitted AFTER defers.
8. Write tests.

## Test Plan

- `TestDeferSimple`: single defer → emitted at block end
- `TestDeferLIFO`: two defers → second emitted first
- `TestDeferEarlyReturn`: defer before return
- `TestDeferNested`: inner block defer doesn't leak to outer
- `TestDeferLoop`: defer in loop body → emitted each iteration
- `TestDeferBreak`: defer emitted before break
- `TestDeferWithDestroy`: defers before `=destroy` calls
- Golden: `tests/golden/cgen/defer_basic.ax` → expected C

## Validation Checklist

- [ ] LIFO ordering correct
- [ ] Early returns include all defers
- [ ] Nested blocks scope correctly
- [ ] break/continue emit block-scoped defers
- [ ] `=destroy` runs after defers
- [ ] Generated C compiles clean

## Acceptance Criteria

- All defer compliance tests pass when compiled and run
- Golden test output matches expected C

## Definition of Done

- [ ] `codegen/cgen/defer.go` implemented
- [ ] `codegen/cgen/stmtgen.go` updated
- [ ] Unit tests pass
- [ ] Golden tests committed

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Deep nesting creates complex paths | Limit to 256 depth; recursive scope walk |
| Defer referencing moved values | Ownership checker prevents at compile time |

## Future Follow-up Tasks

- p09-t07: AIR builder lowers DeferStmt to AIR
- p10-t03: DCE must not eliminate deferred calls
