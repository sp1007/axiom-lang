# RFC 0032 — Position-independent codegen + ET_DYN for ELF shared objects

Status: **CLOSED / IMPLEMENTED** — P1 (ET_DYN + `.dynsym` exports, linker-only), P2 (RIP-relative globals, verified), P3 (`OP_SPAWN` → RIP-relative). Behind `--target linux --shared`.
Author: autopilot
Related: RFC 0009 (FFI dynamic linking — P3 ELF export), RFC 0029 (interface vtable / `OP_FUNC_ADDR`),
RFC 0030 (`.bss`), [[session-state-2026-07-24d]]

## 1. Motivation

RFC 0009 P3 (produce a consumable ELF `.so`) is blocked. Live evidence (2026-07-24d):
`axc build tests/ffi/axmath.ax --shared -o bin/axmath.so --target linux -self-link -O1`
produces an ELF with:

- `Type: EXEC` (ET_EXEC) — a loadable-anywhere library needs **ET_DYN**.
- **No `.dynsym` exports** — `ax_add`/`ax_mul` are absent, so `dlsym` cannot find them.
- **No relocation table** — the self-linker resolves every address at link time and bakes
  absolute `R_X86_64_64` immediates straight into `.text`.

Consequence: a `dlopen`'d `.so` is relocated to an arbitrary base by the dynamic loader, but
every baked absolute address (global-variable addresses, taken-function addresses / RFC 0029
vtable slots) still points at the original fixed load address ⇒ wild reads/calls ⇒ crash.

The COFF/PE side of RFC 0009 (Windows DLL export + consumption) already works. ELF `.so` is
the only remaining gap in the "produce a shared library" success criterion.

## 2. Chosen approach (user decision, 2026-07-24)

**PIC codegen + ET_DYN.** Change x86 instruction selection so that, *on the ELF shared-object
path only*, global-variable and taken-function addresses are formed **RIP-relative**
(`lea reg, [rip + disp32]`) rather than as absolute 64-bit immediates, and emit the library as
**ET_DYN** with exported symbols in `.dynsym`. RIP-relative addressing is inherently
position-independent: the CPU computes the target from the current instruction pointer, so the
loader's base offset applies automatically with **no text relocations**. This is the clean,
long-term-correct answer (what real toolchains emit for `.so`), as opposed to the alternative
of keeping absolute codegen + a full `R_X86_64_RELATIVE` table with `DF_TEXTREL` (writable
text, glibc-warned).

## 3. Design (to be finalized from the codegen/linker surface map)

### 3.0 Binary-level target (measured on current `bin/axmath.so`, 2026-07-24d)

`readelf` on the existing `--shared --target linux` output shows the import-side dynamic
infrastructure is **already present** — export is an extension of it, not new machinery:

- `e_type` = `02 00` at **byte offset 16** of the ELF header (ET_EXEC). P1 flips this to
  `03 00` (ET_DYN) **only** on the shared-object path.
- Program headers (3): INTERP (`/lib64/ld-linux-x86-64.so.2`), one RWE LOAD at vaddr
  `0x400000`, DYNAMIC. **No section header table** (`e_shnum = 0`) — the loader uses the
  DYNAMIC segment, so exports must be reachable purely via `DT_HASH`/`DT_SYMTAB`, not `.shdr`.
- DYNAMIC (16 entries) already has: `NEEDED libc.so.6`, `HASH`, `STRTAB`, `SYMTAB` (SYMENT 24),
  `RELA`/`RELASZ`/`RELAENT`, `PLTGOT`, `JMPREL`, `PLTRELSZ`, `BIND_NOW`, `FLAGS=BIND_NOW`.
  ⇒ `DT_SYMTAB`/`DT_STRTAB`/`DT_HASH` exist for **imports**; **export = add defined
  `STB_GLOBAL`/`STT_FUNC` entries (st_value = function RVA) + rebuild `DT_HASH` chains** so
  `dlsym` resolves them.
- The dynamic loader adjusts `DT_*` vaddrs and processes `RELA` by the load bias automatically
  under ET_DYN, so the **import** side survives relocation without codegen changes.

**Therefore P1 (ET_DYN + export wiring) is independently valuable and shippable:** a `.so`
whose exported functions are **leaves with no internal global/func-addr references** (e.g.
`axmath`'s `ax_add`/`ax_mul` = pure arithmetic) is fully dlopen-able and callable after P1,
because there is nothing absolute in their bodies to relocate. PIC codegen (P2/P3) is required
only when an exported function references a module **global** or a **taken-function address**
(RFC 0029 vtable) — those are the only absolute-in-`.text` sites that break under relocation.

### 3.1 Codegen is ALREADY position-independent (ground-truthed 2026-07-24d)

The "absolute imm baked in `.text`" premise was **largely wrong**. `objdump`/`readelf` on the
current `axmath.so` + a source read of the selector/emitter found:

- **Globals** (`OP_GLOBAL_ADDR`, air.ax:41) → `x86_selector.ax:1861-1866` `vreg=4` →
  `x86_emitter.ax:210-223` emits `lea reg,[rip+disp32]` + `RELOC_PC32`. **Already RIP-relative.**
- **Taken-function addresses** (`OP_FUNC_ADDR`, air.ax:90) → `x86_selector.ax:1853-1859` `vreg=3`
  → `x86_emitter.ax:195-209` `lea reg,[rip+disp32]` + `RELOC_PC32`. **Already RIP-relative.**
  RFC 0029 vtables are heap boxes filled via `OP_FUNC_ADDR` ⇒ inherit RIP-relative; RFC 0028
  jump tables lower to a compare tree (no address table). Both PIC-safe.
- **Calls** → `call rel32` (`x86_emitter.ax:494-535`); linker resolves PC-relative
  (`linker.ax:3084-3087`, `val = target_va - pc + addend`) where the base **cancels**.
- Measured: 110 `(%rip)` operands; 339 `movabs` but only **ONE** is an image address — the
  `OP_SPAWN` thread-entry thunk in bundled runtime (`x86_emitter.ax:184-194`, `RELOC_ABS64`),
  which no leaf export reaches. The other 338 `movabs` are float/constant bit patterns.

⇒ **A leaf `.so`'s `.text` is already position-independent.** The gap is purely the container:
`ET_EXEC`→`ET_DYN` + a `.dynsym` export table. **P1 is a linker-only change; no codegen touch**,
so the Windows/COFF self-build is byte-identical by construction (the shared-ELF path is never
taken when the compiler compiles itself).

### 3.2 P1 edit sites (linker.ax)

- **e_type:** `linker.ax:921` `push_u16_le(out, 2)` → `3` gated on `self.is_shared && elf64`.
- **Exports:** the ELF dynamic block (`linker.ax:2620-2822`, currently imports-only — every
  `.dynsym` entry is `SHN_UNDEF`) must (a) fire even when `thunk_count==0` (a pure leaf lib has
  no imports), and (b) append **defined** entries (`STB_GLOBAL|STT_FUNC`, `st_shndx`=text,
  `st_value`=symbol RVA) for each `self.export_names` (mangled `ax_<name>`, mirroring the COFF
  `build_export_edata` at `linker.ax:1043-1161`), extending `.dynstr`/`.dynsym`/`.hash` (builders
  at `linker.ax:2684-2713`) and fixing the `.hash` `nchain`/bucket chain so `dlsym` resolves them.
- **Base handling — HYPOTHESIS to test minimal-first:** because *all* code is RIP-relative, ET_DYN
  may work with the base kept at `0x400000` (`linker.ax:959/1001`, base_addr `:2067-2072`): glibc
  adds the load bias `l_addr` uniformly to `DT_*` pointers, symbol `st_value`, and RELA relocs, and
  RIP-relative code is bias-independent. So P1 tries **e_type + exports only** and runs the oracle
  under ASLR. Only if that faults do we do the base-0 rewrite (rebase every VA site off 0).
- **Cosmetic (optional in P1):** zero `e_entry` for `is_shared` (`linker.ax:3188` currently yields a
  bogus `0x1003ff200` no-`main` artifact — harmless for `dlopen`); PT_INTERP can stay (ignored by
  `dlopen`).

### 3.3 P2 / P3 status

- **P2 (globals):** codegen already RIP-relative (§3.1) ⇒ **verification only** — a `.so` exporting
  `bump()` over a `mut global` returns 1,2,3 across dlopen calls under ASLR. Code change expected
  = none (any fault would be P1's base rewrite, not the `lea`).
- **P3 (func-addr/vtable):** already RIP-relative (§3.1). **Residual = the single `OP_SPAWN` ABS64**
  (`x86_selector.ax:1874` → `x86_emitter.ax:184-194`). This site is **target-independent**, so
  rewriting it to RIP-relative would perturb the COFF self-build ⇒ **requires the B==C backend
  fixpoint gate**. Deferrable: ship P1/P2 with a documented caveat that an `#[export]` function
  transitively using `spawn` is unsafe in a rebased `.so`.

- **3.1 ET_DYN header** — write `e_type = ET_DYN (3)`; make PT_LOAD `p_vaddr` base-relative
  (starting at 0) and the entry point base-relative; keep alignment.
- **3.2 `.dynsym` exports** — for each `#[export] pub fn`, add a `.dynsym`/`.dynstr` entry (and
  `.hash`/`.gnu.hash` as required) with `STB_GLOBAL`/`STT_FUNC` and the correct `st_value` RVA.
  Mirror the COFF export-directory path that already works.
- **3.3 RIP-relative global addressing** — replace absolute `mov reg, imm64` for
  `OP_GLOBAL_ADDR` with `lea reg, [rip + disp32]`, `disp32 = target_rva - (insn_end_rva)`,
  computed at link time. Requires the linker to know each such fixup site's own address.
- **3.4 RIP-relative function addressing** — same transform for `OP_FUNC_ADDR` (fn-ptr,
  lambda, RFC 0029 vtable slots).
- **3.5 Calls** — if intra-module calls are already `call rel32` (PC-relative), no change; to
  confirm from the map.
- **3.6 disp32 range** — RIP-relative displacement is signed 32-bit (±2 GB); confirm the
  section layout keeps code↔data within range (it does at current sizes).

## 4. Alternatives

- **A. RELATIVE-reloc table + ET_DYN (rejected).** Keep absolute codegen; have the linker
  retain every fixup site and emit one `R_X86_64_RELATIVE` per absolute address, set `DF_TEXTREL`.
  Tractable but leaves writable-text relocations (glibc warns, slower load, larger dirty pages)
  and still touches the heaviest-loaded linker function. Inferior long-term.
- **B. Narrower in-process export (rejected).** `ET_EXEC` + `dlopen(NULL)` + `dlsym` within the
  same process. Does not satisfy the cross-module `.so` use case.
- **C. Defer (rejected).** Leaves the RFC 0009 success criterion unmet.

## 5. Drawbacks

- Touches x86 instruction selection — the single most self-host-sensitive subsystem. Must be
  **strictly gated** to `--target linux --shared` so the Windows/COFF self-build stays
  byte-identical (fixpoint A==B on the daily driver).
- ET_DYN + `.dynsym` widen the ELF writer; more surface to keep deterministic.
- RIP-relative disp32 assumes code and referenced data stay within ±2 GB (true today).

## 6. Migration / compatibility

- Purely additive: the COFF path, the ELF **executable** path (`--target linux` without
  `--shared`), and all existing regression/oracle behavior are unchanged. Only the
  `--shared --target linux` output format changes (ET_EXEC → ET_DYN, gains `.dynsym` + PIC).
- No source-language change ⇒ no user migration.

## 7. Verification plan (each phase gated)

- **Self-host:** every phase must keep the Windows self-build **A==B byte-identical** (PIC code
  is emitted only on the gated ELF-shared path, which the self-build never takes). This is the
  non-negotiable guard.
- **P1 — ET_DYN + `.dynsym` exports.** `axmath.so` (leaf `ax_add`/`ax_mul`, no globals/func-addrs)
  builds as ET_DYN, `readelf -h` shows `Type: DYN`, `readelf --dyn-syms` lists `ax_add`/`ax_mul`.
  Oracle: a WSL C host `dlopen("axmath.so")` + `dlsym("ax_add")` + call → `2+3=5`.
- **P2 — RIP-relative globals.** A `.so` that reads/writes a module global, dlopen'd, returns
  the right value (proves base-relative global access survives relocation).
- **P3 — RIP-relative func-addr / vtable.** A `.so` exporting a function that dispatches through
  a fn-ptr / RFC 0029 interface vtable, dlopen'd, returns the right value.
- Full regression + `elf_linux_check.sh` stay GREEN throughout (ELF **executable** path
  untouched).
- Because this is codegen + linker/ABI, the standard COFF gate is A==B (inert on self-host); the
  ELF-shared artifacts are validated by the dlopen oracles above, run under WSL.

## 7b. P1 result (SHIPPED 2026-07-24)

Linker-only change (`linker.ax`), gated `self.is_shared`:
1. `e_type` → **ET_DYN** for a shared object (`linker_build_elf_headers`, new `is_shared` param).
2. **`PT_GNU_STACK`** (PF_R|PF_W) replaces `PT_INTERP` for a `.so` — without it glibc assumes the
   object wants an executable stack and `dlopen` fails with *"cannot enable executable stack as
   shared object requires: Invalid argument"*. A dlopen'd `.so` needs no interpreter, so dropping
   PT_INTERP keeps the phdr count at 3 and the 512-byte header layout intact.
3. **Exports** appended to `.dynsym` (defined, `STB_GLOBAL|STT_FUNC`, `st_shndx=1` so ld.so adds
   the load bias to `st_value`), names to `.dynstr`, and the single-bucket `.hash` chain extended
   to cover them (`nchain` = imports + exports). Resolution mirrors the COFF `build_export_edata`:
   exported name = `self.export_names[k]`, defined address = RVA of mangled `ax_<name>`.

**Verified** (`scripts/so_export_check.sh`, WSL + python3 ctypes, no gcc needed): `axmath.so` is
`Type: DYN`; `nm -D` shows `T ax_add` / `T ax_mul`; `dlopen` + `dlsym` + call returns
`ax_add(2,3)=5`, `ax_mul(4,5)=20` under ASLR. Gate: fast fixpoint **A==B `0FDCAE49`** (gated change
inert on the COFF self-build), regression **528/528**, `elf_linux_check` **12/12** (ELF executable
path untouched). ⚠️ `readelf --dyn-syms` prints nothing for this file — it needs a section-header
table, which AXIOM does not emit; `nm -D` (dynamic segment) is the correct tool. The dynamic block
still requires `thunk_count > 0` (any bundled-stdlib program imports libc, so this holds in
practice); a truly import-less leaf export is a follow-up.

## 7c. P3 result (SHIPPED 2026-07-24)

The sole image-absolute bake — `OP_SPAWN`'s thread-entry address — was `movabs reg, ABS64(sym)`
(`x86_emitter.ax`, `vreg==1`, the ONE grep-confirmed site). Changed to `lea reg,[rip+disp32]` +
`RELOC_PC32(sym)`, byte-for-byte the same encoding as the `vreg==3` `OP_FUNC_ADDR` path. Same
symbol reference (`sym_name: imm`), just PC-relative — so an `#[export]` function that transitively
`spawn`s no longer bakes a link-time absolute that would be a wild call once the `.so` is dlopen'd
at a random base.

The change is **target-independent** (COFF + ELF), so it perturbs the self-build's spawn bytes ⇒
gated by the **B==C** three-hop fixpoint, not A==B. Gate: **A≠B (expected), B==C `0D46A5C5`**;
regression **529/529** (added `bin/t_spawnsmoke.ax` — the suite exercised spawn NOWHERE before, a
real coverage gap); `so_export_check`/`elf_linux_check` still green.

**Validation basis (transparent):** proven **non-regressing** (B==C + 529/529 + byte-identical exit
behavior old-vs-new across several COFF spawn programs, no crash), and **correct-by-construction on a
directly-verified foundation** — spawn now forms the handler pointer with the exact `lea`+PC32
mechanism that P2's `soprobe` already proved resolves correctly (taken-function addresses via
intra-`.so` calls) in a rebased `.so`. A direct spawn-in-rebased-`.so` actor-dispatch oracle was NOT
built: AXIOM's `spawn` is actor-model (the handler runs on `ax_actor_step` message dispatch, not on
spawn), so observing it end-to-end needs actor-message infrastructure inside a `.so` — disproportionate
for a single-site change resting on the P2-verified mechanism.

## 8. Open questions

- `.gnu.hash` vs classic `.hash`: minimum the glibc loader accepts for `dlsym` on our `.so`.
- Do we need `PT_DYNAMIC`/`DT_SYMTAB`/`DT_STRTAB`/`DT_HASH` entries wired for a bare exported
  leaf, or only once imports are also present? (P1 will answer empirically.)
- Interaction with RFC 0031 `-dfe` pruning on the `.so` path (export symbols must be DFE roots —
  already handled for `--shared`; re-verify under ET_DYN).
