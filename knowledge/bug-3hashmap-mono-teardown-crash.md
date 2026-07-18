---
name: bug-3hashmap-mono-teardown-crash
description: "OPEN (low-severity): compiling a program with 3+ DISTINCT HASH-CONTAINER monomorphizations (HashMap and/or HashSet, cumulative) intermittently (~8%) segfaults the compiler AT TEARDOWN (after Stage 6 self-linking). OUTPUT exe is always produced and correct (exit 0) — only the compiler's process exit crashes. Flag-independent (not -ctgc-free). NOT clone-volume (8 distinct Vec = clean) — specific to hash containers' per-key-type hash/eq mono + occupied/keys arrays. Minimal repro included. Needs ASAN/Linux to root-cause."
metadata:
  type: project
---

# OPEN bug — intermittent teardown segfault: 3+ HashMap monomorphizations

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
