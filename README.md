# AXIOM

A statically typed systems programming language with a **self-hosting** compiler,
a custom IR pipeline, native x86-64 code generation, a custom object writer/linker,
and a small runtime — all written in AXIOM itself.

> Status: self-hosting. The compiler compiles its own source to a byte-identical
> binary (fixpoint), passes a growing regression suite, and emits native
> executables directly (no external assembler or linker required on Windows).

## Highlights

- **Statically typed** with structs, sum types (ADTs), generics + monomorphization,
  `Option[T]` / `Result[T, E]`, pattern-matching `match`, and interfaces.
- **Self-hosting compiler** (`bootstrap/stage1/*.ax`): lexer → parser → AST →
  type checker → AIR (typed IR) → SSA optimization passes → x86-64 machine IR →
  native object emission → self-linking.
- **Native backend**: direct x86-64 instruction selection, register allocation, and
  PE/COFF object + executable emission on Windows (ELF emission is partial).
- **Deterministic, reproducible builds** — verified by a compiler-compiles-itself
  fixpoint gate (`A == B`, and `B == C` for backend changes).
- **UTF-8 strings** (RFC 0020): `str` is UTF-8 by default with byte-indexed `s[i]`
  and codepoint iteration `for c in s`; plus a distinct `bytes` type and `rune`.
- **Collections**: `Vec[T]` (with `v[i]` / `v[i] = x` subscript), `HashMap`,
  `HashSet`, arrays, and `for x in …` iteration.
- **FFI**: `extern "C"` import, `--shared` DLL / `--staticlib` export, and
  import-driven separate compilation (RFC 0009 / 0011).

## Repository layout

```
bootstrap/stage1/   the self-hosted compiler (lexer, parser, typecheck, IR, codegen, linker)
std/                the standard library (string, io, collections, result, …)
bin/                the daily-driver compiler (axc_native.exe) + oracle test programs
scripts/            build, fixpoint, and regression harnesses
rfcs/               design RFCs (0001–0021)
AXIOM SPECIFICATION/ language + subsystem specifications
docs/               architecture notes and task plans
tests/ examples/    test programs and samples
```

## Building & running

The daily-driver compiler is `bin/axc_native.exe`. Compile and run an AXIOM program:

```sh
bin/axc_native.exe build path/to/program.ax -o out.exe -O1
./out.exe
```

Rebuild the compiler from source and verify the self-host fixpoint:

```powershell
# regenerates the concatenated source, builds A from the seed, then B from A,
# and checks SHA-256(A) == SHA-256(B)
scripts/fast_fixpoint.ps1
```

Run the regression suite:

```sh
AXC=bin/axc_native.exe bash scripts/regression_repros.sh
```

> Note: on Windows the built executable needs `ax_runtime.dll` on the PATH (or next
> to the executable). Build programs from the repository root so the standard
> library modules bundle correctly.

## Design & governance

Architecture and language decisions are recorded as RFCs in [`rfcs/`](rfcs/) and as
specifications under [`AXIOM SPECIFICATION/`](AXIOM%20SPECIFICATION/). The guiding
principles are correctness over cleverness, determinism over convenience, and
long-term maintainability — every compiler change is gated by the self-host
fixpoint and the regression suite before it lands.

## License

Released under the [MIT License](LICENSE).
