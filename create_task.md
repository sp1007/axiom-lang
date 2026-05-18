You are the project planning system for the AXIOM programming language project.

Your task is to transform the master implementation roadmap located at:

docs/plan.md

into a COMPLETE EXECUTABLE TASK SYSTEM.

You must generate EXTREMELY DETAILED engineering task files.

The output must be suitable for autonomous execution by Claude Code without human clarification.

# OBJECTIVE

Convert the high-level plan into:

- phase folders
- milestone groups
- dependency-ordered tasks
- executable engineering work units
- verification-driven implementation tasks

Each task must:
- be independently executable
- have measurable completion criteria
- define exact inputs/outputs
- define tests
- define validations
- define risks
- define dependencies

This is NOT a todo list.

This is a PRODUCTION-GRADE COMPILER ENGINEERING TASK GRAPH.

---

# OUTPUT LOCATION

Generate task files under:

docs/tasks/

Naming format:

pXX-tXX-short-task-name.md

Examples:

docs/tasks/p01-t01-repository-bootstrap.md
docs/tasks/p01-t02-build-system-foundation.md
docs/tasks/p02-t01-lexer-token-definition.md
docs/tasks/p02-t02-lexer-implementation.md

Task numbering must:
- follow dependency order
- remain stable
- be deterministic

---

# PHASE ORGANIZATION

Tasks must be grouped by implementation phases.

Example:

Phase 01 — Foundation
Phase 02 — Frontend
Phase 03 — Semantic Analysis
Phase 04 — IR
Phase 05 — Backend
Phase 06 — Runtime
Phase 07 — Optimization
Phase 08 — Tooling
Phase 09 — Self Hosting

You may refine or split phases if needed.

---

# REQUIRED TASK FILE FORMAT

Each task file MUST contain ALL sections below.

# Task Title

## Purpose
Why this task exists.

## Context
Relevant architecture/spec references.

## Inputs
Required:
- specs
- modules
- prior tasks
- APIs
- IR contracts
- runtime assumptions

## Outputs
Concrete deliverables:
- source files
- modules
- binaries
- tests
- docs
- IR validators
- benchmarks

## Dependencies
Explicit task dependencies.

Example:
- p01-t01
- p01-t03

## Subsystems Affected
List all impacted systems.

## Detailed Requirements
VERY DETAILED implementation requirements.

Include:
- invariants
- ownership rules
- architecture constraints
- diagnostics expectations
- determinism requirements
- serialization rules
- error handling rules

## Implementation Steps
Provide step-by-step implementation order.

Each step must:
- be concrete
- be executable
- avoid ambiguity

## Test Plan
Include:
- unit tests
- integration tests
- golden tests
- snapshot tests
- fuzz tests
- regression tests

Define:
- expected inputs
- expected outputs
- failure conditions

## Validation Checklist
Concrete verification checklist.

Examples:
- compiler builds
- tests pass
- IR validates
- deterministic output verified
- diagnostics stable
- no architecture violations

## Acceptance Criteria
Strict measurable completion criteria.

## Definition of Done
Very strict final checklist.

## Risks
Technical and architectural risks.

## Mitigation
How to reduce risks.

## Future Follow-up Tasks
Tasks that become possible afterward.

---

# TASK GRANULARITY RULES

Tasks must be:
- small enough for autonomous execution
- large enough to produce meaningful progress

Avoid:
- giant mega-tasks
- vague tasks
- research-only tasks
- “implement compiler” style tasks

GOOD:
- implement token stream abstraction
- implement AST node arena
- implement diagnostic span system

BAD:
- implement parser
- implement optimizer

---

# ARCHITECTURE RULES

Tasks must preserve:
- compiler layering
- deterministic builds
- modularity
- IR invariants
- future self-hosting compatibility

Never create tasks that:
- bypass validation
- mix frontend/backend
- tightly couple runtime/compiler
- implement optimization before verification

---

# TEST-FIRST POLICY

Every implementation task MUST:
- require tests
- require validation
- define invariants

Compiler tasks MUST define:
- golden tests
- diagnostics validation
- deterministic output validation

Runtime tasks MUST define:
- stress tests
- memory safety checks
- failure-path tests

---

# DEPENDENCY GRAPH REQUIREMENTS

You must derive:
- exact task dependencies
- critical path
- parallelizable tasks

Dependencies must:
- avoid cycles
- reflect real compiler architecture
- preserve implementation order

---

# REQUIRED OUTPUT FILES

Generate:

1. Individual task files
2. docs/tasks/README.md
3. docs/tasks/task-graph.md
4. docs/tasks/milestones.md

README.md must explain:
- task naming
- execution workflow
- dependency rules
- completion policy

task-graph.md must include:
- dependency DAG
- critical path
- parallel execution opportunities

milestones.md must include:
- milestone definitions
- milestone acceptance criteria
- production-readiness gates

---

# IMPORTANT

Tasks must be IMPLEMENTATION-ORIENTED.

Do NOT summarize the plan.

Instead:
- decompose it
- operationalize it
- convert it into executable engineering work

---

# AUTONOMOUS ENGINEERING ASSUMPTION

Assume another Claude instance will later execute tasks automatically.

Therefore each task must:
- contain enough detail
- define exact expectations
- minimize ambiguity
- define success precisely

---

# FINAL REQUIREMENTS

At the end ensure:
- every subsystem from docs/plan.md is covered
- every phase has clear milestones
- every task has deterministic naming
- every task has dependencies
- every task has measurable completion criteria
- every task is independently executable

This task system must be capable of driving the entire AXIOM compiler project from zero to self-hosting.