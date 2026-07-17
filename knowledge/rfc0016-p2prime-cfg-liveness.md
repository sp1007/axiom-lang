---
name: rfc0016-p2prime-cfg-liveness
description: "RFC 0016 P2' SHIPPED e3f9539: x86_regalloc.compute_liveness giờ CFG-aware (live-in/live-out dataflow), không còn linear back-edge hack. Nền cho mọi lowering tạo block giữa biểu thức."
metadata: 
  node_type: memory
  type: project
  originSessionId: d66677ec-a683-4d4f-998c-b3b5aaa4aa27
---

# RFC 0016 P2' — CFG-aware liveness (SHIPPED `e3f9539`, 2026-07-09)

`x86_regalloc.compute_liveness` (bootstrap/stage1/x86_regalloc.ax) từng là **linear-scan theo chỉ số instruction** + back-edge hack — chỉ đúng khi thứ tự mảng ≈ thứ tự CFG. Bất kỳ lowering nào phát block GIỮA biểu thức (if/else merge, short-circuit, tương lai: ternary-expr, `?`-propagation, pattern guard) đều phá giả định này.

**Nay:** dựng basic-block CFG từ mảng MachInst (split ở LABEL + sau JMP/JCC/RET; succ JMP→target, JCC→target+fallthrough, RET→none), giải **live-in/live-out bằng backward bitset dataflow** (`live_in[b]=use[b] | (live_out[b] & ~def[b])`, `live_out[b]=U live_in[succ]`), rồi **MỞ RỘNG** interval tuyến tính cơ sở tới linear-hull của mọi block sống. Chỉ grow → an toàn correctness (tệ nhất là spill thừa, không bao giờ miscompile).

**Helper `inst_defs_dst(op)`**: def-set CHỈ gồm op thực sự ghi dst. ⚠️ CMP/TEST/STORE/FCMP đọc dst nhưng KHÔNG ghi — nếu coi là def sẽ kill sai liveness. Không chắc → không phải def (conservative).

**Ý nghĩa:** đây là fix NỀN. RPO block-reorder (P2) đã thử: self-host được nhưng regress t_math (loop-carried float) → chứng minh phải sửa CHÍNH liveness, không phải thứ tự block. Mở khóa BUG#86 short-circuit ([[bug86-short-circuit-open]]) và mọi mid-expression-block lowering sau này. Gate: A!=B→B==C (`BCEFC38F`), 114/114, t_math=127. Chi tiết đầy đủ ở repo `rfcs/0016-cfg-aware-liveness-block-ordering.md`.
