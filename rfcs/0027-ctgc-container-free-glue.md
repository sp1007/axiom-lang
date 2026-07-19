# RFC 0027 — CTGC free-glue for nested-heap containers

- Status: PROPOSED (2026-07-19)
- Depends on: RFC 0014 (drop-glue), RFC 0015 (CTGC free), the P3 general-free activation (`18db268`)
- Related memory: [[ctgc-p3-scoping-2026-07-18]], [[bug-3hashmap-mono-teardown-crash]]

## 1. Motivation

RFC 0015 P3 general free is now ACTIVE (`18db268`): under the opt-in `-ctgc-free` flag, a
non-escaping owned local whose type declares no `drop` is freed at block-exit with a plain
`OP_DESTROY` (a single `free(ptr)` on the block, `x86_selector.ax` OP_DESTROY = `mov arg0,ptr;
call free`).

For a **container that owns nested heap** — `Vec[T]` (owns `data`), `HashMap[K,V]` (owns
`keys`/`values`/`hashes`/`occupied`), `HashSet[T]` — this frees the **header only**; the inner
buffers leak. This is strictly no worse than the pre-activation "never free", but it leaves the
CTGC free story incomplete: the common case (a scratch `Vec`/`HashMap` local) still leaks its
bulk memory.

The containers ALREADY have the freeing logic — `std/collections.ax` has
`destroy[T](Vec[T])`, `destroy[K,V](HashMap[K,V])`, `destroy[T](HashSet[T])` that `@free` each
buffer — but the CTGC path resolves a method named `drop`, not `destroy`, so they are never
auto-called.

## 2. Constraints (from investigation, 2026-07-19)

1. **Mono is typecheck-driven and runs BEFORE CTGC** (`main_air.ax:1017-1039`;
   `instantiate_function` called from `typecheck.ax:2602/3769/5105` on real source calls). A
   container's `drop[…]` is never *called* in source, so the monomorphizer never creates a
   concrete `drop[Vec[i64]]` instance → `resolve_drop_method` (matches by concrete unwrapped param
   type, `air_builder.ax:1089`) returns 0.
2. **`air_builder` holds no `Monomorphizer`** — it cannot lazily instantiate a template at
   lowering time without new plumbing.
3. **`-ctgc-free` is opt-in and OFF for the self-host build.** `OP_DESTROY` is injected only by the
   CTGC pass under the flag, so `lower_destroy`'s free path never runs when the compiler builds
   itself. **Any change confined to the `OP_DESTROY`/`lower_destroy` free path is therefore
   fixpoint-neutral (A==B) by construction** — the self-host binary is byte-identical.

## 3. Design options

- **A — Eager `drop` instantiation in mono.** When the monomorphizer creates a container
  generic-inst, also instantiate its `drop[…]`. Reuses drop-glue; but touches self-host-critical
  mono, creates `drop` instances in *every* build (symbol-table growth → fixpoint shift risk), and
  runs in the same generic-mono machinery that still hosts the unpinned 3-hashmap heap-corruption
  ([[bug-3hashmap-mono-teardown-crash]]). Highest blast radius.
- **B — Lazy instantiation at air time.** Plumb the `Monomorphizer` into `air_builder` and
  instantiate `drop` on demand in `lower_destroy`. Only instantiates under the flag, but creating
  new symbols/types AFTER typecheck violates pass-ordering invariants (typecheck-established
  symbol/type tables are assumed frozen downstream).
- **C — Container-recognition free-glue in `lower_destroy` (RECOMMENDED).** In `lower_destroy`,
  when the freed local's type is a generic-inst whose base is a known owning container
  (`Vec`/`HashMap`/`HashSet`), emit an `OP_FREE` for each of its owned buffer fields (loaded via
  `OP_GET_FIELD`) BEFORE the header `OP_DESTROY`. No mono/typecheck change; localized to the free
  path; fixpoint-neutral by §2.3. Fragility: couples the compiler to those three stdlib types'
  field layout — mitigated by a narrow name-guard + a comment cross-referencing `collections.ax`,
  and by the fact that these are stable core types.
- **D — Field-ownership annotations + general recursive free (long-term).** Mark struct fields as
  owned-heap vs borrowed; `lower_destroy` recursively frees owned pointer fields for ANY type.
  Principled and general (covers user containers), but a language-level increment (annotation
  syntax + type-system change + escape interaction). Out of scope here; the eventual home for this.

## 4. Recommendation

Ship **C** now (bounded, fixpoint-safe, opt-in-gated), and record **D** as the general successor.
C closes the leak for the three core containers — the 99% case — with the least risk to the
freshly-activated CTGC path and zero risk to the self-host build.

### Sketch (C)
In `air_builder.ax::lower_destroy`, in the `else` (non-drop) branch added by `18db268`, before
emitting the header `OP_DESTROY(reg)`:
1. Resolve `sym.type_id` to its typetable entry; if `TYPE_KIND_GENERIC_INST` and the base name is
   `Vec`/`HashMap`/`HashSet`, look up its pointer buffer fields by name
   (`data`; or `keys`/`values`/`hashes`/`occupied`).
2. For each such field: `OP_GET_FIELD reg -> fptr`, then `OP_FREE fptr`.
3. Then the existing `OP_DESTROY reg` frees the header.
Emit frees for null buffers safely — `ax_free(null)` is a no-op (`std/mem/alloc.ax:free` early-
returns on null), so an unallocated container (cap 0, null buffers) is safe.

## 5. Soundness (all options share this)

The escape analyser must mark a container local escaping whenever it is aliased under AXIOM's
aggregate=reference semantics (`let m2 = m1`; store into a field/container; returned). The
reassign/alias escape edges (`68d2c78`) and container-store edge (`f873948`) already do this for
structs; freeing a container's buffers is only reached for a local the analyser deems
non-escaping, so no other reference survives to the freed buffers. **This must be validated with an
aliased-container oracle** (a `Vec`/`HashMap` bound to a second local, or stored, must NOT be
freed) — a required gate item.

## 6. Drawbacks

- C hardcodes three stdlib types' field names in the backend (fragile to a `collections.ax`
  refactor; guarded by name + comment). D removes this but is a bigger change.
- Frees run only under `-ctgc-free` (opt-in); default builds still leak (acceptable — matches the
  current shipped scope).

## 7. Gate (before commit)

Backend/codegen change → fixpoint + full regression + the broad `-ctgc-free` sweep
(`scratch/ctgc_sweep.sh`, with-vs-without-flag output equality) + `ctgc_free_check.sh` + a new
**aliased-container** oracle proving an aliased container is NOT freed. Revert-on-red.

## 8. Migration / compatibility

No source-level change; no ABI change. `-ctgc-free` behavior becomes "frees more" (less leak);
default builds unchanged. D, if later adopted, subsumes C's hardcoded recognition.
