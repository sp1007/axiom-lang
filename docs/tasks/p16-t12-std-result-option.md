# p16-t12: std.result — Result and Option Utilities

## Purpose
Implement the standard utility methods for `Result[T, E]` and `Option[T]` types — the core error handling and optionality types in AXIOM — including chaining, mapping, and unwrapping operations.

## Context
`Result` and `Option` are built-in sum types in AXIOM (defined in the type system). `std.result` adds the rich utility API: `map`, `flat_map`, `unwrap_or`, `expect`, `ok`, `err`, `?` operator desugaring. These are essential for idiomatic AXIOM code.

## Inputs
- `Result[T, E]` and `Option[T]` built-in type definitions from p05-t03
- `?` operator desugaring from p03/p04

## Outputs
- `stdlib/result/result.ax` — Result[T,E] utility methods
- `stdlib/result/option.ax` — Option[T] utility methods

## Dependencies
- p05-t03: sum-types — built-in Result and Option definitions
- p04-t09: effects-system — propagates errors via `raises` effect

## Detailed Requirements

```axiom
# stdlib/result/result.ax

# Extension methods on Result[T, E]
fn map[T, E, U](self: Result[T, E], f: fn(T) -> U) -> Result[U, E]:
    match self:
        Ok(v)  -> Ok(f(v))
        Err(e) -> Err(e)

fn flat_map[T, E, U](self: Result[T, E], f: fn(T) -> Result[U, E]) -> Result[U, E]:
    match self:
        Ok(v)  -> f(v)
        Err(e) -> Err(e)

fn map_err[T, E, F](self: Result[T, E], f: fn(E) -> F) -> Result[T, F]

fn unwrap(self: Result[T, E]) -> T:  # panics on Err
    match self:
        Ok(v)  -> v
        Err(e) -> panic("unwrap called on Err: {e}")

fn unwrap_or(self: Result[T, E], default: T) -> T
fn unwrap_or_else(self: Result[T, E], f: fn(E) -> T) -> T
fn expect(self: Result[T, E], msg: str) -> T  # panics with msg

fn is_ok(self: Result[T, E]) -> bool
fn is_err(self: Result[T, E]) -> bool
fn ok(self: Result[T, E]) -> Option[T]
fn err(self: Result[T, E]) -> Option[E]

# stdlib/result/option.ax

fn map[T, U](self: Option[T], f: fn(T) -> U) -> Option[U]
fn flat_map[T, U](self: Option[T], f: fn(T) -> Option[U]) -> Option[U]
fn unwrap(self: Option[T]) -> T   # panics on None
fn unwrap_or(self: Option[T], default: T) -> T
fn unwrap_or_else(self: Option[T], f: fn() -> T) -> T
fn expect(self: Option[T], msg: str) -> T
fn is_some(self: Option[T]) -> bool
fn is_none(self: Option[T]) -> bool
fn ok_or[T, E](self: Option[T], err: E) -> Result[T, E]
fn filter[T](self: Option[T], pred: fn(T) -> bool) -> Option[T]
fn or(self: Option[T], other: Option[T]) -> Option[T]
fn zip[T, U](self: Option[T], other: Option[U]) -> Option[(T, U)]
```

`?` operator desugaring (compiler):
```axiom
let val = some_result?
# →
let _r = some_result
match _r:
    Ok(v)  -> v
    Err(e) -> return Err(e)   # requires function returns Result
```

## Implementation Steps

1. Create `stdlib/result/result.ax` with all listed methods.
2. Create `stdlib/result/option.ax` with all listed methods.
3. Wire `?` operator in compiler/parser (p03/p04) to desugar correctly.
4. Write tests for all methods.

## Test Plan
- `TestResultMap`: Ok(1).map(|x| x+1) = Ok(2)
- `TestResultFlatMap`: Ok(1).flat_map(|x| Ok(x*2)) = Ok(2)
- `TestResultUnwrapPanics`: Err("x").unwrap() → panic
- `TestOptionUnwrapOr`: None.unwrap_or(42) = 42
- `TestQuestionMark`: `let v = err_result?` returns early with Err

## Validation Checklist
- [ ] unwrap panic message includes the Err value (Display)
- [ ] ? operator works in functions returning Result and Option
- [ ] All methods preserve the Error type unchanged in map
- [ ] zip: None.zip(Some(1)) = None

## Acceptance Criteria
- AXIOM file-reading code using `?` chains compiles and runs correctly

## Definition of Done
- [ ] All methods implemented
- [ ] `?` operator desugaring verified
- [ ] All tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| `?` in non-Result-returning function → confusing error | Type checker validates enclosing function return type |

## Future Follow-up Tasks
- `try {}` block for `?` scoping without full function propagation
