# RFC 0020 — String Type System (UTF-8 `str` + `bytes` + `rune`, with room for other encodings)

Status: **RESOLVED** — core string type family complete (`str`/`bytes`/`rune` +
iteration + slicing + byte/codepoint view methods). `ascii` and `utf16`/`utf32`
transcoding split out to a future encodings/FFI follow-up (see §9, with rationale).
Author: autopilot
Related: RFC 0018 (for-in iteration; string iteration deferred here), [[string-utf8-default]]

## Implementation status (RESOLVED 2026-07-12)
- ✅ **P1a `rune`** (`edf515a`) — Unicode scalar, alias of u32.
- ✅ **P1b `bytes`** (`14114fe`) — distinct `TYPE_BYTES=22`, shares str's {ptr,len}
  16B repr; `.len`/`b[i]`→u8; zero-cost `s as bytes` / `b as str` reinterpret cast
  (lower_cast reuses the source register for str↔bytes).
- ✅ **P2a `for b in bytes`** (`76ee1f7`) — yields u8 (array-style OP_INDEX + runtime
  bytes.len bound).
- ✅ **P2b `for c in str`** (`13cff98`) — codepoint iteration yielding `rune`. stdlib
  `utf8_decode(s, off) -> i64` (packed `(width<<32)|codepoint`); `lower_for` resolves
  it via `resolve_op_method` (NOT a direct qualified call — that segfaults for str
  args, see BUG#93) and emits an `OP_CALL` per iteration; byte-offset counter advances
  by the variable width via a new `loop_incr_steps` stack (honored by the body-tail
  increment and `continue`). Verified: "AΩ"=2 codepoints, "日本語"=3 (not 9 bytes),
  ASCII "ABC" sum=198.
- ✅ **P3 slice `s[a..b]`** (`7fbb1ce`) — byte-indexed substring (Rust model): `str[a..b]`
  →str, `bytes[a..b]`→bytes, via a bounds-safe `std.string.slice` value-ABI call. See §3.4.
- ✅ **P3 str/bytes UFCS method-form** (`a382f31`) — `s.trim()`/`s.contains(x)`/chaining
  dispatch to `std.string` free fns whose param[0] is str/bytes.
- ✅ **P3 view methods** (`240e696`) — `s.as_bytes()` (byte view → `bytes`),
  `s.chars()` (codepoint view → runs `for c` as `rune`), `s.char_at(n)` (O(n) codepoint
  random access → `rune`). Pure stdlib, dispatched via the P3 UFCS path + the existing
  byte/codepoint for-in (which lowers the full iteree EXPRESSION, so `for b in
  s.as_bytes()` works). Named `as_bytes` (Rust's `str::as_bytes`): a bare `bytes()`
  method is impossible — `bytes` is a TYPE, so `s.bytes()`/`bytes(x)` parse as a
  conversion. Oracle `t_strviews` (exit 244).
- ⛔ **`ascii` type, `utf16`/`utf32` transcoding, `char_indices()`** — **deferred out of
  this RFC** to a future encodings/FFI follow-up. Rationale in §9. These are additive
  (no migration cost) and no current use case is blocked, so RFC 0020 closes with the
  core UTF-8 `str` + raw `bytes` + `rune` family complete.

## 1. Motivation

AXIOM's `str` is **UTF-8 by default**, and the language should **support other
string types** (byte buffers, ASCII, and interop encodings). Today there is a
single `str` type (`TYPE_STRING`, a `{ptr: ptr[u8], len: i64}` byte buffer) whose
operations are *byte*-oriented: `s[i]` yields a `u8` and `.len` is the **byte**
length. That is correct for byte processing (the self-hosting compiler's lexer and
parser read source as bytes and depend on this) but it is the *wrong* default for
user-facing string work:

- `for c in s` "wants" to yield **codepoints**, not bytes — a byte loop splits a
  multi-byte UTF-8 codepoint (`"héllo"` is 5 codepoints / 6 bytes).
- There is no distinct type for **raw bytes** vs **text**, so binary data and text
  share one type with ambiguous semantics.

The stdlib already hints at the intended model: `char_count` (codepoint count via
`str_char_count`), `is_valid_utf8`, `split → Vec[str]`. This RFC makes the model
explicit as a small, coherent **type family**, from which iteration/indexing
semantics fall out per type (the approach chosen over ad-hoc per-operation rules).

## 2. Non-negotiable constraint: do not break the self-host

The compiler's lexer/parser use `s[i] : u8` and byte-`.len` pervasively (22+ sites
in `lexer.ax` alone). Therefore **`str`'s existing byte-level `s[i]` and `.len`
MUST remain** (Rust makes the same choice: `str` indexing is byte-based, O(1); it
never silently does O(n) codepoint indexing). The type system is **additive**: new
types and a new *iteration* semantic, with existing byte operations untouched.

## 3. Design

### 3.1 The type family (all share the `{ptr, len}` 16-byte repr)

| Type    | Meaning                                  | `.len`      | `s[i]`    | `for x in s` yields |
|---------|------------------------------------------|-------------|-----------|---------------------|
| `str`   | UTF-8 text (default). Validated on entry.| byte length | `u8` (byte, O(1)) | `rune` (codepoint) |
| `bytes` | Raw `[u8]` buffer / binary data          | byte length | `u8`      | `u8`                |
| `ascii` | 7-bit ASCII invariant (1 byte = 1 char)  | char length | `char8`   | `char8`             |

Key property: `str`, `bytes`, and `ascii` have the **same runtime representation**
(`{ptr: ptr[u8], len: i64}`, 16-byte inline — the existing `TYPE_STRING` layout).
They differ only in **static type** and the **operations/semantics** attached, so:

- Conversions are mostly **zero-cost views** (no copy): `str.bytes() → bytes`,
  `str.ascii()` (checked), `ascii.as_str() → str` (free — ASCII ⊂ UTF-8).
- `bytes.to_str() → Result[str, Utf8Error]` validates (the one non-free direction).

`char8` (`TYPE_CHAR8`, existing) stays as a single-byte character. This RFC adds:

### 3.2 `rune` — a Unicode scalar value

`rune` = a Unicode scalar (`u32`, range `0..=0x10FFFF`, excluding surrogates). It is
what `for c in str` yields and what a codepoint-aware API returns. Represented as a
`u32` (zero new ABI surface — it is a scalar). Literals: a `'x'` char literal in a
`rune` context is its codepoint; `'\u{1F600}'` for non-ASCII.

(Naming: `rune` chosen over overloading `char`, because the existing `char8` is a
byte; `rune` unambiguously names a codepoint. A `char` alias for `rune` can be added
later if desired.)

### 3.3 Iteration semantics fall out of the type (RFC 0018 P2-string)

- `for c in str`   → decode UTF-8, `c : rune`   (the correct default)
- `for b in bytes` → `b : u8`
- `for c in ascii` → `c : char8`
- Explicit views regardless of the base type:
  - `s.bytes()`  → iterate/index bytes (`u8`)
  - `s.chars()`  → iterate codepoints (`rune`) — same as `for c in str`
  - `s.char_indices()` → `(byte_offset, rune)` pairs (for slicing)

This is why the type system comes **first**: once `str`/`bytes`/`ascii` exist,
`for` iteration is a direct per-type lowering (byte index loop for `bytes`/`ascii`,
UTF-8 decode loop for `str`), reusing the RFC 0018 P2 field-load+index machinery.

### 3.4 Indexing & slicing

- `s[i]` on `str`/`bytes` = **byte** access `u8`, O(1) (unchanged; compiler relies
  on it). Documented as raw byte access.
- Codepoint access is explicit: `s.chars()`, `s.char_at(n)` (O(n)), or slicing on
  `char_indices()`. `str` slicing `s[a..b]` (once RFC 0018/slice support lands) is by
  **byte** offset and must land on codepoint boundaries (checked → `Result` or panic).

### 3.5 UTF-8 validity

`str` carries the **invariant** that its bytes are valid UTF-8. Entry points:
- String literals: validated at compile time (a literal is known bytes).
- `bytes.to_str()`: runtime validation → `Result[str, Utf8Error]`.
- FFI/`ptr[u8]` → `str`: an `unsafe`/checked constructor.
`is_valid_utf8` (exists) backs the checks; `str_char_count` (exists) backs `.chars()`.

## 4. Alternatives considered

1. **One `str`, byte semantics everywhere (status quo).** Rejected: `for c in s`
   silently mis-iterates UTF-8; no type distinction for binary data.
2. **`str` = codepoint-indexed (Swift-like `String[i]` opaque).** Rejected: would
   break the compiler's O(1) byte `s[i]`, and O(n) indexing is a performance
   footgun. Byte-indexing + explicit codepoint views (Rust/Go model) is safer.
3. **Grapheme-cluster iteration by default (Swift).** Rejected for now: needs a
   Unicode segmentation table (large, versioned). `rune`/codepoint is the right
   primitive; graphemes can be a stdlib layer later.
4. **`bytes` as `Vec[u8]` instead of a distinct view type.** Rejected: `bytes`
   should be a zero-copy view sharing the `str` repr so conversions are free; a
   `Vec[u8]` is an owning growable buffer (different role — keep both).

## 5. Drawbacks

- Adds types (`bytes`, `ascii`, `rune`) and view methods — more surface area.
- `str` indexing being *byte*-based while iteration is *codepoint*-based is a subtle
  asymmetry (same as Rust/Go; documented, and the safe choice given the constraint).
- UTF-8 decode in `for c in str` is variable-width (not a fixed stride like arrays),
  so its lowering is more involved than RFC 0018 P2's Vec case.

## 6. Migration / compatibility

- **Fully additive.** `str`'s existing byte `s[i]`/`.len` are unchanged, so the
  self-hosting compiler and all current programs compile identically (A==B expected
  for the type-introduction phase — no self-codegen change until the compiler opts
  into new APIs).
- Existing `char8` unchanged. `rune`/`bytes`/`ascii` are new names (no collision;
  verified `bytes`/`ascii`/`rune` are not currently types).
- `for c in str` is currently *rejected* (RFC 0018 P2 only did arrays + Vec), so
  enabling codepoint iteration only turns a reject into working code — no behavior
  change for valid programs.

## 7. Phased implementation plan (each phase gated; implement only after approval)

- **P1 — types + conversions (frontend, additive).** Register `bytes`, `ascii`
  (repr = `TYPE_STRING`), and `rune` (`u32`). Parser/typecheck recognize the type
  names. Zero-cost view methods `str.bytes()` / `ascii.as_str()`; checked
  `bytes.to_str() → Result`. No codegen change to existing programs ⇒ A==B.
- **P2 — `for c in str` codepoint iteration.** UTF-8 decode loop in `lower_for`
  (advance by the leading-byte width; yield a `rune`), reusing the RFC 0018 P2
  field-load+index scaffolding plus a decode step. `for b in bytes` reuses the
  byte-index path. Oracle: `for c in "héllo"` counts 5, sums codepoints.
- **P3 — `ascii` ops + other encodings.** `ascii` validation + 1:1 iteration;
  `utf16`/`utf32` interop types (FFI). Grapheme segmentation as a later stdlib layer.

## 8. Test / verification strategy

- P1: oracle constructing `bytes` from a `str`, round-tripping, and a
  `bytes.to_str()` that rejects invalid UTF-8 (a `reject`/`Result::Err` oracle).
- P2: `for c in str` over an ASCII string (byte==codepoint) AND a multi-byte string
  (`"héllo"` / `"日本"`) — codepoint count and sum must differ from the byte loop.
- Every phase: fast fixpoint (A==B while the compiler source doesn't use new APIs)
  + full regression; backend touches (P2 decode) require the B==C gate per the
  fixpoint-async rule.

## 9. Open questions — RESOLVED (2026-07-12)

- **`rune` vs `char`:** ✅ RESOLVED — keep `char8` (the single-byte character) and
  `rune` (the codepoint). No rename. `rune` unambiguously names a Unicode scalar and
  does not disturb the byte-oriented `char8`/`u8` the lexer/parser rely on. A `char`
  alias for `rune` may be added later if desired; not needed now.
- **`ascii` in P1 or defer?** ✅ RESOLVED — **deferred (unscheduled).** `bytes` already
  provides raw `u8` access plus a zero-cost `s as bytes` / `s.as_bytes()` view, so an
  `ascii` type would add a validation invariant + a `char8` element that largely
  duplicates `bytes` iteration for marginal value, and it costs new parser/typecheck
  builtin-type surface. No current use case is blocked. If a concrete ASCII-fast-path
  need arises it is purely additive → a small follow-up RFC. Not part of RFC 0020.
- **Invalid-UTF-8 policy for `str[a..b]` on a non-boundary:** ✅ RESOLVED — **byte-indexed
  with clamp (Rust model), no boundary check** (shipped `7fbb1ce`). `s[i]` and `s[a..b]`
  are documented **raw byte** operations (§3.4); the slice is bounds-safe (never OOB) but
  may split a codepoint if misused — the same trade-off Rust/Go accept for O(1) byte
  indexing. A checked codepoint-boundary slice returning `Result` is a possible additive
  future helper (`str.slice_chars` / `char_indices`) but is **not** required for the core
  and is not blocking; deferred with the `char_indices()` iterator work below.
- **Other string types (`utf16`/`utf32`):** ✅ RESOLVED — **deferred to a future
  encodings/FFI RFC.** These are not mere type tags: they require real transcoding
  routines (UTF-8 ↔ UTF-16/UTF-32) and, for `utf16`, Windows wide-char FFI conventions —
  a distinct subsystem from this type family. A user can already hold raw UTF-16/32 in
  `bytes` and transcode via FFI today; typed, safe transcoding is the follow-up. Priority
  when scheduled: `utf16` (Windows FFI) first, then `utf32`.
- **`char_indices()` / iterator views:** ✅ RESOLVED — **deferred.** A `(byte_offset, rune)`
  pair-yielding iterator needs a tuple-in-`for` / iterator-protocol design the language
  does not yet have. `char_at(n)` (O(n) codepoint access) + byte-indexed slicing cover the
  practical need now; the richer iterator surface belongs with a future iterator-protocol RFC.

**Conclusion:** RFC 0020 is **RESOLVED**. The coherent core — UTF-8 `str` (byte index,
codepoint iteration), raw `bytes`, `rune`, slicing, and `.as_bytes()`/`.chars()`/
`.char_at()` views — is shipped and gated (A==B). The deferred items (`ascii`, `utf16`/
`utf32` transcoding, `char_indices` iterator) are additive and carry no migration cost, so
they are cleanly split into a future encodings/iterator follow-up rather than blocking closure.
