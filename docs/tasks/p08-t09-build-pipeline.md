# p08-t09: Full Build Pipeline (`axc build`)

## Purpose
Implement the complete end-to-end build pipeline in `cmd/axc/build.go`. The pipeline takes an AXIOM source file, runs it through every compiler stage (lex, parse, type check, ownership analysis, C-backend codegen), writes the generated `.c` file to a staging area, invokes GCC/Clang as a subprocess, and produces an executable binary. Error messages from GCC are mapped back to AXIOM source locations.

## Context
The `axc build` command is the primary user-facing entry point for the AXIOM compiler. All previous phases have built individual pipeline stages; this task wires them together into a coherent, user-friendly tool. The pipeline must handle errors at every stage cleanly, producing AXIOM-style diagnostics rather than raw C compiler errors.

## Inputs
- All compiler pipeline stages from previous phases:
  - Lexer (Phase 01/02)
  - Parser (Phase 03)
  - Type checker (Phase 04)
  - Ownership analysis (Phase 06)
  - C-Backend (p08-t01 through p08-t08)
- `runtime/` directory with `ax_runtime.h`, `axalloc/`, `panic/`, `arena.c`
- GCC or Clang available in `PATH`
- CLI argument parsing library (use Go's `flag` package)

## Outputs
- `cmd/axc/build.go` — the build command implementation
- `cmd/axc/main.go` — the CLI entry point
- `cmd/axc/errormap.go` — GCC error line mapper

## Dependencies
- p08-t04 (expression codegen — final stage before GCC)
- p08-t05 (ownership codegen)
- p08-t06 (generational check emission)
- p08-t07 (FFI codegen)
- p08-t08 (arena/unsafe codegen)
- p04-t08 (typed AST — input to codegen)
- p04-t09 (type checker entry point)

## Subsystems Affected
- CLI (primary user interface)
- All compiler stages (orchestrated here)
- Runtime (staged alongside generated C)
- Diagnostics (pipeline errors reported in AXIOM format)

## Detailed Requirements

### CLI Interface
```
axc build <file.ax> [flags]

Flags:
  -o <output>       Output binary path (default: same name as input, no extension)
  -O0               No optimization (default)
  -O1               Basic optimization
  -O2               Standard optimization
  -O3               Aggressive optimization
  --emit-c          Write the generated .c file to the current directory
  --keep-c          Do not delete the temp .c file after compilation
  --backend=c       Use C-Backend (default)
  --backend=native  Use native backend (Phase 11)
  --debug           Emit debug info (AX_DEBUG defined, DWARF)
  --target=<triple> Cross-compilation target (Phase 11)
  -v, --verbose     Print each pipeline stage
  --time            Print timing for each stage
```

### Pipeline Stages
```go
func BuildFile(opts BuildOptions) error {
    // Stage 1: Lex
    tokens, err := lexer.Lex(opts.InputFile)
    if err != nil { return formatDiag(err) }

    // Stage 2: Parse
    flatAST, err := parser.Parse(tokens)
    if err != nil { return formatDiag(err) }

    // Stage 3: Resolve names
    resolved, err := sema.ResolveNames(flatAST)
    if err != nil { return formatDiag(err) }

    // Stage 4: Type check
    typedAST, err := typecheck.Check(resolved)
    if err != nil { return formatDiag(err) }

    // Stage 5: Ownership analysis
    ownerAST, err := ownership.Analyze(typedAST)
    if err != nil { return formatDiag(err) }

    // Stage 6: C-Backend codegen
    cCode, err := cgen.GenerateModule(ownerAST)
    if err != nil { return formatDiag(err) }

    // Stage 7: Write .c to temp dir (or current dir if --emit-c)
    cPath, err := writeCFile(opts, cCode)
    if err != nil { return err }
    if !opts.KeepC { defer os.Remove(cPath) }

    // Stage 8: Stage runtime headers
    runtimeDir, cleanup, err := stageRuntime(opts)
    if err != nil { return err }
    defer cleanup()

    // Stage 9: Invoke GCC
    if err := invokeGCC(opts, cPath, runtimeDir); err != nil {
        return err  // GCC errors already mapped and formatted
    }

    return nil
}
```

### Runtime Staging
Copy the `runtime/` directory tree to a temp location. The GCC invocation includes `-I<runtimeDir>` so that `#include "ax_runtime.h"` resolves.

```go
func stageRuntime(opts BuildOptions) (string, func(), error) {
    tmpDir, err := os.MkdirTemp("", "axc-runtime-")
    if err != nil { return "", nil, err }

    // Copy runtime/ax_runtime.h and subdirs
    err = copyDir(opts.RuntimeDir, tmpDir)
    cleanup := func() { os.RemoveAll(tmpDir) }
    return tmpDir, cleanup, err
}
```

The `RuntimeDir` defaults to the directory containing the `axc` binary plus `/runtime`. It can be overridden via the `AXC_RUNTIME` environment variable.

### GCC Invocation
```go
func invokeGCC(opts BuildOptions, cPath, runtimeDir string) error {
    args := []string{
        cPath,
        "-o", opts.Output,
        "-I" + runtimeDir,
        "-std=c11",
        "-Wall",
        opts.OptLevel,  // -O0, -O1, -O2, or -O3
        "-rdynamic",    // for stack traces
    }
    if opts.Debug {
        args = append(args, "-g", "-DAX_DEBUG")
    }
    // Link runtime object files
    args = append(args,
        filepath.Join(runtimeDir, "axalloc/axalloc.c"),
        filepath.Join(runtimeDir, "axalloc/arena.c"),
        filepath.Join(runtimeDir, "panic/panic.c"),
    )

    cmd := exec.Command("gcc", args...)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        // Map GCC error lines back to AXIOM source
        return mapGCCErrors(stderr.String(), cPath, opts.InputFile)
    }
    return nil
}
```

### GCC Error Mapping
GCC errors reference lines in the generated `.c` file. The `errormap.go` module maintains a `LineMap` that maps `.c` line numbers to AXIOM `.ax` source locations. The C-Backend must emit `#line` directives into the generated C code:

```c
#line 42 "main.ax"
ax_i32 result = ax_module_compute(x);
```

The `mapGCCErrors` function:
1. Parses GCC stderr (format: `file.c:42:8: error: message`)
2. Looks up `.c` line 42 in the `LineMap` → AXIOM source location
3. Reformats in AXIOM diagnostic format:
```
error[GCC-E001]: incompatible types
 --> main.ax:15:3
  |
15 | let result: i32 = "hello"
  |             ^^^
```

### `#line` Directive Emission
The `StmtGen` (p08-t03) must emit `#line` directives before each statement:
```go
func (g *StmtGen) EmitStmt(stmt ast.TypedStmt) {
    if g.opts.EmitLineDirectives {
        loc := stmt.SourceLocation()
        g.w.Line(fmt.Sprintf(`#line %d "%s"`, loc.Line, loc.File))
    }
    // ... emit the statement ...
}
```

### Output Binary Name
Default: strip the `.ax` extension and use the resulting name. If no extension, append nothing. Example: `main.ax` → `main`, `src/hello.ax` → `src/hello`.

### Verbose and Timing Mode
With `--verbose`, print each stage name to stderr before running it:
```
[axc] Lexing main.ax...
[axc] Parsing...
[axc] Type checking...
[axc] Ownership analysis...
[axc] C codegen...
[axc] Compiling with gcc...
[axc] Done: ./main
```

With `--time`, append the duration:
```
[axc] Lexing main.ax... 1.2ms
```

## Implementation Steps

### Step 1: Create `cmd/axc/` directory
```
cmd/
  axc/
    main.go         -- CLI entry point
    build.go        -- BuildFile and pipeline
    errormap.go     -- GCC error mapper
    flags.go        -- CLI flag definitions
```

### Step 2: Implement `main.go`
```go
package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }
    switch os.Args[1] {
    case "build":   runBuild(os.Args[2:])
    case "check":   runCheck(os.Args[2:])
    case "emit-c":  runEmitC(os.Args[2:])
    case "dump-air":runDumpAir(os.Args[2:])  // Phase 09
    default:
        fmt.Fprintf(os.Stderr, "axc: unknown command %q\n", os.Args[1])
        os.Exit(1)
    }
}
```

### Step 3: Implement `runBuild` with full pipeline

### Step 4: Implement `mapGCCErrors` in `errormap.go`
Parse GCC stderr line by line using regex `^(.*):(\d+):(\d+): (error|warning): (.*)$`.

### Step 5: Write integration test in `cmd/axc/build_test.go`
Compile a minimal AXIOM program (`let x: i32 = 1 + 1`) and verify the binary is produced and exits 0.

## Test Plan
1. Compile `hello.ax` (prints "hello") → binary produces correct output
2. Compile with `-O2` → binary is smaller/faster than `-O0` (heuristic check)
3. `--emit-c` → `.c` file appears in current directory
4. `--keep-c` → `.c` file in temp dir is not deleted
5. Type error in source → AXIOM-style diagnostic with correct source location
6. GCC error (impossible type) → mapped to AXIOM source line
7. `-o my_output` → binary named `my_output`
8. Default output name: `main.ax` → `main`

## Validation Checklist
- [ ] Pipeline runs all stages in order
- [ ] Runtime headers are staged before GCC invocation
- [ ] `--emit-c` writes to current directory
- [ ] GCC errors are mapped to AXIOM source locations
- [ ] `#line` directives are emitted in generated C
- [ ] Temp directory is cleaned up even if GCC fails
- [ ] All integration tests pass

## Acceptance Criteria
- `axc build hello.ax` produces a runnable binary on Linux, macOS, and Windows
- Compiler errors show AXIOM source locations, not C file line numbers
- The build pipeline handles all error cases without panicking

## Definition of Done
- `cmd/axc/build.go` exists with the full pipeline
- `cmd/axc/main.go` exists with the command dispatcher
- `cmd/axc/errormap.go` exists with GCC error mapping
- Integration test passes on at least Linux
- `go build ./cmd/axc/` succeeds

## Risks & Mitigations
- **Risk**: GCC not in PATH on some systems. **Mitigation**: Check for `clang` as fallback; print a clear error if neither is found with instructions to install.
- **Risk**: `#line` directive mapping breaks if the C-Backend inserts extra lines that are not tracked. **Mitigation**: Only emit `#line` directives for statements and declarations, not for generated boilerplate.
- **Risk**: Runtime staging is slow for large runtime directories. **Mitigation**: In subsequent builds, check modification times and skip staging if runtime is unchanged (implement in p08-t11 benchmark task).

## Future Follow-up Tasks
- p08-t10: E2E compliance tests use `axc build` as the driver
- p08-t12: `axc emit-c` and `axc check` commands added alongside `build`
- p08-t11: Compilation benchmark uses `axc build` with `--time`
- p11-t15: `--backend=native` routes to the native backend
