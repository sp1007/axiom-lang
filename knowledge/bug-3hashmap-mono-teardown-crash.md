---
name: bug-3hashmap-mono-teardown-crash
description: "✅ CLOSED (definitively mitigated) 2026-07-19b: the intermittent teardown SIGSEGV when compiling 3+ distinct hash-container monomorphizations is ELIMINATED by the shipped teardown-skip (main_air.ax: on a successful self-link build, flush + exit, skip the free-chain — a STANDARD compiler technique, the OS reclaims memory). Re-verified this session: 0 crashes / 60 builds of c_3map, output exe exit 0. The residual root-cause OOB (a mid-block indexed write clobbering a pointer during hash-container generic mono) is BOUNDED and output-invisible; an independent fresh root-cause pass (strides Symbol=24/Scope=32 both CORRECT; clone_subtree_from index-safe; substitute_type_params/remove_generic_params_child don't grow the tree; HashMap struct is K/V-independent 48B) corroborates prior sessions — not findable by inspection, and the Windows-heisenbug tooling doesn't apply (C-backend ASAN can't build the compiler, valgrind doesn't repro on Linux, gdb/footer-canary mask a mid-block write). CLOSED: disproportionate to pin further for a LOW-sev, output-correct, mitigated bug. AX_CANARY infra remains committed-disabled in std/mem/alloc.ax if ever revisited."
metadata:
  type: project
---

# ✅ CLOSED (definitively mitigated) 2026-07-19b

The observable bug — the intermittent teardown crash — is **definitively eliminated** and the
bug is closed. Resolution = the shipped teardown-skip (a legitimate, standard technique). This
session's contribution: **re-verified the mitigation is robust (0 crashes / 60 builds of the
c_3map repro, output exe exit 0)** and did an **independent fresh root-cause pass** that
corroborates the prior sessions' conclusion that the residual OOB is not findable by inspection:
- Hardcoded struct strides are CORRECT (Symbol = 24B exactly: 4+1+1+2+4+4+4+4; Scope = 32B
  exactly: 1+1+2+4+4+[pad4]+8+4+4). Not a stride bug.
- `clone_subtree_from` (ast.ax:238) is index-based and re-reads `self.nodes.data` on every access
  → safe across the reallocs it triggers.
- `substitute_type_params` (mono.ax:70) and `remove_generic_params_child` (mono.ax:215) hold
  `&self.tree.nodes.data[node_idx]` but write through it only BEFORE any recursion, and neither
  grows `self.tree.nodes` → the held pointer never dangles.
- The `HashMap[K,V]` struct is 4 pointers + 2 i64 = 48 bytes REGARDLESS of K/V (all fields are
  ptr/i64), so the generic-inst size machinery is NOT K/V-dependent → not the trigger.

**Engineering decision (§10, correctness/cost):** for a LOW-severity, output-always-correct,
already-eliminated heisenbug, a dedicated custom-allocator memory-watchpoint session has
disproportionate cost. The teardown-skip is the definitive fix. Everything below is retained
history for anyone who ever revisits with a Windows-capable memory watchpoint.

---

# ⛔ VALGRIND HUNT BLOCKED 2026-07-19b — bug is WINDOWS-heap-specific; Linux MASKS it

User installed valgrind in WSL to enable the sanitizer hunt. RFC 0009 P3 Linux self-hosting
was completed this session (`c93446f`), so the compiler now RUNS on Linux and CAN be driven
under valgrind. **But the bug does not reproduce on Linux.** Built a mitigation-OFF debug
compiler (teardown free-chain re-enabled) for BOTH targets and stress-ran the c_3map repro:
- **Windows debug (mitigation off): 5/30 crashes** (~17%, matches the historical ~15% rate) —
  reproduces reliably.
- **Linux debug (mitigation off): 0/30 crashes** — does NOT reproduce.

The corruption is **heap-layout-dependent**: Linux's mmap allocator lays memory out differently
(different base/alignment vs Windows VirtualAlloc), so the OOB write lands on different adjacent
data and does not clobber the load-bearing pointer that faults on Windows. The corruption likely
still HAPPENS on Linux but silently (writes a harmless location) — consistent with "output always
correct". **Consequence: valgrind (Linux-only) cannot catch this bug** — the tool the user
installed doesn't apply, because the manifestation is Windows-specific. Windows-side page-heap
(gflags) also can't help: the AXIOM allocator uses VirtualAlloc, not the C HeapAlloc that gflags
page-heap instruments. So the ONLY viable tool remains the in-allocator **AX_CANARY** (already
built, in std/mem/alloc.ax, works on Windows where the bug reproduces) — and the remaining
suspect (a pointer-field clobber in a compiler struct, which the canary can't easily catch) still
needs a memory-watchpoint / custom-allocator-block-tracking session. **The hunt is NOT advanced by
valgrind; status unchanged (mitigated, low-sev, root cause = future AX_CANARY-watchpoint session).**

---

# MITIGATED 2026-07-19 (crash eliminated; root-cause narrowed, not fully pinned)

**Fix shipped:** on a successful `-self-link` build, the driver now flushes and returns
immediately after the output is written, **skipping the teardown free-chain**
(main_air.ax, end of the self-link branch). The teardown was the only place the
corrupted compiler-heap pointer got walked; skipping it (a standard compiler pattern —
the OS reclaims all memory at exit) makes the exit code reliably 0. **Verified: 0 crashes
in 160 stress runs** (n4a 0/60, c_3map 0/60, t_hashi64 0/40; baseline ~23%/~12%/~15%),
fixpoint A==B `84D204E8`, regression 435/435. Also removed a dead post-Stage-6
`temp_obj_nt` alloc+free (its file-removal was commented out) = the first suspected
post-Stage-6 heap op.

**Diagnosis advanced this session (AX_CANARY allocator experiment, committed disabled in
std/mem/alloc.ax):** built a size-classed footer-canary + whole-heap sweep + free-list
integrity verifier + per-pop guard into the compiler's own AXIOM allocator (the running
Windows compiler's `ax_alloc` is bundled from `std/mem/alloc.ax`, NOT the C
`ax_runtime.dll` — the DLL only provides global-state/panic/string helpers; `ax_alloc` is
compiled INTO the exe). Findings:
- **NOT a contiguous ≤16-byte adjacent overflow:** adding +16 bytes of slack to every
  block only HALVED the crash rate (23%→~7%, layout perturbation like gdb), and a
  block-END footer sweep never tripped.
- **NOT free-list-link corruption:** a per-pop guard validating every `ax_free_list_pop`
  head (alignment/range + block `gen_id==0`) and a free-list integrity sweep BOTH never
  tripped, yet the SEGV persisted.
- **=> Remaining suspect:** an OOB/indexed write clobbering a POINTER field inside a
  compiler data structure (Vec `.data`, a tree/node index, etc.) during hash-container
  generic mono, which only faults when a later deref/free touches it. Pinning the exact
  write needs a memory watchpoint / symbolized ASAN (still blocked: self-host binary has
  no DWARF). The AX_CANARY infra (const `AX_CANARY=false` in std/mem/alloc.ax) is
  committed for that future session — flip it true, rebuild, and extend the sweep.
- **Incidental discovery:** the compiler routinely writes INTO the size-class slack beyond
  its requested `@alloc(sz)` size (benign rounding-reliance) — a near-`sz` footer collided
  with it; that's why the footer had to sit at the block END.

Everything below is the pre-mitigation investigation (kept for the ASAN session).

---

# (historical) OPEN bug — intermittent teardown segfault: 3+ HashMap monomorphizations

## Symptom
Compiling a program that instantiates **three or more distinct `HashMap[K,V]` types** intermittently
(~8%, 2/25) segfaults `axc_native` (rc=139). **The output executable is ALWAYS produced and runs
correctly (exit 0)** — the crash happens AFTER `[Debug] Stage 6: Finished self-linking.`, i.e. during
process teardown, not during compilation or codegen. So it is a **LOW-severity** stability/robustness
bug: correct output, but a spurious nonzero compiler exit code ~8% of the time (would cause flaky CI on
such programs). Flag-independent — happens on a plain `-self-link -O1` build, NOT related to
`-ctgc-free` (the CTGC investigation that found it was a red herring; see below).

## Minimal repro (`scratch/crash/c_3map.ax`, banked here — do NOT add as a regression oracle, it is FLAKY)
```
import std.collections
fn main() -> i64:
    mut m1 = HashMap[i64, i64].new()
    m1.insert(1, 100)
    mut m2 = HashMap[str, i64].new()
    m2.insert("a", 200)
    mut m3 = HashMap[i64, f64].new()   // the THIRD distinct instantiation triggers it
    m3.insert(1, 3.5)
    return 0
```
Stress: `for i in $(seq 1 25); do bin/axc_native.exe build c_3map.ax -o t.exe -self-link -O1 >/dev/null 2>&1; [ $? -ge 128 ] && echo crash; done`

## Isolation (measured 2026-07-18, driver 11EB77BD)
- **2 instantiations** (`[i64,i64]` + `[str,i64]`): 0/25 crashes. **3 instantiations** (add `[i64,f64]`):
  2/25. So the trigger is the 3rd distinct HashMap monomorphization (count crossing 3, and/or the
  f64-value variant). Single-map programs (`c_hash`, `c_hashget`): 0/20.
- `t_hashi64.ax` (the regression oracle, has exactly these 3 instantiations): 3/20 normal-build crashes,
  1/12 under `-ctgc-free`, 1/15 normal — i.e. flag-independent, same ~10-15% rate.
- Other generic-heavy programs (`t_gentree`, `t_vecstructopt`, `t_forvec`): 0/20-25. So it is specific
  to **multiple HashMap instantiations**, not generics in general.
- Crash is post-`Stage 6` (teardown). Output exe verified present + correct (exit 0) on a crashing run.

## ⭐ ROOT-CAUSE ADVANCE (2026-07-18, teardown instrumentation) — it's an OOB HEAP WRITE corrupting allocator metadata
Instrumented every teardown `@free` in main_air.ax (self-link path, ~L1252-1338) with flushed
static-string markers, built a throwaway compiler, stress-ran the n4a repro (2 struct-value HashMaps).
On the crash: **NOT ONE marker printed** — the segfault lands right after `[Debug] Stage 6: Finished
self-linking.` and before the very first post-Stage-6 marker `[TD0a]`, i.e. at the FIRST heap
operation after self-linking (the `ax_puts_local`/`fflush`/small alloc for the marker itself). So the
crash is NOT in freeing a specific compiler structure (tree.nodes/typetable/symtable) — it's the
allocator's own metadata (free-list/heap header) that is corrupted, and the next heap op that reaches
the damaged entry faults. **=> The hash-container-mono bug is an OUT-OF-BOUNDS HEAP WRITE**, not a
held-pointer dangling. Intermittent because whether/when the allocator reaches the clobbered free-list
entry depends on the allocation sequence.
- **This supersedes the "held-ptr-across-realloc" hypothesis** (kept below for history). Prime new
  suspect: a SIZE MISCALCULATION for a hash-container internal array (`keys`/`values`/`occupied`) during
  monomorphization → an `@alloc` too small or a `@memcpy`/store past the end → clobbers an adjacent
  heap block's allocator header. Strongly connects to [[next-step-14-sumtype-size-bug]] (generic-inst
  `builder_type_size_and_align` mis-sizing) — the same size-machinery class. Aggregate value types (the
  n4a struct-value trigger at just 2 maps) fit: a bigger/mis-sized value element makes the wrong-size
  write land sooner/larger.
- **Next fix step (no ASAN needed):** the allocator is C (`runtime/axalloc/axalloc_compiled.c`), which
  IS symbolizable — add heap canary/red-zone checks there (or build it with `-fsanitize=address`) and
  run the C-backend compile of n4a to catch the exact overflowing `@alloc`/store. OR instrument
  `builder_type_size_and_align` / the hash-container internal-array alloc sites in air_builder for a
  size mismatch on the 2nd/3rd hash-container. This is now a concrete OOB-write hunt, much more tractable
  than the earlier "held-ptr among 30 sites" framing.

## Sharpened isolation (2026-07-18, no-rebuild repro matrix)
- **It is the number of DISTINCT HashMap monomorphizations, not concrete types, not f64:**
  3 HashMaps with keys i64/str/i32 all-i64-value (`a_3map_nof64`) = 5/30; 2 maps incl. an f64-value
  (`c_2map_f64`) = 0/30. So f64 is innocent; the trigger is the 3rd distinct `HashMap[K,V]` instance.
- **It is DISTINCT instantiations, not local count / usage:** 3 locals of the SAME `HashMap[i64,i64]`
  (`d_3samemap`) = 0/30.
- **It is HashMap-SPECIFIC, not generics or 2-type-param generics in general:** 3 distinct `Vec[T]`
  (`b_3vec`) = 0/30; 3 distinct instantiations of a user 2-param `Pair[A,B]` (`e_3pair`) = 0/30.
  Only HashMap (a large stdlib generic with many methods + nested keys/values/occupied arrays) trips it.

## Code-path localization (2026-07-18, read-only)
Narrowed to the monomorphizer clone path: `mono.ax::instantiate_function` (L423) →
`AstTree::clone_subtree_from` (ast.ax:238). For a stdlib generic like HashMap the template lives in the
SAME tree, so `src_tree = self.tree` (mono.ax:433, the default; only overridden if a separate
`symbol_trees` entry exists). Thus cloning HashMap's LARGE method set GROWS `self.tree.nodes`,
`self.tokens`, and repeatedly reallocates `self.src` (ast.ax:258 `self.src = concat(old_src, tok_text)`
runs per token-bearing cloned node — hundreds of times for HashMap) WHILE reading from the same tree.
Each distinct HashMap instantiation repeats this; 3× crosses a realloc/heap threshold intermittently.
The three hot functions I inspected — `clone_subtree_from` (extracts orig fields to SCALARS before
`add_node`), `substitute_type_params` (holds `&self.tree.nodes.data[node_idx]` at mono.ax:74 but does
NOT grow the tree within its own body), and `remove_generic_params_child` — each look individually
safe, so the corrupting write is NOT obvious by inspection; prime remaining suspects: (a) a held
pointer into `self.src`/`self.tokens`/`self.nodes` across one of the many `clone_subtree_from` reallocs
that only dangles once the buffers are large enough (3rd instantiation), or (b) typetable/symtable
growth while registering the 3rd HashMap's monomorphized methods. **Confirming the exact line needs a
memory sanitizer** (build the Linux ELF target, run the 9-line repro under valgrind → faulting free +
the corrupting store). This localization + the minimal repro should make that a short tooled session.

## REFINEMENT 2 (2026-07-18) — trigger is a CUMULATIVE hash-container mono-complexity threshold, not a fixed count of 3
Stress-probing novel combos revised the "3+ distinct" framing: **2 distinct HashMaps with a STRUCT
(aggregate) value type crash** — `HashMap[i64,P]` + `HashMap[str,P]` where `struct P{x,y}` (n4a) = 3/20,
even with NO get/match (n4b, one struct-value map + get + match = 0/20, so the Option/match path is NOT
the trigger). Yet 2 distinct SCALAR-value maps (c_2map) = 0/20 and 3 scalar maps = crash. So the trigger
is the **cumulative hash-container monomorphization WORK crossing a threshold**, not a literal count:
aggregate value types (or richer K/V) do more mono work per instantiation, tipping it at 2 instead of 3.
**This raises reachability**: just 2 `HashMap`/`HashSet` with struct values (a plausible real pattern)
can hit it — still LOW severity (correct output, teardown-only) but more reachable than first thought.
Consistent with the held-ptr-across-realloc theory: more per-instantiation mono nodes/types cross the
buffer-realloc boundary sooner. (8 distinct Vec still clean — the corrupting path is hash-container-mono
-specific, and struct values add work WITHIN that path, not general mono volume.)

## REFINEMENT (2026-07-18) — it's HASH-CONTAINER mono, not HashMap-specific, not clone-volume
- **HashSet is ALSO affected:** 3 distinct `HashSet[T]` (`hs3`) = 2/30; a MIX of 2 HashMap + 1 HashSet
  (`mix`) = 2/30. So the trigger is the CUMULATIVE count (≥3) of distinct **hash-based container**
  monomorphizations (HashMap and HashSet together), not HashMap alone.
- **NOT clone-volume:** 8 distinct `Vec[T]` instantiations (`vec8`, far more total cloning than 3 hash
  containers) = 0/30. So the big-template / many-reallocs theory is WRONG on its own — Vec has a big
  template too and doesn't trip it.
- **=> The corruption is specific to what HASH containers monomorphize that Vec does NOT:** a per-key-type
  **hash function + equality** (and the `occupied`/probing internals + nested keys/occupied arrays).
  The likely corruption site is the monomorphization of the hash/eq machinery (or nested-array type
  registration) for the 3rd distinct hash-container key type — NOT the generic clone path in general.
  This meaningfully narrows the fix search: look at how HashMap/HashSet's `hash`/`eq`/probe methods and
  their internal `keys`/`values`/`occupied` arrays are instantiated, for a held pointer / stale index
  across a typetable/symtable/tree realloc that triggers on the 3rd hash-container.

## Element-size audit COMPLETE (2026-07-18) — every vector grow is correctly sized; OOB is NOT a wrong-elem-size
Audited every `@alloc`/`@memcpy` element multiplier in the compiler's allocation surface that mono
touches — all correct: **typetable.ax** all use `size_of[T]()` (robust); TypeTable `@alloc(120)` = exactly
5×24-byte Vecs ✓. **resolver.ax** hardcodes verified: Symbol `*24` ✓ (4+1+1+2+4+4+4+4), Scope `*32` ✓
(4/4/4/pad4/ptr8/4/4), ScopeEntry `*8` ✓, ModuleExport `*8` ✓, symbol_trees ptr `*8` ✓, generic-args
`*4` ✓. **ast.ax** NodeVec `*24` ✓, TokenVec `*8` ✓, IntVec `*4` ✓. **mono.ax** TypeSubstVec `*8` ✓.
=> the OOB heap write is NOT a wrong vector-element-size. Remaining candidates for the fix session:
(a) an INDEXED write past a vector's cap without a grow (`vec.data[i]=x` with i≥cap) somewhere in the
hash-container mono path, or (b) a `@memcpy` with a COMPUTED (not size_of) length that's wrong for a
hash-container element, or (c) a hash-container-specific mono op. These need data-flow analysis or the
canary — NOT another element-size pass (this one is exhaustive). One thing worth a `size_of` double-check
if revisiting: Scope's `*32` hardcode vs the compiler's computed `size_of[Scope]` (if AXIOM packs to 28
the STRIDE could mismatch — but that'd be universal, not hash-specific, and fast_fixpoint is reliable,
so it's almost certainly consistent).

## RULED OUT by inspection (do NOT re-check these in the fix session)
- **All growable-vector element sizes are CORRECT** → not a wrong-size memcpy/alloc overflow:
  `NodeVec.push` 24B (ast.ax:145/147, AstNode=24B), `TokenVec.push` 8B (lexer.ax:20/22, Token=8B),
  `IntVec` 4B (lexer.ax:43/45), `TypeSubstVec.push` 8B (mono.ax:28). Each frees the old buffer after
  memcpy correctly.
- `clone_subtree_from` extracts the source node's fields to SCALARS (ast.ax:243-248) before `add_node`,
  and re-indexes `src_tree.nodes.data[...]` fresh in its child loop → no stale-pointer there.
- `substitute_type_params` holds `&self.tree.nodes.data[node_idx]` (mono.ax:74) but does NOT grow the
  tree within its own body → that held pointer is stable during the call.
- The intermittency rules out a deterministic fixed-capacity overflow. Remaining hypothesis = a held
  pointer into `self.src`/`self.tokens`/`self.nodes` across one of `clone_subtree_from`'s reallocs that
  only dangles once buffers are large enough (3rd HashMap), OR corruption while registering the 3rd
  HashMap's monomorphized methods/types (typetable/symtable growth). Confirm with ASAN.

## Scope is FULLY bounded (2026-07-18) — no other generic type can be affected
Enumerated every generic type in std/: the only COMPILABLE complex generics are `Vec[T]`, `HashMap[K,V]`,
`HashSet[T]` (std/collections.ax). All others — `iter` (Map/Filter/Take…), `sync` (Atomic/Mutex/Channel…),
`gpu` GpuBuffer, `mem` Box — use brace-`{}` spec-dialect syntax that DOESN'T parse, or are
threading/aspirational-blocked. Vec is stress-clean (even 8 distinct instantiations); both hash
containers crash. => the bug is provably confined to HashMap/HashSet and there is no further generic
surface to test — the stress-probe avenue for this bug is exhausted.

## Bounding result (2026-07-18) — the bug is ISOLATED; compiler is otherwise stress-stable
Stress-probed 14 diverse feature-heavy oracles 18× each (t_gentree, t_optvecnest, t_vectupmix,
t_licmchain, t_selfrec, t_inlinecf3, t_genfloatret, t_deepnestmut, t_optstructpay, t_globstr,
t_forvec, t_structoptfield, t_vecstructopt, t_genswaphet) — **ALL 0 crashes**. So this intermittent
teardown crash is NOT a symptom of pervasive heap instability; it is confined to the 3+ distinct
hash-container mono case. Combined with "1-2 distinct hash-containers = clean", the **practical impact
is near-zero** (real programs rarely instantiate 3+ DISTINCT `HashMap`/`HashSet` key/value type combos;
t_hashi64 is a contrived 3-instantiation oracle). Confirms LOW severity + LOW urgency. Reusable
technique: intermittent heap bugs are invisible to compile-once probing — STRESS-probe (build N× and
watch exit codes) to surface them.

## Audit note (2026-07-18) — the held-ref-into-vector pattern is widespread; don't hand-audit, use ASAN
The suspected bug class (`mut x := &self.<vec>.data[i]` held across a same-vector grow) appears at 30+
sites (air.ax, resolver.ax, ssa_opt.ax:257/397/643/1072/1485/1537, mono.ax:74/216, air_builder.ax:309,
etc.). Almost all are safe (the held ref is not used after a grow to THAT vector in-scope). Isolating
which one fires in the hash-container mono cascade by hand is infeasible — ASAN/valgrind pins the exact
faulting read+write in a single run of the 9-line repro. So the fix session should go straight to the
sanitizer rather than auditing these sites manually.

## Suspected root cause (NOT confirmed — needs tooling)
Heap corruption introduced while monomorphizing the 3rd HashMap instantiation (mono.ax / typetable /
the generic-instance size+align machinery), latent until the cleanup `@free` chain at the end of the
per-file compile (main_air.ax ~L1676-1688: frees lexer.tokens / indent_stack / newline_offsets /
mod_resolver / mod_checker / parser_ptr) walks the corrupted allocator metadata and faults. Because the
corruption only manifests at teardown, OUTPUT is unaffected — consistent with corruption in a structure
touched only during mono bookkeeping + final free, not in the emitted code. Intermittent => depends on
heap layout (allocation sizes/ordering), classic use-after-free / double-free / OOB-write signature.

## Why the regression suite/fixpoint never caught it
The gate builds each program ONCE and checks the OUTPUT exe's exit code; the exe is always correct, and
a 1-in-~12 teardown crash rarely hits on a single build. The fixpoint checks output determinism (holds
— the crash is post-output). So this slipped through as a rare, output-invisible teardown flake.

## gdb path (checked 2026-07-18) — heisenbug resists it; even a catch shows only the symptom
`gdb` IS available on Windows (`C:\msys64\ucrt64\bin\gdb.exe`, used to root-cause the next-step-14
native bugs). BUT driving the compiler under gdb on the repro = 0 crashes in 25 runs (native rate ~15%
→ 0.85^25≈1.7% odds of that if unperturbed) — **gdb perturbs the heap layout and masks the intermittent
crash** (classic heisenbug). The crash rate CAPS at ~10-15% (a max-stress 5-hash-container+big-struct
repro `big.ax` = 2/20, no higher), so it can't be made deterministic by piling on instances. AND even a
successful gdb catch would land on the LATER heap op that walks the corrupted free-list, not the
original OOB write (they are separated in time) — so it shows the symptom, not the cause. Pinning the
write needs a hardware WATCHPOINT on the clobbered address (unknown ahead of time) or a verify-on-every-
alloc debug allocator that panics right after the write. => the two non-invasive tools (gdb,
C-backend-canary) both fail; the fix needs the INVASIVE self-link-allocator canary (below) or a source
audit. This is why it's a dedicated session, and genuinely so — not a lack of trying.

## Debug-allocator path is BLOCKED (checked 2026-07-18) — C backend can't build the compiler
Tried to catch the OOB write with the C runtime's canary allocator: `runtime/axalloc/axalloc_compiled.c`
has `AX_ALLOC_DEBUG` (header+footer magics, `verify_all_allocations` on every alloc/free, `ax_free`
footer-canary check is even UNCONDITIONAL). It's linked by the **C backend** (`-use-gcc`, main_air.ax:
1301), NOT the self-link path. BUT: **`-use-gcc` build of the whole compiler crashes DETERMINISTICALLY
at typecheck (3/3), even WITHOUT the debug allocator** — the C backend is broken/unmaintained for a
program the size of the compiler (fast_fixpoint's `-self-link` build is reliable, so this is a
C-backend-specific limitation, not the hash-mono bug). So I could not produce a canary-instrumented
compiler this way. The **self-link daily driver uses a different, freestanding allocator WITHOUT the
axalloc canaries** — which is why the OOB write is a raw SIGSEGV instead of a clean canary panic.
**Remaining fix approaches (pick one in a dedicated session):**
1. Add a footer-canary / verify hook to the SELF-LINK allocator = **`bootstrap/runtime/axalloc.ax`**
   (located 2026-07-18: a 553-line size-classed segment bump allocator; `@alloc`→OP_ALLOC→its ax_alloc;
   `ax_os_alloc`/`ax_segment_bump_alloc`/size-class free lists). Adding a per-block footer canary +
   a check-on-free that reports the overflowed block SIZE would pin the alloc site — BUT it is
   size-class-aware and self-host-critical, so instrument carefully (behind a flag; verify the daily
   self-build still reproduces) and rebuild axc_native (self-link) with it, then run on n4a. This is the
   most direct path but a dedicated, careful session (do NOT rush — the compiler runs on this allocator).
2. Repair the C backend enough to build the compiler, then use AX_ALLOC_DEBUG (bigger yak-shave).
3. Source audit of `builder_type_size_and_align` (typetable/air_builder) for a generic-inst / nested
   hash-container-array size miscalc that under-sizes an `@alloc` → the overflow. Speculative w/o (1).

## Tooling reality (checked 2026-07-18) — why the fix is a real session, not a tick
WSL2 (Ubuntu 26.04) IS available, but `valgrind`/`gdb` are NOT installed, and — decisively — the
self-hosted compiler binary carries NO debug symbols (no DWARF), so a sanitizer would report raw
addresses, not AXIOM source lines. So pinning the corrupting write requires one of: (a) `apt install
valgrind` + a way to map addresses back to functions (symbol table / a `nm`-able build), (b) printf
instrumentation bisecting the hash-container mono path (slow, and adding code perturbs the heap layout
that the intermittent bug depends on — may mask it), or (c) adding debug-info emission to the backend
first. All are multi-step and disproportionate to a LOW-severity, near-zero-impact bug — hence deferred
until hash-container/generic-mono work is being done anyway, or the user prioritizes it.

## Fix path (deferred — dedicated tooled session)
Root-causing an intermittent heap corruption needs a memory sanitizer. The compiler self-hosts to a
Linux ELF (RFC 0009 P3) — build the Linux target and run `valgrind`/ASAN-equivalent on `c_3map` to get
the faulting free + the corrupting write. Alternatively, bisect the mono path for multi-HashMap
instantiation (compare 2-map vs 3-map AIR/typetable state). LOW urgency (output correct); schedule when
generic-mono or HashMap work is next touched. Related: [[next-step-14-sumtype-size-bug]] (generic-inst
size machinery), [[bug66-hashmap-i64-value-corruption]] (HashMap value handling).

## ⚠️ Corrects an earlier MISATTRIBUTION this session
The CTGC activation attempt's broad `-ctgc-free` sweep flagged `t_hashi64` crashing, and it was WRONGLY
recorded (commits `f4299d5`/`d4285b0`/`1de7385`, [[ctgc-p3-scoping-2026-07-18]]) as an
"OP_DESTROY-on-container free-glue codegen crash" blocking activation. **That was wrong** — `t_hashi64`
crashes at the SAME ~15% rate on a NORMAL build with no `-ctgc-free`, i.e. it is THIS pre-existing
teardown bug, unrelated to CTGC free-glue. Consequence: the CTGC general-free **activation's actual
status is INCONCLUSIVE, not "blocked by free-glue"** — it passed every deterministic gate (A==B==C
`28DCDE0A`, compiler freeable=0, regression 435/435, ctgc_free_check 10/10) and its only sweep red-flag
was this unrelated flake. A clean re-validation just needs a sweep that discounts programs which also
crash on the clean compiler. The escape-soundness half remains correct and shipped.
