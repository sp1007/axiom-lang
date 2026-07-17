---
name: bug89-for-continue-increment
description: "BUG#89 — `continue` inside a `for` loop hung (skipped the body-tail increment). Fixed by emitting the +1 inline before continue's jump-to-cond, keeping the original CFG (no latch)."
metadata: 
  node_type: memory
  type: project
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

✅ BUG#89 FIXED — `continue` inside a `for` loop advances the counter.

**Symptom:** `for i in 0..n: if cond: continue; ...` HUNG (both range and fixed-array for-loops). Found by probing the freshly-shipped [[rfc0018-for-in-array-shipped]]. `break` was fine; `continue` in a `while` loop was fine.

**Root cause:** `lower_for` emits the counter increment (`i += 1`) at the BODY TAIL, then jumps to cond. `lower_continue` jumps straight to `loop_conds` top (= cond block), BYPASSING the tail increment → counter never advances → infinite loop. Pre-existing, independent of RFC 0018.

**Fix that FAILED (latch block):** first attempt gave `for` a dedicated latch block (increment + jump cond) that both `continue` and fall-through target. Correct at -O0, but at **-O1 the latch was emptied** (LICM/block passes stripped its increment AND terminator → fell through to exit → wrong/zero result). The new back-edge CFG shape tripped the optimizer — same class as BUG#86. **Lesson: introducing new block shapes mid-loop is an -O1 hazard; avoid when a CFG-preserving fix exists.**

**Fix that SHIPPED (inline increment):** keep the ORIGINAL CFG (no latch, body-tail increment, `continue`→cond). Added parallel stacks `loop_incr_regs`/`loop_incr_types` on `FuncLowering` (pushed in `lower_for` = counter iter_reg + counter_type; `lower_while` pushes reg 0 = none). `lower_continue`: if top incr_reg != 0, emit the same `+1` (ICONST/IADD/COPY) INLINE before the jump-to-cond. No new blocks → optimizer sees the same CFG it already handles → no -O1 interaction. Continue-free loops emit byte-identical AIR.

**Gate:** backend/IR-lowering, A==B==C bit-identical (**new daily-driver hash `1ee404b6`**, was `D6DBA89D`); regression **122/122**. Oracle `t_forcontinue` exit **52** (range 0..6 skip 3 = 12; array [10,20,30,40] continue@20 break@40 = 40). Files: `air_builder.ax` (`lower_for`/`lower_while`/`lower_continue` + struct fields).
