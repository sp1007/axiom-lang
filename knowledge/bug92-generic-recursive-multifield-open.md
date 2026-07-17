---
name: bug92-generic-recursive-multifield-open
description: "OPEN BUG#92 — a GENERIC multi-field variant with a generic_inst/aggregate field (e.g. recursive `Tree[T] = Node(T, Tree[T], Tree[T])`) reads child fields wrong/segfaults: the RFC 0019 synth struct uses TEMPLATE field types (never monomorphized). Non-generic recursive works; generic scalar-only works."
metadata: 
  node_type: memory
  type: project
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

✅ **FIXED 2026-07-10 (i)** — `938c48b`, `origin/main`=`938c48b`, backend change **A==B==C `0657BBEE`** (byte-identical: compiler has no such field itself), **145/145**. Daily-driver rebuilt.

**Root cause (confirmed, was close to hypothesis):** NOT template-vs-mono field types — the synth-struct field offsets were fine (each field 8 bytes). The bug was in `x86_selector.field_is_pointer_sum`: it recognized only bare sum/option/result fields (kinds 6/11/12), NOT a **kind-8 GENERIC_INST whose base is a user SUM** (`Tree[T]`). Such a field IS pointer-repr (value = 8-byte box pointer), but `field_is_aggregate` (true for any kind-8) then won in OP_GET_FIELD/OP_SET_FIELD → field handled BY ADDRESS (LEA) → child read as its own field offset → recursion segfault. **Fix:** `field_is_pointer_sum` also returns true for a kind-8 generic_inst field whose base name resolves to a SUM entry (scan). Corrects GET+SET symmetrically. Generic_inst-of-STRUCT (`Pair[i32,i64]`, `Vec[T]`) has no matching SUM entry → stays by-value (verified 112). Oracle `t_gentree`(15). Robust: deep tree(15), non-generic(10), scalar-generic(11).

**Separate finding — ✅ FIXED (REJECT) 2026-07-10 (j)** `62c0619`, A==B `883975F0`, 146/146: a user variant NAMED `None`/`Some`/`Ok`/`Err` (e.g. `type Pair[T] = P(T,T) | None`) silently miscompiled (16 vs 11; 8 vs 42). ROOT deeper than routing: the **constructor** lowers Some/None/Ok/Err tokens to the pointer-tagged BUILTIN layout (null/box|1) UNCONDITIONALLY by token text (air_builder ~1597), so a user sum reusing the name gets a MIXED repr (its variants tag@field0, the shadowed one builtin). First tried routing match by scrutinee sum-name → fragile (monomorphised Option name is qualified, broke genuine Option). Correct fix = **REJECT** in `typecheck.pre_infer_type_alias`: a sum whose base name (via `extract_qualified_base_name`) is NOT Option/Result may not declare Some/None/Ok/Err. Full support = type-directed construction (RFC). Oracle `t_variantshadow` (reject). Minor: emits 2 identical diags (2 passes) — acceptable. See [[bug-match-optstr-payload-length]].

---
🔴 ~~OPEN — BUG#92:~~ (đã đóng ở trên) a GENERIC multi-field variant whose payload includes a generic_inst/aggregate field miscompiles when that field is extracted+used.

**Repro:**
```
type Tree[T] = Node(T, Tree[T], Tree[T]) | Leaf
fn total(t: Tree[i64]) -> i64:
    match t:
        Node(v, l, r): return v + total(l) + total(r)   // recurse into children
        Leaf: return 0
fn main() -> i64:
    return total(Node(7, Node(2, Leaf, Leaf), Leaf))    // want 9 -> SEGFAULT
```
Narrowed: constructing a deep generic tree + reading only the top `v` WORKS (7); reading a CHILD field `l` (type `Tree[T]`, a generic_inst) and matching on it returns garbage (8 — looks like the field OFFSET) → recursion segfaults.

**Works (not affected):**
- Non-generic recursive multi-field: `type List = Cons(i64, List)|Nil`, `type Tree = Node(i64,Tree,Tree)|Leaf` → correct (10, 19). RFC 0019 enabled these.
- Generic multi-field with SCALAR fields only: `Pair[T]=P(T,T)` with T=i64 → correct (11, 12). (Works by luck: T defaults to 8-byte, offsets 0/8 line up.)
- Generic SINGLE-field variant (BUG#91 fix): correct.

**Root cause (hypothesis):** RFC 0019 [[rfc0019-multifield-variant-shipped]] registers the synth payload struct ONCE at sum registration with TEMPLATE field types (`T`, `Tree[T]`), and it is NEVER monomorphized per instantiation. `field_offset`/`field_size` (x86_selector) then compute offsets from the template types: a generic param `T` defaults to 8 bytes (so scalar-only structs accidentally work), but a `Tree[T]` (generic_inst) field's size/offset is mis-derived → child fields read at wrong offset/type (8 = offset leaking, not the value). Same family as the BUG#91 note "payload type stays T, not substituted to concrete".

**Fix direction:** monomorphize the synth struct per generic instantiation (fields substituted T→concrete), OR make the match/construct field offset/type resolution substitute generic params via the scrutinee's generic_inst args. Ties into the broader generic-sum monomorphization (see [[bug67-option-struct-payload-unwrap]], set_sum_generic_args, finish_generic_instantiation). Backend + typecheck; B==C. Oracle: the `Tree[i64]` total above → 9.

**Workaround:** use a non-generic ADT (`type IntTree = Node(i64,IntTree,IntTree)|Leaf`) — works today. Deferred as a deep generic-mono edge; not blocking (non-generic recursive ADTs cover the common case).

---
### 🔴 OPEN sub-case (BUG#92b) — generic MULTI-FIELD variant with a field that monomorphizes to str/>8B (2026-07-11, precisely root-caused)
`type P[T] = Pair(T, i32) | Nil` with `T=str`: `match Pair(s,n): len(s)+n` returns 7 not 10 (str `s` reads len 0); the recursive `Tree[str]` form SEGFAULTS. Single-field generic str (`W[T]=One(T)`, T=str) WORKS; generic i64 multi-field WORKS; generic STRUCT (non-sum) str field WORKS (pc=10).
**Root cause (confirmed via ax_puts_local probe):** the RFC 0019 synth payload struct is registered ONCE with TEMPLATE field types. **Construction** (`lower_variant_construct`, air_builder ~2508-2517) resolves `payload_type` from `vsym.type_id` = the TEMPLATE sum → synth field 0 = generic `T` (8 bytes) → SET_FIELD stores only 8 bytes (the str ptr, drops len) AND lays i32 at offset 8. **Extraction** (`lower_match` is_mf, ~2902) resolves from `scrut_type` = the MONOMORPHISED `P[str]` → synth field 0 = str(12), 16 bytes at offset 0, i32 at offset 16 → reads 16 bytes (len half uninitialized=0). Template-vs-mono LAYOUT MISMATCH. Only bites when template field (8-byte generic) ≠ mono field (>8B, i.e. str); i64 mono==8 so it aligns by luck (same "works by luck" as the scalar case).
**Fix direction (REFINED 2026-07-11 — attempted & reverted, tree clean `62c0619`):** tried plumbing the construct expr's inferred type `node_types[idx]` into `lower_variant_construct` and using `resolve_sum_type(mono_type)` to pick the mono synth struct. **Didn't work** — probe showed `VC-SAMEASTMPL` + `VC-F0OTHER`: the construct expr types as a **GENERIC_INST** `P[str]`, and `resolve_sum_type` (air_builder ~2394) matches a GENERIC_INST to the FIRST `kind==SUM` with the same `name_id` = the **TEMPLATE** sum (field 0 = generic T), NOT the mono sum. Meanwhile the `match` scrutinee's param type `p: P[str]` is a mono **kind-6 SUM** entry directly (str synth struct), so extraction gets str12. So there ARE two same-named sum entries (template + mono) and `resolve_sum_type` can't disambiguate. **Real fix = make `resolve_sum_type` (or a new resolver) pick the mono sum whose `SumInfo.generic_args` match the GENERIC_INST's args** (RFC 0013 stores generic_args) — a monomorphisation-matching change touching a widely-used helper (broad blast radius, B==C). Alternatively construction computes per-field offsets/size from the ARG types directly (bypass the struct-type layout) — fiddly, must exactly mirror `type_size_and_align`. Both non-trivial; deep-mono, dedicated session.
**Probes that PASS:** single-field generic str (qa=5), generic i64 multi-field (qc=42), Option[Tree[i64]] (pb=8), generic struct str field (pc=10). Surfaced by adjacency bug-probe to the optstr/BUG#92 fixes.
