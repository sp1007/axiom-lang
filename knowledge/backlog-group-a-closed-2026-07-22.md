---
name: backlog-group-a-closed-2026-07-22
description: "Group A (residual hẹp) CLOSED 2026-07-22 — A1..A6 all shipped or reclassified; includes the probe-found match-on-aggregate miscompile and the .bss reclassification."
metadata:
  node_type: memory
  type: project
---

User directive 2026-07-22: "thực hiện #A, dứt điểm". All six items are now closed.
Baseline moved **474/474 → 481/481**, driver `4043C7B942EBB203`.

## Shipped

| # | Item | Commit | Gate |
|---|------|--------|------|
| A3 | `to_hex_upper(i64)` (+ numeric UFCS) | `485d02f` | A==B `36122AC2` |
| A6 | RFC 0020 §4: `is_ascii`/`to_ascii_lower`/`to_ascii_upper`, `char_indices`, `to_utf32`/`from_utf32`, `to_utf16`/`from_utf16`, `utf8_encode_len` | `485d02f` | same |
| A5 | uninit globals + ELF `.data` — **already worked**, pinned by oracle | `a1ce9f9` | tests only |
| — | **`match` on aggregate/str/float scrutinee → REJECT** (probe-found miscompile) | `19c1ab1` | A==B `033D629A` |
| A4 | RFC 0022 P5: nested destructuring patterns + `for (a,b) in v` | `e4bea71` | A==B `E1DD4102` |
| A1 | Option/Result arg payload-width reject | `171ea83` | A==B `7D90A4B1` |
| A2 | module-level INTERFACE globals | `a40f3e2` | **A==B==C** `4043C7B9` |

Oracles added: t_tohexup(77), t_strenc(70), t_globuninit(92), elfglobuninit(92),
t_matchaggreject(reject), t_tuple5(91), t_optpayloadreject(reject), t_ifaceglobal(86).

## Findings worth keeping

**`match` on a non-scalar scrutinee was a silent miscompile** (found while scoping A4,
fixed `19c1ab1`). A struct/array/tuple/str/bytes/float scrutinee hits NEITHER lowering
path (sum-tag dispatch, or the int/bool/char switch) → the match lowered to NOTHING and
the function returned garbage RAX. `match p: _: return p.x` returned 0, not 42. The
per-arm literal-pattern reject `07139af` did NOT cover it — that one only fires on a
LITERAL pattern, so a bare `_` or binding arm slipped through. Reject now sits at the
SCRUTINEE level (once per match). GENERIC_INST and TYPE_UNKNOWN are deliberately spared.

**A1's real shape — three gate-caught corrections.** The reject took four cycles because:
(1) the param arrives as kind **SUM** but the unannotated local as kind **GENERIC_INST**,
so both shapes must be measured or the reject silently never fires; (2) those two branches
measure DIFFERENT quantities — variant-payload size vs type-argument size — which coincide
ONLY for Option/Result. For `Tree[T] = Node(T,Tree[T],Tree[T])` they diverge (24 vs 8) and
full regression caught the bogus over-reject of t_gentree. Hence the base-name scope gate
to Option/Result. (3) Reject only ACROSS the 8-byte boundary; two payloads that both fit in
8 bytes (i32 vs i64) are widened correctly and rejecting them would break working programs.
Confirms the standing rule: **compare payload SIZE, never type_id.**

**A2's subtle half:** opening the RFC 0017 gate was not enough — `mut g: Shape = Square(..)`
SIGSEGV'd because the global-init prologue is a VALUE SITE and was storing the raw struct
pointer, not the vtable box. Wiring `coerce_struct_to_interface` there makes it the **8th
and final** T→I coercion site, completing the family closed in `770dfc3`.

## Reclassified OUT of the residual tier

**`.bss` is NOT a residual — it is RFC-scale linker work.** Cost is real and measured
(a `[i64; 200000]` zero global inflates an exe 77KB → 3.2MB), but the change is a FOURTH
object section across COFF + ELF + the self-linker's image layout (SizeOfRawData=0/
VirtualSize=N; SHT_NOBITS with memsz>filesz) plus relocation routing to it. Per CLAUDE.md
§13 that needs an RFC and a B==C gate. It belongs with the milestone work, not a cleanup
pass. Do not slip it in unannounced.

## Still deferred (unchanged)

- RFC 0022 P5 remainder: destructuring in fn PARAMS; tuple patterns in `match` arms
  (the latter is now a clean diagnostic rather than a silent miscompile).
- RFC 0020 §4 `ascii`/`utf16`/`utf32` **TYPES** — the transcoding FUNCTIONS shipped; the
  types still need an RFC amendment. Code units/scalars are carried as `i64` in `Vec[i64]`,
  which holds every UTF-16 unit and Unicode scalar exactly and needs no new type.

## Lesson (cost ~2 cycles)

A python trace-strip left an **orphaned `if` with no body** → parse error, and separately a
gate ran against a **stale concat** (the OOM was the traced compiler, not the change).
Always confirm the edit reached `tmp_concatenated_air.ax` (`grep` a distinctive comment)
before trusting a RED gate. Related: [[fast-fixpoint-workflow]],
[[bug-unannotated-some-aggregate-match]], [[rfc0029-vtable-progress]], [[backlog-open-items]].
