---
name: probe-closure-capture-guard-sound
description: The zero-capture-only closure restriction (RFC 0008 P2 unimplemented) is soundly enforced — every capturing form rejects cleanly, none miscompiles. Clean probe, do not repeat.
metadata:
  type: project
---

**Probed 2026-07-23, CLEAN — no bug.** Capturing closures are not implemented (RFC 0008 P2),
and the restriction is enforced SOUNDLY: every form that captures an enclosing-scope binding is
rejected with `error: closure captures 'x' from an enclosing scope; only zero-capture closures
are currently supported`, before code generation. None miscompiles.

Forms tested, all correctly rejected:

- capture a `let` local, a fn param, a struct field (`c.base`), a loop variable
- capture inside arithmetic, as a call argument, via array index, through a method call on the
  captured var
- capture in a nested closure (inner captures outer`s capture)
- a closure calling another capturing closure

Forms that work (genuinely zero-capture, correct):

- `[|x| -> i64 x*2, |x| -> i64 x*4]` — array of closures, called by index
- closures using only their params and literals

The guard recurses into every expression sub-form, so there is no "captured in an unusual
position" evasion — the class this probe was looking for (cf. the match-scrutinee family, where
`_`/binding arms slipped a literal-only reject). It does not exist here.

**One forward-looking caveat, not a current bug:** block-BODY closures (`|x| -> T:` then an
indented block) do not PARSE — they fail with `expected expression`, so the capture check never
runs on a block body. Moot today. But if block-body closure syntax is ever added, the
capture-lift pass must be re-verified to recurse into block STATEMENTS, not just the single
expression it handles now; otherwise a capture inside a block body could reach codegen
unchecked. Worth a test at that point, nothing to do until then.
