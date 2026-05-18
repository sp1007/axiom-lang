# p01-t04: CI Pipeline

## Purpose
Set up a GitHub Actions continuous integration pipeline that automatically validates every push and pull request to the repository. The CI must run the full build, test, and lint suite across all three target platforms (Linux, macOS, Windows), catching regressions before they merge. A working CI pipeline is non-negotiable for a production-grade compiler project — it is the automated quality gate that enforces all engineering standards defined in CLAUDE.md.

## Context
The AXIOM compiler is a monorepo Go project targeting Linux, macOS, and Windows. The CI pipeline must be fast (under 5 minutes for a cold run), reliable (no flaky tests), and comprehensive (build + test + lint + smoke test). The smoke test compiles a minimal hello-world `.ax` file using the `axc` binary to verify the end-to-end pipeline works. In early phases, `axc` is a stub, so the smoke test verifies the binary exists and exits cleanly with a known error. As the compiler matures, the smoke test evolves into a real compilation test. Go module caching and lint caching are configured to keep CI fast.

## Inputs
- `.github/workflows/` directory (created here)
- `go.mod` from p01-t01
- `Makefile` from p01-t01
- `.golangci.yml` from p01-t01
- `cmd/axc/main.go` from p01-t01
- `tests/ci/hello.ax` (created here — a minimal AXIOM source file)

## Outputs
- `.github/workflows/ci.yml` — main CI workflow
- `.github/workflows/fuzz.yml` — scheduled fuzz testing workflow (runs nightly)
- `tests/ci/hello.ax` — minimal AXIOM source file for smoke test
- `tests/ci/smoke_test.sh` — shell script for smoke test (also inline in CI)

## Dependencies
- p01-t01: repository-bootstrap — go.mod, Makefile, .golangci.yml must exist

## Subsystems Affected
- All: CI validates every subsystem on every push
- `cmd/axc/`: smoke test runs the built `axc` binary
- `tools/`: CI may run additional tool checks in later phases

## Detailed Requirements

1. **Main CI workflow** at `.github/workflows/ci.yml`:
   - Trigger: `push` to any branch, `pull_request` to `main`
   - Matrix: `os: [ubuntu-latest, windows-latest, macos-latest]`, `go: ['1.22.x']`
   - Jobs: `build`, `test`, `lint`, `smoke` (all run in parallel where possible; `smoke` depends on `build`)

2. **Go version pinning**: Use `actions/setup-go@v5` with `go-version: '1.22.x'` and `cache: true` (automatically caches `$GOPATH/pkg/mod`).

3. **Lint job**: Use `golangci/golangci-lint-action@v4` with `version: v1.57.2` and cache enabled. Only run lint on `ubuntu-latest` (lint is platform-independent for Go).

4. **Build job steps**:
   ```yaml
   - uses: actions/checkout@v4
   - uses: actions/setup-go@v5
     with:
       go-version: '1.22.x'
       cache: true
   - name: Build
     run: go build ./...
   - name: Build axc binary
     run: go build -o bin/axc${{ matrix.os == 'windows-latest' && '.exe' || '' }} ./cmd/axc
   - uses: actions/upload-artifact@v4
     with:
       name: axc-${{ matrix.os }}
       path: bin/axc*
   ```

5. **Test job steps**:
   ```yaml
   - name: Run tests
     run: go test -v -race -timeout 120s ./...
   - name: Upload coverage
     if: matrix.os == 'ubuntu-latest'
     run: go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
   ```
   The `-race` flag enables the Go race detector. All tests must pass with the race detector enabled.

6. **Smoke test**: After building `axc`, run:
   ```yaml
   - name: Smoke test (axc stub)
     run: |
       ./bin/axc dump-ast tests/ci/hello.ax || true
       # In Phase 0-1, axc exits with error (no commands implemented)
       # This test verifies the binary exists and runs without segfault
       ./bin/axc 2>&1 | grep -q "usage:" || echo "WARN: usage message not found"
   ```
   On Windows, adjust paths to use `.\bin\axc.exe`.

7. **Fuzz workflow** at `.github/workflows/fuzz.yml`:
   - Trigger: `schedule: - cron: '0 2 * * *'` (nightly at 2 AM UTC) and `workflow_dispatch`
   - Run on `ubuntu-latest` only
   - Run for 5 minutes: `go test -fuzz=FuzzLexer ./compiler/lexer/ -fuzztime=5m`
   - Once p02-t06 and p03-t09 are done, add fuzz targets for parser

8. **`tests/ci/hello.ax`** content — the simplest valid (future) AXIOM program:
   ```
   import std.io

   fn main():
       std.io.println("hello, world")
   ```

9. **Workflow must not use deprecated actions**: Use `actions/checkout@v4`, `actions/setup-go@v5`, `actions/upload-artifact@v4`.

10. **Caching strategy**:
    - Go module cache: handled automatically by `actions/setup-go@v5` with `cache: true`
    - golangci-lint cache: handled by `golangci/golangci-lint-action@v4`
    - Build cache (`GOCACHE`): also handled by `actions/setup-go@v5`

11. **Failure notifications**: The CI should fail fast — if `go build ./...` fails, don't run tests. Use `needs:` to express job dependencies.

12. **Windows path handling**: On Windows, binary is `axc.exe`. Use the expression `${{ matrix.os == 'windows-latest' && 'axc.exe' || 'axc' }}` to handle this in all steps.

## Implementation Steps

1. Create `.github/workflows/` directory.

2. Write `.github/workflows/ci.yml` with the full workflow:
   ```yaml
   name: CI

   on:
     push:
       branches: ['**']
     pull_request:
       branches: [main]

   jobs:
     build:
       name: Build (${{ matrix.os }})
       runs-on: ${{ matrix.os }}
       strategy:
         matrix:
           os: [ubuntu-latest, windows-latest, macos-latest]
           go: ['1.22.x']
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: ${{ matrix.go }}
             cache: true
         - name: go build
           run: go build ./...
         - name: Build axc
           shell: bash
           run: |
             mkdir -p bin
             go build -o bin/axc$([[ "${{ matrix.os }}" == "windows-latest" ]] && echo ".exe" || echo "") ./cmd/axc
         - uses: actions/upload-artifact@v4
           with:
             name: axc-${{ matrix.os }}
             path: bin/

     test:
       name: Test (${{ matrix.os }})
       runs-on: ${{ matrix.os }}
       needs: build
       strategy:
         matrix:
           os: [ubuntu-latest, windows-latest, macos-latest]
           go: ['1.22.x']
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: ${{ matrix.go }}
             cache: true
         - name: go test
           run: go test -v -race -timeout 120s ./...

     lint:
       name: Lint
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: '1.22.x'
             cache: false
         - uses: golangci/golangci-lint-action@v4
           with:
             version: v1.57.2

     smoke:
       name: Smoke Test (${{ matrix.os }})
       runs-on: ${{ matrix.os }}
       needs: build
       strategy:
         matrix:
           os: [ubuntu-latest, windows-latest, macos-latest]
       steps:
         - uses: actions/checkout@v4
         - uses: actions/download-artifact@v4
           with:
             name: axc-${{ matrix.os }}
             path: bin/
         - name: Mark executable (unix)
           if: matrix.os != 'windows-latest'
           run: chmod +x bin/axc
         - name: Smoke test
           shell: bash
           run: |
             BIN=bin/axc
             [[ "${{ matrix.os }}" == "windows-latest" ]] && BIN=bin/axc.exe
             $BIN 2>&1 || true
             echo "Smoke test: binary runs without segfault"
   ```

3. Write `.github/workflows/fuzz.yml`:
   ```yaml
   name: Fuzz

   on:
     schedule:
       - cron: '0 2 * * *'
     workflow_dispatch:

   jobs:
     fuzz-lexer:
       name: Fuzz Lexer
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: '1.22.x'
             cache: true
         - name: Fuzz lexer
           run: go test -fuzz=FuzzLexer ./compiler/lexer/ -fuzztime=5m
           continue-on-error: true
         - name: Fuzz parser
           run: go test -fuzz=FuzzParser ./compiler/parser/ -fuzztime=5m
           continue-on-error: true
   ```
   Note: The fuzz jobs use `continue-on-error: true` because fuzz findings are expected during active development — they are saved as corpus files, not treated as CI blockers.

4. Create `tests/ci/hello.ax` with the content from Requirement 8.

5. Create `tests/ci/smoke_test.sh` as a reference script for local testing:
   ```bash
   #!/usr/bin/env bash
   set -euo pipefail
   BIN=${1:-./bin/axc}
   echo "Running smoke test with: $BIN"
   $BIN 2>&1 || true
   echo "PASS: binary runs without crash"
   ```

6. Push the workflow files and verify GitHub Actions runs them. Check the Actions tab in the repository.

7. Verify all three matrix builds succeed (ubuntu, windows, macos).

8. Verify the lint job passes.

## Test Plan

CI itself is the test — it runs `go test ./...` on every push. The CI is validated by:

- **Manual trigger**: After pushing the workflow file, manually trigger via GitHub UI (`workflow_dispatch` or push a commit).
- **Status badges**: Add a CI status badge to the repository README.
- **Branch protection**: Enable branch protection on `main` requiring CI to pass before merge.
- **Smoke test validation**: Run `tests/ci/smoke_test.sh ./bin/axc` locally before pushing.
- **Matrix verification**: Check GitHub Actions shows 3 build jobs, 3 test jobs, 1 lint job, 3 smoke jobs = 10 jobs total.

Write a local integration test in `ci/test_ci_local.sh`:
```bash
#!/usr/bin/env bash
set -euo pipefail
echo "=== Local CI simulation ==="
go build ./... && echo "PASS: build"
go test -race ./... && echo "PASS: tests"
golangci-lint run && echo "PASS: lint"
go build -o /tmp/axc-test ./cmd/axc && echo "PASS: axc build"
/tmp/axc-test 2>&1 || true && echo "PASS: smoke"
```

## Validation Checklist
- [ ] `.github/workflows/ci.yml` exists and is valid YAML
- [ ] `.github/workflows/fuzz.yml` exists
- [ ] CI matrix includes ubuntu-latest, windows-latest, macos-latest
- [ ] Go version pinned to 1.22.x
- [ ] `go test -race` is used (race detector enabled)
- [ ] `golangci-lint-action@v4` used with version v1.57.2
- [ ] Smoke test downloads built artifact and runs it
- [ ] `tests/ci/hello.ax` exists with valid stub content
- [ ] Fuzz workflow runs nightly
- [ ] `needs:` dependencies correctly express build → test and build → smoke order
- [ ] Windows binary uses `.exe` extension
- [ ] All jobs pass on GitHub Actions (verify in Actions tab)

## Acceptance Criteria
- All 10 CI jobs (3 build + 3 test + 1 lint + 3 smoke) pass on initial push
- CI completes in under 5 minutes (cold run) / under 3 minutes (cached run)
- `go test -race` passes with zero race conditions detected
- `golangci-lint run` reports zero issues
- Fuzz workflow can be triggered manually and runs for 5 minutes without error
- Branch protection on `main` requires CI to pass

## Definition of Done
- [ ] `.github/workflows/ci.yml` committed and running on GitHub
- [ ] All 10 jobs green on first push
- [ ] Fuzz workflow configured and tested with manual trigger
- [ ] Branch protection rule enabled on main
- [ ] Local CI simulation script tested and committed
- [ ] Status badge added to repository README
- [ ] CI reviewed by second engineer

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Windows path separators break shell commands | Use `shell: bash` on all steps; use bash conditional for .exe extension |
| `golangci-lint` version mismatch breaks CI | Pin exact version `v1.57.2`; update via PR when upgrading |
| Race detector causes flaky tests | Fix all races before merging; never disable `-race` flag |
| Go module download failures in CI | Use `cache: true` in setup-go to cache module downloads |
| Fuzz findings break nightly CI | Use `continue-on-error: true` for fuzz jobs; review findings separately |
| GitHub Actions minutes quota exceeded | Optimize matrix to not run lint on all 3 platforms; only ubuntu for lint |
| Smoke test fails as compiler grows | Update smoke test expected output; keep it minimal |

## Future Follow-up Tasks
- p02-t06: Add `FuzzLexer` target — fuzz workflow activates this
- p03-t09: Add `FuzzParser` target — fuzz workflow activates this
- Phase 3+: Update smoke test to actually compile `tests/ci/hello.ax` when lexer+parser ready
- Phase 4+: Add codegen integration test to CI
- Phase 6+: Add benchmark CI job that posts performance numbers as PR comments
