// Generated automatically by AXIOM AirCGen Stage 1
#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

#define r_0 0

// Forward Declarations
typedef struct AxiomString AxiomString;
typedef struct Token Token;
typedef struct TokenVec TokenVec;
typedef struct IntVec IntVec;
typedef struct Lexer Lexer;
typedef struct AstNode AstNode;
typedef struct NodeVec NodeVec;
typedef struct AstTree AstTree;
typedef struct InternEntry InternEntry;
typedef struct InternState InternState;
typedef struct InternPool InternPool;
typedef struct Parser Parser;
typedef struct Symbol Symbol;
typedef struct ScopeEntry ScopeEntry;
typedef struct Scope Scope;
typedef struct SymbolVec SymbolVec;
typedef struct ScopeVec ScopeVec;
typedef struct U32Vec U32Vec;
typedef struct SymbolTable SymbolTable;
typedef struct NameResolver NameResolver;
typedef struct ComptimeValue ComptimeValue;
typedef struct TypeEntry TypeEntry;
typedef struct StructField StructField;
typedef struct StructFieldVec StructFieldVec;
typedef struct StructInfo StructInfo;
typedef struct StructInfoVec StructInfoVec;
typedef struct VariantInfo VariantInfo;
typedef struct VariantInfoVec VariantInfoVec;
typedef struct SumInfo SumInfo;
typedef struct SumInfoVec SumInfoVec;
typedef struct FuncInfo FuncInfo;
typedef struct FuncInfoVec FuncInfoVec;
typedef struct TypeEntryVec TypeEntryVec;
typedef struct TypeTable TypeTable;
typedef struct TypeSubst TypeSubst;
typedef struct TypeSubstVec TypeSubstVec;
typedef struct Monomorphizer Monomorphizer;
typedef struct TypeChecker TypeChecker;
typedef struct CGNode CGNode;
typedef struct CGNodeVec CGNodeVec;
typedef struct CGEdge CGEdge;
typedef struct CGEdgeVec CGEdgeVec;
typedef struct U32VecVec U32VecVec;
typedef struct ConnectionGraph ConnectionGraph;
typedef struct OwnershipChecker OwnershipChecker;
typedef struct EscapeAnalyser EscapeAnalyser;
typedef struct CtgcInjector CtgcInjector;
typedef struct AliasReuseOptimizer AliasReuseOptimizer;
typedef struct AirInst AirInst;
typedef struct AirInstVec AirInstVec;
typedef struct BasicBlock BasicBlock;
typedef struct BasicBlockVec BasicBlockVec;
typedef struct AirFunc AirFunc;
typedef struct AirFuncVec AirFuncVec;
typedef struct AirModule AirModule;
typedef struct AirFuncBuilder AirFuncBuilder;
typedef struct LocalMapEntry LocalMapEntry;
typedef struct LocalMap LocalMap;
typedef struct AirModuleBuilder AirModuleBuilder;
typedef struct FuncLowering FuncLowering;
typedef struct ConstVal ConstVal;
typedef struct SsaOptimizer SsaOptimizer;
typedef struct CGenerator CGenerator;
typedef struct WasmGenerator WasmGenerator;
typedef struct MachOperand MachOperand;
typedef struct MachInst MachInst;
typedef struct MachInstVec MachInstVec;
typedef struct InstructionSelector InstructionSelector;
typedef struct LiveInterval LiveInterval;
typedef struct LiveIntervalVec LiveIntervalVec;
typedef struct RegAllocation RegAllocation;
typedef struct RegAllocResult RegAllocResult;
typedef struct StackFrame StackFrame;
typedef struct ByteVec ByteVec;
typedef struct Relocation Relocation;
typedef struct RelocationVec RelocationVec;
typedef struct Fixup Fixup;
typedef struct FixupVec FixupVec;
typedef struct LabelMap LabelMap;
typedef struct MachEmitter MachEmitter;
typedef struct ELF64Sym ELF64Sym;
typedef struct ELF64SymVec ELF64SymVec;
typedef struct COFFReloc COFFReloc;
typedef struct COFFRelocVec COFFRelocVec;
typedef struct COFFSymbol COFFSymbol;
typedef struct COFFSymbolVec COFFSymbolVec;
typedef struct CompiledFuncInfo CompiledFuncInfo;
typedef struct CompiledFuncInfoVec CompiledFuncInfoVec;
typedef struct LinkerSymbol LinkerSymbol;
typedef struct LinkerSymbolVec LinkerSymbolVec;
typedef struct ParsedReloc ParsedReloc;
typedef struct ParsedRelocVec ParsedRelocVec;
typedef struct LinkerStrVec LinkerStrVec;
typedef struct ParsedObject ParsedObject;
typedef struct ParsedObjectPtrVec ParsedObjectPtrVec;
typedef struct AxiomLinker AxiomLinker;
typedef struct FmtToken FmtToken;
typedef struct FmtTokenVec FmtTokenVec;
typedef struct Formatter Formatter;

// Struct Definitions
struct AxiomString {
    void* f_0;
    ax_i64 f_1;
};

struct Token {
    ax_u8 f_0;
    ax_u8 f_1;
    ax_u16 f_2;
    ax_u32 f_3;
};

struct TokenVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct IntVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct Lexer {
    ax_string f_0;
    ax_i64 f_1;
    ax_i64 f_2;
    TokenVec* f_3;
    IntVec* f_4;
    IntVec* f_5;
};

struct AstNode {
    ax_u8 f_0;
    ax_u8 f_1;
    ax_u16 f_2;
    ax_u32 f_3;
    ax_u32 f_4;
    ax_u32 f_5;
    ax_u32 f_6;
    ax_u32 f_7;
};

struct NodeVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct AstTree {
    NodeVec* f_0;
    IntVec* f_1;
    ax_string f_2;
    TokenVec* f_3;
};

struct InternEntry {
    ax_u32 f_0;
    ax_u32 f_1;
    ax_u32 f_2;
    ax_u32 f_3;
    ax_u32 f_4;
    ax_u32 f_5;
};

struct InternState {
    ax_i64 f_0;
    ax_i64 f_1;
    ax_i64 f_2;
    ax_i64 f_3;
};

struct InternPool {
    void* f_0;
    void* f_1;
    void* f_2;
    void* f_3;
};

struct Parser {
    TokenVec* f_0;
    ax_i64 f_1;
    AstTree* f_2;
    void* f_3;
    ax_string f_4;
    ax_i64 f_5;
};

struct Symbol {
    ax_u32 f_0;
    ax_u8 f_1;
    ax_u8 f_2;
    ax_u16 f_3;
    ax_u32 f_4;
    ax_u32 f_5;
    ax_u32 f_6;
    ax_u32 f_7;
};

struct ScopeEntry {
    ax_u32 f_0;
    ax_u32 f_1;
};

struct Scope {
    ax_u8 f_0;
    ax_u8 f_1;
    ax_u16 f_2;
    ax_u32 f_3;
    ax_u32 f_4;
    void* f_5;
    ax_u32 f_6;
    ax_u32 f_7;
};

struct SymbolVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct ScopeVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct U32Vec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct SymbolTable {
    SymbolVec* f_0;
    ScopeVec* f_1;
    U32Vec* f_2;
};

struct NameResolver {
    AstTree* f_0;
    void* f_1;
    void* f_2;
};

struct ComptimeValue {
    ax_u32 f_0;
    ax_i64 f_1;
    ax_f64 f_2;
    ax_bool f_3;
    ax_string f_4;
};

struct TypeEntry {
    ax_u8 f_0;
    ax_u8 f_1;
    ax_u16 f_2;
    ax_u32 f_3;
    ax_u32 f_4;
    ax_u32 f_5;
    ax_u32 f_6;
};

struct StructField {
    ax_u32 f_0;
    ax_u32 f_1;
};

struct StructFieldVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct StructInfo {
    StructFieldVec* f_0;
};

struct StructInfoVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct VariantInfo {
    ax_u32 f_0;
    ax_u32 f_1;
    ax_u8 f_2;
    ax_u8 f_3;
    ax_u16 f_4;
};

struct VariantInfoVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct SumInfo {
    VariantInfoVec* f_0;
};

struct SumInfoVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct FuncInfo {
    U32Vec* f_0;
    ax_u32 f_1;
    ax_u32 f_2;
};

struct FuncInfoVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct TypeEntryVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct TypeTable {
    TypeEntryVec* f_0;
    StructInfoVec* f_1;
    FuncInfoVec* f_2;
    SumInfoVec* f_3;
};

struct TypeSubst {
    ax_u32 f_0;
    ax_u32 f_1;
};

struct TypeSubstVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct Monomorphizer {
    AstTree* f_0;
    void* f_1;
    void* f_2;
    void* f_3;
};

struct TypeChecker {
    AstTree* f_0;
    void* f_1;
    void* f_2;
    void* f_3;
    void* f_4;
    ax_i64 f_5;
    ax_u32 f_6;
    ax_u32 f_7;
};

struct CGNode {
    ax_u32 f_0;
    ax_u32 f_1;
    ax_u32 f_2;
    ax_bool f_3;
    ax_u32 f_4;
};

struct CGNodeVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct CGEdge {
    ax_u32 f_0;
    ax_u32 f_1;
    ax_u32 f_2;
};

struct CGEdgeVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct U32VecVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct ConnectionGraph {
    CGNodeVec* f_0;
    CGEdgeVec* f_1;
    U32VecVec* f_2;
    U32VecVec* f_3;
    U32Vec* f_4;
};

struct OwnershipChecker {
    AstTree* f_0;
    void* f_1;
    void* f_2;
    void* f_3;
    ax_u32 f_4;
};

struct EscapeAnalyser {
    AstTree* f_0;
    void* f_1;
    void* f_2;
    void* f_3;
    ConnectionGraph* f_4;
    ax_u32 f_5;
};

struct CtgcInjector {
    AstTree* f_0;
    void* f_1;
    void* f_2;
    void* f_3;
    U32Vec* f_4;
    U32Vec* f_5;
};

struct AliasReuseOptimizer {
    AstTree* f_0;
    void* f_1;
    void* f_2;
    void* f_3;
};

struct AirInst {
    ax_u16 f_0;
    ax_u16 f_1;
    ax_u32 f_2;
    ax_u32 f_3;
    ax_u32 f_4;
};

struct AirInstVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct BasicBlock {
    ax_u32 f_0;
    ax_u32 f_1;
    ax_u32 f_2;
    ax_u32 f_3;
    ax_u32 f_4;
    ax_u32 f_5;
    ax_u32 f_6;
    ax_u8 f_7;
    ax_bool f_8;
    ax_bool f_9;
};

struct BasicBlockVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct AirFunc {
    ax_u32 f_0;
    ax_u32 f_1;
    U32Vec* f_2;
    ax_u32 f_3;
    BasicBlockVec* f_4;
    AirInstVec* f_5;
    U32Vec* f_6;
    ax_bool f_7;
    ax_bool f_8;
    U32Vec* f_9;
    U32Vec* f_10;
    U32Vec* f_11;
};

struct AirFuncVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct AirModule {
    AirFuncVec* f_0;
};

struct AirFuncBuilder {
    ax_u32 f_0;
    ax_u32 f_1;
    BasicBlockVec* f_2;
    AirInstVec* f_3;
    U32Vec* f_4;
    ax_i32 f_5;
    ax_u32 f_6;
    U32Vec* f_7;
    U32Vec* f_8;
    U32Vec* f_9;
};

struct LocalMapEntry {
    ax_u32 f_0;
    ax_u32 f_1;
};

struct LocalMap {
    void* f_0;
    ax_u32 f_1;
    ax_u32 f_2;
};

struct AirModuleBuilder {
    AstTree* f_0;
    void* f_1;
    void* f_2;
    void* f_3;
    AirModule* f_4;
    void* f_5;
};

struct FuncLowering {
    void* f_0;
    AirFuncBuilder* f_1;
    LocalMap* f_2;
    U32Vec* f_3;
    ax_bool f_4;
};

struct ConstVal {
    ax_bool f_0;
    ax_u32 f_1;
};

struct SsaOptimizer {
    ax_u8 f_0;
};

struct CGenerator {
    AirModule* f_0;
    AstTree* f_1;
    void* f_2;
    void* f_3;
    void* f_4;
    void* f_5;
    void* f_6;
};

struct WasmGenerator {
    AirModule* f_0;
    AstTree* f_1;
    void* f_2;
    void* f_3;
    void* f_4;
    void* f_5;
    void* f_6;
};

struct MachOperand {
    ax_u8 f_0;
    ax_u8 f_1;
    ax_u16 f_2;
    ax_u32 f_3;
    ax_u32 f_4;
    ax_i64 f_5;
};

struct MachInst {
    ax_u16 f_0;
    ax_u8 f_1;
    ax_u8 f_2;
    MachOperand* f_3;
    MachOperand* f_4;
    MachOperand* f_5;
};

struct MachInstVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct InstructionSelector {
    void* f_0;
    ax_string f_1;
    void* f_2;
    ax_i32 f_3;
    void* f_4;
    void* f_5;
    ax_u32 f_6;
    ax_u32 f_7;
};

struct LiveInterval {
    ax_u32 f_0;
    ax_i32 f_1;
    ax_i32 f_2;
};

struct LiveIntervalVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct RegAllocation {
    ax_u32 f_0;
    ax_u8 f_1;
    ax_bool f_2;
    ax_i32 f_3;
};

struct RegAllocResult {
    void* f_0;
    ax_u32 f_1;
    ax_i32 f_2;
};

struct StackFrame {
    void* f_0;
    ax_i64 f_1;
    ax_i32 f_2;
    ax_i32 f_3;
    ax_i32 f_4;
    ax_i32 f_5;
};

struct ByteVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct Relocation {
    ax_i64 f_0;
    ax_u8 f_1;
    ax_u8 f_2;
    ax_u16 f_3;
    ax_u32 f_4;
    ax_i64 f_5;
};

struct RelocationVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct Fixup {
    ax_i64 f_0;
    ax_u32 f_1;
    ax_i32 f_2;
};

struct FixupVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct LabelMap {
    void* f_0;
    void* f_1;
    ax_i64 f_2;
    ax_i64 f_3;
};

struct MachEmitter {
    ByteVec* f_0;
    RelocationVec* f_1;
    LabelMap* f_2;
    FixupVec* f_3;
};

struct ELF64Sym {
    ax_string f_0;
    ax_u64 f_1;
    ax_u64 f_2;
    ax_u8 f_3;
    ax_u8 f_4;
    ax_u16 f_5;
};

struct ELF64SymVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct COFFReloc {
    ax_u32 f_0;
    ax_u32 f_1;
    ax_u16 f_2;
};

struct COFFRelocVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct COFFSymbol {
    ax_string f_0;
    ax_u32 f_1;
    ax_i16 f_2;
    ax_u16 f_3;
    ax_u8 f_4;
    ax_u8 f_5;
};

struct COFFSymbolVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct CompiledFuncInfo {
    ax_string f_0;
    ax_u32 f_1;
    ax_u32 f_2;
    ax_bool f_3;
    RelocationVec* f_4;
};

struct CompiledFuncInfoVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct LinkerSymbol {
    ax_string f_0;
    ax_i64 f_1;
    ax_u64 f_2;
    ax_u64 f_3;
    ax_bool f_4;
};

struct LinkerSymbolVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct ParsedReloc {
    ax_i64 f_0;
    ax_u32 f_1;
    ax_bool f_2;
    ax_i64 f_3;
};

struct ParsedRelocVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct LinkerStrVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct ParsedObject {
    ByteVec* f_0;
    ByteVec* f_1;
    LinkerSymbolVec* f_2;
    LinkerStrVec* f_3;
    ParsedRelocVec* f_4;
    ax_u64 f_5;
};

struct ParsedObjectPtrVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct AxiomLinker {
    LinkerStrVec* f_0;
    ax_string f_1;
};

struct FmtToken {
    ax_u8 f_0;
    ax_string f_1;
};

struct FmtTokenVec {
    void* f_0;
    ax_i64 f_1;
    ax_i64 f_2;
};

struct Formatter {
    ax_i64 f_0;
};

// Function Prototypes
void* ax_get_str_ptr(ax_string p_0);
void ax_print_raw(ax_string p_0);
void ax_print_raw_ptr(void* p_0, ax_i64 p_1);
ax_i32 ax_puts_local(ax_string p_0);
void ax_print_i64_raw(ax_i64 p_0);
ax_i32 ax_printf_local(ax_string p_0, ax_i64 p_1, ax_i64 p_2, ax_i64 p_3, ax_i64 p_4, ax_i64 p_5, ax_i64 p_6, ax_i64 p_7, ax_i64 p_8);
ax_i32 ax_snprintf_local(void* p_0, ax_i64 p_1, void* p_2, void* p_3, void* p_4);
void ax_print_to_file(void* p_0, ax_string p_1);
void ax_print_i64_to_file(void* p_0, ax_i64 p_1);
ax_i32 ax_fprintf_local(void* p_0, ax_string p_1, ax_i64 p_2, ax_i64 p_3, ax_i64 p_4, ax_i64 p_5, ax_i64 p_6, ax_i64 p_7, ax_i64 p_8, ax_i64 p_9);
ax_i32 ax_sprintf_local(void* p_0, void* p_1, ax_f64 p_2);
TokenVec* ax_new_token_vec();
void ax_TokenVec_push(TokenVec* p_0, Token* p_1);
IntVec* ax_new_int_vec();
void ax_IntVec_push(IntVec* p_0, ax_i32 p_1);
ax_i32 ax_IntVec_pop(IntVec* p_0);
ax_i32 ax_IntVec_top(IntVec* p_0);
Lexer* ax_new_lexer(ax_string p_0);
ax_u8 ax_Lexer_peek1(Lexer* p_0);
void ax_Lexer_emit(Lexer* p_0, ax_u8 p_1, ax_i64 p_2, ax_i64 p_3);
ax_u8 ax_lookup_keyword(ax_string p_0);
ax_bool ax_is_ident_start(ax_u8 p_0);
ax_bool ax_is_ident_continue(ax_u8 p_0);
void ax_Lexer_scan_ident(Lexer* p_0);
ax_bool ax_is_dec_digit(ax_u8 p_0);
ax_bool ax_is_hex_digit(ax_u8 p_0);
ax_bool ax_is_oct_digit(ax_u8 p_0);
ax_bool ax_is_bin_digit(ax_u8 p_0);
void ax_Lexer_scan_dec_digits(Lexer* p_0);
void ax_Lexer_scan_hex_digits(Lexer* p_0);
void ax_Lexer_scan_oct_digits(Lexer* p_0);
void ax_Lexer_scan_bin_digits(Lexer* p_0);
void ax_Lexer_scan_number(Lexer* p_0);
void ax_Lexer_scan_string(Lexer* p_0);
void ax_Lexer_scan_char(Lexer* p_0);
void ax_Lexer_scan_line_comment(Lexer* p_0);
void ax_Lexer_scan_operator_or_punct(Lexer* p_0);
void ax_Lexer_run(Lexer* p_0);
ax_i64 ax_find_line_start(ax_string p_0, ax_i64 p_1);
ax_i64 ax_Lexer_find_next_content_offset(Lexer* p_0, ax_i64 p_1);
ax_i64 ax_count_leading_spaces(ax_string p_0, ax_i64 p_1, ax_i64 p_2);
TokenVec* ax_Lexer_process_indentation(Lexer* p_0);
TokenVec* ax_Lexer_tokenize(Lexer* p_0);
NodeVec* ax_new_node_vec();
ax_u32 ax_NodeVec_push(NodeVec* p_0, AstNode* p_1);
AstTree* ax_new_ast_tree(ax_string p_0, TokenVec* p_1);
ax_u32 ax_AstTree_add_node(AstTree* p_0, ax_u8 p_1, ax_u32 p_2);
ax_u32 ax_AstTree_add_extra(AstTree* p_0, ax_i32 p_1);
ax_u32 ax_AstTree_clone_subtree(AstTree* p_0, ax_u32 p_1);
ax_u32 ax_fnv1a(ax_string p_0);
InternPool* ax_new_intern_pool();
ax_string ax_alloc_str_from_raw(void* p_0, ax_i64 p_1);
ax_string ax_InternPool_get_str(void* p_0, ax_u32 p_1);
ax_string ax_InternPool_get(void* p_0, ax_u32 p_1);
ax_u32 ax_InternPool_intern(void* p_0, ax_string p_1);
ax_u32 ax_InternPool_intern_string(void* p_0, ax_string p_1);
ax_u32 ax_InternPool_insert_at(void* p_0, ax_i64 p_1, ax_string p_2, ax_u32 p_3);
void ax_InternPool_grow_table(void* p_0);
void ax_InternPool_free_pool(void* p_0);
Parser* ax_TokenVec_new_parser(TokenVec* p_0, ax_string p_1, void* p_2);
Token* ax_Parser_peek(Parser* p_0);
Token* ax_Parser_peek_raw(Parser* p_0);
Token* ax_Parser_peek_at(Parser* p_0, ax_i64 p_1);
Token* ax_Parser_consume(Parser* p_0);
ax_bool ax_Parser_check(Parser* p_0, ax_u8 p_1);
ax_bool ax_Parser_check_raw(Parser* p_0, ax_u8 p_1);
void ax_Parser_errorf(Parser* p_0, Token* p_1, ax_string p_2);
Token* ax_Parser_expect(Parser* p_0, ax_u8 p_1);
void ax_Parser_expect_newline(Parser* p_0);
ax_u32 ax_Parser_token_idx(Parser* p_0, Token* p_1);
ax_string ax_Parser_token_text(Parser* p_0, Token* p_1);
void ax_Parser_append_child(Parser* p_0, ax_u32 p_1, ax_u32 p_2);
void ax_Parser_set_payload(Parser* p_0, ax_u32 p_1, ax_u32 p_2);
void ax_Parser_set_flags(Parser* p_0, ax_u32 p_1, ax_u16 p_2);
ax_i32 ax_left_binding_power(ax_u8 p_0);
ax_u32 ax_Parser_parse_expr_with_prec(Parser* p_0, ax_i32 p_1);
ax_u32 ax_Parser_parse_nud(Parser* p_0);
ax_u32 ax_Parser_parse_led(Parser* p_0, ax_u32 p_1, Token* p_2, ax_i32 p_3);
ax_u32 ax_Parser_parse_call_args(Parser* p_0, ax_u32 p_1, Token* p_2);
ax_u32 ax_Parser_parse_stmt(Parser* p_0);
ax_u32 ax_Parser_parse_break_stmt(Parser* p_0);
ax_u32 ax_Parser_parse_continue_stmt(Parser* p_0);
ax_u32 ax_Parser_parse_var_decl(Parser* p_0);
ax_u32 ax_Parser_parse_return_stmt(Parser* p_0);
ax_u32 ax_Parser_parse_if_stmt(Parser* p_0);
ax_u32 ax_Parser_parse_while_loop(Parser* p_0);
ax_u32 ax_Parser_parse_for_loop(Parser* p_0);
ax_u32 ax_Parser_parse_match_stmt(Parser* p_0);
ax_u32 ax_Parser_parse_match_arm(Parser* p_0);
ax_u32 ax_Parser_parse_pattern(Parser* p_0);
ax_u32 ax_Parser_parse_expr_stmt(Parser* p_0);
ax_u32 ax_Parser_parse_block(Parser* p_0);
ax_u32 ax_Parser_parse_type_expr(Parser* p_0);
ax_u32 ax_Parser_parse_generic_params(Parser* p_0);
ax_u32 ax_Parser_parse_generic_param(Parser* p_0);
ax_u32 ax_Parser_parse_func_decl(Parser* p_0, ax_bool p_1);
ax_u32 ax_Parser_parse_struct_decl(Parser* p_0, ax_bool p_1);
ax_u32 ax_Parser_parse_field_decl(Parser* p_0, ax_bool p_1);
ax_u32 ax_Parser_parse_type_alias_decl(Parser* p_0, ax_bool p_1);
ax_u32 ax_Parser_parse_type_variant(Parser* p_0);
void ax_Parser_parse_program(Parser* p_0);
ax_u32 ax_Parser_parse_param(Parser* p_0);
ax_u32 ax_Parser_parse_const_decl(Parser* p_0, ax_bool p_1);
ax_u32 ax_Parser_parse_extern_decl(Parser* p_0, ax_bool p_1);
ax_u32 ax_Parser_parse_import_decl(Parser* p_0);
void ax_Scope_init_scope(void* p_0, ax_u8 p_1, ax_u32 p_2, ax_u32 p_3, ax_u32 p_4);
ax_u32 ax_hash_fnv1a(ax_u32 p_0);
void ax_Scope_scope_put(void* p_0, ax_u32 p_1, ax_u32 p_2);
void ax_Scope_scope_insert(void* p_0, ax_u32 p_1, ax_u32 p_2);
ScopeEntry* ax_Scope_scope_get(void* p_0, ax_u32 p_1);
void ax_Scope_scope_grow(void* p_0);
SymbolVec* ax_new_symbol_vec();
ax_u32 ax_SymbolVec_push(void* p_0, Symbol* p_1);
ScopeVec* ax_new_scope_vec();
ax_u32 ax_ScopeVec_push(void* p_0, Scope* p_1);
U32Vec* ax_new_u32_vec();
ax_u32 ax_U32Vec_push(void* p_0, ax_u32 p_1);
ax_u32 ax_U32Vec_pop(void* p_0);
ax_u32 ax_U32Vec_append_unique(void* p_0, ax_u32 p_1);
SymbolTable* ax_InternPool_new_symbol_table(void* p_0);
void ax_SymbolTable_define_builtin(void* p_0, ax_string p_1, ax_u32 p_2, void* p_3);
ax_u32 ax_SymbolTable_push_scope(void* p_0, ax_u8 p_1);
void ax_SymbolTable_pop_scope(void* p_0);
ax_u32 ax_SymbolTable_current_scope(void* p_0);
ax_u32 ax_SymbolTable_current_depth(void* p_0);
ax_u32 ax_SymbolTable_define(void* p_0, ax_u32 p_1, ax_u8 p_2, ax_u16 p_3, ax_u32 p_4);
ax_u32 ax_SymbolTable_resolve(void* p_0, ax_u32 p_1);
ax_u32 ax_SymbolTable_resolve_type(void* p_0, ax_u32 p_1);
NameResolver* ax_AstTree_new_name_resolver(AstTree* p_0, void* p_1, void* p_2);
void ax_NameResolver_resolve(void* p_0);
void ax_NameResolver_resolve_children(void* p_0, ax_u32 p_1);
void ax_NameResolver_resolve_node(void* p_0, ax_u32 p_1);
void ax_SymbolTable_free_symtable(void* p_0);
void ax_StructFieldVec_push(void* p_0, StructField* p_1);
void ax_StructInfoVec_push(void* p_0, StructInfo* p_1);
void ax_VariantInfoVec_push(void* p_0, VariantInfo* p_1);
void ax_SumInfoVec_push(void* p_0, SumInfo* p_1);
void ax_FuncInfoVec_push(void* p_0, FuncInfo* p_1);
ax_u32 ax_TypeEntryVec_push(void* p_0, TypeEntry* p_1);
SumInfoVec* ax_new_sum_info_vec();
VariantInfoVec* ax_new_variant_info_vec();
void* ax_new_type_table();
ax_u32 ax_TypeTable_register_sum_type(void* p_0, ax_u32 p_1, VariantInfoVec* p_2);
ax_u32 ax_TypeTable_register_struct(void* p_0, ax_u32 p_1, StructFieldVec* p_2);
ax_u32 ax_TypeTable_register_function(void* p_0, U32Vec* p_1, ax_u32 p_2);
ax_u32 ax_TypeTable_register_pointer(void* p_0, ax_u32 p_1);
ax_u32 ax_TypeTable_register_slice(void* p_0, ax_u32 p_1);
ax_u32 ax_TypeTable_register_array(void* p_0, ax_u32 p_1, ax_u32 p_2);
void ax_TypeTable_free_typetable(void* p_0);
TypeSubstVec* ax_new_type_subst_vec();
void ax_TypeSubstVec_push(TypeSubstVec* p_0, TypeSubst* p_1);
ax_u32 ax_TypeSubstVec_lookup_subst(TypeSubstVec* p_0, ax_u32 p_1);
Monomorphizer* ax_AstTree_new_monomorphizer(AstTree* p_0, void* p_1, void* p_2, void* p_3);
void ax_Monomorphizer_substitute_type_params(Monomorphizer* p_0, ax_u32 p_1, TypeSubstVec* p_2);
void ax_Monomorphizer_remove_generic_params_child(Monomorphizer* p_0, ax_u32 p_1);
ax_string ax_Monomorphizer_mangle_name(Monomorphizer* p_0, ax_string p_1, U32Vec* p_2);
ax_u32 ax_Monomorphizer_instantiate_function(Monomorphizer* p_0, ax_u32 p_1, U32Vec* p_2);
ax_string ax_strip_quotes(ax_string p_0);
ax_string ax_format_int(ax_i64 p_0);
ax_string ax_format_float(ax_f64 p_0);
TypeChecker* ax_AstTree_new_type_checker(AstTree* p_0, void* p_1, void* p_2, void* p_3);
void ax_TypeChecker_free_type_checker(void* p_0);
void ax_TypeChecker_set_node_type(void* p_0, ax_u32 p_1, ax_u32 p_2);
ax_i64 ax_parse_comptime_int(ax_string p_0);
ax_string ax_TypeChecker_node_text(void* p_0, ax_u32 p_1);
ComptimeValue* ax_TypeChecker_eval_comptime(void* p_0, ax_u32 p_1);
void ax_TypeChecker_run_type_checker(void* p_0);
void ax_TypeChecker_pre_infer_type_alias(void* p_0, ax_u32 p_1);
void ax_TypeChecker_pre_infer_struct(void* p_0, ax_u32 p_1);
void ax_TypeChecker_pre_infer_func_signature(void* p_0, ax_u32 p_1);
ax_u32 ax_TypeChecker_infer_node(void* p_0, ax_u32 p_1, ax_u32 p_2);
CGNodeVec* ax_new_cg_node_vec();
ax_u32 ax_CGNodeVec_push(CGNodeVec* p_0, CGNode* p_1);
CGEdgeVec* ax_new_cg_edge_vec();
ax_u32 ax_CGEdgeVec_push(CGEdgeVec* p_0, CGEdge* p_1);
U32VecVec* ax_new_u32_vec_vec();
ax_u32 ax_U32VecVec_push(U32VecVec* p_0, U32Vec* p_1);
ConnectionGraph* ax_new_connection_graph();
void ax_ConnectionGraph_free_connection_graph(ConnectionGraph* p_0);
void ax_ConnectionGraph_ensure_adj_capacity(ConnectionGraph* p_0, ax_u32 p_1);
ax_u32 ax_ConnectionGraph_add_value_node(ConnectionGraph* p_0, ax_u32 p_1, ax_u32 p_2, ax_u32 p_3);
ax_u32 ax_ConnectionGraph_add_ref_node(ConnectionGraph* p_0, ax_u32 p_1);
void ax_ConnectionGraph_add_edge(ConnectionGraph* p_0, ax_u32 p_1, ax_u32 p_2, ax_u32 p_3);
ax_u32 ax_ConnectionGraph_node_of_sym(ConnectionGraph* p_0, ax_u32 p_1);
ax_bool ax_ConnectionGraph_escapes(ConnectionGraph* p_0, ax_u32 p_1);
ax_bool ax_ConnectionGraph_escape_dfs(ConnectionGraph* p_0, ax_u32 p_1, void* p_2);
OwnershipChecker* ax_AstTree_new_ownership_checker(AstTree* p_0, void* p_1, void* p_2, void* p_3);
void ax_OwnershipChecker_check(OwnershipChecker* p_0);
void ax_OwnershipChecker_check_node(OwnershipChecker* p_0, ax_u32 p_1);
void ax_OwnershipChecker_check_move(OwnershipChecker* p_0, ax_u32 p_1);
EscapeAnalyser* ax_AstTree_new_escape_analyser(AstTree* p_0, void* p_1, void* p_2, void* p_3);
void ax_EscapeAnalyser_run(EscapeAnalyser* p_0);
void ax_EscapeAnalyser_traverse_nodes(EscapeAnalyser* p_0, ax_u32 p_1);
void ax_EscapeAnalyser_analyze_block(EscapeAnalyser* p_0, ax_u32 p_1);
void ax_EscapeAnalyser_analyze_stmt(EscapeAnalyser* p_0, ax_u32 p_1);
void ax_EscapeAnalyser_analyze_expr(EscapeAnalyser* p_0, ax_u32 p_1, ax_u32 p_2);
CtgcInjector* ax_AstTree_new_ctgc_injector(AstTree* p_0, void* p_1, void* p_2, void* p_3);
void ax_CtgcInjector_run(CtgcInjector* p_0);
void ax_CtgcInjector_traverse_and_inject(CtgcInjector* p_0, ax_u32 p_1, ax_u32 p_2);
void ax_CtgcInjector_append_child_node(CtgcInjector* p_0, ax_u32 p_1, ax_u32 p_2);
void ax_CtgcInjector_insert_before(CtgcInjector* p_0, ax_u32 p_1, ax_u32 p_2, ax_u32 p_3);
AliasReuseOptimizer* ax_AstTree_new_alias_reuse_optimizer(AstTree* p_0, void* p_1, void* p_2, void* p_3);
void ax_AliasReuseOptimizer_run(AliasReuseOptimizer* p_0);
void ax_AliasReuseOptimizer_optimize_node(AliasReuseOptimizer* p_0, ax_u32 p_1);
AirInstVec* ax_new_air_inst_vec();
ax_u32 ax_AirInstVec_push(AirInstVec* p_0, AirInst* p_1);
BasicBlockVec* ax_new_basic_block_vec();
ax_u32 ax_BasicBlockVec_push(BasicBlockVec* p_0, BasicBlock* p_1);
void ax_AirFunc_free_air_func(AirFunc* p_0);
AirFuncVec* ax_new_air_func_vec();
ax_u32 ax_AirFuncVec_push(AirFuncVec* p_0, AirFunc* p_1);
void ax_AirModule_free_air_module(AirModule* p_0);
AirFuncBuilder* ax_new_air_func_builder(ax_u32 p_0, ax_u32 p_1);
ax_u32 ax_AirFuncBuilder_new_block(AirFuncBuilder* p_0);
void ax_AirFuncBuilder_switch_to(AirFuncBuilder* p_0, ax_u32 p_1);
ax_u32 ax_AirFuncBuilder_current_block(AirFuncBuilder* p_0);
ax_u32 ax_AirFuncBuilder_emit(AirFuncBuilder* p_0, AirInst* p_1);
ax_u32 ax_AirFuncBuilder_emit_extra(AirFuncBuilder* p_0, ax_u32 p_1);
void ax_AirFuncBuilder_set_extra(AirFuncBuilder* p_0, ax_u32 p_1, ax_u32 p_2);
ax_u32 ax_AirFuncBuilder_fresh_reg(AirFuncBuilder* p_0);
void ax_AirFuncBuilder_add_edge(AirFuncBuilder* p_0, ax_u32 p_1, ax_u32 p_2);
AirFunc* ax_AirFuncBuilder_build_func(AirFuncBuilder* p_0);
ax_string ax_opcode_mnemonic(ax_u16 p_0);
ax_u16 ax_opcode_class(ax_u16 p_0);
ax_bool ax_opcode_is_binary_alu(ax_u16 p_0);
LocalMap* ax_new_local_map();
void ax_LocalMap_free_local_map(LocalMap* p_0);
ax_u32 ax_local_map_hash(ax_u32 p_0);
void ax_LocalMap_local_map_put(LocalMap* p_0, ax_u32 p_1, ax_u32 p_2);
void ax_LocalMap_local_map_insert(LocalMap* p_0, ax_u32 p_1, ax_u32 p_2);
ax_u32 ax_LocalMap_local_map_get(LocalMap* p_0, ax_u32 p_1);
void ax_LocalMap_local_map_grow(LocalMap* p_0);
AirModuleBuilder* ax_AstTree_new_air_module_builder(AstTree* p_0, void* p_1, void* p_2, void* p_3, void* p_4);
void ax_AirModuleBuilder_free_air_module_builder(void* p_0);
ax_string ax_AirModuleBuilder_get_token_text(void* p_0, ax_u32 p_1);
ax_i64 ax_parse_int_from_str(ax_string p_0);
ax_u16 ax_map_binary_op(ax_string p_0);
FuncLowering* ax_AirModuleBuilder_new_func_lowering(void* p_0, AirFuncBuilder* p_1, U32Vec* p_2);
void ax_FuncLowering_free_func_lowering(FuncLowering* p_0);
void ax_FuncLowering_register_params(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_TypeTable_builder_type_size_and_align(void* p_0, ax_u32 p_1, void* p_2, void* p_3);
ax_u32 ax_FuncLowering_try_lower_sizeof(FuncLowering* p_0, ax_u32 p_1);
ax_u32 ax_FuncLowering_try_lower_intrinsic(FuncLowering* p_0, ax_u32 p_1);
ax_u32 ax_FuncLowering_lower_expr(FuncLowering* p_0, ax_u32 p_1);
ax_u32 ax_FuncLowering_lower_int_lit(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_float_lit(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_bool_lit(FuncLowering* p_0, AstNode* p_1);
ax_u32 ax_FuncLowering_lower_nil_lit(FuncLowering* p_0);
ax_u32 ax_FuncLowering_lower_string_lit(FuncLowering* p_0, AstNode* p_1);
ax_u32 ax_FuncLowering_lower_char_lit(FuncLowering* p_0, AstNode* p_1);
ax_u32 ax_FuncLowering_lower_ident(FuncLowering* p_0, AstNode* p_1);
ax_u32 ax_FuncLowering_lower_binary_expr(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_unary_expr(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_call_expr(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_field_expr(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_index_expr(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_cast_expr(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_deref_expr(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_spawn_expr(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_await_expr(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_struct_lit(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_array_lit(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
ax_u32 ax_FuncLowering_lower_struct_constructor_call(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2, ax_u32 p_3);
ax_u32 ax_FuncLowering_emit_heap_alloc(FuncLowering* p_0, ax_u16 p_1);
ax_u32 ax_FuncLowering_emit_deref(FuncLowering* p_0, ax_u32 p_1, ax_u16 p_2);
ax_u32 ax_FuncLowering_emit_move(FuncLowering* p_0, ax_u32 p_1, ax_u16 p_2);
void ax_FuncLowering_emit_free(FuncLowering* p_0, ax_u32 p_1);
ax_u32 ax_FuncLowering_emit_arena_alloc(FuncLowering* p_0, ax_u32 p_1, ax_u16 p_2);
ax_u32 ax_FuncLowering_emit_alias_reuse(FuncLowering* p_0, ax_u32 p_1, ax_u16 p_2);
ax_u32 ax_FuncLowering_lower_ownership_aware(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2, ax_u32 p_3);
void ax_FuncLowering_lower_alias_stmt(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_lower_block(FuncLowering* p_0, ax_u32 p_1);
void ax_FuncLowering_lower_stmt(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_lower_var_decl(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_lower_assign(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_lower_return(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_lower_if(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_lower_while(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_lower_for(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_lower_destroy(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_lower_defer(FuncLowering* p_0, ax_u32 p_1, AstNode* p_2);
void ax_FuncLowering_ensure_return(FuncLowering* p_0);
void* ax_AirModuleBuilder_lower_func(void* p_0, ax_u32 p_1, AstNode* p_2);
void ax_AirModuleBuilder_build_module(void* p_0);
void* ax_AirModuleBuilder_builder_str_to_null_terminated(void* p_0, ax_string p_1);
SsaOptimizer* ax_new_ssa_optimizer();
ax_u32 ax_AirFunc_max_reg_id(void* p_0);
ax_bool ax_is_unary_foldable(ax_u16 p_0);
ax_bool ax_has_side_effect(ax_u16 p_0);
ax_bool ax_opcode_is_control(ax_u16 p_0);
ax_bool ax_eval_binary(ax_u16 p_0, ax_u32 p_1, ax_u32 p_2, void* p_3);
ax_bool ax_eval_unary(ax_u16 p_0, ax_u32 p_1, void* p_2);
ax_bool ax_AirFunc_fold_func(void* p_0);
ax_bool ax_AirFunc_copy_prop_func(void* p_0);
ax_bool ax_AirFunc_remove_unreachable_blocks(void* p_0);
ax_bool ax_AirFunc_dce_func(void* p_0);
ax_bool ax_AirFunc_licm_func(void* p_0);
ax_u32 ax_AirFunc_insert_inst_at(void* p_0, ax_u32 p_1, ax_u32 p_2, AirInst* p_3);
ax_bool ax_AirFunc_compute_loop_depths(void* p_0);
ax_bool ax_AirFunc_strength_reduction_func(void* p_0);
ax_bool ax_AirFunc_loop_unroll_func(void* p_0);
void ax_SsaOptimizer_run(SsaOptimizer* p_0, void* p_1);
ax_bool ax_opcode_defines_dest(ax_u16 p_0);
CGenerator* ax_AirModule_new_c_generator(AirModule* p_0, AstTree* p_1, void* p_2, void* p_3, void* p_4);
ax_string ax_CGenerator_get_c_type(CGenerator* p_0, ax_u32 p_1);
ax_bool ax_CGenerator_is_stdlib_func(CGenerator* p_0, ax_string p_1);
ax_string ax_CGenerator_get_mangled_name_by_sym(CGenerator* p_0, ax_u32 p_1);
void* ax_str_to_null_terminated(ax_string p_0);
ax_bool ax_CGenerator_generate(CGenerator* p_0, ax_string p_1);
void ax_CGenerator_translate_inst(CGenerator* p_0, AirFunc* p_1, AirInst* p_2);
ax_bool ax_is_type_unsigned(ax_u32 p_0);
ax_string ax_TypeTable_map_wasm_type(void* p_0, ax_u32 p_1);
ax_bool ax_wasm_opcode_defines_dest(ax_u16 p_0);
WasmGenerator* ax_AirModule_new_wasm_generator(AirModule* p_0, AstTree* p_1, void* p_2, void* p_3, void* p_4);
ax_string ax_WasmGenerator_resolve_sym_name(WasmGenerator* p_0, ax_u32 p_1, ax_u32 p_2);
void* ax_wasm_str_to_null_terminated(ax_string p_0);
ax_bool ax_WasmGenerator_generate(WasmGenerator* p_0, ax_string p_1);
void ax_WasmGenerator_compile_func(WasmGenerator* p_0, AirFunc* p_1);
void ax_WasmGenerator_lower_inst(WasmGenerator* p_0, AirFunc* p_1, AirInst* p_2);
ax_bool ax_reg_is_gpr(ax_u8 p_0);
ax_u8 ax_reg_hw_reg(ax_u8 p_0);
ax_bool ax_reg_needs_rex(ax_u8 p_0);
ax_string ax_reg_to_str(ax_u8 p_0);
ax_u8 ax_get_sysv_arg_reg(ax_i64 p_0);
ax_u8 ax_get_win64_arg_reg(ax_i64 p_0);
ax_bool ax_reg_is_sysv_caller_saved(ax_u8 p_0);
ax_bool ax_reg_is_sysv_callee_saved(ax_u8 p_0);
ax_bool ax_reg_is_win64_callee_saved(ax_u8 p_0);
ax_string ax_x86_resolve_sym_name(ax_u32 p_0, void* p_1, void* p_2, void* p_3);
ax_string ax_cond_code_to_str(ax_u8 p_0);
MachInstVec* ax_new_mach_inst_vec();
ax_u32 ax_MachInstVec_push(MachInstVec* p_0, MachInst* p_1);
void ax_TypeTable_type_size_and_align(void* p_0, ax_u32 p_1, void* p_2, void* p_3);
ax_u32 ax_TypeTable_field_offset(void* p_0, ax_u32 p_1, ax_u32 p_2);
ax_u32 ax_TypeTable_field_size(void* p_0, ax_u32 p_1, ax_u32 p_2);
ax_u8 ax_abi_int_arg_reg(ax_string p_0, ax_i64 p_1);
ax_u8 ax_abi_return_reg(ax_string p_0);
ax_u32 ax_InstructionSelector_next_vreg(void* p_0);
ax_u32 ax_InstructionSelector_next_label(void* p_0);
ax_u32 ax_InstructionSelector_get_register_type(void* p_0, ax_u32 p_1);
void ax_AirInst_select_cmp(void* p_0, ax_u8 p_1, void* p_2);
void ax_InstructionSelector_emit_block_copy_ext(void* p_0, ax_u32 p_1, ax_u32 p_2, ax_u32 p_3, ax_u32 p_4, ax_u32 p_5, void* p_6);
void ax_InstructionSelector_emit_block_copy(void* p_0, ax_u32 p_1, ax_u32 p_2, ax_u32 p_3, ax_u32 p_4, void* p_5);
void ax_InstructionSelector_select_inst(void* p_0, void* p_1, void* p_2);
MachInstVec* ax_AirFunc_select_all(void* p_0, ax_string p_1, void* p_2, void* p_3, void* p_4);
LiveIntervalVec* ax_new_live_interval_vec();
ax_u32 ax_LiveIntervalVec_push(LiveIntervalVec* p_0, LiveInterval* p_1);
ax_bool ax_is_two_operand_read(ax_u16 p_0);
LiveIntervalVec* ax_MachInst_compute_liveness(void* p_0, ax_i64 p_1);
void* ax_get_allocatable_gprs(void* p_0);
ax_bool ax_reg_is_caller_saved(ax_string p_0, ax_u8 p_1);
RegAllocResult* ax_LiveIntervalVec_graph_coloring_alloc(LiveIntervalVec* p_0, void* p_1, ax_i64 p_2, void* p_3, ax_i64 p_4, ax_string p_5);
void* ax_get_used_callee_saved(ax_string p_0, void* p_1, ax_u32 p_2, void* p_3);
StackFrame* ax_compute_frame(void* p_0, ax_i64 p_1, ax_i32 p_2, ax_i32 p_3, ax_string p_4);
void ax_StackFrame_emit_prologue(void* p_0, void* p_1);
void ax_StackFrame_emit_epilogue(void* p_0, void* p_1);
ax_u8 ax_get_dst_behavior(ax_u16 p_0);
MachInstVec* ax_MachInst_insert_spill_code(void* p_0, ax_i64 p_1, void* p_2, void* p_3);
ax_string ax_to_byte_reg(ax_u8 p_0);
ax_string ax_to_word_reg(ax_u8 p_0);
ax_string ax_to_dword_reg(ax_u8 p_0);
ax_string ax_MachOperand_format_operand(MachOperand* p_0, ax_string p_1, void* p_2);
void ax_emit_inst(void* p_0, MachInst* p_1, ax_string p_2, void* p_3, ax_string p_4, void* p_5, void* p_6, void* p_7);
void ax_emit_function(void* p_0, ax_string p_1, void* p_2, ax_i64 p_3, void* p_4, void* p_5, ax_string p_6, void* p_7, void* p_8, void* p_9);
void* ax_x86_str_to_null_terminated(ax_string p_0);
ax_bool ax_compile_native_asm(ax_string p_0, AirModule* p_1, void* p_2, void* p_3, void* p_4, ax_string p_5);
ByteVec* ax_new_byte_vec();
void ax_ByteVec_push_byte(void* p_0, ax_u8 p_1);
void ax_ByteVec_push_bytes(void* p_0, void* p_1, ax_i64 p_2);
void ax_ByteVec_push_u16_le(void* p_0, ax_u16 p_1);
void ax_ByteVec_push_u32_le(void* p_0, ax_u32 p_1);
void ax_ByteVec_push_u64_le(void* p_0, ax_u64 p_1);
ax_u8 ax_encode_rex(ax_bool p_0, ax_bool p_1, ax_bool p_2, ax_bool p_3);
void ax_x86_encode_modrm_rr(ax_u8 p_0, ax_u8 p_1, void* p_2, void* p_3, void* p_4);
void ax_x86_encode_modrm_rm(ax_u8 p_0, ax_u8 p_1, ax_i32 p_2, void* p_3);
void ax_x86_encode_modrm_rip(ax_u8 p_0, ax_i32 p_1, void* p_2);
void ax_x86_encode_modrm_sib(ax_u8 p_0, ax_u8 p_1, ax_u8 p_2, ax_u8 p_3, ax_i32 p_4, void* p_5);
void ax_ByteVec_x86_encode_ret(void* p_0);
void ax_ByteVec_x86_encode_nop(void* p_0);
void ax_ByteVec_x86_encode_int3(void* p_0);
void ax_x86_encode_push(ax_u8 p_0, void* p_1);
void ax_x86_encode_pop(ax_u8 p_0, void* p_1);
void ax_x86_encode_mov_rr(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_mov_ri(ax_u8 p_0, ax_i32 p_1, void* p_2);
void ax_x86_encode_mov_ri64(ax_u8 p_0, ax_i64 p_1, void* p_2);
void ax_x86_encode_add_rr(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_add_ri(ax_u8 p_0, ax_i32 p_1, void* p_2);
void ax_x86_encode_sub_rr(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_sub_ri(ax_u8 p_0, ax_i32 p_1, void* p_2);
void ax_x86_encode_imul_rr(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_ByteVec_x86_encode_cqo(void* p_0);
void ax_x86_encode_idiv_r(ax_u8 p_0, void* p_1);
void ax_x86_encode_neg_r(ax_u8 p_0, void* p_1);
void ax_x86_encode_not_r(ax_u8 p_0, void* p_1);
void ax_x86_encode_cmp_rr(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_cmp_ri(ax_u8 p_0, ax_i32 p_1, void* p_2);
void ax_x86_encode_test_rr(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_and_rr(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_or_rr(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_xor_rr(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_shl_cl(ax_u8 p_0, void* p_1);
void ax_x86_encode_sar_cl(ax_u8 p_0, void* p_1);
void ax_x86_encode_xor_zero(ax_u8 p_0, void* p_1);
void ax_x86_encode_setcc(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_movzx_br(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_jmp_rel32(ax_i32 p_0, void* p_1);
void ax_x86_encode_jcc_rel32(ax_u8 p_0, ax_i32 p_1, void* p_2);
void ax_x86_encode_call_rel32(ax_i32 p_0, void* p_1);
void ax_x86_encode_call_r(ax_u8 p_0, void* p_1);
void ax_x86_encode_lea(ax_u8 p_0, ax_u8 p_1, ax_i32 p_2, void* p_3);
void ax_x86_encode_mov_load(ax_u8 p_0, ax_u8 p_1, ax_i32 p_2, void* p_3);
void ax_x86_encode_mov_store(ax_u8 p_0, ax_i32 p_1, ax_u8 p_2, void* p_3);
void ax_ByteVec_x86_encode_syscall(void* p_0);
void ax_x86_encode_lea_rip(ax_u8 p_0, ax_i32 p_1, void* p_2);
void ax_x86_encode_shl_imm(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_sar_imm(ax_u8 p_0, ax_u8 p_1, void* p_2);
void ax_x86_encode_mov_load_sized(ax_u8 p_0, ax_u8 p_1, ax_i32 p_2, ax_u8 p_3, void* p_4);
void ax_x86_encode_mov_store_sized(ax_u8 p_0, ax_i32 p_1, ax_u8 p_2, ax_u8 p_3, void* p_4);
RelocationVec* ax_new_relocation_vec();
void ax_RelocationVec_push_reloc(void* p_0, Relocation* p_1);
FixupVec* ax_new_fixup_vec();
void ax_FixupVec_push_fixup(void* p_0, Fixup* p_1);
LabelMap* ax_new_label_map();
void ax_LabelMap_label_map_set(void* p_0, ax_u32 p_1, ax_i64 p_2);
ax_bool ax_LabelMap_label_map_get(void* p_0, ax_u32 p_1, void* p_2);
MachEmitter* ax_new_mach_emitter();
void ax_MachEmitter_free_mach_emitter(void* p_0);
ax_u8 ax_MachOperand_emitter_resolve_reg(MachOperand* p_0, void* p_1);
void ax_MachEmitter_emit_mach_inst(void* p_0, MachInst* p_1, void* p_2, void* p_3, void* p_4, void* p_5);
void ax_MachEmitter_emitter_resolve_fixups(void* p_0);
void ax_MachEmitter_emit_function_binary(void* p_0, ax_string p_1, void* p_2, ax_i64 p_3, void* p_4, void* p_5, ax_bool p_6, void* p_7, void* p_8, ax_bool p_9);
ELF64SymVec* ax_new_elf64_sym_vec();
void ax_ELF64SymVec_push_elf64_sym(void* p_0, ELF64Sym* p_1);
void ax_ByteVec_elf64_serialize(void* p_0, void* p_1, void* p_2, void* p_3);
ax_bool ax_elf64_write_object_file(ax_string p_0, void* p_1, void* p_2, void* p_3);
COFFRelocVec* ax_new_coff_reloc_vec();
void ax_COFFRelocVec_push_coff_reloc(void* p_0, COFFReloc* p_1);
COFFSymbolVec* ax_new_coff_symbol_vec();
ax_i32 ax_COFFSymbolVec_push_coff_symbol(void* p_0, COFFSymbol* p_1);
void ax_ByteVec_coff_serialize(void* p_0, void* p_1, void* p_2, void* p_3, void* p_4);
ax_bool ax_coff_write_object_file(ax_string p_0, void* p_1, void* p_2, void* p_3, void* p_4);
CompiledFuncInfoVec* ax_new_compiled_func_info_vec();
void ax_CompiledFuncInfoVec_push_compiled_func_info(void* p_0, CompiledFuncInfo* p_1);
ax_string ax_resolve_binary_sym_name(ax_i64 p_0, ax_string p_1, void* p_2, void* p_3, void* p_4);
ax_bool ax_compile_native_binary(ax_string p_0, AirModule* p_1, void* p_2, void* p_3, void* p_4, ax_string p_5, ax_bool p_6);
ax_string ax_x86_slice_to_str(void* p_0, ax_i64 p_1);
ax_u16 ax_read_u16_le(void* p_0, ax_i64 p_1);
ax_u32 ax_read_u32_le(void* p_0, ax_i64 p_1);
ax_u64 ax_read_u64_le(void* p_0, ax_i64 p_1);
void ax_write_u32_le(void* p_0, ax_i64 p_1, ax_u32 p_2);
void ax_write_u64_le(void* p_0, ax_i64 p_1, ax_u64 p_2);
LinkerSymbolVec* ax_new_linker_symbol_vec();
void ax_LinkerSymbolVec_push_linker_symbol(void* p_0, LinkerSymbol* p_1);
ParsedRelocVec* ax_new_parsed_reloc_vec();
void ax_ParsedRelocVec_push_parsed_reloc(void* p_0, ParsedReloc* p_1);
LinkerStrVec* ax_new_linker_str_vec();
void ax_LinkerStrVec_push_linker_str(void* p_0, ax_string p_1);
ParsedObjectPtrVec* ax_new_parsed_object_ptr_vec();
void ax_ParsedObjectPtrVec_push_parsed_object_ptr(void* p_0, void* p_1);
void* ax_linker_parse_elf(void* p_0, ax_i64 p_1);
void* ax_linker_parse_coff(void* p_0, ax_i64 p_1);
AxiomLinker* ax_new_axiom_linker();
void ax_ByteVec_align_byte_vec(void* p_0, ax_i64 p_1);
ax_bool ax_is_valid_runtime_dll_symbol(ax_string p_0);
ax_string ax_get_dll_for_symbol(ax_string p_0);
void ax_LinkerStrVec_add_unique_import(void* p_0, ax_string p_1);
void ax_write_buf_u32_le(void* p_0, ax_i64 p_1, ax_u32 p_2);
void ax_write_buf_u64_le(void* p_0, ax_i64 p_1, ax_u64 p_2);
void ax_ByteVec_linker_build_pe_headers(void* p_0, ax_u32 p_1, ax_u32 p_2, ax_u32 p_3, ax_u32 p_4, ax_u32 p_5, ax_u32 p_6, ax_u32 p_7);
void ax_ByteVec_linker_build_elf_headers(void* p_0, ax_u32 p_1, ax_u64 p_2, ax_u32 p_3, ax_u32 p_4);
void ax_AxiomLinker_axiom_linker_add_input(void* p_0, ax_string p_1);
ax_bool ax_AxiomLinker_axiom_linker_link(void* p_0);
FmtTokenVec* ax_new_fmt_token_vec();
void ax_FmtTokenVec_push(FmtTokenVec* p_0, FmtToken* p_1);
ax_bool ax_fmt_is_digit(ax_u8 p_0);
ax_bool ax_fmt_is_hex_digit(ax_u8 p_0);
ax_bool ax_fmt_is_oct_digit(ax_u8 p_0);
ax_bool ax_fmt_is_ident_start(ax_u8 p_0);
ax_bool ax_fmt_is_ident_part(ax_u8 p_0);
ax_bool ax_is_punctuation(ax_string p_0);
ax_i64 ax_match_operator_or_punct(ax_string p_0, ax_i64 p_1, ax_i64 p_2);
FmtTokenVec* ax_scan_tokens(ax_string p_0);
Formatter* ax_new_formatter();
ax_string ax_Formatter_format(Formatter* p_0, ax_string p_1);
void ax_AirInst_print_inst(AirInst* p_0);
void ax_AirFunc_print_block(AirFunc* p_0, BasicBlock* p_1);
void ax_AirFunc_print_func(AirFunc* p_0);
void ax_AirModule_print_module(AirModule* p_0);
ax_string ax_read_file_content(ax_string p_0);
ax_bool ax_match_prefix(ax_string p_0, ax_i64 p_1, ax_string p_2);
ax_string ax_strip_imports(ax_string p_0);
ax_string ax_strip_package_prefixes(ax_string p_0);
ax_string ax_local_wslice_to_str(void* p_0, ax_i64 p_1);
void* ax_get_freestanding_args(void* p_0);
ax_i32 ax_main();

// Function Definitions
void* ax_get_str_ptr(ax_string p_0) {
    ax_string r_1 = {0};
    void* r_2 = {0};
    r_1 = p_0;
block_0: ;
    r_1 = r_1;
    r_2 = (void*)r_1.ptr;
    return r_2;
}

void ax_print_raw(ax_string p_0) {
    ax_string r_1 = {0};
    ax_i32 r_2 = {0};
    ax_u32 r_3 = {0};
    void* r_4 = {0};
    ax_i32 r_5 = {0};
    ax_u32 r_6 = {0};
    void* r_7 = {0};
    void* r_8 = {0};
    void* r_9 = {0};
    void* r_10 = {0};
    void* r_11 = {0};
    ax_u32 r_12 = {0};
    void* r_13 = {0};
    void* r_14 = {0};
    void* r_15 = {0};
    void* r_16 = {0};
    ax_i32 r_17 = {0};
    r_1 = p_0;
block_0: ;
    r_1 = r_1;
    r_2 = 4294967285;
    r_3 = (ax_u32)r_2;
    r_4 = GetStdHandle(r_3);
    r_5 = 0;
    r_6 = (ax_u32)r_5;
    r_7 = ax_get_str_ptr(r_1);
    r_8 = (void*)r_7;
    r_9 = ((struct void*)r_0)->f_0;
    r_10 = ((struct void*)r_9)->f_0;
    r_11 = (void*)((void*(*)())r_10)(r_1);
    r_12 = (ax_u32)r_11;
    r_14 = (void*)r_13;
    r_15 = 0;
    r_16 = (void*)r_15;
    r_17 = WriteFile(r_4, r_8, r_12, r_14, r_16);
    return;
}

void ax_print_raw_ptr(void* p_0, ax_i64 p_1) {
    void* r_1 = {0};
    ax_i64 r_2 = {0};
    ax_i32 r_3 = {0};
    ax_u32 r_4 = {0};
    void* r_5 = {0};
    ax_i32 r_6 = {0};
    ax_u32 r_7 = {0};
    void* r_8 = {0};
    ax_u32 r_9 = {0};
    void* r_10 = {0};
    void* r_11 = {0};
    void* r_12 = {0};
    void* r_13 = {0};
    ax_i32 r_14 = {0};
    r_1 = p_0;
    r_2 = p_1;
block_0: ;
    r_1 = r_1;
    r_2 = r_2;
    r_3 = 4294967285;
    r_4 = (ax_u32)r_3;
    r_5 = GetStdHandle(r_4);
    r_6 = 0;
    r_7 = (ax_u32)r_6;
    r_8 = (void*)r_1;
    r_9 = (ax_u32)r_2;
    r_11 = (void*)r_10;
    r_12 = 0;
    r_13 = (void*)r_12;
    r_14 = WriteFile(r_5, r_8, r_9, r_11, r_13);
    return;
}

ax_i32 ax_puts_local(ax_string p_0) {
    ax_string r_1 = {0};
    ax_string r_3 = {0};
    ax_i32 r_5 = {0};
    r_1 = p_0;
block_0: ;
    r_1 = r_1;
    ax_print_raw(r_1);
    r_3 = AX_STR("\n");
    ax_print_raw(r_3);
    r_5 = 0;
    return