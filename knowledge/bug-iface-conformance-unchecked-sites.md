---
name: bug-iface-conformance-unchecked-sites
description: check_iface_conformance chạy ở 3/5 site — gán và payload Option không được kiểm, làm SEGFAULT (139) ở mọi mức tối ưu
metadata:
  type: project
---

# BUG MỞ — `check_iface_conformance` chỉ chạy ở 3 trong 5 site ⇒ SEGFAULT

**Trạng thái:** OPEN (phát hiện 2026-07-31, probe8). **Mức nặng nhất trong đợt quét**: đây là ca duy
nhất **segfault**, không phải trả sai số.

## Triệu chứng đo được
| Probe | Nội dung | Kết quả | Đúng phải là |
|---|---|---|---|
| `bin/probe8/f1_iface_optpayload.ax` | struct KHÔNG thoả interface, nhét vào **payload của Option** | **139 (SEGFAULT)** ở `-O0`/default/`-O1` | lỗi biên dịch |
| `bin/probe8/f5_iface_assign.ax` | struct KHÔNG thoả interface, đưa vào biến iface bằng **phép GÁN** | **7** ở cả ba mức | lỗi biên dịch |
| `bin/probe8/f1c_…_ok.ax`, `f5c_…_ok.ax` | bản THOẢ interface, cùng hình dạng | **42** ✅ | 42 |
| `bin/probe8/f1d`, `f5d` | struct **không có method nào na ná** | **139 / 127** — fail y hệt | — |

⭐ **Vai trò của `f1d`/`f5d`**: chúng chứng minh gốc rễ là **THIẾU CHECK**, không phải luật khớp tên
lỏng (rank-1). Nếu là luật tên thì struct không có method na ná đã phải fail KHÁC đi. Đây là cách
tách hai nguyên nhân chồng lên nhau — đừng gộp bug này vào họ rank-1.

## Vị trí
`check_iface_conformance` được gọi ở đúng **3** chỗ:
- `typecheck.ax:764` — khởi tạo **field**
- `typecheck.ax:4193` — khởi tạo **let**
- `typecheck.ax:4208` — **return**

cộng một check param **inline** ở `typecheck.ax:5439`.

**Không phủ:** (1) phép **gán** vào một binding kiểu interface đã tồn tại; (2) **payload của Option**
(và nhiều khả năng payload ADT/tuple/phần tử mảng — CHƯA đo, đừng đọc im lặng thành bằng chứng).

## Vì sao đây đúng hình dạng defect quen thuộc của repo
Lại là **một luật tồn tại ở dạng HẸP HƠN thứ nó phải phủ** — y hệt `coerce_to_float_target` (RFC 0006
§7.2/§7.3) và cửa sổ rank-1 ([[BACKLOG]] hole A/B). Bản sửa đúng vì thế **không** phải thêm lời gọi
thứ tư rời rạc, mà là làm cho mọi value site đi qua **một** điểm — cùng hình dạng chống-trôi mà
`coerce_struct_to_interface` đang có.

## Cần lưu ý khi sửa
- Coercion struct→interface đã có sẵn (`coerce_struct_to_interface`); câu hỏi là **danh sách site**,
  không phải cơ chế. Liệt kê site theo đúng cách RFC 0006 §7.2 đã liệt kê (let/assign/element/field/
  return/argument/global init) rồi đối chiếu từng cái.
- Fix ở typecheck ⇒ **A==B**. Nhưng nếu chạm cả coercion ở air_builder thì **B==C**.
- Phải có control chống **over-reject**: struct THOẢ interface ở đúng các site đó (`f1c`, `f5c`) và
  qua generic/Option lồng nhau.

## Liên quan
- [[BACKLOG]] — mục "Phát hiện KỀ BÊN" của đợt probe8.
- Kiểu **TRẢ VỀ** của operator method cũng không được kiểm (`d1c_eq_f64_exact`): `eq(self,o) -> f64`
  đặt tên CHÍNH XÁC vẫn làm `a == b` ra true vì đọc thanh ghi nguyên trong khi callee trả XMM0.
  Cùng họ "conformance kiểm thiếu", khác site.
