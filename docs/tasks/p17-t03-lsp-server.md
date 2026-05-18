# p17-t03: LSP Server

## Purpose
Implement a Language Server Protocol (LSP) server for AXIOM, providing IDE integration with hover documentation, go-to-definition, completion, diagnostics, and formatting for VS Code, Neovim, and other LSP-compatible editors.

## Context
LSP enables AXIOM to integrate with any editor without per-editor plugins. The server maintains an in-memory compilation state, responds to editor requests, and provides real-time feedback. It leverages `.axmeta` (p11-t14) for cross-module type information.

## Inputs
- LSP protocol (JSON-RPC 2.0 over stdio/TCP)
- AXIOM source via `textDocument/didChange` notifications
- `axc check` output for diagnostics
- `.axmeta` files from compiled modules

## Outputs
- `tools/lsp/server.go` — LSP server
- `tools/lsp/handlers.go` — per-request handlers
- `axc lsp` subcommand (starts server on stdio)

## Dependencies
- p17-t02: axc-check — diagnostics source
- p11-t14: axmeta-writer — cross-module symbol info
- p17-t01: axc-fmt — formatting provider

## Detailed Requirements

LSP capabilities implemented (MVP):
- `textDocument/publishDiagnostics` — real-time error reporting
- `textDocument/completion` — identifier completion
- `textDocument/hover` — type on hover + doc string
- `textDocument/definition` — go to definition (file:line:col)
- `textDocument/formatting` — format document via `axc fmt`
- `textDocument/documentSymbol` — file outline (functions, types)
- `workspace/symbol` — cross-file symbol search

```go
type LSPServer struct {
    Conn       *jsonrpc.Conn
    Workspace  string
    FileCache  map[string]*ParsedFile  // open files → parsed AST
    ModuleCache map[string]*AxMetaSection  // .axmeta for compiled modules
}

func StartServer(stdio bool) error
func (s *LSPServer) HandleRequest(method string, params json.RawMessage) (interface{}, error)
```

Incremental compilation:
- On `textDocument/didChange`: re-parse changed file, re-check only affected files.
- Cache ASTs and TypedAST for unchanged files.
- Provide diagnostics within 200ms of keystroke (target).

Completion algorithm:
1. Parse up to cursor position.
2. Determine context: inside identifier, after `.`, after `import`.
3. Filter symbol table by prefix.
4. Sort by relevance (local scope first, then module, then stdlib).

Hover: lookup symbol at cursor → return TypeInfo.to_str() + doc string.

Go-to-definition: cursor on identifier → find definition site from SemanticGraph.

## Implementation Steps

1. Create `tools/lsp/server.go` with JSON-RPC 2.0 stdio transport.
2. Implement `initialize` / `initialized` handshake.
3. Implement `textDocument/didOpen` / `didChange` → trigger re-check.
4. Implement `publishDiagnostics` from `axc check` output.
5. Implement `completion` with prefix filtering.
6. Implement `hover` from TypeInfo.
7. Implement `definition` from SemanticGraph source locations.
8. Implement `formatting` via `axc fmt`.
9. Test with VS Code extension.

## Test Plan
- `TestLSPInit`: initialize handshake → server capabilities returned
- `TestLSPDiagnostics`: open file with error → diagnostic published within 200ms
- `TestLSPCompletion`: type "std.math." → math functions listed
- `TestLSPHover`: hover on `i32` variable → type "i32" shown
- `TestLSPDefinition`: go-to-def on user function → correct file:line

## Validation Checklist
- [ ] JSON-RPC messages parse correctly (no schema violations)
- [ ] Diagnostics include correct line:col from source
- [ ] Completion sorted: local vars before stdlib
- [ ] Formatting via axc fmt is idempotent

## Acceptance Criteria
- VS Code with AXIOM extension shows real-time errors and completion

## Definition of Done
- [ ] `tools/lsp/server.go` implemented
- [ ] VS Code extension test shows diagnostics + hover

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| LSP latency > 200ms on large files | Incremental parsing; only re-parse changed line range |
| JSON-RPC protocol edge cases | Use well-tested json-rpc library (sourcegraph/jsonrpc2) |

## Future Follow-up Tasks
- `textDocument/rename` — rename all uses of a symbol
- `textDocument/codeAction` — quick fixes for common errors
- VS Code extension packaging
