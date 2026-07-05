# RFC 0012 — Biểu diễn nhất quán cho Option/Result (mở khóa `match` native)

- **Status:** ✅ SHIPPED (2026-07-05) — **backend-only**, KHÔNG đụng type-checker/ABI.
  Fixpoint B==C bit-identical; 24 chương trình regression + 4 test Option/Result xanh.

## SHIPPED — thiết kế backend-only (khác kế hoạch Part A ban đầu)

Kế hoạch (A) "gán Option/Result type kind 11/12 xuyên type-checker" đã thử và
**REVERT**: type `Option[T]` ở annotation (param/return) làm hỏng self-host — các
method generic trong std/result.ax (`is_some[T](self: Option[T])`, `unwrap[T]`…)
bị đổi biểu diễn param/mono → B (compiler build bằng A) segfault khi compile. Bisect
xác nhận **chỉ riêng thay đổi type-checker** (NODE_GENERIC_TYPE + call-typing) đủ để
vỡ, bất kể size entry 8 hay 16.

**Cách đã ship — toàn bộ nằm trong backend, KHÔNG sửa typecheck.ax:**
1. `typetable.ax`: thêm `register_option(inner)` / `register_result(ok,err)` (kind
   11/12, **size 16 như sum kind-6**, aggregate → value là con trỏ box 8B, copy như
   sum). Chỉ air_builder gọi — type-checker không bao giờ tạo entry này.
2. `air_builder.ax` constructor (Some/Ok/Err): tính `box_ty` **cục bộ** bằng
   register_option/result(payload_t) và dùng cho `OP_ALLOC` → box_reg mang type
   kind-11/12 (KHÔNG phải str-12) ⇒ EQ/COPY/MAKE_REF hết đi đường str-deref. None =
   ICONST i64 0. Err = box|1.
3. `air_builder.ax` `match_arms_tagged_kind`: phân loại match theo **TÊN ARM**
   (Some/None ⇒ Option, Ok/Err ⇒ Result) — không phụ thuộc type tĩnh của scrutinee
   (vốn có thể là std-sum kind-6 hoặc untyped). Giữ off-ABI.
4. `air_builder.ax` `lower_match_tagged`: dispatch pointer-tagged (None `==0`, Some
   `!=0`, Ok `(x&1)==0`, Err `(x&1)!=0`; payload = LOAD[box&~1], width lấy từ symbol
   binding, fallback i64). scrut_reg untyped (function-return) ⇒ EQ số nguyên; box
   kind-11/12 ⇒ cũng số nguyên. Cả hai đều đúng.

Vì các giá trị Option/Result kind-11/12 chỉ xuất hiện **cục bộ** (box constructor +
match), chúng luân chuyển y hệt user-sum kind-6 (đã self-host ổn từ lâu) → an toàn.

---

- **Status(cũ):** Draft (2026-07-05) — chờ implement; root cause đã xác định đầy đủ bằng dump-air + objdump
- **Author:** self-host team
- **Tracking:** BUG#57 (native `match` trên Option/Result sinh 0 dispatch), liên quan BUG#59
- **Liên quan:** typecheck.ax (NODE_GENERIC_TYPE / infer_node, không gán type cho Option/Result),
  air_builder.ax (constructor Ok/Err/Some/None ~L1356 + `lower_match`/`lower_match_tagged`),
  x86_selector.ax (OP_EQ/NE str-path, OP_MAKE_REF/OP_COPY/OP_CAST/OP_LOAD trên giá trị by-address,
  `type_is_aggregate`, `regalloc_is_16byte`), std/result.ax (đọc box qua `(&self as ptr[u64])[0]`).
- **Blocks:** mọi user code dùng `match` trên Option/Result; `.unwrap()`/`?`-ergonomics tin cậy;
  gỡ bỏ SKIP của tests/generics/match_option_BUG57.ax.

---

## 1. Motivation — `match` trên Option/Result im lặng trả rác

```axiom
fn main() -> i32:
    match Some(42):
        Some(v): return v      // kỳ vọng 42
        None:    return 0
// native: trả 0 (rơi qua CẢ HAI arm). C-emit: undefined 'Some' nếu không có prelude.
```

`match` trên Option/Result **không sinh dispatch nào** trên native backend. Đây KHÔNG
phải lỗi cục bộ trong `lower_match` — nó là hệ quả của **hai khiếm khuyết biểu diễn**
xác định bằng dump-air + objdump (2026-07-05):

### 1a. Scrutinee Option/Result là UNTYPED

`node_types[scrutinee] == 0` cho **mọi** dạng: `match make()`, `match Some(42)`,
`let x: Option[i32] = Some(7); match x`, `let x = make(); match x`. Lý do: không có
`register_option`/`register_result`; `Option[i32]` với type-arg cụ thể rơi khỏi nhánh
generic-template trong `infer_node`/`NODE_GENERIC_TYPE` vì symbol builtin `Option`
không mang cờ generic → `result_type` giữ nguyên `TYPE_UNKNOWN`. ⇒ Test kind trong
`lower_match` (kind 11/12) **không bao giờ chạy**; phải phân loại theo TÊN arm
(Some/None ⇒ Option, Ok/Err ⇒ Result) — mảnh `match_arms_tagged_kind` từ lần thử
trước tái dùng được.

### 1b. Box biểu diễn KHÔNG nhất quán, thù địch với dispatch sạch

| Variant | Biểu diễn hiện tại |
|---|---|
| `None`        | `OP_ICONST 0` — một **i64 = 0** |
| `Some(x)`/`Ok(x)` | `OP_ALLOC type_id=12 (str, 16B)` → box heap; `Some/Ok` = con trỏ box |
| `Err(x)`      | như trên nhưng OR bit-0 (`box \| 1`) |
| payload       | ghi tại `[box+0]` (str payload 16B, scalar 8B) |

Box được **gán kiểu `str` (type_id 12, 16 byte)**. Giá trị kiểu-str được giữ
**BY-ADDRESS**: thanh ghi chứa con trỏ heap, còn "giá trị" là các byte TẠI con trỏ đó.
Hệ quả (đã phá mọi cách vá chỉ-ở-lowering):

- `OP_EQ`/`OP_NE` trên reg kiểu-str đi đường **so-sánh-chuỗi** (`ax_str_eq`, call -22),
  không phải so sánh con trỏ/số nguyên.
- `OP_COPY`/`OP_CAST`/`OP_MAKE_REF`/`OP_LOAD` trên box đều **DEREF** ra `[box+0]`
  (payload) thay vì trả con trỏ box thô cần cho test null/low-bit. objdump cho thấy
  toàn bộ "đọc box" gộp thành một lệnh `mov (%rsi),%rbx` (= load payload) dù dùng op nào.
- `std/result.ax` `(&self as ptr[u64])[0]` chạy đúng **CHỈ vì** `self` là **tham số**:
  param-passing chuẩn hóa giá trị str vào home-slot chứa con trỏ 8B, nên `&self`=địa chỉ
  slot và `[slot]`=con trỏ. `match` scrutinee (call-result / local / constructor trực
  tiếp) không được chuẩn hóa đó → `&scrut` deref ra payload.
- Net: `match Some(42){Some=>77,None=>99}` trả 0 — discriminant so **payload (42)** với
  0/1, không phải con trỏ.

**Kết luận:** cần thay đổi type + representation (đúng như cảnh báo "dedicated pass"
cũ), KHÔNG phải chỉnh lowering. RFC bắt buộc theo CLAUDE.md §13 (IR/representation redesign).

## 2. Ngữ nghĩa đề xuất (normative)

`Option[T]` và `Result[T,E]` là **kiểu con-trỏ-cỡ-8-byte** xuyên suốt pipeline:
- Giá trị runtime = **con trỏ box đã-tag** (8B, held-by-value như scalar), KHÔNG phải
  cấu trúc 16B held-by-address.
- Box heap 16B vẫn tồn tại (chứa tag + payload) nhưng chỉ là vùng payload — "giá trị"
  là con trỏ 8B tới nó, y như biểu diễn hiện tại của user-sum kind 6 (BUG#56 đã xác nhận
  sum = con trỏ box 8B, `regalloc_is_16byte` phải loại kind 6/11/12).
- Tag: Option — `None`=null(0), `Some`=con trỏ non-null; Result — `Ok`=box chẵn (bit0=0),
  `Err`=`box|1`. Payload = `[box & ~1 + 0]`.

## 3. Design — hai phần độc lập, ưu tiên (A) trước

### (A) Type system: gán Option/Result một type entry THẬT, size 8, kind 11/12

- Thêm `register_option(inner)` / `register_result(ok, err)` vào typetable (dedup theo
  args như `register_generic_inst`), **size=8, align=8, kind=TYPE_KIND_OPTION/RESULT**.
- `infer_node`/`NODE_GENERIC_TYPE` (typecheck.ax ~L1906): nhận biết base-name
  `Option`/`Result` (mirror detection ở L2145) → gọi register_option/result thay vì rơi
  về UNKNOWN. Set `node_types[scrutinee]`, `let`-binding, và return-type của hàm.
- Constructor Ok/Err/Some/None: gán reg kết quả kind 11/12 (không phải str/12) để
  `regalloc_is_16byte`/`type_is_aggregate` xử lý như con-trỏ-8B (đã có nhánh loại kind
  6/11/12 từ BUG#56/OP_MAKE_REF L1715-1727) → EQ/NE đi đường số nguyên, COPY/LOAD
  không deref nhầm.

### (B) Lowering `match` (sau khi (A) cho scrut kind 11/12)

- `lower_match`: test kind 11/12 → `lower_match_tagged` (giữ `match_arms_tagged_kind`
  làm fallback nếu scrut vẫn untyped ở site nào đó).
- Dispatch (scrut giờ là con trỏ 8B kind 11/12, EQ/NE = số nguyên):
  - Option: `None` ⇒ `scrut == 0`; `Some` ⇒ `scrut != 0`; payload `v = LOAD[scrut]`.
  - Result: `Ok` ⇒ `(scrut & 1) == 0`; `Err` ⇒ `(scrut & 1) != 0`;
    payload `= LOAD[scrut]` (Ok) / `LOAD[scrut-1]` (Err).
  - payload type = type-arg từ register_option/result (không còn phải fallback i64).

## 4. Rủi ro & fixpoint

- **Đụng ABI/param-passing**: đổi kind của Option/Result thay đổi cách truyền/trả
  (str-16B → pointer-8B). `std/result.ax` đọc qua `ax_sum_layout_is_pointer()` +
  `(&self as ptr)` — phải kiểm tra layout vẫn khớp (self-host dùng Option/Result nặng).
- **Fixpoint BẮT BUỘC** (backend/ABI, theo [[feedback-fixpoint-async-rule]]): stage2==stage3
  VÀ stage3==stage4 bit-identical trước commit; regression bằng native binary.
- Nếu (A) làm vỡ self-host (result.ax layout), tách nhỏ: chỉ đổi representation ở
  constructor+match trước, giữ ABI cũ, đo bằng test cô lập rồi mới lan sang ABI.

## 5. Alternatives

- **(b) Chỉ đổi constructor**: giữ 16B alloc nhưng reg kết quả pointer-typed (8B). Nhẹ
  hơn (A) nhưng vẫn để scrut untyped → `match` vẫn phải phân loại theo tên; nửa vời.
- **Giữ nguyên + lowering hack**: đã CHỨNG MINH bất khả (mọi op deref; str-compare).
  Loại bỏ.

## 6. Migration

1. Landmine test đã có: `tests/generics/match_option_BUG57.ax` (SKIP). Gỡ SKIP khi xanh.
2. Thêm test Result match + payload-binding + nested + None/Err arms.
3. Cập nhật spec §Option/Result nêu rõ "pointer-tagged, size 8".
