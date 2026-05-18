# p18-t05: Triple-Build Verification

## Purpose
Implement the triple-build verification script and test harness that proves the self-hosted AXIOM compiler is deterministic: compiling the compiler source three times in succession must produce bit-identical binaries. This is the definitive proof of self-hosting correctness and determinism.

## Context
Plan §7 (Self-Hosting Strategy) specifies:
```bash
./axc_stage2 build ./bootstrap/stage1/compiler/ -o axc_stage2b
./verify/triple_build.sh axiom_compliance_suite.ax  # 3× identical hashes
```

The triple-build test is the **final acceptance gate** for v1.0.0. It proves:
1. The compiler is deterministic (no map iteration order, no timestamps)
2. The compiler is self-consistent (compiling itself produces a functionally equivalent binary)
3. No untested code paths exist in the self-hosted compiler

## Inputs
- Self-hosted compiler binary from p18-t04 (`axc_self`)
- All 19 compliance test suites
- SHA-256 hashing utility

## Outputs
- `verify/triple_build.sh` — main verification script
- `verify/compare_binaries.sh` — binary hash comparison
- `verify/verify_compliance.sh` — compliance suite runner
- `verify/README.md` — documentation of the verification process
- CI job definition for automated triple-build

## Dependencies
- p18-t04: stage4-full-compiler — self-hosted compiler exists
- p11-t15: native-backend-integration — native binary output (for hash comparison)
- p08-t10: e2e-compliance-tests — compliance test infrastructure

## Subsystems Affected
- Build system: new `make verify-triple` target
- CI: new verification job
- Release process: triple-build is a release gate

## Detailed Requirements

### 1. Triple-Build Script

```bash
#!/usr/bin/env bash
# verify/triple_build.sh — Triple-build verification for self-hosting
set -euo pipefail

COMPILER_SRC="./bootstrap/stage1/compiler/"
AXC_STAGE0="$1"  # path to Go-compiled axc (or previous stage)

echo "=== AXIOM Triple-Build Verification ==="
echo "Stage 0 compiler: $AXC_STAGE0"

# Build 1: Stage 0 compiles compiler source → axc_build1
echo "[1/3] Building compiler (stage 0 → build 1)..."
$AXC_STAGE0 build $COMPILER_SRC -o axc_build1 --deterministic
HASH1=$(sha256sum axc_build1 | awk '{print $1}')
echo "  Hash: $HASH1"

# Build 2: Build 1 compiles compiler source → axc_build2
echo "[2/3] Building compiler (build 1 → build 2)..."
./axc_build1 build $COMPILER_SRC -o axc_build2 --deterministic
HASH2=$(sha256sum axc_build2 | awk '{print $1}')
echo "  Hash: $HASH2"

# Build 3: Build 2 compiles compiler source → axc_build3
echo "[3/3] Building compiler (build 2 → build 3)..."
./axc_build2 build $COMPILER_SRC -o axc_build3 --deterministic
HASH3=$(sha256sum axc_build3 | awk '{print $1}')
echo "  Hash: $HASH3"

# Verify: build2 == build3 (fixed point reached)
echo ""
echo "=== Results ==="
echo "Build 1: $HASH1"
echo "Build 2: $HASH2"
echo "Build 3: $HASH3"

if [ "$HASH2" = "$HASH3" ]; then
    echo "✅ PASS: Fixed point reached (build2 == build3)"
else
    echo "❌ FAIL: Builds 2 and 3 differ!"
    echo "  This indicates non-determinism in the compiler."
    exit 1
fi

# Optional: check build1 == build2 (indicates stage0 == self-hosted)
if [ "$HASH1" = "$HASH2" ]; then
    echo "✅ BONUS: Stage 0 produces identical output to self-hosted"
else
    echo "ℹ️  NOTE: Stage 0 output differs from self-hosted (expected during bootstrap)"
fi

# Run compliance suite with the final binary
echo ""
echo "=== Running compliance suite with axc_build3 ==="
./verify/verify_compliance.sh ./axc_build3
```

### 2. Compliance Verification

```bash
#!/usr/bin/env bash
# verify/verify_compliance.sh
AXC="$1"
PASS=0; FAIL=0; TOTAL=0

for suite in tests/axiom_*_suite.ax; do
    TOTAL=$((TOTAL + 1))
    if $AXC build "$suite" -o /tmp/ax_test && /tmp/ax_test; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        echo "FAIL: $suite"
    fi
done

echo "Results: $PASS/$TOTAL passed, $FAIL failed"
[ "$FAIL" -eq 0 ] || exit 1
```

### 3. Determinism Requirements

For triple-build to succeed, the compiler MUST NOT:
- Embed timestamps in output
- Use map iteration order (Go maps are non-deterministic)
- Use goroutine scheduling order in output
- Use random values anywhere

The `--deterministic` flag ensures:
- Symbol table sorted lexicographically before emission
- All parallel operations produce deterministic output via sorted merge
- No `time.Now()` calls in codegen path

### 4. Benchmark Comparison

After triple-build, compare performance:
```bash
# axc_build3 must not be >20% slower than stage 0
time ./axc_build3 build benchmarks/compile_time/bench_1kloc.ax
time $AXC_STAGE0 build benchmarks/compile_time/bench_1kloc.ax
```

## Implementation Steps

1. Create `verify/` directory with `README.md`.
2. Write `verify/triple_build.sh` (see above).
3. Write `verify/verify_compliance.sh`.
4. Write `verify/compare_binaries.sh` — generic binary hash comparison utility.
5. Add `--deterministic` flag to `axc build` (ensures sorted output, no timestamps).
6. Add `make verify-triple` Makefile target.
7. Add CI job in `ci/.github/workflows/verify.yml` (runs on release tags only).
8. Document the verification process in `verify/README.md`.

## Test Plan

- `TestTripleBuildScript`: run the script with a known-deterministic Go axc → verify exit 0
- `TestDeterministicFlag`: compile same file twice with `--deterministic` → identical hashes
- `TestNonDeterministicDetection`: intentionally add `time.Now()` → triple-build fails

## Validation Checklist

- [ ] `verify/triple_build.sh` executes without errors
- [ ] Build 2 and Build 3 produce identical SHA-256 hashes
- [ ] All 19 compliance suites pass with the triple-built binary
- [ ] Performance regression < 20% vs stage 0
- [ ] `--deterministic` flag works correctly

## Acceptance Criteria

- Triple-build produces a fixed point (hash2 == hash3)
- Compliance suite: 100% pass rate with the final binary

## Definition of Done

- [ ] `verify/triple_build.sh` implemented and tested
- [ ] `verify/verify_compliance.sh` implemented
- [ ] `--deterministic` flag added to `axc build`
- [ ] CI job configured for release verification
- [ ] Documentation in `verify/README.md`

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Non-determinism in Go's map iteration | Use sorted keys everywhere; lint rule to catch `range map` without sort |
| Cross-platform hash differences | Normalize line endings; use platform-specific golden hashes |
| Triple-build too slow for CI | Run only on release tags, not on every PR |

## Future Follow-up Tasks

- p18-t06: Runtime self-hosting (AxAlloc in AXIOM) — extends triple-build to cover runtime
- Release pipeline: triple-build is a mandatory gate before tagging v1.0.0
