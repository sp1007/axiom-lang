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

## ⭐⭐ CƠ CHẾ — **KHÔNG phải "backend intercept", mà là OVERLOAD RESOLUTION thua stdlib bundled**
Các tên này **CÓ THẬT** dưới dạng **hàm tự do** trong stdlib được bundle:
| tên | khai báo | số param |
|---|---|---|
| `get` | `std/collections.ax:42` `get[T](self: Vec[T], index: i64)`; `:375` HashMap; `std/sync.ax:71` `get[T](self: MutexGuard[T]) -> T` | có bản **1 param** |
| `len` | `std/string.ax:11` `len(s: str) -> i64`; `collections.ax:290/498` HashMap/HashSet | **1 param** |
| `find` | `std/collections.ax:131` `find[T](self: Vec[T], pred)` | 2 |
| `push` | `std/collections.ax:21` `push[T](mut self: Vec[T], item: T)` | 2 |
| `set` | `std/mem.ax:26` `set(dst: ptr[u8], value: u8, n: i64)` | **3** |

⇒ **Giải thích KHỚP toàn bộ bảng kết quả**:
- `set` an toàn vì bản stdlib cần **3 tham số** ⇒ lời gọi 1 tham số không thể khớp ⇒ hàm user thắng.
- `len` trả **8** vì khớp `len(s: str)` — đọc "độ dài" từ word thứ hai của giá trị interface.
- `get` trả **0** vì khớp một bản `get` 1-tham-số (vd `get[T](self: MutexGuard[T])`).

⚠️ **NHƯNG chỉ INTERFACE param mới thua** — đã đo: `fn get(p: Sq)` (struct thường) → **36 ĐÚNG**,
`fn get(s: str)` → **36 ĐÚNG**, `fn get(x: i64)` → **42 ĐÚNG**. Các overload stdlib tồn tại y hệt
trong cả ba ca đó mà hàm user vẫn thắng. ⇒ **Giá trị kiểu INTERFACE khớp lỏng với tham số generic
(`T` / `Vec[T]` / `MutexGuard[T]`) theo cách struct cụ thể KHÔNG khớp** — đó mới là điểm gãy.
⇒ Chỗ cần sửa: **luật chọn overload** khi ứng viên là hàm stdlib GENERIC và đối số là giá trị
interface — hàm user khai báo TRỰC TIẾP phải thắng (hoặc báo lỗi rõ ràng).

## 📍 ĐÃ TÌM RA SITE — `typecheck.ax:1241` `free_call_gate` (+ tie-break HOLE#5 quanh L1154–1240)
**Cơ chế đã có sẵn để chữa ĐÚNG họ này** — đọc comment tại `typecheck.ax:1154–1168` và
`:1234–1240`: chúng mô tả nguyên văn "user free fn bị **shadow** bởi bundled stdlib overload
⇒ **SILENTLY MISCOMPILED**", và đã vá 2 lỗ trước (`HOLE#5` tie-break 2 pass, `HOLE#6` float
widening), có tham chiếu [[bug-freefn-stdlib-collision-noarg]].

`free_call_gate(ci, arg0_type)` nhận một ứng viên khi **bất kỳ** điều nào đúng:
`arg0_type` unknown / **param đầu là GENERIC trần** (`typecheck.ax:1253`) / widening int↔int /
`is_method_compatible` khớp.

⇒ **Giả thuyết cần probe (ĐỪNG suy luận — hãy in ra tập ứng viên)**: với `fn get(sh: Shape)`,
ứng viên stdlib **1 tham số** `get[T](self: MutexGuard[T])` (`std/sync.ax:71`) được **NHẬN** qua
nhánh generic (`ci_first_generic`), rồi **tie-break chọn nó thay vì hàm user khớp CHÍNH XÁC**.
Đây sẽ là **HOLE#7** cùng họ: *giá trị INTERFACE* lọt qua gate generic theo cách struct cụ thể
không lọt (đã đo: `fn get(p: Sq)` → 36 ĐÚNG).

### 🔎 ĐỌC KỸ `resolve_free_call_overload` (`:1294`) — giả thuyết SẮC HƠN (⚠️ mới ĐỌC CODE, **CHƯA probe**)
Cấu trúc thực tế:
```
pass 1: đếm ứng viên có ĐÚNG arity VÀ qua free_call_gate  -> n_gated, first_gated (HEAD-FIRST)
if n_gated >= 1 and (arg0 unknown or n_gated == 1): return first_gated   // <-- KHÔNG hề tie-break
if n_gated >= 2: chấm điểm toàn bộ param list (HOLE#5)                   // exact > coercion
```
⭐ **stdlib được nối vào TRƯỚC nên overload stdlib chính là HEAD của chuỗi** (comment `:1300–1303`
nói rõ). ⇒ Nếu **hàm USER TRƯỢT gate** còn bản stdlib generic **QUA gate**, thì `n_gated == 1`
⇒ **return thẳng bản stdlib, KHÔNG bao giờ tới tie-break HOLE#5.**

⇒ **Giả thuyết sắc nhất: không phải "tie-break chọn sai", mà là "ứng viên của USER không được
NHẬN"** — nhiều khả năng `is_method_compatible(param=Shape, arg0=Shape)` **thất bại với kiểu
INTERFACE**, nên `free_call_gate` loại chính hàm người dùng.
🔬 **Probe để xác nhận (đừng suy luận tiếp)**: in trong `resolve_free_call_overload` mỗi ứng viên
+ `sym_decl_param_count` + kết quả `free_call_gate` + `n_gated`, khi biên dịch
`bin/t_ifacefnbuiltinname.ax`. Nếu thấy `n_gated == 1` và ứng viên đó là stdlib ⇒ giả thuyết ĐÚNG,
và chỗ sửa là **`is_method_compatible` / gate với arg kiểu interface**, KHÔNG phải khối HOLE#5.
(Đúng thành ngữ probe đã dùng cho peephole 1d phiên này: dump thứ pass THỰC SỰ nhìn thấy.)

**Cách sửa có khả năng đúng nhất (NẾU giả thuyết trên sai và thật sự là tie-break)**: một ứng viên khớp **CHÍNH XÁC/cấu trúc**
(hàm user khai báo trực tiếp) phải **THẮNG** ứng viên chỉ khớp nhờ **param đầu generic trần** —
đúng tinh thần HOLE#5 nhưng hiện chưa phủ arg kiểu interface.
⚠️ Oracle bắt buộc **HAI CHIỀU**: (1) `fn get(sh: Shape)` phải được gọi; (2) **guard chống
over-fix**: `v.get(0)` trên `Vec` thật, `s.len` trên `str`, `zip[A,B,C]`, và ca HOLE#6 float
widening **phải giữ nguyên** hành vi. Chạy cả `-O0`.

## Gợi ý điều tra cũ (đã thu hẹp MỘT bước)
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
