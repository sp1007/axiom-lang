---
name: group-a-design-decisions-2026-07-22
description: "Group A (the four design-decision items) executed 2026-07-22. A3/A4 closed as documented decisions; A2 shipped RFC 0027 path D with a PROVEN soundness limit on the no-annotation model; A1 found and fixed a 2x executable-size defect (global padded to its own SIZE not ALIGNMENT) and drafted RFC 0030 for .bss."
metadata:
  node_type: memory
  type: project
---

# Group A executed 2026-07-22 — four design decisions, two real defects found

User: "Thực hiện #A. Đưa ra các option để tôi lựa chọn." All four answered, then executed in
risk order (docs → opt-in code → linker). Mid-task the user added a standing constraint:
**"không muốn làm phình các file thực thi"** ([[feedback-no-exe-bloat]]).

## The decisions

| | Item | Decision | Where recorded |
|---|---|---|---|
| A3 | `ascii`/`utf16`/`utf32` TYPES | **No new types** — `Vec[i64]` is canonical | RFC 0020 §10 |
| A4 | M4 dialect (`=>`, `.length()`, `impl Trait`) | **Adopt none** | [[m4-compliance-suite-spec-vs-impl-gap]] |
| A2 | RFC 0027 path D ownership model | **Infer from type + provenance, no syntax** | RFC 0027 §9 |
| A1 | `.bss` | **RFC 0030 + implement** | rfcs/0030 |

## ⭐ A2's real result: the chosen model has a PROVEN hard limit

Path D shipped (`f1a7f9b`) as: free a field iff it is pointer-typed AND every write to it is
a **syntactically fresh allocation**. Path C's audited `{Vec,HashMap,HashSet}` list is kept
as a fallback so the derivation can only ADD coverage.

**The instructive part is the failed strengthening.** Following a local one binding — to
recognise the idiomatic `let d = @alloc(..); Buf(data: d)` — is UNSOUND and produced a real
use-after-free. `Buf(data: d)` (owning) and `View(borrowed: d)` (borrowing) are
*syntactically identical*; what separates them is whether anyone else still holds the
pointer, which is an **aliasing** question that provenance cannot answer. Oracle
`t_ctgcborrow` caught it (42 → 99: the borrowed buffer was freed and the next allocation
overwrote it).

⚠️ **So the annotation option the user declined is the one that actually generalises.** This
is now recorded as measured evidence, not opinion. Coverage limits, all in the leak (safe)
direction: the idiomatic local form is uncovered; field names are matched program-wide so
`data` is permanently demoted by `Vec[T](data: data)` in `with_capacity`; no recursion into
nested aggregate fields.

⚠️ **Oracle honesty:** `t_ctgcuser` pins `on == off` but **cannot detect a silent no-fire** —
the allocator does not return a just-freed block to the next same-size request in that
shape, so a program cannot observe its own heap. Firing was confirmed with a temporary
trace in `emit_owned_field_frees`. The dangerous direction has a sharp oracle; the inert
direction does not. Recorded rather than papered over (RFC 0028 lesson).

## ⭐ A1's real result: a 2× exe-size defect that was NOT a .bss problem

Every executable carrying a large global was **exactly twice** the necessary size. Root:
`x86_coff.ax` padded each global with `data_buf.len % gsize` — aligning to the global's
**SIZE** instead of its **ALIGNMENT** — so a large global padded `.data` up to a full
multiple of itself and then wrote itself again. Fix: align to natural alignment, capped at
16.

Measured, `[i64; 200000]`: **3,278,336 → 1,678,336 bytes (−48.8%)**, program still correct.
Compiler's own size unchanged (2,536,960) — it has no large globals.

⚠️ **The old note scoping `.bss` as "RFC-scale" was sized against the 3.2 MB figure, half of
which was this one-line bug.** Always separate "the format wastes space" from "we have a
padding bug" before sizing the work.

**Method note:** the doubling was found by *measuring the slope* (50k/100k/200k elements →
delta exactly 2× each time), then bisecting the pipeline (exe → static lib → COFF section
headers) rather than reading code. A distinguishable-value probe also ruled out a false
lead: the 4 copies of a sentinel found in the exe were `mov imm64` operands in `.text`, not
duplicated storage.

## RFC 0030 (.bss) — design settled, P3/P4 open

Cheaper than the "fourth section everywhere" framing: **image** needs no new section, just
`VirtualSize > SizeOfRawData` (PE) / `p_memsz > p_filesz` (ELF) with zero globals ordered
last. But the **object** does need a real `.bss` section — the linker sees each object's
globals as one opaque byte block and cannot tell which bytes are zero-eligible
(`init_lo`/`init_node` live in `AirGlobal`). Threshold 4096 so small programs stay
byte-identical. P1 (alignment fix) + P2 (`scripts/exe_size_check.sh`) done; P3 (PE) and P4
(ELF) open, both needing **B==C**.

New standing guard: **`scripts/exe_size_check.sh`** — asserts exe size against a no-global
baseline. The regression suite compares exit codes and was blind to the 2× defect; this is
the suite that would have caught it. Tighten its budgets when P3/P4 land.

Related: [[feedback-no-exe-bloat]], [[backlog-open-items]], [[ctgc-p3-scoping-2026-07-18]].
