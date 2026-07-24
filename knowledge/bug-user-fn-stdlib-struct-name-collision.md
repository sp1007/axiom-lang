---
name: bug-user-fn-stdlib-struct-name-collision
description: "OPEN (medium) — a top-level user FUNCTION whose name matches a bundled stdlib STRUCT name silently mis-links. `fn worker(...)` collides with `struct worker` in std/scheduler.ax (bundled), so `call worker` resolves to a different symbol; the user fn is never called → wrong result / crash. 5th hole in the free-fn/stdlib collision family (prior 4 were fn-vs-fn). Workaround: don't name symbols after bundled stdlib names. Repro bin/known_fail_worker_name_collision.ax."
metadata:
  node_type: memory
  type: project
---

**OPEN, medium severity, pre-existing.** Found 2026-07-24 during RFC 0033 Phase C (threads).
A new hole in the free-fn/stdlib collision family ([[bug-freefn-stdlib-collision-noarg]], whose
4 fixed holes were all fn-vs-fn): this one is **user FUNCTION vs bundled stdlib STRUCT name**.

## Symptom
`fn worker(n: i64) -> i64: ...` + `worker(5)` → main's `call worker` links to a DIFFERENT
symbol (a bundled stdlib function ~`0x140007c77`), so the user `worker` is never called. The
user body (incl. any global write) is correct but dead. Result: wrong value / crash depending on
what the mis-linked target does. Repro `bin/known_fail_worker_name_collision.ax` (want 105, gets 8).

## Root
`std/scheduler.ax:113` defines `pub struct worker:` and it is bundled into EVERY program
(concatenate_stdlib). A user top-level `fn worker` shares the name; the call resolves/links to
the bundled symbol instead of the user function. Confirmed by disassembly: `worker`'s real code
(at its own address) is never the target of any `call`; main calls the stdlib symbol. Renaming
the user function to anything distinctive → correct (verified: `zqmyuniquefn` → right answer).

## ⭐ CRITICAL LESSON — minimization fooled by a co-varying name
This was first mis-diagnosed as a "param + non-void return + module-global-write miscompile"
because every FAILING minimal test happened to be named `worker` while every PASSING one was
named `setit`/`dbl`/`wf`. The name was the real independent variable; the param/return/global
features were coincidental. **When minimizing, hold the SYMBOL NAME fixed (or vary it explicitly
as its own axis) — a name that collides with bundled stdlib is invisible in the feature matrix.**
The disassembly (main calling the wrong address; nothing calling the user fn) is what broke the
false theory — cf. the standing rule "đổi TÊN hàm khi thu hẹp repro" ([[bug80-free-call-overload-collision]]).

## Impact / fix direction
- Blocked nothing that matters once names are chosen well; RFC 0033 Phase C threads SHIP fine
  with a distinctively-named entry (do NOT name a thread entry `worker`).
- Proper fix = extend the collision handling to also disambiguate user free-fn names against
  bundled STRUCT (and other non-fn) symbol names — same MODDUP-mangle / resolver approach as the
  fn-vs-fn holes, but the resolver must consider struct/type symbols too. Frontend; A==B likely
  (compiler self-build doesn't define a `fn worker`). Add an oracle once fixed.
- Lower priority than it looks: the workaround (avoid bundled names) is trivial, and the set of
  bundled names is small. But it is a genuine silent-miscompile footgun.

Related: [[bug-freefn-stdlib-collision-noarg]], [[bug80-free-call-overload-collision]], [[rfc0033-concurrency-v1]].
