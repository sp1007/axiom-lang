# p01-t05: Coding Standards

## Purpose
Document the engineering conventions for the AXIOM compiler Go codebase in `docs/CONTRIBUTING.md`. These standards ensure all contributors (human and AI-assisted) produce consistent, maintainable, production-grade code. The document covers error handling, naming conventions, test organization, use of `unsafe`, comment style, and architectural rules specific to this project. Without written standards, every engineer makes local decisions that diverge over time, producing an inconsistent codebase that is expensive to maintain.

## Context
The AXIOM compiler is a long-lived production-grade system designed for 10+ year maintainability. The codebase will be read far more often than it is written. Contributors include systems engineers, compiler engineers, and AI-assisted development sessions. The standards defined here codify the engineering philosophy from CLAUDE.md into concrete, checkable rules. Where possible, standards are enforced automatically (via `golangci-lint`, `gofmt`, `go vet`) rather than relying on human review.

## Inputs
- `go.mod` and `.golangci.yml` from p01-t01
- `CLAUDE.md` (the project operating manual)
- `compiler/diagnostics/diagnostics.go` from p01-t01

## Outputs
- `docs/CONTRIBUTING.md` — comprehensive coding standards document
- `.golangci.yml` — updated with any additional linter rules needed to enforce standards

## Dependencies
- p01-t01: repository-bootstrap — directory structure and initial files must exist

## Subsystems Affected
- All: standards apply to every Go file in the repository
- `docs/`: this document lives here

## Detailed Requirements

1. **Error handling rule — no panic in compiler passes**:
   - Compiler passes (lexer, parser, type checker, IR builder, code generator) MUST NOT call `panic()`.
   - Instead, passes return `[]diagnostics.Diagnostic` or `(*T, []diagnostics.Diagnostic)`.
   - `panic()` is only permitted in:
     - `runtime/` packages (catastrophic runtime failures)
     - Test files (`t.Fatal` preferred, but `panic` allowed for test infrastructure)
     - `internal/assert/` package (invariant checking in debug builds, behind a build tag)
   - Example of CORRECT error handling:
     ```go
     func (p *Parser) parseIdent() (uint32, []diagnostics.Diagnostic) {
         tok := p.peek()
         if tok.Kind != lexer.TokenIdent {
             diag := diagnostics.Diagnostic{
                 Severity: diagnostics.SeverityError,
                 Code:     2001,
                 Pos:      p.pos(tok),
                 Message:  fmt.Sprintf("expected identifier, got %s", tok.Kind),
             }
             return 0, []diagnostics.Diagnostic{diag}
         }
         return p.consume(), nil
     }
     ```
   - Example of INCORRECT (forbidden) pattern:
     ```go
     // WRONG: panics on bad input
     func (p *Parser) parseIdent() uint32 {
         if p.peek().Kind != lexer.TokenIdent {
             panic("expected ident") // FORBIDDEN in compiler passes
         }
         return p.consume()
     }
     ```

2. **Naming conventions**:
   - **Packages**: short, lowercase, single word. `lexer`, `ast`, `sema`, `types`, `air`, `cgen`. No underscores except `panic_`.
   - **Types**: PascalCase. `AstNode`, `TokenKind`, `AirInst`, `TypeInfo`.
   - **Functions/Methods**: PascalCase for exported, camelCase for unexported. `ParseExpr`, `parseIdent`.
   - **Variables**: camelCase. `tok`, `nodeIdx`, `typeID`.
   - **Constants**: Use PascalCase for exported (`NodeFuncDecl`), SCREAMING_SNAKE_CASE only for external ABI constants.
   - **Test functions**: `TestXxxYyy` format. `TestLexerIntLiterals`, `TestParserFuncDecl`.
   - **Benchmark functions**: `BenchmarkXxxYyy` format. `BenchmarkLexerThroughput`.
   - **Fuzz functions**: `FuzzXxx` format. `FuzzLexer`, `FuzzParser`.

3. **File organization within packages**:
   - One primary concept per file. `lexer.go` for the lexer core, `token.go` for Token type, `token_kind.go` for TokenKind enum.
   - Test files in the same package as the code they test: `lexer_test.go` in `package lexer`.
   - Use `_test` suffix for black-box tests: `package lexer_test` in `lexer_integration_test.go`.
   - No file longer than 800 lines. If a file grows beyond this, split by responsibility.

4. **Comment style**:
   - Every exported type and function MUST have a doc comment.
   - Doc comments describe what (not how), starting with the symbol name. `// Parse parses the next expression from the token stream.`
   - Do NOT write obvious comments: `// increment i` above `i++` is forbidden.
   - Use `// NOTE:` for non-obvious behavior. `// FIXME:` for known bugs. `// HACK:` for temporary workarounds that need an issue.
   - Architecture comments belong in the package-level `doc.go` file, not scattered in implementation files.
   - FROZEN structures must have comment `// FROZEN: do not modify without RFC`.

5. **Use of `unsafe` package**:
   - `unsafe` is ONLY permitted in:
     - `runtime/` packages (axalloc, actors, panic_)
     - Size/offset assertion code in struct definition files (the `var _ = [1]struct{}{}[...]` pattern)
     - Explicitly marked files with build tag `//go:build unsafe_allowed`
   - Every use of `unsafe` must have a comment explaining WHY it is safe.
   - `unsafe.Pointer` casts must be accompanied by a comment citing the Go spec rule that makes them valid.
   - Forbidden in: all compiler passes, all IR code, all codegen code (use typed abstractions instead).

6. **Test organization**:
   - Unit tests: same package, in `*_test.go` files.
   - Integration tests: `tests/` subdirectory, compiled as separate test binary.
   - Golden tests: `tests/<subsystem>/*.ax` + `tests/<subsystem>/*.golden` pairs. Update with `-update` flag.
   - Fuzz targets: in the package under test, in `*_fuzz_test.go` files.
   - Benchmark tests: in `benchmarks/` directory.
   - Every test must be deterministic. No `time.Sleep`, no random seeds without `t.Setenv("SEED", ...)`.
   - Test names must be descriptive: `TestLexerHexLiteral_WithUnderscores` not `TestLex3`.

7. **Import organization** (enforced by `goimports`):
   ```go
   import (
       // Standard library
       "fmt"
       "os"

       // Internal packages (github.com/axiom-lang/axiom/...)
       "github.com/axiom-lang/axiom/compiler/lexer"
       "github.com/axiom-lang/axiom/compiler/diagnostics"

       // Third-party (none expected in compiler core)
   )
   ```

8. **Architecture boundary rules**:
   - Frontend packages (`lexer`, `parser`, `ast`) MUST NOT import backend packages (`cgen`, `native`).
   - `compiler/` packages MUST NOT import `ir/` packages directly — only via interfaces.
   - `runtime/` packages MUST NOT import `compiler/` packages.
   - Violations are caught by `depguard` linter rule (add to `.golangci.yml`).
   - Import graph must be a DAG (no cycles). Verified by `go build ./...` (Go itself enforces this).

9. **No global mutable state** (except explicitly documented):
   - Global variables must be `const` or effectively immutable after `init()`.
   - Mutable globals require a comment explaining why they are global and how thread safety is ensured.
   - Compiler passes must be stateless functions or methods on a context struct (e.g., `Parser`, `TypeChecker`) passed explicitly.
   - Forbidden: `var globalTypeTable TypeTable` in a package without explicit justification.

10. **Performance guidelines**:
    - Do NOT prematurely optimize. Correctness first.
    - Hot paths (lexer inner loop, AST traversal) may use unsafe tricks — document them.
    - Benchmark before claiming performance improvements.
    - Never allocate in the lexer inner loop (zero-copy design).
    - Arena/pool allocations in the parser are acceptable and encouraged.

11. **Diagnostic code numbering**:
    - Lexer errors: E0001–E0999
    - Parser errors: E1000–E1999
    - Name resolution errors: E2000–E2999
    - Type errors: E3000–E3999
    - Ownership errors: E4000–E4999
    - IR errors: E5000–E5999
    - Codegen errors: E6000–E6999
    - Warnings: W1000+
    - Notes: N1000+
    - Document each code in `docs/DIAGNOSTICS.md` (future task).

12. **Git commit message format**:
    ```
    <type>(<scope>): <short description>

    <optional body>
    ```
    Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`.
    Scope: subsystem name (`lexer`, `parser`, `sema`, `air`, `cgen`).
    Example: `feat(lexer): add INDENT/DEDENT token generation`.

## Implementation Steps

1. Create `docs/CONTRIBUTING.md` with sections:
   - Getting Started (build, test, lint commands)
   - Directory Structure (reference p01-t01 outputs)
   - Error Handling (Requirement 1 — with code examples)
   - Naming Conventions (Requirement 2)
   - File Organization (Requirement 3)
   - Comment Style (Requirement 4)
   - Use of `unsafe` (Requirement 5)
   - Test Organization (Requirement 6)
   - Import Organization (Requirement 7)
   - Architecture Boundaries (Requirement 8)
   - Global State Policy (Requirement 9)
   - Performance Guidelines (Requirement 10)
   - Diagnostic Code Ranges (Requirement 11)
   - Git Commit Format (Requirement 12)
   - Review Checklist (condensed form of Validation Checklist below)

2. Update `.golangci.yml` to add:
   ```yaml
   linters:
     enable:
       - depguard
       - godot      # doc comments end with period
       - misspell
       - nilerr     # common error handling mistake
       - gocritic
   linters-settings:
     depguard:
       rules:
         frontend-no-backend:
           list-mode: lax
           files:
             - "**/compiler/lexer/**"
             - "**/compiler/parser/**"
             - "**/compiler/ast/**"
           deny:
             - pkg: "github.com/axiom-lang/axiom/codegen"
               desc: "Frontend packages must not import backend packages"
         runtime-isolation:
           list-mode: lax
           files:
             - "**/runtime/**"
           deny:
             - pkg: "github.com/axiom-lang/axiom/compiler"
               desc: "Runtime must not import compiler packages"
   ```

3. Create `internal/assert/assert.go` for debug-build invariant checking:
   ```go
   //go:build debug

   package assert

   import "fmt"

   // Invariant panics if cond is false. Only active in debug builds.
   // Use: assert.Invariant(len(nodes) > 0, "nodes must not be empty")
   func Invariant(cond bool, msg string, args ...any) {
       if !cond {
           panic(fmt.Sprintf("INVARIANT VIOLATION: "+msg, args...))
       }
   }
   ```
   And a no-op release version:
   ```go
   //go:build !debug

   package assert

   // Invariant is a no-op in release builds.
   func Invariant(cond bool, msg string, args ...any) {}
   ```

4. Run `golangci-lint run` after adding new linters — fix any new issues surfaced in existing code.

5. Verify the depguard rules work by temporarily adding a forbidden import and confirming lint fails.

## Test Plan

Coding standards are primarily enforced by automated tooling, not tests. However:

- **`TestCodingStandardsLintPasses`** (conceptual): `golangci-lint run` in CI is the test. It must pass on every commit.
- **`TestNoGlobalMutableState`**: Write a Go analysis tool (future) that uses `go/analysis` to detect package-level `var` declarations of mutable types outside `runtime/`.
- **`TestNoPanicInCompilerPasses`**: Write a `go/analysis` check (future) that reports `panic(...)` calls in compiler packages.
- **Manual review checklist**: The Validation Checklist below is used during code review.
- **New contributor test**: A new contributor should be able to read `docs/CONTRIBUTING.md` and set up their environment without asking anyone for help.

## Validation Checklist
- [ ] `docs/CONTRIBUTING.md` exists and covers all 12 requirement areas
- [ ] Error handling section has both CORRECT and INCORRECT code examples
- [ ] `unsafe` usage policy is clearly documented with allowed locations
- [ ] Diagnostic code ranges are defined for all subsystems
- [ ] Import organization rules documented with example
- [ ] Architecture boundary rules documented
- [ ] `.golangci.yml` updated with `depguard`, `godot`, `misspell`, `nilerr`
- [ ] `internal/assert/` package created with both debug and release builds
- [ ] `golangci-lint run` passes with updated config
- [ ] Git commit message format documented with examples
- [ ] Test naming conventions documented with examples (`TestLexerHexLiteral_WithUnderscores`)
- [ ] "Getting Started" section has `make build`, `make test`, `make lint` commands

## Acceptance Criteria
- A new Go developer can read `docs/CONTRIBUTING.md` and understand all coding requirements without additional context
- `golangci-lint run` enforces at least architecture boundaries (depguard) and doc comment style (godot)
- `internal/assert/` compiles in both debug (`-tags debug`) and release builds
- No existing code violates the documented standards at the time of writing
- The document is under 1000 lines (concise, not exhaustive)

## Definition of Done
- [ ] `docs/CONTRIBUTING.md` committed
- [ ] `.golangci.yml` updated and `golangci-lint run` passes
- [ ] `internal/assert/` package committed with both build tags
- [ ] Reviewed by second engineer
- [ ] CI pipeline (p01-t04) still passes after `.golangci.yml` changes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Depguard rules too strict, breaking legitimate cross-package use | Start with `lax` mode; tune rules as codebase grows; document exceptions |
| Contributors ignore CONTRIBUTING.md | Make standards machine-enforceable where possible; CI is the enforcer |
| Standards become outdated as codebase evolves | Review CONTRIBUTING.md quarterly; update when new patterns emerge |
| `godot` linter too noisy (every comment needs a period) | Disable godot if it generates too many false positives; use `nolint` sparingly |
| `internal/assert` package causes confusion about panic policy | Add explicit note in CONTRIBUTING.md: "assert.Invariant is not a panic in production" |

## Future Follow-up Tasks
- Phase 3+: Write `go/analysis` linter for `panic()` in compiler passes
- Phase 3+: Write `go/analysis` linter for global mutable state
- Phase 4+: Add `docs/DIAGNOSTICS.md` with full diagnostic code catalog
- Phase 9+: `tools/lsp/` conventions added to CONTRIBUTING.md
- Ongoing: Review and update CONTRIBUTING.md as project evolves
