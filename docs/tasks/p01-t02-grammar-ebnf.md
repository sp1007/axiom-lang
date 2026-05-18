# p01-t02: Grammar EBNF

## Purpose
Write the complete, authoritative, formal EBNF grammar for the AXIOM programming language to `docs/GRAMMAR.ebnf`. This document is the single source of truth for every syntactic decision in the compiler. The parser (Phase 03) must implement exactly this grammar — no more, no less. Having a formal grammar before writing the parser prevents ambiguity, guides test case design, and enables future tooling (syntax highlighters, language servers, documentation generators) to work from a specification rather than reverse-engineering the implementation.

## Context
AXIOM is an indentation-based language (like Python) using exactly 4-space indentation — no tabs allowed. It is statically typed with type inference, supports algebraic data types (sum types), interfaces (structural/duck typing), async/await, ownership annotations (`lent`, `!T`/sink), and actor-style concurrency (`spawn`/`await`). The grammar must be LL(1) compatible — no left recursion, no ambiguity that requires lookahead beyond 1 token except where explicitly documented. This grammar will be validated against real AXIOM programs in the test suite. All names use snake_case for variables/functions and PascalCase for types/structs. Keywords are reserved and cannot be used as identifiers.

## Inputs
- `AXIOM LANGUAGE SPECIFICATION v1.0.md` (primary specification)
- `01.minimal core.md`
- `03. Thiết kế parser thực tế.md`
- Existing `.ax` example programs in `tests/` and `examples/`

## Outputs
- `docs/GRAMMAR.ebnf` — complete formal EBNF grammar, ~300-500 lines

## Dependencies
- p01-t01: repository-bootstrap — `docs/` directory must exist

## Subsystems Affected
- `compiler/parser/`: Parser implementation must conform exactly to this grammar
- `compiler/lexer/`: Token kinds must cover all terminals referenced in this grammar
- `docs/`: Grammar is part of the language specification documentation

## Detailed Requirements

1. **EBNF notation conventions** used in the file:
   ```
   (* Comment *)
   Rule = Alternative1 | Alternative2 ;
   Rule = TermA , TermB ;          (* sequence *)
   Rule = [ Optional ] ;
   Rule = { Repeated } ;           (* zero or more *)
   Rule = TermA - TermB ;          (* except *)
   'keyword'                        (* literal terminal *)
   UPPER_CASE                       (* lexer token reference *)
   ```

2. **Top-level program structure**:
   ```ebnf
   Program = { ImportDecl } , { TopLevelDecl } , EOF ;
   TopLevelDecl = FnDecl | StructDecl | InterfaceDecl | TypeAliasDecl | ConstDecl ;
   ```

3. **Import declarations**:
   ```ebnf
   ImportDecl = 'import' , ModulePath , [ '{' , ImportList , '}' ] , NEWLINE ;
   ModulePath = IDENT , { '.' , IDENT } ;
   ImportList = IDENT , { ',' , IDENT } ;
   ```

4. **Function declarations** (must handle pub, async, extern, generic params):
   ```ebnf
   FnDecl = [ 'pub' ] , [ 'async' ] , [ 'extern' ] ,
            'fn' , IDENT , [ GenericParams ] ,
            '(' , [ ParamList ] , ')' ,
            [ '->' , TypeExpr , [ EffectAnnotation ] ] ,
            ':' , Block ;

   GenericParams = '[' , GenericParam , { ',' , GenericParam } , ']' ;
   GenericParam  = IDENT , [ ':' , TypeExpr ] ;

   ParamList = Param , { ',' , Param } ;
   Param = [ ParamModifier ] , IDENT , ':' , TypeExpr ;
   ParamModifier = 'mut' | 'lent' | '!' ;

   EffectAnnotation = '{' , '.' , IDENT , ':' , '[' , TypeList , ']' , '.' , '}' ;
   ```

5. **Struct and interface declarations**:
   ```ebnf
   StructDecl    = [ 'pub' ] , [ 'packed' ] , 'struct' , IDENT ,
                   [ GenericParams ] , ':' , INDENT ,
                   { FieldDecl } , { FnDecl } , DEDENT ;

   FieldDecl     = [ 'pub' ] , [ 'mut' ] , IDENT , ':' , TypeExpr , NEWLINE ;

   InterfaceDecl = [ 'pub' ] , 'interface' , IDENT ,
                   [ GenericParams ] , ':' , INDENT ,
                   { MethodSig } , DEDENT ;

   MethodSig     = [ 'async' ] , 'fn' , IDENT ,
                   [ GenericParams ] , '(' , [ ParamList ] , ')' ,
                   [ '->' , TypeExpr ] , NEWLINE ;
   ```

6. **Type alias and sum types**:
   ```ebnf
   TypeAliasDecl = [ 'pub' ] , 'type' , IDENT , [ GenericParams ] ,
                   '=' , SumTypeExpr , NEWLINE ;

   SumTypeExpr   = TypeVariant , { '|' , TypeVariant } ;
   TypeVariant   = IDENT , [ '(' , TypeList , ')' ] ;
   TypeList      = TypeExpr , { ',' , TypeExpr } ;
   ```

7. **Type expressions** (full type grammar):
   ```ebnf
   TypeExpr = PtrTypeExpr
            | SliceTypeExpr
            | ArrayTypeExpr
            | FuncTypeExpr
            | GenericTypeExpr
            | IsolatedTypeExpr
            | FutureTypeExpr
            | IDENT ;

   PtrTypeExpr      = '*' , [ 'mut' ] , TypeExpr ;
   SliceTypeExpr    = '[' , TypeExpr , ']' ;
   ArrayTypeExpr    = '[' , TypeExpr , ';' , INT_LIT , ']' ;
   FuncTypeExpr     = 'fn' , '(' , [ TypeList ] , ')' , [ '->' , TypeExpr ] ;
   GenericTypeExpr  = IDENT , '[' , TypeList , ']' ;
   IsolatedTypeExpr = 'Isolated' , '[' , TypeExpr , ']' ;
   FutureTypeExpr   = 'Future' , '[' , TypeExpr , ']' ;
   ```

8. **Statements**:
   ```ebnf
   Block = INDENT , { Stmt } , DEDENT ;

   Stmt = VarDecl
        | AssignStmt
        | ReturnStmt
        | IfStmt
        | ForStmt
        | WhileStmt
        | MatchStmt
        | DeferStmt
        | UnsafeBlock
        | ArenaBlock
        | ExprStmt ;

   VarDecl    = ( 'let' | 'mut' ) , IDENT , [ ':' , TypeExpr ] ,
                '=' , Expr , NEWLINE ;
   AssignStmt = Expr , AssignOp , Expr , NEWLINE ;
   AssignOp   = '=' | '+=' | '-=' | '*=' | '/=' | '%=' ;
   ReturnStmt = 'return' , [ Expr ] , NEWLINE ;
   ExprStmt   = Expr , NEWLINE ;

   IfStmt   = 'if' , Expr , ':' , Block ,
              { 'elif' , Expr , ':' , Block } ,
              [ 'else' , ':' , Block ] ;

   ForStmt  = 'for' , IDENT , 'in' , Expr , ':' , Block ;
   WhileStmt = 'while' , Expr , ':' , Block ;

   DeferStmt  = 'defer' , Expr , NEWLINE ;
   UnsafeBlock = 'unsafe' , ':' , Block ;
   ArenaBlock  = 'in' , '[' , IDENT , ']' , ':' , Block ;
   ```

9. **Match statement** with exhaustiveness requirement:
   ```ebnf
   MatchStmt  = 'match' , Expr , ':' , INDENT ,
                { MatchArm } , DEDENT ;
   MatchArm   = Pattern , ':' , ( Block | ( Expr , NEWLINE ) ) ;
   Pattern    = WildcardPat | LiteralPat | BindingPat | VariantPat | TuplePat ;
   WildcardPat = '_' ;
   LiteralPat  = INT_LIT | FLOAT_LIT | STRING_LIT | 'true' | 'false' | 'nil' ;
   BindingPat  = IDENT ;
   VariantPat  = IDENT , [ '(' , PatternList , ')' ] ;
   TuplePat    = '(' , PatternList , ')' ;
   PatternList = Pattern , { ',' , Pattern } ;
   ```

10. **Expressions** — document the Pratt precedence table inline as EBNF comments:
    ```ebnf
    Expr = OrExpr ;
    OrExpr   = AndExpr , { 'or' , AndExpr } ;
    AndExpr  = NotExpr , { 'and' , NotExpr } ;
    NotExpr  = [ 'not' ] , CmpExpr ;
    CmpExpr  = BitOrExpr , { CmpOp , BitOrExpr } ;
    CmpOp    = '==' | '!=' | '<' | '>' | '<=' | '>=' ;
    BitOrExpr  = BitXorExpr , { '|' , BitXorExpr } ;
    BitXorExpr = BitAndExpr , { '^' , BitAndExpr } ;
    BitAndExpr = ShiftExpr , { '&' , ShiftExpr } ;
    ShiftExpr  = AddExpr , { ( '<<' | '>>' ) , AddExpr } ;
    AddExpr    = MulExpr , { ( '+' | '-' ) , MulExpr } ;
    MulExpr    = PowerExpr , { ( '*' | '/' | '%' ) , PowerExpr } ;
    PowerExpr  = UnaryExpr , [ '**' , PowerExpr ] ;  (* right-assoc *)
    UnaryExpr  = ( '-' | '~' | 'not' ) , UnaryExpr | PostfixExpr ;
    PostfixExpr = PrimaryExpr , { PostfixOp } ;
    PostfixOp  = '.' , IDENT
               | '[' , Expr , ']'
               | '(' , [ ArgList ] , ')'
               | '.*'
               | 'as' , TypeExpr ;
    ```

11. **Primary expressions** (atoms):
    ```ebnf
    PrimaryExpr = IDENT
                | INT_LIT | FLOAT_LIT | STRING_LIT | CHAR_LIT
                | 'true' | 'false' | 'nil'
                | '(' , Expr , ')'
                | ArrayLit
                | StructLit
                | SpawnExpr
                | AwaitExpr
                | ClosureExpr ;

    ArrayLit   = '[' , [ ExprList ] , ']' ;
    StructLit  = IDENT , '{' , [ FieldInitList ] , '}' ;
    FieldInitList = FieldInit , { ',' , FieldInit } ;
    FieldInit  = IDENT , ':' , Expr ;
    SpawnExpr  = 'spawn' , Expr ;
    AwaitExpr  = 'await' , Expr ;
    ClosureExpr = '|' , [ ParamList ] , '|' , ( Expr | Block ) ;

    ExprList   = Expr , { ',' , Expr } ;
    ArgList    = Arg , { ',' , Arg } ;
    Arg        = [ IDENT , ':' ] , Expr ;  (* named args supported *)
    ```

12. **Lexical terminals** — document all token kinds referenced:
    ```ebnf
    IDENT      = LETTER , { LETTER | DIGIT | '_' } - KEYWORD ;
    INT_LIT    = DecInt | HexInt | OctInt | BinInt ;
    DecInt     = DIGIT , { DIGIT | '_' } ;
    HexInt     = '0x' , HEXDIGIT , { HEXDIGIT | '_' } ;
    OctInt     = '0o' , OCTDIGIT , { OCTDIGIT | '_' } ;
    BinInt     = '0b' , ( '0' | '1' ) , { '0' | '1' | '_' } ;
    FLOAT_LIT  = DIGIT , { DIGIT } , '.' , DIGIT , { DIGIT } ,
                 [ ( 'e' | 'E' ) , [ '+' | '-' ] , DIGIT , { DIGIT } ] ;
    STRING_LIT = '"' , { StrChar } , '"' ;
    CHAR_LIT   = "'" , StrChar , "'" ;
    StrChar    = ? any UTF-8 char except '"' and '\' ?
               | '\n' | '\t' | '\\' | '\"' | "\u{" , HEXDIGIT , { HEXDIGIT } , '}' ;
    KEYWORD    = 'fn' | 'let' | 'mut' | 'struct' | 'interface' | 'import'
               | 'pub' | 'return' | 'if' | 'elif' | 'else' | 'for' | 'while'
               | 'in' | 'match' | 'type' | 'spawn' | 'await' | 'defer'
               | 'unsafe' | 'and' | 'or' | 'not' | 'true' | 'false' | 'nil'
               | 'async' | 'extern' | 'packed' | 'lent' | 'Isolated' | 'Future' ;
    INDENT  = (* increase in indentation level by exactly 4 spaces *) ;
    DEDENT  = (* decrease in indentation level by exactly 4 spaces *) ;
    NEWLINE = (* line feed, possibly preceded by carriage return *) ;
    EOF     = (* end of input *) ;
    ```

13. **Const declarations** (for completeness):
    ```ebnf
    ConstDecl = [ 'pub' ] , 'const' , IDENT , ':' , TypeExpr , '=' , Expr , NEWLINE ;
    ```

14. **LL(1) compliance note**: Document any rules that require more than 1 token of lookahead and how the parser resolves them. Key case: distinguishing `AssignStmt` from `ExprStmt` requires seeing if an `AssignOp` follows the LHS expression. The parser speculatively parses an expression, then checks if the next token is an `AssignOp`.

15. **Naming conventions** must be documented in the grammar file header:
    - Variables and functions: `snake_case`
    - Types, structs, interfaces: `PascalCase`
    - Constants: `SCREAMING_SNAKE_CASE`
    - Module paths: `lower.dotted.path`

## Implementation Steps

1. Create `docs/GRAMMAR.ebnf` with a file header comment block:
   ```
   (* AXIOM Language Grammar — Formal EBNF Definition
      Version: 0.1.0
      Date: <date>
      This grammar is the authoritative specification for AXIOM syntax.
      The parser in compiler/parser/ must implement this grammar exactly.
      Notation: ISO/IEC 14977 EBNF with minor extensions noted in comments.
   *)
   ```

2. Write the grammar sections in this order, with section header comments:
   - `(* === PROGRAM STRUCTURE === *)`
   - `(* === DECLARATIONS === *)`
   - `(* === TYPE EXPRESSIONS === *)`
   - `(* === STATEMENTS === *)`
   - `(* === EXPRESSIONS === *)`
   - `(* === PATTERNS === *)`
   - `(* === LEXICAL TERMINALS === *)`

3. For each production rule, add a comment if there is any non-obvious design decision (e.g., why PowerExpr is right-associative).

4. Verify no left recursion: walk every rule and check that no rule's first alternative begins with the rule name itself. PostfixExpr uses iteration (`{ PostfixOp }`) to avoid left recursion.

5. Mark all keywords explicitly in the KEYWORD terminal so the lexer can use this list for keyword recognition.

6. Add a section `(* === DISAMBIGUATION NOTES === *)` at the end documenting:
   - AssignStmt vs ExprStmt: parse as expr, check next token
   - StructLit vs Block: `IDENT '{'` is always StructLit in expression context
   - GenericTypeExpr vs comparison: `IDENT '[' TypeExpr ']'` only when type context is expected

7. Validate the grammar against at least 5 example programs mentally (hello world, fibonacci, sum type, generic sort, async function) by tracing derivations.

8. Cross-reference the grammar against `compiler/lexer/token.go` (established in p02-t01) to confirm every terminal in the grammar maps to a TokenKind.

## Test Plan

The grammar file itself is not directly testable as Go code, but it must be validated:

- **Grammar validation tool** (optional, future): Write `tools/grammarcheck/main.go` that parses the EBNF file and checks for undefined non-terminals, left recursion, and unreachable rules.
- **Parser conformance**: When the parser is implemented (p03-t04 through p03-t07), every construct in the grammar must have at least one test in `tests/parser/`. The grammar serves as the checklist.
- **Manual trace test**: For each of the following programs, manually derive the parse tree using the grammar and verify it matches what the parser produces:
  1. `tests/parser/hello_world.ax`
  2. `tests/parser/fibonacci.ax`
  3. `tests/parser/sum_type.ax`
  4. `tests/parser/generic_sort.ax`
  5. `tests/parser/async_fn.ax`
- **Keyword exhaustiveness**: Write a test `TestKeywordCoverage` in `compiler/lexer/lexer_test.go` (p02-t02) that verifies every keyword listed in the EBNF KEYWORD terminal has a corresponding `TokenKind` constant.

## Validation Checklist
- [ ] `docs/GRAMMAR.ebnf` exists and is non-empty (target: 300+ lines)
- [ ] All keywords listed: fn, let, mut, struct, interface, import, pub, return, if, elif, else, for, while, in, match, type, spawn, await, defer, unsafe, and, or, not, true, false, nil, async, extern, packed, lent, Isolated, Future
- [ ] No left recursion in any production rule (manually verified)
- [ ] INDENT/DEDENT/NEWLINE/EOF terminals are defined and documented
- [ ] Type expressions cover: primitives, ptr, slice, array, func, generic, Isolated, Future
- [ ] Sum type syntax `type X = A | B` is present
- [ ] Effect annotation syntax `{.raises: [T].}` is present
- [ ] Arena block syntax `in [arena]:` is present
- [ ] All operator precedence levels (10 through 130) are documented
- [ ] LL(1) disambiguation notes are present
- [ ] Naming convention rules are documented in the header
- [ ] Grammar cross-referenced against at least 5 example programs

## Acceptance Criteria
- `docs/GRAMMAR.ebnf` is syntactically valid EBNF (parseable by a standard EBNF checker)
- Every AXIOM keyword has a grammar rule that introduces it
- No production rule has left recursion
- The grammar can derive all 5 required example programs
- The grammar precisely rejects the following invalid programs (document them in the file as `(* INVALID: ... *)` comments):
  - Tab-indented code
  - 2-space indentation
  - Using a keyword as an identifier
  - Missing `:` after function signature

## Definition of Done
- [ ] `docs/GRAMMAR.ebnf` committed to repository
- [ ] All required keywords present
- [ ] No left recursion
- [ ] Disambiguation notes written
- [ ] Example program derivations manually verified
- [ ] Grammar reviewed against AXIOM LANGUAGE SPECIFICATION v1.0.md
- [ ] Cross-referenced with p02-t01 TokenKind list for terminal coverage
- [ ] Approved by project lead / second engineer review

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Grammar ambiguity discovered when implementing parser | Add disambiguation notes immediately to GRAMMAR.ebnf; update grammar before updating parser |
| Indentation-based grammar difficult to express in EBNF | Use explicit INDENT/DEDENT terminal tokens as lexer-emitted tokens; document this in grammar header |
| Operator precedence mistakes | Cross-reference against spec document; write precedence table explicitly in comments |
| Missing syntax for edge cases (e.g., multiline expressions) | Add continuation line rule: expression continues if next line is indented more than statement start |
| Keywords added later break existing programs | Grammar is versioned; breaking keyword additions require RFC (p01-t05) |

## Future Follow-up Tasks
- p02-t01: TokenKind enum must include every terminal defined in this grammar
- p02-t02: Lexer must recognize every keyword listed in KEYWORD terminal
- p03-t04: Parser statements must implement every Stmt production
- p03-t05: Pratt parser must implement every expression production with correct precedence
- p03-t08: Parser golden tests use this grammar as the checklist for test coverage
