---
name: bug-undefined-name-accept
description: "FIXED (probe 2026-07-18): a bare value-position reference to an UNDECLARED name (`return foo` / `let y = bar + 5` where foo/bar are never defined) was SILENTLY ACCEPTED — the resolver left the raw name_id as the ident payload, which accidentally indexed an unrelated symbol → garbage/0 at runtime. Now rejected with `error: undefined name 'X'` before codegen. Fix: resolver forces payload=0 on unresolved; typecheck rejects a NODE_IDENT that ends UNKNOWN with payload==0."
metadata:
  node_type: memory
  type: project
---

# Undefined-name silently accepted — FIXED `D79B62DE` (A==B, frontend)

Found by the malformed-input robustness probe (2026-07-18, same class as m1/m2/m4/m5). `return
totally_undeclared_xyz` built an exe and returned 0; `let y = undeclared_abc + 5; return y` returned
garbage. A fundamental hole: an undefined variable/name must be a hard name-resolution error.

## Root cause
[resolver.ax ~L1006] NODE_IDENT resolution: `sym_idx = symtable.resolve(name_id)`. On success it
rewrites `payload = sym_idx`; on FAILURE (unresolved) it left `payload = name_id` unchanged and emitted
NOTHING. In typecheck ([typecheck.ax ~L4086] NODE_IDENT), `sym_idx = payload = name_id`, and
`is_sym = (sym_idx != 0 and sym_idx < symbols.len)` — a small name_id accidentally indexes a real but
unrelated symbol slot → `result_type` = that stray symbol's type → no error, runtime garbage.

## Fix (2 sites, frontend → A==B)
1. **resolver.ax** unresolved path: `self.tree.nodes.data[node_idx].payload = 0` — a definitive
   "no symbol" marker (instead of leaving the raw name_id).
2. **typecheck.ax** end of NODE_IDENT branch: if `result_type == UNKNOWN and not is_substituted and
   payload == 0` → `error: undefined name 'X'` + `diags_count++` (driver halts before codegen).
   Builtin TYPE names used in value position already set `result_type` via node_text matching above,
   so they never reach this reject.

## Why bounded-safe (the m2/m5 over-rejection trap avoided)
Before writing the reject, TRACED every unresolved bare ident during the compiler's OWN self-build
(instrumented the resolver's else-branch, ran the traced compiler on tmp_concatenated_air.ax):
**ZERO unresolved idents** — every ident in the entire compiler source binds in the resolve pass
(top-levels are pre-registered by `pre_define_top_levels`, locals/params defined as encountered). So
forcing payload=0 on unresolved changes nothing for valid code. Confirmed: A==B fixpoint `D79B62DE`
held, full regression 384/384 (no valid program newly rejected), and a valid program with
forward-referenced fn + locals + params still builds. Oracle **t_undefname** (reject mode).

## Lesson
The resolver silently tolerating an unresolved bare ident is the SAME class as m2 (call-non-fn) and
m5 (str↔numeric let) — accept-then-miscompile. Unlike m2 (unsound to reject — infer imprecision),
this one IS cleanly rejectable because the RESOLVER (scopes live) is the reliable oracle for
"undefined", not infer_node. Key discriminator: emit the reject from the resolver's knowledge
(payload=0 marker), consumed by typecheck's halt gate.
