# AXIOM Compiler — Task System

## Overview

This directory contains the complete executable task graph for building the AXIOM programming language compiler, runtime, standard library, and toolchain from zero to self-hosting. Every task is designed for autonomous execution by an AI coding agent without human clarification.

## Task Naming Convention

```
pXX-tYY-short-task-name.md
```

| Component | Meaning |
|-----------|---------|
| `pXX` | Phase number (01–18), zero-padded |
| `tYY` | Task number within phase, zero-padded |
| `short-task-name` | Kebab-case descriptor of the deliverable |

**Examples:**
- `p01-t01-repository-bootstrap.md` — Phase 1, Task 1: Initialize monorepo
- `p09-t06-air-builder-expressions.md` — Phase 9, Task 6: AIR builder for expressions

Task numbers are **dependency-ordered within each phase** — lower numbers must complete before higher numbers (unless explicitly marked parallel).

## Phase Organization

| Phase | Name | Focus | Key Milestone |
|-------|------|-------|---------------|
| **p01** | Foundation | Repository, CI, EBNF, frozen struct layouts, diagnostics | M0: CI green, structs frozen |
| **p02** | Lexer | Token definitions, DFA scanner, INDENT/DEDENT, dump-tokens | M1a: Lex all 19 test files |
| **p03** | Parser + AST | FlatAST, recursive descent, Pratt expressions | M1b: Parse all test files |
| **p04** | Semantic Analysis Core | Symbol table, type table, name resolution, type checker, `axc check` | M2: `axc check` works |
| **p05** | Advanced Type System | Generics, monomorphization, sum types, structural typing, `#run` | M2b: Generics working |
| **p06** | Ownership + CTGC | Connection Graph, escape analysis, `=destroy` injection | M3: Ownership enforced |
| **p07** | Runtime MVP | AxAlloc (malloc wrapper), gen-ref, panic handler | M3b: Runtime linked |
| **p08** | C-Backend | AST→C11 translation, `axc build`, defer, sum types/match | **M4: MVC v0.1.0** |
| **p09** | AIR Definition + Builder | AIR instruction set, basic blocks, AST→AIR lowering | M5a: AIR emitted |
| **p10** | Optimization Pipeline | Constant folding, DCE, inlining, CTGC on AIR | M5b: Optimized AIR |
| **p11** | Native x86-64 Backend | Instruction selection, reg alloc, ModRM/SIB, ELF emission | M6: Native ELF binary |
| **p12** | Linker + Multi-Format | PE/COFF, Mach-O, symbol mangling, demangling | M6b: Cross-platform |
| **p13** | ARM64 + RISC-V | Additional architecture backends | M7: Multi-arch |
| **p14** | AxAlloc Production | Size-classed segments, NUMA, GPU pinned memory | M8a: Production allocator |
| **p15** | Actor Runtime | M:N scheduler, actors, channels, async, distributed stub | M8b: Actor concurrency |
| **p16** | Standard Library | Collections, IO, net, crypto, mem, SIMD, AI, quantum/GPU stubs | M9: stdlib complete |
| **p17** | Tooling | Formatter, LSP, package manager, incremental compilation | M10: Full toolchain |
| **p18** | Self-Hosting | Compiler in AXIOM, triple-build, runtime self-hosting | **M11: Self-hosting v1.0** |

## Execution Workflow

### For Each Task

1. **Read the task file completely** before starting any implementation.
2. **Verify dependencies** — all listed dependency tasks must be complete.
3. **Create source files** listed in Outputs.
4. **Implement** following the Implementation Steps in order.
5. **Write tests** defined in the Test Plan.
6. **Run validation** — complete every item in the Validation Checklist.
7. **Verify acceptance** — all Acceptance Criteria must be met.
8. **Mark complete** — check all items in Definition of Done.

### Dependency Rules

- A task **MUST NOT** begin until all tasks listed in its Dependencies section are complete.
- Dependencies are specified by task ID (e.g., `p01-t01`).
- Tasks within the same phase are ordered sequentially unless explicitly noted as parallelizable.
- Cross-phase dependencies are always explicit.

### Parallel Execution Opportunities

After Phase 8 (MVC complete), three tracks can proceed in parallel:
- **Track A (Compiler):** AIR, optimizations, native backend
- **Track B (Runtime):** Production allocator, scheduler, actors
- **Track C (Stdlib):** Collections, IO, networking, crypto

See `task-graph.md` for the full dependency DAG.

## Completion Policy

A task is **complete** when:
1. All source files in Outputs exist and compile
2. All tests in Test Plan pass
3. All items in Validation Checklist are checked
4. All Acceptance Criteria are met
5. All items in Definition of Done are checked
6. No regressions in any previously passing tests (`go test ./...` still green)

A task is **blocked** when:
- Any dependency task is incomplete
- A specification ambiguity prevents implementation (document and escalate)

A task **fails** when:
- An architectural incompatibility with a dependency is discovered
- The specification is contradictory (document in `docs/rfcs/`)

## Task File Format

Every task file contains these sections (all mandatory):

| Section | Purpose |
|---------|---------|
| Purpose | Why this task exists |
| Context | Architecture/spec references |
| Inputs | Required specs, modules, prior tasks |
| Outputs | Concrete deliverables (files, binaries, tests) |
| Dependencies | Explicit task IDs |
| Subsystems Affected | All impacted components |
| Detailed Requirements | Implementation specifics, invariants, constraints |
| Implementation Steps | Step-by-step execution order |
| Test Plan | Unit, integration, golden, fuzz, property tests |
| Validation Checklist | Concrete verification items |
| Acceptance Criteria | Measurable completion criteria |
| Definition of Done | Final checklist |
| Risks & Mitigations | Technical and architectural risks |
| Future Follow-up Tasks | What this task unlocks |

## Quick Reference

- **Total tasks:** 182
- **Critical path:** p01→p02→p03→p04→p06→p08→p09→p10→p11→p18-t05→p18-t06
- **Bootstrap language:** Go 1.22+
- **Module path:** `github.com/axiom-lang/axiom`
- **Test command:** `go test ./...`
- **Lint command:** `golangci-lint run`
- **Build command:** `go build ./cmd/axc`

> **Note on phase numbering:** The task system uses `p01–p18` (1-indexed), while the master plan uses `Phase 0–Phase 11`. The mapping is: plan Phase N maps to task phase `p(N+1)` for Phases 0-3, with later plan phases split across multiple task phases for granularity. See `task-graph.md` for the full mapping.

## Related Documents

- [task-graph.md](task-graph.md) — Dependency DAG with critical path and parallel opportunities
- [milestones.md](milestones.md) — Milestone definitions, acceptance criteria, production gates
- [../plan.md](../plan.md) — Master implementation roadmap (source of truth)
