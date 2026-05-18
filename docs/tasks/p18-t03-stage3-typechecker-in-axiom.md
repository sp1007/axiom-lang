# p18-t03: Stage 3 — Type Checker in AXIOM

## Purpose
Implement the AXIOM type checker in AXIOM (Stage 3), including symbol resolution, type inference, and semantic analysis — the most complex component of self-hosting.

## Context
Stage 3 is the hardest self-hosting step: the type checker must type-check AXIOM code including generics, sum types, and ownership rules. When Stage 3 is complete, AXIOM can type-check its own frontend. The type checker in AXIOM is still compiled by the Go bootstrap compiler at this stage.

## Inputs
- `bootstrap/stage2/parser.ax` — AST from Stage 2
- Type system design from p04 and p05

## Outputs
- `bootstrap/stage3/typetable.ax` — TypeInfo in AXIOM
- `bootstrap/stage3/resolver.ax` — name resolution in AXIOM
- `bootstrap/stage3/checker.ax` — type checker in AXIOM

## Dependencies
- p18-t02: stage2-parser-in-axiom — AST input
- p04: all type checker passes — algorithms to port

## Detailed Requirements

```axiom
# bootstrap/stage3/typetable.ax
type TypeKind: u8
const TK_PRIMITIVE: TypeKind = 0
const TK_STRUCT:    TypeKind = 1
const TK_FUNC:      TypeKind = 2
const TK_GENERIC:   TypeKind = 3

type TypeInfo:
    var kind:    TypeKind
    var size:    u32
    var align:   u8
    var name:    str
    var fields:  Array[FieldInfo]  # for structs
    var params:  Array[u32]        # for funcs: param TypeIDs
    var ret:     u32               # for funcs: return TypeID

type TypeTable:
    var types:   Array[TypeInfo]
    var by_name: HashMap[str, u32]

    fn register(mut self, info: TypeInfo) -> u32
    fn lookup(self, name: str) -> Option[u32]
    fn get(self, id: u32) -> TypeInfo

# bootstrap/stage3/checker.ax
type TypeChecker:
    var ast:    Array[AstNode]
    var table:  TypeTable
    var scope:  ScopeStack
    var errors: Array[TypeError]

    fn new(ast: Array[AstNode], table: TypeTable) -> TypeChecker
    fn check_module(mut self)
    fn check_func(mut self, node: u32)
    fn check_stmt(mut self, node: u32)
    fn infer_expr(mut self, node: u32) -> u32  # returns TypeID
    fn unify(mut self, a: u32, b: u32) -> bool  # HM unification
```

Validation: compile the same AXIOM file with Go type checker and AXIOM type checker; compare TypeID assignments and error messages.

## Implementation Steps

1. Create `bootstrap/stage3/typetable.ax` — TypeInfo, TypeTable.
2. Create `bootstrap/stage3/resolver.ax` — ScopeStack, name resolution.
3. Create `bootstrap/stage3/checker.ax` — type checking passes.
4. Port HM type inference (union-find) to AXIOM.
5. Port overload resolution to AXIOM.
6. Run on stdlib corpus; compare against Go type checker output.

## Test Plan
- `TestStage3CheckerCorpus`: all stdlib files type-check identically to Go checker
- `TestStage3GenericInstantiation`: generic function type-checks correctly
- `TestStage3TypeError`: type error produces same diagnostic as Go checker

## Validation Checklist
- [ ] TypeID assignments match Go checker for all test files
- [ ] Error messages match (or are structurally equivalent)
- [ ] Generic monomorphization produces same type IDs
- [ ] Union-find consistent after unification

## Acceptance Criteria
- AXIOM type checker in AXIOM checks the stdlib without errors

## Definition of Done
- [ ] `bootstrap/stage3/checker.ax` implemented
- [ ] Corpus comparison passes on stdlib

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| HM unification in AXIOM stack-overflows on recursive types | Use explicit work stack instead of recursion |

## Future Follow-up Tasks
- p18-t04: Stage 4 — AIR builder in AXIOM (codegen frontend)
