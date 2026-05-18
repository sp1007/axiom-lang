# p16-t11: std.fmt — Formatting and Display

## Purpose
Implement the `std.fmt` module providing the `Display` and `Debug` interfaces, format string processing, and the `print!`/`println!` macro equivalents using AXIOM's string interpolation.

## Context
AXIOM uses `"text {expr}"` string interpolation which desugars to `std.fmt` calls. The `Display` interface defines how values appear in user-facing output; `Debug` provides structured debug output. Types implementing `Display` can be used in string interpolation.

## Inputs
- AXIOM string interpolation syntax from p03 (parser)
- `std.string.StringBuilder` from p16-t02
- TypedAST from p04 — Display/Debug interface implementations

## Outputs
- `stdlib/fmt/fmt.ax` — Display, Debug interfaces + format function
- `stdlib/fmt/impls.ax` — built-in implementations for all primitives

## Dependencies
- p16-t02: std-string — StringBuilder
- p04-t04: name-resolver — resolves Display.to_str() call in interpolation
- p05-t04: structural-duck-typing — Display satisfied implicitly

## Detailed Requirements

```axiom
# stdlib/fmt/fmt.ax

interface Display:
    fn to_str(self) -> str

interface Debug:
    fn debug_str(self) -> str

fn format(template: str, args: []DisplayVal) -> str
# Note: actual interpolation is compile-time desugared, not runtime dispatch

# stdlib/fmt/impls.ax

impl Display for i32: fn to_str(self) -> str { ... }
impl Display for i64: fn to_str(self) -> str { ... }
impl Display for f32: fn to_str(self) -> str { ... }
impl Display for f64: fn to_str(self) -> str { ... }
impl Display for bool: fn to_str(self) -> str { "true" or "false" }
impl Display for str:  fn to_str(self) -> str { self }

impl Debug for i32: fn debug_str(self) -> str { "i32({self})" }
impl Debug for bool: fn debug_str(self) -> str { "bool({self})" }
impl Debug for str: fn debug_str(self) -> str { "str(\"{self}\")" }

# Array display
impl[T: Display] Display for Array[T]:
    fn to_str(self) -> str:
        var sb = StringBuilder.new()
        sb.write("[")
        for i, val in self.iter().enumerate():
            if i > 0: sb.write(", ")
            sb.write(val.to_str())
        sb.write("]")
        sb.to_str()

# Option display
impl[T: Display] Display for Option[T]:
    fn to_str(self) -> str:
        match self:
            Some(v) -> "Some({v})"
            None    -> "None"
```

String interpolation desugaring (compiler-side):
```axiom
"hello {name}, you are {age} years old"
# →
var _sb = StringBuilder.new()
_sb.write("hello ")
_sb.write(name.to_str())    # requires name: Display
_sb.write(", you are ")
_sb.write(age.to_str())     # requires age: Display
_sb.write(" years old")
_sb.to_str()
```

## Implementation Steps

1. Create `stdlib/fmt/fmt.ax` — Display and Debug interfaces.
2. Create `stdlib/fmt/impls.ax` — impl Display for all primitives.
3. Implement i32/i64 → str via itoa algorithm.
4. Implement f64 → str via Grisu2 or sprintf-based algorithm.
5. Implement generic Display for Array[T], Option[T], Result[T,E].
6. Wire compiler string interpolation to StringBuilder + Display.to_str().
7. Write tests for all Display implementations.

## Test Plan
- `TestDisplayI32`: `42.to_str()` = "42"; `-1.to_str()` = "-1"
- `TestDisplayF64`: `3.14.to_str()` = "3.14" (6 significant digits)
- `TestDisplayBool`: `true.to_str()` = "true"
- `TestDisplayArray`: `[1,2,3].to_str()` = "[1, 2, 3]"
- `TestInterpolation`: `"x = {x}"` produces "x = 42" for i32 x=42

## Validation Checklist
- [ ] All primitive types implement Display
- [ ] Float formatting matches printf %g behavior
- [ ] Nested generic Display works (Array[Option[i32]])
- [ ] Interpolation with missing Display → compile-time error (not runtime)

## Acceptance Criteria
- `println("pi ≈ {std.math.PI}")` outputs "pi ≈ 3.141592653589793"

## Definition of Done
- [ ] `stdlib/fmt/impls.ax` with all primitive Display impls
- [ ] Interpolation desugaring wired in compiler

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Float formatting rounding differences | Use Grisu2 or Ryu algorithm for shortest round-trip representation |

## Future Follow-up Tasks
- `{x:08b}` format specifiers (padding, radix, precision)
- `{x:.3}` float precision specifier
