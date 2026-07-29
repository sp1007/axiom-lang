# RFC 0035 — Module-namespaced symbols and link-time collision detection

- Status: **P1 SHIPPED (diagnostic)** / **P2 SHIPPED for the `.lib` path** (see §9) / P3 PROPOSED
- Author: autopilot, 2026-07-29
- Approved in advance by the user (standing decision D3, `knowledge/user-decisions-2026-07-29.md`):
  the real cure is pre-approved, RFC + `B==C` still mandatory.
- Related: BUG#50 (cross-module same-name fn), `knowledge/task-cross-library-name-collision.md`,
  `knowledge/bug-user-fn-stdlib-struct-name-collision.md`, RFC 0011 (static/auto libraries),
  RFC 0031 (DFE root set matches by NAME).

## 1. Motivation

AXIOM emits a symbol name per function derived from the source name (`ax_<name>`). Nothing in
that scheme carries the defining module, so two modules that both define `helper` produce the
same emitted name. The linker resolves every reference by a **linear first-match scan** over
`func_names` (`linker.ax`, the reloc resolution loops), so the first definition parsed wins and
the second is discarded **silently** — calls intended for it are bound to the other body.

Five holes in this family have been recorded. Four were fn-vs-fn; the fifth was fn-vs-struct,
mitigated 2026-07-29 by rejecting the clash in the typechecker (`C432EA9E`). Those are point
fixes to a scheme that is wrong at the root.

## 2. The defect this RFC was written from (found by P1, 2026-07-29)

The existing mitigation for fn-vs-fn is `SYM_FLAG_MODDUP` (2048), set in `typecheck.ax` when two
**bare module-level symbols in the same compilation** share a `name_id`; `x86_resolve_sym_name`
then emits `ax_<name>__m<sym_idx>` instead of `ax_<name>`.

Both halves of that are per-compilation, and that is fatal once libraries are compiled
separately (RFC 0011 `--auto-lib`). Reproduced on the shipped compiler:

```
libpa.ax:  pub fn helper() -> i32: return 10
libpb.ax:  pub fn helper() -> i32: return 20
appcol.ax: import libpa / import libpb / libpa.helper() + libpb.helper()

axc build appcol.ax --auto-lib -self-link -O1
  warning[E0501]: symbol defined more than once ... ax_helper
  error: linker: unresolved external symbol 'ax_helper__m1755'
```

Two independent failures, both caused by the same root:

1. **The callee does not mangle.** Each library is compiled ALONE, so within that compilation
   nothing shares `helper`'s name, flag 2048 is never set, and both libraries emit plain
   `ax_helper`. Two definitions, one name — a silent first-wins mis-link.
2. **The caller does mangle, to a name nobody defines.** The app sees two same-named imports,
   sets flag 2048, and emits a call to `ax_helper__m1755`. No library ever emitted that name,
   because `sym_idx` is an index into the *importing* compilation's symbol table.

So the caller and the callee disagree about the name of the same function. **A mangling derived
from per-compilation state cannot be a link-time contract.** `sym_idx` in particular is not
stable across compilations — the task note predicted this, and P1 has now demonstrated it.

## 3. What P1 shipped

`report_duplicate_definitions` (`linker.ax`) scans the defined TEXT symbols and warns
(`E0501`) when one emitted name is defined more than once, naming it. Detection is exact — two
defined symbols with byte-identical names — not heuristic.

It is a **warning, not an error**, and that is a finding rather than timidity: the same run
shows `ax_Ok`, `ax_Err`, `ax_Some`, `ax_None`, `ax_sum_layout_is_pointer`, `ax_block_size`
duplicated too, because each library embeds its own copy of those runtime shims. Duplicate
definitions are therefore NORMAL on the multi-library path today, and erroring would break it.
Separating "benign duplicate of an identical shim" from "two different functions" is exactly
what P2 makes decidable.

Gated on a multi-object link: one compilation cannot emit a name twice (the frontend rejects
same-name clashes before codegen), so with a single object the scan can only come up empty.

Calibrated, per the rule that a guard never seen to fire is not a guard: it fires on the case
above, and reports **zero** duplicates across the full 554-program regression and the compiler's
own self-link.

## 4. Proposed P2 — a stable, module-qualified symbol scheme

Replace the flag-2048 heuristic with an unconditional scheme:

```
ax_<module-id>_<name>            free function
axS_<module-id>_<type>_<name>    method / associated function
axG_<module-id>_<name>           module-level global
axC_<module-id>_<type>           type constructor / ctor glue
```

`<module-id>` must be derived from the module's canonical import path (e.g. a hash of
`axprobe/liba`), **never** from `sym_idx` or any table index, so that a library and everything
that imports it compute the same name independently. This is the property flag-2048 lacks and
the whole reason §2 fails.

Distinct prefixes per symbol CATEGORY are what closes fn-vs-struct structurally, rather than by
the typecheck rejection currently standing in for it.

### Constraints this must respect
- **Determinism (§3)** — the module-id must be a pure function of the canonical path, stable
  across machines and runs. Reproducible builds depend on it.
- **`B==C` before commit (§24)** — this is an ABI change by definition.
- **Name-matching predicates elsewhere must move with it.** At least three places compare
  emitted names across a layer boundary and would silently misbehave: RFC 0031's DFE root set
  (`dfe_is_abi_name`, which already strips one `ax_` prefix), `#[export]` matching (which had
  to match both intern-id AND mangled name), and the runtime ABI-shadow binding in
  `x86_resolve_callee_name`. This is the recorded lesson *"a predicate comparing names across a
  layer boundary has mangling as part of the comparison"* — the third such bug in one session.
- **Runtime/ABI symbols keep their fixed names.** `ax_malloc`, `ax_panic`, syscall shims and
  every DLL import are an external contract and must NOT be namespaced.

## 5. Proposed P3 — promote E0501 to an error

Once P2 gives every non-ABI symbol a unique name, a duplicate definition can only mean a genuine
clash, and the benign shim duplicates of §3 disappear (they become either module-qualified or
recognised ABI names). At that point E0501 becomes a hard error with both definition sites named.

## 6. Alternatives considered

- **Keep widening the flag-2048 heuristic.** Rejected: §2 shows the failure is not a missing
  case but the per-compilation premise itself.
- **Make the linker prefer the "closest" definition.** Rejected: silently picking a winner is
  the existing bug with better manners.
- **Reject clashes in the frontend only** (what fn-vs-struct does today). Rejected as the
  general answer: it cannot see across separately-compiled libraries, which is where §2 lives.

## 7. Drawbacks

Longer symbol names inflate the COFF/ELF string table (bounded — names are not in `.text`, and
RFC 0031 already prunes dead functions). Debuggers and `nm` output become less readable; a
demangler in `axc` would be a follow-up. P2 is a flag day for any prebuilt `.lib`, which must be
rebuilt — acceptable now, when no `.lib` is distributed outside this repo.

## 7bis. P2 PREREQUISITE — there is no module identity to mangle with yet (found 2026-07-29)

P2 cannot begin with the mangler. **`Symbol` (`resolver.ax:59`) has no module field**:

```
name_id, kind, padding, flags, type_id, decl_node, scope_id, next_overload
```

`scope_id` is a lexical scope, not a module, and the existing flag-2048 path had to fall back on
`sym_idx` precisely because nothing better exists — which is the root cause in §2. So the first
sub-task of P2 is **giving the symbol table a stable module id**, derived from the canonical
import path, and only then changing `x86_resolve_sym_name`.

Two hazards to plan for, both already recorded in this repo:
- `Symbol` is a **fixed-layout struct populated positionally** in places; BUG#21 was exactly a
  field-order/omission bug of this kind in `LinkerSymbol`, where a missing field shifted `defined`
  into `size` and made every parsed symbol read as undefined. Adding a field means auditing every
  construction site, not just the declaration.
- The module id must be a pure function of the canonical path (§4), so it must be computed where
  the import path is still known — the resolver — and not re-derived later from a table index.

Sequencing for P2, each step independently gated:
1. add the module id to `Symbol` + populate it in the resolver (inert: nothing reads it yet, so
   this should be `A==B`);
2. switch `x86_resolve_sym_name` to the module-qualified scheme, retiring flag 2048;
3. update the cross-layer name predicates of §4 in the SAME commit as step 2.

### Groundwork done 2026-07-29 (so step 1 does not have to re-derive it)

- **The stride footgun is already defused** (`518264c`). `SymbolVec.push` spelled Symbol's size
  as the literal `24` in both the allocation and the memcpy. Adding a field would have
  under-allocated the buffer and corrupted the heap from inside the symbol table on the next
  push, with nothing reporting it. It now derives from
  `@compiler_intrinsic("size_of")[Symbol]()`, the idiom `air.ax` already uses. Shipped
  separately and proven inert (`A==B`), so the field addition cannot be blamed for a corruption
  it would merely have triggered.
- **There are only 7 `Symbol(...)` construction sites** — `resolver.ax` x4 (`:477`, `:517`,
  `:561`, `:576`) and `main_air.ax` x3 (`:1570`, `:1608`, `:1631`). Small enough to audit by
  hand, which is required: they are positional, and BUG#21 was exactly a field-order mistake of
  this kind.
- **The module id has a real source.** `ModuleInfo` (`resolver.ax:322`) already carries
  `name_id` and `file_path`, and `LazyResolver.modules` tracks every imported module;
  `lazy_resolver_register_import` defines a `SYM_MODULE` symbol per import. `file_path` is the
  canonical path §4 asks for.
- **ANSWERED, AND IT KILLED THE FIRST DESIGN — there is no module nesting to hang a cursor on.**
  The plan above (a `module_id` field on `Symbol` plus a `current_module` cursor on
  `SymbolTable`, set and restored around `ax_ax_driver_load_module` in
  `lazy_resolver_preload_module`) was **built, gated, and reverted**. It reached `A==B B0AEA1C0`
  and 557/557 — and that green result was **meaningless**, because the field was structurally
  always `0`.

  A temporary probe printing every `define()` where `current_module != 0` reported **zero hits**
  compiling a normal program, and **zero on the `--auto-lib` path too**. `ax_ax_driver_load_module`
  simply does not fire in these flows. Had the probe not been run, an always-zero field would have
  shipped behind a genuine-looking `A==B`, and step 2 would have been built on top of it.
  (This is the standing lesson — assert the premise, do not infer it from a null result — and it
  is the second time in the same session that a green gate was proving nothing.)

  **Corrected model of the compilation unit.** AXIOM's native path has no module nesting at
  compile time: each `axc build X.ax` is ONE flat unit (bundled stdlib + X's source), and a
  library is a SEPARATE compilation (RFC 0011 `--auto-lib` shells out per library). There is no
  point in resolution at which "the module being entered" is a meaningful, nested notion — which
  is also the real reason the flag-2048 mitigation had to reach for `sym_idx`.

  **So the module id must come from the compilation UNIT, not from a cursor during resolution.**
  When the compiler builds `libpa.ax`, the identity `libpa` is known from the outset — it is the
  unit being compiled. Deriving the qualifier from that gives exactly the symmetry §2 lacks:
  - compiling `libpa.ax`, `helper` emits `ax_libpa_helper`;
  - compiling `libpb.ax`, `helper` emits `ax_libpb_helper`;
  - the app, which resolves the call as `libpa.helper`, computes `ax_libpa_helper` from the
    module name in the call — the SAME string, with no shared table and no index.

  This is simpler than the reverted design (no `Symbol` field, no cursor, no positional-construction
  audit) and it is the only form that can work across separate compilations. Step 1 is therefore
  **not** a symbol-table change; it is "establish the compilation unit's canonical module name and
  make it reachable from `x86_resolve_sym_name`", after which steps 2–3 proceed as written.

  Note for step 2: symbols that are NOT owned by the unit (bundled stdlib, runtime/ABI names)
  must keep their present names — the qualifier applies to the unit's OWN definitions only.

- **Both sites are already located, and each already has the module name in hand.**
  1. **App side — `register_module_from_lib`** (`main_air.ax:1539`), which takes `mod_name: str`
     as a parameter. Its own comment states the defect in so many words: *"The symbol carries
     only SYM_FLAG_PUB, so cgen mangles the call target as plain `ax_<name>`, which matches the
     library's exported symbol."* That match is exactly what makes two libraries indistinguishable.
     This is where an imported symbol would be marked as owned by `mod_name`.
  2. **Library side — the `--staticlib` compilation**, where the unit's own name is known from
     the source/output path, and `build_lib_iface_text` (`linker.ax:1495`) writes the `F` lines
     that the app later reads back. The emitted symbol and the iface entry must agree.

  Because `--auto-lib` is opt-in and currently BROKEN for same-name libraries (RFC 0035 §2), a
  first increment can change ONLY that path: qualify a library's own public symbols and the
  imports registered from a `.lib`. The self-host build uses neither, so it should be inert —
  which also gives a clean `A==B` signal that nothing else moved. **But note the trap this RFC
  already fell into once: `A==B` inert is ALSO what a dead no-op looks like.** Any such increment
  must be accompanied by the §2 repro (two libraries, same function name, both called, expected
  30) actually returning 30, not merely by a green gate.

## 8. Migration

P1 (shipped) is additive and inert on the normal path. P2 must land in one commit with all the
name-matching predicates of §4 updated together, gated on `B==C` plus the full regression, ELF,
`so_export`, exe-size and ctgc suites, and it should carry an oracle built on the §2 repro (two
libraries, same function name, both actually called, correct answer 30).

## 9. What P2 shipped (2026-07-29, fixpoint `DC4FD242`, 557/557)

The §2 repro returns **30**. That, and not the gate, is the evidence — §7bis records this same
RFC reaching a clean `A==B` on a change that was a structural no-op, so an inert green result
is the one thing that cannot distinguish a working increment from a dead one here.

**The qualifier is the compilation unit's canonical name**, mapped through one pure function
(`module_qualifier`, `main_air.ax`) that collapses the two spellings the two sides hold:
`libpa.ax` / `libpa.lib` on the library side and the import name `libpa` on the app side, with
`/`, `\` and `.` all folding to `_`. No hash, no ordering, no table index — so `pkg/mod.ax`
compiled alone and `import pkg.mod` resolved elsewhere compute the same string with nothing
shared between the compilations. This is the property §2 shows `sym_idx` cannot have.

Two sites, one name, both returning ahead of the `ax_<Struct>_<fn>` method decoration so the
definition and the reference cannot disagree for a function whose first parameter is a struct:

- **Library side** — `SymbolTable.unit_qualifier`, set by the driver only for `--staticlib`,
  read by `x86_resolve_sym_name` for the unit's own public functions. Scoped to `--staticlib`
  because that is the only output linked into a *different* compilation, which keeps every
  ordinary build byte-identical.
- **App side** — `SYM_FLAG_MODQUAL` (8192) on the body-less symbols
  `register_module_from_lib` registers from a `.lib` interface, whose `name_id` is set to the
  qualified base. The local name still drives all lookup (global-scope `mod.fn` binding and
  the `ModuleExport` list), so source-level resolution is untouched, and the now-distinct
  name_ids also stop flag 2048 from firing — which is what produced the call to
  `ax_helper__m1755` that nothing defined.

**Runtime/ABI names are excluded** by reading `is_valid_runtime_dll_symbol` from `linker.ax`
rather than restating the list, testing both the source spelling and the default emitted one.
The §4 warning about cross-layer name predicates turned out to need no further work: the DFE
root set already computes every name it compares *through* `x86_resolve_sym_name`, so it moved
with the scheme automatically, and `#[export]` already matched on intern-id as well as name.

Gate: `A==B DC4FD242` (inert on the self-host build, as intended), regression **557/557**,
ELF 12/12, ctgc 16/16, exe-size 4/4, so_export ✓, and the new
`scripts/lib_collision_check.sh` 5/5 — **calibrated**: on the shipped compiler it fails 3 of 5
with exactly the §2 symptoms (`unresolved external symbol 'ax_helper__m1755'`, both libraries
emitting plain `ax_helper`), while its two over-qualification guards pass on both compilers.

### Still open after P2

- **Methods, globals and constructors keep the old scheme.** `axS_`/`axG_`/`axC_` from §4 are
  not implemented; only free functions are qualified, because only they cross the `.lib`
  interface today. fn-vs-struct therefore still rests on the typecheck rejection.
- **P2 is a flag day, and nothing detects it.** The `.lib` staleness manifest hashes only the
  library source, so a `.lib` built by an older compiler is considered fresh and is silently
  linked against the new names. The oracle deletes stale `.lib`s for this reason. A compiler
  identity in the manifest would close it.
- **`mod_name` reaches `register_module_from_lib` empty** (probed, which is why the qualifier
  is derived from `lib_path`). The `mod.NAME` global binding written from it is therefore dead
  — resolution reaches these symbols through the `ModuleExport` list. Harmless today, but it
  is a live trap for anyone who reads that code and assumes the binding works.
- **Multi-segment non-std module paths do not resolve at call sites.** `import bin.libcol.liba`
  compiles the library correctly but `bin.libcol.liba.helper()` fails with
  "undefined name 'bin'". Pre-existing — the shipped compiler fails identically — and the
  reason the oracle uses single-segment module names.
- **P3 (E0501 as an error)** is still blocked as §5 describes: the benign duplicated runtime
  shims remain, since each library embeds its own copy and those names are deliberately not
  qualified.
