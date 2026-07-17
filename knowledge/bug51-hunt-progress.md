---
name: bug51-hunt-progress
description: "🎉 BUG#51+52 FIXED+SHIPPED (4d7949a) — stale aggregate-alias qua mono vec-grow; daily-driver native bin/axc_native.exe mở khóa (1200x); còn RFC 0010 landmine class"
metadata: 
  node_type: memory
  type: project
  originSessionId: ed12f2e7-f4ab-41c0-9bb3-940acfa7aaec
---

**🎉 BUG#51 + BUG#52 FIXED + SHIPPED (2026-07-03, commit 4d7949a):**

**Root cause:** `let callee_node = self.tree.nodes.data[callee]` — native backend giữ aggregate BY-ADDRESS (alias; cgen thì COPY) → mono instantiation grow nodes vec (realloc+free old) → alias dangling. gcc-built đọc trang stale còn mapped + nội dung nguyên → "chạy đúng"; native (VirtualFree block lớn) → segfault deterministic (t_mathx) + module-load flaky theo ngưỡng free-list/VirtualFree (#52). Fix targeted: fresh reads sau điểm instantiation trong NODE_CALL handler (typecheck.ax).

**Verify:** t_mathx 28/28 ×3; t_modcollide 10/10; fast-fixpoint n2==n3; FULL regression 93/93 **bằng binary native** (chuẩn gate mới).

**DAILY DRIVER MỚI: `bin/axc_native.exe`** (= n2 converged, có mọi fix). Self-build 790 hàm = **8s** (vs stage0-built 2h50m). Regression full = vài phút (vs 1.5h). Quy trình mới:
- Dev loop: sửa source → rebuild_stage1.ps1 (concat) → `bin/axc_native.exe build concat -o <new> -self-link -O1` (8s) → test.
- Gate: FULL regression `AXC=bin/axc_native.exe bash scripts/regression_repros.sh` + fast-fixpoint (self-build ×2, SHA).
- stage0-rooted chain (rebuild_stage1.ps1 stage1-gcc) vẫn giữ làm trusted root, re-verify định kỳ.
- **Sau mỗi fix mới: promote binary converged mới → bin/axc_native.exe.**

**Kỹ thuật hunt đáng tái dùng:** bisect size-threshold (non-monotonic = grow-boundary, không phải construct) → gdb crash addr → instrument-probe `ax_printf_local+fflush` loop <1phút qua converged binary → **build instrument ở -O0** (probe -O1 bị reorder → bracket sai).

**⚠️ LANDMINE CLASS CÒN LẠI → RFC 0010 (Draft, e6fdf51+bdfd95e):** mọi `let s = arr[i]` (aggregate) giữ qua vec-grow = bom. Fix = value-semantics block-copy trong air_builder lower_let (frontend-only, dùng OP_ALLOCA+OP_STORE sẵn có, KHÔNG đụng backend). **Audit sơ bộ (RFC §8): flip CÓ VẺ AN TOÀN** — style codebase = value-bind-để-đọc + direct-index-để-ghi (`arr[i].f=v`), các `&arr[i]` con-trỏ-tường-minh không bị đụng. Chưa rà nghiêm ngặt 196 site (điều kiện tiên quyết P1).

**RFC 0010 → NEEDS REDESIGN (a7f6057):** đã thử blanket copy-at-bind (OP_ALLOC+OP_STORE trong lower_var_decl cho `let/mut x = arr[i]` addr-aggregate). Oracle t_aggcopy OK (copy=5 vs alias=99), regression 94/94, NHƯNG **FIXPOINT FAIL** — compiler tự-build với copy-semantics hỏng. Root cause negative: compiler DỰA aliasing cho **read-after-mutate cùng buffer** (`let node=nodes.data[i]; nodes.data[i].f=X; đọc node.f qua alias`). Hai loại alias khác nhau: (1) qua-realloc=BUG, (2) read-after-mutate=relied-upon. Targeted 4d7949a (re-fetch) mới đúng hướng. Đã revert code, giữ bin/t_aggcopy.ax oracle. Redesign: re-fetch-lint / escape-across-call-copy / realloc-stable-buffers.

**`--time` flag SHIPPED (3cb40d4):** `axc build ... --time` in `[time] <phase>: <ms>` (concat/lex/parse/resolve/typecheck/air-build/ssa-opt/codegen/self-link), inert khi không có flag. Gate: regression 93/93 + fixpoint OK. **DỮ LIỆU PERF then chốt:** t_mathx -O1 = **codegen 2074ms (~80%)**, typecheck 214ms, ssa-opt 98ms, còn lại <40ms. ⟹ bottleneck THẬT = native codegen (regalloc).

**🎉 PERF#2 FIXED+SHIPPED (bb58f18 + docs dde8772):** compiler tự-build codegen vẫn 9560ms sau PERF#1. Probe `fn-cost` → tập trung vài mega-func THẬT: `select_inst` insts=27071/spills=12882=3045ms (select 1341 + regalloc 2076), infer_node 660, translate_inst 524. Root cause: `is_float_vreg` linear-scan `fn_ptr.insts` gọi per-interval trong `allocate_registers_orchestrator` split-loop ⇒ O(V×N). Fix: `build_vreg_def_idx` (1 lượt O(N), vreg→def-inst-đầu) + `is_float_vreg_cached` (cùng logic, lookup O(1), byte-identical). **codegen 9560→4452ms (~2.1×)**, select_inst regalloc 2076→689ms. Fixpoint c5316008 + 93/93. Lưu ý: hàm insts>5000 đi spill-all fallback (return sớm trước graph-coloring) nên chi phí ở split+prescan, KHÔNG ở interference.

**🎉 PERF#3 FIXED+SHIPPED (cc9acf9 + docs d434ce9):** `regalloc_is_16byte` cùng bệnh linear-scan (spill-all fallback 482 + insert_spill_code 886/929/956) = O(V×N). **KEY INSIGHT an toàn:** không nhân đôi 120 dòng (đệ quy/inner-scan-2/cf-skip) — dùng **wrapper mỏng giữ signature (16 caller nguyên, null→linear)** + `regalloc_is_16byte_cached(...def_idx...)` = CÙNG body, chỉ đổi **START scan tại def_idx[vreg] thay vì 0** (chứng minh identical vì [0,def_idx) toàn dest≠vreg; def_idx==-1→false). **codegen 4452→2396ms; TỔNG PERF#2+#3: 9560→2396ms (~4×).** Fixpoint 13ac5d05 + 93/93. **Pattern tái dùng: cache hàm scan-tìm-def phức tạp = jump-to-def-index, ĐỪNG replicate logic.**

**🎉 PERF#4 FIXED+SHIPPED (1411787):** `get_register_type` (x86_selector.ax) cùng bệnh linear-scan, gọi ~25 lần trong select_inst → quadratic ~747ms. Fix: cache def_idx **trên struct InstructionSelector** (đã threaded qua `sel`), build 1 lần trong select_all, get_register_type jump-start tại def_idx[reg]. codegen **2396→1442ms**. Fixpoint 966dfb09 + 93/93.

**📊 TỔNG KẾT PERF SESSION (2026-07-04) — 4 fix, đều fixpoint+93/93:** codegen native compiler-self-build **9560→1442ms (~6.6×)**; float-heavy t_mathx **2074→126ms (~16×)**. Tổng self-build wall-clock **~14.8s → ~4.1s (~3.6×)**. PERF#1 dense-vreg (+60000→next_vreg), #2 is_float_vreg memoize, #3 regalloc_is_16byte memoize, #4 get_register_type memoize. Quy trình: `--time`→probe per-func/sub-phase→phát hiện quadratic→def_idx memoize (jump-to-def-index, không replicate logic). **Nguyên tắc vàng: hàm scan-tìm-def gọi trong vòng O(V) = quadratic ẩn; luôn nghĩ def_idx.**

**🎉 PERF#5+#6 FIXED+SHIPPED (2ca4843, e448e06):** đào typecheck. Probe: phase3 infer_node=883ms; `resolve_method_sym`+`method_ret_type` gọi 4854× scan 9482 syms = 46M iter, mỗi iter alloc str (intern.get + match_method_name concat). #5: match_method_name zero-alloc suffix-check → phase3 883→669ms (~24%). #6: `intern.get`=alloc_str_from_raw COPY → thêm `InternPool.name_matches_method` so-khớp-trên-arena (borrow, no-copy) → bỏ ~14M alloc/free. Fixpoint e737a3a2 + 93/93. Nguyên tắc mới: **intern.get/get_str ALLOCATE — đừng gọi trong vòng nóng để so sánh; dùng arena-borrow.**

**📊 TỔNG KẾT PERF SESSION (2026-07-04) — 6 FIX, đều fixpoint+93/93:** codegen compiler-self-build **9560→~1150ms (~6.6×)**; float t_mathx **2074→126ms (~16×)**; typecheck phase3 **883→669ms**; tổng self-build wall-clock **~14.8s→~3.5s (~4×)**. PERF#1 dense-vreg, #2 is_float memoize, #3 is_16byte memoize, #4 get_register_type memoize, #5 match_method_name zero-alloc, #6 intern arena-borrow. **2 nguyên tắc vàng: (a) scan-tìm-def trong vòng O(V)=quadratic ẩn→def_idx; (b) alloc/concat/intern.get trong vòng nóng=alloc-storm→borrow.** Còn lại lành mạnh (diminishing): codegen ~1.1s, typecheck ~0.7-0.9s, self-link ~0.6s, ssa-opt ~0.4s. Playbook đầy đủ: knowledge/bugs.md PERF#1-6.

**🎉 PERF#1 FIXED+SHIPPED (ee139a2 + docs c209aa0):** probe sâu → `allocate_registers_orchestrator`=2008ms → root cause: `x86_selector.ax` lower OP_FCONST/OP_NEG cấp GPR-scratch = `inst.dest + 60000` (hack "không-phải-AIR-dest"). Hệ quả mọi hàm float có vreg≈60000 → `graph_size=max_vreg+1≈60000` dù chỉ vài chục vreg sống → mọi vòng O(graph_size) trong graph_coloring_alloc chạy 60000×/hàm = ~90% codegen. **Fix: `inst.dest+60000` → `next_vreg(sel)`** (dense, cùng phân loại byte-for-byte). Kết quả: **codegen 2074→126ms (~16×), tổng build ~2.5s→0.34s (~7×)**, GS 60000→13-95. Fixpoint SHA=72c9aefc + regression 93/93 + exit=28. Playbook: knowledge/bugs.md PERF#1. Bài học: hack unique-vreg-bằng-offset-lớn = anti-pattern (mọi array keyed max_vreg trả giá); ssa_opt cũng dùng max_reg+1 arrays (cùng rủi ro nếu vreg thưa-cao).

**Cập nhật cuối session 2026-07-03:** Tất cả GREEN+pushed. Ngoài BUG#51/52: RFC 0010 (value-semantics) + RFC 0009 §10 (thiết kế FFI COFF .idata multi-DLL đã reverse-engineer). **Tooling mới `scripts/build_native.ps1`** (+regen_concat.ps1, commit 2467c38): daily-driver 20s tái lập (regen concat→self-build×2→promote nếu SHA fixpoint; `-Bootstrap` cold-start). Gate chuẩn: `AXC=bin/axc_native.exe bash scripts/regression_repros.sh` (93/93). **FFI P1 descope có chủ đích:** linker `.idata` = rewrite ~250 dòng hàm tải-nặng-nhất → focused effort riêng có kiểm chứng tăng dần, KHÔNG blind-rewrite (CLAUDE.md §16); frontend prototype (parser `from`+resolver `dll_binds`) đã chứng minh khả thi rồi revert giữ cây sạch. **NEXT session:** RFC 0010 P1 (hoàn tất audit→lower_let copy) HOẶC FFI linker focused — cả hai là focused effort riêng.
