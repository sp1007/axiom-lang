# plan_stage4.md — Axiom Compiler Stage 1 Expansion: Advanced Features

This document details the engineering execution plan for expanding the self-hosted AXIOM compiler (Stage 1) to natively support **struct declarations, type aliases, sum types (variants), and match expressions**.

---

## Task 1: Extend Self-Hosted Parser (`bootstrap/stage1/parser.ax`)
### Objectives
- Add parser methods for parsing top-level structs and type declarations (alias and sum types).
- Add parser methods for parsing match statements, match arms, and match patterns.
- Keep `AstNode` exactly aligned to the 24-byte layout.

### Input
- `bootstrap/stage1/parser.ax`
- Specification of struct and sum type parsing in `compiler/parser/parser.go`

### Output
- Upgraded `parser.ax` with new top-level parsing rules and statement parsing.
- Ability to parse:
  - `struct` declarations with fields and functions (methods).
  - `type` declarations containing sum type variants (e.g., `type Option[T] = Some(T) | None`).
  - `match` statements with patterns and arms.

### Implementation Steps
1. **Top-Level Keywords**: Update `parse_program` to recognize `TK_STRUCT` and `TK_TYPE` as top-level declarations (including when prefixed with `TK_PUB`).
2. **Struct Parsing**:
   - `parse_struct_decl(is_pub: bool) -> u32`: Parse struct identifier, optional generic parameters, colon `:`, and an indented block of fields or methods.
   - `parse_field_decl(is_pub: bool) -> u32`: Parse optionally mutable struct fields (`mut name: type_expr`).
3. **Type Alias & Sum Type Parsing**:
   - `parse_type_alias_decl(is_pub: bool) -> u32`: Parse type identifier, optional generic parameters, `=` sign, and variants separated by `|` (NodeSumType containing NodeVariantDecl children).
   - `parse_type_variant() -> u32`: Parse variant name and optional parenthesized payload type expressions.
4. **Match Statement Parsing**:
   - Update `parse_stmt` to recognize `TK_MATCH` and delegate to `parse_match_stmt`.
   - `parse_match_stmt() -> u32`: Parse match expression (scrutinee), colon `:`, and an indented block of match arms.
   - `parse_match_arm() -> u32`: Parse pattern, colon `:`, followed by an expression or block body.
   - `parse_pattern() -> u32`: Parse wildcards (`_`), variant patterns (`Some(v)`), binding patterns (`v`), literal patterns (`42`), or tuple patterns.

---

## Task 2: Extend Self-Hosted Name Resolver (`bootstrap/stage1/resolver.ax`)
### Objectives
- Handle scoped symbol resolution for match pattern bindings, struct types, and sum type variant constructors.
- Elevate sum type variant symbols to the parent scope of the type alias so they are globally/locally constructible.

### Input
- `bootstrap/stage1/resolver.ax`
- Symbol resolver specification in `compiler/sema/resolver.go`

### Output
- Upgraded `resolver.ax` correctly resolving variants and match patterns.

### Implementation Steps
1. **Type Alias Variant Elevation**:
   - Modify `NODE_TYPE_ALIAS_DECL` in `resolve_node`. After resolving children in the pushed block scope, look at all entries in the current scope.
   - For every entry, if its symbol's kind is `SYM_VARIANT`, put it in the parent scope. This exposes constructor variants (e.g., `Ok`, `Err`) to the outer scope.
2. **Match Scoping**:
   - For `NODE_MATCH_ARM`, push a new block scope before resolving its children, allowing pattern bindings (e.g., `v` in `Some(v)`) to exist exclusively inside that arm's body.
3. **Pattern Resolution**:
   - For `NODE_BINDING_PAT`, define the variable in the current arm scope.
   - For `NODE_VARIANT_PAT`, resolve the variant constructor identifier.

---

## Task 3: Extend Self-Hosted Type Checker (`bootstrap/stage1/typecheck.ax`)
### Objectives
- Register and validate `TypeTable` entries for struct fields and sum types.
- Infer and check types for match expressions, validating exhaustiveness and typing of pattern bindings.

### Input
- `bootstrap/stage1/typecheck.ax`
- Type checker specification in `compiler/sema/check_stmt.go` and `compiler/types/typetable.go`

### Output
- Upgraded `typecheck.ax` correctly validating structs, sum types, and match expressions.

### Implementation Steps
1. **Sum Type Registration**:
   - Add `TYPE_KIND_SUM` constant to `typecheck.ax`.
   - Add sum type structures to `TypeTable` (e.g. `SumType` structures with variants containing name, payload type, and tag).
   - Implement `register_sum_type(name_id: u32, variants: VariantVec) -> u32`.
2. **Type Checking for `NodeTypeAliasDecl`**:
   - Walk the sum type children, resolve their payload types, and register the sum type in the type table.
   - Assign the registered `TypeID` to the type alias symbol and to each child variant symbol.
3. **Type Checking for `NodeMatchStmt` & `NodeMatchArm`**:
   - Infer the scrutinee type and track it.
   - For each arm, type check the pattern and the body.
   - For variant patterns `Variant(arg)`, find the corresponding variant in the sum type definition, get its payload type, and assign it to the binding argument.
   - Ensure the return type of all arms are consistent and compatible.

---

## Task 4: Verification and Triple-Build Validation
### Objectives
- Run a full end-to-end self-hosting triple-build loop to verify correctness.
- Ensure that the extended compiler compiles itself cleanly and produces byte-for-byte identical binaries.

### Input
- Modified Stage 1 compiler source files.
- `scripts/triple_build.ps1`

### Output
- Green triple-build confirmation.
- Safe executable binaries (`axc.exe` and `axc_stage1.exe`).

### Implementation Steps
1. Concatenate modified compiler files.
2. Build Stage 0, and use Stage 0 to compile Stage 1.
3. Verify Stage 1 compiles successfully and passes all corpus tests.
