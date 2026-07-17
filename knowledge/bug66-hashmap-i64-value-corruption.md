---
name: bug66-hashmap-i64-value-corruption
description: "BUG#66 FIXED (73f396d): std.collections.HashMap[K, i64] insert().get().unwrap() từng trả 0 thay vì giá trị thật. Root cause THẬT SỰ (khác với nghi ngờ ban đầu): infer_generic_type_args ghi đè vô điều kiện binding của 1 generic param mỗi lần khớp param position, khiến literal-arg mặc định i32 (param `value: V`) đè lên binding ĐÚNG lấy từ self's GENERIC_INST args (param `self: M[K,V]`, xử lý trước). Fix: first-binding-wins."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#66 FIXED — 73f396d, 2026-07-06. Fixpoint A==B verified, regression 94/94 (thêm row `t_hashi64`).**

## Triệu chứng (đã CONFIRMED qua `import std.collections` thật)

```
HashMap[i64, i64] / HashMap[str, i64]: insert(k, v).get(k).unwrap() → 0 (SAI)
HashMap[K, i32] / HashMap[K, f64]: hoạt động bình thường
```

## Root cause THẬT SỰ (khác hẳn nghi ngờ ban đầu trong phiên trước)

Phiên trước nghi ngờ `NODE_FIELD_EXPR`'s GENERIC_INST name-matching loop (typecheck.ax dòng ~2245-2256) — **ĐÓ LÀ ĐÁNH LẠC HƯỚNG**. Trong repro tối giản (chỉ 1 struct `M`), loop đó luôn tìm đúng struct duy nhất — không phải nguồn lỗi.

**Nguồn lỗi thật nằm ở `infer_generic_type_args`** (typecheck.ax dòng 291, cụ thể dòng 300-304), hàm suy luận generic param binding khi gọi 1 hàm generic:

```
mut g_idx := 0 as i64
while g_idx < gen_params.len:
    if gen_params.data[g_idx] == node_name_id:
        inferred.data[g_idx] = arg_type   // <-- GHI ĐÈ VÔ ĐIỀU KIỆN, bug ở đây
        break
```

Hàm này được gọi 1 lần cho MỖI param position của lời gọi hàm generic (typecheck.ax dòng 1581-1590, vòng lặp theo THỨ TỰ khai báo param). Với:
```
fn set_at[K, V](mut self: M[K, V], idx: i64, key: K, value: V): ...
set_at(m, 7, 1, 100)   // m: M[i64,i64]
```
- param 0 (`self: M[K,V]`) xử lý TRƯỚC: khớp cấu trúc `M[K,V]` với arg thật `GENERIC_INST(M,[i64,i64])` → bind ĐÚNG `K=i64, V=i64`.
- param 2 (`key: K`) và param 3 (`value: V`) xử lý SAU: arg thật là LITERAL `1`/`100` — kiểu mặc định của literal chưa gõ kiểu là **i32** — GHI ĐÈ lên `K`/`V` đã đúng, biến thành `K=i32, V=i32`.
- Hàm được monomorphize với `V=i32` → `self.values[idx] = value` (values thật là `ptr[i64]`) sinh OP_STORE scale=4 thay vì 8.

`get_at[K,V](self: M[K,V], idx: i64)` KHÔNG có param nào khác kiểu K/V ngoài `self` → không bị ghi đè → luôn đúng. **Giải thích chính xác vì sao write sai mà read đúng** (bất đối xứng quan sát được từ đầu).

## Fix (minimal, đã ship)

`typecheck.ax` dòng ~300-304: chỉ set `inferred.data[g_idx]` nếu CHƯA có binding (`== TYPE_UNKNOWN`) — first-successful-binding wins. Vì `self`/receiver luôn là param đầu tiên trong thứ tự khai báo, binding cấu trúc-chính-xác của nó luôn thắng, các param sau (dễ bị literal mặc định sai kiểu) không còn ghi đè được nữa.

```
if gen_params.data[g_idx] == node_name_id:
    if inferred.data[g_idx] == TYPE_UNKNOWN:
        inferred.data[g_idx] = arg_type
    break
```

**Giới hạn đã biết của fix (chấp nhận được, không phải bug mới)**: đây là "first-wins", không phải hợp nhất/kiểm-tra-nhất-quán thật sự. Nếu param ĐẦU TIÊN tự nó là literal chưa gõ kiểu (kiểu mặc định sai) VÀ có param sau đó mang kiểu đúng, fix này sẽ vẫn giữ binding sai. Trường hợp đó KHÔNG xảy ra với receiver-style method (`self` luôn đầu), là pattern phổ biến nhất (HashMap/Vec/collections). Một fix "đúng" đầy đủ cần bidirectional literal inference — quá lớn so với scope bug này, không làm ở đây theo đúng nguyên tắc tối giản của CLAUDE.md.

## Bài học / lịch sử điều tra

Phiên trước tốn nhiều công sức nghi oan cho `NODE_FIELD_EXPR`'s name-based struct lookup (xem [[bug64-vec-big-aggregate-element]] BUG#65 — đó VẪN LÀ một rủi ro kiến trúc thật, nhưng KHÔNG PHẢI nguyên nhân của BUG#66). Bài học: khi 1 giá trị "lẽ ra generic placeholder" hoá ra là 1 concrete type_id cụ thể (ở đây: 3 = i32), luôn nghi ngờ TRƯỚC HẾT cơ chế suy luận/binding generic param của lời gọi hàm (call-site inference), không chỉ cơ chế tra cứu field/struct.

Test regression: `bin/t_hashi64.ax` (row `t_hashi64|exit|42` trong `scripts/regression_repros.sh`).

Liên quan: [[bug64-vec-big-aggregate-element]] (BUG#65 — họ kiến trúc gần, nhưng khác cơ chế lỗi thật sự).
