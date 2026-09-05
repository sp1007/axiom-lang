---
name: bug-cross-tree-decl-node-segv
description: FIXED 2026-09-05 — B3 and B6 were ONE bug; pre_infer_func_signature indexed another unit's decl_node into self.tree and SEGV'd the compiler
metadata:
  type: project
---

# B3 + B6 = MỘT bug: `decl_node` xuyên cây trong `pre_infer_func_signature`

✅ **ĐÃ SỬA 2026-09-05.** A==B `F7146F3DBCDD3280E838B06538E946B33D2F3D19F32A18909BBAD81E22BC77E3`,
**2.329.600 byte**, regression **711/711 ở CẢ default LẪN `-O0`** ⇒ **baseline mới = 711**.
Frontend-only ⇒ chỉ cần **A==B**; mốc **B==C** gần nhất vẫn là `c3eae77` / `52D1ABD4…`.

## Triệu chứng
Module nạp trễ (không nằm trong bundle) ⇒ **compiler SEGV 139, không chẩn đoán**, deterministic ở
mọi mức tối ưu. `import std.{sync,thread,process}`, `import std.collections`, và bất kỳ module user
nào dùng `@compiler_intrinsic`.

## Điều kiện kích hoạt THẬT (đo bằng ma trận, không suy đoán)
Không phải `@compiler_intrinsic`, cũng không phải `std.collections`. Điều kiện là **module nạp trễ
gọi bất kỳ hàm nào khai báo trong unit gốc** (tức stdlib đã bundle):

| callee trong module nạp trễ | kết quả | vì sao |
|---|---|---|
| `compiler_intrinsic(..)`, `streq(..)`, `ax_str_len(..)`, `concat(..)` | **139** | hàm bundled thật (`std/runtime.ax:165`, `std/string.ax`) |
| `v.push(1)` (dạng method), `with_capacity[i64](4)` (type-arg tường minh) | **139** | hai call shape khác, cùng gốc |
| `@alloc`, `@memcpy`, `@assert`, `panic`, `print`, `malloc`, `ptr`, `Vec`, `Option` | **0** | resolve về builtin/extern ⇒ `decl_node == 0` |
| hàm generic khai báo trong **một module nạp trễ khác** | **0** | `decl_node` nằm trong cây của chính module đó |

⇒ **B6 không phải bug riêng.** Hai "bug" được mô tả bằng *triệu chứng bề mặt* hoá ra là một nguyên
nhân gốc — bài học phân loại đáng nhớ.

## Nguyên nhân gốc
`bootstrap/stage1/typecheck.ax:3770` — `pre_infer_func_signature(node_idx)` đánh chỉ số vào **`self.tree`**:
```axiom
let sym_idx = self.tree.nodes.data[node_idx].payload      // :3771  đọc ngoài biên
let sym = self.symtable.symbols.data[sym_idx]             // :3774  chỉ số không kiểm biên
if sym.type_id != 0 as u32:                               // :3775  <-- SIGSEGV
```
**Ba** chỗ gọi đưa vào một `Symbol.decl_node` **thuộc cây của unit KHÁC**:
- `typecheck.ax:5898` — đường gọi hàm tự do
- `typecheck.ax:1626` — fallback phân giải tên method
- `typecheck.ax:7081` — đường `f[T](..)` type-arg tường minh

⭐ **Bản sao thứ tư ở `typecheck.ax:1985` ĐÃ được canh sẵn, và comment tại đó mô tả đúng cơ chế này**
("*sym.decl_node may index a DIFFERENT module's AST tree … regression: t_mathx bignum methods*").
Dạng defect **"một luật, N bản sao, chỉ một bản được sửa"** — cùng họ P6 / `ownership.ax:138,162`.
⇒ Lặp lại bài học của [[lesson-exit-code-8bit-masking]]: **một sự thật ghi trong comment chỉ bảo vệ
đúng dòng nó nằm trên; muốn bảo vệ mọi dòng thì phải có MỘT chỗ ở duy nhất cho luật.**

**Vì sao chỉ nổ với module nạp trễ:** việc nạp module chạy **trong lúc resolve của unit GỐC**
(`main_air.ax:2055`), tức **trước** pre-pass `pre_infer_func_signature` của chính unit gốc
(`typecheck.ax:3271-3276`). Nên mọi symbol bundled còn `type_id == TYPE_UNKNOWN`, fallback theo yêu
cầu nổ, và một `decl_node` của root tree (~10⁴–10⁵) bị áp lên cây 8 node của module.

## Chứng minh (đo, không suy luận) — 3 mảnh độc lập
1. **gdb**: `SIGSEGV` tại `mov 0x8(%rcx),%edx`, `rcx` = địa chỉ **chưa map**. Offset 8 trong `Symbol`
   là `type_id` (`name_id`u32/`kind`u8/`padding`u8/`flags`u16/`type_id`u32) ⇒ đúng `:3775`.
2. ⭐⭐ **Thí nghiệm nhồi (quyết định)**: cùng module `streq("a","b")` đó, nhồi thêm 4000 hàm rác lên
   **44.011 node** ⇒ biên dịch **exit 0**. **Chỉ có biên mảng đổi.** ⇒ **Biến thể IM LẶNG nguy hiểm
   hơn biến thể crash**: khi chỉ số rơi trúng vùng hợp lệ, nó đọc một node vô can và **đóng dấu chữ
   ký sai** lên symbol dùng chung.
3. `symbol_trees[sym_idx]` (`resolver.ax:562/606/621`, ghi từ `symtable.current_tree`, đặt theo từng
   cây ở `:708`) **đã** ghi sẵn cây sở hữu — bốn chỗ nói trên chỉ khác nhau ở chỗ có tra nó hay không.

## Hai "manh mối" ban đầu ĐỀU LÀ HỆ QUẢ, không phải nguyên nhân
- **`undefined name 'raw32'`** (`raw32` là **biến cục bộ** ở `std/result.ax:24,35,105,115,136`):
  cascade downstream. Repro tối giản không có generic, không có `raw32`, vẫn SEGV. Trong biến thể
  `import std.collections`, `decl_node` của `result.ax` (cây 1091 node) **rơi trúng vùng hợp lệ** của
  cây collections (2840 node) ⇒ thay vì crash thì đọc nhầm node, monomorphizer nhân bản token text lạ
  vào `src` của collections (`ast.ax:265-290 clone_subtree_from`) ⇒ `node_text` mới "đánh vần" ra
  `raw32`. Cùng gốc, khác chỗ tiếp đất (trong biên vs ngoài biên).
- **`type 'hash_key' does not implement 'rem'`**: `hash_key[K](key)` đi đường `:7044-7087`; nhánh
  override `sym.kind == SYM_FUNC` ở `:7078-7087` lấy `fnty` bằng cách gọi chính hàm hỏng ở `:7081`.
  Hỏng ⇒ `result_type` kẹt lại là `GENERIC_INST` mang tên hàm ⇒ thông báo đó.
- Đối chứng trên đường **bundled** (`build`, cả `-O0`/`-O1`) đúng những cấu trúc này: `Option.is_some/
  unwrap` (thân hàm chứa `raw32`) và `HashMap.insert/get` (chỗ gọi `hash_key[K]`) **chạy đúng**
  (`bundled OK: 7`, `hashmap: 9`) ⇒ `std/collections.ax` + `std/result.ax` **hợp lệ**; lỗi là **giả**.

## Cách sửa (đã ship)
1. **Một chỗ ở duy nhất cho luật** — `typecheck.ax:3803 sym_decl_tree(sym_idx) -> ptr[AstTree]`
   (tra `symtable.symbol_trees`, fallback `self.tree`).
2. **`typecheck.ax:3824 pre_infer_symbol_signature(sym_idx)`** — kiểm biên `sym_idx`/`decl_node`; nếu
   cây sở hữu **khác**, chạy `pre_infer_func_signature` trên một `TypeChecker` **gắn với cây đó**
   (cache một helper theo cây sở hữu — `node_types` cấp phát theo từng cây nên **không được** hoán
   `self.tree` tại chỗ, và **không được** dựng helper mỗi lần gọi). Kết quả rơi vào `SymbolTable`/
   `TypeTable` **dùng chung** nên checker của module thấy được. Helper đặt `allow_cross_tree_infer =
   false` ⇒ đệ quy chặn ở đúng **một** tầng. Delta `diags_count` được cộng vào (luật B5 — không nuốt lỗi).
3. **`pre_infer_func_signature` tự kiểm đầu vào** (`:3856-3866`, CLAUDE.md §9) ⇒ giết luôn biến thể
   im lặng ở mục (2) của phần chứng minh.
4. Ba call site `:1646/:5993/:7177` (số dòng sau khi sửa) đi qua accessor.
5. `:2003-2010` — bản đã canh: **cố ý giữ nguyên hành vi**, chỉ trỏ về accessor. Đó là một **bộ lọc
   ứng viên** quét mọi symbol; ép suy diễn ở đó sẽ khiến kết quả phụ thuộc vào **số ứng viên đã quét**.

⛔ **KHÔNG phải cách sửa**: hoán `self.tree` tại chỗ; nhét lại các module đó vào bảng bôi trắng
(= khôi phục đúng defect "nuốt import im lặng" vừa mới sửa).

⚠️ **Bản vá KHÔNG byte-identical với seed** — code mới vẫn nằm trong self-image dù quá trình tự biên
dịch không hề đi qua đường nạp trễ. **A==B mới là tiêu chí**, đừng kỳ vọng byte-identical.

## Oracle (§7.1 — stdout, không phải exit code)
- `bin/t_b3lazyintrinsic.ax` + fixture `b3mod.ax` (gốc repo, cùng tiền lệ đặt chỗ với `std_util.ax`)
  ⇒ `Mô-đun nạp trễ gọi được stdlib đã bundle: 42`
- `bin/t_b3stdsync.ax` ⇒ `Nạp trễ std.sync không còn làm sập trình biên dịch: 42`
- **Hiệu chuẩn đã đo**: trên driver trước-fix cả hai **chết 139, không sinh exe**, ở `-O0` lẫn `-O1`.

## CÒN MỞ — B3b (defect KHÁC, đã tách bạch bằng thí nghiệm)
`import std.{iter,process,cli}` **vẫn 139**. Nguồn của module nạp trễ **không được tiền xử lý**, nên
`import std.collections` bên trong nó không bị bôi trắng ⇒ loader đọc `std/collections.ax`
(+ `std/result.ax`) **lần thứ hai**; typecheck bản trùng in 10× `undefined name 'raw32'` rồi SIGSEGV.
Đã chứng minh **độc lập với bản vá này** bằng cách dựng biến thể **chỉ-P0**: y hệt lỗi, y hệt crash
(P1 chỉ bỏ đi mấy dòng `hash_key ... 'rem'`). Đây là **stage 2** của lộ trình stdlib (đăng ký 8 tên
bundled thành `SYM_MODULE`, tra scope toàn cục thay vì nạp file) — **đổi thiết kế, không phải vá**.
Repro đã gửi ngân hàng: `nestbundledmod.ax` + `bin/probe_nestbundled.ax`.
`std/iter.ax` còn có 23 lỗi parse của riêng nó.

⚠️ **Ghi chép cũ cần đính chính**: BACKLOG từng ghi `std.string`/`std.collections` crash — **không
chính xác** trên đường `build` (chúng được bundle nên import bị bôi trắng).

Liên quan: [[lesson-taskstop-leaves-suite-running]] (bẫy hạ tầng gặp ngay trong lượt gate của bug này)
