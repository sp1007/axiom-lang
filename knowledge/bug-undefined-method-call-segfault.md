---
name: bug-undefined-method-call-segfault
description: "FIXED: calling an undefined method/UFCS name (p.ghost()) was silently accepted then segfaulted; typecheck now rejects. Fixpoint 8905FB39, 147/147."
metadata: 
  node_type: memory
  type: project
  originSessionId: b9556362-e358-4881-8a42-df3ba54b80bb
---

✅ **Undefined-method call silent-segfault FIXED** — 2026-07-11. daily-driver = A==B **`8905FB39`**, `origin/main`=`1fa2d63`, **147/147**.

`recv.name(...)` where `name` is neither a data field, a real method, nor a free
function usable via UFCS (`fn name(self: T, ...)`) was **silently accepted**:
typecheck left the call UNTYPED (`callee_type == UNKNOWN`, method flag 2048 unset)
and `air_builder.lower_call_expr`, finding no method/UFCS match (loop ~1565), fell
through to `lower_expr(callee)` → `getfld` of a non-existent field (garbage reg) →
indirect `call reg` → **SIGSEGV**, no diagnostic. Same accept-then-miscompile class
as BUG#53/68/81/93. Repro: `p.ghost()`, `x.nonexistent()`, `s.totallyfake()`.
Contrast: `p.x()` where `x` IS a field already gave "field, not a function".

**Fix** = REJECT in `typecheck.ax` (after the BUG#61 return-type recovery block, ~line 2917).
Gate: `result_type==UNKNOWN and callee_type==UNKNOWN and callee.kind==NODE_FIELD_EXPR
and (flags & 2048)==0` + receiver known & not INTERFACE.

⚠️ **KEY LESSON (first attempt broke self-host):** a receiver-TYPE-matched check
(method_ret_type on the inferred receiver type) FALSE-POSITIVES — a cross-module
param type (`in_file: File` from std.io) still reads as a placeholder ("R") during
inference, so `in_file.read()` (valid UFCS → `pub fn read(mut self: File,...)`) got
wrongly rejected and A!=B (stage A errored building itself: "no method 'read' on
type 'R'"). Air_builder resolves UFCS with FULL type info later; typecheck-time
receiver types are incomplete for cross-module types.
→ Fix uses a **RECEIVER-AGNOSTIC** test: reject only when NO `SYM_FUNC` anywhere
matches the name (`self.intern.name_matches_method`). If any same-named function
exists, defer to normal resolution (conservative, never breaks a real call).
`p.ghost()` (no `ghost` fn) rejects; `in_file.read()` (a `read` fn exists) passes.

Oracle `t_nomethod` (reject-mode). Frontend-only → A==B. Surfaced by bug-probe.
Related: [[bug93-qualified-str-call-segfault]] (module-qualified sibling), [[inline-match-arm-unsupported]].
