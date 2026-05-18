# p17-t09: REPL (Interactive Interpreter)

## Purpose
Implement an interactive AXIOM REPL (Read-Eval-Print Loop) that compiles and executes expressions and statements incrementally, enabling rapid exploration and prototyping.

## Context
A REPL is invaluable for learning and prototyping. The AXIOM REPL compiles each line/expression through the full pipeline (lex → parse → typecheck → codegen → execute), accumulating declared variables and functions across prompts. It uses the JIT-like approach of compiling small fragments to machine code and executing via dlopen/dlsym.

## Inputs
- AXIOM source lines from stdin
- Previous session state (accumulated bindings)
- C backend for quick compilation (not native backend, for simplicity)

## Outputs
- `tools/repl/repl.go` — REPL driver
- `axc repl` subcommand (or just `axc` with no args)

## Dependencies
- p02-p04: frontend pipeline — incremental compilation
- p08: C backend — compile fragment to C → gcc → shared library → dlopen
- p16-t11: std-fmt — Display for printing results

## Detailed Requirements

```go
type REPL struct {
    Session  *ReplSession
    History  []string
    Readline *readline.Instance
}

type ReplSession struct {
    Bindings  map[string]TypedValue  // accumulated var/fn bindings
    ImportedModules []string
    TempDir   string  // for compiled fragments
}

func (r *REPL) Run() error
func (r *REPL) Eval(input string) (string, error)
func (r *REPL) compileFragment(code string, session *ReplSession) (*CompiledFrag, error)
func (r *REPL) execFragment(frag *CompiledFrag) (interface{}, error)
```

REPL interaction:
```
axiom> let x = 42
x: i32 = 42

axiom> x * 2
84

axiom> fn greet(name: str) -> str: "hello, {name}"
greet: fn(str) -> str

axiom> greet("world")
"hello, world"

axiom> import std.math
axiom> std.math.sqrt(2.0)
1.4142135623730951

axiom> :help     # REPL commands
axiom> :quit
```

REPL special commands:
- `:help` — show commands
- `:quit` / `:q` — exit
- `:type <expr>` — show type without evaluating
- `:history` — show command history
- `:clear` — clear session bindings
- `:load <file>` — load AXIOM file into session

Compilation strategy:
1. Wrap input in synthetic module with all accumulated bindings.
2. Compile to C (fast), pipe to gcc → `.so`.
3. `dlopen()` the `.so`, call the fragment function.
4. Print result using Display.to_str().

## Implementation Steps

1. Create `tools/repl/repl.go`.
2. Integrate readline library for line editing + history.
3. Implement session state accumulation.
4. Implement fragment wrapping (inject accumulated bindings).
5. Implement compilation + dlopen execution loop.
6. Implement special commands (`:help`, `:quit`, etc.).
7. Write tests for expression evaluation.

## Test Plan
- `TestReplExpression`: `2 + 2` → prints "4"
- `TestReplBinding`: `let x = 42`, then `x` → prints "42"
- `TestReplFunction`: define + call function → correct result
- `TestReplImport`: `import std.math` → math functions available
- `TestReplHistory`: up-arrow → previous command recalled

## Validation Checklist
- [ ] Session accumulates bindings across prompts
- [ ] Type errors shown with location (best-effort for single-line input)
- [ ] Special commands all work
- [ ] REPL exits cleanly on Ctrl-D and Ctrl-C

## Acceptance Criteria
- AXIOM REPL can run a fibonacci function defined across multiple prompts

## Definition of Done
- [ ] `tools/repl/repl.go` implemented
- [ ] Expression evaluation test passes

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Compilation latency per keystroke | Compile on Enter only; don't attempt partial compilation |
| dlopen symbol collisions between fragments | Use unique symbol names per fragment (fragment_N prefix) |

## Future Follow-up Tasks
- Persistent session: save/load session to file
- Syntax highlighting in REPL (via ANSI escape codes)
