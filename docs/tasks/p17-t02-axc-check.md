# p17-t02: axc check — Static Analysis

## Purpose
Implement `axc check` — a fast type-checking and static analysis pass that reports errors and warnings without emitting code, suitable for editor integration and CI pre-flight.

## Context
`axc check` runs the frontend pipeline (lexer → parser → semantic → type check) without backend code generation, providing fast feedback. It also runs additional linting passes (unused variables, unreachable code, deprecated APIs) and outputs structured diagnostics.

## Inputs
- AXIOM source files
- Full compiler frontend pipeline (p02-p06)
- Linting rules (configurable via `axiom.toml`)

## Outputs
- `tools/check/check.go` — check driver
- Structured diagnostic output (JSON or human-readable)
- `axc check` subcommand

## Dependencies
- p02: lexer — tokenization
- p03: parser — AST
- p04: semantic/type checker — type errors
- p09-t04: air-verifier — optional AIR verification

## Subsystems Affected
- Editor/LSP: LSP server calls `axc check` for diagnostics
- CI: `axc check` as pre-commit gate

## Detailed Requirements

Check passes run in order:
1. Lexer (syntax errors)
2. Parser (grammar errors)
3. Name resolution (undefined symbols)
4. Type checker (type errors)
5. Linting (warnings)

Lint rules:
```
E0001: unused variable
E0002: unused import
E0003: unreachable code after return/panic
E0004: unused function parameter
W0001: deprecated function used
W0002: integer overflow risk in constant expression
W0003: comparison of float equality with ==
W0004: shadowed variable in inner scope
```

Output formats:
```
# Human-readable (default)
error[E0001]: unused variable 'x'
 --> src/main.ax:12:5
  |
12 |     let x = 42
  |         ^ variable 'x' is never used

# JSON (--json flag)
{"severity":"error","code":"E0001","message":"unused variable 'x'",
 "file":"src/main.ax","line":12,"col":5}
```

```go
type CheckResult struct {
    Diagnostics []Diagnostic
    ErrorCount  int
    WarnCount   int
}

func Check(paths []string, flags CheckFlags) CheckResult
```

Performance target: < 100ms for a 10K-line project (no backend invoked).

## Implementation Steps

1. Create `tools/check/check.go`.
2. Invoke frontend pipeline without backend.
3. Collect all diagnostics from each pass.
4. Implement lint rules as separate passes over TypedAST.
5. Implement `--json` output format.
6. Add `axc check` subcommand.
7. Write tests: each lint rule triggers correctly.

## Test Plan
- `TestCheckUnusedVar`: unused variable → E0001 diagnostic
- `TestCheckTypeError`: type mismatch → error with correct location
- `TestCheckJSON`: --json flag → parseable JSON output
- `TestCheckClean`: correct file → exit 0
- `TestCheckPerf`: 10K-line file → check completes < 100ms

## Validation Checklist
- [ ] All error codes documented
- [ ] Source locations accurate (line:col match actual position)
- [ ] JSON output valid (parseable by std.json)
- [ ] Exit code 0 = no errors, 1 = errors, 2 = internal error

## Acceptance Criteria
- LSP server uses `axc check --json` for real-time diagnostics

## Definition of Done
- [ ] `tools/check/check.go` implemented
- [ ] All lint rules implemented and tested
- [ ] JSON output format finalized

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Lint rules produce false positives | Each rule behind a flag; disabled by default until stable |

## Future Follow-up Tasks
- Incremental check: only re-check modified files
- Custom lint rules via plugin API
