# RFC 0038 — The `print` / `println` builtin contract

- Status: accepted
- Author: AXIOM compiler team
- Created: 2026-08-05
- Affects: frontend (`typecheck`, `air_builder`); no IR opcode, ABI, linker or backend change
- Supersedes/extends: nothing (no RFC has ever covered `print`)

---

## 1. Motivation

`print` and `println` are the only builtins a user meets on line one of their first
program, and they were the least specified thing in the language. `resolver.ax:484-485`
declares them as `SYM_BUILTIN_TYPE` with no signature; `typecheck.ax` contained no
occurrence of the strings `"print"` / `"println"` at all; and the entire behaviour lived
in ONE branch of the x86 instruction selector (`x86_selector.ax:1737-1776`), which
rewrote the call target to a **one-parameter** runtime shim
(`ax_{print,println}_{str,i64,f64,bool}`).

Three defects followed directly from that, all measured (repros in `bin/probe9/`):

**P1 — every argument after the first is silently dropped.**
`println("val: ", s)` prints `val: ` and stops. The AIR is *correct* — it carries all
N arguments — but the selector reads only `extras.data[arg_start+1]`, retargets to a
1-parameter shim, and never adjusts `arg_count`, so arguments 2..N are marshalled into
RDX/R8/... where nothing reads them. This is the accept-then-miscompile class (BUG#53):
the program compiles clean and produces the wrong output.

**P2 — `println()` with no arguments segfaults.** The `arg_count == 0` branch selects
`ax_println_str` with no argument at all, so the runtime dereferences whatever is in RCX
(exit 139).

**P3 — the type classification has a fall-through hole.** `x86_selector.ax:1742` maps
type ids 1..8/15/16 to int, 11 to bool, 9/10 to float, and **everything else to `str`**,
which is then dereferenced as a string pointer. Measured: `char8` (id 13) segfaults;
a struct, a `ptr[T]` or a `void` value prints a blank line and exits 0.

P3 is arity-independent and pre-existing, but fixing P1 **multiplies exposure to it**: an
argument in position 2 is harmlessly dropped today and becomes a real call tomorrow. The
two must therefore ship together.

Finally, `std/fmt.ax:47,52` already declares

```
pub fn print(args: ...)
pub fn println(args: ...)
```

so a variadic `print` is the signature the standard library has been advertising all
along. This RFC does not widen the language; it makes the implementation honour a
contract the stdlib already published.

---

## 2. The contract

1. **Arity is `>= 0`.** `print()` / `println()` are legal; there is no upper bound.
2. **All arguments are evaluated, left to right, BEFORE any output is produced.**
   This is printf's rule and it is observable: an argument whose evaluation itself
   prints (or panics) must not interleave with the output of an earlier argument.
3. **No separator is inserted between arguments.** `print("a", "b")` writes `ab`.
   Formatting each value is unchanged from the single-argument case
   (`i64` -> decimal, `bool` -> `true`/`false`, `f64` -> 6 fractional digits, `str`/
   `bytes` -> raw bytes).
4. **Only the FINAL call of a `println` emits the newline.** `println(a, b, c)` writes
   `abc\n`, not `a\nb\nc\n`.
5. **`println()` is exactly `println("")`** — it writes a single newline.
6. **`print()` writes nothing** and produces no code.
7. **Permitted argument types** are exactly
   `{i8, i16, i32, i64, u8, u16, u32 (rune), u64, isize, usize, bool, f32, f64, str,
   bytes}`.
   Every other type — struct, sum/enum, `Option`/`Result`, tuple, array, `ptr[T]`,
   interface, `char8`, `void` — is a **compile error, `error[E3033]`** (§4).
8. The builtins produce **no value**. `let x = print(...)` is not meaningful and is not
   supported.

`print`/`println` remain BUILTINS, not ordinary functions. A user-defined
`fn println(...)` is a different symbol and this contract does not apply to it. (The
selector's separate defect of hijacking such a function by NAME — "P4" — is out of scope
here and is tracked separately; it needs the runtime-symbol choice moved out of the
selector, which is a backend change.)

---

## 3. Design: desugar in AIR, not in the selector

The rewrite is performed in `air_builder.lower_call_expr`
(`bootstrap/stage1/air_builder.ax`), at the point where the argument registers have all
been lowered:

```
print(a, b, c)    =>  print(a);  print(b);  print(c)
println(a, b, c)  =>  print(a);  print(b);  println(c)
println()         =>  println("")
print()           =>  <nothing>
```

Three reasons for AIR rather than the selector:

- **CLAUDE.md §4 (stage isolation).** An instruction selector selects instructions for
  the IR it is given; synthesizing N calls out of one is a *lowering* decision, and
  putting it in the selector is exactly the "backend does frontend work" coupling the
  manual forbids.
- **One fix, three backends.** `x86_selector`, `cgen` (C backend) and the wasm backend
  each consume the same AIR. Desugaring in AIR fixes all three; patching the selector
  would fix one and leave the other two silently dropping arguments.
- **The AIR was never wrong.** The dump of `bin/probe9/pn1.ax` shows `extras[0]=2`
  followed by both argument vregs. The information loss is entirely in the selector, so
  the repair belongs at or above it, not deeper.

### 3.1 Evaluation order

The existing argument-lowering loop is left in place and runs to completion FIRST; only
then are the N calls emitted. This is deliberate and is the reason the desugar is not
written as a source-level rewrite to `print(a); print(b)`: that form would evaluate `b`
*after* `a` had already been printed, changing observable interleaving for effectful
arguments. Keeping the loop also keeps `coerce_float_arg` / `coerce_interface_arg`
running on every argument exactly as before.

### 3.2 Fixpoint safety

`temp_count == 1` takes the **pre-existing path unchanged**, so the AIR emitted for every
single-argument call is byte-identical to before this RFC. An exhaustive scan of all 1202
`*.ax` files in the repository found **zero** `print`/`println` call sites with two or
more arguments in `bootstrap/stage1/`, `std/`, `stdlib/` or `tests/`; the compiler's own
uses are `print_helpers.ax:75` and `:80`, both a single `str`, and it never calls a bare
`println()`. The compiler's self-image therefore cannot change: the required gate is
**A == B**.

### 3.3 Symbol identity, not spelling

The non-final calls of a desugared `println` must target the **`print` builtin symbol**.
That symbol is *looked up* — a scan of the symbol table for
`name_id == intern("print") and kind == SYM_BUILTIN_TYPE`, which terminates immediately
because builtins occupy the first ~40 slots — and is **not** derived as `println_sym - 1`,
even though `resolver.ax:484-485` happens to declare the two adjacently. Deriving a
symbol index from another symbol's index is the same "match the writing, not the
identity" defect class recorded in RFC 0037. If no builtin `print` is found, the desugar
declines and the old path runs.

The desugar is likewise gated on `kind == SYM_BUILTIN_TYPE`, not on the callee's name
text, so a user-defined `fn println(a: str, b: i64)` is never desugared.

### 3.4 The synthesized empty string

`println()` synthesizes an `OP_ICONST` of string type whose payload is the interned text
`""` **including the two quote characters**. `lower_string_lit` interns the RAW TOKEN
text (quotes included) and the backends strip them with `unescape_string_literal`
(`print_helpers.ax:24-37`). The nearby `assert` desugar interns UNQUOTED text; that
survives only because `unescape_string_literal` no-ops when the first character is not a
quote, and it must not be copied blindly here — an unquoted 0-length payload would leave
`unescape_string_literal` in its `s_len < 2` branch.

---

## 4. Diagnostic `error[E3033]`

New type-checker diagnostic, in the E3xxx (type/semantic) band alongside E3030..E3032:

```
error[E3033]: cannot print a value of type `Point`
  |
  |     println("p = ", p)
  |                     ^ only primitive values can be printed
  |
   = note: `print`/`println` accept i8..u64, isize, usize, bool, f32, f64, str and bytes
   = help: print a field (`p.x`), convert with `as u32`/`as i64`, or add a `to_str()` method and print its result
```

Emitted from `typecheck.ax` at the `NODE_CALL_EXPR` site, after the arguments have been
inferred, for every argument of a call whose callee is the `print`/`println` **builtin
symbol**. An argument whose type is still `TYPE_UNKNOWN` is never diagnosed
(conservative: never a false reject). `diags_count` is bumped, so the driver halts before
codegen — no object file is produced.

This is the first arity/type checking these builtins have ever had.

---

## 5. Alternatives considered

**(a) Reject calls with more than one argument (an arity error).** Simplest, and it
converts a silent miscompile into a diagnostic, which satisfies the BUG#53 policy. It was
rejected because `std/fmt.ax:47,52` *already declares these builtins variadic*: the user
intent is not ambiguous, the stdlib has published the signature, and the manual's rule is
to reject only when intent is unclear. Rejecting would also make the newly-adopted
stdout-oracle test convention (`println("label: ", value)`) unwritable.

**(b) Insert a separator (space) between arguments, like Python's `print`.** Rejected:
it makes `print("a", "b")` non-concatenative and would break the natural
`println("label: ", v)` idiom, which wants no space beyond the one the user wrote.

**(c) Fix the selector to emit N calls.** Rejected on layering (CLAUDE.md §4) and because
it would have to be re-implemented identically in `cgen` and the wasm backend — two more
copies of one rule, which is the defect shape this project has paid for repeatedly.

**(d) A real variadic ABI for the builtins.** Rejected as vastly out of proportion: it
would need a calling-convention change (an ABI change, §13 territory of its own) to solve
a problem that N sequential 1-argument calls solve exactly, with identical output.

**(e) Auto-`to_str()` for aggregates instead of E3033.** Rejected for now: it needs a
`Display`-style contract and a decision on how a user opts in. E3033 keeps the door open
— it is a rejection, not a semantic commitment — and can be relaxed later without
breaking any program that compiles today.

---

## 6. Drawbacks

- N arguments cost N calls instead of one. `print` is I/O-bound; the cost is
  irrelevant next to the write syscall, and it is exactly what the single-argument
  form already costs per value.
- Output is no longer atomic per statement: another thread writing to stdout could
  interleave between the calls of one `println(a, b)`. This is already true between two
  statements, and AXIOM does not promise atomic stdout.
- E3033 turns some programs that compile today into compile errors. All such programs
  were already broken (blank output, or a segfault for `char8`); none of them printed
  anything meaningful.

---

## 7. Migration

- Single-argument `print`/`println`: **no change whatsoever** — identical AIR, identical
  machine code.
- Multi-argument calls: previously dropped every argument after the first; now all are
  printed. Any program relying on the old behaviour was relying on a bug.
- `println()`: previously segfaulted; now prints a newline.
- Printing an aggregate / `ptr` / `void` / `char8`: previously a blank line or a
  segfault; now `error[E3033]` with a `help:` naming the three ways out.

## 8. Oracles

- `bin/t_printvariadic.ax` — mixed types in ONE call (`str`, `i64`, `bool`, `f64`),
  adjacent `print` calls, a trailing `println`, and a non-ASCII UTF-8 label
  (`Kết quả: `) whose bytes are pinned. The mixed-type row is load-bearing: it fails if
  anyone "fixes" this by applying the FIRST argument's type to all arguments.
- `bin/t_printbarenewline.ax` — bare `println()` between two `print`s (exit 139 pre-fix).
- `bin/t_printbadtype.ax` — printing a struct: rejected (E3033).
- `bin/t_printchar8.ax` — printing a `char8`: rejected (E3033).
