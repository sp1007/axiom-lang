---
name: bug-generic-struct-inline-method
description: "FULLY FIXED 2026-07-24g (641D2653, 543/543, A==B): generic type-arg inference through a nested ptr-self param (self: ptr[Hold[T]]) defaulted T to i32 for ANY generic fn (inline method OR free-fn, UFCS or direct) — str returns crashed, i64 returns silently TRUNCATED. The long-banked '>8B return codegen' framing was WRONG; root was a 2-line auto-ref gap in infer_generic_type_args. Earlier 8B-T inline-method dispatch fix shipped ccfd1af7. Oracle t_genptrselfmono."
metadata:
  node_type: memory
  type: project
---

## ✅✅ FULLY FIXED 2026-07-24g — driver `641D2653`, 543/543, A==B — ROOT WAS NOT ">8B RETURN CODEGEN"
**The entire ">8B return codegen / sret-with-receiver" framing below (banked across many ticks) was
WRONG.** A clean re-probe with a UNIQUELY-named method (`fetchxq`, to avoid the stdlib-`get` trace
pollution the notes kept hitting) + an `XQCALL` trace at the OP_CALL selection showed the mono'd
instance was named **`ax__AX_std_Hold__i32__...fetchxq__i32`** with `ret_tid=3 size=4 kind=0` — i.e.
`Hold[str]` was being monomorphized as **`Hold[i32]`**. The type param `T` was DEFAULTING TO i32.
- **The "8-byte-T works" evidence was a coincidence:** `Hold[i64]`.item=42 also mono'd as i32 and
  returned 42 only because 42 fits in 4 bytes. `bigval` (i64 = 5000000123 > 2^32) returned the
  TRUNCATED low 32 bits → the i64 case was ALSO broken, just invisibly.
- **NOT inline-method-specific, NOT >8B-specific:** the free-fn forms (`h.fetchxq2()` UFCS AND
  `fetchxq3(h)` direct) truncated identically. It is ANY generic fn whose `T` appears ONLY inside a
  nested `ptr[Hold[T]]` param.
**ROOT (typecheck.ax `infer_generic_type_args` ~L990/L1026):** the `ptr[..]`-param handling recursed
into the inner type ONLY when the ARG type was itself POINTER/REF kind. But for UFCS/auto-ref calls
the receiver `h: Hold[i64]` is passed as a VALUE (not yet `ptr[Hold[i64]]`), so the condition failed,
the recursion was skipped, `T` never bound, and mono defaulted it to i32. Vec/HashMap methods escape
this because they are backend-intercepted builtins, not routed through this inference.
**FIX:** in both `ptr` branches (the `is_ptr` NODE_GENERIC_TYPE case and NODE_PTR_TYPE), add an `else`:
when the arg is NOT pointer/ref (the auto-ref case), recurse matching the inner param against the arg
DIRECTLY (the arg IS the pointee). 2 small edits, typecheck-only. Gate: fast fixpoint **A==B
`641D2653`** (self-host inert — the compiler's own generic-ptr-self code was already correct or unused
via this path), regression **543/543** (+`t_genptrselfmono`: inline method str-return=5 + i64>2^32
non-truncation + free-fn direct, O0==O1=42). Oracle `bin/t_genptrselfmono.ax`.
⭐ **LESSON:** a bug banked as "return-value ABI codegen, no working reference, needs deep session" was
actually a 2-line **type-INFERENCE** fix — the wrong layer entirely. The misdirection came from (1)
tracing on the common name `get` (stdlib pollution, self-corrected once but the conclusion drifted),
and (2) the i64 case "passing" hid that mono was wrong for it too. A unique-named repro + a call-site
trace printing the MANGLED instance name (`Hold__i32`) exposed the real cause in one build. When a bug
resists a documented layer, re-probe with a unique name and read the mono'd SYMBOL NAME, not the codegen.

## ✅ EXTENDED 2026-07-24g — TUPLE & ARRAY param positions had the SAME gap — driver `09E41CC5`, 545/545, A==B
Right after the ptr-self fix, probed sibling inference positions: `infer_generic_type_args` had branches
for ptr/slice/generic/func but **none for NODE_TUPLE_TYPE or NODE_ARRAY_TYPE**. So a `T` appearing only
inside a `(T, i64)` tuple param or a `[T; N]` array param was never inferred → i32 default → same silent
truncation (`firstel((5000000123, 7))` and `head([5000000123,..])` returned the truncated low 32 bits).
FIX: added both branches — ARRAY mirrors the slice branch (recurse the element node against the arg
array's `entry.extra`); TUPLE matches each param element type node against the canonical `__tupN` STRUCT's
field types (`structs.data[extra].fields[i].type_id`). A==B `09E41CC5`, regression 545/545, oracle
`t_gentuplearrayinfer` (tuple + array, large-i64 non-truncation, O0==O1=42). NOT-a-bug boundary: a BARE
int literal `wrap(5000000123)` still defaults to i32 (separate int-literal-width behavior, not inference);
use a typed local to get i64. Mutator methods (`fn set(self: ptr[Hold[T]], v: T)`) were already fine (T
comes from the direct `v: T` param).

## (superseded) ✅ FIXED 2026-07-24e (8-byte-T cases) — driver `ccfd1af7`, 534/534, B==C
The fix turned out to be **parser-only** (the dispatch/mono "just works" once the method is
structurally identical to the free-fn form): `parse_struct_decl` now calls `inherit_struct_generics`
on each inline method — clones the struct's generic-params subtree (`clone_subtree`) as the method's
FIRST child + sets `FLAG_IS_GENERIC` (only when the method declares no generics of its own). So
`fn get(self: ptr[Box[T]])` becomes effectively `fn get[T](self: ptr[Box[T]])` and mono instantiates
it per `Box[i64]`/`Box[i32]`. Gate: B==C `ccfd1af7` (A≠seed — added the helper; A==B==C), regression
**534/534** (+`t_genstructmethod`, two 8-byte instantiations × two methods), p4/p4b/p4c/i2/i4 all
correct. Inert on the self-host (it uses the free-fn workaround → the parser change adds nothing to
existing code).
**REMAINING — a DISTINCT, PRE-EXISTING ABI bug (not caused by this fix):** a generic METHOD (one with
a `self: ptr[Box[T]]` receiver) that returns a **>8-byte `T` BY VALUE** (`get(self) -> T` with
`T=str`, 16B) SIGSEGVs — AND it does so in the FREE-FN method form too (`fn get[T](self: ptr[Box[T]])
-> T`, `gstr2.ax` → SEGV), which this parser fix never touched, so it is pre-existing. Crucially a
PLAIN generic fn returning `T=str` WORKS (`fn id[T](x: T) -> T`, `id[str]("hello").len` = 5,
`gstr.ax`), and a method returning an 8-byte type works — so the failure is specifically the
**receiver + >8B-`T`-return-by-value** combination (likely the sret return-slot ABI conflicting with
the struct-pointer `self`, or the >8B return not set up when a receiver is present). Separate
return-value-ABI investigation; repros `/tmp/probe4/gstr2.ax` (SEGV) vs `gstr.ax` (ok, =5).
**SHARPENED (probe matrix):** NON-generic method→str WORKS (`ng.ax`=5); free-fn(ptr arg)→str WORKS
(`ng2.ax`=5); plain generic `id[T]->T` with str WORKS (`gstr.ax`=5). ONLY the combination **generic
method + return type IS the generic `T` (not a direct value param) + `T` mono'd to >8B** fails. So it
is a MONOMORPHIZATION return-width bug: when `get[T](self: ptr[Box[T]]) -> T` is instantiated with
`T=str`, the mono'd function's RETURN type/width is not updated from the 8-byte generic placeholder to
the 16-byte str, so codegen returns it on the wrong (8-byte) ABI path → SIGSEGV. `id[T]` escapes it
because its return `T` is also its value PARAM `x: T`, so the width is pinned by the arg; `get[T]`'s
return `T` comes only from the field, decoupled from the `self: ptr[Box[T]]` param. FIX AREA: mono's
handling of a return type that is a bare generic param — ensure the instantiated function's return
type_id (and its ABI width) is the concrete `T`. Fresh session; gate B==C (codegen/ABI).
**ROOT NAILED (probe, 2026-07-24e):** it's the CALL-SITE RESULT TYPE, not the callee's return codegen.
`let r = s.get()` (unannotated) infers `r` as the wrong type (generic `T` NOT resolved to `str`) → `r`
gets an 8-byte slot → `r.len` reads garbage/off-slot → SIGSEGV (`use.ax`). With `let r: str = s.get()`
the slot is 16B so NO crash but `r.len` = 39 (garbage — the value still arrives wrong) (`use2.ax`).
And `let r = s.get()` with `r` UNUSED runs fine (`asg.ax`=5) — the value is only mis-handled when read.
⇒ **the call `s.get()` on a mono'd generic method does not resolve its result type to the method's
CONCRETE return type** (`str`), defaulting to the generic placeholder width. This is the SAME CLASS as
the shipped [[bug-generic-fn-float-return-default-int]] fix (typecheck.ax:4560, the `flags & 2048`
mono-instance path that overrides result_type with the instance signature's return) — but that fix is
on the free-CALL path; the METHOD-CALL path (`b.get()` via NODE_FIELD_EXPR callee) needs the same
instance-return-type override. FIX: in the call handler, when the callee is a mono'd generic METHOD,
set the call's result_type from the instantiated method's return type_id (not the template `T`).
Localized once found; gate B==C. Repros `use.ax`(SEGV)/`use2.ax`(=39 garbage)/`asg.ax`(=5, unused ok).
**FINAL REFINEMENT — it is TWO facets, the deeper one is VALUE TRANSFER, not just result type:**
`let r: str = s.get()` (correct 16B slot via annotation) STILL gives garbage (`use2.ax`=39, want 5),
and `let r = s.get()` with `r` UNUSED runs fine (`asg.ax`=5). So even when the type is correct, the
mono'd `get[str]` **mis-transfers the 16-byte str return VALUE itself** — the crash (unannotated) vs
garbage (annotated) difference is only the caller's slot size (8B → off-slot read → SIGSEGV; 16B →
wrong bytes). So beyond the call-site result-TYPE override (facet 1, mirrors the 4560 float-return
fix), the real defect is in **CODEGEN return-value ABI**: a mono'd generic METHOD (has a `self`
receiver) returning a >8B `T` by value does not emit/transfer the 16-byte return correctly. Contrast:
a NON-generic method→str works (ng.ax=5), a plain generic fn→str works (id[str]=5) — only
generic + receiver + >8B-return breaks. FIX AREA is air_builder/x86 return lowering for the mono'd
method (likely the sret/16B-return path interacting with the self param), NOT the typecheck sites
already read. Genuine fresh codegen session; gate B==C. This is the ONE bug this session that has no
working reference to mirror and lives in an unexplored (return-ABI) area — hence banked, not rushed.
**GENERAL to any >8B `T`, not str-specific** (`s16.ax`: generic method → 16-byte struct P16 also
SEGVs; `s16n.ax`: non-generic method → P16 works =42). The substitution machinery exists
(`substitute_type_params` mono.ax:70; subst map built mono.ax:503-522 from the struct/generic params),
and the i64 (8B) return works — so `T` IS substituted in the return node, yet the >8B return still
mis-transfers. ⇒ defect is NOT the substitution but how the mono'd function's return type_id/WIDTH
reaches the return codegen. **NEXT SESSION (trace, bash grep only):** dump the mono'd `get[<T>]`'s
registered return type_id (is it the concrete 16B type, or a stale 8B placeholder?) at pre_infer of
the instance + at air_builder's return lowering; the fix is wherever the width defaults to 8B for a
mono'd generic method's >8B return. Repros: `s16.ax`(SEGV, 16B struct)/`s16n.ax`(=42, non-generic)/
`use.ax`/`use2.ax`/`asg.ax`. Fully characterized; needs mono/codegen trace instrumentation.
**✅ TRACED TO THE EXACT SYMPTOM (2026-07-24e, RETDBG at pre_infer_func_signature:3098, bash grep):**
a mono'd `get` instance registers **`ret_type` = a GENERIC placeholder (kind 7 TYPE_KIND_GENERIC,
size 0)** — e.g. type_id 53 — instead of the concrete return type; contrast a WORKING mono'd
`get[i64]` which registers `ret_type=4 (i64), kind 0 PRIMITIVE, size 8`. So the mono'd generic
method's **return type is never resolved from the generic `T` to the concrete type** (stays kind-7,
size-0); codegen then emits the return on a wrong/default width → SIGSEGV / garbage. ROOT is that when
`T` is inferred from a NESTED param position (`self: ptr[Box[T]]`) rather than a direct value param
(`x: T`, which `id[T]` has and works), the return-type node's `T` is not substituted/resolved to the
concrete type in `pre_infer_func_signature` (the return `infer_node` at :3082 yields the kind-7
generic). **RAZOR-SHARP FIX TARGET:** in the mono instantiation, ensure the cloned function's
return-type node `T` is substituted to the concrete type (like the param `T`s are), OR in
`pre_infer_func_signature` resolve a kind-7-generic `ret_type` via the instance's type-arg map. Verify
the mono'd `get[str]`/`get[P16]` then registers a size-16 return. Trace with **bash grep only**
(RETDBG recipe above). Delicate (mono/generic-inference, self-host-critical) → gate B==C + full
regression + oracle `t_genmethodbig`(str + 16B-struct return). Fresh focused session.
**⚠️ CORRECTION — the "substitution failure" reading was a RED HERRING (over-broad filter).** The
RNODE trace filtered on the NAME `"get"`, which ALSO matches the bundled stdlib's `get` methods
(`Vec.get`, `HashMap.get`, …) — legitimately generic, so THEIR return `T` shows kind-7/rettype-53.
Those were the lines I read, NOT the user's `get[str]`. A deeper SUBTRACE inside `substitute_type_params`
(printing `lookup_subst` for every `NODE_TYPE_EXPR` named `"T"`) shows **`lookup_subst("T")` DOES
return concrete types — 12 (str), 4 (i64), etc.** So **substitution WORKS**; the subst map has `T→str`.
⇒ the >8B segfault is NOT a subst/return-type-resolution failure. The mono'd `get[str]` most likely
DOES register a str (16B) return; the defect is downstream in the **16-byte-return CODEGEN for a
method with a receiver** (air_builder/x86 return lowering, or the call-site materialisation of the 16B
result) — consistent with the earlier facts (non-generic method→str works, plain generic fn→str works,
`asg.ax` unused-result works, `use2.ax` annotated gives garbage). **NEXT (focused session, PRECISE
filter):** distinguish the USER's `get` from stdlib (filter by the mangled instance name or the decl
token position, NOT the bare name `"get"`), then trace the mono'd user `get[str]`'s return type_id AND
the emitted return sequence in air_builder/x86 for that one function. ⭐ LESSON: filtering a compiler
trace by a COMMON name (`get`/`len`/`v`) silently mixes in stdlib and yields wrong conclusions — filter
by the unique mangled instance name. See [[lesson-bash-grep-not-powershell-selectstring]] (same family:
trust the data only after the filter is proven specific). Kept 8B-`T` working (shipped `ccfd1af7`);
this is the >8B residual only, and it lives in return CODEGEN, not mono substitution.

## (history) OPEN — inline method on a GENERIC struct segfaults when called
**Probe-found on driver `cf42579b`, 2026-07-24e.**
```
struct Box[T]:
    v: T
    fn get(self: ptr[Box[T]]) -> T:   // INLINE method (in the struct body)
        return self.v
fn main() -> i64:
    let b = Box[i64](v: 42)
    return b.get()                     // SEGFAULT (want 42), O0 and O1
```

## Precisely characterized (probe matrix)
| case | shape | result |
|---|---|---|
| p4a | `Box[T]` generic, DIRECT field `b.v` (no method) | ✅ 42 |
| p4b | NON-generic `Boxi` + inline method `get()` | ✅ 42 |
| p4c | `Box[T]` generic + FREE-fn method `fn get[T](self: ptr[Box[T]])` | ✅ 42 |
| **p4** | `Box[T]` generic + **INLINE** method `get()` returning `self.v` | ❌ **SEGV** |
| **p4d** | `Box[T]` generic + **INLINE** method returning a CONSTANT `7` (no T) | ❌ **SEGV** |
⇒ The crash is NOT about the generic `T` in the body (p4d has none) — it is the **INLINE-method ×
GENERIC-struct** combination itself. Calling the method dispatches to a wrong/unmonomorphized address.
**WORKAROUND (works today):** write the method as a FREE FUNCTION `fn m[T](self: ptr[Box[T]]) -> ...`
outside the struct body (p4c).

## ROOT PINPOINTED (read-only, 2026-07-24e) — parser drops the struct's generic params
`parse_struct_decl` (parser.ax:1367): the struct's generic params `[T]` are parsed at 1385-1389
(gp node, `FLAG_IS_GENERIC` set on the STRUCT node). Inline methods are parsed at **1408-1418** via
`parse_func_decl(...)` and appended as struct children — **but they DON'T receive the struct's `[T]`
or `FLAG_IS_GENERIC`.** So `fn get(self: ptr[Box[T]])` is a NON-generic function that merely mentions
`T`; mono never instantiates it for `Box[i64]`, so `b.get()` calls the un-instantiated template body
with a wrong layout → SIGSEGV. The free-fn form `fn get[T](self)` works because it carries an explicit
`[T]` → normal `mono.instantiate_function` path.

## Fix plan (dedicated session — likely A==B inert on self-host, which uses the free-fn workaround)
MULTI-PART (tractable but not a one-liner):
1. **Parser (parser.ax:1408-1418):** when the struct is generic (gp != 0), give each inline method the
   struct's generic params + mark it generic. `clone_subtree` (ast.ax:208) can clone the gp subtree;
   insert it at the METHOD node's correct child position (generic-params come FIRST, before params —
   check parse_func_decl's child order; append_child adds at END so a front-insert helper or building
   the method with gp first is needed) and `set_flags(m, FLAG_IS_GENERIC)`.
2. **Dispatch/mono:** verify `b.get()` on a `Box[i64]` receiver resolves to the method AND infers
   `T=i64` from the receiver so mono instantiates `get[i64]` (the free-fn form does this via UFCS +
   arg-type inference; confirm inline-method dispatch reaches the same instantiation, else route it
   there). Trace with **bash grep** ([[lesson-bash-grep-not-powershell-selectstring]]).
Gate: A==B expected (compiler declares no inline method on a generic struct — it uses the free-fn
form, so the change is inert on the self-host); full regression + oracle `t_genstructmethod`(42) +
a 2-instantiation oracle (`Box[i64]` and `Box[str]`) to prove per-instance mono. B==C if any codegen
shifts.

## (superseded) Fix direction (dedicated session — NOT yet attempted)
An inline method in a struct body is (RFC 0002 `current_struct`) parsed with an implicit/typed `self`
and attached to the struct. For a NON-generic struct it monomorphizes trivially (p4b works). For a
GENERIC struct `Box[T]`, when `Box[i64]` is instantiated, the struct's inline methods must ALSO be
monomorphized with `T=i64` and their `self: ptr[Box[T]]` bound to `ptr[Box[i64]]` — this apparently
does not happen (or the call binds to the un-instantiated template body), so `b.get()` calls a bad
address → SIGSEGV. The free-fn form works because `fn get[T]` goes through the normal generic-free-fn
monomorphization path (mono.instantiate_function). Likely fix: when monomorphizing a generic struct
instance, also instantiate its inline methods (mirror the free-fn `[T]` path), and make method-call
resolution (`b.get()`) target the instantiated method for the concrete `Box[i64]`, not the template.
INVESTIGATE: how inline methods are registered on a struct (parser `parse_struct_decl`) vs how
`b.get()` resolves the callee (typecheck method dispatch) vs mono of generic struct instances. Gate:
B==C (touches mono/dispatch = self-host-critical) + regression + oracle `t_genstructmethod`(42).
Verify traces with **bash grep**, never PowerShell Select-String ([[lesson-bash-grep-not-powershell-selectstring]]).
Repro `/tmp/probe4/p4.ax`, `p4d.ax`; contrast `p4b`(non-generic ok)/`p4c`(free-fn ok).
