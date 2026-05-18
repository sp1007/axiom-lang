# p02-t07: `axc dump-tokens` Command

## Purpose
Implement the `axc dump-tokens <file.ax>` CLI command that tokenizes a source file and prints the token stream as JSON to stdout. This is a debugging and testing tool for verifying lexer correctness.

## Context
Plan §3.1 specifies: _"`axc dump-tokens <file.ax>` prints kind names from source positions"_. This command is essential for debugging INDENT/DEDENT balance issues and verifying the lexer produces correct token streams.

## Inputs
- Lexer from p02-t02/t03/t04
- CLI infrastructure from p01-t01
- Token kind string names from p02-t01

## Outputs
- `cmd/axc/dump_tokens.go` — `runDumpTokens()` function
- Updated `cmd/axc/main.go` — `dump-tokens` subcommand
- JSON output format for token stream

## Dependencies
- p02-t05: lexer-golden-tests — lexer verified working
- p01-t01: repository-bootstrap — CLI entry point

## Subsystems Affected
- CLI: new `dump-tokens` subcommand

## Detailed Requirements

### Output Format
```json
[
  {"kind": "FN", "offset": 0, "len": 2, "text": "fn"},
  {"kind": "IDENT", "offset": 3, "len": 3, "text": "foo"},
  {"kind": "LPAREN", "offset": 6, "len": 1, "text": "("},
  {"kind": "INDENT", "offset": 15, "len": 0, "text": ""},
  {"kind": "DEDENT", "offset": 30, "len": 0, "text": ""},
  {"kind": "EOF", "offset": 31, "len": 0, "text": ""}
]
```

### Flags
```
axc dump-tokens <file.ax> [flags]
  --compact     One token per line, no indentation
  --no-text     Omit "text" field (faster for large files)
  --stats       Print token count summary by kind
```

### Stats Mode
```
Token Statistics:
  IDENT:    142
  INT_LIT:   38
  INDENT:    23
  DEDENT:    23
  NEWLINE:   89
  ...
  Total:    512
```

## Implementation Steps

1. Create `cmd/axc/dump_tokens.go`.
2. Add `TokenKind.String()` method to p02-t01's token kinds (if not already present).
3. Lex the input file, iterate tokens, emit JSON array.
4. For `--stats` mode, count tokens by kind and print summary.
5. Update `cmd/axc/main.go` to route `"dump-tokens"`.
6. Write test: known file → expected JSON output.

## Test Plan

- `TestDumpTokensHello`: `fn main(): println("hello")` → correct JSON
- `TestDumpTokensIndent`: file with indentation → INDENT/DEDENT in output
- `TestDumpTokensStats`: `--stats` → summary with correct counts
- `TestDumpTokensEmpty`: empty file → `[{"kind": "EOF", ...}]`

## Validation Checklist

- [ ] All token kinds have String() representation
- [ ] JSON output is valid (parseable by `jq`)
- [ ] INDENT/DEDENT tokens included in output
- [ ] `--stats` mode counts correctly

## Acceptance Criteria

- `axc dump-tokens` produces valid JSON for all 19 test suite files
- Token count matches expected for known test files

## Definition of Done

- [ ] `cmd/axc/dump_tokens.go` implemented
- [ ] `cmd/axc/main.go` updated
- [ ] Tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Large files produce huge JSON output | `--compact` and `--no-text` flags reduce output size |

## Future Follow-up Tasks

- p03-t10: `axc dump-ast` builds on same CLI infrastructure
- p09-t11: `axc dump-air` follows same pattern
