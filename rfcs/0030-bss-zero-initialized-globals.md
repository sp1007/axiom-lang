# RFC 0030 — `.bss`: zero-initialized globals must not occupy file bytes

Status: **DRAFT** (design settled, implementation phased)
Author: autopilot
Related: RFC 0017 (global variable storage), [[feedback-no-exe-bloat]], [[backlog-group-a-closed-2026-07-22]]

## 1. Motivation

A zero-initialized module-level global is stored **literally, as zero bytes, in the
executable file**. AXIOM emits its own objects and links them with its own linker, so no
external tool trims this: every byte the backend writes is a byte the user ships.

Measured on `bin/axc_native.exe` at `f1a7f9b`, program = one global plus a `main` that
touches it:

| global | exe size | delta over a no-global program (77,824 B) |
|---|---|---|
| none | 77,824 | — |
| `[i64; 50000]` (400 KB) | 478,208 | 400,384 |
| `[i64; 100000]` (800 KB) | 878,080 | 800,256 |
| `[i64; 200000]` (1.6 MB) | 1,678,336 | 1,600,512 |

The delta is now exactly the declared storage. It should be **≈0**: none of those bytes
carry information, and the OS loader can supply zeroed pages for free.

### A prior 2× defect, fixed separately — do not re-attribute it to `.bss`

Before `f1a7f9b`'s successor commit, the same three rows measured 878,080 / 1,678,336 /
3,278,336 — **exactly twice** the storage. That was not a `.bss` problem but a bug in
`x86_coff.ax`: the per-global padding loop aligned `data_buf` to the global's **size**
(`data_buf.len % gsize`) rather than its **alignment**, so a large global padded the
section up to a full multiple of itself before writing itself again. It is fixed
(align to the natural alignment, capped at 16). The remaining 1× is what this RFC targets.
Recording the distinction matters: the memory note that scoped `.bss` as "RFC-scale" was
sized against the 3.2 MB figure, half of which was a one-line bug.

## 2. Design — extend VirtualSize, do NOT add a fourth section

The obvious reading of "add `.bss`" is a fourth output section across COFF + ELF + the
self-linker's image layout. **That is not necessary**, and the cheaper design is also the
more standard one.

Both target formats already express "this region exists in memory but not in the file":

- **PE/COFF**: a section whose `VirtualSize > SizeOfRawData` has its tail zero-filled by
  the loader.
- **ELF**: a `PT_LOAD` segment whose `p_memsz > p_filesz` has its tail zero-filled by the
  loader. (`SHT_NOBITS` is the section-level spelling of the same thing.)

The self-linker already places all module-level globals contiguously at the end of the
writable region (`linker.ax` ~2407 for PE, ~2864 for ELF). So:

1. **Order globals so every zero-initialized one comes last** within the global block.
2. **Write file bytes only for the initialized prefix.**
3. **Set the region's memory size to the full extent**, leaving the loader to zero the tail.

No new **image** section, no new relocation routing, no change to how `OP_GLOBAL_ADDR`
resolves — a global's address is still its offset within the same region.

### Correction (found while reading `linker.ax`): the OBJECT still needs a `.bss` section

An earlier draft of this section claimed no new section anywhere. That is wrong, and the
reason is worth recording. The linker receives each object's globals as a **single opaque
byte block** (`ParsedObject.data`) with symbol offsets into it; it has no idea which of
those bytes came from a zero-initialized global, because `init_lo`/`init_node` live in
`AirGlobal`, back in the emitter. So the split cannot be decided at link time.

Therefore:

- **Object level:** a real fourth section `.bss` (`IMAGE_SCN_CNT_UNINITIALIZED_DATA`,
  `SizeOfRawData = 0`, `VirtualSize = N`; ELF `SHT_NOBITS`). Every downstream
  `section == 3` comparison must learn about section 4 — this is the fiddly part, and the
  part the original "RFC-scale" sizing was right about.
- **Image level:** still no fourth section. The linker appends each object's `.bss` extent
  as a **virtual-only tail** of the writable region: `SizeOfRawData` covers the initialized
  prefix, `VirtualSize` covers everything.

The rejected alternative — ordering zero globals last within `.data` and marking the
boundary with a synthetic symbol like `__ax_bss_start` — avoids the object section but
makes a symbol name load-bearing for layout correctness. A section is what the format
provides for exactly this; use it.

### Which globals qualify

A global is `.bss`-eligible iff it has **no initializer at all**: in `AirGlobal` terms,
`init_lo == 0` **and** `init_node == 0`. `init_node != 0` means a runtime initializer runs
at entry and writes the storage, so the file bytes are already dead weight — those are
eligible too, and are the common case for aggregates (RFC 0017 P2 makes every aggregate
runtime-init). A folded constant (`init_lo != 0`) is not eligible.

### Threshold — do not trade a size win for a size loss

Emitting the machinery unconditionally would grow small programs (padding to reach the
split point). So the split is applied only when the eligible bytes exceed one page
(**4096**); below that, zero globals stay in `.data` exactly as today. This keeps the
common small-program output **byte-identical**, which is also what makes the fixpoint
argument easy.

## 3. Alternatives considered

1. **A real fourth `.bss` section.** Rejected as unnecessary: it adds a section header, a
   COFF section number that every `section == N` comparison downstream must learn, and an
   ELF `SHT_NOBITS` header — all to express what `VirtualSize` already expresses. Keep it
   in reserve if a future need requires `.bss` to be separately addressable.
2. **Leave it to the OS's page-zeroing on a demand-paged file.** Rejected: the bytes are
   still in the file on disk and still transferred on copy/download; the user's constraint
   is about the artifact, not resident memory.
3. **Compress zero runs in the image.** Rejected: PE/ELF have no such facility, and the
   format already offers the intended mechanism.

## 4. Drawbacks

- Global ordering becomes load-bearing (zero-init globals must sort last), so the layout
  code gains an ordering invariant that a future refactor could silently break. Mitigated
  by an explicit size oracle (§6).
- Two size regimes (below/above the threshold) mean the small-program path and the
  large-program path differ — the threshold boundary needs its own test.

## 5. Migration / compatibility

No source-level change, no ABI change, no change to `OP_GLOBAL_ADDR` semantics. Programs
under the threshold are byte-identical. The compiler itself has no large zero globals, so
its own build is expected to be byte-identical ⇒ the fixpoint criterion is unaffected.

## 6. Gate (before commit)

Linker/section-layout change ⇒ **B==C is mandatory**, not just A==B (per the
fixpoint-async rule). Plus:

- Full regression.
- **A size oracle** (`scripts/exe_size_check.sh`): a program with a large zero global must
  link to an executable whose size is within a small constant of the no-global baseline.
  This is the test that would have caught the 2× alignment defect above, and it is the
  standing guard for [[feedback-no-exe-bloat]].
- Linux ELF check (`scripts/elf_linux_check.sh`) — the zero-fill path differs per format,
  so both must be exercised.
- The program must still *run*: a `.bss` global must read as zero and be writable.

## 7. Phased plan

- **P1 — the alignment fix.** Shipped separately; removes the 2× multiplier. (Done.)
- **P2 — size oracle.** Institutionalize the measurement before changing layout, so P3 is
  gated by a test that already passes at the current 1× and must reach ≈0×.
- **P3 — PE/COFF zero-tail.** Order + split + `VirtualSize`; B==C.
- **P4 — ELF zero-tail.** `p_memsz > p_filesz`; verified under WSL.

### P3/P4 edit sites (surveyed 2026-07-22 — the work is bounded, not open-ended)

`x86_coff.ax`
1. Global emission loop (~L682): partition `mod.globals` into initialized (bytes into
   `data_buf`) and zero-eligible (offset assigned in a parallel `bss_size` counter), once
   the eligible total clears the 4096 threshold.
2. Section headers (~L104-190): `num_sections` 3 → 4; add `.bss`
   (`IMAGE_SCN_CNT_UNINITIALIZED_DATA|READ|WRITE` = 0x C0000080), `SizeOfRawData = 0`,
   `VirtualSize = bss_size`, `PointerToRawData = 0`. Emit no raw bytes for it.
3. Symbols: zero-eligible globals get `section_num: 4` (COFF) / `section: 4` (ELF
   `SHT_NOBITS`).

`linker.ax`
4. Object parse (~L273-290): map the `.bss` header index to logical section **4**, beside
   the existing text/rodata/gdata mapping.
5. New `ParsedObject.bss_size` (no byte vector — there are no bytes to carry).
6. PE placement (~L2407): after appending each object's `.data`, reserve `bss_size` of
   **virtual-only** space; register `section == 4` symbols at those RVAs. Mirror at ELF
   (~L2864).
7. `section == 3` comparisons that must gain a `== 4` sibling: L1935, L2419, L2876 (plus
   the ELF symbol pass). Surveyed — there are only four, which is what makes this bounded.
8. `linker_build_pe_headers`: new `idata_virt_size` param → section `VirtualSize`,
   `SizeOfUninitializedData`, and `last_sec_end` (which feeds `SizeOfImage`).

The threshold check lives in `x86_coff.ax` alone; below it, nothing above is exercised and
the output is byte-identical to today.
