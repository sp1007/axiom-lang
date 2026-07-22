---
name: dfe-elf-runtime-is-in-program
description: On --target linux the ax_* runtime is compiled INTO the program, so RFC 0031 roots that are DLL imports on Windows are prunable functions on ELF
metadata:
  type: project
---

**The asymmetry:** on Windows the `ax_*` runtime family (`ax_println_str`, `ax_str_eq`,
`ax_panic`, `ax_alloc`, …) is imported from `ax_runtime.dll`. On `--target linux` there is no
such DLL: the same family is compiled INTO the program as ordinary AXIOM functions
(`std/runtime.ax` plus the `sys_write` / `sys_mmap` / `ax_lx_write1` syscall layer). ELF
builds carry 206 functions where the COFF build of the same source carries 184.

**Why that is dangerous for [[rfc-0031-dead-function-elimination]]:** those functions are
reached only through the magic negative callee indices (`-1` `ax_alloc`, `-2` `ax_free`,
`-10..-17` print/println, `-22` `ax_str_eq`), which carry no call edge in the AIR graph. On
COFF they are imports — not members of `mod.funcs` — so pruning cannot touch them and their
absence from the root set is harmless. On ELF they are exactly the kind of thing pruning
deletes.

**The bug this actually produced:** the ABI root test called
`is_valid_runtime_dll_symbol(nm)` directly. That list is written in ABI names
(`ax_println_str`), but codegen emits AXIOM definitions with its own `ax_` prefix, so the same
function appears as **`ax_ax_println_str`** and never matched. Result: every pruned ELF binary
SIGSEGV'd (exit 11), while COFF regression passed **511 / 511**. Fixed by `dfe_is_abi_name`,
which tests the raw name and the name with one leading `ax_` stripped. ELF ABI roots went
10 → 28 and the gate went 4/12 → 12/12.

**How to apply.** Two things generalise:

1. **A green suite on one target says nothing about the other when the two targets have
   different amounts of the runtime compiled in.** COFF hid a systematic error completely
   because the affected functions do not exist there. Any pass that reasons about "which
   functions exist" must be gated on ELF too, not just PE.
2. **When a predicate compares NAMES across a layer boundary, the mangling is part of the
   comparison.** This is the second time in one session: `#[export]` roots reported zero
   because `export_syms` holds plain intern ids while the emitted symbol is mangled. Both
   failures look identical from outside — a root category that silently matches nothing.

Both `scripts/regression_repros.sh` and `scripts/elf_linux_check.sh` now honour `AXEXTRA`, so
the whole suite can be re-run under `-dfe` on both targets. That hook is how this was found
and is the only practical guard for the root set.
