---
name: bug93-qualified-str-call-segfault
description: "BUG#93 (silent-segfault half) FIXED via typecheck reject: unresolved module-qualified calls (std.math.gcd, short-alias string.concat, typos) no longer miscompile to a garbage indirect call. NOT a str-ABI issue — original diagnosis was wrong. Making UNBUNDLED modules actually work = separate OPEN follow-up (import-driven bundling)."
metadata: 
  node_type: memory
  type: project
  originSessionId: 0787a3ec-d780-4021-ba63-9d15cc181559
---

✅ **BUG#93 silent-segfault class FIXED** — `468f581` (typecheck reject, frontend-only, A==B `824572CE...`, 137/137). ⚠️ **Original root-cause in this file was WRONG** — it is NOT a 16-byte `str` argument ABI issue.

## True root cause (investigator, 2026-07-10)
A **direct module-qualified call whose member does not resolve** was silently accepted, then air_builder lowered the unresolved `NODE_FIELD_EXPR` callee to a garbage register and emitted `call *reg` → jump to unmapped addr → **SIGSEGV (exit 139)**, no diagnostic. Same accept-then-miscompile class as BUG#53/#68/#88. Affected: `std.math.gcd(...)` (module not bundled), `string.concat(a,b)` **short alias** (only `std.string.` stripped, bare `string.` not), any typo'd member. NOT str-specific — `math.gcd(int,int)` and 16-byte struct qualified calls segfaulted too.

**Mechanism:** the native driver strips `import` lines textually before parsing (`main_air.strip_imports`) and bundles a **hardcoded whitelist** of stdlib files (`concatenate_stdlib`) + rewrites a **hardcoded prefix whitelist** (`strip_package_prefixes`: only mem.alloc/scheduler/os*/string/io/net). So on the native path **no `SYM_MODULE` exists** (imports gone) and any non-whitelisted module member survives unresolved. The operator path (`a+b`) works because it resolves by name-scanning the bundled symtable (`resolve_op_method`), bypassing field resolution.

## The fix (what shipped)
`typecheck.ax` call-inference `NODE_FIELD_EXPR` flag-2048-unset branch (~L2446): when `sym_idx==0`, walk receiver chain to ROOT ident; root is BOUND iff `symbols[root.payload].name_id == intern(node_text(root))`. Bound value receiver (`r`/`self`/global) → spared; undefined namespace (`std`/`math`/`string`) → **reject** `error: unresolved call to '<m>' on undefined namespace '<root>'`. Test: `tests/sema/err_unresolved_module_call.ax`.

**KEY LESSON (3 RED cycles to get here):** at typecheck time local scopes are POPPED, so `symtable.resolve` sees ONLY module/global symbols — it CANNOT detect a local/param receiver (`r`, `self` → returns 0 → false "namespace" → self-build rejects its own `r.unwrap()`/`self.read()` → RED). Must use the **resolver-stamped `payload`** on the ROOT ident (set to sym_idx during resolution when scopes were live) and match its symbol NAME to the spelling. Also do NOT gate on `rec_type==TYPE_UNKNOWN` — Option/Result value methods (BUG#90 backend-intercept) legitimately have UNKNOWN rec_type at typecheck (v1 RED). Also don't rely on `str==str` (untested lowering) — compare interned name_ids (u32).

## Follow-up "import-driven bundling" is MOOT for now (2026-07-10 finding)
`std.math.gcd` etc. now give a clean ERROR, not a working call — and that is **correct**, not a stopgap. Import-driven bundling (scan `import std.<mod>` → bundle `std/<mod>.ax` → extend `strip_package_prefixes`) has **near-zero payoff**: per `std/MODULE_STATUS.md`, the ONLY grammar-conformant stdlib modules (string/io/collections/os/result/option/process/scheduler/runtime/alloc) are **already bundled**. Every unbundled module (`std.math` 218 parse errs, `net`/`iter`/`json`/`fmt`/`mem`/`crypto`/`cli`/`arch`/`gpu`/`compiler`/`ffi`…) is **ASPIRATIONAL** — non-grammar dialect (braces `{}`, `enum{}`, Go-style `fn (recv)`, `Self`, assoc `type`). Bundling one would just inject hundreds of parse errors. So DON'T build import-driven bundling to "fix" BUG#93 — the reject IS the fix. To actually enable e.g. `std.math`, first **rewrite math.ax to grammar** (large per-file, indent/sum-type/self-method) OR **RFC to extend grammar** (enum/Self/assoc-type) — a separate feature backlog, tracked in `docs/next-step-15.md`, NOT a self-host blocker. Short-alias `string.foo` (vs full `std.string.foo`) is a lesser gap (idiomatic = full path, now guarded by the reject). See [[string-utf8-default]].
