# p10-t06: Compile-Time Expression Interpreter (#run)

## Purpose
Implement the compile-time expression interpreter that evaluates `#run` expressions and `OpComptime`-marked AIR subgraphs at compile time, replacing them with constant values.

## Context
`#run` in AXIOM allows arbitrary pure expressions to be evaluated at compile time: `#run fibonacci(20)` computes the 20th Fibonacci number during compilation and replaces the expression with the constant result. The interpreter executes AIR instructions in a virtual machine with a register file, handling all pure operations (no I/O, no allocation, no side effects).

## Inputs
- `AirFunc` with `OpComptime` markers around compile-time evaluable subgraphs
- TypeTable for value type information

## Outputs
- `ir/opt/comptime.go` — CompTimeInterpreter
- OpComptime subgraphs replaced with `OpIConst`/`OpFConst` results
- `[]Diagnostic` — for unsupported operations or errors

## Dependencies
- p10-t01: opt-pipeline-manager — implements OptPass
- p09-t01: air-instruction-set — all ALU opcodes interpreted

## Subsystems Affected
- Compile-time execution: `#run` and `@computed` expressions
- Constant folding: comptime results feed back into constant folder
- User experience: meaningful error for `#run` attempting I/O

## Detailed Requirements

1. `CompTimeInterpreter` struct: `regs []Value, tt *TypeTable`
2. `Value` type: `{TypeID uint32; IVal int64; FVal float64; IsNil bool}`
3. `Interpret(fn *AirFunc, args []Value) (Value, error)` — execute the function with given args.
4. Supported operations: all ALU opcodes (OpIAdd, OpISub, OpIMul, OpIDiv, OpFAdd, etc.), OpIConst, OpFConst, OpICmp*, OpFCmp*, OpBAnd, OpBOr, OpBXor, OpBNot, OpBShl, OpBShr, OpNeg, OpZExt, OpSExt, OpTrunc, OpIToF, OpFToI, OpJump, OpBranch, OpPhi, OpReturn, OpCall (to other pure functions recursively).
5. Unsupported: OpAlloc, OpFree, OpStore, OpLoad (heap), OpSpawn, OpSend, OpRecv → error: `"#run cannot use memory allocation"`.
6. Max instruction count: 100,000 steps before aborting with `"#run: computation exceeded step limit"`.
7. Stack depth limit: 1000 recursive calls.
8. In the optimization pass: find all `OpComptime` markers, extract the subgraph, interpret it, replace with result OpIConst/OpFConst.

## Implementation Steps

1. Create `ir/opt/comptime.go`.
2. Implement `Value` type with all primitive representations.
3. Implement `Interpret()` — step through AIR instructions.
4. Handle OpCall recursively (call Interpret on the callee).
5. Implement `ComptimePass` that implements `OptPass`.
6. In `ComptimePass.Run()`: find OpComptime blocks, interpret, replace with constant.
7. Write tests: `TestComptimeFib`, `TestComptimeFail`, `TestComptimeStepLimit`.

## Test Plan

- `TestComptimeFib`: `#run fibonacci(10)` → interprets fib, returns 55 as iconst
- `TestComptimeArith`: `#run 2 ** 32` → `iconst 4294967296`
- `TestComptimeFail`: `#run read_file("x")` → error: cannot use I/O in #run
- `TestComptimeStepLimit`: `#run loop_1B_times()` → "exceeded step limit" error
- `TestComptimePure`: `#run pure_complex_formula()` → correct result

## Validation Checklist

- [ ] All ALU ops supported
- [ ] Memory ops rejected with clear error
- [ ] Step limit enforced
- [ ] Stack depth limit enforced
- [ ] Result replaces OpComptime with correct type OpIConst/OpFConst
- [ ] AIR verifier passes after comptime replacement

## Acceptance Criteria

- Compliance tests 071-080 (#run group) pass
- `#run fibonacci(20)` evaluates to `iconst 6765` at compile time

## Definition of Done

- [ ] `ir/opt/comptime.go` implemented
- [ ] Registered in O1+ pipeline
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Infinite loop in comptime (no step limit) | 100K step limit, clear error message |
| Comptime accesses global mutable state | Only pure functions allowed; type checker enforces `pure` effect |

## Future Follow-up Tasks

- p16-t16: std.compiler.ai uses comptime to query semantic info at build time
