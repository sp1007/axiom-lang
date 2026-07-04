# RFC 0011 — Static libraries (`.lib`) + precompiled-stdlib cache

- **Status:** Accepted (2026-07-04) — **P1+P2+P3 SHIPPED**. P1 producer
  (`--staticlib` → COFF `!<arch>`), P2 consumer (`-l` static link), P3 source-hash
  staleness (`--staticlib` skips the build when the source is unchanged). P4
  (precompiled-stdlib cache via separate compilation) + P5 (generics) remain.
- **Author:** self-host team
- **Tracking:** follows [[0009-ffi-dynamic-linking]] (P1 import / P2 export shipped);
  addresses the "compile chậm" (slow build) thread from the perf session.
- **Liên quan:** `main_air.ax` (`concatenate_stdlib`, driver, arg parsing), `linker.ax`
  (`axiom_linker_link`, COFF parse), resolver/typecheck (separate-compilation), a new
  archive module.
- **Blocks:** fast incremental builds of large apps; shipping AXIOM libraries as
  binary artifacts; not recompiling stdlib on every build.

---

## 1. Motivation (user-raised, 2026-07-04)

> "Phải xuất được thư viện được liên kết tĩnh và phải sử dụng được nó. Tối ưu: precompile
> các thư viện vào `.lib` (vd `std.lib`) để khi biên dịch code standalone dùng chúng thì
> chỉ việc lấy ra để link — tiết kiệm nhiều thời gian. Nhưng phải có cơ chế quản lý mã
> nguồn: nếu bị thay đổi thì biên dịch lại và lưu `.lib` mới."

Two intertwined goals:

1. **Ship static libraries.** `axc` should emit a `.lib` archive of compiled object
   code, and another AXIOM build should link against it (static linking — bodies copied
   into the final EXE, no runtime DLL dependency). This complements RFC 0009 (dynamic).
2. **Precompiled-stdlib cache — the real time saver.** Today `concatenate_stdlib`
   (main_air.ax:343) reads every `std/*.ax` file, strips imports, and **splices the full
   stdlib source into every user translation unit**, which is then lexed + parsed +
   typechecked + code-generated *from scratch on every build*. Compiling stdlib once into
   `std.lib` and linking it would remove that repeated work — the dominant fixed cost of
   a small-program build.
3. **Staleness management.** `std.lib` must be rebuilt when its source changes; a build
   must never link a stale archive. Keyed on a content hash of the library sources.

## 2. Current state (surveyed)

- **Whole-program concatenation.** `-self-link` builds concatenate stdlib source
  (main_air.ax:656-657) + user code into ONE unit. There is **no separate compilation**:
  every symbol is defined and compiled together; the codegen emits a single
  `axiom_temp.obj`; the linker links that one object (+ hardcoded runtime imports).
- **Linker already parses COFF `.obj`** (`linker_parse_coff`) and links **multiple**
  inputs (`AxiomLinker.input_files`, `axiom_linker_add_input`). It resolves relocs by
  symbol name across all inputs (name-based, see RFC 0009 §10). So *consuming* extra
  object code is largely a matter of feeding more inputs.
- **No archive (`ar`/`.lib`) reader/writer** exists. No object-cache. No manifest.
- **Typecheck needs bodies?** No — typecheck needs *signatures* (types of functions,
  structs, consts). Bodies are needed only by codegen. This is the key lever for the
  cache (see §3.3).

## 3. Design

### 3.1 Archive format — standard COFF `!<arch>` (`.lib`)

Use the ubiquitous Unix `ar` "thin/second linker member" layout MSVC/binutils use, so
the artifact interoperates with `lib.exe`/`llvm-ar`/`dumpbin` and future toolchains:

```
"!<arch>\n"
1st linker member  "/  "  — symbol index: for each exported symbol, the offset of the
                            member object that defines it (big-endian u32 table + names)
(2nd linker member "/  " — MSVC sorted index; optional, can emit for compatibility)
longnames member   "//"  — long member filenames
member 1           "name/"  <header 60B>  <COFF .obj bytes>  (2-byte aligned)
member 2           ...
```

For AXIOM's own consumption the **symbol index is the load-bearing part**: given an
undefined symbol, the linker looks it up to pull the defining member. Producing a valid
index also makes the `.lib` usable by external linkers (ship AXIOM libs to C projects).

### 3.2 Producer — `axc build --staticlib`

```
axc build --staticlib std_bundle.ax -o std.lib
```

- Compiles the source to object code (reuse `compile_native_binary` → one `.obj`), then
  wraps it in an archive with a symbol index built from the object's **defined** symbols.
- MVP: a **single-member** archive (the whole compiled unit as one `.obj`). Correct and
  simple; the linker pulls "the member" when any of its symbols is needed. Multi-member
  (one object per source module, for finer dead-strip) is a later refinement.
- Emits a sidecar **manifest** `std.lib.manifest` = SHA-256 of the concatenated library
  sources + the compiler binary's own hash (so a compiler change also invalidates).

### 3.3 Consumer — two layers

**Layer A — link a `.lib` (mechanical, low-risk).** `axc build … -l std.lib` (or an
implicit `std.lib` when present): the linker detects the `!<arch>` magic, and for each
still-undefined symbol pulls the member that defines it from the symbol index, parses it
as COFF, and adds it to the link set. Reuses the existing name-based reloc resolution.
This alone lets AXIOM ship + consume static libs.

**Layer B — skip recompiling stdlib (the payoff; needs separate compilation).** To *not*
recompile stdlib, the user unit must be typechecked against stdlib **signatures** without
its bodies, then linked against `std.lib`. Prerequisite: an **interface/`.axi` mechanism**
— a compiled digest of stdlib's public signatures (structs, fn types, consts, monomorph
instantiations) the resolver/typechecker loads instead of stdlib source. This is real
separate-compilation infrastructure (a module-interface system) and is the largest piece.
Generics complicate it: monomorphization currently happens whole-program, so a stdlib
generic used with a *new* type argument from user code still needs the generic body →
either (a) keep generic bodies in the interface (compile on demand) or (b) restrict the
cache to non-generic stdlib. MVP for Layer B: **non-generic stdlib core** precompiled;
generic/monomorphized parts stay in-source until generic separate-comp lands.

### 3.4 Staleness / cache mechanism

- On `--staticlib`, write `<out>.lib.manifest`: `sha256(sources) + sha256(compiler.exe) + axiom_lib_format_version`.
- On consume, before using `std.lib`, recompute the source hash; if it differs from the
  manifest → **rebuild** the `.lib` (or error with `--frozen`). Deterministic:
  same sources + same compiler ⇒ same `.lib` (builds already reproducible/fixpoint).
- A tiny `axc lib` driver verb can front the build/refresh (`axc lib build std`,
  `axc lib check`).

## 4. Alternatives

- **Object-file cache (no archive).** Cache `axiom_temp.obj` keyed by TU hash. Doesn't
  help — the stdlib is concatenated *into the same TU* as user code, so the hash changes
  every build. Separating stdlib into its own object (§3.2) is the prerequisite anyway.
- **Precompiled AXIOM-specific archive (not `ar`).** Simpler to emit, but loses external
  interop (can't hand the lib to a C linker) and re-invents a solved format. Rejected.
- **Dynamic-only (RFC 0009).** DLLs already avoid recompiling the lib, but impose a
  runtime DLL dependency and cross-`axc` ABI hazards (RFC-12). Static linking is the
  right tool for a single self-contained binary. Complementary, not a substitute.

## 5. Drawbacks

- Layer B needs a module-interface/separate-compilation subsystem — a substantial new
  frontend capability (the biggest architectural lift on the roadmap after codegen).
- Archive symbol-index correctness + `ar` alignment quirks are fiddly (must match what
  external linkers expect if interop is desired).
- Cache invalidation is a classic footgun; must be conservative (hash sources +
  compiler) and default to rebuild-on-mismatch, never silently link stale code.

## 6. Phased plan (each phase gated: regression + self-host fixpoint)

- **P1 — Archive I/O + producer.** `!<arch>` reader/writer module; `axc build
  --staticlib x.ax -o x.lib` (single-member) + `x.lib.manifest`. Gate: emit a `.lib`,
  verify `llvm-ar t`/`dumpbin` reads it; round-trip parse.
- **P2 — Static consume (Layer A).** Linker pulls members from a `.lib` by symbol.
  Gate: `axmath.lib` (static) linked into an EXE that calls it — no DLL dependency; exit
  code oracle; fixpoint.
- **P3 — Manifest/staleness driver.** `axc lib build/check`; auto-rebuild on source-hash
  mismatch. Gate: touch a source → rebuild; unchanged → reuse (timing proof).
  **SHIPPED:** `axc build --staticlib` is now idempotent — it hashes the (post-concat)
  source with djb2 and, if `<out>.lib.manifest` matches and the `.lib` exists, prints
  `[lib] up to date — skipping rebuild` and exits before lexing (skips the whole
  pipeline). Any source edit changes the hash → full rebuild + refreshed manifest.
  Verified: fresh build → codegen runs; re-run unchanged → skipped; append a fn →
  rebuilds with the new symbol. A dedicated `axc lib` verb is deferred (folded into
  `--staticlib` for now); multi-file source hashing (stdlib) is P4's concern.
- **P4 — Separate compilation (Layer B), non-generic stdlib.** Module-interface digest
  (`.axi`); typecheck user unit against stdlib signatures; link `std.lib`. Gate:
  build a program with `--no-stdlib-source -l std.lib` faster than the concat path,
  identical runtime behavior, fixpoint.
- **P5 — Generic separate-comp.** On-demand monomorphization of archived generics.

## 7. Success criteria

- `axc build --staticlib std.ax -o std.lib` produces a valid COFF archive
  (`llvm-ar t std.lib` lists members; `dumpbin /symbols` shows the index).
- An EXE statically linked against `std.lib` runs with **no** extra DLL dependency.
- Editing a library source and rebuilding regenerates `std.lib` (manifest hash changes);
  an unchanged source reuses it — and a stdlib-cached build is measurably faster than the
  current concat-everything path.
- All deterministic; every phase passes regression + self-host fixpoint.

## 8. Open questions

- Interface digest format for Layer B: reuse the AST/typetable serialization, or a
  purpose-built `.axi`? How are monomorphized instances keyed and versioned?
- Multi-member archives (per-module objects) for dead-strip vs single-member simplicity —
  when is the granularity worth it?
- Where do cached libs live (`build/`, a global cache dir)? Reproducibility across machines.
- Interaction with `#[export]`/`--shared`: a `.lib` as an *import library* for a DLL
  (RFC 0009 P2 leftover) vs a static-code archive — keep the two `.lib` roles distinct.
