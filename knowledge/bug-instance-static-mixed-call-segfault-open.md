---
name: bug-instance-static-mixed-call-segfault-open
description: "✅ FIXED 491575a: static-form method call `Type.method(recv,args)` lowered the type name as a bogus receiver reg, shifting all args by one slot -> SIGSEGV for 2-param/two-static/instance+static-mix. Fix = skip receiver synth when receiver symbol kind is a type. Oracle t_staticcall(41), A==B, 165/165."
metadata:
  node_type: memory
  type: project
  originSessionId: fb0bd2b4-5b71-43e6-95c8-daa4ef9a6f1e
---

✅ **FIXED 2026-07-11 `491575a`** (A==B fixpoint `8F062917`, 165/165). ROOT: static-form
call `Type.method(recv, args)` — `air_builder.lower_call_expr` always lowered the
`receiver_node` as the `self` register. For a static call the receiver is the bare TYPE
name, so it lowered a bogus value and pushed it as an EXTRA first arg (`tempc` = params+1),
shifting every real arg by one slot → arg/home-slot corruption. Instance-only had
`recvsymk=SYM_VAR(0)`; static had `recvsymk=SYM_STRUCT(2)`. Fix: skip receiver synth when
the receiver symbol kind is a type (SYM_STRUCT/SYM_TYPE_ALIAS/SYM_INTERFACE); explicit args
carry `self`. Oracle `t_staticcall` (exit=41: instance + 2 static + 2-param static + mix).
**LESSON:** air_builder debug prints via `ax_printf_local` starting with `[D...` (e.g.
`[DBGTOP]`) are SWALLOWED by `is_verbose_debug` (whitelists only known `[Debug] Stage/…`
prefixes) — use a NON-bracket prefix like `ZZSTATIC` or they're INERT (same trap as the
run_type_checker/ssa_opt debug-print lessons).

Original diagnosis (2026-07-11), found while building the `t_methodok` oracle for the
method-arity fix [[bug-accept-then-miscompile-cluster-0711]]:

**Symptom.** A function that calls the SAME struct's method both ways segfaults:
```
struct Pt:
    x: i64
    fn getx(self) -> i64: return self.x
pub fn main() -> i32:
    let p = Pt(x: 7 as i64)
    let a = p.getx()      // instance
    let b = Pt.getx(p)    // static (receiver named as bare type)
    return (a + b) as i32 // SIGSEGV (exit 139) at BOTH -O0 and -O1
```
- INSTANCE-only (`p.getx()`) → OK. STATIC-only (`Pt.getx(p)`) → OK (returns 7).
- The MIX (both in one fn) → SIGSEGV. Confirmed on the OLD daily driver `bin/axc_native.exe`
  too → **PRE-EXISTING**, NOT caused by the typecheck arity change (which only adds rejects).

**Likely area.** Static-call `Type.method(recv)` codegen in air_builder/x86 — the two call
forms probably resolve the same method symbol to two different call shapes (self as
first_child receiver vs. self as an explicit arg), and mixing them corrupts arg-slot /
home-slot layout. Static method call is a niche form; instance is the common one.

**Status.** Deferred — needs backend investigation (B==C gate mandatory before commit).
Kept OUT of the `t_methodok` oracle (which now uses instance-only calls). Not urgent:
the self-host compiler uses instance calls; static-call syntax is rare.
