# p13-t07: Multi-Target Cross-Compilation Integration Tests

## Purpose
Validate end-to-end cross-compilation for all three native backends (x86-64, ARM64, RISC-V 64) by building real AXIOM programs targeting each platform and verifying correctness using QEMU user-mode emulation on Linux CI.

## Context
Cross-compilation allows developers to build AXIOM programs on one host (e.g., Linux x86-64) for a different target (e.g., ARM64 or RISC-V). This is essential for embedded workflows, CI/CD pipelines, and ensuring the compiler produces correct code for targets that may not have a native CI runner.

QEMU user-mode emulation (`qemu-user-static`) allows running ARM64 and RISC-V binaries on an x86-64 Linux host without full system emulation. The binary runs as if it were on the target architecture, using the host OS kernel for syscalls (via architecture translation).

This task integrates the three native backends and validates them against the compliance test suite (or a subset thereof) using QEMU.

## Inputs
- x86-64 native backend (p11-t01 through p11-t16)
- ARM64 native backend (p13-t01 through p13-t04)
- RISC-V 64 native backend (p13-t05, p13-t06)
- ELF linker (p11-t16)
- Mach-O linker (p12-t03, p13-t04)
- Test programs in `tests/compliance/` (subset, at least 10 programs)

## Outputs
- `ci/cross-compile-test.yml` — GitHub Actions workflow for cross-compile CI
- `scripts/cross-compile-test.sh` — local test runner script
- `tests/cross_compile_test.go` — Go integration tests for cross-compilation
- Updated `compiler/driver/` — `--target` flag fully wired to all three backends

## Dependencies
- p13-t04: ARM64 Mach-O integration (ARM64 backend complete)
- p13-t06: RISC-V ABI (RISC-V backend complete)
- p11-t16: ELF linker (produces runnable ELF binaries for Linux targets)

## Subsystems Affected
- `compiler/driver/` — `--target` flag handling for all three architectures
- `ci/` — new CI workflow
- `scripts/` — new test script

## Detailed Requirements

### Target Triples Supported
| Target Triple | Binary Format | Test Method |
|---|---|---|
| `x86_64-linux-gnu` | ELF64 x86-64 | Run natively on CI |
| `aarch64-linux-gnu` | ELF64 ARM64 | Run via `qemu-aarch64-static` |
| `aarch64-apple-macos13` | Mach-O ARM64 | Run on macos-14 GitHub Actions runner |
| `riscv64-linux-gnu` | ELF64 RISC-V 64 | Run via `qemu-riscv64-static` |

### Cross-Compile Command Interface
```bash
# Linux x86-64 target
axc build --target=x86_64-linux-gnu main.ax -o main_x86

# ARM64 Linux target
axc build --target=aarch64-linux-gnu main.ax -o main_arm64

# ARM64 macOS target
axc build --target=aarch64-apple-macos13 main.ax -o main_macos_arm64

# RISC-V 64 Linux target
axc build --target=riscv64-linux-gnu main.ax -o main_riscv64
```

### Test Programs for Cross-Compile Validation
The following programs must compile and run correctly on all targets:

1. `tests/cross/hello.ax` — `println("Hello, World!")` — tests basic output
2. `tests/cross/arith.ax` — integer arithmetic (add, sub, mul, div, rem)
3. `tests/cross/float.ax` — floating-point operations (basic f64 math)
4. `tests/cross/branches.ax` — if/else, loops, early return
5. `tests/cross/functions.ax` — multiple function calls, recursion (fibonacci)
6. `tests/cross/structs.ax` — struct creation, field access, passing structs
7. `tests/cross/strings.ax` — basic string operations
8. `tests/cross/memory.ax` — heap allocation, pointer operations
9. `tests/cross/stdlib_io.ax` — `std.io.println`, `std.io.readline`
10. `tests/cross/stdlib_math.ax` — `std.math.sqrt`, `std.math.pow`

For each program, define expected stdout output in a `.expected` file.

### QEMU User-Mode Setup
On Ubuntu CI:
```bash
sudo apt-get install -y qemu-user-static binfmt-support
sudo update-binfmts --enable qemu-aarch64
sudo update-binfmts --enable qemu-riscv64
```

After this, ARM64 and RISC-V ELF binaries can be run directly:
```bash
qemu-aarch64-static ./main_arm64    # or just ./main_arm64 via binfmt
qemu-riscv64-static ./main_riscv64
```

### Correctness Validation
For each test program × target combination:
1. Compile with `axc build --target=<triple> <program.ax> -o <output>`
2. Run the output binary (natively or via QEMU)
3. Compare stdout with `<program>.expected`
4. Compare exit code with expected exit code
5. Verify runtime is ≤ 5 seconds (to detect infinite loops)

```go
func TestCrossCompile(t *testing.T) {
    programs := listCrossTestPrograms()
    targets := []Target{
        {Triple: "x86_64-linux-gnu",      Runner: ""},
        {Triple: "aarch64-linux-gnu",     Runner: "qemu-aarch64-static"},
        {Triple: "riscv64-linux-gnu",     Runner: "qemu-riscv64-static"},
    }
    for _, prog := range programs {
        for _, target := range targets {
            t.Run(prog.Name+"/"+target.Triple, func(t *testing.T) {
                binary := compile(t, prog.Path, target.Triple)
                output := run(t, binary, target.Runner, 5*time.Second)
                require.Equal(t, prog.Expected, output)
            })
        }
    }
}
```

### CI Workflow
Create `.github/workflows/cross-compile.yml`:
```yaml
name: Cross-Compile Tests

on: [push, pull_request]

jobs:
  cross-compile-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - name: Install QEMU
        run: |
          sudo apt-get update
          sudo apt-get install -y qemu-user-static binfmt-support
      - name: Build axc
        run: go build ./cmd/axc
      - name: Run cross-compile tests
        run: go test ./tests/... -run TestCrossCompile -v -timeout 300s

  cross-compile-macos-arm64:
    runs-on: macos-14
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - name: Build axc
        run: go build ./cmd/axc
      - name: Test ARM64 Mach-O
        run: go test ./tests/... -run TestCrossMachO -v -timeout 120s
```

### Target Triple Parser
In `compiler/driver/`:
```go
type TargetTriple struct {
    Arch    string  // x86_64, aarch64, riscv64
    OS      string  // linux, apple
    ABI     string  // gnu, macos13
}

func ParseTriple(s string) (TargetTriple, error) {
    parts := strings.Split(s, "-")
    if len(parts) < 3 {
        return TargetTriple{}, fmt.Errorf("invalid target triple: %q", s)
    }
    return TargetTriple{Arch: parts[0], OS: parts[1], ABI: parts[2]}, nil
}

func (t TargetTriple) Backend() (Backend, error) {
    switch t.Arch {
    case "x86_64":  return x86.NewBackend(t), nil
    case "aarch64": return arm64.NewBackend(t), nil
    case "riscv64": return riscv64.NewBackend(t), nil
    default:
        return nil, fmt.Errorf("unsupported architecture: %s", t.Arch)
    }
}
```

### Differential Testing
For programs that can run on both x86-64 (natively) and RISC-V (via QEMU), compare outputs:
```go
func TestDifferential(t *testing.T) {
    for _, prog := range crossTestPrograms {
        nativeOut := runNative(t, prog, "x86_64-linux-gnu")
        arm64Out  := runQEMU(t,   prog, "aarch64-linux-gnu", "qemu-aarch64-static")
        riscvOut  := runQEMU(t,   prog, "riscv64-linux-gnu",  "qemu-riscv64-static")
        
        require.Equal(t, nativeOut, arm64Out,  "ARM64 output differs from x86-64")
        require.Equal(t, nativeOut, riscvOut,  "RISC-V output differs from x86-64")
    }
}
```

## Implementation Steps

### Step 1: Wire --target Flag
In `cmd/axc/main.go` and `compiler/driver/driver.go`:
```go
flag.StringVar(&config.Target, "target", hostTriple(), "target triple")
// ...
backend, err := ParseTriple(config.Target).Backend()
if err != nil { log.Fatal(err) }
```

### Step 2: Write Cross-Compile Test Programs
Create `tests/cross/` directory with 10 test `.ax` files and corresponding `.expected` files.

### Step 3: Write Integration Test
Create `tests/cross_compile_test.go` with the test structure above.

### Step 4: Set Up CI Workflow
Create `.github/workflows/cross-compile.yml` as specified.

### Step 5: Local Test Script
Create `scripts/cross-compile-test.sh`:
```bash
#!/usr/bin/env bash
set -e

AXCC=./axc
PROGRAMS=(tests/cross/*.ax)
TARGETS=("x86_64-linux-gnu" "aarch64-linux-gnu" "riscv64-linux-gnu")
RUNNERS=("" "qemu-aarch64-static" "qemu-riscv64-static")

for prog in "${PROGRAMS[@]}"; do
    for i in "${!TARGETS[@]}"; do
        target=${TARGETS[$i]}
        runner=${RUNNERS[$i]}
        out_bin=$(mktemp /tmp/axc_XXXXXX)
        $AXCC build --target=$target "$prog" -o "$out_bin"
        if [ -n "$runner" ]; then
            actual=$($runner "$out_bin")
        else
            actual=$("$out_bin")
        fi
        expected=$(cat "${prog%.ax}.expected")
        if [ "$actual" != "$expected" ]; then
            echo "FAIL: $prog on $target"
            echo "  Expected: $expected"
            echo "  Got:      $actual"
            exit 1
        fi
        echo "PASS: $prog on $target"
        rm -f "$out_bin"
    done
done
```

## Test Plan

### Compilation Tests
For each program × target:
1. Compilation succeeds with exit code 0
2. Output binary is valid ELF64 (for Linux targets) or Mach-O (for macOS target)
3. Binary runs to completion (no SIGILL, SIGSEGV, SIGBUS)

### Output Correctness Tests
1. `hello.ax` prints `Hello, World!\n` on all targets
2. `arith.ax` prints correct arithmetic results (e.g., `7 * 6 = 42`)
3. `float.ax` prints `sqrt(2.0) = 1.414213562...`
4. `functions.ax` prints `fib(10) = 55`

### Error Handling Tests
1. `axc build --target=invalid-triple main.ax` → error message, exit 1
2. `axc build --target=sparc64-linux-gnu main.ax` → "unsupported architecture: sparc64", exit 1

### Performance Tests
1. Compilation time: each test program compiles in < 5 seconds
2. QEMU execution: each test binary runs in < 2 seconds under QEMU

## Validation Checklist
- [ ] `--target` flag parsed and wired to correct backend
- [ ] All 10 test programs compile for all 3 Linux targets
- [ ] QEMU emulation works for ARM64 and RISC-V
- [ ] Differential tests pass (same output across targets for deterministic programs)
- [ ] CI workflow passes on `ubuntu-latest` and `macos-14`
- [ ] Error messages are clear for invalid targets

## Acceptance Criteria
1. All 10 × 3 = 30 compilation + execution combinations pass
2. CI passes on both `ubuntu-latest` and `macos-14` GitHub Actions runners
3. `axc build --target=<any_supported_triple>` produces a runnable binary
4. Differential tests confirm byte-identical stdout across architectures for deterministic programs
5. Invalid target triple gives a clear, actionable error message

## Definition of Done
- `cross_compile_test.go`, CI workflow, and test programs implemented
- All 30 target/program combinations pass in CI
- Local test script runs successfully on developer machines
- `README` for `tests/cross/` explains how to add new cross-compile tests

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| QEMU version differences across Ubuntu releases | Pin QEMU version in CI; test on ubuntu-22.04 and ubuntu-24.04 |
| binfmt_misc not available in some CI environments | Use explicit `qemu-aarch64-static ./binary` instead of relying on binfmt |
| RISC-V backend not complete when this task starts | Run RISC-V tests as optional (`-short` flag skips them) until p13-t06 is done |
| macOS-14 runner availability | macos-14 is Apple Silicon; fall back to macos-13 (x86-64) if unavailable |

## Future Follow-up Tasks
- Extend cross-compile tests to include all 100 compliance tests (once p16 is complete)
- Windows cross-compile target: `x86_64-windows-msvc` (Phase 17+)
- Android target: `aarch64-linux-android` (future)
