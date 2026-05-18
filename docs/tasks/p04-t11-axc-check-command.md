# p04-t11: `axc check` Command (Early Integration)

## Purpose
Implement the `axc check <file.ax>` command that runs the full frontend pipeline (lex → parse → name resolution → type check) and reports all diagnostics to stderr. This is the primary developer feedback loop tool — it must be available as soon as the type checker is functional (Phase 4), not deferred to Phase 17 tooling.

## Context
The plan specifies `axc check <file.ax>` as a Phase 2 output (plan §Phase 2): _"`axc check <file.ax>` — reports type errors, exits 0 if clean"_. This must be wired into the CLI immediately after the type checker is functional. The later `p17-t02` task will extend this with LSP integration and incremental checking; this task provides the baseline command.

## Inputs
- `cmd/axc/main.go` from p01-t01 — CLI entry point
- Lexer pipeline from p02
- Parser pipeline from p03
- Name resolver from p04-t04
- Type checker from p04-t05/t06/t07
- Diagnostic formatter from p01-t06
- Overload resolution from p04-t08
- Effects system from p04-t09

## Outputs
- `cmd/axc/check.go` — `runCheck()` function
- Updated `cmd/axc/main.go` — `check` subcommand routing
- Integration tests for `axc check`

## Dependencies
- p04-t10: sema-golden-tests — type checker verified working
- p01-t06: diagnostic-formatter — error display
- p03-t10: axc-dump-ast-command — CLI infrastructure established

## Subsystems Affected
- CLI — new `check` subcommand
- All frontend passes — orchestrated here for validation

## Detailed Requirements

### 1. CLI Interface

```
axc check <file.ax> [flags]

Exit codes:
  0  — no errors (warnings may be present)
  1  — one or more type/semantic errors
  2  — internal compiler error (ICE)

Flags:
  --warnings-as-errors    Treat warnings as errors (exit 1 if any warnings)
  --no-color              Disable colored output
  -v, --verbose           Print each pipeline stage
```

### 2. Pipeline

```go
func runCheck(args []string) {
    opts := parseCheckFlags(args)
    
    // Read source
    src, err := os.ReadFile(opts.InputFile)
    if err != nil { fatal(err) }
    
    // Stage 1: Lex
    tokens, lexErrs := lexer.Tokenize(src)
    allDiags := toDiagnostics(lexErrs)
    
    // Stage 2: Parse
    tree := ast.NewTree()
    parser := parser.New(tokens, tree)
    _, parseErrs := parser.ParseFile()
    allDiags = append(allDiags, toDiagnostics(parseErrs)...)
    
    // Stage 3: Name resolution
    symtab := sema.NewSymbolTable()
    resolver := sema.NewNameResolver(tree, symtab)
    resolveErrs := resolver.Resolve()
    allDiags = append(allDiags, resolveErrs...)
    
    // Stage 4: Type check (only if no fatal parse errors)
    if !hasFatal(allDiags) {
        tc := sema.NewTypeChecker(tree, symtab)
        typeErrs := tc.Check()
        allDiags = append(allDiags, typeErrs...)
    }
    
    // Report
    fmt.Fprint(os.Stderr, diagnostics.FormatDiagnostics(allDiags, src, opts.InputFile, fmtOpts))
    
    if hasErrors(allDiags) || (opts.WarningsAsErrors && hasWarnings(allDiags)) {
        os.Exit(1)
    }
    os.Exit(0)
}
```

### 3. Multi-File Support (Future)

For now, `axc check` operates on a single file. Multi-file checking (`axc check ./...`) is deferred to p17-t02 (requires import resolution and module graph).

### 4. Error Deduplication

If the same error is reported at the same position by multiple passes, deduplicate before display. Use `(Pos, Code)` as the dedup key.

### 5. Sorting

Diagnostics are displayed sorted by `(Line, Col, Severity)` — errors before warnings at the same position.

### 6. Summary Line

After all diagnostics, print a summary:
```
error: 3 errors, 1 warning emitted
```
Or on success (with possible warnings):
```
axc: 0 errors, 2 warnings in main.ax
```

## Implementation Steps

1. Create `cmd/axc/check.go` with `runCheck()` function.
2. Update `cmd/axc/main.go` to route `"check"` to `runCheck()`.
3. Implement flag parsing with `flag.NewFlagSet("check", ...)`.
4. Wire the lex → parse → resolve → typecheck pipeline.
5. Collect all diagnostics, deduplicate, sort, format.
6. Print summary line.
7. Exit with appropriate code (0/1/2).
8. Write integration test: valid file → exit 0, invalid file → exit 1 with error on stderr.

## Test Plan

### Unit Tests
- `TestCheckValidFile`: well-typed AXIOM file → exit 0, no stderr output (except optional warnings)
- `TestCheckTypeError`: `let x: i32 = "hello"` → exit 1, stderr contains `type mismatch`
- `TestCheckParseError`: syntax error → exit 1, stderr contains parser error
- `TestCheckMultipleErrors`: file with 3 errors → all 3 reported, exit 1
- `TestCheckWarningsAsErrors`: `--warnings-as-errors` + warning → exit 1
- `TestCheckICE`: trigger ICE (if possible with mock) → exit 2

### Integration Tests
- Compile `tests/golden/sema/valid_arithmetic.ax` → `axc check` exits 0
- Compile `tests/golden/sema/type_error.ax` → `axc check` exits 1, stderr matches golden

## Validation Checklist

- [ ] `axc check valid.ax` exits 0
- [ ] `axc check invalid.ax` exits 1 with diagnostic on stderr
- [ ] Diagnostics include file:line:col
- [ ] Summary line printed after diagnostics
- [ ] `--no-color` disables ANSI output
- [ ] Multiple errors all reported (not just first)
- [ ] Error deduplication works

## Acceptance Criteria

- `axc check axiom_compliance_suite.ax` exits 0 for groups 1–5 (per plan §Phase 2)
- Error messages match format: `error[E####]: message` with source snippet

## Definition of Done

- [ ] `cmd/axc/check.go` implemented
- [ ] `cmd/axc/main.go` updated with `check` routing
- [ ] Integration tests pass
- [ ] `axc check` works on at least 5 golden test files

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Type checker not ready when this task starts | This task depends on p04-t10 (sema golden tests), ensuring TC is functional |
| ICE handling complex | Wrap pipeline in recover(); format ICE with `FormatICE()` |

## Future Follow-up Tasks

- p08-t09: `axc build` reuses the same pipeline with additional codegen stages
- p17-t02: extends `axc check` with incremental mode, multi-file, and LSP integration
- p17-t05: incremental compilation uses `axc check` as the validation step
