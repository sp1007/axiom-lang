---
name: session-state-2026-07-24b
description: "STATE 2026-07-24b — feature backlog fully closed (for_each/reduce shipped; iface-dispatch + hex-upper were stale-notes). 19-probe clean sweep. Safe tick-sized backlog EXHAUSTED; next step needs user direction on heavy RFC/milestone items."
metadata:
  node_type: memory
  type: project
---

**ĐỌC ĐẦU TIÊN.** HEAD `000971c`, regression **528/528** GREEN, driver `bin/axc_native.exe` unchanged (this session touched NO compiler source — only stdlib + oracles + docs, so A==B by construction / no fixpoint needed).

## What shipped this session (user: "thực hiện dứt điểm 3 feature backlog … tiếp tục")
Of the three "Feature backlog" items, **only ONE needed code** — the other two were STALE backlog notes:
1. **`Vec.for_each` / `Vec.reduce` — NEW CODE** (`000971c`, std/collections.ax after `fold`). `for_each[T](self: Vec[T], f: fn(T))` (void fn-type param verified to parse+lower); `reduce[T](self, f: fn(T,T)->T) -> Option[T]` (FIRST element as init, None on empty, distinct from fold's explicit accumulator). Lambdas stay zero-capture (BUG#73) so for_each effects flow via a named fn / global — the methods NEVER needed capture; the old "blocked on capturing lambda" note conflated USE with signature. Oracle `t_vecreduce(88)`.
2. **`to_hex_upper` — ALREADY SHIPPED** (std/string.ax:471, oracle `t_tohexup(77)` wired at regression:190). "STILL OPEN" note predated `485d02f`. No code.
3. **Interface vtable dynamic dispatch (BUG#71) — ALREADY SHIPPED** in RFC 0029: `ensure_iface_box_type`/`build_interface_value` (air_builder.ax:1678/1699) build an inline per-value box `{data, m0..m(N-1)}` at each T→I coercion; `iface.m()` → `data=field0`, `fptr=field(slot+1)`, indirect `OP_CALL` (air_builder.ax:1837–1887, "Closes BUG#71"). `t_ifaceconsumer(46)` already proved the std/log `Box[LogSink]` pattern. Banked `t_ifacevecpoly(70)` to lock polymorphic dispatch through Vec[Interface] + mut-reassign + 2-method vtable slot ordering. No code. [[backlog-open-items]] updated (all three marked done/stale-closed).

## 19-probe clean sweep (correctness-first, backlog thin)
Crossed the new features + historically-buggy surfaces, ALL correct (O0/O1, exit-oracle-checked), 0 silent miscompiles:
- for_each/reduce × {struct-accumulator reduce, tuple-element reduce, Option[Interface] unwrap+dispatch, map(iface)→for_each chain, single-elem, empty-Vec, struct-lambda-with-if-expression}
- nested generic `Box[Option[str]]` match; `HashMap[str,i64]` get/match; `Result[i64,str]` Ok match
- **bug92b guard**: generic multi-field variant `Pair[str] = P(str,i64)` + recursive `Tree[str] = Node(str,…)` through a fn param — both correct (the `7f3ec7a` fix holds broadly)
- cast-chain width (`300 as u8 as i64`=44, `-1 as u8 as i64`=255), shift/bitwise (`5<<3|2`=42), string slice/concat/contains/index_of/split/replace
Not banked as oracles — these surfaces already have regression coverage (t_castwidth, t_strsplit, t_ifaceconsumer, t_gentreestr, etc.); banking would duplicate.

## Next direction — NEEDS USER (safe tick-sized backlog exhausted)
No OPEN bugs. Probing across 5 batches yields clean. Remaining backlog is all heavy/risky/blocked:
- **RFC 0009 P3** — ELF export (`.edata`): backend/linker → mandatory B==C fixpoint (~2h) AND ELF verification needs a linux/WSL runner for `elf_linux_check.sh`. Confirm the runner exists before starting, else pick a Windows-verifiable direction.
- **RFC 0015 P3** — CTGC compile-time free: HIGH-RISK (needs borrow/alias tracking; wrong => UAF in self-host). Dedicated session.
- **RFC 0014 P3** — bignum auto-free: formally SCOPED OUT (needs interprocedural return-value ownership, RFC-scale) — see [[rfc0014-drop-glue-complete]].
- **M4** (100 compliance tests — partly contradicts AXIOM grammar) / **M6** (perf gate Fib(40)≤5% clang, Mach-O/ARM64) — milestone-scale.
Autopilot guardrail applied: when only risky/blocked work remains, keep the loop alive with low-risk probing/oracle-banking rather than force a destabilizing backend change that can't be properly verified here. Heartbeat monitor stays armed.
