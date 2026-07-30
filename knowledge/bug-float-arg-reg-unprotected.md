---
name: bug-float-arg-reg-unprotected
description: "LATENT (not reachable on HEAD today) — param_idx_of_vreg protects incoming INTEGER arg registers (RCX/RDX/R8/R9) from being colored/clobbered by an earlier param, but does NOT list XMM0-XMM3, so the float side has no backstop. The float spill scratch XMM2 IS the 3rd float arg register on win64."
metadata:
  node_type: memory
  type: project
---

# LATENT HOLE — incoming FLOAT arg registers have no allocator protection

Status: **LATENT, deliberately not fixed** (2026-07-30). Nothing on HEAD generates the shape that
reaches it. Recorded because peephole 1f reached it within one session of being written, and the
next thing that produces `MOV vFloat, PHYS(xmm_k)` will reach it again — silently.

## The hole

`x86_regalloc.ax:519` scans for `MOV vreg, PHYS(r)` and records `param_idx_of_vreg[vreg]`. At
coloring time (`:915`) that index forbids the vreg from taking any **later** parameter's incoming
arg register, so an earlier param's value cannot clobber a later param before it is read.

The register list is **integer only**:

    win64:  RCX RDX R8 R9
    sysv:   RDI RSI RDX RCX R8 R9

**XMM0–XMM3 are absent.** A vreg defined from a float arg register therefore gets `p_idx == -1`
and no protection at all.

Compounding it: the **float spill scratch is XMM2** (chosen 2026-07-29e when spilled float ALU
destinations were fixed — R10/R11 alias XMM10/XMM11 which are allocatable, so XMM2 was picked).
On win64 **XMM2 is also the third float argument register**. So the hazard is reachable through
spilling, not only through coloring.

## Why HEAD is correct today anyway — and it is STRUCTURAL, not luck

`emit_param_prologue` (`x86_selector.ax:2436`) snapshots **every** incoming register-passed
argument into a fresh vreg in phase 1, **before** phase 2 reads any of them. Its own comment
(`:2430`) states this is the point: "after this phase all original arg registers are safely
captured and later value-loading scratch may use any register."

Float params route through **GPR-classified** snapshots — `is_float_vreg` is false for a
selector-invented temp with no AIR def, so it cannot be colored XMM at all, so no XMM arg
register can be touched before all four are captured. That is what makes the missing XMM entries
in `param_idx_of_vreg` invisible.

⚠️ So the correctness of float parameter passing rests on **prologue ORDERING**, enforced nowhere
except by that pass's shape. Any pass that reorders, folds, or deletes the phase-1 snapshots
breaks it. See [[session-handoff-2026-07-30c]].

## Proof it is real (observed, not reasoned)

Peephole 1f folded the snapshots away. Disassembly of `ip_catmull_rom`:

    movsd %xmm0,%xmm2      ; v1 spilled, through the float spill scratch
    movsd %xmm1,%xmm2      ; v2 spilled, same scratch
    movsd %xmm2,%xmm9      ; v3 := p2 -- reads a CLOBBERED %xmm2

Symptom: `t_interpolation` 127→79, `t_colorhsl` 127→120, `t_quatrot` 8→3, identical at O0/O1/O2.
Minimal repro pinned as `bin/t_floatparamchain.ax` (42 correct; 1 on the unguarded build).

⭐ **Float PRESSURE was ruled out by measurement, and was the wrong suspect.** A function with 18
simultaneously-live f64 values compiles correctly on HEAD (spills fine). Four f64 params with a
trivial body are fine under 1f. The break needs 4+ register-passed float params AND enough
pressure to spill a param vreg.

## If it is ever fixed

Add XMM0–XMM3 (win64) / XMM0–XMM7 (sysv) to the `param_idx_of_vreg` scan, and make the forbid
loop at `:915` walk the float arg registers for a float param rather than `abi_int_arg_reg`.
Backend change ⇒ **B==C mandatory**. Then also reconsider whether the float spill scratch should
be a register that is never an arg register (XMM4+ on win64), which would remove the spill path
independently of coloring.

⚠️ **Do not fix it speculatively to "harden" things.** It is currently unreachable, and §10 of
CLAUDE.md plus the 2026-07-29e revert both say an unmeasurable benefit does not buy complexity in
the most self-host-critical component. Fix it when something needs the shape.

## Guard that keeps 1f out of it

`collapse_copy_chain` refuses any fold whose def reads `OPND_PHYS`. Rationale is in the source
comment: that is ABI establishment, not a value copy, and `counts[vT] == 2` proves nothing about a
physical register. Related: [[m6-perf-baseline]].
