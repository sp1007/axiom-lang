---
name: bug-array-return-type-dropped
description: "FIXED: `fn f() -> [T;N]` (and `let a: [T;N]=..`, plus slice/ptr-star/fn-type annotations) were dropped to VOID because pre_infer_func_signature only recognized NODE_TYPE_EXPR/GENERIC_TYPE — causing array literals not to coerce to the declared element type -> i32/i64 stride miscompile (`for x in mk()` hung, `mk()[i]` garbage)."
metadata:
  node_type: memory
  type: project
  originSessionId: 7fa6650d-d306-4f61-9562-d7dcda064554
---

✅ **FIXED** `6845608` (frontend-only typecheck.ax, fast fixpoint **A==B `E13AB200`**, regression **138/138**). Surfaced by RFC 0020 string-surface bug-probing — the visible symptom was `for x in mk()` (mk returns a fixed array) HANGING.

## Root cause
A `[i64;3]` return type parses to **NODE_ARRAY_TYPE (49)** directly (not wrapped in NODE_TYPE_EXPR). `pre_infer_func_signature` (typecheck.ax ~L1751) only matched `NODE_TYPE_EXPR`/`NODE_GENERIC_TYPE` as the return annotation, so an array-returning fn was silently typed **VOID** (`current_return` = TYPE_VOID = 14). The declared element type was thus lost: a bare-int array literal `[10,20,30]` (default **[i32;3]**, stride 4) was never coerced to the declared **[i64;3]** (stride 8) because NODE_ARRAY_LIT's element-coercion reads `expected.extra` and `expected` was VOID. The caller then read the returned value with the i64 stride into i32-strided data → `mk()[i]` garbage, `for x in mk()` hung. Same gap at the `let a: [T;N] = ..` site (~L2007) misclassified the array-type annotation as the initializer (annotation ignored → `let: [i64;3]` also broke).

`ptr[T]`/`Vec[T]`/`Option[T]` returns were unaffected because `foo[..]` parses to NODE_GENERIC_TYPE (already handled); plain `i64`/`MyStruct` to NODE_TYPE_EXPR. Only the bracket/star/fn forms (ARRAY/SLICE/PTR-star/FUNC) were dropped.

## Fix
Added `is_type_annotation_node(k)` = TYPE_EXPR|GENERIC_TYPE|ARRAY_TYPE|SLICE_TYPE|PTR_TYPE|FUNC_TYPE, used at both the signature-return site and the let-annotation site. No self-host/stdlib code returns `[T;N]`/`[T]`/`*T`/`fn(..)` or annotates a fixed-array local (only aspirational unbundled crypto.ax uses `-> [u8;32]`), so the compiler's own typing is unchanged → **fixpoint preserved**. The array-return **ABI already works** (verified: with explicit `[10 as i64,..]` elements, array param+return returned correct values pre-fix) — this was purely a frontend type-drop, NOT an sret/backend bug.

## Residual follow-up — ✅ CLOSED `c8dd125` (A==B `64D988C7`, 139/139)
`let a = [10,20,30]  (unannotated -> [i32;3]);  rd(a)  where rd(a:[i64;3])` silently miscompiled (exit 8). Fixed in the call-arg inference (NODE_CALL_EXPR, ~L2762): for a fixed-array param of a **NODE_IDENT non-generic** free call, pass the param type as the arg's expected type → a **direct** `[..]` literal now COERCES (`rd([10,20,30])` → 40, RFC 0005 ergonomic win); if the arg is a DIFFERENT fixed-array type (a variable already [i32;N]), **reject** cleanly (annotate `let a:[i64;3]=..` or cast). Gated NODE_IDENT+non-generic+array<->array so method self/generic positions never misalign and all other args untouched; no bundled code uses array params → fixpoint safe. Oracle `bin/t_arrargmismatch.ax` (reject). `register_array` DEDUPES so array-type equality = id equality.

Oracle: `bin/t_arrret.ax` (array return + `for x in mk()` + `let: [i64;3]` → 70). Row `t_arrret|exit|70` in scripts/regression_repros.sh.

## Cluster part 3 — struct-constructor array field — ✅ `c103b9f` (A==B `366C1B13`, 140/140)
Same element-coercion class: `Grid(cells: [11,22,33,44])` where `cells:[i64;4]` corrupted the field — inline literal defaulted to [i32;4], element stores used i32 stride (4) into the 32-byte slot → `g.cells[1]`=33 (bytes 8..15), `[2]/[3]`=0. Any non-i32 element field (`[u8;N]`/`[i16;N]`/`[f32;N]`) hit. Root: `try_instantiate_struct_ctor` coerces GENERIC-struct fields but early-returns for NON-generic (typecheck.ax:452); general call-arg loop infers field values with UNKNOWN. Fix: in call-arg inference, when callee is a struct ctor, walk the struct's RESOLVED field types (StructInfo, tree-independent) positionally (named args are positional per BUG#21) and infer ARRAY-field values with the field type as expected; scalar/ptr/struct fields keep UNKNOWN → fixpoint-safe. Oracle `bin/t_ctorarrfield.ax` (110). **Whole "int-array-literal element coercion" cluster (param/return/for-in/let-annotation/ctor-field) now complete.**
