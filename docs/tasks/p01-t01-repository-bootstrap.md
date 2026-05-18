# p01-t01: Repository Bootstrap

## Purpose
Initialize the AXIOM compiler monorepo as a Go module with the full canonical directory structure, build system scaffolding, and a working `go test ./...` baseline. This is the foundation every subsequent task depends on — without a well-structured monorepo and a passing zero-test baseline, no other engineering work can be integrated cleanly.

## Context
AXIOM's bootstrap compiler is written in Go 1.22+. The module path is `github.com/axiom-lang/axiom`. The repository is a monorepo hosting the compiler, runtime, standard library, tools, tests, examples, and documentation. Every directory listed here corresponds to a subsystem defined in the compiler pipeline: Lexer → FlatAST → SemanticGraph → TypedAST+ConnectionGraph → AIR → OptimizedAIR → [C-Backend|NativeBackend] → ELF/PE/MachO + .axmeta. Getting this structure right at the start prevents structural refactoring later, which would break cross-package imports across many files.

## Inputs
- No files from the project (this is the bootstrap task)
- Go 1.22+ toolchain installed on the developer machine
- `golangci-lint` v1.57+ installed

## Outputs
- `go.mod` with module `github.com/axiom-lang/axiom`, Go 1.22
- `go.sum` (empty initially)
- `cmd/axc/main.go` — CLI entry point (stub)
- `compiler/lexer/` — package `lexer`
- `compiler/parser/` — package `parser`
- `compiler/ast/` — package `ast`
- `compiler/sema/` — package `sema`
- `compiler/types/` — package `types`
- `compiler/diagnostics/` — package `diagnostics`
- `compiler/driver/` — package `driver`
- `ir/air/` — package `air`
- `ir/builder/` — package `builder`
- `ir/opt/` — package `opt`
- `codegen/cgen/` — package `cgen`
- `codegen/native/` — package `native`
- `runtime/axalloc/` — package `axalloc`
- `runtime/actors/` — package `actors`
- `runtime/panic/` — package `panic_`
- `tools/lsp/` — package `lsp`
- `tools/pkg/` — package `pkg`
- `std/` — placeholder for stdlib `.ax` source files
- `tests/` — test data directory (not a Go package)
- `docs/` — documentation directory
- `rfcs/` — RFC documents
- `benchmarks/` — Go benchmark files
- `fuzz/` — fuzz corpus directories
- `bootstrap/` — self-hosting bootstrap artifacts
- `ci/` — CI scripts
- `scripts/` — developer utility scripts
- `.golangci.yml` — linter configuration
- `Makefile` — common dev commands

## Dependencies
None. This is the first task.

## Subsystems Affected
- All: every subsystem depends on the directory structure and module path established here.

## Detailed Requirements

1. **Go module initialization**
   ```
   module github.com/axiom-lang/axiom

   go 1.22
   ```
   The module path `github.com/axiom-lang/axiom` must be used consistently in all import paths. No relative imports allowed anywhere in the codebase.

2. **Stub package files**: Every directory that is a Go package must contain at least one `.go` file with the correct `package` declaration so `go build ./...` succeeds. Example for `compiler/lexer/`:
   ```go
   // Package lexer implements the AXIOM source lexer.
   // It converts raw UTF-8 source bytes into a flat []Token slice
   // with no string allocations (zero-copy design).
   package lexer
   ```

3. **CLI entry point** at `cmd/axc/main.go`:
   ```go
   package main

   import (
       "fmt"
       "os"
   )

   func main() {
       if len(os.Args) < 2 {
           fmt.Fprintln(os.Stderr, "usage: axc <command> [args]")
           os.Exit(1)
       }
       fmt.Fprintln(os.Stderr, "axc: unknown command:", os.Args[1])
       os.Exit(1)
   }
   ```

4. **Diagnostics package** at `compiler/diagnostics/diagnostics.go` — define the shared `Diagnostic` type used by all passes:
   ```go
   package diagnostics

   // Severity classifies how serious a diagnostic is.
   type Severity uint8

   const (
       SeverityError   Severity = iota
       SeverityWarning
       SeverityNote
   )

   // Pos identifies a byte offset in the source file.
   type Pos struct {
       Offset uint32
       Line   uint32
       Col    uint32
   }

   // Diagnostic is a compiler message attached to a source location.
   // Compiler passes MUST NOT panic; they return []Diagnostic instead.
   type Diagnostic struct {
       Severity Severity
       Code     uint32
       Pos      Pos
       Message  string
       Hint     string // optional actionable hint
   }

   func (d *Diagnostic) Error() string { return d.Message }
   ```

5. **Linter configuration** at `.golangci.yml`:
   ```yaml
   run:
     timeout: 5m
   linters:
     enable:
       - errcheck
       - gosimple
       - govet
       - ineffassign
       - staticcheck
       - unused
       - gofmt
       - goimports
   linters-settings:
     goimports:
       local-prefixes: github.com/axiom-lang/axiom
   ```

6. **Makefile** with targets:
   - `make build` → `go build ./...`
   - `make test` → `go test ./...`
   - `make lint` → `golangci-lint run`
   - `make fuzz-lexer` → `go test -fuzz=FuzzLexer ./compiler/lexer/ -fuzztime=60s`
   - `make clean` → remove `bin/`

7. **`runtime/panic/` package name**: Because `panic` is a Go builtin, the package at `runtime/panic/` must declare `package panic_` to avoid collision. This must be documented in the stub file.

8. **`std/` directory**: This is NOT a Go package. It will contain `.ax` source files. Place a `README.md` inside explaining this distinction.

9. **`tests/` directory**: Not a Go package. Contains test data subdirectories: `tests/lexer/`, `tests/parser/`, `tests/sema/`. Place a `.gitkeep` or stub file.

10. **Verify `go build ./...` and `go test ./...` pass** with zero errors and zero test failures after bootstrapping. All packages with only stub files will have no test functions, which is valid.

## Implementation Steps

1. Create the root `go.mod`:
   ```
   go mod init github.com/axiom-lang/axiom
   ```
   Then edit it to set `go 1.22`.

2. Create every directory listed in Outputs. In Go, empty directories are not tracked by git. Every Go package directory needs at least one `.go` file.

3. Write stub `.go` files for each package. Follow this pattern for each:
   - File: `<dir>/doc.go` or `<dir>/<pkgname>.go`
   - Content: package declaration + a one-paragraph doc comment explaining the package's role

4. Write `compiler/diagnostics/diagnostics.go` with the full `Diagnostic` type (see Detailed Requirements #4).

5. Write `cmd/axc/main.go` with the stub main function (see Detailed Requirements #3).

6. Create `.golangci.yml` (see Detailed Requirements #5).

7. Create `Makefile` with all targets (see Detailed Requirements #6).

8. Create `std/README.md` explaining that `std/` contains AXIOM source files (`.ax`), not Go source.

9. Create stub directories for non-Go content: `tests/lexer/`, `tests/parser/`, `tests/sema/`, `rfcs/`, `benchmarks/`, `fuzz/`, `bootstrap/`, `ci/`, `scripts/`, `docs/`.

10. Run `go build ./...` — fix any compilation errors.

11. Run `go test ./...` — should output `ok` or `[no test files]` for every package, zero failures.

12. Run `golangci-lint run` — fix any lint issues.

13. Commit with message: `chore: initialize axiom monorepo structure`.

## Test Plan

- **TestPackageCompiles** (implicit): `go build ./...` must exit 0. This is verified in CI (p01-t04).
- **TestDiagnosticFields**: Write `compiler/diagnostics/diagnostics_test.go`:
  ```go
  package diagnostics

  import "testing"

  func TestDiagnosticZeroValue(t *testing.T) {
      var d Diagnostic
      if d.Severity != SeverityError {
          t.Fatalf("expected zero value severity=SeverityError, got %d", d.Severity)
      }
  }

  func TestDiagnosticMessage(t *testing.T) {
      d := Diagnostic{
          Severity: SeverityError,
          Code:     1001,
          Pos:      Pos{Offset: 42, Line: 3, Col: 7},
          Message:  "undefined variable",
      }
      if d.Error() != "undefined variable" {
          t.Fatalf("Error() = %q, want %q", d.Error(), "undefined variable")
      }
  }
  ```

- **No panics in main**: `go run ./cmd/axc` with no args should print usage to stderr and exit 1 (tested manually or in a shell integration test).

## Validation Checklist
- [ ] `go.mod` has module `github.com/axiom-lang/axiom` and `go 1.22`
- [ ] `go build ./...` exits 0 with no errors
- [ ] `go test ./...` exits 0, all packages pass or have no tests
- [ ] `golangci-lint run` exits 0 with no issues
- [ ] All package directories have at least one `.go` stub file
- [ ] `compiler/diagnostics/` package is defined with `Diagnostic`, `Severity`, `Pos` types
- [ ] `cmd/axc/main.go` compiles and prints usage when run with no args
- [ ] `runtime/panic/` package declares `package panic_`
- [ ] `Makefile` has `build`, `test`, `lint`, `clean` targets
- [ ] `std/README.md` exists explaining `.ax` content
- [ ] No circular imports anywhere

## Acceptance Criteria
- `go build ./...` passes on Go 1.22+ on Linux, macOS, and Windows
- `go test ./...` reports zero failures
- `golangci-lint run` reports zero issues
- Directory tree matches the layout specified in Outputs exactly
- Module path is `github.com/axiom-lang/axiom` throughout
- `axc` binary can be built with `go build -o bin/axc ./cmd/axc`

## Definition of Done
- [ ] All tests pass (`go test ./...`)
- [ ] No linter errors (`golangci-lint run`)
- [ ] `go build ./...` succeeds
- [ ] Directory structure reviewed and matches spec
- [ ] `Makefile` tested on the target platform
- [ ] `compiler/diagnostics/` reviewed by second engineer
- [ ] Committed to main branch with passing CI (p01-t04 must be set up in parallel or immediately after)

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| `runtime/panic` package name conflicts with Go builtin | Declare `package panic_`; document this in the stub file header comment |
| Circular import between `compiler/ast` and `compiler/types` | Keep `types` dependency-free from `ast`; `ast` may reference `types` only by ID (uint32), never by pointer to TypeInfo |
| Developer adds Go files to `std/` or `tests/` by mistake | Add `// +build ignore` files or `.gitignore` notes; document clearly in README files |
| Go module cache issues in CI | Pin Go version in CI matrix (p01-t04); use `go mod download` cache step |
| Mismatched package names vs directory names | Enforce `goimports` lint rule; package name must match directory name (except `panic_`) |

## Future Follow-up Tasks
- p01-t02: Write the EBNF grammar (depends on this structure being in place)
- p01-t03: Define FROZEN struct layouts in the established packages
- p01-t04: Set up CI pipeline that runs `go build`, `go test`, `golangci-lint` against this structure
- p01-t05: Document coding standards referencing this directory layout
- p02-t01: Implement TokenKind enum in `compiler/lexer/` package established here
