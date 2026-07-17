---
name: bug57-match-option-native
description: "BUG#57 FIXED (6ef4176, backend-only) — native `match` on Option/Result now dispatches. Fix lives entirely in air_builder+typetable, OFF the stdlib generic ABI; the type-checker approach broke self-host and was reverted. Fixpoint B==C. See [[next-step-18-bug57-shipped]]."
metadata:
  node_type: memory
  type: project
  originSessionId: 05d3f904-e67c-4f1a-9bfa-33caeb26ab45
---

**✅ BUG#57 FIXED — commit 6ef4176 (2026-07-05), backend-only, fixpoint B==C
bit-identical, 24 regression + 4 Option/Result tests green.** RFC 0012 shipped.

**What was wrong:** native `match` on a builtin Option/Result scrutinee emitted
NO dispatch — `lower_match` only handled user sums (kind 6) + integers and
returned early for Option/Result, so `match Some(42)` silently returned register
garbage (usually 0). Option/Result are POINTER-tagged (None=null, Some=box; Ok=even
box, Err=box|1; payload at [box&~1]), not the user-sum tag@field0 box.

**The winning design (all in the BACKEND, no typecheck/ABI change):**
1. `typetable.ax`: `register_option(inner)` / `register_result(ok,err)` — kind
   11/12, **size 16 like a sum (kind-6)**, aggregate ⇒ value is an 8-byte box
   pointer copied like a sum. ONLY air_builder creates these entries.
2. `air_builder.ax` constructors (Some/Ok/Err): compute `box_ty` LOCALLY via
   register_option/result(payload_t) and use it for `OP_ALLOC`, so the box reg is
   typed kind-11/12 (NOT str-12) ⇒ EQ/COPY/MAKE_REF stop taking the str-by-address
   deref path. None = ICONST i64 0.
3. `air_builder.ax` `match_arms_tagged_kind`: classify a match as Option (0) /
   Result (1) by **ARM NAMES** (Some/None ⇒ Option, Ok/Err ⇒ Result), independent
   of the scrutinee's static type (which may be a std-sum kind-6 or untyped).
4. `air_builder.ax` `lower_match_tagged`: pointer-tagged dispatch — None `==0`,
   Some `!=0`, Ok `(x&1)==0`, Err `(x&1)!=0`; payload = `LOAD[box&~1]`, load width
   from the binding symbol's type (i64 fallback).

**KEY LESSON — why the "obvious" type-system fix (RFC 0012 Part A) FAILED and was
reverted:** typing `Option[T]`/`Result[T,E]` ANNOTATIONS as kind 11/12 in
typecheck.ax (NODE_GENERIC_TYPE + Some/Ok/Err call-typing) broke self-host — B (the
compiler built by A) segfaulted compiling anything. Root cause: std/result.ax
methods are GENERIC (`is_some[T](self: Option[T])`, `unwrap[T]`…); retyping the
`self: Option[T]` param changed its representation through monomorphization and
crashed. Bisect proved the typecheck change ALONE breaks it, regardless of entry
size (8 or 16). So the fix must NOT touch how the stdlib's generic Option/Result
values are typed — keep kind-11/12 values purely LOCAL to construction + match.
Since those flow exactly like user-sums (kind-6, long self-hosting), they're safe.

**Fast iteration unlocked this:** `scripts/fast_fixpoint.ps1` — seed from the
native daily-driver `bin/axc_native.exe` (A), A builds B, B builds C, check B==C
(~9s total) instead of the hours-long gcc `axc_stage1` path (10000s+ w/ machine
sleep). B==C (both under new rules) is the real fixpoint; A≠B is just the one-time
codegen transition. See [[feedback-fixpoint-async-rule]].

**✅ BUG#59 ALSO FIXED — commit 0b921f5 (2026-07-05), fixpoint B==C.** Matching a
SUM-typed FIELD of a struct (`match bx.c` where `c: Color`) segfaulted native. Root
cause (dump-air): GET_FIELD/SET_FIELD sized a sum field by its type entry (16 = box
alloc size) and took the 16-byte-inline path, but regalloc homes a sum-typed dest
(kind 6/11/12) as an 8-byte POINTER → the 16-byte write overran the 8-byte slot →
stack corruption. Fix: `field_is_pointer_sum()` in x86_selector — a struct field of
kind 6/11/12 is LOADED/STORED as an 8-byte box pointer (first branch in GET_FIELD/
SET_FIELD, before the size==16 path), consistent with sum values everywhere. Field
still reserves entry size in layout; only load/store width changes. Test
match_sum_field_BUG59.ax (un-SKIP, print-based). Same cluster as
[[bug56-nested-sum-payload]] / [[bug54-qualified-variant]] / [[bug51-hunt-progress]].
The tagged-type representation cluster (BUG#51/54/55/56/57/58/59) is now CLOSED.
