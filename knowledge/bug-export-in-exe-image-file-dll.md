---
name: bug-export-in-exe-image-file-dll
description: CLOSED — #[export] in a -self-link executable forced IMAGE_FILE_DLL, making every such .exe unloadable by Windows
metadata:
  type: project
---

**Status: FIXED 2026-07-22** (`linker.ax` ~L3165, B==C `7EAFE11B…`).

`AxiomLinker.link()` set `is_dll = true` whenever `export_names` was non-empty, on top of the
correct `is_dll = self.is_shared`. That put `IMAGE_FILE_DLL` (0x2000) into the PE
`Characteristics` of any EXECUTABLE containing an `#[export]` function — `0x2026` instead of
`0x0026`. Windows then refused to launch it:

```
The specified executable is not a valid application for this OS platform
```

The rejection happens in the loader, before `main`, so the program produced no output and no
exit code — nothing to attribute the failure to.

**The confusion the bug encodes:** *having an export directory* and *being a library* are
different facts. A PE executable may legally export symbols; Windows decides image kind from
the DLL characteristic bit alone. Only `--shared` may set it.

**How it was found:** not by a user and not by the regression suite — by writing
`bin/t_dfeexport.ax` as a root-set oracle for [[rfc-0031-dead-function-elimination]] and
running it. `#[export]` had no executable-shaped coverage at all, so the whole feature was
broken in the exe case and the suite was silent. Confirmed pre-existing (reproduced with the
flag-free build, and with a two-function program whose only difference was the attribute).

**How to apply:** when a PE/ELF header field is set from a *derived* condition
(`exports exist` ⇒ `is DLL`), check whether the condition is the actual invariant or merely
correlated with it in the cases that were tested. And: a feature with no oracle in the default
output shape is untested, however many RFC sections describe it.

Both directions are now pinned — `t_dfeexport` (exit 42) in `regression_repros.sh` for the exe
case, and `--shared` verified to still emit `0x2026`.
