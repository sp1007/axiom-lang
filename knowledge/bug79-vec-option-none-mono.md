---
name: bug79-vec-option-none-mono
description: "BUG#79 FIXED — v.push(None) into Vec[Option[T]] không còn hỏng monomorphization (unresolved ax_push)"
metadata: 
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

**BUG#79 FIXED** ✅ d28a3e8 (2026-07-08, frontend-only, fixpoint A==B `97EE6215…`, regression 107/107).

Đây là follow-up MỞ còn lại của [[bug78-array-of-option-none]] ("Vec[Option[T]] với None vẫn crash qua đường push/get"). Hóa ra **KHÔNG phải bug backend/store** như nghi ban đầu — mà là **monomorphization miss ở typecheck**.

## Triệu chứng
`v.push(None)` trên `Vec[Option[i32]]` → x86 self-linker báo `Unresolved external symbol 'ax_push'`. `v.push(Some(3))` thì OK; chỉ `None` hỏng. `ax_push` = codegen fallback dùng TÊN TEMPLATE thô "push" (prefix "ax_") vì call node vẫn trỏ template sym, KHÔNG có instance mangled `_AX_std_push__...`.

## Root cause
Method-call generic-instantiation path trong `typecheck.ax` (~line 2064-2091). Receiver `Vec[Option[i32]]` xác định ĐẦY ĐỦ T=Option[i32] (first-binding-wins, [[bug66-hashmap-i64-value-corruption]]) → `inferred` toàn concrete. NHƯNG guard `has_generic_arg` quét CẢ `arg_types` (kiểu các đối số thực): đối số `None` có kiểu Option under-determined (generic-looking) → `is_generic`=true → `has_generic_arg`=true → **SKIP instantiation** → call giữ template → `ax_push`.

## Fix (nguyên lý)
Một ĐỐI SỐ generic-looking chỉ nên chặn instantiation khi có type-param **thực sự chưa resolve** (bị i32-default ở line 2067). Khi receiver resolve mọi param concrete, `None` KHÔNG được defer. Thêm `has_unresolved` (quét inferred==UNKNOWN TRƯỚC khi i32-default); chỉ khi `has_unresolved` mới cho arg_types-generic set `has_generic_arg`. Param INFERRED genuinely-generic (trong thân generic) VẪN luôn defer (không đổi).

## Bài học
- `Vec.get[T]` trả `Option[T]` → `Vec[Option[i32]].get()` = `Option[Option[i32]]` (nested match); `len` là FIELD không phải method (`v.len` không `v.len()`). Construct = `new_vec[T]()` không `Vec[T]()`.
- Direct-store path `v.data[i]` VÀ full `get()` path đều xanh (=110). Cụm None-trong-container [[bug78-array-of-option-none]] giờ ĐÓNG cho array literal + Vec.
- Path direct-call explicit `f[T](...)` (typecheck.ax ~2748) dùng `args` = type-args tường minh, KHÁC path này (infer từ receiver) → không đụng.
