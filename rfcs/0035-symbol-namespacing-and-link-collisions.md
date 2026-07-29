# RFC 0035 — Module-namespaced symbols and link-time collision detection

- Status: **P1 SHIPPED (diagnostic)** / P2–P3 PROPOSED
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
  canonical path §4 asks for. **Open question for step 1:** whether the resolver knows the
  *current* module while walking a loaded module's AST, or whether that context has to be
  threaded through — that is the one thing left to establish before writing code, and it decides
  whether step 1 is a field-plus-assignment or also a plumbing change.

## 8. Migration

P1 (shipped) is additive and inert on the normal path. P2 must land in one commit with all the
name-matching predicates of §4 updated together, gated on `B==C` plus the full regression, ELF,
`so_export`, exe-size and ctgc suites, and it should carry an oracle built on the §2 repro (two
libraries, same function name, both actually called, correct answer 30).
