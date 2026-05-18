# p09-t05: AIR Text Printer

## Purpose
Implement a human-readable text printer for AIR that enables debugging of the AIR builder and optimization passes. The printed format should be clear enough for a compiler engineer to read and verify correctness by inspection.

## Context
The AIR text printer is the primary debugging tool for all AIR-level work. It is used by `axc dump-air` and in golden tests. The format must be stable (golden tests depend on it), readable, and complete (no information loss).

## Inputs
- `AirModule` or `AirFunc` from the AIR builder
- `InternPool` for resolving name IDs
- `MetaTable` for source annotations

## Outputs
- `ir/air/printer.go` — AirPrinter with `Print(module)` method
- Text output format: one instruction per line with register names

## Dependencies
- p09-t03: air-metadata-table — source locations for `--debug` mode
- p09-t02: air-basic-blocks — CFG structure to traverse
- p09-t01: air-instruction-set — mnemonic strings

## Subsystems Affected
- Debugging: primary tool for inspecting AIR
- Golden tests (p09-t12): printer output compared against expected
- axc dump-air command (p09-t11)

## Detailed Requirements

1. Output format:
   ```
   fn _AX_main():
     block_0:  ; entry
       %0: i32 = iconst 42
       %1: i32 = iconst 58
       %2: i32 = iadd %0, %1
       return %2
   ```
2. Register names: `%N` where N is the virtual register index.
3. Block labels: `block_N:` with optional `;` comments for entry/exit.
4. Instruction format: `[%dest: type] = mnemonic [%src1] [%src2]` or `mnemonic %src1, %src2` for void instructions.
5. Phi nodes: `%r = phi [%v1, block_2], [%v2, block_4]`.
6. Function calls: `%r: i32 = call @fn_name(%arg1, %arg2)`.
7. `--debug` flag adds `; file.ax:12` source location comment.
8. `--no-types` flag omits type annotations (shorter output).
9. `Print(w io.Writer, module *AirModule)` — writes to any Writer.
10. `PrintFunc(w io.Writer, fn *AirFunc)` — prints single function.

## Implementation Steps

1. Create `ir/air/printer.go` with `AirPrinter` struct.
2. Implement `PrintFunc()`: iterate blocks in order, print each block's instructions.
3. Implement `printInst(inst AirInst)`: format based on opcode class.
4. Handle special cases: phi nodes (multi-source), calls (variadic args from Extras).
5. Implement `resolveTypeName(typeID uint32) string` using TypeTable.
6. Add `--debug` mode: append `;file.ax:N` from MetaTable.
7. Write round-trip test: parse `.air` text → build AIR → print → compare.

## Test Plan

- `TestPrintSimpleFunc`: function with arithmetic → verify format
- `TestPrintIfElse`: function with branch → verify block labels and branch inst
- `TestPrintPhi`: phi node → verify `phi [%v1, block_2], [%v2, block_4]` format
- `TestPrintCall`: function call with 3 args → verify args printed
- `TestPrintDebugMode`: verify source location comments in --debug mode
- Golden tests in p09-t12 validate printer output

## Validation Checklist

- [ ] All opcodes have non-empty printed form
- [ ] Register names consistent (always %N)
- [ ] Block order: entry block first
- [ ] Phi nodes printed with all incoming values
- [ ] Source annotations appear with --debug flag

## Acceptance Criteria

- `axc dump-air hello.ax` produces readable, stable output
- Golden tests (p09-t12) all pass against printer output

## Definition of Done

- [ ] `ir/air/printer.go` implemented
- [ ] All instruction kinds printed correctly
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Format changes break golden tests | Version the format; use --update to regenerate |
| Very long arg lists in calls break readability | Wrap at 80 chars or show first N args with ... |

## Future Follow-up Tasks

- p09-t11: axc-dump-air command uses this printer
- p09-t12: AIR golden tests use printer output
