# RFC 0037 — Method name resolution is RANKED, and exact names win

- Status: implemented; **amended 2026-07-31 — rank 1 (the loose window) is RETIRED, see §"Amendment"**
- Date: 2026-07-31
- Affects: `bootstrap/stage1/air_builder.ax` (AIR-build method resolution), `bootstrap/stage1/typecheck.ax`
  (the diagnostic mirror). No IR, ABI, linker or syntax change.
- Supersedes nothing. Makes explicit a rule RFC 0029 already required.

## Motivation

Method-form calls (`x.m(...)`), interface vtable slot filling, operator methods and drop glue are all
resolved in air_builder by scanning the symbol table for a SYM_FUNC whose name "matches" the called
name and whose first parameter unwraps to the receiver type. The name test
(`match_mangled_method_raw_bytes`) accepted the called name as a **substring** of the symbol name,
bounded by `.` or `_` on the left and by end-of-name or `__` on the right. Every loop then took the
**first** such symbol in symbol order.

That combination silently selected the wrong function whenever two methods on the SAME receiver type
were named so that the shorter is a `_`-bounded suffix (or `__`-bounded infix) of the longer, and the
longer was declared first:

```axiom
struct S:
    fn p32_r32(self: S) -> i64: return 7
    fn r32(self: S) -> i64:     return 42
S(...).r32()   // called p32_r32 and returned 7
```

It is type-agnostic, it hits static calls as well as interface dispatch, and when the two signatures
differ it is type confusion (an f64 return read out of RAX) or arity confusion (an extra parameter
read from a garbage register), not merely a wrong number. RFC 0029 §"Vtable synth + init" already
specifies the intended rule — *for slot k, resolve T's method whose name = `iface_methods[I][k]`* —
i.e. name EQUALITY. So this was an implementation bug, not a design question.

## Why not simply tighten the match

The loose rule is load-bearing in two ways:

1. A monomorphized instance is named `_AX_std_<name>__<typeargs>` (`mono.ax mangle_name`), so `push`
   must still find `_AX_std_push__i64`.
2. A module-level function can act as a namespaced method: `os.pathbuf_to_str(self: PathBuf)` is what
   would answer `p.to_str()`.

Rejecting (1) breaks generics outright. Rejecting (2) would change the meaning of programs that are
not covered by any test — **measured, not assumed**: a compiler with rank 1 removed entirely still
passes 619/619 regressions and rebuilds the compiler+stdlib to a byte-identical binary (see
"Evidence"), because every namespaced stdlib form is in fact called by its full name
(`pathbuf_to_str(parent)` at std/os.ax:221, `self.locals.local_map_get(..)`). So a tightening would
have been safe *for this repo*; ranking is preferred because it cannot break user code outside it and
because it is a strictly smaller behaviour change. Retiring rank 1 (with a deprecation diagnostic) is
a reasonable follow-up, but it is a language-semantics decision, not a bug fix.

## Design

`method_name_rank_raw_bytes(pool, sym_id, method_id) -> i64` ranks a candidate; each resolution loop
selects the **highest-ranked** candidate that also passes its own receiver/parameter filter, and stops
early on the top rank. Ties keep the previous behaviour: first in symbol order.

| rank | meaning |
|------|---------|
| 3 | the symbol's PLAIN name — the segment after the last `.`, minus an `_AX_std_` prefix — **equals** the called name |
| 2 | that plain name is the called name followed by the `__` type-argument separator (`_AX_std_push__i64` answers `push`) |
| 1 | the legacy loose window hit (`os.pathbuf_to_str` answers `to_str`) |
| 0 | no match |

`match_mangled_method_raw_bytes` is retained as `rank != 0`, so every *existence* test (ownership.ax's
`scan_has_drop` / `type_has_drop`) keeps its exact previous meaning; only the *selection* loops change.
The loose window itself is unchanged, moved verbatim into `match_mangled_method_loose`, so there is
exactly one implementation of each rule.

Updated selection sites (all of them):

- `air_builder.ax find_struct_method_sym` — interface vtable slot k (RFC 0029)
- `air_builder.ax` UFCS method-dispatch repair loop — the plain static `x.m()` path
- `air_builder.ax resolve_op_method` — operator overloads (RFC 0007)
- `air_builder.ax resolve_drop_method` — drop glue (RFC 0014)
- `typecheck.ax diag_resolve_op_method` — the diagnostic mirror of resolve_op_method; a diagnostic
  that disagrees with the resolver is its own bug

Consequence: this only re-decides calls where an exactly-named candidate **and** a looser candidate
both pass the receiver filter — precisely the shadowed set. Programs with no name collision resolve
identically.

## Drawbacks / limits

- Two methods on one receiver whose plain names are *both* exactly the called name cannot exist, so
  rank 3 is unique per receiver; but rank 1 ties are still resolved by symbol order, which is
  declaration order — unspecified, and unchanged by this RFC.
- The rule is still name-based, not signature-based: an exactly-named method with the wrong arity
  still wins over a loose candidate with the right arity. That is intentional — the interface
  conformance check (typecheck `interface_signature_mismatch`) is where signature mismatches are
  rejected.

## Evidence

- Oracles: `bin/t_methshadow.ax` (11 rows: static, interface i64, interface str, `__`-infix,
  module-level UFCS, plus 5 already-correct controls) and `bin/t_methshadowsig.ax` (type confusion,
  arity confusion). Pre-fix they return 1 / 1@-O0 and 2@-O1; post-fix 42 at -O0/-O1/-O2. Registered in
  `scripts/regression_repros.sh` main list, the -O0-only list and the -O2/-O3 list.
- 31 probe programs in `bin/probe6/` (the full trigger matrix) plus `bin/probe4/f1.ax`,
  `bin/probe5/r6d,r6e,f1r6` go from wrong to 42; the already-correct probes stay 42.
- Fixpoint A==B holds; the compiler's own sources contain zero inline struct methods and always call
  the namespaced form explicitly (`self.locals.local_map_get(...)`, `scope.scope_get(...)`), i.e. rank
  3, so self-compilation is unaffected — a compiler built with rank 1 removed entirely emits
  byte-identical code size for the same source.

## Amendment (2026-07-31) — rank 1 is RETIRED

Ranking fixed only the case where an exactly-named candidate **also** existed. When none did, the
loose window was still the ANSWER, and that turned out not to be a compatibility niche but three live
defects — each measured, each an accept-then-miscompile (BUG#53 class):

| symptom | measured |
|---------|----------|
| `a == b` on a type declaring only `deep_eq` called deep_eq and returned true for unequal values. Same for `total_lt`/`<` and `checked_add`/`+`. The RFC 0007 §2.2 diagnostic the compiler **already had** was suppressed by rank 1. | exit 7 / 7 / 8 → now rejected |
| RFC 0014 drop glue INVENTED a call the user never wrote: a type declaring only `pre_drop(self)` had it invoked at every scope exit, including a `late_drop(self) -> ptr[i64]` called through a `drop(self)` shape. Running an arbitrary method on a block about to be freed is a UAF, not a wrong number. | 5 and 3 spurious calls → now 0 |
| `ownership.ax type_has_drop` reported "has drop" for `pre_drop`, so `let b = a` on a type with no drop was rejected with **E4003**. | rejected → accepted |

The third one is also where the *two copies* of the rule bit. `match_mangled_method_raw_bytes` was
retained as `rank != 0` for ownership.ax's existence tests (see §Design above) — but that only kept
ownership in step by accident of the rank-1 fallback, and it was never converted when the rule became
rank-based. There is now **exactly one** function, `method_name_rank_raw_bytes`, and every consumer
calls it directly: `resolve_op_method`, `resolve_drop_method`, `find_struct_method_sym`, the UFCS
dispatch loop, `typecheck.diag_resolve_op_method`, and ownership's `scan_has_drop` / `type_has_drop`
(as `rank != 0`, which is exactly when `resolve_drop_method` returns a non-zero symbol). Three further
dead near-copies of the same idea — `match_base_names`, `match_mangled_method_name`,
`match_mangled_method_loose` — are deleted rather than left unused; leaving a fourth copy of a rule
that just caused a bug is how it comes back.

Ranks are now `{3, 2, 0}`; the selection loops are unchanged (highest rank wins, rank 3 short-circuits,
rank 2 remembered as a fallback for the monomorphized `_AX_std_<name>__<typeargs>` form).

**Named consequence, accepted:** a module-level function used as a namespaced method
(`os.pathbuf_to_str(self: PathBuf)` answering `p.to_str()`) no longer resolves. Nothing in this repo
does that — every such form is called by its full name — and the alternative is silently calling a
function the user did not name.

**Known remaining over-reach, NOT addressed here** (both introduced by the mangling-aware rules of the
original RFC, not by the loose window, and both need their own decision):

- rank 2 accepts a **user-declared** `eq__fast` as an answer to `eq` (`bin/probe8/h1_rank2_eq.ax`,
  still exit 7). The `__` separator is mono's, but the rule does not require the name to have carried
  a mono prefix.
- the `_AX_std_` strip makes a user-declared `_AX_std_eq` a rank-3 **tie** with a genuine `eq`, decided
  by declaration order (`bin/probe8/h2_axstd_prefix.ax` = 7 vs `h3_axstd_prefix_rev.ax` = 42).

Both are the same shape: a user name that *mimics* the monomorphizer's mangling is treated as if the
monomorphizer had produced it. A fix would key ranks 2/3 on the symbol actually being an instantiation
rather than on its spelling.

### Evidence (amendment)

- `bin/t_dropglueexact.ax` — REJECTED (E4003) before, **42** after; pins 0 spurious `pre_drop`/
  `late_drop` calls, 6 real `drop` calls when a real `drop` exists, and the legality of copying a
  `pre_drop`-only type.
- `bin/t_opnooverload.ax` / `t_opnooverloadlt.ax` / `t_opnooverloadarith.ax` — exit 7 / 7 / 8 before,
  **rejected** after.
- `bin/t_ufcsnoexact.ax`, `bin/t_ifacenoexact.ax` — rejected before and after (guards against
  re-loosening).
- `bin/t_methnamestrict.ax` — **42 before and after**: the over-reach guard. 12 calls that all name a
  method that exists (both declaration orders of `len`/`buf_len`, two structs sharing `get`, the
  static call form, an interface declaring both `r32` and `p32_r32`, a generic struct method, generic
  free fns, a stdlib call beside a superstring-named free fn, exact `eq` beside `deep_eq`).
- All rows registered in `scripts/regression_repros.sh` (main list, the -O0 list and the -O2/-O3 list).
