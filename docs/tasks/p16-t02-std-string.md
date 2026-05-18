# p16-t02: std.string — String Operations

## Purpose
Implement the AXIOM standard string library providing UTF-8 string manipulation, formatting, splitting, searching, and conversion operations.

## Context
AXIOM strings are immutable UTF-8 byte sequences with a length prefix. `std.string` provides the fundamental operations programs need: concatenation, slicing, searching, formatting, and conversion to/from numeric types. The implementation must be allocation-efficient, preferring stack allocation for small strings.

## Inputs
- AXIOM `str` built-in type (stored as `{ptr: *u8, len: u32}`)
- Runtime allocator from p14 (for heap-allocated string operations)
- UTF-8 encoding rules

## Outputs
- `stdlib/string/string.ax` — string operations
- `stdlib/string/utf8.ax` — UTF-8 validation and iteration
- `stdlib/string/format.ax` — string formatting (interpolation backend)

## Dependencies
- p04-t02: type-table — TypeString definition
- p14-t01: axalloc — for dynamic string allocation

## Subsystems Affected
- String interpolation: `"hello {name}"` desugars to `std.string.format()`
- Diagnostics: error messages use string formatting

## Detailed Requirements

```axiom
# stdlib/string/string.ax

fn len(s: str) -> u32
fn bytes(s: str) -> []u8   # raw bytes (UTF-8)
fn concat(a: str, b: str) -> str
fn slice(s: str, start: u32, end: u32) -> str
fn contains(s: str, sub: str) -> bool
fn starts_with(s: str, prefix: str) -> bool
fn ends_with(s: str, suffix: str) -> bool
fn index_of(s: str, sub: str) -> Option[u32]
fn split(s: str, sep: str) -> []str
fn trim(s: str) -> str
fn trim_start(s: str) -> str
fn trim_end(s: str) -> str
fn to_upper(s: str) -> str
fn to_lower(s: str) -> str
fn repeat(s: str, n: u32) -> str
fn replace(s: str, old: str, new: str) -> str
fn to_i64(s: str) -> Result[i64, str]
fn to_f64(s: str) -> Result[f64, str]
fn from_i64(n: i64) -> str
fn from_f64(f: f64, precision: u8 = 6) -> str

# StringBuilder for efficient concatenation
type StringBuilder:
    var buf: []u8

    fn write(mut self, s: str)
    fn write_char(mut self, c: u32)
    fn to_str(self) -> str
```

String interpolation desugaring:
```axiom
"hello {name}!"
# → std.string.concat("hello ", std.string.concat(to_str(name), "!"))
# Or via StringBuilder for efficiency
```

UTF-8 operations:
```axiom
fn char_count(s: str) -> u32        # Unicode code point count
fn char_at(s: str, idx: u32) -> u32 # code point at position idx
fn is_valid_utf8(s: str) -> bool
```

## Implementation Steps

1. Create `stdlib/string/string.ax`.
2. Implement all basic string operations using AXIOM.
3. Where performance requires, use C shims via `extern`.
4. Create `stdlib/string/format.ax` — `StringBuilder` implementation.
5. Create `stdlib/string/utf8.ax` — UTF-8 iteration.
6. Wire string interpolation in codegen to `StringBuilder`.
7. Write tests for all functions.

## Test Plan
- `TestConcat`: "hello" + " " + "world" = "hello world"
- `TestSlice`: "hello"[1:4] = "ell"
- `TestSplit`: "a,b,c".split(",") = ["a","b","c"]
- `TestToI64`: "42".to_i64() = Ok(42); "abc".to_i64() = Err(...)
- `TestUTF8CharCount`: "héllo" char count = 5
- `TestStringBuilder`: write 1000 strings → correct result

## Validation Checklist
- [ ] All operations handle empty string without panic
- [ ] Slice bounds checked (panic on out-of-range)
- [ ] UTF-8 validity checked in char_at
- [ ] to_i64/to_f64 return Err on invalid input (not panic)

## Acceptance Criteria
- String interpolation in hello-world program works correctly

## Definition of Done
- [ ] `stdlib/string/string.ax` implemented
- [ ] All tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Unicode slice boundaries — slice mid-codepoint | slice() operates on byte indices; char-based slice requires char_at iteration |

## Future Follow-up Tasks
- Regular expressions (stdlib/regex)
- String hashing (for HashMap keys)
