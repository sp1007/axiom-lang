---
name: bug90-option-method-segfault-open
description: "BUG#90 PARTIAL FIX — Option/Result .unwrap()/.is_some()/.is_ok() on a TYPED receiver now lower to deterministic pointer-tagged inline logic (backend intercept), fixing the crash/flakiness. REMAINING: an inline UNTYPED receiver (m.get(k).unwrap()) still needs typecheck to propagate the mono'd return type."
metadata: 
  node_type: memory
  type: project
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

🟢 **BUG#90 PARTIAL FIX** — Option/Result methods on a TYPED receiver now deterministic + correct.

**Fixed (backend intercept, `air_builder.ax` `lower_call_expr`):** `unwrap`/`is_some`/`is_none`/`is_ok`/`is_err` on a receiver whose type is Option/Result (kind 11/12, or generic_inst/sum whose base name starts "Option"/"Result", via `opt_res_kind`/`name_has_prefix`) are emitted as the SAME deterministic pointer-tagged inline logic `lower_match_tagged` uses:
- `is_some`/`is_ok` ⇒ `recv != 0` / `(recv & 1) == 0`; `is_none`/`is_err` inverse.
- `unwrap` ⇒ `OP_DEREF(recv, pl_type)`, pl_type = the call's own result type (Some/Ok are bit0=0 so no mask needed).
This bypasses `std/result.ax`'s `(&self as ptr)[0]` read, which crashed when self's param home-slot didn't hold the box pointer. Off the stdlib ABI (RFC 0012 tactic).

**Gate:** A==B==C bit-identical (`f8460d00`) — the compiler's own 16 `.unwrap()` sites are on UNTYPED mono'd results, so the intercept doesn't fire there → self-codegen unchanged, ZERO risk. Regression **129/129**. Oracle `t_optmethod` (42, 5/5 deterministic: Option/Result unwrap+is_some+is_none+is_ok+is_err). Simple Option/Result unwrap that used to be flaky (o1) is now deterministic.

**UPDATE (next iteration): get/unwrap crash class is effectively CLOSED.** Re-probed exhaustively: `m.get(k).unwrap()` standalone (u2), unannotated local (u1), two gets (u4), `is_some`+`unwrap` (hm), typed local (hm3) — ALL work 5/5 deterministically. The earlier "hmc still segfaults" was a MISDIAGNOSIS: the crash was `m.contains(k)` (see below), NOT `.unwrap()`. So BUG#90 needs no further work for realistic Option/Result method usage.

**Discovered while re-probing → BUG#93 (separate, OPEN)** [[bug93-contains-method-resolution-open]]: `HashMap` has NO `contains`/`contains_key`, so `m.contains(k)` on a HashMap mis-resolves to `HashSet.contains` (same name, wrong self type) → segfault. Method resolution is **self-type-blind** for same-named generic methods (adding a HashMap.contains breaks HashSet dispatch). ALSO fixed this iteration: `HashSet.contains` itself returned true for ANY key because it used INLINE match arms (`Some(_): true`) which mis-lower [[inline-match-arm-unsupported]] — rewrote to block-form.
