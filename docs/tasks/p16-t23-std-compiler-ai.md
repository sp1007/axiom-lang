# p16-t23: `std.compiler.ai` — AI Semantic Layer API

## Purpose
Implement the `std.compiler.ai` module providing the programmatic API for AI tooling to query `.axmeta` semantic graphs, emit optimization suggestions, and integrate with the `@[ai::*]` annotation system.

## Context
Plan §Phase 9 lists: _"`std/compiler/ai.ax` — AI optimization API (std.compiler.ai)"_. The `.axmeta` section (p11-t14) contains compressed semantic data; this module provides the runtime query API for AI assistants (Copilot, etc.) to read and act on that data.

## Inputs
- `.axmeta` format from p11-t14 (Zstd-compressed JSON)
- ConnectionGraph structure from p06-t01
- Effect system data from p04-t09
- `@[ai::*]` annotation parsing from parser

## Outputs
- `std/compiler/ai.ax` — AI query API
- `std/compiler/axmeta.ax` — `.axmeta` reader/decoder
- Tests

## Dependencies
- p16-t01: std-testing-assert — test framework
- p11-t14: axmeta-writer — `.axmeta` format definition
- p16-t09: std-json — JSON parsing for `.axmeta` content

## Detailed Requirements

### API Surface

```axiom
pub struct AxMeta:
    fn load(path: string) -> Result[AxMeta, Error]
    fn symbols(self) -> Seq[SymbolInfo]
    fn type_of(self, name: string) -> Option[TypeInfo]
    fn connection_graph(self) -> ConnectionGraph
    fn effects_of(self, fn_name: string) -> Seq[Effect]

pub struct SymbolInfo:
    name: string
    kind: SymbolKind   // Function, Struct, Const, etc.
    type_info: TypeInfo
    source_loc: SourceLoc
    doc: Option[string]

// AI annotation verification
pub fn assert_pure(fn_name: string) -> bool
pub fn suggest_layout(struct_name: string) -> Option[LayoutSuggestion]
```

### Integration Points
- LSP server (p17-t03) uses this API for hover/completion with AI context
- `axc fix` (p17-t10) uses this API for AI-assisted migrations
- External AI tools read `.axmeta` via this API

### `.axmeta` Decoding
1. Read `.axmeta` section from ELF/PE/Mach-O binary
2. Decompress Zstd payload
3. Parse JSON into `AxMeta` struct
4. Expose queries via typed API

## Implementation Steps

1. Create `std/compiler/axmeta.ax` with Zstd decompression + JSON parsing.
2. Create `std/compiler/ai.ax` with query API.
3. Implement `assert_pure` verification against effect data.
4. Implement `suggest_layout` based on ConnectionGraph analysis.
5. Write tests using test binaries with known `.axmeta` content.

## Test Plan

- `TestAxMetaLoad`: load `.axmeta` from test binary → symbols enumerable
- `TestAxMetaTypeOf`: query known function → correct type signature
- `TestAssertPure`: pure function → true, impure → false
- `TestAxMetaRoundTrip`: write `.axmeta` (p11-t14) → read back → identical data

## Acceptance Criteria

- AI tools can query type, effect, and ownership data from compiled binaries
- `@[ai::assert_pure]` verified at compile time via this module

## Definition of Done

- [ ] `std/compiler/ai.ax` implemented
- [ ] `std/compiler/axmeta.ax` implemented
- [ ] Tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| `.axmeta` format changes | Version field in JSON; backward-compat reader |
| Zstd dependency | Use pure-AXIOM Zstd decoder or link C library via FFI |
