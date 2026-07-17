---
name: bug86-short-circuit-open
description: "BUG#86 FIXED 2026-07-09 (RFC 0016 P2'+P3): short-circuit and/or shipped. Cần P2' (CFG-aware liveness) LÀM NỀN + fix lower_while CFG-edge (bug thứ 2). B==C fixpoint c777ef7b, 114/114, t_math=127."
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

## ✅ FIXED 2026-07-09 (RFC 0016 P2' + P3) — hash `c777ef7b`
Đóng hoàn toàn sau khi P2' (CFG-aware liveness, commit `e3f9539`) làm NỀN. P3:
1. **Lowering** `lower_short_circuit` (air_builder.ax) + interception theo TOKEN trong `lower_binary_expr` (`and`/`&&`/`or`/`||`, KHÔNG `&`/`|`). Diamond: `sc_res`=lhs → OP_BRANCH → rhs_block ghi đè sc_res → merge. `sc_res` 2-def an toàn nhờ optimizer guard `def_counts<=1` (né copy-prop/fold) + P2' liveness.
2. **⭐ BUG THỨ 2 (thủ phạm B hang) — `lower_while` CFG-edge SAI:** short-circuit trong điều kiện `while` tạo diamond nên OP_BRANCH nằm ở MERGE block, không phải `cond_block`. `lower_while` add_edge body/exit từ `cond_block` (cũ) → merge block THIẾU succ→exit → `remove_unreachable_blocks` (DCE, -O1) NOP `ret` của exit → **vòng lặp vô hạn CHỈ ở -O1** (-O0 không chạy SSA-opt nên OK → đây là cách khoanh vùng: -O0 pass, -O1 hang = lỗi opt/CFG-metadata). Fix: `branch_block = current_block()` SAU khi lower cond (giống `lower_if_chain` ĐÃ làm → chính bất đối xứng này khiến `if…and` chạy mà `while…and` hang). `lower_for`/`**`-loop điều kiện generated → không đụng.
3. **Gate:** oracle sc1/sc2=42 (RHS fault bị skip), scv=3 (truth table), scw=5 (while+and), scstress=32 (nested/while/precedence/as-value); B compile trivial; **B==C `c777ef7b`**; regression 114/114; t_math=127. Oracles ở scratchpad sc1/sc2/scv/scw/scstress.

Bài học: bug backend mà -O0 pass/-O1 fail ⇒ soi SSA-opt (`ssa_opt.ax`), không phải regalloc. `dump-air -O0` vs `-O1` so trực tiếp AIR tìm pass phá (ở đây exit block mất `ret`). CFG-edge metadata (add_edge) phải theo current block THỰC, không phải block giả định trước khi lower sub-expression.

---
## (lịch sử) BUG#86 — OPEN (fix attempt reverted 2026-07-09). Phát hiện qua proactive probing (batch #7).

## Triệu chứng (spec violation, CONFIRMED)
`and`/`&&`/`or`/`||` KHÔNG short-circuit — evaluate CẢ HAI operand. `if false and boom(0)` → boom(0) VẪN chạy → div-by-zero crash (exit `-1073741676`=0xC0000094 STATUS_INTEGER_DIVIDE_BY_ZERO). Tương tự `true or boom(0)`. Idiom `if p != null and p[0] > 0` sẽ fault RHS dù LHS false. VALUE đúng cho operand thuần (a&b==a and b với bool 0/1), chỉ sai khi RHS có side-effect/fault.

## Root cause
[air_builder.ax] `map_binary_op`: `and`/`&&`/`&` → **OP_AND** (eager bitwise), `or`/`||`/`|` → **OP_OR** (eager). `lower_binary_expr` lower cả hai operand rồi emit 1 op. KHÔNG branch.

## Spec ĐÒI short-circuit (authoritative)
`docs/tasks/p09-t06-air-builder-expressions.md`: "`a and b` — evaluate b only if a is true"; "Handle short-circuit for and/or: create branch blocks"; gợi ý helper `emitShortCircuit`. `docs/tasks/p08-t04-cgen-expressions.md:69`: `a and b → (a && b)` short-circuit preserved. → Đây là BUG thật, không phải design choice.

## Fix attempt #1 (REVERTED — vỡ self-host)
Intercept `and`/`or`/`&&`/`||` (KHÔNG `&`/`|` bitwise) trong lower_binary_expr: eval LHS, OP_COPY vào `sc_res` vreg, OP_BRANCH (AND: true→rhs_block/false→merge; OR: true→merge/false→rhs_block), rhs_block eval RHS + OP_COPY sc_res, jump merge, return sc_res. OP_BRANCH codegen (x86 L1413) emit JCC src2(then) + JMP dest(else) — **explicit cả 2, KHÔNG fall-through**, nên logic branch ĐÚNG. Fixpoint A!=B (đổi self-codegen — đúng vì compiler dùng and/or khắp nơi) NHƯNG **B hang cả `return 7` trivial** (build C timeout, fpB build trivial rc=124). → B miscompiled.

## ROOT CAUSE THẬT (xác nhận 2026-07-09 qua isolation test) ⭐
**Lowering ĐÚNG HOÀN TOÀN** — không phải lỗi backend/regalloc/block-order. Test cô lập: build `fpA` = short-circuit compiler (eager-built bởi axc_native nên fpA CHẠY), fpA compile chương trình nhỏ → **sc1/sc2/sc3=42 (short-circuit thật, `false and boom(0)` KHÔNG crash), scv=3 (bảng chân trị TT/TF/FT/FF đúng)**. Lowering hoàn hảo.
**Blocker THẬT: chính SOURCE compiler có site `and`/`or` DỰA VÀO EAGER eval** (RHS có side-effect PHẢI chạy dù LHS quyết định). `scB` = short-circuit compiler build BỞI short-circuit compiler → **hang trivial** (reproduce). Vì 1+ site `and`/`or` trong compiler có RHS side-effecting (vd `while cond and advance()`, `if x and mutate()`) — eager chạy cả 2, short-circuit bỏ RHS → mất side-effect → vòng lặp vô hạn / logic sai trong chính compiler.
**CẬP NHẬT phân tích culprit**: grep site `and/or` có RHS `self.method()` trong hot-path (parser/typecheck) → TẤT CẢ là predicate THUẦN (`check`/`peek_at`/`is_generic`/`is_method_compatible`, read-only). ⇒ giả thuyết "eager-reliance side-effect" YẾU đi. **Leading hypothesis giờ: REGALLOC mishandle `sc_res` vreg (ghi 2 block, diamond-merge) dưới ÁP LỰC THANH GHI CAO của hàm compiler phức tạp** — chương trình nhỏ OK (fpA proves), hàm lớn/nhiều biến/nested-loop thì sc_res bị spill/reload sai → giá trị `and/or` sai → hang. (Simple function regalloc home sc_res đúng; complex thì không.)
**❌ Stack-slot plan BÁC BỎ**: `OP_ALLOC` KHÔNG phải stack slot — x86 (L1648) emit **MACH_CALL imm=-1 = gọi runtime allocator (HEAP)**. Dùng cho mỗi `and/or` = malloc mỗi lần + leak trong loop. Vô dụng. "Stack slot" tôi muốn CHÍNH LÀ 1 spilled vreg (backend tự home vreg xuống stack) — tức đúng là sc_res. Nên vấn đề = **regalloc home multi-def vreg sc_res**.
**Next attempt (regalloc)**: sc_res dùng ĐÚNG cơ chế mut-local (OP_COPY vào vreg ghi 2 block — L2881-2889 mut-local diamond CŨNG vậy và CHẠY ở compiler scale). Nên NGHI: (a) mut-local vreg có được ĐĂNG KÝ/track đặc biệt để regalloc spill nhất quán mà fresh_reg KHÔNG có? → soi cách local_map vreg vs fresh_reg khác nhau ở regalloc (interval computation: vreg 2-def có tính live-interval UNION đúng hay fragment?). (b) HOẶC high-register-pressure trong hàm compiler phức tạp làm sc_res spill/reload sai (small program áp lực thấp → OK; compiler → vỡ). Cần đọc x86_regalloc interval/liveness cho multi-def vreg. Gate cuối: scB compile trivial OK + B==C + regression + sc1/sc2/sc3=42 + scv=3. **Cần phiên riêng tập trung regalloc.**

## ⭐⭐ ROOT CAUSE CHÍNH XÁC NHẤT (đọc x86_regalloc.compute_liveness 2026-07-09)
`compute_liveness` (x86_regalloc.ax:55) là **LINEAR-scan theo CHỈ SỐ INSTRUCTION**, KHÔNG control-flow-aware. Interval 1 vreg = `[first_def_index, last_use_index]` (def thứ 2 của multi-def bị BỎ QUA — L78 chỉ tạo khi null; use mở rộng `end`). Back-edge extension (L164-182) chỉ nới end tới vị trí JMP/JCC ngược. **Giả định CỐT LÕI: thứ tự mảng instruction ≈ thứ tự execution.** Short-circuit tạo `sc_rhs_block`/`sc_merge_block` GIỮA biểu thức; đặc biệt trong điều kiện `if`, `lower_if_chain` ĐÃ tạo then/else TRƯỚC → stream instruction phát ra có **control-flow order ≠ linear-index order** → linear liveness tính interval SAI → gán thanh ghi chồng chéo/clobber → giá trị sai → hang. Simple function (fpA): ít block, order tình cờ đúng → OK. Complex compiler function (nhiều nested block, áp lực cao): order xáo trộn → VỠ. **Đây là lý do "lowering đúng ở isolation nhưng vỡ ở self-compile".**
**FIX THẬT (2 hướng, đều là backend lớn)**: (A) phát block theo **RPO (reverse postorder)** để linear-index khớp control-flow trước khi compute_liveness; HOẶC (B) làm **liveness CFG-aware** (theo block + edge, không theo linear index). Cả hai đụng regalloc/emission core → RFC + phiên riêng. Short-circuit chỉ ship ĐƯỢC sau khi sửa liveness/ordering này (nó là bug NỀN, ảnh hưởng mọi lowering tạo block phi-tuyến — không chỉ short-circuit). Có thể đây cũng là lý do các lowering khác né tạo block giữa biểu thức.

## ⭐⭐⭐ THÍ NGHIỆM RPO (P2) — 2026-07-09: SELF-HOST ĐƯỢC nhưng REGRESS t_math → cần CFG-aware liveness
Đã implement RPO block serialization (`compute_rpo`+`dfs_postorder` trong x86_selector, emit block theo reverse-postorder từ block 0, safety-net append block unreachable). Kết quả:
- **RPO compiler SELF-HOST**: `scB` compile trivial OK, **B==C fixpoint** (`40E531D6…`), ~30 spot-check tests PASS → **chứng minh short-circuit compiler ĐẠT ĐƯỢC khi liveness đúng**.
- **NHƯNG t_math REGRESS 127→124**: giá trị float loop-carried (Newton sqrt / chuỗi ln/exp/sin) bị miscompile. RPO tạo THỨ TỰ LOOP ĐÚNG (postorder+reverse tự xếp cond,body,exit), nên lỗi KHÔNG phải block-order → **chính `compute_liveness` SAI cho loop-carried value** (linear interval + back-edge extension fragile).
- **KẾT LUẬN: reorder block KHÔNG đủ — chỉ dời CFG-shape nào vỡ. Fix THẬT = CFG-aware liveness (live-in/live-out dataflow), RFC 0016 P2'.** Đã REVERT sạch, tree về BUG#88 fixpoint `0D672CC8`, t_math=127 lại.
- Full plan + kết quả ở `rfcs/0016-cfg-aware-liveness-block-ordering.md`. **Next: implement P2' (backward dataflow liveness) với t_math làm canary bắt buộc, rồi P3 (re-apply short-circuit).**

## Xác nhận cuối (block serialization)
Block serialize theo **CREATION order** (x86_selector L2211-2215: lặp `f.blocks.data` theo id; new_block id=blocks.len). `lower_if` tạo **merge_block TRƯỚC** (id M) rồi then(M+1)/else(M+2) → if/else HIỆN TẠI đã serialize merge TRƯỚC then/else (NGƯỢC control-flow) mà VẪN chạy — vì merge là join RỖNG + giá trị đi qua mut-local vreg (def trước if, use SAU if, interval phủ hết). Short-circuit VỠ pattern mong manh này vì **giá trị sc_res được TIÊU THỤ NGAY TẠI join (sc_merge)** — không có use sau toàn bộ cấu trúc để interval linear phủ đúng qua các block phi-tuyến. ⇒ khẳng định fix = RPO serialization HOẶC CFG-aware liveness (bug NỀN, không riêng short-circuit).

## (cũ, ĐÃ BÁC) Nghi ngờ root cause #2 — block-order vs control-flow
`lower_if_chain` tạo `then_block`/`else_block` bằng new_block() **TRƯỚC** khi lower_expr(cond) (giả định cond eval trong block hiện tại). Short-circuit chèn `sc_rhs_block`/`sc_merge_block` với ID **CAO HƠN** then/else nhưng nằm control-flow **GIỮA** cond và then/else → thứ tự tạo block (ID) NGƯỢC thứ tự control-flow. Nếu backend giả định block-id-order == RPO/emission-order cho liveness/regalloc → mọi `if X and Y:` (guard bounds khắp compiler) sinh CFG lệch → regalloc sai → B hang toàn cục. **Fix có thể cần**: tạo sc blocks TRƯỚC khi caller tạo then/else, HOẶC đảm bảo order khớp. Đây là lý do vỡ RỘNG (không chỉ 1 site).

## Nghi ngờ root cause #1 (chưa xác nhận)
`sc_res` = fresh_reg ghi ở HAI block (sc_cur + rhs_block), đọc ở merge (diamond). Mut-local qua diamond cũng dùng OP_COPY vào **fixed vreg** (comment air_builder L964-966) — NHƯNG local's fixed vreg có thể được TRACK để regalloc cho stack home ổn định; `sc_res` ad-hoc fresh_reg ghi 2 block có thể KHÔNG được regalloc (linear-scan) home nhất quán qua diamond → đọc rác → sai điều kiện `while ... and ...` trong compiler → vòng lặp vô hạn → B hang.

## Next khi làm (cẩn thận, incremental)
1. Test cross-block vreg-merge ISO trước: chương trình nhỏ `let r = (a and pure_b); return r` build bằng compiler-có-fix, chạy — nếu sai value → xác nhận regalloc không home diamond vreg.
2. Nếu đúng vậy: dùng cơ chế local mut thật (alloc stack slot + store/load) HOẶC đăng ký sc_res như "cross-block fixed vreg" mà regalloc hiểu (tìm cách compiler đánh dấu local fixed vreg).
3. Test `and`/`or` trong WHILE-condition + nested `a and b and c` + as-expression trước khi self-host.
4. Gate: A!=B (đổi self-codegen) → **B==C BẮT BUỘC** + B build trivial + regression + oracle sc1/sc2/sc3 (exit 42, không crash). Backend → fixpoint trước commit.
Oracles đã có: scratchpad sc1.ax/sc2.ax (div-by-zero nếu RHS chạy), sc3.ax (both true→42). Cùng họ [[bug82-global-var-semantics-open]] (cũng cần cẩn thận vì hot path compiler).
