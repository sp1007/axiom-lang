---
name: bug-generic-struct-inline-method
description: "OPEN bug (probe-found 2026-07-24e): calling an INLINE method (defined in the struct body) on a GENERIC struct instance SEGFAULTS — regardless of the method body (even `return 7`). Non-generic inline methods work; the SAME method written as a free function `fn m[T](self: ptr[Box[T]])` works (the workaround). So inline methods on generic structs are not monomorphized/dispatched correctly for the instance."
metadata:
  node_type: memory
  type: project
---

## ✅ FIXED 2026-07-24e (8-byte-T cases) — driver `ccfd1af7`, 534/534, B==C
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
