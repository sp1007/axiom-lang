# p03-t08: Parser Golden Tests

## Purpose
Establish a comprehensive golden-file test suite for the parser that validates correct AST structure for representative AXIOM programs. Golden tests provide regression protection — any future change that alters AST output for existing programs will be caught immediately.

## Context
Golden tests work by storing expected output in `.ast` files alongside `.ax` input files. The test framework compiles the `.ax` file, runs the AST printer, and diffs the output against the `.ast` file. A `--update` flag regenerates expected files when intentional changes are made.

## Inputs
- `tests/parser/` directory — test case files (to be created)
- `compiler/ast/printer.go` — AST printer from p03-t03
- `compiler/parser/parser.go` — parser from p03-t04, p03-t05, p03-t06, p03-t07

## Outputs
- `tests/parser/*.ax` — AXIOM source input files (20+ test cases)
- `tests/parser/*.ast` — expected AST text output
- `compiler/parser/golden_test.go` — test runner

## Dependencies
- p03-t03: ast-printer — produces the text to compare
- p03-t06: parser-indentation — needed for block parsing
- p03-t07: parser-error-recovery — needed for error cases

## Subsystems Affected
- Testing infrastructure: establishes golden test pattern used by all future subsystems
- Parser: all parse paths covered

## Detailed Requirements

1. Test runner in `compiler/parser/golden_test.go`:
   ```go
   func TestParserGolden(t *testing.T) {
       entries, _ := os.ReadDir("../../tests/parser")
       for _, e := range entries {
           if !strings.HasSuffix(e.Name(), ".ax") { continue }
           t.Run(e.Name(), func(t *testing.T) {
               runGoldenTest(t, "../../tests/parser/"+e.Name())
           })
       }
   }
   ```
2. `runGoldenTest`: lex + parse + print AST, compare to `.ast` file (or write if `--update`).
3. `--update` flag via `go test -run TestParserGolden -update`.

Required test cases:
- `hello_world.ax` — minimal program with println call
- `fibonacci.ax` — recursive function, if/else, arithmetic
- `struct_basic.ax` — struct with fields and a method
- `interface_basic.ax` — interface with one method
- `for_loop.ax` — for x in list, nested for
- `while_loop.ax` — while condition loop
- `match_basic.ax` — match expression with 3 arms
- `defer_stmt.ax` — defer with function call
- `unsafe_block.ax` — unsafe block with pointer ops
- `generics_fn.ax` — generic function `fn sort[T](list: [T])`
- `generics_struct.ax` — generic struct `struct Stack[T]`
- `sum_type.ax` — `type Result = Ok(i32) | Err(string)`
- `import_selective.ax` — `import std.fs { read, write }`
- `pub_fn.ax` — public function declaration
- `multiline_expr.ax` — chained field access and method calls
- `nested_blocks.ax` — deeply nested if/for/while
- `error_missing_colon.ax` — missing `:` after fn signature (error recovery)
- `error_bad_expr.ax` — invalid expression (error recovery)
- `error_multiple.ax` — 3 syntax errors in one file
- `comptime_run.ax` — `#run` compile-time expression

## Implementation Steps

1. Create `tests/parser/` directory.
2. Write each `.ax` test file with representative AXIOM code.
3. Run `go test ./compiler/parser/ -run TestParserGolden -update` to generate initial `.ast` files.
4. Review generated `.ast` files manually — verify correctness of AST structure.
5. Commit both `.ax` and `.ast` files.
6. From now on, CI will run `go test ./compiler/parser/ -run TestParserGolden` without `--update`.

## Test Plan

The golden tests ARE the test plan. Each `.ax` file tests a specific language construct. Additional unit tests:
- `TestParserGoldenIdempotent`: parse → print → parse → print → verify identical output
- `TestParserGoldenNoErrors`: all non-error test cases produce zero diagnostics

## Validation Checklist

- [ ] All 20 `.ax` files created and syntactically valid
- [ ] All `.ast` golden files generated and manually verified
- [ ] `go test ./compiler/parser/ -run TestParserGolden` passes with no diffs
- [ ] Error test cases produce expected error nodes in AST
- [ ] CI runs golden tests on every PR

## Acceptance Criteria

- Zero diffs between actual and golden output for all 20 test cases
- Error test cases produce error nodes with correct diagnostic messages
- Test suite runs in < 2 seconds

## Definition of Done

- [ ] 20+ test case pairs created in `tests/parser/`
- [ ] Golden test runner implemented
- [ ] All tests pass in CI (ubuntu, windows, macos)
- [ ] README in `tests/parser/` explains how to add new test cases

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| AST printer format changes break all golden tests | Versioned printer format; `--update` makes regeneration easy |
| Platform differences in line endings cause false failures | Normalize to `\n` in test runner |

## Future Follow-up Tasks

- p04-t10: sema golden tests follow the same pattern
- p09-t12: AIR builder golden tests
- p03-t09: fuzz target uses golden test inputs as seed corpus
