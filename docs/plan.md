# AXIOM COMPILER — COMPLETE PRODUCTION-GRADE IMPLEMENTATION PLAN

*Compiled by: Lead Architect & Principal Compiler Engineer*
*Date: 2026-05-16 | Specification version: v1.0 | Target: Production-grade AXIOM ecosystem*

---

## EXECUTIVE SUMMARY

AXIOM is a systems-level, AOT-compiled, statically typed language with: indentation-based syntax, single-ownership memory safety via Generational References (no GC, no borrow checker), an Actor concurrency model with isolated heaps, a first-class AI semantic layer (`.axmeta`), and a dual-backend strategy (C-transpiler for MVP, native x86-64/ARM64/RISC-V for production). The compiler is named `axc` and is a single monolithic binary embedding compiler, formatter, LSP, and package manager.

Bootstrap language: **Go** (for maximum AI code-generation assistance and speed to first milestone). The spec contradicts itself on this point (sections mention Go, Rust, and Zig variously); Go is chosen because it was the most consistently mentioned, it has the lowest barrier for rapid prototyping, its GC eliminates manual memory management during the bootstrap phase, and AI tooling generates Go fluently.

---

## SECTION 1 — HIGH-LEVEL ARCHITECTURE

### 1.1 Compiler Stages (Full Pipeline)

```
Source (.ax)
    │
    ▼
[Stage 1] Lexer
    │  Input:  UTF-8 source text
    │  Output: FlatTokenArray (zero-copy, 8 bytes/token: kind:u8, offset:u32, len:u16, pad:u8)
    │  Method: DFA, SIMD-accelerated whitespace scanning
    │  Key:    Emits INDENT/DEDENT synthetic tokens via indentation stack
    ▼
[Stage 2] Parser
    │  Input:  FlatTokenArray
    │  Output: FlatAST ([]AstNode, index-based, no pointers)
    │  Method: Recursive Descent (statements) + Pratt (expressions)
    │  Key:    Parallel per-file, error-recovering via panic-mode sync
    ▼
[Stage 3] Name Resolution & Semantic Graph
    │  Input:  FlatAST set (all files in compilation unit)
    │  Output: SemanticGraph (symbol table + import resolution)
    │  Method: Lazy Field Analysis (Zig-style): only resolve symbols reachable from main()
    │  Key:    Open-addressing hash map, O(1) lookup
    ▼
[Stage 4] Type Checker + Connection Graph
    │  Input:  SemanticGraph
    │  Output: TypedAST + ConnectionGraph + EscapeMap
    │  Method: Hindley-Milner local inference, structural duck-typing for interfaces,
    │           monomorphization, effect system propagation, async→state-machine rewrite
    │  Key:    Connection Graph tracks Owns/Borrows/FlowsTo/EscapesTo edges
    ▼
[Stage 5] AIR Builder (AST→IR Lowering)
    │  Input:  TypedAST
    │  Output: AIR (Axiom IR) — flat []AirInst, 16 bytes/inst, SSA form
    │  Method: Post-order AST traversal; loop_region preservation
    │  Key:    Every inst carries meta_idx → AI metadata table
    ▼
[Stage 6] Optimization Pipeline
    │  Input:  Raw AIR
    │  Output: Optimized AIR
    │  Passes (ordered):
    │    T1: Compile-time execution (#run) → constant substitution
    │    T1: Monomorphization pass → erase generics, clone AIR blocks
    │    T1: Connection Graph / CTGC → alloc.heap→alloc.stack (escape), alias reuse
    │    T2: Auto-SoA mutator (opt-in @SOA or --ai-suggest)
    │    T2: Loop vectorization (SIMD) → v_add, v_mul 256/512-bit
    │    T3: Dead code elimination (CFG reachability)
    │    T3: Constant folding + propagation
    │    T3: Inlining (small functions, call-graph analysis)
    │    T3: Register allocation (linear scan, native backend only)
    ▼
[Stage 7A] C-Backend (MVP / stage 0–2)
    │  Input:  Optimized AIR (or TypedAST directly for MVP)
    │  Output: C11/C23 source text
    │  Method: Direct AIR→C statement mapping
    │  Key:    Invokes gcc/clang as subprocess for final binary
    ▼
[Stage 7B] Native Backend (production, stage 3+)
    │  Input:  Optimized AIR
    │  Output: Machine code (x86-64, ARM64, RISC-V)
    │  Method: Instruction selection (tree-pattern matching) →
    │           liveness analysis → linear scan register allocation →
    │           stack frame generation → in-memory relocation → ELF/PE/Mach-O emission
    ▼
[Stage 8] Linker (In-Memory)
    │  Input:  Machine code byte arrays from all modules
    │  Output: Final executable (ELF / PE / Mach-O) + .axmeta section
    │  Method: Global symbol table (open-addressing hash map), PC-relative relocations,
    │           backpatching from PatchList, in-memory only (no .o files written to disk)
    │  Key:    Appends .axmeta (Zstd-compressed JSON semantic graph)
    ▼
Executable Binary
```

### 1.2 Runtime Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    AXIOM RUNTIME (embedded microkernel)          │
├─────────────────────────────────────────────────────────────────┤
│  M:N Adaptive Work-Stealing Scheduler                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ OS Thread│ │ OS Thread│ │ OS Thread│ │ OS Thread│  (N cores) │
│  │ RunQueue │ │ RunQueue │ │ RunQueue │ │ RunQueue │            │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘           │
│       │work-steal──┘             │             │                  │
│       ▼  (M lightweight actors)  ▼             ▼                  │
│  ┌────────────────────────────────────────────────────┐          │
│  │  Actor Pool  [Actor₁][Actor₂]...[ActorN]            │          │
│  │  Each actor: isolated heap, channel mailbox,        │          │
│  │              reduction budget (preemption)          │          │
│  └────────────────────────────────────────────────────┘          │
├─────────────────────────────────────────────────────────────────┤
│  AxAlloc (Per-Actor Heap)                                        │
│  ┌─────────────────────────────────────────────────────┐        │
│  │ Size-classed segments (64KB each, bump pointer)     │        │
│  │ Generational header (8-byte per alloc)              │        │
│  │ Free-list sharding by size class                    │        │
│  │ NUMA-node affinity aware                            │        │
│  │ GPU pinned memory (ax_alloc_pinned)                 │        │
│  └─────────────────────────────────────────────────────┘        │
├─────────────────────────────────────────────────────────────────┤
│  Panic Handler → .axtrace emission → semantic debugger          │
│  Async State Machine executor (zero-heap coroutines)            │
│  Supervisor Tree (Erlang OTP-style fault tolerance)             │
│  Distributed Actor Transport (TCP/QUIC, location transparency)  │
└─────────────────────────────────────────────────────────────────┘
```

### 1.3 Memory Model

| Layer | Mechanism | Who inserts | Cost |
|---|---|---|---|
| Stack allocation | Escape analysis: alloc.heap→alloc.stack | Compiler (opt pass) | Zero at runtime |
| Heap allocation | AxAlloc size-classed bump pointer | Runtime | ~3–4 CPU cycles |
| Generational safety | 8-byte header + gen_id comparison | Runtime + compiler | 1 CMP instruction per deref |
| CTGC/object reuse | Alias instruction in AIR | Compiler (opt pass) | Zero (in-place reuse) |
| Arena regions | `in [Arena]` block, O(1) dealloc | Developer + compiler | Near zero |
| Actor crash cleanup | Free all segments of crashed actor | Runtime scheduler | O(1) per actor |

### 1.4 IR Layers

```
Source Text
    │
    ▼ [Lexer]
FlatTokenArray          — zero-copy 8-byte tokens
    │
    ▼ [Parser]
FlatAST                 — []AstNode, index-based, 24 bytes/node
                          (kind:u8, token_idx:u32, first_child:u32,
                           next_sibling:u32, payload:u32, flags:u16)
    │
    ▼ [Type Checker]
TypedAST                — FlatAST with type_idx resolved on every node
                          + ConnectionGraph (ownership edges)
                          + EscapeMap (per-variable escape status)
    │
    ▼ [IR Builder]
AIR (raw)               — []AirInst (16 bytes each, SSA, infinite virtual regs)
                          + BasicBlock graph
                          + MetadataTable (for .axmeta)
                          + LoopRegions (preserved for SIMD/GPU)
    │
    ▼ [Optimizer]
AIR (optimized)         — same structure, transformed:
                          - alloc.heap may become alloc.stack
                          - alias instructions inserted (CTGC)
                          - dead blocks removed
                          - constants folded
                          - generics monomorphized
    │
    ▼ [Native Backend]
MachineIR               — Target-specific instruction sequence
                          (virtual regs still present)
    │
    ▼ [Register Allocator]
PhysicalMachineCode     — Virtual regs mapped to physical regs
                          + spill slots assigned
    │
    ▼ [Emitter]
[]byte                  — Raw machine code bytes, relocations patched in-memory
```

### 1.5 Subsystem Dependency Graph

```
Grammar/EBNF
    │
    ├──► Lexer ──► Parser ──► FlatAST
    │                              │
    │              ┌───────────────┤
    │              ▼               ▼
    │         StringPool      AstNode defs
    │              │
    │              ▼
    │         NameResolver ──► SymbolTable ──► ScopeStack
    │              │
    │              ▼
    │         TypeChecker ──► TypeTable
    │              │        ──► ConnectionGraph
    │              │        ──► EffectChecker
    │              │        ──► MonomorphizationEngine
    │              ▼
    │         AIRBuilder ──► AIR + MetadataTable
    │              │
    │         ┌────┤
    │         ▼    ▼
    │    OptPasses  CTGCPass
    │    (DCE,CF,   (EscapeAnalysis,
    │    Inline,    ObjectReuse)
    │    SoA)
    │         │
    │    ┌────┴────────────┐
    │    ▼                 ▼
    │  CBackend         NativeBackend
    │  (MVP→)           (Prod→)
    │    │               │
    │    │         InstrSelector
    │    │         LivenessAnalysis
    │    │         LinearScanRegAlloc
    │    │         StackFrameGen
    │    │               │
    │    └──────┬─────────┘
    │           ▼
    │       InMemoryLinker ──► SymbolResolver
    │           │           ──► RelocationEngine
    │           │           ──► ELF/PE/MachO Emitter
    │           ▼
    │       .axmeta Writer (Zstd JSON)
    │
    └──► Runtime (separate binary/lib)
            AxAlloc
            Scheduler (M:N)
            ActorRuntime
            AsyncExecutor
            PanicHandler
            SupervisorTree
```

### 1.6 Critical Path Analysis

```
BLOCKING (nothing can proceed without these):
1. FlatAST node layout (frozen early)
2. Token kind enum (drives parser and error messages)
3. TypeID scheme (drives type table, monomorphization, codegen)
4. AIR instruction set (frozen before any optimization work)
5. Generational reference ABI (8-byte header layout)

HIGH PRIORITY (unlock multiple parallel workstreams):
6. C-Backend (proves AST correctness, enables self-hosting path)
7. Symbol Table + Scope Stack (required by type checker)
8. Connection Graph data structure (required by escape analysis + CTGC)

PARALLEL-CAPABLE (after critical path):
- Standard library (needs C-backend working)
- LSP server (needs type checker working)
- Native backend (independent of C-backend)
- .axmeta emission (needs typed AST)
- AxAlloc (independent of compiler)
- M:N Scheduler (independent of compiler)
```

### 1.7 Subsystem Coupling Analysis

| Pair | Coupling | Contract | Notes |
|---|---|---|---|
| Lexer ↔ Parser | Low | FlatTokenArray slice | Lexer is pure function |
| Parser ↔ TypeChecker | Medium | FlatAST + StringPool | TypeChecker mutates payload fields |
| TypeChecker ↔ ConnectionGraph | High | Shared node indices | Must be designed together |
| ConnectionGraph ↔ CTGC Pass | High | Edge enumeration API | Tight coupling acceptable |
| AIRBuilder ↔ Optimizer | Low | []AirInst slice | Optimizer is pure transform |
| Optimizer ↔ CBackend | Low | Optimized AIR | C-backend only reads AIR |
| Optimizer ↔ NativeBackend | Low | Optimized AIR | Same interface |
| Runtime ↔ Compiler | Low | 8-byte header ABI | Only ABI contract |
| Linker ↔ NativeBackend | Medium | Symbol table + byte[] | Linker owns final layout |

---

## SECTION 2 — IMPLEMENTATION PHASE ROADMAP

---

## Phase 0 — Foundation & Tooling Setup

### Goals
Establish repository structure, build system, CI skeleton, coding standards, and the Go bootstrap environment. No AXIOM code compiles yet.

### Inputs
- AXIOM LANGUAGE SPECIFICATION v1.0.md (grammar section)
- EBNF from spec section 3
- Repository design from spec `14. implementation plan.md`

### Outputs
- `axiom-lang/` monorepo initialized
- `go.mod` with Go 1.22+
- `Makefile` / `task.yaml` for common commands
- CI pipeline (GitHub Actions): `lint`, `test`, `build`
- `GRAMMAR.ebnf` — complete formal grammar file
- `docs/design/token-kinds.md`
- `docs/design/ast-schema.md`
- `docs/design/air-schema.md`
- Skeleton directory tree (all directories created, placeholder `README.md`)
- Coding standards document (`CONTRIBUTING.md`)
- Pre-commit hooks (gofmt, staticcheck)

### Components Implemented
- Repository skeleton
- Build toolchain (Go 1.22, golangci-lint, gotestsum)
- CI/CD (GitHub Actions: build matrix Linux/Windows/macOS)
- EBNF grammar (complete, reviewable by team)
- Token kind enum definition (canonical, frozen)
- AstNode struct definition (frozen layout — this is the single most important early decision)
- AirInst struct definition (frozen layout)

### Internal APIs Introduced

```go
// Token — frozen layout, 8 bytes
type TokenKind uint8
type Token struct {
    Kind   TokenKind  // 1 byte
    Offset uint32     // 4 bytes (byte offset in source)
    Len    uint16     // 2 bytes
    _      uint8      // padding to 8 bytes
}

// AstNode — frozen layout, 24 bytes
type NodeKind uint8
type AstNode struct {
    Kind        NodeKind  // 1 byte
    Flags       uint16    // 2 bytes (is_pub, is_mut, is_async, is_extern...)
    TokenIdx    uint32    // 4 bytes (source location)
    FirstChild  uint32    // 4 bytes (0 = no children)
    NextSibling uint32    // 4 bytes (0 = no sibling)
    Payload     uint32    // 4 bytes (TypeID after typecheck, or literal value index)
    ExtraIdx    uint32    // 4 bytes (points into ExtraData pool for overflow)
    // Total: 23 bytes (pad to 24)
}

// AirInst — frozen layout, 16 bytes
type AirInst struct {
    Opcode  uint16  // 2 bytes (instruction kind)
    TypeID  uint16  // 2 bytes (result type)
    Dest    uint32  // 4 bytes (SSA virtual register)
    Src1    uint32  // 4 bytes
    Src2    uint32  // 4 bytes (or immediate value)
    MetaIdx uint32  // embedded in Src2 high bits or separate pass
}
```

### Test Strategy
- Unit tests: Token struct size assertions (must be 8 bytes)
- Unit tests: AstNode struct size assertions (must be 24 bytes)
- Unit tests: AirInst struct size assertions (must be 16 bytes)
- Grammar validation: hand-parse all 19 test suite files manually against EBNF to find gaps

### Acceptance Criteria
- All struct sizes verified via `unsafe.Sizeof` tests
- EBNF successfully describes all constructs found in `axiom_compliance_suite.ax`
- CI passing on Linux, Windows, macOS
- Zero linter warnings

### Definition of Done
- [ ] `axiom-lang/` repository with full directory tree
- [ ] GRAMMAR.ebnf committed, reviewed, and signed off
- [ ] Token/AstNode/AirInst structs committed and layout-tested
- [ ] CI green on all three platforms
- [ ] CONTRIBUTING.md written

### Risks
- Grammar ambiguities discovered late forcing AST redesign

### Mitigation
- Hand-parse all 19 test files against EBNF in this phase before any code is written
- Grammar review gate before Phase 1 begins

### Estimated Complexity: Low

### Dependencies: None

---

## Phase 1 — Lexer + Parser + FlatAST

### Goals
Given a `.ax` source file, produce a correct FlatAST. Emit AST as JSON for debugging. No type checking yet.

### Inputs
- Phase 0 outputs (frozen AstNode layout, EBNF, Token layout)
- `axiom_compliance_suite.ax` (behavioral spec, parse-level tests)

### Outputs
- `compiler/lexer/` — complete, tested zero-copy lexer
- `compiler/parser/` — complete Recursive Descent + Pratt parser
- `compiler/ast/` — FlatAST definition + printer
- `axc dump-ast <file.ax>` — CLI command that prints AST as JSON
- Test corpus: 50+ `.ax` snippet files covering every grammar rule
- Golden test output: expected JSON AST for each snippet

### Components Implemented

**Lexer:**
- DFA scanner, O(N)
- Indentation stack → INDENT/DEDENT tokens
- UTF-8 validation
- Zero-copy: tokens point into source buffer via offset+len
- All token kinds: keywords, operators, literals, delimiters, INDENT, DEDENT, NEWLINE, EOF
- Error recovery: unknown chars → ErrorToken, continue scanning

**Parser:**
- Recursive Descent for statements: `fn`, `struct`, `interface`, `impl`, `const`, `let`/`mut`, `if`/`elif`/`else`, `for`/`in`, `match`, `return`, `break`, `continue`, `spawn`, `await`, `await_all`, `yield`, `lock`, `unsafe`, `in [Arena]`, `defer`, `#run`, `@[annotation]`, `extern`, `import`, `pub`, `async`
- Pratt Parser for expressions: all binary/unary operators with correct precedence, function calls, index access, field access, closures, generics `[T]`, tuple destructuring
- FlatAST builder: write nodes into `[]AstNode` slice, return root index
- Indentation-based block parsing: block starts with `:`, block is sequence of statements at indent+4
- Error recovery: panic mode with synchronization at DEDENT/fn/struct keywords

**AST Printer:**
- Walk FlatAST by index, emit indented JSON
- Used for golden testing and debugging

### Internal APIs Introduced

```go
// Lexer API
type Lexer struct { src []byte; pos int; indentStack []int; tokens []Token }
func NewLexer(src []byte) *Lexer
func (l *Lexer) Tokenize() ([]Token, []LexError)

// AST API
type Ast struct {
    Nodes     []AstNode
    ExtraData []uint32   // overflow payload storage
    Source    []byte     // original source (for string extraction)
    Strings   StringPool // interned strings
    RootIdx   uint32
}
func NewAst() *Ast
func (a *Ast) AddNode(kind NodeKind, tokenIdx uint32) uint32
func (a *Ast) SetChild(parent, child uint32)
func (a *Ast) SetSibling(node, sibling uint32)
func (a *Ast) NodeString(idx uint32) string // extract source text for token

// Parser API
type Parser struct { tokens []Token; ast *Ast; pos int; errors []ParseError }
func NewParser(tokens []Token, ast *Ast) *Parser
func (p *Parser) ParseFile() (rootIdx uint32, errs []ParseError)
```

### Test Strategy

**Unit tests:**
- Lexer: each token kind, indentation sequences, INDENT/DEDENT balance, error tokens
- Parser: each statement form, each expression form, precedence rules, error recovery

**Golden tests:**
- Input: `tests/golden/lexer/*.ax` → expected token sequence JSON
- Input: `tests/golden/parser/*.ax` → expected AST JSON
- Test runner: `go test ./compiler/lexer/... -update` to regenerate goldens

**Fuzzing:**
- `go test -fuzz FuzzLex` — random bytes as input, verify no panic
- `go test -fuzz FuzzParse` — random token sequences, verify no panic, only ErrorNodes produced

**Compliance suite smoke test:**
- Parse all 19 test suite `.ax` files, verify no ICE (Internal Compiler Error), ErrorNode count = 0 for valid files

### Acceptance Criteria
- Parse `axiom_compliance_suite.ax` with 0 ErrorNodes
- Lexer fuzz: 1M iterations, 0 panics
- Parser fuzz: 1M iterations, 0 panics
- All 50+ golden tests pass
- `axc dump-ast` produces valid JSON

### Definition of Done
- [ ] Lexer handles all 19 test files without panic
- [ ] Parser handles all 19 test files without panic
- [ ] All golden tests committed and passing
- [ ] Fuzz targets written (not necessarily run to completion)
- [ ] `axc dump-ast` command works

### Risks
- Indentation ambiguity (e.g., inconsistent mixed indent levels)
- Pratt parser precedence bugs in complex expressions

### Mitigation
- Strict 4-space rule: any non-4 indentation is a LexError, never silently misparse
- Precedence table reviewed against all test files manually before implementation

### Estimated Complexity: Medium

### Dependencies: Phase 0

---

## Phase 2 — Name Resolution + Type Checker (Core)

### Goals
Given a FlatAST, resolve all names, infer types, and produce a TypedAST. Basic types only: `i8–i64`, `f32/f64`, `bool`, `string`, `char8`, `void`. Structs, functions, basic generics. No ownership semantics yet.

### Inputs
- Phase 1 outputs (FlatAST, StringPool)
- Type checker spec (spec file 04)
- `axiom_compliance_suite.ax` groups 1–5 (primitives, control flow, functions, structs, generics)

### Outputs
- `compiler/sema/` — symbol table, scope stack, name resolver
- `compiler/types/` — type table, type inference engine
- `compiler/sema/mono.go` — monomorphization engine
- TypedAST (FlatAST with payload = TypeID on every node)
- `axc check <file.ax>` — reports type errors, exits 0 if clean
- Error messages: `file.ax:42:7: type mismatch: expected i32, found f64`

### Components Implemented

**Symbol Table:**
- `StringPool`: arena-backed string intern table, O(1) dedup
- `SymbolTable`: flat `[]Symbol` array
- `ScopeStack`: `[]map[uint32]uint32` (hash(name) → symbol_idx), push/pop on block entry/exit
- `Symbol` struct: `{ NameIdx, TypeIdx, AstNodeIdx, ScopeLevel, Flags }`
- Lazy Field Analysis: when `import X` encountered, load only module name; only resolve `X.field` on first access

**Type System:**
- `TypeTable`: `[]TypeEntry` — primitive types pre-registered at index 0–15
- Primitive types: `i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32`, `u64`, `f32`, `f64`, `bool`, `string`, `char8`, `char16`, `char32`, `isize`, `usize`, `void`
- `StructType`: name, fields `[]FieldEntry{name, typeIdx, offset}`
- `FunctionType`: params `[]TypeIdx`, return `TypeIdx`, effects `[]EffectID`
- `GenericType`: unresolved `[T: Constraint]`
- `MonomorphicType`: concrete instantiation of generic
- `SumType` (`type X = A | B`): discriminated union
- `TupleType`: fixed-size heterogeneous
- `ArrayType[T, N]` and `Seq[T]` (dynamic)
- `Option[T]`, `Result[T, E]` — built-in sum types

**Type Inference (local HM):**
- Bottom-up: leaf nodes get types from literals, propagate upward
- Bidirectional for `let x = expr`: infer type from RHS
- Function return type: walk all `return` statements, unify types (fail if conflicting)
- Error on global inference: function signatures must always be explicit

**Monomorphization:**
- When `Box[i32]` first appears: clone function/struct AST subtree, substitute T→i32, register under mangled name
- Mangle scheme: `_AX_<module>_<name>_<type1>_<type2>`
- Deduplication: same concrete instantiation used multiple times → single copy

**Overload Resolution:**
- Scoring: exact match (4), generic match (3), subtype match (2), implicit widening (1)
- Ambiguous if two candidates tie at max score → compile error

**Effects System (basic):**
- `{.raises: [X].}` propagates through call graph
- Pure functions: no `{.raises}` annotation + no I/O calls → marked `Pure` in TypeID flags
- `@[ai::assert_pure]` → verified at compile time

**Async typing:**
- `async fn` annotated in TypeID flags; return type remains `T` (not `Promise[T]`)
- State-machine rewrite deferred to IR builder phase

### Internal APIs Introduced

```go
// Type resolution
type TypeID uint32
type TypeChecker struct {
    ast    *Ast
    types  *TypeTable
    syms   *SymbolTable
    scope  *ScopeStack
    errors []TypeError
}
func (tc *TypeChecker) Check() []TypeError
func (tc *TypeChecker) TypeOf(nodeIdx uint32) TypeID

// MonoEngine
type MonoEngine struct { cache map[MonoKey]uint32; queue []MonoWork }
func (m *MonoEngine) Instantiate(genericFnIdx uint32, typeArgs []TypeID) uint32

// Connection Graph (placeholder in this phase)
type ConnectionGraph struct { nodes []CGNode; edges []CGEdge }
```

### Test Strategy

**Unit tests:**
- Type inference for each primitive
- Overload resolution scoring
- Monomorphization: `Box[i32]` and `Box[f64]` produce distinct type entries
- Effect propagation: function calling I/O function must declare IOError

**Golden tests:**
- Input: typed snippet → expected TypeID annotations as JSON
- `axc check` exit codes: 0 on clean, 1 on type errors, 2 on ICE

**Compliance tests:**
- `axiom_compliance_suite.ax` groups 1–5: `axc check` must exit 0

**Error message tests:**
- Input: intentionally broken code → expected error string snapshots

### Acceptance Criteria
- `axc check axiom_compliance_suite.ax` exits 0 for groups 1–5
- Monomorphization produces distinct TypeIDs for `Box[i32]` vs `Box[f64]`
- 0 regressions across Phase 1 golden tests

### Definition of Done
- [ ] `axc check` command reports type errors with file/line/col
- [ ] All primitive types, structs, generics working
- [ ] Effects system propagating through call graph
- [ ] Monomorphization creating concrete copies
- [ ] 100% unit test coverage of type inference rules

### Risks
- Generic type inference interaction with structural duck-typing is subtle
- Lazy Field Analysis may miss cross-file cycles

### Mitigation
- Implement structural matching as a separate pass with an explicit `satisfies(T, Interface)` function
- Cycle detection in lazy analysis via a `Resolving` flag per symbol (detect re-entrancy)

### Estimated Complexity: High

### Dependencies: Phase 1

---

## Phase 3 — Ownership Semantics + Connection Graph + CTGC

### Goals
Implement single ownership, `lent` borrows, `sink` transfer, `Isolated[T]`, generational reference ABI, and CTGC (compile-time GC: auto `=destroy` insertion + in-place reuse). This is the core safety innovation of AXIOM.

### Inputs
- Phase 2 outputs (TypedAST, SymbolTable)
- Type checker spec section 4 (ownership rules)
- `axiom_compliance_suite.ax` group 4 (structs + ownership)

### Outputs
- `compiler/sema/ownership.go` — ownership rules + linear type checking
- `compiler/sema/connection_graph.go` — full Connection Graph implementation
- `compiler/sema/escape.go` — escape analysis on Connection Graph
- `compiler/sema/ctgc.go` — CTGC pass: insert `=destroy` nodes + alias reuse
- TypedAST augmented with `=destroy` / `alias` AST nodes at correct scopes
- `runtime/axalloc/genref.go` (or `.c`) — generational reference runtime: 8-byte header, gen_id compare

### Components Implemented

**Ownership Rules (enforced by type checker):**
- `let x = expr` — x is immutable owner
- `mut x := expr` — x is mutable owner
- Assignment `a = b` where b is heap-owning → **move semantics**: b is invalidated, flag b's symbol as `Moved`
- Accessing a `Moved` symbol → compile error: "use after move"
- `lent` parameter / `&x` borrow → creates a `RefNode` with `Borrows` edge, lifetime bounded to owner's scope
- `sink` parameter (`!T`) → transfers ownership into function, caller's symbol invalidated
- `Isolated[T]` → proved at compile time to have no external references (Connection Graph: no incoming edges from outside isolation boundary)

**Connection Graph:**
```go
type CGNodeKind uint8  // Value, Ref, Operation
type CGEdgeKind uint8  // Owns, Borrows, FlowsTo, EscapesTo
type CGNode struct { ID uint32; Kind CGNodeKind; SymbolIdx uint32; MemClass MemoryClass }
type CGEdge struct { From, To uint32; Kind CGEdgeKind }
type ConnectionGraph struct {
    Nodes []CGNode
    Edges []CGEdge
    OutEdges map[uint32][]uint32
    InEdges  map[uint32][]uint32
}
func (g *ConnectionGraph) EscapesFunction(nodeID uint32) bool
func (g *ConnectionGraph) IsIsolated(nodeID uint32) bool
func (g *ConnectionGraph) LivenessInterval(nodeID uint32) (start, end uint32)
```

**Escape Analysis:**
- For each heap-allocated variable: if no EscapesTo edge exists AND no Borrows edge leaves the function → mark `StackAlloc`
- CTGC pass converts `alloc.heap` → `alloc.stack` for stack-allocatable nodes
- Object reuse: when variable A's liveness ends and variable B is allocated with same type/size → insert `alias B = A` instruction

**Generational Reference Runtime:**
```c
// axalloc ABI: every heap allocation gets 8-byte header
typedef struct { uint64_t gen_id; } AxHeader;
// Pointer stores: (actual_address, gen_id_at_time_of_borrow)
typedef struct { void* ptr; uint64_t gen_id; } AxGenRef;
// On dereference: if (ref.gen_id != ((AxHeader*)ref.ptr - 1)->gen_id) panic()
// On free: ((AxHeader*)ptr - 1)->gen_id++ (invalidate all references)
```

**`=destroy` Insertion:**
- At end of each scope block: for every owned variable declared in that scope, insert `AstNode{Kind: DestroyExpr, Payload: symbolIdx}` as the last child of the block
- For sink parameters: insert destroy at the return point of the callee (ownership transferred)
- C-Backend translates `DestroyExpr` → `ax_free(ptr)` which increments gen_id then calls AxAlloc

**`in [Arena]` Block:**
- Disable generational ID check for all allocations within the block
- At block exit: single `ax_arena_free(arena)` call
- Type checker: verify no `lent` reference escapes the arena block

### Internal APIs Introduced

```go
// Ownership checker
type OwnershipChecker struct {
    ast  *Ast; types *TypeTable; graph *ConnectionGraph
    symbols *SymbolTable; errors []OwnershipError
}
func (oc *OwnershipChecker) Check() []OwnershipError

// CTGC pass
type CTGCPass struct { air *AirProgram; graph *ConnectionGraph }
func (p *CTGCPass) Run() // mutates air in-place: alloc.heap→alloc.stack, inserts alias
```

### Test Strategy

**Unit tests:**
- Ownership: use-after-move is detected
- Ownership: use-after-free via gen_id mismatch triggers panic (runtime test)
- Isolated: passing non-isolated value to spawn → compile error
- Arena: lent reference escaping arena → compile error
- CTGC: verify alloc.heap converted to alloc.stack for non-escaping objects
- CTGC: verify alias insertion reduces alloc count

**Property tests (fuzzing):**
- Generate random ownership patterns, verify type checker never ICE
- Verify: for any valid program, `=destroy` count ≤ owner declaration count

**Compliance tests:**
- `axiom_compliance_suite.ax` group 4 (tests 031–040) must pass
- `axiom_lowlevel_suite.ax` unsafe/arena tests must pass

**Runtime tests:**
- Compile and run programs that trigger gen_id mismatch: verify panic (not segfault)

### Acceptance Criteria
- Use-after-move detected at compile time for all test cases
- Gen_id mismatch triggers clean panic (verified via runtime test)
- CTGC reduces heap allocation count by ≥ 20% on Fibonacci benchmark
- 0 regressions on Phase 1–2 tests

### Definition of Done
- [ ] Ownership rules enforced for all test cases in group 4
- [ ] Connection Graph builds correctly for all compliance suite cases
- [ ] CTGC pass runs, stack-allocates non-escaping objects
- [ ] Gen-ref runtime wired into C-backend output
- [ ] Arena blocks tested

### Risks
- Path-sensitive analysis of Connection Graph may produce false positives (forcing heap when stack is safe)

### Mitigation
- Start with flow-insensitive analysis (conservative: if any path escapes, treat as heap), add path sensitivity incrementally
- Escape analysis false positive rate tracked as a metric: must be < 10% on standard benchmarks

### Estimated Complexity: Extreme

### Dependencies: Phase 2

---

## Phase 4 — C-Backend (MVP Compiler Complete)

### Goals
Translate TypedAST (with ownership annotations, `=destroy` nodes) directly to C11 source. `axc build main.ax` produces a working executable by invoking GCC/Clang.

### Inputs
- Phase 3 outputs (TypedAST with ownership, `=destroy`, gen-ref)
- C-backend spec (spec file 02 section 7)
- MVP targets from `01.minimal core.md`

### Outputs
- `codegen/cgen/` — complete C-backend
- `runtime/axalloc/axalloc.c` — MVP allocator (malloc + 8-byte header + gen_id)
- `runtime/panic/panic.c` — panic handler (stack trace + exit)
- `axc build <file.ax>` — end-to-end command: produces runnable executable
- All 100 tests in `axiom_compliance_suite.ax` compiling and passing
- `axc build` on Hello World: produces correct output
- `axc build --emit-c` flag: write intermediate .c file to disk for debugging

### Components Implemented

**C-Backend Mapping Table:**

| AXIOM construct | C11 output |
|---|---|
| `fn foo(a: i32) -> i32` | `int32_t _AX_foo(int32_t a)` |
| `pub fn main()` | `int main(void)` |
| `let x: i32 = 42` | `int32_t x = 42;` |
| `mut x := 42` | `int32_t x = 42;` |
| `if cond:` | `if (cond) {` |
| `for i in 0..N:` | `for (int32_t i = 0; i < N; i++) {` |
| `for x in arr:` | range-based loop via AxArray iterator |
| `match x: A => B` | `switch`/if-else chain |
| `struct Foo { x: i32 }` | `typedef struct { int32_t x; } Foo;` |
| `spawn f()` | `pthread_create(...)` (MVP) |
| `await expr` | direct call (MVP: synchronous) |
| `async fn f()` | synchronous in MVP |
| `=destroy(x)` | `ax_free(&x)` |
| `alias b = a` | pointer cast (CTGC reuse) |
| `extern "C" fn puts(s: string) -> i32` | `extern int32_t puts(const char* s);` |
| `unsafe { }` | code block, no gen_id checks |
| `in [Arena] { }` | `ax_arena_t _arena_N; ... ax_arena_free(&_arena_N);` |
| `#run expr` | compile-time evaluated constant |
| `Result[T, E]` / `Option[T]` | tagged union struct |
| `Isolated[T]` | same as T in C (constraint is compile-time only) |

**Runtime (MVP):**
```c
void* ax_alloc(size_t size) {
    AxHeader* h = (AxHeader*)malloc(sizeof(AxHeader) + size);
    h->gen_id = 1;
    return h + 1;
}
void ax_free(void* ptr) {
    AxHeader* h = ((AxHeader*)ptr) - 1;
    h->gen_id++;  // invalidate all refs
    free(h);
}
void ax_deref_check(void* ptr, uint64_t expected_gen) {
    AxHeader* h = ((AxHeader*)ptr) - 1;
    if (h->gen_id != expected_gen) ax_panic("Generational reference mismatch");
}
```

**Build Automation:**
```go
func Build(file string, opts BuildOpts) error {
    src     := readFile(file)
    tokens  := Lex(src)
    ast     := Parse(tokens)
    typed   := TypeCheck(ast)
    owned   := OwnershipCheck(typed)
    ctgc    := CTGCPass(owned)
    ccode   := CBackend(ctgc)
    cfile   := writeTempFile(ccode, ".c")
    return execGCC(cfile, opts.Output, opts.Optimize)
}
```

### Test Strategy

**End-to-end tests:**
- Compile and run: Hello World, Fibonacci, Factorial
- Compile and run: all 100 tests in `axiom_compliance_suite.ax`
- Compile and run: `axiom_lowlevel_suite.ax`
- Compile and run: `axiom_functional_suite.ax`

**C output snapshot tests:**
- Golden `.c` files for known inputs; diffs reported on changes

**Memory safety tests:**
- Programs that trigger gen_id mismatch → verify panic exit (not segfault)
- Programs with intentional use-after-free → compiler rejects at compile time

**Benchmark baseline:**
- Record: compile time for Hello World, 1000-line file, 10000-line file
- Record: runtime performance of Fibonacci(40) vs equivalent C

### Acceptance Criteria
- `axc build hello.ax` produces running executable: exits 0, outputs "Hello, World!"
- All 100 compliance suite tests pass when compiled and run
- Compile time for a 1000-line file: < 500ms on dev machine
- Gen-id mismatch detected at runtime: clean panic (no segfault)

### Definition of Done
- [ ] C-backend covers all grammar constructs
- [ ] MVP runtime (axalloc + panic) integrated
- [ ] `axc build` command working end-to-end
- [ ] All 100 compliance suite tests passing
- [ ] Compile time benchmarks baseline committed

### Risks
- C name mangling conflicts with existing C library symbols
- GCC/Clang availability varies across target platforms

### Mitigation
- All AXIOM symbols prefixed `_AX_` — reserved namespace in C standard
- Ship bundled Zig C compiler (which bundles clang) as optional fallback for hermetic builds

### Estimated Complexity: High

### Dependencies: Phase 3

---

## Phase 5 — AIR (Axiom IR) + Optimization Pipeline

### Goals
Introduce the full AIR layer between TypedAST and code generation. Implement core optimization passes.

### Outputs
- `ir/air/` — AirInst definitions, basic block graph, CFG
- `ir/builder/` — TypedAST → AIR lowerer
- `ir/opt/` — optimization passes: ctgc, const_fold, dce, inline, mono, loop_region
- `axc dump-air` — print AIR as text
- Updated C-backend: now reads from AIR instead of TypedAST directly

### AIR Instruction Set (complete):
```
// Memory
alloc.stack <type> → %r          // stack allocation
alloc.heap  <type> → %r          // heap allocation (may become alloc.stack)
load.T  [%r + offset] → %r      // typed load
store.T [%r + offset], %r       // typed store
free    %r                       // explicit free
alias   %r1 = %r2               // CTGC: reuse r2's memory for r1
gen_check %r, %gen_id           // generational ref check

// ALU
add.T / sub.T / mul.T / div.T / rem.T
icmp.{eq,ne,lt,le,gt,ge} → %bool
zext / sext / trunc / fpext / fptrunc
bitand / bitor / bitxor / shl / shr

// Control flow
jmp <block> / br %cond, <bt>, <bf>
call @func, [args] → %r
ret [%r]
phi [(%r1, <b1>), (%r2, <b2>)] → %r
loop_region <start>, <end>, parallel=bool
unreachable

// SIMD/Vector
v_load.T / v_store.T / v_add.T / v_mul.T / v_fma.T
dispatch <kernel>, %grid, %block  // GPU

// Quantum (stub)
qalloc → %q / qgate %q, GateKind / qmeasure %q → %bool
```

**Optimization Passes (ordered):**
1. Inline (small functions ≤ 20 instructions)
2. Constant Folding + Propagation
3. CTGC + Escape Analysis (alloc.heap → alloc.stack, alias insertion)
4. DCE (unreachable block removal)
5. Auto-SoA (opt-in only: @SOA annotation)
6. Loop Vectorization (loop_region parallel=true)

### Acceptance Criteria
- All Phase 4 tests continue passing through new AIR pipeline
- CTGC pass reduces `alloc.heap` count by ≥ 20% on standard benchmarks
- DCE eliminates `if (false)` branches
- Constant folding eliminates `2+3` at compile time

### Estimated Complexity: High

### Dependencies: Phase 4

---

## Phase 6 — Native Backend (x86-64)

### Goals
Implement full native code generation for x86-64. Produce ELF executables without invoking GCC/Clang.

### Outputs
- `codegen/native/x86/` — x86-64 instruction selector, linear scan reg alloc, stack frame, emitter
- `codegen/linker/` — in-memory linker (ELF, PE, Mach-O)
- `.axmeta` writer (Zstd-compressed JSON)
- `axc build --target x86-linux` — produces ELF binary

### Key Components

**Linear Scan Register Allocator:**
- Liveness analysis: compute [start, end] for each virtual register
- Sort by start; assign free physical reg; if none: spill longest-end interval
- x86-64 GPR pool: 14 usable (after RSP/RBP reserved)
- XMM pool: XMM0–XMM15 for float/SIMD

**Stack Frame (x86-64 System V):**
- Args 1–6: RDI, RSI, RDX, RCX, R8, R9
- Return: RAX (or RAX+RDX for structs ≤ 16 bytes)
- Large struct return: hidden `sret` pointer as first argument

**In-Memory Linker:**
- Global symbol table: open-addressing hash map
- Relocation queue + backpatching (PC-relative offsets)
- ELF64: `.text` (RX), `.rodata` (R), `.data` (RW), `.axmeta` (R)
- PE/COFF: IAT for kernel32.dll/msvcrt.dll
- Mach-O: ad-hoc code signing required on Apple Silicon

**`.axmeta` Section:**
- JSON: ConnectionGraph + SymbolTable + EffectProfiles
- Compressed with Zstd (level 3)
- Emitted only with `--emit-meta` or debug builds

### Acceptance Criteria
- All 100 compliance tests pass via native backend (ELF on Linux)
- Fibonacci(40) native backend: ≤ 5% slower than clang -O2
- GDB can step through source lines via DWARF info
- `.axmeta` section present and decompressable

### Estimated Complexity: Extreme

### Dependencies: Phase 5

---

## Phase 7 — ARM64 + RISC-V Backends

### Goals
Extend native backend to ARM64 (Apple Silicon, Linux ARM) and RISC-V 64-bit.

### Outputs
- `codegen/native/arm64/` — AAPCS64 calling convention, NEON SIMD
- `codegen/native/riscv64/` — RV64I base, psABI
- `axc build --target aarch64-linux|riscv64-linux`

### Estimated Complexity: High

### Dependencies: Phase 6

---

## Phase 8 — AxAlloc Production + Actor Runtime

### Goals
Replace MVP malloc-wrapper with production AxAlloc. Implement M:N work-stealing scheduler and full Actor runtime.

### Outputs
- `runtime/axalloc/` — size-classed, bump-pointer, NUMA-aware, lock-free
- `runtime/scheduler/` — M:N work-stealing with reduction budget preemption
- `runtime/actors/` — Actor struct, spawn, mailbox, supervisor tree
- `runtime/async/` — state machine executor (epoll/kqueue/IOCP)
- `runtime/channels/` — bounded + unbounded channels

### AxAlloc Key Design
- Size classes: 8, 16, 32, 64, 128, 256, 512, 1024 bytes → segments
- Segment: 64KB, all blocks same size, bump pointer allocation
- Per-actor heap: no global lock, 100% lock-free for normal allocations
- Free: push to per-size-class LIFO free list, increment gen_id
- Actor crash: return all segments to OS in O(1) via munmap

### Actor Model
```
Actor { id, state, heap: *AxHeap, mailbox: Deque[Msg],
        stack, budget: u32, supervisor: ActorID }
```

### Benchmark Targets
- Actor spawn: < 1μs per actor
- Message throughput: > 10M msgs/sec single-core
- AxAlloc throughput: > 500M alloc/free pairs/sec
- NUMA benefit: < 5% latency increase for cross-NUMA access

### Estimated Complexity: Extreme

### Dependencies: Phase 6

---

## Phase 9 — Standard Library

### Goals
Implement core standard library in AXIOM itself. Primary dogfooding test.

### Outputs
- `std/collections/` — List[T], Map[K,V], Set[T], Deque[T]
- `std/string.ax` — UTF-8 operations
- `std/math.ax` — abs, pow, sqrt, trig
- `std/io.ax` — print, println, read_line
- `std/fs.ax` — file open/read/write/close/path
- `std/net/` — async TCP/UDP/HTTP, URL parsing
- `std/concurrency.ax` — Channel, Actor, Locker, spawn, await_all
- `std/crypto.ax` — SHA-256, SHA-512, ChaCha20
- `std/time.ax` — timestamps, duration, sleep
- `std/testing.ax` — assert, test runner
- `std/mem.ax` — Arena, addr, size_of, align_of, byte_swap
- `std/arch/x86.ax` — SIMD intrinsics (_mm256_*)
- `std/compiler/ai.ax` — AI optimization API (std.compiler.ai)
- `std/quantum.ax` — stub (bool-based simulator)
- `std/gpu.ax` — stub (CPU fallback)

### Estimated Complexity: High (volume)

### Dependencies: Phase 7, Phase 8

---

## Phase 10 — Tooling (LSP, Formatter, Package Manager)

### Outputs
- `axc fmt` — zero-configuration formatter (idempotent, 4-space canonical)
- `axc lsp` — LSP 3.17 server (diagnostics, hover, completion, definition)
- `axc get` — package manager (axiom.toml, axiom.lock, git-based, PubGrub SAT solver)
- `axc doc` — documentation generator (/// docstrings → HTML/MD)
- `axc fix` — AI-assisted automatic code migration
- Real-time `.axmeta` emission on save for AI tooling

### Estimated Complexity: Medium

### Dependencies: Phase 5 (type checker for LSP), Phase 9 (stdlib)

---

## Phase 11 — Self-Hosting Bootstrap

See Section 7 for full detail.

### Dependencies: Phase 9, Phase 10

---

## SECTION 3 — DETAILED COMPILER PIPELINE PLAN

### 3.1 FlatTokenArray

**Purpose:** Zero-allocation token stream, 8 bytes per token. Source buffer never copied.

**Ownership:** Lexer owns source buffer (borrowed from caller). Token array owned by caller after `Tokenize()`.

**Verification rules:**
- INDENT/DEDENT tokens balanced in final stream
- Each token: `Offset + Len ≤ len(source)`
- No overlapping tokens

**Serialization:** `[{"kind":"FN","offset":0,"len":2}, ...]`

**Debug:** `axc dump-tokens <file.ax>` prints kind names from source positions

**Testing:** Fuzz with arbitrary bytes; assert no panic, INDENT/DEDENT balance

---

### 3.2 FlatAST

**Purpose:** Universal intermediate between lexical and semantic analysis.

**Ownership:** `Ast` owns `[]AstNode`, `StringPool`, `ExtraData`. Passes borrow immutably.

**Mutability:** Parser writes once. Semantic passes annotate `payload`/`flags` in-place. `=destroy` nodes appended.

**Key invariants:**
- Node 0 = module root
- `first_child`/`next_sibling` indices within bounds or 0
- No cycles (verified by DFS)
- After type check: every non-error node has `payload` = valid TypeID

**Serialization:** `{"kind":"FnDecl","token":"fn","children":[...],"type":"i32→i32"}`

**Debug:** `axc dump-ast --types <file.ax>`

**Testing:** DFS count == live node count. Snapshot tests for every grammar construct.

---

### 3.3 TypedAST (annotated FlatAST)

**Purpose:** FlatAST where every expression node has a resolved TypeID.

**Key invariants:**
- Every `BinaryExpr`: `payload` = TypeID of result
- Every `FnDecl`: `payload` = FunctionTypeID (params + return + effects)
- No TypeID(0) on non-error nodes (TypeID(0) = Unknown = error)

---

### 3.4 AIR (Axiom IR)

**Purpose:** SSA-form IR. Machine-independent. Preserves loop regions and ownership metadata.

**Ownership:** `AirProgram` owns `[]AirInst`, `[]BasicBlock`, `MetadataTable`.

**Mutability:** Passes modify in-place. Dead instructions marked `Opcode = NOP`.

**Verification rules:**
- SSA: each virtual register defined exactly once
- Every register defined before first use (dominance check)
- Every basic block ends with terminator (jmp/br/ret/unreachable)
- `loop_region` start/end = valid block indices

**Text format:**
```
fn _AX_main():
  bb0:
    %r1 = alloc.stack i32
    store.i32 [%r1], 42
    %r2 = load.i32 [%r1]
    call @_AX_io_print, %r2
    ret void
```

**Debug:** `axc dump-air [--before-opt | --after-opt] <file.ax>`

**Testing:** AIR verifier runs after every optimization pass in debug mode.

---

### 3.5 MachineIR (post-instruction-selection)

**Purpose:** Target-specific instructions with virtual registers. Input to register allocator.

```go
type MachInst struct {
    Op       x86Op; Dest VReg; Src1 VReg; Src2 VReg
    Size     OpSize; AddrMode AddrMode
}
```

---

### 3.6 PhysicalMachineCode (post-register-allocation)

**Purpose:** Final byte sequence + relocation table + symbol table.

- `[]byte code` — raw machine code
- `[]Relocation` — (offset, symbol_name, reloc_type) to backpatch
- `[]Symbol` — exported names and code offsets
- Stack layout map — for DWARF debug info

---

## SECTION 4 — DIRECTORY STRUCTURE PLAN

```
axiom-lang/
│
├── cmd/axc/main.go              # CLI entry: build, check, fmt, lsp, get, fix,
│                                #            dump-ast, dump-air, dump-tokens
│
├── compiler/
│   ├── lexer/
│   │   ├── lexer.go             # DFA scanner, INDENT/DEDENT emission
│   │   └── token.go             # Token struct (8 bytes), TokenKind enum
│   ├── ast/
│   │   ├── ast.go               # AstNode struct, Ast, NodeKind enum
│   │   ├── printer.go           # JSON printer
│   │   └── visitor.go           # DFS traversal helpers
│   ├── parser/
│   │   ├── parser.go            # Recursive Descent
│   │   ├── pratt.go             # Pratt expression parser
│   │   ├── indent.go            # Indentation stack
│   │   └── recovery.go          # Panic-mode error recovery
│   ├── sema/
│   │   ├── symbol.go            # Symbol, SymbolTable, ScopeStack
│   │   ├── resolver.go          # Name resolution + Lazy Field Analysis
│   │   ├── typechecker.go       # Main type checker pass
│   │   ├── inference.go         # Local HM type inference
│   │   ├── overload.go          # Overload resolution scoring
│   │   ├── effects.go           # Effect system propagation
│   │   ├── ownership.go         # mut / lent / sink / Isolated rules
│   │   ├── connection_graph.go  # Connection Graph: nodes, edges, queries
│   │   ├── escape.go            # Escape analysis
│   │   ├── ctgc.go              # =destroy injection, alias insertion
│   │   ├── mono.go              # Monomorphization engine
│   │   └── isolated.go          # Isolated[T] constraint verification
│   └── types/
│       ├── types.go             # TypeEntry, TypeID, TypeTable
│       ├── primitive.go         # Built-in primitives
│       ├── generic.go           # Generic type representation
│       └── sum.go               # Sum type / union type
│
├── ir/
│   ├── air/
│   │   ├── inst.go              # AirInst struct (16 bytes), Opcode enum
│   │   ├── block.go             # BasicBlock, CFG
│   │   ├── program.go           # AirProgram
│   │   ├── verifier.go          # SSA invariant checker
│   │   └── printer.go           # Text format printer
│   ├── builder/
│   │   ├── builder.go           # TypedAST → AIR lowering
│   │   └── statemachine.go      # async fn → state machine lowering
│   └── opt/
│       ├── pipeline.go          # Pass manager (ordered execution)
│       ├── inline.go            # Function inlining
│       ├── constfold.go         # Constant folding + propagation
│       ├── dce.go               # Dead code elimination
│       ├── ctgc.go              # Compile-time GC on AIR
│       ├── soa.go               # Auto-SoA struct layout transform
│       ├── vectorize.go         # Loop vectorization (SIMD)
│       └── comptime.go          # #run bytecode interpreter
│
├── codegen/
│   ├── cgen/
│   │   ├── cgen.go              # AIR → C11 source
│   │   └── runtime_h.go         # Embedded axruntime.h (gen_ref, axalloc)
│   └── native/
│       ├── target.go            # Target triple (Arch, OS, ABI)
│       ├── x86/
│       │   ├── selector.go      # AIR → MachInst (x86-64)
│       │   ├── regalloc.go      # Linear scan register allocator
│       │   ├── frame.go         # Stack frame layout
│       │   ├── emit.go          # MachInst → bytes
│       │   ├── abi_sysv.go      # System V AMD64 ABI
│       │   └── abi_win64.go     # Windows x64 ABI
│       ├── arm64/
│       │   ├── selector.go; regalloc.go; frame.go; emit.go
│       │   └── abi_aapcs64.go
│       └── riscv64/
│           ├── selector.go; emit.go
│           └── abi_riscv.go
│
├── linker/
│   ├── linker.go                # In-memory linker orchestration
│   ├── symbols.go               # Global symbol table, mangling
│   ├── reloc.go                 # Relocation queue + backpatching
│   ├── elf.go                   # ELF64 format writer
│   ├── pe.go                    # PE/COFF format writer (Windows)
│   ├── macho.go                 # Mach-O format writer (macOS)
│   └── axmeta.go                # .axmeta section: JSON → Zstd → bytes
│
├── runtime/
│   ├── axalloc/
│   │   ├── axalloc.c            # Production allocator
│   │   ├── axalloc.h            # Public API header
│   │   ├── genref.c             # Generational reference check
│   │   ├── numa.c               # NUMA-aware segment allocation
│   │   └── gpu_pinned.c         # Pinned memory for GPU transfer
│   ├── scheduler/
│   │   ├── scheduler.c          # M:N work-stealing scheduler
│   │   └── thread_pool.c        # OS thread management
│   ├── actors/
│   │   ├── actor.c              # Actor struct, spawn, mailbox
│   │   ├── supervisor.c         # Supervisor tree, restart policies
│   │   └── channel.c            # Bounded + unbounded channels
│   ├── async/
│   │   ├── executor.c           # State machine executor
│   │   ├── epoll.c / kqueue.c / iocp.c   # Platform I/O reactors
│   └── panic/
│       ├── panic.c              # Panic handler
│       └── stacktrace.c         # Stack unwinding (frame pointer)
│
├── std/                         # Standard library in AXIOM (.ax)
│   ├── collections/ (list, map, set, deque)
│   ├── string.ax / math.ax / io.ax
│   ├── fs.ax / time.ax / crypto.ax
│   ├── net/ (tcp, http, url)
│   ├── concurrency.ax / testing.ax / mem.ax
│   ├── arch/ (x86.ax, arm.ax)
│   ├── compiler/ (ast.ax, ai.ax)
│   ├── gpu.ax (stub) / quantum.ax (stub)
│
├── tools/
│   ├── lsp/
│   │   ├── server.go            # JSON-RPC LSP server
│   │   ├── hover.go / completion.go / diagnostics.go
│   │   └── axmeta.go            # .axmeta emission for AI tooling
│   └── pkg/
│       ├── manager.go           # Package manager
│       ├── resolver.go          # PubGrub SAT solver
│       ├── git.go               # Git-based package fetching
│       └── lockfile.go          # axiom.lock SHA-256
│
├── tests/                       # 19 behavioral specification test suites
│   ├── axiom_compliance_suite.ax
│   ├── axiom_lowlevel_suite.ax
│   ├── (all 19 suites)
│   └── runner.go                # Compile + run + verify exit 0
│
├── bootstrap/
│   ├── stage0/                  # Symlink → compiler/ (Go implementation)
│   ├── stage1/compiler/*.ax     # Compiler rewritten in AXIOM
│   └── verify/triple_build.sh   # Build 3×, compare hashes
│
├── benchmarks/
│   ├── compile_time/ (bench_hello, bench_1kloc)
│   ├── runtime/ (fib.ax, alloc_stress.ax, actor_spawn.ax)
│   └── vs_clang/compare.sh
│
├── fuzz/
│   ├── fuzz_lex.go / fuzz_parse.go / fuzz_typecheck.go
│   └── corpus/                  # Seed corpus from test suites
│
├── docs/
│   ├── design/
│   │   ├── token-kinds.md / ast-schema.md / air-schema.md
│   │   ├── type-system.md / ownership.md / abi.md
│   │   └── axmeta-format.md
│   ├── rfcs/                    # Accepted RFCs
│   └── changelog/
│
├── ci/.github/workflows/
│   ├── build.yml                # Build + test on push/PR
│   ├── fuzz.yml                 # Scheduled fuzzing (daily, 1hr)
│   ├── bench.yml                # Benchmark tracking (weekly)
│   └── release.yml              # Release pipeline (tags)
│
├── examples/
│   ├── hello/main.ax
│   ├── fibonacci/main.ax
│   ├── http-server/             # Actor-based HTTP server
│   ├── game-loop/               # Real-time game loop with Arena memory
│   └── ai-pipeline/             # .axmeta + AI annotation example
│
├── GRAMMAR.ebnf                 # Canonical formal grammar
├── CONTRIBUTING.md              # Coding standards, review process
├── go.mod                       # Bootstrap compiler module
├── Makefile                     # build, test, bench, fuzz, release
└── axiom.toml                   # Package manifest for self-hosting
```

---

## SECTION 5 — TEST INFRASTRUCTURE PLAN

### 5.1 Test Categories

**Category 1: Unit Tests** (`*_test.go`)
- Each package tests its own public API in isolation
- Convention: `Test<Component>_<Scenario>`, e.g., `TestLexer_IndentDedentBalance`
- All unit tests complete in < 30s

**Category 2: Golden Tests** (snapshot files)
- `tests/golden/<component>/<input>.ax` + `<input>.expected.json`
- Runner: `go test ./... -update-golden` to regenerate
- Components: lexer, parser, typechecker, air-builder, cgen, all opt passes

**Category 3: Compiler End-to-End Tests**
- `tests/e2e/*.ax` — compile + run, verify exit code and stdout
- Expected behavior embedded in comments: `// EXPECT_EXIT: 0`, `// EXPECT_STDOUT: hello`

**Category 4: Compliance Tests (behavioral specs)**
- All 19 test suites are compliance tests
- A compiler version is "compliant" only when all 19 suites pass
- Priority: compliance_suite → lowlevel_suite → functional_suite → concurrency_suite → remaining 15

**Category 5: Memory Safety Tests**
- Use-after-free: compile-time rejection verified
- Gen_id mismatch: runtime panic verified (not segfault)
- Data races on non-Isolated data: compile-time rejection
- Run generated C under AddressSanitizer (C-backend phase)
- Run native binary under `rr` (record-replay) for non-determinism

**Category 6: Fuzzing**
- Targets: `FuzzLex`, `FuzzParse`, `FuzzTypeCheck`, `FuzzCodegen`
- Property: no panic, no ICE — only clean user-facing errors
- Schedule: 1 hour daily in CI (`ci/fuzz.yml`)
- Corpus: `fuzz/corpus/`, grown by CI findings

**Category 7: Property-Based Tests**
- Round-trip: `parse(print(ast)) == ast`
- Idempotency: `fmt(fmt(src)) == fmt(src)`
- SSA invariant: after IR builder, every vReg defined exactly once
- Determinism: compiling same source twice produces identical bytes

**Category 8: Differential Testing**
- C-backend output ≡ native backend output for all compliance tests
- Debug build ≡ release build (semantic behavior identical)

**Category 9: Performance Regression Tests**
- Benchmarks tracked with historical data in CI artifacts
- Fail if any benchmark regresses > 5% vs 10-build rolling average
- Key metrics: compile time (Hello World, 1k LOC, 10k LOC), Fibonacci(40) runtime, allocator throughput

**Category 10: ABI Compatibility Tests**
- `extern "C"` functions callable from C
- C libraries callable from AXIOM
- `@packed` struct layout matches C layout

**Category 11: Self-Hosting Tests** (Phase 11+)
- Bootstrap verification: compile compiler 3×, all three binaries produce identical output

### 5.2 Test Naming Convention

```
Go unit tests:   Test<Component>_<Scenario>_<Expected>
Fuzz targets:    Fuzz<Component>
Benchmarks:      Benchmark<Operation>_<Scale>
E2E test files:  <suite>_<number>_<description>.ax
```

### 5.3 Failure Diagnostics

Every test failure outputs:
1. Test name and file:line
2. Input (source snippet or token sequence)
3. Expected output
4. Actual output
5. Unified diff

Compiler ICE format:
```
axc: internal compiler error at compiler/sema/typechecker.go:342
  function: TypeChecker.inferBinaryExpr
  node: BinaryExpr [file.ax:10:5]
  please report at https://github.com/axiom-lang/axiom/issues
```

---

## SECTION 6 — BUILD + TOOLCHAIN PLAN

### 6.1 Bootstrap Compiler Strategy

```
Stage 0: Go-written axc → C output → GCC → binary         [Phase 1–4]
Stage 1: Go axc compiles Axiom-written axc → C → GCC       [Phase 11.1]
Stage 2: Axiom axc (stage 1) compiles itself → native      [Phase 11.2]
Stage 3: Stage 2 binary compiles itself → identical bytes  [Phase 11.3]
```

### 6.2 Build System

```
make build    # compile Go bootstrap compiler
make test     # all Go unit + integration tests
make e2e      # all AXIOM e2e compliance tests
make bench    # run benchmarks, save results
make fuzz     # fuzz targets for 60 seconds (quick mode)
make release  # cross-compile all targets, full test suite
```

**Incremental compilation:**
- Hash-based caching: `sha256(source)` → cached TypedAST and AIR in `~/.axiom/cache/`
- Cache key: `hash(source) + hash(all imports' hashes)` (transitive)
- Only recompile files whose hash changed

**Parallel compilation:**
- Parsing: one goroutine per source file (fully parallel)
- Type checking: topological order of import DAG
- Code generation: one goroutine per function

**Reproducible builds:**
- No timestamps in emitted output
- Symbol sort order: always lexicographic (never map iteration)
- Verify: `sha256(axc build --deterministic a.ax) == sha256(axc build --deterministic a.ax)`

**Cross compilation:**
- `AXIOM_TARGET=x86_64-linux|aarch64-linux|x86_64-windows|aarch64-macos`
- No cross-toolchain required (own emitters for all targets)
- Static linking against musl for hermetic Linux binaries

### 6.3 Package Manager

```toml
# axiom.toml
[package]
name = "my-app"
version = "1.0.0"

[dependencies]
http-client = { git = "https://github.com/...", tag = "v1.2.0" }
```

- `axiom.lock` — `{ pkg, git, commit, tree_hash: sha256_of_source_tree }`
- `axc get <name>@<version>` — add dependency
- `axc update` — update within semver constraints
- PubGrub SAT-based version solver (deterministic, no backtracking ambiguity)

### 6.4 LSP Integration

- Protocol: LSP 3.17 over stdio (JSON-RPC 2.0)
- Editor: VS Code extension (`axiom-lang.vscode`)
- Incremental sync: re-parse only changed functions (Phase 1 incremental parsing)
- Diagnostic latency target: < 100ms for single-file changes

---

## SECTION 7 — SELF-HOSTING STRATEGY

### Stage 0 — Go Bootstrap Compiler (Months 1–12)

**Language:** Go 1.22+
**Capabilities:** Full frontend, C-backend, basic optimizations
**Limitations:** No native backend, no M:N scheduler, MVP runtime only
**Verification:** All 19 compliance suites pass on Linux x86-64
**Transition:** `axc build axiom_compliance_suite.ax` → 100% pass

### Stage 1 — Compiler in AXIOM, Compiled via Stage 0 (Months 13–18)

**Goal:** Rewrite complete compiler in AXIOM source (`bootstrap/stage1/compiler/*.ax`). Compile with Stage 0 Go compiler → `axc.c` → GCC → new `axc` binary.

**Verification:**
```bash
# Build stage 1 binary using stage 0 (Go)
go run ./cmd/axc build ./bootstrap/stage1/compiler/ -o axc_stage1

# Stage 1 compiles itself
./axc_stage1 build ./bootstrap/stage1/compiler/ -o axc_stage1b

# Functional equivalence
./axc_stage1  build axiom_compliance_suite.ax -o test1 && ./test1
./axc_stage1b build axiom_compliance_suite.ax -o test2 && ./test2
diff <(./test1) <(./test2)  # Must be identical
```

**Transition:** `axc_stage1b` passes all 100 compliance tests; outputs identical between stage1 and stage1b.

### Stage 2 — Self-Hosting with Native Backend (Months 19–24)

**Goal:** Stage 1 compiler emits native code (no GCC). Port native backend from Go to AXIOM.

**Verification:**
```bash
./axc_stage2 build ./bootstrap/stage1/compiler/ -o axc_stage2b
./verify/triple_build.sh axiom_compliance_suite.ax  # 3× identical hashes
```

**Transition:** `axc_stage2 build main.ax` produces native ELF; compile speed ≥ 100k LOC/sec

### Stage 3 — Production Self-Hosting + Full Runtime (Months 25–36)

**Goal:** Runtime (AxAlloc, M:N scheduler) written in AXIOM. Zero dependency on libc for core runtime.

**Self-hosting definition:**
- Compiler source (`*.ax`) compiles itself
- Runtime source (`*.ax`, formerly `*.c`) compiles itself
- Standard library compiles itself
- Reproducible: `hash(axc build axc_source) == hash(axc build axc_source)` from clean state

**Transition:** `axc build --runtime=none` works; AxAlloc replaces libc malloc; compile speed ≥ 500k LOC/sec

### Transition Gate Protocol

For each stage transition:
1. Run all 19 compliance suites with new stage binary
2. Run triple-build verification (compile 3×, compare binary hashes)
3. Benchmarks: new stage must not be > 20% slower than previous
4. Code review by 2 engineers

---

## SECTION 8 — RUNTIME + MEMORY PLAN

### 8.1 AxAlloc Architecture

**Size classes:**
```
8, 10, 12, 16, 20, 24, 28, 32, 40, 48, 56, 64, 80, 96, 112, 128,
160, 192, 224, 256, 320, 384, 448, 512, 640, 768, 1024, 2048, 4096, 8192 bytes
>8192 bytes → mmap directly
```

**Segment structure (64KB):**
```
┌───────────────────────────────────────────┐
│ AxSegmentHeader (64 bytes)                │
│   magic, size_class, used_count,          │
│   free_list_head, next_segment, numa_node │
├───────────────────────────────────────────┤
│ Block 0: [AxHeader(8 bytes)] [data...]    │
│ Block 1: [AxHeader(8 bytes)] [data...]    │
│ ...                                       │
└───────────────────────────────────────────┘
```

**AxHeader layout (8 bytes):**
```c
typedef struct {
    uint64_t gen_id : 63;   // generation counter
    uint64_t is_free : 1;   // 0=alive, 1=freed
} AxHeader;
```

**Allocation fast path (~3 cycles):**
```c
void* ax_alloc(AxHeap* heap, size_t size) {
    int sc = size_class_index(size);
    AxSegment* seg = heap->segments[sc];
    if (seg->bump < seg->end) {
        AxHeader* h = (AxHeader*)seg->bump;
        h->gen_id = 1; h->is_free = 0;
        seg->bump += seg->block_size;
        return h + 1;
    }
    return ax_alloc_slow(heap, sc, size);
}
```

### 8.2 Thread Model

- N OS threads (one per logical CPU core, configurable via `AXIOM_THREADS=N`)
- Thread affinity: `pthread_setaffinity_np` / `SetThreadAffinityMask`
- Each OS thread owns one AxHeap
- Work-stealing: atomic deque per OS thread, lock-free CAS on tail

### 8.3 Async Runtime

- State machines (zero-heap): async fn lowered to switch-on-state at IR level
- I/O reactor: Linux epoll / macOS kqueue / Windows IOCP → unified as `AxIO`
- Wake: I/O completion → `ax_scheduler_wake(actor_id)` → push to run queue
- Timer: `timerfd` (Linux) / `kqueue EVFILT_TIMER` (macOS) / `CreateWaitableTimer` (Windows)

### 8.4 Memory Safety Model

| Violation | Detection | Response |
|---|---|---|
| Use-after-free | Gen_id mismatch at runtime | Safe panic + .axtrace |
| Use-after-move | Type checker at compile time | Compile error |
| Dangling lent ref | Type checker (lifetime scope) | Compile error |
| Data race (non-Isolated) | Type checker at compile time | Compile error |
| Arena ref escape | Type checker at compile time | Compile error |
| Out-of-bounds | Runtime bounds check | Safe panic |
| Integer overflow | Runtime check in debug mode | Safe panic |

### 8.5 Generational Reference ABI (only ABI contract between compiler and runtime)

```c
typedef struct { uint64_t gen_id; } AxHeader;
typedef struct { void* ptr; uint64_t gen_id; } AxRef;

static inline void* ax_deref(AxRef ref) {
    AxHeader* h = ((AxHeader*)ref.ptr) - 1;
    if (__builtin_expect(h->gen_id != ref.gen_id, 0))
        ax_panic("GenerationalID mismatch: use-after-free detected");
    return ref.ptr;
}
```

### 8.6 FFI Boundary

```axiom
// Import C function
extern "C" fn printf(fmt: *char8, ...) -> i32

// Export AXIOM function to C
@[export("my_func")]
pub fn my_func(x: i32) -> i32:
    return x * 2
```

C-ABI functions skip gen_id checking. `extern "C"` parameters treated as unsafe.

### 8.7 Panic Handling

```c
void ax_panic(const char* msg) {
    // 1. Stop actors on current OS thread
    // 2. Print message + stack trace to stderr
    // 3. Write .axtrace if --emit-meta was set
    // 4. exit(101)
}
```

---

## SECTION 9 — OPTIMIZATION ROADMAP

### 9.1 Early Optimizations (Phase 4–5)

| Optimization | IR Level | Heuristic |
|---|---|---|
| Constant folding | TypedAST / AIR | Always apply |
| Dead code elimination | AIR CFG | Always apply |
| `=destroy` elimination | TypedAST | Always apply |
| Escape analysis → stack alloc | AIR + Connection Graph | Conservative: any escaping path → heap |
| `#run` compile-time eval | TypedAST | Pure functions only |

### 9.2 Mid-Level Optimizations (Phase 5–6)

| Optimization | IR Level | Heuristic |
|---|---|---|
| Function inlining | AIR | Callee < 20 instructions |
| Monomorphization | TypedAST / AIR | Always for generics |
| Object reuse (CTGC alias) | AIR + Connection Graph | Same type, non-escaping, liveness non-overlapping |
| Constant propagation | AIR dataflow | Always for const operands |
| Auto-SoA transform | AIR + @SOA | Only with @SOA or --ai-suggest |

### 9.3 Backend Optimizations (Phase 6+)

| Optimization | Level | Heuristic |
|---|---|---|
| Linear scan reg alloc | Machine IR | Spill longest live interval |
| Instruction selection fusion | Machine IR | `mul+add` → `IMAD` |
| Loop vectorization (SIMD) | AIR (loop_region) | Independent iterations, aligned arrays |
| LTO | Post-link | Release builds only |

### 9.4 Advanced Optimizations (Phase 8+)

| Optimization | Heuristic |
|---|---|
| PGO | Hot paths > 10% execution time |
| Layout reorganization (Adaptive) | Cache miss rate > threshold |
| Devirtualization | Interface call with single implementor |

### 9.5 What NOT to Optimize Early

- Peephole optimization — < 1% benefit, high complexity. Defer to Stage 3.
- Global Value Numbering — CTGC already reduces redundancy.
- Speculative inlining — only with PGO data.
- TBAA / alias-based opts — Connection Graph already provides aliasing info.
- Auto-vectorization without annotation — explicit opt-in only.

---

## SECTION 10 — ENGINEERING MANAGEMENT PLAN

### 10.1 RFC Workflow

1. **Pre-RFC**: GitHub Discussion (temperature check, ≥ 3 reactions to proceed)
2. **Draft**: PR to `rfcs/` using template (bot auto-closes if missing any of 7 sections)
3. **Review**: 2 Core Team approvals; community 14-day window
4. **FCP**: 14-day Final Comment Period; 2+ Core Team objections = postpone
5. **Accepted**: merged; implementation behind feature flag `axc -Z <feature>`
6. **Shipped**: feature flag removed in next minor version

**RFC template sections (all required):**
1. Summary, 2. Motivation, 3. Detailed Design, 4. AI Tooling & Metadata Impact,
5. Drawbacks, 6. Alternatives, 7. Unresolved Questions

### 10.2 Coding Standards

**Go (bootstrap):**
- `gofmt` + `golangci-lint` (errcheck, staticcheck, govet)
- No `interface{}` / `any` in compiler core
- Comments: only WHY, never WHAT

**AXIOM (self-hosting):**
- `axc fmt` enforced (pre-commit hook)
- No `unsafe` in compiler core (only runtime + axalloc)
- Pure functions annotated `@[ai::assert_pure]`
- Effect annotations required for all I/O functions

### 10.3 Review Process

- Every PR: 1 approver (any team member)
- Changes to: type system, AIR spec, ABI, GRAMMAR.ebnf → 2 approvers (1 must be Core Team)
- CI must be green before merge
- No direct pushes to `main` (branch protection)

### 10.4 Branch Strategy

```
main              — always releasable, all tests green
feature/<name>    — squash-merge to main
fix/<issue>       — squash-merge to main
release/v<N>      — release stabilization
rfcs/             — RFC drafts (merged when accepted)
```

### 10.5 CI Pipeline

```yaml
jobs:
  unit-test:       # go test ./... (always)
  e2e-compliance:  # all 19 suites (always)
  lint:            # golangci-lint (always)
  format-check:    # go fmt --check (always)
  fuzz-short:      # 60s per target (on PR)
  cross-compile:   # all 4 targets (on PR)
  bench-check:     # fail if >5% regression (on main merge)
```

### 10.6 Release Strategy

- Semantic versioning: `vMAJOR.MINOR.PATCH`
- Major: breaking source changes (with `axc fix` auto-migration)
- Minor: new features, backward compatible
- Patch: bug fixes only
- Cadence: monthly minor, weekly patch
- Artifacts: pre-built binaries for Linux-x86_64, Linux-ARM64, macOS-ARM64, Windows-x86_64

### 10.7 Benchmark Tracking

- Results stored as JSON CI artifacts
- Dashboard: GitHub Pages with history charts
- Alert: > 5% regression over 10-build rolling average → auto-open `perf-regression` issue

### 10.8 Compatibility Rules (from spec 44)

- **ABI**: No stable internal ABI. Cross-version struct exchange requires `extern "C"`.
- **Source**: No breaking changes within same Major version.
- **Migration**: Every breaking change (major version bump) ships with `axc fix` auto-refactor.
- **Deprecation**: 4-stage sunset (Doc → Soft Warning → AI-assisted fix → Hard Error); minimum 12 months from Stage 1 to Stage 4.

---

## SECTION 11 — MISSING SPEC ANALYSIS

### Issue 1: Bootstrap Language Contradiction

**Problem:** Spec file 11 says Rust/Zig; files 01, 14, main spec say Go.

**Resolution:** **Go is authoritative.** Fastest path to working compiler, best AI code generation, GC eliminates manual memory issues during prototyping. Zig excluded (pre-1.0 API churn). Rust excluded (Borrow Checker friction on cyclic AST/Connection Graph).

---

### Issue 2: `let` vs `var` Keyword

**Problem:** Spec file 01 says `var`. All 19 test files and BASE_1 EBNF use `let`.

**Resolution:** `let` is canonical. `var` is not a keyword in AXIOM (historical artifact in spec 01). BASE_1 EBNF is authoritative.

---

### Issue 3: Borrow Syntax — `&x` vs `lent`

**Problem:** Tests use `&x` (borrow expression); spec uses `lent` as parameter qualifier.

**Resolution:** Both valid in different positions:
- `lent` = parameter qualifier: `fn foo(lent x: Point)`
- `&x` = borrow expression: `let view = &x`
- Both create `RefNode` with `Borrows` edge in Connection Graph.

---

### Issue 4: Quantum Extension Scope

**Problem:** Spec 01 says quantum is scrapped for MVP. Compliance tests 097–099 test `qbit`, `quantum.H()`, `quantum.measure()`. AIR spec includes `qalloc`/`qgate`/`qmeasure`.

**Resolution:** Implement `std.quantum` as a **stub module**:
- `quantum.alloc_qbit()` → returns a `qbit` (internally a `bool` with random initial value)
- `quantum.H(q)` → probabilistic coin flip
- `quantum.measure(q)` → read internal bool, "collapse" it
- AIR quantum opcodes → generate C stub calls to `std.quantum`
- Real QPU: Phase 10+ (marked [Future])

---

### Issue 5: `!T` Notation vs `sink` Keyword

**Problem:** Main spec section 3 uses `!T` type notation. Type checker spec 4.1 uses `sink` keyword.

**Resolution:** `!T` is the **canonical type-level notation** for sink parameters. `sink` as standalone keyword is dropped. Grammar: `SinkType ::= "!" Type`. At call site, type checker marks argument as moved.

---

### Issue 6: `async`/`await` in MVP

**Problem:** Spec 01 says async/await is MVP-out-of-scope. Compliance tests 061–063 test `await` and `await_all`.

**Resolution:** Async is **synchronous in MVP** (C-backend phase): `async fn` = regular `fn`, `await expr` = evaluate synchronously. State machine rewriting happens in AIR phase (Phase 5). Tests 061–063 pass because mock functions return immediately.

---

### Issue 7: M:N Scheduler Timing

**Resolution:** Clear milestone gate:
- Phases 1–7: `spawn` = `pthread_create` (1:1 OS threads). Tests pass.
- Phase 8: Full M:N scheduler. Re-run all concurrency tests. Gate: spawn 1M actors without OOM.

---

### Issue 8: `defer` Keyword Missing from Keyword List

**Problem:** `defer` mentioned in pipeline spec but not in reserved keyword list (BASE_1 section 2.4).

**Resolution:** Add `defer` to reserved keywords. Grammar: `DeferStmt ::= "defer" CallExpr NEWLINE`. C-backend: emit deferred calls before each `return` and at block end.

---

### Issue 9: `type X = A | B` Sum Type Syntax

**Problem:** `type` used for both type aliases and sum types. Match uses `i32(v)` destructuring.

**Resolution:**
```
TypeDecl ::= "type" Identifier "=" Type { "|" Type }
```
If RHS contains `|` → sum type. Pattern `i32(v)` = structural discrimination. C-backend: tagged union.

---

### Issue 10: `Locker[T]` and `lock` Block Missing from Spec

**Problem:** Compliance test 068 uses `Locker[i32]` and `lock x as y:` — not specified anywhere.

**Resolution:**
- `Locker[T]` = stdlib type: `{ value: T, mutex: PlatformMutex }`
- `lock x as y:` = acquire mutex, bind inner value as `y`, execute block, release mutex
- Grammar: `LockStmt ::= "lock" Expr "as" Identifier ":" Block`
- C-backend: `pthread_mutex_lock` → bind → body → `pthread_mutex_unlock`

---

### Issue 11: Missing `?` Error Propagation Operator

**Problem:** Test 055 comments "simulating `?`". Spec never defines `?` operator.

**Resolution:** **No `?` operator in v1.0.** Error propagation is explicit via `match`. RFC required for v1.1 with full type system implications. Tests 055 already demonstrate the pattern.

---

### Issue 12: `@SOA` vs `@[ai::suggest_layout(layout="SoA")]`

**Problem:** Two different annotations for the same concept.

**Resolution:** Distinct semantics:
- `@SOA` = compiler directive: immediately transform struct layout to SoA
- `@[ai::suggest_layout(layout="SoA")]` = AI hint: Copilot suggests, human approves
- `@SOA` is a shorthand keyword (no brackets). `@[...]` is the AI annotation namespace.

---

### Issue 13: `pub fn main()` vs `fn main()` Entrypoint

**Resolution:** Both are valid entrypoints. `pub` only affects symbol visibility for external linking.

---

### Issue 14: `and`/`or`/`not` vs `&&`/`||`/`!`

**Resolution:** AXIOM uses only `and`, `or`, `not`. `&&`/`||` are not valid. Emit helpful error: "use `and`/`or` instead of `&&`/`||`".

---

### Issue 15: `axiom.lock` SHA-256 Definition

**Resolution:** Lock file stores:
```json
{ "pkg": "name", "git": "url", "commit": "sha1", "tree_hash": "sha256_of_source_tree" }
```
`tree_hash` = SHA-256 of sorted list of `sha256(file_content)` for all `.ax` files in package root. Verified on `axc get` by recomputing tree_hash.

---

## SECTION 12 — FINAL EXECUTION GRAPH

### 12.1 Implementation Order (Dependency-Ordered)

```
Tier 0 (no dependencies):
  Grammar EBNF · Repository/CI · Struct layouts · AxHeader ABI

Tier 1 (requires 0):
  Lexer · StringPool · Runtime MVP (malloc wrapper + gen_id)

Tier 2 (requires 1):
  Parser · AST Printer

Tier 3 (requires 2):
  Symbol Table + Scope Stack · Type Table (primitives)

Tier 4 (requires 3):
  Name Resolver · Type Inference · Basic Type Checker

Tier 5 (requires 4):
  Generics + Monomorphization · Effects System · Async Typing

Tier 6 (requires 5):
  Connection Graph · Escape Analysis · Ownership Checker

Tier 7 (requires 6):
  CTGC Pass · C-Backend

Tier 8 (requires 7):
  axc build end-to-end · All 100 compliance tests passing   ◄── MVP (v0.1.0)

Tier 9 (requires 8, parallelizable):
  [A] AIR Builder
  [B] Standard Library (collections, string, math, io)
  [C] AxAlloc MVP refinement

Tier 10 (requires 9A):
  Optimization Passes · Updated C-Backend (reads from AIR)

Tier 11 (requires 10, parallelizable):
  [A] Native x86-64 backend + ELF linker
  [B] .axmeta JSON writer
  [C] std.fs, std.net, std.concurrency, std.crypto

Tier 12 (requires 11):
  ARM64 backend · RISC-V backend · PE + Mach-O linkers

Tier 13 (requires 12 + 11C):
  AxAlloc production · M:N Scheduler · Actor Runtime + Channels

Tier 14 (requires 13):
  LSP server · Package Manager · Formatter

Tier 15 (requires 14):
  Self-hosting bootstrap (Stage 1–3)                        ◄── Production (v1.0.0)
```

### 12.2 Critical Path

```
EBNF → Lexer → Parser → TypeChecker → ConnectionGraph → CTGC
→ C-Backend → MVP Complete → AIR Builder → Native Backend
→ ELF Linker → AxAlloc Prod → M:N Scheduler → Self-hosting
```

Estimated duration:
- 1 engineer: 30 months
- 3-person core team: 18 months

### 12.3 Parallelizable Tasks (After Phase 3)

| Track A (Compiler) | Track B (Runtime) | Track C (Stdlib) |
|---|---|---|
| AIR Builder | AxAlloc production | std.collections |
| Optimization passes | M:N Scheduler | std.string / std.math |
| Native x86-64 backend | Actor runtime | std.io / std.fs |
| ARM64 backend | Channels | std.net |
| RISC-V backend | Async executor | std.concurrency |
| .axmeta emitter | Distributed runtime | std.crypto |
| LSP server | NUMA awareness | std.compiler.ai |

### 12.4 Minimum Viable AXIOM Compiler (MVC) — End of Phase 4

**`axc build hello.ax` → working native executable (via C-backend + GCC)**

Supported:
- All 100 tests in `axiom_compliance_suite.ax`
- All primitives, structs, generics, sum types, interfaces
- All control flow: if/elif/else, for/in, match, break/continue, return
- Functions: closures, higher-order, recursion, effects system
- Concurrency: `spawn` (1:1 OS threads), `await` (synchronous), `Channel` (basic)
- Error handling: `Result[T,E]`, `Option[T]`, match-based propagation
- Interop: `extern "C"`, basic FFI
- Runtime: malloc-based AxAlloc, gen_id safety, panic handler
- Tooling: `axc build`, `axc check`, `axc dump-ast`

NOT in MVC: native backend, M:N scheduler, full stdlib, LSP, package manager, .axmeta, GPU/quantum

### 12.5 Production-Grade AXIOM — End of Phase 11 + Phase 13

- Self-hosted: compiler written and compiled in AXIOM
- Native backend: x86-64, ARM64 (no GCC/Clang)
- M:N work-stealing scheduler (Erlang-class actor concurrency)
- AxAlloc: production allocator (NUMA-aware, GPU-pinned, lock-free)
- Full standard library
- Complete toolchain: `axc build/check/fmt/lsp/get/fix/doc`
- Reproducible builds (triple-build verified)
- Compile speed: ≥ 500k LOC/sec
- `.axmeta` AI semantic layer
- Cross-compilation: Linux, macOS, Windows, bare-metal

### 12.6 What Should NOT Be Implemented Early

| Feature | Reason to Defer |
|---|---|
| Native backend | C-backend + GCC gives working compiler faster |
| M:N scheduler | 1:1 OS threads sufficient for all MVC tests |
| AxAlloc production | MVP malloc wrapper passes all tests |
| Auto-SoA without annotation | May break C ABI interop |
| Quantum opcodes (real QPU) | No hardware; stubs sufficient |
| GPU dispatch (real) | Phase 10+ |
| Distributed runtime | Overkill before self-hosting |
| PGO | Requires two compilation passes; defer Phase 8+ |
| LTO | Requires full program IR; Phase 6+ |
| `axc fix` AI migration | Requires LSP + stable AST API; Phase 10 |
| WASM backend | Not in spec; add via RFC after production |

### 12.7 What Should Be Stubbed Initially

| Component | Stub | Real |
|---|---|---|
| `std.quantum` | bool-based sim | Phase 10+ |
| `std.gpu` | CPU fallback | Phase 10+ |
| M:N Scheduler | `pthread_create` | Phase 8 |
| AxAlloc | `malloc` + 8-byte header | Phase 8 |
| Distributed runtime | local-only messaging | Phase 9+ |
| `axc lsp` | echo server (no-op) | Phase 10 |
| `axc get` | manual copy | Phase 10 |
| `#run` (full VM) | constant folding only | Phase 5 |

---

## MILESTONE TABLE

| Milestone | Phase | Target (3-person team) | Key Deliverable |
|---|---|---|---|
| M0: Foundation | 0 | Month 1 | Frozen structs, CI, EBNF |
| M1: Parser | 1 | Month 2 | FlatAST from any .ax file |
| M2: Type Checker | 2 | Month 4 | `axc check` reports type errors |
| M3: Ownership | 3 | Month 6 | Gen-ref safety, CTGC |
| **M4: MVC v0.1.0** | 4 | **Month 8** | **`axc build` works, 100 tests pass** |
| M5: AIR + Opts | 5 | Month 10 | Full IR pipeline, DCE/fold/inline |
| M6: Native x86-64 | 6 | Month 13 | ELF binary, no GCC needed |
| M7: Multi-arch | 7 | Month 15 | ARM64 + RISC-V |
| M8: Runtime v0.5.0 | 8 | Month 18 | AxAlloc + M:N + Actors |
| M9: Stdlib | 9 | Month 21 | std.* complete |
| M10: Tooling | 10 | Month 23 | LSP + pkg mgr + formatter |
| **M11: Self-hosting v1.0.0** | 11 | **Month 27** | **Compiler written in AXIOM** |
| M12: Production | 12 | Month 30 | Reproducible builds, 500k LOC/s |

---

## SUBSYSTEM DEPENDENCY DAG

```
Grammar ──────────► Lexer ──────────────────────► StringPool
                      │                                │
                      ▼                                │
                   Parser ◄──────────────────────────┘
                      │
                      ▼
                  FlatAST
                      │
         ┌────────────┼─────────────┐
         ▼            ▼             ▼
  NameResolver    TypeTable    AstPrinter
         │            │
         └─────┬───────┘
               ▼
         TypeChecker
               │
    ┌──────────┼──────────────────┐
    ▼          ▼                  ▼
ConnectionGraph  MonoEngine  EffectChecker
    │
 ┌──┴──┐
 ▼     ▼
Escape CTGC
          │
          ▼
     TypedAST ──────────────────────► CBackend ──► GCC/Clang
          │                                │
          ▼                                ▼
     AIRBuilder                     axc build (MVP)
          │
          ▼
        AIR
          │
 ┌────────┼─────────────────┐
 ▼        ▼                 ▼
Inline   DCE          CTGCPass (AIR)
 │        │                 │
 └────────┴─────────────────┘
                 │
                 ▼
         OptimizedAIR
                 │
    ┌────────────┼──────────────┐
    ▼            ▼              ▼
x86Backend   ARM64Bk       CBackend v2
    │
    ▼
MachineIR
    │
    ▼
LinScanRA
    │
    ▼
InMemoryLinker ──► ELF/PE/MachO
    │
    ▼
AxmetaWriter ──► .axmeta section


Runtime Track (parallel):
AxAlloc MVP ──► AxAlloc Prod ──► M:N Scheduler ──► ActorRuntime
                                                 ──► AsyncExecutor
                                                 ──► Channels

Stdlib Track (parallel after MVC):
std.collections / std.string / std.math / std.io
std.fs / std.net / std.crypto / std.concurrency / std.compiler.ai
```

---

## PRIORITIZED NEXT-ACTION CHECKLIST

### Week 1 — Day 1 Actions

- [ ] `mkdir axiom-lang && git init` — create monorepo
- [ ] `go mod init github.com/axiom-lang/axiom`
- [ ] Create full directory tree per Section 4
- [ ] Write `GRAMMAR.ebnf` — complete formal grammar covering all constructs in `axiom_compliance_suite.ax`
- [ ] Write `compiler/ast/ast.go` — `AstNode` struct (24 bytes), `NodeKind` enum, `Ast` struct
- [ ] Write test: `unsafe.Sizeof(AstNode) == 24` — must never regress
- [ ] Write `compiler/lexer/token.go` — `Token` struct (8 bytes), `TokenKind` enum
- [ ] Write test: `unsafe.Sizeof(Token) == 8`
- [ ] Write `ir/air/inst.go` — `AirInst` struct (16 bytes), `Opcode` enum
- [ ] Write test: `unsafe.Sizeof(AirInst) == 16`
- [ ] Setup GitHub Actions: build + test on Linux/macOS/Windows
- [ ] Commit `docs/design/ast-schema.md`, `docs/design/token-kinds.md`, `docs/design/air-schema.md`

### Week 2–4 — Lexer

- [ ] Implement DFA scanner in `compiler/lexer/lexer.go`
- [ ] Implement INDENT/DEDENT emission with indentation stack
- [ ] 30+ lexer unit tests: all keywords, operators, literals, INDENT/DEDENT balance, UTF-8, error tokens
- [ ] Write fuzz target `fuzz/fuzz_lex.go`
- [ ] Parse all 19 `.ax` test files: zero panics, zero out-of-bounds

### Week 5–8 — Parser

- [ ] Implement Recursive Descent parser for all statement forms
- [ ] Implement Pratt expression parser with correct precedence table
- [ ] Implement error recovery (panic mode + synchronization)
- [ ] Implement `axc dump-ast` command
- [ ] Golden test corpus: 50+ `.ax` snippets with expected AST JSON
- [ ] Parse `axiom_compliance_suite.ax`: 0 ErrorNodes

### Week 9–16 — Type Checker

- [ ] Implement `StringPool`, `SymbolTable`, `ScopeStack`
- [ ] Implement `TypeTable` with all primitives
- [ ] Implement `NameResolver` with Lazy Field Analysis
- [ ] Implement type inference (local HM, bidirectional for `let`)
- [ ] Implement overload resolution scoring
- [ ] Implement `MonoEngine` (generic instantiation)
- [ ] Implement effects system propagation
- [ ] `axc check axiom_compliance_suite.ax` groups 1–5: exit 0

### Week 17–24 — Ownership + CTGC

- [ ] Implement `ConnectionGraph` (nodes, edges, EscapesFunction, IsIsolated)
- [ ] Implement `OwnershipChecker` (mut/lent/sink/Isolated rules)
- [ ] Implement escape analysis on Connection Graph
- [ ] Implement CTGC: `=destroy` injection, `alias` insertion
- [ ] Implement `runtime/axalloc/axalloc.c` MVP (malloc + 8-byte AxHeader)
- [ ] Implement `runtime/panic/panic.c`
- [ ] Runtime tests: gen_id mismatch → panic (not segfault)

### Week 25–32 — C-Backend + MVP Complete

- [ ] Implement `codegen/cgen/cgen.go`: all AST/AIR → C11 mappings
- [ ] Implement `axc build` end-to-end pipeline
- [ ] All 100 compliance tests: compile and run, exit 0
- [ ] Benchmark baseline: commit compile time for Hello World
- [ ] **Tag v0.0.1**

### Ongoing Throughout

- [ ] Every new subsystem: add fuzz target
- [ ] Every bug: add regression test before fix
- [ ] Every week: run benchmark suite, track in CI artifact
- [ ] Every month: review fuzz corpus findings, triage ICE reports
- [ ] Every phase: update `docs/changelog/` with decisions made

---

*This plan is implementation-ready. A compiler engineer can pick up any phase, read the inputs/outputs/APIs, and begin coding without further architecture meetings. The critical path is clear, risks are identified with mitigations, and the self-hosting strategy has concrete verification gates. Total estimated timeline to production-grade self-hosting: 27–30 months for a 3-person core team.*
