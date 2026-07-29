---
name: bug-uninitialized-local-array-segfault
description: "OPEN BUG (probe-found 2026-07-30) — mảng LOCAL khai báo KHÔNG có initializer (`mut a: [i64; 2]`) được chấp nhận IM LẶNG rồi SEGFAULT ở mọi opt level. Bản GLOBAL cùng cú pháp chạy đúng; bản local CÓ initializer chạy đúng. Accept-then-miscompile."
metadata:
  type: project
---

# 🐞 OPEN — mảng LOCAL không có initializer: nhận im lặng rồi **SEGFAULT**

**Tình trạng: OPEN, chưa sửa.** Probe-found 2026-07-30, driver `99225522`.

## Reproducer tối giản (KHÔNG dính literal lớn, KHÔNG dính tối ưu)
```axiom
fn main() -> i64:
    mut arr: [i64; 2]
    arr[0] = 7
    arr[1] = 5
    return arr[0] + arr[1]
```
→ **SEGFAULT (139)** ở **CẢ `-O0` LẪN `-O1`**. Compiler **KHÔNG in một chẩn đoán nào** (exit 0,
không error/warning) — nó phát ra một binary chạy là chết.

## Đối chứng — đã khoanh, không suy đoán
| ca | kết quả |
|---|---|
| `mut arr: [i64; 2]` (LOCAL, **không** init) rồi gán từng phần tử | ❌ **SEGFAULT** |
| `mut a: [i64; 2] = [7, 5]` (LOCAL, **có** init) | ✅ 12 |
| `mut a: [i64; 2] = [0, 0]` rồi gán từng phần tử | ✅ 12 |
| `mut g: [i64; 2]` (**GLOBAL**, không init) rồi gán | ✅ 42 |

⇒ Chỉ hỏng ở **local + không initializer**. Cùng cú pháp ở **global thì CHẠY ĐÚNG**
(`t_globarrnoinit`, `elfglobuninit`, `t_bssglobal` đều pin ca global).

## Vì sao không ai thấy: **KHÔNG CÓ TEST NÀO** cho hình dạng này
Quét `bin/*.ax`: **mọi** khai báo mảng local trong suite đều CÓ initializer
(`t_arrassign.ax: mut a: [i64; 3] = [1, 1, 1]`); khai báo **không** initializer chỉ xuất hiện ở
**global scope**. ⇒ Vùng hoàn toàn chưa được test, không phải regression.

## ⭐⭐⭐ KHÔNG PHẢI câu hỏi thiết kế — **ngôn ngữ ĐÃ TRẢ LỜI cho scalar; AGGREGATE mới là chỗ lệch**
Bản ghi ĐẦU của mục này đóng khung đây là lựa chọn 50/50 "hỗ trợ hay từ chối, cần user quyết".
**Sai.** Probe tiếp cho thấy local KHÔNG khởi tạo **ĐÃ ĐƯỢC HỖ TRỢ SẴN** với scalar:

| local KHÔNG initializer | kết quả |
|---|---|
| `mut x: i64` (scalar) | ✅ chạy đúng (7) |
| `mut s: str` | ✅ chạy đúng (2) |
| `mut p: P` (**struct**) | ❌ **SEGFAULT** |
| `mut arr: [i64; 2]` (**array**) | ❌ **SEGFAULT** |

⇒ Đây là **lỗi NHẤT QUÁN**, không phải câu hỏi ngữ nghĩa: scalar và str đã chạy, nên hướng sửa
**ĐÃ ĐƯỢC QUYẾT** — làm **aggregate hành xử GIỐNG scalar** (cấp slot stack), **KHÔNG** phải reject.
Reject bây giờ sẽ **phá tính nhất quán** mà scalar đang thiết lập.

⚠️ **VÀ NÓ LÀ CẢ MỘT HỌ, không riêng mảng**: struct hỏng y hệt. Sửa riêng mảng sẽ là **partial
fix** — đúng cái sai lặp đi lặp lại trong dự án ([[bug-f32-compare-float-literal]]: "thêm rule ở
1 nhánh mà quên nhánh sibling"). **Phải phủ MỌI kiểu aggregate** (array, struct, và kiểm luôn
tuple/sum nếu có).

Hành vi hiện tại (nhận im lặng → binary crash) vẫn thuộc họ **accept-then-miscompile**
([[bug-option-as-call-arg-not-rejected]], [[bug-option-as-let-binding-not-rejected]]).

## Gợi ý điều tra (chưa làm)
Nghi vấn: `NODE_VAR_DECL` cho kiểu ARRAY không có initializer → không cấp phát `local_bytes`
trên frame ⇒ `arr[0] = ...` ghi qua một địa chỉ chưa được đặt. So sánh với nhánh global
(RFC 0017/RFC 0030 P3 `.bss`) vốn cấp phát đàng hoàng. Xem `compute_frame` / chỗ tính
`local_bytes` trong `x86_coff.ax`, và nhánh VAR_DECL trong `air_builder.ax`.
⚠️ Khi sửa: oracle phải chạy **cả `-O0`** (xem [[bug-negative-literal-compare-o0]] — gate không
chạy cấu hình nào thì mù ở cấu hình đó), và cần **cả hai** oracle: ca dùng được (nếu chọn lối 1)
hoặc ca reject (nếu chọn lối 2), **cộng** một ca guard chống over-reject (mảng local CÓ init).
