---
name: bug-user-fn-builtin-name-iface-param
description: "OPEN BUG (probe-found 2026-07-30) — hàm user trùng tên với method BUILTIN-INTERCEPTED (get/len/find/push) VÀ có tham số kiểu INTERFACE thì lời gọi bị chặn thành builtin ⇒ trả về SAI im lặng (get→0, len→8=size, find→0, push→build fail). Đổi tên (zqarea/set) thì đúng. Cùng họ collision đã biết."
metadata:
  type: project
---

# 🐞 OPEN — hàm user trùng tên builtin + tham số INTERFACE ⇒ gọi nhầm builtin, sai IM LẶNG

**Tình trạng: OPEN, chưa sửa.** Probe-found 2026-07-30, driver `BF00A447`.

## Reproducer (`bin/t_ifacefnbuiltinname.ax`)
```axiom
interface Shape:
    fn area(self: Self) -> i64

struct Sq:
    s: i64
    fn area(self: Sq) -> i64:
        return self.s * self.s

fn get(sh: Shape) -> i64:      // <-- CHỈ CẦN ĐỔI TÊN THÀNH zqarea LÀ ĐÚNG
    return sh.area()

fn main() -> i64:
    return get(Sq(s: 6))
```
→ **0** ở `-O0/-O1/-O3` (đúng phải là **36**).

## Phạm vi — đã đo, cùng một chương trình chỉ đổi TÊN HÀM
| tên hàm user (tham số `sh: Shape`) | kết quả | |
|---|---|---|
| `get` | **0** | ❌ |
| `len` | **8** | ❌ — **8 = SIZE của interface value**, dấu hiệu rõ nhất |
| `find` | **0** | ❌ |
| `push` | **BUILD FAIL** | ❌ |
| `set` | 36 | ✅ |
| `zqarea` | 36 | ✅ |

⭐ **`len` trả về 8 là bằng chứng cơ chế**: lời gọi bị **intercept thành BUILTIN** trên giá trị
interface (lấy size/len) thay vì dispatch tới hàm user. Không phải "hàm user chạy sai" mà là
**hàm user KHÔNG HỀ ĐƯỢC GỌI**.

## Trigger là TỔ HỢP, không phải riêng tên
| dạng | kết quả |
|---|---|
| `fn get(x: i64) -> i64` (tham số **scalar**) | ✅ 42 — **KHÔNG hỏng** |
| `fn get(sh: Shape) -> i64` (tham số **interface**) | ❌ 0 |
Đã thử 8 tên builtin với tham số scalar (`get/len/push/set/find/count/insert/remove`) — **tất cả
ĐÚNG**. ⇒ Chỉ hỏng khi tham số là **INTERFACE**.

⚠️ Cũng KHÔNG phải lỗi của interface dispatch: `let sh: Shape = sq` rồi `sh.area()` trực tiếp
**ĐÚNG (36)**; `t_ifacedispatch` trong suite vẫn xanh (37). Chỉ hỏng ở **lời gọi hàm user có
tham số interface, khi tên hàm trùng builtin**.

## Vì sao suite không thấy
`t_ifacedispatch` đặt tên hàm là `total`, `t_ifaceconsumer`/`t_ifacevecpoly`… đều dùng tên không
trùng builtin. **Không có test nào** đặt tên hàm nhận interface trùng với method builtin ⇒ vùng
chưa có coverage, không phải regression.

## Bối cảnh: đây là HỌ COLLISION đã biết, thêm một lỗ mới
Xem [[task-cross-library-name-collision]] và [[bug-user-fn-stdlib-struct-name-collision]].
Ghi chú cũ nói **"Vec/HashMap thoát vì builtin-intercepted"** — chính **builtin interception** là
cơ chế gây ra ca này. Các lỗ trước đã xử: fn-vs-fn (mangling flag 2048), fn-vs-struct (reject ở
typecheck). Lỗ NÀY: **user fn vs BUILTIN method**, và nó không reject cũng không mangle — nó
**im lặng gọi builtin**.

## Gợi ý điều tra (đã thu hẹp MỘT bước, chưa tìm ra site)
⚠️ **ĐÃ LOẠI TRỪ**: `typecheck.ax:4342` (comment có nhắc "backend-intercepted") **KHÔNG phải**
chỗ này — nó là kiểm ARITY của **method call** `recv.m(...)`. Bug này là **lời gọi HÀM TỰ DO**
`get(Sq(s:6))` bị chặn, nên phải tìm ở đường lowering **CALL_EXPR hàm tự do** (nhiều khả năng
trong `air_builder`), nơi một call theo TÊN bị biến thành builtin op.
Tìm chỗ intercept builtin method (`get/len/push/find`) — nhiều khả năng nhận diện theo **TÊN**
trên receiver có kiểu "container-ish", và giá trị **interface** lọt vào diện đó. Cần: khi call
site resolve ra một **hàm user do người dùng khai báo**, hàm đó phải **THẮNG** builtin intercept
(hoặc reject rõ ràng như hole fn-vs-struct đã làm).
⚠️ Khi sửa phải có oracle cho **cả hai chiều**: ca user-fn phải được gọi, VÀ ca guard chống
over-reject (`v.get(0)` trên Vec thật vẫn phải là builtin). Chạy cả `-O0`.
