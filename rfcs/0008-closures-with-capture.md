# RFC 0008 — Closures with captured environment

- **Status:** Draft (2026-07-03)
- **Author:** self-host team
- **Tracking:** #38.7 (closures) / follows BUG#49 (bare function pointers)
- **Liên quan:** air_builder (lower_ident SYM_FUNC / OP_FUNC_ADDR), x86_selector
  (MACH_CALL_INDIRECT), escape.ax, ownership.ax, ctgc.ax, connection_graph.ax
- **Blocks:** generic higher-order với TRẠNG THÁI (partial application, callback
  mang context, iterator adaptor bắt biến), std.iter đầy đủ.

---

## 1. Motivation

BUG#49 đã cho **con trỏ hàm trần** (`let f = add`, `fn(f64)->f64` param). Đủ cho
quadrature/root-find/sort-comparator vì hàm là *thuần* (không bắt biến). Nhưng
rất nhiều pattern cần hàm MANG TRẠNG THÁI bắt từ scope bao quanh:

```
let k = 3.0
let scaled = |x| x * k          // bắt k
integrate(scaled, 0.0, 1.0, 1000)

fn make_adder(n: i64) -> fn(i64)->i64:
    return |x| x + n             // bắt n, thoát khỏi frame
```

Con trỏ hàm trần KHÔNG biểu diễn được: `scaled` cần mang `k` theo. Đây là closure
= **con trỏ code + môi trường bắt được (captured environment)**.

Đây là feature ngôn ngữ + ABI + tương tác ownership → theo CLAUDE.md phải có RFC
trước khi implement.

---

## 2. Thiết kế — không gian lựa chọn

### 2.1 Biểu diễn closure (2 phương án)

**A. Fat pointer thống nhất (mọi `fn(...)` = {code, env}).**
Mỗi giá trị hàm là 16-byte `{code_ptr: i64, env_ptr: i64}`. Hàm trần: `env_ptr = 0`.
Gọi: luôn nạp env vào thanh ghi arg ẩn thứ 0 rồi `call code_ptr`.
- Ưu: một kiểu duy nhất, fn-ptr trần và closure hoán đổi tự do.
- Nhược: **đổi ABI của MỌI con trỏ hàm hiện có** (BUG#49 dùng con trỏ 8-byte trần).
  Hàm trần phải nhận thêm 1 tham số env ẩn (bỏ qua) → hoặc cần thunk.

**B. Kiểu closure RIÊNG (`Fn[...]`), fn-ptr trần giữ nguyên 8-byte.**
Con trỏ hàm trần (BUG#49) không đổi. Closure là kiểu mới, biểu diễn 16-byte
`{code_ptr, env_ptr}`. `code_ptr` của closure LUÔN có chữ ký `(env, args...)`.
- Ưu: **không đụng ABI BUG#49** (self-host an toàn hơn); tách bạch rõ.
- Nhược: hai kiểu (`fn(...)` trần và closure) → cần quy tắc ép/hội khi API nhận
  một trong hai. MVP: API nhận closure; hàm trần tự bọc thành closure env=0.

→ **Khuyến nghị: Phương án B** cho MVP. Ít blast-radius nhất lên phần đã fixpoint
(BUG#49). Có thể tiến hoá sang A sau nếu muốn thống nhất.

### 2.2 Layout môi trường

`env` = struct sinh tự động chứa **các biến bắt được**, cấp phát trên heap qua
`@alloc` (giống struct hiện tại = con trỏ heap 8-byte). Trình biên dịch:
1. Escape/lambda analysis liệt kê tập biến tự do (free vars) trong thân lambda.
2. Sinh `EnvN { v0: T0, v1: T1, ... }` theo thứ tự xác định (tất định).
3. Tại điểm tạo closure: `@alloc(sizeof EnvN)`, copy/ghi các biến bắt được, tạo
   `{code_ptr = &lambda_body, env_ptr = env}`.
4. Thân lambda được hạ thành một **hàm top-level** `lambda_body(env: ptr[EnvN],
   args...)`, đọc biến bắt được qua `env.vK`.

### 2.3 Ngữ nghĩa bắt biến (capture mode)

- **MVP: capture-by-value** (copy giá trị vào env tại thời điểm tạo closure). Đơn
  giản, tất định, an toàn với escape (giá trị sống trong env trên heap, không phụ
  thuộc frame gốc). Đủ cho `make_adder`, `scaled`.
- **Follow-up: capture-by-reference** (`&x`): env giữ con trỏ tới ô của biến. Cần
  escape analysis chứng minh biến bắt được sống lâu hơn closure, hoặc heap-promote
  ô đó. Hoãn — tương tác phức tạp với ownership.

### 2.4 Cú pháp

```
|params| expr                       // lambda biểu thức
|params| :                          // lambda khối (indent block)
    stmt
    ...
```
Kiểu: suy từ ngữ cảnh (tham số API) hoặc annotate `|x: f64| -> f64: ...`. MVP có
thể yêu cầu annotate để tránh đụng type-inference (giảm rủi ro).

---

## 3. Tương tác OWNERSHIP / CTGC / ESCAPE (mấu chốt của AXIOM)

Đây là phần khiến closure KHÁC hẳn ngôn ngữ có GC. AXIOM là single-owner + CTGC.

1. **env là một allocation có chủ.** Ai sở hữu env? → **closure sở hữu env.** Khi
   closure bị drop, drop luôn env (và drop-glue cho các trường có chủ trong env).
   Điều này CHƯA có: cần drop-glue cho struct-fields có chủ (xem [[bignum-ctgc-conflict]]
   — cùng lỗ hổng "owned-struct-fields/drop-glue").
2. **Capture-by-value một giá trị CÓ CHỦ** (vd một struct/vec sở hữu) = **move** vào
   env (chuyển quyền sở hữu), không copy con trỏ (nếu copy → nghịch single-owner,
   double-free). Connection graph + ownership.ax phải coi capture là move.
3. **Closure thoát khỏi frame** (`return |x| x+n`) → env PHẢI trên heap (đã vậy) và
   escape analysis phải đánh dấu closure + env là escaping (không stack-alloc, không
   alias-reuse nhầm).
4. **Copy-by-value một closure** = copy fat-pointer {code, env}: chia sẻ env → NGHỊCH
   single-owner (hai chủ cùng env). MVP: **cấm copy closure** (move-only), hoặc chỉ
   cho phép closure không-sở-hữu-gì-có-chủ (env toàn giá trị POD) được copy an toàn.

→ **Kết luận:** MVP nên GIỚI HẠN capture ở **giá trị POD/số** (i64/f64/bool/ptr thô,
không phải kiểu có chủ). Khi đó env không cần drop-glue có chủ, closure move-only là
đủ, và ta né hoàn toàn lỗ hổng drop-glue chưa giải. Closure bắt kiểu-có-chủ = follow-up
sau khi RFC drop-glue (owned-struct-fields) land.

---

## 4. ABI / calling convention

- Closure value = 16-byte `{code_ptr, env_ptr}`, truyền **bằng con trỏ** như mọi
  aggregate 16-byte của AXIOM (RFC 0001), nhất quán.
- Gọi closure `c(a, b)`:
  1. nạp `c.env_ptr` vào thanh ghi arg 0 (RCX/RDI theo nền tảng),
  2. marshaling `a, b` vào arg 1..n (dịch mọi arg sang phải 1 chỗ),
  3. nạp `c.code_ptr` → `MACH_CALL_INDIRECT` (tái dùng hạ tầng BUG#49: bắt target
     vào R11 TRƯỚC marshaling, gọi qua R11).
- `code_ptr` là địa chỉ hàm top-level `lambda_body` → RIP-relative/RELOC_PC32
  (ASLR-safe, đã có từ BUG#49).

Điểm khác BUG#49 duy nhất: **chèn env làm arg ẩn 0** + call target lấy từ trường
`code_ptr` của một aggregate (thay vì OP_FUNC_ADDR trực tiếp).

---

## 5. Tại sao an toàn cho self-host (fixpoint)

- Phương án B **không đổi** con đường con trỏ hàm trần đã fixpoint (BUG#49): mã
  compiler hiện tại không dùng closure → đường mới chỉ kích hoạt khi có `|...|`.
- Env struct + `@alloc` + gọi gián tiếp đều là cơ chế đã tồn tại (struct heap,
  MACH_CALL_INDIRECT). Không opcode máy mới bắt buộc (có thể tái dùng OP_CALL
  indirect + một OP dựng closure hoặc lower thẳng thành alloc+store+struct).
- Bắt buộc chạy `scripts/regression_repros.sh` + `verify_bug29_selfhost.sh` sau khi
  implement (đổi air_builder/backend). Library-only phần std.iter theo sau = không
  cần fixpoint.

---

## 6. Lộ trình implement (phân pha)

- **P0 (RFC này):** chốt biểu diễn (B: kiểu closure riêng), capture-by-value POD-only,
  move-only, cú pháp `|params| expr`.
- **P1 — front-end:** parser lambda → AST node Lambda(params, body, free_vars);
  resolver/typecheck: thu thập free vars, sinh EnvN, kiểu Closure.
- **P2 — lower:** air_builder hạ Lambda → (a) hàm top-level `lambda_body(env, args)`,
  (b) tại site: alloc env + ghi captures + dựng {code_ptr, env_ptr}; call-site closure
  → chèn env arg 0 + indirect call. Reuse BUG#49 R11-capture.
- **P3 — verify:** oracle repro (make_adder, scaled-integrate) O0+O1; regression +
  fixpoint.
- **P4 — follow-up:** capture-by-ref; capture kiểu-có-chủ (sau RFC drop-glue);
  std.iter map/filter/fold nhận closure; tiến hoá lên fat-pointer thống nhất (A) nếu
  muốn hợp nhất `fn(...)` và closure.

---

## 7. Drawbacks / Alternatives

- **Drawback:** hai kiểu (fn trần vs closure) tạo ma sát API (phải nhận đúng loại).
  Giảm bằng auto-wrap hàm trần → closure env=0 khi API cần closure.
- **Alt 1 — chỉ fat pointer (A):** thanh lịch hơn nhưng đổi ABI đã-fixpoint → rủi ro
  self-host cao hơn ngay lúc này. Để dành tiến hoá sau.
- **Alt 2 — không có closure, chỉ truyền struct-context thủ công** (`f(ctx, x)`):
  đúng là workaround hiện tại (comparator/predicate của std.sort). Ergonomic kém,
  không hỗ trợ `return`-closure. Không đạt mục tiêu #38.7.

## 8. Migration / compatibility

- Thuần bổ sung: cú pháp `|...|` mới, kiểu Closure mới. Không đổi ngữ nghĩa hàm trần
  (BUG#49) → mọi code hiện tại giữ nguyên hành vi. std.sort/std.numerical vẫn dùng
  `fn(...)` trần; có thể thêm overload nhận closure ở follow-up.
