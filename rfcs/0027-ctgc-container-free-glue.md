# RFC 0027 — CTGC free-glue for nested-heap containers

- Status: SHIPPED (2026-07-19, `10eceb6`) — path C implemented in `air_builder.ax::
  emit_container_buffer_frees`. Gate GREEN: fixpoint A==B `9A178747`, regression 436/436,
  ctgc_free_check 12/12 (+`t_ctgccont`), broad sweep 419 checked/0 crashes/only intended
  `t_drop` diff, 8 aliasing/escape probes correct. D (field-ownership annotations) remains the
  long-term general successor.
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

## 9. Amendment 2026-07-22 — path D shipped, with a proven limit on the no-syntax model

User decision (2026-07-22), choosing between the §3 D variants: derive field ownership **from
the type plus assignment provenance, adding no annotation syntax**. Path D is now implemented on
that basis (`emit_owned_field_frees`, formerly `emit_container_buffer_frees`).

### The shipped rule

A field is freed before its owner's header iff **both** halves hold:

1. **Type half** — the field's type is `TYPE_KIND_POINTER`. Non-pointer fields, and nested
   aggregate fields, are never freed.
2. **Provenance half** — every write to that field, across the whole program (constructor named
   arguments and `x.f = ..` assignments), is either a **syntactically fresh allocation**
   (`@alloc`/`@realloc`, through any casts) or `null`; and at least one is a real allocation.

Path C's audited `{Vec,HashMap,HashSet}` × `{data,keys,values,hashes,occupied}` list is **kept as
a fallback** beside the derivation. It is not redundancy: the derived rule is deliberately
conservative, so retaining the audited answer guarantees path D can only ADD coverage for user
containers and can never silently REMOVE what path C already shipped.

### The limit, established by counterexample rather than argument

The natural strengthening — follow a local one binding, so that the idiomatic
`let d = @alloc(..)` followed by `Buf(data: d)` is recognised — **is unsound, and was caught in
bring-up as a real use-after-free.** `Buf(data: d)` (owning) and `View(borrowed: d)` (borrowing)
are *syntactically identical*. What separates them is whether anyone else still holds the
pointer, which is an **aliasing** question that no amount of provenance analysis answers. The
`t_ctgcborrow` oracle freed the borrowed buffer and the next allocation overwrote it (returned 99
instead of 42).

So the decidable fragment of the no-syntax model is exactly "a fresh allocation expression at the
write site", because nothing else can hold that pointer yet. **This is the concrete evidence that
§3 D's original framing was right: annotations are what generalise this.** Path D as shipped is
the sound sublanguage of that design, not a substitute for it.

### Known coverage limits (all in the safe direction — leak, never a wrong free)

- **The idiomatic form is not covered.** `let d = @alloc(..); S(data: d)` — the shape
  `std/collections.ax` itself uses — is classified borrowed.
- **Name-sensitivity.** The scan matches field names program-wide, not per declaring type, so one
  same-named borrowed field anywhere (including bundled stdlib) demotes the field everywhere. In
  practice `data` is permanently demoted by `Vec[T](data: data)` in `with_capacity`; user
  containers must use a distinct field name for the derivation to fire. Adding per-type precision
  would remove this wart but would NOT extend coverage past the limit above.
- **No recursion into nested aggregate fields.** Under reference semantics (RFC 0001 §5) a nested
  `Vec`-in-struct field may alias a live local; freeing it transitively needs the alias tracking
  this design lacks.

### Gate result

Fixpoint A==B `F7E84810EF6544C7950ADCAE2E460ED3B6CE1D443CF1F56026FD06FDBA49CE70` (inert on the
self-host build by §2.3 — the flag is off there), `ctgc_free_check.sh` 14/14, full regression.
Oracles: `t_ctgcuser` (user container, correctness) + **`t_ctgcborrow`** (the sharp one — it is
the oracle that failed loudly on the unsound rule, and is the reason this amendment can state the
limit as a measured fact).

**A caveat on what the oracles prove.** `t_ctgcuser` pins `on == off`; it canNOT detect a silent
no-fire, because the runtime allocator does not return a just-freed block to the next same-size
request in that shape, and a program cannot otherwise observe its own heap. Firing was confirmed
out-of-band with a temporary trace in `emit_owned_field_frees`. Per the RFC 0028 lesson this is
recorded rather than papered over: the dangerous direction (over-freeing) has a sharp oracle, the
inert direction does not.

## 10. Annotations reconsidered 2026-07-22 — better design, deferred on value not cost

After §9 established that the no-syntax model has a hard limit, the user asked whether the
annotation variant should simply be adopted instead. Reassessed, and **deferred** — with the
trigger written down so this is a scheduling decision, not a rejection.

**Cost is NOT the blocker.** `#[export]` already exists (parser.ax `TK_HASH`, ~L1572), so
`#[owns]` reuses the attribute lexing/parsing path rather than inventing syntax. The work is
`StructField` gaining a flags word, typecheck propagating it, and `emit_owned_field_frees`
reading the flag instead of scanning the AST. Bounded — a few gated cycles.

**Value is the blocker.** Every benefit — of path D *and* of annotations — is realised only
under `-ctgc-free`, which is opt-in, off for the self-host build, and off for every default
user build. Today both models therefore have **zero effect on any shipped artifact**; what
separates them is the coverage of a feature nobody's build enables. Against that, annotations
move the change out of the provably-inert free path (§2.3) and into the parser/typecheck —
self-host-critical code — which is a real risk increase for an unrealised gain. CLAUDE.md §3,
§13 and §19 all counsel against adding language surface before the need is demonstrated.

**Trigger to adopt:** a decision that `-ctgc-free` becomes DEFAULT-ON, i.e. that AXIOM commits
to reclaiming memory in shipped programs. At that moment path D's coverage gap stops being
theoretical (the idiomatic `let d = @alloc(..); S(data: d)` is the shape `std/collections.ax`
itself uses) and annotations become the right answer immediately. Revisit here first.

**Cheaper intermediate, available any time, no syntax:** replace the program-wide field-NAME
match with a per-declaring-TYPE match. That removes the wart where `data` is permanently
demoted by `Vec[T](data: data)`, so user containers stop depending on their field's name. It
stays inside the free path and does NOT extend coverage past the fresh-allocation limit — it
only makes the existing coverage predictable.
