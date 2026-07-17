---
name: bug-match-optstr-payload-length
description: "match on Option[str]/Result[str,E] dropped the str length half (garbage s.len) — synth-struct payload width; FIXED 6c0b976"
metadata: 
  node_type: memory
  type: project
  originSessionId: 102de304-2f65-4ad8-b994-972ef133a05a
---

✅ **FIXED 2026-07-10 (h)** — `6c0b976`, `origin/main`=`6c0b976`, A==B fixpoint **`C3EBD800`**, **144/144**. Daily-driver `bin/axc_native.exe` rebuilt from this fixpoint.

**Symptom (silent miscompile):** `match` on `Option[str]` / `Result[str,E]` extracting the str payload returned a str with a CORRECT pointer half but a GARBAGE length half — `s.len` / `std.string.len(s)` read junk (e.g. `Some("found")` → `s.len` gave 108 instead of 5). `.unwrap()` worked fine (different, backend-intercept path), which masked it.

**Root cause:** `lower_match_tagged` (air_builder.ax ~2677) sets the payload `OP_DEREF` width from `find_variant_info`, which for a str variant reports the **RFC 0019 SYNTH struct** that WRAPS the str (single `_f0: str` field, `flags&1` set) — a type id ≠ str(12). The x86 selector only emits the 16-byte load for type_id==12, so the DEREF took the 8-byte path and dropped the length word. **Asymmetry with the constructor** (air_builder ~1656): Some/Ok/Err store the payload INLINE as a 16-byte str whenever the payload EXPRESSION is str (`store_t=12`). So the box holds a raw 16B str; extraction must read 16B too.

**Fix (frontend-only, air_builder, mirrors constructor):**
1. Unwrap a single-str-field SYNTH struct (`flags&1` + 1 field of type 12) payload back to `pl_type=12` → full 16-byte DEREF. Genuine multi-field / >8B user-struct payloads are heap-boxed 8-byte pointers and are NOT synth-flagged → left untouched (no false unwrap).
2. Also derive payload from a kind-11/12 Option/Result scrutinee directly (`extra`=Some/Ok inner, `name_id`=Result Err) since find_variant_info can't see kinds 11/12.

**Key diagnostic lesson:** the scrutinee here types as a **kind-6 SUM** (not kind-11 OPTION as first assumed) — verified by static-marker probes. `find_variant_info` returned a kind-1 STRUCT payload (`flags&1` synth). Two debug dead-ends: (a) `extern "C" printf` is NOT in ax_runtime.dll → uncommenting it makes the compiler fail to load (ENTRYPOINT_NOT_FOUND); (b) `ax_printf_local` called from `lower_match_tagged` segfaults the compiler. Use `ax_puts_local` (static-string, pub, the `[Debug]` printer) with distinct string markers, or a "force `pl_type=12`" binary experiment, instead of formatted tracing. See [[rfc0019-multifield-variant-shipped]], [[bug90-option-method-segfault-open]] (unwrap path), [[string-utf8-default]].

Oracle: `bin/t_optstrmatch.ax` (exit 10). Surfaced by autopilot bug-probe.
