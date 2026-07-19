---
name: rfc0009-p3-elf-linux-target-wip
description: "SHIPPED (milestone): Linux ELF64 target. `axc build p.ax --target linux -self-link -O1` emits a runnable ELF executable, VERIFIED under WSL Ubuntu (elf42=42, elfloop=7, elfglob=55==Windows). SysV ABI, inline syscalls, ELF .text/.rodata/.data (RFC 0017 globals), freestanding runtime bridge, exit(2) via syscall. Gate B==C=A0B44A50, 337/337 (+elfglob oracle). Scope: pure-compute + globals; heap/print/actor-runtime on Linux = follow-up."
metadata:
  node_type: memory
  type: project
  originSessionId: 3228306b-52d7-4378-bb1c-a0b6cef57eba
---

# ⭐ COMPILER SELF-HOSTS ON LINUX — SHIPPED 2026-07-19b (`c93446f`)

The compiler now **RUNS as a compiler on Linux**, not just cross-compiles to it. Blocker was
`get_freestanding_args` (main_air.ax): the ELF `main` IS the entry, so main's prologue clobbers
the kernel's stack-passed argc/argv before the arg reader runs, and the Linux `else` branch was
empty (`let dummy = 0`) → argc 0 → the Linux binary always printed usage. Fix: on Linux read
`/proc/self/cmdline` (NUL-separated argv) via SYS_open/read/close, splitting into the freestanding
args array (mirrors std/os.ax::linux_read_file). Inert on Windows (is_windows folds true → dead
branch) → **A==B==C=`f0885975`**, full Windows regression 439/439. Validated under WSL: a
Linux-built compiler parses `build … --target linux -self-link -O1`, reads source, self-links, and
emits a runnable ELF whose output runs correctly (HashMap → 35). Remaining P3 refinements (split
PT_LOAD RX/RW, DT_INIT/fini, static ELF) are non-blocking. NB: this UNBLOCKED but did NOT crack the
3-hashmap teardown bug — that's Windows-heap-specific and Linux masks it (0/30), see
[[bug-3hashmap-mono-teardown-crash]].

# Linux ELF64 target — SHIPPED milestone (2026-07-17)

`axc build prog.ax --target linux -self-link -O1` produces a **runnable Linux ELF64
executable**, verified end-to-end under **WSL Ubuntu**: `elf42`→42, `elfloop`→7 (while-loop),
`elfglob`→55 (module global + recursion(fib) + calls + control-flow) — and `elfglob` returns
**55 on Windows COFF too** (same program, both targets). Gate `scripts/elf_linux_check.sh`.

The ELF path had been scaffolded but NEVER run: `@compiler_intrinsic("is_windows")` was
hardcoded to 1. This was an unfinished target port; brought it to life.

## Everything shipped (one commit)
1. **Driver `--target linux`** (main_air.ax): flag → threads `is_windows=0` into the AIR
   builder (`target_is_windows` field on AirModuleBuilder, default true → self-build stays
   Windows/A==B) so the bundled runtime folds to its syscall paths; passes obj format "elf".
   Also bundles `bootstrap/runtime/panic.ax` for Linux (freestanding runtime shims).
2. **Intrinsic fold** (air_builder.ax ~L1573): is_windows/is_linux/os_name/path_separator now
   branch on `self.mb.target_is_windows`. 7-arg `syscall` extern added to the inline-syscall
   list (== syscall6: arg0→RAX, args1..6→RDI/RSI/RDX/R10/R8/R9).
3. **`main` exit** (x86_selector.ax OP_RETURN): SysV `main` IS the ELF entry (no CRT to `ret`
   into) → emit `mov rdi,rax; mov rax,60; syscall` (exit(2)); skip trailing `ret`. The eager
   `__ax_runtime_init`(-4)/`shutdown`(-5) actor-runtime calls are GATED to win64 (see scope).
   ⚠️ BUG fixed here: an immediate move MUST use `MACH_MOV_IMM`, not `MACH_MOV` — MACH_MOV
   resolves its source through the reg table and silently emits NOTHING for an OPND_IMM (this
   dropped `mov rax,60`, so exit ran with RAX=42 → syscall #42=connect → segfault).
4. **ELF object writer** (x86_elf64.ax): fixed a heap-overflow (`sec_names` `4*8`→`n*size_of[str]`,
   str=16B — the first-ever-write segfault). Rewrote elf64_serialize to a fixed 7-section layout
   whose PROGBITS indices MATCH the linker's canonical COFF scheme: 0 NULL | 1 .text | 2 .rodata |
   3 .data | 4 .symtab | 5 .strtab | 6 .rela.text. Now takes rdata+data params.
5. **Codegen globals for ELF** (x86_coff.ax `compile_native_binary`): emit module globals for
   ELF too (STT_OBJECT, section 3=.data), pass rdata_buf+data_buf to the writer. Fixed an
   elf_sym_map off-by-one (defined syms are at symtab index i+1 because of the NULL entry).
6. **Linker** (linker.ax): `linker_parse_elf` routes PROGBITS by sh_flags (EXECINSTR→.text,
   WRITE→.data, else→.rodata) and remaps section indices to the canonical 1/2/3; fills obj
   rdata+data. ELF exe emit: lays out a writable data region after code+dyn (base-relative
   offsets → standard reloc resolution), registers each global's VA into func_names/offsets
   before relocation, PT_LOAD made RW, writes the region. Dead Windows/ax_runtime.dll imports
   (kernel32 + ax_runtime buckets zeroed for elf64) downgraded to warnings + patched to 0
   (they sit in DCE-surviving dead code — no whole-fn DCE yet). Unresolved C-ABI runtime names
   bind to their bundled `ax_`-twin (generalized syscall→ax_syscall). Added **DT_JMPREL** (was
   missing while PLTRELSZ≠0 → ld.so walked a NULL rela array → segfault); fixed PT_DYNAMIC file
   offset (`512 + dyn_file_off + offset_dynamic`, not `512 + offset_dynamic` which pointed into
   .text); size_dynamic +1 tag.

## Gate — GREEN
Backend change but ELF additions are INERT on the Windows self-build → **A==B==C = `A0B44A50…`**
(the earlier syscall-inline transition already settled in the seed). Full regression **337/337**
(+ new `elfglob` COFF oracle = 55). ELF smoke: `AXC=bin/axc_native.exe bash scripts/elf_linux_check.sh`.
Oracles: `bin/elf42.ax`(42) `bin/elfloop.ax`(7) `bin/elfglob.ax`(55). WSL run form (git-bash):
`export MSYS2_ARG_CONV_EXCL="*"; wsl /mnt/d/projects/compiler/Axiom/bin/x.elf; echo $?`
(NB: `wsl bash -c "...; echo $?"` LOSES the exit code — capture wsl's own `$?` on the git-bash side).

## ✅✅ HEAP ON LINUX — SOLVED & SHIPPED `59c61d5` (2026-07-17). Digs #1-6 below = history.
**ROOT CAUSE (the real blocker): OP_SYSCALL register-shuffle bug.** The selector moved each
syscall arg vreg DIRECTLY into its fixed register (num→RAX, a1→RDI … a6→R9); an arg whose
regalloc home was a *later* target register got clobbered before it was read. All-constant
syscalls (print) worked; any VARIABLE arg (`mmap(size,…)`) got garbage flags/fd → -EBADF →
allocator slab/segment mmap returned junk → SIGSEGV/OOM. **Fix (x86_selector OP_SYSCALL):**
spill every arg to the SysV red zone (128B below RSP, untouched by `syscall`), then load the
fixed registers from there — store-then-load breaks all register deps, no RSP adjust. Plus:
std/mem/alloc.ax slab else-mmap (memset-null) + ax_os_alloc neg-errno check; std/runtime.ax
gates actor-system/scheduler to win64 (thread spawning not ported — Vec/String/HashMap don't
need it); main_air bundles bootstrap/runtime/syscall.ax for --target linux. Gate A!=B, B==C
`D99A78FB`, 338/338, elf_linux_check 8/8 (Vec=42/HashMap=42/String "Hello, Linux!"). The digs
below (memset-null, ax_alloc mangling, collision theories) were WRONG TURNS — the live path is
std/runtime.ax `malloc`→__ax_runtime_init→std/mem/alloc.ax (NOT axalloc.ax, which is dead), and
the crash was the OP_SYSCALL bug all along. Remaining Linux gap: actor/scheduler (threads).

## (history) Heap/Vec/String on Linux — DIAGNOSED (2026-07-17), was blocked
Probed after the stdout fix: `@alloc`, `Vec.push/get` → SIGSEGV on Linux. Re-enabling
`__ax_runtime_init` (ungate -4/-5) does NOT help — it makes EVEN `elf42` crash **inside
libc.so.6** (`ip in libc[...]`, segfault at 0). Narrowed down (2026-07-17):
- **mmap WORKS on Linux** (verified: `syscall(9,0,4096,3,0x22,-1,0)` maps + RW round-trips →
  77). So ax_os_alloc is fine — do NOT re-chase mmap.
- Therefore the fault is DEEPER in `__ax_runtime_init`'s chain: `ax_segment_manager_init` /
  `ax_actor_heap_create` / `ax_actor_system_init` — a libc `memset`/`memcpy` on a bad/null
  ptr, or scheduler threading (libc.so.6 is a real DT_NEEDED from calloc/malloc/free/memset).
### Heap dig #2 (2026-07-17 pm) — ROOT of the libc crash found + a mangling snag (both reverted)
Attempted the freestanding-@alloc approach; got PARTWAY, reverted to keep main clean at
`e5be7e2`. Two concrete findings (re-apply next session):
1. **SLAB BUG (ready fix).** `std_mem_alloc_get_slab` (std/mem/alloc.ax:143-149) allocates the
   196608-byte slab-metadata region ONLY `if @compiler_intrinsic("is_windows")` (VirtualAlloc) —
   **no else branch** → on Linux returns NULL → `ax_segment_manager_init` does
   `@memset(slab, 0, 196608)` = **memset(NULL) → the exact libc crash** (`ip in libc`, addr 0).
   FIX (verified compiles): add `else: p_g_slab.* = syscall(9 as u64,0,196608,3,0x22,
   0xFFFFFFFFFFFFFFFF,0) as ptr[Segment]` (mmap; `syscall` extern already declared at L81).
   This is inert on Windows (is_windows branch) → A==B safe.
2. **ax_alloc wiring + mangling snag.** `@alloc`→OP_ALLOC→`MACH_CALL imm=-1`→resolve_binary_sym_name(-1)
   =`"ax_alloc"` (bare). Real `ax_alloc`/`ax_free`/`get_global_heap` live in
   bootstrap/runtime/axalloc.ax (NOT bundled; Windows gets them from ax_runtime.dll). The BUNDLED
   std/mem/alloc.ax has the low-level `ax_actor_alloc`/`ax_actor_heap_create`/
   `ax_segment_manager_init`/`get_global_state`(+AxGlobalState, global_heap field @off40) but NO
   ax_alloc. I added a lazy-init `pub fn ax_alloc/ax_free/ax_realloc` to std/mem/alloc.ax (init
   segment-mgr+heap on first use, route to ax_actor_alloc, NO actor-system/scheduler). Result:
   the object then had **TWO** symbols — bare `ax_alloc` (341B) AND `ax_ax_alloc` (140B) — and it
   still crashed. UNRESOLVED: which one OP_ALLOC binds to on Linux, and why two exist (mangler
   handling of `ax_`-prefixed names + MODDUP collision machinery). **NEXT:** disassemble both
   symbols to see which is called + whether the slab fix took effect in that build; the bare
   ax_alloc still hit -9 (an mmap error used as a ptr) → confirm the slab-fix path actually runs.
   Cleaner alternative to dodge the name collision: name the Linux allocator something unique
   (e.g. `ax_lx_alloc`) and add a `-1`-sentinel → twin mapping, OR bundle axalloc.ax's ax_alloc
   for Linux (watch for duplicate low-level symbols with std/mem/alloc.ax).
3. **mmap WORKS** standalone (verified 77) — so ax_os_alloc is fine; the -9 comes from a specific
   mmap call (likely the slab, before the fix, or a wrong size). NOTE: std/mem/alloc.ax is
   COMPILER-BUNDLED (concatenate_stdlib), so any change there → A!=B → **B==C gate required**.
The ungate experiment + this heap attempt were reverted (tree clean); -4/-5 gate stays win64-only.

### Heap dig #3 (2026-07-17, refined — CORRECTS #2) — the real allocator + where it dies
- The bare `ax_alloc` (341B) in a Linux Vec build is **bootstrap/runtime/axalloc.ax:501**, pulled
  in NOT by concatenate_stdlib (that lists only result/mem.alloc/scheduler/runtime/os/string/io/
  collections) but by the **lazy resolver via the import graph**: `import std.collections` →
  std.mem.alloc → `bootstrap.runtime.axalloc`. So there are TWO allocator impls compiled in
  (std/mem/alloc.ax AND axalloc.ax, overlapping symbols → MODDUP). Adding a 3rd ax_alloc (#2)
  was WRONG — don't. axalloc.ax is the one OP_ALLOC(-1→"ax_alloc") binds to on Linux.
- **axalloc.ax IS already Linux-aware**: `axalloc_get_slab` (L116-123) and `ax_os_alloc`
  (L263-270) both have `else: sys_mmap(...)` branches. So the slab-memset-null theory (#2, which
  was about std/mem/alloc.ax's get_slab) is NOT the live path — axalloc.ax's get_slab is fine.
- **Where it dies:** SIGSEGV **inside libc.so.6** (error 6 = write to unmapped, addr 0) — a
  `memset`/`memcpy` (the `@memset`/`@memcpy` intrinsics lower to libc calls, UND memset/memcpy →
  libc) called with a NULL/bad ptr. So an alloc step returns NULL and Vec then memcpy/memset's it,
  OR ax_segment_manager_init memset's a null slab. `sys_mmap` (extern, from
  bootstrap/runtime/syscall.ax) — VERIFY it's actually bundled+resolved on Linux (no sys_mmap
  symbol showed in the obj; if unresolved it'd patch to 0 → but crash is in libc not ip=0, so
  more likely mmap runs but a downstream ptr is null).
- **NEXT (dedicated session), pick one:** (a) `apt install gdb` in WSL → `gdb --batch -ex run
  -ex bt` for the exact caller of the libc memset/memcpy; or (b) instrument axalloc.ax's ax_alloc/
  ax_os_alloc/ax_segment_manager_init with `sys_write` traces (print WORKS now) to find the step
  returning null; or (c) check the two-allocator MODDUP — confirm the ax_alloc that runs calls the
  MATCHING ax_segment_manager_init/ax_os_alloc (not a cross-wired std/mem/alloc.ax twin). Then
  Vec/HashMap/String light up. std/mem/alloc.ax + axalloc.ax are compiler-bundled → B==C gate.

### Heap dig #4 (2026-07-17, KEY reframe) — crash is at STARTUP, not in @alloc
Diagnostic: a program `println("before alloc"); let p=@alloc(64); ...` prints **NOTHING** —
it SIGSEGVs (libc, addr 0) BEFORE the first println. But `elfhello` (println only, no @alloc)
prints fine. So merely PULLING the allocator into the bundle (via `@alloc` → import-graph loads
bootstrap.runtime.axalloc + std.mem.alloc) breaks the program at startup — NOT ax_alloc's logic.
Strong hypothesis: **multi-file runtime symbol COLLISION**. My freestanding print/panic runtime
(bootstrap/runtime/panic.ax, Linux-only bundle) defines ax_print_*/ax_println_*/ax_panic/
ax_get_global_state_internal/my_strlen; axalloc.ax + std/mem/alloc.ax ALSO define overlapping
runtime symbols (ax_panic_str, get_global_state, ax_size_class_*, ax_segment_*, print helpers).
When both are present the MODDUP mangler renames the dupes (`__mN`), and something the startup/
print path calls gets mis-bound to the wrong copy (or an unresolved twin patched to 0). That
matches "prints fine WITHOUT the allocator, crashes at startup WITH it." **NEXT (dedicated):**
(1) build the @alloc program, dump the obj symtab, grep for `__m` MODDUP-renamed dupes among
ax_print*/ax_panic*/get_global_state/ax_get_global_state_internal, and check which copy the
println/startup relocs bind to; (2) likely fix = DON'T redefine in panic.ax any symbol that
axalloc.ax/std.mem.alloc already provide — instead put the Linux print runtime in a file with
NO name overlap, or gate panic.ax's helpers to avoid dup with axalloc.ax; OR make the Linux
build bundle ONE coherent runtime set. This is the crux; solve the collision first, THEN the
alloc-null (dig #3) if it still reproduces. gdb-in-WSL (apt) or sys_write instrumentation to
confirm. All experiments reverted; main clean at e5be7e2, driver 29757E64.

### Heap dig #5 (2026-07-17, concrete leads for the collision) — START HERE next session
- **`ax_panic` is DUAL-DEFINED**: `std/runtime.ax:16` (always bundled via concatenate_stdlib)
  AND `bootstrap/runtime/panic.ax:39` (my Linux-only bundle) → guaranteed MODDUP dup. It did NOT
  break print-only elfhello (ax_panic unreachable there), but it's a latent hazard the moment the
  allocator (which calls panic on OOM) is reachable. **Likely the print runtime should NOT live in
  panic.ax** — move ax_print_*/ax_println_*/ax_lx_* into a NEW Linux-only file with zero name
  overlap, and drop panic.ax's re-def of ax_panic / ax_get_global_state_internal (let
  std/runtime.ax + the allocator provide those). Reconsider whether panic.ax needs bundling at all.
- **std/mem/alloc.ax is DOUBLE-INCLUDED** when a program imports std.collections: once by
  concatenate_stdlib (always) and again by the lazy resolver via the import graph
  (collections→mem.alloc). Every symbol duplicated → heavy MODDUP. Works on Windows (Vec progs
  fine) so MODDUP mostly handles it, but the ELF twin-fallback + MODDUP interaction may mis-bind a
  startup/print symbol on Linux. Check the driver's dedup: does concatenate_stdlib + lazy-resolver
  guard against loading the same module twice? (bug82/RFC0011 lazy resolver.)
- **Two distinct failure modes seen**: (a) `@alloc` with NO `import` → ax_alloc unresolved (axalloc
  not pulled) → likely patched-to-0 call; (b) `import std.collections` + Vec → ax_alloc defined
  (341B axalloc) → crash in libc memset/memcpy. Handle both. Recommend: build each, dump obj symtab,
  grep `__m` MODDUP suffixes on ax_print*/ax_panic*/ax_alloc/get_global_state, and disasm main's
  first println call target to see if it binds to the right twin. This collision is THE blocker;
  the alloc-null (dig #3) is secondary. Est: 1 focused session with gdb-in-WSL or sys_write traces.

### HARD BLOCKER for autonomous work (2026-07-17): gdb needs interactive sudo
`sudo apt-get install gdb` → "interactive authentication is required" (no passwordless sudo).
So the definitive-backtrace path needs the USER to either install gdb once
(`sudo apt-get install -y gdb` in the WSL Ubuntu) OR approve a dedicated sys_write-instrumentation
session (iterative edits to the compiler-bundled allocator + rebuilds; destabilizing to do blind).
Heap/Vec/String on Linux is PAUSED awaiting that decision — NOT a tick-sized autonomous task.
Everything else on Linux (compute/globals/calls/control-flow/print str-i64-bool-f64) SHIPPED &
verified. Suggested next-session first step once unblocked: move the Linux print runtime OUT of
panic.ax into a collision-free file (kills the ax_panic dup), then gdb-bt the Vec crash.

## ✅ Linux target ≈ WINDOWS PARITY (scoped 2026-07-17) — no "concurrency port" needed
Investigated the actor/scheduler thinking it was the big remaining Linux gap. Findings:
- The scheduler is **COOPERATIVE / single-threaded** — `scheduler.run()` (std/scheduler.ax:332)
  is a plain loop that steps actors on the calling thread; `worker_loop` is defined but never
  called; NO CreateThread/pthread/clone anywhere. So there are no OS threads to port.
- **BUT spawn/await is BROKEN ON WINDOWS TOO.** A minimal `let a = spawn work(10); await a`
  returns **0**, not the result, on the COFF build. So the async/actor feature is an INCOMPLETE
  language feature (aspirational; the axiom_*_suite tests use it in the un-parseable dialect),
  NOT a Linux-specific gap. Do NOT frame it as "Linux concurrency runtime to port" — it doesn't
  work on either platform. Completing async is a separate cross-platform feature effort.
- **WHY async is broken (both platforms):** `OP_AWAIT` has NO selector case in x86_selector
  (only OP_SPAWN at 1737) → falls through → dest reg never set → `await x` returns 0/garbage.
  There is also no actor-execution/result-capture model: spawn creates an actor id (ax_actor_spawn),
  but nothing runs it to completion and captures a "return value" (actors are message-based, not
  function-return). Completing async = RFC-scale (design the await/result semantics + OP_AWAIT
  codegen + drive the cooperative scheduler on await + result plumbing). A defensible SMALLER
  step: REJECT `spawn`/`await` at typecheck as unimplemented (BUG#53 accept-then-miscompile
  convention — silent wrong result is worse than a clean reject) — BUT that option is CLOSED:
  std/scheduler.ax:492 uses `spawn supervisor_handler(...)` and scheduler.ax is bundled into
  EVERY program incl. the compiler self-build (main_air.ax:352 read_file_content), so rejecting
  `spawn` would break self-hosting. So async can ONLY be completed (RFC-scale), not rejected.
  → Genuinely needs user direction / a dedicated RFC + design session; NOT autonomous work.
- Net: the Linux ELF target has **parity with Windows for every feature that actually works**
  (compute/globals/print/heap/strings/collections/parse/for-in/nested/to_str). The only true
  Linux-only follow-ups are rare runtime-symbol twins (ax_time_now_ns) when a program uses them.

## Scope / follow-ups (NOT yet on Linux)
- ✅ **print DONE** (str/i64/bool `83b22af`, **f64 `e5be7e2`**) — full console output works.
- **Heap/Vec/HashMap/String**: STILL blocked on the allocator (see "Heap/Vec/String" section
  above — actor runtime faults in libc; needs a freestanding mmap allocator init). Pure-compute
  + module globals + calls + control-flow + ALL print WORK. This is the main remaining gap.
- **SysV param ABI suspicion**: disasm of a runtime fn showed win64-style param loading (RCX-first)
  — verify emit_param_prologue honors `abi=="sysv"` (RDI-first) before trusting multi-arg calls
  on Linux beyond what elfglob exercises (fib's add(a,b) 2-arg worked → likely OK, but confirm).
- Split PT_LOAD into RX+RW segments (currently one RWX LOAD); DT_INIT/fini; static (non-dyn) ELF.
- Future targets user wants (later): macOS/Mach-O, iOS/iPadOS, Android. The `target_is_windows`
  → generalize to a target enum; is_macos fold already stubbed.

### Heap dig #6 (2026-07-17, final for session) — pin FACTS, then STOP
- println WORKS; crash is strictly inside `@alloc` (Vec prog prints "start"; `let p=@alloc(64)`
  prints "A: before alloc" then SIGSEGVs in libc memset/memcpy on null BEFORE @alloc returns).
- **Bundling bootstrap/runtime/syscall.ax for --target linux is REQUIRED** (axalloc uses sys_mmap/
  munmap/exit; `ax_sys_mmap` then resolves). Re-apply — but it did NOT fix the crash alone.
- **THE WALL — resolution mystery:** the linked `ax_alloc` (341B) did NOT change when I edited
  bootstrap/runtime/axalloc.ax (axtrace never fired), and `scratch/self_linked_concatenated.ax`
  has ZERO `fn ax_alloc`. So the compiled `ax_alloc` comes from a lazy-resolved module NOT in the
  concatenate bundle, and my source edit isn't the copy compiled. **Resolve FIRST next session:**
  WHERE does `ax_alloc` for `import std.collections` actually come from? Trace std.collections →
  std.mem.alloc → ? ; check `library`/.lib/auto-lib cache; instrument the ACTUAL module.
- **DECISION:** heap-on-Linux = large multi-layer runtime-integration (module resolution + coherent
  freestanding runtime bundle + allocator init), no debugger. PAUSED pending user `sudo apt-get
  install -y gdb` in WSL OR a dedicated session. Everything else SHIPPED at e5be7e2. Do NOT
  re-attempt blind on autopilot ticks — it only produces reverts.

Related: [[backlog-open-items]] (RFC 0009 P3), [[bug82-global-var-semantics-open]] (RFC 0017
globals — same subsystem, now extended to ELF), [[ffi-dynamic-linking-priority]].
