---
name: bug-method-call-overload-ignores-arity
description: lời gọi viết bằng CÚ PHÁP METHOD (s.f(41), method inline, S.f(&s,41)) chọn overload chỉ theo RECEIVER, bỏ qua arity — trong khi lời gọi tự do đã gate theo arity + kiểu đối số
metadata:
  type: project
---

# BUG MỞ — chọn overload ở lời gọi **cú pháp method** bỏ qua ARITY

**Trạng thái:** OPEN (tách ra 2026-07-31, khi sửa
[[bug-struct-receiver-overload-symbol-collision]]). **Đây là defect KHÁC**, không phải phần còn
lại của cái kia.

## Hiện tượng

```axiom
struct St:
    n: i64
    fn k(self: St) -> i64:      return 1
    fn k(self: St, x: i64) -> i64:  return x

St.k(&st, 41)   // => 1   (đúng phải 41)
St.k(&st)       // => 1
st.k(41) + st.k()   // => 2  (đúng phải 42)
```

Cùng hình dạng với free fn gọi kiểu UFCS: `fn f(s: S)` + `fn f(s: S, x: i64)` rồi
`s.f(41) + s.f()` = **2**.

## Quy trách nhiệm CHÍNH XÁC (không suy đoán)

Trên **MỘT bản compiler** (sau khi hole C đã sửa, `A==B==C 52D1ABD4`), **cùng một bộ khai báo**:
- gọi dạng **tự do** `f(s, 41) + f(s)` ⇒ **42** ✅
- gọi dạng **method** `s.f(41) + s.f()` ⇒ **2** ❌

Symbol phát ra là như nhau ở cả hai chương trình ⇒ khác biệt **chỉ nằm ở khâu BIND**, không phải
ở tên symbol. Đây là lý do không được gộp nó vào hole C: hole C đã sửa xong và đo được, còn cái
này vẫn nguyên.

## Cơ chế

`typecheck.ax:1135` `resolve_method_overload(sym_idx, rec_type)` duyệt chuỗi `next_overload` và
trả về **phần tử ĐẦU TIÊN tương thích RECEIVER** (`is_method_compatible`), **không hề nhìn arity**
— nó thậm chí không nhận call node nên không biết số đối số. Trong khi đó
`resolve_free_call_overload` (`typecheck.ax:1340`) gate theo `sym_decl_param_count(ci) ==
call_argc` **và** kiểu đối số, có cả tie-break chấm điểm toàn bộ danh sách tham số.

⇒ **Đúng lớp defect chủ đạo của repo: MỘT luật, HAI bản sao, một bản không bao giờ được mở rộng.**
Cùng họ với `ownership.ax:138,162` (chưa từng chuyển sang rank mới) và
`verify_air_no_int_into_float` (không có khái niệm bề rộng).

## Vì sao chưa sửa luôn (quyết định, không phải lười)

Hai điểm cần quyết trước khi viết code, cả hai đều ở tầng **ngữ nghĩa ngôn ngữ**:
1. **Overload theo arity ở cú pháp method có phải feature được hỗ trợ không?** Đường free-call
   nói CÓ. Nếu câu trả lời là KHÔNG thì cách sửa đúng theo chính sách BUG#53 là **REJECT** kèm
   chẩn đoán, chứ không phải bind cho đúng.
2. `resolve_method_sym` (2 call site ở `typecheck.ax`) **không có** call node/argc; và
   `air_builder` còn có vòng phân giải method **riêng** của nó, dính vào bộ máy rank của
   RFC 0037 (`a281992`, `a538983` vừa đụng vào). Luồn `call_argc` qua là thay đổi nhiều tầng,
   không phải one-liner.

## Oracle khi sửa

Repro tối thiểu: `bin/probe8/g1_arity_overload.ax` (method inline) và
`bin/probe8/g2_freefn_arity.ax` (free fn gọi kiểu UFCS), cả hai = **2**, đúng phải 42.
Dạng gọi tĩnh `St.k(&st, 41)` = **1**, đúng phải 41.
Khi sửa xong, thêm lại các hàng này vào `bin/t_structoverload.ax` (header của file đã ghi rõ
chúng bị loại vì defect này) và chạy ở CẢ `-O0` lẫn `-O1`.

## Liên quan
[[bug-struct-receiver-overload-symbol-collision]] · RFC 0035 §2bis · RFC 0037 (rank) ·
[[BACKLOG]].
