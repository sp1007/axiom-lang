---
name: bug-const-fold-narrow-int-wrap
description: "FIXED backend bug (probe 2026-07-18): -O1/-O2 constant-folding of narrow-int (u8/u16/u32/i8/i16/i32) add/sub/mul/neg did NOT wrap to the type width — computed in full 64-bit — so e.g. `u32 4000000000 + 500000000` folded to 4.5e9 instead of wrapping to 205032704 (O0-vs-O1 divergence, silent miscompile). Fixed in ssa_opt.ax fold_func by masking the folded result to the type width, GUARDED by operand-fit (AIR type_id is unreliable — a negated i64::MIN carries an i32 type_id)."
metadata:
  node_type: memory
  type: project
---

# Const-fold narrow-int wrap — silent O0-vs-O1 miscompile (FIXED `633913E9`, A==B)

Found by the arithmetic/bitwise bug-probe batch 2026-07-18 (`a4`: `u32 4000000000 + 500000000`,
`/1000000` → O0=205 (correct wrap), **O1/O2=148** (no wrap)). Confirmed general across u16/u32/i32
and add & mul (all narrow < 64-bit ints), both `let c = a + b` const operands.

## Root cause
`fold_func` (ssa_opt.ax) folds a binary/unary ALU op whose operands are both known constants via
`eval_binary`/`eval_unary`, which compute in **full 64-bit**. The folded `OP_ICONST` was stored WITHOUT
truncating to the operation's type width. At -O0 the same op runs in a 32/16/8-bit register and wraps;
at -O1+ the folded constant keeps all 64 bits → the two paths diverge. `4000000000+500000000 =
4500000000` stayed 4.5e9 (should wrap to `4500000000 - 2^32 = 205032704`). Runtime (opaque-operand)
path was always correct — only the compile-time folder was wrong.

## Fix (`633913E9`, ssa_opt.ax)
New helper `fold_narrow_mask(type_id, val)`: masks `val` to the width of the primitive `type_id`
(i8=1,i16=2,i32=3,u8=5,u16=6,u32=7,bool=11,char8=13 → 1/2/4-byte; signed types sign-extend to 64
bits, unsigned zero-extend; i64/u64/isize/usize/unknown/float → unchanged). Applied to the folded
result in BOTH the binary and unary fold sites.

## ⚠️ CRITICAL GUARD (why the naive version broke self-host)
First version masked unconditionally on `inst.type_id`. It **passed A==B fixpoint** but BROKE
`t_tostr` (regression 376/1-fail): `to_str(i64::MIN)` negates `-9223372036854775808`, and that
`OP_NEG` carries **`type_id=3` (i32)** — a MISLABEL (the value is i64). Masking truncated i64::MIN → 0.
**Lesson: AIR `inst.type_id` is NOT reliably the true value width.** Fix = only wrap when the
OPERAND(s) already fit the declared width (`fold_narrow_mask(tid, operand) == operand`). Then the
type_id is consistent with the values and the wrap reflects a genuine narrow-type overflow; a
mislabeled full-width value (operand doesn't fit) is left untouched. This is the same "trust the value,
not the imprecise type_id" lesson as m5/m2 ([[bug-malformed-input-robustness-cluster]]).

## Gate (backend change)
A==B `633913E9` (compiler's own constants don't overflow their types → deterministic self-build);
regression **377/377 on BOTH -O1-built AND -O2-built compiler** (mandatory optimizer-change
acceptance); t_tostr back to 88; all narrow-int fold cases (u16/u32/i32, add/mul) now O0==O1==O2.
Oracle **t_foldu32wrap** (exit 205) in main rows + opt_rows (O2/O3). LESSON reinforced: A==B alone is
NOT correctness — the full regression caught the to_str break that the fixpoint missed.
