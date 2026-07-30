# RFC 0037 — Method name resolution is RANKED, and exact names win

- Status: implemented
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
