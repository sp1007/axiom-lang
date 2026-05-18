You are the lead architect and principal compiler engineer for a new programming language project named AXIOM.

Your task is NOT to implement code yet.

Your task is to deeply analyze the existing AXIOM specifications and create a COMPLETE PRODUCTION-GRADE IMPLEMENTATION PLAN for the entire language ecosystem.

# CONTEXT

Project structure:

AXIOM SPECIFICATION/
├── 01.minimal core.md
├── 02. Pipeline compiler chi tiết.md
├── 03. Thiết kế parser thực tế.md
├── 04. Type checker.md
├── 05. IR thật sự.md
├── 06. Optimization passes.md
├── 07. Native code generation.md
├── 08. Linker riêng.md
├── 09. Runtime architecture production-grade.md
├── 10. Allocator thật.md
├── 11. Self-hosting roadmap.md
├── 12. Internal RFC system.md
├── 13. AI semantic layer.md
├── 14. implementation plan.md
└── AXIOM LANGUAGE SPECIFICATION v1.0.md

There is also:
tests/

The tests directory contains language examples and should be treated as executable behavioral specifications.

# YOUR OBJECTIVES

You must produce a COMPLETE ENGINEERING EXECUTION PLAN for building AXIOM from scratch.

The final output must be practical enough that a team of compiler engineers can immediately start implementation phase-by-phase.

You must:
- infer architecture
- resolve contradictions between specs
- identify missing pieces
- propose implementation order
- define milestone gates
- define test strategy
- define CI/CD strategy
- define bootstrap strategy
- define self-hosting strategy
- define risk mitigation

You are expected to think like:
- LLVM architects
- Rust compiler team
- Zig compiler engineers
- Jai/Odin language implementers
- production systems engineers

# IMPORTANT

DO NOT summarize the specs.

Instead:
- reverse engineer the intended architecture
- derive implementation dependencies
- derive compiler layering
- derive subsystem contracts
- derive internal APIs
- derive executable milestones

Treat this as a REAL compiler project.

# REQUIRED OUTPUT FORMAT

Produce the result in the following sections.

-------------------------------------------------------------------------------
1. HIGH LEVEL ARCHITECTURE
-------------------------------------------------------------------------------

Describe:

- compiler stages
- runtime architecture
- memory model
- IR layers
- optimization pipeline
- code generation strategy
- linker architecture
- package/module system
- build system
- tooling ecosystem
- self-hosting evolution
- AI semantic layer integration

Also provide:
- dependency graph between all subsystems
- critical path analysis
- subsystem coupling analysis

-------------------------------------------------------------------------------
2. IMPLEMENTATION PHASE ROADMAP
-------------------------------------------------------------------------------

Create VERY DETAILED phases.

For EACH phase include:

## Phase Name

### Goals
What must be achieved.

### Inputs
Which specs/files/modules are required.

### Outputs
Concrete deliverables:
- binaries
- libraries
- compiler stages
- test artifacts
- documentation
- benchmark tools
etc.

### Components Implemented
Detailed subsystem list.

### Internal APIs Introduced
Describe interfaces/contracts.

### Test Strategy
Unit tests
Integration tests
Golden tests
Snapshot tests
Compiler tests
IR verification
Fuzzing
Property testing
Performance regression tests

### Acceptance Criteria
Objective measurable success criteria.

### Definition of Done
Strict completion checklist.

### Risks
Technical risks and architectural risks.

### Mitigation
How to reduce risks.

### Estimated Complexity
Low / Medium / High / Extreme

### Dependencies
Which previous phases are required.

-------------------------------------------------------------------------------
3. DETAILED COMPILER PIPELINE PLAN
-------------------------------------------------------------------------------

Describe EXACTLY:

frontend
lexer
parser
AST
HIR
MIR
typed IR
SSA IR
backend IR
machine IR

For EACH layer describe:
- purpose
- ownership model
- mutability rules
- verification rules
- serialization format
- optimization opportunities
- debug strategy
- testing strategy

-------------------------------------------------------------------------------
4. DIRECTORY STRUCTURE PLAN
-------------------------------------------------------------------------------

Propose a REAL production repository structure.

Include:
- compiler/
- runtime/
- stdlib/
- tests/
- tools/
- benchmarks/
- docs/
- ci/
- examples/
- bootstrap/
- fuzz/
- integration/
etc.

Explain the purpose of every major directory.

-------------------------------------------------------------------------------
5. TEST INFRASTRUCTURE PLAN
-------------------------------------------------------------------------------

Design a COMPLETE compiler testing strategy.

Include:
- parser tests
- type checker tests
- IR validation
- optimizer correctness
- codegen correctness
- linker validation
- runtime validation
- memory safety testing
- allocator stress testing
- ABI compatibility tests
- self-hosting tests
- differential testing
- fuzzing
- snapshot testing
- golden tests

Also define:
- test naming conventions
- expected outputs
- failure diagnostics
- automated test runners

Use the tests/ directory as behavioral specification input.

-------------------------------------------------------------------------------
6. BUILD + TOOLCHAIN PLAN
-------------------------------------------------------------------------------

Design:
- bootstrap compiler strategy
- build system
- package manager
- incremental compilation
- caching
- parallel compilation
- reproducible builds
- cross compilation
- debug info generation
- LSP support
- formatter
- linter
- documentation generator

-------------------------------------------------------------------------------
7. SELF-HOSTING STRATEGY
-------------------------------------------------------------------------------

Provide a MULTI-STAGE bootstrap plan.

Include:
Stage 0
Stage 1
Stage 2
Stage 3

For EACH:
- implementation language
- compiler capabilities
- limitations
- verification method
- transition criteria

Define EXACTLY when AXIOM becomes self-hosting.

-------------------------------------------------------------------------------
8. RUNTIME + MEMORY PLAN
-------------------------------------------------------------------------------

Design:
- allocator architecture
- thread model
- async runtime
- GC or non-GC strategy
- ownership model
- memory arenas
- object layout
- ABI
- panic handling
- exception model
- FFI boundary
- OS abstraction layer

-------------------------------------------------------------------------------
9. OPTIMIZATION ROADMAP
-------------------------------------------------------------------------------

Describe:
- early optimizations
- mid-level optimizations
- backend optimizations
- SSA passes
- escape analysis
- inlining
- dead code elimination
- monomorphization
- devirtualization
- vectorization
- PGO
- LTO

For EACH optimization:
- IR level
- prerequisites
- validation strategy
- profitability heuristic

-------------------------------------------------------------------------------
10. ENGINEERING MANAGEMENT PLAN
-------------------------------------------------------------------------------

Design:
- RFC workflow
- coding standards
- review process
- branch strategy
- CI pipeline
- release strategy
- semantic versioning
- benchmark tracking
- regression prevention
- documentation standards

-------------------------------------------------------------------------------
11. MISSING SPEC ANALYSIS
-------------------------------------------------------------------------------

Identify:
- underspecified systems
- contradictions
- dangerous ambiguities
- impossible requirements
- missing ABI details
- missing syntax details
- missing runtime guarantees

For EACH:
- explain the issue
- explain implementation risk
- propose concrete resolution

-------------------------------------------------------------------------------
12. FINAL EXECUTION GRAPH
-------------------------------------------------------------------------------

Produce:

- implementation order graph
- critical path
- parallelizable tasks
- estimated milestone sequence
- likely blockers
- minimum viable compiler definition
- first self-hosting milestone
- production-ready milestone

Also define:
- what should NOT be implemented early
- what should be stubbed initially
- what should be deferred

# IMPORTANT ENGINEERING RULES

You must optimize for:
- maintainability
- compiler correctness
- debuggability
- incremental evolution
- future self-hosting
- deterministic behavior
- testability
- modularity
- performance later, correctness first

Avoid:
- premature optimization
- overengineering early parser stages
- unstable IR design
- tightly coupled passes
- runtime/compiler circular dependencies

# FINAL REQUIREMENT

At the end, produce:

1. A COMPLETE milestone table
2. A subsystem dependency DAG
3. A “minimum viable AXIOM compiler” definition
4. A “production-grade AXIOM” definition
5. A prioritized next-action checklist for engineers

Be extremely concrete.
Be implementation-oriented.
Be production-grade.
Assume this language will eventually compete with Rust/Zig/LLVM-class ecosystems.