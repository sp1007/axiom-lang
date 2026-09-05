---
name: bug-b3b-cross-module-index-and-loader-defects
description: B3b is NOT a stdlib/bundling problem — it is a general cross-module compilation defect reachable by two plain user modules; plus two loader defects and a refutation of roadmap stage 2
metadata:
  type: project
---

# B3b — cơ chế mà lộ trình ghi là SAI; và stage 2 như đề xuất **vi phạm quyết định D1-3**

Đo 2026-09-05 (read-only investigator), trên driver `bin/axc_native.exe` A==B
`F7146F3D…`, 2.329.600 byte, baseline 711/711. **Không build compiler, không sửa source.**

---

## 1. Sự thật ĐÚNG, nhưng KHÔNG phải nguyên nhân

**Đúng (sự kiện):** loader nạp trễ **không tiền xử lý** nguồn module. Import của unit gốc bị bôi
trắng ở `main_air.ax:619-620` (`strip_imports` → `strip_package_prefixes`; bảng `:275-296`, vòng bôi
trắng `:335-399`). Đường nạp trễ đọc file ở `main_air.ax:1994` rồi đưa **thẳng vào `new_lexer`**
(`:2004`) — không `strip_imports`, không `strip_package_prefixes`. Đo: `bin/probe_nestbundled.ax` gây
ba lần `read_file_content` dài 998 / 17941 / 7082 = `nestbundledmod.ax` / `std/collections.ax` /
`std/result.ax`. **Việc nạp trùng là có thật.**

**BÁC BỎ (quan hệ nhân quả) — ba phép đo độc lập:**

| thí nghiệm | kết quả |
|---|---|
| `import std.collections` ở gốc + `--no-stdlib` (⇒ **không hề có bản bundled**, 2 lần nạp, **không trùng lặp**) | **vẫn** 10× `undefined name 'raw32'` + SEGV |
| module nạp trễ chỉ `import std.result` (có trùng lặp result.ax) | **build sạch, exit 0** — trùng lặp là **vô hại** |
| **hai module user thuần**, không dính `std` ở đâu cả | **tái hiện y hệt** |

⇒ **Nạp trùng là hiện tượng ĐI KÈM, không phải nguyên nhân.**

## 2. Defect thật: chỉ số node xuyên cây (lại đúng họ B3/B6)

Module A được module B `import`; B được gốc `import`. **B không tham chiếu gì của A.** Yếu tố kích
hoạt là **SỐ NODE của A** — không gì khác.

Repro tối giản (đã gửi ngân hàng, untracked): `b3bpkg/moda.ax` (5 hàm `pub fn a_padN(aparam: i64)`,
**không cần generic**), `b3bpkg/modb.ax` (`import b3bpkg.moda`; generic `BVec[T]`, `b_new`, `b_push`,
`b_map`), `bin/probe_b3b_dotted.ax`.

```
error: undefined name 'aparam'      (×2)
error: 2 type error(s) in imported module 'b3bpkg.modb'
```
⭐ **`aparam` chỉ tồn tại trong module A**, nhưng bị báo trong lúc typecheck **module B**.
Quét kích thước: 1 pad ⇒ sạch; 5/10/20/40 pad ⇒ hỏng. Y hệt ở `-O0` và `-O1`, y hệt khi bundled và
khi `--no-stdlib`. **Bỏ generic của A không đổi gì; bỏ NODE của A thì hết.**

Đây đúng vân tay của [[bug-cross-tree-decl-node-segv]] ("thí nghiệm nhồi"). Chuỗi `undefined name` do
`typecheck.ax:6464-6467` đọc `node_text` trên một node mà token text đã bị
`ast.ax:254-302 clone_subtree_from` nối vào `src` của B — tức monomorphizer **nhân bản subtree ra khỏi
cây của A** khi instantiate generic của B. Bản vá B3/B6 chỉ làm cứng `pre_infer_func_signature`
(`typecheck.ax:3793-3866`), **không phủ đường này**.
`mono.ax:483-487` và `typecheck.ax:5419-5423` **có** tra `symbol_trees`, nhưng
**`typecheck.ax:5416-5427` dùng `decl_node` KHÔNG kiểm biên** (đối lập `:818`, `:1213`, `:2327`,
`:2454` — có kiểm). Chỉ đích danh dòng cần một bản build có instrument ⇒ việc của implementer.

⇒ **B3b KHÔNG phải vấn đề stdlib/bundling. Nó là defect biên dịch xuyên module tổng quát, hai module
user bất kỳ chạm tới được.** `std.iter/process/cli` chỉ là nạn nhân ồn ào nhất.

## 3. Hai defect nữa tìm được trên đường (mỗi cái filable riêng)

### B3b-2 — use-after-free trong loader (im lặng, phụ thuộc mức tối ưu)
`std/string.ax:760-761`: `replace` trả về **chính `s`** khi không có match nào.
`main_air.ax:1942` `let rel_path = std.string.replace(mod_name, ".", "/")`, rồi `main_air.ax:1957`
`@free(get_str_ptr(rel_path))`. Với **mọi tên module không có dấu chấm** (mọi module user một đoạn),
`rel_path` **CHÍNH LÀ** `mod_name` ⇒ `:1957` giải phóng `mod_name`. Mọi lần dùng sau là UAF: `:2034`,
`:2059`, và **tệ nhất `:2088`, nơi `full_name = mod_name + "." + local_name`** — tên đủ điều kiện mà
**mỗi symbol xuất khẩu được đăng ký vào scope toàn cục**; cộng double-free ở `:2148`.
Đo: cùng một lệnh build in `'\xef\xbf\xbd\xef\xbf\xbd'` ở `-O0` và `'Sj'` ở `-O1`; bản có dấu chấm in
đúng `b3bpkg.modb`. ⭐ `resolver.ax:801-803` có **đúng lời gọi đó với dòng free đã bị comment lại** —
ai đó từng đụng phải và **né tại chỗ** thay vì sửa hợp đồng aliasing.
Sửa ở call site (đừng free một alias) là rẻ nhất; đổi `replace` thành luôn-copy chạm hợp đồng stdlib
nóng khắp self-image ⇒ phải đo A==B.

### B3b-3 — cổng B5 kiểm quá muộn
`main_air.ax:2033-2035` ghi `parser_ptr.diags_count` rồi **cố ý đi tiếp** (comment `:2029`) vào
resolve/typecheck/export-walk trên **AST dựng dở**. Đo: `import std.iter` báo `23 parse error(s)` rồi
**SEGV 139**; `import std.cli` báo `9 parse error(s)` rồi **SEGV 139**. Trả về khác 0 ngay sau khi đếm
lỗi parse biến **2 trong 3** SEGV của std thành reject sạch, tốn ~3 dòng.
⇒ **Giá trị trên rủi ro cao nhất cả khu vực này.** (Đang thực hiện.)

## 4. 8 tên bundled — lấy từ bảng THẬT
`main_air.ax:275-296` `preprocessed_module_name(i)`, 11 mục; `bundled_module_count() = 8` (`:272`).
**Bundled (0-7, thứ tự `concatenate_stdlib` nối vào):** `std.result`, `std.mem.alloc`,
`std.scheduler`, `std.runtime`, `std.os`, `std.string`, `std.io`, `std.collections`.
**Tiền xử lý nhưng KHÔNG bundle (8-10):** `std.os.win32`, `std.os.linux_sys`, `std.net`.

## 5. ⛔ Stage 2 như đề xuất **VI PHẠM quyết định D1-3** (khớp theo chính tả)

`concatenate_stdlib` nối cả 8 module vào **MỘT scope toàn cục phẳng** ⇒ symbol bundled **không mang
danh tính module nào cả**. Nên "tra tên trong scope toàn cục" chỉ có thể nghĩa là: **cắt bỏ
`std.collections.` rồi tra `new_vec` trần**. Đó chính là `strip_package_prefixes` **dời từ tầng văn
bản vào resolver** — cùng một luật, tầng mới, đúng lớp defect P4 / RFC 0037 / `is_verbose_debug`.

Nó còn **vỡ trên overload set thật**: `unwrap` có ở cả `Option` lẫn `Result` (`std/result.ax:38`,
`:118`); `map` ở cả `Option` lẫn `Vec` (`std/result.ax:66`, `std/collections.ax:62`); `get`/`filter`
tương tự. Tra tên trần trả về đầu chuỗi overload, mà `resolver.ax:653/662` **đã ghi rõ** rằng việc
chọn giữa nhiều `M.foo` là **TÙY TIỆN (theo thứ tự hash-slot)**. ⇒ `std.collections.map` và
`std.result.map` sẽ không phân biệt được, và lời hứa "codegen delta = 0" chỉ đúng cho những tên **tình
cờ đang là duy nhất**.

**Có bản ĐÚNG, và tiền đề của nó đã tồn tại:** `main_air.ax:682-689` đã ghi sẵn một vùng `srcmap`
chính xác cho **từng** module bundled (`path_res`, `path_coll`, …). Vậy module gốc của mọi symbol
được nối là suy ra được: `Symbol.decl_node → node.offset → srcmap_find → path → chỉ số module`. Khoá
`std.collections.new_vec` theo **cái đó** mới là theo danh tính và mới tôn trọng D1-3.
⇒ **Đó chính là "stage 4" của lộ trình — nên stage 4 là TIỀN ĐỀ của một stage 2 trung thực với spec,
không phải phần đi sau. Lộ trình đang ghi ngược thứ tự phụ thuộc.**

## 6. Quan hệ với `strip_package_prefixes` — stage 3 KHÔNG rẻ hơn, và B1 có đường rẻ hơn cả hai
- Stage 2 **không phụ thuộc** `strip_package_prefixes` và **không làm nó xoá được**. Chúng độc lập:
  bộ viết-lại-văn-bản phục vụ call site `std.string.foo(...)` của **unit gốc**; stage 2 phục vụ phân
  giải của **module nạp trễ**.
- Stage 3 (xoá nó) **phụ thuộc** stage 2 tồn tại, vì hôm nay chính lời gọi `std.string.replace(...)`
  của compiler chỉ chạy được nhờ viết lại văn bản. ⇒ Stage 3 **đắt hơn**, không phải rẻ hơn.
- **B1 (hỏng hằng chuỗi) không cần cái nào trong hai.** Vi phạm §3 nằm ở chỗ bộ viết lại **mù với
  literal**. Cho `strip_package_prefixes` (`main_air.ax:401-…`) **bỏ qua vùng `"…"`** là một hàm,
  **không cần RFC** (nó *khôi phục* nghĩa đã ghi của hằng chuỗi), không đổi ngữ nghĩa đường nào, và
  **mở khoá luôn số CỘT trong chẩn đoán** (`print_helpers.ax:718`, `typecheck.ax:2284`) — đúng thứ mà
  stage 3 vốn được muốn vì nó. ⚠️ Cảnh báo phải ĐO chứ đừng giả định: `cgen.ax:773,777` và
  `resolver.ax:807-818` so sánh với literal `"std.string.len"` mà **hiện đang bị viết lại thành
  `"len"`** trong self-image ⇒ làm bộ viết lại biết-literal sẽ đổi chính các literal đó. Cần A==B + suite 711.

## 7. Thứ tự khuyến nghị (rủi ro thấp trước) và phán quyết RFC
**Thứ tự:** B3b-3 (loader dừng khi có lỗi parse) → B3b-2 (UAF loader) → B3b thật (chỉ số xuyên cây:
`sym_decl_tree` + kiểm biên §9) → `strip_package_prefixes` biết-literal (đóng B1, mở khoá cột) →
**rồi mới** định giá lại stage 2 trên nền danh tính theo module.
⇒ **Stage 2 không sửa được nguyên nhân nào của B3b**; nó chỉ ngăn compiler đi tiếp trên con đường
hỏng cho đúng 8 cái tên.

**RFC:**
- **B3b, B3b-2, B3b-3: KHÔNG cần RFC.** Sửa lỗi thuần, khôi phục bất biến đã tuyên bố (§9; một `str`
  không được free hai lần; một chẩn đoán không được theo sau bởi crash).
- **`strip_package_prefixes` biết-literal: KHÔNG cần RFC** (khôi phục ngữ nghĩa literal đã ghi).
- **Stage 2: BẮT BUỘC RFC.** Nó đổi `std.X.f` **biểu thị cái gì**, `import std.X` bên trong module có
  phải no-op không, user có che được tên bundled không, và thêm cờ/bảng phụ vào `Symbol`. Đó là hợp
  đồng ngữ nghĩa ở bề mặt ngôn ngữ ⇒ §13 phủ, dù ngữ pháp không đổi.

## 8. Chỗ sửa (đúng lớp) và rủi ro
Defect B3b **không nằm ở resolver/loader/driver**, mà ở **typecheck/mono**. Các site cần audit và đưa
qua `typecheck.ax:3793 sym_decl_tree(sym_idx)` + kiểm biên bắt buộc: `typecheck.ax:5416-5427`
(**không kiểm biên**), `typecheck.ax:1206-1216`, `mono.ax:418-437`, `mono.ax:473-529`, và
`air_builder.ax:769 / :2096 / :2705 / :4005` (đều index `self.mb.tree` bằng `sym.decl_node` thô;
`:2704` và `:4002` có kiểm biên, `:769` và `:2096` **không**). B3b-2 và B3b-3 ở `main_air.ax` (driver).
**Không cái nào chạm backend.**

⭐ **Self-image KHÔNG BAO GIỜ dùng loader nạp trễ.** `scripts/fast_fixpoint.ps1:20` build
`bootstrap/stage1/tmp_concatenated_air.ax` — một file đã nối sẵn, 3 import duy nhất
(`std.collections`, `std.io`, `std.string`) đều nằm trong bảng bôi trắng; 33 dòng
`import bootstrap.stage1.*` chỉ có trong `main_air.ax` **chưa nối**, vốn không phải đầu vào build.
⇒ **không đường self-host nào chạm `ax_driver_load_module`** ⇒ thay đổi ở đây là frontend ⇒ **A==B đủ**,
và **đừng kỳ vọng byte-identical với seed**.

**8 hàng regression chạm đường module** (rủi ro nếu redesign stage 2): `t_modcollide` (:172),
`t_stdmath` (:962), `t_stdprefixmod` (:968), `t_modparseerr` (:971), `t_modnomember` (:975),
`t_stdmathnomember` (:977), `t_b3stdsync` (:988), `t_b3lazyintrinsic` (:989).

## 9. Oracle + hiệu chuẩn (§7.1)
`bin/t_b3bcrossmod.ax` (fixture `b3bpkg/moda.ax`, `b3bpkg/modb.ax`) ⇒
`mô-đun nhập mô-đun khác biên dịch được: 42`, chạy ở `-O0` **và** default.
**Hiệu chuẩn đã đo trên driver hôm nay:** exit **1**, không sinh exe, `error: undefined name 'aparam'`
×2 — y hệt ở `-O0` và `-O1`.
Phụ (UAF): assert tên module in ra **nguyên vẹn**; hôm nay in rác **khác nhau theo mức tối ưu**.
⚠️ **KHÔNG dùng `std/iter.ax` làm acceptance test** — nó có 23 lỗi parse của riêng nó.

Liên quan: [[bug-cross-tree-decl-node-segv]] · [[lesson-taskstop-leaves-suite-running]]
