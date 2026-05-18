# p04-t08: Overload Resolution

## Purpose
Implement overload resolution for function calls with multiple candidates — primarily operator overloading (built-in `+` on strings, integers, floats) and method dispatch on interfaces. The resolver selects the best matching overload using a scoring system and reports ambiguity errors when no unique best match exists.

## Context
AXIOM's overload resolution uses a four-level scoring system to rank candidates: exact type match (4), type that is coercible to the expected type (3), generic type parameter match (2), interface satisfaction match (1). The highest-scoring unique candidate wins. Ties are ambiguity errors. This system is simple enough for fast compilation but expressive enough to handle the most common cases.

## Inputs
- Function name (interned NameID)
- Argument TypeIDs (from type-checked expressions)
- SymbolTable — all functions with the given name
- TypeTable — for coercion and interface checking

## Outputs
- `ResolvedCall{SymbolID uint32, Score int}` — the winning overload
- `[]Diagnostic` — ambiguity errors or "no matching overload" errors

## Dependencies
- p04-t07: type-checker-expressions — calls overload resolver for each CallExpr
- p04-t02: type-table-primitives — coercion rules use TypeTable
- p04-t01: symbol-table — candidate lookup

## Subsystems Affected
- Type checker: expression type checking delegates to overload resolver
- Operator semantics: built-in operators implemented as pseudo-overloads

## Detailed Requirements

1. `OverloadResolver` struct: `st *SymbolTable, tt *TypeTable`
2. `Resolve(name uint32, argTypes []uint32) (ResolvedCall, error)`:
   - Collect all symbols with `name` across all scopes
   - Score each candidate against the argument types
   - Return candidate with highest total score; error on tie or no match
3. Scoring per argument position:
   - 4: exact TypeID match
   - 3: arg type is coercible (e.g., i32 literal to i64 parameter)
   - 2: matches generic type parameter (T in `fn foo[T](x: T)`)
   - 1: arg type implements required interface
   - 0: no match (candidate is eliminated)
4. Built-in operator pseudo-overloads (hardcoded, not in symbol table):
   - `+`: `(i32,i32)→i32`, `(i64,i64)→i64`, `(f32,f32)→f32`, `(f64,f64)→f64`, `(string,string)→string`
   - `==`,`!=`: any type → bool (all types are comparable by default)
   - Arithmetic: all numeric combinations
5. Method dispatch: for `obj.method(args)`, look up `method` in the type of `obj`'s struct fields, then in interface implementations.
6. `isCoercible(from, to uint32) bool`: integer literal → any numeric, smaller int → larger int (explicit only for non-literal), nil → any pointer/optional.

## Implementation Steps

1. Create `compiler/sema/overload.go`.
2. Implement built-in operator table as `map[string][]OverloadEntry`.
3. Implement `collectCandidates(name uint32) []SymbolID` — searches all scopes.
4. Implement `scoreCandidate(sym Symbol, argTypes []uint32) int` — sum of per-argument scores.
5. Implement `Resolve()`: collect → score → select max → check uniqueness.
6. Implement method dispatch: `ResolveMethod(receiverTypeID uint32, methodName uint32, argTypes []uint32)`.
7. Write tests: exact match, coercible match, ambiguity error, no match error.

## Test Plan

- `TestOverloadExact`: `fn foo(x: i32)` called with i32 arg → exact match, score=4
- `TestOverloadCoercible`: `fn foo(x: i64)` called with i32 literal → coercible, score=3
- `TestOverloadAmbiguous`: two functions both score 3 → ambiguity error
- `TestOverloadNoMatch`: `fn foo(x: string)` called with i32 → "no matching overload"
- `TestBuiltinPlus`: `1 + 2` with TypeI32 args → resolves to int addition
- `TestBuiltinStringPlus`: `"a" + "b"` with TypeString args → string concatenation

## Validation Checklist

- [ ] Exact match always preferred over coercible
- [ ] Ambiguity error when two candidates have equal score
- [ ] Built-in operators resolve correctly for all numeric types
- [ ] Method dispatch works on struct receiver types
- [ ] Error messages name the conflicting overloads

## Acceptance Criteria

- `println("hello")` resolves to `std.io.println(string)` without ambiguity
- `1 + 2` resolves to integer addition
- `"a" + "b"` resolves to string concatenation

## Definition of Done

- [ ] `compiler/sema/overload.go` implemented
- [ ] Built-in operator table populated
- [ ] `go test ./compiler/sema/ -run TestOverload` passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Overload resolution is too slow for large candidate sets | Limit search to current scope chain; stop at first exact match |
| Implicit coercions causing unexpected behavior | Only coerce literals; never implicitly coerce named variables |

## Future Follow-up Tasks

- p05-t04: structural duck typing uses interface-based overload scoring
- p05-t02: generic instantiation adds generic candidates to the overload set
