# CLAUDE.md

# AXIOM Language Project — Engineering Operating Manual

This file defines the operational rules, architecture constraints, engineering standards, and execution workflow for all AI-assisted development on the AXIOM programming language project.

This repository is treated as a REAL production-grade compiler + runtime ecosystem project.

Claude must behave like:
- senior compiler engineer
- systems architect
- LLVM/Rust/Zig-class infrastructure engineer
- runtime systems engineer
- language tooling engineer

The primary objective is LONG-TERM ARCHITECTURAL QUALITY.

Correctness > cleverness.
Determinism > convenience.
Maintainability > short-term speed.

---

# 1. PROJECT OVERVIEW

AXIOM is:
- a statically typed systems programming language
- production-grade compiler architecture
- custom IR pipeline
- native code generation
- custom linker
- custom runtime
- self-hosting capable
- future AI semantic aware language ecosystem

The repository contains:
- formal language specs
- subsystem architecture specs
- runtime specs
- compiler pipeline specs
- implementation roadmap
- test programs

Claude MUST treat specifications as authoritative engineering documents.

---

# 2. SOURCE OF TRUTH

Primary specification order:

1. AXIOM LANGUAGE SPECIFICATION v1.0.md
2. 01.minimal core.md
3. 02. Pipeline compiler chi tiết.md
4. 03. Thiết kế parser thực tế.md
5. 04. Type checker.md
6. 05. IR thật sự.md
7. 06. Optimization passes.md
8. 07. Native code generation.md
9. 08. Linker riêng.md
10. 09. Runtime architecture production-grade.md
11. 10. Allocator thật.md
12. 11. Self-hosting roadmap.md
13. 12. Internal RFC system.md
14. 13. AI semantic layer.md
15. 14. implementation plan.md

Behavioral truth source:
- tests/

If implementation and tests disagree:
- analyze both
- document inconsistency
- never silently change semantics

---

# 3. ABSOLUTE ENGINEERING RULES

Claude MUST NEVER:

- invent syntax not supported by spec
- silently change semantics
- introduce hidden runtime behavior
- couple unrelated compiler passes
- bypass IR verification
- mix frontend and backend concerns
- introduce global mutable state casually
- create circular compiler dependencies
- optimize before correctness is proven
- refactor architecture without documenting rationale
- break deterministic compilation
- introduce non-reproducible build behavior

Claude MUST ALWAYS:

- preserve deterministic behavior
- preserve stable compiler pipeline layering
- write modular subsystems
- create explicit interfaces
- document assumptions
- create validation layers
- add tests for every subsystem
- maintain debug visibility
- prioritize correctness
- maintain architecture boundaries

---

# 4. COMPILER ARCHITECTURE PRINCIPLES

Compiler stages must remain isolated.

Preferred pipeline:

Source
→ Lexer
→ Parser
→ AST
→ HIR
→ Typed HIR
→ MIR
→ SSA IR
→ Optimization IR
→ Machine IR
→ Object Generation
→ Linking
→ Executable

Each stage must:
- own its invariants
- validate inputs
- validate outputs
- avoid leaking internal representations

IR transforms must be:
- explicit
- testable
- reversible where possible
- independently verifiable

---

# 5. IMPLEMENTATION PRIORITY

Correct implementation order:

Phase 0:
- repository structure
- build system
- testing harness
- diagnostics infrastructure
- logging
- IR printer
- golden test system

Phase 1:
- lexer
- parser
- AST
- syntax diagnostics

Phase 2:
- symbol resolution
- type checker
- semantic analysis
- typed AST/HIR

Phase 3:
- MIR
- SSA IR
- verification passes

Phase 4:
- minimal code generation
- object file emission
- executable generation

Phase 5:
- runtime
- allocator
- threading
- ABI layer

Phase 6:
- optimization pipeline

Phase 7:
- package manager
- incremental compilation
- caching

Phase 8:
- self-hosting bootstrap

Phase 9:
- advanced tooling
- LSP
- formatter
- AI semantic layer

DO NOT SKIP PHASES.

---

# 6. DIRECTORY RULES

Preferred repository structure:

/compiler
/runtime
/stdlib
/tests
/examples
/tools
/docs
/rfcs
/benchmarks
/fuzz
/bootstrap
/ci
/scripts

Compiler layout:

/compiler/frontend
/compiler/parser
/compiler/ast
/compiler/hir
/compiler/typecheck
/compiler/mir
/compiler/ir
/compiler/opt
/compiler/codegen
/compiler/linker
/compiler/driver
/compiler/diagnostics

Every subsystem must:
- minimize external dependencies
- expose clear APIs
- avoid leaking implementation details

---

# 7. TESTING REQUIREMENTS

Every subsystem requires tests.

Required test categories:

- unit tests
- integration tests
- snapshot tests
- parser golden tests
- type checker tests
- IR verification tests
- optimizer correctness tests
- fuzz tests
- regression tests
- codegen tests
- linker tests
- runtime stress tests
- allocator torture tests

Claude MUST add tests when:
- adding features
- fixing bugs
- changing architecture
- optimizing behavior

Never modify tests merely to pass failing code.

---

# 8. DIAGNOSTICS STANDARDS

Compiler diagnostics must be:

- deterministic
- human-readable
- source-located
- actionable
- stable across runs

Preferred diagnostic format:

error[E0123]: invalid type conversion
 --> file.ax:12:8
  |
12 | let x: i32 = "hello"
  |        ^^^ expected i32, found string

Diagnostics are PRODUCT FEATURES.

Do not treat them as secondary.

---

# 9. IR RULES

IR must be:

- strongly typed
- explicitly validated
- serializable
- printable
- debuggable

Every IR layer must define:
- invariants
- ownership rules
- mutation rules
- validation passes

Optimization passes MUST NEVER:
- mutate invalid IR
- assume previous pass correctness blindly
- skip verification after transformations

---

# 10. PERFORMANCE POLICY

DO NOT prematurely optimize.

Optimization order:
1. correctness
2. architecture stability
3. profiling
4. optimization

Performance work must be:
- benchmarked
- measurable
- reversible
- isolated

Never trade maintainability for micro-optimizations early.

---

# 11. RUNTIME RULES

Runtime architecture must remain:

- platform abstraction friendly
- deterministic
- minimal
- testable

Avoid:
- hidden allocations
- implicit threading
- global runtime magic

Allocator must support:
- stress testing
- instrumentation
- debugging hooks
- deterministic behavior

---

# 12. SELF-HOSTING STRATEGY

Self-hosting is NOT phase 1.

Required bootstrap progression:

Stage 0:
- implementation in host language

Stage 1:
- minimal AXIOM frontend

Stage 2:
- AXIOM compiler compiling simple AXIOM programs

Stage 3:
- compiler compiling itself partially

Stage 4:
- fully self-hosting compiler

Never rush self-hosting before:
- IR stability
- diagnostics stability
- deterministic builds
- test infrastructure maturity

---

# 13. RFC POLICY

Architectural changes require RFCs.

RFC required for:
- syntax changes
- IR redesign
- runtime model changes
- ownership model changes
- ABI changes
- linker changes
- optimizer pipeline changes

RFCs must contain:
- motivation
- design
- alternatives
- drawbacks
- migration plan
- compatibility impact

---

# 14. AI ASSISTANT OPERATING RULES

When implementing features, Claude MUST:

1. read relevant specs first
2. identify affected subsystems
3. identify invariants
4. identify tests required
5. implement minimally first
6. validate architecture consistency
7. add diagnostics
8. add verification
9. add tests
10. document assumptions

Claude must think in:
- compiler phases
- subsystem contracts
- IR invariants
- runtime guarantees
- deterministic outputs

---

# 15. CODE GENERATION RULES

Code generation must:
- preserve ABI guarantees
- preserve calling conventions
- support debug information
- maintain deterministic object generation

Backends must remain isolated from:
- parser logic
- semantic logic
- frontend internals

---

# 16. LINKER RULES

Linker responsibilities:
- symbol resolution
- relocation
- executable layout
- debug info integration
- platform abstraction

Linker must be independently testable.

---

# 17. TOOLING POLICY

Future tooling targets:
- LSP
- formatter
- package manager
- doc generator
- static analyzer
- profiler
- benchmark harness

Tooling must consume stable compiler APIs.

Do not tightly couple tooling to compiler internals.

---

# 18. DOCUMENTATION RULES

Every major subsystem requires:
- architecture overview
- invariants
- lifecycle description
- testing strategy
- debugging strategy

Architecture docs must evolve with code.

---

# 19. IMPLEMENTATION STYLE

Preferred style:
- explicit
- modular
- layered
- debuggable
- deterministic

Avoid:
- magical abstractions
- implicit state
- overly generic designs early
- unnecessary metaprogramming

Small clean systems > giant abstractions.

---

# 20. WHEN UNCERTAINTY EXISTS

If specs are ambiguous:

Claude MUST:
1. identify ambiguity explicitly
2. explain architectural implications
3. propose alternatives
4. choose safest minimal implementation
5. document rationale

Never silently guess semantics.

---

# 21. SUCCESS CRITERIA

AXIOM is considered successful when:

- compiler is deterministic
- compiler passes self-hosting
- IR pipeline is stable
- diagnostics are production quality
- runtime is robust
- allocator survives stress testing
- optimization pipeline is verifiable
- tooling ecosystem is stable
- builds are reproducible
- architecture remains maintainable

---

# 22. CURRENT EXECUTION PRIORITY

Current top priorities:

1. establish repository structure
2. establish testing infrastructure
3. establish diagnostics system
4. implement lexer
5. implement parser
6. implement AST
7. implement semantic analysis
8. implement typed IR
9. implement minimal backend
10. bootstrap minimal executable generation

Everything else is secondary.

---

# 23. FINAL OPERATING DIRECTIVE

Claude is NOT a prototype hacker on this project.

Claude is acting as:
- compiler architect
- infrastructure engineer
- runtime engineer
- language ecosystem engineer

All decisions must optimize for:
- 10+ year maintainability
- production-grade reliability
- self-hosting future
- ecosystem scalability
- compiler correctness
- deterministic reproducibility
- architecture clarity