# p16-t19: Stdlib Compliance Tests

## Purpose
Validate that all standard library modules meet their specifications through a comprehensive compliance test suite — ensuring correctness, edge cases, API stability, and cross-platform behavior.

## Context
Individual stdlib modules have unit tests, but compliance tests verify the complete public API as a contracted surface. These tests run against the stdlib API specification and must pass on all supported platforms (Linux, macOS, Windows). They also serve as regression protection when stdlib internals change.

## Inputs
- All stdlib modules from p16-t01 through p16-t18
- Target platforms: Linux (ELF), macOS (Mach-O), Windows (PE-COFF)
- API specification from each module's task file

## Outputs
- `tests/stdlib/compliance_test.go` — Go test wrapper calling AXIOM tests
- `tests/stdlib/*.ax` — AXIOM compliance tests per module
- CI jobs per platform

## Dependencies
- All p16-t01 through p16-t18: stdlib modules under test
- p16-t01: std.testing — assert functions used by compliance tests

## Subsystems Affected
- CI: compliance tests run on every PR, on all platforms

## Detailed Requirements

One compliance test file per module:

```
tests/stdlib/test_testing.ax     # std.testing compliance
tests/stdlib/test_string.ax      # std.string compliance
tests/stdlib/test_collections.ax # std.collections compliance
tests/stdlib/test_io.ax          # std.io compliance
tests/stdlib/test_math.ax        # std.math compliance
tests/stdlib/test_net.ax         # std.net compliance (localhost only)
tests/stdlib/test_process.ax     # std.process compliance
tests/stdlib/test_sync.ax        # std.sync compliance
tests/stdlib/test_json.ax        # std.json compliance
tests/stdlib/test_time.ax        # std.time compliance
tests/stdlib/test_fmt.ax         # std.fmt compliance
tests/stdlib/test_result.ax      # std.result + std.option
tests/stdlib/test_log.ax         # std.log compliance
tests/stdlib/test_os.ax          # std.os compliance
tests/stdlib/test_iter.ax        # std.iter compliance
tests/stdlib/test_random.ax      # std.random compliance
tests/stdlib/test_cli.ax         # std.cli compliance
tests/stdlib/test_ffi.ax         # std.ffi compliance
```

Each test file follows the pattern:
```axiom
import std.testing
import std.<module>

#[test]
fn test_<api_function>_<case>():
    # Tests every documented behavior and edge case
    assert_eq(expected, actual)
```

Coverage requirements:
- Every public function tested with: happy path, error path, empty/zero input, boundary values.
- Cross-module: `std.json.parse()` on a `std.io.File` read result.
- Platform-specific: test file path separator on each platform.

CI matrix:
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
```

## Implementation Steps

1. Create `tests/stdlib/` directory.
2. Write compliance test for each module (18 files).
3. Create Go test wrapper `tests/stdlib/compliance_test.go` that compiles + runs each .ax test file.
4. Add edge cases not covered by unit tests.
5. Add cross-module integration tests.
6. Add CI matrix job for all three platforms.
7. Measure and report coverage.

## Test Plan
- 18 compliance test files, each covering 100% of public API surface
- Cross-platform validation on Linux, macOS, Windows

## Validation Checklist
- [ ] All 18 stdlib modules have compliance tests
- [ ] Every public function appears in at least one test
- [ ] Tests pass on Linux, macOS, Windows
- [ ] Edge cases: empty string, zero values, max values, None/Err paths

## Acceptance Criteria
- All compliance tests pass on all three platforms in CI

## Definition of Done
- [ ] 18 compliance test files created
- [ ] CI matrix runs all three platforms
- [ ] All tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Windows-specific failures (path separators, CRLF) | Add platform guards in os-specific tests |
| Network tests flaky in CI | Use localhost only; disable external network in CI |

## Future Follow-up Tasks
- Fuzz testing for string parsing functions
- Performance compliance: assert functions meet latency SLAs
