---
name: bug-windows-to-i64-panic
description: "FIXED 4cf8239 (2026-07-17, A==B==C F846D40C, 339/339): Windows std.string.to_i64/to_f64 broke via the C-runtime ax_str_parse_* (unwrap panicked exit 101). Reimplemented in pure AXIOM (value ABI) in std/string.ax, dropping the extern C calls — works on Windows AND Linux now. Oracle t_parse=40. (was: OPEN pre-existing Windows C-runtime bug.)"
metadata:
  node_type: memory
  type: project
  originSessionId: 3228306b-52d7-4378-bb1c-a0b6cef57eba
---

# ✅ FIXED `4cf8239` — Windows `std.string.to_i64` / `to_f64` broken (C-runtime ax_str_parse_*)

**FIXED 2026-07-17** (`4cf8239`, A==B==C `F846D40C`, 339/339 + t_parse oracle=40):
reimplemented `to_i64`/`to_f64` in pure AXIOM (value ABI) in std/string.ax — dropped the
broken `extern "C" ax_str_parse_i64/f64` calls. Works on Windows AND Linux now; no C-runtime
dep. Found while adding the Linux parse twins (`1f96d3b`, now redundant but harmless).
Original report:

# (history) OPEN — Windows `std.string.to_i64` / `to_f64` broken (C-runtime ax_str_parse_*)

Found 2026-07-17 while shipping the Linux freestanding parse twins (`1f96d3b`).

## Symptom
On the WINDOWS COFF/self-link build:
```
import std.string
pub fn main() -> i32:
    return std.string.to_i64("42").unwrap() as i32   // → PANIC, exit 101
```
`to_i64("42")` returns `Err("Invalid integer format")` → `.unwrap()` panics
("AXIOM PANIC", exit 101). So `ax_str_parse_i64(s, &val)` (the ax_runtime.dll C symbol
declared `extern "C"` in std/string.ax:573) returns `false` for a valid "42".

## Scope
- **Windows only.** The **Linux** ELF target works correctly — the freestanding twins added
  in `1f96d3b` (bootstrap/runtime/panic.ax) parse "40"/"-8"/"2.5" fine (oracle elfparse=40).
  So the Linux path is more correct than the Windows C runtime here.
- Pre-existing; NOT caused by the Linux target work. Not covered by regression_repros.sh
  (still 338/338), so it went unnoticed.
- Root is in the C runtime (`runtime/ax_string_ops.c` ax_str_parse_i64) OR the value-ABI
  boundary of that `extern "C"` call (str arg + out-ptr) on the native COFF path — the same
  sret/by-ref C-ABI class that bit BUG#29. Likely the `out_val: ptr[i64]` out-param or the
  `s: str` arg isn't marshalled to the C function correctly on the native path.

## Fix options (when prioritized)
Simplest: replace std.string.to_i64/to_f64's `extern "C" ax_str_parse_*` calls with the
pure-AXIOM parse logic (same as the Linux twins in panic.ax) so BOTH platforms use the
freestanding implementation — drops the C-runtime dependency entirely. Would need A==B (or
B==C if it shifts codegen) + a regression oracle. Low urgency (parsing is not on the
self-host critical path), but it's a real correctness bug for user programs.
