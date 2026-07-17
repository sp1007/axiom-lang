---
name: bug-f32-boundary-abi-open
description: "✅ FIXED c7f6b78 (2026-07-16, A==B 82CC482F, 333/333): f32 silently miscompiled at fn boundaries (param-in / return-out read as 0). NOT a backend/ABI bug as first guessed — root was FRONTEND: NODE_FLOAT_LIT always typed TYPE_F64, so `let v: f32 = 42.0` stayed double-repr; fine locally, wrong at the f32 ABI (movss reads low 32 of the double). One-line fix: coerce float lit to f32 when expected==F32. Oracle t_f32boundary."
metadata:
  node_type: memory
  type: project
  originSessionId: 1c44f2f4-e5c7-43b5-aac0-3f73f5f7eb36
---

# ✅ FIXED `c7f6b78` — f32 miscompiled across function boundaries (silent → 0)

**FIX (frontend, A==B `82CC482F`, 333/333):** the real root was NOT the backend ABI
(my initial guess below was wrong). `typecheck.ax infer_node` had `NODE_FLOAT_LIT ->
result_type = TYPE_F64` UNCONDITIONALLY, ignoring the expected type — unlike the
integer-literal path right above it which coerces to `expected`. So a float literal in
an f32 context (`let v: f32 = 42.0`, an f32 param/return arg) stayed TYPE_F64;
air_builder's `lower_float_lit` reads `node_types[lit]` and materialized it as a DOUBLE
(OP_FCONST 64-bit, MOVDQ pad=8 — CONFIRMED by a MachInst dump: pad 8 before, pad 4
after). Every LOCAL op treated it consistently as f64 (so `v as i64` = cvttsd2si worked)
but the f32 ABI boundary reads XMM0 as single (movss = low 32 of the double) -> 0. One-
line fix: `if expected == TYPE_F32: result_type = TYPE_F32 else TYPE_F64`. Now the
literal is genuinely single (movd) so all downstream single-precision ops (return move,
param, cvttss2si cast) align. Fixpoint-safe: compiler source has no f32 literals ->
f64-expected path unchanged. Oracle `bin/t_f32boundary.ax`. LESSON: disassembly/MachInst
dump (`[dbgmi]` in x86_coff.ax func loop, print op/padding/operand kinds) pinned the
DOUBLE-width constant instantly; the "backend ABI" hypothesis below was a red herring —
the boundary MOV was fine, the VALUE was double.

**FOLLOW-UP `4f0df4e`** (frontend, A==B `2B240CC6`, 334/334) — same root in more contexts. The `c7f6b78` fix covered let/param/return, but a float literal in a STRUCT-CONSTRUCTOR field (`P(x: 40.0)`) or a FUNCTION ARG (`takef(40.0)`) still stayed f64 → stored DOUBLE (for a 4-byte f32 field this overflows the slot AND reads back wrong). Fix: the struct-ctor arg coercion (`typecheck.ax ~3853`) and the non-generic call-arg coercion (`~3910`) already thread the target type as `expected` for ARRAY/TUPLE fields/params — extended both to `elif cand/fty == TYPE_F32: vexp/pexp = fty`. Scoped to F32 → fixpoint-safe. Oracle t_f32aggregate. Probe matrix: sfield/sret/nf FIXED; sarr/scmp already OK.
✅ **GENERIC-ARG FIXED `6eb63db`** (frontend, A==B `850458B4`, 335/335) — `Vec[f32].push(40.0)` (push[T], T=f32 from receiver) now works. Two changes reusing the RFC 0022 tuple×generic machinery: (1) the generic-call arg re-inference loop (`typecheck.ax ~3460`), after `inferred[]` is final, re-infers an arg whose bare-generic param resolved to `TYPE_F32` with f32 expected; (2) a STICKY GUARD on NODE_FLOAT_LIT (`~4196`): on a later `expected==UNKNOWN` re-visit, keep a prior f32 pin instead of reverting to f64 (WITHOUT this the re-infer set f32 but a hint-less re-visit reverted it → still DOUBLE; found via MachInst dump). Oracle t_f32generic. 🔴 REMAINING (small, documented): EXPLICIT `f[f32](40.0)` — callee is a NODE_INDEX_EXPR, so `is_generic_call` (set only for IDENT/FIELD_EXPR callees) is FALSE and it takes a separate explicit-instantiation path this fix doesn't reach (probe `gf2`=0). Rare (explicit type-arg + float literal); defer.

---
(original OPEN diagnosis kept for reference — the backend-fix sketch was NOT needed)

# 🔴 (was OPEN) — f32 miscompiles across function boundaries (silent → 0)

Found 2026-07-16 while probing the tail of the free-fn/stdlib collision cluster
([[bug-freefn-stdlib-collision-noarg]] — that cluster is CLOSED; this is a SEPARATE,
pre-existing bug it uncovered). Daily driver at discovery: `018d49c`, A==B `489CDDAE`.

## Symptom (all silent wrong answers, exit 0/garbage — NO diagnostic)
Minimal isolations (each `main` returns the value `as i64`, want 42):
- `f_id`   `fn idf(x: f32) -> f32: return x` ; `idf(42.0f32)` → **0** ❌
- `f_pin`  `fn pin(x: f32) -> i64: return x as i64` ; `pin(42.0f32)` → **0** ❌ (f32 PARAM doesn't arrive)
- `f_ret`  `fn mkf() -> f32: return 42.0` ; `mkf() as i64` → **0** ❌ (f32 RETURN wrong)
- `f_ar`   local only: `let a:f32=41.0; let b:f32=a+1.0; return b as i64` → **42** ✅ (LOCAL f32 arithmetic fine)
- `iso2`   `fn xmap(x: f64)->f64: return x+1.0` ; `xmap(41.0f32)` → **42** ✅ (f64 param + f32→f64 arg COERCION works)
- `d_id`   f64 identity across boundary → **42** ✅ (f64 boundary fine)

**So: f64 at boundaries = OK; LOCAL f32 = OK; only f32 crossing a call boundary (param-in OR return-out) is broken.**

## Diagnosis (read-only; not yet fixed)
XMM registers hold f32 in the low 32 bits. The boundary code moves float values with a
64-bit `MACH_MOV` (movq) which is bit-exact for a *properly single-precision* value — but
two things go wrong for f32:
1. **Return path** `x86_selector.ax ~1406-1411` (OP_RETURN, non-16byte): for `ret_type ==
   9 (f32) or 10 (f64)` it emits `MACH_MOV inst.src1 -> XMM0`. If the returned value is a
   *literal* (`return 42.0`) that was materialized as an f64 double bit-pattern (literals
   default to f64), moving it to XMM0 and having the CALLER read it as f32 (movss = low 32
   bits of the double) yields ≈0. No `cvtsd2ss` narrows an f64-typed value to f32 before
   the XMM0 move, and no width is carried on the return move.
2. **Param prologue** `x86_selector.ax ~2166-2207` (emit_param_prologue): snapshots the
   float arg reg via movq and restores to the param vreg via movq. For `pin(x: f32)` the
   value should arrive as single in xmm0; `f_pin`→0 says either the CALLER placed the f32
   arg wrongly (call-site arg setup, search the CALL emission for the float-arg move —
   `abi_float_arg_reg`, ~1519 handles `arg_type == 9 or 10` together) or the param read /
   subsequent `x as i64` (cvttss2si) operates on the wrong reg/width.

Root theme: the backend treats f32 and f64 IDENTICALLY at boundaries (both just "a float →
xmm via movq"), but f32 needs single-precision awareness — width-4 moves (movss) and
cvtsd2ss/cvtss2sd conversions where an f64-typed value meets an f32 slot (and vice-versa).
Contrast LOCAL f32 which works because RFC 0006 carries width in `padding` (4 = single) on
the FADD/FMUL/ITOF/FTOI/MOVDQ MachInsts (`x86_selector.ax:52-53`) — the boundary MOV/RET/
CALL-arg paths do NOT carry/honor that width.

## Fix sketch (BACKEND → B==C MANDATORY, own session)
- Return: when `ret_type == TYPE_F32`, ensure the value in XMM0 is single — if the value's
  type is f64, emit `MACH_CVTSD2SS` before the XMM0 move; and/or make the XMM0 move width-4.
  Simplest correct: narrow at the `return` lowering when the returned expr type != declared
  f32 ret. Also the CALLER must read a f32 return as single (movss) — check the call-result
  read path for a float ret_type.
- Param: verify the call-site float-arg setup narrows an f64 value to f32 for an f32 param
  (cvtsd2ss) and that same-width f32→f32 doesn't accidentally movsd garbage high bits; and
  that the param prologue movq→param-vreg + downstream use honor single width.
- Add oracles: t_f32param (f32 param in), t_f32ret (f32 return), t_f32roundtrip. Gate B==C.

## Guardrails
Backend/ABI change → A!=B expected, require hand-built **B==C** + full regression BEFORE
commit (CLAUDE.md §24). Do NOT "fix" by making everything f64 (f32 is a real type, RFC 0006).
Low day-to-day impact (idiomatic AXIOM float code uses f64), but it is a SILENT miscompile
(BUG#53 accept-then-miscompile class) so it must be fixed or the f32-at-boundary case
rejected with a diagnostic if a full fix is deferred.
