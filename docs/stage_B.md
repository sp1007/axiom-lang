# AXIOM Toolchain — Stage B Task Specifications: Static Data, Type Mangling & Cross-Platform Linker Autonomy

This document defines the technical task specifications, structural design, inputs, outputs, and success criteria for implementing Stage B of AXIOM's self-hosted compiler pipeline.

Stage B resolves key remaining barriers in native code emission, standard library capability, and platform-agnostic compilation:
1. **Static Data Segment Emission (`.rdata`/`.data`)** to support string literals and static pointers natively.
2. **Type-Directed Name Mangling** to properly overload and resolve printer functions (`print`/`println`) natively without memory faults.
3. **ELF64 PLT/GOT Relocation Resolver** for full autonomy on Linux targets.

---

## 🎯 Task 1: PE/COFF Static Data Segment Emission (`.rdata` / `.data`)
* **Objective**: Upgrade AXIOM's native PE/COFF object writer (`x86_coff.ax`) and custom linker (`linker.ax`) to compile, emit, relocate, and address static global variables and raw string literals natively.

### 📋 Technical Design:
1. **Emitter Segment Support**:
   * Add a `data` byte vector to the machine emitter structure: `pub data: ByteVec` in `x86_emitter.ax`.
   * When the compiler linearizer/generator encounters a string literal (e.g., `"hello"`), allocate a unique label and push the raw UTF-8 bytes (null-terminated) into the `data` segment.
2. **COFF Section Generation**:
   * Modify `x86_coff.ax` to output a `.rdata` (Read-Only Data) section header alongside `.text`.
   * Set section characteristics: `IMAGE_SCN_CNT_INITIALIZED_DATA | IMAGE_SCN_MEM_READ` (`0x40000040`).
3. **RIP-Relative Relocation Resolution**:
   * Map reference pointers using RIP-relative addressing: `lea rsi, [rip + displacement]`.
   * Emit an `IMAGE_REL_AMD64_REL32` relocation record pointing from the `.text` reference site to the corresponding string symbol in the `.rdata` section.
   * Update the custom linker (`linker.ax`) to correctly resolve cross-section relocations.

### 🏁 Task Criteria:
* **Input**: AXIOM source files carrying raw string literals or static global variables.
* **Output**: A standalone `.exe` carrying fully resolved, non-zero `.rdata` segments containing correct string byte vectors.
* **Success Indicator**: Programs like `valid_hello.ax` compile, self-link, and print text to stdout flawlessly without crashing.

---

## 🎯 Task 2: Type-Directed Name Mangling & Overloading for Prints
* **Objective**: Natively support polymorphic and overloaded printing via type-directed name mangling inside AXIOM's instruction selector (`x86_selector.ax`) and standard symbol parser.

### 📋 Technical Design:
1. **Inspect AST Argument Types**:
   * Modify the name resolver/selector to examine the precise type IDs of arguments passed to `print()` and `println()` calls during IR lowering.
2. **Type-to-Mangled-Symbol Mapping**:
   * Map `println()` and `print()` function calls dynamically based on type signatures:
     - `println(string)` or `println(str)` $\rightarrow$ `ax_println_str`
     - `println(i32)` or `println(i64)` $\rightarrow$ `ax_println_i64`
     - `println(f32)` or `println(f64)` $\rightarrow$ `ax_println_f64`
     - `println(bool)` $\rightarrow$ `ax_println_bool`
3. **Register Binding**:
   * Ensure that the argument value is correctly bound to `RCX` (Win64) or `RDI` (System V) depending on the target ABI, and call the mangled FFI dynamic target correctly.

### 🏁 Task Criteria:
* **Input**: An AXIOM source compiling multiple `println` expressions on primitive types (e.g., `println(10)`, `println(true)`).
* **Output**: A compiler-generated object file referencing differentiated dynamic imported targets (`ax_println_i64`, `ax_println_bool`) in `ax_runtime.dll`.
* **Success Indicator**: Executables like `valid_struct_e2e.ax` and numerical outputs run natively without triggering Access Violations (`0xC0000005`).

---

## 🎯 Task 3: ELF64 Linux PLT/GOT Relocation Resolver
* **Objective**: Complete cross-platform linker autonomy by upgrading AXIOM's custom linker (`linker.ax`) to generate and patch Procedure Linkage Table (PLT) and Global Offset Table (GOT) segments on ELF64 Linux systems.

### 📋 Technical Design:
1. **Dynamic Section Construction**:
   * Program `linker.ax` to dynamically output ELF64 standard segments:
     - `.got` (Global Offset Table) holding absolute addresses of shared library variables/functions.
     - `.plt` (Procedure Linkage Table) containing execution branch stubs.
     - `.dynsym` and `.dynstr` containing imported dynamic symbols and their string names.
2. **Relocation & Lazy Resolution**:
   * Intercept `R_X86_64_PLT32` and `R_X86_64_GOTPCREL` relocations in the linker.
   * Redirect relative calls targeting dynamic C/POSIX functions to their respective PLT slots (`jmp qword ptr [got_entry]`).
   * Properly populate the `.dynamic` program header enabling standard Linux dynamic loaders (`ld-linux-x86-64.so`) to resolve imports at load-time.

### 🏁 Task Criteria:
* **Input**: An ELF64 relocatable object file (`.o`) carrying dynamic shared library FFI references.
* **Output**: A stand-alone dynamically-linked ELF64 executable carrying fully functional GOT and PLT structures.
* **Success Indicator**: Running the compiled program on Linux yields correct exit codes and outputs, confirmed by `readelf -d` and `ldd` verification.
