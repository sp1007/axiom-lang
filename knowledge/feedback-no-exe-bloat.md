---
name: feedback-no-exe-bloat
description: "User standing constraint (2026-07-22): do not bloat the emitted executables. Measure output size before/after any codegen/linker change; size regressions need justification."
metadata:
  node_type: memory
  type: feedback
---

User, 2026-07-22, mid-task while approving the group-A plan: **"chú ý, tôi không muốn làm
phình các file thực thi"** — do not let changes inflate the emitted executables.

**Why:** AXIOM emits its own object files and links them with its own linker, so nothing
external trims the output. Every byte the backend or linker decides to write is a byte in
the user's exe permanently. The measured case that prompted this: a `[i64; 200000]`
zero-initialized global inflates an exe **77KB → 3.2MB** because the zero bytes are stored
literally in the file — this is exactly what RFC 0030 `.bss` exists to fix.

**How to apply:**
- Any change touching codegen, the linker, section layout, or stdlib bundling: **measure
  the output exe size before and after** and report the delta. Treat a size increase as a
  finding that needs a stated reason, not an acceptable side effect.
- Prefer designs where new capability costs zero bytes for programs that do not use it
  (functions linked on demand > always-emitted builtin surface — this reasoning is what
  settled [[m4-compliance-suite-spec-vs-impl-gap]] and RFC 0020 §10 against new types).
- Compiler self-build size counts too: `bin/axc_native.exe` is a shipped artifact.
- This is a size constraint, not a speed one — do not trade correctness for bytes
  (CLAUDE.md §10 still governs: correctness first).

Related: [[feedback-ergonomics]], [[infra-defender-build-throttle]].
