# RFC 0010 — Value semantics cho aggregate `let` bindings (diệt lớp bug stale-alias)

- **Status:** Draft (2026-07-03)
- **Author:** self-host team
- **Tracking:** follow-up bắt buộc của BUG#51/#52 (fix targeted 4d7949a)
- **Liên quan:** air_builder (lower_let / lower_index_expr / OP_INDEX / OP_STORE),
  x86_selector (OP_INDEX aggregate branch "held by-address"), cgen (OP_INDEX copy),
  ownership.ax, ctgc.ax
- **Blocks:** độ tin cậy dài hạn của MỌI code AXIOM dùng `let s = arr[i]` với struct;
  tiền đề an toàn cho std.collections lớn + compiler tự-host không còn bom UAF.

---

## 1. Motivation — lớp bug "bom nổ chậm" đã nổ hai lần

BUG#51 (segfault deterministic khi native-built compiler compile t_mathx) và BUG#52
(module-load flaky) có CÙNG root cause:

```axiom
let callee_node = self.tree.nodes.data[callee]   // (1) lấy "giá trị" phần tử
...
mono.instantiate_function(...)                    // (2) tree.nodes GROW → realloc + free buffer cũ
...
if callee_node.kind == ... :                      // (3) BOOM (native) / đọc stale "may mắn đúng" (gcc)
```

Nguyên nhân gốc: **hai backend bất đồng ngữ nghĩa cho cùng một AIR**:

| Backend | `dest = OP_INDEX(arr, i)` với element = aggregate |
|---|---|
| cgen (C) | `r_dest = &arr[i]` — cũng là ĐỊA CHỈ (alias) |
| x86 native | địa chỉ phần tử (alias) — comment: *"Aggregates are held by-address in this backend"* |

Cả hai đều alias ở mức OP_INDEX, nhưng **hành vi lỗi khác nhau** khi alias dangle:
gcc/malloc giữ trang cũ mapped + nội dung nguyên (memcpy khi grow) → đọc stale trả
giá trị đúng y hệt → **bug bị che hoàn toàn**; native ax_free → VirtualFree block lớn
→ unmap → segfault. Kết quả: fixpoint SHA hội tụ trên code sai, chỉ lộ khi input
đủ lớn để vượt ngưỡng free-list→VirtualFree.

**Quy mô lớp bug:** ≥196 site `let X = <vec>.data[i]` với element aggregate chỉ riêng
bootstrap/stage1 (nodes/symbols/entries/tokens/...). Mỗi site là một quả bom nếu có
grow xen giữa bind và last-use. Fix targeted (4d7949a) chỉ gỡ MỘT site đã nổ.

## 2. Ngữ nghĩa hiện tại vs đề xuất

Spec v1.0 **không định nghĩa tường minh** copy-vs-alias cho `let x = arr[i]`
(đã grep: không có "value semantics"/"by value" trong spec — ambiguity per
CLAUDE.md §20). Hành vi thực tế hiện nay = **alias** (cả 2 backend), tức `let` một
aggregate từ index KHÔNG tạo bản sao — trái trực giác `let`-là-giá-trị và trái với
hành vi scalar (scalar được LOAD ra vreg = copy thật).

**Đề xuất ngữ nghĩa (normative):** `let x = <aggregate lvalue>` tạo **bản sao độc lập**
(block-copy) tại thời điểm bind. Ghi vào nguồn sau đó (kể cả realloc/free nguồn)
không ảnh hưởng `x`. `mut x := arr[i]` + `x.f = v` ghi vào BẢN SAO (không write-through).
Ai muốn alias thì dùng con trỏ tường minh: `let p = &arr[i]` (giữ nguyên hành vi).

## 3. Design — fix tại AIR lowering, KHÔNG đụng backend

Chìa khóa: backend đã có sẵn mọi cơ chế cần thiết:
- `OP_ALLOCA` (slot stack cỡ N) — dùng cho struct literal/local struct.
- `OP_STORE` với aggregate = **block-copy theo địa chỉ** (selector comment: "aggregate
  value held by its ADDRESS in src1, must be block-copied").

**Thay đổi duy nhất — air_builder `lower_let`** (và `lower_assign` cùng dạng):

```
NODE_LET với init-type = aggregate (size > 16 hoặc is_aggregate) và init đến từ
lvalue-alias (OP_INDEX / GET_FIELD trả địa chỉ / DEREF):
  1. slot = OP_ALLOCA(size)
  2. addr = lower init như hiện tại (ra địa chỉ nguồn)
  3. OP_STORE(slot, addr)      // block-copy
  4. bind tên biến → slot      // thay vì → addr
```

- cgen: cùng AIR → emit `T r_slot; memcpy/struct-assign` — C backend xử OP_ALLOCA/
  OP_STORE sẵn → tự đúng.
- 16-byte (str, RFC 0001 aggregates): đã copy-by-value qua cơ chế 2-slot sẵn có —
  **loại khỏi scope** (không đổi gì).
- Downstream không đổi: mọi consumer aggregate (`.field`, call-arg, OP_STORE) đều
  nhận ĐỊA CHỈ như trước — chỉ khác là địa chỉ của bản sao trong frame.

## 4. Rủi ro & audit bắt buộc trước khi flip

1. **Write-through reliance:** code dựa vào `mut x := arr[i]; x.f = v` để ghi về
   mảng sẽ HỎNG ÂM THẦM (ghi vào copy). Audit stage1: 32 site `mut x := ...data[i]`,
   soi tay = đều read-only hoặc reassign (vd `op_unwrap_type`); style codebase ghi
   trực tiếp `arr[i].f = v`. → rủi ro thấp nhưng **audit lại toàn bộ khi implement**
   (grep `mut \w+ := .*\.data\[` + kiểm từng site có ghi field).
2. **Perf:** copy 24-48B × triệu lần trong self-build. Ước tính rẻ (memcpy nhỏ,
   locality tốt); đo bằng self-build time trước/sau (baseline 8s). Nếu >15% chậm →
   cân nhắc chỉ copy khi escape-analysis không chứng minh được no-grow-between.
3. **Frame size:** mỗi let-aggregate thêm slot N byte → mega-fn (infer_node ~trăm
   bindings) phình frame vài KB — chấp nhận được (frame hiện 18KB).
4. **CTGC/ownership:** bản sao stack không đổi ownership của dữ liệu trỏ-tới bên
   trong (con trỏ trong struct copy = shared view) — giữ nguyên ngữ nghĩa hiện có
   của single-owner (không double-free vì CTGC track theo biến gốc). Cần xác nhận
   với ctgc.ax khi implement.

## 5. Alternatives

- **Giữ alias + cấm grow-while-borrowed (borrow checker):** đúng đắn dài hạn nhưng
  là cả một hệ thống (RFC riêng, nhiều tháng); không chặn bom đang nằm sẵn.
- **Vá từng site khi nổ (hiện trạng):** không scale — 196+ site, chỉ lộ khi input
  vượt ngưỡng; chi phí hunt mỗi lần ~nửa ngày dù đã có workflow.
- **Đổi allocator native sang không-bao-giờ-unmap:** che bug thay vì fix (xác định
  hành vi bằng may rủi trang nhớ), phá tính deterministic-fail — bác bỏ.
- **Copy tại backend (selector OP_INDEX):** cần regalloc hiểu vreg-slot cỡ N tùy ý
  (hiện chỉ 8/16B) — blast radius lớn hơn nhiều frontend-lowering. Bác bỏ.

## 6. Kế hoạch thực hiện

- **P0 (RFC này):** chốt ngữ nghĩa + audit write-through toàn stage1/std.
- **P1:** implement `lower_let` copy cho aggregate-from-index (scope hẹp nhất);
  gates: full regression bằng axc_native + fast-fixpoint + self-build time so sánh.
- **P2:** mở rộng cho GET_FIELD-nested/DEREF inits + `lower_assign`; gỡ fix targeted
  4d7949a (không còn cần); thêm test t_aggcopy (bind → grow → đọc lại, oracle khác nhau
  giữa alias/copy).
- **P3:** cập nhật spec (ngữ nghĩa normative mục 2) + docs.

## 7. Success criteria

- t_aggcopy oracle PASS ở cả native lẫn C backend (hai backend đồng ngữ nghĩa).
- Fix targeted BUG#51 gỡ được mà t_mathx vẫn 28/28.
- Self-build time tăng ≤15% so với baseline 8s.
- Full regression + fixpoint GREEN trên native driver.
