# p09-t11: axc dump-air Command

## Purpose
Implement the `axc dump-air <file.ax>` CLI command that runs the full pipeline through AIR generation and prints the resulting AIR to stdout. This is the primary debugging tool for AIR builder and optimization pass development.

## Context
`axc dump-air` is the AIR-level equivalent of `axc dump-ast`. It runs: lex → parse → sema → ownership → CTGC → AIR builder → verify → print. The command is essential for verifying that complex programs produce correct AIR before testing codegen.

## Inputs
- Source `.ax` file
- Full pipeline: lexer + parser + sema + ownership + AIR builder
- AIR printer (p09-t05)
- AIR verifier (p09-t04)

## Outputs
- `cmd/axc/cmd_dump_air.go` — CLI command implementation
- AIR text printed to stdout
- Diagnostics (from all stages) printed to stderr
- Exit 0 if AIR valid, exit 1 otherwise

## Dependencies
- p09-t05: air-text-printer — the printer used
- p09-t04: air-verifier — validates the produced AIR
- p09-t10: air-builder-ownership — completes the AIR builder
- p03-t10: axc-dump-ast-command — CLI infrastructure to extend

## Subsystems Affected
- CLI toolchain: new subcommand added to axc
- Integration: end-to-end test of full frontend + AIR builder

## Detailed Requirements

1. `axc dump-air <file.ax>` command:
   - Run full pipeline through AIR builder
   - Run AIR verifier on each AirFunc
   - Print AIR text to stdout using AirPrinter
   - Print any diagnostics to stderr
   - Exit 0 if 0 errors, exit 1 if any errors
2. Flags:
   - `--debug`: include source location comments in AIR output
   - `--verify`: run verifier (default: on; `--no-verify` to skip)
   - `--func=<name>`: print only the specified function's AIR
   - `--json`: output AIR as JSON (machine-readable)
3. Add `dump-air` to the dispatch table in `cmd/axc/main.go`.
4. Show verifier errors inline: `; VERIFIER ERROR: %r5 defined twice` comment in output.

## Implementation Steps

1. Create `cmd/axc/cmd_dump_air.go` with `cmdDumpAIR(args)`.
2. Call the full pipeline: Lex → Parse → Resolve → TypeCheck → Ownership → CTGC → BuildAIR.
3. For each function in AirModule: run verifier, collect errors.
4. Call `air.PrintModule(os.Stdout, module)` for the text output.
5. Add verifier errors as comments in the output.
6. Parse `--func` flag to filter output.
7. Write integration test: `TestDumpAIRFibonacci` — verify output contains expected AIR structure.

## Test Plan

- `TestDumpAIRHelloWorld`: `axc dump-air hello.ax` → stdout contains `fn _AX_main`
- `TestDumpAIRFibonacci`: verify recursive call in AIR
- `TestDumpAIRErrors`: source with type error → exit 1, diagnostic on stderr
- `TestDumpAIRVerifierFail`: artificially inject bad AIR (only via test) → verifier error shown
- `TestDumpAIRFuncFilter`: `--func=main` → only main function printed

## Validation Checklist

- [ ] `axc dump-air hello.ax` exits 0 with AIR output
- [ ] Verifier errors appear as comments in output
- [ ] Diagnostics on stderr, AIR on stdout (cleanly separated)
- [ ] `--func` filter works correctly
- [ ] Exit code 0/1 correct

## Acceptance Criteria

- `axc dump-air tests/parser/fibonacci.ax` produces AIR with recursive call visible
- `axc dump-air --debug tests/parser/fibonacci.ax` shows source line annotations

## Definition of Done

- [ ] `cmd/axc/cmd_dump_air.go` implemented
- [ ] Integration tests pass
- [ ] Added to CLI dispatch in main.go

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Long AIR output for large files | `--func` filter; paginate or truncate at 10K lines with notice |
| Verifier errors don't correspond to readable locations | Print block_ID:inst_N to locate |

## Future Follow-up Tasks

- p09-t12: AIR golden tests use dump-air output
- p10-t01: optimization pipeline adds `--opt=O2` flag to dump-air
