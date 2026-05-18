# p17-t10: axc fix — Automated Migration Tool

## Purpose
Implement `axc fix` — an automated code migration tool that applies mechanical transformations to AXIOM source code when the language evolves, ensuring user code is updated to new idioms and APIs automatically.

## Context
As AXIOM evolves, APIs change and syntax improves. `axc fix` applies registered "fix passes" that transform outdated patterns to current idioms — like `go fix` or Rust's `rustfix`. Fixes are deterministic, safe (no semantic changes), and idempotent.

## Inputs
- AXIOM source files
- Fix registry: list of registered fix passes with version ranges
- `axiom.toml` specifying minimum AXIOM version
- AST from parser (p03)

## Outputs
- `tools/fix/fix.go` — fix pass framework and driver
- `axc fix [--dry-run] [--fix=<name>]` subcommand
- `fixes/` directory with individual fix pass implementations

## Dependencies
- p03: parser — produces editable AST
- p17-t01: axc-fmt — apply formatting after fixes
- p04-t10: sema-golden-tests — fixes validated by existing golden tests

## Detailed Requirements

Fix pass interface:
```go
type FixPass struct {
    Name        string
    Description string
    Since       string  // AXIOM version this fix applies from
    Until       string  // AXIOM version where this fix is no longer needed
    Apply       func(file *AstFile) []Edit
}

type Edit struct {
    Start   SourcePos
    End     SourcePos
    Replace string
}

var RegisteredFixes = []FixPass{
    // Example fixes:
    {
        Name: "update-print-syntax",
        Description: "Update print() to println() for statements",
        Since: "0.5.0",
        Apply: func(file *AstFile) []Edit { ... },
    },
    {
        Name: "option-unwrap-to-expect",
        Description: "Add message to bare .unwrap() calls",
        Since: "0.8.0",
        Apply: func(file *AstFile) []Edit { ... },
    },
}
```

```
axc fix [options] [files...]
  --dry-run      Show changes without applying
  --fix=<name>   Apply only specific fix
  --all          Apply all applicable fixes
  --list         List all available fixes
```

Fix application flow:
1. Parse source to AST.
2. For each applicable fix: collect edits (non-overlapping).
3. Sort edits by position (reverse order to preserve offsets).
4. Apply edits to source bytes.
5. Run `axc fmt` on modified files.
6. Write back to files (or print diff in --dry-run mode).

## Implementation Steps

1. Create `tools/fix/fix.go` with FixPass framework.
2. Implement edit collection and non-overlapping validation.
3. Implement reverse-order edit application.
4. Implement `--dry-run` diff output.
5. Register at least 3 initial fix passes.
6. Add `axc fix` subcommand.
7. Test: apply fix → file unchanged on second run (idempotent).

## Test Plan
- `TestFixIdempotent`: apply fix twice → no change on second apply
- `TestFixDryRun`: --dry-run → source unchanged, diff printed
- `TestFixSpecific`: --fix=<name> → only named fix applied
- `TestFixList`: --list → all registered fixes shown
- `TestFixRealMigration`: old-style source → fixed to new style correctly

## Validation Checklist
- [ ] Fixes idempotent: applying twice produces same result as once
- [ ] Non-overlapping edits: no two fixes modify same byte range
- [ ] Format applied after fixes (clean output)
- [ ] --dry-run never modifies files

## Acceptance Criteria
- `axc fix --all` updates all Phase 1 example programs to current idioms without breaking them

## Definition of Done
- [ ] `tools/fix/fix.go` implemented
- [ ] 3 fix passes registered and tested
- [ ] Idempotency test passes

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Overlapping edits from different fixes corrupt file | Validate no overlap before applying; error if conflict |
| Fix changes semantics (not safe) | Each fix requires test proving semantic equivalence |

## Future Follow-up Tasks
- Fix passes for generics syntax changes
- Automatic deprecation warning → fix suggestion integration with LSP
