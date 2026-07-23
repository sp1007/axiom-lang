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
- **RFC 0009 P3** — ELF export: **SCOPED + VERIFIABLE 2026-07-24b (investigation, read-only).** ⭐ **Verification tooling IS present on this machine** (corrects my earlier "needs linux, confirm runner" worry): `/c/WINDOWS/system32/wsl` AND `/c/msys64/ucrt64/bin/readelf` both exist → can `readelf` for structural checks AND `wsl` to `dlopen` the `.so` at runtime. **Current behavior:** `axc build x.ax --shared --target linux -o x.so` already emits a valid 82KB ELF but with `e_type=0x0002` = **ET_EXEC** (wrong — a shared object needs **ET_DYN=3**), and its `.dynsym` holds only IMPORTS (SHN_UNDEF), so `#[export]` funcs are not exposed. **The ELF dynamic-link infra already exists** (linker.ax ~2684–2779: `.dynsym` STB_GLOBAL/STT_FUNC template at 0x12, `.hash` single-bucket nbucket=1 + linear chain, `.dynstr`, `.dynamic` with DT_* tags, `.rela.plt`, `.got.plt`) — built for imports; **export = EXTEND these, not build new.** Concrete edit list: (1) thread `is_shared` into `linker_build_elf_headers` (linker.ax:905) → set `e_type` ET_DYN(3) when shared; (2) append exported funcs to `.dynsym` as STB_GLOBAL STT_FUNC with `st_value`=func VA, `st_shndx`=text section idx (NOT SHN_UNDEF); (3) add export names to `.dynstr`; (4) extend `.hash` nchain + chain to cover the export symbols (single bucket already hashes-all-to-0 + linear-chains, so lookup stays correct — just lengthen the chain); (5) ensure `.dynamic` carries DT_HASH/DT_SYMTAB/DT_STRTAB/DT_STRSZ (present for imports — verify they point at the export-inclusive tables). Still a real backend/linker change ⇒ **mandatory B==C fixpoint (~2h) before commit**, and touches every ELF output (risk). Ready to implement; needs the ~2h autonomous-build commitment greenlit (or a dedicated session).
- **RFC 0015 P3** — CTGC compile-time free: HIGH-RISK (needs borrow/alias tracking; wrong => UAF in self-host). Dedicated session.
- **RFC 0014 P3** — bignum auto-free: formally SCOPED OUT (needs interprocedural return-value ownership, RFC-scale) — see [[rfc0014-drop-glue-complete]].
- **M4** (100 compliance tests — partly contradicts AXIOM grammar) / **M6** (perf gate Fib(40)≤5% clang, Mach-O/ARM64) — milestone-scale.
Autopilot guardrail applied: when only risky/blocked work remains, keep the loop alive with low-risk probing/oracle-banking rather than force a destabilizing backend change that can't be properly verified here. Heartbeat monitor stays armed.
