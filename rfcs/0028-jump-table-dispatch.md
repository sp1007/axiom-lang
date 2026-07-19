# RFC 0028 — Jump-table dispatch for dense integer/tag chains

- Status: DRAFT (2026-07-19c) — investigation complete; implementation is a dedicated
  backend+linker session. User greenlit (2026-07-19c backlog).
- Depends on: air.ax (new opcode), x86_selector/emitter (codegen), linker.ax (new
  relocation kind for a code-address table), ssa_opt.ax (recognition pass).
- Related: [[m6-perf-gate-fib-benchmark]] (perf work), the self-host dispatch hot paths.

## 1. Motivation

The compiler's hottest code is opcode dispatch written as long `if op == OP_X: … elif op ==
OP_Y: …` chains (x86_selector.ax `lower_instruction`, air_builder, ssa_opt) and `match` on
sum tags. Today these lower to a **linear chain of `OP_BRANCH` compares** (air_builder.ax:
match → chained tag-compares; if/elif → sequential `cmp; je`). For an N-way dispatch that is
O(N) compares per dispatch — the selector's main switch has ~80 arms, so a worst-case opcode
pays ~80 compares. A jump table turns this into O(1): one bounds check + one indirect jump
through a table indexed by `op - min`. This speeds BOTH the compiler itself and any user
program with a dense `match`/if-chain. Semantics-preserving.

## 2. Current state (investigation 2026-07-19c)

- AIR control flow is ONLY `OP_JUMP` (0x0301, unconditional) and `OP_BRANCH` (0x0302,
  two-way: dest=false-target, src1=cond, src2=true-target). **No indirect/computed jump, no
  switch, no jump-table opcode exists.**
- `match` on a sum lowers to sequential tag `OP_BRANCH`es; `if/elif` likewise. Both are the
  linear form this RFC optimizes.
- The custom linker (linker.ax) resolves code relocations for calls/branches but has **no
  relocation kind for a table of code addresses in `.rodata`** — this is the main new
  infrastructure piece.

## 3. Design

Three parts, each independently gated:

### 3a. New AIR opcode `OP_JUMP_TABLE`
`OP_JUMP_TABLE`: src1 = index vreg (already normalized to `0..N-1`), dest/extra = a table of
M block-id targets + a default block id. Encoded via the `emit_extra` side array (like OP_CALL
args): extra[0]=default_block, extra[1]=min_value, extra[2]=count, extra[3..3+count]=block ids.
Verifier: index in range, all targets valid blocks.

### 3b. Codegen (x86_selector + emitter)
Lower `OP_JUMP_TABLE` to:
```
    cmp  idx, count            ; unsigned
    jae  .Ldefault
    lea  rTmp, [rip + .Ltable] ; RIP-relative table base
    movsxd rIdx, idx
    jmp  qword [rTmp + rIdx*8]
.Ltable: (in .rodata) .quad .Lblk0, .Lblk1, …   ; each needs a reloc to the block's code addr
```
The `.quad .LblkK` entries are **code addresses** → each needs a relocation resolved at link
time. This is the new linker work (3c). Alternative: store 32-bit RIP-relative offsets
(`.long .LblkK - .Ltable`) and `jmp` via `lea base; add; jmp` — avoids absolute relocs but
needs offset relocs. Pick the RIP-relative-offset form (position-independent, matches ELF PIE
+ COFF; smaller table).

### 3c. Linker relocation for the code-address table
Add a relocation kind "code-offset-in-rodata": for each table slot, patch `target_code_va -
table_slot_va` (32-bit) once block addresses are assigned. Mirrors existing branch-offset
relocation math but the fixup site is in `.rodata`, not `.text`. Both COFF and ELF writers +
linker fixup pass.

### 3d. Recognition pass (ssa_opt.ax, opt-level ≥ O1)
A CFG pass that detects a block ending in a chain: `cmp v, cK; je BK; …` on the SAME vreg `v`
with distinct integer constants, terminating in a default. If the constants are DENSE ENOUGH
(heuristic: `count >= 4` AND `(max-min+1) <= 2*count`, i.e. ≥50% density) rewrite the chain
head into `idx = v - min; OP_JUMP_TABLE idx, [BK…], default`. Non-dense or small chains keep
the linear form (a 3-way `if` is faster linear). Must run BEFORE regalloc; preserve the
non-SSA def-count discipline (cf. the shift-in-loop / const_divisor_pow2 lesson — the compared
vreg `v` must be single-valued at the chain, or bail).

## 4. Alternatives

- **Binary-search dispatch** (balanced compare tree, O(log N)) — no linker work, but slower
  than O(1) and more code. Good fallback for SPARSE constants; complements 3d (use table for
  dense, tree for sparse). Defer the tree to a follow-up.
- **Leave as-is** — the linear chains are correct and the compiler is fast enough (~4.5s
  self-build); this is a perf nicety, not a correctness need. Justifies the DRAFT status /
  low urgency vs. the other greenlit features.

## 5. Drawbacks

- Largest new-infrastructure surface of the greenlit backlog: a new opcode + codegen + a NEW
  linker relocation kind + a CFG recognition pass. High blast radius (linker + backend +
  optimizer all touched).
- Backend/linker change → **B==C mandatory + full regression + -O2/-O3 acceptance** (the
  RFC 0025/0026 lesson: B==C necessary NOT sufficient; run the -O2-BUILT compiler regression).
- Jump tables can pessimize a small/sparse dispatch (branch predictor handles a short linear
  chain better) — hence the density heuristic in 3d.

## 6. Migration / compatibility

No source-level or ABI change; purely an internal codegen optimization. Default-off at -O0,
on at ≥O1 (like other opt passes). Fully semantics-preserving.

## 7. Gate (before commit, when implemented)

Backend+linker change → build compiler at -O2, run regression on the -O2-built compiler
(loop/opt passes are invisible to an -O1 self-build); B==C fixpoint; full regression 439+/439+;
new oracles: a dense `match`/if-chain returning the right arm for every value incl. below-min /
above-max / gap values (default path), + the compiler's own selector dispatch exercised. Verify
the `.rodata` table relocations on BOTH COFF (Windows) and ELF (Linux, now self-hostable).

## 7b. De-risking findings (2026-07-19c investigation)

- **No new LINKER relocation is actually required** (corrects §3c). A `match`/if-chain
  dispatches only to blocks WITHIN THE SAME FUNCTION, so emit the table **INLINE in `.text`**
  (a trailing data block after the function body) storing **intra-function 32-bit offsets**
  `target_label - table_label`. Both the table and its targets are local labels in the same
  `.text` region, resolved at the existing function-assembly/label-fixup phase — no link-time
  reloc, no `.rodata`. Dispatch: `lea rBase,[rip+Ltable]; movsxd rOff,[rBase+idx*4]; add
  rBase,rOff; jmp rBase`. This is significantly lower-risk than the original `.rodata`+reloc
  plan and does NOT block on RFC 0029's shared reloc.
- **Codegen infra present:** `MACH_JMP`(19) exists but takes an OPND_LABEL; a **register-
  indirect `jmp rBase`** variant must be added (x86 `FF /4`). `MACH_CALL_INDIRECT`(22) already
  exists (fn-pointer calls, BUG#49) — mirror its encoding for the indirect jmp. The table's
  32-bit label-difference entries need a new emitter fixup kind (data word = `Ltarget -
  Ltable`), analogous to the existing rel32 branch fixup but written into a `.text` data slot.
- **Self-host safety:** gate the recognition pass behind an opt-in `-jumptable` flag (like
  `-ctgc-free`). Default self-build never emits OP_JUMP_TABLE → **A==B** (the new opcode +
  codegen sit as dead code on the self-build), so the change is self-host-inert until validated
  and later enabled by default. This removes the main stability risk.

## 7c. Emitter fixup mechanism (2026-07-19c — path fully confirmed)

The inline-table entries reuse the EXISTING `Fixup` machinery (x86_emitter.ax:38
`Fixup{offset,label_id,inst_size,pad0}`; resolved in `emitter_resolve_fixups`:494). Today a
local-label fixup patches `rel = target - (offset+4)` (rel32 to the next instruction, line 528).
For a **table entry** the wanted value is `target - table_base` (int32, position-independent).
Concrete plan:
- Place a local label `Ltable` at the table start; record it in `e.labels`.
- Emit each 4-byte entry as a placeholder `0` and push a `Fixup` with `pad0 = Ltable's byte
  offset` (currently pad0 is unused/spare). In `emitter_resolve_fixups`, add a branch: if
  `pad0 != 0`, patch `target - pad0` (no `-4`) as the int32 — i.e. offset-from-table-base.
- Dispatch code (selector → MACH): `lea rBase,[rip+Ltable]` (RIP-relative to a local label —
  reuse the same rel32-to-label fixup path as MACH_JMP, but for `lea`), `movsxd rOff,
  [rBase+idx*4]`, `add rBase,rOff`, `jmp rBase` (register-indirect = `FF /4`; ADD this encoding —
  MACH_JMP currently only does rel32-to-label).
- `OP_JUMP_TABLE` operands via the `emit_extra` side array (like OP_CALL args):
  `[default_block, min_value, count, blk0..blk_{count-1}]`; src1 = the (already `- min`,
  bounds-checked-unsigned `< count`) index vreg.
So NO new struct fields and NO new linker reloc — one new emitter branch (pad0-based fixup),
one new instruction encoding (indirect jmp), and the `lea`-to-local-label. Confirmed tractable;
the risk is purely getting the encodings + fixup arithmetic right, so validate with a
`-jumptable`-built dense-match oracle at O0/O1/O2 on BOTH COFF and ELF before enabling by default.

## 8. Implementation order (dedicated session)

1. `OP_JUMP_TABLE` opcode + verifier (inert; A==B).
2. Codegen with RIP-relative-offset table + the linker reloc kind; hand-emit a test AIR
   using it (bypass the pass) to validate codegen+reloc in isolation, both targets.
3. Recognition pass at O1 behind a `-jumptable` flag first (opt-in, measure), then enable by
   default once B==C + -O2 regression is green.
4. Density-heuristic tuning + a sparse-chain binary-tree follow-up (optional).
