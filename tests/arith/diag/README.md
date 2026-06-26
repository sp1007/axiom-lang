# Arithmetic diagnostics tests (RFC 0006)

Each `eNN_*.ax` is a snippet whose FIRST comment states the expected outcome:
- `EXPECT-ERROR`: must produce a typecheck error once RFC 0006 conversion policy is
  implemented (today the compiler wrongly accepts these).
- `EXPECT-OK`: must compile (sanity guard against over-strict rules).

Policy (RFC 0006): float->int and f64->f32 require explicit `as`; int->float and
f32->f64 are implicit. bool is not numeric. Integer overflow wraps (see matrix_gen
`wrap_*` value tests). Division/modulo by **constant** zero is a compile error;
by a runtime zero is a runtime trap.

Open (TBD): mixing two different INT types as variables (e.g. `i32 + u32`) — whether
implicit is allowed or requires `as` (BUG#34.4). Not yet encoded here.

Runner (after RFC 0006): for each EXPECT-ERROR file, `axc build` must FAIL; for
EXPECT-OK, must succeed.
