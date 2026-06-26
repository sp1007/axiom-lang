# RFC 0001 — ABI cho aggregate 16-byte truyền/return BY-VALUE

- **Status:** Implemented (2026-06-26) — Phương án B; fixpoint giữ (stage3==stage4 = d7f14c2c)
- **Author:** (self-host team)
- **Created:** 2026-06-26
- **Tracking bug:** Family C (knowledge/bugs.md), repro `bin/tsp.ax`
- **Liên quan:** BUG#29 (str C-ABI sret), BUG#30 (struct = by-address 8-byte pointer)

---

## 1. Motivation (Động lực)

Hiện tại truyền/return một `struct` 16-byte BY-VALUE sinh code SAI:

```axiom
struct P:        # size = 16 (a:i64, b:i64)
    a: i64
    b: i64
fn mk(x: i64, y: i64) -> P:
    return P(a: x, b: y)
fn sum(p: P) -> i64:
    return p.a + p.b
pub fn main() -> i32:
    let q = mk(3, 4)
    return sum(q) as i32     # KỲ VỌNG 7
```

- stage1 `-O1` → trả **16** (sai).
- stage1 `-O0` → trả **127** (sai khác).

Lỗi KHÔNG block self-hosting (compiler nội bộ truyền struct qua `ptr`/`ref`, không by-value 16-byte), nên đã **defer**. Nhưng đây là lỗ hổng ngữ nghĩa ngôn ngữ cần đóng đúng bài.

## 2. Bối cảnh kỹ thuật hiện tại

Hai biểu diễn 16-byte ĐANG TỒN TẠI và XUNG ĐỘT:

1. **`str`** (type_id 12, kind PRIMITIVE) = giá trị 16-byte INLINE `{ptr:i64, len:i64}`.
   - Return: RAX (low 8) : RDX (high 8) — xem `x86_selector.ax:1040` (OP_RETURN), `:1293` (capture RAX:RDX sau call).
   - Args: truyền BY-POINTER (callee đọc 2 nửa qua con trỏ).
   - `regalloc_is_16byte` → true ⇒ vreg chiếm **2 spill slot** (16-byte inline home).

2. **`struct`/sum/option/result/array/tuple** (aggregate) = giá trị **8-byte HEAP POINTER** (by-address, BUG#30).
   - `let q = mk()` ⇒ q giữ con trỏ tới data struct.
   - field access = deref `[ptr+disp]`.
   - `regalloc_is_16byte` PHẢI → false (1 slot, 8-byte pointer).

`regalloc_is_16byte` (`x86_selector.ax:435`) là nơi quyết định 1-slot-pointer vs 2-slot-inline. Case OP_COPY (`:496`) đã guard `not type_is_aggregate` (BUG#30). Nhưng case **param (442)**, **call-return direct (469) + dynamic (477)**, **cast (511)** CHƯA guard → một struct 16-byte by-value bị coi là inline-16.

## 3. Tại sao "fix 1 dòng" THẤT BẠI (đã thử + revert 2026-06-26)

Thêm `and not type_is_aggregate(...)` vào param(442)+call-return(469/477):
- `tsp` đổi từ trả-sai-16 → **SEGFAULT** (str regression vẫn OK).
- **Nguyên nhân:** chỉ đổi `regalloc_is_16byte` làm callee coi param/return là 8-byte pointer, NHƯNG:
  - `mk` return một con trỏ tới struct nằm trên **stack frame của mk** (hoặc tạm) → sau khi `mk` ret, con trỏ trỏ vùng đã chết → `sum` deref → crash.
  - Không có phía nào CẤP storage bền (heap/caller-frame) + COPY 16 byte. Cả 3 phía (call-arg lowering, param receipt, return emission) phải nhất quán; sửa lẻ regalloc chỉ tạo mismatch mới.

⇒ Đây là thay đổi **ABI**, theo CLAUDE.md §13 phải qua RFC. Không vá bằng guard regalloc đơn lẻ.

## 4. Design đề xuất (KHUYẾN NGHỊ: Phương án B — sret/by-address nhất quán)

Giữ mô hình BUG#30 (aggregate = by-address). Truyền/return by-value = **copy nội dung qua con trỏ**, KHÔNG copy con trỏ.

### 4.1 Return (sret — structural return pointer)
- Hàm `-> P` (P aggregate, size > 8 hoặc bất kỳ aggregate by-value) nhận **hidden param sret** = con trỏ tới vùng caller cấp (RCX trên Win64 / RDI trên SysV, đẩy các arg thật lùi 1 ô).
- `return P(...)`: callee `memcpy(sret, &data, size)` rồi `return` (RAX = sret theo convention).
- Caller `let q = mk()`: cấp slot/heap cho q TRƯỚC, truyền địa chỉ làm sret, sau call q = vùng đó (đã có data bền).

### 4.2 Param by-value
- `fn sum(p: P)`: caller `memcpy` (hoặc cấp bản sao) data vào vùng tạm, truyền **địa chỉ bản sao** (đảm bảo ngữ nghĩa value: callee sửa p KHÔNG ảnh hưởng caller).
- callee `p` = con trỏ (8-byte), field access deref như aggregate thường (đúng BUG#30).
- `regalloc_is_16byte(param)` → **false** cho aggregate (guard `not type_is_aggregate`) — KHỚP với 8-byte pointer.

### 4.3 `str` GIỮ NGUYÊN inline-16 (RAX:RDX)
- `str` là PRIMITIVE, không aggregate ⇒ tất cả guard `not type_is_aggregate` bỏ qua nó ⇒ vẫn inline-16. Không đụng path str (đã đúng, đang dùng khắp compiler).

### 4.4 Điểm chạm code
1. `x86_selector.ax regalloc_is_16byte` — thêm guard `not type_is_aggregate` vào param/call-return/cast (NHƯNG chỉ sau khi các điểm dưới xong).
2. Call lowering (OP_CALL): với arg aggregate by-value → cấp temp + memcpy + truyền địa chỉ. Với hàm trả aggregate → chèn hidden sret arg.
3. OP_RETURN: hàm trả aggregate → memcpy vào sret thay vì nạp RAX:RDX.
4. air_builder/typecheck: đánh dấu hàm có sret; truyền sret qua signature.
5. ABI layer (Win64 vs SysV): vị trí thanh ghi sret + dịch arg.

## 5. Alternatives (Phương án thay thế)

- **A. Inline-16 như `str`:** coi struct 16-byte by-value là inline (RAX:RDX, 2-slot). ✗ Xung đột BUG#30 (struct là by-address ở mọi chỗ khác) → cần biểu diễn struct "split" inline-vs-pointer theo ngữ cảnh → phức tạp, dễ vỡ. Không chỉ áp dụng được cho struct >16 byte.
- **C. Cấm struct by-value:** bắt buộc `ptr[P]`/`ref[P]`. ✓ Đơn giản nhất, an toàn. ✗ Hạn chế ngôn ngữ, khác kỳ vọng systems-language; chỉ nên là giải pháp tạm.
- **B (chọn):** sret/by-address. ✓ Nhất quán BUG#30, tổng quát cho mọi size aggregate, đúng chuẩn (giống Win64/SysV cho struct lớn). ✗ Nhiều điểm chạm + cần test kỹ.

## 6. Drawbacks

- Thay đổi calling convention nội bộ (signature có hidden sret) → mọi caller/callee phải đồng bộ trong CÙNG một lần build (self-host: stage1 phải sinh stage2 với ABI mới nhất quán).
- Rủi ro vỡ self-host fixpoint (vừa đạt: stage3==stage4 SHA cf5a7c6a). Phải verify lại 2.5h sau khi implement.
- Tăng copy (memcpy) cho mỗi truyền by-value — đánh đổi đúng-đắn lấy hiệu năng (chấp nhận; có thể tối ưu elide sau).

## 7. Migration plan

1. Implement §4.4 điểm 2–5 (call/return/abi) TRƯỚC, giữ `regalloc_is_16byte` cũ (để không vỡ giữa chừng).
2. Bật guard `not type_is_aggregate` (điểm 1) CUỐI CÙNG khi 3 phía đã nhất quán.
3. Test: `bin/tsp.ax`=7; struct >16 byte (3+ field) by-value; nested; `str` param/return vẫn đúng (`t_strip`, `t_movrr`); mutation semantics (callee sửa param không ảnh hưởng caller).
4. `scripts/verify_bug29_selfhost.sh` — fixpoint stage3==stage4 PHẢI giữ.
5. Golden backup `bin/axc_stage{2,3}_selfhost_fixpoint.exe` để rollback nếu vỡ.

## 8. Compatibility impact

- ABI nội bộ (chưa có ABI ổn định public) → an toàn đổi.
- `str` không đổi → mọi code str hiện tại giữ nguyên.
- Aggregate by `ptr`/`ref` (cách compiler đang dùng) không đổi.
- Object/exe sinh ra đổi bytes (signature mới) → cần verify lại SHA fixpoint.

## 9. Test plan tối thiểu (definition of done)

- [x] `tsp.ax` → 7 (O0 và O1)
- [x] struct 24-byte by-value param+return đúng (`tsp3.ax` → 12; `tsp2.ax` qua 2 call → 9)
- [x] `t_strip`, `t_movrr`, `t_param5`, `t_cse`, `t_cpaddr`, `t_modrm` không regression
- [x] `verify_bug29_selfhost.sh`: stage3==stage4 fixpoint giữ (d7f14c2c)
- [x] nested struct by-value (`bin/tstruct_abi.ax` D=6), str-field by-value (C=15)
- [x] semantics tham chiếu nhất quán đã xác minh (E=99) — xem §11

## 10. Đã implement (2026-06-26)

Theo Phương án B. 4 điểm guard `not type_is_aggregate` trong `x86_selector.ax`:
1. `regalloc_is_16byte` — param case (~442), call-return direct (~471) + dynamic (~481).
2. `emit_param_prologue` — register-passed (~1688) + stack-passed (~1709): aggregate param lưu thẳng con trỏ 8-byte (else branch) thay vì deref+copy 16 byte.

Cách tìm root (gdb): fault `mov (%r10),%rcx` với r10 = giá trị field (3) ⇒ param materialize đang copy 16-byte-inline rồi getfld lại deref ⇒ lần ra `emit_param_prologue` (KHÔNG phải path OP_COPY lazy — đã chết vì `param_idx_processed` đặt quá cuối).

## 11. Semantics quyết định (2026-06-26): REFERENCE — KHÔNG phải value-copy

Đã điều tra triệt để ngữ nghĩa truyền struct (để không tích lũy bug). **Quyết định: struct dùng REFERENCE semantics nhất quán** (không memcpy bản sao). Lý do:

1. **Khớp model spec:** AXIOM LANGUAGE SPEC dùng Single-Ownership + heap object + Generational References. Struct value = con trỏ tới heap object (BUG#30). Move-semantics chỉ bắt buộc cho `qbit`.
2. **Nhất quán đã verify** (`bin/tstruct_abi.ax`):
   - `let r = q` rồi mutate r ⇒ q đổi theo (alias) — đã ghi ở BUG#30 (`let x=y; x.a=9 ⇒ y.a==9`).
   - `bump(mut p: P)` mutate p ⇒ caller's object đổi (E=99). GIỐNG hệt assignment alias ⇒ NHẤT QUÁN.
3. **Compiler tự host OK** với model này (cả codebase giả định reference; nếu đổi sang value-copy sẽ phá fixpoint + toàn bộ ownership/CTGC).

⇒ KHÔNG thêm memcpy. Truyền con trỏ 8-byte tới object gốc là ĐÚNG và nhất quán. §4.2 (memcpy bản sao) bị BÁC vì sẽ tạo bất nhất với assignment-alias. Nếu tương lai muốn value-type rõ ràng, đó là feature ngôn ngữ riêng (vd `Copy[T]` / explicit clone), cần RFC khác — KHÔNG sửa ngầm ở ABI.

**Definition of done: ĐẠT.** Family C đóng triệt để, không còn TODO treo.
