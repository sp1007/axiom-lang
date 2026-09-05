---
name: probe8-scan-2026-07-31-live-holes
description: The probe8 scan after a281992 — NOT a clean sweep; records which name-resolution holes were fixed and which remain live, plus adjacent defects (iface conformance unchecked sites, operator return type unchecked, mono generic -> T as f64)
metadata:
  type: project
---

# probe8 (2026-07-31, sau `a281992`) — **KHÔNG phải quét sạch: 3 lỗ SỐNG**

> Tách khỏi `BACKLOG.md` ngày 2026-09-05 (§24: file con trỏ không giữ sự thật). Nguyên văn, không
> lược. ⭐ **Bài học bao trùm:** `a281992` làm tên CHÍNH XÁC **thắng điểm** cửa sổ lỏng, nhưng
> **không chặn cửa sổ lỏng LÀM CÂU TRẢ LỜI khi không có tên chính xác nào** — tức là **sửa một nửa
> của một luật**. Đây cùng lớp với "hai bản sao một luật, một bản không bao giờ được sửa" xuất hiện
> nhiều lần trong repo này.

## 🐞🐞 BUG MỞ — quét sau `a281992` (probe8, 2026-07-31). **KHÔNG phải quét sạch: 3 lỗ SỐNG.**
Tất cả đều đo được, có control đối chứng, ở `bin/probe8/` (runner `run8.sh`, ma trận `matrix8.sh`).
Bài học bao trùm: `a281992` làm tên CHÍNH XÁC **thắng điểm** cửa sổ lỏng, nhưng **không chặn cửa sổ
lỏng LÀM CÂU TRẢ LỜI khi không có tên chính xác nào**. Sửa một nửa của một luật.

**A+B+B′ — rank-1 "loose": ✅ ĐÃ SỬA `a538983`** (A==B `B10DABE6`, regression 662/662, breakage
audit 1215 file: 7 reject mới đúng như dự định, 2 accept mới, **0 collateral**). Đã xoá luôn **bốn**
bản sao chết của cùng luật. Ngoài ra **hoàn thiện nhánh SỐ HỌC của RFC 0007 §3.1** (§2.2 chỉ hoãn
tới khi có error infra — nay đã có): `a + b` trên struct không có `add` trước đây **SEGFAULT (139)**,
nay là lỗi biên dịch. Tự kiểm chứng độc lập trên compiler CŨ: accept + exit 139.
⚠️ **`h1_rank2_eq` KHÔNG được sửa và không thể sửa bằng việc bỏ rank 1** — nó là **rank 2**
(`Num.eq__fast` khớp `eq` + dấu phân cách type-arg `__`). Cùng hình dạng với `h2_axstd_prefix`:
**tên do user đặt bắt chước cách mangle của monomorphizer**. Xem "Over-reach còn lại" trong RFC 0037
— sửa đúng là khoá rank 2/3 vào việc symbol **THỰC SỰ là một instantiation**, không phải vào cách
viết; hạn chế rank 2 theo prefix `_AX_std_` **sẽ làm hỏng** method instantiated của struct generic
(mangle đặt prefix TRƯỚC dấu chấm). Đó là **quyết định thiết kế**, không phải siết cơ học.
Mô tả gốc — chọn method theo **substring chặn `_`**:
- `==`/`<`/`+` gọi hàm user chưa từng đặt tên: chỉ có `deep_eq` ⇒ `a == b` trên hai struct KHÁC nhau
  ra **true** (`a1_op_deepeq` = 7). Chẩn đoán RFC 0007 §2.2 **đã tồn tại** trong compiler và bị
  rank 1 **bịt miệng**. Cùng cơ chế: `total_lt`→`<`, `checked_add`→`+`, `eq__fast` (rank 2).
- **Drop glue TỰ CHẾ lời gọi**: kiểu chỉ khai báo `pre_drop(self)` bị gọi làm drop glue ở MỌI lối ra
  scope (`a2` 5 lời gọi; `d4` 3 lời gọi vào method trả `ptr[i64]`) ⇒ **UAF đang chờ xảy ra**.
  Quy trách nhiệm chính xác: 42 dưới `-no-ctgc-free`, 7 mặc định ⇒ `resolve_drop_method`+`lower_destroy`.
  RFC 0014 định nghĩa hook là `Type.drop(self)`; `pre_drop` **không phải** `drop`.
- **`ownership.ax:138,162` CHƯA TỪNG được chuyển** sang rank mới ⇒ `type_has_drop` true cho
  `pre_drop` ⇒ `let b = a` bị **E4003 oan** (`a3`; control `a3c` dùng `cleanup` = 42). **Đúng hình
  dạng "hai bản sao một luật, một bản không bao giờ được sửa".**
- Bằng chứng an toàn: commit `a281992` ghi rằng bản compiler **bỏ hẳn rank 1** qua **619/619 và tự
  dựng lại byte-identical** (stdlib đều gọi bằng tên đầy đủ = rank 3). Còn `air_builder.ax:1571`
  `match_base_names` và `:1755` `match_mangled_method_name` = **bản sao CHẾT, 0 caller** ⇒ xóa.

**C — overload cùng tên trên receiver struct đè nhau ở symbol: ✅ ĐÃ SỬA 2026-07-31** cho dạng gọi
**TỰ DO** (A==B==C `52D1ABD4`, RFC 0035 §2bis, oracle `bin/t_structoverload{,dfe}.ax`).
⚠️ **Nửa còn lại VẪN MỞ và là bug KHÁC:** gọi bằng **cú pháp method** (`s.f(41)`, method inline,
`S.f(&s,41)`) bind theo RECEIVER, **bỏ qua arity** ⇒ [[bug-method-call-overload-ignores-arity]]
(`resolve_method_overload` không có arity, `resolve_free_call_overload` thì có — lại là **một luật,
hai bản sao**). Quy trách nhiệm chắc: cùng khai báo, **một** compiler, tự do ra 42 / method ra 2.
Mô tả gốc: `typecheck.ax:1233-1246` `free_fn_bare_mangles` trả false khi param 0 là struct/sum, tin rằng
`x86_regs.ax:338` mangle ra `ax_<Struct>_<fn>` "duy nhất theo receiver" — nhưng đó là duy nhất theo
**(receiver, tên)**, KHÔNG theo chữ ký. Vòng uniquing Phase-3.5 (`typecheck.ax:3227-3240`) không gắn
`MODDUP` cho overload thứ hai ⇒ **mọi lời gọi bind vào body khai báo TRƯỚC**.
⚠️ **Phân kỳ theo mức tối ưu**: `g10_seq` = **2 ở -O0/default, 42 ở -O1** (deterministic 5/5, dựng
lại 2 lần, binary byte-identical; default ≡ -O0 vì `optimize` mặc định false ở `main_air.ax:854/943`).
**-O1 chỉ CHE bằng inlining.** `g11_revboth` (đảo thứ tự khai báo) = 82. `g5_only2` = 22 nhưng **11
dưới `-no-dfe`** ⇒ các ca "đúng" chỉ là **ảo giác do DFE** xóa hàm bị che. Receiver i64/str (`g12`,
`g13`) = 42 vì chúng CÓ bare-mangle nên Phase 3.5 unique được.

**Phát hiện KỀ BÊN (không phải luật tên) — mỗi cái là một item riêng:**
- `check_iface_conformance` chỉ được gọi ở **3 site** (`typecheck.ax:764` field init, `:4193` let
  init, `:4208` return) + check param inline `:5439`. **Gán (assignment)** và **payload Option**
  KHÔNG được phủ ⇒ `f1_iface_optpayload` **SEGFAULT (139)** ở cả ba mức, `f5_iface_assign` = 7.
  Control `f1d`/`f5d` (struct không có method na ná) fail y hệt ⇒ gốc là **thiếu check**, không phải
  luật tên. → [[bug-iface-conformance-unchecked-sites]].
- **Kiểu TRẢ VỀ của operator method không được kiểm**: `eq(self,o) -> f64` đặt tên CHÍNH XÁC vẫn làm
  `a == b` ra true (đọc thanh ghi nguyên trong khi callee trả XMM0) — `d1c_eq_f64_exact`.
- **Method generic mono hoá trả `-> T` với T ↦ f64 ra 0.0**: `i8_genf64ret` = 40 ở cả ba mức.
  Controls thu hẹp rất gọn: `i6` đọc **field** f64 generic = 41 ✅, `i7` method f64 không generic
  = 41 ✅, `i9` struct generic có `-> f64` **cụ thể** = 42 ✅, `i5` `Box[i32].get` = 42 ✅ ⇒ khuyết tật
  đúng ở **`-> T` mono hoá sang f64**, nghi lớp thanh ghi trả về (**chưa xác minh** — xem objdump
  trước khi tin). → [[bug-mono-generic-ret-typaram-f64]].
- `h2_axstd_prefix` = 7 (mà `h3` đảo thứ tự = 42): phép **strip `_AX_std_`** biến một
  `_AX_std_eq` do user định nghĩa thành **tie rank-3** cướp `==` khỏi `eq` thật — over-reach do
  **chính `a281992`** đẻ ra, nằm ở luật phá hoà, không phải ở cửa sổ lỏng.
- `resolver.ax:653/662`: biên `.` đúng, nhưng **việc CHỌN** giữa nhiều `M.foo` là tuỳ tiện (theo thứ
  tự slot hash) — chưa bị chạm tới, ghi lại để không ai tưởng đã kiểm.

