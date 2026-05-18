# p08-t12: `axc emit-c` and `axc check` Commands

## Purpose
Implement two additional `axc` subcommands: `axc emit-c <file.ax>` which outputs the generated C code to stdout (or a file), and `axc check <file.ax>` which runs the compiler pipeline through type checking and ownership analysis, reporting any errors without producing a binary. These are essential developer tools for debugging the C-Backend output and for fast syntax/type checking in editors.

## Context
`axc emit-c` is the primary debugging tool for the C-Backend: it lets engineers inspect exactly what C code is generated for any AXIOM program without running GCC. `axc check` provides fast feedback during development — it is the basis for editor integration (the language server can call `axc check` on save to surface errors instantly).

Both commands must be available immediately after the build pipeline (p08-t09) is complete, as they are simpler subsets of the full build pipeline.

## Inputs
- `cmd/axc/build.go` (p08-t09) — pipeline stages to reuse
- `cmd/axc/main.go` — the command dispatcher to extend

## Outputs
- `cmd/axc/emitc.go` — `axc emit-c` implementation
- `cmd/axc/check.go` — `axc check` implementation
- Updated `cmd/axc/main.go` — adds new commands to the dispatcher

## Dependencies
- p08-t09 (build pipeline — stages reused here)

## Subsystems Affected
- CLI (new user-facing commands)
- C-Backend (emit-c exposes its output directly)
- Diagnostics (check surfaces all errors in AXIOM format)

## Detailed Requirements

### `axc emit-c <file.ax>`
Runs the full pipeline through C codegen, then writes the generated C to stdout (or to a file with `-o`).

```
Usage: axc emit-c <file.ax> [flags]

Flags:
  -o <output.c>     Write to file instead of stdout (default: stdout)
  --no-line-dirs    Suppress #line directives in the output
  --pretty          Format the C output with clang-format (if available)
  --backend=c       C-Backend (default; only option in Phase 08)
```

Pipeline stages run:
1. Lex
2. Parse
3. Name resolution
4. Type checking
5. Ownership analysis
6. C codegen
7. Output C to stdout or file

Does NOT invoke GCC.

Example usage:
```bash
$ axc emit-c main.ax
#include "ax_runtime.h"

struct ax_Point {
    ax_i32 x;
    ax_i32 y;
};

ax_i32 ax_main_distance(struct ax_Point* a, struct ax_Point* b);

ax_i32 ax_main_distance(struct ax_Point* a, struct ax_Point* b) {
    ...
}
```

If the input has errors, `axc emit-c` exits with code 1 and prints diagnostics to stderr (same format as `axc build`).

### `axc check <file.ax>`
Runs the pipeline through type checking and ownership analysis. Reports all errors and warnings. Does NOT run codegen or GCC.

```
Usage: axc check <file.ax> [flags]

Flags:
  --json            Output diagnostics as JSON (for editor integration)
  --no-color        Disable ANSI color codes in diagnostics
  --max-errors N    Stop after N errors (default: 50)
```

Pipeline stages run:
1. Lex
2. Parse
3. Name resolution
4. Type checking
5. Ownership analysis

Exit codes:
- 0: No errors (warnings may be present)
- 1: One or more errors

Example (error case):
```bash
$ axc check bad.ax
error[E0042]: type mismatch
 --> bad.ax:5:9
  |
5 | let x: i32 = "hello"
  |         ^^^ expected i32, found string

1 error found.
$ echo $?
1
```

Example (clean):
```bash
$ axc check good.ax
No errors. (3 warnings)
$ echo $?
0
```

### JSON Output Format for `--json`
```json
{
  "file": "bad.ax",
  "diagnostics": [
    {
      "code": "E0042",
      "severity": "error",
      "message": "type mismatch: expected i32, found string",
      "location": {
        "file": "bad.ax",
        "line": 5,
        "column": 9,
        "end_line": 5,
        "end_column": 14
      },
      "source_text": "let x: i32 = \"hello\"",
      "hint": "try: let x: string = \"hello\""
    }
  ],
  "summary": {
    "errors": 1,
    "warnings": 0
  }
}
```

### Color Output
When stdout is a terminal (`isatty`), use ANSI color codes:
- Error code: bold red
- Arrow `-->`: cyan
- Bar `|`: cyan
- Underline `^^^`: red
- Hint text: green

When stdout is not a terminal (piped to a file or another program), suppress colors (unless `--color` is explicitly set).

### `axc dump-air` Placeholder
While not implemented until p09-t11, add a stub in `main.go`:
```go
case "dump-air":
    fmt.Fprintln(os.Stderr, "axc dump-air: not yet implemented (Phase 09)")
    os.Exit(1)
```

## Implementation Steps

### Step 1: Implement `cmd/axc/emitc.go`
```go
package main

import (
    "flag"
    "fmt"
    "io"
    "os"
    "axiom/codegen/cgen"
    "axiom/compiler/..."
)

func runEmitC(args []string) {
    fs := flag.NewFlagSet("emit-c", flag.ExitOnError)
    output := fs.String("o", "", "output file (default: stdout)")
    noLineDirs := fs.Bool("no-line-dirs", false, "suppress #line directives")
    _ = fs.Bool("pretty", false, "format with clang-format")
    fs.Parse(args)

    if fs.NArg() < 1 {
        fmt.Fprintln(os.Stderr, "usage: axc emit-c <file.ax>")
        os.Exit(1)
    }
    inputFile := fs.Arg(0)

    // Run pipeline through codegen
    cCode, err := runPipelineToC(inputFile, !*noLineDirs)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

    // Output
    var w io.Writer = os.Stdout
    if *output != "" {
        f, err := os.Create(*output)
        if err != nil { fmt.Fprintf(os.Stderr, "create: %v\n", err); os.Exit(1) }
        defer f.Close()
        w = f
    }
    fmt.Fprint(w, cCode)
}
```

### Step 2: Implement `cmd/axc/check.go`
```go
func runCheck(args []string) {
    fs := flag.NewFlagSet("check", flag.ExitOnError)
    jsonOut := fs.Bool("json", false, "JSON output")
    maxErrors := fs.Int("max-errors", 50, "maximum errors to report")
    fs.Parse(args)

    if fs.NArg() < 1 {
        fmt.Fprintln(os.Stderr, "usage: axc check <file.ax>")
        os.Exit(1)
    }
    inputFile := fs.Arg(0)

    diags, err := runPipelineToOwnership(inputFile, *maxErrors)
    if err != nil {
        // internal compiler error
        fmt.Fprintf(os.Stderr, "internal error: %v\n", err)
        os.Exit(2)
    }

    if *jsonOut {
        printDiagsJSON(diags, inputFile)
    } else {
        printDiagsHuman(diags)
    }

    errorCount := countErrors(diags)
    if errorCount > 0 {
        os.Exit(1)
    }
    fmt.Fprintf(os.Stderr, "No errors. (%d warnings)\n", countWarnings(diags))
}
```

### Step 3: Update `cmd/axc/main.go`
Add `"emit-c"` and `"check"` to the command dispatcher.

### Step 4: Add `isatty` check
Use `golang.org/x/term` package's `term.IsTerminal(int(os.Stdout.Fd()))`.

### Step 5: Write `cmd/axc/emitc_test.go` and `check_test.go`
Test both commands with a known-good and known-bad AXIOM file.

## Test Plan
1. `axc emit-c hello.ax` → C code with `#include "ax_runtime.h"` first line
2. `axc emit-c --no-line-dirs hello.ax` → no `#line` directives in output
3. `axc emit-c -o out.c hello.ax` → file `out.c` created
4. `axc emit-c bad.ax` → exits 1, diagnostics on stderr
5. `axc check hello.ax` → exits 0
6. `axc check bad.ax` → exits 1, errors on stderr
7. `axc check --json bad.ax` → valid JSON with correct structure
8. `axc check --max-errors 2 many_errors.ax` → only 2 errors reported, then stops

## Validation Checklist
- [ ] `axc emit-c` output can be passed to GCC and compiled
- [ ] `axc check` exits 0 for clean files, 1 for files with errors
- [ ] JSON output is valid JSON
- [ ] Color codes suppressed when output is not a terminal
- [ ] `axc dump-air` stub prints a helpful message
- [ ] All tests pass

## Acceptance Criteria
- `axc emit-c main.ax | gcc -x c - -o main -I runtime/ runtime/axalloc/axalloc.c runtime/panic/panic.c` produces a working binary
- `axc check` exits 0 for 30 compliance test files
- JSON output contains all required fields

## Definition of Done
- `cmd/axc/emitc.go` exists and works
- `cmd/axc/check.go` exists and works
- `cmd/axc/main.go` updated with new commands
- Tests pass for both commands

## Risks & Mitigations
- **Risk**: `axc emit-c` produces C that GCC rejects due to a codegen bug. **Mitigation**: This is expected during development. The command is a debugging tool; the fact it surfaces the broken C is valuable.
- **Risk**: `isatty` check differs across platforms. **Mitigation**: Use `golang.org/x/term` which handles Linux, macOS, and Windows correctly.

## Future Follow-up Tasks
- p09-t11: `axc dump-air` stub becomes a real command
- Future: `axc fmt <file.ax>` code formatter command
- Future: `axc lsp` language server that wraps `axc check --json` for editor integration
