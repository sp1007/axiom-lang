# p17-t01: axc fmt — Code Formatter

## Purpose
Implement the AXIOM code formatter (`axc fmt`) that normalizes source code to a canonical style: consistent indentation, spacing, line length, and import ordering — idempotent and deterministic.

## Context
A canonical formatter eliminates style debates and reduces diff noise. `axc fmt` reads AXIOM source, parses to AST, and pretty-prints according to a fixed style (no configuration). Idempotent: running twice produces the same output. Integrated into CI and editor save-hooks.

## Inputs
- AXIOM source files (`.ax`)
- AST from parser (p03)
- Style rules (hard-coded, not configurable)

## Outputs
- `tools/fmt/fmt.go` — formatter implementation
- `axc fmt` subcommand (modifies files in-place or outputs to stdout)

## Dependencies
- p03: parser — produces AST from source
- p03-t03: ast-printer — reuse AST pretty-printer infrastructure

## Subsystems Affected
- Editor integration: LSP server calls `axc fmt` on save
- CI: `axc fmt --check` fails PR if files are not formatted

## Detailed Requirements

Style rules:
1. **Indentation**: 4 spaces (not tabs)
2. **Line length**: soft limit 100, hard limit 120
3. **Blank lines**: 1 blank line between top-level declarations; 0 inside function body (except logical groups)
4. **Import ordering**: stdlib first, then third-party, then local; alphabetical within group
5. **Trailing whitespace**: removed
6. **Trailing newline**: exactly one at end of file
7. **Operator spacing**: `a + b` not `a+b`; no space before `:` in type annotations
8. **Function call**: no space between name and `(`; space after `,`
9. **Comment alignment**: inline comments aligned to column 40 if multiple in same block

```go
type Formatter struct {
    IndentWidth int    // 4
    MaxWidth    int    // 100
}

func (f *Formatter) Format(src []byte) ([]byte, error)
func (f *Formatter) FormatFile(path string) error
func (f *Formatter) Check(path string) (bool, error)  // returns needsFormatting
```

`axc fmt` CLI:
```
axc fmt [--check] [--write] [files...]
  --check: exit 1 if any file needs formatting (CI mode)
  --write: write formatted output back to file (default)
  no args: format all .ax files in current directory tree
```

## Implementation Steps

1. Create `tools/fmt/fmt.go`.
2. Implement Formatter — parse AST, then print with style rules.
3. Implement import group detection and sorting.
4. Implement line-length wrapping for long function signatures.
5. Implement `--check` mode for CI.
6. Add `axc fmt` subcommand to compiler driver.
7. Write idempotency tests.

## Test Plan
- `TestFmtIdempotent`: format already-formatted file → no change
- `TestFmtIndent`: wrong indentation → corrected to 4 spaces
- `TestFmtImports`: unordered imports → sorted correctly
- `TestFmtCheck`: unformatted file → exit code 1 with --check
- `TestFmtWriteback`: --write → file modified in place

## Validation Checklist
- [ ] Idempotent: `fmt(fmt(src)) == fmt(src)`
- [ ] Deterministic: same input → same output always
- [ ] --check exit code usable in CI
- [ ] Preserves all semantic content (no tokens lost)

## Acceptance Criteria
- Entire AXIOM stdlib formatted with `axc fmt` produces no changes

## Definition of Done
- [ ] `tools/fmt/fmt.go` implemented
- [ ] Idempotency test passes on 20 real AXIOM files
- [ ] CI check mode integrated

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Comment placement ambiguous after reformatting | Attach comments to nearest AST node before formatting |
| Line wrapping decisions non-deterministic | Prefer width-first: wrap at first opportunity beyond soft limit |

## Future Follow-up Tasks
- Editor integration via LSP formatting protocol
- `axc fmt --diff` mode showing what would change
