# p18-t04: Stage 4 — Full Compiler in AXIOM

## Purpose
Assemble the complete AXIOM compiler written in AXIOM by wiring Stages 1-3 (lexer, parser, type checker) together with AIR builder and code generation backend, producing a compiler that can compile AXIOM programs.

## Context
Stage 4 completes the frontend of the self-hosted compiler. With lexer + parser + type checker + AIR builder all written in AXIOM, the compiler can produce AIR from AXIOM source. For Stage 4, the C backend is used for code generation (native backend porting to AXIOM is Stage 5).

## Inputs
- `bootstrap/stage1/lexer.ax`, `stage2/parser.ax`, `stage3/checker.ax`
- AIR instruction set from p09
- C backend from p08/p10

## Outputs
- `bootstrap/stage4/air_builder.ax` — AIR builder in AXIOM
- `bootstrap/stage4/compiler.ax` — complete compiler driver in AXIOM
- `bootstrap/axc_self` — self-compiled compiler binary (compiled by Go axc)

## Dependencies
- p18-t01 through p18-t03: Stages 1-3
- p09: AIR instruction set — opcodes to emit in AXIOM
- p08: C backend — code generation (initially)

## Detailed Requirements

```axiom
# bootstrap/stage4/compiler.ax

type AxcCompiler:
    var lexer:   Lexer
    var parser:  Parser
    var checker: TypeChecker
    var builder: AirBuilder
    var backend: CBackend   # or NativeBackend

    fn new() -> AxcCompiler

    fn compile(mut self, src: str) -> Result[Array[u8], Array[CompileError]]:
        let tokens = self.lexer.tokenize(src)?
        let ast    = self.parser.parse(tokens)?
        self.checker.check(ast)?
        let air    = self.builder.build(ast, self.checker.types)?
        let code   = self.backend.generate(air)?
        Ok(code)
```

Milestone: `bootstrap/axc_self` compiled by Go `axc` can:
1. Compile `hello.ax` → C → gcc → executable → prints "hello world"
2. Output matches Go `axc`'s output

This is the partial self-hosting milestone.

Stage 4 focus is on correctness of the pipeline, not performance. The AXIOM-compiled compiler is expected to be 3-5x slower than Go axc initially.

## Implementation Steps

1. Create `bootstrap/stage4/air_builder.ax` — port Go AIR builder to AXIOM.
2. Create `bootstrap/stage4/cbackend.ax` — port C backend (AIR→C) to AXIOM.
3. Create `bootstrap/stage4/compiler.ax` — wire all stages.
4. Compile with Go `axc`: `axc compile bootstrap/stage4/compiler.ax -o axc_self`.
5. Run `axc_self` on `hello.ax` → verify output matches Go axc.
6. Run `axc_self` on entire stdlib → compare output with Go axc.

## Test Plan
- `TestStage4HelloWorld`: axc_self compiles hello.ax → correct output
- `TestStage4StdlibCompile`: axc_self compiles all stdlib files without error
- `TestStage4OutputEquivalence`: axc_self output matches Go axc on 50 test programs

## Validation Checklist
- [ ] axc_self produces identical C code for test programs (or identical execution results)
- [ ] All stdlib tests pass when compiled by axc_self
- [ ] Error messages match Go axc for all test error cases

## Acceptance Criteria
- `./axc_self compile hello.ax && ./hello` prints "hello world"

## Definition of Done
- [ ] `bootstrap/stage4/compiler.ax` implemented
- [ ] axc_self compiles and runs hello.ax correctly

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Memory usage: AXIOM compiler uses more RAM than Go axc | Profile and optimize critical paths; AXIOM allocator helps |
| AIR builder too complex to port | Port incrementally: scalar types first, generics last |

## Future Follow-up Tasks
- p18-t05: Stage 5 — axc_self compiles axc_self (full self-hosting)
