# p11-t14: .axmeta Section Writer

## Purpose
Serialize and embed the AXIOM semantic graph (type info, symbol metadata, AI hints) into a `.axmeta` section in the object file, enabling tooling, LSP, and AI-assisted features to read compiled artifacts.

## Context
`.axmeta` is AXIOM's custom section containing zstd-compressed JSON with semantic metadata: exported symbols, their types, AIR hints, doc strings, and AI semantic annotations. This enables the LSP server and AI tools to provide rich code intelligence without re-parsing source files.

## Inputs
- TypedAST/SemanticGraph from type checker (p04)
- AIR metadata table from p09-t03 (AI hints, complexity annotations)
- Symbol table: exported function names, types, locations

## Outputs
- `codegen/native/x86/axmeta.go` — .axmeta serializer
- `.axmeta` section bytes added to ELF/PE/Mach-O output

## Dependencies
- p11-t12: elf64-emitter — adds .axmeta as custom section
- p09-t03: air-metadata-table — AI semantic hints
- p04-t02: type-table — TypeInfo for all exported symbols

## Subsystems Affected
- ELF/PE/Mach-O emitters: add .axmeta section with SHT_PROGBITS, no SHF_ALLOC
- LSP server (phase 17): reads .axmeta for code intelligence
- AI semantic layer (phase 18): reads AI hints from .axmeta

## Detailed Requirements

```go
type AxMetaSection struct {
    Version  uint32          `json:"version"`   // 1
    Module   string          `json:"module"`
    Symbols  []AxMetaSym     `json:"symbols"`
    AIHints  []AxMetaAIHint  `json:"ai_hints,omitempty"`
}

type AxMetaSym struct {
    Name     string   `json:"name"`
    Kind     string   `json:"kind"`    // "func", "type", "const", "global"
    TypeSig  string   `json:"type_sig"` // "(i32, i32) -> i32"
    DocStr   string   `json:"doc,omitempty"`
    SourceLoc string  `json:"loc"`     // "file.ax:12:4"
    Exported bool     `json:"exported"`
}

type AxMetaAIHint struct {
    Symbol     string `json:"symbol"`
    Complexity string `json:"complexity"` // "O(n)", "O(1)"
    SuggestSoA bool   `json:"suggest_soa,omitempty"`
    PureFunc   bool   `json:"pure,omitempty"`
}

func BuildAxMeta(module string, symbols []AxMetaSym, hints []AxMetaAIHint) AxMetaSection
func SerializeAxMeta(meta AxMetaSection) ([]byte, error)  // JSON + zstd compress
func DeserializeAxMeta(data []byte) (AxMetaSection, error) // zstd decompress + JSON parse
```

Section format:
```
[4 bytes] Magic: "AXMT"
[4 bytes] Version: uint32 LE
[4 bytes] UncompressedSize: uint32 LE
[N bytes] zstd-compressed JSON
```

In ELF: section name `.axmeta`, type `SHT_PROGBITS`, flags = 0 (not loaded at runtime).

## Implementation Steps

1. Create `codegen/native/x86/axmeta.go`.
2. Define `AxMetaSection`, `AxMetaSym`, `AxMetaAIHint` structs.
3. Implement `BuildAxMeta()` — collect from TypedAST and AIR metadata.
4. Implement `SerializeAxMeta()` — JSON marshal + zstd.Compress.
5. Implement `DeserializeAxMeta()` — zstd.Decompress + JSON unmarshal.
6. Wire into ELF64 emitter: add `.axmeta` section.
7. Write tests: serialize → deserialize → compare.

## Test Plan
- `TestAxMetaRoundtrip`: serialize + deserialize → identical struct
- `TestAxMetaMagic`: first 4 bytes are "AXMT"
- `TestAxMetaCompression`: compressed size < uncompressed for non-trivial input
- `TestAxMetaELFSection`: produced .o has `.axmeta` section visible in `readelf -S`

## Validation Checklist
- [ ] Magic bytes "AXMT" present
- [ ] zstd compression applied
- [ ] All exported symbols present in .axmeta
- [ ] Source locations included (file:line:col)

## Acceptance Criteria
- `axc dump-meta output.o` prints human-readable symbol metadata

## Definition of Done
- [ ] `codegen/native/x86/axmeta.go` implemented
- [ ] Round-trip serialize/deserialize tests pass
- [ ] `.axmeta` section present in produced ELF objects

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| zstd dependency adds complexity | Use pure-Go zstd library (klauspost/compress) |
| .axmeta not loaded at runtime (ELF flags) | Ensure SHF_ALLOC not set — section discarded by loader |

## Future Follow-up Tasks
- LSP server reads .axmeta for hover/completion (phase 17)
- AI semantic layer annotates .axmeta with learned patterns (phase 18)
