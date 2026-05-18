# p11-t13: DWARF Line Info Emission

## Purpose
Emit DWARF `.debug_line` section mapping machine code addresses to source file line numbers, enabling `gdb`/`lldb` to show source locations during debugging.

## Context
Without debug info, AXIOM programs are nearly undebuggable — stack traces show raw addresses, not source lines. DWARF is the standard debug format on Linux/macOS. The `.debug_line` section is the highest-value addition: it enables `gdb` to show "file.ax:42" instead of "0x401234".

## Inputs
- Machine code address → AIR instruction mapping (from emitter)
- AIR instruction → source location (file, line, col) from p09-t03 (AirMeta)
- Function symbol addresses from ELF symbol table

## Outputs
- `codegen/native/x86/dwarf.go` — DWARF line number program emitter
- `.debug_line` section bytes added to ELF output

## Dependencies
- p11-t10: x86-machine-code-emitter — machine code offsets
- p11-t12: elf64-emitter — adds .debug_line section to object file
- p09-t03: air-metadata-table — source location for each AIR instruction

## Subsystems Affected
- ELF64 emitter: receives .debug_line bytes as additional section
- Debug experience: enables line-level debugging with gdb

## Detailed Requirements

DWARF4 line number program (simplified state machine):

```go
type LineEntry struct {
    PC   uint64  // machine code address
    File uint16  // index into file table
    Line uint32
    Col  uint16
    IsStmt bool
}

type DWARFLineWriter struct {
    Header    DWARFLineHeader
    Entries   []LineEntry
    FileTable []string
}

type DWARFLineHeader struct {
    UnitLength        uint32
    Version           uint16  // 4
    HeaderLength      uint32
    MinInstrLen       uint8   // 1
    DefaultIsStmt     uint8   // 1
    LineBase          int8    // -5
    LineRange         uint8   // 14
    OpcodeBase        uint8   // 13
    StandardOpLengths [12]uint8
}

func (w *DWARFLineWriter) AddFile(filename string) uint16
func (w *DWARFLineWriter) EmitEntry(pc uint64, file uint16, line, col uint32)
func (w *DWARFLineWriter) Serialize() []byte
```

Line number program opcodes used:
- `DW_LNS_advance_pc` (0x02): advance PC by arg * min_instr_length
- `DW_LNS_advance_line` (0x03): advance line by signed LEB128
- `DW_LNS_set_file` (0x04): change current file
- `DW_LNS_copy` (0x01): emit a row with current state
- Special opcodes: encode small PC+line deltas in one byte

Encoding strategy (simple):
1. Sort LineEntry by PC.
2. Emit DW_LNS_set_file for each new file.
3. Emit DW_LNS_advance_pc for each PC delta.
4. Emit DW_LNS_advance_line for each line delta.
5. Emit DW_LNS_copy to record the row.
6. Emit DW_LNE_end_sequence at function end.

## Implementation Steps

1. Create `codegen/native/x86/dwarf.go`.
2. Define DWARF4 line header constants.
3. Implement `DWARFLineWriter` with file table and entry accumulation.
4. Implement LEB128 encoding (unsigned and signed).
5. Implement `Serialize()` — write header + line number program.
6. Wire into ELF64 emitter: add `.debug_line` section.
7. Test with `readelf --debug-dump=line` and `gdb` source display.

## Test Plan
- `TestDWARFHeader`: serialized header has correct DWARF4 magic fields
- `TestDWARFLineEntry`: single function, 3 lines → 3 rows in line program
- `TestDWARFLEB128`: verify signed/unsigned LEB128 encoding
- `TestDWARFGDBIntegration`: `gdb -ex "list main"` shows source lines

## Validation Checklist
- [ ] DWARF4 version (0x0004) in header
- [ ] File table contains all source files referenced
- [ ] PC addresses match actual machine code offsets
- [ ] `readelf --debug-dump=line` shows correct file:line mappings

## Acceptance Criteria
- `gdb` shows `file.ax:12` when stopped at a breakpoint in AXIOM code

## Definition of Done
- [ ] `codegen/native/x86/dwarf.go` implemented
- [ ] .debug_line section added to ELF output
- [ ] `readelf` verification passes

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| LEB128 encoding error → corrupted debug info | Unit test each LEB128 value against known-good encoding |
| PC offsets wrong (relative vs absolute) | DWARF uses segment-relative; set initial PC from symbol offset |

## Future Follow-up Tasks
- DWARF `.debug_info` and `.debug_abbrev` for type/variable info (post-MVP)
- p12-t02: PE-COFF debug info (CodeView PDB) for Windows
