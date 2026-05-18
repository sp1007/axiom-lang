# p09-t03: AIR Metadata Table

## Purpose
Implement the metadata table that stores debug information and AI-semantic annotations for each AIR instruction without inflating the core 16-byte instruction size. Metadata is stored in a parallel array indexed by instruction index.

## Context
AIR instructions are 16 bytes and must stay that way for cache performance. But each instruction may need: source location (file:line:col), ownership annotations, AI hints (`@[ai::assert_pure]`), and DWARF debug info. The metadata table is a parallel array — `meta[i]` corresponds to `insts[i]`. Most instructions have no metadata (null entry), so the table is sparse and memory-efficient.

## Inputs
- `ir/air/cfg.go` — AirFunc to extend with metadata
- Source location info from lexer LineTable
- Semantic annotations from sema passes

## Outputs
- `ir/air/meta.go` — AirMeta struct and MetaTable
- Extended AirFunc with metadata access

## Dependencies
- p09-t02: air-basic-blocks — AirFunc is the container
- p03-t03: ast-printer — source location format reference

## Subsystems Affected
- AIR printer (p09-t05): uses metadata for source annotations
- DWARF emitter (p11-t13): reads source locations
- .axmeta writer (p11-t14): exports AI annotations
- Debugger: uses metadata for stepping

## Detailed Requirements

1. `AirMeta` struct:
   ```go
   type AirMeta struct {
       SourceFile   uint32  // interned file path
       SourceLine   uint32
       SourceCol    uint16
       OwnerInfo    uint8   // 0=none, 1=stack, 2=heap, 3=arena
       AIHints      uint32  // index into AIHintTable (0=none)
   }
   ```
2. `MetaTable` struct: sparse `map[uint32]*AirMeta` (instIdx → meta)
3. `AirFunc.Meta *MetaTable` — field added to AirFunc
4. API:
   - `MetaTable.Set(instIdx uint32, meta AirMeta)`
   - `MetaTable.Get(instIdx uint32) *AirMeta` — returns nil if no metadata
5. `AIHintTable`: flat array of `AIHint{Kind:AIHintKind, Data:string}` entries
   - `AIHintKind`: AssertPure, SuggestSoA, Explain, SuggestVectorize
6. Source location set for: all VarDecl-derived instructions, all CallExpr-derived instructions, function entry.
7. Ownership info: `OwnerInfo` set by escape analysis results from sema.
8. JSON serialization for `.axmeta` export.

## Implementation Steps

1. Create `ir/air/meta.go`.
2. Implement `AirMeta`, `MetaTable`, `AIHintTable`.
3. Add `Meta *MetaTable` field to `AirFunc`.
4. In AIR builder: when emitting VarDecl-derived instructions, call `meta.Set(instIdx, AirMeta{SourceLine:...})`.
5. In escape analysis output integration: set OwnerInfo based on escape flag.
6. Implement JSON marshaling for .axmeta export.
7. Write unit tests: `TestMetaSetGet`, `TestMetaSparse`, `TestMetaJSON`.

## Test Plan

- `TestMetaSetGet`: set metadata for instruction 5, get back same data
- `TestMetaSparse`: instructions without metadata return nil from Get()
- `TestMetaSourceLocation`: AIR builder sets source lines correctly
- `TestMetaJSON`: round-trip JSON serialization

## Validation Checklist

- [ ] AirMeta does not affect AirInst size (separate array)
- [ ] Sparse storage — unset entries return nil
- [ ] Source locations set for debug-relevant instructions
- [ ] AIHints stored and retrievable
- [ ] JSON serialization complete

## Acceptance Criteria

- `axc dump-air --debug` shows source locations for each instruction
- MetaTable adds < 100 bytes per function on average

## Definition of Done

- [ ] `ir/air/meta.go` implemented
- [ ] AirFunc extended with Meta field
- [ ] AIR builder sets source locations
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Memory usage of MetaTable for large functions | Use sparse map, only store where metadata exists |
| Source location accuracy after AIR transformations | Preserve metadata through transformations (copy meta when cloning insts) |

## Future Follow-up Tasks

- p09-t05: air-printer uses metadata for annotation
- p11-t13: dwarf-line-info reads source locations from MetaTable
- p11-t14: axmeta-writer exports AI hints
