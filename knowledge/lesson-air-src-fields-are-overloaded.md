---
name: lesson-air-src-fields-are-overloaded
description: "AIR's src1/src2 are NOT registers in general — they hold an immediate for OP_ICONST, the two 32-bit IEEE halves for OP_FCONST, and a block index for OP_JUMP/OP_BRANCH. Sizing a per-vreg structure by max(dest,src1,src2) made one 3.0 literal request a gigabyte, and the compiler APPEARED TO HANG (rc=124, no error, no output file)."
metadata:
  node_type: memory
  type: reference
---

# AIR `src1`/`src2` are overloaded fields, not registers

Any pass that walks AIR instructions and treats `src1`/`src2` as **virtual register numbers** is
wrong for at least four opcodes:

| opcode | what `src1`/`src2` actually hold |
|---|---|
| `OP_ICONST` | the immediate **value** |
| `OP_FCONST` | the **two 32-bit halves** of the IEEE bit pattern |
| `OP_JUMP` | a **block index** |
| `OP_BRANCH` | a **block index** |

## How it bit (2026-07-31, during `6febd02`)

The first version of `verify_air_no_int_into_float` sized its per-vreg map by
`max(dest, src1, src2)` over all instructions. A single `3.0` literal lowers to `OP_FCONST` with
`src2 = 1074266112` (the high half of the double `3.0`), so the pass asked for a map of ~10^9
entries.

⭐ **The failure mode is what makes this worth a file.** It did not report an allocation error. The
compiler **appeared to hang**: `rc=124` from the timeout, **no diagnostic, no output file**. Nothing
in that signature says "bad allocation size" — it reads as an infinite loop in whatever you last
touched.

⚠️ **And it did not reproduce on the compiler's own source.** `bootstrap/stage1/*.ax` is
integer-only, so it has no `OP_FCONST` with a large half — the pass ran clean over all 1053
functions. It only broke on programs containing float literals: **exactly the class of program the
change was written for.** A pass can be self-host-clean and still be broken for users.

## The rule

- **Bound a per-vreg structure from `dest` only.** Every genuine register operand is some
  instruction's `dest`, so `max(dest)` is a sound upper bound and needs no per-opcode knowledge.
- Add a sanity cap anyway, so a future opcode with a surprising `dest` fails loudly instead of
  allocating.
- If you must read `src1`/`src2` as registers, **switch on the opcode first** — there is no generic
  "is this operand a register" predicate.

## Family

Same category as [[lesson-exit-code-8bit-masking]] and
[[lesson-bash-grep-not-powershell-selectstring]]: the tool used to *inspect* the program silently
did something other than what it appeared to do. Here the inspecting pass was the compiler's own
verifier, which is the most expensive place for it to happen — a verifier that hangs is worse than
no verifier, because it looks like the code under test is at fault.
