# AXIOM Compiler — Contributing Guide & Coding Standards

This document defines the engineering conventions for the AXIOM compiler Go codebase.
All contributors (human and AI-assisted) must follow these standards to produce
consistent, maintainable, production-grade code.

## Getting Started

```bash
# Build all packages
go build ./...
# or: make build

# Run all tests
go test ./...
# or: make test

# Run linter
golangci-lint run
# or: make lint

# Build the axc binary
go build -o bin/axc ./cmd/axc

# Run fuzz tests (lexer)
make fuzz-lexer
```

## Directory Structure

```
cmd/axc/            — CLI entry point
compiler/
  lexer/            — zero-copy lexer
  parser/           — recursive descent + Pratt expression parser
  ast/              — AST node definitions (flat array, u32 indices)
  sema/             — semantic analysis (name resolution, type checking)
  types/            — type system representation (no dependency on ast)
  diagnostics/      — shared Diagnostic type + formatter
  driver/           — compiler pipeline orchestration
ir/
  air/              — AXIOM Intermediate Representation (SSA)
  builder/          — AIR construction API
  opt/              — optimization passes
codegen/
  cgen/             — C code generation backend (C11)
  native/           — native x86-64/ARM64/RISC-V backend
runtime/
  axalloc/          — memory allocator
  actors/           — actor runtime system
  panic/            — panic handler (package panic_)
tools/
  lsp/              — Language Server Protocol
  pkg/              — package manager
internal/
  assert/           — debug-build invariant checking
std/                — AXIOM stdlib (.ax files, NOT Go)
tests/              — test data (.ax inputs, .golden snapshots)
docs/               — documentation
rfcs/               — RFC documents
benchmarks/         — Go benchmarks
fuzz/               — fuzz corpus
```

## Error Handling

### Rule: No `panic()` in compiler passes

Compiler passes (lexer, parser, type checker, IR builder, code generator) **MUST NOT**
call `panic()`. Instead, passes return `[]diagnostics.Diagnostic` or `(*T, []diagnostics.Diagnostic)`.

**CORRECT:**
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

**FORBIDDEN:**
```go
// WRONG: panics on bad input
func (p *Parser) parseIdent() uint32 {
    if p.peek().Kind != lexer.TokenIdent {
        panic("expected ident") // FORBIDDEN in compiler passes
    }
    return p.consume()
}
```

**`panic()` is ONLY permitted in:**
- `runtime/` packages (catastrophic runtime failures)
- Test files (`t.Fatal` preferred, but `panic` allowed for test infrastructure)
- `internal/assert/` package (invariant checking in debug builds, behind `//go:build debug`)

## Naming Conventions

| Element | Convention | Examples |
|---------|-----------|----------|
| Packages | short, lowercase, single word | `lexer`, `ast`, `sema`, `types`, `air`, `cgen` |
| Types | PascalCase | `AstNode`, `TokenKind`, `AirInst`, `TypeInfo` |
| Exported functions/methods | PascalCase | `ParseExpr`, `TokenizeFile` |
| Unexported functions/methods | camelCase | `parseIdent`, `skipWhitespace` |
| Variables | camelCase | `tok`, `nodeIdx`, `typeID` |
| Exported constants | PascalCase | `NodeFuncDecl`, `SeverityError` |
| External ABI constants | SCREAMING_SNAKE_CASE | (only for external C ABI constants) |
| Test functions | `TestXxxYyy` | `TestLexerIntLiterals`, `TestParserFuncDecl` |
| Benchmark functions | `BenchmarkXxxYyy` | `BenchmarkLexerThroughput` |
| Fuzz functions | `FuzzXxx` | `FuzzLexer`, `FuzzParser` |

Exception: `runtime/panic/` package declares `package panic_` due to Go builtin collision.

## File Organization

- **One primary concept per file.** `lexer.go` for the lexer core, `token.go` for Token type, `token_kind.go` for TokenKind enum.
- **Test files in the same package:** `lexer_test.go` in `package lexer`.
- **Black-box tests:** Use `package lexer_test` in `lexer_integration_test.go`.
- **No file longer than 800 lines.** If a file grows beyond this, split by responsibility.

## Comment Style

- Every exported type and function **MUST** have a doc comment.
- Doc comments describe **what** (not how), starting with the symbol name:
  `// Parse parses the next expression from the token stream.`
- Do NOT write obvious comments: `// increment i` above `i++` is forbidden.
- Use `// NOTE:` for non-obvious behavior.
- Use `// FIXME:` for known bugs.
- Use `// HACK:` for temporary workarounds that need an issue.
- Architecture comments belong in the package-level `doc.go` file.
- FROZEN structures must have comment `// FROZEN: do not modify without RFC`.

## Use of `unsafe` Package

`unsafe` is **ONLY** permitted in:
- `runtime/` packages (`axalloc`, `actors`, `panic_`)
- Size/offset assertion code in struct definition files
- Explicitly marked files with build tag `//go:build unsafe_allowed`

Every use of `unsafe` must have a comment explaining **WHY** it is safe.
`unsafe.Pointer` casts must cite the Go spec rule that makes them valid.

**Forbidden in:** all compiler passes, all IR code, all codegen code.

## Test Organization

| Type | Location | Naming |
|------|----------|--------|
| Unit tests | Same package, `*_test.go` | `TestXxxYyy` |
| Integration tests | `tests/` subdirectory | Separate test binary |
| Golden tests | `tests/<subsystem>/*.ax` + `*.golden` | Update with `-update` flag |
| Fuzz targets | Same package, `*_fuzz_test.go` | `FuzzXxx` |
| Benchmarks | `benchmarks/` directory | `BenchmarkXxxYyy` |

Rules:
- Every test **must** be deterministic. No `time.Sleep`, no random seeds without `t.Setenv`.
- Test names must be descriptive: `TestLexerHexLiteral_WithUnderscores` not `TestLex3`.

## Import Organization

Enforced by `goimports`. Three groups, separated by blank lines:

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

## Architecture Boundaries

- **Frontend** (`lexer`, `parser`, `ast`) MUST NOT import **backend** (`cgen`, `native`).
- `compiler/` packages MUST NOT import `ir/` packages directly — only via interfaces.
- `runtime/` packages MUST NOT import `compiler/` packages.
- Import graph must be a DAG (no cycles). Go enforces this at compile time.

## Global State Policy

- Global variables must be `const` or effectively immutable after `init()`.
- Mutable globals require a comment explaining why and how thread safety is ensured.
- Compiler passes must be stateless functions or methods on a context struct.
- Forbidden: `var globalTypeTable TypeTable` without explicit justification.

## Performance Guidelines

- Do NOT prematurely optimize. **Correctness first.**
- Hot paths (lexer inner loop, AST traversal) may use unsafe — document them.
- Benchmark before claiming performance improvements.
- Never allocate in the lexer inner loop (zero-copy design).
- Arena/pool allocations in the parser are acceptable and encouraged.

## Diagnostic Code Ranges

| Subsystem | Range | Example |
|-----------|-------|---------|
| Lexer errors | E0001–E0999 | E0042: invalid character |
| Parser errors | E1000–E1999 | E1001: expected identifier |
| Name resolution errors | E2000–E2999 | E2001: undefined variable |
| Type errors | E3000–E3999 | E3001: type mismatch |
| Ownership errors | E4000–E4999 | E4001: use after move |
| IR errors | E5000–E5999 | E5001: invalid instruction |
| Codegen errors | E6000–E6999 | E6001: unsupported target |
| Warnings | W1000+ | W1001: unused variable |
| Notes | N1000+ | N1001: declared here |

## Git Commit Format

```
<type>(<scope>): <short description>

<optional body>
```

**Types:** `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`
**Scope:** subsystem name (`lexer`, `parser`, `sema`, `air`, `cgen`)
**Example:** `feat(lexer): add INDENT/DEDENT token generation`

## Review Checklist

Before submitting code, verify:

- [ ] All tests pass: `go test ./...`
- [ ] No linter errors: `golangci-lint run`
- [ ] All exported symbols have doc comments
- [ ] No `panic()` in compiler pass code
- [ ] No `unsafe` outside permitted locations
- [ ] No file exceeds 800 lines
- [ ] No global mutable state without justification
- [ ] Test names are descriptive
- [ ] Architecture boundaries respected (no circular imports)
- [ ] FROZEN structs not modified without RFC
