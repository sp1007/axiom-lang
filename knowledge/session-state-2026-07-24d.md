---
name: session-state-2026-07-24d
description: "STATE 2026-07-24d — #1 ELF re-verify under CTGC default-on = 12/12 GREEN (DONE). #2 ELF .so export re-confirmed HARD BLOCKER by live readelf: --shared --target linux emits ET_EXEC, no .dynsym exports, no relocs. Cross-module .so needs ET_DYN + PIC-or-RELATIVE-reloc-table = RFC-scale; product decision pending."
metadata:
  node_type: memory
  type: project
---

**ĐỌC ĐẦU TIÊN.** Driver `bin/axc_native.exe` = CTGC default-on (`AF899BDF` lineage,
recognizes `-no-ctgc-free`). Continues [[session-state-2026-07-24c]].

## #1 — ELF re-verify under CTGC default-on → DONE (12/12 GREEN)
`bash scripts/elf_linux_check.sh` = **12 passed, 0 failed → ELF_LINUX_OK** under the
default-on CTGC compiler. Includes `elfvec`/`elfmap`/`elfsmap` (Vec/HashMap aggregates =
exactly what CTGC compile-time free targets) + `t_bssglobal`. Confirms the flip is safe on
the ELF target too. Verification-only, no source change → no commit. Closes the "ELF path not
re-verified under default-on" NOT-done item from 24c.

## #2 — ELF .so export (RFC 0009 P3) → re-confirmed HARD BLOCKER (live evidence)
Built `bin/axmath.so` via `axc build tests/ffi/axmath.ax --shared -o bin/axmath.so
--target linux -self-link -O1` (note: **source arg MUST come before `--shared`** or the
driver treats `--shared` as the source file). `readelf` (WSL, `MSYS2_ARG_CONV_EXCL="*"`):
- **`Type: EXEC`** (ET_EXEC) — a real `.so` needs **ET_DYN**.
- Fixed entry `0x1003ff200`, absolute addressing baked in.
- **No dynamic symbols** — `ax_add`/`ax_mul` NOT in `.dynsym` (not exported).
- **No relocations** at all — the self-linker resolves every address at link time and bakes
  absolute `R_X86_64_64` values directly into `.text`; there is no residual `.rela.dyn`.
⇒ A `dlopen`'d `.so` relocated to an arbitrary base has every baked absolute address wrong
(globals + `OP_FUNC_ADDR` vtable slots). Fixing = **ET_DYN + PIC codegen OR a full
`R_X86_64_RELATIVE` table**, retaining the fixup-site list through the heaviest-loaded
linker function (§16 = incremental + RFC). This is RFC-scale + a ~2h B==C fixpoint; per
CLAUDE.md §13 a linker/ABI change of this scale REQUIRES an RFC. **Product decision pending
from user** (which architectural path / whether to greenlight). Recorded so no session burns
the dead-end path. See commit `1072129` for the same diagnosis.

## #2 RESOLVED via RFC 0032 P1 — real cross-module ELF `.so` now ships (linker-only)
User chose **PIC codegen + ET_DYN**. Investigator ground-truth flipped the plan: codegen is
**already fully RIP-relative** (`lea [rip+disp32]` for globals `OP_GLOBAL_ADDR`/func-addr
`OP_FUNC_ADDR`, `call rel32`); the sole image-absolute bake is one `OP_SPAWN` thunk in bundled
runtime, unreachable from leaf exports. ⇒ **P1 = linker-only, no codegen touch** → Windows/COFF
self-build byte-identical by construction. New RFC `rfcs/0032-pic-codegen-elf-shared-object.md`.

**P1 SHIPPED (linker.ax, gated `self.is_shared`):** (1) `e_type` → ET_DYN via new `is_shared`
param on `linker_build_elf_headers`; (2) `PT_GNU_STACK` (RW) replaces `PT_INTERP` for a `.so` —
without it `dlopen` dies *"cannot enable executable stack…"*; (3) `#[export]` funcs appended as
DEFINED `.dynsym` entries (`STB_GLOBAL|STT_FUNC`, `st_shndx=1` so ld.so adds load bias to
`st_value`), names into `.dynstr`, single-bucket `.hash` chain extended. Resolution mirrors COFF
`build_export_edata` (exported name = `export_names[k]`, addr = RVA of mangled `ax_<name>`).

**Verified** `scripts/so_export_check.sh` (WSL + python3 ctypes, no gcc): `axmath.so` = Type DYN;
`nm -D` → `T ax_add`/`T ax_mul`; dlopen+dlsym+call → `ax_add(2,3)=5`, `ax_mul(4,5)=20` under ASLR.
Gate: fast fixpoint **A==B `0FDCAE49`** (daily driver promoted), regression **528/528**,
elf_linux_check **12/12**. ⚠️ `readelf --dyn-syms` shows nothing (no section headers emitted) —
use `nm -D`. Dynamic block still gated `thunk_count>0` (bundled stdlib always imports libc).

## P2 / P3 — BOTH SHIPPED, RFC 0032 CLOSED
- **P2 (RIP-relative globals):** VERIFIED + oracle banked (`30d814b`). `soglobal.so`
  bump→[1,2,3]/addg(10)→13 under ASLR; `soprobe.so` intra-`.so` call (`ax_calc(4)=34`) +
  string-return (`ax_greet()='hi-from-so'`) (`9c97294`). All in `scripts/so_export_check.sh`.
- **P3 (OP_SPAWN ABS64→lea+PC32):** SHIPPED. The ONE grep-confirmed `vreg==1` site in
  `x86_emitter.ax` (spawn thread-entry) now uses the exact `lea [rip]`+PC32 form as `vreg==3`
  OP_FUNC_ADDR. Target-independent ⇒ gated by **B==C** (A≠B expected, **B==C `0D46A5C5`**,
  driver promoted), regression **529/529** (+`bin/t_spawnsmoke.ax` — spawn had ZERO runtime
  coverage before). ⚠️ **Validation caveat:** proven non-regressing + correct-by-construction on
  the P2-verified `lea`+PC32 foundation, but NO direct spawn-in-rebased-`.so` oracle — AXIOM
  `spawn` is actor-model (handler runs on `ax_actor_step` message dispatch, not on spawn), so an
  end-to-end observation needs actor-message infra in a `.so` (disproportionate). See RFC §7c.

## Lessons
- `readelf --dyn-syms` needs a section-header table (we emit none) → shows nothing; use `nm -D`.
- A `.so` with no `PT_GNU_STACK` → glibc tries an exec stack → `dlopen` refuses; emit it.
- The regression suite exercised `spawn` NOWHERE — a whole IR op with no runtime gate. Found
  while validating P3; closed with `t_spawnsmoke`. Check op-level coverage before trusting "green".

## Next (loop stays alive)
No OPEN bugs; RFC 0032 fully closed. Backlog = proactive probing / M4-M6 milestones /
p13 ARM64 backend / p14 allocator size-classes. Needs user direction for the heavy items.
