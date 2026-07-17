---
name: session-handoff-2026-07-09
description: "HANDOFF 2026-07-09 → next session. State: tree clean at BUG#88 fixpoint 0D672CC8; 3 local commits UNPUSHED (git credential expired). Shipped BUG#83-88 + for-collection reject. BUG#86 short-circuit blocked on RFC 0016 P2' (CFG-aware liveness)."
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

# HANDOFF phiên 2026-07-09 (đọc đầu tiên khi vào phiên mới)

## ✅ CẬP NHẬT 2026-07-12 — RFC 0017 GLOBAL STORAGE HOÀN CHỈNH + probing
`origin/main`=`ceceb7e`, 243/243, daily driver `bin/axc_native.exe` **A==B==C `81522e76`** (đổi ở `288c86a`=backend; sau đó chỉ test-only).
**RFC 0017 storage COMPLETE mọi value category** (scalar const/non-const, aggregate struct/array, pointer-repr Option/Result/sum, 16B-inline str/bytes). Commit chain:
- `288c86a` = **STR/BYTES (16B-inline) globals**: `mut g: str="…"` → 16B `.data` slot, runtime-init, đọc/ghi CẢ HAI nửa 8B inline. 3 mảnh: typecheck accept 16B PRIMITIVE; `collect_global` `is_inline16`→16B slot+runtime-init; `x86_selector` OP_LOAD 16B-non-agg dest→two-8B-halves inline load vào dest home (single 16B MACH_LOAD vào GP reg INVALID — mirror OP_GET_FIELD str). OP_STORE ko đổi (emit_block_copy LEA 16B inline src). Oracle t_globstr/t_globstridx(42).
- `3a44577` = **POINTER-REPR globals** (Option/Result/user-sum). Value=8B tagged box ptr → slot=8B store/load pointer TRỰC TIẾP (KHÔNG block-copy). 3 mảnh: (1) `typecheck check_module_global` accept sum/opt/res; Option/Result **annotation** UNRESOLVED bởi general inference → resolve register_option/result **scoped module-global-decl only** (compiler ko có global loại này → fixpoint-safe), pin symbol type_id; (2) `collect_global` ptr-repr `size=8` override + runtime-init; (3) `x86_selector` OP_LOAD force 8B + OP_STORE plain-path (src2==0) `store_is_ptr_sum` 8B ptr store (mirror BUG#78). Oracle t_globopt/globresult/globsum.
- `6264ff6` = **aggregate globals** (struct/array). Tái dùng by-address machinery: typecheck accept STRUCT/ARRAY/TUPLE; `collect_global` slot=full size+runtime-init block-copy; `lower_global_read` aggregate trả OP_GLOBAL_ADDR (địa chỉ, KHÔNG load) → GET_FIELD/INDEX compose như local. `lower_global_write` ko đổi.
- `f5ef298` = parser no-init var-decl fix (`check_raw` thay `check(TK_EQ)`→peek nuốt newline); no-init global → zeroed `.data`.
- `5d49d9d`/`1a66360` = bank 5/6 aggregate-global write+interaction oracle (O0==O1).
- `ceceb7e` = **bank 5 DEEP-CROSS oracle (test-only, binary ko đổi)**: 19 probe/4 batch (globals × ctrl-flow/generics/methods/str-16B/ptr-repr/init-order) = 0 miscompile. Bank: t_globarrstr([str;3]), t_globoptstr(Option[str] payload=str), t_globarrstructstr([{i64,str};2]), t_globarropt([Option;3] None-store), t_globinitorder(cross-global init top-to-bottom).

**2 finding KHÔNG-bug (defer):** (a) **tuple-literal expr `(a,b)` KHÔNG parse** — parser chỉ có `NODE_TUPLE_PAT` (pattern), KHÔNG tuple-EXPR node → tuple globals unconstructible (typecheck-accept-list liệt kê TUPLE nhưng KHÔNG có oracle nào). Grammar gap → cần RFC §13. (b) aggregate `return g` = REFERENCE alias (RFC 0001 §5), KHÔNG copy — đúng design.
**BẪY probe:** build từ repo ROOT WITHOUT `-self-link` (import resolve theo CWD; `-self-link`=compiler-self-build → segfault giả); self-build có thể OOM → `rm -f bin/axc_fp*.exe` trước mỗi hop. RFC 0015 P2 escape OPEN (blocked, reverted `cff2552`).

## ✅ CẬP NHẬT 2026-07-09 (phiên sau) — BUG#86 ĐÓNG
- **Tất cả ĐÃ PUSH.** `origin/main` = `755d7b8`.
- **RFC 0016 P2' (CFG-aware liveness) SHIP** = `e3f9539`. Chi tiết [[rfc0016-p2prime-cfg-liveness]].
- **RFC 0016 P3 (short-circuit and/or) SHIP** = `755d7b8` → **đóng BUG#86** [[bug86-short-circuit-open]]. 2 bug: (1) lowering diamond, (2) `lower_while` CFG-edge (bug làm B hang: -O0 pass/-O1 hang → DCE `remove_unreachable_blocks` NOP `ret` exit).
- **Daily driver `bin/axc_native.exe` hash `c777ef7b`** (thay `BCEFC38F`←`0D672CC8`).

## Trạng thái tree (SẠCH)
- HEAD local = `755d7b8` = `origin/main`. Tree clean (chỉ `scratch/self_linked_concatenated.ax` + `.claude/` pre-existing, KHÔNG đụng).
- **`bin/axc_native.exe` fixpoint `c777ef7b`.** Regression baseline **114 tests + 5 sc oracle** (sc1/sc2/scv/scw/scstress). t_math=127.
- Gate nhanh: `scripts/fast_fixpoint.ps1` (A==B, hash = `c777ef7b`). Regression: `AXC=bin/axc_native.exe bash scripts/regression_repros.sh`; spot-check nhanh `bin/axc_native.exe build bin/<t>.ax -o /tmp/x -O1; /tmp/x; echo $?`.

## Đã SHIP (đã gate)
BUG#83/84/85/87/88 · **RFC 0016 P2' (CFG liveness) + P3 (BUG#86 short-circuit)**. Tất cả pushed.

## OPEN — ưu tiên cao
- **BUG#82 globals** ([[bug82-global-var-semantics-open]]): module-level `let`/`mut` init non-zero không chạy + cross-fn RMW sai. Cần RFC riêng.
- `for x in <collection>` iteration thật (iterator protocol) — hiện REJECT sạch; feature tương lai.

## Tài liệu chính
- `rfcs/0016-cfg-aware-liveness-block-ordering.md` — thiết kế fix nền cho BUG#86 (P1 terminator-norm, P2 RPO=insufficient, **P2' CFG-aware liveness=fix thật**, P3 short-circuit). ĐỌC KỸ trước khi làm BUG#86.
- Design: [[axiom-struct-reference-semantics]] — struct = REFERENCE semantics (RFC 0001 §5), `mut cpy:=src` alias KHÔNG copy; ĐỪNG "fix" (suýt phá self-host).

## Bài học phiên này
- Proactive probing (batch feature-combo, so exit code) tìm được BUG#83-88; yield về 0 sau ~18 batch (compiler rất chắc).
- Exit code bash 8-bit: giá trị >255 bị truncate (vd 921→153); dùng PowerShell `$LASTEXITCODE` cho full 32-bit.
- CHECK SPEC trước khi "fix" hành vi lạ (struct alias là design, không phải bug).
- Backend đổi self-codegen → A!=B là transition ĐÚNG, gate = B==C tay (build C từ B). Frontend-thuần → A==B.
