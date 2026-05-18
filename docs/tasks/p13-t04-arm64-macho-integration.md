# p13-t04: ARM64 Mach-O Integration (Apple Silicon)

## Purpose
Integrate the ARM64 native backend with the Mach-O object file emitter to produce valid executables for Apple Silicon (M1/M2/M3) macOS. This includes Mach-O binary encoding of ARM64 instructions, segment/section layout, ad hoc code signing, and the 16KB page size requirement specific to Apple Silicon.

## Context
Apple Silicon uses the ARM64 (AArch64) ISA but has Apple-specific requirements on top of the standard AArch64 Linux ABI. Key differences:
- Page size is 16KB on Apple Silicon (vs 4KB on Linux ARM64 and x86-64)
- All executable binaries must be code-signed — for local development, ad hoc signing (`codesign --sign -`) is sufficient
- The `__TEXT` segment must be page-aligned at 16KB boundaries
- The `LC_BUILD_VERSION` load command must specify the minimum macOS version
- Symbol table format follows the Mach-O N_LIST nlist64 structure
- Dynamic linking uses dyld (Apple's dynamic linker) with `LC_LOAD_DYLINKER` and `LC_LOAD_DYLIB`

This task builds on the existing Mach-O emitter from p12-t03 (which was built for x86-64 Mach-O). The ARM64 variant reuses the Mach-O structure layer and adds ARM64-specific instruction encoding and code signing.

## Inputs
- `MachineInstr` list from ARM64 instruction selector (p13-t01)
- `FrameLayout` from ARM64 register allocator (p13-t02)
- `ABIAssignment` from AAPCS64 ABI (p13-t03)
- Existing Mach-O emitter from p12-t03 (`linker/macho/`)
- Target triple: `aarch64-apple-macos13`

## Outputs
- `codegen/native/arm64/encoder.go` — ARM64 32-bit instruction binary encoding
- `linker/macho/arm64_emitter.go` — ARM64-specific Mach-O object file emission
- `linker/macho/codesign.go` — ad hoc code signing implementation
- Test file: `linker/macho/arm64_emitter_test.go`
- Test binary: `tests/codegen/arm64/hello_macho` — runnable on Apple Silicon

## Dependencies
- p13-t03: AAPCS64 ABI (call convention used in object file)
- p12-t03: Existing Mach-O emitter infrastructure for x86-64

## Subsystems Affected
- `linker/macho/` — new ARM64 emitter file, codesign file
- `codegen/native/arm64/` — new encoder file
- `compiler/driver/` — `--target=aarch64-apple-macos13` triggers this path

## Detailed Requirements

### ARM64 Instruction Encoding
ARM64 uses fixed 32-bit little-endian instruction encoding. Each instruction class has its own encoding format. Key encodings needed:

#### Data Processing (Register)
```
ADD (shifted register):
  [31]    = sf (1 = 64-bit)
  [30:29] = op/S (00 = ADD)
  [28:24] = 01011
  [23:22] = shift (00=LSL, 01=LSR, 10=ASR)
  [21]    = 0
  [20:16] = Rm
  [15:10] = imm6 (shift amount)
  [9:5]   = Rn
  [4:0]   = Rd

MUL (alias for MADD with Ra=XZR):
  [31]    = sf (1)
  [30:29] = 00
  [28:24] = 11011
  [23:21] = 000
  [20:16] = Rm
  [15]    = 0 (Ra != 31 means MADD)
  [14:10] = Ra (11111 = XZR)
  [9:5]   = Rn
  [4:0]   = Rd
```

#### Load/Store
```
LDR (unsigned offset, 64-bit):
  [31:30] = 11 (size = 8 bytes)
  [29:27] = 111
  [26]    = 0 (not SIMD)
  [25:24] = 01
  [23:22] = 01 (load)
  [21:10] = imm12 (byte offset / 8 for 64-bit)
  [9:5]   = Rn (base)
  [4:0]   = Rt (dest)

STR (unsigned offset, 64-bit):
  Similar but [23:22] = 00 (store)
```

#### Branches
```
BL (branch with link):
  [31:26] = 100101
  [25:0]  = imm26 (PC-relative word offset)

B.cond:
  [31:24] = 01010100
  [23:5]  = imm19 (PC-relative word offset)
  [4]     = 0
  [3:0]   = cond (0000=EQ, 0001=NE, 1011=LT, 1101=LE, 1100=GT, 1010=GE)

RET:
  [31:10] = 1101011001011111000000
  [9:5]   = Rn (11110 = X30)
  [4:0]   = 00000
```

#### STP/LDP (pair)
```
STP (pre-index, 64-bit):
  [31:30] = 10 (64-bit)
  [29:27] = 101
  [26]    = 0
  [25:23] = 011 (pre-index)
  [22]    = 0 (store)
  [21:15] = imm7 (byte offset / 8)
  [14:10] = Rt2
  [9:5]   = Rn (base)
  [4:0]   = Rt1
```

### Mach-O ARM64 Object Layout

#### Segment/Section Layout
```
Mach-O Header (32 bytes, magic = 0xFEEDFACF for 64-bit)
Load Commands:
  LC_SEGMENT_64 "__TEXT" {r-x, addr=0x100000000, size=aligned_to_16KB}
    Section "__TEXT.__text"      — machine code
    Section "__TEXT.__stubs"     — PLT stubs (for dynamic symbols)
    Section "__TEXT.__stub_helper" — dyld stub helper
    Section "__TEXT.__cstring"   — string literals
    Section "__TEXT.__unwind_info" — compact unwind (for stack unwinding)
  LC_SEGMENT_64 "__DATA_CONST" {rw-, addr=after __TEXT}
    Section "__DATA_CONST.__got" — global offset table
  LC_SEGMENT_64 "__DATA" {rw-}
    Section "__DATA.__data"      — initialized data
    Section "__DATA.__bss"       — zero-initialized data
  LC_SEGMENT_64 "__LINKEDIT" {r--} — symbol table, string table, code signature
  LC_DYLD_INFO_ONLY — dyld rebase/bind info
  LC_SYMTAB — symbol table
  LC_DYSYMTAB — dynamic symbol table
  LC_LOAD_DYLINKER "/usr/lib/dyld"
  LC_LOAD_DYLIB "/usr/lib/libSystem.B.dylib"
  LC_BUILD_VERSION {platform=macOS, minos=13.0.0, sdk=13.0.0}
  LC_MAIN {entryoff, stacksize}
  LC_CODE_SIGNATURE — ad hoc code signing
```

#### Page Size: 16KB
```go
const AppleSiliconPageSize = 0x4000  // 16384 bytes

// Segments must be aligned to 16KB on Apple Silicon
func alignToPage(offset int) int {
    return (offset + AppleSiliconPageSize - 1) &^ (AppleSiliconPageSize - 1)
}
```

### Ad Hoc Code Signing
macOS on Apple Silicon requires all executables to be signed. For local development:
```
codesign --sign - --force <binary>
```
But since `axc` must produce signed binaries without shell tools, implement ad hoc signing inline:

Ad hoc signing uses a special code signature blob with a null hash identity. The `LC_CODE_SIGNATURE` load command points to a `CS_SuperBlob` → `CS_CodeDirectory` structure.

```go
type CSCodeDirectory struct {
    Magic        uint32  // 0xFADE0C02
    Length       uint32
    Version      uint32  // 0x20400
    Flags        uint32  // CS_ADHOC = 0x0002
    HashOffset   uint32
    IdentOffset  uint32
    NSpecialSlots uint32 // 0
    NCodeSlots    uint32 // ceil(codeSize / pageSize)
    CodeLimit     uint32
    HashSize      uint8  // 32 for SHA-256
    HashType      uint8  // CS_HASHTYPE_SHA256 = 2
    // ...
}
```

For ad hoc: `Flags = CS_ADHOC (0x0002)`. The code hashes are SHA-256 of each 4KB (not 16KB) page of the `__TEXT` segment. Leave hash slots as zeros for ad hoc — the OS accepts this.

### LC_BUILD_VERSION
Required on macOS 11+:
```go
type LCBuildVersion struct {
    Cmd       uint32  // 0x32
    CmdSize   uint32
    Platform  uint32  // PLATFORM_MACOS = 1
    MinOS     uint32  // 0x000D0000 = 13.0.0
    SDK       uint32  // 0x000D0000 = 13.0.0
    NumTools  uint32  // 0
}
```

## Implementation Steps

### Step 1: ARM64 Instruction Encoder
Create `codegen/native/arm64/encoder.go`:
```go
func EncodeInstr(instr MachineInstr) uint32 {
    switch instr.Opcode {
    case ADD:
        return encodeDataProcReg(1, 0, 0, instr.Rm, 0, instr.Rn, instr.Rd)
    case LDR:
        offset12 := uint32(instr.Imm / 8)  // scaled offset
        return 0xF9400000 | (offset12 << 10) | (uint32(instr.Rn) << 5) | uint32(instr.Rd)
    case BL:
        // imm26 is PC-relative word offset
        return 0x94000000 // placeholder; linker fills in relocation
    case RET:
        return 0xD65F03C0  // RET X30
    case STP:
        return encodeSTPPreIndex(instr.Rn, instr.Rm, instr.Base, instr.Offset)
    // ...
    default:
        panic(fmt.Sprintf("encoder: unhandled opcode %v", instr.Opcode))
    }
}
```

### Step 2: ARM64 Mach-O Emitter
Create `linker/macho/arm64_emitter.go`:
```go
type ARM64MachoEmitter struct {
    base    *MachoEmitter  // reuse from p12-t03
    target  TargetTriple
    instrs  []MachineInstr
}

func (e *ARM64MachoEmitter) Emit(fn *arm64.CompiledFunction) error {
    // Encode each instruction to 4 bytes
    code := make([]byte, 0, len(fn.Instrs)*4)
    for _, instr := range fn.Instrs {
        enc := arm64.EncodeInstr(instr)
        code = binary.LittleEndian.AppendUint32(code, enc)
    }
    e.base.AddCode(fn.Name, code)
    return nil
}
```

### Step 3: Code Signing
Create `linker/macho/codesign.go`:
```go
func AppendAdHocSignature(binary []byte, codeSize int) []byte {
    pageSize := 4096  // CS uses 4KB pages even on Apple Silicon
    nPages := (codeSize + pageSize - 1) / pageSize
    
    cdirSize := 88 + 32*nPages  // CodeDirectory header + SHA-256 hashes
    blobSize := 12 + cdirSize   // SuperBlob header + CodeDirectory
    
    superBlob := buildSuperBlob(cdirSize)
    cdir := buildCodeDirectory(codeSize, nPages)
    
    // For ad hoc: leave all page hashes as zeros
    sig := append(superBlob, cdir...)
    sig = append(sig, make([]byte, 32*nPages)...)  // zero hashes
    
    return append(binary, sig...)
}
```

### Step 4: LC_CODE_SIGNATURE Load Command
```go
func (e *ARM64MachoEmitter) addCodeSignatureLC(codeSize int) {
    sigOffset := e.linkeditOffset  // offset of signature in file
    sigSize := computeSigSize(codeSize)
    
    lc := LCCodeSignature{
        Cmd:        0x1D,  // LC_CODE_SIGNATURE
        CmdSize:    16,
        DataOffset: uint32(sigOffset),
        DataSize:   uint32(sigSize),
    }
    e.addLoadCommand(lc)
}
```

### Step 5: End-to-End Test
```go
func TestHelloWorldMachO(t *testing.T) {
    // Compile hello.ax → ARM64 MachineInstrs → Mach-O binary
    binary := compileToMachO("tests/hello.ax", "aarch64-apple-macos13")
    
    // Write to temp file
    tmpf := writeTempFile(binary)
    
    // Sign it (ad hoc)
    exec.Command("codesign", "--sign", "-", "--force", tmpf).Run()
    
    // Run it (only on Apple Silicon)
    if runtime.GOARCH == "arm64" && runtime.GOOS == "darwin" {
        out, err := exec.Command(tmpf).Output()
        require.NoError(t, err)
        require.Equal(t, "Hello, World!\n", string(out))
    }
}
```

## Test Plan

### Unit Tests
1. `TestEncodeADD` — encode `ADD X0, X1, X2` → expect `0x8B020020`
2. `TestEncodeLDR` — encode `LDR X0, [X1, #8]` → expect `0xF9400420`
3. `TestEncodeRET` — encode `RET` → expect `0xD65F03C0`
4. `TestEncodeSTPPreIndex` — verify STP encoding with pre-index offset
5. `TestMachOHeaderARM64` — verify magic = 0xFEEDFACF, cputype = 0x0100000C (ARM64)
6. `TestSegmentAlignment16KB` — `__TEXT` segment addr aligned to 16384
7. `TestCodeSignatureBlob` — SuperBlob magic = 0xFADE0CC0, ad hoc flag set
8. `TestBuildVersionLC` — platform=1 (macOS), minos=0x000D0000

### Integration Tests
1. Compile `tests/hello.ax` for `aarch64-apple-macos13`:
   - Output file is valid Mach-O (check with `file` command)
   - `otool -h` shows ARM64 cpu type
   - `codesign -v` succeeds
   - Binary runs on Apple Silicon (if CI runner is macos-14)

### CI Configuration
GitHub Actions `macos-14` runner is Apple Silicon. Add job:
```yaml
- name: Test ARM64 Mach-O
  runs-on: macos-14
  steps:
    - run: go test ./linker/macho/... -run TestHelloWorldMachO
```

## Validation Checklist
- [ ] Instruction encoder produces correct 32-bit encodings for ADD, SUB, MUL, LDR, STR, BL, RET, STP, LDP
- [ ] Mach-O header has correct ARM64 cpu type (0x0100000C) and subtype (0x0)
- [ ] `__TEXT` segment start address 16KB-aligned
- [ ] `LC_BUILD_VERSION` present with macOS platform
- [ ] `LC_CODE_SIGNATURE` present with valid SuperBlob structure
- [ ] Ad hoc signature accepted by macOS (binary runs without "killed: 9" error)
- [ ] Relocations for BL instructions emitted as ARM64_RELOC_BRANCH26

## Acceptance Criteria
1. `axc build --target=aarch64-apple-macos13 tests/hello.ax` produces a Mach-O binary
2. Binary passes `codesign -v` verification
3. Binary runs successfully on a macos-14 GitHub Actions runner producing expected output
4. `otool -l` shows correct load commands including `LC_BUILD_VERSION`
5. 16KB page alignment verified with `pagestuff -a <binary>`

## Definition of Done
- `encoder.go`, `arm64_emitter.go`, `codesign.go` implemented and reviewed
- CI job passes on `macos-14` runner
- All unit tests pass
- Binary executes on Apple Silicon without code signing errors

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| Code signing format changes between macOS versions | Target macOS 13+ (Ventura+); test on macos-14 runner |
| 16KB alignment causes larger-than-expected binaries | Expected — document that small programs have large files due to page alignment |
| BL branch range limit (±128MB) | For large programs, emit thunks; not needed for MVP |
| Compact unwind section required for exception handling | Emit minimal `__unwind_info` with a single "no unwind" entry for MVP |

## Future Follow-up Tasks
- p13-t07: Cross-compile integration tests (test ARM64 output from Linux x86-64 host)
- Universal binary (fat binary) support: combine ARM64 and x86-64 Mach-O in one file — future task
- macOS codesign with Developer ID for App Store distribution — future task
