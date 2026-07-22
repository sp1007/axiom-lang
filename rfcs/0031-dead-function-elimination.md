# RFC 0031 — Dead function elimination

Status: **DRAFT** (measured, design proposed, not yet implemented)
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

## 6. Expected result

If reachability is accurate, `return 42` should link to a few KB rather than 77 KB — roughly
a 90% reduction on small programs, and a proportional win on every program that uses part of
the stdlib rather than all of it.

## 7. Status / next step

Measured and specified only. Implementation is NOT started: the root-set enumeration (§3) is
the whole difficulty and deserves its own session with the indirect-call oracle written
FIRST, so the dangerous direction is under test before any function is dropped.
