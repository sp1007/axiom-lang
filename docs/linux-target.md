# Linux ELF Target (RFC 0009 P3)

AXIOM cross-compiles to native **Linux x86-64 ELF64 executables** from the Windows
self-hosting toolchain. This document covers usage, the validated feature surface, the
implementation architecture, and known gaps.

## Usage

```sh
axc build prog.ax -o prog.elf --target linux -self-link -O1
```

- `--target linux` (aliases: `x86_64-linux`, `x86_64-unknown-linux`, `x86_64-linux-gnu`)
  emits an ELF64 executable using the **SysV** ABI instead of the default Windows COFF/PE.
- `-O1` is recommended: the `is_windows`/`is_linux` compile-time fold relies on the
  constant-branch DCE pass to drop the dead Windows runtime paths.
- The output is a dynamically-linked ELF (interpreter `/lib64/ld-linux-x86-64.so.2`,
  `DT_NEEDED libc.so.6` for `memset`/`memcpy`). It runs on any glibc Linux, incl. **WSL**.

### Running / testing

The verification gate is `scripts/elf_linux_check.sh` (needs WSL with a Linux distro):

```sh
export MSYS2_ARG_CONV_EXCL="*"
AXC=bin/axc_native.exe bash scripts/elf_linux_check.sh   # 10/10
# Manual run + exit code:
wsl /mnt/d/projects/compiler/Axiom/bin/prog.elf; echo $?
```

> **WSL gotcha:** run the ELF directly as a `wsl` *process* and read git-bash's `$?`.
> Do **not** wrap it as `wsl bash -c "…; echo $?"` — the interop layer swallows both the
> exit code and stdout in that form.

## What works (verified under WSL, cross-checked vs. Windows)

The Linux target has **parity with the Windows build for every feature that works there**:

- Computation, module-level globals, function calls, recursion, control flow, structs
- `for-in` loops (fixed arrays, `Vec`, string codepoints), nested collections (`Vec[Vec]`,
  `Vec[struct]`)
- `print` / `println` of `str`, `i64`, `bool`, `f64`; `to_str`
- **Heap:** `Vec`, `String` (concat), `HashMap` (int- **and** str-keyed) — including the
  large-object (`>4 KB`) path and multi-segment growth
- String ops: `slice`, `contains`, byte index, `==`
- `std.string.to_i64` / `to_f64` (pure-AXIOM, value-ABI)

## Architecture

| Concern | Implementation |
|---|---|
| Target select | `main_air.ax` `--target linux` → threads `is_windows=0` via `AirModuleBuilder.target_is_windows` (default `true`, so the Windows self-build stays byte-identical) and emits the `"elf"` object format. |
| Platform fold | `air_builder.ax` folds `@compiler_intrinsic("is_windows"/"is_linux"/"os_name"/"path_separator")` per target; the bundled runtime (`syscall.ax`, `alloc.ax`) picks the syscall path. |
| Syscalls | Inline `syscall` instruction (`0F 05`); the 7-arg `syscall(num,a1..a6)` and `syscall0..6` lower to `OP_SYSCALL`. `main` is the ELF entry (no CRT), so it exits via `exit(2)` syscall. |
| Object + link | `x86_elf64.ax` writes a 7-section ELF object (`.text/.rodata/.data/.symtab/.strtab/.rela.text`); the self-linker (`linker.ax`) auto-detects ELF vs COFF from the object magic and lays out `.text/.rodata/.data` + a PT_LOAD/interp/dynamic image. |
| Globals | Module-level globals emit to ELF `.data` (RFC 0017), same machinery as COFF. |
| Runtime | Freestanding shims bundled **only** for `--target linux`: `bootstrap/runtime/panic.ax` (print `str/i64/bool/f64`, `ax_panic`, `ax_str_eq`, mmap-backed global state) + `bootstrap/runtime/syscall.ax` (`sys_mmap`/`munmap`/`write`/`exit`). The allocator (`std/mem/alloc.ax`) mmaps its slab + segments directly; the actor system/scheduler init is gated off (not needed for the sequential heap). |

### The key bug fixed to unblock the heap

`OP_SYSCALL` moved each argument vreg **directly** into its fixed syscall register
(`num→RAX`, `a1→RDI`, … `a6→R9`). When an argument's register-allocator home happened to be
a *later* target register it was clobbered before use — so all-constant syscalls (the print
path) worked, but any variable-argument syscall (`mmap(size,…)`) received garbage flags/fd,
failed with `-EBADF`, and the allocator crashed/OOM'd. The fix (`x86_selector.ax`) spills every
syscall argument to the **SysV red zone** (128 B below `RSP`, untouched by `syscall`), then
loads the fixed registers from there — a store-then-load that breaks all register dependencies.

## Known gaps

- **`async` (`spawn`/`await`)** — unimplemented on **both** platforms (not a Linux gap):
  `OP_AWAIT` has no codegen and there is no actor-result model, so `await` returns 0. It cannot
  be rejected (the compiler's own `std/scheduler.ax` uses `spawn`), so completing it is an
  RFC-scale design + implementation effort.
- **Rare runtime symbols** (`ax_time_now_ns`, actor ops) — patched to 0 on the ELF link; a
  program that *uses* one crashes. Fix per case: a freestanding twin in the Linux runtime, or
  reimplement in pure AXIOM in `std/` (as was done for `to_i64`).
- **PT_LOAD** is a single RWX segment (not split RX/RW); no static (non-dynamic) ELF yet.
- **Other OSes:** macOS/Mach-O, iOS, Android are future work — generalize `target_is_windows`
  into a target enum (the `is_macos` fold is already stubbed).
