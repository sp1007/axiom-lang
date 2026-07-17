---
name: bignum-ctgc-conflict
description: "std.bignum đứng ngoài hệ CTGC, rò rỉ bộ nhớ + copy-by-value con trỏ sở hữu (nghịch single-owner); cần xử lý sau (RFC owned-struct-fields)"
metadata: 
  node_type: memory
  type: project
  originSessionId: 9720703b-c141-45a7-9477-748fe53e183d
---

**Vấn đề (phát hiện 2026-06-29, để giải quyết sau):** Cách quản lý bộ nhớ của `std.bignum` xung đột với *tinh thần* (không phải luật) quản lý bộ nhớ AXIOM.

Triết lý AXIOM: single-owner + compile-time GC. Pipeline = Ownership → Escape → CTGC → Alias-reuse. CTGC chèn `DestroyStmt` cuối scope cho VarDecl gắn cờ `FlagEscapesToHeap` (chưa move/return), lower thành **`ax_free` một con trỏ** ([cgen.ax:1059]). Không GC runtime, không free thủ công ở code idiomatic.

`BigUint = {limbs: ptr[u64], n: i64}` (struct 16B value); heap thật là buffer `@alloc` sau con trỏ thô. **Không có `@free`/`bu_free`/`destroy` nào trong std/bignum.ax** (grep rỗng).

Ba điểm:
1. **Không phá luật**: `@alloc`+`ptr[T]` thô là tầng thủ công hợp lệ (runtime/compiler cũng dùng).
2. **Rò rỉ thật + bỏ qua CTGC**: heap sau con trỏ thô không được ownership-graph track; struct 16B nằm off-heap nên CTGC không sinh `ax_free`. `bu_shl`(loop gọi `bu_shl1`), `bf_mul`(buffer `full` 2n), mọi op trả-new → trung gian rò hết.
3. **Nghịch single-owner**: copy-by-value `let b = a`/`mut r := a` chia sẻ cùng `limbs` → 2 chủ 1 buffer; chưa nổ vì không free, nhưng thêm free đúng chỗ = double-free/UAF.

**Hướng giải quyết (ưu tiên):**
- Idiomatic nhất: `BigUint` thành *owned type có destructor* → CTGC tự `ax_free(limbs)`, alias-reuse recycle qua vòng. **Vướng: chưa có cơ chế "drop glue cho field con trỏ"** (grep destructor/drop/owned-field rỗng) → **cần RFC: owned struct fields + drop glue**.
- Hoặc: đại diện `BigUint` bằng MỘT khối `@alloc` (header+limbs chung), trả như một con trỏ owned.
- Thực dụng ngay: thêm `bu_free` tường minh + arena/scratch cho chuỗi phép tính; cấm copy-by-value-rồi-free-2-lần.

Liên quan [[next-step-15-selfhost-status]]. Chưa sửa gì — mới phân tích.

**Cập nhật 2026-07-06:** khi thiết kế fix (xem [[rfc0014-drop-glue-blocked]], `drop(self)`
hook), phát hiện vấn đề GỐC RỄ còn sâu hơn nhiều: [[bug69-ctgc-ownership-escape-noop]] —
CTGC (cùng OwnershipChecker + EscapeAnalyser) hiện là NO-OP HOÀN TOÀN (bug guard
`node_idx==0` bắt nhầm AST root), tức là KHÔNG CÓ compile-time GC nào từng chạy cho bất kỳ
chương trình AXIOM nào, không riêng gì bignum. Việc fix bignum cụ thể giờ phụ thuộc vào một
nỗ lực "CTGC activation" lớn hơn nhiều — xem memory đó để biết chi tiết + lý do chưa bật
ngay được (regression 0/98 khi thử).
