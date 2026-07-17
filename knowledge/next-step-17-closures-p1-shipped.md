---
name: next-step-17-closures-p1-shipped
description: "Closures P1 + BUG#50 SHIPPED; fixpoint OK; phát hiện BUG#51 (native-built compiler miscompiled, chặn fast-driver 1200x) + BUG#52 (module loading UAF phi-tất-định)"
metadata: 
  node_type: memory
  type: project
  originSessionId: ed12f2e7-f4ab-41c0-9bb3-940acfa7aaec
---

**2026-07-03 — SHIPPED:**
- **Closures P1** `|a,b|->T` (9d35e54): lambda-lift MVP, regression 92/92, **fixpoint stage0-rooted CONVERGED** (stage3==stage4, xác nhận 14:51). RFC 0008 → P1 Implemented.
- **BUG#50 fix** (fb74dd8): "first import wins" — 2 tầng: (1) resolver define() dedup decl_node phải so thêm cây (`symbol_trees`), flag `SYM_FLAG_MODDUP=2048`; (2) cả 2 mangler emit `ax_<name>__m<sym_idx>` cho MODDUP. Oracle t_modcollide=101 hai chiều. RFC 0009 Accepted (3a76ca3).

**HAI BUG LỚN MỚI (OPEN, ưu tiên đầu):**
- **BUG#51:** binary compiler build bằng NATIVE backend segfault **deterministic** khi compile bin/t_mathx.ax; stage0-built compile OK. ⟹ native backend miscompile ≥1 hàm compiler; **SHA fixpoint ≠ correctness** — quy trình mới: chạy FULL regression bằng converged binary. **Fix này mở khóa daily-driver 1200x nhanh** (native-built self-build 790 hàm = 8s vs stage0-built = 2h50m — perf root-cause thật, không phải regalloc O(n²) như đoán đầu; regalloc có spill-all fallback >5000 insts).
- **BUG#52:** lazy module loading (ax_driver_load_module, main_air.ax ~1085) **UAF phi-tất-định**: cùng binary+input → OK/crash/hang ngẫu nhiên (ASLR-dependent). Nghi cleanup free lexer arrays khi ModuleInfo còn giữ ptr. Mọi test import local = flaky (t_modcollide registered nhưng flaky). Đốt 1h debug giả — **bài học: crash lạ → chạy lặp ≥5 lần trước khi đổ cho code mới.**

**Workflow mới đã chứng minh:** fast-fixpoint qua converged binary (build concat → self-build ×2 → SHA) = ~25s thay 3h; dùng để gate backend changes (a3==a4 OK cho BUG#50). Defender exclusion đã bật (repo + %TEMP%, qua UAC).

**Queue:** 1) BUG#51 (mở fast driver) → 2) BUG#52 (determinism) → 3) perf --time instrumentation → 4) FFI P1 (RFC 0009 accepted) → 5) BUG#44/TRẦN-STAGE0/Closures-P2.
