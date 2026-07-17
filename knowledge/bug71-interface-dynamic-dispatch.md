---
name: bug71-interface-dynamic-dispatch
description: "BUG#71 FIXED (455e706): calling a method through an interface-TYPED value (`fn f(s: Shape): s.area()`) segfaulted — interface names never got a real type_id in typecheck (NODE_INTERFACE_DECL was never processed at all), so the receiver was untyped and method resolution silently picked an arbitrary same-named method by scope lookup. Fixed with a diagnostic reject (same class as BUG#53/64/68/70) + minimal register_interface giving interfaces a real TYPE_KIND_INTERFACE type_id. Direct calls on concrete structs implementing an interface still work fine — only true dynamic dispatch through the interface type is rejected."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#71 FIXED — 455e706, 2026-07-06.** Found via proactive probing right after shipping
[[bug70-array-literal-shipped]] (continuing the "probe an under-tested language area" pattern
this session, after [[bug69-ctgc-ownership-escape-noop]] turned out to need its own dedicated
effort rather than a quick fix).

## Triệu chứng

```
interface Shape:
    fn area(self: Self) -> i32

struct Square:
    side: i32
    fn area(self: Square) -> i32:
        return self.side * self.side

fn total_area(s: Shape) -> i32:
    return s.area()      // <-- SEGFAULT at runtime

fn main() -> i32:
    return total_area(Square(side: 3 as i32))
```
Gọi PHƯƠNG THỨC trực tiếp trên struct cụ thể hoạt động bình thường (`Square(...).area()` ổn).
Truyền struct cho một tham số KIỂU LÀ INTERFACE (`s: Shape`) rồi gọi `s.area()` bên trong
hàm đó (dynamic dispatch qua kiểu interface) → segfault. Basic parsing của `interface` đã có
sẵn test (`tests/generics/interface_basic.ax`) NHƯNG test đó không hề exercise dispatch thật
(chỉ khai báo, không gọi qua interface-typed value) — nên chưa từng lộ ra.

## Root cause

`typecheck.ax` **KHÔNG XỬ LÝ `NODE_INTERFACE_DECL` Ở ĐÂU CẢ** (grep rỗng trước fix) — khác
hẳn `NODE_STRUCT_DECL` vốn có hẳn 1 "Phase 0" pre-registration loop gán `sym.type_id` qua
`register_struct` TRƯỚC inference sâu. Vì vậy symbol "Shape" (đăng ký `SYM_INTERFACE` ở
`resolver.ax`) **KHÔNG BAO GIỜ có `type_id` thật** — mãi mãi = 0 (`TYPE_UNKNOWN`). Tham số
`s: Shape` do đó có kiểu THỰC TẾ = `TYPE_UNKNOWN`, không phải "một interface" theo nghĩa nào
cả.

Khi gọi `s.area()`, `resolve_method_sym(field_name_id="area", rec_type=TYPE_UNKNOWN)`
(typecheck.ax:716) — bước đầu tiên (`self.symtable.resolve(field_name_id)`, scope-based,
KHÔNG lọc theo receiver type) tìm được MỘT symbol "area" nào đó (Square.area hoặc Rect.area,
tùy thứ tự đăng ký) — và vì `rec_type=TYPE_UNKNOWN`, các bước lọc compatibility sau đó
(`resolve_method_overload`/`is_method_compatible`) không đủ nghiêm để reject đúng "receiver
không xác định = không tương thích", nên trả về MỘT method BẤT KỲ. Gọi method đó với `self`
là giá trị interface-typed value → layout `self` KHÔNG khớp với method thực sự nhận (có thể
khác struct, khác field offset) → đọc sai bộ nhớ → segfault.

**AXIOM chưa có bất kỳ runtime representation nào cho interface value** (không vtable,
không fat-pointer {data_ptr, vtable_ptr}) — `TYPE_KIND_INTERFACE` (typetable.ax, kind=13)
trước fix CHỈ xuất hiện ở đúng 1 chỗ (khởi tạo built-in placeholder trong `new_type_table`),
KHÔNG được check ở BẤT KỲ đâu trong air_builder/typecheck cho method dispatch thật. Đây
đúng lớp bug "tính năng tồn tại 1 phần (parse OK) nhưng chưa từng hoàn thiện" — giống hệt
[[bug70-array-literal-shipped]] (array literal có backend nhưng thiếu parser) chỉ khác
chiều: interface có PARSER + resolver symbol nhưng thiếu type-registration + dispatch thật.

## Fix — diagnostic reject (KHÔNG implement vtable/dynamic dispatch thật)

Theo đúng tinh thần tối giản CLAUDE.md và pattern đã dùng nhiều lần (BUG#53/64/68/70):
KHÔNG cố implement full trait-object dispatch (cần thiết kế ABI riêng cho interface value —
fat pointer hay gì khác — RFC riêng, phạm vi lớn hơn nhiều). Thay vào đó:

1. **`typetable.ax::register_interface(name_id) -> u32`** (mirror `register_array`'s độ đơn
   giản, KHÔNG cần side-table như StructInfo — interface chưa lưu method signature gì cả):
   push 1 `TypeEntry(kind: TYPE_KIND_INTERFACE, name_id: name_id, ...)`, dedupe theo name_id.
2. **`typecheck.ax`**: thêm 1 nhánh trong Phase-0 pre-registration loop (cạnh
   `NODE_STRUCT_DECL`'s), xử lý `NODE_INTERFACE_DECL`: gán `sym.type_id =
   register_interface(sym.name_id)` nếu chưa có. **Đây là điều kiện TIÊN QUYẾT** để bước
   sau (diagnostic) có gì đó THẬT để phát hiện — nếu bỏ qua bước này, receiver vẫn
   `TYPE_UNKNOWN` và mọi guard `entry.kind == TYPE_KIND_INTERFACE` sẽ KHÔNG BAO GIỜ fire
   (đúng loại lỗi "guard trông đúng nhưng inert" mà session này đã gặp ở c2a2a15 — ĐÃ RÚT
   KINH NGHIỆM, verify bằng debug-build TRỰC TIẾP trước khi tin).
3. Tại điểm gọi method thật (`NODE_CALL_EXPR` với callee = `NODE_FIELD_EXPR`,
   typecheck.ax ~1737): TRƯỚC KHI gọi `resolve_method_sym`, unwrap ptr/ref layer của
   `rec_type`, nếu kind == `TYPE_KIND_INTERFACE` → emit lỗi rõ ràng ("dynamic dispatch
   through an interface-typed value is not yet supported") + `diags_count++`, KHÔNG gọi
   resolve_method_sym nữa (chặn từ gốc, không dựa vào resolve trả về gì).

**Verify EMPIRICAL (không chỉ tin code "trông đúng")**: build lại 3 test case bằng compiler
đã sửa — (a) case segfault gốc → giờ lỗi biên dịch rõ ràng, KHÔNG segfault; (b) gọi method
trực tiếp trên struct cụ thể (không qua interface param) → vẫn chạy đúng (exit 16); (c)
tham số interface-typed KHÔNG dispatch (chỉ nhận, không gọi method) → vẫn chạy đúng (exit
7, không bị false-positive reject). Cả 3 test sẵn có `tests/generics/interface_*.ax` build
sạch không lỗi mới. Fixpoint A==B, regression 99/99, negative test mới
`tests/sema/err_interface_dynamic_dispatch.ax`.

**Ngụ ý**: interface trong AXIOM hiện tại CHỈ hữu ích như static structural constraint
(giống Go interface dùng ở compile-time cho generic bound, xem `tests/generics/
generic_constrained.ax`) — KHÔNG dùng được cho runtime polymorphism/dynamic dispatch qua
interface-typed value/tham số/collection. Nếu cần trait-object thật (Box<dyn Trait> kiểu
Rust, hay interface value kiểu Go), cần RFC riêng thiết kế ABI (fat pointer hay
monomorphization tại mỗi call site) — CHƯA BẮT ĐẦU, ngoài phạm vi fix này.

Liên quan: [[bug70-array-literal-shipped]] (cùng session, cùng pattern "tính năng
half-shipped lộ ra qua proactive probing"), BUG#53/64/68 (cùng lớp "diagnostic reject thay
vì miscompile im lặng").
