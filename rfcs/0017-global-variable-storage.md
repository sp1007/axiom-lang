# RFC 0017 — Module-level global variable storage & initialization semantics

Status: **DRAFT** (2026-07-09). Fixes: BUG#82 (module-level `let`/`mut` globals miscompile).

---

## 1. Problem — module-level mutable globals have NO backing storage

A module-level `let`/`mut` variable (not `const`) is accepted by the front end but the
backend gives it **no storage at all**. Confirmed 2026-07-09 via `dump-air`:

```axiom
mut g := 0
fn setit(x: i32): g = x
fn getit() -> i32: return g
fn main() -> i32: setit(42); return getit()   // returns 42 — BUT BY LUCK
```

```
fn @2(t3) -> t14:               ; setit
    %1: t3 = copy %1            ; `g = x` lowered to a self-copy no-op (no store)
    ret
fn @5() -> t3:                  ; getit
    ret                         ; `return g` — empty; returns whatever is in RAX
fn @6() -> t3:                  ; main
    %1 = iconst %42
    %2 = call                  ; setit(42)  -> leaves 42 in RAX
    %3 = call %2               ; getit()    -> returns stale RAX == 42
    ret %3
```

The earlier note (memory `bug82-global-var-semantics-open`) claimed *"plain set/get
works"*. **That was a false positive**: `getit` returns a stale register, which happens
to equal the value the caller just passed. There is no shared storage. The three observed
symptoms are all one root cause — globals are not materialized:

- **Non-zero initializer never runs** — `mut g := 7; return g` → `0`. The module-level
  decl is never lowered; the initializer expression is dropped.
- **Cross-function read-modify-write miscompiles** — `counter = counter + 1` in a function
  reads/writes a per-function vreg that does not persist; two `bump()` calls return
  garbage (`4` instead of `1+2=3`).
- **"Plain set/get works"** — coincidental register leftover, not storage.

`lower_ident` (`air_builder.ax:691`) resolves a `SYM_VAR` via
`self.locals.local_map_get(sym_idx)` — a **per-function** vreg map. A module-level symbol
has no entry in any function's map, so a global read yields a fresh/zero vreg and a global
write updates a vreg that dies at function return.

### Why the self-hosted compiler still reaches a fixpoint

The entire compiler + stdlib source contains **exactly one** module-level mutable global:
`lsp.ax:15  mut g_documents: u64 = 0`. It is (a) zero-initialized (so the missing
initializer is harmless) and (b) only touched by the `lsp` command, which is **never
exercised during `build`**. So self-compilation never depends on global storage working,
and the bug stays latent. `pub const` (used pervasively in `air.ax`) is unaffected: consts
are inlined at each use site (the `SYM_CONST` path in `air_builder.lower_ident`), not
runtime storage.

---

## 2. Storage model

A module-level global needs **static, process-lifetime, writable storage** addressed
identically from every function. The backend already has the two mechanisms required:

1. **RIP-relative addressing + relocations** — `OP_FUNC_ADDR` emits `lea reg,[rip+sym]` +
   `RELOC_PC32` to a code symbol (`x86_emitter.ax:188`); string literals emit the same
   PC32 form to a data symbol in section 2 (`x86_emitter.ax:203`). Both are ASLR-safe and
   already resolved by the custom linker.
2. **A per-object data section** — string literals live in the COFF `.rdata` section
   (`x86_coff.ax` section 2). **But `.rdata` is not writable**: the linker *merges* each
   object's rdata into the read-execute `.text` image (`linker.ax:1920-1926`,
   `0x60000020`). Writing a global there would fault.

Therefore globals require a **new, genuinely writable section** in both the object file and
the final image.

### Decision — writable `.data` section (COFF object) merged into the existing RW image region

- **Object (COFF)**: add a 3rd section `.data`, characteristics
  `IMAGE_SCN_CNT_INITIALIZED_DATA | READ | WRITE` (`0xC0000040`). Each global is emitted at
  a fixed offset with its constant-folded initial bytes (zero-filled when the initializer is
  zero or absent). A COFF symbol (section 3) is registered per global at its offset.
- **Image (PE)**: the linker already emits exactly one **writable** section — `.idata`
  (`0xC0000040`), built after code layout. Phase 1 appends the merged global bytes to the
  **end** of that writable section (after all import structures) and resolves global symbols
  to `idata_rva + globals_region_start + sym.offset`. This reuses the proven writable region
  and — critically — does **not** perturb any existing code/rdata/import offset math (globals
  are appended last, and the section's raw size is read after the append). A brand-new,
  dedicated `.data` PE section is a Phase 2 refinement; the storage semantics are identical.
  (ELF path: deferred — see §4; the self-host target is win64/COFF.)
- **Addressing**: a new AIR op `OP_GLOBAL_ADDR` produces the global's address via
  `lea reg,[rip+gsym]` + `RELOC_PC32` — structurally identical to `OP_FUNC_ADDR`, only the
  target symbol lives in the writable data region instead of `.text`. Reads compose
  `OP_GLOBAL_ADDR` + `OP_LOAD`; writes compose `OP_GLOBAL_ADDR` + `OP_STORE`. Because every
  linked AXIOM executable imports `ax_runtime`/`kernel32` (runtime init + `ExitProcess`), the
  writable `.idata` section is always present, so a global always has a home.

### Critical safety invariant (fixpoint preservation)

> **When a compilation unit declares zero globals, the emitted object and image are
> byte-for-byte identical to today.**

The `.data` section is emitted only when `globals.len > 0`. This guarantees that every
existing test and every source file *without* globals is unaffected. The compiler's own
source has one global (`g_documents`), so the self-built compiler binary *will* change
(it gains a `.data` section holding one 8-byte zero slot + one symbol) — the expected
`A != B` backend-transition; the gate is the hand-built `B == C`. Self-compilation
*behavior* is unchanged because `build` never reads or writes `g_documents`.

---

## 3. Initialization semantics

- **Constant scalar initializers** (integer / bool / char / float literals, and
  compile-time-constant expressions) are folded into the `.data` section's initial bytes.
  This fixes the "non-zero initializer never runs" symptom deterministically with **no
  runtime init sequence** — the value is present in the image at load time. Covers
  `mut g := 7`, `mut g: u64 = 0`, etc.
- **Zero / absent initializer** → zero bytes (still emitted in `.data`; a `.bss`
  optimization is deferred — determinism and simplicity first).
- **Non-constant initializers** (`mut g := compute()`, `mut v := Vec.new()`): require a
  synthetic init routine run before `main`. **Out of scope for Phase 1** — rejected with a
  clear diagnostic (per the BUG#53 "reject, don't silently miscompile" convention). The
  compiler itself uses none, so this rejection cannot break self-build.

## 4. Scope

### Phase 1 (this RFC, shippable)
- Scalar globals with size ≤ 8 bytes (`i8..i64`, `u8..u64`, `bool`, `char`, `f32`, `f64`,
  `ptr[T]`). Aggregate/struct/array globals → rejected (deferred to Phase 2).
- Constant scalar initializers folded into `.data`; zero otherwise.
- `OP_GLOBAL_ADDR` + read/write lowering.
- COFF `.data` section + per-global symbol; PE `.data` section + PC32 fixups.
- Non-constant initializer → diagnostic reject.
- Oracles: non-zero init, cross-fn RMW counter, multiple globals, mixed types.

### Phase 2 (future)
- Aggregate/struct/array globals.
- Non-constant initializers via a synthetic `__axiom_global_init` called at `main` entry.
- `.bss` for zero-initialized globals (image-size optimization).
- ELF `.data` parity hardening (Phase 1 mirrors COFF; ELF exercised only under `--target`).

## 5. Invariants & verification

- **Determinism**: global layout is assigned in source-declaration order; offsets and
  symbol names are a pure function of the program. No hashing, no address-dependent output.
- **Isolation**: no change to the lexer/parser/typechecker semantics; this is purely a
  lowering + backend-emission addition. `const` handling untouched.
- **Gate**: fast fixpoint (`B == C` — backend change), full regression
  (`regression_repros.sh`), plus new global oracles. Self-build must still produce a working
  compiler that passes the whole suite.

## 6. Alternatives considered

- **Reject all module-level mutable globals** (BUG#53 style): rejected — the compiler
  declares `g_documents`, so a blanket reject breaks self-build. Only *non-constant
  initializers* are safe to reject.
- **Reuse `.rdata`**: rejected — merged into read-execute `.text` by the linker; writes
  fault.
- **Runtime-allocated globals block**: rejected — needs a root pointer, which is itself a
  global (circular), plus hidden allocation (violates §11 "no hidden allocations").

## 7. Migration / compatibility

No source changes required. Programs that previously *appeared* to work by register luck
now work by real storage; programs that relied on the buggy behavior did not exist (the
behavior was garbage). Non-constant module-level initializers that were silently
miscompiled now produce a diagnostic — a strict improvement.
