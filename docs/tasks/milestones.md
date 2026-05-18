# AXIOM Compiler — Milestones

## Milestone Overview

| ID | Name | Phase | Target | Gate |
|----|------|-------|--------|------|
| M0 | Foundation Complete | p01 | Month 1 | CI green, structs frozen, EBNF reviewed |
| M1 | Frontend Complete | p02–p03 | Month 3 | Parse all 19 test files, 0 ErrorNodes |
| M2 | Type Checker Complete | p04–p05 | Month 6 | `axc check` exits 0 on groups 1–5 |
| M3 | Ownership Complete | p06 | Month 8 | Use-after-move detected, Connection Graph builds |
| M3b | Runtime MVP | p07 | Month 8 | Gen-ref runtime linked, panic handler works |
| **M4** | **MVC v0.1.0** | **p08** | **Month 10** | **`axc build` works, 100 compliance tests pass** |
| M5 | AIR Pipeline | p09–p10 | Month 13 | AIR emitted + optimized, C-backend v2 from AIR |
| M6 | Native x86-64 | p11 | Month 16 | ELF binary without GCC, Fib(40) ≤ 5% slower than clang -O2 |
| M6b | Multi-Format | p12 | Month 17 | PE + Mach-O linking works |
| M7 | Multi-Architecture | p13 | Month 19 | ARM64 + RISC-V backends pass compliance tests |
| M8a | Production Allocator | p14 | Month 21 | AxAlloc: ≥500M alloc/free/sec, NUMA-aware |
| M8b | Actor Runtime | p15 | Month 22 | M:N scheduler, 1M actors, ≥10M msgs/sec |
| M9 | Standard Library | p16 | Month 24 | All std modules implemented and tested |
| M10 | Tooling Complete | p17 | Month 26 | LSP + formatter + package manager working |
| **M11** | **Self-Hosting v1.0** | **p18** | **Month 30** | **Compiler in AXIOM, triple-build verified** |

---

## M0 — Foundation Complete

**Phase:** p01 (5 tasks)

### Deliverables
- Monorepo with full directory structure
- GRAMMAR.ebnf — complete formal grammar
- Frozen struct layouts: Token (8 bytes), AstNode (24 bytes), AirInst (16 bytes)
- CI pipeline on Linux/macOS/Windows
- Coding standards (CONTRIBUTING.md)
- Diagnostic formatter (source snippets, colored output, ICE format)

### Acceptance Criteria
- [ ] `go build ./...` passes on all three platforms
- [ ] `go test ./...` — zero failures
- [ ] `golangci-lint run` — zero warnings
- [ ] `unsafe.Sizeof(Token{}) == 8` (test enforced)
- [ ] `unsafe.Sizeof(AstNode{}) == 24` (test enforced)
- [ ] `unsafe.Sizeof(AirInst{}) == 16` (test enforced)
- [ ] GRAMMAR.ebnf covers all constructs in `axiom_compliance_suite.ax`

### Production Gate
- Grammar review by second engineer
- Struct layouts signed off as FROZEN (no changes after this point without RFC)

---

## M1 — Frontend Complete

**Phase:** p02 (7 tasks) + p03 (10 tasks)

### Deliverables
- Complete lexer with INDENT/DEDENT, UTF-8, error recovery
- Complete parser: recursive descent + Pratt expressions
- FlatAST printer (JSON output)
- `axc dump-tokens <file.ax>` command
- `axc dump-ast <file.ax>` command
- 50+ golden tests
- Fuzz targets for lexer and parser

### Acceptance Criteria
- [ ] Parse `axiom_compliance_suite.ax` with 0 ErrorNodes
- [ ] Parse all 19 test suite files with 0 panics
- [ ] Lexer fuzz: 1M iterations, 0 panics
- [ ] Parser fuzz: 1M iterations, 0 panics
- [ ] All golden tests pass
- [ ] `axc dump-ast` produces valid JSON
- [ ] INDENT/DEDENT tokens always balanced

### Production Gate
- 100% of grammar rules have corresponding golden tests
- Error recovery tested: malformed input produces ErrorNodes (not panics)

---

## M2 — Type Checker Complete

**Phase:** p04 (11 tasks) + p05 (7 tasks)

### Deliverables
- Symbol table + scope stack (O(1) lookup)
- Type table with all primitives + structs + functions
- Name resolver with Lazy Field Analysis
- Type inference (local Hindley-Milner)
- Overload resolution scoring
- Monomorphization engine
- Effects system propagation
- Basic compile-time constant evaluation (`#run` stub)
- `axc check <file.ax>` command

### Acceptance Criteria
- [ ] `axc check axiom_compliance_suite.ax` exits 0 for groups 1–5
- [ ] Monomorphization produces distinct TypeIDs for `Box[i32]` vs `Box[f64]`
- [ ] Effects propagation through call graph verified
- [ ] Error messages include file:line:col
- [ ] 0 regressions on Phase 1 golden tests

### Production Gate
- 100% unit test coverage of type inference rules
- Structural duck-typing tested with interfaces
- Cycle detection in Lazy Field Analysis verified

---

## M3 — Ownership Complete

**Phase:** p06 (8 tasks)

### Deliverables
- Connection Graph (nodes, edges, queries)
- Ownership checker (mut/lent/sink/Isolated rules)
- Escape analysis
- CTGC: `=destroy` injection + alias reuse
- Arena block handling

### Acceptance Criteria
- [ ] Use-after-move detected at compile time for all test cases
- [ ] `Isolated[T]` constraint verified via Connection Graph
- [ ] Escape analysis converts non-escaping heap allocs to stack
- [ ] `=destroy` inserted at correct scope boundaries
- [ ] Arena block: `lent` reference escaping → compile error
- [ ] CTGC reduces heap allocation count by ≥20% on Fibonacci benchmark
- [ ] 0 regressions on Phase 1–2 tests

### Production Gate
- Path-sensitive analysis false positive rate < 10%
- Connection Graph builds for all compliance suite test cases

---

## M3b — Runtime MVP

**Phase:** p07 (5 tasks)

### Deliverables
- `runtime/axalloc/axalloc.c` — malloc + 8-byte header
- `runtime/axalloc/genref.c` — generational reference check
- `runtime/panic/panic.c` — panic handler with stack trace
- `runtime/axruntime.h` — unified C header

### Acceptance Criteria
- [ ] Gen-id mismatch triggers clean panic (not segfault)
- [ ] `ax_alloc` / `ax_free` manage 8-byte AxHeader correctly
- [ ] Panic handler prints message + stack trace to stderr
- [ ] Runtime compiles with GCC and Clang

### Production Gate
- AddressSanitizer clean on all runtime tests
- Gen-ref ABI (8-byte header) signed off as FROZEN

---

## M4 — MVC v0.1.0 (Minimum Viable Compiler)

**Phase:** p08 (14 tasks)

### Deliverables
- Complete C-backend (TypedAST → C11)
- `axc build <file.ax>` — end-to-end compilation
- `axc build --emit-c` — intermediate C output
- C codegen for `defer` statements (LIFO ordering)
- C codegen for sum types (tagged unions) and `match` expressions
- All 100 compliance suite tests compiling and running
- Compile time benchmarks baseline

### Acceptance Criteria
- [ ] `axc build hello.ax` → executable prints "Hello, World!" and exits 0
- [ ] All 100 compliance suite tests pass when compiled and run
- [ ] `axc build --emit-c` writes readable C11 output
- [ ] Compile time for 1000-line file: < 500ms
- [ ] Gen-id mismatch at runtime: clean panic (not segfault)
- [ ] Programs intentionally triggering use-after-free: rejected at compile time

### Production Gate
- **This is the first externally-demonstrable milestone**
- Benchmark baseline committed: Hello World compile time, Fib(40) runtime
- C output snapshot tests committed for all grammar constructs
- Tag `v0.0.1` in repository

---

## M5 — AIR Pipeline

**Phase:** p09 (12 tasks) + p10 (11 tasks)

### Deliverables
- AIR instruction set (50+ opcodes)
- AIR basic blocks + CFG
- AIR verifier (SSA invariants)
- AIR text printer
- TypedAST → AIR builder
- Optimization passes: const fold, DCE, inlining, CTGC on AIR
- Updated C-backend reading from AIR
- `axc dump-air` command

### Acceptance Criteria
- [ ] All Phase 4 tests pass through AIR pipeline
- [ ] AIR verifier catches: duplicate vReg definitions, missing terminators
- [ ] CTGC on AIR reduces `alloc.heap` by ≥20%
- [ ] DCE eliminates `if (false)` branches
- [ ] Constant folding eliminates `2+3` at compile time
- [ ] C-backend v2 (from AIR) produces identical behavior to v1

### Production Gate
- AIR instruction set signed off as FROZEN
- Differential tests: C-backend v1 vs v2 produce identical execution results

---

## M6 — Native x86-64

**Phase:** p11 (16 tasks)

### Deliverables
- x86-64 instruction selector
- Linear scan register allocator
- Stack frame generation (System V + Win64 ABI)
- Machine code emitter
- ELF64 format writer
- In-memory linker with backpatching
- DWARF line info
- `.axmeta` section writer
- `axc build --target x86-linux` command

### Acceptance Criteria
- [ ] All 100 compliance tests pass via native backend (ELF on Linux)
- [ ] Fibonacci(40) native: ≤ 5% slower than `clang -O2`
- [ ] GDB can step through source lines via DWARF info
- [ ] `.axmeta` section present and decompressible
- [ ] No GCC/Clang invocation in native build path

### Production Gate
- Register allocator: no spill bugs on compliance suite
- ELF binaries accepted by `readelf` and `objdump`
- Deterministic: same source → identical binary bytes

---

## M7 — Multi-Architecture

**Phase:** p13 (7 tasks)

### Deliverables
- ARM64 instruction selector + register allocator + AAPCS64 ABI
- RISC-V 64 instruction selector + psABI
- Mach-O integration (ARM64 macOS)
- Cross-compilation from any host to any target

### Acceptance Criteria
- [ ] ARM64 backend passes all compliance tests on Linux ARM64
- [ ] RISC-V backend passes all compliance tests on RISC-V (emulated)
- [ ] `axc build --target aarch64-macos` produces signed Mach-O
- [ ] Cross-compilation: x86 host → ARM64 target works

### Production Gate
- NEON SIMD instructions verified on ARM64
- Multi-target differential tests: same source → same behavior on all architectures

---

## M8a — Production Allocator

**Phase:** p14 (8 tasks)

### Deliverables
- Size-classed segments (30 size classes, 64KB segments)
- Bump-pointer fast path (~3 CPU cycles)
- Per-actor heap (lock-free allocation)
- Free-list sharding
- NUMA-aware segment allocation
- GPU pinned memory
- Actor crash cleanup (O(1))

### Acceptance Criteria
- [ ] Throughput: ≥ 500M alloc/free pairs/sec
- [ ] Actor crash cleanup: all segments returned in O(1)
- [ ] NUMA: < 5% latency increase for cross-NUMA access
- [ ] Zero lock contention for per-actor allocations

### Production Gate
- AddressSanitizer + ThreadSanitizer clean
- Torture test: 10M allocations of random sizes, zero memory leaks

---

## M8b — Actor Runtime

**Phase:** p15 (10 tasks)

### Deliverables
- Actor struct with isolated heap + mailbox
- M:N work-stealing scheduler
- Bounded + unbounded channels
- Async state-machine executor
- Supervisor tree (Erlang OTP-style)
- I/O event loop (epoll/kqueue/IOCP)

### Acceptance Criteria
- [ ] Spawn 1M actors without OOM
- [ ] Actor spawn latency: < 1μs
- [ ] Message throughput: ≥ 10M msgs/sec single-core
- [ ] Work-stealing: balanced load across N OS threads
- [ ] Supervisor restart: crashed actor replaced within 100μs

### Production Gate
- Concurrency stress tests pass under ThreadSanitizer
- No deadlocks in 24-hour soak test

---

## M9 — Standard Library

**Phase:** p16 (19 tasks)

### Deliverables
- `std/collections/` — Array, HashMap, Set, Deque
- `std/string.ax`, `std/math.ax`, `std/io.ax`
- `std/fs.ax`, `std/net/`, `std/crypto.ax`
- `std/concurrency.ax`, `std/testing.ax`, `std/mem.ax`
- `std/time.ax`, `std/json.ax`, `std/fmt.ax`
- `std/os.ax`, `std/iter.ax`, `std/random.ax`
- `std/cli.ax`, `std/ffi.ax`

### Acceptance Criteria
- [ ] All stdlib modules compile and test with `axc test ./std/...`
- [ ] HashMap: lookup < 100ns for 1M string keys
- [ ] HTTP client/server: GET request round-trip works
- [ ] Crypto: SHA-256 produces correct hashes (verified against test vectors)
- [ ] All ownership rules enforced within stdlib code

### Production Gate
- Stdlib is the primary dogfooding test for the compiler
- Every stdlib module has ≥ 80% test coverage

---

## M10 — Tooling Complete

**Phase:** p17 (10 tasks)

### Deliverables
- `axc fmt` — zero-config formatter (idempotent)
- `axc check` — type checker with rich diagnostics
- `axc lsp` — LSP 3.17 server
- `axc get` — package manager (PubGrub solver)
- `axc doc` — documentation generator
- Incremental compilation (hash-based caching)
- REPL for interactive exploration

### Acceptance Criteria
- [ ] `axc fmt`: idempotent — `fmt(fmt(src)) == fmt(src)`
- [ ] LSP: diagnostics latency < 100ms for single-file changes
- [ ] Package manager: resolve + install dependency from git
- [ ] Incremental: recompile only changed files (verified by cache hit rate)

### Production Gate
- VS Code extension published (axiom-lang.vscode)
- Package manager handles diamond dependency resolution

---

## M11 — Self-Hosting v1.0.0

**Phase:** p18 (4+ tasks)

### Deliverables
- Complete compiler rewritten in AXIOM
- `axc_self` binary: AXIOM compiler compiled by Go `axc`
- Triple-build verification: compile 3×, identical binary hashes
- Runtime (AxAlloc, scheduler) written in AXIOM

### Acceptance Criteria
- [ ] `axc_self` passes all 19 compliance suites
- [ ] Triple-build: `hash(build1) == hash(build2) == hash(build3)`
- [ ] Compile speed: ≥ 100k LOC/sec (Stage 2), ≥ 500k LOC/sec (Stage 3)
- [ ] `axc_self` not > 20% slower than Go `axc` on benchmarks
- [ ] Zero dependency on libc for core runtime (Stage 3)
- [ ] Runtime (AxAlloc, panic handler) written in AXIOM

### Production Gate
- **This is the production release milestone**
- Reproducible builds verified from clean state
- All 19 compliance suites pass with self-hosted compiler
- Runtime is fully in AXIOM (no C runtime files)
- Triple-build verification passes (p18-t05)
- Tag `v1.0.0` in repository

---

## Transition Gate Protocol

For every milestone transition:

1. Run all 19 compliance suites with the current binary
2. Run `go test ./...` — zero failures
3. Run benchmark suite — no > 5% regression vs previous milestone
4. Update `docs/changelog/` with decisions made
5. Tag the milestone in git
