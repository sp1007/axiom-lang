# RFC 0031 — Dead function elimination

Status: **DRAFT** — measured end-to-end (`-dfe-report` shipped, inert); pruning NOT yet activated
Author: autopilot
Related: RFC 0030 (`.bss`), [[feedback-no-exe-bloat]], RFC 0011 (static libs / bundling)

## 1. The measurement

Every AXIOM executable carries the whole bundled stdlib whether or not it is used. Measured
on driver `8D944EF5`:

| program | exe size |
|---|---|
| `fn main() -> i64: return 42` | **77,824** |
| Vec + HashMap + string concat + Option unwrap | 82,432 |

The difference between "uses nothing" and "uses three containers and the string library" is
**4,608 bytes**. So roughly **75 KB is fixed overhead present in every executable**, and for
the trivial program essentially all of it is unreachable.

Section breakdown of the trivial program: `.text` 75,264 bytes (**96.7%** of the file),
`.idata` 1,536. Codegen reports **272 symbols / 74,118 bytes of code** for `return 42`.

For comparison, RFC 0030 fought hard for a 512-byte constant and won a 97.6% reduction on
programs with large globals. This is the same order of win for *every* program, including
the ones with no globals at all.

## 2. Why it happens

The native path **bundles the entire stdlib** into one compilation unit (RFC 0011 P4 /
BUG#93: there is no import-driven selection on this path — everything is concatenated, which
is what makes `std.*` work without an import). Every bundled function is then lowered,
emitted, and linked.

Existing dead-code elimination is **intra-procedural only**: `remove_unreachable_blocks`
(`ssa_opt.ax:517`) prunes unreachable BLOCKS within a function. Nothing prunes whole
functions, so an unused `HashMap.remove` is emitted in full.

## 3. Design — prune at codegen, not in the linker

Two places could do it:

- **Linker (`--gc-sections` style).** Requires per-function sections or ranges, plus
  renumbering every relocation and symbol offset after removal. High blast radius in the
  one component whose invariants are hardest to test.
- **Codegen (RECOMMENDED).** Compute reachability over the AIR call graph and simply do not
  emit unreachable functions. No relocation renumbering — offsets are assigned during
  emission, so skipped functions never enter the layout. The linker is untouched.

### The reachability roots

1. `main` (and the runtime entry / `_start` path).
2. Anything `#[export]`-ed, and everything reachable from an exported symbol (`--shared`,
   `--staticlib`: for a library, the exports ARE the roots and nothing else may be pruned).
3. **Every function whose ADDRESS is taken** — `OP_FUNC_ADDR` operands. AXIOM has first-class
   function pointers (BUG#49) and closures, so an indirectly-called function has no direct
   call edge. Missing this is the one way this optimization becomes a miscompile rather than
   a size win.
4. `drop` methods reachable through CTGC (`resolve_drop_method` binds them by type, not by a
   call site), and any other compiler-synthesized call target.

### Conservative default

Anything the analysis cannot classify must be treated as a root. The failure direction must
be "emitted unnecessarily" (a size regression), never "omitted while reachable" (a link
error at best, a wild call at worst).

## 4. Drawbacks

- The root set is the whole risk. Every mechanism that reaches a function without a direct
  call edge — fn pointers, closures, drop glue, interface vtables (RFC 0029 stores
  `OP_FUNC_ADDR` into a vtable box), operator overloads resolved by name, the `ax_*` runtime
  ABI — must be enumerated. RFC 0029's vtables are the newest such mechanism and are exactly
  the kind of thing an older audit would have missed.
- Debug/profiling flows that expect every symbol to exist would see fewer symbols.

## 5. Gate

Codegen change ⇒ **B==C mandatory**. Plus full regression, `elf_linux_check.sh`,
`ctgc_free_check.sh` (drop glue is a root category), and `exe_size_check.sh` extended with a
baseline row so the win is pinned and cannot silently regress.

**A dedicated indirect-call oracle is required**: a program whose only reference to a
function is through a function pointer / closure / interface method must still link and run.
That is the test that distinguishes this from a miscompile.

## 5b. Measured reachability (`-dfe-report`, 2026-07-22)

A dump-only pass landed first, mirroring the `-ctgc-free-report` precedent: it computes
reachability and PRINTS the count, injecting nothing. Proven inert — the executable is
byte-identical with and without the flag.

A dump-only pass landed first, mirroring the `-ctgc-free-report` precedent: it computes
reachability and PRINTS the count, injecting nothing. Proven inert — the executable is
byte-identical with and without the flag.

**The indirect-edge check passed on its own terms.** `t_indirectcall` reported 8 from `main`
alone — exactly `main` + the four functions reachable only indirectly + the three
intermediaries. A pass walking direct call edges would have said about 3 and silently marked
the fn-pointer targets, the lambda and the vtable method dead. That number is the evidence
that `OP_FUNC_ADDR` is followed.

### The first numbers were wrong by 60×, and the pass caught it itself

The seed-from-`main`-only counts (`return 42` → **1 / 184**) were published in this RFC as a
lower bound. They were a very loose one. Seeding the remaining root categories moves the same
program to **61 / 184**.

| program | main only | + export + ABI roots | dead `.text` bytes |
|---|---|---|---|
| `return 42` | 1 / 184 | **61 / 184** | **55,518 of 74,088 (74.9%)** |
| `t_indirectcall` | 8 / 191 | 68 / 191 | 55,518 of 74,723 |
| `t_dfeexport` | — | 63 / 186 (2 export roots) | 55,518 of 74,169 |

The correction came from an **audit counter, not from a crash**. §7 finding 2 below claimed
`ax_free` was a hidden root needing seeding, and finding 3 claimed the runtime ABI allow-list
was another. Both were reasoned out by reading. Rather than trust that, the pass shipped with
a counter asserting the premise — that no BUNDLED function answers to a runtime ABI name — and
the very first run reported **ten of them**. `std/runtime.ax` defines `ax_str_len`,
`ax_str_concat`, `ax_str_slice`, `ax_str_replace`, `ax_panic` and more as ordinary AXIOM
functions.

That inverts both findings. The DLL imports were never at risk, because imports are not
members of `mod.funcs` and pruning cannot reach them. The real hazard is the **shadow**: a
bundled definition whose name collides with an ABI symbol is bound **by name** —
`x86_resolve_callee_name` rewrites AXIOM `free` to `"ax_free"` and `std.string.len` to
`"ax_str_len"`, and the magic negative callee indices resolve to the same names — so its
callers leave no trace at all in the AIR graph. Sixty of the 61 functions now reachable in
`return 42` hang off those ten shims.

Had pruning been activated on the "1 / 184" reading, it would have deleted them and produced
calls into freed address space. The lesson is not that the reading was careless; it is that a
premise cheap enough to assert in code should never be left as prose.

The third root, `#[export]`, was equally unproven: it shipped matching only the resolved
symbol name, and `t_dfeexport` reported **0 export roots** because `export_syms` stores plain
intern name-ids while the emitted symbol is mangled. Matching both id and name fixed it. A
root category that silently matches nothing looks exactly like a root category that is not
needed.

### Bytes, not function counts

Function counts are the cheap number and the misleading one. What this RFC buys is `.text`
bytes, and the two ratios differ: 67% of functions are dead in `return 42` but **74.9% of the
bytes** are, because the dead set skews large — the ABI shims' closure is string machinery.
The report therefore runs AFTER emission and sums real emitted sizes rather than inferring a
win from a ratio.

**`dead_bytes` is 55,518 in all three programs, to the byte.** That is the fixed stdlib tail
of §1 measured directly rather than inferred from file-size differences, and its invariance
across three unrelated programs is a consistency check the count-based number could not give.

## 6. Expected result

Measured, no longer estimated: removing 55,518 bytes of dead `.text` takes `return 42` from
**77,824 to roughly 22 KB — about a 71% reduction**. The earlier "a few KB / ~90%" guess was
too optimistic; it assumed the 1/184 count, which the root seeding corrected. 71% of every
executable is still by far the largest size win available.

## 7. Status / next step

Measured and specified only; implementation NOT started. The root-set enumeration (§3) is the
whole difficulty and deserves its own session.

**Groundwork already banked (so the next session starts with the guard in place):**

- **`t_indirectcall`** (regression, exit 42) — four functions reachable ONLY through indirect
  edges: a function pointer in a variable, a lambda passed as a fn-typed argument, an
  interface method reached through its vtable, and an array of function pointers indexed at
  runtime. A reachability pass that walks direct call edges alone drops all four. Verified
  identical at O0–O3.
- **Calibrated, not assumed.** `dump-air` on that oracle shows **five `funcaddr`
  instructions**, and the table call lowers to `%36 = index …` / `%37 = call %36` — a genuine
  indirect call, not a direct one the compiler folded. So the oracle really does exercise the
  edges it claims to.
- **The AIR printer can now name them.** `OP_FUNC_ADDR` and `OP_GLOBAL_ADDR` previously
  printed as `???` — the exact instruction this RFC's analysis must trace. Both now print as
  `funcaddr` / `globaddr` (CLAUDE.md §9: the IR must be printable).

### Root survey (read-only, 2026-07-22) — three findings

1. **Drop glue and operator overloads are NOT hidden roots.** Both were listed as suspects,
   and both turn out to emit an ordinary `OP_CALL` carrying the resolved symbol
   (`lower_destroy` for `drop`, `lower_op_overload` for operators). A reachability walk over
   the FINAL AIR therefore sees them, provided it runs after CTGC injection. Better still,
   with `-ctgc-free` off no `OP_DESTROY` is injected at all, so unused `drop` methods are
   genuinely unreachable and pruning them is correct rather than merely safe.

2. ⚠️ **`ax_free` IS a hidden root.** `OP_FREE` and `OP_DESTROY` do not lower to a symbol
   reference — `x86_selector.ax` emits `MACH_CALL` with the immediate **-2**, a magic callee
   index resolved late. So the deallocator has NO call edge in the AIR graph and an
   otherwise-correct reachability pass would drop it, leaving a call to a pruned address.
   This is the concrete instance of the failure mode §3 warns about, found by reading rather
   than by a crash.

3. **The runtime ABI already has an allow-list.** `linker.ax` (~L635) enumerates the
   `__ax_runtime_init` / `ax_alloc` / `ax_str_*` / `ax_println_*` / scheduler family that must
   survive regardless of call edges. That list is a ready-made seed for the root set rather
   than something to re-derive — and it should be READ from one place, not duplicated, so the
   two cannot drift.

The entry root is `main` or `_AX_main_main_v_v` (`linker.ax` ~L3101).

Remaining before implementation: decide where the shared root list lives so the linker and the
new pass cannot disagree, then add oracle coverage for the `ax_free` path (a program that
frees under `-ctgc-free` and must still link) the same way `t_indirectcall` covers the
indirect-call edges.
