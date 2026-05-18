# p03-t10: axc dump-ast Command

## Purpose
Implement the `axc dump-ast <file.ax>` CLI subcommand that runs the full lex+parse pipeline and prints the resulting AST to stdout. This is the first end-to-end integration of the compiler frontend and serves as the primary debugging tool for parser development.

## Context
`axc dump-ast` is the first user-visible compiler command. It validates that the lexer, parser, and AST printer work together correctly. It's also the foundation for the full `axc build` command added in Phase 08. The command must handle errors gracefully — printing the partial AST plus diagnostics to stderr, then exiting with code 1.

## Inputs
- `compiler/lexer/lexer.go` — from p02-t02, p02-t03
- `compiler/parser/parser.go` — from p03-t04 through p03-t07
- `compiler/ast/printer.go` — from p03-t03

## Outputs
- `cmd/axc/main.go` — CLI entry point with subcommand dispatch
- `cmd/axc/cmd_dump_ast.go` — dump-ast subcommand implementation

## Dependencies
- p03-t03: ast-printer — the printer used for output
- p03-t07: parser-error-recovery — ensures no panics on bad input
- p01-t01: repository-bootstrap — `cmd/axc/` directory exists

## Subsystems Affected
- CLI toolchain: establishes the axc command structure
- Integration: first end-to-end test of lexer + parser + printer

## Detailed Requirements

1. CLI structure in `cmd/axc/main.go`:
   ```go
   func main() {
       if len(os.Args) < 2 {
           printUsage()
           os.Exit(1)
       }
       switch os.Args[1] {
       case "dump-ast":
           cmdDumpAST(os.Args[2:])
       case "build":
           cmdBuild(os.Args[2:])  // placeholder for Phase 08
       default:
           fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
           os.Exit(1)
       }
   }
   ```
2. `cmdDumpAST(args []string)`:
   - Read source file into `[]byte`
   - Run `Lex()` → `ParseProgram()` → `PrintAST()`
   - Print AST to stdout
   - Print diagnostics to stderr (format: `file.ax:12:8: error: message`)
   - Exit 0 if no errors, exit 1 if any errors
3. Flags: `--json` for JSON output, `--no-color` to disable ANSI colors.
4. Source file must exist and be readable; emit friendly error if not.
5. Binary name: `axc` (built via `go build ./cmd/axc/`).

## Implementation Steps

1. Create `cmd/axc/main.go` with argument dispatch.
2. Create `cmd/axc/cmd_dump_ast.go` with `cmdDumpAST()`.
3. In `cmdDumpAST`: `os.ReadFile(path)` → `lexer.Lex()` → `parser.ParseProgram()` → `ast.PrintTree()`.
4. Print diagnostics to stderr with source location.
5. Add `go build -o bin/axc ./cmd/axc/` to the Makefile.
6. Write integration test: `TestDumpASTHelloWorld` — compile hello_world.ax, verify output contains "FuncDecl" and "main".
7. Add `axc dump-ast` to the CI smoke test.

## Test Plan

- `TestDumpASTHelloWorld`: verify stdout contains expected AST structure
- `TestDumpASTFibonacci`: verify recursive function AST
- `TestDumpASTErrors`: file with syntax errors — exit code 1, diagnostics on stderr
- `TestDumpASTMissingFile`: non-existent file — friendly error message, exit 1
- Integration: `go build ./cmd/axc/ && ./axc dump-ast tests/parser/hello_world.ax`

## Validation Checklist

- [ ] `axc dump-ast hello.ax` prints AST to stdout
- [ ] `axc dump-ast` with no args prints usage and exits 1
- [ ] Error diagnostics go to stderr, not stdout
- [ ] Exit code 0 on success, 1 on errors
- [ ] `axc` binary builds on Linux, Windows, macOS

## Acceptance Criteria

- `axc dump-ast tests/parser/hello_world.ax` produces AST matching golden file
- `axc dump-ast nonexistent.ax` prints "error: file not found" to stderr, exits 1
- Binary builds with `go build ./cmd/axc/` in under 10 seconds

## Definition of Done

- [ ] `cmd/axc/main.go` and `cmd_dump_ast.go` implemented
- [ ] Integration tests pass
- [ ] `axc` binary produced by `go build`
- [ ] CI smoke test runs `axc dump-ast` on a sample file

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| AST printer output format changes break integration tests | Tests check for key substrings, not exact output |
| Windows path handling issues | Use `filepath.Clean` for all path operations |

## Future Follow-up Tasks

- p08-t09: axc build command (adds codegen steps)
- p08-t12: axc emit-c and axc check commands
- p09-t11: axc dump-air command
