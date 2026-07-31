---
name: bug-mono-generic-ret-typaram-f64
description: method generic mono hoá trả `-> T` với T ↦ f64 trả về 0.0; bốn control thu hẹp đúng vào lớp thanh ghi trả về
metadata:
  type: project
---

# BUG MỞ — method generic mono hoá trả `-> T` với T ↦ f64 cho ra 0.0

**Trạng thái:** OPEN (phát hiện 2026-07-31, probe8). Lớp **accept-then-miscompile** (BUG#53) — nhận,
không chẩn đoán, chạy ra số sai.

## Triệu chứng
`bin/probe8/i8_genf64ret.ax` = **40**, đúng phải **42**. Sai ở **cả ba** mức `-O0` / default / `-O1`
⇒ **không** phải họ phân kỳ theo mức tối ưu (khác hole C, xem [[BACKLOG]]).

## ⭐ Bốn control đã thu hẹp sẵn — đừng đo lại, hãy dùng
| Probe | Hình dạng | Kết quả |
|---|---|---|
| `i6` | đọc **FIELD** f64 trong struct **generic** | 41 ✅ đúng |
| `i7` | method f64 **không generic** | 41 ✅ đúng |
| `i9` | struct generic, method trả **`-> f64` CỤ THỂ** (không phải `-> T`) | 42 ✅ đúng |
| `i5` | `Box[i32].get` — cùng hình dạng `-> T`, T ↦ **i32** | 42 ✅ đúng |

Đọc bảng này theo đúng nghĩa của nó: hỏng **không** phải vì generic (i6, i9 đúng), **không** phải vì
f64 (i7 đúng), **không** phải vì hình dạng `-> T` (i5 đúng). Chỉ hỏng ở **giao** của cả ba:
**`-> T` được mono hoá sang f64**.

## Giả thuyết (CHƯA xác minh — đừng chép thành kết luận)
Lớp **thanh ghi trả về**: giá trị f64 về XMM0 nhưng call site đọc RAX (hoặc ngược lại) vì kiểu trả
về của instance mono hoá vẫn mang lớp của **tham số kiểu**, không phải của **kiểu đã thay**. Cùng họ
cơ chế với `d1c_eq_f64_exact` (operator method trả f64 bị đọc như thanh ghi nguyên) — nếu đúng thì
**một** fix có thể đóng cả hai, nhưng phải chứng minh bằng **disassembly**, không bằng độ hợp lý.
Tiền lệ bắt buộc đọc trước: bug #4 đã bị chẩn đoán nhầm ba lần là "thanh ghi float cũ" trong khi thủ
phạm là chọn nhầm hàm ([[BACKLOG]] bug #4) — **hãy xem objdump trước khi tin bất kỳ câu chuyện nào
về thanh ghi.**

## Khi sửa
- Xác minh bằng objdump rằng call site và callee bất đồng về lớp thanh ghi, TRƯỚC khi sửa.
- Oracle phải phủ cả **i32/i64/str/struct** cho `-> T` (chống over-reach), và cả f32 — `i8` mới chỉ
  đo f64. **f32 chưa được đo**; đừng giả định nó đúng.
- Chạm mono hoá / lớp thanh ghi trả về ⇒ nhiều khả năng **B==C**, không phải A==B.

## Liên quan
[[BACKLOG]] (mục "Phát hiện KỀ BÊN" của probe8) · [[bug-iface-conformance-unchecked-sites]] ·
`bin/probe8/` (runner `run8.sh`, ma trận `matrix8.sh`).
