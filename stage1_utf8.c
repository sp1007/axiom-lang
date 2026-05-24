#define AX_EMIT_MAIN
#define AX_MAIN_WITH_ARGS
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax_Token;
struct ax_TokenVec;
struct ax_IntVec;
struct ax_Lexer;
struct ax_AstNode;
struct ax_NodeVec;
struct ax_AstTree;
struct ax_InternEntry;
struct ax_InternPool;
struct ax_Parser;
struct ax_Symbol;
struct ax_ScopeEntry;
struct ax_Scope;
struct ax_SymbolVec;
struct ax_ScopeVec;
struct ax_U32Vec;
struct ax_SymbolTable;
struct ax_NameResolver;
struct ax_ComptimeValue;
struct ax_TypeEntry;
struct ax_StructField;
struct ax_StructFieldVec;
struct ax_StructInfo;
struct ax_StructInfoVec;
struct ax_VariantInfo;
struct ax_VariantInfoVec;
struct ax_SumInfo;
struct ax_SumInfoVec;
struct ax_FuncInfo;
struct ax_FuncInfoVec;
struct ax_TypeEntryVec;
struct ax_TypeTable;
struct ax_TypeSubst;
struct ax_TypeSubstVec;
struct ax_Monomorphizer;
struct ax_TypeChecker;
struct ax_CGNode;
struct ax_CGNodeVec;
struct ax_CGEdge;
struct ax_CGEdgeVec;
struct ax_U32VecVec;
struct ax_ConnectionGraph;
struct ax_OwnershipChecker;
struct ax_EscapeAnalyser;
struct ax_CtgcInjector;
struct ax_AliasReuseOptimizer;
struct ax_AirInst;
struct ax_AirInstVec;
struct ax_BasicBlock;
struct ax_BasicBlockVec;
struct ax_AirFunc;
struct ax_AirFuncVec;
struct ax_AirModule;
struct ax_AirFuncBuilder;
struct ax_LocalMapEntry;
struct ax_LocalMap;
struct ax_AirModuleBuilder;
struct ax_FuncLowering;
struct ax_ConstVal;
struct ax_SsaOptimizer;
struct ax_CGenerator;
struct ax_WasmGenerator;
struct ax_MachOperand;
struct ax_MachInst;
struct ax_MachInstVec;
struct ax_InstructionSelector;
struct ax_LiveInterval;
struct ax_LiveIntervalVec;
struct ax_RegAllocation;
struct ax_RegAllocResult;
struct ax_StackFrame;
struct ax_ByteVec;
struct ax_Relocation;
struct ax_RelocationVec;
struct ax_Fixup;
struct ax_FixupVec;
struct ax_LabelMap;
struct ax_MachEmitter;
struct ax_ELF64Sym;
struct ax_ELF64SymVec;
struct ax_COFFReloc;
struct ax_COFFRelocVec;
struct ax_COFFSymbol;
struct ax_COFFSymbolVec;
struct ax_CompiledFuncInfo;
struct ax_CompiledFuncInfoVec;
struct ax_LinkerSymbol;
struct ax_LinkerSymbolVec;
struct ax_ParsedReloc;
struct ax_ParsedRelocVec;
struct ax_LinkerStrVec;
struct ax_ParsedObject;
struct ax_ParsedObjectPtrVec;
struct ax_AxiomLinker;

/* Type definitions */
struct ax_Token {
    ax_u8 kind;
    ax_u8 padding;
    ax_u16 len;
    ax_u32 offset;
};
struct ax_TokenVec {
    struct ax_Token* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_IntVec {
    ax_i32* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_Lexer {
    ax_string src;
    ax_i64 len;
    ax_i64 pos;
    struct ax_TokenVec tokens;
    struct ax_IntVec indent_stack;
    struct ax_IntVec newline_offsets;
};
struct ax_AstNode {
    ax_u8 kind;
    ax_u8 padding;
    ax_u16 flags;
    ax_u32 token_idx;
    ax_u32 first_child;
    ax_u32 next_sibling;
    ax_u32 payload;
    ax_u32 extra_idx;
};
struct ax_NodeVec {
    struct ax_AstNode* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_AstTree {
    struct ax_NodeVec nodes;
    struct ax_IntVec extras;
    ax_string src;
    struct ax_TokenVec tokens;
};
struct ax_InternEntry {
    ax_u32 hash;
    ax_u32 padding;
    ax_u32 start;
    ax_u32 len;
    ax_u32 id;
    ax_u32 dummy;
};
struct ax_InternPool {
    ax_u8* arena;
    ax_i64 arena_len;
    ax_i64 arena_cap;
    struct ax_InternEntry* table;
    ax_i64 table_size;
    struct ax_InternEntry* ids;
    ax_i64 count;
};
struct ax_Parser {
    struct ax_TokenVec tokens;
    ax_i64 pos;
    struct ax_AstTree tree;
    struct ax_InternPool pool;
    ax_string src;
    ax_i64 diags_count;
};
struct ax_Symbol {
    ax_u32 name_id;
    ax_u8 kind;
    ax_u8 padding;
    ax_u16 flags;
    ax_u32 type_id;
    ax_u32 decl_node;
    ax_u32 scope_id;
    ax_u32 next_overload;
};
struct ax_ScopeEntry {
    ax_u32 name_id;
    ax_u32 symbol_idx;
};
struct ax_Scope {
    ax_u8 kind;
    ax_u8 padding1;
    ax_u16 padding2;
    ax_u32 parent_id;
    ax_u32 depth;
    struct ax_ScopeEntry* entries;
    ax_u32 count;
    ax_u32 capacity;
};
struct ax_SymbolVec {
    struct ax_Symbol* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_ScopeVec {
    struct ax_Scope* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_U32Vec {
    ax_u32* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_SymbolTable {
    struct ax_SymbolVec symbols;
    struct ax_ScopeVec scopes;
    struct ax_U32Vec stack;
};
struct ax_NameResolver {
    struct ax_AstTree tree;
    struct ax_InternPool intern;
    struct ax_SymbolTable symtable;
};
struct ax_ComptimeValue {
    ax_u32 kind;
    ax_i64 int_val;
    ax_f64 float_val;
    ax_bool bool_val;
    ax_string str_val;
};
struct ax_TypeEntry {
    ax_u8 kind;
    ax_u8 padding;
    ax_u16 flags;
    ax_u32 name_id;
    ax_u32 size;
    ax_u32 align;
    ax_u32 extra;
};
struct ax_StructField {
    ax_u32 name_id;
    ax_u32 type_id;
};
struct ax_StructFieldVec {
    struct ax_StructField* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_StructInfo {
    struct ax_StructFieldVec fields;
};
struct ax_StructInfoVec {
    struct ax_StructInfo* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_VariantInfo {
    ax_u32 name_id;
    ax_u32 payload_type;
    ax_u8 tag;
    ax_u8 padding1;
    ax_u16 padding2;
};
struct ax_VariantInfoVec {
    struct ax_VariantInfo* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_SumInfo {
    struct ax_VariantInfoVec variants;
};
struct ax_SumInfoVec {
    struct ax_SumInfo* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_FuncInfo {
    struct ax_U32Vec params;
    ax_u32 ret;
    ax_u32 dummy;
};
struct ax_FuncInfoVec {
    struct ax_FuncInfo* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_TypeEntryVec {
    struct ax_TypeEntry* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_TypeTable {
    struct ax_TypeEntryVec entries;
    struct ax_StructInfoVec structs;
    struct ax_FuncInfoVec funcs;
    struct ax_SumInfoVec sumtypes;
};
struct ax_TypeSubst {
    ax_u32 name_id;
    ax_u32 type_id;
};
struct ax_TypeSubstVec {
    struct ax_TypeSubst* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_Monomorphizer {
    struct ax_AstTree tree;
    struct ax_InternPool pool;
    struct ax_SymbolTable symtable;
    struct ax_TypeTable typetable;
};
struct ax_TypeChecker {
    struct ax_AstTree tree;
    struct ax_InternPool intern;
    struct ax_SymbolTable symtable;
    struct ax_TypeTable types;
    ax_u32* node_types;
    ax_i64 node_types_cap;
    ax_u32 current_return;
    ax_u32 current_match_scrutinee;
};
struct ax_CGNode {
    ax_u32 id;
    ax_u32 sym_id;
    ax_u32 type_id;
    ax_bool is_ref;
    ax_u32 lifetime;
};
struct ax_CGNodeVec {
    struct ax_CGNode* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_CGEdge {
    ax_u32 from;
    ax_u32 to;
    ax_u32 kind;
};
struct ax_CGEdgeVec {
    struct ax_CGEdge* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_U32VecVec {
    struct ax_U32Vec* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_ConnectionGraph {
    struct ax_CGNodeVec nodes;
    struct ax_CGEdgeVec edges;
    struct ax_U32VecVec adj_out;
    struct ax_U32VecVec adj_in;
    struct ax_U32Vec sym_to_node;
};
struct ax_OwnershipChecker {
    struct ax_AstTree tree;
    struct ax_InternPool pool;
    struct ax_SymbolTable symtable;
    struct ax_TypeTable typetable;
    ax_u32 errors;
};
struct ax_EscapeAnalyser {
    struct ax_AstTree tree;
    struct ax_InternPool pool;
    struct ax_SymbolTable symtable;
    struct ax_TypeTable typetable;
    struct ax_ConnectionGraph curr_cg;
    ax_u32 escape_node_idx;
};
struct ax_CtgcInjector {
    struct ax_AstTree tree;
    struct ax_InternPool pool;
    struct ax_SymbolTable symtable;
    struct ax_TypeTable typetable;
    struct ax_U32Vec active_vars;
    struct ax_U32Vec active_var_decls;
};
struct ax_AliasReuseOptimizer {
    struct ax_AstTree tree;
    struct ax_InternPool pool;
    struct ax_SymbolTable symtable;
    struct ax_TypeTable typetable;
};
struct ax_AirInst {
    ax_u16 opcode;
    ax_u16 type_id;
    ax_u32 dest;
    ax_u32 src1;
    ax_u32 src2;
};
struct ax_AirInstVec {
    struct ax_AirInst* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_BasicBlock {
    ax_u32 id;
    ax_u32 instrs_start;
    ax_u32 instrs_len;
    ax_u32 succs_start;
    ax_u32 succs_len;
    ax_u32 preds_start;
    ax_u32 preds_len;
    ax_u8 loop_depth;
    ax_bool is_entry;
    ax_bool is_exit;
};
struct ax_BasicBlockVec {
    struct ax_BasicBlock* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_AirFunc {
    ax_u32 sym_id;
    ax_u32 name;
    struct ax_U32Vec params;
    ax_u32 ret_type;
    struct ax_BasicBlockVec blocks;
    struct ax_AirInstVec insts;
    struct ax_U32Vec extras;
    ax_bool is_async;
    ax_bool is_extern;
    struct ax_U32Vec block_instrs;
    struct ax_U32Vec block_succs;
    struct ax_U32Vec block_preds;
};
struct ax_AirFuncVec {
    struct ax_AirFunc* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_AirModule {
    struct ax_AirFuncVec funcs;
};
struct ax_AirFuncBuilder {
    ax_u32 name;
    ax_u32 ret_type;
    struct ax_BasicBlockVec blocks;
    struct ax_AirInstVec insts;
    struct ax_U32Vec extras;
    ax_i32 cur_block;
    ax_u32 next_reg;
    struct ax_U32Vec block_instrs;
    struct ax_U32Vec block_succs;
    struct ax_U32Vec block_preds;
};
struct ax_LocalMapEntry {
    ax_u32 name_id;
    ax_u32 reg;
};
struct ax_LocalMap {
    struct ax_LocalMapEntry* entries;
    ax_u32 count;
    ax_u32 capacity;
};
struct ax_AirModuleBuilder {
    struct ax_AstTree tree;
    struct ax_SymbolTable symbols;
    struct ax_TypeTable typetable;
    struct ax_InternPool pool;
    struct ax_AirModule module;
    ax_u32* node_types;
};
struct ax_FuncLowering {
    struct ax_AirModuleBuilder* mb;
    struct ax_AirFuncBuilder fb;
    struct ax_LocalMap locals;
    struct ax_U32Vec params;
    ax_bool terminated;
};
struct ax_ConstVal {
    ax_bool known;
    ax_u32 val;
};
struct ax_SsaOptimizer {
    ax_u8 level;
};
struct ax_CGenerator {
    struct ax_AirModule module;
    struct ax_AstTree tree;
    struct ax_InternPool pool;
    struct ax_SymbolTable symtable;
    struct ax_TypeTable typetable;
    void* file;
    ax_u32* reg_types;
};
struct ax_WasmGenerator {
    struct ax_AirModule module;
    struct ax_AstTree tree;
    struct ax_InternPool pool;
    struct ax_SymbolTable symtable;
    struct ax_TypeTable typetable;
    void* file;
    ax_u32* reg_types;
};
struct ax_MachOperand {
    ax_u8 kind;
    ax_u8 phys;
    ax_u16 padding;
    ax_u32 vreg;
    ax_u32 label;
    ax_i64 imm;
};
struct ax_MachInst {
    ax_u16 op;
    ax_u8 cc;
    ax_u8 padding;
    struct ax_MachOperand dst;
    struct ax_MachOperand src1;
    struct ax_MachOperand src2;
};
struct ax_MachInstVec {
    struct ax_MachInst* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_InstructionSelector {
    struct ax_AirFunc* fn_ptr;
    ax_string abi_name;
    struct ax_TypeTable table;
    ax_i32 param_idx_processed;
    struct ax_SymbolTable symbols;
    struct ax_InternPool pool;
    ax_u32 max_vreg;
    ax_u32 max_label;
};
struct ax_LiveInterval {
    ax_u32 vreg;
    ax_i32 start;
    ax_i32 end;
};
struct ax_LiveIntervalVec {
    struct ax_LiveInterval* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_RegAllocation {
    ax_u32 vreg;
    ax_u8 phys;
    ax_bool spilled;
    ax_i32 spill_idx;
};
struct ax_RegAllocResult {
    struct ax_RegAllocation* allocs;
    ax_u32 max_vreg;
    ax_i32 spill_count;
};
struct ax_StackFrame {
    ax_u8* callee_saved;
    ax_i64 callee_saved_len;
    ax_i32 spill_slots;
    ax_i32 local_bytes;
    ax_i32 align_padding;
    ax_i32 total_size;
};
struct ax_ByteVec {
    ax_u8* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_Relocation {
    ax_i64 offset;
    ax_u8 kind;
    ax_u8 padding1;
    ax_u16 padding2;
    ax_u32 sym_name;
    ax_i64 addend;
};
struct ax_RelocationVec {
    struct ax_Relocation* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_Fixup {
    ax_i64 offset;
    ax_u32 label_id;
    ax_i32 inst_size;
};
struct ax_FixupVec {
    struct ax_Fixup* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_LabelMap {
    ax_u32* keys;
    ax_i64* values;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_MachEmitter {
    struct ax_ByteVec code;
    struct ax_RelocationVec relocs;
    struct ax_LabelMap labels;
    struct ax_FixupVec fixups;
};
struct ax_ELF64Sym {
    ax_string name_str;
    ax_u64 value;
    ax_u64 size;
    ax_u8 binding;
    ax_u8 sym_type;
    ax_u16 section;
};
struct ax_ELF64SymVec {
    struct ax_ELF64Sym* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_COFFReloc {
    ax_u32 virt_addr;
    ax_u32 sym_table_idx;
    ax_u16 reloc_type;
};
struct ax_COFFRelocVec {
    struct ax_COFFReloc* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_COFFSymbol {
    ax_string name_str;
    ax_u32 value;
    ax_i16 section_num;
    ax_u16 sym_type;
    ax_u8 storage_class;
    ax_u8 num_aux;
};
struct ax_COFFSymbolVec {
    struct ax_COFFSymbol* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_CompiledFuncInfo {
    ax_string name;
    ax_u32 offset;
    ax_u32 size;
    ax_bool is_global;
    struct ax_RelocationVec relocs;
};
struct ax_CompiledFuncInfoVec {
    struct ax_CompiledFuncInfo* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_LinkerSymbol {
    ax_string name;
    ax_i64 section;
    ax_u64 offset;
    ax_u64 size;
    ax_bool defined;
};
struct ax_LinkerSymbolVec {
    struct ax_LinkerSymbol* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_ParsedReloc {
    ax_i64 offset;
    ax_u32 sym_idx;
    ax_bool is_pc;
    ax_i64 addend;
};
struct ax_ParsedRelocVec {
    struct ax_ParsedReloc* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_LinkerStrVec {
    ax_string* data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_ParsedObject {
    struct ax_ByteVec text;
    struct ax_ByteVec rdata;
    struct ax_LinkerSymbolVec symbols;
    struct ax_LinkerStrVec sym_names;
    struct ax_ParsedRelocVec relocs;
    ax_u64 va;
};
struct ax_ParsedObjectPtrVec {
    struct ax_ParsedObject** data;
    ax_i64 len;
    ax_i64 cap;
};
struct ax_AxiomLinker {
    struct ax_LinkerStrVec input_files;
    ax_string output_path;
};

/* Global variables */
extern const ax_u8 ax_TK_INT_LIT;
const ax_u8 ax_TK_INT_LIT = 0;
extern const ax_u8 ax_TK_FLOAT_LIT;
const ax_u8 ax_TK_FLOAT_LIT = 1;
extern const ax_u8 ax_TK_STRING_LIT;
const ax_u8 ax_TK_STRING_LIT = 2;
extern const ax_u8 ax_TK_CHAR_LIT;
const ax_u8 ax_TK_CHAR_LIT = 3;
extern const ax_u8 ax_TK_IDENT;
const ax_u8 ax_TK_IDENT = 4;
extern const ax_u8 ax_TK_AND;
const ax_u8 ax_TK_AND = 5;
extern const ax_u8 ax_TK_AS;
const ax_u8 ax_TK_AS = 6;
extern const ax_u8 ax_TK_ASYNC;
const ax_u8 ax_TK_ASYNC = 7;
extern const ax_u8 ax_TK_AWAIT;
const ax_u8 ax_TK_AWAIT = 8;
extern const ax_u8 ax_TK_CONST;
const ax_u8 ax_TK_CONST = 9;
extern const ax_u8 ax_TK_DEFER;
const ax_u8 ax_TK_DEFER = 10;
extern const ax_u8 ax_TK_ELIF;
const ax_u8 ax_TK_ELIF = 11;
extern const ax_u8 ax_TK_ELSE;
const ax_u8 ax_TK_ELSE = 12;
extern const ax_u8 ax_TK_EXTERN;
const ax_u8 ax_TK_EXTERN = 13;
extern const ax_u8 ax_TK_FALSE;
const ax_u8 ax_TK_FALSE = 14;
extern const ax_u8 ax_TK_FN;
const ax_u8 ax_TK_FN = 15;
extern const ax_u8 ax_TK_FOR;
const ax_u8 ax_TK_FOR = 16;
extern const ax_u8 ax_TK_FUTURE;
const ax_u8 ax_TK_FUTURE = 17;
extern const ax_u8 ax_TK_IF;
const ax_u8 ax_TK_IF = 18;
extern const ax_u8 ax_TK_IMPORT;
const ax_u8 ax_TK_IMPORT = 19;
extern const ax_u8 ax_TK_IN;
const ax_u8 ax_TK_IN = 20;
extern const ax_u8 ax_TK_INTERFACE;
const ax_u8 ax_TK_INTERFACE = 21;
extern const ax_u8 ax_TK_ISOLATED;
const ax_u8 ax_TK_ISOLATED = 22;
extern const ax_u8 ax_TK_LENT;
const ax_u8 ax_TK_LENT = 23;
extern const ax_u8 ax_TK_LET;
const ax_u8 ax_TK_LET = 24;
extern const ax_u8 ax_TK_MATCH;
const ax_u8 ax_TK_MATCH = 25;
extern const ax_u8 ax_TK_MUT;
const ax_u8 ax_TK_MUT = 26;
extern const ax_u8 ax_TK_NIL;
const ax_u8 ax_TK_NIL = 27;
extern const ax_u8 ax_TK_NOT;
const ax_u8 ax_TK_NOT = 28;
extern const ax_u8 ax_TK_OR;
const ax_u8 ax_TK_OR = 29;
extern const ax_u8 ax_TK_PACKED;
const ax_u8 ax_TK_PACKED = 30;
extern const ax_u8 ax_TK_PUB;
const ax_u8 ax_TK_PUB = 31;
extern const ax_u8 ax_TK_RETURN;
const ax_u8 ax_TK_RETURN = 32;
extern const ax_u8 ax_TK_SPAWN;
const ax_u8 ax_TK_SPAWN = 33;
extern const ax_u8 ax_TK_STRUCT;
const ax_u8 ax_TK_STRUCT = 34;
extern const ax_u8 ax_TK_TRUE;
const ax_u8 ax_TK_TRUE = 35;
extern const ax_u8 ax_TK_TYPE;
const ax_u8 ax_TK_TYPE = 36;
extern const ax_u8 ax_TK_UNSAFE;
const ax_u8 ax_TK_UNSAFE = 37;
extern const ax_u8 ax_TK_WHILE;
const ax_u8 ax_TK_WHILE = 38;
extern const ax_u8 ax_TK_PLUS;
const ax_u8 ax_TK_PLUS = 39;
extern const ax_u8 ax_TK_MINUS;
const ax_u8 ax_TK_MINUS = 40;
extern const ax_u8 ax_TK_STAR;
const ax_u8 ax_TK_STAR = 41;
extern const ax_u8 ax_TK_SLASH;
const ax_u8 ax_TK_SLASH = 42;
extern const ax_u8 ax_TK_PERCENT;
const ax_u8 ax_TK_PERCENT = 43;
extern const ax_u8 ax_TK_STAR_STAR;
const ax_u8 ax_TK_STAR_STAR = 44;
extern const ax_u8 ax_TK_EQ_EQ;
const ax_u8 ax_TK_EQ_EQ = 45;
extern const ax_u8 ax_TK_BANG_EQ;
const ax_u8 ax_TK_BANG_EQ = 46;
extern const ax_u8 ax_TK_LT;
const ax_u8 ax_TK_LT = 47;
extern const ax_u8 ax_TK_GT;
const ax_u8 ax_TK_GT = 48;
extern const ax_u8 ax_TK_LT_EQ;
const ax_u8 ax_TK_LT_EQ = 49;
extern const ax_u8 ax_TK_GT_EQ;
const ax_u8 ax_TK_GT_EQ = 50;
extern const ax_u8 ax_TK_AMP;
const ax_u8 ax_TK_AMP = 51;
extern const ax_u8 ax_TK_PIPE;
const ax_u8 ax_TK_PIPE = 52;
extern const ax_u8 ax_TK_CARET;
const ax_u8 ax_TK_CARET = 53;
extern const ax_u8 ax_TK_TILDE;
const ax_u8 ax_TK_TILDE = 54;
extern const ax_u8 ax_TK_LT_LT;
const ax_u8 ax_TK_LT_LT = 55;
extern const ax_u8 ax_TK_GT_GT;
const ax_u8 ax_TK_GT_GT = 56;
extern const ax_u8 ax_TK_EQ;
const ax_u8 ax_TK_EQ = 57;
extern const ax_u8 ax_TK_COLON_EQ;
const ax_u8 ax_TK_COLON_EQ = 58;
extern const ax_u8 ax_TK_PLUS_EQ;
const ax_u8 ax_TK_PLUS_EQ = 59;
extern const ax_u8 ax_TK_MINUS_EQ;
const ax_u8 ax_TK_MINUS_EQ = 60;
extern const ax_u8 ax_TK_STAR_EQ;
const ax_u8 ax_TK_STAR_EQ = 61;
extern const ax_u8 ax_TK_SLASH_EQ;
const ax_u8 ax_TK_SLASH_EQ = 62;
extern const ax_u8 ax_TK_PERCENT_EQ;
const ax_u8 ax_TK_PERCENT_EQ = 63;
extern const ax_u8 ax_TK_DOT;
const ax_u8 ax_TK_DOT = 64;
extern const ax_u8 ax_TK_DOT_STAR;
const ax_u8 ax_TK_DOT_STAR = 65;
extern const ax_u8 ax_TK_COMMA;
const ax_u8 ax_TK_COMMA = 66;
extern const ax_u8 ax_TK_COLON;
const ax_u8 ax_TK_COLON = 67;
extern const ax_u8 ax_TK_SEMICOLON;
const ax_u8 ax_TK_SEMICOLON = 68;
extern const ax_u8 ax_TK_ARROW;
const ax_u8 ax_TK_ARROW = 69;
extern const ax_u8 ax_TK_BANG;
const ax_u8 ax_TK_BANG = 70;
extern const ax_u8 ax_TK_L_PAREN;
const ax_u8 ax_TK_L_PAREN = 71;
extern const ax_u8 ax_TK_R_PAREN;
const ax_u8 ax_TK_R_PAREN = 72;
extern const ax_u8 ax_TK_L_BRACKET;
const ax_u8 ax_TK_L_BRACKET = 73;
extern const ax_u8 ax_TK_R_BRACKET;
const ax_u8 ax_TK_R_BRACKET = 74;
extern const ax_u8 ax_TK_L_BRACE;
const ax_u8 ax_TK_L_BRACE = 75;
extern const ax_u8 ax_TK_R_BRACE;
const ax_u8 ax_TK_R_BRACE = 76;
extern const ax_u8 ax_TK_INDENT;
const ax_u8 ax_TK_INDENT = 77;
extern const ax_u8 ax_TK_DEDENT;
const ax_u8 ax_TK_DEDENT = 78;
extern const ax_u8 ax_TK_NEWLINE;
const ax_u8 ax_TK_NEWLINE = 79;
extern const ax_u8 ax_TK_EOF;
const ax_u8 ax_TK_EOF = 80;
extern const ax_u8 ax_TK_ERROR;
const ax_u8 ax_TK_ERROR = 81;
extern const ax_u8 ax_TK_DOT_DOT;
const ax_u8 ax_TK_DOT_DOT = 82;
extern const ax_u8 ax_TK_HASH;
const ax_u8 ax_TK_HASH = 83;
extern const ax_u8 ax_TK_BREAK;
const ax_u8 ax_TK_BREAK = 84;
extern const ax_u8 ax_TK_CONTINUE;
const ax_u8 ax_TK_CONTINUE = 85;
extern const ax_u8 ax_NODE_INVALID;
const ax_u8 ax_NODE_INVALID = 0;
extern const ax_u8 ax_NODE_PROGRAM;
const ax_u8 ax_NODE_PROGRAM = 1;
extern const ax_u8 ax_NODE_FUNC_DECL;
const ax_u8 ax_NODE_FUNC_DECL = 2;
extern const ax_u8 ax_NODE_STRUCT_DECL;
const ax_u8 ax_NODE_STRUCT_DECL = 3;
extern const ax_u8 ax_NODE_INTERFACE_DECL;
const ax_u8 ax_NODE_INTERFACE_DECL = 4;
extern const ax_u8 ax_NODE_IMPORT_DECL;
const ax_u8 ax_NODE_IMPORT_DECL = 5;
extern const ax_u8 ax_NODE_CONST_DECL;
const ax_u8 ax_NODE_CONST_DECL = 6;
extern const ax_u8 ax_NODE_TYPE_ALIAS_DECL;
const ax_u8 ax_NODE_TYPE_ALIAS_DECL = 7;
extern const ax_u8 ax_NODE_PARAM_DECL;
const ax_u8 ax_NODE_PARAM_DECL = 8;
extern const ax_u8 ax_NODE_FIELD_DECL;
const ax_u8 ax_NODE_FIELD_DECL = 9;
extern const ax_u8 ax_NODE_METHOD_SIG;
const ax_u8 ax_NODE_METHOD_SIG = 10;
extern const ax_u8 ax_NODE_VARIANT_DECL;
const ax_u8 ax_NODE_VARIANT_DECL = 11;
extern const ax_u8 ax_NODE_BLOCK;
const ax_u8 ax_NODE_BLOCK = 12;
extern const ax_u8 ax_NODE_VAR_DECL;
const ax_u8 ax_NODE_VAR_DECL = 13;
extern const ax_u8 ax_NODE_ASSIGN_STMT;
const ax_u8 ax_NODE_ASSIGN_STMT = 14;
extern const ax_u8 ax_NODE_RETURN_STMT;
const ax_u8 ax_NODE_RETURN_STMT = 15;
extern const ax_u8 ax_NODE_IF_STMT;
const ax_u8 ax_NODE_IF_STMT = 16;
extern const ax_u8 ax_NODE_ELIF_CLAUSE;
const ax_u8 ax_NODE_ELIF_CLAUSE = 17;
extern const ax_u8 ax_NODE_ELSE_CLAUSE;
const ax_u8 ax_NODE_ELSE_CLAUSE = 18;
extern const ax_u8 ax_NODE_FOR_STMT;
const ax_u8 ax_NODE_FOR_STMT = 19;
extern const ax_u8 ax_NODE_WHILE_STMT;
const ax_u8 ax_NODE_WHILE_STMT = 20;
extern const ax_u8 ax_NODE_MATCH_STMT;
const ax_u8 ax_NODE_MATCH_STMT = 21;
extern const ax_u8 ax_NODE_MATCH_ARM;
const ax_u8 ax_NODE_MATCH_ARM = 22;
extern const ax_u8 ax_NODE_DEFER_STMT;
const ax_u8 ax_NODE_DEFER_STMT = 23;
extern const ax_u8 ax_NODE_UNSAFE_BLOCK;
const ax_u8 ax_NODE_UNSAFE_BLOCK = 24;
extern const ax_u8 ax_NODE_ARENA_BLOCK;
const ax_u8 ax_NODE_ARENA_BLOCK = 25;
extern const ax_u8 ax_NODE_BINARY_EXPR;
const ax_u8 ax_NODE_BINARY_EXPR = 26;
extern const ax_u8 ax_NODE_UNARY_EXPR;
const ax_u8 ax_NODE_UNARY_EXPR = 27;
extern const ax_u8 ax_NODE_CALL_EXPR;
const ax_u8 ax_NODE_CALL_EXPR = 28;
extern const ax_u8 ax_NODE_INDEX_EXPR;
const ax_u8 ax_NODE_INDEX_EXPR = 29;
extern const ax_u8 ax_NODE_FIELD_EXPR;
const ax_u8 ax_NODE_FIELD_EXPR = 30;
extern const ax_u8 ax_NODE_CAST_EXPR;
const ax_u8 ax_NODE_CAST_EXPR = 31;
extern const ax_u8 ax_NODE_DEREF_EXPR;
const ax_u8 ax_NODE_DEREF_EXPR = 32;
extern const ax_u8 ax_NODE_SPAWN_EXPR;
const ax_u8 ax_NODE_SPAWN_EXPR = 33;
extern const ax_u8 ax_NODE_AWAIT_EXPR;
const ax_u8 ax_NODE_AWAIT_EXPR = 34;
extern const ax_u8 ax_NODE_CLOSURE_EXPR;
const ax_u8 ax_NODE_CLOSURE_EXPR = 35;
extern const ax_u8 ax_NODE_INT_LIT;
const ax_u8 ax_NODE_INT_LIT = 36;
extern const ax_u8 ax_NODE_FLOAT_LIT;
const ax_u8 ax_NODE_FLOAT_LIT = 37;
extern const ax_u8 ax_NODE_STRING_LIT;
const ax_u8 ax_NODE_STRING_LIT = 38;
extern const ax_u8 ax_NODE_CHAR_LIT;
const ax_u8 ax_NODE_CHAR_LIT = 39;
extern const ax_u8 ax_NODE_BOOL_LIT;
const ax_u8 ax_NODE_BOOL_LIT = 40;
extern const ax_u8 ax_NODE_NIL_LIT;
const ax_u8 ax_NODE_NIL_LIT = 41;
extern const ax_u8 ax_NODE_IDENT;
const ax_u8 ax_NODE_IDENT = 42;
extern const ax_u8 ax_NODE_ARRAY_LIT;
const ax_u8 ax_NODE_ARRAY_LIT = 43;
extern const ax_u8 ax_NODE_STRUCT_LIT;
const ax_u8 ax_NODE_STRUCT_LIT = 44;
extern const ax_u8 ax_NODE_NAMED_ARG;
const ax_u8 ax_NODE_NAMED_ARG = 45;
extern const ax_u8 ax_NODE_TYPE_EXPR;
const ax_u8 ax_NODE_TYPE_EXPR = 46;
extern const ax_u8 ax_NODE_PTR_TYPE;
const ax_u8 ax_NODE_PTR_TYPE = 47;
extern const ax_u8 ax_NODE_SLICE_TYPE;
const ax_u8 ax_NODE_SLICE_TYPE = 48;
extern const ax_u8 ax_NODE_ARRAY_TYPE;
const ax_u8 ax_NODE_ARRAY_TYPE = 49;
extern const ax_u8 ax_NODE_FUNC_TYPE;
const ax_u8 ax_NODE_FUNC_TYPE = 50;
extern const ax_u8 ax_NODE_GENERIC_TYPE;
const ax_u8 ax_NODE_GENERIC_TYPE = 51;
extern const ax_u8 ax_NODE_ISOLATED_TYPE;
const ax_u8 ax_NODE_ISOLATED_TYPE = 52;
extern const ax_u8 ax_NODE_FUTURE_TYPE;
const ax_u8 ax_NODE_FUTURE_TYPE = 53;
extern const ax_u8 ax_NODE_SUM_TYPE;
const ax_u8 ax_NODE_SUM_TYPE = 54;
extern const ax_u8 ax_NODE_WILDCARD_PAT;
const ax_u8 ax_NODE_WILDCARD_PAT = 55;
extern const ax_u8 ax_NODE_LITERAL_PAT;
const ax_u8 ax_NODE_LITERAL_PAT = 56;
extern const ax_u8 ax_NODE_BINDING_PAT;
const ax_u8 ax_NODE_BINDING_PAT = 57;
extern const ax_u8 ax_NODE_VARIANT_PAT;
const ax_u8 ax_NODE_VARIANT_PAT = 58;
extern const ax_u8 ax_NODE_TUPLE_PAT;
const ax_u8 ax_NODE_TUPLE_PAT = 59;
extern const ax_u8 ax_NODE_GENERIC_PARAMS;
const ax_u8 ax_NODE_GENERIC_PARAMS = 60;
extern const ax_u8 ax_NODE_GENERIC_PARAM;
const ax_u8 ax_NODE_GENERIC_PARAM = 61;
extern const ax_u8 ax_NODE_EFFECT_ANNOTATION;
const ax_u8 ax_NODE_EFFECT_ANNOTATION = 62;
extern const ax_u8 ax_NODE_ERROR;
const ax_u8 ax_NODE_ERROR = 63;
extern const ax_u8 ax_NODE_DESTROY_STMT;
const ax_u8 ax_NODE_DESTROY_STMT = 64;
extern const ax_u8 ax_NODE_ALIAS_STMT;
const ax_u8 ax_NODE_ALIAS_STMT = 65;
extern const ax_u8 ax_NODE_COMPTIME;
const ax_u8 ax_NODE_COMPTIME = 66;
extern const ax_u8 ax_NODE_BREAK_STMT;
const ax_u8 ax_NODE_BREAK_STMT = 67;
extern const ax_u8 ax_NODE_CONTINUE_STMT;
const ax_u8 ax_NODE_CONTINUE_STMT = 68;
extern const ax_u16 ax_FLAG_IS_PUB;
const ax_u16 ax_FLAG_IS_PUB = 1;
extern const ax_u16 ax_FLAG_IS_MUT;
const ax_u16 ax_FLAG_IS_MUT = 2;
extern const ax_u16 ax_FLAG_IS_ASYNC;
const ax_u16 ax_FLAG_IS_ASYNC = 4;
extern const ax_u16 ax_FLAG_IS_EXTERN;
const ax_u16 ax_FLAG_IS_EXTERN = 8;
extern const ax_u16 ax_FLAG_IS_SINK;
const ax_u16 ax_FLAG_IS_SINK = 16;
extern const ax_u16 ax_FLAG_IS_LENT;
const ax_u16 ax_FLAG_IS_LENT = 32;
extern const ax_u64 ax_FLAG_IS_PACKED;
const ax_u64 ax_FLAG_IS_PACKED = 64;
extern const ax_u16 ax_FLAG_ESCAPES_TO_HEAP;
const ax_u16 ax_FLAG_ESCAPES_TO_HEAP = 128;
extern const ax_u16 ax_FLAG_USES_ARENA;
const ax_u16 ax_FLAG_USES_ARENA = 256;
extern const ax_u16 ax_FLAG_IS_GENERIC;
const ax_u16 ax_FLAG_IS_GENERIC = 512;
extern const ax_u16 ax_FLAG_IS_MOVED;
const ax_u16 ax_FLAG_IS_MOVED = 1024;
extern const ax_u32 ax_NULL_IDX;
const ax_u32 ax_NULL_IDX = 0;
extern const ax_i32 ax_BP_NONE;
const ax_i32 ax_BP_NONE = 0;
extern const ax_i32 ax_BP_OR;
const ax_i32 ax_BP_OR = 10;
extern const ax_i32 ax_BP_AND;
const ax_i32 ax_BP_AND = 20;
extern const ax_i32 ax_BP_NOT;
const ax_i32 ax_BP_NOT = 30;
extern const ax_i32 ax_BP_CMP;
const ax_i32 ax_BP_CMP = 40;
extern const ax_i32 ax_BP_BIT_OR;
const ax_i32 ax_BP_BIT_OR = 50;
extern const ax_i32 ax_BP_BIT_XOR;
const ax_i32 ax_BP_BIT_XOR = 60;
extern const ax_i32 ax_BP_BIT_AND;
const ax_i32 ax_BP_BIT_AND = 70;
extern const ax_i32 ax_BP_SHIFT;
const ax_i32 ax_BP_SHIFT = 80;
extern const ax_i32 ax_BP_ADD;
const ax_i32 ax_BP_ADD = 90;
extern const ax_i32 ax_BP_MUL;
const ax_i32 ax_BP_MUL = 100;
extern const ax_i32 ax_BP_POWER;
const ax_i32 ax_BP_POWER = 110;
extern const ax_i32 ax_BP_UNARY;
const ax_i32 ax_BP_UNARY = 120;
extern const ax_i32 ax_BP_POSTFIX;
const ax_i32 ax_BP_POSTFIX = 130;
extern const ax_u8 ax_SYM_VAR;
const ax_u8 ax_SYM_VAR = 0;
extern const ax_u8 ax_SYM_FUNC;
const ax_u8 ax_SYM_FUNC = 1;
extern const ax_u8 ax_SYM_STRUCT;
const ax_u8 ax_SYM_STRUCT = 2;
extern const ax_u8 ax_SYM_INTERFACE;
const ax_u8 ax_SYM_INTERFACE = 3;
extern const ax_u8 ax_SYM_TYPE_ALIAS;
const ax_u8 ax_SYM_TYPE_ALIAS = 4;
extern const ax_u8 ax_SYM_VARIANT;
const ax_u8 ax_SYM_VARIANT = 5;
extern const ax_u8 ax_SYM_PARAM;
const ax_u8 ax_SYM_PARAM = 6;
extern const ax_u8 ax_SYM_FIELD;
const ax_u8 ax_SYM_FIELD = 7;
extern const ax_u8 ax_SYM_GENERIC_PARAM;
const ax_u8 ax_SYM_GENERIC_PARAM = 8;
extern const ax_u8 ax_SYM_MODULE;
const ax_u8 ax_SYM_MODULE = 9;
extern const ax_u8 ax_SYM_BUILTIN_TYPE;
const ax_u8 ax_SYM_BUILTIN_TYPE = 10;
extern const ax_u8 ax_SYM_ENUM_VARIANT;
const ax_u8 ax_SYM_ENUM_VARIANT = 11;
extern const ax_u8 ax_SYM_CONST;
const ax_u8 ax_SYM_CONST = 12;
extern const ax_u16 ax_SYM_FLAG_PUB;
const ax_u16 ax_SYM_FLAG_PUB = 1;
extern const ax_u16 ax_SYM_FLAG_MUT;
const ax_u16 ax_SYM_FLAG_MUT = 2;
extern const ax_u16 ax_SYM_FLAG_EXTERN;
const ax_u16 ax_SYM_FLAG_EXTERN = 4;
extern const ax_u16 ax_SYM_FLAG_SINK;
const ax_u16 ax_SYM_FLAG_SINK = 8;
extern const ax_u16 ax_SYM_FLAG_LENT;
const ax_u16 ax_SYM_FLAG_LENT = 16;
extern const ax_u16 ax_SYM_FLAG_ASYNC;
const ax_u16 ax_SYM_FLAG_ASYNC = 32;
extern const ax_u16 ax_SYM_FLAG_PURE;
const ax_u16 ax_SYM_FLAG_PURE = 64;
extern const ax_u16 ax_SYM_FLAG_MOVED;
const ax_u16 ax_SYM_FLAG_MOVED = 128;
extern const ax_u16 ax_SYM_FLAG_USED;
const ax_u16 ax_SYM_FLAG_USED = 256;
extern const ax_u16 ax_SYM_FLAG_COMPTIME;
const ax_u16 ax_SYM_FLAG_COMPTIME = 512;
extern const ax_u16 ax_SYM_FLAG_GENERIC;
const ax_u16 ax_SYM_FLAG_GENERIC = 1024;
extern const ax_u8 ax_SCOPE_GLOBAL;
const ax_u8 ax_SCOPE_GLOBAL = 0;
extern const ax_u8 ax_SCOPE_FUNCTION;
const ax_u8 ax_SCOPE_FUNCTION = 1;
extern const ax_u8 ax_SCOPE_BLOCK;
const ax_u8 ax_SCOPE_BLOCK = 2;
extern const ax_u8 ax_SCOPE_CLOSURE;
const ax_u8 ax_SCOPE_CLOSURE = 3;
extern const ax_u8 ax_SCOPE_LOOP;
const ax_u8 ax_SCOPE_LOOP = 4;
extern const ax_u32 ax_TYPE_UNKNOWN;
const ax_u32 ax_TYPE_UNKNOWN = 0;
extern const ax_u32 ax_TYPE_I8;
const ax_u32 ax_TYPE_I8 = 1;
extern const ax_u32 ax_TYPE_I16;
const ax_u32 ax_TYPE_I16 = 2;
extern const ax_u32 ax_TYPE_I32;
const ax_u32 ax_TYPE_I32 = 3;
extern const ax_u32 ax_TYPE_I64;
const ax_u32 ax_TYPE_I64 = 4;
extern const ax_u32 ax_TYPE_U8;
const ax_u32 ax_TYPE_U8 = 5;
extern const ax_u32 ax_TYPE_U16;
const ax_u32 ax_TYPE_U16 = 6;
extern const ax_u32 ax_TYPE_U32;
const ax_u32 ax_TYPE_U32 = 7;
extern const ax_u32 ax_TYPE_U64;
const ax_u32 ax_TYPE_U64 = 8;
extern const ax_u32 ax_TYPE_F32;
const ax_u32 ax_TYPE_F32 = 9;
extern const ax_u32 ax_TYPE_F64;
const ax_u32 ax_TYPE_F64 = 10;
extern const ax_u32 ax_TYPE_BOOL;
const ax_u32 ax_TYPE_BOOL = 11;
extern const ax_u32 ax_TYPE_STRING;
const ax_u32 ax_TYPE_STRING = 12;
extern const ax_u32 ax_TYPE_CHAR8;
const ax_u32 ax_TYPE_CHAR8 = 13;
extern const ax_u32 ax_TYPE_VOID;
const ax_u32 ax_TYPE_VOID = 14;
extern const ax_u32 ax_TYPE_ISIZE;
const ax_u32 ax_TYPE_ISIZE = 15;
extern const ax_u32 ax_TYPE_USIZE;
const ax_u32 ax_TYPE_USIZE = 16;
extern const ax_u32 ax_TYPE_ORD;
const ax_u32 ax_TYPE_ORD = 17;
extern const ax_u32 ax_TYPE_EQ;
const ax_u32 ax_TYPE_EQ = 18;
extern const ax_u32 ax_TYPE_HASH;
const ax_u32 ax_TYPE_HASH = 19;
extern const ax_u32 ax_TYPE_DISPLAY;
const ax_u32 ax_TYPE_DISPLAY = 20;
extern const ax_u32 ax_TYPE_ACTOR_REF;
const ax_u32 ax_TYPE_ACTOR_REF = 21;
extern const ax_u8 ax_TYPE_KIND_PRIMITIVE;
const ax_u8 ax_TYPE_KIND_PRIMITIVE = 0;
extern const ax_u8 ax_TYPE_KIND_STRUCT;
const ax_u8 ax_TYPE_KIND_STRUCT = 1;
extern const ax_u8 ax_TYPE_KIND_FUNC;
const ax_u8 ax_TYPE_KIND_FUNC = 2;
extern const ax_u8 ax_TYPE_KIND_GENERIC;
const ax_u8 ax_TYPE_KIND_GENERIC = 3;
extern const ax_u8 ax_TYPE_KIND_INTERFACE;
const ax_u8 ax_TYPE_KIND_INTERFACE = 4;
extern const ax_u8 ax_TYPE_KIND_SUM;
const ax_u8 ax_TYPE_KIND_SUM = 5;
extern const ax_u8 ax_TYPE_KIND_POINTER;
const ax_u8 ax_TYPE_KIND_POINTER = 6;
extern const ax_u8 ax_TYPE_KIND_SLICE;
const ax_u8 ax_TYPE_KIND_SLICE = 7;
extern const ax_u8 ax_TYPE_KIND_ARRAY;
const ax_u8 ax_TYPE_KIND_ARRAY = 8;
extern const ax_u32 ax_EDGE_OWNS;
const ax_u32 ax_EDGE_OWNS = 0;
extern const ax_u32 ax_EDGE_BORROWS;
const ax_u32 ax_EDGE_BORROWS = 1;
extern const ax_u32 ax_EDGE_FLOWS_TO;
const ax_u32 ax_EDGE_FLOWS_TO = 2;
extern const ax_u32 ax_EDGE_ESCAPES_TO;
const ax_u32 ax_EDGE_ESCAPES_TO = 3;
extern const ax_u32 ax_EDGE_REUSED_BY;
const ax_u32 ax_EDGE_REUSED_BY = 4;
extern const ax_u16 ax_OP_NOP;
const ax_u16 ax_OP_NOP = 0x0000;
extern const ax_u16 ax_OP_ALLOC;
const ax_u16 ax_OP_ALLOC = 0x0101;
extern const ax_u16 ax_OP_FREE;
const ax_u16 ax_OP_FREE = 0x0102;
extern const ax_u16 ax_OP_LOAD;
const ax_u16 ax_OP_LOAD = 0x0103;
extern const ax_u16 ax_OP_STORE;
const ax_u16 ax_OP_STORE = 0x0104;
extern const ax_u16 ax_OP_GEP;
const ax_u16 ax_OP_GEP = 0x0105;
extern const ax_u16 ax_OP_COPY;
const ax_u16 ax_OP_COPY = 0x0106;
extern const ax_u16 ax_OP_MOVE;
const ax_u16 ax_OP_MOVE = 0x0107;
extern const ax_u16 ax_OP_MAKE_REF;
const ax_u16 ax_OP_MAKE_REF = 0x0108;
extern const ax_u16 ax_OP_DEREF;
const ax_u16 ax_OP_DEREF = 0x0109;
extern const ax_u16 ax_OP_ARENA_ALLOC;
const ax_u16 ax_OP_ARENA_ALLOC = 0x010A;
extern const ax_u16 ax_OP_DESTROY;
const ax_u16 ax_OP_DESTROY = 0x010B;
extern const ax_u16 ax_OP_ALIAS_REUSE;
const ax_u16 ax_OP_ALIAS_REUSE = 0x010C;
extern const ax_u16 ax_OP_GET_FIELD;
const ax_u16 ax_OP_GET_FIELD = 0x010D;
extern const ax_u16 ax_OP_SET_FIELD;
const ax_u16 ax_OP_SET_FIELD = 0x010E;
extern const ax_u16 ax_OP_INDEX;
const ax_u16 ax_OP_INDEX = 0x010F;
extern const ax_u16 ax_OP_SLICE;
const ax_u16 ax_OP_SLICE = 0x0110;
extern const ax_u16 ax_OP_ICONST;
const ax_u16 ax_OP_ICONST = 0x0201;
extern const ax_u16 ax_OP_FCONST;
const ax_u16 ax_OP_FCONST = 0x0202;
extern const ax_u16 ax_OP_IADD;
const ax_u16 ax_OP_IADD = 0x0203;
extern const ax_u16 ax_OP_ISUB;
const ax_u16 ax_OP_ISUB = 0x0204;
extern const ax_u16 ax_OP_IMUL;
const ax_u16 ax_OP_IMUL = 0x0205;
extern const ax_u16 ax_OP_IDIV;
const ax_u16 ax_OP_IDIV = 0x0206;
extern const ax_u16 ax_OP_IMOD;
const ax_u16 ax_OP_IMOD = 0x0207;
extern const ax_u16 ax_OP_FADD;
const ax_u16 ax_OP_FADD = 0x0208;
extern const ax_u16 ax_OP_FSUB;
const ax_u16 ax_OP_FSUB = 0x0209;
extern const ax_u16 ax_OP_FMUL;
const ax_u16 ax_OP_FMUL = 0x020A;
extern const ax_u16 ax_OP_FDIV;
const ax_u16 ax_OP_FDIV = 0x020B;
extern const ax_u16 ax_OP_EQ;
const ax_u16 ax_OP_EQ = 0x020C;
extern const ax_u16 ax_OP_NE;
const ax_u16 ax_OP_NE = 0x020D;
extern const ax_u16 ax_OP_LT;
const ax_u16 ax_OP_LT = 0x020E;
extern const ax_u16 ax_OP_LE;
const ax_u16 ax_OP_LE = 0x020F;
extern const ax_u16 ax_OP_GT;
const ax_u16 ax_OP_GT = 0x0210;
extern const ax_u16 ax_OP_GE;
const ax_u16 ax_OP_GE = 0x0211;
extern const ax_u16 ax_OP_AND;
const ax_u16 ax_OP_AND = 0x0212;
extern const ax_u16 ax_OP_OR;
const ax_u16 ax_OP_OR = 0x0213;
extern const ax_u16 ax_OP_XOR;
const ax_u16 ax_OP_XOR = 0x0214;
extern const ax_u16 ax_OP_SHL;
const ax_u16 ax_OP_SHL = 0x0215;
extern const ax_u16 ax_OP_SHR;
const ax_u16 ax_OP_SHR = 0x0216;
extern const ax_u16 ax_OP_NOT;
const ax_u16 ax_OP_NOT = 0x0217;
extern const ax_u16 ax_OP_NEG;
const ax_u16 ax_OP_NEG = 0x021D;
extern const ax_u16 ax_OP_ITOF;
const ax_u16 ax_OP_ITOF = 0x021E;
extern const ax_u16 ax_OP_FTOI;
const ax_u16 ax_OP_FTOI = 0x021F;
extern const ax_u16 ax_OP_ZEXT;
const ax_u16 ax_OP_ZEXT = 0x0220;
extern const ax_u16 ax_OP_SEXT;
const ax_u16 ax_OP_SEXT = 0x0221;
extern const ax_u16 ax_OP_TRUNC;
const ax_u16 ax_OP_TRUNC = 0x0222;
extern const ax_u16 ax_OP_CAST;
const ax_u16 ax_OP_CAST = 0x0223;
extern const ax_u16 ax_OP_JUMP;
const ax_u16 ax_OP_JUMP = 0x0301;
extern const ax_u16 ax_OP_BRANCH;
const ax_u16 ax_OP_BRANCH = 0x0302;
extern const ax_u16 ax_OP_CALL;
const ax_u16 ax_OP_CALL = 0x0303;
extern const ax_u16 ax_OP_RETURN;
const ax_u16 ax_OP_RETURN = 0x0304;
extern const ax_u16 ax_OP_PHI;
const ax_u16 ax_OP_PHI = 0x0305;
extern const ax_u16 ax_OP_LOOP_BEGIN;
const ax_u16 ax_OP_LOOP_BEGIN = 0x0306;
extern const ax_u16 ax_OP_LOOP_END;
const ax_u16 ax_OP_LOOP_END = 0x0307;
extern const ax_u16 ax_OP_SPAWN;
const ax_u16 ax_OP_SPAWN = 0x0308;
extern const ax_u16 ax_OP_SEND;
const ax_u16 ax_OP_SEND = 0x0309;
extern const ax_u16 ax_OP_RECV;
const ax_u16 ax_OP_RECV = 0x030A;
extern const ax_u16 ax_OP_AWAIT;
const ax_u16 ax_OP_AWAIT = 0x030B;
extern const ax_u16 ax_OP_SYSCALL;
const ax_u16 ax_OP_SYSCALL = 0x030C;
extern const ax_u16 ax_OP_SIMD_LOAD;
const ax_u16 ax_OP_SIMD_LOAD = 0x0401;
extern const ax_u16 ax_OP_SIMD_STORE;
const ax_u16 ax_OP_SIMD_STORE = 0x0402;
extern const ax_u16 ax_OP_SIMD_ADD;
const ax_u16 ax_OP_SIMD_ADD = 0x0403;
extern const ax_u16 ax_OP_SIMD_MUL;
const ax_u16 ax_OP_SIMD_MUL = 0x0404;
extern const ax_u16 ax_OP_SIMD_FMA;
const ax_u16 ax_OP_SIMD_FMA = 0x0405;
extern const ax_u16 ax_OP_COMPTIME;
const ax_u16 ax_OP_COMPTIME = 0x0501;
extern const ax_u8 ax_REG_RAX;
const ax_u8 ax_REG_RAX = 0;
extern const ax_u8 ax_REG_RCX;
const ax_u8 ax_REG_RCX = 1;
extern const ax_u8 ax_REG_RDX;
const ax_u8 ax_REG_RDX = 2;
extern const ax_u8 ax_REG_RBX;
const ax_u8 ax_REG_RBX = 3;
extern const ax_u8 ax_REG_RSP;
const ax_u8 ax_REG_RSP = 4;
extern const ax_u8 ax_REG_RBP;
const ax_u8 ax_REG_RBP = 5;
extern const ax_u8 ax_REG_RSI;
const ax_u8 ax_REG_RSI = 6;
extern const ax_u8 ax_REG_RDI;
const ax_u8 ax_REG_RDI = 7;
extern const ax_u8 ax_REG_R8;
const ax_u8 ax_REG_R8 = 8;
extern const ax_u8 ax_REG_R9;
const ax_u8 ax_REG_R9 = 9;
extern const ax_u8 ax_REG_R10;
const ax_u8 ax_REG_R10 = 10;
extern const ax_u8 ax_REG_R11;
const ax_u8 ax_REG_R11 = 11;
extern const ax_u8 ax_REG_R12;
const ax_u8 ax_REG_R12 = 12;
extern const ax_u8 ax_REG_R13;
const ax_u8 ax_REG_R13 = 13;
extern const ax_u8 ax_REG_R14;
const ax_u8 ax_REG_R14 = 14;
extern const ax_u8 ax_REG_R15;
const ax_u8 ax_REG_R15 = 15;
extern const ax_u8 ax_REG_NONE;
const ax_u8 ax_REG_NONE = 255;
extern const ax_u16 ax_MACH_NOP;
const ax_u16 ax_MACH_NOP = 0;
extern const ax_u16 ax_MACH_MOV;
const ax_u16 ax_MACH_MOV = 1;
extern const ax_u16 ax_MACH_MOV_IMM;
const ax_u16 ax_MACH_MOV_IMM = 2;
extern const ax_u16 ax_MACH_ADD;
const ax_u16 ax_MACH_ADD = 3;
extern const ax_u16 ax_MACH_SUB;
const ax_u16 ax_MACH_SUB = 4;
extern const ax_u16 ax_MACH_IMUL;
const ax_u16 ax_MACH_IMUL = 5;
extern const ax_u16 ax_MACH_IDIV;
const ax_u16 ax_MACH_IDIV = 6;
extern const ax_u16 ax_MACH_CQO;
const ax_u16 ax_MACH_CQO = 7;
extern const ax_u16 ax_MACH_NEG;
const ax_u16 ax_MACH_NEG = 8;
extern const ax_u16 ax_MACH_NOT;
const ax_u16 ax_MACH_NOT = 9;
extern const ax_u16 ax_MACH_AND;
const ax_u16 ax_MACH_AND = 10;
extern const ax_u16 ax_MACH_OR;
const ax_u16 ax_MACH_OR = 11;
extern const ax_u16 ax_MACH_XOR;
const ax_u16 ax_MACH_XOR = 12;
extern const ax_u16 ax_MACH_SHL;
const ax_u16 ax_MACH_SHL = 13;
extern const ax_u16 ax_MACH_SAR;
const ax_u16 ax_MACH_SAR = 14;
extern const ax_u16 ax_MACH_CMP;
const ax_u16 ax_MACH_CMP = 15;
extern const ax_u16 ax_MACH_TEST;
const ax_u16 ax_MACH_TEST = 16;
extern const ax_u16 ax_MACH_SETCC;
const ax_u16 ax_MACH_SETCC = 17;
extern const ax_u16 ax_MACH_MOVZX_B;
const ax_u16 ax_MACH_MOVZX_B = 18;
extern const ax_u16 ax_MACH_JMP;
const ax_u16 ax_MACH_JMP = 19;
extern const ax_u16 ax_MACH_JCC;
const ax_u16 ax_MACH_JCC = 20;
extern const ax_u16 ax_MACH_CALL;
const ax_u16 ax_MACH_CALL = 21;
extern const ax_u16 ax_MACH_CALL_INDIRECT;
const ax_u16 ax_MACH_CALL_INDIRECT = 22;
extern const ax_u16 ax_MACH_RET;
const ax_u16 ax_MACH_RET = 23;
extern const ax_u16 ax_MACH_PUSH;
const ax_u16 ax_MACH_PUSH = 24;
extern const ax_u16 ax_MACH_POP;
const ax_u16 ax_MACH_POP = 25;
extern const ax_u16 ax_MACH_LEA;
const ax_u16 ax_MACH_LEA = 26;
extern const ax_u16 ax_MACH_LOAD;
const ax_u16 ax_MACH_LOAD = 27;
extern const ax_u16 ax_MACH_STORE;
const ax_u16 ax_MACH_STORE = 28;
extern const ax_u16 ax_MACH_XOR_ZERO;
const ax_u16 ax_MACH_XOR_ZERO = 29;
extern const ax_u16 ax_MACH_LABEL;
const ax_u16 ax_MACH_LABEL = 30;
extern const ax_u16 ax_MACH_SYSCALL;
const ax_u16 ax_MACH_SYSCALL = 31;
extern const ax_u8 ax_OPND_NONE;
const ax_u8 ax_OPND_NONE = 0;
extern const ax_u8 ax_OPND_VREG;
const ax_u8 ax_OPND_VREG = 1;
extern const ax_u8 ax_OPND_PHYS;
const ax_u8 ax_OPND_PHYS = 2;
extern const ax_u8 ax_OPND_IMM;
const ax_u8 ax_OPND_IMM = 3;
extern const ax_u8 ax_OPND_LABEL;
const ax_u8 ax_OPND_LABEL = 4;
extern const ax_u8 ax_OPND_MEM;
const ax_u8 ax_OPND_MEM = 5;
extern const ax_u8 ax_CC_O;
const ax_u8 ax_CC_O = 0x00;
extern const ax_u8 ax_CC_NO;
const ax_u8 ax_CC_NO = 0x01;
extern const ax_u8 ax_CC_B;
const ax_u8 ax_CC_B = 0x02;
extern const ax_u8 ax_CC_AE;
const ax_u8 ax_CC_AE = 0x03;
extern const ax_u8 ax_CC_E;
const ax_u8 ax_CC_E = 0x04;
extern const ax_u8 ax_CC_NE;
const ax_u8 ax_CC_NE = 0x05;
extern const ax_u8 ax_CC_BE;
const ax_u8 ax_CC_BE = 0x06;
extern const ax_u8 ax_CC_A;
const ax_u8 ax_CC_A = 0x07;
extern const ax_u8 ax_CC_S;
const ax_u8 ax_CC_S = 0x08;
extern const ax_u8 ax_CC_NS;
const ax_u8 ax_CC_NS = 0x09;
extern const ax_u8 ax_CC_PE;
const ax_u8 ax_CC_PE = 0x0A;
extern const ax_u8 ax_CC_PO;
const ax_u8 ax_CC_PO = 0x0B;
extern const ax_u8 ax_CC_L;
const ax_u8 ax_CC_L = 0x0C;
extern const ax_u8 ax_CC_GE;
const ax_u8 ax_CC_GE = 0x0D;
extern const ax_u8 ax_CC_LE;
const ax_u8 ax_CC_LE = 0x0E;
extern const ax_u8 ax_CC_G;
const ax_u8 ax_CC_G = 0x0F;
extern const ax_u8 ax_DST_UNUSED;
const ax_u8 ax_DST_UNUSED = 0;
extern const ax_u8 ax_DST_WRITE_ONLY;
const ax_u8 ax_DST_WRITE_ONLY = 1;
extern const ax_u8 ax_DST_READ_WRITE;
const ax_u8 ax_DST_READ_WRITE = 2;
extern const ax_u8 ax_DST_READ_ONLY;
const ax_u8 ax_DST_READ_ONLY = 3;
extern const ax_u8 ax_MOD_INDIRECT;
const ax_u8 ax_MOD_INDIRECT = 0b00;
extern const ax_u8 ax_MOD_DISP8;
const ax_u8 ax_MOD_DISP8 = 0b01;
extern const ax_u8 ax_MOD_DISP32;
const ax_u8 ax_MOD_DISP32 = 0b10;
extern const ax_u8 ax_MOD_REG_DIRECT;
const ax_u8 ax_MOD_REG_DIRECT = 0b11;
extern const ax_u8 ax_REX_BASE;
const ax_u8 ax_REX_BASE = 0x40;
extern const ax_u8 ax_REX_W;
const ax_u8 ax_REX_W = 0x08;
extern const ax_u8 ax_REX_R;
const ax_u8 ax_REX_R = 0x04;
extern const ax_u8 ax_REX_X;
const ax_u8 ax_REX_X = 0x02;
extern const ax_u8 ax_REX_B;
const ax_u8 ax_REX_B = 0x01;
extern const ax_u8 ax_RELOC_PC32;
const ax_u8 ax_RELOC_PC32 = 0;
extern const ax_u8 ax_RELOC_ABS64;
const ax_u8 ax_RELOC_ABS64 = 1;
extern const ax_u8 ax_RELOC_PLT32;
const ax_u8 ax_RELOC_PLT32 = 2;
extern const ax_u8 ax_ELF_CLASS64;
const ax_u8 ax_ELF_CLASS64 = 2;
extern const ax_u8 ax_ELF_DATA2LSB;
const ax_u8 ax_ELF_DATA2LSB = 1;
extern const ax_u8 ax_EV_CURRENT;
const ax_u8 ax_EV_CURRENT = 1;
extern const ax_u16 ax_ET_REL;
const ax_u16 ax_ET_REL = 1;
extern const ax_u16 ax_EM_X86_64;
const ax_u16 ax_EM_X86_64 = 62;
extern const ax_u32 ax_SHT_NULL;
const ax_u32 ax_SHT_NULL = 0;
extern const ax_u32 ax_SHT_PROGBITS;
const ax_u32 ax_SHT_PROGBITS = 1;
extern const ax_u32 ax_SHT_SYMTAB;
const ax_u32 ax_SHT_SYMTAB = 2;
extern const ax_u32 ax_SHT_STRTAB;
const ax_u32 ax_SHT_STRTAB = 3;
extern const ax_u32 ax_SHT_RELA;
const ax_u32 ax_SHT_RELA = 4;
extern const ax_u64 ax_SHF_ALLOC;
const ax_u64 ax_SHF_ALLOC = 0x02;
extern const ax_u64 ax_SHF_EXECINSTR;
const ax_u64 ax_SHF_EXECINSTR = 0x04;
extern const ax_u8 ax_STB_LOCAL;
const ax_u8 ax_STB_LOCAL = 0;
extern const ax_u8 ax_STB_GLOBAL;
const ax_u8 ax_STB_GLOBAL = 1;
extern const ax_u8 ax_STT_OBJECT;
const ax_u8 ax_STT_OBJECT = 1;
extern const ax_u8 ax_STT_FUNC;
const ax_u8 ax_STT_FUNC = 2;
extern const ax_u8 ax_STT_SECTION;
const ax_u8 ax_STT_SECTION = 3;
extern const ax_u16 ax_IMAGE_REL_AMD64_ADDR64;
const ax_u16 ax_IMAGE_REL_AMD64_ADDR64 = 1;
extern const ax_u16 ax_IMAGE_REL_AMD64_REL32;
const ax_u16 ax_IMAGE_REL_AMD64_REL32 = 4;
extern const ax_u8 ax_IMAGE_SYM_CLASS_EXTERNAL;
const ax_u8 ax_IMAGE_SYM_CLASS_EXTERNAL = 2;
extern const ax_u8 ax_IMAGE_SYM_CLASS_STATIC;
const ax_u8 ax_IMAGE_SYM_CLASS_STATIC = 3;

/* Function prototypes */
struct ax_TokenVec ax_new_token_vec(void);
void ax_TokenVec_push(struct ax_TokenVec* self, struct ax_Token t);
struct ax_IntVec ax_new_int_vec(void);
void ax_IntVec_push(struct ax_IntVec* self, ax_i32 val);
ax_i32 ax_IntVec_pop(struct ax_IntVec* self);
ax_i32 ax_IntVec_top(struct ax_IntVec self);
struct ax_Lexer ax_new_lexer(ax_string src);
ax_u8 ax_Lexer_peek1(struct ax_Lexer self);
void ax_Lexer_emit(struct ax_Lexer* self, ax_u8 kind, ax_i64 offset, ax_i64 length);
static ax_u8 ax_lookup_keyword(ax_string s);
static ax_bool ax_is_ident_start(ax_u8 b);
static ax_bool ax_is_ident_continue(ax_u8 b);
void ax_Lexer_scan_ident(struct ax_Lexer* self);
static ax_bool ax_is_dec_digit(ax_u8 b);
static ax_bool ax_is_hex_digit(ax_u8 b);
static ax_bool ax_is_oct_digit(ax_u8 b);
static ax_bool ax_is_bin_digit(ax_u8 b);
void ax_Lexer_scan_dec_digits(struct ax_Lexer* self);
void ax_Lexer_scan_hex_digits(struct ax_Lexer* self);
void ax_Lexer_scan_oct_digits(struct ax_Lexer* self);
void ax_Lexer_scan_bin_digits(struct ax_Lexer* self);
void ax_Lexer_scan_number(struct ax_Lexer* self);
void ax_Lexer_scan_string(struct ax_Lexer* self);
void ax_Lexer_scan_char(struct ax_Lexer* self);
void ax_Lexer_scan_line_comment(struct ax_Lexer* self);
void ax_Lexer_scan_operator_or_punct(struct ax_Lexer* self);
void ax_Lexer_run(struct ax_Lexer* self);
static ax_i64 ax_find_line_start(ax_string src, ax_i64 pos);
ax_i64 ax_Lexer_find_next_content_offset(struct ax_Lexer self, ax_i64 start_idx);
static ax_i64 ax_count_leading_spaces(ax_string src, ax_i64 offset, ax_i64 len);
struct ax_TokenVec ax_Lexer_process_indentation(struct ax_Lexer* self);
struct ax_TokenVec ax_Lexer_tokenize(struct ax_Lexer* self);
struct ax_NodeVec ax_new_node_vec(void);
ax_u32 ax_NodeVec_push(struct ax_NodeVec* self, struct ax_AstNode node);
struct ax_AstTree ax_new_ast_tree(ax_string src, struct ax_TokenVec tokens);
ax_u32 ax_AstTree_add_node(struct ax_AstTree* self, ax_u8 kind, ax_u32 token_idx);
ax_u32 ax_AstTree_add_extra(struct ax_AstTree* self, ax_i32 val);
ax_u32 ax_AstTree_clone_subtree(struct ax_AstTree* self, ax_u32 node_idx);
ax_u32 ax_fnv1a(ax_string s);
struct ax_InternPool ax_new_intern_pool(void);
static ax_string ax_alloc_str_from_raw(ax_u8* data, ax_i64 len);
ax_string ax_InternPool_get_str(struct ax_InternPool self, ax_u32 id);
ax_string ax_InternPool_get(struct ax_InternPool self, ax_u32 id);
ax_u32 ax_InternPool_intern(struct ax_InternPool* self, ax_string s);
ax_u32 ax_InternPool_intern_string(struct ax_InternPool* self, ax_string s);
static ax_u32 ax_InternPool_insert_at(struct ax_InternPool* self, ax_i64 slot, ax_string s, ax_u32 h);
static void ax_InternPool_grow_table(struct ax_InternPool* self);
void ax_InternPool_free_pool(struct ax_InternPool* self);
struct ax_Parser ax_TokenVec_new_parser(struct ax_TokenVec tokens, ax_string src, struct ax_InternPool pool);
struct ax_Token ax_Parser_peek(struct ax_Parser* self);
struct ax_Token ax_Parser_peek_raw(struct ax_Parser self);
struct ax_Token ax_Parser_peek_at(struct ax_Parser self, ax_i64 offset);
struct ax_Token ax_Parser_consume(struct ax_Parser* self);
ax_bool ax_Parser_check(struct ax_Parser* self, ax_u8 kind);
ax_bool ax_Parser_check_raw(struct ax_Parser self, ax_u8 kind);
void ax_Parser_errorf(struct ax_Parser* self, struct ax_Token tok, ax_string msg);
struct ax_Token ax_Parser_expect(struct ax_Parser* self, ax_u8 kind);
void ax_Parser_expect_newline(struct ax_Parser* self);
ax_u32 ax_Parser_token_idx(struct ax_Parser self, struct ax_Token tok);
ax_string ax_Parser_token_text(struct ax_Parser self, struct ax_Token tok);
void ax_Parser_append_child(struct ax_Parser* self, ax_u32 parent, ax_u32 child);
void ax_Parser_set_payload(struct ax_Parser* self, ax_u32 node, ax_u32 payload);
void ax_Parser_set_flags(struct ax_Parser* self, ax_u32 node, ax_u16 flags);
static ax_i32 ax_left_binding_power(ax_u8 kind);
ax_u32 ax_Parser_parse_expr_with_prec(struct ax_Parser* self, ax_i32 min_bp);
ax_u32 ax_Parser_parse_nud(struct ax_Parser* self);
ax_u32 ax_Parser_parse_led(struct ax_Parser* self, ax_u32 left, struct ax_Token op_tok, ax_i32 bp);
ax_u32 ax_Parser_parse_call_args(struct ax_Parser* self, ax_u32 callee, struct ax_Token lparen);
ax_u32 ax_Parser_parse_stmt(struct ax_Parser* self);
ax_u32 ax_Parser_parse_break_stmt(struct ax_Parser* self);
ax_u32 ax_Parser_parse_continue_stmt(struct ax_Parser* self);
ax_u32 ax_Parser_parse_var_decl(struct ax_Parser* self);
ax_u32 ax_Parser_parse_return_stmt(struct ax_Parser* self);
ax_u32 ax_Parser_parse_if_stmt(struct ax_Parser* self);
ax_u32 ax_Parser_parse_while_loop(struct ax_Parser* self);
ax_u32 ax_Parser_parse_for_loop(struct ax_Parser* self);
ax_u32 ax_Parser_parse_match_stmt(struct ax_Parser* self);
ax_u32 ax_Parser_parse_match_arm(struct ax_Parser* self);
ax_u32 ax_Parser_parse_pattern(struct ax_Parser* self);
ax_u32 ax_Parser_parse_expr_stmt(struct ax_Parser* self);
ax_u32 ax_Parser_parse_block(struct ax_Parser* self);
ax_u32 ax_Parser_parse_type_expr(struct ax_Parser* self);
ax_u32 ax_Parser_parse_generic_params(struct ax_Parser* self);
ax_u32 ax_Parser_parse_generic_param(struct ax_Parser* self);
ax_u32 ax_Parser_parse_func_decl(struct ax_Parser* self, ax_bool is_pub);
ax_u32 ax_Parser_parse_struct_decl(struct ax_Parser* self, ax_bool is_pub);
ax_u32 ax_Parser_parse_field_decl(struct ax_Parser* self, ax_bool is_pub);
ax_u32 ax_Parser_parse_type_alias_decl(struct ax_Parser* self, ax_bool is_pub);
ax_u32 ax_Parser_parse_type_variant(struct ax_Parser* self);
void ax_Parser_parse_program(struct ax_Parser* self);
ax_u32 ax_Parser_parse_param(struct ax_Parser* self);
ax_u32 ax_Parser_parse_const_decl(struct ax_Parser* self, ax_bool is_pub);
ax_u32 ax_Parser_parse_extern_decl(struct ax_Parser* self, ax_bool is_pub);
ax_u32 ax_Parser_parse_import_decl(struct ax_Parser* self);
void ax_Scope_init_scope(struct ax_Scope* self, ax_u8 kind, ax_u32 parent_id, ax_u32 depth, ax_u32 capacity);
static ax_u32 ax_hash_fnv1a(ax_u32 v);
void ax_Scope_scope_put(struct ax_Scope* self, ax_u32 name_id, ax_u32 symbol_idx);
static void ax_Scope_scope_insert(struct ax_Scope self, ax_u32 name_id, ax_u32 symbol_idx);
struct ax_ScopeEntry ax_Scope_scope_get(struct ax_Scope self, ax_u32 name_id);
static void ax_Scope_scope_grow(struct ax_Scope* self);
struct ax_SymbolVec ax_new_symbol_vec(void);
ax_u32 ax_SymbolVec_push(struct ax_SymbolVec* self, struct ax_Symbol sym);
struct ax_ScopeVec ax_new_scope_vec(void);
ax_u32 ax_ScopeVec_push(struct ax_ScopeVec* self, struct ax_Scope scope);
struct ax_U32Vec ax_new_u32_vec(void);
ax_u32 ax_U32Vec_push(struct ax_U32Vec* self, ax_u32 val);
ax_u32 ax_U32Vec_pop(struct ax_U32Vec* self);
ax_u32 ax_U32Vec_append_unique(struct ax_U32Vec* self, ax_u32 val);
struct ax_SymbolTable ax_InternPool_new_symbol_table(struct ax_InternPool* intern);
void ax_SymbolTable_define_builtin(struct ax_SymbolTable* self, ax_string name, ax_u32 type_id, struct ax_InternPool* intern);
ax_u32 ax_SymbolTable_push_scope(struct ax_SymbolTable* self, ax_u8 kind);
void ax_SymbolTable_pop_scope(struct ax_SymbolTable* self);
ax_u32 ax_SymbolTable_current_scope(struct ax_SymbolTable self);
ax_u32 ax_SymbolTable_current_depth(struct ax_SymbolTable self);
ax_u32 ax_SymbolTable_define(struct ax_SymbolTable* self, ax_u32 name_id, ax_u8 kind, ax_u16 flags, ax_u32 decl_node);
ax_u32 ax_SymbolTable_resolve(struct ax_SymbolTable self, ax_u32 name_id);
struct ax_NameResolver ax_AstTree_new_name_resolver(struct ax_AstTree tree, struct ax_InternPool intern, struct ax_SymbolTable symtable);
void ax_NameResolver_resolve(struct ax_NameResolver* self);
void ax_NameResolver_resolve_children(struct ax_NameResolver* self, ax_u32 node_idx);
void ax_NameResolver_resolve_node(struct ax_NameResolver* self, ax_u32 node_idx);
void ax_SymbolTable_free_symtable(struct ax_SymbolTable* self);
void ax_StructFieldVec_push(struct ax_StructFieldVec* self, struct ax_StructField sf);
void ax_StructInfoVec_push(struct ax_StructInfoVec* self, struct ax_StructInfo si);
void ax_VariantInfoVec_push(struct ax_VariantInfoVec* self, struct ax_VariantInfo vi);
void ax_SumInfoVec_push(struct ax_SumInfoVec* self, struct ax_SumInfo si);
void ax_FuncInfoVec_push(struct ax_FuncInfoVec* self, struct ax_FuncInfo fi);
ax_u32 ax_TypeEntryVec_push(struct ax_TypeEntryVec* self, struct ax_TypeEntry te);
struct ax_SumInfoVec ax_new_sum_info_vec(void);
struct ax_VariantInfoVec ax_new_variant_info_vec(void);
struct ax_TypeTable ax_new_type_table(void);
ax_u32 ax_TypeTable_register_sum_type(struct ax_TypeTable* self, ax_u32 name_id, struct ax_VariantInfoVec variants);
ax_u32 ax_TypeTable_register_struct(struct ax_TypeTable* self, ax_u32 name_id, struct ax_StructFieldVec fields);
ax_u32 ax_TypeTable_register_function(struct ax_TypeTable* self, struct ax_U32Vec params, ax_u32 ret);
ax_u32 ax_TypeTable_register_pointer(struct ax_TypeTable* self, ax_u32 inner_id);
ax_u32 ax_TypeTable_register_slice(struct ax_TypeTable* self, ax_u32 elem_id);
ax_u32 ax_TypeTable_register_array(struct ax_TypeTable* self, ax_u32 elem_id);
void ax_TypeTable_free_typetable(struct ax_TypeTable* self);
struct ax_TypeSubstVec ax_new_type_subst_vec(void);
void ax_TypeSubstVec_push(struct ax_TypeSubstVec* self, struct ax_TypeSubst subst);
ax_u32 ax_TypeSubstVec_lookup_subst(struct ax_TypeSubstVec self, ax_u32 name_id);
struct ax_Monomorphizer ax_AstTree_new_monomorphizer(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable);
void ax_Monomorphizer_substitute_type_params(struct ax_Monomorphizer* self, ax_u32 node_idx, struct ax_TypeSubstVec subst);
void ax_Monomorphizer_remove_generic_params_child(struct ax_Monomorphizer* self, ax_u32 node_idx);
ax_string ax_Monomorphizer_mangle_name(struct ax_Monomorphizer self, ax_string orig_name, struct ax_U32Vec args);
ax_u32 ax_Monomorphizer_instantiate_function(struct ax_Monomorphizer* self, ax_u32 template_sym_id, struct ax_U32Vec args);
ax_string ax_strip_quotes(ax_string s);
ax_string ax_format_int(ax_i64 val);
ax_string ax_format_float(ax_f64 val);
struct ax_TypeChecker ax_AstTree_new_type_checker(struct ax_AstTree tree, struct ax_InternPool intern, struct ax_SymbolTable symtable, struct ax_TypeTable types);
void ax_TypeChecker_free_type_checker(struct ax_TypeChecker* self);
void ax_TypeChecker_set_node_type(struct ax_TypeChecker* self, ax_u32 node_idx, ax_u32 type_id);
ax_i64 ax_parse_comptime_int(ax_string s);
ax_string ax_TypeChecker_node_text(struct ax_TypeChecker self, ax_u32 node_idx);
struct ax_ComptimeValue ax_TypeChecker_eval_comptime(struct ax_TypeChecker* self, ax_u32 node_idx);
void ax_TypeChecker_run_type_checker(struct ax_TypeChecker* self);
void ax_TypeChecker_pre_infer_type_alias(struct ax_TypeChecker* self, ax_u32 node_idx);
void ax_TypeChecker_pre_infer_struct(struct ax_TypeChecker* self, ax_u32 node_idx);
void ax_TypeChecker_pre_infer_func_signature(struct ax_TypeChecker* self, ax_u32 node_idx);
ax_u32 ax_TypeChecker_infer_node(struct ax_TypeChecker* self, ax_u32 node_idx, ax_u32 expected);
struct ax_CGNodeVec ax_new_cg_node_vec(void);
ax_u32 ax_CGNodeVec_push(struct ax_CGNodeVec* self, struct ax_CGNode n);
struct ax_CGEdgeVec ax_new_cg_edge_vec(void);
ax_u32 ax_CGEdgeVec_push(struct ax_CGEdgeVec* self, struct ax_CGEdge e);
struct ax_U32VecVec ax_new_u32_vec_vec(void);
ax_u32 ax_U32VecVec_push(struct ax_U32VecVec* self, struct ax_U32Vec v);
struct ax_ConnectionGraph ax_new_connection_graph(void);
void ax_ConnectionGraph_free_connection_graph(struct ax_ConnectionGraph* self);
void ax_ConnectionGraph_ensure_adj_capacity(struct ax_ConnectionGraph* self, ax_u32 node_id);
ax_u32 ax_ConnectionGraph_add_value_node(struct ax_ConnectionGraph* self, ax_u32 sym_id, ax_u32 type_id, ax_u32 lifetime);
ax_u32 ax_ConnectionGraph_add_ref_node(struct ax_ConnectionGraph* self, ax_u32 target_node_id);
void ax_ConnectionGraph_add_edge(struct ax_ConnectionGraph* self, ax_u32 from, ax_u32 to, ax_u32 kind);
ax_u32 ax_ConnectionGraph_node_of_sym(struct ax_ConnectionGraph self, ax_u32 sym_id);
ax_bool ax_ConnectionGraph_escapes(struct ax_ConnectionGraph self, ax_u32 node_id);
ax_bool ax_ConnectionGraph_escape_dfs(struct ax_ConnectionGraph self, ax_u32 node_id, ax_bool* visited);
struct ax_OwnershipChecker ax_AstTree_new_ownership_checker(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable);
void ax_OwnershipChecker_check(struct ax_OwnershipChecker* self);
void ax_OwnershipChecker_check_node(struct ax_OwnershipChecker* self, ax_u32 node_idx);
void ax_OwnershipChecker_check_move(struct ax_OwnershipChecker* self, ax_u32 node_idx);
struct ax_EscapeAnalyser ax_AstTree_new_escape_analyser(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable);
void ax_EscapeAnalyser_run(struct ax_EscapeAnalyser* self);
void ax_EscapeAnalyser_traverse_nodes(struct ax_EscapeAnalyser* self, ax_u32 node_idx);
void ax_EscapeAnalyser_analyze_block(struct ax_EscapeAnalyser* self, ax_u32 block_idx);
void ax_EscapeAnalyser_analyze_stmt(struct ax_EscapeAnalyser* self, ax_u32 stmt_idx);
void ax_EscapeAnalyser_analyze_expr(struct ax_EscapeAnalyser* self, ax_u32 expr_idx, ax_u32 flow_dest);
struct ax_CtgcInjector ax_AstTree_new_ctgc_injector(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable);
void ax_CtgcInjector_run(struct ax_CtgcInjector* self);
void ax_CtgcInjector_traverse_and_inject(struct ax_CtgcInjector* self, ax_u32 node_idx, ax_u32 parent_idx);
void ax_CtgcInjector_append_child_node(struct ax_CtgcInjector* self, ax_u32 parent, ax_u32 child);
void ax_CtgcInjector_insert_before(struct ax_CtgcInjector* self, ax_u32 parent, ax_u32 target, ax_u32 new_node);
struct ax_AliasReuseOptimizer ax_AstTree_new_alias_reuse_optimizer(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable);
void ax_AliasReuseOptimizer_run(struct ax_AliasReuseOptimizer* self);
void ax_AliasReuseOptimizer_optimize_node(struct ax_AliasReuseOptimizer* self, ax_u32 node_idx);
struct ax_AirInstVec ax_new_air_inst_vec(void);
ax_u32 ax_AirInstVec_push(struct ax_AirInstVec* self, struct ax_AirInst inst);
struct ax_BasicBlockVec ax_new_basic_block_vec(void);
ax_u32 ax_BasicBlockVec_push(struct ax_BasicBlockVec* self, struct ax_BasicBlock bb);
void ax_AirFunc_free_air_func(struct ax_AirFunc* f);
struct ax_AirFuncVec ax_new_air_func_vec(void);
ax_u32 ax_AirFuncVec_push(struct ax_AirFuncVec* self, struct ax_AirFunc f);
void ax_AirModule_free_air_module(struct ax_AirModule* m);
struct ax_AirFuncBuilder ax_new_air_func_builder(ax_u32 name, ax_u32 ret_type);
ax_u32 ax_AirFuncBuilder_new_block(struct ax_AirFuncBuilder* self);
void ax_AirFuncBuilder_switch_to(struct ax_AirFuncBuilder* self, ax_u32 block_id);
ax_u32 ax_AirFuncBuilder_current_block(struct ax_AirFuncBuilder self);
ax_u32 ax_AirFuncBuilder_emit(struct ax_AirFuncBuilder* self, struct ax_AirInst inst);
ax_u32 ax_AirFuncBuilder_emit_extra(struct ax_AirFuncBuilder* self, ax_u32 val);
void ax_AirFuncBuilder_set_extra(struct ax_AirFuncBuilder* self, ax_u32 idx, ax_u32 val);
ax_u32 ax_AirFuncBuilder_fresh_reg(struct ax_AirFuncBuilder* self);
void ax_AirFuncBuilder_add_edge(struct ax_AirFuncBuilder* self, ax_u32 src, ax_u32 dst);
struct ax_AirFunc ax_AirFuncBuilder_build_func(struct ax_AirFuncBuilder* self);
ax_string ax_opcode_mnemonic(ax_u16 op);
ax_u16 ax_opcode_class(ax_u16 op);
ax_bool ax_opcode_is_binary_alu(ax_u16 op);
struct ax_LocalMap ax_new_local_map(void);
void ax_LocalMap_free_local_map(struct ax_LocalMap* self);
static ax_u32 ax_local_map_hash(ax_u32 v);
void ax_LocalMap_local_map_put(struct ax_LocalMap* self, ax_u32 name_id, ax_u32 reg);
static void ax_LocalMap_local_map_insert(struct ax_LocalMap self, ax_u32 name_id, ax_u32 reg);
ax_u32 ax_LocalMap_local_map_get(struct ax_LocalMap self, ax_u32 name_id);
static void ax_LocalMap_local_map_grow(struct ax_LocalMap* self);
struct ax_AirModuleBuilder ax_AstTree_new_air_module_builder(struct ax_AstTree tree, struct ax_SymbolTable symbols, struct ax_TypeTable typetable, struct ax_InternPool pool, ax_u32* node_types);
void ax_AirModuleBuilder_free_air_module_builder(struct ax_AirModuleBuilder* self);
ax_string ax_AirModuleBuilder_get_token_text(struct ax_AirModuleBuilder* self, ax_u32 token_idx);
ax_i64 ax_parse_int_from_str(ax_string s);
ax_u16 ax_map_binary_op(ax_string op);
struct ax_FuncLowering ax_AirModuleBuilder_new_func_lowering(struct ax_AirModuleBuilder* mb, struct ax_AirFuncBuilder fb, struct ax_U32Vec params);
void ax_FuncLowering_free_func_lowering(struct ax_FuncLowering* self);
void ax_FuncLowering_register_params(struct ax_FuncLowering* self, ax_u32 func_idx, struct ax_AstNode func_node);
ax_u32 ax_FuncLowering_lower_expr(struct ax_FuncLowering* self, ax_u32 idx);
ax_u32 ax_FuncLowering_lower_int_lit(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_float_lit(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_bool_lit(struct ax_FuncLowering* self, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_nil_lit(struct ax_FuncLowering* self);
ax_u32 ax_FuncLowering_lower_string_lit(struct ax_FuncLowering* self, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_char_lit(struct ax_FuncLowering* self, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_ident(struct ax_FuncLowering* self, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_binary_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_unary_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_call_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_field_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_index_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_cast_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_deref_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_spawn_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_await_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_struct_lit(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_array_lit(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
ax_u32 ax_FuncLowering_lower_struct_constructor_call(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node, ax_u32 type_id);
ax_u32 ax_FuncLowering_emit_heap_alloc(struct ax_FuncLowering* self, ax_u16 type_id);
ax_u32 ax_FuncLowering_emit_deref(struct ax_FuncLowering* self, ax_u32 ref_reg, ax_u16 type_id);
ax_u32 ax_FuncLowering_emit_move(struct ax_FuncLowering* self, ax_u32 src_reg, ax_u16 type_id);
void ax_FuncLowering_emit_free(struct ax_FuncLowering* self, ax_u32 ptr_reg);
ax_u32 ax_FuncLowering_emit_arena_alloc(struct ax_FuncLowering* self, ax_u32 arena_reg, ax_u16 type_id);
ax_u32 ax_FuncLowering_emit_alias_reuse(struct ax_FuncLowering* self, ax_u32 ptr_reg, ax_u16 type_id);
ax_u32 ax_FuncLowering_lower_ownership_aware(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node, ax_u32 init_reg);
void ax_FuncLowering_lower_alias_stmt(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_lower_block(struct ax_FuncLowering* self, ax_u32 block_idx);
void ax_FuncLowering_lower_stmt(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_lower_var_decl(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_lower_assign(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_lower_return(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_lower_if(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_lower_while(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_lower_for(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_lower_destroy(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_lower_defer(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node);
void ax_FuncLowering_ensure_return(struct ax_FuncLowering* self);
struct ax_AirFunc* ax_AirModuleBuilder_lower_func(struct ax_AirModuleBuilder* self, ax_u32 idx, struct ax_AstNode node);
void ax_AirModuleBuilder_build_module(struct ax_AirModuleBuilder* self);
ax_u8* ax_AirModuleBuilder_builder_str_to_null_terminated(struct ax_AirModuleBuilder* self, ax_string s);
struct ax_SsaOptimizer ax_new_ssa_optimizer(void);
ax_u32 ax_AirFunc_max_reg_id(struct ax_AirFunc* f);
ax_bool ax_is_unary_foldable(ax_u16 op);
ax_bool ax_has_side_effect(ax_u16 op);
ax_bool ax_opcode_is_control(ax_u16 op);
ax_bool ax_eval_binary(ax_u16 op, ax_u32 a, ax_u32 b, ax_u32* out_val);
ax_bool ax_eval_unary(ax_u16 op, ax_u32 a, ax_u32* out_val);
ax_bool ax_AirFunc_fold_func(struct ax_AirFunc* f);
ax_bool ax_AirFunc_copy_prop_func(struct ax_AirFunc* f);
ax_bool ax_AirFunc_remove_unreachable_blocks(struct ax_AirFunc* f);
ax_bool ax_AirFunc_dce_func(struct ax_AirFunc* f);
void ax_SsaOptimizer_run(struct ax_SsaOptimizer* self, struct ax_AirModule* m);
static ax_bool ax_opcode_defines_dest(ax_u16 op);
struct ax_CGenerator ax_AirModule_new_c_generator(struct ax_AirModule mod, struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable);
ax_string ax_CGenerator_get_c_type(struct ax_CGenerator self, ax_u32 type_id);
ax_bool ax_CGenerator_is_stdlib_func(struct ax_CGenerator self, ax_string name);
ax_string ax_CGenerator_get_mangled_name_by_sym(struct ax_CGenerator self, ax_u32 sym_idx);
static ax_u8* ax_str_to_null_terminated(ax_string s);
ax_bool ax_CGenerator_generate(struct ax_CGenerator* self, ax_string out_filename);
void ax_CGenerator_translate_inst(struct ax_CGenerator* self, struct ax_AirFunc f, struct ax_AirInst inst);
static ax_bool ax_is_type_unsigned(ax_u32 type_id);
static ax_string ax_TypeTable_map_wasm_type(struct ax_TypeTable typetable, ax_u32 type_id);
static ax_bool ax_wasm_opcode_defines_dest(ax_u16 op);
struct ax_WasmGenerator ax_AirModule_new_wasm_generator(struct ax_AirModule mod, struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable);
ax_string ax_WasmGenerator_resolve_sym_name(struct ax_WasmGenerator self, ax_u32 sym_id, ax_u32 name_id);
static ax_u8* ax_wasm_str_to_null_terminated(ax_string s);
ax_bool ax_WasmGenerator_generate(struct ax_WasmGenerator* self, ax_string out_filename);
void ax_WasmGenerator_compile_func(struct ax_WasmGenerator* self, struct ax_AirFunc f);
void ax_WasmGenerator_lower_inst(struct ax_WasmGenerator* self, struct ax_AirFunc f, struct ax_AirInst inst);
ax_bool ax_reg_is_gpr(ax_u8 r);
ax_u8 ax_reg_hw_reg(ax_u8 r);
ax_bool ax_reg_needs_rex(ax_u8 r);
ax_string ax_reg_to_str(ax_u8 r);
ax_u8 ax_get_sysv_arg_reg(ax_i64 idx);
ax_u8 ax_get_win64_arg_reg(ax_i64 idx);
ax_bool ax_reg_is_sysv_caller_saved(ax_u8 r);
ax_bool ax_reg_is_sysv_callee_saved(ax_u8 r);
ax_bool ax_reg_is_win64_callee_saved(ax_u8 r);
ax_string ax_x86_resolve_sym_name(ax_u32 sym_idx, struct ax_SymbolTable symbols, struct ax_InternPool pool);
ax_string ax_cond_code_to_str(ax_u8 cc);
struct ax_MachInstVec ax_new_mach_inst_vec(void);
ax_u32 ax_MachInstVec_push(struct ax_MachInstVec* self, struct ax_MachInst inst);
void ax_TypeTable_type_size_and_align(struct ax_TypeTable table, ax_u32 type_id, ax_u32* out_size, ax_u32* out_align);
ax_u32 ax_TypeTable_field_offset(struct ax_TypeTable table, ax_u32 struct_type_id, ax_u32 field_idx);
ax_u8 ax_abi_int_arg_reg(ax_string abi, ax_i64 idx);
ax_u8 ax_abi_return_reg(ax_string abi);
ax_u32 ax_InstructionSelector_next_vreg(struct ax_InstructionSelector* sel);
ax_u32 ax_InstructionSelector_next_label(struct ax_InstructionSelector* sel);
ax_u32 ax_InstructionSelector_get_register_type(struct ax_InstructionSelector* sel, ax_u32 reg);
void ax_AirInst_select_cmp(struct ax_AirInst* inst, ax_u8 cc, struct ax_MachInstVec* out_insts);
void ax_InstructionSelector_select_inst(struct ax_InstructionSelector* sel, struct ax_AirInst* inst, struct ax_MachInstVec* out_insts);
struct ax_MachInstVec ax_AirFunc_select_all(struct ax_AirFunc* f, ax_string abi, struct ax_TypeTable table, struct ax_SymbolTable symbols, struct ax_InternPool pool);
struct ax_LiveIntervalVec ax_new_live_interval_vec(void);
ax_u32 ax_LiveIntervalVec_push(struct ax_LiveIntervalVec* self, struct ax_LiveInterval iv);
ax_bool ax_is_two_operand_read(ax_u16 op);
struct ax_LiveIntervalVec ax_MachInst_compute_liveness(struct ax_MachInst* insts, ax_i64 insts_len);
ax_u8* ax_get_allocatable_gprs(ax_i64* out_len);
struct ax_RegAllocResult ax_LiveIntervalVec_graph_coloring_alloc(struct ax_LiveIntervalVec intervals, ax_u8* avail_regs, ax_i64 avail_regs_len);
ax_u8* ax_get_used_callee_saved(ax_string abi, struct ax_RegAllocation* allocs, ax_u32 max_vreg, ax_i64* out_len);
struct ax_StackFrame ax_compute_frame(ax_u8* callee_saved, ax_i64 callee_saved_len, ax_i32 spill_count, ax_i32 local_bytes);
void ax_StackFrame_emit_prologue(struct ax_StackFrame* frame, struct ax_MachInstVec* out_insts);
void ax_StackFrame_emit_epilogue(struct ax_StackFrame* frame, struct ax_MachInstVec* out_insts);
ax_u8 ax_get_dst_behavior(ax_u16 op);
struct ax_MachInstVec ax_MachInst_insert_spill_code(struct ax_MachInst* insts, ax_i64 insts_len, struct ax_RegAllocation* allocs, struct ax_StackFrame* frame);
ax_string ax_to_byte_reg(ax_u8 reg);
ax_string ax_MachOperand_format_operand(struct ax_MachOperand op, ax_string format, struct ax_RegAllocation* allocs);
void ax_emit_inst(void* file, struct ax_MachInst inst, ax_string fn_name, struct ax_RegAllocation* allocs, ax_string format, struct ax_SymbolTable symbols, struct ax_InternPool pool);
void ax_emit_function(void* file, ax_string fn_name, struct ax_MachInst* insts, ax_i64 insts_len, struct ax_RegAllocation* allocs, struct ax_StackFrame* frame, ax_string format, struct ax_SymbolTable symbols, struct ax_InternPool pool);
static ax_u8* ax_x86_str_to_null_terminated(ax_string s);
ax_bool ax_compile_native_asm(ax_string output_asm_file, struct ax_AirModule mod, struct ax_SymbolTable symbols, struct ax_InternPool pool, struct ax_TypeTable table, ax_string format);
struct ax_ByteVec ax_new_byte_vec(void);
void ax_ByteVec_push_byte(struct ax_ByteVec* self, ax_u8 b);
void ax_ByteVec_push_bytes(struct ax_ByteVec* self, ax_u8* bytes, ax_i64 bytes_len);
void ax_ByteVec_push_u16_le(struct ax_ByteVec* self, ax_u16 val);
void ax_ByteVec_push_u32_le(struct ax_ByteVec* self, ax_u32 val);
void ax_ByteVec_push_u64_le(struct ax_ByteVec* self, ax_u64 val);
ax_u8 ax_encode_rex(ax_bool w, ax_bool r, ax_bool x, ax_bool b);
void ax_x86_encode_modrm_rr(ax_u8 reg, ax_u8 rm, ax_u8* out_modrm, ax_u8* out_rex, ax_bool* out_need_rex);
void ax_x86_encode_modrm_rm(ax_u8 reg, ax_u8 base, ax_i32 disp, struct ax_ByteVec* buf);
void ax_x86_encode_modrm_rip(ax_u8 reg, ax_i32 disp32, struct ax_ByteVec* buf);
void ax_x86_encode_modrm_sib(ax_u8 reg, ax_u8 base, ax_u8 index, ax_u8 scale, ax_i32 disp, struct ax_ByteVec* buf);
void ax_ByteVec_x86_encode_ret(struct ax_ByteVec* buf);
void ax_ByteVec_x86_encode_nop(struct ax_ByteVec* buf);
void ax_ByteVec_x86_encode_int3(struct ax_ByteVec* buf);
void ax_x86_encode_push(ax_u8 reg, struct ax_ByteVec* buf);
void ax_x86_encode_pop(ax_u8 reg, struct ax_ByteVec* buf);
void ax_x86_encode_mov_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_x86_encode_mov_ri(ax_u8 dst, ax_i32 imm, struct ax_ByteVec* buf);
void ax_x86_encode_mov_ri64(ax_u8 dst, ax_i64 imm, struct ax_ByteVec* buf);
void ax_x86_encode_add_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_x86_encode_add_ri(ax_u8 dst, ax_i32 imm, struct ax_ByteVec* buf);
void ax_x86_encode_sub_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_x86_encode_sub_ri(ax_u8 dst, ax_i32 imm, struct ax_ByteVec* buf);
void ax_x86_encode_imul_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_ByteVec_x86_encode_cqo(struct ax_ByteVec* buf);
void ax_x86_encode_idiv_r(ax_u8 divisor, struct ax_ByteVec* buf);
void ax_x86_encode_neg_r(ax_u8 reg, struct ax_ByteVec* buf);
void ax_x86_encode_not_r(ax_u8 reg, struct ax_ByteVec* buf);
void ax_x86_encode_cmp_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_x86_encode_cmp_ri(ax_u8 reg, ax_i32 imm, struct ax_ByteVec* buf);
void ax_x86_encode_test_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_x86_encode_and_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_x86_encode_or_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_x86_encode_xor_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_x86_encode_shl_cl(ax_u8 reg, struct ax_ByteVec* buf);
void ax_x86_encode_sar_cl(ax_u8 reg, struct ax_ByteVec* buf);
void ax_x86_encode_xor_zero(ax_u8 reg, struct ax_ByteVec* buf);
void ax_x86_encode_setcc(ax_u8 cc, ax_u8 dst, struct ax_ByteVec* buf);
void ax_x86_encode_movzx_br(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf);
void ax_x86_encode_jmp_rel32(ax_i32 rel32, struct ax_ByteVec* buf);
void ax_x86_encode_jcc_rel32(ax_u8 cc, ax_i32 rel32, struct ax_ByteVec* buf);
void ax_x86_encode_call_rel32(ax_i32 rel32, struct ax_ByteVec* buf);
void ax_x86_encode_call_r(ax_u8 reg, struct ax_ByteVec* buf);
void ax_x86_encode_lea(ax_u8 dst, ax_u8 base, ax_i32 disp, struct ax_ByteVec* buf);
void ax_x86_encode_mov_load(ax_u8 dst, ax_u8 base, ax_i32 disp, struct ax_ByteVec* buf);
void ax_x86_encode_mov_store(ax_u8 base, ax_i32 disp, ax_u8 src, struct ax_ByteVec* buf);
void ax_ByteVec_x86_encode_syscall(struct ax_ByteVec* buf);
void ax_x86_encode_lea_rip(ax_u8 dst, ax_i32 disp32, struct ax_ByteVec* buf);
void ax_x86_encode_shl_imm(ax_u8 reg, ax_u8 imm, struct ax_ByteVec* buf);
void ax_x86_encode_sar_imm(ax_u8 reg, ax_u8 imm, struct ax_ByteVec* buf);
struct ax_RelocationVec ax_new_relocation_vec(void);
void ax_RelocationVec_push_reloc(struct ax_RelocationVec* self, struct ax_Relocation r);
struct ax_FixupVec ax_new_fixup_vec(void);
void ax_FixupVec_push_fixup(struct ax_FixupVec* self, struct ax_Fixup f);
struct ax_LabelMap ax_new_label_map(void);
void ax_LabelMap_label_map_set(struct ax_LabelMap* self, ax_u32 key, ax_i64 val);
ax_bool ax_LabelMap_label_map_get(struct ax_LabelMap* self, ax_u32 key, ax_i64* out_val);
struct ax_MachEmitter ax_new_mach_emitter(void);
void ax_MachEmitter_free_mach_emitter(struct ax_MachEmitter* self);
ax_u8 ax_MachOperand_emitter_resolve_reg(struct ax_MachOperand op, struct ax_RegAllocation* allocs);
void ax_MachEmitter_emit_mach_inst(struct ax_MachEmitter* e, struct ax_MachInst inst, struct ax_RegAllocation* allocs);
void ax_MachEmitter_emitter_resolve_fixups(struct ax_MachEmitter* e);
void ax_MachEmitter_emit_function_binary(struct ax_MachEmitter* e, ax_string fn_name, struct ax_MachInst* insts, ax_i64 insts_len, struct ax_RegAllocation* allocs, struct ax_StackFrame* frame);
struct ax_ELF64SymVec ax_new_elf64_sym_vec(void);
void ax_ELF64SymVec_push_elf64_sym(struct ax_ELF64SymVec* self, struct ax_ELF64Sym s);
void ax_ByteVec_elf64_serialize(struct ax_ByteVec* code, struct ax_RelocationVec* relocs, struct ax_ELF64SymVec* symbols, struct ax_ByteVec* out_bytes);
ax_bool ax_elf64_write_object_file(ax_string filename, struct ax_ByteVec* code, struct ax_RelocationVec* relocs, struct ax_ELF64SymVec* symbols);
struct ax_COFFRelocVec ax_new_coff_reloc_vec(void);
void ax_COFFRelocVec_push_coff_reloc(struct ax_COFFRelocVec* self, struct ax_COFFReloc r);
struct ax_COFFSymbolVec ax_new_coff_symbol_vec(void);
ax_i32 ax_COFFSymbolVec_push_coff_symbol(struct ax_COFFSymbolVec* self, struct ax_COFFSymbol s);
void ax_ByteVec_coff_serialize(struct ax_ByteVec* code, struct ax_ByteVec* rdata, struct ax_COFFRelocVec* relocs, struct ax_COFFSymbolVec* symbols, struct ax_ByteVec* out_bytes);
ax_bool ax_coff_write_object_file(ax_string filename, struct ax_ByteVec* code, struct ax_ByteVec* rdata, struct ax_COFFRelocVec* relocs, struct ax_COFFSymbolVec* symbols);
struct ax_CompiledFuncInfoVec ax_new_compiled_func_info_vec(void);
void ax_CompiledFuncInfoVec_push_compiled_func_info(struct ax_CompiledFuncInfoVec* self, struct ax_CompiledFuncInfo info);
ax_string ax_resolve_binary_sym_name(ax_i64 sym_idx, ax_string current_fn_name, struct ax_SymbolTable symbols, struct ax_InternPool pool);
ax_bool ax_compile_native_binary(ax_string output_obj_file, struct ax_AirModule mod, struct ax_SymbolTable symbols, struct ax_InternPool pool, struct ax_TypeTable table, ax_string format);
ax_string ax_x86_slice_to_str(ax_u8* data, ax_i64 len);
ax_u16 ax_read_u16_le(ax_u8* data, ax_i64 off);
ax_u32 ax_read_u32_le(ax_u8* data, ax_i64 off);
ax_u64 ax_read_u64_le(ax_u8* data, ax_i64 off);
void ax_write_u32_le(ax_u8* data, ax_i64 off, ax_u32 val);
void ax_write_u64_le(ax_u8* data, ax_i64 off, ax_u64 val);
struct ax_LinkerSymbolVec ax_new_linker_symbol_vec(void);
void ax_LinkerSymbolVec_push_linker_symbol(struct ax_LinkerSymbolVec* self, struct ax_LinkerSymbol s);
struct ax_ParsedRelocVec ax_new_parsed_reloc_vec(void);
void ax_ParsedRelocVec_push_parsed_reloc(struct ax_ParsedRelocVec* self, struct ax_ParsedReloc r);
struct ax_LinkerStrVec ax_new_linker_str_vec(void);
void ax_LinkerStrVec_push_linker_str(struct ax_LinkerStrVec* self, ax_string s);
struct ax_ParsedObjectPtrVec ax_new_parsed_object_ptr_vec(void);
void ax_ParsedObjectPtrVec_push_parsed_object_ptr(struct ax_ParsedObjectPtrVec* self, struct ax_ParsedObject* p);
struct ax_ParsedObject* ax_linker_parse_elf(ax_u8* data, ax_i64 data_len);
struct ax_ParsedObject* ax_linker_parse_coff(ax_u8* data, ax_i64 data_len);
struct ax_AxiomLinker ax_new_axiom_linker(void);
void ax_ByteVec_align_byte_vec(struct ax_ByteVec* vec, ax_i64 alignment);
ax_string ax_get_dll_for_symbol(ax_string name);
void ax_LinkerStrVec_add_unique_import(struct ax_LinkerStrVec* imports, ax_string name);
void ax_write_buf_u32_le(ax_u8* data, ax_i64 off, ax_u32 val);
void ax_write_buf_u64_le(ax_u8* data, ax_i64 off, ax_u64 val);
void ax_ByteVec_linker_build_pe_headers(struct ax_ByteVec* out, ax_u32 code_raw_size, ax_u32 idata_raw_size, ax_u32 entry_rva, ax_u32 idata_rva, ax_u32 idata_size, ax_u32 iat_rva, ax_u32 iat_size);
void ax_ByteVec_linker_build_elf_headers(struct ax_ByteVec* out, ax_u32 code_raw_size, ax_u64 entry_va, ax_u32 dynamic_offset, ax_u32 dynamic_size);
void ax_AxiomLinker_axiom_linker_add_input(struct ax_AxiomLinker* self, ax_string file);
ax_bool ax_AxiomLinker_axiom_linker_link(struct ax_AxiomLinker* self);
static void ax_AirInst_print_inst(struct ax_AirInst inst);
static void ax_AirFunc_print_block(struct ax_AirFunc f, struct ax_BasicBlock* bb);
static void ax_AirFunc_print_func(struct ax_AirFunc f);
static void ax_AirModule_print_module(struct ax_AirModule mod);
static ax_string ax_read_file_content(ax_string path);
static ax_bool ax_match_prefix(ax_string s, ax_i64 i, ax_string prefix);
static ax_string ax_strip_imports(ax_string s);
ax_i32 ax_main_usr(ax_i32 argc, ax_u8** argv);


struct ax_TokenVec ax_new_token_vec(void) {
    return ((struct ax_TokenVec){.data=((struct ax_Token*)(NULL)), .len=0, .cap=0});
}

void ax_TokenVec_push(struct ax_TokenVec* self, struct ax_Token t) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_Token* new_data = ((struct ax_Token*)(ax_alloc((new_cap * 8))));
        if ((self->data != ((struct ax_Token*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 8));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = t;
    self->len = (self->len + 1);
}

struct ax_IntVec ax_new_int_vec(void) {
    return ((struct ax_IntVec){.data=((ax_i32*)(NULL)), .len=0, .cap=0});
}

void ax_IntVec_push(struct ax_IntVec* self, ax_i32 val) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        ax_i32* new_data = ((ax_i32*)(ax_alloc((new_cap * 4))));
        if ((self->data != ((ax_i32*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 4));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = val;
    self->len = (self->len + 1);
}

ax_i32 ax_IntVec_pop(struct ax_IntVec* self) {
    if ((self->len == 0)) {
        return 0;
    }
    self->len = (self->len - 1);
    return ((self->data)[self->len]);
}

ax_i32 ax_IntVec_top(struct ax_IntVec self) {
    if ((self.len == 0)) {
        return 0;
    }
    return ((self.data)[(self.len - 1)]);
}

struct ax_Lexer ax_new_lexer(ax_string src) {
    struct ax_IntVec indent_stack = ax_new_int_vec();
    ax_IntVec_push(&(indent_stack), 0);
    return ((struct ax_Lexer){.src=src, .len=ax_str_len(src), .pos=0, .tokens=ax_new_token_vec(), .indent_stack=indent_stack, .newline_offsets=ax_new_int_vec()});
}

ax_u8 ax_Lexer_peek1(struct ax_Lexer self) {
    if (((self.pos + 1) < self.len)) {
        return ((ax_u8)((ax_bounds_check((ax_u64)((self.pos + 1)), (self.src).len), (self.src).ptr[(self.pos + 1)])));
    }
    return ((ax_u8)(0));
}

void ax_Lexer_emit(struct ax_Lexer* self, ax_u8 kind, ax_i64 offset, ax_i64 length) {
    ax_TokenVec_push(&(self->tokens), ((struct ax_Token){.kind=kind, .padding=((ax_u8)(0)), .len=((ax_u16)(length)), .offset=((ax_u32)(offset))}));
}

static ax_u8 ax_lookup_keyword(ax_string s) {
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"and", .len=3})) {
        return ax_TK_AND;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"as", .len=2})) {
        return ax_TK_AS;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"async", .len=5})) {
        return ax_TK_ASYNC;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"await", .len=5})) {
        return ax_TK_AWAIT;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"break", .len=5})) {
        return ax_TK_BREAK;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"const", .len=5})) {
        return ax_TK_CONST;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"continue", .len=8})) {
        return ax_TK_CONTINUE;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"defer", .len=5})) {
        return ax_TK_DEFER;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"elif", .len=4})) {
        return ax_TK_ELIF;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"else", .len=4})) {
        return ax_TK_ELSE;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"extern", .len=6})) {
        return ax_TK_EXTERN;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"false", .len=5})) {
        return ax_TK_FALSE;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"fn", .len=2})) {
        return ax_TK_FN;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"for", .len=3})) {
        return ax_TK_FOR;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"Future", .len=6})) {
        return ax_TK_FUTURE;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"if", .len=2})) {
        return ax_TK_IF;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"import", .len=6})) {
        return ax_TK_IMPORT;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"in", .len=2})) {
        return ax_TK_IN;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"interface", .len=9})) {
        return ax_TK_INTERFACE;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"Isolated", .len=8})) {
        return ax_TK_ISOLATED;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"lent", .len=4})) {
        return ax_TK_LENT;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"let", .len=3})) {
        return ax_TK_LET;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"match", .len=5})) {
        return ax_TK_MATCH;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"mut", .len=3})) {
        return ax_TK_MUT;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"nil", .len=3})) {
        return ax_TK_NIL;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"not", .len=3})) {
        return ax_TK_NOT;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"or", .len=2})) {
        return ax_TK_OR;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"packed", .len=6})) {
        return ax_TK_PACKED;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"pub", .len=3})) {
        return ax_TK_PUB;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"return", .len=6})) {
        return ax_TK_RETURN;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"spawn", .len=5})) {
        return ax_TK_SPAWN;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"struct", .len=6})) {
        return ax_TK_STRUCT;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"true", .len=4})) {
        return ax_TK_TRUE;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"type", .len=4})) {
        return ax_TK_TYPE;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"unsafe", .len=6})) {
        return ax_TK_UNSAFE;
    }
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"while", .len=5})) {
        return ax_TK_WHILE;
    }
    return ax_TK_IDENT;
}

static ax_bool ax_is_ident_start(ax_u8 b) {
    return ((((b >= 97) && (b <= 122)) || ((b >= 65) && (b <= 90))) || (b == 95));
}

static ax_bool ax_is_ident_continue(ax_u8 b) {
    return (ax_is_ident_start(b) || ((b >= 48) && (b <= 57)));
}

void ax_Lexer_scan_ident(struct ax_Lexer* self) {
    ax_i64 start = self->pos;
    self->pos = (self->pos + 1);
    while ((self->pos < self->len)) {
        ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
        if (ax_is_ident_continue(b)) {
            self->pos = (self->pos + 1);
        } else {
            {
                break;
            }
        }
    }
    ax_string text = ax_str_slice(self->src, start, self->pos);
    ax_u8 kind = ax_lookup_keyword(text);
    ax_Lexer_emit(self, kind, start, (self->pos - start));
}

static ax_bool ax_is_dec_digit(ax_u8 b) {
    return ((b >= 48) && (b <= 57));
}

static ax_bool ax_is_hex_digit(ax_u8 b) {
    return ((ax_is_dec_digit(b) || ((b >= 97) && (b <= 102))) || ((b >= 65) && (b <= 70)));
}

static ax_bool ax_is_oct_digit(ax_u8 b) {
    return ((b >= 48) && (b <= 55));
}

static ax_bool ax_is_bin_digit(ax_u8 b) {
    return ((b == 48) || (b == 49));
}

void ax_Lexer_scan_dec_digits(struct ax_Lexer* self) {
    while ((self->pos < self->len)) {
        ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
        if ((ax_is_dec_digit(b) || (b == 95))) {
            self->pos = (self->pos + 1);
        } else {
            {
                break;
            }
        }
    }
}

void ax_Lexer_scan_hex_digits(struct ax_Lexer* self) {
    while ((self->pos < self->len)) {
        ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
        if ((ax_is_hex_digit(b) || (b == 95))) {
            self->pos = (self->pos + 1);
        } else {
            {
                break;
            }
        }
    }
}

void ax_Lexer_scan_oct_digits(struct ax_Lexer* self) {
    while ((self->pos < self->len)) {
        ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
        if ((ax_is_oct_digit(b) || (b == 95))) {
            self->pos = (self->pos + 1);
        } else {
            {
                break;
            }
        }
    }
}

void ax_Lexer_scan_bin_digits(struct ax_Lexer* self) {
    while ((self->pos < self->len)) {
        ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
        if ((ax_is_bin_digit(b) || (b == 95))) {
            self->pos = (self->pos + 1);
        } else {
            {
                break;
            }
        }
    }
}

void ax_Lexer_scan_number(struct ax_Lexer* self) {
    ax_i64 start = self->pos;
    ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
    if (((b == 48) && ((self->pos + 1) < self->len))) {
        ax_u8 next = ((ax_u8)((ax_bounds_check((ax_u64)((self->pos + 1)), (self->src).len), (self->src).ptr[(self->pos + 1)])));
        if (((next == 120) || (next == 88))) {
            self->pos = (self->pos + 2);
            ax_Lexer_scan_hex_digits(self);
            ax_Lexer_emit(self, ax_TK_INT_LIT, start, (self->pos - start));
            return;
        } else if (((next == 111) || (next == 79))) {
            self->pos = (self->pos + 2);
            ax_Lexer_scan_oct_digits(self);
            ax_Lexer_emit(self, ax_TK_INT_LIT, start, (self->pos - start));
            return;
        } else if (((next == 98) || (next == 66))) {
            self->pos = (self->pos + 2);
            ax_Lexer_scan_bin_digits(self);
            ax_Lexer_emit(self, ax_TK_INT_LIT, start, (self->pos - start));
            return;
        }
    }
    ax_Lexer_scan_dec_digits(self);
    if (((self->pos < self->len) && (((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos]))) == 46))) {
        if (((self->pos + 1) < self->len)) {
            ax_u8 next = ((ax_u8)((ax_bounds_check((ax_u64)((self->pos + 1)), (self->src).len), (self->src).ptr[(self->pos + 1)])));
            if (ax_is_dec_digit(next)) {
                self->pos = (self->pos + 1);
                ax_Lexer_scan_dec_digits(self);
                if ((self->pos < self->len)) {
                    ax_u8 exp = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
                    if (((exp == 101) || (exp == 69))) {
                        self->pos = (self->pos + 1);
                        if ((self->pos < self->len)) {
                            ax_u8 sign = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
                            if (((sign == 43) || (sign == 45))) {
                                self->pos = (self->pos + 1);
                            }
                        }
                        ax_Lexer_scan_dec_digits(self);
                    }
                }
                ax_Lexer_emit(self, ax_TK_FLOAT_LIT, start, (self->pos - start));
                return;
            }
        }
    }
    ax_Lexer_emit(self, ax_TK_INT_LIT, start, (self->pos - start));
}

void ax_Lexer_scan_string(struct ax_Lexer* self) {
    ax_i64 start = self->pos;
    self->pos = (self->pos + 1);
    while ((self->pos < self->len)) {
        ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
        if ((b == 34)) {
            self->pos = (self->pos + 1);
            ax_Lexer_emit(self, ax_TK_STRING_LIT, start, (self->pos - start));
            return;
        }
        if ((b == 92)) {
            self->pos = (self->pos + 1);
            if ((self->pos < self->len)) {
                self->pos = (self->pos + 1);
            }
            continue;
        }
        if ((b == 10)) {
            ax_Lexer_emit(self, ax_TK_STRING_LIT, start, (self->pos - start));
            return;
        }
        self->pos = (self->pos + 1);
    }
    ax_Lexer_emit(self, ax_TK_STRING_LIT, start, (self->pos - start));
}

void ax_Lexer_scan_char(struct ax_Lexer* self) {
    ax_i64 start = self->pos;
    self->pos = (self->pos + 1);
    if ((self->pos < self->len)) {
        ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
        if ((b == 92)) {
            self->pos = (self->pos + 1);
            if ((self->pos < self->len)) {
                self->pos = (self->pos + 1);
            }
        } else if ((b != 39)) {
            ax_i32 size = 1;
            if (((b >= ((ax_u8)(192))) && (b < ((ax_u8)(224))))) {
                size = 2;
            } else if (((b >= ((ax_u8)(224))) && (b < ((ax_u8)(240))))) {
                size = 3;
            } else if (((b >= ((ax_u8)(240))) && (b < ((ax_u8)(248))))) {
                size = 4;
            }
            self->pos = (self->pos + size);
            if ((self->pos > self->len)) {
                self->pos = self->len;
            }
        }
    }
    if (((self->pos < self->len) && (((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos]))) == 39))) {
        self->pos = (self->pos + 1);
    }
    ax_Lexer_emit(self, ax_TK_CHAR_LIT, start, (self->pos - start));
}

void ax_Lexer_scan_line_comment(struct ax_Lexer* self) {
    self->pos = (self->pos + 2);
    while (((self->pos < self->len) && (((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos]))) != 10))) {
        self->pos = (self->pos + 1);
    }
}

void ax_Lexer_scan_operator_or_punct(struct ax_Lexer* self) {
    ax_i64 start = self->pos;
    ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
    if (((self->pos + 1) < self->len)) {
        ax_u8 next = ((ax_u8)((ax_bounds_check((ax_u64)((self->pos + 1)), (self->src).len), (self->src).ptr[(self->pos + 1)])));
        if (((b == 61) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_EQ_EQ, start, 2);
            return;
        }
        if (((b == 33) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_BANG_EQ, start, 2);
            return;
        }
        if (((b == 60) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_LT_EQ, start, 2);
            return;
        }
        if (((b == 62) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_GT_EQ, start, 2);
            return;
        }
        if (((b == 42) && (next == 42))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_STAR_STAR, start, 2);
            return;
        }
        if (((b == 60) && (next == 60))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_LT_LT, start, 2);
            return;
        }
        if (((b == 62) && (next == 62))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_GT_GT, start, 2);
            return;
        }
        if (((b == 45) && (next == 62))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_ARROW, start, 2);
            return;
        }
        if (((b == 58) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_COLON_EQ, start, 2);
            return;
        }
        if (((b == 43) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_PLUS_EQ, start, 2);
            return;
        }
        if (((b == 45) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_MINUS_EQ, start, 2);
            return;
        }
        if (((b == 42) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_STAR_EQ, start, 2);
            return;
        }
        if (((b == 47) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_SLASH_EQ, start, 2);
            return;
        }
        if (((b == 37) && (next == 61))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_PERCENT_EQ, start, 2);
            return;
        }
        if (((b == 46) && (next == 42))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_DOT_STAR, start, 2);
            return;
        }
        if (((b == 46) && (next == 46))) {
            self->pos = (self->pos + 2);
            ax_Lexer_emit(self, ax_TK_DOT_DOT, start, 2);
            return;
        }
    }
    if ((b == 43)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_PLUS, start, 1);
        return;
    }
    if ((b == 45)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_MINUS, start, 1);
        return;
    }
    if ((b == 42)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_STAR, start, 1);
        return;
    }
    if ((b == 47)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_SLASH, start, 1);
        return;
    }
    if ((b == 37)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_PERCENT, start, 1);
        return;
    }
    if ((b == 61)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_EQ, start, 1);
        return;
    }
    if ((b == 60)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_LT, start, 1);
        return;
    }
    if ((b == 62)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_GT, start, 1);
        return;
    }
    if ((b == 38)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_AMP, start, 1);
        return;
    }
    if ((b == 124)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_PIPE, start, 1);
        return;
    }
    if ((b == 94)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_CARET, start, 1);
        return;
    }
    if ((b == 126)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_TILDE, start, 1);
        return;
    }
    if ((b == 46)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_DOT, start, 1);
        return;
    }
    if ((b == 44)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_COMMA, start, 1);
        return;
    }
    if ((b == 58)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_COLON, start, 1);
        return;
    }
    if ((b == 59)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_SEMICOLON, start, 1);
        return;
    }
    if ((b == 33)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_BANG, start, 1);
        return;
    }
    if ((b == 40)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_L_PAREN, start, 1);
        return;
    }
    if ((b == 41)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_R_PAREN, start, 1);
        return;
    }
    if ((b == 91)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_L_BRACKET, start, 1);
        return;
    }
    if ((b == 93)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_R_BRACKET, start, 1);
        return;
    }
    if ((b == 123)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_L_BRACE, start, 1);
        return;
    }
    if ((b == 125)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_R_BRACE, start, 1);
        return;
    }
    if ((b == 35)) {
        self->pos = (self->pos + 1);
        ax_Lexer_emit(self, ax_TK_HASH, start, 1);
        return;
    }
    self->pos = (self->pos + 1);
    ax_Lexer_emit(self, ax_TK_ERROR, start, 1);
}

void ax_Lexer_run(struct ax_Lexer* self) {
    while ((self->pos < self->len)) {
        ax_u8 b = ((ax_u8)((ax_bounds_check((ax_u64)(self->pos), (self->src).len), (self->src).ptr[self->pos])));
        if (((b == 32) || (b == 13))) {
            self->pos = (self->pos + 1);
        } else if ((b == 64)) {
            if ((((self->pos + 1) < self->len) && ax_is_ident_start(((ax_u8)((ax_bounds_check((ax_u64)((self->pos + 1)), (self->src).len), (self->src).ptr[(self->pos + 1)])))))) {
                self->pos = (self->pos + 1);
            } else {
                {
                    ax_Lexer_scan_operator_or_punct(self);
                }
            }
        } else if ((b == 9)) {
            self->pos = (self->pos + 1);
        } else if ((b == 10)) {
            ax_i64 offset = self->pos;
            ax_IntVec_push(&(self->newline_offsets), ((ax_i32)(offset)));
            ax_Lexer_emit(self, ax_TK_NEWLINE, offset, 1);
            self->pos = (self->pos + 1);
        } else if (((b == 47) && (ax_Lexer_peek1(*(self)) == 47))) {
            ax_Lexer_scan_line_comment(self);
        } else if (((b >= 48) && (b <= 57))) {
            ax_Lexer_scan_number(self);
        } else if ((b == 34)) {
            ax_Lexer_scan_string(self);
        } else if ((b == 39)) {
            ax_Lexer_scan_char(self);
        } else if (ax_is_ident_start(b)) {
            ax_Lexer_scan_ident(self);
        } else {
            {
                ax_Lexer_scan_operator_or_punct(self);
            }
        }
    }
    ax_Lexer_emit(self, ax_TK_EOF, self->pos, 0);
}

static ax_i64 ax_find_line_start(ax_string src, ax_i64 pos) {
    ax_i64 i = (pos - 1);
    while ((i >= 0)) {
        if ((((ax_u8)((ax_bounds_check((ax_u64)(i), (src).len), (src).ptr[i]))) == 10)) {
            return (i + 1);
        }
        i = (i - 1);
    }
    return 0;
}

ax_i64 ax_Lexer_find_next_content_offset(struct ax_Lexer self, ax_i64 start_idx) {
    ax_i64 j = start_idx;
    while ((j < self.tokens.len)) {
        struct ax_Token tok = ((self.tokens.data)[j]);
        if ((tok.kind == ax_TK_NEWLINE)) {
            j = (j + 1);
            continue;
        }
        if ((tok.kind == ax_TK_EOF)) {
            return (-1);
        }
        ax_i64 line_start = ax_find_line_start(self.src, ((ax_i64)(tok.offset)));
        return line_start;
    }
    return (-1);
}

static ax_i64 ax_count_leading_spaces(ax_string src, ax_i64 offset, ax_i64 len) {
    ax_i32 count = 0;
    ax_i64 i = offset;
    while ((i < len)) {
        if ((((ax_u8)((ax_bounds_check((ax_u64)(i), (src).len), (src).ptr[i]))) == 32)) {
            count = (count + 1);
        } else {
            {
                break;
            }
        }
        i = (i + 1);
    }
    return count;
}

struct ax_TokenVec ax_Lexer_process_indentation(struct ax_Lexer* self) {
    struct ax_TokenVec out = ax_new_token_vec();
    struct ax_IntVec stack = ax_new_int_vec();
    ax_IntVec_push(&(stack), 0);
    ax_i32 nesting_depth = 0;
    ax_i32 i = 0;
    while ((i < self->tokens.len)) {
        struct ax_Token tok = ((self->tokens.data)[i]);
        if ((((tok.kind == ax_TK_L_PAREN) || (tok.kind == ax_TK_L_BRACKET)) || (tok.kind == ax_TK_L_BRACE))) {
            nesting_depth = (nesting_depth + 1);
        } else if ((((tok.kind == ax_TK_R_PAREN) || (tok.kind == ax_TK_R_BRACKET)) || (tok.kind == ax_TK_R_BRACE))) {
            nesting_depth = (nesting_depth - 1);
            if ((nesting_depth < 0)) {
                nesting_depth = 0;
            }
        }
        if ((tok.kind != ax_TK_NEWLINE)) {
            if ((tok.kind == ax_TK_EOF)) {
                while ((ax_IntVec_top(stack) > 0)) {
                    ax_TokenVec_push(&(out), ((struct ax_Token){.kind=ax_TK_DEDENT, .padding=((ax_u8)(0)), .len=((ax_u16)(0)), .offset=tok.offset}));
                    ax_IntVec_pop(&(stack));
                }
                ax_TokenVec_push(&(out), tok);
                i = (i + 1);
                continue;
            }
            ax_TokenVec_push(&(out), tok);
            i = (i + 1);
            continue;
        }
        if ((nesting_depth > 0)) {
            i = (i + 1);
            continue;
        }
        ax_i64 next_start = ax_Lexer_find_next_content_offset(*(self), (i + 1));
        if ((next_start < 0)) {
            ax_TokenVec_push(&(out), tok);
            i = (i + 1);
            continue;
        }
        ax_i64 next_indent = ax_count_leading_spaces(self->src, next_start, self->len);
        ax_i32 current_indent = ax_IntVec_top(stack);
        if ((next_indent > current_indent)) {
            ax_IntVec_push(&(stack), ((ax_i32)(next_indent)));
            ax_TokenVec_push(&(out), tok);
            ax_TokenVec_push(&(out), ((struct ax_Token){.kind=ax_TK_INDENT, .padding=((ax_u8)(0)), .len=((ax_u16)(0)), .offset=tok.offset}));
        } else if ((next_indent < current_indent)) {
            ax_TokenVec_push(&(out), tok);
            while ((ax_IntVec_top(stack) > ((ax_i32)(next_indent)))) {
                ax_TokenVec_push(&(out), ((struct ax_Token){.kind=ax_TK_DEDENT, .padding=((ax_u8)(0)), .len=((ax_u16)(0)), .offset=tok.offset}));
                ax_IntVec_pop(&(stack));
            }
        } else {
            {
                ax_TokenVec_push(&(out), tok);
            }
        }
        i = (i + 1);
    }
    if ((stack.data != ((ax_i32*)(NULL)))) {
        ax_free(((ax_u8*)(stack.data)));
    }
    return out;
    /* skip destroy for value type out */
}

struct ax_TokenVec ax_Lexer_tokenize(struct ax_Lexer* self) {
    ax_Lexer_run(self);
    struct ax_TokenVec processed = ax_Lexer_process_indentation(self);
    return processed;
    /* skip destroy for value type processed */
}

struct ax_NodeVec ax_new_node_vec(void) {
    return ((struct ax_NodeVec){.data=((struct ax_AstNode*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_NodeVec_push(struct ax_NodeVec* self, struct ax_AstNode node) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_AstNode* new_data = ((struct ax_AstNode*)(ax_alloc((new_cap * 24))));
        if ((self->data != ((struct ax_AstNode*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 24));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = node;
    self->len = (self->len + 1);
    return idx;
}

struct ax_AstTree ax_new_ast_tree(ax_string src, struct ax_TokenVec tokens) {
    struct ax_NodeVec nodes = ax_new_node_vec();
    ax_NodeVec_push(&(nodes), ((struct ax_AstNode){.kind=ax_NODE_PROGRAM, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .token_idx=((ax_u32)(0)), .first_child=ax_NULL_IDX, .next_sibling=ax_NULL_IDX, .payload=((ax_u32)(0)), .extra_idx=((ax_u32)(0))}));
    return ((struct ax_AstTree){.nodes=nodes, .extras=ax_new_int_vec(), .src=src, .tokens=tokens});
}

ax_u32 ax_AstTree_add_node(struct ax_AstTree* self, ax_u8 kind, ax_u32 token_idx) {
    return ax_NodeVec_push(&(self->nodes), ((struct ax_AstNode){.kind=kind, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .token_idx=token_idx, .first_child=ax_NULL_IDX, .next_sibling=ax_NULL_IDX, .payload=((ax_u32)(0)), .extra_idx=((ax_u32)(0))}));
}

ax_u32 ax_AstTree_add_extra(struct ax_AstTree* self, ax_i32 val) {
    ax_u32 idx = ((ax_u32)(self->extras.len));
    ax_IntVec_push(&(self->extras), val);
    return idx;
}

ax_u32 ax_AstTree_clone_subtree(struct ax_AstTree* self, ax_u32 node_idx) {
    if (((node_idx == ((ax_u32)(0))) || (node_idx == ((ax_u32)(0xffffffff))))) {
        return ((ax_u32)(0));
    }
    struct ax_AstNode orig_node = ((self->nodes.data)[node_idx]);
    ax_u32 new_idx = ax_AstTree_add_node(self, orig_node.kind, orig_node.token_idx);
    ((self->nodes.data)[new_idx]).flags = orig_node.flags;
    ((self->nodes.data)[new_idx]).payload = orig_node.payload;
    ((self->nodes.data)[new_idx]).extra_idx = orig_node.extra_idx;
    ax_u32 child_idx = orig_node.first_child;
    ax_u32 prev_cloned_child = ((ax_u32)(0));
    while ((child_idx != ((ax_u32)(0)))) {
        ax_u32 cloned_child = ax_AstTree_clone_subtree(self, child_idx);
        if ((prev_cloned_child == ((ax_u32)(0)))) {
            ((self->nodes.data)[new_idx]).first_child = cloned_child;
        } else {
            {
                ((self->nodes.data)[prev_cloned_child]).next_sibling = cloned_child;
            }
        }
        prev_cloned_child = cloned_child;
        child_idx = ((self->nodes.data)[child_idx]).next_sibling;
    }
    return new_idx;
}

ax_u32 ax_fnv1a(ax_string s) {
    ax_u32 h = ((ax_u32)(2166136261));
    ax_i64 length = ax_str_len(s);
    ax_i32 i = 0;
    while ((i < length)) {
        h = (h ^ ((ax_u32)((ax_bounds_check((ax_u64)(i), (s).len), (s).ptr[i]))));
        h = (h * ((ax_u32)(16777619)));
        i = (i + 1);
    }
    return h;
}

struct ax_InternPool ax_new_intern_pool(void) {
    ax_i64 table_size = ((ax_i64)(512));
    struct ax_InternEntry* table = ((struct ax_InternEntry*)(ax_alloc((table_size * 24))));
    ax_i32 i = 0;
    while ((i < table_size)) {
        ((table)[i]) = ((struct ax_InternEntry){.hash=((ax_u32)(0)), .padding=((ax_u32)(0)), .start=((ax_u32)(0)), .len=((ax_u32)(0)), .id=((ax_u32)(0)), .dummy=((ax_u32)(0))});
        i = (i + 1);
    }
    struct ax_InternEntry* ids = ((struct ax_InternEntry*)(ax_alloc((256 * 24))));
    return ((struct ax_InternPool){.arena=((ax_u8*)(NULL)), .arena_len=0, .arena_cap=0, .table=table, .table_size=table_size, .ids=ids, .count=0});
}

static ax_string ax_alloc_str_from_raw(ax_u8* data, ax_i64 len) {
    ax_u8* buf = ((ax_u8*)(ax_alloc((len + 1))));
    memcpy(buf, data, len);
    ((buf)[len]) = ((ax_u8)(0));
    return ((ax_string){.ptr = (const ax_u8*)(buf), .len = strlen((const char*)(buf))});
}

ax_string ax_InternPool_get_str(struct ax_InternPool self, ax_u32 id) {
    if (((id == ((ax_u32)(0))) || (id > ((ax_u32)(self.count))))) {
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    struct ax_InternEntry entry = ((self.ids)[(((ax_i64)(id)) - 1)]);
    ax_u8* start_ptr = ((ax_u8*)((((ax_i64)(self.arena)) + ((ax_i64)(entry.start)))));
    return ax_alloc_str_from_raw(start_ptr, ((ax_i64)(entry.len)));
}

ax_string ax_InternPool_get(struct ax_InternPool self, ax_u32 id) {
    return ax_InternPool_get_str(self, id);
}

ax_u32 ax_InternPool_intern(struct ax_InternPool* self, ax_string s) {
    ax_i64 length = ax_str_len(s);
    if ((length == 0)) {
        return ((ax_u32)(0));
    }
    ax_u32 h = ax_fnv1a(s);
    ax_i64 mask = (self->table_size - 1);
    ax_i64 idx = (((ax_i64)(h)) & mask);
    while (AX_TRUE) {
        struct ax_InternEntry entry = ((self->table)[idx]);
        if ((entry.id == ((ax_u32)(0)))) {
            return ax_InternPool_insert_at(self, idx, s, h);
        }
        if (((entry.hash == h) && (((ax_i64)(entry.len)) == length))) {
            ax_bool is_match = AX_TRUE;
            ax_i32 i = 0;
            while ((i < length)) {
                if ((((self->arena)[(((ax_i64)(entry.start)) + i)]) != ((ax_u8)((ax_bounds_check((ax_u64)(i), (s).len), (s).ptr[i]))))) {
                    is_match = AX_FALSE;
                    break;
                }
                i = (i + 1);
            }
            if (is_match) {
                return entry.id;
            }
        }
        idx = ((idx + 1) & mask);
    }
    return ((ax_u32)(0));
}

ax_u32 ax_InternPool_intern_string(struct ax_InternPool* self, ax_string s) {
    return ax_InternPool_intern(self, s);
}

static ax_u32 ax_InternPool_insert_at(struct ax_InternPool* self, ax_i64 slot, ax_string s, ax_u32 h) {
    ax_i64 length = ax_str_len(s);
    if (((self->count > 0) && ((self->count % 256) == 0))) {
        struct ax_InternEntry* new_ids = ((struct ax_InternEntry*)(ax_alloc(((self->count + 256) * 24))));
        memcpy(((ax_u8*)(new_ids)), ((ax_u8*)(self->ids)), (self->count * 24));
        self->ids = new_ids;
    }
    if (((self->arena_len + length) >= self->arena_cap)) {
        ax_i64 new_cap = ((ax_i64)(1024));
        if ((self->arena_cap != 0)) {
            new_cap = (self->arena_cap * 2);
        }
        if ((new_cap < (self->arena_len + length))) {
            new_cap = ((self->arena_len + length) + 1024);
        }
        ax_u8* new_arena = ((ax_u8*)(ax_alloc(new_cap)));
        if ((self->arena != ((ax_u8*)(NULL)))) {
            memcpy(new_arena, self->arena, self->arena_len);
        }
        self->arena = new_arena;
        self->arena_cap = new_cap;
    }
    ax_i64 start = self->arena_len;
    ax_i32 i = 0;
    while ((i < length)) {
        ((self->arena)[(start + i)]) = ((ax_u8)((ax_bounds_check((ax_u64)(i), (s).len), (s).ptr[i])));
        i = (i + 1);
    }
    self->arena_len = (self->arena_len + length);
    ax_u32 new_id = ((ax_u32)((self->count + 1)));
    struct ax_InternEntry entry = ((struct ax_InternEntry){.hash=h, .padding=((ax_u32)(0)), .start=((ax_u32)(start)), .len=((ax_u32)(length)), .id=new_id, .dummy=((ax_u32)(0))});
    ((self->table)[slot]) = entry;
    ((self->ids)[self->count]) = entry;
    self->count = (self->count + 1);
    if ((((self->count * 100) / self->table_size) > 70)) {
        ax_InternPool_grow_table(self);
    }
    return new_id;
}

static void ax_InternPool_grow_table(struct ax_InternPool* self) {
    ax_i64 old_size = self->table_size;
    ax_i64 new_size = (old_size * 2);
    struct ax_InternEntry* new_table = ((struct ax_InternEntry*)(ax_alloc((new_size * 24))));
    ax_i32 i = 0;
    while ((i < new_size)) {
        ((new_table)[i]) = ((struct ax_InternEntry){.hash=((ax_u32)(0)), .padding=((ax_u32)(0)), .start=((ax_u32)(0)), .len=((ax_u32)(0)), .id=((ax_u32)(0)), .dummy=((ax_u32)(0))});
        i = (i + 1);
    }
    ax_i64 mask = (new_size - 1);
    i = 0;
    while ((i < old_size)) {
        struct ax_InternEntry entry = ((self->table)[i]);
        if ((entry.id != ((ax_u32)(0)))) {
            ax_i64 idx = (((ax_i64)(entry.hash)) & mask);
            while ((((new_table)[idx]).id != ((ax_u32)(0)))) {
                idx = ((idx + 1) & mask);
            }
            ((new_table)[idx]) = entry;
        }
        i = (i + 1);
    }
    self->table = new_table;
    self->table_size = new_size;
}

void ax_InternPool_free_pool(struct ax_InternPool* self) {
    if ((self->arena != ((ax_u8*)(NULL)))) {
        ax_free(self->arena);
    }
    if ((self->table != ((struct ax_InternEntry*)(NULL)))) {
        ax_free(((ax_u8*)(self->table)));
    }
    if ((self->ids != ((struct ax_InternEntry*)(NULL)))) {
        ax_free(((ax_u8*)(self->ids)));
    }
}

struct ax_Parser ax_TokenVec_new_parser(struct ax_TokenVec tokens, ax_string src, struct ax_InternPool pool) {
    return ((struct ax_Parser){.tokens=tokens, .pos=0, .tree=ax_new_ast_tree(src, tokens), .pool=pool, .src=src, .diags_count=0});
}

struct ax_Token ax_Parser_peek(struct ax_Parser* self) {
    while (((self->pos < self->tokens.len) && (((self->tokens.data)[self->pos]).kind == ax_TK_NEWLINE))) {
        self->pos = (self->pos + 1);
    }
    if ((self->pos >= self->tokens.len)) {
        return ((struct ax_Token){.kind=ax_TK_EOF, .padding=((ax_u8)(0)), .len=((ax_u16)(0)), .offset=((ax_u32)(0))});
    }
    return ((self->tokens.data)[self->pos]);
}

struct ax_Token ax_Parser_peek_raw(struct ax_Parser self) {
    if ((self.pos >= self.tokens.len)) {
        return ((struct ax_Token){.kind=ax_TK_EOF, .padding=((ax_u8)(0)), .len=((ax_u16)(0)), .offset=((ax_u32)(0))});
    }
    return ((self.tokens.data)[self.pos]);
}

struct ax_Token ax_Parser_peek_at(struct ax_Parser self, ax_i64 offset) {
    ax_i64 pos = (self.pos + offset);
    while (((pos < self.tokens.len) && (((self.tokens.data)[pos]).kind == ax_TK_NEWLINE))) {
        pos = (pos + 1);
    }
    if ((pos >= self.tokens.len)) {
        return ((struct ax_Token){.kind=ax_TK_EOF, .padding=((ax_u8)(0)), .len=((ax_u16)(0)), .offset=((ax_u32)(0))});
    }
    return ((self.tokens.data)[pos]);
}

struct ax_Token ax_Parser_consume(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_peek(self);
    if ((self->pos < self->tokens.len)) {
        self->pos = (self->pos + 1);
    }
    return tok;
    /* skip destroy for value type tok */
}

ax_bool ax_Parser_check(struct ax_Parser* self, ax_u8 kind) {
    return (ax_Parser_peek(self).kind == kind);
}

ax_bool ax_Parser_check_raw(struct ax_Parser self, ax_u8 kind) {
    return (ax_Parser_peek_raw(self).kind == kind);
}

void ax_Parser_errorf(struct ax_Parser* self, struct ax_Token tok, ax_string msg) {
    self->diags_count = (self->diags_count + 1);
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"error: %s at offset %d\n", .len=23}).ptr, ((ax_i64)(((ax_u8*)(msg.ptr)))), ((ax_i64)(tok.offset)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
}

struct ax_Token ax_Parser_expect(struct ax_Parser* self, ax_u8 kind) {
    struct ax_Token tok = ax_Parser_peek(self);
    if ((tok.kind != kind)) {
        ax_Parser_errorf(self, tok, (ax_string){.ptr=(const ax_u8*)"unexpected token", .len=16});
        return tok;
    }
    return ax_Parser_consume(self);
    /* skip destroy for value type tok */
}

void ax_Parser_expect_newline(struct ax_Parser* self) {
    if (((self->pos < self->tokens.len) && (((self->tokens.data)[self->pos]).kind == ax_TK_NEWLINE))) {
        self->pos = (self->pos + 1);
        return;
    }
    struct ax_Token raw = ax_Parser_peek_raw(*(self));
    if (((raw.kind != ax_TK_EOF) && (raw.kind != ax_TK_DEDENT))) {
        ax_Parser_errorf(self, raw, (ax_string){.ptr=(const ax_u8*)"expected newline", .len=16});
    }
}

ax_u32 ax_Parser_token_idx(struct ax_Parser self, struct ax_Token tok) {
    ax_i64 lo = ((ax_i64)(0));
    ax_i64 hi = self.tokens.len;
    while ((lo < hi)) {
        ax_i64 mid = ((lo + hi) / 2);
        if ((((self.tokens.data)[mid]).offset < tok.offset)) {
            lo = (mid + 1);
        } else {
            {
                hi = mid;
            }
        }
    }
    if ((lo < self.tokens.len)) {
        return ((ax_u32)(lo));
    }
    return ((ax_u32)(0));
}

ax_string ax_Parser_token_text(struct ax_Parser self, struct ax_Token tok) {
    if ((tok.len == ((ax_u16)(0)))) {
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    ax_u8* start_ptr = ((ax_u8*)((((ax_i64)(((ax_u8*)(self.src.ptr)))) + ((ax_i64)(tok.offset)))));
    return ax_alloc_str_from_raw(start_ptr, ((ax_i64)(tok.len)));
}

void ax_Parser_append_child(struct ax_Parser* self, ax_u32 parent, ax_u32 child) {
    if ((((self->tree.nodes.data)[parent]).first_child == ax_NULL_IDX)) {
        ((self->tree.nodes.data)[parent]).first_child = child;
        return;
    }
    ax_u32 cur = ((self->tree.nodes.data)[parent]).first_child;
    while ((((self->tree.nodes.data)[cur]).next_sibling != ax_NULL_IDX)) {
        cur = ((self->tree.nodes.data)[cur]).next_sibling;
    }
    ((self->tree.nodes.data)[cur]).next_sibling = child;
}

void ax_Parser_set_payload(struct ax_Parser* self, ax_u32 node, ax_u32 payload) {
    ((self->tree.nodes.data)[node]).payload = payload;
}

void ax_Parser_set_flags(struct ax_Parser* self, ax_u32 node, ax_u16 flags) {
    ((self->tree.nodes.data)[node]).flags = (((self->tree.nodes.data)[node]).flags | flags);
}

static ax_i32 ax_left_binding_power(ax_u8 kind) {
    if ((kind == ax_TK_OR)) {
        return ax_BP_OR;
    }
    if ((kind == ax_TK_AND)) {
        return ax_BP_AND;
    }
    if (((((((kind == ax_TK_EQ_EQ) || (kind == ax_TK_BANG_EQ)) || (kind == ax_TK_LT)) || (kind == ax_TK_GT)) || (kind == ax_TK_LT_EQ)) || (kind == ax_TK_GT_EQ))) {
        return ax_BP_CMP;
    }
    if ((kind == ax_TK_PIPE)) {
        return ax_BP_BIT_OR;
    }
    if ((kind == ax_TK_CARET)) {
        return ax_BP_BIT_XOR;
    }
    if ((kind == ax_TK_AMP)) {
        return ax_BP_BIT_AND;
    }
    if (((kind == ax_TK_LT_LT) || (kind == ax_TK_GT_GT))) {
        return ax_BP_SHIFT;
    }
    if (((kind == ax_TK_PLUS) || (kind == ax_TK_MINUS))) {
        return ax_BP_ADD;
    }
    if ((((kind == ax_TK_STAR) || (kind == ax_TK_SLASH)) || (kind == ax_TK_PERCENT))) {
        return ax_BP_MUL;
    }
    if ((kind == ax_TK_STAR_STAR)) {
        return ax_BP_POWER;
    }
    if ((kind == ax_TK_DOT_DOT)) {
        return 35;
    }
    if ((((((kind == ax_TK_DOT) || (kind == ax_TK_DOT_STAR)) || (kind == ax_TK_L_BRACKET)) || (kind == ax_TK_L_PAREN)) || (kind == ax_TK_AS))) {
        return ax_BP_POSTFIX;
    }
    return ax_BP_NONE;
}

ax_u32 ax_Parser_parse_expr_with_prec(struct ax_Parser* self, ax_i32 min_bp) {
    ax_u32 left = ax_Parser_parse_nud(self);
    if ((left == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    while (AX_TRUE) {
        struct ax_Token tok = ax_Parser_peek_raw(*(self));
        ax_i32 bp = ax_left_binding_power(tok.kind);
        if ((bp <= min_bp)) {
            break;
        }
        left = ax_Parser_parse_led(self, left, tok, bp);
        if ((left == ((ax_u32)(0)))) {
            break;
        }
    }
    return left;
}

ax_u32 ax_Parser_parse_nud(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_peek(self);
    if ((tok.kind == ax_TK_IDENT)) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_IDENT, ax_Parser_token_idx(*(self), tok));
        ax_string text = ax_Parser_token_text(*(self), tok);
        ax_u32 name_id = ax_InternPool_intern(&(self->pool), text);
        ax_free(((ax_u8*)(text.ptr)));
        ax_Parser_set_payload(self, node, name_id);
        return node;
    }
    if ((tok.kind == ax_TK_INT_LIT)) {
        ax_Parser_consume(self);
        return ax_AstTree_add_node(&(self->tree), ax_NODE_INT_LIT, ax_Parser_token_idx(*(self), tok));
    }
    if ((tok.kind == ax_TK_FLOAT_LIT)) {
        ax_Parser_consume(self);
        return ax_AstTree_add_node(&(self->tree), ax_NODE_FLOAT_LIT, ax_Parser_token_idx(*(self), tok));
    }
    if ((tok.kind == ax_TK_STRING_LIT)) {
        ax_Parser_consume(self);
        return ax_AstTree_add_node(&(self->tree), ax_NODE_STRING_LIT, ax_Parser_token_idx(*(self), tok));
    }
    if ((tok.kind == ax_TK_CHAR_LIT)) {
        ax_Parser_consume(self);
        return ax_AstTree_add_node(&(self->tree), ax_NODE_CHAR_LIT, ax_Parser_token_idx(*(self), tok));
    }
    if (((tok.kind == ax_TK_TRUE) || (tok.kind == ax_TK_FALSE))) {
        ax_Parser_consume(self);
        return ax_AstTree_add_node(&(self->tree), ax_NODE_BOOL_LIT, ax_Parser_token_idx(*(self), tok));
    }
    if ((tok.kind == ax_TK_NIL)) {
        ax_Parser_consume(self);
        return ax_AstTree_add_node(&(self->tree), ax_NODE_NIL_LIT, ax_Parser_token_idx(*(self), tok));
    }
    if ((((tok.kind == ax_TK_MINUS) || (tok.kind == ax_TK_TILDE)) || (tok.kind == ax_TK_AMP))) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_UNARY_EXPR, ax_Parser_token_idx(*(self), tok));
        ax_u32 operand = ax_Parser_parse_expr_with_prec(self, ax_BP_UNARY);
        if ((operand != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, operand);
        }
        return node;
    }
    if ((tok.kind == ax_TK_NOT)) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_UNARY_EXPR, ax_Parser_token_idx(*(self), tok));
        ax_u32 operand = ax_Parser_parse_expr_with_prec(self, ax_BP_NOT);
        if ((operand != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, operand);
        }
        return node;
    }
    if ((tok.kind == ax_TK_L_PAREN)) {
        ax_Parser_consume(self);
        ax_u32 inner = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
        ax_Parser_expect(self, ax_TK_R_PAREN);
        return inner;
    }
    if ((tok.kind == ax_TK_SPAWN)) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_SPAWN_EXPR, ax_Parser_token_idx(*(self), tok));
        ax_u32 expr = ax_Parser_parse_expr_with_prec(self, ax_BP_UNARY);
        if ((expr != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, expr);
        }
        return node;
    }
    if ((tok.kind == ax_TK_AWAIT)) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_AWAIT_EXPR, ax_Parser_token_idx(*(self), tok));
        ax_u32 expr = ax_Parser_parse_expr_with_prec(self, ax_BP_UNARY);
        if ((expr != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, expr);
        }
        return node;
    }
    if ((tok.kind == ax_TK_HASH)) {
        struct ax_Token hash_tok = ax_Parser_consume(self);
        struct ax_Token run_tok = ax_Parser_expect(self, ax_TK_IDENT);
        ax_string run_text = ax_Parser_token_text(*(self), run_tok);
        ax_bool is_run = AX_FALSE;
        if ((ax_str_len(run_text) > 0)) {
            if (ax_str_eq(run_text, (ax_string){.ptr=(const ax_u8*)"run", .len=3})) {
                is_run = AX_TRUE;
            }
            ax_free(((ax_u8*)(run_text.ptr)));
        }
        if (is_run) {
            ax_u32 comptime_node = ax_AstTree_add_node(&(self->tree), ax_NODE_COMPTIME, ax_Parser_token_idx(*(self), hash_tok));
            if (ax_Parser_check(self, ax_TK_L_BRACE)) {
                ax_Parser_consume(self);
                ax_u32 block_node = ax_AstTree_add_node(&(self->tree), ax_NODE_BLOCK, ax_Parser_token_idx(*(self), run_tok));
                while (((!ax_Parser_check(self, ax_TK_R_BRACE)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
                    ax_i64 prev_pos = self->pos;
                    ax_u32 stmt = ax_Parser_parse_stmt(self);
                    if ((stmt != ((ax_u32)(0)))) {
                        ax_Parser_append_child(self, block_node, stmt);
                    }
                    if ((self->pos == prev_pos)) {
                        ax_Parser_consume(self);
                    }
                }
                ax_Parser_expect(self, ax_TK_R_BRACE);
                ax_Parser_append_child(self, comptime_node, block_node);
            } else if (ax_Parser_check(self, ax_TK_COLON)) {
                ax_Parser_consume(self);
                ax_u32 block_node = ax_Parser_parse_block(self);
                if ((block_node != ((ax_u32)(0)))) {
                    ax_Parser_append_child(self, comptime_node, block_node);
                }
            } else {
                {
                    ax_u32 expr_node = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
                    if ((expr_node != ((ax_u32)(0)))) {
                        ax_Parser_append_child(self, comptime_node, expr_node);
                    }
                }
            }
            return comptime_node;
        } else {
            {
                return ax_AstTree_add_node(&(self->tree), ax_NODE_ERROR, ax_Parser_token_idx(*(self), hash_tok));
            }
        }
    }
    ax_Parser_errorf(self, tok, (ax_string){.ptr=(const ax_u8*)"expected expression nud", .len=23});
    ax_Parser_consume(self);
    return ((ax_u32)(0));
}

ax_u32 ax_Parser_parse_led(struct ax_Parser* self, ax_u32 left, struct ax_Token op_tok, ax_i32 bp) {
    ax_Parser_consume(self);
    if ((((((((((((((((((((op_tok.kind == ax_TK_OR) || (op_tok.kind == ax_TK_AND)) || (op_tok.kind == ax_TK_EQ_EQ)) || (op_tok.kind == ax_TK_BANG_EQ)) || (op_tok.kind == ax_TK_LT)) || (op_tok.kind == ax_TK_GT)) || (op_tok.kind == ax_TK_LT_EQ)) || (op_tok.kind == ax_TK_GT_EQ)) || (op_tok.kind == ax_TK_PIPE)) || (op_tok.kind == ax_TK_CARET)) || (op_tok.kind == ax_TK_AMP)) || (op_tok.kind == ax_TK_LT_LT)) || (op_tok.kind == ax_TK_GT_GT)) || (op_tok.kind == ax_TK_PLUS)) || (op_tok.kind == ax_TK_MINUS)) || (op_tok.kind == ax_TK_STAR)) || (op_tok.kind == ax_TK_SLASH)) || (op_tok.kind == ax_TK_PERCENT)) || (op_tok.kind == ax_TK_DOT_DOT))) {
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_BINARY_EXPR, ax_Parser_token_idx(*(self), op_tok));
        ax_u16 flags = ((ax_u16)(0));
        if (((((((op_tok.kind == ax_TK_EQ_EQ) || (op_tok.kind == ax_TK_BANG_EQ)) || (op_tok.kind == ax_TK_LT)) || (op_tok.kind == ax_TK_GT)) || (op_tok.kind == ax_TK_LT_EQ)) || (op_tok.kind == ax_TK_GT_EQ))) {
            flags = ((ax_u16)(1));
        } else if (((op_tok.kind == ax_TK_AND) || (op_tok.kind == ax_TK_OR))) {
            flags = ((ax_u16)(2));
        }
        if ((flags != ((ax_u16)(0)))) {
            ax_Parser_set_flags(self, node, flags);
        }
        ax_Parser_append_child(self, node, left);
        ax_u32 right = ax_Parser_parse_expr_with_prec(self, bp);
        if ((right != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, right);
        }
        return node;
    }
    if ((op_tok.kind == ax_TK_STAR_STAR)) {
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_BINARY_EXPR, ax_Parser_token_idx(*(self), op_tok));
        ax_Parser_append_child(self, node, left);
        ax_u32 right = ax_Parser_parse_expr_with_prec(self, (bp - 1));
        if ((right != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, right);
        }
        return node;
    }
    if ((op_tok.kind == ax_TK_DOT)) {
        struct ax_Token tok = ax_Parser_peek(self);
        struct ax_Token field_tok = ax_Parser_expect(self, ax_TK_IDENT);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_FIELD_EXPR, ax_Parser_token_idx(*(self), tok));
        ax_Parser_append_child(self, node, left);
        ax_string text = ax_Parser_token_text(*(self), field_tok);
        ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
        ax_free(((ax_u8*)(text.ptr)));
        ax_u32 field_name = ax_AstTree_add_node(&(self->tree), ax_NODE_IDENT, ax_Parser_token_idx(*(self), field_tok));
        ax_Parser_append_child(self, node, field_name);
        return node;
    }
    if ((op_tok.kind == ax_TK_DOT_STAR)) {
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_DEREF_EXPR, ax_Parser_token_idx(*(self), op_tok));
        ax_Parser_append_child(self, node, left);
        return node;
    }
    if ((op_tok.kind == ax_TK_L_BRACKET)) {
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_INDEX_EXPR, ax_Parser_token_idx(*(self), op_tok));
        ax_Parser_append_child(self, node, left);
        ax_u32 idx = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
        if ((idx != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, idx);
        }
        ax_Parser_expect(self, ax_TK_R_BRACKET);
        return node;
    }
    if ((op_tok.kind == ax_TK_L_PAREN)) {
        return ax_Parser_parse_call_args(self, left, op_tok);
    }
    if ((op_tok.kind == ax_TK_AS)) {
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_CAST_EXPR, ax_Parser_token_idx(*(self), op_tok));
        ax_Parser_append_child(self, node, left);
        ax_u32 type_node = ax_Parser_parse_type_expr(self);
        if ((type_node != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, type_node);
        }
        return node;
    }
    return left;
}

ax_u32 ax_Parser_parse_call_args(struct ax_Parser* self, ax_u32 callee, struct ax_Token lparen) {
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_CALL_EXPR, ax_Parser_token_idx(*(self), lparen));
    ax_Parser_append_child(self, node, callee);
    while (((!ax_Parser_check(self, ax_TK_R_PAREN)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
        ax_i64 prev_pos = self->pos;
        if ((ax_Parser_check(self, ax_TK_IDENT) && (ax_Parser_peek_at(*(self), ((ax_i64)(1))).kind == ax_TK_COLON))) {
            ax_u32 arg_node = ax_AstTree_add_node(&(self->tree), ax_NODE_NAMED_ARG, ax_Parser_token_idx(*(self), ax_Parser_peek(self)));
            struct ax_Token name_tok = ax_Parser_consume(self);
            ax_Parser_consume(self);
            ax_string text = ax_Parser_token_text(*(self), name_tok);
            ax_u32 name_id = ax_InternPool_intern(&(self->pool), text);
            ax_free(((ax_u8*)(text.ptr)));
            ax_Parser_set_payload(self, arg_node, name_id);
            ax_u32 expr = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
            if ((expr != ((ax_u32)(0)))) {
                ax_Parser_append_child(self, arg_node, expr);
            }
            ax_Parser_append_child(self, node, arg_node);
        } else {
            {
                ax_u32 expr = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
                if ((expr != ((ax_u32)(0)))) {
                    ax_Parser_append_child(self, node, expr);
                }
            }
        }
        if ((!ax_Parser_check(self, ax_TK_R_PAREN))) {
            ax_Parser_expect(self, ax_TK_COMMA);
        }
        if ((self->pos == prev_pos)) {
            ax_Parser_consume(self);
        }
    }
    ax_Parser_expect(self, ax_TK_R_PAREN);
    return node;
}

ax_u32 ax_Parser_parse_stmt(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_peek(self);
    if (((tok.kind == ax_TK_LET) || (tok.kind == ax_TK_MUT))) {
        return ax_Parser_parse_var_decl(self);
    }
    if ((tok.kind == ax_TK_RETURN)) {
        return ax_Parser_parse_return_stmt(self);
    }
    if ((tok.kind == ax_TK_IF)) {
        return ax_Parser_parse_if_stmt(self);
    }
    if ((tok.kind == ax_TK_WHILE)) {
        return ax_Parser_parse_while_loop(self);
    }
    if ((tok.kind == ax_TK_FOR)) {
        return ax_Parser_parse_for_loop(self);
    }
    if ((tok.kind == ax_TK_MATCH)) {
        return ax_Parser_parse_match_stmt(self);
    }
    if ((tok.kind == ax_TK_BREAK)) {
        return ax_Parser_parse_break_stmt(self);
    }
    if ((tok.kind == ax_TK_CONTINUE)) {
        return ax_Parser_parse_continue_stmt(self);
    }
    return ax_Parser_parse_expr_stmt(self);
}

ax_u32 ax_Parser_parse_break_stmt(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_BREAK);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_BREAK_STMT, ax_Parser_token_idx(*(self), tok));
    ax_Parser_expect_newline(self);
    return node;
}

ax_u32 ax_Parser_parse_continue_stmt(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_CONTINUE);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_CONTINUE_STMT, ax_Parser_token_idx(*(self), tok));
    ax_Parser_expect_newline(self);
    return node;
}

ax_u32 ax_Parser_parse_var_decl(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_consume(self);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_VAR_DECL, ax_Parser_token_idx(*(self), tok));
    if ((tok.kind == ax_TK_MUT)) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_MUT)));
    }
    struct ax_Token name_tok = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string text = ax_Parser_token_text(*(self), name_tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
    ax_free(((ax_u8*)(text.ptr)));
    if (ax_Parser_check(self, ax_TK_COLON)) {
        ax_Parser_consume(self);
        ax_u32 t = ax_Parser_parse_type_expr(self);
        if ((t != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, t);
        }
    }
    if ((ax_Parser_check(self, ax_TK_EQ) || ax_Parser_check(self, ax_TK_COLON_EQ))) {
        ax_Parser_consume(self);
        ax_u32 val = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
        if ((val != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, val);
        }
    }
    ax_Parser_expect_newline(self);
    return node;
}

ax_u32 ax_Parser_parse_return_stmt(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_RETURN);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_RETURN_STMT, ax_Parser_token_idx(*(self), tok));
    struct ax_Token raw = ax_Parser_peek_raw(*(self));
    if ((((raw.kind != ax_TK_NEWLINE) && (raw.kind != ax_TK_DEDENT)) && (raw.kind != ax_TK_EOF))) {
        ax_u32 val = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
        if ((val != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, val);
        }
    }
    ax_Parser_expect_newline(self);
    return node;
}

ax_u32 ax_Parser_parse_if_stmt(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_IF);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_IF_STMT, ax_Parser_token_idx(*(self), tok));
    ax_u32 cond = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
    if ((cond != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, cond);
    }
    ax_Parser_expect(self, ax_TK_COLON);
    ax_u32 body = ax_Parser_parse_block(self);
    if ((body != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, body);
    }
    while (ax_Parser_check(self, ax_TK_ELIF)) {
        struct ax_Token elif_tok = ax_Parser_consume(self);
        ax_u32 elif_node = ax_AstTree_add_node(&(self->tree), ax_NODE_ELIF_CLAUSE, ax_Parser_token_idx(*(self), elif_tok));
        ax_u32 elif_cond = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
        if ((elif_cond != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, elif_node, elif_cond);
        }
        ax_Parser_expect(self, ax_TK_COLON);
        ax_u32 elif_body = ax_Parser_parse_block(self);
        if ((elif_body != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, elif_node, elif_body);
        }
        ax_Parser_append_child(self, node, elif_node);
    }
    if (ax_Parser_check(self, ax_TK_ELSE)) {
        struct ax_Token else_tok = ax_Parser_consume(self);
        ax_u32 else_node = ax_AstTree_add_node(&(self->tree), ax_NODE_ELSE_CLAUSE, ax_Parser_token_idx(*(self), else_tok));
        ax_Parser_expect(self, ax_TK_COLON);
        ax_u32 else_body = ax_Parser_parse_block(self);
        if ((else_body != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, else_node, else_body);
        }
        ax_Parser_append_child(self, node, else_node);
    }
    return node;
}

ax_u32 ax_Parser_parse_while_loop(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_WHILE);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_WHILE_STMT, ax_Parser_token_idx(*(self), tok));
    ax_u32 cond = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
    if ((cond != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, cond);
    }
    ax_Parser_expect(self, ax_TK_COLON);
    ax_u32 body = ax_Parser_parse_block(self);
    if ((body != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, body);
    }
    return node;
}

ax_u32 ax_Parser_parse_for_loop(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_FOR);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_FOR_STMT, ax_Parser_token_idx(*(self), tok));
    struct ax_Token var_tok = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string text = ax_Parser_token_text(*(self), var_tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
    ax_free(((ax_u8*)(text.ptr)));
    ax_Parser_expect(self, ax_TK_IN);
    ax_u32 iter = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
    if ((iter != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, iter);
    }
    ax_Parser_expect(self, ax_TK_COLON);
    ax_u32 body = ax_Parser_parse_block(self);
    if ((body != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, body);
    }
    return node;
}

ax_u32 ax_Parser_parse_match_stmt(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_MATCH);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_MATCH_STMT, ax_Parser_token_idx(*(self), tok));
    ax_u32 expr = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
    if ((expr != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, expr);
    }
    ax_Parser_expect(self, ax_TK_COLON);
    if ((!ax_Parser_check(self, ax_TK_INDENT))) {
        struct ax_Token p = ax_Parser_peek(self);
        ax_Parser_errorf(self, p, (ax_string){.ptr=(const ax_u8*)"expected INDENT after match", .len=27});
        return node;
    }
    ax_Parser_consume(self);
    while (((!ax_Parser_check(self, ax_TK_DEDENT)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
        ax_i64 prev_pos = self->pos;
        ax_u32 arm = ax_Parser_parse_match_arm(self);
        if ((arm != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, arm);
        }
        if ((self->pos == prev_pos)) {
            self->pos = (self->pos + 1);
        }
    }
    if (ax_Parser_check(self, ax_TK_DEDENT)) {
        ax_Parser_consume(self);
    }
    return node;
}

ax_u32 ax_Parser_parse_match_arm(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_peek(self);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_MATCH_ARM, ax_Parser_token_idx(*(self), tok));
    ax_u32 pat = ax_Parser_parse_pattern(self);
    if ((pat != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, pat);
    }
    ax_Parser_expect(self, ax_TK_COLON);
    if (ax_Parser_check(self, ax_TK_INDENT)) {
        ax_u32 body = ax_Parser_parse_block(self);
        if ((body != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, body);
        }
    } else {
        {
            ax_u32 expr = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
            if ((expr != ((ax_u32)(0)))) {
                ax_Parser_append_child(self, node, expr);
            }
            ax_Parser_expect_newline(self);
        }
    }
    return node;
}

ax_u32 ax_Parser_parse_pattern(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_peek(self);
    if ((tok.kind == ax_TK_IDENT)) {
        ax_Parser_consume(self);
        ax_string text = ax_Parser_token_text(*(self), tok);
        ax_i64 len = ax_str_len(text);
        if (((len == 1) && (((ax_u8)((ax_bounds_check((ax_u64)(0), (text).len), (text).ptr[0]))) == ((ax_u8)(95))))) {
            ax_free(((ax_u8*)(text.ptr)));
            return ax_AstTree_add_node(&(self->tree), ax_NODE_WILDCARD_PAT, ax_Parser_token_idx(*(self), tok));
        }
        if (ax_Parser_check(self, ax_TK_L_PAREN)) {
            ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_VARIANT_PAT, ax_Parser_token_idx(*(self), tok));
            ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
            ax_free(((ax_u8*)(text.ptr)));
            ax_Parser_consume(self);
            while (((!ax_Parser_check(self, ax_TK_R_PAREN)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
                ax_i64 prev_pos = self->pos;
                ax_u32 inner = ax_Parser_parse_pattern(self);
                if ((inner != ((ax_u32)(0)))) {
                    ax_Parser_append_child(self, node, inner);
                }
                if ((!ax_Parser_check(self, ax_TK_R_PAREN))) {
                    ax_Parser_expect(self, ax_TK_COMMA);
                }
                if ((self->pos == prev_pos)) {
                    ax_Parser_consume(self);
                }
            }
            ax_Parser_expect(self, ax_TK_R_PAREN);
            return node;
        }
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_BINDING_PAT, ax_Parser_token_idx(*(self), tok));
        ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
        ax_free(((ax_u8*)(text.ptr)));
        return node;
    } else if ((((((((tok.kind == ax_TK_INT_LIT) || (tok.kind == ax_TK_FLOAT_LIT)) || (tok.kind == ax_TK_STRING_LIT)) || (tok.kind == ax_TK_CHAR_LIT)) || (tok.kind == ax_TK_TRUE)) || (tok.kind == ax_TK_FALSE)) || (tok.kind == ax_TK_NIL))) {
        ax_Parser_consume(self);
        return ax_AstTree_add_node(&(self->tree), ax_NODE_LITERAL_PAT, ax_Parser_token_idx(*(self), tok));
    } else if ((tok.kind == ax_TK_L_PAREN)) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_TUPLE_PAT, ax_Parser_token_idx(*(self), tok));
        while (((!ax_Parser_check(self, ax_TK_R_PAREN)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
            ax_i64 prev_pos = self->pos;
            ax_u32 inner = ax_Parser_parse_pattern(self);
            if ((inner != ((ax_u32)(0)))) {
                ax_Parser_append_child(self, node, inner);
            }
            if ((!ax_Parser_check(self, ax_TK_R_PAREN))) {
                ax_Parser_expect(self, ax_TK_COMMA);
            }
            if ((self->pos == prev_pos)) {
                ax_Parser_consume(self);
            }
        }
        ax_Parser_expect(self, ax_TK_R_PAREN);
        return node;
    } else {
        {
            ax_Parser_errorf(self, tok, (ax_string){.ptr=(const ax_u8*)"expected pattern", .len=16});
            ax_Parser_consume(self);
            return ((ax_u32)(0));
        }
    }
}

ax_u32 ax_Parser_parse_expr_stmt(struct ax_Parser* self) {
    ax_u32 expr = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
    if ((expr == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    struct ax_Token tok = ax_Parser_peek_raw(*(self));
    if (((((((tok.kind == ax_TK_EQ) || (tok.kind == ax_TK_PLUS_EQ)) || (tok.kind == ax_TK_MINUS_EQ)) || (tok.kind == ax_TK_STAR_EQ)) || (tok.kind == ax_TK_SLASH_EQ)) || (tok.kind == ax_TK_PERCENT_EQ))) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_ASSIGN_STMT, ax_Parser_token_idx(*(self), tok));
        ax_Parser_append_child(self, node, expr);
        ax_u32 rhs = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
        if ((rhs != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, rhs);
        }
        ax_Parser_expect_newline(self);
        return node;
    }
    ax_Parser_expect_newline(self);
    return expr;
}

ax_u32 ax_Parser_parse_block(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_INDENT);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_BLOCK, ax_Parser_token_idx(*(self), tok));
    while (((!ax_Parser_check(self, ax_TK_DEDENT)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
        ax_i64 prev_pos = self->pos;
        ax_u32 stmt = ax_Parser_parse_stmt(self);
        if ((stmt != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, stmt);
        }
        if ((self->pos == prev_pos)) {
            self->pos = (self->pos + 1);
        }
    }
    ax_Parser_expect(self, ax_TK_DEDENT);
    return node;
}

ax_u32 ax_Parser_parse_type_expr(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_peek(self);
    if ((((tok.kind == ax_TK_IDENT) || (tok.kind == ax_TK_FUTURE)) || (tok.kind == ax_TK_ISOLATED))) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_TYPE_EXPR, ax_Parser_token_idx(*(self), tok));
        ax_string text = ax_Parser_token_text(*(self), tok);
        ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
        ax_free(((ax_u8*)(text.ptr)));
        if (ax_Parser_check_raw(*(self), ax_TK_L_BRACKET)) {
            ax_u32 gen_node = ax_AstTree_add_node(&(self->tree), ax_NODE_GENERIC_TYPE, ax_Parser_token_idx(*(self), tok));
            ax_Parser_append_child(self, gen_node, node);
            ax_Parser_consume(self);
            while (((!ax_Parser_check(self, ax_TK_R_BRACKET)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
                ax_i64 prev_pos = self->pos;
                ax_u32 t = ax_Parser_parse_type_expr(self);
                if ((t != ((ax_u32)(0)))) {
                    ax_Parser_append_child(self, gen_node, t);
                }
                if ((!ax_Parser_check(self, ax_TK_R_BRACKET))) {
                    ax_Parser_expect(self, ax_TK_COMMA);
                }
                if ((self->pos == prev_pos)) {
                    ax_Parser_consume(self);
                }
            }
            ax_Parser_expect(self, ax_TK_R_BRACKET);
            return gen_node;
        }
        return node;
    }
    if ((tok.kind == ax_TK_STAR)) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_PTR_TYPE, ax_Parser_token_idx(*(self), tok));
        if (ax_Parser_check(self, ax_TK_MUT)) {
            ax_Parser_consume(self);
            ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_MUT)));
        }
        ax_u32 inner = ax_Parser_parse_type_expr(self);
        if ((inner != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, inner);
        }
        return node;
    }
    if ((tok.kind == ax_TK_L_BRACKET)) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_SLICE_TYPE, ax_Parser_token_idx(*(self), tok));
        ax_u32 inner = ax_Parser_parse_type_expr(self);
        if ((inner != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, inner);
        }
        if (ax_Parser_check(self, ax_TK_SEMICOLON)) {
            ax_Parser_consume(self);
            ((self->tree.nodes.data)[node]).kind = ax_NODE_ARRAY_TYPE;
            struct ax_Token size_tok = ax_Parser_expect(self, ax_TK_INT_LIT);
            ax_u32 size_node = ax_AstTree_add_node(&(self->tree), ax_NODE_INT_LIT, ax_Parser_token_idx(*(self), size_tok));
            ax_Parser_append_child(self, node, size_node);
        }
        ax_Parser_expect(self, ax_TK_R_BRACKET);
        return node;
    }
    if ((tok.kind == ax_TK_FN)) {
        ax_Parser_consume(self);
        ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_FUNC_TYPE, ax_Parser_token_idx(*(self), tok));
        ax_Parser_expect(self, ax_TK_L_PAREN);
        while (((!ax_Parser_check(self, ax_TK_R_PAREN)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
            ax_i64 prev_pos = self->pos;
            ax_u32 t = ax_Parser_parse_type_expr(self);
            if ((t != ((ax_u32)(0)))) {
                ax_Parser_append_child(self, node, t);
            }
            if ((!ax_Parser_check(self, ax_TK_R_PAREN))) {
                ax_Parser_expect(self, ax_TK_COMMA);
            }
            if ((self->pos == prev_pos)) {
                ax_Parser_consume(self);
            }
        }
        ax_Parser_expect(self, ax_TK_R_PAREN);
        if (ax_Parser_check_raw(*(self), ax_TK_ARROW)) {
            ax_Parser_consume(self);
            ax_u32 ret = ax_Parser_parse_type_expr(self);
            if ((ret != ((ax_u32)(0)))) {
                ax_Parser_append_child(self, node, ret);
                ax_Parser_set_flags(self, node, ((ax_u16)(1)));
            }
        }
        return node;
    }
    ax_Parser_errorf(self, tok, (ax_string){.ptr=(const ax_u8*)"expected type expression", .len=24});
    return ((ax_u32)(0));
}

ax_u32 ax_Parser_parse_generic_params(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_peek(self);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_GENERIC_PARAMS, ax_Parser_token_idx(*(self), tok));
    ax_Parser_expect(self, ax_TK_L_BRACKET);
    while (((!ax_Parser_check(self, ax_TK_R_BRACKET)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
        ax_i64 prev_pos = self->pos;
        ax_u32 gp = ax_Parser_parse_generic_param(self);
        if ((gp != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, gp);
        }
        if ((!ax_Parser_check(self, ax_TK_R_BRACKET))) {
            ax_Parser_expect(self, ax_TK_COMMA);
        }
        if ((self->pos == prev_pos)) {
            ax_Parser_consume(self);
        }
    }
    ax_Parser_expect(self, ax_TK_R_BRACKET);
    return node;
}

ax_u32 ax_Parser_parse_generic_param(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_peek(self);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_GENERIC_PARAM, ax_Parser_token_idx(*(self), tok));
    struct ax_Token name_tok = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string text = ax_Parser_token_text(*(self), name_tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
    ax_free(((ax_u8*)(text.ptr)));
    if (ax_Parser_check(self, ax_TK_COLON)) {
        ax_Parser_consume(self);
        ax_u32 constraint = ax_Parser_parse_type_expr(self);
        if ((constraint != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, constraint);
        }
    }
    return node;
}

ax_u32 ax_Parser_parse_func_decl(struct ax_Parser* self, ax_bool is_pub) {
    struct ax_Token fn_tok = ax_Parser_expect(self, ax_TK_FN);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_FUNC_DECL, ax_Parser_token_idx(*(self), fn_tok));
    if (is_pub) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_PUB)));
    }
    struct ax_Token name_tok = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string text = ax_Parser_token_text(*(self), name_tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
    ax_free(((ax_u8*)(text.ptr)));
    if (ax_Parser_check(self, ax_TK_L_BRACKET)) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_GENERIC)));
        ax_u32 gp = ax_Parser_parse_generic_params(self);
        if ((gp != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, gp);
        }
    }
    ax_Parser_expect(self, ax_TK_L_PAREN);
    while (((!ax_Parser_check(self, ax_TK_R_PAREN)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
        ax_i64 prev_pos = self->pos;
        struct ax_Token param_tok = ax_Parser_peek(self);
        ax_u32 param = ax_AstTree_add_node(&(self->tree), ax_NODE_PARAM_DECL, ax_Parser_token_idx(*(self), param_tok));
        if (ax_Parser_check(self, ax_TK_LENT)) {
            ax_Parser_consume(self);
            ax_Parser_set_flags(self, param, ax_FLAG_IS_LENT);
        } else if (ax_Parser_check(self, ax_TK_BANG)) {
            ax_Parser_consume(self);
            ax_Parser_set_flags(self, param, ax_FLAG_IS_SINK);
        } else if (ax_Parser_check(self, ax_TK_MUT)) {
            ax_Parser_consume(self);
            ax_Parser_set_flags(self, param, ((ax_u16)(ax_FLAG_IS_MUT)));
        }
        struct ax_Token param_name = ax_Parser_expect(self, ax_TK_IDENT);
        ax_string param_text = ax_Parser_token_text(*(self), param_name);
        ax_Parser_set_payload(self, param, ax_InternPool_intern(&(self->pool), param_text));
        ax_free(((ax_u8*)(param_text.ptr)));
        ax_Parser_expect(self, ax_TK_COLON);
        ax_u32 t = ax_Parser_parse_type_expr(self);
        if ((t != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, param, t);
        }
        ax_Parser_append_child(self, node, param);
        if ((!ax_Parser_check(self, ax_TK_R_PAREN))) {
            ax_Parser_expect(self, ax_TK_COMMA);
        }
        if ((self->pos == prev_pos)) {
            ax_Parser_consume(self);
        }
    }
    ax_Parser_expect(self, ax_TK_R_PAREN);
    if (ax_Parser_check(self, ax_TK_ARROW)) {
        ax_Parser_consume(self);
        ax_u32 ret = ax_Parser_parse_type_expr(self);
        if ((ret != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, ret);
        }
    }
    ax_Parser_expect(self, ax_TK_COLON);
    ax_u32 body = ax_Parser_parse_block(self);
    if ((body != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, body);
    }
    return node;
}

ax_u32 ax_Parser_parse_struct_decl(struct ax_Parser* self, ax_bool is_pub) {
    ax_bool is_packed = AX_FALSE;
    if (ax_Parser_check(self, ax_TK_PACKED)) {
        ax_Parser_consume(self);
        is_packed = AX_TRUE;
    }
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_STRUCT);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_STRUCT_DECL, ax_Parser_token_idx(*(self), tok));
    if (is_pub) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_PUB)));
    }
    if (is_packed) {
        ax_Parser_set_flags(self, node, ((ax_u16)(64)));
    }
    struct ax_Token name_tok = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string name_text = ax_Parser_token_text(*(self), name_tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), name_text));
    ax_free(((ax_u8*)(name_text.ptr)));
    if (ax_Parser_check(self, ax_TK_L_BRACKET)) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_GENERIC)));
        ax_u32 gp = ax_Parser_parse_generic_params(self);
        if ((gp != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, gp);
        }
    }
    ax_Parser_expect(self, ax_TK_COLON);
    if ((!ax_Parser_check(self, ax_TK_INDENT))) {
        struct ax_Token p = ax_Parser_peek(self);
        ax_Parser_errorf(self, p, (ax_string){.ptr=(const ax_u8*)"expected INDENT after struct declaration", .len=40});
        return node;
    }
    ax_Parser_consume(self);
    while (((!ax_Parser_check(self, ax_TK_DEDENT)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
        ax_i64 prev_pos = self->pos;
        struct ax_Token inner = ax_Parser_peek(self);
        if ((inner.kind == ax_TK_FN)) {
            ax_u32 m = ax_Parser_parse_func_decl(self, AX_FALSE);
            if ((m != ((ax_u32)(0)))) {
                ax_Parser_append_child(self, node, m);
            }
        } else if ((inner.kind == ax_TK_PUB)) {
            ax_Parser_consume(self);
            struct ax_Token next = ax_Parser_peek(self);
            if ((next.kind == ax_TK_FN)) {
                ax_u32 m = ax_Parser_parse_func_decl(self, AX_TRUE);
                if ((m != ((ax_u32)(0)))) {
                    ax_Parser_append_child(self, node, m);
                }
            } else {
                {
                    ax_u32 f = ax_Parser_parse_field_decl(self, AX_TRUE);
                    if ((f != ((ax_u32)(0)))) {
                        ax_Parser_append_child(self, node, f);
                    }
                }
            }
        } else {
            {
                ax_u32 f = ax_Parser_parse_field_decl(self, AX_FALSE);
                if ((f != ((ax_u32)(0)))) {
                    ax_Parser_append_child(self, node, f);
                }
            }
        }
        if ((self->pos == prev_pos)) {
            ax_Parser_consume(self);
        }
    }
    if (ax_Parser_check(self, ax_TK_DEDENT)) {
        ax_Parser_consume(self);
    }
    return node;
}

ax_u32 ax_Parser_parse_field_decl(struct ax_Parser* self, ax_bool is_pub) {
    ax_bool is_mut = AX_FALSE;
    if (ax_Parser_check(self, ax_TK_MUT)) {
        ax_Parser_consume(self);
        is_mut = AX_TRUE;
    }
    struct ax_Token tok = ax_Parser_peek(self);
    if ((tok.kind != ax_TK_IDENT)) {
        ax_Parser_errorf(self, tok, (ax_string){.ptr=(const ax_u8*)"expected field name", .len=19});
        ax_Parser_consume(self);
        return ((ax_u32)(0));
    }
    ax_Parser_consume(self);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_FIELD_DECL, ax_Parser_token_idx(*(self), tok));
    if (is_pub) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_PUB)));
    }
    if (is_mut) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_MUT)));
    }
    ax_string text = ax_Parser_token_text(*(self), tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
    ax_free(((ax_u8*)(text.ptr)));
    ax_Parser_expect(self, ax_TK_COLON);
    ax_u32 type_node = ax_Parser_parse_type_expr(self);
    if ((type_node != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, type_node);
    }
    ax_Parser_expect_newline(self);
    return node;
}

ax_u32 ax_Parser_parse_type_alias_decl(struct ax_Parser* self, ax_bool is_pub) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_TYPE);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_TYPE_ALIAS_DECL, ax_Parser_token_idx(*(self), tok));
    if (is_pub) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_PUB)));
    }
    struct ax_Token name_tok = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string name_text = ax_Parser_token_text(*(self), name_tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), name_text));
    ax_free(((ax_u8*)(name_text.ptr)));
    if (ax_Parser_check(self, ax_TK_L_BRACKET)) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_GENERIC)));
        ax_u32 gp = ax_Parser_parse_generic_params(self);
        if ((gp != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, gp);
        }
    }
    ax_Parser_expect(self, ax_TK_EQ);
    struct ax_Token sum_tok = ax_Parser_peek(self);
    ax_u32 sum_node = ax_AstTree_add_node(&(self->tree), ax_NODE_SUM_TYPE, ax_Parser_token_idx(*(self), sum_tok));
    ax_u32 v = ax_Parser_parse_type_variant(self);
    if ((v != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, sum_node, v);
    }
    while (ax_Parser_check_raw(*(self), ax_TK_PIPE)) {
        ax_Parser_consume(self);
        ax_u32 next_v = ax_Parser_parse_type_variant(self);
        if ((next_v != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, sum_node, next_v);
        }
    }
    ax_Parser_append_child(self, node, sum_node);
    ax_Parser_expect_newline(self);
    return node;
}

ax_u32 ax_Parser_parse_type_variant(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_peek(self);
    if ((tok.kind != ax_TK_IDENT)) {
        return ((ax_u32)(0));
    }
    ax_Parser_consume(self);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_VARIANT_DECL, ax_Parser_token_idx(*(self), tok));
    ax_string text = ax_Parser_token_text(*(self), tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), text));
    ax_free(((ax_u8*)(text.ptr)));
    if (ax_Parser_check_raw(*(self), ax_TK_L_PAREN)) {
        ax_Parser_consume(self);
        while (((!ax_Parser_check(self, ax_TK_R_PAREN)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
            ax_i64 prev_pos = self->pos;
            ax_u32 t = ax_Parser_parse_type_expr(self);
            if ((t != ((ax_u32)(0)))) {
                ax_Parser_append_child(self, node, t);
            }
            if ((!ax_Parser_check(self, ax_TK_R_PAREN))) {
                ax_Parser_expect(self, ax_TK_COMMA);
            }
            if ((self->pos == prev_pos)) {
                ax_Parser_consume(self);
            }
        }
        ax_Parser_expect(self, ax_TK_R_PAREN);
    }
    return node;
}

void ax_Parser_parse_program(struct ax_Parser* self) {
    while ((!ax_Parser_check(self, ax_TK_EOF))) {
        ax_i64 prev_pos = self->pos;
        struct ax_Token tok = ax_Parser_peek(self);
        ax_u32 decl = ((ax_u32)(0));
        if ((tok.kind == ax_TK_IMPORT)) {
            decl = ax_Parser_parse_import_decl(self);
        } else if ((tok.kind == ax_TK_PUB)) {
            ax_Parser_consume(self);
            struct ax_Token next = ax_Parser_peek(self);
            if ((next.kind == ax_TK_FN)) {
                decl = ax_Parser_parse_func_decl(self, AX_TRUE);
            } else if (((next.kind == ax_TK_STRUCT) || (next.kind == ax_TK_PACKED))) {
                decl = ax_Parser_parse_struct_decl(self, AX_TRUE);
            } else if ((next.kind == ax_TK_TYPE)) {
                decl = ax_Parser_parse_type_alias_decl(self, AX_TRUE);
            } else if ((next.kind == ax_TK_CONST)) {
                decl = ax_Parser_parse_const_decl(self, AX_TRUE);
            } else if ((next.kind == ax_TK_EXTERN)) {
                decl = ax_Parser_parse_extern_decl(self, AX_TRUE);
            } else if (((next.kind == ax_TK_LET) || (next.kind == ax_TK_MUT))) {
                decl = ax_Parser_parse_var_decl(self);
                ax_Parser_set_flags(self, decl, ((ax_u16)(ax_FLAG_IS_PUB)));
            }
        } else if ((tok.kind == ax_TK_FN)) {
            decl = ax_Parser_parse_func_decl(self, AX_FALSE);
        } else if (((tok.kind == ax_TK_STRUCT) || (tok.kind == ax_TK_PACKED))) {
            decl = ax_Parser_parse_struct_decl(self, AX_FALSE);
        } else if ((tok.kind == ax_TK_TYPE)) {
            decl = ax_Parser_parse_type_alias_decl(self, AX_FALSE);
        } else if ((tok.kind == ax_TK_CONST)) {
            decl = ax_Parser_parse_const_decl(self, AX_FALSE);
        } else if ((tok.kind == ax_TK_EXTERN)) {
            decl = ax_Parser_parse_extern_decl(self, AX_FALSE);
        } else if (((tok.kind == ax_TK_LET) || (tok.kind == ax_TK_MUT))) {
            decl = ax_Parser_parse_var_decl(self);
        } else {
            {
                ax_Parser_errorf(self, tok, (ax_string){.ptr=(const ax_u8*)"expected top level declaration", .len=30});
                ax_Parser_consume(self);
            }
        }
        if ((decl != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, ((ax_u32)(0)), decl);
        }
        if ((self->pos == prev_pos)) {
            ax_Parser_consume(self);
        }
    }
}

ax_u32 ax_Parser_parse_param(struct ax_Parser* self) {
    struct ax_Token param_tok = ax_Parser_peek(self);
    ax_u32 param = ax_AstTree_add_node(&(self->tree), ax_NODE_PARAM_DECL, ax_Parser_token_idx(*(self), param_tok));
    if (ax_Parser_check(self, ax_TK_LENT)) {
        ax_Parser_consume(self);
        ax_Parser_set_flags(self, param, ax_FLAG_IS_LENT);
    } else if (ax_Parser_check(self, ax_TK_BANG)) {
        ax_Parser_consume(self);
        ax_Parser_set_flags(self, param, ax_FLAG_IS_SINK);
    } else if (ax_Parser_check(self, ax_TK_MUT)) {
        ax_Parser_consume(self);
        ax_Parser_set_flags(self, param, ((ax_u16)(ax_FLAG_IS_MUT)));
    }
    struct ax_Token param_name = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string param_text = ax_Parser_token_text(*(self), param_name);
    ax_Parser_set_payload(self, param, ax_InternPool_intern(&(self->pool), param_text));
    ax_free(((ax_u8*)(param_text.ptr)));
    ax_Parser_expect(self, ax_TK_COLON);
    ax_u32 t = ax_Parser_parse_type_expr(self);
    if ((t != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, param, t);
    }
    return param;
}

ax_u32 ax_Parser_parse_const_decl(struct ax_Parser* self, ax_bool is_pub) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_CONST);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_CONST_DECL, ax_Parser_token_idx(*(self), tok));
    if (is_pub) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_PUB)));
    }
    struct ax_Token name_tok = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string name_text = ax_Parser_token_text(*(self), name_tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), name_text));
    ax_free(((ax_u8*)(name_text.ptr)));
    ax_Parser_expect(self, ax_TK_COLON);
    ax_u32 type_node = ax_Parser_parse_type_expr(self);
    if ((type_node != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, type_node);
    }
    ax_Parser_expect(self, ax_TK_EQ);
    ax_u32 expr = ax_Parser_parse_expr_with_prec(self, ax_BP_NONE);
    if ((expr != ((ax_u32)(0)))) {
        ax_Parser_append_child(self, node, expr);
    }
    ax_Parser_expect_newline(self);
    return node;
}

ax_u32 ax_Parser_parse_extern_decl(struct ax_Parser* self, ax_bool is_pub) {
    ax_Parser_expect(self, ax_TK_EXTERN);
    ax_Parser_expect(self, ax_TK_STRING_LIT);
    struct ax_Token fn_tok = ax_Parser_expect(self, ax_TK_FN);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_FUNC_DECL, ax_Parser_token_idx(*(self), fn_tok));
    ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_EXTERN)));
    if (is_pub) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_PUB)));
    }
    struct ax_Token name_tok = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string name_text = ax_Parser_token_text(*(self), name_tok);
    ax_Parser_set_payload(self, node, ax_InternPool_intern(&(self->pool), name_text));
    ax_free(((ax_u8*)(name_text.ptr)));
    if (ax_Parser_check(self, ax_TK_L_BRACKET)) {
        ax_Parser_set_flags(self, node, ((ax_u16)(ax_FLAG_IS_GENERIC)));
        ax_u32 gp = ax_Parser_parse_generic_params(self);
        if ((gp != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, gp);
        }
    }
    ax_Parser_expect(self, ax_TK_L_PAREN);
    while (((!ax_Parser_check(self, ax_TK_R_PAREN)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
        ax_i64 prev_pos = self->pos;
        ax_u32 param = ax_Parser_parse_param(self);
        if ((param != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, param);
        }
        if ((!ax_Parser_check(self, ax_TK_R_PAREN))) {
            ax_Parser_expect(self, ax_TK_COMMA);
        }
        if ((self->pos == prev_pos)) {
            ax_Parser_consume(self);
        }
    }
    ax_Parser_expect(self, ax_TK_R_PAREN);
    if (ax_Parser_check_raw(*(self), ax_TK_ARROW)) {
        ax_Parser_consume(self);
        ax_u32 ret_type = ax_Parser_parse_type_expr(self);
        if ((ret_type != ((ax_u32)(0)))) {
            ax_Parser_append_child(self, node, ret_type);
        }
    }
    ax_Parser_expect_newline(self);
    return node;
}

ax_u32 ax_Parser_parse_import_decl(struct ax_Parser* self) {
    struct ax_Token tok = ax_Parser_expect(self, ax_TK_IMPORT);
    ax_u32 node = ax_AstTree_add_node(&(self->tree), ax_NODE_IMPORT_DECL, ax_Parser_token_idx(*(self), tok));
    struct ax_Token path_tok = ax_Parser_expect(self, ax_TK_IDENT);
    ax_string path_text = ax_Parser_token_text(*(self), path_tok);
    ax_u32 path_id = ax_InternPool_intern(&(self->pool), path_text);
    ax_free(((ax_u8*)(path_text.ptr)));
    while (ax_Parser_check_raw(*(self), ax_TK_DOT)) {
        ax_Parser_consume(self);
        struct ax_Token seg_tok = ax_Parser_expect(self, ax_TK_IDENT);
        ax_string seg_text = ax_Parser_token_text(*(self), seg_tok);
        ax_string base_str = ax_InternPool_get(self->pool, path_id);
        ax_i64 len_base = ax_str_len(base_str);
        ax_i64 len_seg = ax_str_len(seg_text);
        ax_i64 total_len = ((len_base + 1) + len_seg);
        ax_u8* new_buf = ((ax_u8*)(ax_alloc((total_len + 1))));
        memcpy(new_buf, ((ax_u8*)(base_str.ptr)), len_base);
        ((new_buf)[len_base]) = ((ax_u8)('.'));
        memcpy(((ax_u8*)(((((ax_i64)(new_buf)) + len_base) + ((ax_i64)(1))))), ((ax_u8*)(seg_text.ptr)), len_seg);
        ((new_buf)[total_len]) = ((ax_u8)(0));
        path_id = ax_InternPool_intern(&(self->pool), ((ax_string){.ptr = (const ax_u8*)(new_buf), .len = strlen((const char*)(new_buf))}));
        ax_free(new_buf);
        ax_free(((ax_u8*)(seg_text.ptr)));
    }
    ax_Parser_set_payload(self, node, path_id);
    if (ax_Parser_check_raw(*(self), ax_TK_L_BRACE)) {
        ax_Parser_consume(self);
        while (((!ax_Parser_check(self, ax_TK_R_BRACE)) && (!ax_Parser_check(self, ax_TK_EOF)))) {
            ax_i64 prev_pos = self->pos;
            struct ax_Token name_tok = ax_Parser_expect(self, ax_TK_IDENT);
            ax_string name_text = ax_Parser_token_text(*(self), name_tok);
            ax_u32 name_node = ax_AstTree_add_node(&(self->tree), ax_NODE_IDENT, ax_Parser_token_idx(*(self), name_tok));
            ax_Parser_set_payload(self, name_node, ax_InternPool_intern(&(self->pool), name_text));
            ax_free(((ax_u8*)(name_text.ptr)));
            ax_Parser_append_child(self, node, name_node);
            if ((!ax_Parser_check(self, ax_TK_R_BRACE))) {
                ax_Parser_expect(self, ax_TK_COMMA);
            }
            if ((self->pos == prev_pos)) {
                ax_Parser_consume(self);
            }
        }
        ax_Parser_expect(self, ax_TK_R_BRACE);
    }
    ax_Parser_expect_newline(self);
    return node;
}

void ax_Scope_init_scope(struct ax_Scope* self, ax_u8 kind, ax_u32 parent_id, ax_u32 depth, ax_u32 capacity) {
    self->kind = kind;
    self->parent_id = parent_id;
    self->depth = depth;
    self->capacity = capacity;
    self->count = ((ax_u32)(0));
    self->entries = ((struct ax_ScopeEntry*)(ax_alloc((((ax_i64)(capacity)) * 8))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < ((ax_i64)(capacity)))) {
        ((self->entries)[i]) = ((struct ax_ScopeEntry){.name_id=((ax_u32)(0)), .symbol_idx=((ax_u32)(0))});
        i = (i + 1);
    }
}

static ax_u32 ax_hash_fnv1a(ax_u32 v) {
    ax_u32 hash = ((ax_u32)(2166136261));
    hash = (hash ^ (v & ((ax_u32)(255))));
    hash = (hash * ((ax_u32)(16777619)));
    hash = (hash ^ ((v >> ((ax_u32)(8))) & ((ax_u32)(255))));
    hash = (hash * ((ax_u32)(16777619)));
    hash = (hash ^ ((v >> ((ax_u32)(16))) & ((ax_u32)(255))));
    hash = (hash * ((ax_u32)(16777619)));
    hash = (hash ^ ((v >> ((ax_u32)(24))) & ((ax_u32)(255))));
    hash = (hash * ((ax_u32)(16777619)));
    return hash;
}

void ax_Scope_scope_put(struct ax_Scope* self, ax_u32 name_id, ax_u32 symbol_idx) {
    if (((self->count * ((ax_u32)(4))) > (self->capacity * ((ax_u32)(3))))) {
        ax_Scope_scope_grow(self);
    }
    ax_Scope_scope_insert(*(self), name_id, symbol_idx);
    self->count = (self->count + ((ax_u32)(1)));
}

static void ax_Scope_scope_insert(struct ax_Scope self, ax_u32 name_id, ax_u32 symbol_idx) {
    ax_u32 mask = (self.capacity - ((ax_u32)(1)));
    ax_u32 idx = (ax_hash_fnv1a(name_id) & mask);
    while (AX_TRUE) {
        if ((((self.entries)[idx]).name_id == ((ax_u32)(0)))) {
            ((self.entries)[idx]).name_id = name_id;
            ((self.entries)[idx]).symbol_idx = symbol_idx;
            return;
        }
        idx = ((idx + ((ax_u32)(1))) & mask);
    }
}

struct ax_ScopeEntry ax_Scope_scope_get(struct ax_Scope self, ax_u32 name_id) {
    if ((self.capacity == ((ax_u32)(0)))) {
        return ((struct ax_ScopeEntry){.name_id=((ax_u32)(0)), .symbol_idx=((ax_u32)(0))});
    }
    ax_u32 mask = (self.capacity - ((ax_u32)(1)));
    ax_u32 idx = (ax_hash_fnv1a(name_id) & mask);
    ax_u32 start_idx = idx;
    while (AX_TRUE) {
        struct ax_ScopeEntry entry = ((self.entries)[idx]);
        if ((entry.name_id == ((ax_u32)(0)))) {
            return ((struct ax_ScopeEntry){.name_id=((ax_u32)(0)), .symbol_idx=((ax_u32)(0))});
        }
        if ((entry.name_id == name_id)) {
            return entry;
        }
        idx = ((idx + ((ax_u32)(1))) & mask);
        if ((idx == start_idx)) {
            break;
        }
        /* skip destroy for value type entry */
    }
    return ((struct ax_ScopeEntry){.name_id=((ax_u32)(0)), .symbol_idx=((ax_u32)(0))});
}

static void ax_Scope_scope_grow(struct ax_Scope* self) {
    struct ax_ScopeEntry* old_entries = self->entries;
    ax_u32 old_capacity = self->capacity;
    ax_u32 new_cap = (old_capacity * ((ax_u32)(2)));
    self->capacity = new_cap;
    self->entries = ((struct ax_ScopeEntry*)(ax_alloc((((ax_i64)(new_cap)) * 8))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < ((ax_i64)(new_cap)))) {
        ((self->entries)[i]) = ((struct ax_ScopeEntry){.name_id=((ax_u32)(0)), .symbol_idx=((ax_u32)(0))});
        i = (i + 1);
    }
    i = ((ax_i64)(0));
    while ((i < ((ax_i64)(old_capacity)))) {
        struct ax_ScopeEntry entry = ((old_entries)[i]);
        if ((entry.name_id != ((ax_u32)(0)))) {
            ax_Scope_scope_insert(*(self), entry.name_id, entry.symbol_idx);
        }
        i = (i + 1);
    }
    ax_free(((ax_u8*)(old_entries)));
}

struct ax_SymbolVec ax_new_symbol_vec(void) {
    return ((struct ax_SymbolVec){.data=((struct ax_Symbol*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_SymbolVec_push(struct ax_SymbolVec* self, struct ax_Symbol sym) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_Symbol* new_data = ((struct ax_Symbol*)(ax_alloc((new_cap * 24))));
        if ((self->data != ((struct ax_Symbol*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 24));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = sym;
    self->len = (self->len + 1);
    return idx;
}

struct ax_ScopeVec ax_new_scope_vec(void) {
    return ((struct ax_ScopeVec){.data=((struct ax_Scope*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_ScopeVec_push(struct ax_ScopeVec* self, struct ax_Scope scope) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_Scope* new_data = ((struct ax_Scope*)(ax_alloc((new_cap * 32))));
        if ((self->data != ((struct ax_Scope*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 32));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = scope;
    self->len = (self->len + 1);
    return idx;
}

struct ax_U32Vec ax_new_u32_vec(void) {
    return ((struct ax_U32Vec){.data=((ax_u32*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_U32Vec_push(struct ax_U32Vec* self, ax_u32 val) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        ax_u32* new_data = ((ax_u32*)(ax_alloc((new_cap * 4))));
        if ((self->data != ((ax_u32*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 4));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = val;
    self->len = (self->len + 1);
    return idx;
}

ax_u32 ax_U32Vec_pop(struct ax_U32Vec* self) {
    if ((self->len == 0)) {
        return ((ax_u32)(0));
    }
    self->len = (self->len - 1);
    return ((self->data)[self->len]);
}

ax_u32 ax_U32Vec_append_unique(struct ax_U32Vec* self, ax_u32 val) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->len)) {
        if ((((self->data)[i]) == val)) {
            return ((ax_u32)(i));
        }
        i = (i + 1);
    }
    return ax_U32Vec_push(self, val);
}

struct ax_SymbolTable ax_InternPool_new_symbol_table(struct ax_InternPool* intern) {
    struct ax_SymbolTable st = ((struct ax_SymbolTable){.symbols=ax_new_symbol_vec(), .scopes=ax_new_scope_vec(), .stack=ax_new_u32_vec()});
    struct ax_Scope global_scope = ((struct ax_Scope){.kind=ax_SCOPE_GLOBAL, .padding1=((ax_u8)(0)), .padding2=((ax_u16)(0)), .parent_id=((ax_u32)(0)), .depth=((ax_u32)(0)), .entries=((struct ax_ScopeEntry*)(NULL)), .count=0, .capacity=0});
    ax_Scope_init_scope(&(global_scope), ax_SCOPE_GLOBAL, ((ax_u32)(0)), ((ax_u32)(0)), ((ax_u32)(64)));
    ax_ScopeVec_push(&(st.scopes), global_scope);
    ax_U32Vec_push(&(st.stack), ((ax_u32)(0)));
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"i8", .len=2}, 1, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"i16", .len=3}, 2, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"i32", .len=3}, 3, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"i64", .len=3}, 4, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"u8", .len=2}, 5, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"u16", .len=3}, 6, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"u32", .len=3}, 7, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"u64", .len=3}, 8, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"f32", .len=3}, 9, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"f64", .len=3}, 10, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"bool", .len=4}, 11, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"string", .len=6}, 12, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"char8", .len=5}, 13, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"void", .len=4}, 14, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"isize", .len=5}, 15, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"usize", .len=5}, 16, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"ActorRef", .len=8}, 21, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"str", .len=3}, 12, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"ptr", .len=3}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"null", .len=4}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"alloc", .len=5}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"free", .len=4}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"memcpy", .len=6}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"memset", .len=6}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"print", .len=5}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"println", .len=7}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"compiler_intrinsic", .len=18}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"assert", .len=6}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"panic", .len=5}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"syscall0", .len=8}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"syscall1", .len=8}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"syscall2", .len=8}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"syscall3", .len=8}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"syscall4", .len=8}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"syscall5", .len=8}, 0, intern);
    ax_SymbolTable_define_builtin(&(st), (ax_string){.ptr=(const ax_u8*)"syscall6", .len=8}, 0, intern);
    return st;
    /* skip destroy for value type st */
}

void ax_SymbolTable_define_builtin(struct ax_SymbolTable* self, ax_string name, ax_u32 type_id, struct ax_InternPool* intern) {
    ax_u32 name_id = ax_InternPool_intern_string(intern, name);
    ax_u32 sym_idx = ((ax_u32)(self->symbols.len));
    ax_SymbolVec_push(&(self->symbols), ((struct ax_Symbol){.name_id=name_id, .kind=ax_SYM_BUILTIN_TYPE, .flags=ax_SYM_FLAG_PUB, .type_id=type_id, .decl_node=((ax_u32)(0)), .scope_id=((ax_u32)(0)), .next_overload=((ax_u32)(0))}));
    ax_Scope_scope_put(&(((self->scopes.data)[0])), name_id, sym_idx);
}

ax_u32 ax_SymbolTable_push_scope(struct ax_SymbolTable* self, ax_u8 kind) {
    ax_u32 parent_id = ax_SymbolTable_current_scope(*(self));
    ax_u32 depth = (ax_SymbolTable_current_depth(*(self)) + ((ax_u32)(1)));
    struct ax_Scope new_scope = ((struct ax_Scope){.kind=kind, .padding1=((ax_u8)(0)), .padding2=((ax_u16)(0)), .parent_id=parent_id, .depth=depth, .entries=((struct ax_ScopeEntry*)(NULL)), .count=0, .capacity=0});
    ax_Scope_init_scope(&(new_scope), kind, parent_id, depth, ((ax_u32)(16)));
    ax_u32 idx = ax_ScopeVec_push(&(self->scopes), new_scope);
    ax_U32Vec_push(&(self->stack), idx);
    return idx;
}

void ax_SymbolTable_pop_scope(struct ax_SymbolTable* self) {
    ax_U32Vec_pop(&(self->stack));
}

ax_u32 ax_SymbolTable_current_scope(struct ax_SymbolTable self) {
    return ((self.stack.data)[(self.stack.len - 1)]);
}

ax_u32 ax_SymbolTable_current_depth(struct ax_SymbolTable self) {
    return ((ax_u32)((self.stack.len - 1)));
}

ax_u32 ax_SymbolTable_define(struct ax_SymbolTable* self, ax_u32 name_id, ax_u8 kind, ax_u16 flags, ax_u32 decl_node) {
    ax_u32 scope_idx = ax_SymbolTable_current_scope(*(self));
    struct ax_ScopeEntry prev = ax_Scope_scope_get(((self->scopes.data)[scope_idx]), name_id);
    if ((prev.name_id != ((ax_u32)(0)))) {
        if (((((self->symbols.data)[prev.symbol_idx]).kind == ax_SYM_FUNC) && (kind == ax_SYM_FUNC))) {
            ax_u32 curr_idx = prev.symbol_idx;
            while (AX_TRUE) {
                if ((((self->symbols.data)[curr_idx]).next_overload == ((ax_u32)(0)))) {
                    break;
                }
                curr_idx = ((self->symbols.data)[curr_idx]).next_overload;
            }
            ax_u32 sym_idx = ((ax_u32)(self->symbols.len));
            ax_SymbolVec_push(&(self->symbols), ((struct ax_Symbol){.name_id=name_id, .kind=kind, .flags=flags, .type_id=((ax_u32)(0)), .decl_node=decl_node, .scope_id=scope_idx, .next_overload=((ax_u32)(0))}));
            ((self->symbols.data)[curr_idx]).next_overload = sym_idx;
            return sym_idx;
        }
        return prev.symbol_idx;
    }
    ax_u32 sym_idx = ((ax_u32)(self->symbols.len));
    ax_SymbolVec_push(&(self->symbols), ((struct ax_Symbol){.name_id=name_id, .kind=kind, .flags=flags, .type_id=((ax_u32)(0)), .decl_node=decl_node, .scope_id=scope_idx, .next_overload=((ax_u32)(0))}));
    ax_Scope_scope_put(&(((self->scopes.data)[scope_idx])), name_id, sym_idx);
    return sym_idx;
}

ax_u32 ax_SymbolTable_resolve(struct ax_SymbolTable self, ax_u32 name_id) {
    ax_i64 i = (self.stack.len - 1);
    while ((i >= 0)) {
        ax_u32 scope_idx = ((self.stack.data)[i]);
        struct ax_ScopeEntry entry = ax_Scope_scope_get(((self.scopes.data)[scope_idx]), name_id);
        if ((entry.name_id != ((ax_u32)(0)))) {
            return entry.symbol_idx;
        }
        i = (i - 1);
    }
    return ((ax_u32)(0));
}

struct ax_NameResolver ax_AstTree_new_name_resolver(struct ax_AstTree tree, struct ax_InternPool intern, struct ax_SymbolTable symtable) {
    return ((struct ax_NameResolver){.tree=tree, .intern=intern, .symtable=symtable});
}

void ax_NameResolver_resolve(struct ax_NameResolver* self) {
    if ((self->tree.nodes.len == 0)) {
        return;
    }
    ax_NameResolver_resolve_node(self, ((ax_u32)(0)));
}

void ax_NameResolver_resolve_children(struct ax_NameResolver* self, ax_u32 node_idx) {
    struct ax_AstNode node = ((self->tree.nodes.data)[node_idx]);
    ax_u32 child = node.first_child;
    while ((child != ((ax_u32)(0)))) {
        ax_NameResolver_resolve_node(self, child);
        child = ((self->tree.nodes.data)[child]).next_sibling;
    }
}

void ax_NameResolver_resolve_node(struct ax_NameResolver* self, ax_u32 node_idx) {
    ax_u8 kind = ((self->tree.nodes.data)[node_idx]).kind;
    ax_u16 flags = ((self->tree.nodes.data)[node_idx]).flags;
    if ((kind == ax_NODE_PROGRAM)) {
        ax_NameResolver_resolve_children(self, node_idx);
    } else if ((kind == ax_NODE_IMPORT_DECL)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_MODULE, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_string name = ax_InternPool_get(self->intern, name_id);
        if (ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"std.string", .len=10})) {
            ax_u32 len_name = ax_InternPool_intern_string(&(self->intern), (ax_string){.ptr=(const ax_u8*)"std.string.len", .len=14});
            ax_SymbolTable_define(&(self->symtable), len_name, ax_SYM_FUNC, ((ax_u16)(0)), node_idx);
            ax_u32 slice_name = ax_InternPool_intern_string(&(self->intern), (ax_string){.ptr=(const ax_u8*)"std.string.slice", .len=16});
            ax_SymbolTable_define(&(self->symtable), slice_name, ax_SYM_FUNC, ((ax_u16)(0)), node_idx);
            ax_u32 concat_name = ax_InternPool_intern_string(&(self->intern), (ax_string){.ptr=(const ax_u8*)"std.string.concat", .len=17});
            ax_SymbolTable_define(&(self->symtable), concat_name, ax_SYM_FUNC, ((ax_u16)(0)), node_idx);
            ax_u32 replace_name = ax_InternPool_intern_string(&(self->intern), (ax_string){.ptr=(const ax_u8*)"std.string.replace", .len=18});
            ax_SymbolTable_define(&(self->symtable), replace_name, ax_SYM_FUNC, ((ax_u16)(0)), node_idx);
        }
    } else if ((kind == ax_NODE_FIELD_EXPR)) {
        ax_u32 lhs_idx = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((lhs_idx != ((ax_u32)(0)))) {
            struct ax_AstNode lhs_node = ((self->tree.nodes.data)[lhs_idx]);
            ax_bool is_module_import = AX_FALSE;
            if ((lhs_node.kind == ax_NODE_IDENT)) {
                ax_string lhs_name = ax_InternPool_get(self->intern, lhs_node.payload);
                ax_string rhs_name = ax_InternPool_get(self->intern, ((self->tree.nodes.data)[node_idx]).payload);
                ax_string full_name = ax_str_concat(lhs_name, ax_str_concat((ax_string){.ptr=(const ax_u8*)".", .len=1}, rhs_name));
                ax_u32 full_name_id = ax_InternPool_intern_string(&(self->intern), full_name);
                ax_u32 sym_idx = ax_SymbolTable_resolve(self->symtable, full_name_id);
                if ((sym_idx != ((ax_u32)(0)))) {
                    struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
                    if ((sym.kind == ax_SYM_MODULE)) {
                        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
                        is_module_import = AX_TRUE;
                    }
                }
            }
            if ((!is_module_import)) {
                ax_NameResolver_resolve_node(self, lhs_idx);
                struct ax_AstNode updated_lhs_node = ((self->tree.nodes.data)[lhs_idx]);
                if (((updated_lhs_node.kind == ax_NODE_IDENT) || (updated_lhs_node.kind == ax_NODE_FIELD_EXPR))) {
                    ax_u32 sym_idx = updated_lhs_node.payload;
                    if ((sym_idx != ((ax_u32)(0)))) {
                        struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
                        if ((sym.kind == ax_SYM_MODULE)) {
                            ax_string field_name = ax_InternPool_get(self->intern, ((self->tree.nodes.data)[node_idx]).payload);
                            ax_string module_name = ax_InternPool_get(self->intern, sym.name_id);
                            ax_string full_name = ax_str_concat(module_name, ax_str_concat((ax_string){.ptr=(const ax_u8*)".", .len=1}, field_name));
                            ax_u32 full_name_id = ax_InternPool_intern_string(&(self->intern), full_name);
                            ax_u32 resolved_idx = ax_SymbolTable_resolve(self->symtable, full_name_id);
                            if ((resolved_idx != ((ax_u32)(0)))) {
                                ((self->tree.nodes.data)[node_idx]).payload = resolved_idx;
                                ((self->symtable.symbols.data)[resolved_idx]).flags = (((self->symtable.symbols.data)[resolved_idx]).flags | ax_SYM_FLAG_USED);
                            }
                        }
                    }
                }
            }
        }
    } else if ((kind == ax_NODE_FUNC_DECL)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u16 sym_flags = ((ax_u16)(0));
        if (((flags & ax_FLAG_IS_PUB) != ((ax_u16)(0)))) {
            sym_flags = (sym_flags | ax_SYM_FLAG_PUB);
        }
        if (((flags & ax_FLAG_IS_EXTERN) != ((ax_u16)(0)))) {
            sym_flags = (sym_flags | ax_SYM_FLAG_EXTERN);
        }
        if (((flags & ax_FLAG_IS_ASYNC) != ((ax_u16)(0)))) {
            sym_flags = (sym_flags | ax_SYM_FLAG_ASYNC);
        }
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_FUNC, sym_flags, node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_SymbolTable_push_scope(&(self->symtable), ax_SCOPE_FUNCTION);
        ax_NameResolver_resolve_children(self, node_idx);
        ax_SymbolTable_pop_scope(&(self->symtable));
    } else if ((kind == ax_NODE_PARAM_DECL)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_PARAM, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_NameResolver_resolve_children(self, node_idx);
    } else if ((kind == ax_NODE_STRUCT_DECL)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_STRUCT, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_SymbolTable_push_scope(&(self->symtable), ax_SCOPE_BLOCK);
        ax_NameResolver_resolve_children(self, node_idx);
        ax_SymbolTable_pop_scope(&(self->symtable));
    } else if ((kind == ax_NODE_FIELD_DECL)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_FIELD, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_NameResolver_resolve_children(self, node_idx);
    } else if ((kind == ax_NODE_INTERFACE_DECL)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_INTERFACE, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_SymbolTable_push_scope(&(self->symtable), ax_SCOPE_BLOCK);
        ax_u32 self_name_id = ax_InternPool_intern_string(&(self->intern), (ax_string){.ptr=(const ax_u8*)"Self", .len=4});
        ax_SymbolTable_define(&(self->symtable), self_name_id, ax_SYM_TYPE_ALIAS, ((ax_u16)(0)), node_idx);
        ax_NameResolver_resolve_children(self, node_idx);
        ax_SymbolTable_pop_scope(&(self->symtable));
    } else if ((kind == ax_NODE_VAR_DECL)) {
        ax_NameResolver_resolve_children(self, node_idx);
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u16 sym_flags = ((ax_u16)(0));
        if (((flags & ax_FLAG_IS_MUT) != ((ax_u16)(0)))) {
            sym_flags = (sym_flags | ax_SYM_FLAG_MUT);
        }
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_VAR, sym_flags, node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
    } else if ((kind == ax_NODE_CONST_DECL)) {
        ax_NameResolver_resolve_children(self, node_idx);
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_CONST, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
    } else if ((kind == ax_NODE_TYPE_ALIAS_DECL)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_TYPE_ALIAS, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_SymbolTable_push_scope(&(self->symtable), ax_SCOPE_BLOCK);
        ax_NameResolver_resolve_children(self, node_idx);
        ax_u32 curr_scope_idx = ax_SymbolTable_current_scope(self->symtable);
        ax_SymbolTable_pop_scope(&(self->symtable));
        ax_u32 parent_scope_idx = ax_SymbolTable_current_scope(self->symtable);
        struct ax_Scope curr_scope = ((self->symtable.scopes.data)[curr_scope_idx]);
        ax_i64 i = ((ax_i64)(0));
        while ((i < ((ax_i64)(curr_scope.capacity)))) {
            struct ax_ScopeEntry entry = ((curr_scope.entries)[i]);
            if ((entry.name_id != ((ax_u32)(0)))) {
                struct ax_Symbol sym = ((self->symtable.symbols.data)[entry.symbol_idx]);
                if ((sym.kind == ax_SYM_VARIANT)) {
                    ax_Scope_scope_put(&(((self->symtable.scopes.data)[parent_scope_idx])), entry.name_id, entry.symbol_idx);
                }
            }
            i = (i + 1);
        }
    } else if ((kind == ax_NODE_VARIANT_DECL)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_VARIANT, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_NameResolver_resolve_children(self, node_idx);
    } else if ((((((((kind == ax_NODE_IF_STMT) || (kind == ax_NODE_ELIF_CLAUSE)) || (kind == ax_NODE_ELSE_CLAUSE)) || (kind == ax_NODE_MATCH_ARM)) || (kind == ax_NODE_ARENA_BLOCK)) || (kind == ax_NODE_UNSAFE_BLOCK)) || (kind == ax_NODE_WHILE_STMT))) {
        ax_SymbolTable_push_scope(&(self->symtable), ax_SCOPE_BLOCK);
        ax_NameResolver_resolve_children(self, node_idx);
        ax_SymbolTable_pop_scope(&(self->symtable));
    } else if ((kind == ax_NODE_FOR_STMT)) {
        ax_SymbolTable_push_scope(&(self->symtable), ax_SCOPE_LOOP);
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_VAR, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_NameResolver_resolve_children(self, node_idx);
        ax_SymbolTable_pop_scope(&(self->symtable));
    } else if ((kind == ax_NODE_CLOSURE_EXPR)) {
        ax_SymbolTable_push_scope(&(self->symtable), ax_SCOPE_CLOSURE);
        ax_NameResolver_resolve_children(self, node_idx);
        ax_SymbolTable_pop_scope(&(self->symtable));
    } else if ((kind == ax_NODE_IDENT)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_string name = ax_InternPool_get(self->intern, name_id);
        if (((!ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"break", .len=5})) && (!ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"continue", .len=8})))) {
            ax_u32 sym_idx = ax_SymbolTable_resolve(self->symtable, name_id);
            if ((sym_idx != ((ax_u32)(0)))) {
                ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
                ((self->symtable.symbols.data)[sym_idx]).flags = (((self->symtable.symbols.data)[sym_idx]).flags | ax_SYM_FLAG_USED);
            }
        }
        ax_NameResolver_resolve_children(self, node_idx);
    } else if ((kind == ax_NODE_TYPE_EXPR)) {
        if ((((self->tree.nodes.data)[node_idx]).payload != ((ax_u32)(0)))) {
            ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
            ax_u32 sym_idx = ax_SymbolTable_resolve(self->symtable, name_id);
            if ((sym_idx != ((ax_u32)(0)))) {
                ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
                ((self->symtable.symbols.data)[sym_idx]).flags = (((self->symtable.symbols.data)[sym_idx]).flags | ax_SYM_FLAG_USED);
            }
        }
        ax_NameResolver_resolve_children(self, node_idx);
    } else if ((kind == ax_NODE_BINDING_PAT)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_VAR, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
    } else if ((kind == ax_NODE_VARIANT_PAT)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_resolve(self->symtable, name_id);
        if ((sym_idx != ((ax_u32)(0)))) {
            ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
            ((self->symtable.symbols.data)[sym_idx]).flags = (((self->symtable.symbols.data)[sym_idx]).flags | ax_SYM_FLAG_USED);
        }
        ax_NameResolver_resolve_children(self, node_idx);
    } else if ((kind == ax_NODE_GENERIC_PARAM)) {
        ax_u32 name_id = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 sym_idx = ax_SymbolTable_define(&(self->symtable), name_id, ax_SYM_GENERIC_PARAM, ((ax_u16)(0)), node_idx);
        ((self->tree.nodes.data)[node_idx]).payload = sym_idx;
        ax_NameResolver_resolve_children(self, node_idx);
    } else {
        {
            ax_NameResolver_resolve_children(self, node_idx);
        }
    }
}

void ax_SymbolTable_free_symtable(struct ax_SymbolTable* self) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->scopes.len)) {
        struct ax_Scope scope = ((self->scopes.data)[i]);
        if ((scope.entries != ((struct ax_ScopeEntry*)(NULL)))) {
            ax_free(((ax_u8*)(scope.entries)));
        }
        i = (i + 1);
    }
    if ((self->symbols.data != ((struct ax_Symbol*)(NULL)))) {
        ax_free(((ax_u8*)(self->symbols.data)));
    }
    if ((self->scopes.data != ((struct ax_Scope*)(NULL)))) {
        ax_free(((ax_u8*)(self->scopes.data)));
    }
    if ((self->stack.data != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(self->stack.data)));
    }
}

void ax_StructFieldVec_push(struct ax_StructFieldVec* self, struct ax_StructField sf) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_StructField* new_data = ((struct ax_StructField*)(ax_alloc((new_cap * 8))));
        if ((self->data != ((struct ax_StructField*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 8));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = sf;
    self->len = (self->len + 1);
}

void ax_StructInfoVec_push(struct ax_StructInfoVec* self, struct ax_StructInfo si) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_StructInfo* new_data = ((struct ax_StructInfo*)(ax_alloc((new_cap * 24))));
        if ((self->data != ((struct ax_StructInfo*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 24));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = si;
    self->len = (self->len + 1);
}

void ax_VariantInfoVec_push(struct ax_VariantInfoVec* self, struct ax_VariantInfo vi) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_VariantInfo* new_data = ((struct ax_VariantInfo*)(ax_alloc((new_cap * 16))));
        if ((self->data != ((struct ax_VariantInfo*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 16));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = vi;
    self->len = (self->len + 1);
}

void ax_SumInfoVec_push(struct ax_SumInfoVec* self, struct ax_SumInfo si) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_SumInfo* new_data = ((struct ax_SumInfo*)(ax_alloc((new_cap * 24))));
        if ((self->data != ((struct ax_SumInfo*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 24));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = si;
    self->len = (self->len + 1);
}

void ax_FuncInfoVec_push(struct ax_FuncInfoVec* self, struct ax_FuncInfo fi) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_FuncInfo* new_data = ((struct ax_FuncInfo*)(ax_alloc((new_cap * 32))));
        if ((self->data != ((struct ax_FuncInfo*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 32));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = fi;
    self->len = (self->len + 1);
}

ax_u32 ax_TypeEntryVec_push(struct ax_TypeEntryVec* self, struct ax_TypeEntry te) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_TypeEntry* new_data = ((struct ax_TypeEntry*)(ax_alloc((new_cap * 20))));
        if ((self->data != ((struct ax_TypeEntry*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 20));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = te;
    self->len = (self->len + 1);
    return idx;
}

struct ax_SumInfoVec ax_new_sum_info_vec(void) {
    return ((struct ax_SumInfoVec){.data=((struct ax_SumInfo*)(NULL)), .len=0, .cap=0});
}

struct ax_VariantInfoVec ax_new_variant_info_vec(void) {
    return ((struct ax_VariantInfoVec){.data=((struct ax_VariantInfo*)(NULL)), .len=0, .cap=0});
}

struct ax_TypeTable ax_new_type_table(void) {
    struct ax_TypeTable tt = ((struct ax_TypeTable){.entries=((struct ax_TypeEntryVec){.data=((struct ax_TypeEntry*)(NULL)), .len=0, .cap=0}), .structs=((struct ax_StructInfoVec){.data=((struct ax_StructInfo*)(NULL)), .len=0, .cap=0}), .funcs=((struct ax_FuncInfoVec){.data=((struct ax_FuncInfo*)(NULL)), .len=0, .cap=0}), .sumtypes=((struct ax_SumInfoVec){.data=((struct ax_SumInfo*)(NULL)), .len=0, .cap=0})});
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(0)), .align=((ax_u32)(0)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(1)), .align=((ax_u32)(1)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(2)), .align=((ax_u32)(2)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(4)), .align=((ax_u32)(4)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(1)), .align=((ax_u32)(1)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(2)), .align=((ax_u32)(2)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(4)), .align=((ax_u32)(4)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(4)), .align=((ax_u32)(4)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(1)), .align=((ax_u32)(1)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(1)), .align=((ax_u32)(1)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(0)), .align=((ax_u32)(0)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=((ax_u32)(0))}));
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_PRIMITIVE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=((ax_u32)(0))}));
    ax_i32 i = 0;
    while ((i < 4)) {
        ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_INTERFACE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(16)), .align=((ax_u32)(8)), .extra=((ax_u32)(0))}));
        i = (i + 1);
    }
    ax_TypeEntryVec_push(&(tt.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_STRUCT, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=((ax_u32)(0))}));
    return tt;
    /* skip destroy for value type tt */
}

ax_u32 ax_TypeTable_register_sum_type(struct ax_TypeTable* self, ax_u32 name_id, struct ax_VariantInfoVec variants) {
    ax_u32 s_idx = ((ax_u32)(self->sumtypes.len));
    ax_SumInfoVec_push(&(self->sumtypes), ((struct ax_SumInfo){.variants=variants}));
    return ax_TypeEntryVec_push(&(self->entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_SUM, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=name_id, .size=((ax_u32)(0)), .align=((ax_u32)(0)), .extra=s_idx}));
}

ax_u32 ax_TypeTable_register_struct(struct ax_TypeTable* self, ax_u32 name_id, struct ax_StructFieldVec fields) {
    ax_u32 s_idx = ((ax_u32)(self->structs.len));
    ax_StructInfoVec_push(&(self->structs), ((struct ax_StructInfo){.fields=fields}));
    return ax_TypeEntryVec_push(&(self->entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_STRUCT, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=name_id, .size=((ax_u32)(0)), .align=((ax_u32)(0)), .extra=s_idx}));
}

ax_u32 ax_TypeTable_register_function(struct ax_TypeTable* self, struct ax_U32Vec params, ax_u32 ret) {
    ax_u32 f_idx = ((ax_u32)(self->funcs.len));
    ax_FuncInfoVec_push(&(self->funcs), ((struct ax_FuncInfo){.params=params, .ret=ret, .dummy=((ax_u32)(0))}));
    return ax_TypeEntryVec_push(&(self->entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_FUNC, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=f_idx}));
}

ax_u32 ax_TypeTable_register_pointer(struct ax_TypeTable* self, ax_u32 inner_id) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->entries.len)) {
        struct ax_TypeEntry entry = ((self->entries.data)[i]);
        if (((entry.kind == ax_TYPE_KIND_POINTER) && (entry.extra == inner_id))) {
            return ((ax_u32)(i));
        }
        i = (i + 1);
    }
    return ax_TypeEntryVec_push(&(self->entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_POINTER, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=inner_id}));
}

ax_u32 ax_TypeTable_register_slice(struct ax_TypeTable* self, ax_u32 elem_id) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->entries.len)) {
        struct ax_TypeEntry entry = ((self->entries.data)[i]);
        if (((entry.kind == ax_TYPE_KIND_SLICE) && (entry.extra == elem_id))) {
            return ((ax_u32)(i));
        }
        i = (i + 1);
    }
    return ax_TypeEntryVec_push(&(self->entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_SLICE, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(24)), .align=((ax_u32)(8)), .extra=elem_id}));
}

ax_u32 ax_TypeTable_register_array(struct ax_TypeTable* self, ax_u32 elem_id) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->entries.len)) {
        struct ax_TypeEntry entry = ((self->entries.data)[i]);
        if (((entry.kind == ax_TYPE_KIND_ARRAY) && (entry.extra == elem_id))) {
            return ((ax_u32)(i));
        }
        i = (i + 1);
    }
    return ax_TypeEntryVec_push(&(self->entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_ARRAY, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=((ax_u32)(0)), .size=((ax_u32)(8)), .align=((ax_u32)(8)), .extra=elem_id}));
}

void ax_TypeTable_free_typetable(struct ax_TypeTable* self) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->structs.len)) {
        struct ax_StructInfo s = ((self->structs.data)[i]);
        if ((s.fields.data != ((struct ax_StructField*)(NULL)))) {
            ax_free(((ax_u8*)(s.fields.data)));
        }
        i = (i + 1);
    }
    i = ((ax_i64)(0));
    while ((i < self->funcs.len)) {
        struct ax_FuncInfo f = ((self->funcs.data)[i]);
        if ((f.params.data != ((ax_u32*)(NULL)))) {
            ax_free(((ax_u8*)(f.params.data)));
        }
        i = (i + 1);
    }
    i = ((ax_i64)(0));
    while ((i < self->sumtypes.len)) {
        struct ax_SumInfo s = ((self->sumtypes.data)[i]);
        if ((s.variants.data != ((struct ax_VariantInfo*)(NULL)))) {
            ax_free(((ax_u8*)(s.variants.data)));
        }
        i = (i + 1);
    }
    if ((self->sumtypes.data != ((struct ax_SumInfo*)(NULL)))) {
        ax_free(((ax_u8*)(self->sumtypes.data)));
    }
    if ((self->entries.data != ((struct ax_TypeEntry*)(NULL)))) {
        ax_free(((ax_u8*)(self->entries.data)));
    }
    if ((self->structs.data != ((struct ax_StructInfo*)(NULL)))) {
        ax_free(((ax_u8*)(self->structs.data)));
    }
    if ((self->funcs.data != ((struct ax_FuncInfo*)(NULL)))) {
        ax_free(((ax_u8*)(self->funcs.data)));
    }
}

struct ax_TypeSubstVec ax_new_type_subst_vec(void) {
    return ((struct ax_TypeSubstVec){.data=((struct ax_TypeSubst*)(NULL)), .len=0, .cap=0});
}

void ax_TypeSubstVec_push(struct ax_TypeSubstVec* self, struct ax_TypeSubst subst) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(8));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_TypeSubst* new_data = ((struct ax_TypeSubst*)(ax_alloc((new_cap * 8))));
        if ((self->data != ((struct ax_TypeSubst*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 8));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = subst;
    self->len = (self->len + ((ax_i64)(1)));
}

ax_u32 ax_TypeSubstVec_lookup_subst(struct ax_TypeSubstVec self, ax_u32 name_id) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self.len)) {
        if ((((self.data)[i]).name_id == name_id)) {
            return ((self.data)[i]).type_id;
        }
        i = (i + ((ax_i64)(1)));
    }
    return ((ax_u32)(0));
}

struct ax_Monomorphizer ax_AstTree_new_monomorphizer(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable) {
    return ((struct ax_Monomorphizer){.tree=tree, .pool=pool, .symtable=symtable, .typetable=typetable});
}

void ax_Monomorphizer_substitute_type_params(struct ax_Monomorphizer* self, ax_u32 node_idx, struct ax_TypeSubstVec subst) {
    if (((node_idx == ((ax_u32)(0))) || (node_idx == ((ax_u32)(0xffffffff))))) {
        return;
    }
    struct ax_AstNode* node = &(((self->tree.nodes.data)[node_idx]));
    ax_u8 kind = node->kind;
    if (((((((((((((kind == ax_NODE_IDENT) || (kind == ax_NODE_PARAM_DECL)) || (kind == ax_NODE_FIELD_DECL)) || (kind == ax_NODE_FUNC_DECL)) || (kind == ax_NODE_STRUCT_DECL)) || (kind == ax_NODE_TYPE_ALIAS_DECL)) || (kind == ax_NODE_VARIANT_DECL)) || (kind == ax_NODE_TYPE_EXPR)) || (kind == ax_NODE_VAR_DECL)) || (kind == ax_NODE_CONST_DECL)) || (kind == ax_NODE_BINDING_PAT)) || (kind == ax_NODE_VARIANT_PAT))) {
        ax_u32 sym_idx = node->payload;
        if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symtable.symbols.len))) {
            struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
            if (((((kind == ax_NODE_IDENT) || (kind == ax_NODE_TYPE_EXPR)) || (kind == ax_NODE_BINDING_PAT)) || (kind == ax_NODE_VARIANT_PAT))) {
                ax_u32 type_id = ax_TypeSubstVec_lookup_subst(subst, sym.name_id);
                if ((type_id != ((ax_u32)(0)))) {
                    if ((((ax_i64)(type_id)) < self->typetable.entries.len)) {
                        struct ax_TypeEntry entry = ((self->typetable.entries.data)[type_id]);
                        ax_u32 name_id = entry.name_id;
                        if ((name_id == ((ax_u32)(0)))) {
                            if ((type_id == ((ax_u32)(3)))) {
                                name_id = ax_InternPool_intern(&(self->pool), (ax_string){.ptr=(const ax_u8*)"i32", .len=3});
                            } else if ((type_id == ((ax_u32)(4)))) {
                                name_id = ax_InternPool_intern(&(self->pool), (ax_string){.ptr=(const ax_u8*)"i64", .len=3});
                            } else if ((type_id == ((ax_u32)(11)))) {
                                name_id = ax_InternPool_intern(&(self->pool), (ax_string){.ptr=(const ax_u8*)"bool", .len=4});
                            } else if ((type_id == ((ax_u32)(12)))) {
                                name_id = ax_InternPool_intern(&(self->pool), (ax_string){.ptr=(const ax_u8*)"string", .len=6});
                            } else {
                                {
                                    name_id = ax_InternPool_intern(&(self->pool), (ax_string){.ptr=(const ax_u8*)"type_val", .len=8});
                                }
                            }
                        }
                        node->payload = name_id;
                    }
                } else {
                    {
                        node->payload = sym.name_id;
                    }
                }
            } else {
                {
                    node->payload = sym.name_id;
                }
            }
        }
    }
    ax_u32 child = node->first_child;
    while ((child != ((ax_u32)(0)))) {
        ax_Monomorphizer_substitute_type_params(self, child, subst);
        child = ((self->tree.nodes.data)[child]).next_sibling;
    }
}

void ax_Monomorphizer_remove_generic_params_child(struct ax_Monomorphizer* self, ax_u32 node_idx) {
    struct ax_AstNode* node = &(((self->tree.nodes.data)[node_idx]));
    ax_u32 prev = ((ax_u32)(0));
    ax_u32 curr = node->first_child;
    while ((curr != ((ax_u32)(0)))) {
        if ((((self->tree.nodes.data)[curr]).kind == ax_NODE_GENERIC_PARAMS)) {
            ax_u32 next = ((self->tree.nodes.data)[curr]).next_sibling;
            if ((prev == ((ax_u32)(0)))) {
                node->first_child = next;
            } else {
                {
                    ((self->tree.nodes.data)[prev]).next_sibling = next;
                }
            }
            break;
        }
        prev = curr;
        curr = ((self->tree.nodes.data)[curr]).next_sibling;
    }
}

ax_string ax_Monomorphizer_mangle_name(struct ax_Monomorphizer self, ax_string orig_name, struct ax_U32Vec args) {
    ax_string base_mangled = (ax_string){.ptr=(const ax_u8*)"_AX_std_", .len=8};
    ax_i64 len_base = ax_str_len(base_mangled);
    ax_i64 len_orig = ax_str_len(orig_name);
    ax_i64 total_len = (len_base + len_orig);
    ax_i64 idx = ((ax_i64)(0));
    while ((idx < args.len)) {
        ax_u32 arg_type = ((args.data)[idx]);
        ax_string name_str = (ax_string){.ptr=(const ax_u8*)"type", .len=4};
        if ((((ax_i64)(arg_type)) < self.typetable.entries.len)) {
            struct ax_TypeEntry entry = ((self.typetable.entries.data)[arg_type]);
            if ((entry.name_id != ((ax_u32)(0)))) {
                name_str = ax_InternPool_get(self.pool, entry.name_id);
            } else {
                {
                    if ((arg_type == ((ax_u32)(3)))) {
                        name_str = (ax_string){.ptr=(const ax_u8*)"i32", .len=3};
                    } else if ((arg_type == ((ax_u32)(4)))) {
                        name_str = (ax_string){.ptr=(const ax_u8*)"i64", .len=3};
                    } else if ((arg_type == ((ax_u32)(11)))) {
                        name_str = (ax_string){.ptr=(const ax_u8*)"bool", .len=4};
                    } else if ((arg_type == ((ax_u32)(12)))) {
                        name_str = (ax_string){.ptr=(const ax_u8*)"string", .len=6};
                    }
                }
            }
        }
        total_len = ((total_len + ((ax_i64)(1))) + ax_str_len(name_str));
        idx = (idx + ((ax_i64)(1)));
    }
    ax_u8* buf = ((ax_u8*)(ax_alloc((total_len + ((ax_i64)(1))))));
    memcpy(buf, ((ax_u8*)(base_mangled.ptr)), len_base);
    memcpy(((ax_u8*)((((ax_i64)(buf)) + len_base))), ((ax_u8*)(orig_name.ptr)), len_orig);
    ax_i64 curr_offset = (len_base + len_orig);
    idx = 0;
    while ((idx < args.len)) {
        ax_u32 arg_type = ((args.data)[idx]);
        ax_string name_str = (ax_string){.ptr=(const ax_u8*)"type", .len=4};
        if ((((ax_i64)(arg_type)) < self.typetable.entries.len)) {
            struct ax_TypeEntry entry = ((self.typetable.entries.data)[arg_type]);
            if ((entry.name_id != ((ax_u32)(0)))) {
                name_str = ax_InternPool_get(self.pool, entry.name_id);
            } else {
                {
                    if ((arg_type == ((ax_u32)(3)))) {
                        name_str = (ax_string){.ptr=(const ax_u8*)"i32", .len=3};
                    } else if ((arg_type == ((ax_u32)(4)))) {
                        name_str = (ax_string){.ptr=(const ax_u8*)"i64", .len=3};
                    } else if ((arg_type == ((ax_u32)(11)))) {
                        name_str = (ax_string){.ptr=(const ax_u8*)"bool", .len=4};
                    } else if ((arg_type == ((ax_u32)(12)))) {
                        name_str = (ax_string){.ptr=(const ax_u8*)"string", .len=6};
                    }
                }
            }
        }
        ((buf)[curr_offset]) = ((ax_u8)('_'));
        curr_offset = (curr_offset + ((ax_i64)(1)));
        ax_i64 len_name = ax_str_len(name_str);
        memcpy(((ax_u8*)((((ax_i64)(buf)) + curr_offset))), ((ax_u8*)(name_str.ptr)), len_name);
        curr_offset = (curr_offset + len_name);
        idx = (idx + ((ax_i64)(1)));
    }
    ((buf)[total_len]) = ((ax_u8)(0));
    return ((ax_string){.ptr = (const ax_u8*)(buf), .len = strlen((const char*)(buf))});
}

ax_u32 ax_Monomorphizer_instantiate_function(struct ax_Monomorphizer* self, ax_u32 template_sym_id, struct ax_U32Vec args) {
    if ((((ax_i64)(template_sym_id)) >= self->symtable.symbols.len)) {
        return ((ax_u32)(0));
    }
    struct ax_Symbol orig_sym = ((self->symtable.symbols.data)[template_sym_id]);
    ax_u32 node_idx = orig_sym.decl_node;
    if ((node_idx == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 cloned_root = ax_AstTree_clone_subtree(&(self->tree), node_idx);
    ax_u32 struct_type_id = ((ax_u32)(0));
    ax_u8 cloned_kind = ((self->tree.nodes.data)[cloned_root]).kind;
    if ((cloned_kind == ax_NODE_STRUCT_DECL)) {
        struct_type_id = ((ax_u32)(self->typetable.entries.len));
    }
    struct ax_TypeSubstVec subst = ax_new_type_subst_vec();
    if ((cloned_kind == ax_NODE_STRUCT_DECL)) {
        ax_TypeSubstVec_push(&(subst), ((struct ax_TypeSubst){.name_id=orig_sym.name_id, .type_id=struct_type_id}));
    }
    ax_u32 curr = ((self->tree.nodes.data)[node_idx]).first_child;
    while ((curr != ((ax_u32)(0)))) {
        if ((((self->tree.nodes.data)[curr]).kind == ax_NODE_GENERIC_PARAMS)) {
            ax_u32 gp = ((self->tree.nodes.data)[curr]).first_child;
            ax_i64 idx = ((ax_i64)(0));
            while (((gp != ((ax_u32)(0))) && (idx < args.len))) {
                ax_u32 gp_name_id = ((self->tree.nodes.data)[gp]).payload;
                ax_TypeSubstVec_push(&(subst), ((struct ax_TypeSubst){.name_id=gp_name_id, .type_id=((args.data)[idx])}));
                gp = ((self->tree.nodes.data)[gp]).next_sibling;
                idx = (idx + ((ax_i64)(1)));
            }
            break;
        }
        curr = ((self->tree.nodes.data)[curr]).next_sibling;
    }
    ax_Monomorphizer_substitute_type_params(self, cloned_root, subst);
    ((self->tree.nodes.data)[cloned_root]).flags = (((self->tree.nodes.data)[cloned_root]).flags & (~ax_FLAG_IS_GENERIC));
    ax_Monomorphizer_remove_generic_params_child(self, cloned_root);
    ax_string orig_name = ax_InternPool_get(self->pool, orig_sym.name_id);
    ax_string mangled = ax_Monomorphizer_mangle_name(*(self), orig_name, args);
    ax_u32 mangled_id = ax_InternPool_intern(&(self->pool), mangled);
    ((self->tree.nodes.data)[cloned_root]).payload = mangled_id;
    ax_u32 inst_sym_idx = ((ax_u32)(self->symtable.symbols.len));
    struct ax_Symbol new_sym = orig_sym;
    new_sym.name_id = mangled_id;
    new_sym.decl_node = cloned_root;
    new_sym.flags = (new_sym.flags & (~ax_SYM_FLAG_GENERIC));
    if ((cloned_kind == ax_NODE_STRUCT_DECL)) {
        ax_u32 s_idx = ((ax_u32)(self->typetable.structs.len));
        struct ax_StructFieldVec fields = ((struct ax_StructFieldVec){.data=((struct ax_StructField*)(NULL)), .len=0, .cap=0});
        ax_StructInfoVec_push(&(self->typetable.structs), ((struct ax_StructInfo){.fields=fields}));
        ax_TypeEntryVec_push(&(self->typetable.entries), ((struct ax_TypeEntry){.kind=ax_TYPE_KIND_STRUCT, .padding=((ax_u8)(0)), .flags=((ax_u16)(0)), .name_id=mangled_id, .size=((ax_u32)(0)), .align=((ax_u32)(0)), .extra=s_idx}));
        new_sym.type_id = struct_type_id;
        ((self->tree.nodes.data)[cloned_root]).payload = inst_sym_idx;
    } else {
        {
            ((self->tree.nodes.data)[cloned_root]).payload = inst_sym_idx;
        }
    }
    ax_SymbolVec_push(&(self->symtable.symbols), new_sym);
    ax_free(((ax_u8*)(subst.data)));
    return inst_sym_idx;
}

ax_string ax_strip_quotes(ax_string s) {
    ax_i64 len = ax_str_len(s);
    if ((((len >= 2) && (((ax_u8)((ax_bounds_check((ax_u64)(0), (s).len), (s).ptr[0]))) == ((ax_u8)('"')))) && (((ax_u8)((ax_bounds_check((ax_u64)((len - 1)), (s).len), (s).ptr[(len - 1)]))) == ((ax_u8)('"'))))) {
        ax_u8* ptr_start = ((ax_u8*)((((ax_i64)(((ax_u8*)(s.ptr)))) + ((ax_i64)(1)))));
        return ax_alloc_str_from_raw(ptr_start, (len - 2));
    }
    return s;
}

ax_string ax_format_int(ax_i64 val) {
    if ((val == ((ax_i64)(0)))) {
        return (ax_string){.ptr=(const ax_u8*)"0", .len=1};
    }
    if ((val == (-((ax_i64)(9223372036854775808))))) {
        return (ax_string){.ptr=(const ax_u8*)"-9223372036854775808", .len=20};
    }
    ax_i64 temp = val;
    ax_bool is_neg = AX_FALSE;
    if ((val < ((ax_i64)(0)))) {
        is_neg = AX_TRUE;
        temp = (((ax_i64)(0)) - val);
    }
    ax_u8* buf = ((ax_u8*)(ax_alloc(24)));
    ax_i32 len = 0;
    while ((temp > ((ax_i64)(0)))) {
        ax_u8 rem = ((ax_u8)((temp % ((ax_i64)(10)))));
        ((buf)[len]) = (((ax_u8)('0')) + rem);
        len = (len + 1);
        temp = (temp / ((ax_i64)(10)));
    }
    if (is_neg) {
        ((buf)[len]) = ((ax_u8)('-'));
        len = (len + 1);
    }
    ax_i32 i = 0;
    ax_i32 j = (len - 1);
    while ((i < j)) {
        ax_u8 tmp = ((buf)[i]);
        ((buf)[i]) = ((buf)[j]);
        ((buf)[j]) = tmp;
        i = (i + 1);
        j = (j - 1);
    }
    ((buf)[len]) = ((ax_u8)(0));
    ax_string result = ax_alloc_str_from_raw(buf, ((ax_i64)(len)));
    ax_free(buf);
    return result;
}

ax_string ax_format_float(ax_f64 val) {
    ax_u8* buf = ((ax_u8*)(ax_alloc(64)));
    sprintf(buf, ((ax_u8*)((ax_string){.ptr=(const ax_u8*)"%g", .len=2}.ptr)), val);
    ax_i64 len = strlen(buf);
    ax_string result = ax_alloc_str_from_raw(buf, len);
    ax_free(buf);
    return result;
}

struct ax_TypeChecker ax_AstTree_new_type_checker(struct ax_AstTree tree, struct ax_InternPool intern, struct ax_SymbolTable symtable, struct ax_TypeTable types) {
    ax_i64 node_count = tree.nodes.len;
    ax_i64 initial_cap = node_count;
    if ((initial_cap < ((ax_i64)(16)))) {
        initial_cap = ((ax_i64)(16));
    }
    ax_u32* node_types = ((ax_u32*)(ax_alloc((initial_cap * 4))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < initial_cap)) {
        ((node_types)[i]) = ((ax_u32)(0));
        i = (i + 1);
    }
    return ((struct ax_TypeChecker){.tree=tree, .intern=intern, .symtable=symtable, .types=types, .node_types=node_types, .node_types_cap=initial_cap, .current_return=((ax_u32)(0)), .current_match_scrutinee=((ax_u32)(0))});
}

void ax_TypeChecker_free_type_checker(struct ax_TypeChecker* self) {
    ax_free(((ax_u8*)(self->node_types)));
}

void ax_TypeChecker_set_node_type(struct ax_TypeChecker* self, ax_u32 node_idx, ax_u32 type_id) {
    if ((node_idx == ((ax_u32)(0xffffffff)))) {
        return;
    }
    ax_i64 idx = ((ax_i64)(node_idx));
    if ((idx >= self->node_types_cap)) {
        ax_i64 new_cap = (self->node_types_cap * 2);
        if ((idx >= new_cap)) {
            new_cap = (idx + ((ax_i64)(100)));
        }
        ax_u32* new_ptr = ((ax_u32*)(ax_alloc((new_cap * 4))));
        ax_i64 i = ((ax_i64)(0));
        while ((i < self->node_types_cap)) {
            ((new_ptr)[i]) = ((self->node_types)[i]);
            i = (i + 1);
        }
        while ((i < new_cap)) {
            ((new_ptr)[i]) = ((ax_u32)(0));
            i = (i + 1);
        }
        ax_free(((ax_u8*)(self->node_types)));
        self->node_types = new_ptr;
        self->node_types_cap = new_cap;
    }
    ((self->node_types)[idx]) = type_id;
}

ax_i64 ax_parse_comptime_int(ax_string s) {
    ax_i64 len = ax_str_len(s);
    if ((len == 0)) {
        return ((ax_i64)(0));
    }
    ax_i32 i = 0;
    ax_i64 sign = ((ax_i64)(1));
    if ((((ax_u8)((ax_bounds_check((ax_u64)(0), (s).len), (s).ptr[0]))) == ((ax_u8)('-')))) {
        sign = (-((ax_i64)(1)));
        i = 1;
    } else if ((((ax_u8)((ax_bounds_check((ax_u64)(0), (s).len), (s).ptr[0]))) == ((ax_u8)('+')))) {
        i = 1;
    }
    ax_i64 base = ((ax_i64)(10));
    if ((((i + 1) < len) && (((ax_u8)((ax_bounds_check((ax_u64)(i), (s).len), (s).ptr[i]))) == ((ax_u8)('0'))))) {
        ax_u8 next_char = ((ax_u8)((ax_bounds_check((ax_u64)((i + 1)), (s).len), (s).ptr[(i + 1)])));
        if (((next_char == ((ax_u8)('x'))) || (next_char == ((ax_u8)('X'))))) {
            base = ((ax_i64)(16));
            i = (i + 2);
        } else if (((next_char == ((ax_u8)('o'))) || (next_char == ((ax_u8)('O'))))) {
            base = ((ax_i64)(8));
            i = (i + 2);
        } else if (((next_char == ((ax_u8)('b'))) || (next_char == ((ax_u8)('B'))))) {
            base = ((ax_i64)(2));
            i = (i + 2);
        }
    }
    ax_i64 val = ((ax_i64)(0));
    while ((i < len)) {
        ax_u8 c = ((ax_u8)((ax_bounds_check((ax_u64)(i), (s).len), (s).ptr[i])));
        ax_i64 digit = (-((ax_i64)(1)));
        if (((c >= ((ax_u8)('0'))) && (c <= ((ax_u8)('9'))))) {
            digit = ((ax_i64)((c - ((ax_u8)('0')))));
        } else if (((c >= ((ax_u8)('a'))) && (c <= ((ax_u8)('f'))))) {
            digit = ((ax_i64)(((c - ((ax_u8)('a'))) + ((ax_u8)(10)))));
        } else if (((c >= ((ax_u8)('A'))) && (c <= ((ax_u8)('F'))))) {
            digit = ((ax_i64)(((c - ((ax_u8)('A'))) + ((ax_u8)(10)))));
        }
        if (((digit < 0) || (digit >= base))) {
            break;
        }
        val = ((val * base) + digit);
        i = (i + 1);
    }
    return (val * sign);
}

ax_string ax_TypeChecker_node_text(struct ax_TypeChecker self, ax_u32 node_idx) {
    struct ax_AstNode node = ((self.tree.nodes.data)[node_idx]);
    ax_u32 tok_idx = node.token_idx;
    if ((((ax_i64)(tok_idx)) >= self.tree.tokens.len)) {
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    struct ax_Token tok = ((self.tree.tokens.data)[((ax_i64)(tok_idx))]);
    if ((tok.len == ((ax_u16)(0)))) {
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    ax_u8* start_ptr = ((ax_u8*)((((ax_i64)(((ax_u8*)(self.tree.src.ptr)))) + ((ax_i64)(tok.offset)))));
    return ax_alloc_str_from_raw(start_ptr, ((ax_i64)(tok.len)));
}

struct ax_ComptimeValue ax_TypeChecker_eval_comptime(struct ax_TypeChecker* self, ax_u32 node_idx) {
    struct ax_AstNode node = ((self->tree.nodes.data)[node_idx]);
    ax_u8 kind = node.kind;
    if ((kind == ax_NODE_INT_LIT)) {
        ax_string text = ax_TypeChecker_node_text(*(self), node_idx);
        ax_i64 val = ax_parse_comptime_int(text);
        return ((struct ax_ComptimeValue){.kind=ax_TYPE_I64, .int_val=val, .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
    } else if ((kind == ax_NODE_FLOAT_LIT)) {
        ax_string text = ax_TypeChecker_node_text(*(self), node_idx);
        ax_f64 val = atof((const char*)(text).ptr);
        return ((struct ax_ComptimeValue){.kind=ax_TYPE_F64, .int_val=((ax_i64)(0)), .float_val=val, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
    } else if ((kind == ax_NODE_BOOL_LIT)) {
        ax_string text = ax_TypeChecker_node_text(*(self), node_idx);
        ax_bool is_true = ax_str_eq(text, (ax_string){.ptr=(const ax_u8*)"true", .len=4});
        return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=is_true, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
    } else if ((kind == ax_NODE_STRING_LIT)) {
        ax_string text = ax_TypeChecker_node_text(*(self), node_idx);
        ax_string stripped = ax_strip_quotes(text);
        return ((struct ax_ComptimeValue){.kind=ax_TYPE_STRING, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=stripped});
    } else if ((kind == ax_NODE_COMPTIME)) {
        ax_u32 child = node.first_child;
        if ((child != ((ax_u32)(0)))) {
            return ax_TypeChecker_eval_comptime(self, child);
        }
        return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
    } else if ((kind == ax_NODE_BLOCK)) {
        ax_u32 child = node.first_child;
        if ((child == ((ax_u32)(0)))) {
            return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
        }
        struct ax_ComptimeValue last_val = ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
        while ((child != ((ax_u32)(0)))) {
            last_val = ax_TypeChecker_eval_comptime(self, child);
            child = ((self->tree.nodes.data)[child]).next_sibling;
        }
        return last_val;
        /* skip destroy for value type last_val */
    } else if ((kind == ax_NODE_RETURN_STMT)) {
        ax_u32 child = node.first_child;
        if ((child != ((ax_u32)(0)))) {
            return ax_TypeChecker_eval_comptime(self, child);
        }
        return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
    } else if ((kind == ax_NODE_UNARY_EXPR)) {
        ax_u32 child = node.first_child;
        if ((child == ((ax_u32)(0)))) {
            return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
        }
        struct ax_ComptimeValue val = ax_TypeChecker_eval_comptime(self, child);
        ax_string op = ax_TypeChecker_node_text(*(self), node_idx);
        if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"-", .len=1})) {
            if ((val.kind == ax_TYPE_I64)) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_I64, .int_val=(((ax_i64)(0)) - val.int_val), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            } else if ((val.kind == ax_TYPE_F64)) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_F64, .int_val=((ax_i64)(0)), .float_val=(0.0 - val.float_val), .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            }
        } else if ((ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"not", .len=3}) || ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"!", .len=1}))) {
            if ((val.kind == ax_TYPE_BOOL)) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(!val.bool_val), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            }
        } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"~", .len=1})) {
            if ((val.kind == ax_TYPE_I64)) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_I64, .int_val=((((ax_i64)(0)) - val.int_val) - ((ax_i64)(1))), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            }
        }
        return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
    } else if ((kind == ax_NODE_BINARY_EXPR)) {
        ax_u32 lhs_idx = node.first_child;
        if ((lhs_idx == ((ax_u32)(0)))) {
            return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
        }
        ax_u32 rhs_idx = ((self->tree.nodes.data)[lhs_idx]).next_sibling;
        if ((rhs_idx == ((ax_u32)(0)))) {
            return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
        }
        struct ax_ComptimeValue lhs = ax_TypeChecker_eval_comptime(self, lhs_idx);
        struct ax_ComptimeValue rhs = ax_TypeChecker_eval_comptime(self, rhs_idx);
        ax_string op = ax_TypeChecker_node_text(*(self), node_idx);
        ax_bool is_lhs_num = ((lhs.kind == ax_TYPE_I64) || (lhs.kind == ax_TYPE_F64));
        ax_bool is_rhs_num = ((rhs.kind == ax_TYPE_I64) || (rhs.kind == ax_TYPE_F64));
        if ((is_lhs_num && is_rhs_num)) {
            if (((lhs.kind == ax_TYPE_F64) || (rhs.kind == ax_TYPE_F64))) {
                ax_f64 lf = lhs.float_val;
                if ((lhs.kind == ax_TYPE_I64)) {
                    lf = ((ax_f64)(lhs.int_val));
                }
                ax_f64 rf = rhs.float_val;
                if ((rhs.kind == ax_TYPE_I64)) {
                    rf = ((ax_f64)(rhs.int_val));
                }
                if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"+", .len=1})) {
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_F64, .int_val=((ax_i64)(0)), .float_val=(lf + rf), .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"-", .len=1})) {
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_F64, .int_val=((ax_i64)(0)), .float_val=(lf - rf), .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"*", .len=1})) {
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_F64, .int_val=((ax_i64)(0)), .float_val=(lf * rf), .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"/", .len=1})) {
                    if ((rf == 0.0)) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    }
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_F64, .int_val=((ax_i64)(0)), .float_val=(lf / rf), .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"==", .len=2})) {
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lf == rf), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"!=", .len=2})) {
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lf != rf), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"<", .len=1})) {
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lf < rf), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"<=", .len=2})) {
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lf <= rf), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)">", .len=1})) {
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lf > rf), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)">=", .len=2})) {
                    return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lf >= rf), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                }
            } else {
                {
                    ax_i64 lh = lhs.int_val;
                    ax_i64 rh = rhs.int_val;
                    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"+", .len=1})) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_I64, .int_val=(lh + rh), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"-", .len=1})) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_I64, .int_val=(lh - rh), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"*", .len=1})) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_I64, .int_val=(lh * rh), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"/", .len=1})) {
                        if ((rh == ((ax_i64)(0)))) {
                            return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                        }
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_I64, .int_val=(lh / rh), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"%", .len=1})) {
                        if ((rh == ((ax_i64)(0)))) {
                            return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                        }
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_I64, .int_val=(lh % rh), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"==", .len=2})) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lh == rh), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"!=", .len=2})) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lh != rh), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"<", .len=1})) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lh < rh), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"<=", .len=2})) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lh <= rh), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)">", .len=1})) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lh > rh), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)">=", .len=2})) {
                        return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lh >= rh), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
                    }
                }
            }
        }
        if (((lhs.kind == ax_TYPE_BOOL) && (rhs.kind == ax_TYPE_BOOL))) {
            ax_bool lb = lhs.bool_val;
            ax_bool rb = rhs.bool_val;
            if ((ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"and", .len=3}) || ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"&&", .len=2}))) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lb && rb), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            } else if ((ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"or", .len=2}) || ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"||", .len=2}))) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lb || rb), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"==", .len=2})) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lb == rb), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"!=", .len=2})) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(lb != rb), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            }
        }
        if (((lhs.kind == ax_TYPE_STRING) && (rhs.kind == ax_TYPE_STRING))) {
            ax_string ls = lhs.str_val;
            ax_string rs = rhs.str_val;
            if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"+", .len=1})) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_STRING, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=ax_str_concat(ls, rs)});
            } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"==", .len=2})) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=ax_str_eq(ls, rs), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            } else if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"!=", .len=2})) {
                return ((struct ax_ComptimeValue){.kind=ax_TYPE_BOOL, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=(!ax_str_eq(ls, rs)), .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
            }
        }
    }
    return ((struct ax_ComptimeValue){.kind=ax_TYPE_VOID, .int_val=((ax_i64)(0)), .float_val=0.0, .bool_val=AX_FALSE, .str_val=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
}

void ax_TypeChecker_run_type_checker(struct ax_TypeChecker* self) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->tree.nodes.len)) {
        struct ax_AstNode node = ((self->tree.nodes.data)[i]);
        if ((node.kind == ax_NODE_TYPE_ALIAS_DECL)) {
            ax_TypeChecker_pre_infer_type_alias(self, ((ax_u32)(i)));
        }
        i = (i + 1);
    }
    i = ((ax_i64)(0));
    while ((i < self->tree.nodes.len)) {
        struct ax_AstNode node = ((self->tree.nodes.data)[i]);
        if ((node.kind == ax_NODE_STRUCT_DECL)) {
            ax_TypeChecker_pre_infer_struct(self, ((ax_u32)(i)));
        }
        i = (i + 1);
    }
    i = ((ax_i64)(0));
    while ((i < self->tree.nodes.len)) {
        struct ax_AstNode node2 = ((self->tree.nodes.data)[i]);
        if ((node2.kind == ax_NODE_FUNC_DECL)) {
            ax_TypeChecker_pre_infer_func_signature(self, ((ax_u32)(i)));
        }
        i = (i + 1);
    }
    ax_TypeChecker_infer_node(self, ((ax_u32)(0)), ax_TYPE_UNKNOWN);
}

void ax_TypeChecker_pre_infer_type_alias(struct ax_TypeChecker* self, ax_u32 node_idx) {
    ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
    if ((sym_idx != ((ax_u32)(0)))) {
        struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
        if ((sym.type_id == ((ax_u32)(0)))) {
            ax_u32 sum_node = ((ax_u32)(0));
            ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
            while ((child != ((ax_u32)(0)))) {
                struct ax_AstNode child_node = ((self->tree.nodes.data)[child]);
                if ((child_node.kind == ax_NODE_SUM_TYPE)) {
                    sum_node = child;
                    break;
                }
                child = child_node.next_sibling;
            }
            if ((sum_node != ((ax_u32)(0)))) {
                struct ax_VariantInfoVec variants = ax_new_variant_info_vec();
                ax_u8 tag = ((ax_u8)(0));
                ax_u32 v_node = ((self->tree.nodes.data)[sum_node]).first_child;
                while ((v_node != ((ax_u32)(0)))) {
                    struct ax_AstNode v_node_struct = ((self->tree.nodes.data)[v_node]);
                    if ((v_node_struct.kind == ax_NODE_VARIANT_DECL)) {
                        ax_u32 v_sym_idx = v_node_struct.payload;
                        if ((v_sym_idx != ((ax_u32)(0)))) {
                            struct ax_Symbol v_sym = ((self->symtable.symbols.data)[v_sym_idx]);
                            ax_u32 payload_type = ax_TYPE_UNKNOWN;
                            ax_u32 type_expr_node = v_node_struct.first_child;
                            if ((type_expr_node != ((ax_u32)(0)))) {
                                payload_type = ax_TypeChecker_infer_node(self, type_expr_node, ax_TYPE_UNKNOWN);
                            }
                            ax_VariantInfoVec_push(&(variants), ((struct ax_VariantInfo){.name_id=v_sym.name_id, .payload_type=payload_type, .tag=tag, .padding1=((ax_u8)(0)), .padding2=((ax_u16)(0))}));
                            tag = (tag + ((ax_u8)(1)));
                        }
                    }
                    v_node = v_node_struct.next_sibling;
                }
                ax_u32 type_id = ax_TypeTable_register_sum_type(&(self->types), sym.name_id, variants);
                ((self->symtable.symbols.data)[sym_idx]).type_id = type_id;
                ax_TypeChecker_set_node_type(self, node_idx, type_id);
                v_node = ((self->tree.nodes.data)[sum_node]).first_child;
                while ((v_node != ((ax_u32)(0)))) {
                    struct ax_AstNode v_node_struct = ((self->tree.nodes.data)[v_node]);
                    if ((v_node_struct.kind == ax_NODE_VARIANT_DECL)) {
                        ax_u32 v_sym_idx = v_node_struct.payload;
                        if ((v_sym_idx != ((ax_u32)(0)))) {
                            ((self->symtable.symbols.data)[v_sym_idx]).type_id = type_id;
                        }
                    }
                    v_node = v_node_struct.next_sibling;
                }
            }
        }
    }
}

void ax_TypeChecker_pre_infer_struct(struct ax_TypeChecker* self, ax_u32 node_idx) {
    ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
    if ((sym_idx != ((ax_u32)(0)))) {
        struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
        struct ax_StructFieldVec fields = ((struct ax_StructFieldVec){.data=((struct ax_StructField*)(NULL)), .len=0, .cap=0});
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        while ((child != ((ax_u32)(0)))) {
            struct ax_AstNode child_node = ((self->tree.nodes.data)[child]);
            if ((child_node.kind == ax_NODE_FIELD_DECL)) {
                ax_u32 f_sym_idx = child_node.payload;
                if ((f_sym_idx != ((ax_u32)(0)))) {
                    struct ax_Symbol f_sym = ((self->symtable.symbols.data)[f_sym_idx]);
                    ax_u32 f_type_node = child_node.first_child;
                    ax_u32 f_type_id = ax_TYPE_UNKNOWN;
                    if ((f_type_node != ((ax_u32)(0)))) {
                        f_type_id = ax_TypeChecker_infer_node(self, f_type_node, ax_TYPE_UNKNOWN);
                    }
                    ((self->symtable.symbols.data)[f_sym_idx]).type_id = f_type_id;
                    ax_StructFieldVec_push(&(fields), ((struct ax_StructField){.name_id=f_sym.name_id, .type_id=f_type_id}));
                }
            }
            child = child_node.next_sibling;
        }
        ax_u32 type_id = sym.type_id;
        if ((type_id == ((ax_u32)(0)))) {
            ax_u32 struct_type_id = ax_TypeTable_register_struct(&(self->types), sym.name_id, fields);
            ((self->symtable.symbols.data)[sym_idx]).type_id = struct_type_id;
            ax_TypeChecker_set_node_type(self, node_idx, struct_type_id);
        } else {
            {
                struct ax_TypeEntry entry = ((self->types.entries.data)[((ax_i64)(type_id))]);
                ((self->types.structs.data)[((ax_i64)(entry.extra))]) = ((struct ax_StructInfo){.fields=fields});
                ax_TypeChecker_set_node_type(self, node_idx, type_id);
            }
        }
    }
}

void ax_TypeChecker_pre_infer_func_signature(struct ax_TypeChecker* self, ax_u32 node_idx) {
    ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
    ax_bool has_existing = AX_FALSE;
    if ((sym_idx != ((ax_u32)(0)))) {
        struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
        if ((sym.type_id != ((ax_u32)(0)))) {
            struct ax_TypeEntry entry = ((self->types.entries.data)[sym.type_id]);
            if ((entry.kind == ax_TYPE_KIND_FUNC)) {
                has_existing = AX_TRUE;
            }
        }
    }
    if ((!has_existing)) {
        struct ax_U32Vec param_types = ax_new_u32_vec();
        ax_u32 ret_type = ax_TYPE_UNKNOWN;
        ax_bool has_ret_expr = AX_FALSE;
        struct ax_AstNode node = ((self->tree.nodes.data)[node_idx]);
        ax_u32 child = node.first_child;
        while ((child != ((ax_u32)(0)))) {
            struct ax_AstNode child_node = ((self->tree.nodes.data)[child]);
            if ((child_node.kind == ax_NODE_PARAM_DECL)) {
                ax_u32 p_type = ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
                ax_U32Vec_push(&(param_types), p_type);
            } else if (((child_node.kind == ax_NODE_TYPE_EXPR) || (child_node.kind == ax_NODE_GENERIC_TYPE))) {
                ret_type = ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
                has_ret_expr = AX_TRUE;
            }
            child = child_node.next_sibling;
        }
        if ((!has_ret_expr)) {
            ax_bool is_main = AX_FALSE;
            if ((sym_idx != ((ax_u32)(0)))) {
                ax_u32 name_id = ((self->symtable.symbols.data)[sym_idx]).name_id;
                ax_string text = ax_InternPool_get(self->intern, name_id);
                if (ax_str_eq(text, (ax_string){.ptr=(const ax_u8*)"main", .len=4})) {
                    is_main = AX_TRUE;
                }
            }
            if (is_main) {
                ret_type = ax_TYPE_I32;
            } else {
                {
                    ret_type = ax_TYPE_VOID;
                }
            }
        }
        ax_u32 func_type_id = ax_TypeTable_register_function(&(self->types), param_types, ret_type);
        if ((sym_idx != ((ax_u32)(0)))) {
            ((self->symtable.symbols.data)[sym_idx]).type_id = func_type_id;
        }
        ax_TypeChecker_set_node_type(self, node_idx, func_type_id);
    }
}

ax_u32 ax_TypeChecker_infer_node(struct ax_TypeChecker* self, ax_u32 node_idx, ax_u32 expected) {
    if ((node_idx == ((ax_u32)(0)))) {
        if ((self->tree.nodes.len == ((ax_i64)(0)))) {
            return ax_TYPE_UNKNOWN;
        }
        if ((((self->tree.nodes.data)[0]).kind == ((ax_u8)(0)))) {
            return ax_TYPE_UNKNOWN;
        }
    }
    ax_u8 kind = ((self->tree.nodes.data)[node_idx]).kind;
    ax_u16 flags = ((self->tree.nodes.data)[node_idx]).flags;
    ax_u32 result_type = ax_TYPE_UNKNOWN;
    if (((kind == ax_NODE_PROGRAM) || (kind == ax_NODE_BLOCK))) {
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        while ((child != ((ax_u32)(0)))) {
            ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
            child = ((self->tree.nodes.data)[child]).next_sibling;
        }
    } else if ((kind == ax_NODE_FUNC_DECL)) {
        ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 ret_type = ax_TYPE_UNKNOWN;
        if ((sym_idx != ((ax_u32)(0)))) {
            struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
            if ((sym.type_id != ((ax_u32)(0)))) {
                struct ax_TypeEntry entry = ((self->types.entries.data)[sym.type_id]);
                if ((entry.kind == ax_TYPE_KIND_FUNC)) {
                    struct ax_FuncInfo fi = ((self->types.funcs.data)[entry.extra]);
                    ret_type = fi.ret;
                    result_type = sym.type_id;
                }
            }
        }
        ax_u32 prev_ret = self->current_return;
        self->current_return = ret_type;
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        while ((child != ((ax_u32)(0)))) {
            struct ax_AstNode child_node = ((self->tree.nodes.data)[child]);
            if (((child_node.kind != ax_NODE_PARAM_DECL) && (child_node.kind != ax_NODE_TYPE_EXPR))) {
                ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
            }
            child = child_node.next_sibling;
        }
        self->current_return = prev_ret;
    } else if ((kind == ax_NODE_STRUCT_DECL)) {
        ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
        if ((sym_idx != ((ax_u32)(0)))) {
            result_type = ((self->symtable.symbols.data)[sym_idx]).type_id;
        }
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        while ((child != ((ax_u32)(0)))) {
            struct ax_AstNode child_node = ((self->tree.nodes.data)[child]);
            if ((child_node.kind != ax_NODE_FIELD_DECL)) {
                ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
            }
            child = child_node.next_sibling;
        }
    } else if ((kind == ax_NODE_TYPE_ALIAS_DECL)) {
        ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
        if ((sym_idx != ((ax_u32)(0)))) {
            result_type = ((self->symtable.symbols.data)[sym_idx]).type_id;
        }
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        while ((child != ((ax_u32)(0)))) {
            struct ax_AstNode child_node = ((self->tree.nodes.data)[child]);
            if ((child_node.kind != ax_NODE_SUM_TYPE)) {
                ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
            }
            child = child_node.next_sibling;
        }
    } else if ((kind == ax_NODE_PARAM_DECL)) {
        ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 p_type = ax_TYPE_UNKNOWN;
        ax_u32 type_node = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((type_node != ((ax_u32)(0)))) {
            p_type = ax_TypeChecker_infer_node(self, type_node, ax_TYPE_UNKNOWN);
        }
        if ((sym_idx != ((ax_u32)(0)))) {
            ((self->symtable.symbols.data)[sym_idx]).type_id = p_type;
        }
        result_type = p_type;
    } else if (((kind == ax_NODE_VAR_DECL) || (kind == ax_NODE_CONST_DECL))) {
        ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 exp_type = ax_TYPE_UNKNOWN;
        if ((sym_idx != ((ax_u32)(0)))) {
            exp_type = ((self->symtable.symbols.data)[sym_idx]).type_id;
        }
        ax_u32 type_node = ((ax_u32)(0));
        ax_u32 init_expr = ((ax_u32)(0));
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        while ((child != ((ax_u32)(0)))) {
            struct ax_AstNode child_node = ((self->tree.nodes.data)[child]);
            if (((child_node.kind == ax_NODE_TYPE_EXPR) || (child_node.kind == ax_NODE_GENERIC_TYPE))) {
                type_node = child;
            } else {
                {
                    init_expr = child;
                }
            }
            child = child_node.next_sibling;
        }
        if ((type_node != ((ax_u32)(0)))) {
            exp_type = ax_TypeChecker_infer_node(self, type_node, ax_TYPE_UNKNOWN);
            if ((sym_idx != ((ax_u32)(0)))) {
                ((self->symtable.symbols.data)[sym_idx]).type_id = exp_type;
            }
        }
        if ((init_expr != ((ax_u32)(0)))) {
            ax_u32 inferred = ax_TypeChecker_infer_node(self, init_expr, exp_type);
            if ((exp_type != ax_TYPE_UNKNOWN)) {
                result_type = exp_type;
            } else {
                {
                    result_type = inferred;
                    if ((sym_idx != ((ax_u32)(0)))) {
                        ((self->symtable.symbols.data)[sym_idx]).type_id = inferred;
                    }
                }
            }
        } else {
            {
                result_type = exp_type;
            }
        }
    } else if ((kind == ax_NODE_RETURN_STMT)) {
        ax_u32 expr = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((expr != ((ax_u32)(0)))) {
            ax_TypeChecker_infer_node(self, expr, self->current_return);
        }
    } else if ((kind == ax_NODE_MATCH_STMT)) {
        ax_u32 scrutinee = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((scrutinee != ((ax_u32)(0)))) {
            ax_u32 scrutinee_type = ax_TypeChecker_infer_node(self, scrutinee, ax_TYPE_UNKNOWN);
            ax_u32 prev_scrutinee = self->current_match_scrutinee;
            self->current_match_scrutinee = scrutinee_type;
            ax_u32 arm = ((self->tree.nodes.data)[scrutinee]).next_sibling;
            ax_u32 arm_type = ax_TYPE_UNKNOWN;
            while ((arm != ((ax_u32)(0)))) {
                ax_u32 t = ax_TypeChecker_infer_node(self, arm, expected);
                if ((arm_type == ax_TYPE_UNKNOWN)) {
                    arm_type = t;
                }
                arm = ((self->tree.nodes.data)[arm]).next_sibling;
            }
            self->current_match_scrutinee = prev_scrutinee;
            result_type = arm_type;
        }
    } else if ((kind == ax_NODE_MATCH_ARM)) {
        ax_u32 pattern = ((self->tree.nodes.data)[node_idx]).first_child;
        ax_u32 body = ((ax_u32)(0));
        if ((pattern != ((ax_u32)(0)))) {
            body = ((self->tree.nodes.data)[pattern]).next_sibling;
        }
        ax_u32 scrutinee_type = self->current_match_scrutinee;
        if ((scrutinee_type != ax_TYPE_UNKNOWN)) {
            struct ax_TypeEntry s_entry = ((self->types.entries.data)[scrutinee_type]);
            if ((s_entry.kind == ax_TYPE_KIND_SUM)) {
                struct ax_SumInfo sum_info = ((self->types.sumtypes.data)[s_entry.extra]);
                if ((pattern != ((ax_u32)(0)))) {
                    struct ax_AstNode pat_node = ((self->tree.nodes.data)[pattern]);
                    if ((pat_node.kind == ax_NODE_VARIANT_PAT)) {
                        ax_u32 v_sym_idx = pat_node.payload;
                        if ((v_sym_idx != ((ax_u32)(0)))) {
                            struct ax_Symbol v_sym = ((self->symtable.symbols.data)[v_sym_idx]);
                            ax_u32 payload_type = ax_TYPE_UNKNOWN;
                            ax_bool found = AX_FALSE;
                            ax_i64 j = ((ax_i64)(0));
                            while ((j < sum_info.variants.len)) {
                                struct ax_VariantInfo v_info = ((sum_info.variants.data)[j]);
                                if ((v_info.name_id == v_sym.name_id)) {
                                    payload_type = v_info.payload_type;
                                    found = AX_TRUE;
                                    break;
                                }
                                j = (j + 1);
                            }
                            ax_u32 arg = pat_node.first_child;
                            if (((arg != ((ax_u32)(0))) && (((self->tree.nodes.data)[arg]).kind == ax_NODE_BINDING_PAT))) {
                                ax_u32 arg_sym_idx = ((self->tree.nodes.data)[arg]).payload;
                                if ((arg_sym_idx != ((ax_u32)(0)))) {
                                    ((self->symtable.symbols.data)[arg_sym_idx]).type_id = payload_type;
                                    ax_TypeChecker_set_node_type(self, arg, payload_type);
                                }
                            }
                        }
                    } else if ((pat_node.kind == ax_NODE_BINDING_PAT)) {
                        ax_u32 b_sym_idx = pat_node.payload;
                        if ((b_sym_idx != ((ax_u32)(0)))) {
                            ((self->symtable.symbols.data)[b_sym_idx]).type_id = scrutinee_type;
                            ax_TypeChecker_set_node_type(self, pattern, scrutinee_type);
                        }
                    }
                }
            }
        }
        if ((body != ((ax_u32)(0)))) {
            result_type = ax_TypeChecker_infer_node(self, body, expected);
        }
    } else if ((kind == ax_NODE_IF_STMT)) {
        ax_u32 cond = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((cond != ((ax_u32)(0)))) {
            ax_TypeChecker_infer_node(self, cond, ax_TYPE_BOOL);
            ax_u32 then_branch = ((self->tree.nodes.data)[cond]).next_sibling;
            if ((then_branch != ((ax_u32)(0)))) {
                result_type = ax_TypeChecker_infer_node(self, then_branch, expected);
                ax_u32 sibling = ((self->tree.nodes.data)[then_branch]).next_sibling;
                while ((sibling != ((ax_u32)(0)))) {
                    ax_TypeChecker_infer_node(self, sibling, expected);
                    sibling = ((self->tree.nodes.data)[sibling]).next_sibling;
                }
            }
        }
    } else if ((kind == ax_NODE_ELIF_CLAUSE)) {
        ax_u32 cond = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((cond != ((ax_u32)(0)))) {
            ax_TypeChecker_infer_node(self, cond, ax_TYPE_BOOL);
            ax_u32 body = ((self->tree.nodes.data)[cond]).next_sibling;
            if ((body != ((ax_u32)(0)))) {
                result_type = ax_TypeChecker_infer_node(self, body, expected);
            }
        }
    } else if ((kind == ax_NODE_ELSE_CLAUSE)) {
        ax_u32 body = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((body != ((ax_u32)(0)))) {
            result_type = ax_TypeChecker_infer_node(self, body, expected);
        }
    } else if ((kind == ax_NODE_WHILE_STMT)) {
        ax_u32 cond = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((cond != ((ax_u32)(0)))) {
            ax_TypeChecker_infer_node(self, cond, ax_TYPE_BOOL);
            ax_u32 body = ((self->tree.nodes.data)[cond]).next_sibling;
            if ((body != ((ax_u32)(0)))) {
                ax_TypeChecker_infer_node(self, body, ax_TYPE_UNKNOWN);
            }
        }
    } else if ((kind == ax_NODE_FOR_STMT)) {
        ax_u32 iter_var = ((self->tree.nodes.data)[node_idx]).payload;
        ax_u32 range_expr = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((range_expr != ((ax_u32)(0)))) {
            ax_TypeChecker_infer_node(self, range_expr, ax_TYPE_UNKNOWN);
            ax_u32 body = ((self->tree.nodes.data)[range_expr]).next_sibling;
            if ((body != ((ax_u32)(0)))) {
                ax_TypeChecker_infer_node(self, body, ax_TYPE_UNKNOWN);
            }
        }
        if ((iter_var != ((ax_u32)(0)))) {
            ((self->symtable.symbols.data)[iter_var]).type_id = ax_TYPE_I32;
        }
    } else if ((kind == ax_NODE_BINARY_EXPR)) {
        ax_u32 lhs = ((self->tree.nodes.data)[node_idx]).first_child;
        ax_u32 rhs = ((ax_u32)(0));
        if ((lhs != ((ax_u32)(0)))) {
            rhs = ((self->tree.nodes.data)[lhs]).next_sibling;
        }
        if (((lhs != ((ax_u32)(0))) && (rhs != ((ax_u32)(0))))) {
            ax_u16 op = flags;
            if ((op == ((ax_u16)(1)))) {
                ax_TypeChecker_infer_node(self, lhs, ax_TYPE_UNKNOWN);
                ax_TypeChecker_infer_node(self, rhs, ax_TYPE_UNKNOWN);
                result_type = ax_TYPE_BOOL;
            } else if ((op == ((ax_u16)(2)))) {
                ax_TypeChecker_infer_node(self, lhs, ax_TYPE_BOOL);
                ax_TypeChecker_infer_node(self, rhs, ax_TYPE_BOOL);
                result_type = ax_TYPE_BOOL;
            } else {
                {
                    ax_u32 t1 = ax_TypeChecker_infer_node(self, lhs, ax_TYPE_UNKNOWN);
                    ax_u32 t2 = ax_TypeChecker_infer_node(self, rhs, ax_TYPE_UNKNOWN);
                    if (((t1 == ax_TYPE_F64) || (t2 == ax_TYPE_F64))) {
                        result_type = ax_TYPE_F64;
                    } else if (((t1 == ax_TYPE_F32) || (t2 == ax_TYPE_F32))) {
                        result_type = ax_TYPE_F32;
                    } else if (((t1 == ax_TYPE_I64) || (t2 == ax_TYPE_I64))) {
                        result_type = ax_TYPE_I64;
                    } else {
                        {
                            result_type = t1;
                        }
                    }
                }
            }
        }
    } else if ((kind == ax_NODE_CALL_EXPR)) {
        ax_u32 callee = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((callee != ((ax_u32)(0)))) {
            struct ax_AstNode callee_node = ((self->tree.nodes.data)[callee]);
            ax_bool is_generic_call = AX_FALSE;
            ax_u32 template_sym_idx = ((ax_u32)(0));
            if ((callee_node.kind == ax_NODE_IDENT)) {
                ax_u32 sym_idx = callee_node.payload;
                if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symtable.symbols.len))) {
                    struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
                    if (((sym.kind == ax_SYM_FUNC) && (((sym.flags & ax_SYM_FLAG_GENERIC) != ((ax_u16)(0))) || ((((self->tree.nodes.data)[sym.decl_node]).flags & ax_FLAG_IS_GENERIC) != ((ax_u16)(0)))))) {
                        is_generic_call = AX_TRUE;
                        template_sym_idx = sym_idx;
                    }
                }
            }
            if (is_generic_call) {
                ax_u32 decl_node = ((self->symtable.symbols.data)[template_sym_idx]).decl_node;
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[DEBUG-TC-GEN] is_generic_call=true, template_sym_idx=%d, decl_node=%d\n", .len=71}).ptr, ((ax_i64)(template_sym_idx)), ((ax_i64)(decl_node)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                struct ax_U32Vec gen_params = ax_new_u32_vec();
                ax_u32 c = ((self->tree.nodes.data)[decl_node]).first_child;
                while ((c != ((ax_u32)(0)))) {
                    struct ax_AstNode cn = ((self->tree.nodes.data)[c]);
                    if ((cn.kind == ax_NODE_GENERIC_PARAMS)) {
                        ax_u32 gp = cn.first_child;
                        while ((gp != ((ax_u32)(0)))) {
                            ax_U32Vec_push(&(gen_params), ((self->tree.nodes.data)[gp]).payload);
                            gp = ((self->tree.nodes.data)[gp]).next_sibling;
                        }
                        break;
                    }
                    c = cn.next_sibling;
                }
                struct ax_U32Vec arg_types = ax_new_u32_vec();
                ax_u32 arg_idx = ((self->tree.nodes.data)[callee]).next_sibling;
                while ((arg_idx != ((ax_u32)(0)))) {
                    ax_u32 a_type = ax_TypeChecker_infer_node(self, arg_idx, ax_TYPE_UNKNOWN);
                    ax_U32Vec_push(&(arg_types), a_type);
                    arg_idx = ((self->tree.nodes.data)[arg_idx]).next_sibling;
                }
                struct ax_U32Vec inferred = ax_new_u32_vec();
                ax_i64 i = ((ax_i64)(0));
                while ((i < gen_params.len)) {
                    ax_U32Vec_push(&(inferred), ax_TYPE_UNKNOWN);
                    i = (i + 1);
                }
                ax_i64 p_idx = ((ax_i64)(0));
                c = ((self->tree.nodes.data)[decl_node]).first_child;
                while ((c != ((ax_u32)(0)))) {
                    struct ax_AstNode cn = ((self->tree.nodes.data)[c]);
                    if ((cn.kind == ax_NODE_PARAM_DECL)) {
                        ax_u32 type_node_idx = cn.first_child;
                        if (((type_node_idx != ((ax_u32)(0))) && (p_idx < arg_types.len))) {
                            struct ax_AstNode tn = ((self->tree.nodes.data)[type_node_idx]);
                            if ((tn.kind == ax_NODE_TYPE_EXPR)) {
                                ax_u32 param_sym_idx = tn.payload;
                                if (((param_sym_idx != ((ax_u32)(0))) && (((ax_i64)(param_sym_idx)) < self->symtable.symbols.len))) {
                                    struct ax_Symbol param_sym = ((self->symtable.symbols.data)[param_sym_idx]);
                                    if ((param_sym.kind == ax_SYM_GENERIC_PARAM)) {
                                        ax_i64 g_idx = ((ax_i64)(0));
                                        while ((g_idx < gen_params.len)) {
                                            if ((((gen_params.data)[g_idx]) == param_sym.name_id)) {
                                                ((inferred.data)[g_idx]) = ((arg_types.data)[p_idx]);
                                                break;
                                            }
                                            g_idx = (g_idx + ((ax_i64)(1)));
                                        }
                                    }
                                }
                            }
                        }
                        p_idx = (p_idx + 1);
                    }
                    c = cn.next_sibling;
                }
                i = 0;
                while ((i < inferred.len)) {
                    if ((((inferred.data)[i]) == ax_TYPE_UNKNOWN)) {
                        ((inferred.data)[i]) = ax_TYPE_I32;
                    }
                    i = (i + 1);
                }
                struct ax_Monomorphizer mono = ax_AstTree_new_monomorphizer(self->tree, self->intern, self->symtable, self->types);
                ax_string mangled_name = ax_Monomorphizer_mangle_name(mono, ax_InternPool_get(self->intern, ((self->symtable.symbols.data)[template_sym_idx]).name_id), inferred);
                ax_u32 mangled_id = ax_InternPool_intern(&(self->intern), mangled_name);
                ax_u32 existing_sym_idx = ax_SymbolTable_resolve(self->symtable, mangled_id);
                ax_u32 inst_sym_idx = ((ax_u32)(0));
                if ((existing_sym_idx != ((ax_u32)(0)))) {
                    inst_sym_idx = existing_sym_idx;
                } else {
                    {
                        inst_sym_idx = ax_Monomorphizer_instantiate_function(&(mono), template_sym_idx, inferred);
                        struct ax_Symbol inst_sym = ((self->symtable.symbols.data)[((ax_i64)(inst_sym_idx))]);
                        ax_u32 cloned_root = inst_sym.decl_node;
                        struct ax_NameResolver resolver = ax_AstTree_new_name_resolver(self->tree, self->intern, self->symtable);
                        ax_NameResolver_resolve_node(&(resolver), cloned_root);
                        ax_TypeChecker_pre_infer_func_signature(self, cloned_root);
                        ax_TypeChecker_infer_node(self, cloned_root, ax_TYPE_UNKNOWN);
                    }
                }
                ((self->tree.nodes.data)[callee]).payload = inst_sym_idx;
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[DEBUG-TC-GEN] specialized call ok, inst_sym_idx=%d\n", .len=52}).ptr, ((ax_i64)(inst_sym_idx)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                if ((gen_params.data != ((ax_u32*)(NULL)))) {
                    ax_free(((ax_u8*)(gen_params.data)));
                }
                if ((arg_types.data != ((ax_u32*)(NULL)))) {
                    ax_free(((ax_u8*)(arg_types.data)));
                }
                if ((inferred.data != ((ax_u32*)(NULL)))) {
                    ax_free(((ax_u8*)(inferred.data)));
                }
            }
            ax_u32 callee_type = ax_TypeChecker_infer_node(self, callee, ax_TYPE_UNKNOWN);
            if ((callee_type != ax_TYPE_UNKNOWN)) {
                struct ax_TypeEntry entry = ((self->types.entries.data)[callee_type]);
                if ((entry.kind == ax_TYPE_KIND_FUNC)) {
                    struct ax_FuncInfo fi = ((self->types.funcs.data)[entry.extra]);
                    result_type = fi.ret;
                } else if ((entry.kind == ax_TYPE_KIND_STRUCT)) {
                    result_type = callee_type;
                } else if ((entry.kind == ax_TYPE_KIND_SUM)) {
                    result_type = callee_type;
                }
            } else {
                {
                    if ((callee_node.kind == ax_NODE_IDENT)) {
                        ax_u32 sym_idx = callee_node.payload;
                        if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symtable.symbols.len))) {
                            struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
                            ax_string name = ax_InternPool_get(self->intern, sym.name_id);
                            if (((((((ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall0", .len=8}) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall1", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall2", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall3", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall4", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall5", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall6", .len=8}))) {
                                result_type = ax_TYPE_I64;
                            }
                        }
                    }
                }
            }
            ax_u32 arg = ((self->tree.nodes.data)[callee]).next_sibling;
            while ((arg != ((ax_u32)(0)))) {
                ax_TypeChecker_infer_node(self, arg, ax_TYPE_UNKNOWN);
                arg = ((self->tree.nodes.data)[arg]).next_sibling;
            }
        }
    } else if ((kind == ax_NODE_IDENT)) {
        ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
        if ((sym_idx != ((ax_u32)(0)))) {
            result_type = ((self->symtable.symbols.data)[sym_idx]).type_id;
        }
    } else if ((kind == ax_NODE_TYPE_EXPR)) {
        ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
        if ((sym_idx != ((ax_u32)(0)))) {
            result_type = ((self->symtable.symbols.data)[sym_idx]).type_id;
        }
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((child != ((ax_u32)(0)))) {
            ax_u32 inner = ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
            if ((result_type == ax_TYPE_UNKNOWN)) {
                result_type = inner;
            }
        }
    } else if ((kind == ax_NODE_GENERIC_TYPE)) {
        ax_u32 base_idx = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((base_idx != ((ax_u32)(0)))) {
            struct ax_AstNode base_node = ((self->tree.nodes.data)[base_idx]);
            ax_u32 name_id = base_node.payload;
            if ((base_node.kind == ax_NODE_IDENT)) {
                ax_u32 sym_idx = base_node.payload;
                if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symtable.symbols.len))) {
                    name_id = ((self->symtable.symbols.data)[((ax_i64)(sym_idx))]).name_id;
                }
            }
            ax_string text = ax_InternPool_get(self->intern, name_id);
            if (ax_str_eq(text, (ax_string){.ptr=(const ax_u8*)"ptr", .len=3})) {
                ax_u32 arg_idx = base_node.next_sibling;
                if ((arg_idx != ((ax_u32)(0)))) {
                    ax_u32 inner = ax_TypeChecker_infer_node(self, arg_idx, ax_TYPE_UNKNOWN);
                    result_type = ax_TypeTable_register_pointer(&(self->types), inner);
                }
            } else if (ax_str_eq(text, (ax_string){.ptr=(const ax_u8*)"slice", .len=5})) {
                ax_u32 arg_idx = base_node.next_sibling;
                if ((arg_idx != ((ax_u32)(0)))) {
                    ax_u32 inner = ax_TypeChecker_infer_node(self, arg_idx, ax_TYPE_UNKNOWN);
                    result_type = ax_TypeTable_register_slice(&(self->types), inner);
                }
            } else {
                {
                    ax_u32 sym_idx = base_node.payload;
                    if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symtable.symbols.len))) {
                        struct ax_Symbol sym = ((self->symtable.symbols.data)[((ax_i64)(sym_idx))]);
                        if ((((sym.flags & ax_SYM_FLAG_GENERIC) != ((ax_u16)(0))) || ((((self->tree.nodes.data)[sym.decl_node]).flags & ax_FLAG_IS_GENERIC) != ((ax_u16)(0))))) {
                            struct ax_U32Vec args = ax_new_u32_vec();
                            ax_u32 arg_idx = base_node.next_sibling;
                            while ((arg_idx != ((ax_u32)(0)))) {
                                ax_u32 arg_type = ax_TypeChecker_infer_node(self, arg_idx, ax_TYPE_UNKNOWN);
                                ax_U32Vec_push(&(args), arg_type);
                                arg_idx = ((self->tree.nodes.data)[arg_idx]).next_sibling;
                            }
                            struct ax_Monomorphizer mono = ax_AstTree_new_monomorphizer(self->tree, self->intern, self->symtable, self->types);
                            ax_string mangled_name = ax_Monomorphizer_mangle_name(mono, ax_InternPool_get(self->intern, sym.name_id), args);
                            ax_u32 mangled_id = ax_InternPool_intern(&(self->intern), mangled_name);
                            ax_u32 existing_sym_idx = ax_SymbolTable_resolve(self->symtable, mangled_id);
                            ax_u32 inst_sym_idx = ((ax_u32)(0));
                            if ((existing_sym_idx != ((ax_u32)(0)))) {
                                inst_sym_idx = existing_sym_idx;
                            } else {
                                {
                                    inst_sym_idx = ax_Monomorphizer_instantiate_function(&(mono), sym_idx, args);
                                    struct ax_Symbol inst_sym = ((self->symtable.symbols.data)[((ax_i64)(inst_sym_idx))]);
                                    ax_u32 cloned_root = inst_sym.decl_node;
                                    struct ax_NameResolver resolver = ax_AstTree_new_name_resolver(self->tree, self->intern, self->symtable);
                                    ax_NameResolver_resolve_node(&(resolver), cloned_root);
                                    ax_u8 cloned_kind = ((self->tree.nodes.data)[cloned_root]).kind;
                                    if ((cloned_kind == ax_NODE_STRUCT_DECL)) {
                                        ax_TypeChecker_pre_infer_struct(self, cloned_root);
                                        ax_u32 m_child = ((self->tree.nodes.data)[cloned_root]).first_child;
                                        while ((m_child != ((ax_u32)(0)))) {
                                            struct ax_AstNode mc = ((self->tree.nodes.data)[m_child]);
                                            if ((mc.kind == ax_NODE_FUNC_DECL)) {
                                                ax_TypeChecker_pre_infer_func_signature(self, m_child);
                                            }
                                            m_child = mc.next_sibling;
                                        }
                                        ax_TypeChecker_infer_node(self, cloned_root, ax_TYPE_UNKNOWN);
                                    } else if ((cloned_kind == ax_NODE_FUNC_DECL)) {
                                        ax_TypeChecker_pre_infer_func_signature(self, cloned_root);
                                        ax_TypeChecker_infer_node(self, cloned_root, ax_TYPE_UNKNOWN);
                                    } else if ((cloned_kind == ax_NODE_TYPE_ALIAS_DECL)) {
                                        ax_TypeChecker_pre_infer_type_alias(self, cloned_root);
                                        ax_TypeChecker_infer_node(self, cloned_root, ax_TYPE_UNKNOWN);
                                    }
                                }
                            }
                            struct ax_Symbol inst_sym = ((self->symtable.symbols.data)[((ax_i64)(inst_sym_idx))]);
                            result_type = inst_sym.type_id;
                            if ((args.data != ((ax_u32*)(NULL)))) {
                                ax_free(((ax_u8*)(args.data)));
                            }
                        }
                    }
                }
            }
        }
    } else if ((kind == ax_NODE_UNARY_EXPR)) {
        ax_u32 operand = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((operand != ((ax_u32)(0)))) {
            ax_u32 operand_type = ax_TypeChecker_infer_node(self, operand, expected);
            ax_string op = ax_TypeChecker_node_text(*(self), node_idx);
            if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"&", .len=1})) {
                result_type = ax_TypeTable_register_pointer(&(self->types), operand_type);
            } else if ((ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"not", .len=3}) || ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"!", .len=1}))) {
                result_type = ax_TYPE_BOOL;
            } else {
                {
                    result_type = operand_type;
                }
            }
        }
    } else if ((kind == ax_NODE_DEREF_EXPR)) {
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((child != ((ax_u32)(0)))) {
            ax_u32 child_type = ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
            if ((child_type != ax_TYPE_UNKNOWN)) {
                struct ax_TypeEntry entry = ((self->types.entries.data)[child_type]);
                if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                    result_type = entry.extra;
                } else {
                    {
                        result_type = child_type;
                    }
                }
            }
        }
    } else if ((kind == ax_NODE_INT_LIT)) {
        if (((((((((((((expected == ax_TYPE_I8) || (expected == ax_TYPE_I16)) || (expected == ax_TYPE_I32)) || (expected == ax_TYPE_I64)) || (expected == ax_TYPE_U8)) || (expected == ax_TYPE_U16)) || (expected == ax_TYPE_U32)) || (expected == ax_TYPE_U64)) || (expected == ax_TYPE_F32)) || (expected == ax_TYPE_F64)) || (expected == ax_TYPE_ISIZE)) || (expected == ax_TYPE_USIZE))) {
            result_type = expected;
        } else {
            {
                result_type = ax_TYPE_I32;
            }
        }
    } else if ((kind == ax_NODE_FLOAT_LIT)) {
        result_type = ax_TYPE_F64;
    } else if ((kind == ax_NODE_STRING_LIT)) {
        result_type = ax_TYPE_STRING;
    } else if ((kind == ax_NODE_BOOL_LIT)) {
        result_type = ax_TYPE_BOOL;
    } else if ((kind == ax_NODE_CAST_EXPR)) {
        ax_u32 expr = ((self->tree.nodes.data)[node_idx]).first_child;
        ax_u32 target_type = ax_TYPE_UNKNOWN;
        if ((expr != ((ax_u32)(0)))) {
            ax_TypeChecker_infer_node(self, expr, ax_TYPE_UNKNOWN);
            ax_u32 target_node = ((self->tree.nodes.data)[expr]).next_sibling;
            if ((target_node != ((ax_u32)(0)))) {
                target_type = ax_TypeChecker_infer_node(self, target_node, ax_TYPE_UNKNOWN);
            }
        }
        if ((target_type != ax_TYPE_UNKNOWN)) {
            ((self->tree.nodes.data)[node_idx]).payload = target_type;
            result_type = target_type;
        } else {
            {
                result_type = ((self->tree.nodes.data)[node_idx]).payload;
            }
        }
    } else if ((kind == ax_NODE_FIELD_EXPR)) {
        ax_u32 obj = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((obj != ((ax_u32)(0)))) {
            ax_u32 obj_type = ax_TypeChecker_infer_node(self, obj, ax_TYPE_UNKNOWN);
            if ((obj_type != ax_TYPE_UNKNOWN)) {
                struct ax_TypeEntry entry = ((self->types.entries.data)[obj_type]);
                if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                    obj_type = entry.extra;
                    entry = ((self->types.entries.data)[obj_type]);
                }
                if ((entry.kind == ax_TYPE_KIND_STRUCT)) {
                    struct ax_StructInfo struct_info = ((self->types.structs.data)[entry.extra]);
                    ax_u32 field_name_id = ((self->tree.nodes.data)[node_idx]).payload;
                    ax_bool found = AX_FALSE;
                    ax_i64 j = ((ax_i64)(0));
                    while ((j < struct_info.fields.len)) {
                        struct ax_StructField f = ((struct_info.fields.data)[j]);
                        if ((f.name_id == field_name_id)) {
                            result_type = f.type_id;
                            ((self->tree.nodes.data)[node_idx]).extra_idx = ((ax_u32)(j));
                            found = AX_TRUE;
                            break;
                        }
                        j = (j + 1);
                    }
                    if ((!found)) {
                        ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
                        if ((sym_idx != ((ax_u32)(0)))) {
                            struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
                            result_type = sym.type_id;
                        }
                    }
                }
            } else {
                {
                    ax_u32 sym_idx = ((self->tree.nodes.data)[node_idx]).payload;
                    if ((sym_idx != ((ax_u32)(0)))) {
                        struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
                        result_type = sym.type_id;
                    }
                }
            }
        }
    } else if ((kind == ax_NODE_INDEX_EXPR)) {
        ax_u32 collection = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((collection != ((ax_u32)(0)))) {
            ax_u32 col_type = ax_TypeChecker_infer_node(self, collection, ax_TYPE_UNKNOWN);
            ax_u32 idx = ((self->tree.nodes.data)[collection]).next_sibling;
            if ((idx != ((ax_u32)(0)))) {
                ax_TypeChecker_infer_node(self, idx, ax_TYPE_UNKNOWN);
            }
            if ((col_type != ax_TYPE_UNKNOWN)) {
                struct ax_TypeEntry c_entry = ((self->types.entries.data)[col_type]);
                if ((((c_entry.kind == ax_TYPE_KIND_POINTER) || (c_entry.kind == ax_TYPE_KIND_SLICE)) || (c_entry.kind == ax_TYPE_KIND_ARRAY))) {
                    result_type = c_entry.extra;
                }
            }
        }
    } else if ((kind == ax_NODE_COMPTIME)) {
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        if ((child != ((ax_u32)(0)))) {
            ax_u32 child_type = ax_TypeChecker_infer_node(self, child, expected);
            if ((child_type == ax_TYPE_UNKNOWN)) {
                result_type = ax_TYPE_UNKNOWN;
            } else {
                {
                    struct ax_ComptimeValue val = ax_TypeChecker_eval_comptime(self, child);
                    ax_u8 lit_kind = ax_NODE_INT_LIT;
                    ax_string val_str = (ax_string){.ptr=(const ax_u8*)"", .len=0};
                    ax_u8 tok_kind = ax_TK_INT_LIT;
                    if ((val.kind == ax_TYPE_BOOL)) {
                        lit_kind = ax_NODE_BOOL_LIT;
                        if (val.bool_val) {
                            val_str = (ax_string){.ptr=(const ax_u8*)"true", .len=4};
                            tok_kind = ax_TK_TRUE;
                        } else {
                            {
                                val_str = (ax_string){.ptr=(const ax_u8*)"false", .len=5};
                                tok_kind = ax_TK_FALSE;
                            }
                        }
                    } else if ((val.kind == ax_TYPE_STRING)) {
                        lit_kind = ax_NODE_STRING_LIT;
                        val_str = ax_str_concat((ax_string){.ptr=(const ax_u8*)"\"", .len=1}, ax_str_concat(val.str_val, (ax_string){.ptr=(const ax_u8*)"\"", .len=1}));
                        tok_kind = ax_TK_STRING_LIT;
                    } else if ((val.kind == ax_TYPE_F64)) {
                        lit_kind = ax_NODE_FLOAT_LIT;
                        val_str = ax_format_float(val.float_val);
                        tok_kind = ax_TK_FLOAT_LIT;
                    } else {
                        {
                            lit_kind = ax_NODE_INT_LIT;
                            val_str = ax_format_int(val.int_val);
                            tok_kind = ax_TK_INT_LIT;
                        }
                    }
                    ax_u32 new_offset = ((ax_u32)(ax_str_len(self->tree.src)));
                    self->tree.src = ax_str_concat(self->tree.src, val_str);
                    ax_u32 new_tok_idx = ((ax_u32)(self->tree.tokens.len));
                    ax_TokenVec_push(&(self->tree.tokens), ((struct ax_Token){.kind=tok_kind, .padding=((ax_u8)(0)), .len=((ax_u16)(ax_str_len(val_str))), .offset=new_offset}));
                    ((self->tree.nodes.data)[node_idx]).kind = lit_kind;
                    ((self->tree.nodes.data)[node_idx]).token_idx = new_tok_idx;
                    ((self->tree.nodes.data)[node_idx]).first_child = ax_NULL_IDX;
                    ((self->tree.nodes.data)[node_idx]).payload = val.kind;
                    result_type = val.kind;
                }
            }
        } else {
            {
                result_type = ax_TYPE_VOID;
            }
        }
    } else if ((kind == ax_NODE_PTR_TYPE)) {
        ax_u32 inner_node = ((self->tree.nodes.data)[node_idx]).first_child;
        ax_u32 inner_type = ax_TYPE_UNKNOWN;
        if ((inner_node != ((ax_u32)(0)))) {
            inner_type = ax_TypeChecker_infer_node(self, inner_node, ax_TYPE_UNKNOWN);
        }
        result_type = ax_TypeTable_register_pointer(&(self->types), inner_type);
    } else if ((kind == ax_NODE_SLICE_TYPE)) {
        ax_u32 inner_node = ((self->tree.nodes.data)[node_idx]).first_child;
        ax_u32 inner_type = ax_TYPE_UNKNOWN;
        if ((inner_node != ((ax_u32)(0)))) {
            inner_type = ax_TypeChecker_infer_node(self, inner_node, ax_TYPE_UNKNOWN);
        }
        result_type = ax_TypeTable_register_slice(&(self->types), inner_type);
    } else if ((kind == ax_NODE_ARRAY_TYPE)) {
        ax_u32 inner_node = ((self->tree.nodes.data)[node_idx]).first_child;
        ax_u32 inner_type = ax_TYPE_UNKNOWN;
        if ((inner_node != ((ax_u32)(0)))) {
            inner_type = ax_TypeChecker_infer_node(self, inner_node, ax_TYPE_UNKNOWN);
        }
        result_type = ax_TypeTable_register_array(&(self->types), inner_type);
    } else if ((kind == ax_NODE_FUNC_TYPE)) {
        struct ax_U32Vec param_types = ax_new_u32_vec();
        ax_u32 ret_type = ax_TYPE_VOID;
        ax_bool has_ret = ((((self->tree.nodes.data)[node_idx]).flags & ((ax_u16)(1))) != ((ax_u16)(0)));
        ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
        while ((child != ((ax_u32)(0)))) {
            struct ax_AstNode child_node = ((self->tree.nodes.data)[child]);
            ax_u32 next = child_node.next_sibling;
            if ((has_ret && (next == ((ax_u32)(0))))) {
                ret_type = ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
            } else {
                {
                    ax_u32 p_type = ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
                    ax_U32Vec_push(&(param_types), p_type);
                }
            }
            child = next;
        }
        result_type = ax_TypeTable_register_function(&(self->types), param_types, ret_type);
    } else {
        {
            ax_u32 child = ((self->tree.nodes.data)[node_idx]).first_child;
            while ((child != ((ax_u32)(0)))) {
                ax_TypeChecker_infer_node(self, child, ax_TYPE_UNKNOWN);
                child = ((self->tree.nodes.data)[child]).next_sibling;
            }
        }
    }
    ax_TypeChecker_set_node_type(self, node_idx, result_type);
    return result_type;
}

struct ax_CGNodeVec ax_new_cg_node_vec(void) {
    return ((struct ax_CGNodeVec){.data=((struct ax_CGNode*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_CGNodeVec_push(struct ax_CGNodeVec* self, struct ax_CGNode n) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_CGNode* new_data = ((struct ax_CGNode*)(ax_alloc((new_cap * 20))));
        if ((self->data != ((struct ax_CGNode*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 20));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = n;
    self->len = (self->len + 1);
    return idx;
}

struct ax_CGEdgeVec ax_new_cg_edge_vec(void) {
    return ((struct ax_CGEdgeVec){.data=((struct ax_CGEdge*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_CGEdgeVec_push(struct ax_CGEdgeVec* self, struct ax_CGEdge e) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_CGEdge* new_data = ((struct ax_CGEdge*)(ax_alloc((new_cap * 12))));
        if ((self->data != ((struct ax_CGEdge*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 12));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = e;
    self->len = (self->len + 1);
    return idx;
}

struct ax_U32VecVec ax_new_u32_vec_vec(void) {
    return ((struct ax_U32VecVec){.data=((struct ax_U32Vec*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_U32VecVec_push(struct ax_U32VecVec* self, struct ax_U32Vec v) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_U32Vec* new_data = ((struct ax_U32Vec*)(ax_alloc((new_cap * 24))));
        if ((self->data != ((struct ax_U32Vec*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 24));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = v;
    self->len = (self->len + 1);
    return idx;
}

struct ax_ConnectionGraph ax_new_connection_graph(void) {
    return ((struct ax_ConnectionGraph){.nodes=ax_new_cg_node_vec(), .edges=ax_new_cg_edge_vec(), .adj_out=ax_new_u32_vec_vec(), .adj_in=ax_new_u32_vec_vec(), .sym_to_node=ax_new_u32_vec()});
}

void ax_ConnectionGraph_free_connection_graph(struct ax_ConnectionGraph* self) {
    if ((self->nodes.data != ((struct ax_CGNode*)(NULL)))) {
        ax_free(((ax_u8*)(self->nodes.data)));
    }
    if ((self->edges.data != ((struct ax_CGEdge*)(NULL)))) {
        ax_free(((ax_u8*)(self->edges.data)));
    }
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->adj_out.len)) {
        struct ax_U32Vec v = ((self->adj_out.data)[i]);
        if ((v.data != ((ax_u32*)(NULL)))) {
            ax_free(((ax_u8*)(v.data)));
        }
        i = (i + 1);
    }
    if ((self->adj_out.data != ((struct ax_U32Vec*)(NULL)))) {
        ax_free(((ax_u8*)(self->adj_out.data)));
    }
    i = 0;
    while ((i < self->adj_in.len)) {
        struct ax_U32Vec v = ((self->adj_in.data)[i]);
        if ((v.data != ((ax_u32*)(NULL)))) {
            ax_free(((ax_u8*)(v.data)));
        }
        i = (i + 1);
    }
    if ((self->adj_in.data != ((struct ax_U32Vec*)(NULL)))) {
        ax_free(((ax_u8*)(self->adj_in.data)));
    }
    if ((self->sym_to_node.data != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(self->sym_to_node.data)));
    }
}

void ax_ConnectionGraph_ensure_adj_capacity(struct ax_ConnectionGraph* self, ax_u32 node_id) {
    while ((self->adj_out.len <= ((ax_i64)(node_id)))) {
        ax_U32VecVec_push(&(self->adj_out), ax_new_u32_vec());
    }
    while ((self->adj_in.len <= ((ax_i64)(node_id)))) {
        ax_U32VecVec_push(&(self->adj_in), ax_new_u32_vec());
    }
}

ax_u32 ax_ConnectionGraph_add_value_node(struct ax_ConnectionGraph* self, ax_u32 sym_id, ax_u32 type_id, ax_u32 lifetime) {
    ax_u32 id = ((ax_u32)(self->nodes.len));
    ax_CGNodeVec_push(&(self->nodes), ((struct ax_CGNode){.id=id, .sym_id=sym_id, .type_id=type_id, .is_ref=AX_FALSE, .lifetime=lifetime}));
    ax_ConnectionGraph_ensure_adj_capacity(self, id);
    if ((sym_id != 0)) {
        while ((self->sym_to_node.len <= ((ax_i64)(sym_id)))) {
            ax_U32Vec_push(&(self->sym_to_node), ((ax_u32)(0xffffffff)));
        }
        ((self->sym_to_node.data)[sym_id]) = id;
    }
    return id;
}

ax_u32 ax_ConnectionGraph_add_ref_node(struct ax_ConnectionGraph* self, ax_u32 target_node_id) {
    ax_u32 type_id = ((ax_u32)(0));
    ax_u32 lifetime = ((ax_u32)(0));
    if ((((ax_i64)(target_node_id)) < self->nodes.len)) {
        type_id = ((self->nodes.data)[target_node_id]).type_id;
        lifetime = ((self->nodes.data)[target_node_id]).lifetime;
    }
    ax_u32 id = ((ax_u32)(self->nodes.len));
    ax_CGNodeVec_push(&(self->nodes), ((struct ax_CGNode){.id=id, .sym_id=((ax_u32)(0)), .type_id=type_id, .is_ref=AX_TRUE, .lifetime=lifetime}));
    ax_ConnectionGraph_ensure_adj_capacity(self, id);
    ax_ConnectionGraph_add_edge(self, id, target_node_id, ax_EDGE_BORROWS);
    return id;
}

void ax_ConnectionGraph_add_edge(struct ax_ConnectionGraph* self, ax_u32 from, ax_u32 to, ax_u32 kind) {
    ax_u32 edge_idx = ax_CGEdgeVec_push(&(self->edges), ((struct ax_CGEdge){.from=from, .to=to, .kind=kind}));
    ax_ConnectionGraph_ensure_adj_capacity(self, from);
    ax_ConnectionGraph_ensure_adj_capacity(self, to);
    ax_U32Vec_push(&(((self->adj_out.data)[from])), edge_idx);
    ax_U32Vec_push(&(((self->adj_in.data)[to])), edge_idx);
}

ax_u32 ax_ConnectionGraph_node_of_sym(struct ax_ConnectionGraph self, ax_u32 sym_id) {
    if ((((ax_i64)(sym_id)) < self.sym_to_node.len)) {
        return ((self.sym_to_node.data)[sym_id]);
    }
    return ((ax_u32)(0xffffffff));
}

ax_bool ax_ConnectionGraph_escapes(struct ax_ConnectionGraph self, ax_u32 node_id) {
    if ((self.nodes.len == 0)) {
        return AX_FALSE;
    }
    ax_bool* visited = ((ax_bool*)(ax_alloc(self.nodes.len)));
    ax_i64 idx = ((ax_i64)(0));
    while ((idx < self.nodes.len)) {
        ((visited)[idx]) = AX_FALSE;
        idx = (idx + 1);
    }
    ax_bool res = ax_ConnectionGraph_escape_dfs(self, node_id, visited);
    ax_free(((ax_u8*)(visited)));
    return res;
}

ax_bool ax_ConnectionGraph_escape_dfs(struct ax_ConnectionGraph self, ax_u32 node_id, ax_bool* visited) {
    if ((((ax_i64)(node_id)) >= self.nodes.len)) {
        return AX_FALSE;
    }
    if (((visited)[node_id])) {
        return AX_FALSE;
    }
    ((visited)[node_id]) = AX_TRUE;
    if ((((ax_i64)(node_id)) >= self.adj_out.len)) {
        return AX_FALSE;
    }
    struct ax_U32Vec* out_edges = &(((self.adj_out.data)[node_id]));
    ax_i64 i = ((ax_i64)(0));
    while ((i < out_edges->len)) {
        ax_u32 edge_idx = ((out_edges->data)[i]);
        struct ax_CGEdge edge = ((self.edges.data)[edge_idx]);
        if ((edge.kind == ax_EDGE_ESCAPES_TO)) {
            return AX_TRUE;
        }
        if (((edge.kind == ax_EDGE_OWNS) || (edge.kind == ax_EDGE_FLOWS_TO))) {
            if (ax_ConnectionGraph_escape_dfs(self, edge.to, visited)) {
                return AX_TRUE;
            }
        }
        i = (i + 1);
    }
    return AX_FALSE;
}

struct ax_OwnershipChecker ax_AstTree_new_ownership_checker(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable) {
    return ((struct ax_OwnershipChecker){.tree=tree, .pool=pool, .symtable=symtable, .typetable=typetable, .errors=((ax_u32)(0))});
}

void ax_OwnershipChecker_check(struct ax_OwnershipChecker* self) {
    ax_OwnershipChecker_check_node(self, ((ax_u32)(0)));
    if ((self->errors > ((ax_u32)(0)))) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"Ownership check failed.", .len=23}).ptr);
        exit(1);
    }
}

void ax_OwnershipChecker_check_node(struct ax_OwnershipChecker* self, ax_u32 node_idx) {
    if (((node_idx == ((ax_u32)(0xffffffff))) || (node_idx == ((ax_u32)(0))))) {
        return;
    }
    struct ax_AstNode node = ((self->tree.nodes.data)[node_idx]);
    ax_u8 kind = node.kind;
    if ((kind == ax_NODE_VAR_DECL)) {
        ax_u32 init_idx = node.first_child;
        if ((init_idx != ((ax_u32)(0)))) {
            ax_OwnershipChecker_check_node(self, init_idx);
            ax_OwnershipChecker_check_move(self, init_idx);
        }
    } else if ((kind == ax_NODE_ASSIGN_STMT)) {
        ax_u32 lhs_idx = node.first_child;
        ax_u32 rhs_idx = ((self->tree.nodes.data)[lhs_idx]).next_sibling;
        ax_OwnershipChecker_check_node(self, lhs_idx);
        ax_OwnershipChecker_check_node(self, rhs_idx);
        if ((lhs_idx != ((ax_u32)(0)))) {
            ax_u8 lhs_kind = ((self->tree.nodes.data)[lhs_idx]).kind;
            if ((lhs_kind == ax_NODE_IDENT)) {
                ax_u32 sym_idx = ((self->tree.nodes.data)[lhs_idx]).payload;
                if ((sym_idx != ((ax_u32)(0)))) {
                    struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
                    if (((sym.flags & ax_SYM_FLAG_MUT) == ((ax_u16)(0)))) {
                        struct ax_Token tok = ((self->tree.tokens.data)[node.token_idx]);
                        printf((const char*)((ax_string){.ptr=(const ax_u8*)"error[E4002]: cannot assign to immutable variable at offset %d\n", .len=63}).ptr, ((ax_i64)(tok.offset)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                        self->errors = (self->errors + ((ax_u32)(1)));
                    }
                }
            }
            ax_OwnershipChecker_check_move(self, rhs_idx);
        }
    } else if ((kind == ax_NODE_IDENT)) {
        ax_u32 sym_idx = node.payload;
        if ((sym_idx != ((ax_u32)(0)))) {
            struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
            if (((sym.flags & ax_SYM_FLAG_MOVED) != ((ax_u16)(0)))) {
                struct ax_Token tok = ((self->tree.tokens.data)[node.token_idx]);
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"error[E4001]: use of moved value at offset %d\n", .len=46}).ptr, ((ax_i64)(tok.offset)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                self->errors = (self->errors + ((ax_u32)(1)));
            }
        }
    } else if ((kind == ax_NODE_CALL_EXPR)) {
        ax_u32 arg_idx = node.first_child;
        if ((arg_idx != ((ax_u32)(0)))) {
            ax_OwnershipChecker_check_node(self, arg_idx);
            arg_idx = ((self->tree.nodes.data)[arg_idx]).next_sibling;
            while ((arg_idx != ((ax_u32)(0)))) {
                ax_OwnershipChecker_check_node(self, arg_idx);
                ax_OwnershipChecker_check_move(self, arg_idx);
                arg_idx = ((self->tree.nodes.data)[arg_idx]).next_sibling;
            }
        }
    } else {
        {
            ax_u32 child_idx = node.first_child;
            while ((child_idx != ((ax_u32)(0)))) {
                ax_OwnershipChecker_check_node(self, child_idx);
                child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
            }
        }
    }
}

void ax_OwnershipChecker_check_move(struct ax_OwnershipChecker* self, ax_u32 node_idx) {
    if (((node_idx == ((ax_u32)(0))) || (node_idx == ((ax_u32)(0xffffffff))))) {
        return;
    }
    struct ax_AstNode node = ((self->tree.nodes.data)[node_idx]);
    if ((node.kind == ax_NODE_IDENT)) {
        ax_u32 sym_idx = node.payload;
        if ((sym_idx != ((ax_u32)(0)))) {
            struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
            ax_bool is_copyable = AX_TRUE;
            ax_u32 type_id = sym.type_id;
            if (((type_id != ((ax_u32)(0))) && (((ax_i64)(type_id)) < self->typetable.entries.len))) {
                struct ax_TypeEntry entry = ((self->typetable.entries.data)[type_id]);
                if ((((entry.kind == ax_TYPE_KIND_STRUCT) || (entry.kind == ax_TYPE_KIND_SUM)) || (entry.kind == ax_TYPE_KIND_POINTER))) {
                    is_copyable = AX_FALSE;
                }
            }
            if ((!is_copyable)) {
                ((self->symtable.symbols.data)[sym_idx]).flags = (((self->symtable.symbols.data)[sym_idx]).flags | ax_SYM_FLAG_MOVED);
            }
        }
    }
}

struct ax_EscapeAnalyser ax_AstTree_new_escape_analyser(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable) {
    return ((struct ax_EscapeAnalyser){.tree=tree, .pool=pool, .symtable=symtable, .typetable=typetable, .curr_cg=ax_new_connection_graph(), .escape_node_idx=((ax_u32)(0))});
}

void ax_EscapeAnalyser_run(struct ax_EscapeAnalyser* self) {
    self->curr_cg = ax_new_connection_graph();
    self->escape_node_idx = ((ax_u32)(0));
    ax_EscapeAnalyser_traverse_nodes(self, ((ax_u32)(0)));
    ax_ConnectionGraph_free_connection_graph(&(self->curr_cg));
}

void ax_EscapeAnalyser_traverse_nodes(struct ax_EscapeAnalyser* self, ax_u32 node_idx) {
    if (((node_idx == ((ax_u32)(0xffffffff))) || (node_idx == ((ax_u32)(0))))) {
        return;
    }
    struct ax_AstNode node = ((self->tree.nodes.data)[node_idx]);
    ax_u8 kind = node.kind;
    if ((kind == ax_NODE_FUNC_DECL)) {
        struct ax_ConnectionGraph old_cg = self->curr_cg;
        self->curr_cg = ax_new_connection_graph();
        self->escape_node_idx = ax_ConnectionGraph_add_value_node(&(self->curr_cg), ((ax_u32)(0)), ((ax_u32)(0)), ((ax_u32)(0)));
        ax_u32 child_idx = node.first_child;
        while ((child_idx != ((ax_u32)(0)))) {
            struct ax_AstNode child = ((self->tree.nodes.data)[child_idx]);
            if ((child.kind == ax_NODE_PARAM_DECL)) {
                ax_u32 sym_idx = child.payload;
                if ((sym_idx != ((ax_u32)(0)))) {
                    struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
                    ax_ConnectionGraph_add_value_node(&(self->curr_cg), sym_idx, sym.type_id, ((ax_u32)(0)));
                }
            }
            child_idx = child.next_sibling;
        }
        child_idx = node.first_child;
        while ((child_idx != ((ax_u32)(0)))) {
            struct ax_AstNode child = ((self->tree.nodes.data)[child_idx]);
            if ((child.kind == ax_NODE_BLOCK)) {
                ax_EscapeAnalyser_analyze_block(self, child_idx);
            }
            child_idx = child.next_sibling;
        }
        ax_i64 idx = ((ax_i64)(0));
        while ((idx < self->curr_cg.nodes.len)) {
            struct ax_CGNode cg_node = ((self->curr_cg.nodes.data)[idx]);
            if ((cg_node.sym_id != ((ax_u32)(0)))) {
                ax_u32 sym_idx = cg_node.sym_id;
                ax_bool is_escaping = ax_ConnectionGraph_escapes(self->curr_cg, cg_node.id);
                if (is_escaping) {
                    ((self->symtable.symbols.data)[sym_idx]).flags = (((self->symtable.symbols.data)[sym_idx]).flags | ((ax_u16)(ax_FLAG_ESCAPES_TO_HEAP)));
                    ax_u32 decl_node = ((self->symtable.symbols.data)[sym_idx]).decl_node;
                    if ((decl_node != ((ax_u32)(0)))) {
                        ((self->tree.nodes.data)[decl_node]).flags = (((self->tree.nodes.data)[decl_node]).flags | ((ax_u16)(ax_FLAG_ESCAPES_TO_HEAP)));
                    }
                }
            }
            idx = (idx + 1);
        }
        ax_ConnectionGraph_free_connection_graph(&(self->curr_cg));
        self->curr_cg = old_cg;
    } else {
        {
            ax_u32 child_idx = node.first_child;
            while ((child_idx != ((ax_u32)(0)))) {
                ax_EscapeAnalyser_traverse_nodes(self, child_idx);
                child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
            }
        }
    }
}

void ax_EscapeAnalyser_analyze_block(struct ax_EscapeAnalyser* self, ax_u32 block_idx) {
    if (((block_idx == ((ax_u32)(0))) || (block_idx == ((ax_u32)(0xffffffff))))) {
        return;
    }
    struct ax_AstNode node = ((self->tree.nodes.data)[block_idx]);
    ax_u32 child_idx = node.first_child;
    while ((child_idx != ((ax_u32)(0)))) {
        ax_EscapeAnalyser_analyze_stmt(self, child_idx);
        child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
    }
}

void ax_EscapeAnalyser_analyze_stmt(struct ax_EscapeAnalyser* self, ax_u32 stmt_idx) {
    if (((stmt_idx == ((ax_u32)(0))) || (stmt_idx == ((ax_u32)(0xffffffff))))) {
        return;
    }
    struct ax_AstNode node = ((self->tree.nodes.data)[stmt_idx]);
    ax_u8 kind = node.kind;
    if ((kind == ax_NODE_VAR_DECL)) {
        ax_u32 sym_idx = node.payload;
        if ((sym_idx != ((ax_u32)(0)))) {
            struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
            ax_u32 cg_idx = ax_ConnectionGraph_add_value_node(&(self->curr_cg), sym_idx, sym.type_id, ((ax_u32)(0)));
            ax_u32 init_idx = node.first_child;
            if ((init_idx != ((ax_u32)(0)))) {
                ax_EscapeAnalyser_analyze_expr(self, init_idx, cg_idx);
            }
        }
    } else if ((kind == ax_NODE_ASSIGN_STMT)) {
        ax_u32 lhs_idx = node.first_child;
        ax_u32 rhs_idx = ((self->tree.nodes.data)[lhs_idx]).next_sibling;
        if ((lhs_idx != ((ax_u32)(0)))) {
            ax_u8 lhs_kind = ((self->tree.nodes.data)[lhs_idx]).kind;
            if ((lhs_kind == ax_NODE_IDENT)) {
                ax_u32 sym_idx = ((self->tree.nodes.data)[lhs_idx]).payload;
                if ((sym_idx != ((ax_u32)(0)))) {
                    ax_u32 lhs_cg_idx = ax_ConnectionGraph_node_of_sym(self->curr_cg, sym_idx);
                    if ((lhs_cg_idx != ((ax_u32)(0xffffffff)))) {
                        ax_EscapeAnalyser_analyze_expr(self, rhs_idx, lhs_cg_idx);
                    }
                }
            } else {
                {
                    ax_EscapeAnalyser_analyze_expr(self, rhs_idx, self->escape_node_idx);
                }
            }
        }
    } else if ((kind == ax_NODE_RETURN_STMT)) {
        ax_u32 val_idx = node.first_child;
        if ((val_idx != ((ax_u32)(0)))) {
            ax_EscapeAnalyser_analyze_expr(self, val_idx, self->escape_node_idx);
        }
    } else if ((kind == ax_NODE_IF_STMT)) {
        ax_u32 child_idx = node.first_child;
        while ((child_idx != ((ax_u32)(0)))) {
            struct ax_AstNode child = ((self->tree.nodes.data)[child_idx]);
            if ((child.kind == ax_NODE_BLOCK)) {
                ax_EscapeAnalyser_analyze_block(self, child_idx);
            } else {
                {
                    ax_EscapeAnalyser_analyze_stmt(self, child_idx);
                }
            }
            child_idx = child.next_sibling;
        }
    } else if (((kind == ax_NODE_WHILE_STMT) || (kind == ax_NODE_FOR_STMT))) {
        ax_u32 child_idx = node.first_child;
        while ((child_idx != ((ax_u32)(0)))) {
            struct ax_AstNode child = ((self->tree.nodes.data)[child_idx]);
            if ((child.kind == ax_NODE_BLOCK)) {
                ax_EscapeAnalyser_analyze_block(self, child_idx);
            }
            child_idx = child.next_sibling;
        }
    } else {
        {
            ax_u32 child_idx = node.first_child;
            while ((child_idx != ((ax_u32)(0)))) {
                ax_EscapeAnalyser_analyze_stmt(self, child_idx);
                child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
            }
        }
    }
}

void ax_EscapeAnalyser_analyze_expr(struct ax_EscapeAnalyser* self, ax_u32 expr_idx, ax_u32 flow_dest) {
    if (((expr_idx == ((ax_u32)(0))) || (expr_idx == ((ax_u32)(0xffffffff))))) {
        return;
    }
    struct ax_AstNode node = ((self->tree.nodes.data)[expr_idx]);
    ax_u8 kind = node.kind;
    if ((kind == ax_NODE_IDENT)) {
        ax_u32 sym_idx = node.payload;
        if ((sym_idx != ((ax_u32)(0)))) {
            ax_u32 cg_src = ax_ConnectionGraph_node_of_sym(self->curr_cg, sym_idx);
            if (((cg_src != ((ax_u32)(0xffffffff))) && (flow_dest != ((ax_u32)(0xffffffff))))) {
                ax_ConnectionGraph_add_edge(&(self->curr_cg), cg_src, flow_dest, ax_EDGE_FLOWS_TO);
            }
        }
    } else if ((kind == ax_NODE_CALL_EXPR)) {
        ax_u32 child_idx = node.first_child;
        if ((child_idx != ((ax_u32)(0)))) {
            ax_EscapeAnalyser_analyze_expr(self, child_idx, self->escape_node_idx);
            child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
            while ((child_idx != ((ax_u32)(0)))) {
                ax_EscapeAnalyser_analyze_expr(self, child_idx, self->escape_node_idx);
                child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
            }
        }
    } else if ((kind == ax_NODE_STRUCT_LIT)) {
        ax_u32 child_idx = node.first_child;
        while ((child_idx != ((ax_u32)(0)))) {
            struct ax_AstNode child = ((self->tree.nodes.data)[child_idx]);
            ax_u32 val_idx = child.first_child;
            if ((val_idx != ((ax_u32)(0)))) {
                ax_EscapeAnalyser_analyze_expr(self, val_idx, flow_dest);
            }
            child_idx = child.next_sibling;
        }
    } else {
        {
            ax_u32 child_idx = node.first_child;
            while ((child_idx != ((ax_u32)(0)))) {
                ax_EscapeAnalyser_analyze_expr(self, child_idx, flow_dest);
                child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
            }
        }
    }
}

struct ax_CtgcInjector ax_AstTree_new_ctgc_injector(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable) {
    return ((struct ax_CtgcInjector){.tree=tree, .pool=pool, .symtable=symtable, .typetable=typetable, .active_vars=ax_new_u32_vec(), .active_var_decls=ax_new_u32_vec()});
}

void ax_CtgcInjector_run(struct ax_CtgcInjector* self) {
    ax_CtgcInjector_traverse_and_inject(self, ((ax_u32)(0)), ((ax_u32)(0)));
}

void ax_CtgcInjector_traverse_and_inject(struct ax_CtgcInjector* self, ax_u32 node_idx, ax_u32 parent_idx) {
    if (((node_idx == ((ax_u32)(0xffffffff))) || (node_idx == ((ax_u32)(0))))) {
        return;
    }
    struct ax_AstNode node = ((self->tree.nodes.data)[node_idx]);
    ax_u8 kind = node.kind;
    if ((kind == ax_NODE_BLOCK)) {
        ax_i64 start_len = self->active_vars.len;
        ax_u32 child_idx = node.first_child;
        while ((child_idx != ((ax_u32)(0)))) {
            struct ax_AstNode child = ((self->tree.nodes.data)[child_idx]);
            if ((child.kind == ax_NODE_VAR_DECL)) {
                ax_u32 sym_idx = child.payload;
                if ((sym_idx != ((ax_u32)(0)))) {
                    struct ax_Symbol sym = ((self->symtable.symbols.data)[sym_idx]);
                    if (((sym.flags & ((ax_u16)(ax_FLAG_ESCAPES_TO_HEAP))) == ((ax_u16)(0)))) {
                        ax_bool is_heap = AX_FALSE;
                        ax_u32 type_id = sym.type_id;
                        if (((type_id != ((ax_u32)(0))) && (((ax_i64)(type_id)) < self->typetable.entries.len))) {
                            struct ax_TypeEntry entry = ((self->typetable.entries.data)[type_id]);
                            if ((((entry.kind == ax_TYPE_KIND_STRUCT) || (entry.kind == ax_TYPE_KIND_SUM)) || (entry.kind == ax_TYPE_KIND_POINTER))) {
                                is_heap = AX_TRUE;
                            }
                        }
                        if (is_heap) {
                            ax_U32Vec_push(&(self->active_vars), sym_idx);
                            ax_U32Vec_push(&(self->active_var_decls), child_idx);
                        }
                    }
                }
            }
            child_idx = child.next_sibling;
        }
        child_idx = node.first_child;
        while ((child_idx != ((ax_u32)(0)))) {
            ax_CtgcInjector_traverse_and_inject(self, child_idx, node_idx);
            child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
        }
        ax_i64 num_block_vars = (self->active_vars.len - start_len);
        if ((num_block_vars > ((ax_i64)(0)))) {
            ax_i64 idx = (self->active_vars.len - ((ax_i64)(1)));
            while ((idx >= start_len)) {
                ax_u32 sym_idx = ((self->active_vars.data)[idx]);
                ax_u32 destroy_node = ax_AstTree_add_node(&(self->tree), ax_NODE_DESTROY_STMT, node.token_idx);
                ((self->tree.nodes.data)[destroy_node]).payload = sym_idx;
                ax_CtgcInjector_append_child_node(self, node_idx, destroy_node);
                idx = (idx - 1);
            }
            self->active_vars.len = start_len;
            self->active_var_decls.len = start_len;
        }
    } else if ((kind == ax_NODE_RETURN_STMT)) {
        if ((self->active_vars.len > ((ax_i64)(0)))) {
            ax_i64 idx = (self->active_vars.len - ((ax_i64)(1)));
            while ((idx >= ((ax_i64)(0)))) {
                ax_u32 sym_idx = ((self->active_vars.data)[idx]);
                ax_u32 destroy_node = ax_AstTree_add_node(&(self->tree), ax_NODE_DESTROY_STMT, node.token_idx);
                ((self->tree.nodes.data)[destroy_node]).payload = sym_idx;
                ax_CtgcInjector_insert_before(self, parent_idx, node_idx, destroy_node);
                idx = (idx - 1);
            }
        }
    } else {
        {
            ax_u32 child_idx = node.first_child;
            while ((child_idx != ((ax_u32)(0)))) {
                ax_CtgcInjector_traverse_and_inject(self, child_idx, node_idx);
                child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
            }
        }
    }
}

void ax_CtgcInjector_append_child_node(struct ax_CtgcInjector* self, ax_u32 parent, ax_u32 child) {
    if ((((self->tree.nodes.data)[parent]).first_child == ((ax_u32)(0)))) {
        ((self->tree.nodes.data)[parent]).first_child = child;
        return;
    }
    ax_u32 cur = ((self->tree.nodes.data)[parent]).first_child;
    while ((((self->tree.nodes.data)[cur]).next_sibling != ((ax_u32)(0)))) {
        cur = ((self->tree.nodes.data)[cur]).next_sibling;
    }
    ((self->tree.nodes.data)[cur]).next_sibling = child;
}

void ax_CtgcInjector_insert_before(struct ax_CtgcInjector* self, ax_u32 parent, ax_u32 target, ax_u32 new_node) {
    if ((((self->tree.nodes.data)[parent]).first_child == target)) {
        ((self->tree.nodes.data)[new_node]).next_sibling = target;
        ((self->tree.nodes.data)[parent]).first_child = new_node;
        return;
    }
    ax_u32 cur = ((self->tree.nodes.data)[parent]).first_child;
    while ((cur != ((ax_u32)(0)))) {
        if ((((self->tree.nodes.data)[cur]).next_sibling == target)) {
            ((self->tree.nodes.data)[new_node]).next_sibling = target;
            ((self->tree.nodes.data)[cur]).next_sibling = new_node;
            return;
        }
        cur = ((self->tree.nodes.data)[cur]).next_sibling;
    }
}

struct ax_AliasReuseOptimizer ax_AstTree_new_alias_reuse_optimizer(struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable) {
    return ((struct ax_AliasReuseOptimizer){.tree=tree, .pool=pool, .symtable=symtable, .typetable=typetable});
}

void ax_AliasReuseOptimizer_run(struct ax_AliasReuseOptimizer* self) {
    ax_AliasReuseOptimizer_optimize_node(self, ((ax_u32)(0)));
}

void ax_AliasReuseOptimizer_optimize_node(struct ax_AliasReuseOptimizer* self, ax_u32 node_idx) {
    if (((node_idx == ((ax_u32)(0xffffffff))) || (node_idx == ((ax_u32)(0))))) {
        return;
    }
    struct ax_AstNode node = ((self->tree.nodes.data)[node_idx]);
    ax_u8 kind = node.kind;
    if ((kind == ax_NODE_BLOCK)) {
        ax_u32 child_idx = node.first_child;
        while ((child_idx != ((ax_u32)(0)))) {
            struct ax_AstNode child = ((self->tree.nodes.data)[child_idx]);
            if ((child.kind == ax_NODE_DESTROY_STMT)) {
                ax_u32 next_idx = child.next_sibling;
                if ((next_idx != ((ax_u32)(0)))) {
                    struct ax_AstNode next_node = ((self->tree.nodes.data)[next_idx]);
                    if ((next_node.kind == ax_NODE_VAR_DECL)) {
                        ax_u32 destroy_sym = child.payload;
                        ax_u32 alloc_sym = next_node.payload;
                        if (((destroy_sym != ((ax_u32)(0))) && (alloc_sym != ((ax_u32)(0))))) {
                            struct ax_Symbol sym_d = ((self->symtable.symbols.data)[destroy_sym]);
                            struct ax_Symbol sym_a = ((self->symtable.symbols.data)[alloc_sym]);
                            if ((sym_d.type_id == sym_a.type_id)) {
                                ((self->tree.nodes.data)[next_idx]).flags = (((self->tree.nodes.data)[next_idx]).flags | ((ax_u16)(ax_FLAG_IS_MOVED)));
                            }
                        }
                    }
                }
            }
            ax_AliasReuseOptimizer_optimize_node(self, child_idx);
            child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
        }
    } else {
        {
            ax_u32 child_idx = node.first_child;
            while ((child_idx != ((ax_u32)(0)))) {
                ax_AliasReuseOptimizer_optimize_node(self, child_idx);
                child_idx = ((self->tree.nodes.data)[child_idx]).next_sibling;
            }
        }
    }
}

struct ax_AirInstVec ax_new_air_inst_vec(void) {
    return ((struct ax_AirInstVec){.data=((struct ax_AirInst*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_AirInstVec_push(struct ax_AirInstVec* self, struct ax_AirInst inst) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_AirInst* new_data = ((struct ax_AirInst*)(ax_alloc((new_cap * 16))));
        if ((self->data != ((struct ax_AirInst*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 16));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = inst;
    self->len = (self->len + 1);
    return idx;
}

struct ax_BasicBlockVec ax_new_basic_block_vec(void) {
    return ((struct ax_BasicBlockVec){.data=((struct ax_BasicBlock*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_BasicBlockVec_push(struct ax_BasicBlockVec* self, struct ax_BasicBlock bb) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_BasicBlock* new_data = ((struct ax_BasicBlock*)(ax_alloc((new_cap * 32))));
        if ((self->data != ((struct ax_BasicBlock*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 32));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = bb;
    self->len = (self->len + 1);
    return idx;
}

void ax_AirFunc_free_air_func(struct ax_AirFunc* f) {
    if ((f->params.data != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(f->params.data)));
    }
    if ((f->blocks.data != ((struct ax_BasicBlock*)(NULL)))) {
        ax_free(((ax_u8*)(f->blocks.data)));
    }
    if ((f->insts.data != ((struct ax_AirInst*)(NULL)))) {
        ax_free(((ax_u8*)(f->insts.data)));
    }
    if ((f->extras.data != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(f->extras.data)));
    }
    if ((f->block_instrs.data != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(f->block_instrs.data)));
    }
    if ((f->block_succs.data != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(f->block_succs.data)));
    }
    if ((f->block_preds.data != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(f->block_preds.data)));
    }
}

struct ax_AirFuncVec ax_new_air_func_vec(void) {
    return ((struct ax_AirFuncVec){.data=((struct ax_AirFunc*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_AirFuncVec_push(struct ax_AirFuncVec* self, struct ax_AirFunc f) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_AirFunc* new_data = ((struct ax_AirFunc*)(ax_alloc((new_cap * 192))));
        if ((self->data != ((struct ax_AirFunc*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 192));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = f;
    self->len = (self->len + 1);
    return idx;
}

void ax_AirModule_free_air_module(struct ax_AirModule* m) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < m->funcs.len)) {
        ax_AirFunc_free_air_func(&(((m->funcs.data)[i])));
        i = (i + 1);
    }
    if ((m->funcs.data != ((struct ax_AirFunc*)(NULL)))) {
        ax_free(((ax_u8*)(m->funcs.data)));
    }
}

struct ax_AirFuncBuilder ax_new_air_func_builder(ax_u32 name, ax_u32 ret_type) {
    struct ax_AirFuncBuilder builder = ((struct ax_AirFuncBuilder){.name=name, .ret_type=ret_type, .blocks=ax_new_basic_block_vec(), .insts=ax_new_air_inst_vec(), .extras=ax_new_u32_vec(), .cur_block=(-((ax_i32)(1))), .next_reg=((ax_u32)(1)), .block_instrs=ax_new_u32_vec(), .block_succs=ax_new_u32_vec(), .block_preds=ax_new_u32_vec()});
    ax_u32 entry_id = ax_AirFuncBuilder_new_block(&(builder));
    ((builder.blocks.data)[entry_id]).is_entry = AX_TRUE;
    ax_AirFuncBuilder_switch_to(&(builder), entry_id);
    return builder;
    /* skip destroy for value type builder */
}

ax_u32 ax_AirFuncBuilder_new_block(struct ax_AirFuncBuilder* self) {
    ax_u32 id = ((ax_u32)(self->blocks.len));
    ax_BasicBlockVec_push(&(self->blocks), ((struct ax_BasicBlock){.id=id, .instrs_start=((ax_u32)(0)), .instrs_len=((ax_u32)(0)), .succs_start=((ax_u32)(0)), .succs_len=((ax_u32)(0)), .preds_start=((ax_u32)(0)), .preds_len=((ax_u32)(0)), .loop_depth=((ax_u8)(0)), .is_entry=AX_FALSE, .is_exit=AX_FALSE}));
    return id;
}

void ax_AirFuncBuilder_switch_to(struct ax_AirFuncBuilder* self, ax_u32 block_id) {
    self->cur_block = ((ax_i32)(block_id));
}

ax_u32 ax_AirFuncBuilder_current_block(struct ax_AirFuncBuilder self) {
    if ((self.cur_block < 0)) {
        return ((ax_u32)(0xffffffff));
    }
    return ((ax_u32)(self.cur_block));
}

ax_u32 ax_AirFuncBuilder_emit(struct ax_AirFuncBuilder* self, struct ax_AirInst inst) {
    ax_u32 idx = ax_AirInstVec_push(&(self->insts), inst);
    if (((self->cur_block >= 0) && (((ax_i64)(self->cur_block)) < self->blocks.len))) {
        struct ax_BasicBlock* bb = &(((self->blocks.data)[self->cur_block]));
        if ((bb->instrs_len == 0)) {
            bb->instrs_start = ((ax_u32)(self->block_instrs.len));
        }
        ax_U32Vec_push(&(self->block_instrs), idx);
        bb->instrs_len = (bb->instrs_len + ((ax_u32)(1)));
    }
    return idx;
}

ax_u32 ax_AirFuncBuilder_emit_extra(struct ax_AirFuncBuilder* self, ax_u32 val) {
    return ax_U32Vec_push(&(self->extras), val);
}

void ax_AirFuncBuilder_set_extra(struct ax_AirFuncBuilder* self, ax_u32 idx, ax_u32 val) {
    if ((((ax_i64)(idx)) < self->extras.len)) {
        ((self->extras.data)[idx]) = val;
    }
}

ax_u32 ax_AirFuncBuilder_fresh_reg(struct ax_AirFuncBuilder* self) {
    ax_u32 r = self->next_reg;
    self->next_reg = (self->next_reg + ((ax_u32)(1)));
    return r;
}

void ax_AirFuncBuilder_add_edge(struct ax_AirFuncBuilder* self, ax_u32 src, ax_u32 dst) {
    if (((((ax_i64)(src)) < self->blocks.len) && (((ax_i64)(dst)) < self->blocks.len))) {
        struct ax_BasicBlock* src_bb = &(((self->blocks.data)[src]));
        if ((src_bb->succs_len == 0)) {
            src_bb->succs_start = ((ax_u32)(self->block_succs.len));
        }
        ax_bool exists = AX_FALSE;
        ax_i64 i = ((ax_i64)(0));
        while ((i < ((ax_i64)(src_bb->succs_len)))) {
            ax_i64 idx = (((ax_i64)(src_bb->succs_start)) + i);
            if ((((self->block_succs.data)[idx]) == dst)) {
                exists = AX_TRUE;
            }
            i = (i + 1);
        }
        if ((!exists)) {
            ax_U32Vec_push(&(self->block_succs), dst);
            src_bb->succs_len = (src_bb->succs_len + ((ax_u32)(1)));
        }
        struct ax_BasicBlock* dst_bb = &(((self->blocks.data)[dst]));
        if ((dst_bb->preds_len == 0)) {
            dst_bb->preds_start = ((ax_u32)(self->block_preds.len));
        }
        exists = AX_FALSE;
        i = 0;
        while ((i < ((ax_i64)(dst_bb->preds_len)))) {
            ax_i64 idx = (((ax_i64)(dst_bb->preds_start)) + i);
            if ((((self->block_preds.data)[idx]) == src)) {
                exists = AX_TRUE;
            }
            i = (i + 1);
        }
        if ((!exists)) {
            ax_U32Vec_push(&(self->block_preds), src);
            dst_bb->preds_len = (dst_bb->preds_len + ((ax_u32)(1)));
        }
    }
}

struct ax_AirFunc ax_AirFuncBuilder_build_func(struct ax_AirFuncBuilder* self) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->blocks.len)) {
        struct ax_BasicBlock* blk = &(((self->blocks.data)[i]));
        if ((blk->instrs_len > 0)) {
            ax_i64 last_idx_idx = ((((ax_i64)(blk->instrs_start)) + ((ax_i64)(blk->instrs_len))) - 1);
            ax_i64 last_idx = ((ax_i64)(((self->block_instrs.data)[last_idx_idx])));
            if (((last_idx < self->insts.len) && (((self->insts.data)[last_idx]).opcode == ax_OP_RETURN))) {
                blk->is_exit = AX_TRUE;
            }
        }
        i = (i + 1);
    }
    return ((struct ax_AirFunc){.sym_id=((ax_u32)(0)), .name=self->name, .params=ax_new_u32_vec(), .ret_type=self->ret_type, .blocks=self->blocks, .insts=self->insts, .extras=self->extras, .is_async=AX_FALSE, .is_extern=AX_FALSE, .block_instrs=self->block_instrs, .block_succs=self->block_succs, .block_preds=self->block_preds});
}

ax_string ax_opcode_mnemonic(ax_u16 op) {
    if ((op == ax_OP_NOP)) {
        return (ax_string){.ptr=(const ax_u8*)"nop", .len=3};
    }
    if ((op == ax_OP_ALLOC)) {
        return (ax_string){.ptr=(const ax_u8*)"alloc", .len=5};
    }
    if ((op == ax_OP_FREE)) {
        return (ax_string){.ptr=(const ax_u8*)"free", .len=4};
    }
    if ((op == ax_OP_LOAD)) {
        return (ax_string){.ptr=(const ax_u8*)"load", .len=4};
    }
    if ((op == ax_OP_STORE)) {
        return (ax_string){.ptr=(const ax_u8*)"store", .len=5};
    }
    if ((op == ax_OP_GEP)) {
        return (ax_string){.ptr=(const ax_u8*)"gep", .len=3};
    }
    if ((op == ax_OP_COPY)) {
        return (ax_string){.ptr=(const ax_u8*)"copy", .len=4};
    }
    if ((op == ax_OP_MOVE)) {
        return (ax_string){.ptr=(const ax_u8*)"move", .len=4};
    }
    if ((op == ax_OP_MAKE_REF)) {
        return (ax_string){.ptr=(const ax_u8*)"mkref", .len=5};
    }
    if ((op == ax_OP_DEREF)) {
        return (ax_string){.ptr=(const ax_u8*)"deref", .len=5};
    }
    if ((op == ax_OP_ARENA_ALLOC)) {
        return (ax_string){.ptr=(const ax_u8*)"aalloc", .len=6};
    }
    if ((op == ax_OP_DESTROY)) {
        return (ax_string){.ptr=(const ax_u8*)"destroy", .len=7};
    }
    if ((op == ax_OP_ALIAS_REUSE)) {
        return (ax_string){.ptr=(const ax_u8*)"areuse", .len=6};
    }
    if ((op == ax_OP_GET_FIELD)) {
        return (ax_string){.ptr=(const ax_u8*)"getfld", .len=6};
    }
    if ((op == ax_OP_SET_FIELD)) {
        return (ax_string){.ptr=(const ax_u8*)"setfld", .len=6};
    }
    if ((op == ax_OP_INDEX)) {
        return (ax_string){.ptr=(const ax_u8*)"index", .len=5};
    }
    if ((op == ax_OP_SLICE)) {
        return (ax_string){.ptr=(const ax_u8*)"slice", .len=5};
    }
    if ((op == ax_OP_ICONST)) {
        return (ax_string){.ptr=(const ax_u8*)"iconst", .len=6};
    }
    if ((op == ax_OP_FCONST)) {
        return (ax_string){.ptr=(const ax_u8*)"fconst", .len=6};
    }
    if ((op == ax_OP_IADD)) {
        return (ax_string){.ptr=(const ax_u8*)"iadd", .len=4};
    }
    if ((op == ax_OP_ISUB)) {
        return (ax_string){.ptr=(const ax_u8*)"isub", .len=4};
    }
    if ((op == ax_OP_IMUL)) {
        return (ax_string){.ptr=(const ax_u8*)"imul", .len=4};
    }
    if ((op == ax_OP_IDIV)) {
        return (ax_string){.ptr=(const ax_u8*)"idiv", .len=4};
    }
    if ((op == ax_OP_IMOD)) {
        return (ax_string){.ptr=(const ax_u8*)"imod", .len=4};
    }
    if ((op == ax_OP_FADD)) {
        return (ax_string){.ptr=(const ax_u8*)"fadd", .len=4};
    }
    if ((op == ax_OP_FSUB)) {
        return (ax_string){.ptr=(const ax_u8*)"fsub", .len=4};
    }
    if ((op == ax_OP_FMUL)) {
        return (ax_string){.ptr=(const ax_u8*)"fmul", .len=4};
    }
    if ((op == ax_OP_FDIV)) {
        return (ax_string){.ptr=(const ax_u8*)"fdiv", .len=4};
    }
    if ((op == ax_OP_EQ)) {
        return (ax_string){.ptr=(const ax_u8*)"eq", .len=2};
    }
    if ((op == ax_OP_NE)) {
        return (ax_string){.ptr=(const ax_u8*)"ne", .len=2};
    }
    if ((op == ax_OP_LT)) {
        return (ax_string){.ptr=(const ax_u8*)"lt", .len=2};
    }
    if ((op == ax_OP_LE)) {
        return (ax_string){.ptr=(const ax_u8*)"le", .len=2};
    }
    if ((op == ax_OP_GT)) {
        return (ax_string){.ptr=(const ax_u8*)"gt", .len=2};
    }
    if ((op == ax_OP_GE)) {
        return (ax_string){.ptr=(const ax_u8*)"ge", .len=2};
    }
    if ((op == ax_OP_AND)) {
        return (ax_string){.ptr=(const ax_u8*)"and", .len=3};
    }
    if ((op == ax_OP_OR)) {
        return (ax_string){.ptr=(const ax_u8*)"or", .len=2};
    }
    if ((op == ax_OP_XOR)) {
        return (ax_string){.ptr=(const ax_u8*)"xor", .len=3};
    }
    if ((op == ax_OP_SHL)) {
        return (ax_string){.ptr=(const ax_u8*)"shl", .len=3};
    }
    if ((op == ax_OP_SHR)) {
        return (ax_string){.ptr=(const ax_u8*)"shr", .len=3};
    }
    if ((op == ax_OP_NOT)) {
        return (ax_string){.ptr=(const ax_u8*)"not", .len=3};
    }
    if ((op == ax_OP_NEG)) {
        return (ax_string){.ptr=(const ax_u8*)"neg", .len=3};
    }
    if ((op == ax_OP_ITOF)) {
        return (ax_string){.ptr=(const ax_u8*)"itof", .len=4};
    }
    if ((op == ax_OP_FTOI)) {
        return (ax_string){.ptr=(const ax_u8*)"ftoi", .len=4};
    }
    if ((op == ax_OP_ZEXT)) {
        return (ax_string){.ptr=(const ax_u8*)"zext", .len=4};
    }
    if ((op == ax_OP_SEXT)) {
        return (ax_string){.ptr=(const ax_u8*)"sext", .len=4};
    }
    if ((op == ax_OP_TRUNC)) {
        return (ax_string){.ptr=(const ax_u8*)"trunc", .len=5};
    }
    if ((op == ax_OP_CAST)) {
        return (ax_string){.ptr=(const ax_u8*)"cast", .len=4};
    }
    if ((op == ax_OP_JUMP)) {
        return (ax_string){.ptr=(const ax_u8*)"jump", .len=4};
    }
    if ((op == ax_OP_BRANCH)) {
        return (ax_string){.ptr=(const ax_u8*)"branch", .len=6};
    }
    if ((op == ax_OP_CALL)) {
        return (ax_string){.ptr=(const ax_u8*)"call", .len=4};
    }
    if ((op == ax_OP_RETURN)) {
        return (ax_string){.ptr=(const ax_u8*)"ret", .len=3};
    }
    if ((op == ax_OP_PHI)) {
        return (ax_string){.ptr=(const ax_u8*)"phi", .len=3};
    }
    if ((op == ax_OP_LOOP_BEGIN)) {
        return (ax_string){.ptr=(const ax_u8*)"loopbeg", .len=7};
    }
    if ((op == ax_OP_LOOP_END)) {
        return (ax_string){.ptr=(const ax_u8*)"loopend", .len=7};
    }
    if ((op == ax_OP_SPAWN)) {
        return (ax_string){.ptr=(const ax_u8*)"spawn", .len=5};
    }
    if ((op == ax_OP_SEND)) {
        return (ax_string){.ptr=(const ax_u8*)"send", .len=4};
    }
    if ((op == ax_OP_RECV)) {
        return (ax_string){.ptr=(const ax_u8*)"recv", .len=4};
    }
    if ((op == ax_OP_AWAIT)) {
        return (ax_string){.ptr=(const ax_u8*)"await", .len=5};
    }
    if ((op == ax_OP_SYSCALL)) {
        return (ax_string){.ptr=(const ax_u8*)"syscall", .len=7};
    }
    if ((op == ax_OP_SIMD_LOAD)) {
        return (ax_string){.ptr=(const ax_u8*)"vload", .len=5};
    }
    if ((op == ax_OP_SIMD_STORE)) {
        return (ax_string){.ptr=(const ax_u8*)"vstore", .len=6};
    }
    if ((op == ax_OP_SIMD_ADD)) {
        return (ax_string){.ptr=(const ax_u8*)"vadd", .len=4};
    }
    if ((op == ax_OP_SIMD_MUL)) {
        return (ax_string){.ptr=(const ax_u8*)"vmul", .len=4};
    }
    if ((op == ax_OP_SIMD_FMA)) {
        return (ax_string){.ptr=(const ax_u8*)"vfma", .len=4};
    }
    if ((op == ax_OP_COMPTIME)) {
        return (ax_string){.ptr=(const ax_u8*)"comptime", .len=8};
    }
    return (ax_string){.ptr=(const ax_u8*)"???", .len=3};
}

ax_u16 ax_opcode_class(ax_u16 op) {
    return (op >> ((ax_u16)(8)));
}

ax_bool ax_opcode_is_binary_alu(ax_u16 op) {
    if ((op == ax_OP_IADD)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_ISUB)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_IMUL)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_IDIV)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_IMOD)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_FADD)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_FSUB)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_FMUL)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_FDIV)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_EQ)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_NE)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_LT)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_LE)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_GT)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_GE)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_AND)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_OR)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_XOR)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_SHL)) {
        return AX_TRUE;
    }
    if ((op == ax_OP_SHR)) {
        return AX_TRUE;
    }
    return AX_FALSE;
}

struct ax_LocalMap ax_new_local_map(void) {
    ax_u32 capacity = ((ax_u32)(16));
    struct ax_LocalMapEntry* entries = ((struct ax_LocalMapEntry*)(ax_alloc((((ax_i64)(capacity)) * 8))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < ((ax_i64)(capacity)))) {
        ((entries)[i]) = ((struct ax_LocalMapEntry){.name_id=((ax_u32)(0)), .reg=((ax_u32)(0))});
        i = (i + 1);
    }
    return ((struct ax_LocalMap){.entries=entries, .count=((ax_u32)(0)), .capacity=capacity});
}

void ax_LocalMap_free_local_map(struct ax_LocalMap* self) {
    if ((self->entries != ((struct ax_LocalMapEntry*)(NULL)))) {
        ax_free(((ax_u8*)(self->entries)));
    }
}

static ax_u32 ax_local_map_hash(ax_u32 v) {
    ax_u32 hash = ((ax_u32)(2166136261));
    hash = (hash ^ (v & ((ax_u32)(255))));
    hash = (hash * ((ax_u32)(16777619)));
    hash = (hash ^ ((v >> ((ax_u32)(8))) & ((ax_u32)(255))));
    hash = (hash * ((ax_u32)(16777619)));
    hash = (hash ^ ((v >> ((ax_u32)(16))) & ((ax_u32)(255))));
    hash = (hash * ((ax_u32)(16777619)));
    hash = (hash ^ ((v >> ((ax_u32)(24))) & ((ax_u32)(255))));
    hash = (hash * ((ax_u32)(16777619)));
    return hash;
}

void ax_LocalMap_local_map_put(struct ax_LocalMap* self, ax_u32 name_id, ax_u32 reg) {
    if (((self->count * ((ax_u32)(4))) > (self->capacity * ((ax_u32)(3))))) {
        ax_LocalMap_local_map_grow(self);
    }
    ax_LocalMap_local_map_insert(*(self), name_id, reg);
    self->count = (self->count + ((ax_u32)(1)));
}

static void ax_LocalMap_local_map_insert(struct ax_LocalMap self, ax_u32 name_id, ax_u32 reg) {
    ax_u32 mask = (self.capacity - ((ax_u32)(1)));
    ax_u32 idx = (ax_local_map_hash(name_id) & mask);
    while (AX_TRUE) {
        if (((((self.entries)[idx]).name_id == ((ax_u32)(0))) || (((self.entries)[idx]).name_id == name_id))) {
            ((self.entries)[idx]).name_id = name_id;
            ((self.entries)[idx]).reg = reg;
            return;
        }
        idx = ((idx + ((ax_u32)(1))) & mask);
    }
}

ax_u32 ax_LocalMap_local_map_get(struct ax_LocalMap self, ax_u32 name_id) {
    if ((self.capacity == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 mask = (self.capacity - ((ax_u32)(1)));
    ax_u32 idx = (ax_local_map_hash(name_id) & mask);
    ax_u32 start_idx = idx;
    while (AX_TRUE) {
        struct ax_LocalMapEntry entry = ((self.entries)[idx]);
        if ((entry.name_id == ((ax_u32)(0)))) {
            return ((ax_u32)(0));
        }
        if ((entry.name_id == name_id)) {
            return entry.reg;
        }
        idx = ((idx + ((ax_u32)(1))) & mask);
        if ((idx == start_idx)) {
            break;
        }
    }
    return ((ax_u32)(0));
}

static void ax_LocalMap_local_map_grow(struct ax_LocalMap* self) {
    struct ax_LocalMapEntry* old_entries = self->entries;
    ax_u32 old_capacity = self->capacity;
    ax_u32 new_cap = (old_capacity * ((ax_u32)(2)));
    self->capacity = new_cap;
    self->entries = ((struct ax_LocalMapEntry*)(ax_alloc((((ax_i64)(new_cap)) * 8))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < ((ax_i64)(new_cap)))) {
        ((self->entries)[i]) = ((struct ax_LocalMapEntry){.name_id=((ax_u32)(0)), .reg=((ax_u32)(0))});
        i = (i + 1);
    }
    i = ((ax_i64)(0));
    while ((i < ((ax_i64)(old_capacity)))) {
        struct ax_LocalMapEntry entry = ((old_entries)[i]);
        if ((entry.name_id != ((ax_u32)(0)))) {
            ax_LocalMap_local_map_insert(*(self), entry.name_id, entry.reg);
        }
        i = (i + 1);
    }
    ax_free(((ax_u8*)(old_entries)));
}

struct ax_AirModuleBuilder ax_AstTree_new_air_module_builder(struct ax_AstTree tree, struct ax_SymbolTable symbols, struct ax_TypeTable typetable, struct ax_InternPool pool, ax_u32* node_types) {
    return ((struct ax_AirModuleBuilder){.tree=tree, .symbols=symbols, .typetable=typetable, .pool=pool, .module=((struct ax_AirModule){.funcs=ax_new_air_func_vec()}), .node_types=node_types});
}

void ax_AirModuleBuilder_free_air_module_builder(struct ax_AirModuleBuilder* self) {
    ax_AirModule_free_air_module(&(self->module));
}

ax_string ax_AirModuleBuilder_get_token_text(struct ax_AirModuleBuilder* self, ax_u32 token_idx) {
    struct ax_Token tok = ((self->tree.tokens.data)[((ax_i64)(token_idx))]);
    if ((tok.len == ((ax_u16)(0)))) {
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    ax_u8* start_ptr = ((ax_u8*)((((ax_i64)(((ax_u8*)(self->tree.src.ptr)))) + ((ax_i64)(tok.offset)))));
    return ax_alloc_str_from_raw(start_ptr, ((ax_i64)(tok.len)));
}

ax_i64 ax_parse_int_from_str(ax_string s) {
    ax_i64 len = ax_str_len(s);
    if ((len == 0)) {
        return ((ax_i64)(0));
    }
    ax_i32 i = 0;
    ax_i64 sign = ((ax_i64)(1));
    if ((((ax_u8)((ax_bounds_check((ax_u64)(0), (s).len), (s).ptr[0]))) == ((ax_u8)('-')))) {
        sign = (-((ax_i64)(1)));
        i = 1;
    } else if ((((ax_u8)((ax_bounds_check((ax_u64)(0), (s).len), (s).ptr[0]))) == ((ax_u8)('+')))) {
        i = 1;
    }
    ax_i64 base = ((ax_i64)(10));
    if ((((i + 1) < len) && (((ax_u8)((ax_bounds_check((ax_u64)(i), (s).len), (s).ptr[i]))) == ((ax_u8)('0'))))) {
        ax_u8 next_char = ((ax_u8)((ax_bounds_check((ax_u64)((i + 1)), (s).len), (s).ptr[(i + 1)])));
        if (((next_char == ((ax_u8)('x'))) || (next_char == ((ax_u8)('X'))))) {
            base = ((ax_i64)(16));
            i = (i + 2);
        } else if (((next_char == ((ax_u8)('o'))) || (next_char == ((ax_u8)('O'))))) {
            base = ((ax_i64)(8));
            i = (i + 2);
        } else if (((next_char == ((ax_u8)('b'))) || (next_char == ((ax_u8)('B'))))) {
            base = ((ax_i64)(2));
            i = (i + 2);
        }
    }
    ax_i64 val = ((ax_i64)(0));
    while ((i < len)) {
        ax_u8 c = ((ax_u8)((ax_bounds_check((ax_u64)(i), (s).len), (s).ptr[i])));
        ax_i64 digit = (-((ax_i64)(1)));
        if (((c >= ((ax_u8)('0'))) && (c <= ((ax_u8)('9'))))) {
            digit = ((ax_i64)((c - ((ax_u8)('0')))));
        } else if (((c >= ((ax_u8)('a'))) && (c <= ((ax_u8)('f'))))) {
            digit = ((ax_i64)(((c - ((ax_u8)('a'))) + ((ax_u8)(10)))));
        } else if (((c >= ((ax_u8)('A'))) && (c <= ((ax_u8)('F'))))) {
            digit = ((ax_i64)(((c - ((ax_u8)('A'))) + ((ax_u8)(10)))));
        }
        if (((digit < 0) || (digit >= base))) {
            break;
        }
        val = ((val * base) + digit);
        i = (i + 1);
    }
    return (val * sign);
}

ax_u16 ax_map_binary_op(ax_string op) {
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"+", .len=1})) {
        return ax_OP_IADD;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"-", .len=1})) {
        return ax_OP_ISUB;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"*", .len=1})) {
        return ax_OP_IMUL;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"/", .len=1})) {
        return ax_OP_IDIV;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"%", .len=1})) {
        return ax_OP_IMOD;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"==", .len=2})) {
        return ax_OP_EQ;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"!=", .len=2})) {
        return ax_OP_NE;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"<", .len=1})) {
        return ax_OP_LT;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"<=", .len=2})) {
        return ax_OP_LE;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)">", .len=1})) {
        return ax_OP_GT;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)">=", .len=2})) {
        return ax_OP_GE;
    }
    if ((ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"and", .len=3}) || ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"&&", .len=2}))) {
        return ax_OP_AND;
    }
    if ((ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"or", .len=2}) || ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"||", .len=2}))) {
        return ax_OP_OR;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"^", .len=1})) {
        return ax_OP_XOR;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)"<<", .len=2})) {
        return ax_OP_SHL;
    }
    if (ax_str_eq(op, (ax_string){.ptr=(const ax_u8*)">>", .len=2})) {
        return ax_OP_SHR;
    }
    return ax_OP_NOP;
}

struct ax_FuncLowering ax_AirModuleBuilder_new_func_lowering(struct ax_AirModuleBuilder* mb, struct ax_AirFuncBuilder fb, struct ax_U32Vec params) {
    return ((struct ax_FuncLowering){.mb=mb, .fb=fb, .locals=ax_new_local_map(), .params=params, .terminated=AX_FALSE});
}

void ax_FuncLowering_free_func_lowering(struct ax_FuncLowering* self) {
    ax_LocalMap_free_local_map(&(self->locals));
    if ((self->params.data != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(self->params.data)));
    }
}

void ax_FuncLowering_register_params(struct ax_FuncLowering* self, ax_u32 func_idx, struct ax_AstNode func_node) {
    ax_u32 child = func_node.first_child;
    ax_i64 param_idx = ((ax_i64)(0));
    while ((child != ((ax_u32)(0)))) {
        struct ax_AstNode cn = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
        if ((cn.kind == ax_NODE_PARAM_DECL)) {
            ax_u32 name_id = cn.payload;
            if ((name_id == ((ax_u32)(0)))) {
                ax_string tok_str = ax_AirModuleBuilder_get_token_text(self->mb, cn.token_idx);
                name_id = ax_InternPool_intern(&(self->mb->pool), tok_str);
                ax_free(((ax_u8*)(tok_str.ptr)));
            }
            ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
            ax_u16 type_id = ((ax_u16)(0));
            if ((param_idx < self->params.len)) {
                type_id = ((ax_u16)(((self->params.data)[param_idx])));
            }
            ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_COPY, .type_id=type_id, .dest=reg, .src1=((ax_u32)((param_idx + 1))), .src2=((ax_u32)(0))}));
            ax_LocalMap_local_map_put(&(self->locals), name_id, reg);
            param_idx = (param_idx + 1);
        }
        child = cn.next_sibling;
    }
}

ax_u32 ax_FuncLowering_lower_expr(struct ax_FuncLowering* self, ax_u32 idx) {
    if ((idx == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    struct ax_AstNode node = ((self->mb->tree.nodes.data)[((ax_i64)(idx))]);
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]   lower_expr: idx=%d, kind=%d\n", .len=38}).ptr, ((ax_i64)(idx)), ((ax_i64)(node.kind)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    if ((node.kind == ax_NODE_INT_LIT)) {
        return ax_FuncLowering_lower_int_lit(self, idx, node);
    }
    if ((node.kind == ax_NODE_FLOAT_LIT)) {
        return ax_FuncLowering_lower_float_lit(self, idx, node);
    }
    if ((node.kind == ax_NODE_BOOL_LIT)) {
        return ax_FuncLowering_lower_bool_lit(self, node);
    }
    if ((node.kind == ax_NODE_NIL_LIT)) {
        return ax_FuncLowering_lower_nil_lit(self);
    }
    if ((node.kind == ax_NODE_STRING_LIT)) {
        return ax_FuncLowering_lower_string_lit(self, node);
    }
    if ((node.kind == ax_NODE_CHAR_LIT)) {
        return ax_FuncLowering_lower_char_lit(self, node);
    }
    if ((node.kind == ax_NODE_IDENT)) {
        return ax_FuncLowering_lower_ident(self, node);
    }
    if ((node.kind == ax_NODE_BINARY_EXPR)) {
        return ax_FuncLowering_lower_binary_expr(self, idx, node);
    }
    if ((node.kind == ax_NODE_UNARY_EXPR)) {
        return ax_FuncLowering_lower_unary_expr(self, idx, node);
    }
    if ((node.kind == ax_NODE_CALL_EXPR)) {
        return ax_FuncLowering_lower_call_expr(self, idx, node);
    }
    if ((node.kind == ax_NODE_FIELD_EXPR)) {
        return ax_FuncLowering_lower_field_expr(self, idx, node);
    }
    if ((node.kind == ax_NODE_INDEX_EXPR)) {
        return ax_FuncLowering_lower_index_expr(self, idx, node);
    }
    if ((node.kind == ax_NODE_CAST_EXPR)) {
        return ax_FuncLowering_lower_cast_expr(self, idx, node);
    }
    if ((node.kind == ax_NODE_DEREF_EXPR)) {
        return ax_FuncLowering_lower_deref_expr(self, idx, node);
    }
    if ((node.kind == ax_NODE_SPAWN_EXPR)) {
        return ax_FuncLowering_lower_spawn_expr(self, idx, node);
    }
    if ((node.kind == ax_NODE_AWAIT_EXPR)) {
        return ax_FuncLowering_lower_await_expr(self, idx, node);
    }
    if ((node.kind == ax_NODE_STRUCT_LIT)) {
        return ax_FuncLowering_lower_struct_lit(self, idx, node);
    }
    if ((node.kind == ax_NODE_ARRAY_LIT)) {
        return ax_FuncLowering_lower_array_lit(self, idx, node);
    }
    return ((ax_u32)(0));
}

ax_u32 ax_FuncLowering_lower_int_lit(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_string text = ax_AirModuleBuilder_get_token_text(self->mb, node.token_idx);
    ax_i64 val = ax_parse_int_from_str(text);
    ax_free(((ax_u8*)(text.ptr)));
    ax_u16 type_id = ((ax_u16)(3));
    if (((self->mb->node_types != ((ax_u32*)(NULL))) && (((self->mb->node_types)[((ax_i64)(idx))]) != ((ax_u32)(0))))) {
        type_id = ((ax_u16)(((self->mb->node_types)[((ax_i64)(idx))])));
    }
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ICONST, .type_id=type_id, .dest=reg, .src1=((ax_u32)(val)), .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_float_lit(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_string text = ax_AirModuleBuilder_get_token_text(self->mb, node.token_idx);
    ax_f64 val = atof((const char*)(text).ptr);
    ax_free(((ax_u8*)(text.ptr)));
    ax_u16 type_id = ((ax_u16)(10));
    if (((self->mb->node_types != ((ax_u32*)(NULL))) && (((self->mb->node_types)[((ax_i64)(idx))]) != ((ax_u32)(0))))) {
        type_id = ((ax_u16)(((self->mb->node_types)[((ax_i64)(idx))])));
    }
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_FCONST, .type_id=type_id, .dest=reg, .src1=((ax_u32)(val)), .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_bool_lit(struct ax_FuncLowering* self, struct ax_AstNode node) {
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_string text = ax_AirModuleBuilder_get_token_text(self->mb, node.token_idx);
    ax_u32 val = ((ax_u32)(0));
    if (ax_str_eq(text, (ax_string){.ptr=(const ax_u8*)"true", .len=4})) {
        val = ((ax_u32)(1));
    }
    ax_free(((ax_u8*)(text.ptr)));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ICONST, .type_id=((ax_u16)(11)), .dest=reg, .src1=val, .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_nil_lit(struct ax_FuncLowering* self) {
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ICONST, .type_id=((ax_u16)(0)), .dest=reg, .src1=((ax_u32)(0)), .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_string_lit(struct ax_FuncLowering* self, struct ax_AstNode node) {
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_string text = ax_AirModuleBuilder_get_token_text(self->mb, node.token_idx);
    ax_u32 str_id = ax_InternPool_intern(&(self->mb->pool), text);
    ax_free(((ax_u8*)(text.ptr)));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ICONST, .type_id=((ax_u16)(12)), .dest=reg, .src1=str_id, .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_char_lit(struct ax_FuncLowering* self, struct ax_AstNode node) {
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_string text = ax_AirModuleBuilder_get_token_text(self->mb, node.token_idx);
    ax_u32 val = ((ax_u32)(0));
    ax_i64 length = ax_str_len(text);
    if ((length > 0)) {
        if ((((length >= 3) && (((ax_u8)((ax_bounds_check((ax_u64)(0), (text).len), (text).ptr[0]))) == ((ax_u8)('\'')))) && (((ax_u8)((ax_bounds_check((ax_u64)((length - 1)), (text).len), (text).ptr[(length - 1)]))) == ((ax_u8)('\''))))) {
            val = ((ax_u32)((ax_bounds_check((ax_u64)(1), (text).len), (text).ptr[1])));
        }
    }
    ax_free(((ax_u8*)(text.ptr)));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ICONST, .type_id=((ax_u16)(13)), .dest=reg, .src1=val, .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_ident(struct ax_FuncLowering* self, struct ax_AstNode node) {
    ax_u32 name_id = node.payload;
    if ((name_id == ((ax_u32)(0)))) {
        ax_string text = ax_AirModuleBuilder_get_token_text(self->mb, node.token_idx);
        name_id = ax_InternPool_intern(&(self->mb->pool), text);
        ax_free(((ax_u8*)(text.ptr)));
    }
    if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
        struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
        if ((((sym.kind != ax_SYM_VAR) && (sym.kind != ax_SYM_PARAM)) && (sym.kind != ax_SYM_CONST))) {
            return ((ax_u32)(0));
        }
    }
    ax_u32 reg = ax_LocalMap_local_map_get(self->locals, name_id);
    return reg;
}

ax_u32 ax_FuncLowering_lower_binary_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 lhs_reg = ax_FuncLowering_lower_expr(self, child);
    ax_u32 rhs_idx = ((self->mb->tree.nodes.data)[((ax_i64)(child))]).next_sibling;
    if ((rhs_idx == ((ax_u32)(0)))) {
        return lhs_reg;
    }
    ax_u32 rhs_reg = ax_FuncLowering_lower_expr(self, rhs_idx);
    ax_string op_token = ax_AirModuleBuilder_get_token_text(self->mb, node.token_idx);
    ax_u16 opcode = ax_map_binary_op(op_token);
    ax_free(((ax_u8*)(op_token.ptr)));
    ax_u16 type_id = ((ax_u16)(3));
    if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
        struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
        if ((sym.type_id != ((ax_u32)(0)))) {
            type_id = ((ax_u16)(sym.type_id));
        }
    }
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=opcode, .type_id=type_id, .dest=reg, .src1=lhs_reg, .src2=rhs_reg}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_unary_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 operand_reg = ax_FuncLowering_lower_expr(self, child);
    ax_string op_token = ax_AirModuleBuilder_get_token_text(self->mb, node.token_idx);
    ax_u16 opcode = ax_OP_NOP;
    if (ax_str_eq(op_token, (ax_string){.ptr=(const ax_u8*)"-", .len=1})) {
        opcode = ax_OP_NEG;
    } else if ((ax_str_eq(op_token, (ax_string){.ptr=(const ax_u8*)"not", .len=3}) || ax_str_eq(op_token, (ax_string){.ptr=(const ax_u8*)"!", .len=1}))) {
        opcode = ax_OP_NOT;
    } else if (ax_str_eq(op_token, (ax_string){.ptr=(const ax_u8*)"~", .len=1})) {
        opcode = ax_OP_NOT;
    }
    ax_free(((ax_u8*)(op_token.ptr)));
    ax_u16 type_id = ((ax_u16)(3));
    if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
        struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
        if ((sym.type_id != ((ax_u32)(0)))) {
            type_id = ((ax_u16)(sym.type_id));
        }
    }
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=opcode, .type_id=type_id, .dest=reg, .src1=operand_reg, .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_call_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    struct ax_AstNode callee_node = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
    ax_bool is_method_call = AX_FALSE;
    ax_u32 method_sym_idx = ((ax_u32)(0));
    ax_u32 receiver_node_idx = ((ax_u32)(0));
    if ((callee_node.kind == ax_NODE_FIELD_EXPR)) {
        ax_bool resolved = AX_FALSE;
        if ((!resolved)) {
            receiver_node_idx = callee_node.first_child;
            if (((receiver_node_idx != ((ax_u32)(0))) && (self->mb->node_types != ((ax_u32*)(NULL))))) {
                ax_u32 rec_type = ((self->mb->node_types)[((ax_i64)(receiver_node_idx))]);
                if ((rec_type != ((ax_u32)(0)))) {
                    struct ax_TypeEntry entry = ((self->mb->typetable.entries.data)[((ax_i64)(rec_type))]);
                    if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                        rec_type = entry.extra;
                        entry = ((self->mb->typetable.entries.data)[((ax_i64)(rec_type))]);
                    }
                    if ((entry.kind == ax_TYPE_KIND_STRUCT)) {
                        ax_string method_name = ax_InternPool_get(self->mb->pool, callee_node.payload);
                        ax_i64 i = ((ax_i64)(0));
                        while ((i < self->mb->symbols.symbols.len)) {
                            struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[i]);
                            if ((sym.kind == ax_SYM_FUNC)) {
                                ax_string sym_name = ax_InternPool_get(self->mb->pool, sym.name_id);
                                if (ax_str_eq(sym_name, method_name)) {
                                    struct ax_TypeEntry t_entry = ((self->mb->typetable.entries.data)[((ax_i64)(sym.type_id))]);
                                    if ((t_entry.kind == ax_TYPE_KIND_FUNC)) {
                                        struct ax_FuncInfo fi = ((self->mb->typetable.funcs.data)[((ax_i64)(t_entry.extra))]);
                                        if ((fi.params.len > ((ax_i64)(0)))) {
                                            ax_u32 first_param = ((fi.params.data)[0]);
                                            struct ax_TypeEntry p_entry = ((self->mb->typetable.entries.data)[((ax_i64)(first_param))]);
                                            if ((p_entry.kind == ax_TYPE_KIND_POINTER)) {
                                                first_param = p_entry.extra;
                                            }
                                            if ((first_param == rec_type)) {
                                                is_method_call = AX_TRUE;
                                                method_sym_idx = ((ax_u32)(i));
                                                resolved = AX_TRUE;
                                                break;
                                            }
                                        }
                                    }
                                }
                            }
                            i = (i + ((ax_i64)(1)));
                        }
                    }
                }
            }
        }
    }
    if (((!is_method_call) && (callee_node.kind == ax_NODE_IDENT))) {
        ax_u32 sym_idx = callee_node.payload;
        if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->mb->symbols.symbols.len))) {
            struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(sym_idx))]);
            if ((sym.kind == ax_SYM_STRUCT)) {
                return ax_FuncLowering_lower_struct_constructor_call(self, idx, node, sym.type_id);
            }
        }
    }
    ax_u32* temp_args = ((ax_u32*)(ax_alloc((128 * 4))));
    ax_u32 temp_count = ((ax_u32)(0));
    if (is_method_call) {
        ax_u32 receiver_reg = ax_FuncLowering_lower_expr(self, receiver_node_idx);
        ((temp_args)[((ax_i64)(temp_count))]) = receiver_reg;
        temp_count = (temp_count + ((ax_u32)(1)));
    }
    ax_u32 arg = callee_node.next_sibling;
    while ((arg != ((ax_u32)(0)))) {
        ax_u32 arg_reg = ax_FuncLowering_lower_expr(self, arg);
        ((temp_args)[((ax_i64)(temp_count))]) = arg_reg;
        temp_count = (temp_count + ((ax_u32)(1)));
        arg = ((self->mb->tree.nodes.data)[((ax_i64)(arg))]).next_sibling;
    }
    ax_u32 arg_start = ax_AirFuncBuilder_emit_extra(&(self->fb), temp_count);
    ax_u32 i = ((ax_u32)(0));
    while ((i < temp_count)) {
        ax_AirFuncBuilder_emit_extra(&(self->fb), ((temp_args)[((ax_i64)(i))]));
        i = (i + ((ax_u32)(1)));
    }
    ax_free(((ax_u8*)(temp_args)));
    ax_u16 type_id = ((ax_u16)(0));
    ax_u32 callee_reg = ((ax_u32)(0));
    if (is_method_call) {
        type_id = ((ax_u16)(method_sym_idx));
        callee_reg = ((ax_u32)(0));
    } else {
        {
            callee_reg = ax_FuncLowering_lower_expr(self, child);
            if ((callee_node.kind == ax_NODE_IDENT)) {
                type_id = ((ax_u16)(callee_node.payload));
            } else {
                {
                    if (((self->mb->node_types != ((ax_u32*)(NULL))) && (((self->mb->node_types)[((ax_i64)(idx))]) != ((ax_u32)(0))))) {
                        type_id = ((ax_u16)(((self->mb->node_types)[((ax_i64)(idx))])));
                    } else {
                        {
                            if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
                                struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
                                if ((sym.type_id != ((ax_u32)(0)))) {
                                    type_id = ((ax_u16)(sym.type_id));
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    ax_u16 opcode = ax_OP_CALL;
    if (((!is_method_call) && (callee_node.kind == ax_NODE_IDENT))) {
        ax_u32 sym_idx = callee_node.payload;
        if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->mb->symbols.symbols.len))) {
            ax_u32 name_id = ((self->mb->symbols.symbols.data)[((ax_i64)(sym_idx))]).name_id;
            ax_string name = ax_InternPool_get(self->mb->pool, name_id);
            if (((((((ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall0", .len=8}) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall1", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall2", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall3", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall4", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall5", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"syscall6", .len=8}))) {
                opcode = ax_OP_SYSCALL;
            }
        }
    }
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=opcode, .type_id=type_id, .dest=reg, .src1=callee_reg, .src2=arg_start}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_field_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 obj_reg = ax_FuncLowering_lower_expr(self, child);
    ax_u32 field_idx = node.extra_idx;
    ax_u16 type_id = ((ax_u16)(0));
    ax_u32 obj_type = ((self->mb->node_types)[((ax_i64)(child))]);
    if ((obj_type != ((ax_u32)(0)))) {
        struct ax_TypeEntry entry = ((self->mb->typetable.entries.data)[((ax_i64)(obj_type))]);
        if ((entry.kind == ax_TYPE_KIND_POINTER)) {
            type_id = ((ax_u16)(entry.extra));
        } else {
            {
                type_id = ((ax_u16)(obj_type));
            }
        }
    }
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_GET_FIELD, .type_id=type_id, .dest=reg, .src1=obj_reg, .src2=field_idx}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_index_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 arr_reg = ax_FuncLowering_lower_expr(self, child);
    ax_u32 idx_expr = ((self->mb->tree.nodes.data)[((ax_i64)(child))]).next_sibling;
    ax_u32 idx_reg = ((ax_u32)(0));
    if ((idx_expr != ((ax_u32)(0)))) {
        idx_reg = ax_FuncLowering_lower_expr(self, idx_expr);
    }
    ax_u16 type_id = ((ax_u16)(0));
    if (((self->mb->node_types != ((ax_u32*)(NULL))) && (((self->mb->node_types)[((ax_i64)(idx))]) != ((ax_u32)(0))))) {
        type_id = ((ax_u16)(((self->mb->node_types)[((ax_i64)(idx))])));
    } else {
        {
            if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
                struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
                if ((sym.type_id != ((ax_u32)(0)))) {
                    type_id = ((ax_u16)(sym.type_id));
                }
            }
        }
    }
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_INDEX, .type_id=type_id, .dest=reg, .src1=arr_reg, .src2=idx_reg}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_cast_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 src_reg = ax_FuncLowering_lower_expr(self, child);
    ax_u16 type_id = ((ax_u16)(0));
    if (((self->mb->node_types != ((ax_u32*)(NULL))) && (((self->mb->node_types)[((ax_i64)(idx))]) != ((ax_u32)(0))))) {
        type_id = ((ax_u16)(((self->mb->node_types)[((ax_i64)(idx))])));
    } else {
        {
            if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
                struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
                if ((sym.type_id != ((ax_u32)(0)))) {
                    type_id = ((ax_u16)(sym.type_id));
                }
            }
        }
    }
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_CAST, .type_id=type_id, .dest=reg, .src1=src_reg, .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_deref_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 ptr_reg = ax_FuncLowering_lower_expr(self, child);
    ax_u16 type_id = ((ax_u16)(0));
    if (((self->mb->node_types != ((ax_u32*)(NULL))) && (((self->mb->node_types)[((ax_i64)(idx))]) != ((ax_u32)(0))))) {
        type_id = ((ax_u16)(((self->mb->node_types)[((ax_i64)(idx))])));
    } else {
        {
            if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
                struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
                if ((sym.type_id != ((ax_u32)(0)))) {
                    type_id = ((ax_u16)(sym.type_id));
                }
            }
        }
    }
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_DEREF, .type_id=type_id, .dest=reg, .src1=ptr_reg, .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_spawn_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 target_reg = ((ax_u32)(0));
    ax_u16 type_id = ((ax_u16)(0));
    struct ax_AstNode child_node = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
    if ((child_node.kind == ax_NODE_CALL_EXPR)) {
        ax_u32 callee_node_idx = child_node.first_child;
        if ((callee_node_idx != ((ax_u32)(0)))) {
            target_reg = ax_FuncLowering_lower_expr(self, callee_node_idx);
            struct ax_AstNode callee_node = ((self->mb->tree.nodes.data)[((ax_i64)(callee_node_idx))]);
            if ((callee_node.kind == ax_NODE_IDENT)) {
                type_id = ((ax_u16)(callee_node.payload));
            }
        }
    } else {
        {
            target_reg = ax_FuncLowering_lower_expr(self, child);
        }
    }
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_SPAWN, .type_id=type_id, .dest=reg, .src1=target_reg, .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_await_expr(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return ((ax_u32)(0));
    }
    ax_u32 future_reg = ax_FuncLowering_lower_expr(self, child);
    ax_u16 type_id = ((ax_u16)(0));
    if (((self->mb->node_types != ((ax_u32*)(NULL))) && (((self->mb->node_types)[((ax_i64)(idx))]) != ((ax_u32)(0))))) {
        type_id = ((ax_u16)(((self->mb->node_types)[((ax_i64)(idx))])));
    } else {
        {
            if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
                struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
                if ((sym.type_id != ((ax_u32)(0)))) {
                    type_id = ((ax_u16)(sym.type_id));
                }
            }
        }
    }
    ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_AWAIT, .type_id=type_id, .dest=reg, .src1=future_reg, .src2=((ax_u32)(0))}));
    return reg;
}

ax_u32 ax_FuncLowering_lower_struct_lit(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u16 type_id = ((ax_u16)(0));
    if (((self->mb->node_types != ((ax_u32*)(NULL))) && (((self->mb->node_types)[((ax_i64)(idx))]) != ((ax_u32)(0))))) {
        type_id = ((ax_u16)(((self->mb->node_types)[((ax_i64)(idx))])));
    } else {
        {
            if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
                struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
                if ((sym.type_id != ((ax_u32)(0)))) {
                    type_id = ((ax_u16)(sym.type_id));
                }
            }
        }
    }
    ax_u32 struct_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ALLOC, .type_id=type_id, .dest=struct_reg, .src1=((ax_u32)(0)), .src2=((ax_u32)(0))}));
    ax_u32 child = node.first_child;
    ax_u32 field_idx = ((ax_u32)(0));
    while ((child != ((ax_u32)(0)))) {
        struct ax_AstNode cn = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
        if ((cn.kind == ax_NODE_NAMED_ARG)) {
            ax_u32 val_child = cn.first_child;
            if ((val_child != ((ax_u32)(0)))) {
                ax_u32 val_reg = ax_FuncLowering_lower_expr(self, val_child);
                ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_SET_FIELD, .type_id=type_id, .dest=val_reg, .src1=struct_reg, .src2=field_idx}));
            }
            field_idx = (field_idx + ((ax_u32)(1)));
        }
        child = cn.next_sibling;
    }
    return struct_reg;
}

ax_u32 ax_FuncLowering_lower_array_lit(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u16 type_id = ((ax_u16)(0));
    if (((self->mb->node_types != ((ax_u32*)(NULL))) && (((self->mb->node_types)[((ax_i64)(idx))]) != ((ax_u32)(0))))) {
        type_id = ((ax_u16)(((self->mb->node_types)[((ax_i64)(idx))])));
    } else {
        {
            if (((node.payload != ((ax_u32)(0))) && (((ax_i64)(node.payload)) < self->mb->symbols.symbols.len))) {
                struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(node.payload))]);
                if ((sym.type_id != ((ax_u32)(0)))) {
                    type_id = ((ax_u16)(sym.type_id));
                }
            }
        }
    }
    ax_u32 arr_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ALLOC, .type_id=type_id, .dest=arr_reg, .src1=((ax_u32)(0)), .src2=((ax_u32)(0))}));
    ax_u32 child = node.first_child;
    ax_u32 elem_idx = ((ax_u32)(0));
    while ((child != ((ax_u32)(0)))) {
        ax_u32 elem_reg = ax_FuncLowering_lower_expr(self, child);
        ax_u32 idx_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ICONST, .type_id=((ax_u16)(3)), .dest=idx_reg, .src1=elem_idx, .src2=((ax_u32)(0))}));
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_STORE, .type_id=((ax_u16)(0)), .dest=arr_reg, .src1=elem_reg, .src2=idx_reg}));
        elem_idx = (elem_idx + ((ax_u32)(1)));
        child = ((self->mb->tree.nodes.data)[((ax_i64)(child))]).next_sibling;
    }
    return arr_reg;
}

ax_u32 ax_FuncLowering_lower_struct_constructor_call(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node, ax_u32 type_id) {
    ax_u32 struct_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ALLOC, .type_id=((ax_u16)(type_id)), .dest=struct_reg, .src1=((ax_u32)(0)), .src2=((ax_u32)(0))}));
    ax_u32 callee = node.first_child;
    if ((callee == ((ax_u32)(0)))) {
        return struct_reg;
    }
    ax_u32 child = ((self->mb->tree.nodes.data)[((ax_i64)(callee))]).next_sibling;
    ax_u32 field_idx = ((ax_u32)(0));
    while ((child != ((ax_u32)(0)))) {
        struct ax_AstNode cn = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
        if ((cn.kind == ax_NODE_NAMED_ARG)) {
            ax_u32 val_child = cn.first_child;
            if ((val_child != ((ax_u32)(0)))) {
                ax_u32 val_reg = ax_FuncLowering_lower_expr(self, val_child);
                ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_SET_FIELD, .type_id=((ax_u16)(type_id)), .dest=val_reg, .src1=struct_reg, .src2=field_idx}));
            }
            field_idx = (field_idx + ((ax_u32)(1)));
        } else {
            {
                ax_u32 val_reg = ax_FuncLowering_lower_expr(self, child);
                ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_SET_FIELD, .type_id=((ax_u16)(type_id)), .dest=val_reg, .src1=struct_reg, .src2=field_idx}));
                field_idx = (field_idx + ((ax_u32)(1)));
            }
        }
        child = cn.next_sibling;
    }
    return struct_reg;
}

ax_u32 ax_FuncLowering_emit_heap_alloc(struct ax_FuncLowering* self, ax_u16 type_id) {
    ax_u32 ptr_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ALLOC, .type_id=type_id, .dest=ptr_reg, .src1=((ax_u32)(0)), .src2=((ax_u32)(0))}));
    ax_u32 ref_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_MAKE_REF, .type_id=type_id, .dest=ref_reg, .src1=ptr_reg, .src2=((ax_u32)(0))}));
    return ref_reg;
}

ax_u32 ax_FuncLowering_emit_deref(struct ax_FuncLowering* self, ax_u32 ref_reg, ax_u16 type_id) {
    ax_u32 ptr_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_DEREF, .type_id=type_id, .dest=ptr_reg, .src1=ref_reg, .src2=((ax_u32)(0))}));
    return ptr_reg;
}

ax_u32 ax_FuncLowering_emit_move(struct ax_FuncLowering* self, ax_u32 src_reg, ax_u16 type_id) {
    ax_u32 dest_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_MOVE, .type_id=type_id, .dest=dest_reg, .src1=src_reg, .src2=((ax_u32)(0))}));
    return dest_reg;
}

void ax_FuncLowering_emit_free(struct ax_FuncLowering* self, ax_u32 ptr_reg) {
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_FREE, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=ptr_reg, .src2=((ax_u32)(0))}));
}

ax_u32 ax_FuncLowering_emit_arena_alloc(struct ax_FuncLowering* self, ax_u32 arena_reg, ax_u16 type_id) {
    ax_u32 ptr_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ARENA_ALLOC, .type_id=type_id, .dest=ptr_reg, .src1=arena_reg, .src2=((ax_u32)(0))}));
    return ptr_reg;
}

ax_u32 ax_FuncLowering_emit_alias_reuse(struct ax_FuncLowering* self, ax_u32 ptr_reg, ax_u16 type_id) {
    ax_u32 dest_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ALIAS_REUSE, .type_id=type_id, .dest=dest_reg, .src1=ptr_reg, .src2=((ax_u32)(0))}));
    return dest_reg;
}

ax_u32 ax_FuncLowering_lower_ownership_aware(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node, ax_u32 init_reg) {
    if (((node.flags & ax_FLAG_ESCAPES_TO_HEAP) != ((ax_u16)(0)))) {
        ax_u16 type_id = ((ax_u16)(0));
        if ((node.extra_idx != ((ax_u32)(0)))) {
            type_id = ((ax_u16)(node.extra_idx));
        }
        ax_u32 ref_reg = ax_FuncLowering_emit_heap_alloc(self, type_id);
        if ((init_reg != ((ax_u32)(0)))) {
            ax_u32 ptr_reg = ax_FuncLowering_emit_deref(self, ref_reg, type_id);
            ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_STORE, .type_id=((ax_u16)(0)), .dest=ptr_reg, .src1=init_reg, .src2=((ax_u32)(0))}));
        }
        return ref_reg;
    }
    if (((node.flags & ax_FLAG_USES_ARENA) != ((ax_u16)(0)))) {
        ax_u16 type_id = ((ax_u16)(0));
        if ((node.extra_idx != ((ax_u32)(0)))) {
            type_id = ((ax_u16)(node.extra_idx));
        }
        ax_u32 ptr_reg = ax_FuncLowering_emit_arena_alloc(self, ((ax_u32)(0)), type_id);
        if ((init_reg != ((ax_u32)(0)))) {
            ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_STORE, .type_id=((ax_u16)(0)), .dest=ptr_reg, .src1=init_reg, .src2=((ax_u32)(0))}));
        }
        return ptr_reg;
    }
    if ((((node.flags & ax_FLAG_IS_MOVED) != ((ax_u16)(0))) && (init_reg != ((ax_u32)(0))))) {
        ax_u16 type_id = ((ax_u16)(0));
        if ((node.extra_idx != ((ax_u32)(0)))) {
            type_id = ((ax_u16)(node.extra_idx));
        }
        return ax_FuncLowering_emit_move(self, init_reg, type_id);
    }
    return init_reg;
}

void ax_FuncLowering_lower_alias_stmt(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    if ((node.first_child != ((ax_u32)(0)))) {
        struct ax_AstNode cn = ((self->mb->tree.nodes.data)[((ax_i64)(node.first_child))]);
        ax_u32 src_reg = ax_FuncLowering_lower_expr(self, node.first_child);
        ax_u16 type_id = ((ax_u16)(0));
        if ((node.extra_idx != ((ax_u32)(0)))) {
            type_id = ((ax_u16)(node.extra_idx));
        }
        ax_FuncLowering_emit_alias_reuse(self, src_reg, type_id);
    }
}

void ax_FuncLowering_lower_block(struct ax_FuncLowering* self, ax_u32 block_idx) {
    struct ax_AstNode node = ((self->mb->tree.nodes.data)[((ax_i64)(block_idx))]);
    ax_u32 child = node.first_child;
    while ((child != ((ax_u32)(0)))) {
        struct ax_AstNode cn = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
        ax_FuncLowering_lower_stmt(self, child, cn);
        child = cn.next_sibling;
    }
}

void ax_FuncLowering_lower_stmt(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] lower_stmt: idx=%d, kind=%d\n", .len=36}).ptr, ((ax_i64)(idx)), ((ax_i64)(node.kind)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    if (self->terminated) {
        return;
    }
    if ((node.kind == ax_NODE_VAR_DECL)) {
        ax_FuncLowering_lower_var_decl(self, idx, node);
    } else if ((node.kind == ax_NODE_ASSIGN_STMT)) {
        ax_FuncLowering_lower_assign(self, idx, node);
    } else if ((node.kind == ax_NODE_RETURN_STMT)) {
        ax_FuncLowering_lower_return(self, idx, node);
    } else if ((node.kind == ax_NODE_IF_STMT)) {
        ax_FuncLowering_lower_if(self, idx, node);
    } else if ((node.kind == ax_NODE_WHILE_STMT)) {
        ax_FuncLowering_lower_while(self, idx, node);
    } else if ((node.kind == ax_NODE_FOR_STMT)) {
        ax_FuncLowering_lower_for(self, idx, node);
    } else if ((node.kind == ax_NODE_BLOCK)) {
        ax_FuncLowering_lower_block(self, idx);
    } else if ((node.kind == ax_NODE_DESTROY_STMT)) {
        ax_FuncLowering_lower_destroy(self, idx, node);
    } else if ((node.kind == ax_NODE_ALIAS_STMT)) {
        ax_FuncLowering_lower_alias_stmt(self, idx, node);
    } else if ((node.kind == ax_NODE_DEFER_STMT)) {
        ax_FuncLowering_lower_defer(self, idx, node);
    } else {
        {
            ax_FuncLowering_lower_expr(self, idx);
        }
    }
}

void ax_FuncLowering_lower_var_decl(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 name_id = node.payload;
    ax_u16 type_id = ((ax_u16)(0));
    ax_u32 init_reg = ((ax_u32)(0));
    ax_u32 child = node.first_child;
    while ((child != ((ax_u32)(0)))) {
        struct ax_AstNode cn = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
        if (((((((cn.kind != ax_NODE_TYPE_EXPR) && (cn.kind != ax_NODE_PTR_TYPE)) && (cn.kind != ax_NODE_SLICE_TYPE)) && (cn.kind != ax_NODE_ARRAY_TYPE)) && (cn.kind != ax_NODE_FUNC_TYPE)) && (cn.kind != ax_NODE_GENERIC_TYPE))) {
            init_reg = ax_FuncLowering_lower_expr(self, child);
        }
        child = cn.next_sibling;
    }
    if ((init_reg != ((ax_u32)(0)))) {
        if (((node.flags & ((ax_FLAG_ESCAPES_TO_HEAP | ax_FLAG_USES_ARENA) | ax_FLAG_IS_MOVED)) != ((ax_u16)(0)))) {
            ax_u32 owner_reg = ax_FuncLowering_lower_ownership_aware(self, idx, node, init_reg);
            ax_LocalMap_local_map_put(&(self->locals), name_id, owner_reg);
        } else {
            {
                ax_LocalMap_local_map_put(&(self->locals), name_id, init_reg);
            }
        }
    } else {
        {
            ax_u32 reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
            ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ICONST, .type_id=type_id, .dest=reg, .src1=((ax_u32)(0)), .src2=((ax_u32)(0))}));
            ax_LocalMap_local_map_put(&(self->locals), name_id, reg);
        }
    }
}

void ax_FuncLowering_lower_assign(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 lhs_idx = node.first_child;
    if ((lhs_idx == ((ax_u32)(0)))) {
        return;
    }
    struct ax_AstNode lhs_node = ((self->mb->tree.nodes.data)[((ax_i64)(lhs_idx))]);
    if ((lhs_node.kind == ax_NODE_FIELD_EXPR)) {
        ax_u32 obj_idx = lhs_node.first_child;
        ax_u32 obj_reg = ax_FuncLowering_lower_expr(self, obj_idx);
        ax_u32 field_idx = lhs_node.extra_idx;
        ax_u32 rhs_idx = lhs_node.next_sibling;
        if ((rhs_idx == ((ax_u32)(0)))) {
            return;
        }
        ax_u32 val_reg = ax_FuncLowering_lower_expr(self, rhs_idx);
        ax_u16 type_id = ((ax_u16)(0));
        ax_u32 obj_type = ((self->mb->node_types)[((ax_i64)(obj_idx))]);
        if ((obj_type != ((ax_u32)(0)))) {
            struct ax_TypeEntry entry = ((self->mb->typetable.entries.data)[((ax_i64)(obj_type))]);
            if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                type_id = ((ax_u16)(entry.extra));
            } else {
                {
                    type_id = ((ax_u16)(obj_type));
                }
            }
        }
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_SET_FIELD, .type_id=type_id, .dest=val_reg, .src1=obj_reg, .src2=field_idx}));
        return;
    } else if ((lhs_node.kind == ax_NODE_INDEX_EXPR)) {
        ax_u32 arr_idx = lhs_node.first_child;
        ax_u32 arr_reg = ax_FuncLowering_lower_expr(self, arr_idx);
        ax_u32 index_idx = ((self->mb->tree.nodes.data)[((ax_i64)(arr_idx))]).next_sibling;
        ax_u32 index_reg = ((ax_u32)(0));
        if ((index_idx != ((ax_u32)(0)))) {
            index_reg = ax_FuncLowering_lower_expr(self, index_idx);
        }
        ax_u32 rhs_idx = lhs_node.next_sibling;
        if ((rhs_idx == ((ax_u32)(0)))) {
            return;
        }
        ax_u32 val_reg = ax_FuncLowering_lower_expr(self, rhs_idx);
        ax_u16 type_id = ((ax_u16)(0));
        ax_u32 arr_type = ((self->mb->node_types)[((ax_i64)(arr_idx))]);
        if ((arr_type != ((ax_u32)(0)))) {
            struct ax_TypeEntry entry = ((self->mb->typetable.entries.data)[((ax_i64)(arr_type))]);
            if ((((entry.kind == ax_TYPE_KIND_POINTER) || (entry.kind == ax_TYPE_KIND_SLICE)) || (entry.kind == ax_TYPE_KIND_ARRAY))) {
                type_id = ((ax_u16)(entry.extra));
            }
        }
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_STORE, .type_id=type_id, .dest=arr_reg, .src1=val_reg, .src2=index_reg}));
        return;
    } else if ((lhs_node.kind == ax_NODE_DEREF_EXPR)) {
        ax_u32 ptr_idx = lhs_node.first_child;
        ax_u32 ptr_reg = ax_FuncLowering_lower_expr(self, ptr_idx);
        ax_u32 rhs_idx = lhs_node.next_sibling;
        if ((rhs_idx == ((ax_u32)(0)))) {
            return;
        }
        ax_u32 val_reg = ax_FuncLowering_lower_expr(self, rhs_idx);
        ax_u16 type_id = ((ax_u16)(0));
        ax_u32 ptr_type = ((self->mb->node_types)[((ax_i64)(ptr_idx))]);
        if ((ptr_type != ((ax_u32)(0)))) {
            struct ax_TypeEntry entry = ((self->mb->typetable.entries.data)[((ax_i64)(ptr_type))]);
            if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                type_id = ((ax_u16)(entry.extra));
            }
        }
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_STORE, .type_id=type_id, .dest=ptr_reg, .src1=val_reg, .src2=((ax_u32)(0))}));
        return;
    }
    ax_u32 name_id = lhs_node.payload;
    if ((name_id == ((ax_u32)(0)))) {
        ax_string text = ax_AirModuleBuilder_get_token_text(self->mb, lhs_node.token_idx);
        name_id = ax_InternPool_intern(&(self->mb->pool), text);
        ax_free(((ax_u8*)(text.ptr)));
    }
    ax_u32 rhs_idx = lhs_node.next_sibling;
    if ((rhs_idx == ((ax_u32)(0)))) {
        return;
    }
    struct ax_AstNode rhs_node = ((self->mb->tree.nodes.data)[((ax_i64)(rhs_idx))]);
    ax_u32 val_reg = ax_FuncLowering_lower_expr(self, rhs_idx);
    ax_u16 type_id = ((ax_u16)(3));
    if (((lhs_node.payload != ((ax_u32)(0))) && (((ax_i64)(lhs_node.payload)) < self->mb->symbols.symbols.len))) {
        struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(lhs_node.payload))]);
        if ((sym.type_id != ((ax_u32)(0)))) {
            type_id = ((ax_u16)(sym.type_id));
        }
    }
    ax_u32 existing_reg = ax_LocalMap_local_map_get(self->locals, name_id);
    if ((existing_reg != ((ax_u32)(0)))) {
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_COPY, .type_id=type_id, .dest=existing_reg, .src1=val_reg, .src2=((ax_u32)(0))}));
    } else {
        {
            ax_LocalMap_local_map_put(&(self->locals), name_id, val_reg);
        }
    }
}

void ax_FuncLowering_lower_return(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 ret_val = ((ax_u32)(0));
    if ((node.first_child != ((ax_u32)(0)))) {
        ret_val = ax_FuncLowering_lower_expr(self, node.first_child);
    }
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_RETURN, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=ret_val, .src2=((ax_u32)(0))}));
    self->terminated = AX_TRUE;
}

void ax_FuncLowering_lower_if(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 then_block = ax_AirFuncBuilder_new_block(&(self->fb));
    ax_u32 else_block = ax_AirFuncBuilder_new_block(&(self->fb));
    ax_u32 merge_block = ax_AirFuncBuilder_new_block(&(self->fb));
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return;
    }
    struct ax_AstNode cn = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
    ax_u32 cond_reg = ax_FuncLowering_lower_expr(self, child);
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_BRANCH, .type_id=((ax_u16)(0)), .dest=else_block, .src1=cond_reg, .src2=then_block}));
    ax_u32 cur_block = ax_AirFuncBuilder_current_block(self->fb);
    ax_AirFuncBuilder_add_edge(&(self->fb), cur_block, then_block);
    ax_AirFuncBuilder_add_edge(&(self->fb), cur_block, else_block);
    ax_AirFuncBuilder_switch_to(&(self->fb), then_block);
    self->terminated = AX_FALSE;
    child = cn.next_sibling;
    if ((child != ((ax_u32)(0)))) {
        struct ax_AstNode tcn = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
        if ((tcn.kind == ax_NODE_BLOCK)) {
            ax_FuncLowering_lower_block(self, child);
        }
        child = tcn.next_sibling;
    }
    if ((!self->terminated)) {
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_JUMP, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=merge_block, .src2=((ax_u32)(0))}));
        ax_AirFuncBuilder_add_edge(&(self->fb), then_block, merge_block);
    }
    ax_AirFuncBuilder_switch_to(&(self->fb), else_block);
    self->terminated = AX_FALSE;
    while ((child != ((ax_u32)(0)))) {
        struct ax_AstNode ecn = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
        if (((ecn.kind == ax_NODE_ELSE_CLAUSE) || (ecn.kind == ax_NODE_ELIF_CLAUSE))) {
            ax_u32 else_body = ecn.first_child;
            if ((else_body != ((ax_u32)(0)))) {
                struct ax_AstNode eecn = ((self->mb->tree.nodes.data)[((ax_i64)(else_body))]);
                if ((eecn.kind == ax_NODE_BLOCK)) {
                    ax_FuncLowering_lower_block(self, else_body);
                } else if ((ecn.kind == ax_NODE_ELIF_CLAUSE)) {
                    ax_FuncLowering_lower_if(self, child, ecn);
                }
            }
        }
        child = ecn.next_sibling;
    }
    if ((!self->terminated)) {
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_JUMP, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=merge_block, .src2=((ax_u32)(0))}));
        ax_AirFuncBuilder_add_edge(&(self->fb), else_block, merge_block);
    }
    ax_AirFuncBuilder_switch_to(&(self->fb), merge_block);
    self->terminated = AX_FALSE;
}

void ax_FuncLowering_lower_while(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 cond_block = ax_AirFuncBuilder_new_block(&(self->fb));
    ax_u32 body_block = ax_AirFuncBuilder_new_block(&(self->fb));
    ax_u32 exit_block = ax_AirFuncBuilder_new_block(&(self->fb));
    ax_u32 cur_block = ax_AirFuncBuilder_current_block(self->fb);
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_JUMP, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=cond_block, .src2=((ax_u32)(0))}));
    ax_AirFuncBuilder_add_edge(&(self->fb), cur_block, cond_block);
    ax_AirFuncBuilder_switch_to(&(self->fb), cond_block);
    self->terminated = AX_FALSE;
    ax_u32 child = node.first_child;
    if ((child == ((ax_u32)(0)))) {
        return;
    }
    struct ax_AstNode cn = ((self->mb->tree.nodes.data)[((ax_i64)(child))]);
    ax_u32 cond_reg = ax_FuncLowering_lower_expr(self, child);
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_BRANCH, .type_id=((ax_u16)(0)), .dest=exit_block, .src1=cond_reg, .src2=body_block}));
    ax_AirFuncBuilder_add_edge(&(self->fb), cond_block, body_block);
    ax_AirFuncBuilder_add_edge(&(self->fb), cond_block, exit_block);
    ax_AirFuncBuilder_switch_to(&(self->fb), body_block);
    self->terminated = AX_FALSE;
    ax_u32 body_child = cn.next_sibling;
    if ((body_child != ((ax_u32)(0)))) {
        struct ax_AstNode bcn = ((self->mb->tree.nodes.data)[((ax_i64)(body_child))]);
        if ((bcn.kind == ax_NODE_BLOCK)) {
            ax_FuncLowering_lower_block(self, body_child);
        }
    }
    if ((!self->terminated)) {
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_JUMP, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=cond_block, .src2=((ax_u32)(0))}));
        ax_AirFuncBuilder_add_edge(&(self->fb), body_block, cond_block);
    }
    ax_AirFuncBuilder_switch_to(&(self->fb), exit_block);
    self->terminated = AX_FALSE;
}

void ax_FuncLowering_lower_for(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    ax_u32 cond_block = ax_AirFuncBuilder_new_block(&(self->fb));
    ax_u32 body_block = ax_AirFuncBuilder_new_block(&(self->fb));
    ax_u32 exit_block = ax_AirFuncBuilder_new_block(&(self->fb));
    ax_u32 sym_idx = node.payload;
    ax_u32 name_id = ((ax_u32)(0));
    ax_u16 type_id = ((ax_u16)(3));
    if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->mb->symbols.symbols.len))) {
        struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(sym_idx))]);
        name_id = sym.name_id;
        if ((sym.type_id != ((ax_u32)(0)))) {
            type_id = ((ax_u16)(sym.type_id));
        }
    } else {
        {
            name_id = sym_idx;
        }
    }
    ax_u32 range_expr_idx = node.first_child;
    ax_u32 start_reg = ((ax_u32)(0));
    ax_u32 limit_reg = ((ax_u32)(0));
    ax_bool is_range = AX_FALSE;
    if ((range_expr_idx != ((ax_u32)(0)))) {
        struct ax_AstNode range_node = ((self->mb->tree.nodes.data)[((ax_i64)(range_expr_idx))]);
        if ((range_node.kind == ax_NODE_BINARY_EXPR)) {
            ax_string op_text = ax_AirModuleBuilder_get_token_text(self->mb, range_node.token_idx);
            if (ax_str_eq(op_text, (ax_string){.ptr=(const ax_u8*)"..", .len=2})) {
                is_range = AX_TRUE;
                ax_u32 start_node_idx = range_node.first_child;
                if ((start_node_idx != ((ax_u32)(0)))) {
                    start_reg = ax_FuncLowering_lower_expr(self, start_node_idx);
                    struct ax_AstNode start_node = ((self->mb->tree.nodes.data)[((ax_i64)(start_node_idx))]);
                    ax_u32 end_node_idx = start_node.next_sibling;
                    if ((end_node_idx != ((ax_u32)(0)))) {
                        limit_reg = ax_FuncLowering_lower_expr(self, end_node_idx);
                    }
                }
            }
            ax_free(((ax_u8*)(op_text.ptr)));
        }
    }
    ax_u32 iter_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    if (is_range) {
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_COPY, .type_id=type_id, .dest=iter_reg, .src1=start_reg, .src2=((ax_u32)(0))}));
    } else {
        {
            ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ICONST, .type_id=type_id, .dest=iter_reg, .src1=((ax_u32)(0)), .src2=((ax_u32)(0))}));
            if ((range_expr_idx != ((ax_u32)(0)))) {
                limit_reg = ax_FuncLowering_lower_expr(self, range_expr_idx);
            }
        }
    }
    ax_LocalMap_local_map_put(&(self->locals), name_id, iter_reg);
    ax_u32 cur_block = ax_AirFuncBuilder_current_block(self->fb);
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_JUMP, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=cond_block, .src2=((ax_u32)(0))}));
    ax_AirFuncBuilder_add_edge(&(self->fb), cur_block, cond_block);
    ax_AirFuncBuilder_switch_to(&(self->fb), cond_block);
    self->terminated = AX_FALSE;
    ax_u32 cmp_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
    ax_u32 current_iter = ax_LocalMap_local_map_get(self->locals, name_id);
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_LT, .type_id=((ax_u16)(11)), .dest=cmp_reg, .src1=current_iter, .src2=limit_reg}));
    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_BRANCH, .type_id=((ax_u16)(0)), .dest=exit_block, .src1=cmp_reg, .src2=body_block}));
    ax_AirFuncBuilder_add_edge(&(self->fb), cond_block, body_block);
    ax_AirFuncBuilder_add_edge(&(self->fb), cond_block, exit_block);
    ax_AirFuncBuilder_switch_to(&(self->fb), body_block);
    self->terminated = AX_FALSE;
    if ((range_expr_idx != ((ax_u32)(0)))) {
        struct ax_AstNode range_node = ((self->mb->tree.nodes.data)[((ax_i64)(range_expr_idx))]);
        ax_u32 body_node_idx = range_node.next_sibling;
        if ((body_node_idx != ((ax_u32)(0)))) {
            struct ax_AstNode body_node = ((self->mb->tree.nodes.data)[((ax_i64)(body_node_idx))]);
            if ((body_node.kind == ax_NODE_BLOCK)) {
                ax_FuncLowering_lower_block(self, body_node_idx);
            }
        }
    }
    if ((!self->terminated)) {
        ax_u32 one_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_ICONST, .type_id=type_id, .dest=one_reg, .src1=((ax_u32)(1)), .src2=((ax_u32)(0))}));
        ax_u32 current_iter2 = ax_LocalMap_local_map_get(self->locals, name_id);
        ax_u32 new_val_reg = ax_AirFuncBuilder_fresh_reg(&(self->fb));
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_IADD, .type_id=type_id, .dest=new_val_reg, .src1=current_iter2, .src2=one_reg}));
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_COPY, .type_id=type_id, .dest=current_iter2, .src1=new_val_reg, .src2=((ax_u32)(0))}));
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_JUMP, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=cond_block, .src2=((ax_u32)(0))}));
        ax_AirFuncBuilder_add_edge(&(self->fb), body_block, cond_block);
    }
    ax_AirFuncBuilder_switch_to(&(self->fb), exit_block);
    self->terminated = AX_FALSE;
}

void ax_FuncLowering_lower_destroy(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    if ((node.first_child != ((ax_u32)(0)))) {
        ax_u32 val_reg = ax_FuncLowering_lower_expr(self, node.first_child);
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_DESTROY, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=val_reg, .src2=((ax_u32)(0))}));
    } else {
        {
            ax_u32 sym_id = node.payload;
            if (((sym_id != ((ax_u32)(0))) && (((ax_i64)(sym_id)) < self->mb->symbols.symbols.len))) {
                struct ax_Symbol sym = ((self->mb->symbols.symbols.data)[((ax_i64)(sym_id))]);
                ax_u32 reg = ax_LocalMap_local_map_get(self->locals, sym.name_id);
                if ((reg != ((ax_u32)(0)))) {
                    ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_DESTROY, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=reg, .src2=((ax_u32)(0))}));
                }
            }
        }
    }
}

void ax_FuncLowering_lower_defer(struct ax_FuncLowering* self, ax_u32 idx, struct ax_AstNode node) {
    if ((node.first_child != ((ax_u32)(0)))) {
        ax_FuncLowering_lower_expr(self, node.first_child);
    }
}

void ax_FuncLowering_ensure_return(struct ax_FuncLowering* self) {
    if ((!self->terminated)) {
        ax_AirFuncBuilder_emit(&(self->fb), ((struct ax_AirInst){.opcode=ax_OP_RETURN, .type_id=((ax_u16)(0)), .dest=((ax_u32)(0)), .src1=((ax_u32)(0)), .src2=((ax_u32)(0))}));
        self->terminated = AX_TRUE;
    }
}

struct ax_AirFunc* ax_AirModuleBuilder_lower_func(struct ax_AirModuleBuilder* self, ax_u32 idx, struct ax_AstNode node) {
    if (((node.flags & ax_FLAG_IS_EXTERN) != ((ax_u16)(0)))) {
        return ((struct ax_AirFunc*)(NULL));
    }
    ax_u32 name_id = ((ax_u32)(0));
    ax_u32 ret_type_id = ((ax_u32)(0));
    struct ax_U32Vec param_types = ax_new_u32_vec();
    ax_u32 sym_idx = node.payload;
    if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symbols.symbols.len))) {
        struct ax_Symbol sym = ((self->symbols.symbols.data)[((ax_i64)(sym_idx))]);
        name_id = sym.name_id;
        if ((sym.type_id != ((ax_u32)(0)))) {
            struct ax_TypeEntry entry = ((self->typetable.entries.data)[((ax_i64)(sym.type_id))]);
            if ((entry.kind == ax_TYPE_KIND_FUNC)) {
                struct ax_FuncInfo fi = ((self->typetable.funcs.data)[((ax_i64)(entry.extra))]);
                ret_type_id = fi.ret;
                ax_i64 i = ((ax_i64)(0));
                while ((i < fi.params.len)) {
                    ax_U32Vec_push(&(param_types), ((fi.params.data)[i]));
                    i = (i + 1);
                }
            }
        }
    }
    if ((name_id == ((ax_u32)(0)))) {
        ax_string text = ax_AirModuleBuilder_get_token_text(self, node.token_idx);
        name_id = ax_InternPool_intern(&(self->pool), text);
        ax_free(((ax_u8*)(text.ptr)));
    }
    struct ax_AirFuncBuilder fb = ax_new_air_func_builder(name_id, ret_type_id);
    struct ax_FuncLowering fl = ax_AirModuleBuilder_new_func_lowering(self, fb, param_types);
    ax_FuncLowering_register_params(&(fl), idx, node);
    ax_u32 body_idx = ((ax_u32)(0));
    ax_u32 child = node.first_child;
    while ((child != ((ax_u32)(0)))) {
        struct ax_AstNode cn = ((self->tree.nodes.data)[((ax_i64)(child))]);
        if ((cn.kind == ax_NODE_BLOCK)) {
            body_idx = child;
        }
        child = cn.next_sibling;
    }
    if ((body_idx != ((ax_u32)(0)))) {
        ax_FuncLowering_lower_block(&(fl), body_idx);
    }
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] lower_func: block lowered, ensuring return\n", .len=51}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    ax_FuncLowering_ensure_return(&(fl));
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] lower_func: building func\n", .len=34}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    struct ax_AirFunc fn_built = ax_AirFuncBuilder_build_func(&(fl.fb));
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] lower_func: fn built, allocating fn_ptr\n", .len=48}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    struct ax_AirFunc* fn_ptr = ((struct ax_AirFunc*)(ax_alloc(192)));
    ((fn_ptr)[0]) = fn_built;
    fn_ptr->sym_id = sym_idx;
    fn_ptr->is_async = ((node.flags & ax_FLAG_IS_ASYNC) != ((ax_u16)(0)));
    fn_ptr->params = fl.params;
    fl.params.data = ((ax_u32*)(NULL));
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] lower_func: freeing fl\n", .len=31}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    ax_FuncLowering_free_func_lowering(&(fl));
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] lower_func: returning fn_ptr\n", .len=37}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    return fn_ptr;
    ax_free(fn_ptr);
}

void ax_AirModuleBuilder_build_module(struct ax_AirModuleBuilder* self) {
    struct ax_AstNode root = ((self->tree.nodes.data)[0]);
    ax_u32 child = root.first_child;
    while ((child != ((ax_u32)(0)))) {
        struct ax_AstNode node = ((self->tree.nodes.data)[((ax_i64)(child))]);
        if ((node.kind == ax_NODE_FUNC_DECL)) {
            ax_string name = ax_AirModuleBuilder_get_token_text(self, node.token_idx);
            ax_u8* name_ptr = ax_AirModuleBuilder_builder_str_to_null_terminated(self, name);
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] AIR Builder: lowering top-level func %s (nodeIdx=%d)\n", .len=61}).ptr, ((ax_i64)(name_ptr)), ((ax_i64)(child)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_free(name_ptr);
            struct ax_AirFunc* fn_ptr = ax_AirModuleBuilder_lower_func(self, child, node);
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] build_module: lower_func returned\n", .len=42}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            if ((fn_ptr != ((struct ax_AirFunc*)(NULL)))) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] build_module: pushing fn_built\n", .len=39}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                ax_AirFuncVec_push(&(self->module.funcs), ((fn_ptr)[0]));
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] build_module: pushing done, freeing fn_ptr\n", .len=51}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                ax_free(((ax_u8*)(fn_ptr)));
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] build_module: freeing done\n", .len=35}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
            }
        } else if ((node.kind == ax_NODE_STRUCT_DECL)) {
            ax_string struct_name = ax_AirModuleBuilder_get_token_text(self, node.token_idx);
            ax_u8* struct_name_ptr = ax_AirModuleBuilder_builder_str_to_null_terminated(self, struct_name);
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] AIR Builder: lowering struct %s\n", .len=40}).ptr, ((ax_i64)(struct_name_ptr)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_free(struct_name_ptr);
            ax_u32 member = node.first_child;
            while ((member != ((ax_u32)(0)))) {
                struct ax_AstNode member_node = ((self->tree.nodes.data)[((ax_i64)(member))]);
                if ((member_node.kind == ax_NODE_FUNC_DECL)) {
                    ax_string name = ax_AirModuleBuilder_get_token_text(self, member_node.token_idx);
                    ax_u8* name_ptr = ax_AirModuleBuilder_builder_str_to_null_terminated(self, name);
                    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] AIR Builder: lowering struct method %s\n", .len=47}).ptr, ((ax_i64)(name_ptr)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    fflush(((void*)(NULL)));
                    ax_free(name_ptr);
                    struct ax_AirFunc* fn_ptr = ax_AirModuleBuilder_lower_func(self, member, member_node);
                    if ((fn_ptr != ((struct ax_AirFunc*)(NULL)))) {
                        ax_AirFuncVec_push(&(self->module.funcs), ((fn_ptr)[0]));
                        ax_free(((ax_u8*)(fn_ptr)));
                    }
                }
                member = member_node.next_sibling;
            }
        }
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] build_module: getting next sibling\n", .len=43}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        child = node.next_sibling;
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] build_module: next child is %d\n", .len=39}).ptr, ((ax_i64)(child)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
    }
}

ax_u8* ax_AirModuleBuilder_builder_str_to_null_terminated(struct ax_AirModuleBuilder* self, ax_string s) {
    ax_i64 len = ax_str_len(s);
    ax_u8* p = ((ax_u8*)(ax_alloc((len + 1))));
    memcpy(p, ((ax_u8*)(s.ptr)), len);
    ((p)[len]) = ((ax_u8)(0));
    return p;
    ax_free(p);
}

struct ax_SsaOptimizer ax_new_ssa_optimizer(void) {
    return ((struct ax_SsaOptimizer){.level=((ax_u8)(1))});
}

ax_u32 ax_AirFunc_max_reg_id(struct ax_AirFunc* f) {
    ax_u32 max_id = ((ax_u32)(0));
    ax_i64 i = ((ax_i64)(0));
    while ((i < f->insts.len)) {
        struct ax_AirInst inst = ((f->insts.data)[i]);
        if ((inst.dest > max_id)) {
            max_id = inst.dest;
        }
        if ((((inst.opcode != ax_OP_ICONST) && (inst.opcode != ax_OP_FCONST)) && (inst.opcode != ax_OP_JUMP))) {
            if ((inst.src1 > max_id)) {
                max_id = inst.src1;
            }
            if (((inst.src2 > max_id) && (inst.opcode != ax_OP_BRANCH))) {
                max_id = inst.src2;
            }
        }
        i = (i + ((ax_i64)(1)));
    }
    return max_id;
}

ax_bool ax_is_unary_foldable(ax_u16 op) {
    return ((op == ax_OP_NEG) || (op == ax_OP_NOT));
}

ax_bool ax_has_side_effect(ax_u16 op) {
    if (((((op == ax_OP_STORE) || (op == ax_OP_FREE)) || (op == ax_OP_DESTROY)) || (op == ax_OP_ALIAS_REUSE))) {
        return AX_TRUE;
    }
    if ((((op == ax_OP_CALL) || (op == ax_OP_SPAWN)) || (op == ax_OP_SEND))) {
        return AX_TRUE;
    }
    if ((((op == ax_OP_RETURN) || (op == ax_OP_JUMP)) || (op == ax_OP_BRANCH))) {
        return AX_TRUE;
    }
    if ((op == ax_OP_SET_FIELD)) {
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_bool ax_opcode_is_control(ax_u16 op) {
    if (((((op == ax_OP_JUMP) || (op == ax_OP_BRANCH)) || (op == ax_OP_CALL)) || (op == ax_OP_RETURN))) {
        return AX_TRUE;
    }
    if (((op == ax_OP_LOOP_BEGIN) || (op == ax_OP_LOOP_END))) {
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_bool ax_eval_binary(ax_u16 op, ax_u32 a, ax_u32 b, ax_u32* out_val) {
    ax_i32 ai = ((ax_i32)(a));
    ax_i32 bi = ((ax_i32)(b));
    if ((op == ax_OP_IADD)) {
        ((out_val)[0]) = ((ax_u32)((ai + bi)));
        return AX_TRUE;
    } else if ((op == ax_OP_ISUB)) {
        ((out_val)[0]) = ((ax_u32)((ai - bi)));
        return AX_TRUE;
    } else if ((op == ax_OP_IMUL)) {
        ((out_val)[0]) = ((ax_u32)((ai * bi)));
        return AX_TRUE;
    } else if ((op == ax_OP_IDIV)) {
        if ((bi == 0)) {
            return AX_FALSE;
        }
        ((out_val)[0]) = ((ax_u32)((ai / bi)));
        return AX_TRUE;
    } else if ((op == ax_OP_IMOD)) {
        if ((bi == 0)) {
            return AX_FALSE;
        }
        ((out_val)[0]) = ((ax_u32)((ai % bi)));
        return AX_TRUE;
    } else if ((op == ax_OP_EQ)) {
        ax_u32 res = ((ax_u32)(0));
        if ((ai == bi)) {
            res = ((ax_u32)(1));
        }
        ((out_val)[0]) = res;
        return AX_TRUE;
    } else if ((op == ax_OP_NE)) {
        ax_u32 res = ((ax_u32)(0));
        if ((ai != bi)) {
            res = ((ax_u32)(1));
        }
        ((out_val)[0]) = res;
        return AX_TRUE;
    } else if ((op == ax_OP_LT)) {
        ax_u32 res = ((ax_u32)(0));
        if ((ai < bi)) {
            res = ((ax_u32)(1));
        }
        ((out_val)[0]) = res;
        return AX_TRUE;
    } else if ((op == ax_OP_LE)) {
        ax_u32 res = ((ax_u32)(0));
        if ((ai <= bi)) {
            res = ((ax_u32)(1));
        }
        ((out_val)[0]) = res;
        return AX_TRUE;
    } else if ((op == ax_OP_GT)) {
        ax_u32 res = ((ax_u32)(0));
        if ((ai > bi)) {
            res = ((ax_u32)(1));
        }
        ((out_val)[0]) = res;
        return AX_TRUE;
    } else if ((op == ax_OP_GE)) {
        ax_u32 res = ((ax_u32)(0));
        if ((ai >= bi)) {
            res = ((ax_u32)(1));
        }
        ((out_val)[0]) = res;
        return AX_TRUE;
    } else if ((op == ax_OP_AND)) {
        ((out_val)[0]) = (a & b);
        return AX_TRUE;
    } else if ((op == ax_OP_OR)) {
        ((out_val)[0]) = (a | b);
        return AX_TRUE;
    } else if ((op == ax_OP_XOR)) {
        ((out_val)[0]) = (a ^ b);
        return AX_TRUE;
    } else if ((op == ax_OP_SHL)) {
        if ((b >= ((ax_u32)(32)))) {
            ((out_val)[0]) = ((ax_u32)(0));
            return AX_TRUE;
        }
        ((out_val)[0]) = (a << b);
        return AX_TRUE;
    } else if ((op == ax_OP_SHR)) {
        if ((b >= ((ax_u32)(32)))) {
            ((out_val)[0]) = ((ax_u32)(0));
            return AX_TRUE;
        }
        ((out_val)[0]) = ((ax_u32)((ai >> ((ax_i32)(b)))));
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_bool ax_eval_unary(ax_u16 op, ax_u32 a, ax_u32* out_val) {
    if ((op == ax_OP_NEG)) {
        ((out_val)[0]) = ((ax_u32)((-((ax_i32)(a)))));
        return AX_TRUE;
    } else if ((op == ax_OP_NOT)) {
        ax_u32 res = ((ax_u32)(0));
        if ((a == ((ax_u32)(0)))) {
            res = ((ax_u32)(1));
        }
        ((out_val)[0]) = res;
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_bool ax_AirFunc_fold_func(struct ax_AirFunc* f) {
    ax_u32 max_reg = ax_AirFunc_max_reg_id(f);
    if ((max_reg == ((ax_u32)(0)))) {
        return AX_FALSE;
    }
    ax_u32* def_counts = ((ax_u32*)(ax_alloc(((((ax_i64)(max_reg)) + ((ax_i64)(1))) * ((ax_i64)(4))))));
    memset(((void*)(def_counts)), ((ax_u8)(0)), ((((ax_i64)(max_reg)) + ((ax_i64)(1))) * ((ax_i64)(4))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < f->insts.len)) {
        struct ax_AirInst inst = ((f->insts.data)[i]);
        if ((((inst.opcode != ax_OP_NOP) && (inst.dest != ((ax_u32)(0)))) && (inst.opcode != ax_OP_BRANCH))) {
            ((def_counts)[inst.dest]) = (((def_counts)[inst.dest]) + ((ax_u32)(1)));
        }
        i = (i + ((ax_i64)(1)));
    }
    struct ax_ConstVal* vals = ((struct ax_ConstVal*)(ax_alloc(((((ax_i64)(max_reg)) + ((ax_i64)(1))) * ((ax_i64)(8))))));
    ax_u32 reg_idx = ((ax_u32)(0));
    while ((reg_idx <= max_reg)) {
        ((vals)[reg_idx]) = ((struct ax_ConstVal){.known=AX_FALSE, .val=((ax_u32)(0))});
        reg_idx = (reg_idx + ((ax_u32)(1)));
    }
    ax_bool changed = AX_FALSE;
    ax_u32* out_val = ((ax_u32*)(ax_alloc(((ax_i64)(4)))));
    i = ((ax_i64)(0));
    while ((i < f->insts.len)) {
        struct ax_AirInst* inst = &(((f->insts.data)[i]));
        if (((inst->opcode == ax_OP_ICONST) || (inst->opcode == ax_OP_FCONST))) {
            if ((inst->dest != ((ax_u32)(0)))) {
                if ((((def_counts)[inst->dest]) <= ((ax_u32)(1)))) {
                    ((vals)[inst->dest]) = ((struct ax_ConstVal){.known=AX_TRUE, .val=inst->src1});
                } else {
                    {
                        ((vals)[inst->dest]) = ((struct ax_ConstVal){.known=AX_FALSE, .val=((ax_u32)(0))});
                    }
                }
            }
            i = (i + ((ax_i64)(1)));
            continue;
        }
        if (((!ax_opcode_is_binary_alu(inst->opcode)) && (!ax_is_unary_foldable(inst->opcode)))) {
            if ((inst->dest != ((ax_u32)(0)))) {
                ((vals)[inst->dest]) = ((struct ax_ConstVal){.known=AX_FALSE, .val=((ax_u32)(0))});
            }
            i = (i + ((ax_i64)(1)));
            continue;
        }
        if (ax_opcode_is_binary_alu(inst->opcode)) {
            struct ax_ConstVal src1_val = ((vals)[inst->src1]);
            struct ax_ConstVal src2_val = ((vals)[inst->src2]);
            if ((src1_val.known && src2_val.known)) {
                ax_bool ok = ax_eval_binary(inst->opcode, src1_val.val, src2_val.val, out_val);
                if (ok) {
                    inst->opcode = ax_OP_ICONST;
                    inst->src1 = ((out_val)[0]);
                    inst->src2 = ((ax_u32)(0));
                    if ((((def_counts)[inst->dest]) <= ((ax_u32)(1)))) {
                        ((vals)[inst->dest]) = ((struct ax_ConstVal){.known=AX_TRUE, .val=((out_val)[0])});
                    } else {
                        {
                            ((vals)[inst->dest]) = ((struct ax_ConstVal){.known=AX_FALSE, .val=((ax_u32)(0))});
                        }
                    }
                    changed = AX_TRUE;
                    i = (i + ((ax_i64)(1)));
                    continue;
                }
            }
        }
        if (ax_is_unary_foldable(inst->opcode)) {
            struct ax_ConstVal src1_val = ((vals)[inst->src1]);
            if (src1_val.known) {
                ax_bool ok = ax_eval_unary(inst->opcode, src1_val.val, out_val);
                if (ok) {
                    inst->opcode = ax_OP_ICONST;
                    inst->src1 = ((out_val)[0]);
                    inst->src2 = ((ax_u32)(0));
                    if ((((def_counts)[inst->dest]) <= ((ax_u32)(1)))) {
                        ((vals)[inst->dest]) = ((struct ax_ConstVal){.known=AX_TRUE, .val=((out_val)[0])});
                    } else {
                        {
                            ((vals)[inst->dest]) = ((struct ax_ConstVal){.known=AX_FALSE, .val=((ax_u32)(0))});
                        }
                    }
                    changed = AX_TRUE;
                    i = (i + ((ax_i64)(1)));
                    continue;
                }
            }
        }
        if ((inst->dest != ((ax_u32)(0)))) {
            ((vals)[inst->dest]) = ((struct ax_ConstVal){.known=AX_FALSE, .val=((ax_u32)(0))});
        }
        i = (i + ((ax_i64)(1)));
    }
    ax_free(((ax_u8*)(def_counts)));
    ax_free(((ax_u8*)(vals)));
    ax_free(((ax_u8*)(out_val)));
    return changed;
}

ax_bool ax_AirFunc_copy_prop_func(struct ax_AirFunc* f) {
    ax_u32 max_reg = ax_AirFunc_max_reg_id(f);
    if ((max_reg == ((ax_u32)(0)))) {
        return AX_FALSE;
    }
    ax_u32* copy_map = ((ax_u32*)(ax_alloc(((((ax_i64)(max_reg)) + ((ax_i64)(1))) * ((ax_i64)(4))))));
    memset(((void*)(copy_map)), ((ax_u8)(0)), ((((ax_i64)(max_reg)) + ((ax_i64)(1))) * ((ax_i64)(4))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < f->insts.len)) {
        struct ax_AirInst inst = ((f->insts.data)[i]);
        if ((inst.opcode != ax_OP_NOP)) {
            if (((((inst.opcode == ax_OP_COPY) || (inst.opcode == ax_OP_MOVE)) && (inst.dest != ((ax_u32)(0)))) && (inst.src1 != ((ax_u32)(0))))) {
                ((copy_map)[inst.dest]) = inst.src1;
            }
        }
        i = (i + ((ax_i64)(1)));
    }
    ax_bool changed = AX_FALSE;
    i = ((ax_i64)(0));
    while ((i < f->insts.len)) {
        struct ax_AirInst* inst = &(((f->insts.data)[i]));
        if ((((inst->opcode == ax_OP_NOP) || (inst->opcode == ax_OP_ICONST)) || (inst->opcode == ax_OP_FCONST))) {
            i = (i + ((ax_i64)(1)));
            continue;
        }
        if (((inst->src1 != ((ax_u32)(0))) && (inst->opcode != ax_OP_JUMP))) {
            ax_u32 curr = inst->src1;
            ax_i32 depth = 0;
            while (((curr != ((ax_u32)(0))) && (depth < 100))) {
                ax_u32 parent = ((copy_map)[curr]);
                if ((parent == ((ax_u32)(0)))) {
                    break;
                }
                curr = parent;
                depth = (depth + 1);
            }
            if ((curr != inst->src1)) {
                inst->src1 = curr;
                changed = AX_TRUE;
            }
        }
        if (((inst->src2 != ((ax_u32)(0))) && (!ax_opcode_is_control(inst->opcode)))) {
            ax_u32 curr = inst->src2;
            ax_i32 depth = 0;
            while (((curr != ((ax_u32)(0))) && (depth < 100))) {
                ax_u32 parent = ((copy_map)[curr]);
                if ((parent == ((ax_u32)(0)))) {
                    break;
                }
                curr = parent;
                depth = (depth + 1);
            }
            if ((curr != inst->src2)) {
                inst->src2 = curr;
                changed = AX_TRUE;
            }
        }
        if (((inst->dest != ((ax_u32)(0))) && ((inst->opcode == ax_OP_STORE) || (inst->opcode == ax_OP_SET_FIELD)))) {
            ax_u32 curr = inst->dest;
            ax_i32 depth = 0;
            while (((curr != ((ax_u32)(0))) && (depth < 100))) {
                ax_u32 parent = ((copy_map)[curr]);
                if ((parent == ((ax_u32)(0)))) {
                    break;
                }
                curr = parent;
                depth = (depth + 1);
            }
            if ((curr != inst->dest)) {
                inst->dest = curr;
                changed = AX_TRUE;
            }
        }
        if (((inst->opcode == ax_OP_BRANCH) && (inst->src1 != ((ax_u32)(0))))) {
            ax_u32 curr = inst->src1;
            ax_i32 depth = 0;
            while (((curr != ((ax_u32)(0))) && (depth < 100))) {
                ax_u32 parent = ((copy_map)[curr]);
                if ((parent == ((ax_u32)(0)))) {
                    break;
                }
                curr = parent;
                depth = (depth + 1);
            }
            if ((curr != inst->src1)) {
                inst->src1 = curr;
                changed = AX_TRUE;
            }
        }
        if (((inst->opcode == ax_OP_RETURN) && (inst->src1 != ((ax_u32)(0))))) {
            ax_u32 curr = inst->src1;
            ax_i32 depth = 0;
            while (((curr != ((ax_u32)(0))) && (depth < 100))) {
                ax_u32 parent = ((copy_map)[curr]);
                if ((parent == ((ax_u32)(0)))) {
                    break;
                }
                curr = parent;
                depth = (depth + 1);
            }
            if ((curr != inst->src1)) {
                inst->src1 = curr;
                changed = AX_TRUE;
            }
        }
        if ((inst->opcode == ax_OP_CALL)) {
            ax_u32 arg_start = inst->src2;
            ax_u32 arg_count = ((ax_u32)(0));
            if ((arg_start < ((ax_u32)(f->extras.len)))) {
                arg_count = ((f->extras.data)[arg_start]);
            }
            ax_u32 idx = ((ax_u32)(0));
            while ((idx < arg_count)) {
                ax_u32 arg_reg_idx = ((arg_start + ((ax_u32)(1))) + idx);
                ax_u32 arg_reg = ((f->extras.data)[arg_reg_idx]);
                if ((arg_reg != ((ax_u32)(0)))) {
                    ax_u32 curr = arg_reg;
                    ax_i32 depth = 0;
                    while (((curr != ((ax_u32)(0))) && (depth < 100))) {
                        ax_u32 parent = ((copy_map)[curr]);
                        if ((parent == ((ax_u32)(0)))) {
                            break;
                        }
                        curr = parent;
                        depth = (depth + 1);
                    }
                    if ((curr != arg_reg)) {
                        ((f->extras.data)[arg_reg_idx]) = curr;
                        changed = AX_TRUE;
                    }
                }
                idx = (idx + ((ax_u32)(1)));
            }
        }
        i = (i + ((ax_i64)(1)));
    }
    ax_free(((ax_u8*)(copy_map)));
    return changed;
}

ax_bool ax_AirFunc_remove_unreachable_blocks(struct ax_AirFunc* f) {
    if ((f->blocks.len <= ((ax_i64)(1)))) {
        return AX_FALSE;
    }
    ax_bool* reachable = ((ax_bool*)(ax_alloc(f->blocks.len)));
    memset(((void*)(reachable)), ((ax_u8)(0)), f->blocks.len);
    ((reachable)[0]) = AX_TRUE;
    ax_u32* queue = ((ax_u32*)(ax_alloc((f->blocks.len * ((ax_i64)(4))))));
    ax_i64 head = ((ax_i64)(0));
    ax_i64 tail = ((ax_i64)(0));
    ((queue)[tail]) = ((ax_u32)(0));
    tail = (tail + ((ax_i64)(1)));
    while ((head < tail)) {
        ax_u32 cur = ((queue)[head]);
        head = (head + ((ax_i64)(1)));
        if ((((ax_i64)(cur)) < f->blocks.len)) {
            struct ax_BasicBlock blk = ((f->blocks.data)[cur]);
            ax_u32 i = ((ax_u32)(0));
            while ((i < blk.succs_len)) {
                ax_u32 succ = ((f->block_succs.data)[(blk.succs_start + i)]);
                if ((((ax_i64)(succ)) < f->blocks.len)) {
                    if ((!((reachable)[succ]))) {
                        ((reachable)[succ]) = AX_TRUE;
                        ((queue)[tail]) = succ;
                        tail = (tail + ((ax_i64)(1)));
                    }
                }
                i = (i + ((ax_u32)(1)));
            }
        }
    }
    if ((tail == f->blocks.len)) {
        ax_free(((ax_u8*)(reachable)));
        ax_free(((ax_u8*)(queue)));
        return AX_FALSE;
    }
    ax_i64 bi = ((ax_i64)(0));
    while ((bi < f->blocks.len)) {
        if ((!((reachable)[bi]))) {
            struct ax_BasicBlock* blk = &(((f->blocks.data)[bi]));
            ax_u32 i = ((ax_u32)(0));
            while ((i < blk->instrs_len)) {
                ax_u32 inst_idx = ((f->block_instrs.data)[(blk->instrs_start + i)]);
                if ((((ax_i64)(inst_idx)) < f->insts.len)) {
                    ((f->insts.data)[inst_idx]).opcode = ax_OP_NOP;
                    ((f->insts.data)[inst_idx]).type_id = ((ax_u16)(0));
                    ((f->insts.data)[inst_idx]).dest = ((ax_u32)(0));
                    ((f->insts.data)[inst_idx]).src1 = ((ax_u32)(0));
                    ((f->insts.data)[inst_idx]).src2 = ((ax_u32)(0));
                }
                i = (i + ((ax_u32)(1)));
            }
            blk->instrs_len = ((ax_u32)(0));
            blk->succs_len = ((ax_u32)(0));
            blk->preds_len = ((ax_u32)(0));
        }
        bi = (bi + ((ax_i64)(1)));
    }
    ax_free(((ax_u8*)(reachable)));
    ax_free(((ax_u8*)(queue)));
    return AX_TRUE;
}

ax_bool ax_AirFunc_dce_func(struct ax_AirFunc* f) {
    ax_u32 max_reg = ax_AirFunc_max_reg_id(f);
    if ((max_reg == ((ax_u32)(0)))) {
        return AX_FALSE;
    }
    ax_u32* use_count = ((ax_u32*)(ax_alloc(((((ax_i64)(max_reg)) + ((ax_i64)(1))) * ((ax_i64)(4))))));
    memset(((void*)(use_count)), ((ax_u8)(0)), ((((ax_i64)(max_reg)) + ((ax_i64)(1))) * ((ax_i64)(4))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < f->insts.len)) {
        struct ax_AirInst inst = ((f->insts.data)[i]);
        if ((inst.opcode == ax_OP_NOP)) {
            i = (i + ((ax_i64)(1)));
            continue;
        }
        if ((inst.opcode == ax_OP_CALL)) {
            ax_u32 arg_start = inst.src2;
            ax_u32 arg_count = ((ax_u32)(0));
            if ((arg_start < ((ax_u32)(f->extras.len)))) {
                arg_count = ((f->extras.data)[arg_start]);
            }
            ax_u32 idx = ((ax_u32)(0));
            while ((idx < arg_count)) {
                ax_u32 arg_reg = ((f->extras.data)[((arg_start + ((ax_u32)(1))) + idx)]);
                if ((arg_reg != ((ax_u32)(0)))) {
                    ((use_count)[arg_reg]) = (((use_count)[arg_reg]) + ((ax_u32)(1)));
                }
                idx = (idx + ((ax_u32)(1)));
            }
            i = (i + ((ax_i64)(1)));
            continue;
        }
        if ((((inst.opcode == ax_OP_ICONST) || (inst.opcode == ax_OP_FCONST)) || (inst.opcode == ax_OP_JUMP))) {
            i = (i + ((ax_i64)(1)));
            continue;
        }
        if ((inst.opcode == ax_OP_BRANCH)) {
            if ((inst.src1 != ((ax_u32)(0)))) {
                ((use_count)[inst.src1]) = (((use_count)[inst.src1]) + ((ax_u32)(1)));
            }
            i = (i + ((ax_i64)(1)));
            continue;
        }
        if (((inst.dest != ((ax_u32)(0))) && ((inst.opcode == ax_OP_STORE) || (inst.opcode == ax_OP_SET_FIELD)))) {
            ((use_count)[inst.dest]) = (((use_count)[inst.dest]) + ((ax_u32)(1)));
        }
        if ((inst.src1 != ((ax_u32)(0)))) {
            ((use_count)[inst.src1]) = (((use_count)[inst.src1]) + ((ax_u32)(1)));
        }
        if ((inst.src2 != ((ax_u32)(0)))) {
            ((use_count)[inst.src2]) = (((use_count)[inst.src2]) + ((ax_u32)(1)));
        }
        i = (i + ((ax_i64)(1)));
    }
    ax_bool changed = AX_FALSE;
    i = ((ax_i64)(0));
    while ((i < f->insts.len)) {
        struct ax_AirInst* inst = &(((f->insts.data)[i]));
        if (((inst->opcode != ax_OP_NOP) && (inst->dest != ((ax_u32)(0))))) {
            if ((!ax_has_side_effect(inst->opcode))) {
                if ((((use_count)[inst->dest]) == ((ax_u32)(0)))) {
                    inst->opcode = ax_OP_NOP;
                    inst->type_id = ((ax_u16)(0));
                    inst->dest = ((ax_u32)(0));
                    inst->src1 = ((ax_u32)(0));
                    inst->src2 = ((ax_u32)(0));
                    changed = AX_TRUE;
                }
            }
        }
        i = (i + ((ax_i64)(1)));
    }
    ax_free(((ax_u8*)(use_count)));
    if (ax_AirFunc_remove_unreachable_blocks(f)) {
        changed = AX_TRUE;
    }
    return changed;
}

void ax_SsaOptimizer_run(struct ax_SsaOptimizer* self, struct ax_AirModule* m) {
    ax_i64 fi = ((ax_i64)(0));
    while ((fi < m->funcs.len)) {
        struct ax_AirFunc* f = &(((m->funcs.data)[fi]));
        if ((!f->is_extern)) {
            ax_i32 iter = 0;
            while ((iter < 10)) {
                ax_bool changed = AX_FALSE;
                if (ax_AirFunc_fold_func(f)) {
                    changed = AX_TRUE;
                }
                if (ax_AirFunc_copy_prop_func(f)) {
                    changed = AX_TRUE;
                }
                if (ax_AirFunc_dce_func(f)) {
                    changed = AX_TRUE;
                }
                if ((!changed)) {
                    break;
                }
                iter = (iter + 1);
            }
        }
        fi = (fi + ((ax_i64)(1)));
    }
}

static ax_bool ax_opcode_defines_dest(ax_u16 op) {
    if ((op == ((ax_u16)(0)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0102)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x010B)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0104)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x010E)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0301)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0302)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0304)))) {
        return AX_FALSE;
    }
    return AX_TRUE;
}

struct ax_CGenerator ax_AirModule_new_c_generator(struct ax_AirModule mod, struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable) {
    return ((struct ax_CGenerator){.module=mod, .tree=tree, .pool=pool, .symtable=symtable, .typetable=typetable, .file=((void*)(NULL)), .reg_types=((ax_u32*)(NULL))});
}

ax_string ax_CGenerator_get_c_type(struct ax_CGenerator self, ax_u32 type_id) {
    if ((type_id == ((ax_u32)(1)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_i8", .len=5};
    }
    if ((type_id == ((ax_u32)(2)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_i16", .len=6};
    }
    if ((type_id == ((ax_u32)(3)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_i32", .len=6};
    }
    if ((type_id == ((ax_u32)(4)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_i64", .len=6};
    }
    if ((type_id == ((ax_u32)(5)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_u8", .len=5};
    }
    if ((type_id == ((ax_u32)(6)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_u16", .len=6};
    }
    if ((type_id == ((ax_u32)(7)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_u32", .len=6};
    }
    if ((type_id == ((ax_u32)(8)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_u64", .len=6};
    }
    if ((type_id == ((ax_u32)(9)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_f32", .len=6};
    }
    if ((type_id == ((ax_u32)(10)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_f64", .len=6};
    }
    if ((type_id == ((ax_u32)(11)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_bool", .len=7};
    }
    if ((type_id == ((ax_u32)(12)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_string", .len=9};
    }
    if ((type_id == ((ax_u32)(13)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_char", .len=7};
    }
    if ((type_id == ((ax_u32)(14)))) {
        return (ax_string){.ptr=(const ax_u8*)"void", .len=4};
    }
    if ((type_id == ((ax_u32)(15)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_isize", .len=8};
    }
    if ((type_id == ((ax_u32)(16)))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_usize", .len=8};
    }
    if ((type_id == ((ax_u32)(21)))) {
        return (ax_string){.ptr=(const ax_u8*)"AxActorID", .len=9};
    }
    if ((((ax_i64)(type_id)) < self.typetable.entries.len)) {
        struct ax_TypeEntry entry = ((self.typetable.entries.data)[type_id]);
        if ((entry.kind == ax_TYPE_KIND_POINTER)) {
            return (ax_string){.ptr=(const ax_u8*)"void*", .len=5};
        }
        if ((entry.kind == ax_TYPE_KIND_STRUCT)) {
            if ((entry.name_id != ((ax_u32)(0)))) {
                ax_string name = ax_InternPool_get(self.pool, entry.name_id);
                ax_u8* res = ((ax_u8*)(ax_alloc((ax_str_len(name) + 2))));
                memcpy(res, ((ax_u8*)(name.ptr)), ax_str_len(name));
                ((res)[ax_str_len(name)]) = ((ax_u8)('*'));
                ((res)[(ax_str_len(name) + 1)]) = ((ax_u8)(0));
                return ((ax_string){.ptr = (const ax_u8*)(res), .len = strlen((const char*)(res))});
            }
            return (ax_string){.ptr=(const ax_u8*)"void*", .len=5};
        }
        if ((entry.kind == ax_TYPE_KIND_SUM)) {
            return (ax_string){.ptr=(const ax_u8*)"ax_i32", .len=6};
        }
        if ((entry.kind == ax_TYPE_KIND_SLICE)) {
            return (ax_string){.ptr=(const ax_u8*)"ax_slice_void", .len=13};
        }
    }
    return (ax_string){.ptr=(const ax_u8*)"void*", .len=5};
}

ax_bool ax_CGenerator_is_stdlib_func(struct ax_CGenerator self, ax_string name) {
    if ((((((((((((((((((((((((((((((((((((((ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"system", .len=6}) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"remove", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fopen", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fclose", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fseek", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"ftell", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"rewind", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fread", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fwrite", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"printf", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"puts", .len=4})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"putchar", .len=7})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"getchar", .len=7})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fgetc", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fgets", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fputc", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fputs", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"sprintf", .len=7})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"snprintf", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fprintf", .len=7})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"scanf", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fscanf", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"sscanf", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"malloc", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"free", .len=4})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"realloc", .len=7})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"calloc", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"exit", .len=4})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"abort", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"memset", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"memcpy", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"memmove", .len=7})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"println", .len=7})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"print", .len=5})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"eprint", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"eprintln", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"assert", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"fflush", .len=6}))) {
        return AX_TRUE;
    }
    ax_u8* name_ptr = ((ax_u8*)(name.ptr));
    ax_i64 len = ax_str_len(name);
    if (((((len >= 3) && (((name_ptr)[0]) == ((ax_u8)('a')))) && (((name_ptr)[1]) == ((ax_u8)('x')))) && (((name_ptr)[2]) == ((ax_u8)('_'))))) {
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_string ax_CGenerator_get_mangled_name_by_sym(struct ax_CGenerator self, ax_u32 sym_idx) {
    ax_string fn_name = (ax_string){.ptr=(const ax_u8*)"func", .len=4};
    if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self.symtable.symbols.len))) {
        struct ax_Symbol sym = ((self.symtable.symbols.data)[((ax_i64)(sym_idx))]);
        fn_name = ax_InternPool_get(self.pool, sym.name_id);
        if (ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"free", .len=4})) {
            ax_bool is_extern = ((sym.flags & ((ax_u16)(4))) != ((ax_u16)(0)));
            if (is_extern) {
                return (ax_string){.ptr=(const ax_u8*)"free", .len=4};
            }
            return (ax_string){.ptr=(const ax_u8*)"ax_free", .len=7};
        }
        if (ax_CGenerator_is_stdlib_func(self, fn_name)) {
            return fn_name;
        }
        ax_bool is_extern = ((sym.flags & ((ax_u16)(4))) != ((ax_u16)(0)));
        if (is_extern) {
            return fn_name;
        }
        if (((sym.type_id != ((ax_u32)(0))) && (((ax_i64)(sym.type_id)) < self.typetable.entries.len))) {
            struct ax_TypeEntry entry = ((self.typetable.entries.data)[((ax_i64)(sym.type_id))]);
            if ((entry.kind == ax_TYPE_KIND_FUNC)) {
                struct ax_FuncInfo fi = ((self.typetable.funcs.data)[((ax_i64)(entry.extra))]);
                if ((fi.params.len > ((ax_i64)(0)))) {
                    ax_u32 first_param_type = ((fi.params.data)[0]);
                    struct ax_TypeEntry p_entry = ((self.typetable.entries.data)[((ax_i64)(first_param_type))]);
                    if ((p_entry.kind == ax_TYPE_KIND_POINTER)) {
                        first_param_type = p_entry.extra;
                        p_entry = ((self.typetable.entries.data)[((ax_i64)(first_param_type))]);
                    }
                    if (((p_entry.kind == ax_TYPE_KIND_STRUCT) && (p_entry.name_id != ((ax_u32)(0))))) {
                        ax_string struct_name = ax_InternPool_get(self.pool, p_entry.name_id);
                        ax_i64 struct_len = ax_str_len(struct_name);
                        ax_i64 fn_len = ax_str_len(fn_name);
                        ax_i64 total_len = (((3 + struct_len) + 1) + fn_len);
                        ax_u8* res = ((ax_u8*)(ax_alloc((total_len + 1))));
                        memcpy(res, ((ax_u8*)((ax_string){.ptr=(const ax_u8*)"ax_", .len=3}.ptr)), 3);
                        ax_u8* dest1 = ((ax_u8*)((((ax_u64)(res)) + ((ax_u64)(3)))));
                        memcpy(dest1, ((ax_u8*)(struct_name.ptr)), struct_len);
                        ((res)[(3 + struct_len)]) = ((ax_u8)('_'));
                        ax_u8* dest2 = ((ax_u8*)((((((ax_u64)(res)) + ((ax_u64)(3))) + ((ax_u64)(struct_len))) + ((ax_u64)(1)))));
                        memcpy(dest2, ((ax_u8*)(fn_name.ptr)), fn_len);
                        ((res)[total_len]) = ((ax_u8)(0));
                        return ((ax_string){.ptr = (const ax_u8*)(res), .len = strlen((const char*)(res))});
                    }
                }
            }
        }
    }
    ax_i64 fn_len = ax_str_len(fn_name);
    ax_u8* res = ((ax_u8*)(ax_alloc(((3 + fn_len) + 1))));
    memcpy(res, ((ax_u8*)((ax_string){.ptr=(const ax_u8*)"ax_", .len=3}.ptr)), 3);
    ax_u8* dest = ((ax_u8*)((((ax_u64)(res)) + ((ax_u64)(3)))));
    memcpy(dest, ((ax_u8*)(fn_name.ptr)), fn_len);
    ((res)[(3 + fn_len)]) = ((ax_u8)(0));
    return ((ax_string){.ptr = (const ax_u8*)(res), .len = strlen((const char*)(res))});
}

static ax_u8* ax_str_to_null_terminated(ax_string s) {
    ax_i64 len = ax_str_len(s);
    ax_u8* buf = ((ax_u8*)(ax_alloc((len + 1))));
    memcpy(buf, ((ax_u8*)(s.ptr)), len);
    ((buf)[len]) = ((ax_u8)(0));
    return buf;
    ax_free(buf);
}

ax_bool ax_CGenerator_generate(struct ax_CGenerator* self, ax_string out_filename) {
    ax_u8* filename_nt = ax_str_to_null_terminated(out_filename);
    self->file = fopen(((void*)(filename_nt)), (const char*)((ax_string){.ptr=(const ax_u8*)"w", .len=1}).ptr);
    ax_free(filename_nt);
    if ((self->file == ((void*)(NULL)))) {
        return AX_FALSE;
    }
    ax_bool has_main = AX_FALSE;
    ax_bool main_has_args = AX_FALSE;
    ax_i64 fn_idx = ((ax_i64)(0));
    while ((fn_idx < self->module.funcs.len)) {
        struct ax_AirFunc f = ((self->module.funcs.data)[fn_idx]);
        ax_string original_name = ax_InternPool_get(self->pool, f.name);
        if (ax_str_eq(original_name, (ax_string){.ptr=(const ax_u8*)"main", .len=4})) {
            has_main = AX_TRUE;
            if ((f.params.len > ((ax_i64)(0)))) {
                main_has_args = AX_TRUE;
            }
        }
        fn_idx = (fn_idx + ((ax_i64)(1)));
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"// Generated automatically by AXIOM AirCGen Stage 1\n", .len=52}).ptr, self->file);
    if (has_main) {
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"#define AX_EMIT_MAIN\n", .len=21}).ptr, self->file);
        if (main_has_args) {
            fputs((const char*)((ax_string){.ptr=(const ax_u8*)"#define AX_MAIN_WITH_ARGS\n", .len=26}).ptr, self->file);
        }
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"#include \"ax_runtime.h\"\n", .len=24}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"#include \"ax_stdlib.h\"\n\n", .len=24}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"#define r_0 0\n\n", .len=15}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"// Forward Declarations\n", .len=24}).ptr, self->file);
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->typetable.entries.len)) {
        struct ax_TypeEntry entry = ((self->typetable.entries.data)[i]);
        if (((entry.kind == ax_TYPE_KIND_STRUCT) && (entry.name_id != ((ax_u32)(0))))) {
            ax_string name = ax_InternPool_get(self->pool, entry.name_id);
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"typedef struct %s %s;\n", .len=22}).ptr, ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        }
        i = (i + 1);
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"\n", .len=1}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"// Struct Definitions\n", .len=22}).ptr, self->file);
    i = 0;
    while ((i < self->typetable.entries.len)) {
        struct ax_TypeEntry entry = ((self->typetable.entries.data)[i]);
        if (((entry.kind == ax_TYPE_KIND_STRUCT) && (entry.name_id != ((ax_u32)(0))))) {
            ax_string name = ax_InternPool_get(self->pool, entry.name_id);
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"struct %s {\n", .len=12}).ptr, ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            struct ax_StructInfo si = ((self->typetable.structs.data)[((ax_i64)(entry.extra))]);
            ax_i64 f_idx = ((ax_i64)(0));
            while ((f_idx < si.fields.len)) {
                struct ax_StructField field = ((si.fields.data)[f_idx]);
                ax_string field_type = ax_CGenerator_get_c_type(*(self), field.type_id);
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    %s f_%d;\n", .len=13}).ptr, ((ax_i64)(((ax_u8*)(field_type.ptr)))), f_idx, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                f_idx = (f_idx + ((ax_i64)(1)));
            }
            fputs((const char*)((ax_string){.ptr=(const ax_u8*)"};\n\n", .len=4}).ptr, self->file);
        }
        i = (i + 1);
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"// Function Prototypes\n", .len=23}).ptr, self->file);
    ax_i64 f_idx = ((ax_i64)(0));
    while ((f_idx < self->module.funcs.len)) {
        struct ax_AirFunc f = ((self->module.funcs.data)[f_idx]);
        if (f.is_extern) {
            f_idx = (f_idx + ((ax_i64)(1)));
        } else {
            {
                ax_string original_name = ax_InternPool_get(self->pool, f.name);
                ax_string mangled_name = ax_CGenerator_get_mangled_name_by_sym(*(self), f.sym_id);
                ax_string ret_type = ax_CGenerator_get_c_type(*(self), f.ret_type);
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"%s %s(", .len=6}).ptr, ((ax_i64)(((ax_u8*)(ret_type.ptr)))), ((ax_i64)(((ax_u8*)(mangled_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                ax_i64 p_idx = ((ax_i64)(0));
                while ((p_idx < f.params.len)) {
                    if ((p_idx > 0)) {
                        fputs((const char*)((ax_string){.ptr=(const ax_u8*)", ", .len=2}).ptr, self->file);
                    }
                    ax_string param_type = ax_CGenerator_get_c_type(*(self), ((f.params.data)[p_idx]));
                    if ((ax_str_eq(original_name, (ax_string){.ptr=(const ax_u8*)"main", .len=4}) && (p_idx == ((ax_i64)(1))))) {
                        param_type = (ax_string){.ptr=(const ax_u8*)"ax_u8**", .len=7};
                    }
                    fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"%s p_%d", .len=7}).ptr, ((ax_i64)(((ax_u8*)(param_type.ptr)))), p_idx, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    p_idx = (p_idx + ((ax_i64)(1)));
                }
                fputs((const char*)((ax_string){.ptr=(const ax_u8*)");\n", .len=3}).ptr, self->file);
                if ((((!ax_str_eq(mangled_name, original_name)) && (!ax_str_eq(mangled_name, (ax_string){.ptr=(const ax_u8*)"free", .len=4}))) && (!ax_str_eq(mangled_name, (ax_string){.ptr=(const ax_u8*)"ax_free", .len=7})))) {
                    ax_free(((ax_u8*)(mangled_name.ptr)));
                }
                f_idx = (f_idx + ((ax_i64)(1)));
            }
        }
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"\n", .len=1}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"// Function Definitions\n", .len=24}).ptr, self->file);
    f_idx = 0;
    while ((f_idx < self->module.funcs.len)) {
        struct ax_AirFunc f = ((self->module.funcs.data)[f_idx]);
        if (f.is_extern) {
            f_idx = (f_idx + ((ax_i64)(1)));
        } else {
            {
                ax_string original_name = ax_InternPool_get(self->pool, f.name);
                ax_string mangled_name = ax_CGenerator_get_mangled_name_by_sym(*(self), f.sym_id);
                ax_string ret_type = ax_CGenerator_get_c_type(*(self), f.ret_type);
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"%s %s(", .len=6}).ptr, ((ax_i64)(((ax_u8*)(ret_type.ptr)))), ((ax_i64)(((ax_u8*)(mangled_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                ax_i64 p_idx = ((ax_i64)(0));
                while ((p_idx < f.params.len)) {
                    if ((p_idx > 0)) {
                        fputs((const char*)((ax_string){.ptr=(const ax_u8*)", ", .len=2}).ptr, self->file);
                    }
                    ax_string param_type = ax_CGenerator_get_c_type(*(self), ((f.params.data)[p_idx]));
                    if ((ax_str_eq(original_name, (ax_string){.ptr=(const ax_u8*)"main", .len=4}) && (p_idx == ((ax_i64)(1))))) {
                        param_type = (ax_string){.ptr=(const ax_u8*)"ax_u8**", .len=7};
                    }
                    fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"%s p_%d", .len=7}).ptr, ((ax_i64)(((ax_u8*)(param_type.ptr)))), p_idx, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    p_idx = (p_idx + ((ax_i64)(1)));
                }
                fputs((const char*)((ax_string){.ptr=(const ax_u8*)") {\n", .len=4}).ptr, self->file);
                if ((((!ax_str_eq(mangled_name, original_name)) && (!ax_str_eq(mangled_name, (ax_string){.ptr=(const ax_u8*)"free", .len=4}))) && (!ax_str_eq(mangled_name, (ax_string){.ptr=(const ax_u8*)"ax_free", .len=7})))) {
                    ax_free(((ax_u8*)(mangled_name.ptr)));
                }
                ax_u32 max_reg = ((ax_u32)(0));
                ax_i64 inst_idx = ((ax_i64)(0));
                while ((inst_idx < f.insts.len)) {
                    struct ax_AirInst inst = ((f.insts.data)[inst_idx]);
                    if ((inst.dest > max_reg)) {
                        max_reg = inst.dest;
                    }
                    inst_idx = (inst_idx + 1);
                }
                if ((max_reg > ((ax_u32)(0)))) {
                    self->reg_types = ((ax_u32*)(ax_alloc(((((ax_i64)(max_reg)) + 1) * 4))));
                    ax_i64 r = ((ax_i64)(0));
                    while ((r <= ((ax_i64)(max_reg)))) {
                        ((self->reg_types)[r]) = ((ax_u32)(0));
                        r = (r + 1);
                    }
                    if ((f.params.data != ((ax_u32*)(NULL)))) {
                        ax_i64 p_idx = ((ax_i64)(0));
                        while ((p_idx < f.params.len)) {
                            ax_u32 p_reg = ((ax_u32)((p_idx + 1)));
                            if ((p_reg <= max_reg)) {
                                ((self->reg_types)[p_reg]) = ((f.params.data)[p_idx]);
                            }
                            p_idx = (p_idx + 1);
                        }
                    }
                    inst_idx = 0;
                    while ((inst_idx < f.insts.len)) {
                        struct ax_AirInst inst = ((f.insts.data)[inst_idx]);
                        if ((ax_opcode_defines_dest(inst.opcode) && (inst.dest > ((ax_u32)(0))))) {
                            ax_u32 t_id = ((ax_u32)(inst.type_id));
                            if ((inst.opcode == ((ax_u16)(0x010D)))) {
                                ax_u32 obj_type_id = ((ax_u32)(inst.type_id));
                                if (((obj_type_id == ((ax_u32)(0))) && (inst.src1 > ((ax_u32)(0))))) {
                                    obj_type_id = ((self->reg_types)[inst.src1]);
                                }
                                if ((((ax_i64)(obj_type_id)) < self->typetable.entries.len)) {
                                    struct ax_TypeEntry entry = ((self->typetable.entries.data)[obj_type_id]);
                                    if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                                        ax_u32 pointee_id = entry.extra;
                                        if (((pointee_id != ((ax_u32)(0))) && (((ax_i64)(pointee_id)) < self->typetable.entries.len))) {
                                            entry = ((self->typetable.entries.data)[pointee_id]);
                                        }
                                    }
                                    if ((entry.kind == ax_TYPE_KIND_STRUCT)) {
                                        struct ax_StructInfo si = ((self->typetable.structs.data)[((ax_i64)(entry.extra))]);
                                        if ((((ax_i64)(inst.src2)) < si.fields.len)) {
                                            t_id = ((si.fields.data)[((ax_i64)(inst.src2))]).type_id;
                                        }
                                    }
                                }
                            } else if (((((inst.opcode == ((ax_u16)(0x0105))) || (inst.opcode == ((ax_u16)(0x0109)))) || (inst.opcode == ((ax_u16)(0x010F)))) && (t_id == ((ax_u32)(0))))) {
                                if (((inst.src1 > ((ax_u32)(0))) && (((self->reg_types)[inst.src1]) != ((ax_u32)(0))))) {
                                    ax_u32 src_type = ((self->reg_types)[inst.src1]);
                                    if ((((ax_i64)(src_type)) < self->typetable.entries.len)) {
                                        struct ax_TypeEntry entry = ((self->typetable.entries.data)[src_type]);
                                        if ((((entry.kind == ax_TYPE_KIND_POINTER) || (entry.kind == ax_TYPE_KIND_SLICE)) || (entry.kind == ax_TYPE_KIND_ARRAY))) {
                                            t_id = entry.extra;
                                        }
                                    }
                                }
                            } else if ((inst.opcode == ax_OP_CALL)) {
                                ax_u32 sym_idx = ((ax_u32)(inst.type_id));
                                if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symtable.symbols.len))) {
                                    struct ax_Symbol sym = ((self->symtable.symbols.data)[((ax_i64)(sym_idx))]);
                                    ax_string fn_name = ax_InternPool_get(self->pool, sym.name_id);
                                    if ((((((ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"println", .len=7}) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"print", .len=5})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"eprint", .len=6})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"eprintln", .len=8})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"assert", .len=6})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"free", .len=4}))) {
                                        t_id = ((ax_u32)(14));
                                    } else {
                                        {
                                            ax_u32 fn_type_id = sym.type_id;
                                            if (((fn_type_id != ((ax_u32)(0))) && (((ax_i64)(fn_type_id)) < self->typetable.entries.len))) {
                                                struct ax_TypeEntry entry = ((self->typetable.entries.data)[fn_type_id]);
                                                if ((entry.kind == ax_TYPE_KIND_FUNC)) {
                                                    struct ax_FuncInfo fi = ((self->typetable.funcs.data)[((ax_i64)(entry.extra))]);
                                                    t_id = fi.ret;
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                            ((self->reg_types)[inst.dest]) = t_id;
                        }
                        inst_idx = (inst_idx + 1);
                    }
                    r = 1;
                    while ((r <= ((ax_i64)(max_reg)))) {
                        if ((((self->reg_types)[r]) != ((ax_u32)(14)))) {
                            ax_string t_name = ax_CGenerator_get_c_type(*(self), ((self->reg_types)[r]));
                            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    %s r_%d = {0};\n", .len=19}).ptr, ((ax_i64)(((ax_u8*)(t_name.ptr)))), r, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                        }
                        r = (r + 1);
                    }
                    if ((f.params.data != ((ax_u32*)(NULL)))) {
                        ax_i64 p_idx = ((ax_i64)(0));
                        while ((p_idx < f.params.len)) {
                            ax_i64 r_idx = ((ax_i64)((p_idx + 1)));
                            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = p_%d;\n", .len=17}).ptr, r_idx, p_idx, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                            p_idx = (p_idx + 1);
                        }
                    }
                }
                ax_i64 b_idx = ((ax_i64)(0));
                while ((b_idx < f.blocks.len)) {
                    struct ax_BasicBlock bb = ((f.blocks.data)[b_idx]);
                    fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"block_%d: ;\n", .len=12}).ptr, ((ax_i64)(bb.id)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    ax_i64 inst_offset = ((ax_i64)(0));
                    while ((inst_offset < ((ax_i64)(bb.instrs_len)))) {
                        ax_u32 inst_pos = ((f.block_instrs.data)[(((ax_i64)(bb.instrs_start)) + inst_offset)]);
                        struct ax_AirInst inst = ((f.insts.data)[((ax_i64)(inst_pos))]);
                        ax_CGenerator_translate_inst(self, f, inst);
                        inst_offset = (inst_offset + ((ax_i64)(1)));
                    }
                    b_idx = (b_idx + ((ax_i64)(1)));
                }
                if ((max_reg > ((ax_u32)(0)))) {
                    ax_free(((ax_u8*)(self->reg_types)));
                    self->reg_types = ((ax_u32*)(NULL));
                }
                fputs((const char*)((ax_string){.ptr=(const ax_u8*)"}\n\n", .len=3}).ptr, self->file);
                f_idx = (f_idx + ((ax_i64)(1)));
            }
        }
    }
    fclose(self->file);
    return AX_TRUE;
}

void ax_CGenerator_translate_inst(struct ax_CGenerator* self, struct ax_AirFunc f, struct ax_AirInst inst) {
    ax_u16 op = inst.opcode;
    if ((op == ax_OP_NOP)) {
        return;
    }
    if ((op == ax_OP_ICONST)) {
        if ((inst.type_id == ((ax_u16)(12)))) {
            ax_string str_val = ax_InternPool_get(self->pool, inst.src1);
            ax_u8* val_ptr = ((ax_u8*)(str_val.ptr));
            ax_i64 val_len = ax_str_len(str_val);
            ax_bool is_stripped = AX_FALSE;
            if ((((val_len >= ((ax_i64)(2))) && (((val_ptr)[0]) == ((ax_u8)('"')))) && (((val_ptr)[(val_len - 1)]) == ((ax_u8)('"'))))) {
                val_ptr = ((ax_u8*)((((ax_u64)(val_ptr)) + ((ax_u64)(1)))));
                val_len = (val_len - ((ax_i64)(2)));
                is_stripped = AX_TRUE;
            }
            ax_u8* print_buf = val_ptr;
            if (is_stripped) {
                print_buf = ((ax_u8*)(ax_alloc((val_len + 1))));
                memcpy(print_buf, val_ptr, val_len);
                ((print_buf)[val_len]) = ((ax_u8)(0));
            }
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = AX_STR(\"%s\");\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(((ax_u8*)(print_buf)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            if (is_stripped) {
                ax_free(print_buf);
            }
        } else {
            {
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = %d;\n", .len=15}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_OP_FCONST)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = %d.0;\n", .len=17}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_IADD)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d + r_%d;\n", .len=24}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_ISUB)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d - r_%d;\n", .len=24}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_IMUL)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d * r_%d;\n", .len=24}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_IDIV)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d / r_%d;\n", .len=24}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_IMOD)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d %% r_%d;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_EQ)) {
        if ((((self->reg_types != ((ax_u32*)(NULL))) && (inst.src1 > ((ax_u32)(0)))) && (((self->reg_types)[inst.src1]) == ((ax_u32)(12))))) {
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = ax_str_eq(r_%d, r_%d);\n", .len=34}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d == r_%d;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_OP_NE)) {
        if ((((self->reg_types != ((ax_u32*)(NULL))) && (inst.src1 > ((ax_u32)(0)))) && (((self->reg_types)[inst.src1]) == ((ax_u32)(12))))) {
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = !ax_str_eq(r_%d, r_%d);\n", .len=35}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d != r_%d;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_OP_LT)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d < r_%d;\n", .len=24}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_LE)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d <= r_%d;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_GT)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d > r_%d;\n", .len=24}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_GE)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d >= r_%d;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_AND)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d && r_%d;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_OR)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d || r_%d;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_XOR)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d ^ r_%d;\n", .len=24}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_SHL)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d << r_%d;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_SHR)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d >> r_%d;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_NOT)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = !r_%d;\n", .len=18}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_NEG)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = -r_%d;\n", .len=18}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((((op == ax_OP_COPY) || (op == ax_OP_MOVE)) || (op == ax_OP_ALIAS_REUSE))) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = r_%d;\n", .len=17}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_CAST)) {
        if ((inst.type_id == ((ax_u16)(12)))) {
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = (ax_string){.ptr = (const ax_u8*)r_%d, .len = strlen((const char*)r_%d)};\n", .len=85}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                ax_string t_name = ax_CGenerator_get_c_type(*(self), ((ax_u32)(inst.type_id)));
                ax_bool is_str_src = AX_FALSE;
                if (((self->reg_types != ((ax_u32*)(NULL))) && (inst.src1 > ((ax_u32)(0))))) {
                    ax_u32 src_type = ((self->reg_types)[inst.src1]);
                    if ((src_type == ((ax_u32)(12)))) {
                        is_str_src = AX_TRUE;
                    }
                }
                if (is_str_src) {
                    fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = (%s)r_%d.ptr;\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(((ax_u8*)(t_name.ptr)))), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                } else {
                    {
                        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = (%s)r_%d;\n", .len=21}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(((ax_u8*)(t_name.ptr)))), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    }
                }
            }
        }
    } else if ((op == ax_OP_ALLOC)) {
        struct ax_TypeEntry entry = ((self->typetable.entries.data)[((ax_i64)(inst.type_id))]);
        if (((entry.kind == ax_TYPE_KIND_STRUCT) && (entry.name_id != ((ax_u32)(0))))) {
            ax_string name = ax_InternPool_get(self->pool, entry.name_id);
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = (%s*)ax_alloc(sizeof(%s));\n", .len=38}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = ax_alloc(%d);\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(entry.size)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if (((op == ax_OP_FREE) || (op == ax_OP_DESTROY))) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    ax_free(r_%d);\n", .len=19}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_SPAWN)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = (void*)ax_actor_spawn((AxHandlerFn)r_%d, NULL, 0);\n", .len=62}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_MAKE_REF)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = ax_make_ref(r_%d);\n", .len=30}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_DEREF)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = ax_deref(r_%d);\n", .len=27}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_GET_FIELD)) {
        ax_u32 obj_type_id = ((ax_u32)(inst.type_id));
        if ((((obj_type_id == ((ax_u32)(0))) && (self->reg_types != ((ax_u32*)(NULL)))) && (inst.src1 > ((ax_u32)(0))))) {
            obj_type_id = ((self->reg_types)[inst.src1]);
        }
        ax_string s_name = (ax_string){.ptr=(const ax_u8*)"void", .len=4};
        if (((obj_type_id != ((ax_u32)(0))) && (((ax_i64)(obj_type_id)) < self->typetable.entries.len))) {
            struct ax_TypeEntry entry = ((self->typetable.entries.data)[obj_type_id]);
            if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                ax_u32 pointee_id = entry.extra;
                if (((pointee_id != ((ax_u32)(0))) && (((ax_i64)(pointee_id)) < self->typetable.entries.len))) {
                    entry = ((self->typetable.entries.data)[pointee_id]);
                }
            }
            if (((entry.kind == ax_TYPE_KIND_STRUCT) && (entry.name_id != ((ax_u32)(0))))) {
                s_name = ax_InternPool_get(self->pool, entry.name_id);
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = ((struct %s*)r_%d)->f_%d;\n", .len=37}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(((ax_u8*)(s_name.ptr)))), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_SET_FIELD)) {
        ax_u32 obj_type_id = ((ax_u32)(inst.type_id));
        if ((((obj_type_id == ((ax_u32)(0))) && (self->reg_types != ((ax_u32*)(NULL)))) && (inst.src1 > ((ax_u32)(0))))) {
            obj_type_id = ((self->reg_types)[inst.src1]);
        }
        ax_string s_name = (ax_string){.ptr=(const ax_u8*)"void", .len=4};
        if (((obj_type_id != ((ax_u32)(0))) && (((ax_i64)(obj_type_id)) < self->typetable.entries.len))) {
            struct ax_TypeEntry entry = ((self->typetable.entries.data)[obj_type_id]);
            if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                ax_u32 pointee_id = entry.extra;
                if (((pointee_id != ((ax_u32)(0))) && (((ax_i64)(pointee_id)) < self->typetable.entries.len))) {
                    entry = ((self->typetable.entries.data)[pointee_id]);
                }
            }
            if (((entry.kind == ax_TYPE_KIND_STRUCT) && (entry.name_id != ((ax_u32)(0))))) {
                s_name = ax_InternPool_get(self->pool, entry.name_id);
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    ((struct %s*)r_%d)->f_%d = r_%d;\n", .len=37}).ptr, ((ax_i64)(((ax_u8*)(s_name.ptr)))), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_INDEX)) {
        ax_string t_name = (ax_string){.ptr=(const ax_u8*)"void*", .len=5};
        if (((self->reg_types != ((ax_u32*)(NULL))) && (inst.src1 > ((ax_u32)(0))))) {
            ax_u32 t_id = ((self->reg_types)[inst.src1]);
            if (((t_id != ((ax_u32)(0))) && (((ax_i64)(t_id)) < self->typetable.entries.len))) {
                struct ax_TypeEntry entry = ((self->typetable.entries.data)[t_id]);
                if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                    ax_u32 pointee = entry.extra;
                    t_name = ax_CGenerator_get_c_type(*(self), pointee);
                }
            }
        }
        if ((ax_str_eq(t_name, (ax_string){.ptr=(const ax_u8*)"void*", .len=5}) || ax_str_eq(t_name, (ax_string){.ptr=(const ax_u8*)"void", .len=4}))) {
            if (((self->reg_types != ((ax_u32*)(NULL))) && (inst.dest > ((ax_u32)(0))))) {
                ax_u32 dest_t = ((self->reg_types)[inst.dest]);
                if ((dest_t != ((ax_u32)(0)))) {
                    t_name = ax_CGenerator_get_c_type(*(self), dest_t);
                }
            }
        }
        if (ax_str_eq(t_name, (ax_string){.ptr=(const ax_u8*)"void", .len=4})) {
            t_name = (ax_string){.ptr=(const ax_u8*)"void*", .len=5};
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = ((%s*)r_%d)[r_%d];\n", .len=30}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(((ax_u8*)(t_name.ptr)))), ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_STORE)) {
        if ((inst.src2 != ((ax_u32)(0)))) {
            ax_string t_name = (ax_string){.ptr=(const ax_u8*)"void*", .len=5};
            if (((self->reg_types != ((ax_u32*)(NULL))) && (inst.dest > ((ax_u32)(0))))) {
                ax_u32 t_id = ((self->reg_types)[inst.dest]);
                if (((t_id != ((ax_u32)(0))) && (((ax_i64)(t_id)) < self->typetable.entries.len))) {
                    struct ax_TypeEntry entry = ((self->typetable.entries.data)[t_id]);
                    if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                        ax_u32 pointee = entry.extra;
                        t_name = ax_CGenerator_get_c_type(*(self), pointee);
                    }
                }
            }
            if ((ax_str_eq(t_name, (ax_string){.ptr=(const ax_u8*)"void*", .len=5}) || ax_str_eq(t_name, (ax_string){.ptr=(const ax_u8*)"void", .len=4}))) {
                if (((self->reg_types != ((ax_u32*)(NULL))) && (inst.src1 > ((ax_u32)(0))))) {
                    ax_u32 src_t = ((self->reg_types)[inst.src1]);
                    if ((src_t != ((ax_u32)(0)))) {
                        t_name = ax_CGenerator_get_c_type(*(self), src_t);
                    }
                }
            }
            if (ax_str_eq(t_name, (ax_string){.ptr=(const ax_u8*)"void", .len=4})) {
                t_name = (ax_string){.ptr=(const ax_u8*)"void*", .len=5};
            }
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    ((%s*)r_%d)[r_%d] = r_%d;\n", .len=30}).ptr, ((ax_i64)(((ax_u8*)(t_name.ptr)))), ((ax_i64)(inst.dest)), ((ax_i64)(inst.src2)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    *r_%d = r_%d;\n", .len=18}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_OP_JUMP)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    goto block_%d;\n", .len=19}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_BRANCH)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    if (r_%d) goto block_%d; else goto block_%d;\n", .len=49}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_RETURN)) {
        if ((inst.src1 != ((ax_u32)(0)))) {
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    return r_%d;\n", .len=17}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                if ((f.ret_type != ((ax_u32)(14)))) {
                    ax_string t_name = ax_CGenerator_get_c_type(*(self), f.ret_type);
                    fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    return (%s){0};\n", .len=20}).ptr, ((ax_i64)(((ax_u8*)(t_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                } else {
                    {
                        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"    return;\n", .len=12}).ptr, self->file);
                    }
                }
            }
        }
    } else if (((op == ax_OP_CALL) || (op == ax_OP_SYSCALL))) {
        ax_u32 arg_start = inst.src2;
        ax_u32 arg_count = ((f.extras.data)[((ax_i64)(arg_start))]);
        ax_bool is_assert = AX_FALSE;
        ax_bool is_extern = AX_FALSE;
        ax_string fn_name = (ax_string){.ptr=(const ax_u8*)"func", .len=4};
        ax_bool is_allocated = AX_FALSE;
        if ((op == ax_OP_SYSCALL)) {
            fn_name = (ax_string){.ptr=(const ax_u8*)"ax_syscall", .len=10};
        } else if ((inst.src1 == ((ax_u32)(0)))) {
            ax_u32 sym_idx = ((ax_u32)(inst.type_id));
            if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symtable.symbols.len))) {
                fn_name = ax_CGenerator_get_mangled_name_by_sym(*(self), sym_idx);
                ax_string original_name = ax_InternPool_get(self->pool, ((self->symtable.symbols.data)[((ax_i64)(sym_idx))]).name_id);
                if ((((!ax_str_eq(fn_name, original_name)) && (!ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"free", .len=4}))) && (!ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"ax_free", .len=7})))) {
                    is_allocated = AX_TRUE;
                }
                ax_u16 s_flags = ((self->symtable.symbols.data)[((ax_i64)(sym_idx))]).flags;
                if (((s_flags & ((ax_u16)(4))) != ((ax_u16)(0)))) {
                    is_extern = AX_TRUE;
                }
            }
            if (ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"assert", .len=6})) {
                fn_name = (ax_string){.ptr=(const ax_u8*)"ax_assert_axiom", .len=15};
                is_assert = AX_TRUE;
            } else if ((((ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"println", .len=7}) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"print", .len=5})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"eprint", .len=6})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"eprintln", .len=8}))) {
                ax_string suffix = (ax_string){.ptr=(const ax_u8*)"_str", .len=4};
                if ((arg_count > ((ax_u32)(0)))) {
                    ax_u32 first_arg_reg = ((f.extras.data)[(((ax_i64)(arg_start)) + 1)]);
                    if (((self->reg_types != ((ax_u32*)(NULL))) && (first_arg_reg > ((ax_u32)(0))))) {
                        ax_u32 t_id = ((self->reg_types)[first_arg_reg]);
                        if (((((t_id == ((ax_u32)(3))) || (t_id == ((ax_u32)(4)))) || (t_id == ((ax_u32)(7)))) || (t_id == ((ax_u32)(8))))) {
                            suffix = (ax_string){.ptr=(const ax_u8*)"_i64", .len=4};
                        } else if (((t_id == ((ax_u32)(9))) || (t_id == ((ax_u32)(10))))) {
                            suffix = (ax_string){.ptr=(const ax_u8*)"_f64", .len=4};
                        } else if ((t_id == ((ax_u32)(11)))) {
                            suffix = (ax_string){.ptr=(const ax_u8*)"_bool", .len=5};
                        }
                    }
                }
                ax_i64 fn_len = ax_str_len(fn_name);
                ax_i64 suffix_len = ax_str_len(suffix);
                ax_u8* mapped_name = ((ax_u8*)(ax_alloc(((fn_len + suffix_len) + 4))));
                memcpy(mapped_name, ((ax_u8*)((ax_string){.ptr=(const ax_u8*)"ax_", .len=3}.ptr)), 3);
                ax_u8* dest1 = ((ax_u8*)((((ax_u64)(mapped_name)) + ((ax_u64)(3)))));
                memcpy(dest1, ((ax_u8*)(fn_name.ptr)), fn_len);
                ax_u8* dest2 = ((ax_u8*)(((((ax_u64)(mapped_name)) + ((ax_u64)(3))) + ((ax_u64)(fn_len)))));
                memcpy(dest2, ((ax_u8*)(suffix.ptr)), suffix_len);
                ax_i64 term_idx = ((3 + fn_len) + suffix_len);
                ((mapped_name)[term_idx]) = ((ax_u8)(0));
                fn_name = ((ax_string){.ptr = (const ax_u8*)(mapped_name), .len = strlen((const char*)(mapped_name))});
                is_allocated = AX_TRUE;
            }
            if (((inst.dest != ((ax_u32)(0))) && ((self->reg_types == ((ax_u32*)(NULL))) || (((self->reg_types)[inst.dest]) != ((ax_u32)(14)))))) {
                if ((((self->reg_types)[inst.dest]) == ((ax_u32)(0)))) {
                    fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = (void*)%s(", .len=21}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(((ax_u8*)(fn_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                } else {
                    {
                        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = %s(", .len=14}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(((ax_u8*)(fn_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    }
                }
            } else {
                {
                    fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    %s(", .len=7}).ptr, ((ax_i64)(((ax_u8*)(fn_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                }
            }
            if (is_allocated) {
                ax_free(((ax_u8*)(fn_name.ptr)));
            }
        } else {
            {
                if (((inst.dest != ((ax_u32)(0))) && ((self->reg_types == ((ax_u32*)(NULL))) || (((self->reg_types)[inst.dest]) != ((ax_u32)(14)))))) {
                    if ((((self->reg_types)[inst.dest]) == ((ax_u32)(0)))) {
                        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = (void*)((void*(*)())r_%d)(", .len=37}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    } else {
                        {
                            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    r_%d = ((void*(*)())r_%d)(", .len=30}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                        }
                    }
                } else {
                    {
                        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    ((void(*)())r_%d)(", .len=22}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    }
                }
            }
        }
        ax_i64 i = ((ax_i64)(0));
        while ((i < ((ax_i64)(arg_count)))) {
            if ((i > 0)) {
                fputs((const char*)((ax_string){.ptr=(const ax_u8*)", ", .len=2}).ptr, self->file);
            }
            ax_u32 arg_reg = ((f.extras.data)[((((ax_i64)(arg_start)) + 1) + i)]);
            ax_bool is_str = AX_FALSE;
            if (((self->reg_types != ((ax_u32*)(NULL))) && (arg_reg > ((ax_u32)(0))))) {
                ax_u32 t_id = ((self->reg_types)[arg_reg]);
                if ((t_id == ((ax_u32)(12)))) {
                    is_str = AX_TRUE;
                }
            }
            if ((is_extern && is_str)) {
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"(const char*)(r_%d).ptr", .len=23}).ptr, ((ax_i64)(arg_reg)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else {
                {
                    fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"r_%d", .len=4}).ptr, ((ax_i64)(arg_reg)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                }
            }
            i = (i + 1);
        }
        if (is_assert) {
            fputs((const char*)((ax_string){.ptr=(const ax_u8*)", AX_STR(\"assertion failed\")", .len=28}).ptr, self->file);
        }
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)");\n", .len=3}).ptr, self->file);
    }
}

static ax_bool ax_is_type_unsigned(ax_u32 type_id) {
    if ((((((type_id == ((ax_u32)(5))) || (type_id == ((ax_u32)(6)))) || (type_id == ((ax_u32)(7)))) || (type_id == ((ax_u32)(8)))) || (type_id == ((ax_u32)(16))))) {
        return AX_TRUE;
    }
    return AX_FALSE;
}

static ax_string ax_TypeTable_map_wasm_type(struct ax_TypeTable typetable, ax_u32 type_id) {
    if (((type_id == ((ax_u32)(4))) || (type_id == ((ax_u32)(8))))) {
        return (ax_string){.ptr=(const ax_u8*)"i64", .len=3};
    }
    if ((type_id == ((ax_u32)(9)))) {
        return (ax_string){.ptr=(const ax_u8*)"f32", .len=3};
    }
    if ((type_id == ((ax_u32)(10)))) {
        return (ax_string){.ptr=(const ax_u8*)"f64", .len=3};
    }
    if ((type_id == ((ax_u32)(14)))) {
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    return (ax_string){.ptr=(const ax_u8*)"i32", .len=3};
}

static ax_bool ax_wasm_opcode_defines_dest(ax_u16 op) {
    if ((op == ((ax_u16)(0)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0102)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x010B)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0104)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x010E)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0301)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0302)))) {
        return AX_FALSE;
    }
    if ((op == ((ax_u16)(0x0304)))) {
        return AX_FALSE;
    }
    return AX_TRUE;
}

struct ax_WasmGenerator ax_AirModule_new_wasm_generator(struct ax_AirModule mod, struct ax_AstTree tree, struct ax_InternPool pool, struct ax_SymbolTable symtable, struct ax_TypeTable typetable) {
    return ((struct ax_WasmGenerator){.module=mod, .tree=tree, .pool=pool, .symtable=symtable, .typetable=typetable, .file=((void*)(NULL)), .reg_types=((ax_u32*)(NULL))});
}

ax_string ax_WasmGenerator_resolve_sym_name(struct ax_WasmGenerator self, ax_u32 sym_id, ax_u32 name_id) {
    if ((sym_id == ((ax_u32)(0)))) {
        return (ax_string){.ptr=(const ax_u8*)"main", .len=4};
    }
    if ((sym_id == ((ax_u32)(4294967295)))) {
        return (ax_string){.ptr=(const ax_u8*)"malloc", .len=6};
    }
    if ((sym_id == ((ax_u32)(4294967294)))) {
        return (ax_string){.ptr=(const ax_u8*)"free", .len=4};
    }
    ax_string name = ax_InternPool_get(self.pool, name_id);
    if ((ax_str_len(name) > 0)) {
        if ((((ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"main", .len=4}) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"printf", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"malloc", .len=6})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"free", .len=4}))) {
            return name;
        }
        ax_u8* ptr_name = ((ax_u8*)(name.ptr));
        if ((((((ax_str_len(name) >= 4) && (((ptr_name)[0]) == ((ax_u8)('_')))) && (((ptr_name)[1]) == ((ax_u8)('A')))) && (((ptr_name)[2]) == ((ax_u8)('X')))) && (((ptr_name)[3]) == ((ax_u8)('_'))))) {
            return name;
        }
        ax_i64 new_len = (4 + ax_str_len(name));
        ax_u8* buf = ((ax_u8*)(ax_alloc((new_len + 1))));
        memcpy(buf, ((ax_u8*)((ax_string){.ptr=(const ax_u8*)"_AX_", .len=4}.ptr)), 4);
        memcpy(((ax_u8*)((((ax_u64)(buf)) + ((ax_u64)(4))))), ((ax_u8*)(name.ptr)), ax_str_len(name));
        ((buf)[new_len]) = ((ax_u8)(0));
        return ((ax_string){.ptr = (const ax_u8*)(buf), .len = strlen((const char*)(buf))});
    }
    ax_u8* buf = ((ax_u8*)(ax_alloc(32)));
    snprintf(((void*)(buf)), 32, (const char*)((ax_string){.ptr=(const ax_u8*)"_AX_f%d", .len=7}).ptr, ((void*)(sym_id)), ((void*)(NULL)));
    return ((ax_string){.ptr = (const ax_u8*)(buf), .len = strlen((const char*)(buf))});
}

static ax_u8* ax_wasm_str_to_null_terminated(ax_string s) {
    ax_i64 len = ax_str_len(s);
    ax_u8* buf = ((ax_u8*)(ax_alloc((len + 1))));
    memcpy(buf, ((ax_u8*)(s.ptr)), len);
    ((buf)[len]) = ((ax_u8)(0));
    return buf;
    ax_free(buf);
}

ax_bool ax_WasmGenerator_generate(struct ax_WasmGenerator* self, ax_string out_filename) {
    ax_u8* filename_nt = ax_wasm_str_to_null_terminated(out_filename);
    self->file = fopen(((void*)(filename_nt)), (const char*)((ax_string){.ptr=(const ax_u8*)"w", .len=1}).ptr);
    ax_free(filename_nt);
    if ((self->file == ((void*)(NULL)))) {
        return AX_FALSE;
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"(module\n", .len=8}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"  (memory (export \"memory\") 1)\n\n", .len=32}).ptr, self->file);
    ax_i64 f_idx = ((ax_i64)(0));
    while ((f_idx < self->module.funcs.len)) {
        struct ax_AirFunc f = ((self->module.funcs.data)[f_idx]);
        if (f.is_extern) {
            ax_string callee = ax_WasmGenerator_resolve_sym_name(*(self), f.sym_id, f.name);
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"  (import \"env\" \"%s\" (func $%s", .len=30}).ptr, ((ax_i64)(((ax_u8*)(callee.ptr)))), ((ax_i64)(((ax_u8*)(callee.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            ax_i64 p_idx = ((ax_i64)(0));
            while ((p_idx < f.params.len)) {
                ax_string p_type = ax_TypeTable_map_wasm_type(self->typetable, ((f.params.data)[p_idx]));
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)" (param %s)", .len=11}).ptr, ((ax_i64)(((ax_u8*)(p_type.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                p_idx = (p_idx + 1);
            }
            if (((f.ret_type != ((ax_u32)(0))) && (f.ret_type != ((ax_u32)(14))))) {
                ax_string r_type = ax_TypeTable_map_wasm_type(self->typetable, f.ret_type);
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)" (result %s)", .len=12}).ptr, ((ax_i64)(((ax_u8*)(r_type.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
            fputs((const char*)((ax_string){.ptr=(const ax_u8*)"))\n", .len=3}).ptr, self->file);
            ax_free(((ax_u8*)(callee.ptr)));
        }
        f_idx = (f_idx + 1);
    }
    ax_bool has_malloc = AX_FALSE;
    ax_bool has_free = AX_FALSE;
    f_idx = 0;
    while ((f_idx < self->module.funcs.len)) {
        struct ax_AirFunc f = ((self->module.funcs.data)[f_idx]);
        if (f.is_extern) {
            ax_string name = ax_InternPool_get(self->pool, f.name);
            if (ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"malloc", .len=6})) {
                has_malloc = AX_TRUE;
            }
            if (ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"free", .len=4})) {
                has_free = AX_TRUE;
            }
        }
        f_idx = (f_idx + 1);
    }
    if ((!has_malloc)) {
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"  (import \"env\" \"malloc\" (func $malloc (param i32) (result i32)))\n", .len=66}).ptr, self->file);
    }
    if ((!has_free)) {
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"  (import \"env\" \"free\" (func $free (param i32)))\n", .len=49}).ptr, self->file);
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"\n", .len=1}).ptr, self->file);
    f_idx = 0;
    while ((f_idx < self->module.funcs.len)) {
        struct ax_AirFunc f = ((self->module.funcs.data)[f_idx]);
        if ((!f.is_extern)) {
            ax_WasmGenerator_compile_func(self, f);
        }
        f_idx = (f_idx + 1);
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)")\n", .len=2}).ptr, self->file);
    fclose(self->file);
    return AX_TRUE;
}

void ax_WasmGenerator_compile_func(struct ax_WasmGenerator* self, struct ax_AirFunc f) {
    ax_string mangled_name = ax_WasmGenerator_resolve_sym_name(*(self), f.sym_id, f.name);
    fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"  (func $%s (export \"%s\")", .len=25}).ptr, ((ax_i64)(((ax_u8*)(mangled_name.ptr)))), ((ax_i64)(((ax_u8*)(mangled_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    ax_free(((ax_u8*)(mangled_name.ptr)));
    ax_i64 p_idx = ((ax_i64)(0));
    while ((p_idx < f.params.len)) {
        ax_string param_type = ax_TypeTable_map_wasm_type(self->typetable, ((f.params.data)[p_idx]));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)" (param $p%d %s)", .len=16}).ptr, ((ax_i64)((p_idx + 1))), ((ax_i64)(((ax_u8*)(param_type.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        p_idx = (p_idx + 1);
    }
    if (((f.ret_type != ((ax_u32)(0))) && (f.ret_type != ((ax_u32)(14))))) {
        ax_string r_type = ax_TypeTable_map_wasm_type(self->typetable, f.ret_type);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)" (result %s)", .len=12}).ptr, ((ax_i64)(((ax_u8*)(r_type.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"\n", .len=1}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"    (local $state i32)\n", .len=23}).ptr, self->file);
    ax_u32 max_reg = ((ax_u32)(0));
    ax_i64 inst_idx = ((ax_i64)(0));
    while ((inst_idx < f.insts.len)) {
        struct ax_AirInst inst = ((f.insts.data)[inst_idx]);
        if ((inst.dest > max_reg)) {
            max_reg = inst.dest;
        }
        inst_idx = (inst_idx + 1);
    }
    if ((max_reg > ((ax_u32)(0)))) {
        self->reg_types = ((ax_u32*)(ax_alloc(((((ax_i64)(max_reg)) + 1) * 4))));
        ax_i64 r = ((ax_i64)(0));
        while ((r <= ((ax_i64)(max_reg)))) {
            ((self->reg_types)[r]) = ((ax_u32)(0));
            r = (r + 1);
        }
        if ((f.params.data != ((ax_u32*)(NULL)))) {
            ax_i64 p_idx = ((ax_i64)(0));
            while ((p_idx < f.params.len)) {
                ax_u32 p_reg = ((ax_u32)((p_idx + 1)));
                if ((p_reg <= max_reg)) {
                    ((self->reg_types)[p_reg]) = ((f.params.data)[p_idx]);
                }
                p_idx = (p_idx + 1);
            }
        }
        inst_idx = 0;
        while ((inst_idx < f.insts.len)) {
            struct ax_AirInst inst = ((f.insts.data)[inst_idx]);
            if ((ax_wasm_opcode_defines_dest(inst.opcode) && (inst.dest > ((ax_u32)(0))))) {
                ax_u32 t_id = ((ax_u32)(inst.type_id));
                if ((inst.opcode == ((ax_u16)(0x010D)))) {
                    ax_u32 obj_type_id = ((ax_u32)(inst.type_id));
                    if (((obj_type_id == ((ax_u32)(0))) && (inst.src1 > ((ax_u32)(0))))) {
                        obj_type_id = ((self->reg_types)[inst.src1]);
                    }
                    if ((((ax_i64)(obj_type_id)) < self->typetable.entries.len)) {
                        struct ax_TypeEntry entry = ((self->typetable.entries.data)[obj_type_id]);
                        if ((entry.kind == ax_TYPE_KIND_POINTER)) {
                            ax_u32 pointee_id = entry.extra;
                            if (((pointee_id != ((ax_u32)(0))) && (((ax_i64)(pointee_id)) < self->typetable.entries.len))) {
                                entry = ((self->typetable.entries.data)[pointee_id]);
                            }
                        }
                        if ((entry.kind == ax_TYPE_KIND_STRUCT)) {
                            struct ax_StructInfo si = ((self->typetable.structs.data)[((ax_i64)(entry.extra))]);
                            if ((((ax_i64)(inst.src2)) < si.fields.len)) {
                                t_id = ((si.fields.data)[((ax_i64)(inst.src2))]).type_id;
                            }
                        }
                    }
                } else if (((((inst.opcode == ((ax_u16)(0x0105))) || (inst.opcode == ((ax_u16)(0x0109)))) || (inst.opcode == ((ax_u16)(0x010F)))) && (t_id == ((ax_u32)(0))))) {
                    if (((inst.src1 > ((ax_u32)(0))) && (((self->reg_types)[inst.src1]) != ((ax_u32)(0))))) {
                        ax_u32 src_type = ((self->reg_types)[inst.src1]);
                        if ((((ax_i64)(src_type)) < self->typetable.entries.len)) {
                            struct ax_TypeEntry entry = ((self->typetable.entries.data)[src_type]);
                            if ((((entry.kind == ax_TYPE_KIND_POINTER) || (entry.kind == ax_TYPE_KIND_SLICE)) || (entry.kind == ax_TYPE_KIND_ARRAY))) {
                                t_id = entry.extra;
                            }
                        }
                    }
                } else if ((inst.opcode == ax_OP_CALL)) {
                    ax_u32 sym_idx = ((ax_u32)(inst.type_id));
                    if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symtable.symbols.len))) {
                        struct ax_Symbol sym = ((self->symtable.symbols.data)[((ax_i64)(sym_idx))]);
                        ax_string fn_name = ax_InternPool_get(self->pool, sym.name_id);
                        if ((((((ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"println", .len=7}) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"print", .len=5})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"eprint", .len=6})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"eprintln", .len=8})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"assert", .len=6})) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"free", .len=4}))) {
                            t_id = ((ax_u32)(14));
                        } else {
                            {
                                ax_u32 fn_type_id = sym.type_id;
                                if (((fn_type_id != ((ax_u32)(0))) && (((ax_i64)(fn_type_id)) < self->typetable.entries.len))) {
                                    struct ax_TypeEntry entry = ((self->typetable.entries.data)[fn_type_id]);
                                    if ((entry.kind == ax_TYPE_KIND_FUNC)) {
                                        struct ax_FuncInfo fi = ((self->typetable.funcs.data)[((ax_i64)(entry.extra))]);
                                        t_id = fi.ret;
                                    }
                                }
                            }
                        }
                    }
                }
                ((self->reg_types)[inst.dest]) = t_id;
            }
            inst_idx = (inst_idx + 1);
        }
        r = ((ax_i64)(1));
        while ((r <= ((ax_i64)(max_reg)))) {
            if ((((self->reg_types)[r]) != ((ax_u32)(14)))) {
                ax_string t_name = ax_TypeTable_map_wasm_type(self->typetable, ((self->reg_types)[r]));
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    (local $r%d %s)\n", .len=20}).ptr, r, ((ax_i64)(((ax_u8*)(t_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
            r = (r + 1);
        }
    }
    if ((f.blocks.len == ((ax_i64)(0)))) {
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"    return\n  )\n\n", .len=16}).ptr, self->file);
        if ((max_reg > ((ax_u32)(0)))) {
            ax_free(((ax_u8*)(self->reg_types)));
            self->reg_types = ((ax_u32*)(NULL));
        }
        return;
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"    (local.set $state (i32.const 0))\n\n", .len=38}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"    (block $outer\n", .len=18}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"      (loop $loop\n", .len=18}).ptr, self->file);
    ax_i64 bi = (f.blocks.len - 1);
    while ((bi >= 0)) {
        struct ax_BasicBlock bb = ((f.blocks.data)[bi]);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (block $b_%d\n", .len=21}).ptr, ((ax_i64)(bb.id)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        bi = (bi - 1);
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"          (br_table", .len=19}).ptr, self->file);
    bi = 0;
    while ((bi < f.blocks.len)) {
        struct ax_BasicBlock bb = ((f.blocks.data)[bi]);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)" $b_%d", .len=6}).ptr, ((ax_i64)(bb.id)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        bi = (bi + 1);
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)" $b_0 (local.get $state))\n", .len=26}).ptr, self->file);
    bi = 0;
    while ((bi < f.blocks.len)) {
        struct ax_BasicBlock bb = ((f.blocks.data)[bi]);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        ) ;; end $b_%d\n", .len=23}).ptr, ((ax_i64)(bb.id)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        ;; --- block_%d body ---\n", .len=33}).ptr, ((ax_i64)(bb.id)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        ax_i64 inst_offset = ((ax_i64)(0));
        while ((inst_offset < ((ax_i64)(bb.instrs_len)))) {
            ax_u32 inst_pos = ((f.block_instrs.data)[(((ax_i64)(bb.instrs_start)) + inst_offset)]);
            struct ax_AirInst inst = ((f.insts.data)[((ax_i64)(inst_pos))]);
            ax_WasmGenerator_lower_inst(self, f, inst);
            inst_offset = (inst_offset + 1);
        }
        bi = (bi + 1);
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"      )\n", .len=8}).ptr, self->file);
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"    )\n", .len=6}).ptr, self->file);
    if (((f.ret_type != ((ax_u32)(0))) && (f.ret_type != ((ax_u32)(14))))) {
        ax_string r_type = ax_TypeTable_map_wasm_type(self->typetable, f.ret_type);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"    (%s.const 0)\n", .len=17}).ptr, ((ax_i64)(((ax_u8*)(r_type.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    }
    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"  )\n\n", .len=5}).ptr, self->file);
    if ((max_reg > ((ax_u32)(0)))) {
        ax_free(((ax_u8*)(self->reg_types)));
        self->reg_types = ((ax_u32*)(NULL));
    }
}

void ax_WasmGenerator_lower_inst(struct ax_WasmGenerator* self, struct ax_AirFunc f, struct ax_AirInst inst) {
    ax_u16 op = inst.opcode;
    if ((op == ax_OP_NOP)) {
        return;
    }
    if ((op == ax_OP_ICONST)) {
        ax_string t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (%s.const %d)\n", .len=22}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_FCONST)) {
        ax_string t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (%s.const %d)\n", .len=22}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((((op == ax_OP_COPY) || (op == ax_OP_MOVE)) || (op == ax_OP_ALIAS_REUSE))) {
        if (((inst.src1 <= ((ax_u32)(f.params.len))) && (inst.src1 > ((ax_u32)(0))))) {
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $p%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((((((((((op == ax_OP_IADD) || (op == ax_OP_ISUB)) || (op == ax_OP_IMUL)) || (op == ax_OP_IDIV)) || (op == ax_OP_IMOD)) || (op == ax_OP_FADD)) || (op == ax_OP_FSUB)) || (op == ax_OP_FMUL)) || (op == ax_OP_FDIV))) {
        ax_string t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        ax_string op_str = (ax_string){.ptr=(const ax_u8*)"add", .len=3};
        if (((op == ax_OP_IADD) || (op == ax_OP_FADD))) {
            op_str = (ax_string){.ptr=(const ax_u8*)"add", .len=3};
        } else if (((op == ax_OP_ISUB) || (op == ax_OP_FSUB))) {
            op_str = (ax_string){.ptr=(const ax_u8*)"sub", .len=3};
        } else if (((op == ax_OP_IMUL) || (op == ax_OP_FMUL))) {
            op_str = (ax_string){.ptr=(const ax_u8*)"mul", .len=3};
        } else if ((op == ax_OP_IDIV)) {
            if (ax_is_type_unsigned(((ax_u32)(inst.type_id)))) {
                op_str = (ax_string){.ptr=(const ax_u8*)"div_u", .len=5};
            } else {
                {
                    op_str = (ax_string){.ptr=(const ax_u8*)"div_s", .len=5};
                }
            }
        } else if ((op == ax_OP_FDIV)) {
            op_str = (ax_string){.ptr=(const ax_u8*)"div", .len=3};
        } else if ((op == ax_OP_IMOD)) {
            if (ax_is_type_unsigned(((ax_u32)(inst.type_id)))) {
                op_str = (ax_string){.ptr=(const ax_u8*)"rem_u", .len=5};
            } else {
                {
                    op_str = (ax_string){.ptr=(const ax_u8*)"rem_s", .len=5};
                }
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.%s\n", .len=14}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(((ax_u8*)(op_str.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if (((((((op == ax_OP_EQ) || (op == ax_OP_NE)) || (op == ax_OP_LT)) || (op == ax_OP_LE)) || (op == ax_OP_GT)) || (op == ax_OP_GE))) {
        ax_string src_type = (ax_string){.ptr=(const ax_u8*)"i32", .len=3};
        ax_bool is_unsigned_cmp = AX_FALSE;
        if (((inst.src1 > ((ax_u32)(0))) && (self->reg_types != ((ax_u32*)(NULL))))) {
            ax_u32 src1_type = ((self->reg_types)[inst.src1]);
            src_type = ax_TypeTable_map_wasm_type(self->typetable, src1_type);
            is_unsigned_cmp = ax_is_type_unsigned(src1_type);
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        ax_string cmp_op = (ax_string){.ptr=(const ax_u8*)"eq", .len=2};
        if ((op == ax_OP_EQ)) {
            cmp_op = (ax_string){.ptr=(const ax_u8*)"eq", .len=2};
        } else if ((op == ax_OP_NE)) {
            cmp_op = (ax_string){.ptr=(const ax_u8*)"ne", .len=2};
        } else if ((op == ax_OP_LT)) {
            if ((ax_str_eq(src_type, (ax_string){.ptr=(const ax_u8*)"f32", .len=3}) || ax_str_eq(src_type, (ax_string){.ptr=(const ax_u8*)"f64", .len=3}))) {
                cmp_op = (ax_string){.ptr=(const ax_u8*)"lt", .len=2};
            } else if (is_unsigned_cmp) {
                cmp_op = (ax_string){.ptr=(const ax_u8*)"lt_u", .len=4};
            } else {
                {
                    cmp_op = (ax_string){.ptr=(const ax_u8*)"lt_s", .len=4};
                }
            }
        } else if ((op == ax_OP_LE)) {
            if ((ax_str_eq(src_type, (ax_string){.ptr=(const ax_u8*)"f32", .len=3}) || ax_str_eq(src_type, (ax_string){.ptr=(const ax_u8*)"f64", .len=3}))) {
                cmp_op = (ax_string){.ptr=(const ax_u8*)"le", .len=2};
            } else if (is_unsigned_cmp) {
                cmp_op = (ax_string){.ptr=(const ax_u8*)"le_u", .len=4};
            } else {
                {
                    cmp_op = (ax_string){.ptr=(const ax_u8*)"le_s", .len=4};
                }
            }
        } else if ((op == ax_OP_GT)) {
            if ((ax_str_eq(src_type, (ax_string){.ptr=(const ax_u8*)"f32", .len=3}) || ax_str_eq(src_type, (ax_string){.ptr=(const ax_u8*)"f64", .len=3}))) {
                cmp_op = (ax_string){.ptr=(const ax_u8*)"gt", .len=2};
            } else if (is_unsigned_cmp) {
                cmp_op = (ax_string){.ptr=(const ax_u8*)"gt_u", .len=4};
            } else {
                {
                    cmp_op = (ax_string){.ptr=(const ax_u8*)"gt_s", .len=4};
                }
            }
        } else if ((op == ax_OP_GE)) {
            if ((ax_str_eq(src_type, (ax_string){.ptr=(const ax_u8*)"f32", .len=3}) || ax_str_eq(src_type, (ax_string){.ptr=(const ax_u8*)"f64", .len=3}))) {
                cmp_op = (ax_string){.ptr=(const ax_u8*)"ge", .len=2};
            } else if (is_unsigned_cmp) {
                cmp_op = (ax_string){.ptr=(const ax_u8*)"ge_u", .len=4};
            } else {
                {
                    cmp_op = (ax_string){.ptr=(const ax_u8*)"ge_s", .len=4};
                }
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.%s\n", .len=14}).ptr, ((ax_i64)(((ax_u8*)(src_type.ptr)))), ((ax_i64)(((ax_u8*)(cmp_op.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((((((op == ax_OP_AND) || (op == ax_OP_OR)) || (op == ax_OP_XOR)) || (op == ax_OP_SHL)) || (op == ax_OP_SHR))) {
        ax_string t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        ax_string bit_op = (ax_string){.ptr=(const ax_u8*)"and", .len=3};
        if ((op == ax_OP_AND)) {
            bit_op = (ax_string){.ptr=(const ax_u8*)"and", .len=3};
        } else if ((op == ax_OP_OR)) {
            bit_op = (ax_string){.ptr=(const ax_u8*)"or", .len=2};
        } else if ((op == ax_OP_XOR)) {
            bit_op = (ax_string){.ptr=(const ax_u8*)"xor", .len=3};
        } else if ((op == ax_OP_SHL)) {
            bit_op = (ax_string){.ptr=(const ax_u8*)"shl", .len=3};
        } else if ((op == ax_OP_SHR)) {
            if (ax_is_type_unsigned(((ax_u32)(inst.type_id)))) {
                bit_op = (ax_string){.ptr=(const ax_u8*)"shr_u", .len=5};
            } else {
                {
                    bit_op = (ax_string){.ptr=(const ax_u8*)"shr_s", .len=5};
                }
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.%s\n", .len=14}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(((ax_u8*)(bit_op.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_NEG)) {
        ax_string t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        if ((ax_str_eq(t, (ax_string){.ptr=(const ax_u8*)"f32", .len=3}) || ax_str_eq(t, (ax_string){.ptr=(const ax_u8*)"f64", .len=3}))) {
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.neg\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (%s.const 0)\n", .len=21}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.sub\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_NOT)) {
        ax_string t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        if (ax_str_eq(t, (ax_string){.ptr=(const ax_u8*)"i64", .len=3})) {
            fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        (i64.const -1)\n", .len=23}).ptr, self->file);
        } else {
            {
                fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        (i32.const -1)\n", .len=23}).ptr, self->file);
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.xor\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if (((op == ax_OP_ALLOC) || (op == ax_OP_ARENA_ALLOC))) {
        ax_u32 size = ((ax_u32)(8));
        if (((inst.type_id != ((ax_u16)(0))) && (((ax_i64)(inst.type_id)) < self->typetable.entries.len))) {
            struct ax_TypeEntry entry = ((self->typetable.entries.data)[inst.type_id]);
            if ((entry.size != ((ax_u32)(0)))) {
                size = entry.size;
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (i32.const %d)\n", .len=23}).ptr, ((ax_i64)(size)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        (call $malloc)\n", .len=23}).ptr, self->file);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if (((op == ax_OP_FREE) || (op == ax_OP_DESTROY))) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        (call $free)\n", .len=21}).ptr, self->file);
    } else if ((op == ax_OP_LOAD)) {
        ax_string t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.load\n", .len=16}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_STORE)) {
        ax_string t = (ax_string){.ptr=(const ax_u8*)"i32", .len=3};
        if ((inst.src2 != ((ax_u32)(0)))) {
            if (((inst.src1 > ((ax_u32)(0))) && (self->reg_types != ((ax_u32*)(NULL))))) {
                ax_u32 val_type = ((self->reg_types)[inst.src1]);
                t = ax_TypeTable_map_wasm_type(self->typetable, val_type);
            }
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        i32.add\n", .len=16}).ptr, self->file);
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                if (((inst.src1 > ((ax_u32)(0))) && (self->reg_types != ((ax_u32*)(NULL))))) {
                    ax_u32 val_type = ((self->reg_types)[inst.src1]);
                    t = ax_TypeTable_map_wasm_type(self->typetable, val_type);
                }
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.store\n", .len=17}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_GEP)) {
        ax_u32 elem_size = ((ax_u32)(8));
        if (((inst.type_id != ((ax_u16)(0))) && (((ax_i64)(inst.type_id)) < self->typetable.entries.len))) {
            struct ax_TypeEntry entry = ((self->typetable.entries.data)[inst.type_id]);
            if ((entry.size != ((ax_u32)(0)))) {
                elem_size = entry.size;
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (i32.const %d)\n", .len=23}).ptr, ((ax_i64)(elem_size)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        i32.mul\n", .len=16}).ptr, self->file);
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        i32.add\n", .len=16}).ptr, self->file);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_GET_FIELD)) {
        ax_u32 offset = (inst.src2 * ((ax_u32)(8)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (i32.const %d)\n", .len=23}).ptr, ((ax_i64)(offset)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        i32.add\n", .len=16}).ptr, self->file);
        ax_string t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.load\n", .len=16}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_SET_FIELD)) {
        ax_u32 offset = (inst.src2 * ((ax_u32)(8)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (i32.const %d)\n", .len=23}).ptr, ((ax_i64)(offset)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        i32.add\n", .len=16}).ptr, self->file);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        ax_string t = (ax_string){.ptr=(const ax_u8*)"i32", .len=3};
        if (((inst.dest > ((ax_u32)(0))) && (self->reg_types != ((ax_u32*)(NULL))))) {
            ax_u32 dest_type = ((self->reg_types)[inst.dest]);
            t = ax_TypeTable_map_wasm_type(self->typetable, dest_type);
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.store\n", .len=17}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_INDEX)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        i32.add\n", .len=16}).ptr, self->file);
        ax_string t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        %s.load\n", .len=16}).ptr, ((ax_i64)(((ax_u8*)(t.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_CALL)) {
        ax_i64 arg_start = ((ax_i64)(inst.src2));
        ax_u32 arg_count = ((ax_u32)(0));
        if ((arg_start < f.extras.len)) {
            arg_count = ((f.extras.data)[arg_start]);
        }
        ax_i64 j = ((ax_i64)(0));
        while ((j < ((ax_i64)(arg_count)))) {
            ax_u32 arg_reg = ((f.extras.data)[((arg_start + 1) + j)]);
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(arg_reg)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            j = (j + 1);
        }
        ax_string callee = (ax_string){.ptr=(const ax_u8*)"", .len=0};
        if ((inst.src1 == ((ax_u32)(0)))) {
            ax_u32 sym_idx = ((ax_u32)(inst.type_id));
            if (((sym_idx != ((ax_u32)(0))) && (((ax_i64)(sym_idx)) < self->symtable.symbols.len))) {
                struct ax_Symbol sym = ((self->symtable.symbols.data)[((ax_i64)(sym_idx))]);
                callee = ax_WasmGenerator_resolve_sym_name(*(self), sym_idx, sym.name_id);
            } else {
                {
                    callee = ax_WasmGenerator_resolve_sym_name(*(self), ((ax_u32)(0)), ((ax_u32)(0)));
                }
            }
        } else {
            {
                callee = ax_WasmGenerator_resolve_sym_name(*(self), ((ax_u32)(0)), ((ax_u32)(0)));
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (call $%s)\n", .len=19}).ptr, ((ax_i64)(((ax_u8*)(callee.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        if ((inst.dest != ((ax_u32)(0)))) {
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        }
        ax_free(((ax_u8*)(callee.ptr)));
    } else if (((((op == ax_OP_CAST) || (op == ax_OP_ZEXT)) || (op == ax_OP_SEXT)) || (op == ax_OP_TRUNC))) {
        ax_string dst_t = ax_TypeTable_map_wasm_type(self->typetable, ((ax_u32)(inst.type_id)));
        ax_string src_t = (ax_string){.ptr=(const ax_u8*)"i32", .len=3};
        if (((inst.src1 > ((ax_u32)(0))) && (self->reg_types != ((ax_u32*)(NULL))))) {
            ax_u32 src_type = ((self->reg_types)[inst.src1]);
            src_t = ax_TypeTable_map_wasm_type(self->typetable, src_type);
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        if ((ax_str_eq(src_t, (ax_string){.ptr=(const ax_u8*)"i64", .len=3}) && ax_str_eq(dst_t, (ax_string){.ptr=(const ax_u8*)"i32", .len=3}))) {
            fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        i32.wrap_i64\n", .len=21}).ptr, self->file);
        } else if ((ax_str_eq(src_t, (ax_string){.ptr=(const ax_u8*)"i32", .len=3}) && ax_str_eq(dst_t, (ax_string){.ptr=(const ax_u8*)"i64", .len=3}))) {
            if ((op == ax_OP_ZEXT)) {
                fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        i64.extend_i32_u\n", .len=25}).ptr, self->file);
            } else {
                {
                    fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        i64.extend_i32_s\n", .len=25}).ptr, self->file);
                }
            }
        }
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_JUMP)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.set $state (i32.const %d))\n", .len=42}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        (br $loop)\n", .len=19}).ptr, self->file);
    } else if ((op == ax_OP_BRANCH)) {
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        (if\n", .len=12}).ptr, self->file);
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"          (then\n", .len=16}).ptr, self->file);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"            (local.set $state (i32.const %d))\n", .len=46}).ptr, ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"          )\n", .len=12}).ptr, self->file);
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"          (else\n", .len=16}).ptr, self->file);
        fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"            (local.set $state (i32.const %d))\n", .len=46}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"          )\n", .len=12}).ptr, self->file);
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        )\n", .len=10}).ptr, self->file);
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        (br $loop)\n", .len=19}).ptr, self->file);
    } else if ((op == ax_OP_RETURN)) {
        if ((inst.src1 != ((ax_u32)(0)))) {
            fprintf(self->file, (const char*)((ax_string){.ptr=(const ax_u8*)"        (local.get $r%d)\n", .len=25}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        }
        fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        return\n", .len=15}).ptr, self->file);
    } else {
        {
            fputs((const char*)((ax_string){.ptr=(const ax_u8*)"        nop\n", .len=12}).ptr, self->file);
        }
    }
}

ax_bool ax_reg_is_gpr(ax_u8 r) {
    return (r < ((ax_u8)(16)));
}

ax_u8 ax_reg_hw_reg(ax_u8 r) {
    return (r & ((ax_u8)(0x0F)));
}

ax_bool ax_reg_needs_rex(ax_u8 r) {
    return (ax_reg_hw_reg(r) >= ((ax_u8)(8)));
}

ax_string ax_reg_to_str(ax_u8 r) {
    if ((r == ax_REG_RAX)) {
        return (ax_string){.ptr=(const ax_u8*)"rax", .len=3};
    }
    if ((r == ax_REG_RCX)) {
        return (ax_string){.ptr=(const ax_u8*)"rcx", .len=3};
    }
    if ((r == ax_REG_RDX)) {
        return (ax_string){.ptr=(const ax_u8*)"rdx", .len=3};
    }
    if ((r == ax_REG_RBX)) {
        return (ax_string){.ptr=(const ax_u8*)"rbx", .len=3};
    }
    if ((r == ax_REG_RSP)) {
        return (ax_string){.ptr=(const ax_u8*)"rsp", .len=3};
    }
    if ((r == ax_REG_RBP)) {
        return (ax_string){.ptr=(const ax_u8*)"rbp", .len=3};
    }
    if ((r == ax_REG_RSI)) {
        return (ax_string){.ptr=(const ax_u8*)"rsi", .len=3};
    }
    if ((r == ax_REG_RDI)) {
        return (ax_string){.ptr=(const ax_u8*)"rdi", .len=3};
    }
    if ((r == ax_REG_R8)) {
        return (ax_string){.ptr=(const ax_u8*)"r8", .len=2};
    }
    if ((r == ax_REG_R9)) {
        return (ax_string){.ptr=(const ax_u8*)"r9", .len=2};
    }
    if ((r == ax_REG_R10)) {
        return (ax_string){.ptr=(const ax_u8*)"r10", .len=3};
    }
    if ((r == ax_REG_R11)) {
        return (ax_string){.ptr=(const ax_u8*)"r11", .len=3};
    }
    if ((r == ax_REG_R12)) {
        return (ax_string){.ptr=(const ax_u8*)"r12", .len=3};
    }
    if ((r == ax_REG_R13)) {
        return (ax_string){.ptr=(const ax_u8*)"r13", .len=3};
    }
    if ((r == ax_REG_R14)) {
        return (ax_string){.ptr=(const ax_u8*)"r14", .len=3};
    }
    if ((r == ax_REG_R15)) {
        return (ax_string){.ptr=(const ax_u8*)"r15", .len=3};
    }
    return (ax_string){.ptr=(const ax_u8*)"none", .len=4};
}

ax_u8 ax_get_sysv_arg_reg(ax_i64 idx) {
    if ((idx == 0)) {
        return ax_REG_RDI;
    }
    if ((idx == 1)) {
        return ax_REG_RSI;
    }
    if ((idx == 2)) {
        return ax_REG_RDX;
    }
    if ((idx == 3)) {
        return ax_REG_RCX;
    }
    if ((idx == 4)) {
        return ax_REG_R8;
    }
    if ((idx == 5)) {
        return ax_REG_R9;
    }
    return ax_REG_NONE;
}

ax_u8 ax_get_win64_arg_reg(ax_i64 idx) {
    if ((idx == 0)) {
        return ax_REG_RCX;
    }
    if ((idx == 1)) {
        return ax_REG_RDX;
    }
    if ((idx == 2)) {
        return ax_REG_R8;
    }
    if ((idx == 3)) {
        return ax_REG_R9;
    }
    return ax_REG_NONE;
}

ax_bool ax_reg_is_sysv_caller_saved(ax_u8 r) {
    if ((((((((((r == ax_REG_RAX) || (r == ax_REG_RCX)) || (r == ax_REG_RDX)) || (r == ax_REG_RSI)) || (r == ax_REG_RDI)) || (r == ax_REG_R8)) || (r == ax_REG_R9)) || (r == ax_REG_R10)) || (r == ax_REG_R11))) {
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_bool ax_reg_is_sysv_callee_saved(ax_u8 r) {
    if ((((((r == ax_REG_RBX) || (r == ax_REG_R12)) || (r == ax_REG_R13)) || (r == ax_REG_R14)) || (r == ax_REG_R15))) {
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_bool ax_reg_is_win64_callee_saved(ax_u8 r) {
    if (((((((((r == ax_REG_RBX) || (r == ax_REG_RBP)) || (r == ax_REG_RDI)) || (r == ax_REG_RSI)) || (r == ax_REG_R12)) || (r == ax_REG_R13)) || (r == ax_REG_R14)) || (r == ax_REG_R15))) {
        return AX_TRUE;
    }
    return AX_FALSE;
}

ax_string ax_x86_resolve_sym_name(ax_u32 sym_idx, struct ax_SymbolTable symbols, struct ax_InternPool pool) {
    if ((sym_idx == ((ax_u32)(0)))) {
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    if ((((ax_i64)(sym_idx)) >= symbols.symbols.len)) {
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    ax_u32 name_id = ((symbols.symbols.data)[sym_idx]).name_id;
    ax_string name = ax_InternPool_get(pool, name_id);
    return name;
}

ax_string ax_cond_code_to_str(ax_u8 cc) {
    if ((cc == ax_CC_O)) {
        return (ax_string){.ptr=(const ax_u8*)"o", .len=1};
    }
    if ((cc == ax_CC_NO)) {
        return (ax_string){.ptr=(const ax_u8*)"no", .len=2};
    }
    if ((cc == ax_CC_B)) {
        return (ax_string){.ptr=(const ax_u8*)"b", .len=1};
    }
    if ((cc == ax_CC_AE)) {
        return (ax_string){.ptr=(const ax_u8*)"ae", .len=2};
    }
    if ((cc == ax_CC_E)) {
        return (ax_string){.ptr=(const ax_u8*)"e", .len=1};
    }
    if ((cc == ax_CC_NE)) {
        return (ax_string){.ptr=(const ax_u8*)"ne", .len=2};
    }
    if ((cc == ax_CC_BE)) {
        return (ax_string){.ptr=(const ax_u8*)"be", .len=2};
    }
    if ((cc == ax_CC_A)) {
        return (ax_string){.ptr=(const ax_u8*)"a", .len=1};
    }
    if ((cc == ax_CC_S)) {
        return (ax_string){.ptr=(const ax_u8*)"s", .len=1};
    }
    if ((cc == ax_CC_NS)) {
        return (ax_string){.ptr=(const ax_u8*)"ns", .len=2};
    }
    if ((cc == ax_CC_PE)) {
        return (ax_string){.ptr=(const ax_u8*)"pe", .len=2};
    }
    if ((cc == ax_CC_PO)) {
        return (ax_string){.ptr=(const ax_u8*)"po", .len=2};
    }
    if ((cc == ax_CC_L)) {
        return (ax_string){.ptr=(const ax_u8*)"l", .len=1};
    }
    if ((cc == ax_CC_GE)) {
        return (ax_string){.ptr=(const ax_u8*)"ge", .len=2};
    }
    if ((cc == ax_CC_LE)) {
        return (ax_string){.ptr=(const ax_u8*)"le", .len=2};
    }
    if ((cc == ax_CC_G)) {
        return (ax_string){.ptr=(const ax_u8*)"g", .len=1};
    }
    return (ax_string){.ptr=(const ax_u8*)"??", .len=2};
}

struct ax_MachInstVec ax_new_mach_inst_vec(void) {
    return ((struct ax_MachInstVec){.data=((struct ax_MachInst*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_MachInstVec_push(struct ax_MachInstVec* self, struct ax_MachInst inst) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_MachInst* new_data = ((struct ax_MachInst*)(ax_alloc((new_cap * 80))));
        if ((self->data != ((struct ax_MachInst*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 80));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = inst;
    self->len = (self->len + 1);
    return idx;
}

void ax_TypeTable_type_size_and_align(struct ax_TypeTable table, ax_u32 type_id, ax_u32* out_size, ax_u32* out_align) {
    if ((type_id == ((ax_u32)(0)))) {
        ((out_size)[0]) = ((ax_u32)(8));
        ((out_align)[0]) = ((ax_u32)(8));
        return;
    }
    struct ax_TypeEntry* entry = &(((table.entries.data)[type_id]));
    if ((entry->size != ((ax_u32)(0)))) {
        ((out_size)[0]) = entry->size;
        ((out_align)[0]) = entry->align;
        return;
    }
    if ((entry->kind == ax_TYPE_KIND_PRIMITIVE)) {
        ((out_size)[0]) = entry->size;
        ((out_align)[0]) = entry->align;
        return;
    } else if ((entry->kind == ax_TYPE_KIND_STRUCT)) {
        ax_u32 struct_idx = entry->extra;
        struct ax_StructInfo info = ((table.structs.data)[struct_idx]);
        ax_u32 offset = ((ax_u32)(0));
        ax_u32 max_align = ((ax_u32)(1));
        ax_i64 i = ((ax_i64)(0));
        while ((i < info.fields.len)) {
            struct ax_StructField f = ((info.fields.data)[i]);
            ax_u32 f_size = ((ax_u32)(0));
            ax_u32 f_align = ((ax_u32)(0));
            ax_TypeTable_type_size_and_align(table, f.type_id, &(f_size), &(f_align));
            if ((f_align == ((ax_u32)(0)))) {
                f_align = ((ax_u32)(8));
            }
            offset = (((offset + f_align) - ((ax_u32)(1))) & (~(f_align - ((ax_u32)(1)))));
            offset = (offset + f_size);
            if ((f_align > max_align)) {
                max_align = f_align;
            }
            i = (i + 1);
        }
        ax_u32 size = (((offset + max_align) - ((ax_u32)(1))) & (~(max_align - ((ax_u32)(1)))));
        if ((size == ((ax_u32)(0)))) {
            size = ((ax_u32)(8));
        }
        entry->size = size;
        entry->align = max_align;
        ((out_size)[0]) = size;
        ((out_align)[0]) = max_align;
    } else {
        {
            ((out_size)[0]) = ((ax_u32)(8));
            ((out_align)[0]) = ((ax_u32)(8));
        }
    }
}

ax_u32 ax_TypeTable_field_offset(struct ax_TypeTable table, ax_u32 struct_type_id, ax_u32 field_idx) {
    if ((struct_type_id == ((ax_u32)(0)))) {
        return (field_idx * ((ax_u32)(8)));
    }
    struct ax_TypeEntry entry = ((table.entries.data)[struct_type_id]);
    if ((entry.kind != ax_TYPE_KIND_STRUCT)) {
        return (field_idx * ((ax_u32)(8)));
    }
    ax_u32 struct_idx = entry.extra;
    struct ax_StructInfo info = ((table.structs.data)[struct_idx]);
    ax_u32 offset = ((ax_u32)(0));
    ax_i64 i = ((ax_i64)(0));
    while ((i < info.fields.len)) {
        struct ax_StructField f = ((info.fields.data)[i]);
        ax_u32 f_size = ((ax_u32)(0));
        ax_u32 f_align = ((ax_u32)(0));
        ax_TypeTable_type_size_and_align(table, f.type_id, &(f_size), &(f_align));
        if ((f_align == ((ax_u32)(0)))) {
            f_align = ((ax_u32)(8));
        }
        offset = (((offset + f_align) - ((ax_u32)(1))) & (~(f_align - ((ax_u32)(1)))));
        if ((((ax_u32)(i)) == field_idx)) {
            return offset;
        }
        offset = (offset + f_size);
        i = (i + 1);
    }
    return (field_idx * ((ax_u32)(8)));
}

ax_u8 ax_abi_int_arg_reg(ax_string abi, ax_i64 idx) {
    if (ax_str_eq(abi, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
        return ax_get_win64_arg_reg(idx);
    }
    return ax_get_sysv_arg_reg(idx);
}

ax_u8 ax_abi_return_reg(ax_string abi) {
    return ax_REG_RAX;
}

ax_u32 ax_InstructionSelector_next_vreg(struct ax_InstructionSelector* sel) {
    sel->max_vreg = (sel->max_vreg + ((ax_u32)(1)));
    return sel->max_vreg;
}

ax_u32 ax_InstructionSelector_next_label(struct ax_InstructionSelector* sel) {
    sel->max_label = (sel->max_label + ((ax_u32)(1)));
    return sel->max_label;
}

ax_u32 ax_InstructionSelector_get_register_type(struct ax_InstructionSelector* sel, ax_u32 reg) {
    if ((reg == 0)) {
        return 0;
    }
    ax_i64 i = ((ax_i64)(0));
    while ((i < sel->fn_ptr->insts.len)) {
        struct ax_AirInst* inst = &(((sel->fn_ptr->insts.data)[i]));
        if ((inst->dest == reg)) {
            return ((ax_u32)(inst->type_id));
        }
        i = (i + 1);
    }
    return 0;
}

void ax_AirInst_select_cmp(struct ax_AirInst* inst, ax_u8 cc, struct ax_MachInstVec* out_insts) {
    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CMP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SETCC, .cc=cc, .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOVZX_B, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
}

void ax_InstructionSelector_select_inst(struct ax_InstructionSelector* sel, struct ax_AirInst* inst, struct ax_MachInstVec* out_insts) {
    ax_u16 op = inst->opcode;
    if ((op == ax_OP_NOP)) {
        return;
    }
    if ((op == ax_OP_ICONST)) {
        ax_bool is_str = (inst->type_id == ((ax_u16)(12)));
        if (((inst->src1 == ((ax_u32)(0))) && (!is_str))) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_XOR_ZERO, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        } else {
            {
                ax_i64 val = ((ax_i64)(inst->src1));
                ax_u32 vreg_flag = ((ax_u32)(0));
                if (is_str) {
                    vreg_flag = ((ax_u32)(2));
                }
                ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV_IMM, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=vreg_flag, .label=((ax_u32)(0)), .imm=val}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            }
        }
    } else if ((op == ax_OP_FCONST)) {
        ax_i64 val = ((ax_i64)(inst->src1));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV_IMM, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=val}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if (((op == ax_OP_COPY) || (op == ax_OP_MOVE))) {
        if (((sel->param_idx_processed < ((ax_i32)(sel->fn_ptr->params.len))) && (inst->src1 == ((ax_u32)((sel->param_idx_processed + 1)))))) {
            ax_i32 param_idx = sel->param_idx_processed;
            sel->param_idx_processed = (sel->param_idx_processed + 1);
            ax_u8 phys = ax_abi_int_arg_reg(sel->abi_name, ((ax_i64)(param_idx)));
            if ((phys != ax_REG_NONE)) {
                ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=phys, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                return;
            }
        }
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_IADD)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_ADD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_ISUB)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SUB, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_IMUL)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_IMUL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_IDIV)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CQO, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_IDIV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_IMOD)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CQO, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_IDIV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RDX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_NEG)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_NEG, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_NOT)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CMP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SETCC, .cc=ax_CC_E, .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOVZX_B, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_EQ)) {
        ax_AirInst_select_cmp(inst, ax_CC_E, out_insts);
    } else if ((op == ax_OP_NE)) {
        ax_AirInst_select_cmp(inst, ax_CC_NE, out_insts);
    } else if ((op == ax_OP_LT)) {
        ax_AirInst_select_cmp(inst, ax_CC_L, out_insts);
    } else if ((op == ax_OP_LE)) {
        ax_AirInst_select_cmp(inst, ax_CC_LE, out_insts);
    } else if ((op == ax_OP_GT)) {
        ax_AirInst_select_cmp(inst, ax_CC_G, out_insts);
    } else if ((op == ax_OP_GE)) {
        ax_AirInst_select_cmp(inst, ax_CC_GE, out_insts);
    } else if ((op == ax_OP_AND)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_AND, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_OR)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_OR, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_XOR)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_XOR, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_SHL)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RCX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SHL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_SHR)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RCX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SAR, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_RETURN)) {
        ax_string fn_name_ret = ax_InternPool_get(sel->pool, sel->fn_ptr->name);
        if ((inst->src1 != ((ax_u32)(0)))) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        }
        if (ax_str_eq(fn_name_ret, (ax_string){.ptr=(const ax_u8*)"main", .len=4})) {
            if ((inst->src1 != ((ax_u32)(0)))) {
                ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_PUSH, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            }
            if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
                ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SUB, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            }
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=(-((ax_i64)(5)))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
                ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_ADD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            }
            if ((inst->src1 != ((ax_u32)(0)))) {
                ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_POP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            }
        }
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_RET, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_JUMP)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_JMP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=inst->src1, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_BRANCH)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_TEST, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_JCC, .cc=ax_CC_NE, .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=inst->src2, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_JMP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=inst->dest, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_SYSCALL)) {
        ax_u32 arg_start = inst->src2;
        ax_u32 arg_count = ((ax_u32)(0));
        if ((arg_start < ((ax_u32)(sel->fn_ptr->extras.len)))) {
            arg_count = ((sel->fn_ptr->extras.data)[arg_start]);
        }
        if ((arg_count > ((ax_u32)(0)))) {
            ax_u32 sys_num_reg = ((sel->fn_ptr->extras.data)[(arg_start + ((ax_u32)(1)))]);
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=sys_num_reg, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            ax_u32 i = ((ax_u32)(1));
            while ((i < arg_count)) {
                ax_u32 arg_reg = ((sel->fn_ptr->extras.data)[((arg_start + ((ax_u32)(1))) + i)]);
                ax_u8 sys_phys = ((ax_u8)(0));
                if ((i == ((ax_u32)(1)))) {
                    sys_phys = ax_REG_RDI;
                } else if ((i == ((ax_u32)(2)))) {
                    sys_phys = ax_REG_RSI;
                } else if ((i == ((ax_u32)(3)))) {
                    sys_phys = ax_REG_RDX;
                } else if ((i == ((ax_u32)(4)))) {
                    sys_phys = ax_REG_R10;
                } else if ((i == ((ax_u32)(5)))) {
                    sys_phys = ax_REG_R8;
                } else if ((i == ((ax_u32)(6)))) {
                    sys_phys = ax_REG_R9;
                }
                if ((sys_phys != ((ax_u8)(0)))) {
                    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=sys_phys, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=arg_reg, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                }
                i = (i + ((ax_u32)(1)));
            }
        }
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SYSCALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        if ((inst->dest != ((ax_u32)(0)))) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        }
    } else if ((op == ax_OP_CALL)) {
        ax_u32 arg_start = inst->src2;
        ax_u32 arg_count = ((ax_u32)(0));
        if ((arg_start < ((ax_u32)(sel->fn_ptr->extras.len)))) {
            arg_count = ((sel->fn_ptr->extras.data)[arg_start]);
        }
        ax_string target_name = ax_x86_resolve_sym_name(((ax_u32)(inst->type_id)), sel->symbols, sel->pool);
        ax_i64 sym_imm = ((ax_i64)(inst->type_id));
        if ((ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"println", .len=7}) || ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"print", .len=5}))) {
            ax_bool is_println = ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"println", .len=7});
            if ((arg_count > ((ax_u32)(0)))) {
                ax_u32 first_arg_reg = ((sel->fn_ptr->extras.data)[(arg_start + ((ax_u32)(1)))]);
                ax_u32 arg_type = ax_InstructionSelector_get_register_type(sel, first_arg_reg);
                if (((arg_type == ((ax_u32)(3))) || (arg_type == ((ax_u32)(4))))) {
                    if (is_println) {
                        target_name = (ax_string){.ptr=(const ax_u8*)"ax_println_i64", .len=14};
                        sym_imm = (-((ax_i64)(11)));
                    } else {
                        {
                            target_name = (ax_string){.ptr=(const ax_u8*)"ax_print_i64", .len=12};
                            sym_imm = (-((ax_i64)(15)));
                        }
                    }
                } else if ((arg_type == ((ax_u32)(11)))) {
                    if (is_println) {
                        target_name = (ax_string){.ptr=(const ax_u8*)"ax_println_bool", .len=15};
                        sym_imm = (-((ax_i64)(13)));
                    } else {
                        {
                            target_name = (ax_string){.ptr=(const ax_u8*)"ax_print_bool", .len=13};
                            sym_imm = (-((ax_i64)(17)));
                        }
                    }
                } else if (((arg_type == ((ax_u32)(9))) || (arg_type == ((ax_u32)(10))))) {
                    if (is_println) {
                        target_name = (ax_string){.ptr=(const ax_u8*)"ax_println_f64", .len=14};
                        sym_imm = (-((ax_i64)(12)));
                    } else {
                        {
                            target_name = (ax_string){.ptr=(const ax_u8*)"ax_print_f64", .len=12};
                            sym_imm = (-((ax_i64)(16)));
                        }
                    }
                } else {
                    {
                        if (is_println) {
                            target_name = (ax_string){.ptr=(const ax_u8*)"ax_println_str", .len=14};
                            sym_imm = (-((ax_i64)(10)));
                        } else {
                            {
                                target_name = (ax_string){.ptr=(const ax_u8*)"ax_print_str", .len=12};
                                sym_imm = (-((ax_i64)(14)));
                            }
                        }
                    }
                }
            } else {
                {
                    if (is_println) {
                        target_name = (ax_string){.ptr=(const ax_u8*)"ax_println_str", .len=14};
                        sym_imm = (-((ax_i64)(10)));
                    } else {
                        {
                            target_name = (ax_string){.ptr=(const ax_u8*)"ax_print_str", .len=12};
                            sym_imm = (-((ax_i64)(14)));
                        }
                    }
                }
            }
        }
        ax_bool is_syscall = AX_FALSE;
        ax_i64 syscall_num = ((ax_i64)(0));
        if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"sysv", .len=4})) {
            if (ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"exit", .len=4})) {
                is_syscall = AX_TRUE;
                syscall_num = ((ax_i64)(60));
            } else if (ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"write", .len=5})) {
                is_syscall = AX_TRUE;
                syscall_num = ((ax_i64)(1));
            } else if (ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"mmap", .len=4})) {
                is_syscall = AX_TRUE;
                syscall_num = ((ax_i64)(9));
            } else if (ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"munmap", .len=6})) {
                is_syscall = AX_TRUE;
                syscall_num = ((ax_i64)(11));
            }
        }
        if (is_syscall) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV_IMM, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=syscall_num}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            ax_u32 i = ((ax_u32)(0));
            while ((i < arg_count)) {
                ax_u32 arg_reg = ((sel->fn_ptr->extras.data)[((arg_start + ((ax_u32)(1))) + i)]);
                ax_u8 sys_phys = ((ax_u8)(0));
                if ((i == ((ax_u32)(0)))) {
                    sys_phys = ax_REG_RDI;
                } else if ((i == ((ax_u32)(1)))) {
                    sys_phys = ax_REG_RSI;
                } else if ((i == ((ax_u32)(2)))) {
                    sys_phys = ax_REG_RDX;
                } else if ((i == ((ax_u32)(3)))) {
                    sys_phys = ax_REG_R10;
                } else if ((i == ((ax_u32)(4)))) {
                    sys_phys = ax_REG_R8;
                } else if ((i == ((ax_u32)(5)))) {
                    sys_phys = ax_REG_R9;
                }
                if ((sys_phys != ((ax_u8)(0)))) {
                    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=sys_phys, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=arg_reg, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                }
                i = (i + ((ax_u32)(1)));
            }
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SYSCALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            if ((inst->dest != ((ax_u32)(0)))) {
                ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            }
        } else {
            {
                if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
                    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SUB, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                }
                ax_u32 i = ((ax_u32)(0));
                while ((i < arg_count)) {
                    ax_u32 arg_reg = ((sel->fn_ptr->extras.data)[((arg_start + ((ax_u32)(1))) + i)]);
                    ax_u8 phys = ax_abi_int_arg_reg(sel->abi_name, ((ax_i64)(i)));
                    if ((phys != ax_REG_NONE)) {
                        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=phys, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=arg_reg, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                    }
                    i = (i + ((ax_u32)(1)));
                }
                ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=sym_imm}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
                    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_ADD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                }
                if ((inst->dest != ((ax_u32)(0)))) {
                    ax_u8 ret_reg = ax_abi_return_reg(sel->abi_name);
                    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ret_reg, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                }
            }
        }
    } else if ((op == ax_OP_ALLOC)) {
        ax_u32 size = ((ax_u32)(0));
        ax_u32 align = ((ax_u32)(0));
        ax_TypeTable_type_size_and_align(sel->table, ((ax_u32)(inst->type_id)), &(size), &(align));
        ax_u8 arg0 = ax_abi_int_arg_reg(sel->abi_name, 0);
        if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SUB, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        }
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV_IMM, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=arg0, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(size))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=(-((ax_i64)(1)))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_ADD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        }
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_SPAWN)) {
        ax_u8 arg0 = ax_abi_int_arg_reg(sel->abi_name, 0);
        ax_u8 arg1 = ax_abi_int_arg_reg(sel->abi_name, 1);
        ax_u8 arg2 = ax_abi_int_arg_reg(sel->abi_name, 2);
        if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SUB, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        }
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV_IMM, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=arg0, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(1)), .label=((ax_u32)(0)), .imm=((ax_i64)(inst->type_id))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_XOR_ZERO, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=arg1, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_XOR_ZERO, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=arg2, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=(-((ax_i64)(3)))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_ADD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        }
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RAX, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if (((op == ax_OP_FREE) || (op == ax_OP_DESTROY))) {
        ax_u8 arg0 = ax_abi_int_arg_reg(sel->abi_name, 0);
        if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SUB, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        }
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=arg0, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=(-((ax_i64)(2)))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        if (ax_str_eq(sel->abi_name, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_ADD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        }
    } else if ((op == ax_OP_LOAD)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_LOAD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_STORE)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_STORE, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src2, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_GET_FIELD)) {
        ax_u32 struct_type = ax_InstructionSelector_get_register_type(sel, inst->src1);
        ax_u32 disp = ax_TypeTable_field_offset(sel->table, struct_type, inst->src2);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_LOAD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(disp))})}));
    } else if ((op == ax_OP_SET_FIELD)) {
        ax_u32 struct_type = ax_InstructionSelector_get_register_type(sel, inst->src1);
        ax_u32 disp = ax_TypeTable_field_offset(sel->table, struct_type, inst->src2);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_STORE, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(disp))})}));
    } else if ((op == ax_OP_SET_FIELD)) {
        ax_u32 struct_type = ax_InstructionSelector_get_register_type(sel, inst->src1);
        ax_u32 disp = ax_TypeTable_field_offset(sel->table, struct_type, inst->src2);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_STORE, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(disp))})}));
    } else if ((op == ax_OP_MAKE_REF)) {
        ax_u32 lbl_null = ax_InstructionSelector_next_label(sel);
        ax_u32 lbl_end = ax_InstructionSelector_next_label(sel);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_TEST, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_JCC, .cc=((ax_u8)(0x04)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=lbl_null, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_u32 tmp_gen = ax_InstructionSelector_next_vreg(sel);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_LOAD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_gen, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=(-((ax_i64)(8)))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_AND, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_gen, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_gen, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(65535))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SHL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_gen, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(48))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_u32 tmp_ptr = ax_InstructionSelector_next_vreg(sel);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_AND, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(281474976710655))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_OR, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_gen, .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_JMP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=lbl_end, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_LABEL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=lbl_null, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_XOR_ZERO, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_LABEL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=lbl_end, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else if ((op == ax_OP_DEREF)) {
        ax_u32 lbl_ok_null = ax_InstructionSelector_next_label(sel);
        ax_u32 lbl_ok_uaf = ax_InstructionSelector_next_label(sel);
        ax_u32 tmp_ptr = ax_InstructionSelector_next_vreg(sel);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_AND, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(281474976710655))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_TEST, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_JCC, .cc=((ax_u8)(0x05)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=lbl_ok_null, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=(-((ax_i64)(8)))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_LABEL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=lbl_ok_null, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_u32 tmp_gen = ax_InstructionSelector_next_vreg(sel);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_LOAD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_gen, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=(-((ax_i64)(8)))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_AND, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_gen, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_gen, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(65535))})}));
        ax_u32 tmp_cap = ax_InstructionSelector_next_vreg(sel);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_cap, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->src1, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SAR, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_cap, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(48))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_AND, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_cap, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_cap, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(65535))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CMP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_cap, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_gen, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_JCC, .cc=((ax_u8)(0x04)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=lbl_ok_uaf, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_CALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=(-((ax_i64)(8)))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_LABEL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=lbl_ok_uaf, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=inst->dest, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_VREG, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=tmp_ptr, .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    } else {
        {
            ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_NOP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        }
    }
}

struct ax_MachInstVec ax_AirFunc_select_all(struct ax_AirFunc* f, ax_string abi, struct ax_TypeTable table, struct ax_SymbolTable symbols, struct ax_InternPool pool) {
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     select_all: starting actual_max_vreg loop\n", .len=54}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    ax_u32 actual_max_vreg = ((ax_u32)(0));
    ax_i64 vi = ((ax_i64)(0));
    while ((vi < f->insts.len)) {
        struct ax_AirInst* inst = &(((f->insts.data)[vi]));
        if ((inst->dest > actual_max_vreg)) {
            actual_max_vreg = inst->dest;
        }
        vi = (vi + 1);
    }
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     select_all: starting actual_max_label loop\n", .len=55}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    ax_u32 actual_max_label = ((ax_u32)(0));
    ax_i64 bi_init = ((ax_i64)(0));
    while ((bi_init < f->blocks.len)) {
        struct ax_BasicBlock* bb = &(((f->blocks.data)[bi_init]));
        if ((bb->id > actual_max_label)) {
            actual_max_label = bb->id;
        }
        bi_init = (bi_init + 1);
    }
    if ((actual_max_label < ((ax_u32)(1000000)))) {
        actual_max_label = ((ax_u32)(1000000));
    }
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     select_all: creating InstructionSelector\n", .len=53}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    struct ax_InstructionSelector sel = ((struct ax_InstructionSelector){.fn_ptr=f, .abi_name=abi, .table=table, .param_idx_processed=0, .symbols=symbols, .pool=pool, .max_vreg=actual_max_vreg, .max_label=actual_max_label});
    struct ax_MachInstVec result = ax_new_mach_inst_vec();
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     select_all: getting fn_name\n", .len=40}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    ax_string fn_name = ax_InternPool_get(pool, f->name);
    ax_bool is_main = ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"main", .len=4});
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     select_all: bi loop starting\n", .len=41}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    if ((f->blocks.len > 0)) {
        ax_i64 bi = ((ax_i64)(0));
        while ((bi < f->blocks.len)) {
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]       bi loop: bi=%d, blocks.len=%d\n", .len=44}).ptr, ((ax_i64)(bi)), ((ax_i64)(f->blocks.len)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            struct ax_BasicBlock* bb = &(((f->blocks.data)[bi]));
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]       bi loop: bb accessed, pushing MACH_LABEL for bb.id=%d\n", .len=68}).ptr, ((ax_i64)(bb->id)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_LABEL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_LABEL, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=bb->id, .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]       bi loop: block_instrs.len=%d, data=%p, start=%d, len=%d\n", .len=70}).ptr, ((ax_i64)(f->block_instrs.len)), ((ax_i64)(f->block_instrs.data)), ((ax_i64)(bb->instrs_start)), ((ax_i64)(bb->instrs_len)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            if ((is_main && (bi == ((ax_i64)(0))))) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]       bi loop: handling is_main\n", .len=40}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                if (ax_str_eq(abi, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
                    ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_SUB, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                }
                ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_CALL, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=(-((ax_i64)(4)))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                if (ax_str_eq(abi, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
                    ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_ADD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(32))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
                }
            }
            ax_i64 i = ((ax_i64)(0));
            while ((i < ((ax_i64)(bb->instrs_len)))) {
                ax_u32 inst_idx = ((f->block_instrs.data)[(((ax_i64)(bb->instrs_start)) + i)]);
                if ((inst_idx < ((ax_u32)(f->insts.len)))) {
                    struct ax_AirInst* inst = &(((f->insts.data)[((ax_i64)(inst_idx))]));
                    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     select_inst: inst_idx=%d, opcode=%d\n", .len=48}).ptr, ((ax_i64)(inst_idx)), ((ax_i64)(inst->opcode)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    fflush(((void*)(NULL)));
                    ax_InstructionSelector_select_inst(&(sel), inst, &(result));
                }
                i = (i + 1);
            }
            bi = (bi + 1);
        }
    } else {
        {
            ax_i64 i = ((ax_i64)(0));
            while ((i < f->insts.len)) {
                struct ax_AirInst* inst = &(((f->insts.data)[i]));
                ax_InstructionSelector_select_inst(&(sel), inst, &(result));
                i = (i + 1);
            }
        }
    }
    return result;
    /* skip destroy for value type result */
}

struct ax_LiveIntervalVec ax_new_live_interval_vec(void) {
    return ((struct ax_LiveIntervalVec){.data=((struct ax_LiveInterval*)(NULL)), .len=0, .cap=0});
}

ax_u32 ax_LiveIntervalVec_push(struct ax_LiveIntervalVec* self, struct ax_LiveInterval iv) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_LiveInterval* new_data = ((struct ax_LiveInterval*)(ax_alloc((new_cap * 12))));
        if ((self->data != ((struct ax_LiveInterval*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 12));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_u32 idx = ((ax_u32)(self->len));
    ((self->data)[self->len]) = iv;
    self->len = (self->len + 1);
    return idx;
}

ax_bool ax_is_two_operand_read(ax_u16 op) {
    if (((((((((((op == ax_MACH_ADD) || (op == ax_MACH_SUB)) || (op == ax_MACH_IMUL)) || (op == ax_MACH_AND)) || (op == ax_MACH_OR)) || (op == ax_MACH_XOR)) || (op == ax_MACH_SHL)) || (op == ax_MACH_SAR)) || (op == ax_MACH_NEG)) || (op == ax_MACH_NOT))) {
        return AX_TRUE;
    }
    if ((((op == ax_MACH_CMP) || (op == ax_MACH_TEST)) || (op == ax_MACH_STORE))) {
        return AX_TRUE;
    }
    return AX_FALSE;
}

struct ax_LiveIntervalVec ax_MachInst_compute_liveness(struct ax_MachInst* insts, ax_i64 insts_len) {
    ax_u32 max_vreg = ((ax_u32)(0));
    ax_i64 i = ((ax_i64)(0));
    while ((i < insts_len)) {
        struct ax_MachInst* inst = &(((insts)[i]));
        if (((inst->dst.kind == ax_OPND_VREG) && (inst->dst.vreg > max_vreg))) {
            max_vreg = inst->dst.vreg;
        }
        if (((inst->src1.kind == ax_OPND_VREG) && (inst->src1.vreg > max_vreg))) {
            max_vreg = inst->src1.vreg;
        }
        if (((inst->src2.kind == ax_OPND_VREG) && (inst->src2.vreg > max_vreg))) {
            max_vreg = inst->src2.vreg;
        }
        i = (i + 1);
    }
    ax_i64 map_size = ((ax_i64)((max_vreg + ((ax_u32)(1)))));
    struct ax_LiveInterval** interval_map = ((struct ax_LiveInterval**)(ax_alloc((map_size * 8))));
    memset(((ax_u8*)(interval_map)), ((ax_u8)(0)), (map_size * 8));
    i = ((ax_i64)(0));
    while ((i < insts_len)) {
        struct ax_MachInst* inst = &(((insts)[i]));
        if (((inst->dst.kind == ax_OPND_VREG) && (inst->dst.vreg != ((ax_u32)(0))))) {
            ax_u32 v = inst->dst.vreg;
            if ((((interval_map)[v]) == ((struct ax_LiveInterval*)(NULL)))) {
                struct ax_LiveInterval* iv = ((struct ax_LiveInterval*)(ax_alloc(12)));
                iv->vreg = v;
                iv->start = ((ax_i32)(i));
                iv->end = ((ax_i32)(i));
                ((interval_map)[v]) = iv;
            }
        }
        if (((inst->src1.kind == ax_OPND_VREG) && (inst->src1.vreg != ((ax_u32)(0))))) {
            ax_u32 v = inst->src1.vreg;
            struct ax_LiveInterval* iv = ((interval_map)[v]);
            if ((iv == ((struct ax_LiveInterval*)(NULL)))) {
                iv = ((struct ax_LiveInterval*)(ax_alloc(12)));
                iv->vreg = v;
                iv->start = ((ax_i32)(0));
                iv->end = ((ax_i32)(i));
                ((interval_map)[v]) = iv;
            } else if ((((ax_i32)(i)) > iv->end)) {
                iv->end = ((ax_i32)(i));
            }
        }
        if (((inst->src2.kind == ax_OPND_VREG) && (inst->src2.vreg != ((ax_u32)(0))))) {
            ax_u32 v = inst->src2.vreg;
            struct ax_LiveInterval* iv = ((interval_map)[v]);
            if ((iv == ((struct ax_LiveInterval*)(NULL)))) {
                iv = ((struct ax_LiveInterval*)(ax_alloc(12)));
                iv->vreg = v;
                iv->start = ((ax_i32)(0));
                iv->end = ((ax_i32)(i));
                ((interval_map)[v]) = iv;
            } else if ((((ax_i32)(i)) > iv->end)) {
                iv->end = ((ax_i32)(i));
            }
        }
        if (ax_is_two_operand_read(inst->op)) {
            if (((inst->dst.kind == ax_OPND_VREG) && (inst->dst.vreg != ((ax_u32)(0))))) {
                ax_u32 v = inst->dst.vreg;
                struct ax_LiveInterval* iv = ((interval_map)[v]);
                if ((iv == ((struct ax_LiveInterval*)(NULL)))) {
                    iv = ((struct ax_LiveInterval*)(ax_alloc(12)));
                    iv->vreg = v;
                    iv->start = ((ax_i32)(0));
                    iv->end = ((ax_i32)(i));
                    ((interval_map)[v]) = iv;
                } else if ((((ax_i32)(i)) > iv->end)) {
                    iv->end = ((ax_i32)(i));
                }
            }
        }
        i = (i + 1);
    }
    struct ax_LiveIntervalVec result = ax_new_live_interval_vec();
    ax_i64 v = ((ax_i64)(0));
    while ((v < map_size)) {
        struct ax_LiveInterval* iv = ((interval_map)[v]);
        if ((iv != ((struct ax_LiveInterval*)(NULL)))) {
            ax_LiveIntervalVec_push(&(result), ((iv)[0]));
            ax_free(((ax_u8*)(iv)));
        }
        v = (v + 1);
    }
    ax_free(((ax_u8*)(interval_map)));
    ax_i64 idx = ((ax_i64)(1));
    while ((idx < result.len)) {
        struct ax_LiveInterval key = ((result.data)[idx]);
        ax_i64 j = (idx - 1);
        while (((j >= 0) && (((result.data)[j]).start > key.start))) {
            ((result.data)[(j + 1)]) = ((result.data)[j]);
            j = (j - 1);
        }
        ((result.data)[(j + 1)]) = key;
        idx = (idx + 1);
    }
    return result;
    /* skip destroy for value type result */
}

ax_u8* ax_get_allocatable_gprs(ax_i64* out_len) {
    ax_u8* arr = ((ax_u8*)(ax_alloc(12)));
    ((arr)[0]) = ax_REG_RAX;
    ((arr)[1]) = ax_REG_RCX;
    ((arr)[2]) = ax_REG_RDX;
    ((arr)[3]) = ax_REG_RBX;
    ((arr)[4]) = ax_REG_RSI;
    ((arr)[5]) = ax_REG_RDI;
    ((arr)[6]) = ax_REG_R8;
    ((arr)[7]) = ax_REG_R9;
    ((arr)[8]) = ax_REG_R12;
    ((arr)[9]) = ax_REG_R13;
    ((arr)[10]) = ax_REG_R14;
    ((arr)[11]) = ax_REG_R15;
    ((out_len)[0]) = 12;
    return arr;
    ax_free(arr);
}

struct ax_RegAllocResult ax_LiveIntervalVec_graph_coloring_alloc(struct ax_LiveIntervalVec intervals, ax_u8* avail_regs, ax_i64 avail_regs_len) {
    if (((intervals.len == 0) || (avail_regs_len == 0))) {
        return ((struct ax_RegAllocResult){.allocs=((struct ax_RegAllocation*)(NULL)), .max_vreg=((ax_u32)(0)), .spill_count=((ax_i32)(0))});
    }
    ax_u32 max_vreg = ((ax_u32)(0));
    ax_i64 i = ((ax_i64)(0));
    while ((i < intervals.len)) {
        if ((((intervals.data)[i]).vreg > max_vreg)) {
            max_vreg = ((intervals.data)[i]).vreg;
        }
        i = (i + 1);
    }
    ax_i64 graph_size = ((ax_i64)((max_vreg + ((ax_u32)(1)))));
    struct ax_U32Vec* adj = ((struct ax_U32Vec*)(ax_alloc((graph_size * 24))));
    ax_i64 vi = ((ax_i64)(0));
    while ((vi < graph_size)) {
        ((adj)[vi]) = ax_new_u32_vec();
        vi = (vi + 1);
    }
    i = ((ax_i64)(0));
    while ((i < intervals.len)) {
        struct ax_LiveInterval* iv1 = &(((intervals.data)[i]));
        ax_i64 j = (i + 1);
        while ((j < intervals.len)) {
            struct ax_LiveInterval* iv2 = &(((intervals.data)[j]));
            if (((iv1->start <= iv2->end) && (iv2->start <= iv1->end))) {
                ax_U32Vec_append_unique(&(((adj)[iv1->vreg])), iv2->vreg);
                ax_U32Vec_append_unique(&(((adj)[iv2->vreg])), iv1->vreg);
            }
            j = (j + 1);
        }
        i = (i + 1);
    }
    ax_i32* degrees = ((ax_i32*)(ax_alloc((graph_size * 4))));
    ax_i64 di = ((ax_i64)(0));
    while ((di < graph_size)) {
        ((degrees)[di]) = ((ax_i32)(((adj)[di]).len));
        di = (di + 1);
    }
    ax_bool* removed = ((ax_bool*)(ax_alloc(graph_size)));
    memset(((ax_u8*)(removed)), ((ax_u8)(0)), graph_size);
    struct ax_U32Vec stack = ax_new_u32_vec();
    ax_i32 K = ((ax_i32)(avail_regs_len));
    ax_i64 simplified_count = ((ax_i64)(0));
    ax_i64 target_count = intervals.len;
    while ((simplified_count < target_count)) {
        ax_bool found = AX_FALSE;
        ax_u32 best_node = ((ax_u32)(0));
        ax_i64 iv_idx = ((ax_i64)(0));
        while ((iv_idx < intervals.len)) {
            ax_u32 vreg = ((intervals.data)[iv_idx]).vreg;
            if ((!((removed)[vreg]))) {
                if ((((degrees)[vreg]) < K)) {
                    best_node = vreg;
                    found = AX_TRUE;
                    break;
                }
            }
            iv_idx = (iv_idx + 1);
        }
        if (found) {
            ax_U32Vec_push(&(stack), best_node);
            ((removed)[best_node]) = AX_TRUE;
            simplified_count = (simplified_count + 1);
            ax_i64 n_idx = ((ax_i64)(0));
            while ((n_idx < ((adj)[best_node]).len)) {
                ax_u32 neighbor = ((((adj)[best_node]).data)[n_idx]);
                if ((!((removed)[neighbor]))) {
                    ((degrees)[neighbor]) = (((degrees)[neighbor]) - 1);
                }
                n_idx = (n_idx + 1);
            }
        } else {
            {
                ax_u32 spill_candidate = ((ax_u32)(0));
                ax_i32 max_degree = (-((ax_i32)(1)));
                ax_bool has_candidate = AX_FALSE;
                iv_idx = 0;
                while ((iv_idx < intervals.len)) {
                    ax_u32 vreg = ((intervals.data)[iv_idx]).vreg;
                    if ((!((removed)[vreg]))) {
                        if ((((degrees)[vreg]) > max_degree)) {
                            max_degree = ((degrees)[vreg]);
                            spill_candidate = vreg;
                            has_candidate = AX_TRUE;
                        }
                    }
                    iv_idx = (iv_idx + 1);
                }
                if ((!has_candidate)) {
                    break;
                }
                ax_U32Vec_push(&(stack), spill_candidate);
                ((removed)[spill_candidate]) = AX_TRUE;
                simplified_count = (simplified_count + 1);
                ax_i64 n_idx = ((ax_i64)(0));
                while ((n_idx < ((adj)[spill_candidate]).len)) {
                    ax_u32 neighbor = ((((adj)[spill_candidate]).data)[n_idx]);
                    if ((!((removed)[neighbor]))) {
                        ((degrees)[neighbor]) = (((degrees)[neighbor]) - 1);
                    }
                    n_idx = (n_idx + 1);
                }
            }
        }
    }
    ax_u8* assigned_colors = ((ax_u8*)(ax_alloc(graph_size)));
    memset(assigned_colors, ax_REG_NONE, graph_size);
    struct ax_RegAllocation* allocs = ((struct ax_RegAllocation*)(ax_alloc((graph_size * 12))));
    memset(((ax_u8*)(allocs)), ((ax_u8)(0)), (graph_size * 12));
    ax_i32 spill_count = ((ax_i32)(0));
    ax_i64 st_idx = (stack.len - 1);
    while ((st_idx >= 0)) {
        ax_u32 vreg = ((stack.data)[st_idx]);
        ax_bool* forbidden = ((ax_bool*)(ax_alloc(256)));
        memset(((ax_u8*)(forbidden)), ((ax_u8)(0)), 256);
        ax_i64 n_idx = ((ax_i64)(0));
        while ((n_idx < ((adj)[vreg]).len)) {
            ax_u32 neighbor = ((((adj)[vreg]).data)[n_idx]);
            ax_u8 color = ((assigned_colors)[neighbor]);
            if ((color != ax_REG_NONE)) {
                ((forbidden)[color]) = AX_TRUE;
            }
            n_idx = (n_idx + 1);
        }
        ax_bool colored = AX_FALSE;
        ax_i64 c_idx = ((ax_i64)(0));
        while ((c_idx < avail_regs_len)) {
            ax_u8 color = ((avail_regs)[c_idx]);
            if ((!((forbidden)[color]))) {
                ((assigned_colors)[vreg]) = color;
                ((allocs)[vreg]) = ((struct ax_RegAllocation){.vreg=vreg, .phys=color, .spilled=AX_FALSE, .spill_idx=((ax_i32)(0))});
                colored = AX_TRUE;
                break;
            }
            c_idx = (c_idx + 1);
        }
        ax_free(((ax_u8*)(forbidden)));
        if ((!colored)) {
            ((allocs)[vreg]) = ((struct ax_RegAllocation){.vreg=vreg, .phys=ax_REG_NONE, .spilled=AX_TRUE, .spill_idx=spill_count});
            spill_count = (spill_count + 1);
        }
        st_idx = (st_idx - 1);
    }
    ax_free(((ax_u8*)(removed)));
    ax_free(((ax_u8*)(degrees)));
    ax_free(assigned_colors);
    if ((stack.data != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(stack.data)));
    }
    ax_i64 f_idx = ((ax_i64)(0));
    while ((f_idx < graph_size)) {
        if ((((adj)[f_idx]).data != ((ax_u32*)(NULL)))) {
            ax_free(((ax_u8*)(((adj)[f_idx]).data)));
        }
        f_idx = (f_idx + 1);
    }
    ax_free(((ax_u8*)(adj)));
    return ((struct ax_RegAllocResult){.allocs=allocs, .max_vreg=max_vreg, .spill_count=spill_count});
}

ax_u8* ax_get_used_callee_saved(ax_string abi, struct ax_RegAllocation* allocs, ax_u32 max_vreg, ax_i64* out_len) {
    ax_bool* used = ((ax_bool*)(ax_alloc(16)));
    memset(((ax_u8*)(used)), ((ax_u8)(0)), 16);
    ax_i64 i = ((ax_i64)(0));
    while ((i <= ((ax_i64)(max_vreg)))) {
        ax_u8 phys = ((allocs)[i]).phys;
        if ((phys != ax_REG_NONE)) {
            if (ax_str_eq(abi, (ax_string){.ptr=(const ax_u8*)"win64", .len=5})) {
                if (ax_reg_is_win64_callee_saved(phys)) {
                    ((used)[phys]) = AX_TRUE;
                }
            } else {
                {
                    if (ax_reg_is_sysv_callee_saved(phys)) {
                        ((used)[phys]) = AX_TRUE;
                    }
                }
            }
        }
        i = (i + 1);
    }
    ax_i64 count = ((ax_i64)(0));
    ax_i64 r = ((ax_i64)(0));
    while ((r < 16)) {
        if (((used)[r])) {
            count = (count + 1);
        }
        r = (r + 1);
    }
    ax_u8* result = ((ax_u8*)(ax_alloc(count)));
    ax_i64 idx = ((ax_i64)(0));
    r = 0;
    while ((r < 16)) {
        if (((used)[r])) {
            ((result)[idx]) = ((ax_u8)(r));
            idx = (idx + 1);
        }
        r = (r + 1);
    }
    ax_free(((ax_u8*)(used)));
    ((out_len)[0]) = count;
    return result;
    ax_free(result);
}

struct ax_StackFrame ax_compute_frame(ax_u8* callee_saved, ax_i64 callee_saved_len, ax_i32 spill_count, ax_i32 local_bytes) {
    ax_i32 needed = ((spill_count * 8) + local_bytes);
    ax_i64 pushed_bytes = ((callee_saved_len + 2) * 8);
    ax_i32 total = needed;
    ax_i32 padding = ((ax_i32)(0));
    if ((((pushed_bytes + ((ax_i64)(total))) % 16) != 0)) {
        padding = ((ax_i32)((((ax_i64)(16)) - ((pushed_bytes + ((ax_i64)(total))) % 16))));
        total = (total + padding);
    }
    return ((struct ax_StackFrame){.callee_saved=callee_saved, .callee_saved_len=callee_saved_len, .spill_slots=spill_count, .local_bytes=local_bytes, .align_padding=padding, .total_size=total});
}

void ax_StackFrame_emit_prologue(struct ax_StackFrame* frame, struct ax_MachInstVec* out_insts) {
    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_PUSH, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RBP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_MOV, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RBP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    ax_i64 i = ((ax_i64)(0));
    while ((i < frame->callee_saved_len)) {
        ax_u8 reg = ((frame->callee_saved)[i]);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_PUSH, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=reg, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        i = (i + 1);
    }
    if ((frame->total_size > 0)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_SUB, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(frame->total_size))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    }
}

void ax_StackFrame_emit_epilogue(struct ax_StackFrame* frame, struct ax_MachInstVec* out_insts) {
    if ((frame->total_size > 0)) {
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_ADD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RSP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(frame->total_size))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
    }
    ax_i64 i = (frame->callee_saved_len - 1);
    while ((i >= 0)) {
        ax_u8 reg = ((frame->callee_saved)[i]);
        ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_POP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=reg, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
        i = (i - 1);
    }
    ax_MachInstVec_push(out_insts, ((struct ax_MachInst){.op=ax_MACH_POP, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RBP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_NONE, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))})}));
}

ax_u8 ax_get_dst_behavior(ax_u16 op) {
    if ((((((((op == ax_MACH_MOV) || (op == ax_MACH_MOV_IMM)) || (op == ax_MACH_XOR_ZERO)) || (op == ax_MACH_SETCC)) || (op == ax_MACH_MOVZX_B)) || (op == ax_MACH_POP)) || (op == ax_MACH_LOAD))) {
        return ax_DST_WRITE_ONLY;
    }
    if (((((((((((op == ax_MACH_ADD) || (op == ax_MACH_SUB)) || (op == ax_MACH_IMUL)) || (op == ax_MACH_NEG)) || (op == ax_MACH_NOT)) || (op == ax_MACH_AND)) || (op == ax_MACH_OR)) || (op == ax_MACH_XOR)) || (op == ax_MACH_SHL)) || (op == ax_MACH_SAR))) {
        return ax_DST_READ_WRITE;
    }
    if ((((op == ax_MACH_CMP) || (op == ax_MACH_TEST)) || (op == ax_MACH_STORE))) {
        return ax_DST_READ_ONLY;
    }
    return ax_DST_UNUSED;
}

struct ax_MachInstVec ax_MachInst_insert_spill_code(struct ax_MachInst* insts, ax_i64 insts_len, struct ax_RegAllocation* allocs, struct ax_StackFrame* frame) {
    struct ax_MachInstVec result = ax_new_mach_inst_vec();
    ax_i64 i = ((ax_i64)(0));
    while ((i < insts_len)) {
        struct ax_MachInst inst = ((insts)[i]);
        if ((inst.src1.kind == ax_OPND_VREG)) {
            ax_u32 v = inst.src1.vreg;
            struct ax_RegAllocation alloc = ((allocs)[v]);
            if (alloc.spilled) {
                ax_i32 offset = (-((alloc.spill_idx + 1) * 8));
                ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_LOAD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R10, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RBP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(offset))})}));
                inst.src1 = ((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R10, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))});
            }
        }
        if ((inst.src2.kind == ax_OPND_VREG)) {
            ax_u32 v = inst.src2.vreg;
            struct ax_RegAllocation alloc = ((allocs)[v]);
            if (alloc.spilled) {
                ax_i32 offset = (-((alloc.spill_idx + 1) * 8));
                ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_LOAD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R10, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RBP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(offset))})}));
                inst.src2 = ((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R10, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))});
            }
        }
        ax_bool dst_spilled = AX_FALSE;
        struct ax_RegAllocation dst_alloc = ((struct ax_RegAllocation){.vreg=((ax_u32)(0)), .phys=((ax_u8)(0)), .spilled=AX_FALSE, .spill_idx=((ax_i32)(0))});
        if ((inst.dst.kind == ax_OPND_VREG)) {
            ax_u32 v = inst.dst.vreg;
            struct ax_RegAllocation alloc = ((allocs)[v]);
            if (alloc.spilled) {
                dst_spilled = AX_TRUE;
                dst_alloc = alloc;
            }
        }
        if (dst_spilled) {
            ax_u8 behavior = ax_get_dst_behavior(inst.op);
            ax_i32 offset = (-((dst_alloc.spill_idx + 1) * 8));
            if ((behavior == ax_DST_READ_ONLY)) {
                ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_LOAD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R11, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RBP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(offset))})}));
                inst.dst = ((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R11, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))});
                ax_MachInstVec_push(&(result), inst);
            } else if ((behavior == ax_DST_READ_WRITE)) {
                ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_LOAD, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R11, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RBP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(offset))})}));
                inst.dst = ((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R11, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))});
                ax_MachInstVec_push(&(result), inst);
                ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_STORE, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RBP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R11, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(offset))})}));
            } else if ((behavior == ax_DST_WRITE_ONLY)) {
                inst.dst = ((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R11, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))});
                ax_MachInstVec_push(&(result), inst);
                ax_MachInstVec_push(&(result), ((struct ax_MachInst){.op=ax_MACH_STORE, .cc=((ax_u8)(0)), .padding=((ax_u8)(0)), .dst=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_RBP, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src1=((struct ax_MachOperand){.kind=ax_OPND_PHYS, .phys=ax_REG_R11, .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(0))}), .src2=((struct ax_MachOperand){.kind=ax_OPND_IMM, .phys=((ax_u8)(0)), .padding=((ax_u16)(0)), .vreg=((ax_u32)(0)), .label=((ax_u32)(0)), .imm=((ax_i64)(offset))})}));
            } else {
                {
                    ax_MachInstVec_push(&(result), inst);
                }
            }
        } else {
            {
                ax_MachInstVec_push(&(result), inst);
            }
        }
        i = (i + 1);
    }
    return result;
    /* skip destroy for value type result */
}

ax_string ax_to_byte_reg(ax_u8 reg) {
    ax_u8 hw = ax_reg_hw_reg(reg);
    if (ax_reg_is_gpr(reg)) {
        if ((hw == 0)) {
            return (ax_string){.ptr=(const ax_u8*)"al", .len=2};
        }
        if ((hw == 1)) {
            return (ax_string){.ptr=(const ax_u8*)"cl", .len=2};
        }
        if ((hw == 2)) {
            return (ax_string){.ptr=(const ax_u8*)"dl", .len=2};
        }
        if ((hw == 3)) {
            return (ax_string){.ptr=(const ax_u8*)"bl", .len=2};
        }
        if ((hw == 4)) {
            return (ax_string){.ptr=(const ax_u8*)"spl", .len=3};
        }
        if ((hw == 5)) {
            return (ax_string){.ptr=(const ax_u8*)"bpl", .len=3};
        }
        if ((hw == 6)) {
            return (ax_string){.ptr=(const ax_u8*)"sil", .len=3};
        }
        if ((hw == 7)) {
            return (ax_string){.ptr=(const ax_u8*)"dil", .len=3};
        }
        if ((hw == 8)) {
            return (ax_string){.ptr=(const ax_u8*)"r8b", .len=3};
        }
        if ((hw == 9)) {
            return (ax_string){.ptr=(const ax_u8*)"r9b", .len=3};
        }
        if ((hw == 10)) {
            return (ax_string){.ptr=(const ax_u8*)"r10b", .len=4};
        }
        if ((hw == 11)) {
            return (ax_string){.ptr=(const ax_u8*)"r11b", .len=4};
        }
        if ((hw == 12)) {
            return (ax_string){.ptr=(const ax_u8*)"r12b", .len=4};
        }
        if ((hw == 13)) {
            return (ax_string){.ptr=(const ax_u8*)"r13b", .len=4};
        }
        if ((hw == 14)) {
            return (ax_string){.ptr=(const ax_u8*)"r14b", .len=4};
        }
        if ((hw == 15)) {
            return (ax_string){.ptr=(const ax_u8*)"r15b", .len=4};
        }
    }
    return (ax_string){.ptr=(const ax_u8*)"al", .len=2};
}

ax_string ax_MachOperand_format_operand(struct ax_MachOperand op, ax_string format, struct ax_RegAllocation* allocs) {
    if ((op.kind == ax_OPND_PHYS)) {
        return ax_reg_to_str(op.phys);
    } else if ((op.kind == ax_OPND_VREG)) {
        ax_u32 r = op.vreg;
        struct ax_RegAllocation alloc = ((allocs)[r]);
        return ax_reg_to_str(alloc.phys);
    } else if ((op.kind == ax_OPND_IMM)) {
        return ax_format_int(op.imm);
    } else if ((op.kind == ax_OPND_LABEL)) {
        if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
            return ax_str_concat((ax_string){.ptr=(const ax_u8*)"L_b_", .len=4}, ax_format_int(((ax_i64)(op.label))));
        }
        return ax_str_concat((ax_string){.ptr=(const ax_u8*)".L_b_", .len=5}, ax_format_int(((ax_i64)(op.label))));
    }
    return (ax_string){.ptr=(const ax_u8*)"none", .len=4};
}

void ax_emit_inst(void* file, struct ax_MachInst inst, ax_string fn_name, struct ax_RegAllocation* allocs, ax_string format, struct ax_SymbolTable symbols, struct ax_InternPool pool) {
    ax_u16 op = inst.op;
    if ((op == ax_MACH_NOP)) {
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    nop\n", .len=8}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_LABEL)) {
        ax_u32 lbl_id = inst.dst.label;
        if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"L_b_%d:\n", .len=8}).ptr, ((ax_i64)(lbl_id)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)".L_b_%d:\n", .len=9}).ptr, ((ax_i64)(lbl_id)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_MACH_RET)) {
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    ret\n", .len=8}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_MOV)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    mov %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_MOV_IMM)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        if ((inst.src1.vreg == ((ax_u32)(1)))) {
            ax_string sym_name = ax_x86_resolve_sym_name(((ax_u32)(inst.src1.imm)), symbols, pool);
            if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    mov %s, OFFSET %s\n", .len=22}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(sym_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else {
                {
                    fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    mov %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(sym_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                }
            }
        } else {
            {
                ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    mov %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_MACH_XOR_ZERO)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    xor %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_ADD)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    add %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_SUB)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    sub %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_IMUL)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    imul %s, %s\n", .len=16}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_IDIV)) {
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    idiv %s\n", .len=12}).ptr, ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_CQO)) {
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    cqo\n", .len=8}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_NEG)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    neg %s\n", .len=11}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_NOT)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    not %s\n", .len=11}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_AND)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    and %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_OR)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    or %s, %s\n", .len=14}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_XOR)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    xor %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_SHL)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        if ((inst.src1.kind == ax_OPND_IMM)) {
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    shl %s, %d\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), inst.src1.imm, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    shl %s, cl\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_MACH_SAR)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        if ((inst.src1.kind == ax_OPND_IMM)) {
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    sar %s, %d\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), inst.src1.imm, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    sar %s, cl\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_MACH_CMP)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    cmp %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_TEST)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    test %s, %s\n", .len=16}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_SETCC)) {
        ax_u8 dst_reg = ax_REG_NONE;
        if ((inst.dst.kind == ax_OPND_PHYS)) {
            dst_reg = inst.dst.phys;
        } else if ((inst.dst.kind == ax_OPND_VREG)) {
            dst_reg = ((allocs)[inst.dst.vreg]).phys;
        }
        ax_string byte_reg = ax_to_byte_reg(dst_reg);
        ax_string cc_str = ax_cond_code_to_str(inst.cc);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    set%s %s\n", .len=13}).ptr, ((ax_i64)(((ax_u8*)(cc_str.ptr)))), ((ax_i64)(((ax_u8*)(byte_reg.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_MOVZX_B)) {
        ax_u8 src_reg = ax_REG_NONE;
        if ((inst.src1.kind == ax_OPND_PHYS)) {
            src_reg = inst.src1.phys;
        } else if ((inst.src1.kind == ax_OPND_VREG)) {
            src_reg = ((allocs)[inst.src1.vreg]).phys;
        }
        ax_string byte_reg = ax_to_byte_reg(src_reg);
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    movzx %s, %s\n", .len=17}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(byte_reg.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_PUSH)) {
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    push %s\n", .len=12}).ptr, ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_POP)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    pop %s\n", .len=11}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_JMP)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    jmp %s\n", .len=11}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_JCC)) {
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_string cc_str = ax_cond_code_to_str(inst.cc);
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    j%s %s\n", .len=11}).ptr, ((ax_i64)(((ax_u8*)(cc_str.ptr)))), ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_MACH_LOAD)) {
        ax_string base = ax_MachOperand_format_operand(inst.src1, format, allocs);
        ax_i64 disp = ((ax_i64)(0));
        if ((inst.src2.kind == ax_OPND_IMM)) {
            disp = inst.src2.imm;
        }
        ax_string addr_str = (ax_string){.ptr=(const ax_u8*)"", .len=0};
        if ((disp == ((ax_i64)(0)))) {
            addr_str = ax_str_concat((ax_string){.ptr=(const ax_u8*)"[", .len=1}, ax_str_concat(base, (ax_string){.ptr=(const ax_u8*)"]", .len=1}));
        } else if ((disp > ((ax_i64)(0)))) {
            addr_str = ax_str_concat((ax_string){.ptr=(const ax_u8*)"[", .len=1}, ax_str_concat(base, ax_str_concat((ax_string){.ptr=(const ax_u8*)" + ", .len=3}, ax_str_concat(ax_format_int(disp), (ax_string){.ptr=(const ax_u8*)"]", .len=1}))));
        } else {
            {
                addr_str = ax_str_concat((ax_string){.ptr=(const ax_u8*)"[", .len=1}, ax_str_concat(base, ax_str_concat((ax_string){.ptr=(const ax_u8*)" - ", .len=3}, ax_str_concat(ax_format_int((((ax_i64)(0)) - disp)), (ax_string){.ptr=(const ax_u8*)"]", .len=1}))));
            }
        }
        ax_string dst = ax_MachOperand_format_operand(inst.dst, format, allocs);
        if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    mov %s, qword ptr %s\n", .len=25}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(addr_str.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    mov %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(dst.ptr)))), ((ax_i64)(((ax_u8*)(addr_str.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_MACH_STORE)) {
        ax_string base = ax_MachOperand_format_operand(inst.dst, format, allocs);
        ax_i64 disp = ((ax_i64)(0));
        if ((inst.src2.kind == ax_OPND_IMM)) {
            disp = inst.src2.imm;
        }
        ax_string addr_str = (ax_string){.ptr=(const ax_u8*)"", .len=0};
        if ((disp == ((ax_i64)(0)))) {
            addr_str = ax_str_concat((ax_string){.ptr=(const ax_u8*)"[", .len=1}, ax_str_concat(base, (ax_string){.ptr=(const ax_u8*)"]", .len=1}));
        } else if ((disp > ((ax_i64)(0)))) {
            addr_str = ax_str_concat((ax_string){.ptr=(const ax_u8*)"[", .len=1}, ax_str_concat(base, ax_str_concat((ax_string){.ptr=(const ax_u8*)" + ", .len=3}, ax_str_concat(ax_format_int(disp), (ax_string){.ptr=(const ax_u8*)"]", .len=1}))));
        } else {
            {
                addr_str = ax_str_concat((ax_string){.ptr=(const ax_u8*)"[", .len=1}, ax_str_concat(base, ax_str_concat((ax_string){.ptr=(const ax_u8*)" - ", .len=3}, ax_str_concat(ax_format_int((((ax_i64)(0)) - disp)), (ax_string){.ptr=(const ax_u8*)"]", .len=1}))));
            }
        }
        ax_string src = ax_MachOperand_format_operand(inst.src1, format, allocs);
        if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    mov qword ptr %s, %s\n", .len=25}).ptr, ((ax_i64)(((ax_u8*)(addr_str.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    mov %s, %s\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(addr_str.ptr)))), ((ax_i64)(((ax_u8*)(src.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    } else if ((op == ax_MACH_CALL)) {
        if ((inst.src1.imm <= ((ax_i64)(0)))) {
            if ((inst.src1.imm == (-((ax_i64)(1))))) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    call malloc\n", .len=16}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if ((inst.src1.imm == (-((ax_i64)(2))))) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    call free\n", .len=14}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if ((inst.src1.imm == (-((ax_i64)(3))))) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    call ax_actor_spawn\n", .len=24}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if ((inst.src1.imm == (-((ax_i64)(4))))) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    call __ax_runtime_init\n", .len=27}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if ((inst.src1.imm == (-((ax_i64)(5))))) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    call __ax_runtime_shutdown\n", .len=31}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if ((inst.src1.imm == (-((ax_i64)(8))))) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    call abort\n", .len=15}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else {
                {
                    fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    call %s\n", .len=12}).ptr, ((ax_i64)(((ax_u8*)(fn_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                }
            }
        } else {
            {
                ax_string sym_name = ax_x86_resolve_sym_name(((ax_u32)(inst.src1.imm)), symbols, pool);
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"    call %s\n", .len=12}).ptr, ((ax_i64)(((ax_u8*)(sym_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            }
        }
    }
}

void ax_emit_function(void* file, ax_string fn_name, struct ax_MachInst* insts, ax_i64 insts_len, struct ax_RegAllocation* allocs, struct ax_StackFrame* frame, ax_string format, struct ax_SymbolTable symbols, struct ax_InternPool pool) {
    if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"%s PROC\n", .len=8}).ptr, ((ax_i64)(((ax_u8*)(fn_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else {
        {
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"%s:\n", .len=4}).ptr, ((ax_i64)(((ax_u8*)(fn_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        }
    }
    struct ax_MachInstVec prologue = ax_new_mach_inst_vec();
    ax_StackFrame_emit_prologue(frame, &(prologue));
    ax_i64 i = ((ax_i64)(0));
    while ((i < prologue.len)) {
        ax_emit_inst(file, ((prologue.data)[i]), fn_name, allocs, format, symbols, pool);
        i = (i + 1);
    }
    if ((prologue.data != ((struct ax_MachInst*)(NULL)))) {
        ax_free(((ax_u8*)(prologue.data)));
    }
    i = ((ax_i64)(0));
    while ((i < insts_len)) {
        struct ax_MachInst inst = ((insts)[i]);
        if ((inst.op == ax_MACH_RET)) {
            struct ax_MachInstVec epilogue = ax_new_mach_inst_vec();
            ax_StackFrame_emit_epilogue(frame, &(epilogue));
            ax_i64 j = ((ax_i64)(0));
            while ((j < epilogue.len)) {
                ax_emit_inst(file, ((epilogue.data)[j]), fn_name, allocs, format, symbols, pool);
                j = (j + 1);
            }
            if ((epilogue.data != ((struct ax_MachInst*)(NULL)))) {
                ax_free(((ax_u8*)(epilogue.data)));
            }
        }
        ax_emit_inst(file, inst, fn_name, allocs, format, symbols, pool);
        i = (i + 1);
    }
    if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"%s ENDP\n", .len=8}).ptr, ((ax_i64)(((ax_u8*)(fn_name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    }
}

static ax_u8* ax_x86_str_to_null_terminated(ax_string s) {
    ax_i64 len = ax_str_len(s);
    ax_u8* buf = ((ax_u8*)(ax_alloc((len + 1))));
    memcpy(buf, ((ax_u8*)(s.ptr)), len);
    ((buf)[len]) = ((ax_u8)(0));
    return buf;
    ax_free(buf);
}

ax_bool ax_compile_native_asm(ax_string output_asm_file, struct ax_AirModule mod, struct ax_SymbolTable symbols, struct ax_InternPool pool, struct ax_TypeTable table, ax_string format) {
    ax_u8* output_asm_file_nt = ax_x86_str_to_null_terminated(output_asm_file);
    void* file = fopen(((void*)(output_asm_file_nt)), (const char*)((ax_string){.ptr=(const ax_u8*)"wb", .len=2}).ptr);
    ax_free(output_asm_file_nt);
    if ((file == ((void*)(NULL)))) {
        return AX_FALSE;
    }
    if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"fasm", .len=4})) {
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"; Generated by AXIOM Compiler (FASM format)\n", .len=44}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"format ELF64 executable\n", .len=24}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"segment readable executable\n\n", .len=29}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"; Generated by AXIOM Compiler (WinAsm / MASM format)\n", .len=53}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)".code\n\n", .len=7}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else {
        {
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"; Generated by AXIOM Compiler (NASM format)\n", .len=44}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"bits 64\n", .len=8}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"section .text\n\n", .len=15}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        }
    }
    ax_i64 fi = ((ax_i64)(0));
    while ((fi < mod.funcs.len)) {
        struct ax_AirFunc* fn_ptr = &(((mod.funcs.data)[fi]));
        if (fn_ptr->is_extern) {
            ax_string name = ax_x86_resolve_sym_name(fn_ptr->sym_id, symbols, pool);
            if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"fasm", .len=4})) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"extrn %s\n", .len=9}).ptr, ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"EXTERN %s:PROC\n", .len=15}).ptr, ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else {
                {
                    fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"extern %s\n", .len=10}).ptr, ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                }
            }
        }
        fi = (fi + 1);
    }
    fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"\n", .len=1}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fi = 0;
    while ((fi < mod.funcs.len)) {
        struct ax_AirFunc* fn_ptr = &(((mod.funcs.data)[fi]));
        if ((!fn_ptr->is_extern)) {
            ax_string name = ax_x86_resolve_sym_name(fn_ptr->sym_id, symbols, pool);
            if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"fasm", .len=4})) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"public %s\n", .len=10}).ptr, ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
                fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"PUBLIC %s\n", .len=10}).ptr, ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else {
                {
                    fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"global %s\n", .len=10}).ptr, ((ax_i64)(((ax_u8*)(name.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                }
            }
        }
        fi = (fi + 1);
    }
    fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"\n", .len=1}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    ax_string abi = (ax_string){.ptr=(const ax_u8*)"win64", .len=5};
    fi = 0;
    while ((fi < mod.funcs.len)) {
        struct ax_AirFunc* fn_ptr = &(((mod.funcs.data)[fi]));
        if ((!fn_ptr->is_extern)) {
            ax_string fn_name = ax_x86_resolve_sym_name(fn_ptr->sym_id, symbols, pool);
            struct ax_MachInstVec mach_insts = ax_AirFunc_select_all(fn_ptr, abi, table, symbols, pool);
            struct ax_LiveIntervalVec intervals = ax_MachInst_compute_liveness(mach_insts.data, mach_insts.len);
            ax_i64 gpr_len = ((ax_i64)(0));
            ax_u8* avail_gprs = ax_get_allocatable_gprs(&(gpr_len));
            struct ax_RegAllocResult alloc_res = ax_LiveIntervalVec_graph_coloring_alloc(intervals, avail_gprs, gpr_len);
            ax_i64 cs_len = ((ax_i64)(0));
            ax_u8* cs_regs = ax_get_used_callee_saved(abi, alloc_res.allocs, alloc_res.max_vreg, &(cs_len));
            struct ax_StackFrame frame = ax_compute_frame(cs_regs, cs_len, alloc_res.spill_count, ((ax_i32)(0)));
            struct ax_MachInstVec final_insts = ax_MachInst_insert_spill_code(mach_insts.data, mach_insts.len, alloc_res.allocs, &(frame));
            ax_emit_function(file, fn_name, final_insts.data, final_insts.len, alloc_res.allocs, &(frame), format, symbols, pool);
            fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"\n", .len=1}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            ax_free(cs_regs);
            ax_free(avail_gprs);
            ax_free(alloc_res.allocs);
            if ((intervals.data != ((struct ax_LiveInterval*)(NULL)))) {
                ax_free(((ax_u8*)(intervals.data)));
            }
            if ((mach_insts.data != ((struct ax_MachInst*)(NULL)))) {
                ax_free(((ax_u8*)(mach_insts.data)));
            }
            if ((final_insts.data != ((struct ax_MachInst*)(NULL)))) {
                ax_free(((ax_u8*)(final_insts.data)));
            }
        }
        fi = (fi + 1);
    }
    if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"winasm", .len=6})) {
        fprintf(file, (const char*)((ax_string){.ptr=(const ax_u8*)"END\n", .len=4}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    }
    fclose(file);
    return AX_TRUE;
}

struct ax_ByteVec ax_new_byte_vec(void) {
    return ((struct ax_ByteVec){.data=((ax_u8*)(NULL)), .len=0, .cap=0});
}

void ax_ByteVec_push_byte(struct ax_ByteVec* self, ax_u8 b) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        ax_u8* new_data = ((ax_u8*)(ax_alloc(new_cap)));
        if ((self->data != ((ax_u8*)(NULL)))) {
            memcpy(new_data, self->data, self->len);
            ax_free(self->data);
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = b;
    self->len = (self->len + 1);
}

void ax_ByteVec_push_bytes(struct ax_ByteVec* self, ax_u8* bytes, ax_i64 bytes_len) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < bytes_len)) {
        ax_ByteVec_push_byte(self, ((bytes)[i]));
        i = (i + 1);
    }
}

void ax_ByteVec_push_u16_le(struct ax_ByteVec* self, ax_u16 val) {
    ax_ByteVec_push_byte(self, ((ax_u8)((val & ((ax_u16)(0xFF))))));
    ax_ByteVec_push_byte(self, ((ax_u8)(((val >> ((ax_u16)(8))) & ((ax_u16)(0xFF))))));
}

void ax_ByteVec_push_u32_le(struct ax_ByteVec* self, ax_u32 val) {
    ax_ByteVec_push_byte(self, ((ax_u8)((val & ((ax_u32)(0xFF))))));
    ax_ByteVec_push_byte(self, ((ax_u8)(((val >> ((ax_u32)(8))) & ((ax_u32)(0xFF))))));
    ax_ByteVec_push_byte(self, ((ax_u8)(((val >> ((ax_u32)(16))) & ((ax_u32)(0xFF))))));
    ax_ByteVec_push_byte(self, ((ax_u8)(((val >> ((ax_u32)(24))) & ((ax_u32)(0xFF))))));
}

void ax_ByteVec_push_u64_le(struct ax_ByteVec* self, ax_u64 val) {
    ax_ByteVec_push_u32_le(self, ((ax_u32)((val & ((ax_u64)(0xFFFFFFFF))))));
    ax_ByteVec_push_u32_le(self, ((ax_u32)(((val >> ((ax_u64)(32))) & ((ax_u64)(0xFFFFFFFF))))));
}

ax_u8 ax_encode_rex(ax_bool w, ax_bool r, ax_bool x, ax_bool b) {
    ax_u8 rex = ax_REX_BASE;
    if (w) {
        rex = (rex | ax_REX_W);
    }
    if (r) {
        rex = (rex | ax_REX_R);
    }
    if (x) {
        rex = (rex | ax_REX_X);
    }
    if (b) {
        rex = (rex | ax_REX_B);
    }
    return rex;
}

void ax_x86_encode_modrm_rr(ax_u8 reg, ax_u8 rm, ax_u8* out_modrm, ax_u8* out_rex, ax_bool* out_need_rex) {
    ax_u8 reg_field = ax_reg_hw_reg(reg);
    ax_u8 rm_field = ax_reg_hw_reg(rm);
    ((out_modrm)[0]) = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (reg_field << ((ax_u8)(3)))) | rm_field);
    ax_bool reg_need = ax_reg_needs_rex(reg);
    ax_bool rm_need = ax_reg_needs_rex(rm);
    if ((reg_need || rm_need)) {
        ((out_need_rex)[0]) = AX_TRUE;
        ax_u8 rex = ax_REX_BASE;
        if (reg_need) {
            rex = (rex | ax_REX_R);
        }
        if (rm_need) {
            rex = (rex | ax_REX_B);
        }
        ((out_rex)[0]) = rex;
    } else {
        {
            ((out_need_rex)[0]) = AX_FALSE;
            ((out_rex)[0]) = ((ax_u8)(0));
        }
    }
}

void ax_x86_encode_modrm_rm(ax_u8 reg, ax_u8 base, ax_i32 disp, struct ax_ByteVec* buf) {
    ax_u8 base_field = ax_reg_hw_reg(base);
    ax_u8 reg_field = ax_reg_hw_reg(reg);
    ax_u8 rsp_field = ax_reg_hw_reg(ax_REG_RSP);
    ax_u8 rbp_field = ax_reg_hw_reg(ax_REG_RBP);
    ax_bool needs_sib = (base_field == rsp_field);
    ax_bool needs_disp8_for_rbp = ((base_field == rbp_field) && (disp == 0));
    ax_u8 mod = ax_MOD_INDIRECT;
    if (needs_disp8_for_rbp) {
        mod = ax_MOD_DISP8;
    } else if ((disp == 0)) {
        mod = ax_MOD_INDIRECT;
    } else if (((disp >= (-128)) && (disp <= 127))) {
        mod = ax_MOD_DISP8;
    } else {
        {
            mod = ax_MOD_DISP32;
        }
    }
    ax_u8 rex = ((ax_u8)(0));
    ax_bool reg_need = ax_reg_needs_rex(reg);
    ax_bool base_need = ax_reg_needs_rex(base);
    if ((reg_need || base_need)) {
        rex = ax_REX_BASE;
        if (reg_need) {
            rex = (rex | ax_REX_R);
        }
        if (base_need) {
            rex = (rex | ax_REX_B);
        }
    }
    if ((rex != ((ax_u8)(0)))) {
        ax_ByteVec_push_byte(buf, rex);
    }
    if (needs_sib) {
        ax_u8 modrm = (((mod << ((ax_u8)(6))) | (reg_field << ((ax_u8)(3)))) | ((ax_u8)(0x04)));
        ax_u8 sib = (((((ax_u8)(0)) << ((ax_u8)(6))) | (((ax_u8)(0x04)) << ((ax_u8)(3)))) | base_field);
        ax_ByteVec_push_byte(buf, modrm);
        ax_ByteVec_push_byte(buf, sib);
    } else {
        {
            ax_u8 modrm = (((mod << ((ax_u8)(6))) | (reg_field << ((ax_u8)(3)))) | base_field);
            ax_ByteVec_push_byte(buf, modrm);
        }
    }
    if ((mod == ax_MOD_DISP8)) {
        ax_ByteVec_push_byte(buf, ((ax_u8)(disp)));
    } else if ((mod == ax_MOD_DISP32)) {
        ax_ByteVec_push_u32_le(buf, ((ax_u32)(disp)));
    }
}

void ax_x86_encode_modrm_rip(ax_u8 reg, ax_i32 disp32, struct ax_ByteVec* buf) {
    ax_u8 rex = ((ax_u8)(0));
    if (ax_reg_needs_rex(reg)) {
        rex = (ax_REX_BASE | ax_REX_R);
    }
    if ((rex != ((ax_u8)(0)))) {
        ax_ByteVec_push_byte(buf, rex);
    }
    ax_u8 reg_field = ax_reg_hw_reg(reg);
    ax_u8 modrm = (((ax_MOD_INDIRECT << ((ax_u8)(6))) | (reg_field << ((ax_u8)(3)))) | ((ax_u8)(0x05)));
    ax_ByteVec_push_byte(buf, modrm);
    ax_ByteVec_push_u32_le(buf, ((ax_u32)(disp32)));
}

void ax_x86_encode_modrm_sib(ax_u8 reg, ax_u8 base, ax_u8 index, ax_u8 scale, ax_i32 disp, struct ax_ByteVec* buf) {
    ax_u8 scale_bits = ((ax_u8)(0));
    if ((scale == ((ax_u8)(1)))) {
        scale_bits = ((ax_u8)(0));
    } else if ((scale == ((ax_u8)(2)))) {
        scale_bits = ((ax_u8)(1));
    } else if ((scale == ((ax_u8)(4)))) {
        scale_bits = ((ax_u8)(2));
    } else if ((scale == ((ax_u8)(8)))) {
        scale_bits = ((ax_u8)(3));
    } else {
        {
            scale_bits = ((ax_u8)(0));
        }
    }
    ax_u8 base_field = ax_reg_hw_reg(base);
    ax_u8 rbp_field = ax_reg_hw_reg(ax_REG_RBP);
    ax_u8 reg_field = ax_reg_hw_reg(reg);
    ax_u8 index_field = ax_reg_hw_reg(index);
    ax_bool needs_disp8_for_rbp = ((base_field == rbp_field) && (disp == 0));
    ax_u8 mod = ax_MOD_INDIRECT;
    if (needs_disp8_for_rbp) {
        mod = ax_MOD_DISP8;
    } else if ((disp == 0)) {
        mod = ax_MOD_INDIRECT;
    } else if (((disp >= (-128)) && (disp <= 127))) {
        mod = ax_MOD_DISP8;
    } else {
        {
            mod = ax_MOD_DISP32;
        }
    }
    ax_u8 rex = ((ax_u8)(0));
    ax_bool reg_need = ax_reg_needs_rex(reg);
    ax_bool base_need = ax_reg_needs_rex(base);
    ax_bool index_need = ax_reg_needs_rex(index);
    if (((reg_need || base_need) || index_need)) {
        rex = ax_REX_BASE;
        if (reg_need) {
            rex = (rex | ax_REX_R);
        }
        if (index_need) {
            rex = (rex | ax_REX_X);
        }
        if (base_need) {
            rex = (rex | ax_REX_B);
        }
    }
    if ((rex != ((ax_u8)(0)))) {
        ax_ByteVec_push_byte(buf, rex);
    }
    ax_u8 modrm = (((mod << ((ax_u8)(6))) | (reg_field << ((ax_u8)(3)))) | ((ax_u8)(0x04)));
    ax_u8 sib = (((scale_bits << ((ax_u8)(6))) | (index_field << ((ax_u8)(3)))) | base_field);
    ax_ByteVec_push_byte(buf, modrm);
    ax_ByteVec_push_byte(buf, sib);
    if ((mod == ax_MOD_DISP8)) {
        ax_ByteVec_push_byte(buf, ((ax_u8)(disp)));
    } else if ((mod == ax_MOD_DISP32)) {
        ax_ByteVec_push_u32_le(buf, ((ax_u32)(disp)));
    }
}

void ax_ByteVec_x86_encode_ret(struct ax_ByteVec* buf) {
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xC3)));
}

void ax_ByteVec_x86_encode_nop(struct ax_ByteVec* buf) {
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x90)));
}

void ax_ByteVec_x86_encode_int3(struct ax_ByteVec* buf) {
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xCC)));
}

void ax_x86_encode_push(ax_u8 reg, struct ax_ByteVec* buf) {
    if (ax_reg_needs_rex(reg)) {
        ax_ByteVec_push_byte(buf, (ax_REX_BASE | ax_REX_B));
    }
    ax_ByteVec_push_byte(buf, (((ax_u8)(0x50)) + ax_reg_hw_reg(reg)));
}

void ax_x86_encode_pop(ax_u8 reg, struct ax_ByteVec* buf) {
    if (ax_reg_needs_rex(reg)) {
        ax_ByteVec_push_byte(buf, (ax_REX_BASE | ax_REX_B));
    }
    ax_ByteVec_push_byte(buf, (((ax_u8)(0x58)) + ax_reg_hw_reg(reg)));
}

void ax_x86_encode_mov_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 modrm = ((ax_u8)(0));
    ax_u8 rex = ((ax_u8)(0));
    ax_bool need_rex = AX_FALSE;
    ax_x86_encode_modrm_rr(src, dst, &(modrm), &(rex), &(need_rex));
    if (need_rex) {
        rex = (rex | ax_REX_W);
        ax_ByteVec_push_byte(buf, rex);
    } else {
        {
            ax_ByteVec_push_byte(buf, ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE));
        }
    }
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x89)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_mov_ri(ax_u8 dst, ax_i32 imm, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(dst));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(0)) << ((ax_u8)(3)))) | ax_reg_hw_reg(dst));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xC7)));
    ax_ByteVec_push_byte(buf, modrm);
    ax_ByteVec_push_u32_le(buf, ((ax_u32)(imm)));
}

void ax_x86_encode_mov_ri64(ax_u8 dst, ax_i64 imm, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(dst));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, (((ax_u8)(0xB8)) + ax_reg_hw_reg(dst)));
    ax_ByteVec_push_u64_le(buf, ((ax_u64)(imm)));
}

void ax_x86_encode_add_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 modrm = ((ax_u8)(0));
    ax_u8 rex = ((ax_u8)(0));
    ax_bool need_rex = AX_FALSE;
    ax_x86_encode_modrm_rr(src, dst, &(modrm), &(rex), &(need_rex));
    if (need_rex) {
        rex = (rex | ax_REX_W);
    } else {
        {
            rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE);
        }
    }
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x01)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_add_ri(ax_u8 dst, ax_i32 imm, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(dst));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(0)) << ((ax_u8)(3)))) | ax_reg_hw_reg(dst));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x81)));
    ax_ByteVec_push_byte(buf, modrm);
    ax_ByteVec_push_u32_le(buf, ((ax_u32)(imm)));
}

void ax_x86_encode_sub_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 modrm = ((ax_u8)(0));
    ax_u8 rex = ((ax_u8)(0));
    ax_bool need_rex = AX_FALSE;
    ax_x86_encode_modrm_rr(src, dst, &(modrm), &(rex), &(need_rex));
    if (need_rex) {
        rex = (rex | ax_REX_W);
    } else {
        {
            rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE);
        }
    }
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x29)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_sub_ri(ax_u8 dst, ax_i32 imm, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(dst));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(5)) << ((ax_u8)(3)))) | ax_reg_hw_reg(dst));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x81)));
    ax_ByteVec_push_byte(buf, modrm);
    ax_ByteVec_push_u32_le(buf, ((ax_u32)(imm)));
}

void ax_x86_encode_imul_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 modrm = ((ax_u8)(0));
    ax_u8 rex = ((ax_u8)(0));
    ax_bool need_rex = AX_FALSE;
    ax_x86_encode_modrm_rr(dst, src, &(modrm), &(rex), &(need_rex));
    if (need_rex) {
        rex = (rex | ax_REX_W);
    } else {
        {
            rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE);
        }
    }
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x0F)));
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xAF)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_ByteVec_x86_encode_cqo(struct ax_ByteVec* buf) {
    ax_ByteVec_push_byte(buf, ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE));
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x99)));
}

void ax_x86_encode_idiv_r(ax_u8 divisor, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(divisor));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(7)) << ((ax_u8)(3)))) | ax_reg_hw_reg(divisor));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xF7)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_neg_r(ax_u8 reg, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(reg));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(3)) << ((ax_u8)(3)))) | ax_reg_hw_reg(reg));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xF7)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_not_r(ax_u8 reg, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(reg));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(2)) << ((ax_u8)(3)))) | ax_reg_hw_reg(reg));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xF7)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_cmp_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 modrm = ((ax_u8)(0));
    ax_u8 rex = ((ax_u8)(0));
    ax_bool need_rex = AX_FALSE;
    ax_x86_encode_modrm_rr(src, dst, &(modrm), &(rex), &(need_rex));
    if (need_rex) {
        rex = (rex | ax_REX_W);
    } else {
        {
            rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE);
        }
    }
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x39)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_cmp_ri(ax_u8 reg, ax_i32 imm, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(reg));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(7)) << ((ax_u8)(3)))) | ax_reg_hw_reg(reg));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x81)));
    ax_ByteVec_push_byte(buf, modrm);
    ax_ByteVec_push_u32_le(buf, ((ax_u32)(imm)));
}

void ax_x86_encode_test_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 modrm = ((ax_u8)(0));
    ax_u8 rex = ((ax_u8)(0));
    ax_bool need_rex = AX_FALSE;
    ax_x86_encode_modrm_rr(src, dst, &(modrm), &(rex), &(need_rex));
    if (need_rex) {
        rex = (rex | ax_REX_W);
    } else {
        {
            rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE);
        }
    }
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x85)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_and_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 modrm = ((ax_u8)(0));
    ax_u8 rex = ((ax_u8)(0));
    ax_bool need_rex = AX_FALSE;
    ax_x86_encode_modrm_rr(src, dst, &(modrm), &(rex), &(need_rex));
    if (need_rex) {
        rex = (rex | ax_REX_W);
    } else {
        {
            rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE);
        }
    }
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x21)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_or_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 modrm = ((ax_u8)(0));
    ax_u8 rex = ((ax_u8)(0));
    ax_bool need_rex = AX_FALSE;
    ax_x86_encode_modrm_rr(src, dst, &(modrm), &(rex), &(need_rex));
    if (need_rex) {
        rex = (rex | ax_REX_W);
    } else {
        {
            rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE);
        }
    }
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x09)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_xor_rr(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 modrm = ((ax_u8)(0));
    ax_u8 rex = ((ax_u8)(0));
    ax_bool need_rex = AX_FALSE;
    ax_x86_encode_modrm_rr(src, dst, &(modrm), &(rex), &(need_rex));
    if (need_rex) {
        rex = (rex | ax_REX_W);
    } else {
        {
            rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, AX_FALSE);
        }
    }
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x31)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_shl_cl(ax_u8 reg, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(reg));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(4)) << ((ax_u8)(3)))) | ax_reg_hw_reg(reg));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xD3)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_sar_cl(ax_u8 reg, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(reg));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(7)) << ((ax_u8)(3)))) | ax_reg_hw_reg(reg));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xD3)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_xor_zero(ax_u8 reg, struct ax_ByteVec* buf) {
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (ax_reg_hw_reg(reg) << ((ax_u8)(3)))) | ax_reg_hw_reg(reg));
    if (ax_reg_needs_rex(reg)) {
        ax_u8 rex = ((ax_REX_BASE | ax_REX_R) | ax_REX_B);
        ax_ByteVec_push_byte(buf, rex);
    }
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x31)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_setcc(ax_u8 cc, ax_u8 dst, struct ax_ByteVec* buf) {
    if ((ax_reg_needs_rex(dst) || (ax_reg_hw_reg(dst) >= ((ax_u8)(4))))) {
        ax_u8 rex = ax_REX_BASE;
        if (ax_reg_needs_rex(dst)) {
            rex = (rex | ax_REX_B);
        }
        ax_ByteVec_push_byte(buf, rex);
    }
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(0)) << ((ax_u8)(3)))) | ax_reg_hw_reg(dst));
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x0F)));
    ax_ByteVec_push_byte(buf, (((ax_u8)(0x90)) + cc));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_movzx_br(ax_u8 dst, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, ax_reg_needs_rex(dst), AX_FALSE, ax_reg_needs_rex(src));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (ax_reg_hw_reg(dst) << ((ax_u8)(3)))) | ax_reg_hw_reg(src));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x0F)));
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xB6)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_jmp_rel32(ax_i32 rel32, struct ax_ByteVec* buf) {
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xE9)));
    ax_ByteVec_push_u32_le(buf, ((ax_u32)(rel32)));
}

void ax_x86_encode_jcc_rel32(ax_u8 cc, ax_i32 rel32, struct ax_ByteVec* buf) {
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x0F)));
    ax_ByteVec_push_byte(buf, (((ax_u8)(0x80)) + cc));
    ax_ByteVec_push_u32_le(buf, ((ax_u32)(rel32)));
}

void ax_x86_encode_call_rel32(ax_i32 rel32, struct ax_ByteVec* buf) {
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xE8)));
    ax_ByteVec_push_u32_le(buf, ((ax_u32)(rel32)));
}

void ax_x86_encode_call_r(ax_u8 reg, struct ax_ByteVec* buf) {
    if (ax_reg_needs_rex(reg)) {
        ax_ByteVec_push_byte(buf, (ax_REX_BASE | ax_REX_B));
    }
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(2)) << ((ax_u8)(3)))) | ax_reg_hw_reg(reg));
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xFF)));
    ax_ByteVec_push_byte(buf, modrm);
}

void ax_x86_encode_lea(ax_u8 dst, ax_u8 base, ax_i32 disp, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, ax_reg_needs_rex(dst), AX_FALSE, ax_reg_needs_rex(base));
    struct ax_ByteVec temp_buf = ax_new_byte_vec();
    ax_x86_encode_modrm_rm(dst, base, disp, &(temp_buf));
    ax_u8 final_rex = rex;
    ax_i64 start_idx = ((ax_i64)(0));
    if ((temp_buf.len > ((ax_i64)(0)))) {
        ax_u8 first_byte = ((temp_buf.data)[0]);
        if (((first_byte & ((ax_u8)(0xF0))) == ((ax_u8)(0x40)))) {
            final_rex = (final_rex | (first_byte & ((ax_u8)(0x0F))));
            start_idx = 1;
        }
    }
    ax_ByteVec_push_byte(buf, final_rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x8D)));
    ax_i64 i = start_idx;
    while ((i < temp_buf.len)) {
        ax_ByteVec_push_byte(buf, ((temp_buf.data)[i]));
        i = (i + 1);
    }
    if ((temp_buf.data != ((ax_u8*)(NULL)))) {
        ax_free(temp_buf.data);
    }
}

void ax_x86_encode_mov_load(ax_u8 dst, ax_u8 base, ax_i32 disp, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, ax_reg_needs_rex(dst), AX_FALSE, ax_reg_needs_rex(base));
    struct ax_ByteVec temp_buf = ax_new_byte_vec();
    ax_x86_encode_modrm_rm(dst, base, disp, &(temp_buf));
    ax_u8 final_rex = rex;
    ax_i64 start_idx = ((ax_i64)(0));
    if ((temp_buf.len > ((ax_i64)(0)))) {
        ax_u8 first_byte = ((temp_buf.data)[0]);
        if (((first_byte & ((ax_u8)(0xF0))) == ((ax_u8)(0x40)))) {
            final_rex = (final_rex | (first_byte & ((ax_u8)(0x0F))));
            start_idx = 1;
        }
    }
    ax_ByteVec_push_byte(buf, final_rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x8B)));
    ax_i64 i = start_idx;
    while ((i < temp_buf.len)) {
        ax_ByteVec_push_byte(buf, ((temp_buf.data)[i]));
        i = (i + 1);
    }
    if ((temp_buf.data != ((ax_u8*)(NULL)))) {
        ax_free(temp_buf.data);
    }
}

void ax_x86_encode_mov_store(ax_u8 base, ax_i32 disp, ax_u8 src, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, ax_reg_needs_rex(src), AX_FALSE, ax_reg_needs_rex(base));
    struct ax_ByteVec temp_buf = ax_new_byte_vec();
    ax_x86_encode_modrm_rm(src, base, disp, &(temp_buf));
    ax_u8 final_rex = rex;
    ax_i64 start_idx = ((ax_i64)(0));
    if ((temp_buf.len > ((ax_i64)(0)))) {
        ax_u8 first_byte = ((temp_buf.data)[0]);
        if (((first_byte & ((ax_u8)(0xF0))) == ((ax_u8)(0x40)))) {
            final_rex = (final_rex | (first_byte & ((ax_u8)(0x0F))));
            start_idx = 1;
        }
    }
    ax_ByteVec_push_byte(buf, final_rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x89)));
    ax_i64 i = start_idx;
    while ((i < temp_buf.len)) {
        ax_ByteVec_push_byte(buf, ((temp_buf.data)[i]));
        i = (i + 1);
    }
    if ((temp_buf.data != ((ax_u8*)(NULL)))) {
        ax_free(temp_buf.data);
    }
}

void ax_ByteVec_x86_encode_syscall(struct ax_ByteVec* buf) {
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x0F)));
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x05)));
}

void ax_x86_encode_lea_rip(ax_u8 dst, ax_i32 disp32, struct ax_ByteVec* buf) {
    ax_u8 rex = (ax_REX_BASE | ax_REX_W);
    if (ax_reg_needs_rex(dst)) {
        rex = (rex | ax_REX_R);
    }
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0x8D)));
    ax_u8 reg_field = ax_reg_hw_reg(dst);
    ax_u8 modrm = (((ax_MOD_INDIRECT << ((ax_u8)(6))) | (reg_field << ((ax_u8)(3)))) | ((ax_u8)(0x05)));
    ax_ByteVec_push_byte(buf, modrm);
    ax_ByteVec_push_u32_le(buf, ((ax_u32)(disp32)));
}

void ax_x86_encode_shl_imm(ax_u8 reg, ax_u8 imm, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(reg));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(4)) << ((ax_u8)(3)))) | ax_reg_hw_reg(reg));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xC1)));
    ax_ByteVec_push_byte(buf, modrm);
    ax_ByteVec_push_byte(buf, imm);
}

void ax_x86_encode_sar_imm(ax_u8 reg, ax_u8 imm, struct ax_ByteVec* buf) {
    ax_u8 rex = ax_encode_rex(AX_TRUE, AX_FALSE, AX_FALSE, ax_reg_needs_rex(reg));
    ax_u8 modrm = (((ax_MOD_REG_DIRECT << ((ax_u8)(6))) | (((ax_u8)(7)) << ((ax_u8)(3)))) | ax_reg_hw_reg(reg));
    ax_ByteVec_push_byte(buf, rex);
    ax_ByteVec_push_byte(buf, ((ax_u8)(0xC1)));
    ax_ByteVec_push_byte(buf, modrm);
    ax_ByteVec_push_byte(buf, imm);
}

struct ax_RelocationVec ax_new_relocation_vec(void) {
    return ((struct ax_RelocationVec){.data=((struct ax_Relocation*)(NULL)), .len=0, .cap=0});
}

void ax_RelocationVec_push_reloc(struct ax_RelocationVec* self, struct ax_Relocation r) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_Relocation* new_data = ((struct ax_Relocation*)(ax_alloc((new_cap * 24))));
        if ((self->data != ((struct ax_Relocation*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 24));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = r;
    self->len = (self->len + 1);
}

struct ax_FixupVec ax_new_fixup_vec(void) {
    return ((struct ax_FixupVec){.data=((struct ax_Fixup*)(NULL)), .len=0, .cap=0});
}

void ax_FixupVec_push_fixup(struct ax_FixupVec* self, struct ax_Fixup f) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_Fixup* new_data = ((struct ax_Fixup*)(ax_alloc((new_cap * 16))));
        if ((self->data != ((struct ax_Fixup*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 16));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = f;
    self->len = (self->len + 1);
}

struct ax_LabelMap ax_new_label_map(void) {
    return ((struct ax_LabelMap){.keys=((ax_u32*)(NULL)), .values=((ax_i64*)(NULL)), .len=0, .cap=0});
}

void ax_LabelMap_label_map_set(struct ax_LabelMap* self, ax_u32 key, ax_i64 val) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->len)) {
        if ((((self->keys)[i]) == key)) {
            ((self->values)[i]) = val;
            return;
        }
        i = (i + 1);
    }
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        ax_u32* new_keys = ((ax_u32*)(ax_alloc((new_cap * 4))));
        ax_i64* new_values = ((ax_i64*)(ax_alloc((new_cap * 8))));
        if ((self->keys != ((ax_u32*)(NULL)))) {
            memcpy(((ax_u8*)(new_keys)), ((ax_u8*)(self->keys)), (self->len * 4));
            memcpy(((ax_u8*)(new_values)), ((ax_u8*)(self->values)), (self->len * 8));
            ax_free(((ax_u8*)(self->keys)));
            ax_free(((ax_u8*)(self->values)));
        }
        self->keys = new_keys;
        self->values = new_values;
        self->cap = new_cap;
    }
    ((self->keys)[self->len]) = key;
    ((self->values)[self->len]) = val;
    self->len = (self->len + 1);
}

ax_bool ax_LabelMap_label_map_get(struct ax_LabelMap* self, ax_u32 key, ax_i64* out_val) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->len)) {
        if ((((self->keys)[i]) == key)) {
            ((out_val)[0]) = ((self->values)[i]);
            return AX_TRUE;
        }
        i = (i + 1);
    }
    return AX_FALSE;
}

struct ax_MachEmitter ax_new_mach_emitter(void) {
    return ((struct ax_MachEmitter){.code=ax_new_byte_vec(), .relocs=ax_new_relocation_vec(), .labels=ax_new_label_map(), .fixups=ax_new_fixup_vec()});
}

void ax_MachEmitter_free_mach_emitter(struct ax_MachEmitter* self) {
    if ((self->code.data != ((ax_u8*)(NULL)))) {
        ax_free(self->code.data);
    }
    if ((self->relocs.data != ((struct ax_Relocation*)(NULL)))) {
        ax_free(((ax_u8*)(self->relocs.data)));
    }
    if ((self->labels.keys != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(self->labels.keys)));
    }
    if ((self->labels.values != ((ax_i64*)(NULL)))) {
        ax_free(((ax_u8*)(self->labels.values)));
    }
    if ((self->fixups.data != ((struct ax_Fixup*)(NULL)))) {
        ax_free(((ax_u8*)(self->fixups.data)));
    }
}

ax_u8 ax_MachOperand_emitter_resolve_reg(struct ax_MachOperand op, struct ax_RegAllocation* allocs) {
    if ((op.kind == ax_OPND_PHYS)) {
        return op.phys;
    } else if ((op.kind == ax_OPND_VREG)) {
        ax_u32 r = op.vreg;
        struct ax_RegAllocation alloc = ((allocs)[r]);
        return alloc.phys;
    }
    return ax_REG_NONE;
}

void ax_MachEmitter_emit_mach_inst(struct ax_MachEmitter* e, struct ax_MachInst inst, struct ax_RegAllocation* allocs) {
    ax_u16 op = inst.op;
    struct ax_ByteVec* buf = &(e->code);
    if ((op == ax_MACH_NOP)) {
        ax_ByteVec_x86_encode_nop(buf);
    } else if ((op == ax_MACH_LABEL)) {
        ax_LabelMap_label_map_set(&(e->labels), inst.dst.label, buf->len);
    } else if ((op == ax_MACH_RET)) {
        ax_ByteVec_x86_encode_ret(buf);
    } else if ((op == ax_MACH_MOV)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        if (((dst != ax_REG_NONE) && (src != ax_REG_NONE))) {
            ax_x86_encode_mov_rr(dst, src, buf);
        }
    } else if ((op == ax_MACH_MOV_IMM)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        if ((dst != ax_REG_NONE)) {
            ax_i64 imm = inst.src1.imm;
            if ((inst.src1.vreg == ((ax_u32)(1)))) {
                ax_i64 offset = (buf->len + ((ax_i64)(2)));
                ax_RelocationVec_push_reloc(&(e->relocs), ((struct ax_Relocation){.offset=offset, .kind=ax_RELOC_ABS64, .padding1=((ax_u8)(0)), .padding2=((ax_u16)(0)), .sym_name=((ax_u32)(imm)), .addend=((ax_i64)(0))}));
                ax_x86_encode_mov_ri64(dst, ((ax_i64)(0)), buf);
            } else if ((inst.src1.vreg == ((ax_u32)(2)))) {
                ax_i64 offset = (buf->len + ((ax_i64)(3)));
                ax_i64 sym_val = (((ax_i64)(0xE0000000)) + imm);
                ax_RelocationVec_push_reloc(&(e->relocs), ((struct ax_Relocation){.offset=offset, .kind=ax_RELOC_PC32, .padding1=((ax_u8)(0)), .padding2=((ax_u16)(0)), .sym_name=((ax_u32)(sym_val)), .addend=(-((ax_i64)(4)))}));
                ax_x86_encode_lea_rip(dst, ((ax_i32)(0)), buf);
            } else {
                {
                    if (((imm >= (-((ax_i64)(2147483648)))) && (imm <= ((ax_i64)(2147483647))))) {
                        ax_x86_encode_mov_ri(dst, ((ax_i32)(imm)), buf);
                    } else {
                        {
                            ax_x86_encode_mov_ri64(dst, imm, buf);
                        }
                    }
                }
            }
        }
    } else if ((op == ax_MACH_XOR_ZERO)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        if ((dst != ax_REG_NONE)) {
            ax_x86_encode_xor_zero(dst, buf);
        }
    } else if ((op == ax_MACH_ADD)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        if ((inst.src1.kind == ax_OPND_IMM)) {
            ax_x86_encode_add_ri(dst, ((ax_i32)(inst.src1.imm)), buf);
        } else {
            {
                ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
                ax_x86_encode_add_rr(dst, src, buf);
            }
        }
    } else if ((op == ax_MACH_SUB)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        if ((inst.src1.kind == ax_OPND_IMM)) {
            ax_x86_encode_sub_ri(dst, ((ax_i32)(inst.src1.imm)), buf);
        } else {
            {
                ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
                ax_x86_encode_sub_rr(dst, src, buf);
            }
        }
    } else if ((op == ax_MACH_IMUL)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_x86_encode_imul_rr(dst, src, buf);
    } else if ((op == ax_MACH_IDIV)) {
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_x86_encode_idiv_r(src, buf);
    } else if ((op == ax_MACH_CQO)) {
        ax_ByteVec_x86_encode_cqo(buf);
    } else if ((op == ax_MACH_NEG)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_x86_encode_neg_r(dst, buf);
    } else if ((op == ax_MACH_NOT)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_x86_encode_not_r(dst, buf);
    } else if ((op == ax_MACH_AND)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_x86_encode_and_rr(dst, src, buf);
    } else if ((op == ax_MACH_OR)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_x86_encode_or_rr(dst, src, buf);
    } else if ((op == ax_MACH_XOR)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_x86_encode_xor_rr(dst, src, buf);
    } else if ((op == ax_MACH_SHL)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        if ((inst.src1.kind == ax_OPND_IMM)) {
            ax_x86_encode_shl_imm(dst, ((ax_u8)(inst.src1.imm)), buf);
        } else {
            {
                ax_x86_encode_shl_cl(dst, buf);
            }
        }
    } else if ((op == ax_MACH_SAR)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        if ((inst.src1.kind == ax_OPND_IMM)) {
            ax_x86_encode_sar_imm(dst, ((ax_u8)(inst.src1.imm)), buf);
        } else {
            {
                ax_x86_encode_sar_cl(dst, buf);
            }
        }
    } else if ((op == ax_MACH_CMP)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        if ((inst.src1.kind == ax_OPND_IMM)) {
            ax_x86_encode_cmp_ri(dst, ((ax_i32)(inst.src1.imm)), buf);
        } else {
            {
                ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
                ax_x86_encode_cmp_rr(dst, src, buf);
            }
        }
    } else if ((op == ax_MACH_TEST)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_x86_encode_test_rr(dst, src, buf);
    } else if ((op == ax_MACH_SETCC)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_x86_encode_setcc(inst.cc, dst, buf);
    } else if ((op == ax_MACH_MOVZX_B)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_x86_encode_movzx_br(dst, src, buf);
    } else if ((op == ax_MACH_PUSH)) {
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_x86_encode_push(src, buf);
    } else if ((op == ax_MACH_POP)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_x86_encode_pop(dst, buf);
    } else if ((op == ax_MACH_JMP)) {
        if ((inst.dst.kind == ax_OPND_LABEL)) {
            ax_FixupVec_push_fixup(&(e->fixups), ((struct ax_Fixup){.offset=(buf->len + 1), .label_id=inst.dst.label, .inst_size=5}));
            ax_x86_encode_jmp_rel32(((ax_i32)(0)), buf);
        }
    } else if ((op == ax_MACH_JCC)) {
        if ((inst.dst.kind == ax_OPND_LABEL)) {
            ax_FixupVec_push_fixup(&(e->fixups), ((struct ax_Fixup){.offset=(buf->len + 2), .label_id=inst.dst.label, .inst_size=6}));
            ax_x86_encode_jcc_rel32(inst.cc, ((ax_i32)(0)), buf);
        }
    } else if ((op == ax_MACH_CALL)) {
        ax_i32 disp = ((ax_i32)(0));
        if ((inst.src1.imm == ((ax_i64)(0)))) {
            disp = (-((ax_i32)((buf->len + 5))));
        } else {
            {
                ax_FixupVec_push_fixup(&(e->fixups), ((struct ax_Fixup){.offset=(buf->len + 1), .label_id=((ax_u32)(inst.src1.imm)), .inst_size=5}));
            }
        }
        ax_x86_encode_call_rel32(disp, buf);
    } else if ((op == ax_MACH_SYSCALL)) {
        ax_ByteVec_x86_encode_syscall(buf);
    } else if ((op == ax_MACH_LOAD)) {
        ax_u8 dst = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_u8 base = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_i32 disp = ((ax_i32)(0));
        if ((inst.src2.kind == ax_OPND_IMM)) {
            disp = ((ax_i32)(inst.src2.imm));
        }
        ax_x86_encode_mov_load(dst, base, disp, buf);
    } else if ((op == ax_MACH_STORE)) {
        ax_u8 base = ax_MachOperand_emitter_resolve_reg(inst.dst, allocs);
        ax_u8 src = ax_MachOperand_emitter_resolve_reg(inst.src1, allocs);
        ax_i32 disp = ((ax_i32)(0));
        if ((inst.src2.kind == ax_OPND_IMM)) {
            disp = ((ax_i32)(inst.src2.imm));
        }
        ax_x86_encode_mov_store(base, disp, src, buf);
    }
}

void ax_MachEmitter_emitter_resolve_fixups(struct ax_MachEmitter* e) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < e->fixups.len)) {
        struct ax_Fixup fix = ((e->fixups.data)[i]);
        ax_i64 target = ((ax_i64)(0));
        ax_bool found = ax_LabelMap_label_map_get(&(e->labels), fix.label_id, &(target));
        if ((!found)) {
            ax_RelocationVec_push_reloc(&(e->relocs), ((struct ax_Relocation){.offset=fix.offset, .kind=ax_RELOC_PC32, .padding1=((ax_u8)(0)), .padding2=((ax_u16)(0)), .sym_name=fix.label_id, .addend=(-((ax_i64)(4)))}));
        } else {
            {
                ax_i32 rel = ((ax_i32)((target - (fix.offset + ((ax_i64)(4))))));
                ((e->code.data)[(fix.offset + 0)]) = ((ax_u8)((rel & ((ax_i32)(0xFF)))));
                ((e->code.data)[(fix.offset + 1)]) = ((ax_u8)(((rel >> 8) & ((ax_i32)(0xFF)))));
                ((e->code.data)[(fix.offset + 2)]) = ((ax_u8)(((rel >> 16) & ((ax_i32)(0xFF)))));
                ((e->code.data)[(fix.offset + 3)]) = ((ax_u8)(((rel >> 24) & ((ax_i32)(0xFF)))));
            }
        }
        i = (i + 1);
    }
}

void ax_MachEmitter_emit_function_binary(struct ax_MachEmitter* e, ax_string fn_name, struct ax_MachInst* insts, ax_i64 insts_len, struct ax_RegAllocation* allocs, struct ax_StackFrame* frame) {
    struct ax_MachInstVec prologue = ax_new_mach_inst_vec();
    ax_StackFrame_emit_prologue(frame, &(prologue));
    ax_i64 i = ((ax_i64)(0));
    while ((i < prologue.len)) {
        ax_MachEmitter_emit_mach_inst(e, ((prologue.data)[i]), allocs);
        i = (i + 1);
    }
    if ((prologue.data != ((struct ax_MachInst*)(NULL)))) {
        ax_free(((ax_u8*)(prologue.data)));
    }
    i = ((ax_i64)(0));
    while ((i < insts_len)) {
        struct ax_MachInst inst = ((insts)[i]);
        if ((inst.op == ax_MACH_RET)) {
            struct ax_MachInstVec epilogue = ax_new_mach_inst_vec();
            ax_StackFrame_emit_epilogue(frame, &(epilogue));
            ax_i64 j = ((ax_i64)(0));
            while ((j < epilogue.len)) {
                ax_MachEmitter_emit_mach_inst(e, ((epilogue.data)[j]), allocs);
                j = (j + 1);
            }
            if ((epilogue.data != ((struct ax_MachInst*)(NULL)))) {
                ax_free(((ax_u8*)(epilogue.data)));
            }
        }
        ax_MachEmitter_emit_mach_inst(e, inst, allocs);
        i = (i + 1);
    }
    ax_MachEmitter_emitter_resolve_fixups(e);
}

struct ax_ELF64SymVec ax_new_elf64_sym_vec(void) {
    return ((struct ax_ELF64SymVec){.data=((struct ax_ELF64Sym*)(NULL)), .len=0, .cap=0});
}

void ax_ELF64SymVec_push_elf64_sym(struct ax_ELF64SymVec* self, struct ax_ELF64Sym s) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_ELF64Sym* new_data = ((struct ax_ELF64Sym*)(ax_alloc((new_cap * 40))));
        if ((self->data != ((struct ax_ELF64Sym*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 40));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = s;
    self->len = (self->len + 1);
}

void ax_ByteVec_elf64_serialize(struct ax_ByteVec* code, struct ax_RelocationVec* relocs, struct ax_ELF64SymVec* symbols, struct ax_ByteVec* out_bytes) {
    struct ax_ByteVec strtab = ax_new_byte_vec();
    ax_ByteVec_push_byte(&(strtab), ((ax_u8)(0)));
    ax_string* sec_names = ((ax_string*)(ax_alloc((4 * 8))));
    ((sec_names)[0]) = (ax_string){.ptr=(const ax_u8*)".text", .len=5};
    ((sec_names)[1]) = (ax_string){.ptr=(const ax_u8*)".symtab", .len=7};
    ((sec_names)[2]) = (ax_string){.ptr=(const ax_u8*)".strtab", .len=7};
    ((sec_names)[3]) = (ax_string){.ptr=(const ax_u8*)".rela.text", .len=10};
    ax_u32* sec_offsets = ((ax_u32*)(ax_alloc((4 * 4))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < 4)) {
        ((sec_offsets)[i]) = ((ax_u32)(strtab.len));
        ax_string s = ((sec_names)[i]);
        ax_i64 s_len = ax_str_len(s);
        ax_u8* s_ptr = ((ax_u8*)(s.ptr));
        ax_i64 char_i = ((ax_i64)(0));
        while ((char_i < s_len)) {
            ax_ByteVec_push_byte(&(strtab), ((s_ptr)[char_i]));
            char_i = (char_i + 1);
        }
        ax_ByteVec_push_byte(&(strtab), ((ax_u8)(0)));
        i = (i + 1);
    }
    ax_u32* sym_offsets = ((ax_u32*)(ax_alloc((symbols->len * 4))));
    ax_i64 si = ((ax_i64)(0));
    while ((si < symbols->len)) {
        ((sym_offsets)[si]) = ((ax_u32)(strtab.len));
        ax_string s = ((symbols->data)[si]).name_str;
        ax_i64 s_len = ax_str_len(s);
        ax_u8* s_ptr = ((ax_u8*)(s.ptr));
        ax_i64 char_i = ((ax_i64)(0));
        while ((char_i < s_len)) {
            ax_ByteVec_push_byte(&(strtab), ((s_ptr)[char_i]));
            char_i = (char_i + 1);
        }
        ax_ByteVec_push_byte(&(strtab), ((ax_u8)(0)));
        si = (si + 1);
    }
    ax_u16 num_sections = ((ax_u16)(4));
    if ((relocs->len > ((ax_i64)(0)))) {
        num_sections = ((ax_u16)(5));
    }
    ax_i64 ehdr_size = ((ax_i64)(64));
    ax_i64 shdr_size = ((ax_i64)(64));
    ax_i64 shdr_table_size = (((ax_i64)(num_sections)) * shdr_size);
    ax_i64 text_off = ehdr_size;
    ax_i64 text_size = code->len;
    ax_i64 symtab_off = (text_off + text_size);
    if (((symtab_off % ((ax_i64)(8))) != ((ax_i64)(0)))) {
        symtab_off = (symtab_off + (((ax_i64)(8)) - (symtab_off % ((ax_i64)(8)))));
    }
    ax_i64 sym_ent_size = ((ax_i64)(24));
    ax_i64 sym_count = (((ax_i64)(1)) + symbols->len);
    ax_i64 symtab_size = (sym_count * sym_ent_size);
    ax_i64 strtab_off = (symtab_off + symtab_size);
    ax_i64 strtab_size = strtab.len;
    ax_i64 rela_off = (strtab_off + strtab_size);
    ax_i64 rela_ent_size = ((ax_i64)(24));
    ax_i64 rela_size = ((ax_i64)(0));
    if ((relocs->len > ((ax_i64)(0)))) {
        rela_size = (relocs->len * rela_ent_size);
    }
    ax_i64 shdr_off = (rela_off + rela_size);
    if ((relocs->len == ((ax_i64)(0)))) {
        shdr_off = (strtab_off + strtab_size);
    }
    if (((shdr_off % ((ax_i64)(8))) != ((ax_i64)(0)))) {
        shdr_off = (shdr_off + (((ax_i64)(8)) - (shdr_off % ((ax_i64)(8)))));
    }
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0x7F)));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('E')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('L')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('F')));
    ax_ByteVec_push_byte(out_bytes, ax_ELF_CLASS64);
    ax_ByteVec_push_byte(out_bytes, ax_ELF_DATA2LSB);
    ax_ByteVec_push_byte(out_bytes, ax_EV_CURRENT);
    ax_i64 pad_i = ((ax_i64)(0));
    while ((pad_i < 9)) {
        ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
        pad_i = (pad_i + 1);
    }
    ax_ByteVec_push_u16_le(out_bytes, ax_ET_REL);
    ax_ByteVec_push_u16_le(out_bytes, ax_EM_X86_64);
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(ax_EV_CURRENT)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(shdr_off)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(ehdr_size)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(0)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(0)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(shdr_size)));
    ax_ByteVec_push_u16_le(out_bytes, num_sections);
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(3)));
    ax_ByteVec_push_bytes(out_bytes, code->data, text_size);
    while ((out_bytes->len < symtab_off)) {
        ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
    }
    ax_i64 sym_pad = ((ax_i64)(0));
    while ((sym_pad < 24)) {
        ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
        sym_pad = (sym_pad + 1);
    }
    si = 0;
    while ((si < symbols->len)) {
        struct ax_ELF64Sym* sym = &(((symbols->data)[si]));
        ax_u32 name_idx = ((sym_offsets)[si]);
        ax_ByteVec_push_u32_le(out_bytes, name_idx);
        ax_ByteVec_push_byte(out_bytes, ((sym->binding << ((ax_u8)(4))) | sym->sym_type));
        ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
        ax_ByteVec_push_u16_le(out_bytes, sym->section);
        ax_ByteVec_push_u64_le(out_bytes, sym->value);
        ax_ByteVec_push_u64_le(out_bytes, sym->size);
        si = (si + 1);
    }
    ax_ByteVec_push_bytes(out_bytes, strtab.data, strtab_size);
    if ((relocs->len > ((ax_i64)(0)))) {
        ax_i64 ri = ((ax_i64)(0));
        while ((ri < relocs->len)) {
            struct ax_Relocation r = ((relocs->data)[ri]);
            ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(r.offset)));
            ax_u32 r_type = ((ax_u32)(1));
            if ((r.kind == ax_RELOC_PC32)) {
                r_type = ((ax_u32)(2));
            } else if ((r.kind == ax_RELOC_PLT32)) {
                r_type = ((ax_u32)(4));
            } else if ((r.kind == ax_RELOC_ABS64)) {
                r_type = ((ax_u32)(1));
            }
            ax_u64 r_info = (((ax_u64)(r_type)) | (((ax_u64)(r.sym_name)) << ((ax_u64)(32))));
            ax_ByteVec_push_u64_le(out_bytes, r_info);
            ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(r.addend)));
            ri = (ri + 1);
        }
    }
    while ((out_bytes->len < shdr_off)) {
        ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
    }
    ax_i64 sh_pad = ((ax_i64)(0));
    while ((sh_pad < 64)) {
        ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
        sh_pad = (sh_pad + 1);
    }
    ax_ByteVec_push_u32_le(out_bytes, ((sec_offsets)[0]));
    ax_ByteVec_push_u32_le(out_bytes, ax_SHT_PROGBITS);
    ax_ByteVec_push_u64_le(out_bytes, (ax_SHF_ALLOC | ax_SHF_EXECINSTR));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(text_off)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(text_size)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(16)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((sec_offsets)[1]));
    ax_ByteVec_push_u32_le(out_bytes, ax_SHT_SYMTAB);
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(symtab_off)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(symtab_size)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(3)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(1)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(8)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(sym_ent_size)));
    ax_ByteVec_push_u32_le(out_bytes, ((sec_offsets)[2]));
    ax_ByteVec_push_u32_le(out_bytes, ax_SHT_STRTAB);
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(strtab_off)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(strtab_size)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(1)));
    ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
    if ((relocs->len > ((ax_i64)(0)))) {
        ax_ByteVec_push_u32_le(out_bytes, ((sec_offsets)[3]));
        ax_ByteVec_push_u32_le(out_bytes, ax_SHT_RELA);
        ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
        ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(0)));
        ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(rela_off)));
        ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(rela_size)));
        ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(2)));
        ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(1)));
        ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(8)));
        ax_ByteVec_push_u64_le(out_bytes, ((ax_u64)(rela_ent_size)));
    }
    ax_free(((ax_u8*)(sec_names)));
    ax_free(((ax_u8*)(sec_offsets)));
    ax_free(((ax_u8*)(sym_offsets)));
    if ((strtab.data != ((ax_u8*)(NULL)))) {
        ax_free(strtab.data);
    }
}

ax_bool ax_elf64_write_object_file(ax_string filename, struct ax_ByteVec* code, struct ax_RelocationVec* relocs, struct ax_ELF64SymVec* symbols) {
    struct ax_ByteVec out_bytes = ax_new_byte_vec();
    ax_ByteVec_elf64_serialize(code, relocs, symbols, &(out_bytes));
    ax_u8* filename_nt = ax_x86_str_to_null_terminated(filename);
    void* file = fopen(((void*)(filename_nt)), (const char*)((ax_string){.ptr=(const ax_u8*)"wb", .len=2}).ptr);
    ax_free(filename_nt);
    if ((file == ((void*)(NULL)))) {
        if ((out_bytes.data != ((ax_u8*)(NULL)))) {
            ax_free(out_bytes.data);
        }
        return AX_FALSE;
    }
    fwrite(((void*)(out_bytes.data)), 1, out_bytes.len, file);
    fclose(file);
    if ((out_bytes.data != ((ax_u8*)(NULL)))) {
        ax_free(out_bytes.data);
    }
    return AX_TRUE;
}

struct ax_COFFRelocVec ax_new_coff_reloc_vec(void) {
    return ((struct ax_COFFRelocVec){.data=((struct ax_COFFReloc*)(NULL)), .len=0, .cap=0});
}

void ax_COFFRelocVec_push_coff_reloc(struct ax_COFFRelocVec* self, struct ax_COFFReloc r) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_COFFReloc* new_data = ((struct ax_COFFReloc*)(ax_alloc((new_cap * 12))));
        if ((self->data != ((struct ax_COFFReloc*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 12));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = r;
    self->len = (self->len + 1);
}

struct ax_COFFSymbolVec ax_new_coff_symbol_vec(void) {
    return ((struct ax_COFFSymbolVec){.data=((struct ax_COFFSymbol*)(NULL)), .len=0, .cap=0});
}

ax_i32 ax_COFFSymbolVec_push_coff_symbol(struct ax_COFFSymbolVec* self, struct ax_COFFSymbol s) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_COFFSymbol* new_data = ((struct ax_COFFSymbol*)(ax_alloc((new_cap * 32))));
        if ((self->data != ((struct ax_COFFSymbol*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 32));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ax_i32 idx = ((ax_i32)(self->len));
    ((self->data)[self->len]) = s;
    self->len = (self->len + 1);
    return idx;
}

void ax_ByteVec_coff_serialize(struct ax_ByteVec* code, struct ax_ByteVec* rdata, struct ax_COFFRelocVec* relocs, struct ax_COFFSymbolVec* symbols, struct ax_ByteVec* out_bytes) {
    struct ax_ByteVec strtab = ax_new_byte_vec();
    ax_ByteVec_push_u32_le(&(strtab), ((ax_u32)(4)));
    ax_u32* string_offsets = ((ax_u32*)(ax_alloc((symbols->len * 4))));
    ax_i64 si = ((ax_i64)(0));
    while ((si < symbols->len)) {
        ax_string name = ((symbols->data)[si]).name_str;
        ax_i64 name_len = ax_str_len(name);
        if ((name_len > 8)) {
            ((string_offsets)[si]) = ((ax_u32)(strtab.len));
            ax_u8* name_ptr = ((ax_u8*)(name.ptr));
            ax_i64 char_i = ((ax_i64)(0));
            while ((char_i < name_len)) {
                ax_ByteVec_push_byte(&(strtab), ((name_ptr)[char_i]));
                char_i = (char_i + 1);
            }
            ax_ByteVec_push_byte(&(strtab), ((ax_u8)(0)));
        } else {
            {
                ((string_offsets)[si]) = ((ax_u32)(0));
            }
        }
        si = (si + 1);
    }
    if ((strtab.len > ((ax_i64)(4)))) {
        ax_u32 total_len = ((ax_u32)(strtab.len));
        ((strtab.data)[0]) = ((ax_u8)((total_len & ((ax_u32)(0xFF)))));
        ((strtab.data)[1]) = ((ax_u8)(((total_len >> ((ax_u32)(8))) & ((ax_u32)(0xFF)))));
        ((strtab.data)[2]) = ((ax_u8)(((total_len >> ((ax_u32)(16))) & ((ax_u32)(0xFF)))));
        ((strtab.data)[3]) = ((ax_u8)(((total_len >> ((ax_u32)(24))) & ((ax_u32)(0xFF)))));
    }
    ax_i64 text_size = code->len;
    ax_i64 rdata_size = rdata->len;
    ax_i64 text_raw_ptr = ((ax_i64)(100));
    ax_i64 rdata_raw_ptr = (text_raw_ptr + text_size);
    ax_i64 relocs_offset = (rdata_raw_ptr + rdata_size);
    ax_i64 symbol_table_offset = (relocs_offset + (relocs->len * ((ax_i64)(10))));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(0x8664)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(2)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(symbol_table_offset)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(symbols->len)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(0)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(0)));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('.')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('t')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('e')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('x')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('t')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(text_size)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(text_raw_ptr)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(relocs_offset)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(relocs->len)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((((ax_u32)(0x00000020)) | ((ax_u32)(0x20000000))) | ((ax_u32)(0x40000000))));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('.')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('r')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('d')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('a')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('t')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)('a')));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
    ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(rdata_size)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(rdata_raw_ptr)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(0)));
    ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(0)));
    ax_ByteVec_push_u32_le(out_bytes, (((ax_u32)(0x00000040)) | ((ax_u32)(0x40000000))));
    ax_ByteVec_push_bytes(out_bytes, code->data, text_size);
    ax_ByteVec_push_bytes(out_bytes, rdata->data, rdata_size);
    ax_i64 ri = ((ax_i64)(0));
    while ((ri < relocs->len)) {
        struct ax_COFFReloc r = ((relocs->data)[ri]);
        ax_ByteVec_push_u32_le(out_bytes, r.virt_addr);
        ax_ByteVec_push_u32_le(out_bytes, r.sym_table_idx);
        ax_ByteVec_push_u16_le(out_bytes, r.reloc_type);
        ri = (ri + 1);
    }
    si = 0;
    while ((si < symbols->len)) {
        struct ax_COFFSymbol* sym = &(((symbols->data)[si]));
        ax_i64 name_len = ax_str_len(sym->name_str);
        if ((name_len <= 8)) {
            ax_u8* name_ptr = ((ax_u8*)(sym->name_str.ptr));
            ax_i64 char_i = ((ax_i64)(0));
            while ((char_i < 8)) {
                if ((char_i < name_len)) {
                    ax_ByteVec_push_byte(out_bytes, ((name_ptr)[char_i]));
                } else {
                    {
                        ax_ByteVec_push_byte(out_bytes, ((ax_u8)(0)));
                    }
                }
                char_i = (char_i + 1);
            }
        } else {
            {
                ax_u32 offset = ((string_offsets)[si]);
                ax_ByteVec_push_u32_le(out_bytes, ((ax_u32)(0)));
                ax_ByteVec_push_u32_le(out_bytes, offset);
            }
        }
        ax_ByteVec_push_u32_le(out_bytes, sym->value);
        ax_ByteVec_push_u16_le(out_bytes, ((ax_u16)(sym->section_num)));
        ax_ByteVec_push_u16_le(out_bytes, sym->sym_type);
        ax_ByteVec_push_byte(out_bytes, sym->storage_class);
        ax_ByteVec_push_byte(out_bytes, sym->num_aux);
        si = (si + 1);
    }
    ax_ByteVec_push_bytes(out_bytes, strtab.data, strtab.len);
    ax_free(((ax_u8*)(string_offsets)));
    if ((strtab.data != ((ax_u8*)(NULL)))) {
        ax_free(strtab.data);
    }
}

ax_bool ax_coff_write_object_file(ax_string filename, struct ax_ByteVec* code, struct ax_ByteVec* rdata, struct ax_COFFRelocVec* relocs, struct ax_COFFSymbolVec* symbols) {
    struct ax_ByteVec out_bytes = ax_new_byte_vec();
    ax_ByteVec_coff_serialize(code, rdata, relocs, symbols, &(out_bytes));
    ax_u8* filename_nt = ax_x86_str_to_null_terminated(filename);
    void* file = fopen(((void*)(filename_nt)), (const char*)((ax_string){.ptr=(const ax_u8*)"wb", .len=2}).ptr);
    ax_free(filename_nt);
    if ((file == ((void*)(NULL)))) {
        if ((out_bytes.data != ((ax_u8*)(NULL)))) {
            ax_free(out_bytes.data);
        }
        return AX_FALSE;
    }
    fwrite(((void*)(out_bytes.data)), 1, out_bytes.len, file);
    fclose(file);
    if ((out_bytes.data != ((ax_u8*)(NULL)))) {
        ax_free(out_bytes.data);
    }
    return AX_TRUE;
}

struct ax_CompiledFuncInfoVec ax_new_compiled_func_info_vec(void) {
    return ((struct ax_CompiledFuncInfoVec){.data=((struct ax_CompiledFuncInfo*)(NULL)), .len=0, .cap=0});
}

void ax_CompiledFuncInfoVec_push_compiled_func_info(struct ax_CompiledFuncInfoVec* self, struct ax_CompiledFuncInfo info) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_CompiledFuncInfo* new_data = ((struct ax_CompiledFuncInfo*)(ax_alloc((new_cap * 56))));
        if ((self->data != ((struct ax_CompiledFuncInfo*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 56));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = info;
    self->len = (self->len + 1);
}

ax_string ax_resolve_binary_sym_name(ax_i64 sym_idx, ax_string current_fn_name, struct ax_SymbolTable symbols, struct ax_InternPool pool) {
    if ((sym_idx >= ((ax_i64)(0xE0000000)))) {
        ax_i64 str_id = (sym_idx - ((ax_i64)(0xE0000000)));
        ax_u8* sym_name_buf = ((ax_u8*)(ax_alloc(64)));
        snprintf(((void*)(sym_name_buf)), 64, (const char*)((ax_string){.ptr=(const ax_u8*)"Linfo.string.%d", .len=15}).ptr, ((void*)(str_id)), ((void*)(NULL)));
        ax_u32 name_id = ax_InternPool_intern(&(pool), ((ax_string){.ptr = (const ax_u8*)(sym_name_buf), .len = strlen((const char*)(sym_name_buf))}));
        ax_free(sym_name_buf);
        return ax_InternPool_get(pool, name_id);
    }
    if ((sym_idx == (-((ax_i64)(1))))) {
        return (ax_string){.ptr=(const ax_u8*)"malloc", .len=6};
    }
    if ((sym_idx == (-((ax_i64)(2))))) {
        return (ax_string){.ptr=(const ax_u8*)"free", .len=4};
    }
    if ((sym_idx == (-((ax_i64)(3))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_actor_spawn", .len=14};
    }
    if ((sym_idx == (-((ax_i64)(4))))) {
        return (ax_string){.ptr=(const ax_u8*)"__ax_runtime_init", .len=17};
    }
    if ((sym_idx == (-((ax_i64)(5))))) {
        return (ax_string){.ptr=(const ax_u8*)"__ax_runtime_shutdown", .len=21};
    }
    if ((sym_idx == (-((ax_i64)(8))))) {
        return (ax_string){.ptr=(const ax_u8*)"abort", .len=5};
    }
    if ((sym_idx == (-((ax_i64)(10))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_println_str_native", .len=21};
    }
    if ((sym_idx == (-((ax_i64)(11))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_println_i64", .len=14};
    }
    if ((sym_idx == (-((ax_i64)(12))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_println_f64", .len=14};
    }
    if ((sym_idx == (-((ax_i64)(13))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_println_bool", .len=15};
    }
    if ((sym_idx == (-((ax_i64)(14))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_print_str_native", .len=19};
    }
    if ((sym_idx == (-((ax_i64)(15))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_print_i64", .len=12};
    }
    if ((sym_idx == (-((ax_i64)(16))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_print_f64", .len=12};
    }
    if ((sym_idx == (-((ax_i64)(17))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_print_bool", .len=13};
    }
    if ((sym_idx <= ((ax_i64)(0)))) {
        return current_fn_name;
    }
    return ax_x86_resolve_sym_name(((ax_u32)(sym_idx)), symbols, pool);
}

ax_bool ax_compile_native_binary(ax_string output_obj_file, struct ax_AirModule mod, struct ax_SymbolTable symbols, struct ax_InternPool pool, struct ax_TypeTable table, ax_string format) {
    ax_string abi = (ax_string){.ptr=(const ax_u8*)"win64", .len=5};
    if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"elf64", .len=5})) {
        abi = (ax_string){.ptr=(const ax_u8*)"sysv", .len=4};
    }
    struct ax_ByteVec all_code = ax_new_byte_vec();
    struct ax_CompiledFuncInfoVec func_infos = ax_new_compiled_func_info_vec();
    struct ax_COFFSymbolVec coff_syms = ax_new_coff_symbol_vec();
    struct ax_ELF64SymVec elf_syms = ax_new_elf64_sym_vec();
    struct ax_ByteVec rdata_buf = ax_new_byte_vec();
    struct ax_U32Vec str_keys = ax_new_u32_vec();
    struct ax_U32Vec str_offsets = ax_new_u32_vec();
    ax_i64 scan_fi = ((ax_i64)(0));
    while ((scan_fi < mod.funcs.len)) {
        struct ax_AirFunc* fn_ptr = &(((mod.funcs.data)[scan_fi]));
        if ((!fn_ptr->is_extern)) {
            ax_i64 inst_i = ((ax_i64)(0));
            while ((inst_i < fn_ptr->insts.len)) {
                struct ax_AirInst* inst = &(((fn_ptr->insts.data)[inst_i]));
                if (((inst->opcode == ax_OP_ICONST) && (inst->type_id == ((ax_u16)(12))))) {
                    ax_u32 str_id = inst->src1;
                    ax_bool found = AX_FALSE;
                    ax_i64 j = ((ax_i64)(0));
                    while ((j < str_keys.len)) {
                        if ((((str_keys.data)[j]) == str_id)) {
                            found = AX_TRUE;
                            break;
                        }
                        j = (j + 1);
                    }
                    if ((!found)) {
                        ax_u32 offset = ((ax_u32)(rdata_buf.len));
                        ax_U32Vec_push(&(str_keys), str_id);
                        ax_U32Vec_push(&(str_offsets), offset);
                        ax_string text = ax_InternPool_get(pool, str_id);
                        ax_i64 len = ax_str_len(text);
                        ax_u8* text_ptr = ((ax_u8*)(text.ptr));
                        ax_i64 k = ((ax_i64)(0));
                        while ((k < len)) {
                            ax_ByteVec_push_byte(&(rdata_buf), ((text_ptr)[k]));
                            k = (k + 1);
                        }
                        ax_ByteVec_push_byte(&(rdata_buf), ((ax_u8)(0)));
                    }
                }
                inst_i = (inst_i + 1);
            }
        }
        scan_fi = (scan_fi + 1);
    }
    ax_i64 str_i = ((ax_i64)(0));
    while ((str_i < str_keys.len)) {
        ax_u32 str_id = ((str_keys.data)[str_i]);
        ax_u32 offset = ((str_offsets.data)[str_i]);
        ax_u8* sym_name_buf = ((ax_u8*)(ax_alloc(64)));
        snprintf(((void*)(sym_name_buf)), 64, (const char*)((ax_string){.ptr=(const ax_u8*)"Linfo.string.%d", .len=15}).ptr, ((void*)(((ax_i64)(str_id)))), ((void*)(NULL)));
        ax_u32 sym_name = ax_InternPool_intern(&(pool), ((ax_string){.ptr = (const ax_u8*)(sym_name_buf), .len = strlen((const char*)(sym_name_buf))}));
        ax_string sym_name_str = ax_InternPool_get(pool, sym_name);
        if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"elf64", .len=5})) {
            ax_ELF64SymVec_push_elf64_sym(&(elf_syms), ((struct ax_ELF64Sym){.name_str=sym_name_str, .value=((ax_u64)(offset)), .size=((ax_u64)(0)), .binding=ax_STB_GLOBAL, .sym_type=ax_STT_OBJECT, .section=((ax_u16)(2))}));
        } else {
            {
                ax_COFFSymbolVec_push_coff_symbol(&(coff_syms), ((struct ax_COFFSymbol){.name_str=sym_name_str, .value=offset, .section_num=((ax_i16)(2)), .sym_type=((ax_u16)(0)), .storage_class=ax_IMAGE_SYM_CLASS_STATIC, .num_aux=((ax_u8)(0))}));
            }
        }
        ax_free(sym_name_buf);
        str_i = (str_i + 1);
    }
    ax_i64 fi = ((ax_i64)(0));
    while ((fi < mod.funcs.len)) {
        struct ax_AirFunc* fn_ptr = &(((mod.funcs.data)[fi]));
        if (fn_ptr->is_extern) {
            ax_string name = ax_x86_resolve_sym_name(fn_ptr->sym_id, symbols, pool);
            if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"elf64", .len=5})) {
                ax_ELF64SymVec_push_elf64_sym(&(elf_syms), ((struct ax_ELF64Sym){.name_str=name, .value=((ax_u64)(0)), .size=((ax_u64)(0)), .binding=ax_STB_GLOBAL, .sym_type=ax_STT_FUNC, .section=((ax_u16)(0))}));
            } else {
                {
                    ax_COFFSymbolVec_push_coff_symbol(&(coff_syms), ((struct ax_COFFSymbol){.name_str=name, .value=((ax_u32)(0)), .section_num=((ax_i16)(0)), .sym_type=((ax_u16)(0x20)), .storage_class=ax_IMAGE_SYM_CLASS_EXTERNAL, .num_aux=((ax_u8)(0))}));
                }
            }
        }
        fi = (fi + 1);
    }
    fi = 0;
    while ((fi < mod.funcs.len)) {
        struct ax_AirFunc* fn_ptr = &(((mod.funcs.data)[fi]));
        if ((!fn_ptr->is_extern)) {
            ax_string fn_name = ax_x86_resolve_sym_name(fn_ptr->sym_id, symbols, pool);
            ax_u8* name_ptr = ax_str_to_null_terminated(fn_name);
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: compiling function %s\n", .len=53}).ptr, ((ax_i64)(name_ptr)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_free(name_ptr);
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]   running select_all\n", .len=29}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            struct ax_MachInstVec mach_insts = ax_AirFunc_select_all(fn_ptr, abi, table, symbols, pool);
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]   running compute_liveness\n", .len=35}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            struct ax_LiveIntervalVec intervals = ax_MachInst_compute_liveness(mach_insts.data, mach_insts.len);
            ax_i64 gpr_len = ((ax_i64)(0));
            ax_u8* avail_gprs = ax_get_allocatable_gprs(&(gpr_len));
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]   running graph_coloring_alloc\n", .len=39}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            struct ax_RegAllocResult alloc_res = ax_LiveIntervalVec_graph_coloring_alloc(intervals, avail_gprs, gpr_len);
            ax_i64 cs_len = ((ax_i64)(0));
            ax_u8* cs_regs = ax_get_used_callee_saved(abi, alloc_res.allocs, alloc_res.max_vreg, &(cs_len));
            struct ax_StackFrame frame = ax_compute_frame(cs_regs, cs_len, alloc_res.spill_count, ((ax_i32)(0)));
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]   running insert_spill_code\n", .len=36}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            struct ax_MachInstVec final_insts = ax_MachInst_insert_spill_code(mach_insts.data, mach_insts.len, alloc_res.allocs, &(frame));
            struct ax_MachEmitter emitter = ax_new_mach_emitter();
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]   running emit_function_binary\n", .len=39}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_MachEmitter_emit_function_binary(&(emitter), fn_name, final_insts.data, final_insts.len, alloc_res.allocs, &(frame));
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]   emit_function_binary done, pushing compiled info\n", .len=59}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_u32 current_offset = ((ax_u32)(all_code.len));
            ax_ByteVec_push_bytes(&(all_code), emitter.code.data, emitter.code.len);
            ax_CompiledFuncInfoVec_push_compiled_func_info(&(func_infos), ((struct ax_CompiledFuncInfo){.name=fn_name, .offset=current_offset, .size=((ax_u32)(emitter.code.len)), .is_global=AX_TRUE, .relocs=emitter.relocs}));
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]   freeing local resources\n", .len=34}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing cs_regs\n", .len=28}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_free(cs_regs);
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing avail_gprs\n", .len=31}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_free(avail_gprs);
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing alloc_res.allocs\n", .len=37}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_free(alloc_res.allocs);
            if ((intervals.data != ((struct ax_LiveInterval*)(NULL)))) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing intervals.data\n", .len=35}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                ax_free(((ax_u8*)(intervals.data)));
            }
            if ((mach_insts.data != ((struct ax_MachInst*)(NULL)))) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing mach_insts.data\n", .len=36}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                ax_free(((ax_u8*)(mach_insts.data)));
            }
            if ((final_insts.data != ((struct ax_MachInst*)(NULL)))) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing final_insts.data\n", .len=37}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                ax_free(((ax_u8*)(final_insts.data)));
            }
            if ((emitter.code.data != ((ax_u8*)(NULL)))) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing emitter.code.data\n", .len=38}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                ax_free(emitter.code.data);
            }
            if ((emitter.labels.keys != ((ax_u32*)(NULL)))) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing emitter.labels.keys\n", .len=40}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                ax_free(((ax_u8*)(emitter.labels.keys)));
            }
            if ((emitter.labels.values != ((ax_i64*)(NULL)))) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing emitter.labels.values\n", .len=42}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                ax_free(((ax_u8*)(emitter.labels.values)));
            }
            if ((emitter.fixups.data != ((struct ax_Fixup*)(NULL)))) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]     freeing emitter.fixups.data\n", .len=40}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                fflush(((void*)(NULL)));
                ax_free(((ax_u8*)(emitter.fixups.data)));
            }
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug]   freeing local resources done\n", .len=39}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
        }
        fi = (fi + 1);
    }
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: compiling loop finished, populating symbols starting\n", .len=84}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    ax_i32 text_sec_idx = ((ax_i32)(1));
    ax_i64 i = ((ax_i64)(0));
    while ((i < func_infos.len)) {
        struct ax_CompiledFuncInfo* info = &(((func_infos.data)[i]));
        if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"elf64", .len=5})) {
            ax_ELF64SymVec_push_elf64_sym(&(elf_syms), ((struct ax_ELF64Sym){.name_str=info->name, .value=((ax_u64)(info->offset)), .size=((ax_u64)(info->size)), .binding=ax_STB_GLOBAL, .sym_type=ax_STT_FUNC, .section=((ax_u16)(text_sec_idx))}));
        } else {
            {
                ax_COFFSymbolVec_push_coff_symbol(&(coff_syms), ((struct ax_COFFSymbol){.name_str=info->name, .value=info->offset, .section_num=((ax_i16)(text_sec_idx)), .sym_type=((ax_u16)(0x20)), .storage_class=ax_IMAGE_SYM_CLASS_EXTERNAL, .num_aux=((ax_u8)(0))}));
            }
        }
        i = (i + 1);
    }
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: populating symbols done\n", .len=55}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    struct ax_COFFRelocVec coff_relocs = ax_new_coff_reloc_vec();
    struct ax_RelocationVec elf_relocs = ax_new_relocation_vec();
    i = 0;
    while ((i < func_infos.len)) {
        struct ax_CompiledFuncInfo* info = &(((func_infos.data)[i]));
        ax_u32 func_offset = info->offset;
        ax_i64 r_i = ((ax_i64)(0));
        while ((r_i < info->relocs.len)) {
            struct ax_Relocation r = ((info->relocs.data)[r_i]);
            ax_i64 sym_idx = ((ax_i64)(r.sym_name));
            if ((r.sym_name > ((ax_u32)(0xF0000000)))) {
                sym_idx = (sym_idx - ((ax_i64)(4294967296)));
            }
            ax_string target_name = ax_resolve_binary_sym_name(sym_idx, info->name, symbols, pool);
            ax_i32 target_sym_idx = (-((ax_i32)(1)));
            if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"elf64", .len=5})) {
                ax_i64 sym_search = ((ax_i64)(0));
                while ((sym_search < elf_syms.len)) {
                    if (ax_str_eq(((elf_syms.data)[sym_search]).name_str, target_name)) {
                        target_sym_idx = ((ax_i32)((sym_search + ((ax_i64)(1)))));
                        break;
                    }
                    sym_search = (sym_search + 1);
                }
                if ((target_sym_idx == (-((ax_i32)(1))))) {
                    ax_i32 new_idx = (((ax_i32)(elf_syms.len)) + ((ax_i32)(1)));
                    ax_ELF64SymVec_push_elf64_sym(&(elf_syms), ((struct ax_ELF64Sym){.name_str=target_name, .value=((ax_u64)(0)), .size=((ax_u64)(0)), .binding=ax_STB_GLOBAL, .sym_type=ax_STT_FUNC, .section=((ax_u16)(0))}));
                    target_sym_idx = new_idx;
                }
            } else {
                {
                    ax_i64 sym_search = ((ax_i64)(0));
                    while ((sym_search < coff_syms.len)) {
                        if (ax_str_eq(((coff_syms.data)[sym_search]).name_str, target_name)) {
                            target_sym_idx = ((ax_i32)(sym_search));
                            break;
                        }
                        sym_search = (sym_search + 1);
                    }
                    if ((target_sym_idx == (-((ax_i32)(1))))) {
                        ax_i32 new_idx = ((ax_i32)(coff_syms.len));
                        ax_COFFSymbolVec_push_coff_symbol(&(coff_syms), ((struct ax_COFFSymbol){.name_str=target_name, .value=((ax_u32)(0)), .section_num=((ax_i16)(0)), .sym_type=((ax_u16)(0x20)), .storage_class=ax_IMAGE_SYM_CLASS_EXTERNAL, .num_aux=((ax_u8)(0))}));
                        target_sym_idx = new_idx;
                    }
                }
            }
            ax_u32 reloc_offset = (func_offset + ((ax_u32)(r.offset)));
            if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"elf64", .len=5})) {
                ax_RelocationVec_push_reloc(&(elf_relocs), ((struct ax_Relocation){.offset=((ax_i64)(reloc_offset)), .kind=r.kind, .padding1=((ax_u8)(0)), .padding2=((ax_u16)(0)), .sym_name=((ax_u32)(target_sym_idx)), .addend=r.addend}));
            } else {
                {
                    ax_u16 reloc_type = ax_IMAGE_REL_AMD64_REL32;
                    if ((r.kind == ax_RELOC_ABS64)) {
                        reloc_type = ax_IMAGE_REL_AMD64_ADDR64;
                    }
                    ax_COFFRelocVec_push_coff_reloc(&(coff_relocs), ((struct ax_COFFReloc){.virt_addr=reloc_offset, .sym_table_idx=((ax_u32)(target_sym_idx)), .reloc_type=reloc_type}));
                }
            }
            r_i = (r_i + 1);
        }
        i = (i + 1);
    }
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: populating relocs done, writing object file starting\n", .len=84}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    ax_bool success = AX_FALSE;
    if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"elf64", .len=5})) {
        success = ax_elf64_write_object_file(output_obj_file, &(all_code), &(elf_relocs), &(elf_syms));
    } else {
        {
            success = ax_coff_write_object_file(output_obj_file, &(all_code), &(rdata_buf), &(coff_relocs), &(coff_syms));
        }
    }
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: writing object file done, success=%d\n", .len=68}).ptr, ((ax_i64)(success)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: starting cleanup\n", .len=48}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    if ((all_code.data != ((ax_u8*)(NULL)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing all_code\n", .len=48}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_free(all_code.data);
    }
    if ((rdata_buf.data != ((ax_u8*)(NULL)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing rdata_buf\n", .len=49}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_free(rdata_buf.data);
    }
    if ((str_keys.data != ((ax_u32*)(NULL)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing str_keys\n", .len=48}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_free(((ax_u8*)(str_keys.data)));
    }
    if ((str_offsets.data != ((ax_u32*)(NULL)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing str_offsets\n", .len=51}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_free(((ax_u8*)(str_offsets.data)));
    }
    i = 0;
    while ((i < func_infos.len)) {
        struct ax_CompiledFuncInfo* info = &(((func_infos.data)[i]));
        if ((info->relocs.data != ((struct ax_Relocation*)(NULL)))) {
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing info.relocs for func %d\n", .len=63}).ptr, ((ax_i64)(i)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            fflush(((void*)(NULL)));
            ax_free(((ax_u8*)(info->relocs.data)));
        }
        i = (i + 1);
    }
    if ((func_infos.data != ((struct ax_CompiledFuncInfo*)(NULL)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing func_infos\n", .len=50}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_free(((ax_u8*)(func_infos.data)));
    }
    if ((coff_syms.data != ((struct ax_COFFSymbol*)(NULL)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing coff_syms\n", .len=49}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_free(((ax_u8*)(coff_syms.data)));
    }
    if ((coff_relocs.data != ((struct ax_COFFReloc*)(NULL)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing coff_relocs\n", .len=51}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_free(((ax_u8*)(coff_relocs.data)));
    }
    if ((elf_syms.data != ((struct ax_ELF64Sym*)(NULL)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing elf_syms\n", .len=48}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_free(((ax_u8*)(elf_syms.data)));
    }
    if ((elf_relocs.data != ((struct ax_Relocation*)(NULL)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: freeing elf_relocs\n", .len=50}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_free(((ax_u8*)(elf_relocs.data)));
    }
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] compile_native_binary: cleanup done\n", .len=44}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    return success;
}

ax_string ax_x86_slice_to_str(ax_u8* data, ax_i64 len) {
    ax_u8* buf = ((ax_u8*)(ax_alloc((len + 1))));
    memcpy(buf, data, len);
    ((buf)[len]) = ((ax_u8)(0));
    return ((ax_string){.ptr = (const ax_u8*)(buf), .len = strlen((const char*)(buf))});
}

ax_u16 ax_read_u16_le(ax_u8* data, ax_i64 off) {
    ax_u16 b0 = ((ax_u16)(((data)[off])));
    ax_u16 b1 = ((ax_u16)(((data)[(off + 1)])));
    return (b0 | (b1 << ((ax_u16)(8))));
}

ax_u32 ax_read_u32_le(ax_u8* data, ax_i64 off) {
    ax_u32 b0 = ((ax_u32)(((data)[off])));
    ax_u32 b1 = ((ax_u32)(((data)[(off + 1)])));
    ax_u32 b2 = ((ax_u32)(((data)[(off + 2)])));
    ax_u32 b3 = ((ax_u32)(((data)[(off + 3)])));
    return (((b0 | (b1 << ((ax_u32)(8)))) | (b2 << ((ax_u32)(16)))) | (b3 << ((ax_u32)(24))));
}

ax_u64 ax_read_u64_le(ax_u8* data, ax_i64 off) {
    ax_u64 b0 = ((ax_u64)(((data)[off])));
    ax_u64 b1 = ((ax_u64)(((data)[(off + 1)])));
    ax_u64 b2 = ((ax_u64)(((data)[(off + 2)])));
    ax_u64 b3 = ((ax_u64)(((data)[(off + 3)])));
    ax_u64 b4 = ((ax_u64)(((data)[(off + 4)])));
    ax_u64 b5 = ((ax_u64)(((data)[(off + 5)])));
    ax_u64 b6 = ((ax_u64)(((data)[(off + 6)])));
    ax_u64 b7 = ((ax_u64)(((data)[(off + 7)])));
    return (((((((b0 | (b1 << ((ax_u64)(8)))) | (b2 << ((ax_u64)(16)))) | (b3 << ((ax_u64)(24)))) | (b4 << ((ax_u64)(32)))) | (b5 << ((ax_u64)(40)))) | (b6 << ((ax_u64)(48)))) | (b7 << ((ax_u64)(56))));
}

void ax_write_u32_le(ax_u8* data, ax_i64 off, ax_u32 val) {
    ((data)[off]) = ((ax_u8)((val & ((ax_u32)(0xFF)))));
    ((data)[(off + 1)]) = ((ax_u8)(((val >> ((ax_u32)(8))) & ((ax_u32)(0xFF)))));
    ((data)[(off + 2)]) = ((ax_u8)(((val >> ((ax_u32)(16))) & ((ax_u32)(0xFF)))));
    ((data)[(off + 3)]) = ((ax_u8)(((val >> ((ax_u32)(24))) & ((ax_u32)(0xFF)))));
}

void ax_write_u64_le(ax_u8* data, ax_i64 off, ax_u64 val) {
    ((data)[off]) = ((ax_u8)((val & ((ax_u64)(0xFF)))));
    ((data)[(off + 1)]) = ((ax_u8)(((val >> ((ax_u64)(8))) & ((ax_u64)(0xFF)))));
    ((data)[(off + 2)]) = ((ax_u8)(((val >> ((ax_u64)(16))) & ((ax_u64)(0xFF)))));
    ((data)[(off + 3)]) = ((ax_u8)(((val >> ((ax_u64)(24))) & ((ax_u64)(0xFF)))));
    ((data)[(off + 4)]) = ((ax_u8)(((val >> ((ax_u64)(32))) & ((ax_u64)(0xFF)))));
    ((data)[(off + 5)]) = ((ax_u8)(((val >> ((ax_u64)(40))) & ((ax_u64)(0xFF)))));
    ((data)[(off + 6)]) = ((ax_u8)(((val >> ((ax_u64)(48))) & ((ax_u64)(0xFF)))));
    ((data)[(off + 7)]) = ((ax_u8)(((val >> ((ax_u64)(56))) & ((ax_u64)(0xFF)))));
}

struct ax_LinkerSymbolVec ax_new_linker_symbol_vec(void) {
    return ((struct ax_LinkerSymbolVec){.data=((struct ax_LinkerSymbol*)(NULL)), .len=0, .cap=0});
}

void ax_LinkerSymbolVec_push_linker_symbol(struct ax_LinkerSymbolVec* self, struct ax_LinkerSymbol s) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_LinkerSymbol* new_data = ((struct ax_LinkerSymbol*)(ax_alloc((new_cap * 40))));
        if ((self->data != ((struct ax_LinkerSymbol*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 40));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = s;
    self->len = (self->len + 1);
}

struct ax_ParsedRelocVec ax_new_parsed_reloc_vec(void) {
    return ((struct ax_ParsedRelocVec){.data=((struct ax_ParsedReloc*)(NULL)), .len=0, .cap=0});
}

void ax_ParsedRelocVec_push_parsed_reloc(struct ax_ParsedRelocVec* self, struct ax_ParsedReloc r) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_ParsedReloc* new_data = ((struct ax_ParsedReloc*)(ax_alloc((new_cap * 24))));
        if ((self->data != ((struct ax_ParsedReloc*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 24));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = r;
    self->len = (self->len + 1);
}

struct ax_LinkerStrVec ax_new_linker_str_vec(void) {
    return ((struct ax_LinkerStrVec){.data=((ax_string*)(NULL)), .len=0, .cap=0});
}

void ax_LinkerStrVec_push_linker_str(struct ax_LinkerStrVec* self, ax_string s) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        ax_string* new_data = ((ax_string*)(ax_alloc((new_cap * 16))));
        if ((self->data != ((ax_string*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 16));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = s;
    self->len = (self->len + 1);
}

struct ax_ParsedObjectPtrVec ax_new_parsed_object_ptr_vec(void) {
    return ((struct ax_ParsedObjectPtrVec){.data=((struct ax_ParsedObject**)(NULL)), .len=0, .cap=0});
}

void ax_ParsedObjectPtrVec_push_parsed_object_ptr(struct ax_ParsedObjectPtrVec* self, struct ax_ParsedObject* p) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        struct ax_ParsedObject** new_data = ((struct ax_ParsedObject**)(ax_alloc((new_cap * 8))));
        if ((self->data != ((struct ax_ParsedObject**)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 8));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    ((self->data)[self->len]) = p;
    self->len = (self->len + 1);
}

struct ax_ParsedObject* ax_linker_parse_elf(ax_u8* data, ax_i64 data_len) {
    ax_i64 shoff = ((ax_i64)(ax_read_u64_le(data, 40)));
    ax_i64 shnum = ((ax_i64)(ax_read_u16_le(data, 60)));
    ax_i64 text_off = ((ax_i64)(0));
    ax_i64 text_size = ((ax_i64)(0));
    ax_i64 symtab_off = ((ax_i64)(0));
    ax_i64 symtab_size = ((ax_i64)(0));
    ax_i64 strtab_off = ((ax_i64)(0));
    ax_i64 strtab_size = ((ax_i64)(0));
    ax_i64 rela_off = ((ax_i64)(0));
    ax_i64 rela_size = ((ax_i64)(0));
    ax_i64 i = ((ax_i64)(0));
    while ((i < shnum)) {
        ax_i64 off = (shoff + (i * 64));
        ax_u32 sh_type = ax_read_u32_le(data, (off + 4));
        ax_i64 sh_offset = ((ax_i64)(ax_read_u64_le(data, (off + 24))));
        ax_i64 sh_size = ((ax_i64)(ax_read_u64_le(data, (off + 32))));
        if ((sh_type == ((ax_u32)(1)))) {
            text_off = sh_offset;
            text_size = sh_size;
        } else if ((sh_type == ((ax_u32)(2)))) {
            symtab_off = sh_offset;
            symtab_size = sh_size;
        } else if ((sh_type == ((ax_u32)(3)))) {
            strtab_off = sh_offset;
            strtab_size = sh_size;
        } else if ((sh_type == ((ax_u32)(4)))) {
            rela_off = sh_offset;
            rela_size = sh_size;
        }
        i = (i + 1);
    }
    struct ax_LinkerStrVec sym_names = ax_new_linker_str_vec();
    if ((symtab_size > 0)) {
        ax_i64 num_syms = (symtab_size / 24);
        ax_i64 sym_i = ((ax_i64)(0));
        while ((sym_i < num_syms)) {
            ax_i64 off = (symtab_off + (sym_i * 24));
            ax_i64 name_offset = ((ax_i64)(ax_read_u32_le(data, off)));
            ax_string name_str = (ax_string){.ptr=(const ax_u8*)"", .len=0};
            if ((name_offset < strtab_size)) {
                ax_u8* str_ptr = ((ax_u8*)(((((ax_i64)(data)) + strtab_off) + name_offset)));
                ax_i64 char_len = ((ax_i64)(0));
                while ((((str_ptr)[char_len]) != ((ax_u8)(0)))) {
                    char_len = (char_len + 1);
                }
                name_str = ax_x86_slice_to_str(str_ptr, char_len);
            }
            ax_LinkerStrVec_push_linker_str(&(sym_names), name_str);
            sym_i = (sym_i + 1);
        }
    }
    struct ax_LinkerSymbolVec symbols = ax_new_linker_symbol_vec();
    if ((symtab_size > 0)) {
        ax_i64 num_syms = (symtab_size / 24);
        ax_i64 sym_i = ((ax_i64)(1));
        while ((sym_i < num_syms)) {
            ax_i64 off = (symtab_off + (sym_i * 24));
            ax_u16 sec_idx = ax_read_u16_le(data, (off + 6));
            ax_u64 val64 = ax_read_u64_le(data, (off + 8));
            ax_u64 size = ax_read_u64_le(data, (off + 16));
            ax_string name = ((sym_names.data)[sym_i]);
            ax_bool is_defined = (sec_idx != ((ax_u16)(0)));
            ax_LinkerSymbolVec_push_linker_symbol(&(symbols), ((struct ax_LinkerSymbol){.name=name, .section=((ax_i64)(sec_idx)), .offset=val64, .size=size, .defined=is_defined}));
            sym_i = (sym_i + 1);
        }
    }
    struct ax_ParsedRelocVec relocs = ax_new_parsed_reloc_vec();
    if ((rela_size > 0)) {
        ax_i64 num_relocs = (rela_size / 24);
        ax_i64 rel_i = ((ax_i64)(0));
        while ((rel_i < num_relocs)) {
            ax_i64 off = (rela_off + (rel_i * 24));
            ax_i64 r_offset = ((ax_i64)(ax_read_u64_le(data, off)));
            ax_u64 r_info = ax_read_u64_le(data, (off + 8));
            ax_i64 r_addend = ((ax_i64)(ax_read_u64_le(data, (off + 16))));
            ax_u32 sym_idx = ((ax_u32)((r_info >> ((ax_u64)(32)))));
            ax_u32 r_type = ((ax_u32)((r_info & ((ax_u64)(0xFFFFFFFF)))));
            ax_bool is_pc = ((r_type == ((ax_u32)(2))) || (r_type == ((ax_u32)(4))));
            ax_ParsedRelocVec_push_parsed_reloc(&(relocs), ((struct ax_ParsedReloc){.offset=r_offset, .sym_idx=sym_idx, .is_pc=is_pc, .addend=r_addend}));
            rel_i = (rel_i + 1);
        }
    }
    struct ax_ByteVec text_vec = ax_new_byte_vec();
    if ((text_size > 0)) {
        ax_ByteVec_push_bytes(&(text_vec), ((ax_u8*)((((ax_i64)(data)) + text_off))), text_size);
    }
    struct ax_ParsedObject* obj = ((struct ax_ParsedObject*)(ax_alloc(sizeof(struct ax_ParsedObject))));
    obj->text = text_vec;
    obj->rdata = ax_new_byte_vec();
    obj->symbols = symbols;
    obj->sym_names = sym_names;
    obj->relocs = relocs;
    obj->va = ((ax_u64)(0));
    return obj;
    ax_free(obj);
}

struct ax_ParsedObject* ax_linker_parse_coff(ax_u8* data, ax_i64 data_len) {
    ax_i64 num_sections = ((ax_i64)(ax_read_u16_le(data, 2)));
    ax_i64 symtab_off = ((ax_i64)(ax_read_u32_le(data, 8)));
    ax_i64 sym_count = ((ax_i64)(ax_read_u32_le(data, 12)));
    ax_i64 text_off = ((ax_i64)(0));
    ax_i64 text_size = ((ax_i64)(0));
    ax_i64 relocs_off = ((ax_i64)(0));
    ax_i64 num_relocs = ((ax_i64)(0));
    ax_i64 rdata_off = ((ax_i64)(0));
    ax_i64 rdata_size = ((ax_i64)(0));
    ax_i64 i = ((ax_i64)(0));
    while ((i < num_sections)) {
        ax_i64 off = (20 + (i * 40));
        ax_bool is_text = AX_FALSE;
        if ((((((((data)[off]) == ((ax_u8)('.'))) && (((data)[(off + 1)]) == ((ax_u8)('t')))) && (((data)[(off + 2)]) == ((ax_u8)('e')))) && (((data)[(off + 3)]) == ((ax_u8)('x')))) && (((data)[(off + 4)]) == ((ax_u8)('t'))))) {
            is_text = AX_TRUE;
        }
        ax_bool is_rdata = AX_FALSE;
        if (((((((((data)[off]) == ((ax_u8)('.'))) && (((data)[(off + 1)]) == ((ax_u8)('r')))) && (((data)[(off + 2)]) == ((ax_u8)('d')))) && (((data)[(off + 3)]) == ((ax_u8)('a')))) && (((data)[(off + 4)]) == ((ax_u8)('t')))) && (((data)[(off + 5)]) == ((ax_u8)('a'))))) {
            is_rdata = AX_TRUE;
        }
        if (is_text) {
            text_size = ((ax_i64)(ax_read_u32_le(data, (off + 16))));
            text_off = ((ax_i64)(ax_read_u32_le(data, (off + 20))));
            relocs_off = ((ax_i64)(ax_read_u32_le(data, (off + 24))));
            num_relocs = ((ax_i64)(ax_read_u16_le(data, (off + 32))));
        } else if (is_rdata) {
            rdata_size = ((ax_i64)(ax_read_u32_le(data, (off + 16))));
            rdata_off = ((ax_i64)(ax_read_u32_le(data, (off + 20))));
        }
        i = (i + 1);
    }
    ax_i64 strtab_off = (symtab_off + (sym_count * 18));
    ax_i64 strtab_size = (data_len - strtab_off);
    struct ax_LinkerStrVec sym_names = ax_new_linker_str_vec();
    struct ax_LinkerSymbolVec symbols = ax_new_linker_symbol_vec();
    ax_i64 sym_i = ((ax_i64)(0));
    while ((sym_i < sym_count)) {
        ax_i64 off = (symtab_off + (sym_i * 18));
        ax_string name_str = (ax_string){.ptr=(const ax_u8*)"", .len=0};
        ax_u32 first4 = ax_read_u32_le(data, off);
        if ((first4 == ((ax_u32)(0)))) {
            ax_i64 str_off = ((ax_i64)(ax_read_u32_le(data, (off + 4))));
            if ((str_off < strtab_size)) {
                ax_u8* str_ptr = ((ax_u8*)(((((ax_i64)(data)) + strtab_off) + str_off)));
                ax_i64 char_len = ((ax_i64)(0));
                while ((((str_ptr)[char_len]) != ((ax_u8)(0)))) {
                    char_len = (char_len + 1);
                }
                name_str = ax_x86_slice_to_str(str_ptr, char_len);
            }
        } else {
            {
                ax_u8* name_ptr = ((ax_u8*)((((ax_i64)(data)) + off)));
                ax_i64 char_len = ((ax_i64)(0));
                while (((char_len < 8) && (((name_ptr)[char_len]) != ((ax_u8)(0))))) {
                    char_len = (char_len + 1);
                }
                name_str = ax_x86_slice_to_str(name_ptr, char_len);
            }
        }
        ax_u32 val = ax_read_u32_le(data, (off + 8));
        ax_i16 sec_num = ((ax_i16)(ax_read_u16_le(data, (off + 12))));
        ax_u8 aux_count = ((data)[(off + 17)]);
        ax_LinkerStrVec_push_linker_str(&(sym_names), name_str);
        ax_LinkerSymbolVec_push_linker_symbol(&(symbols), ((struct ax_LinkerSymbol){.name=name_str, .section=((ax_i64)(sec_num)), .offset=((ax_u64)(val)), .defined=(sec_num > ((ax_i16)(0)))}));
        sym_i = (sym_i + 1);
        ax_u8 k = ((ax_u8)(0));
        while ((k < aux_count)) {
            ax_LinkerStrVec_push_linker_str(&(sym_names), (ax_string){.ptr=(const ax_u8*)"", .len=0});
            ax_LinkerSymbolVec_push_linker_symbol(&(symbols), ((struct ax_LinkerSymbol){.name=(ax_string){.ptr=(const ax_u8*)"", .len=0}, .section=0, .offset=((ax_u64)(0)), .defined=AX_FALSE}));
            sym_i = (sym_i + 1);
            k = (k + ((ax_u8)(1)));
        }
    }
    struct ax_ParsedRelocVec relocs = ax_new_parsed_reloc_vec();
    ax_i64 rel_idx = ((ax_i64)(0));
    while ((rel_idx < num_relocs)) {
        ax_i64 off = (relocs_off + (rel_idx * 10));
        ax_i64 virt_addr = ((ax_i64)(ax_read_u32_le(data, off)));
        ax_u32 sym_idx = ax_read_u32_le(data, (off + 4));
        ax_u16 rel_type = ax_read_u16_le(data, (off + 8));
        ax_bool is_pc = (rel_type == ((ax_u16)(4)));
        ax_i64 addend = ((ax_i64)(0));
        if (is_pc) {
            addend = (-((ax_i64)(4)));
        }
        ax_ParsedRelocVec_push_parsed_reloc(&(relocs), ((struct ax_ParsedReloc){.offset=virt_addr, .sym_idx=sym_idx, .is_pc=is_pc, .addend=addend}));
        rel_idx = (rel_idx + 1);
    }
    struct ax_ByteVec text_vec = ax_new_byte_vec();
    if ((text_size > 0)) {
        ax_ByteVec_push_bytes(&(text_vec), ((ax_u8*)((((ax_i64)(data)) + text_off))), text_size);
    }
    struct ax_ByteVec rdata_vec = ax_new_byte_vec();
    if ((rdata_size > 0)) {
        ax_ByteVec_push_bytes(&(rdata_vec), ((ax_u8*)((((ax_i64)(data)) + rdata_off))), rdata_size);
    }
    struct ax_ParsedObject* obj = ((struct ax_ParsedObject*)(ax_alloc(sizeof(struct ax_ParsedObject))));
    obj->text = text_vec;
    obj->rdata = rdata_vec;
    obj->symbols = symbols;
    obj->sym_names = sym_names;
    obj->relocs = relocs;
    obj->va = ((ax_u64)(0));
    return obj;
    ax_free(obj);
}

struct ax_AxiomLinker ax_new_axiom_linker(void) {
    return ((struct ax_AxiomLinker){.input_files=ax_new_linker_str_vec(), .output_path=(ax_string){.ptr=(const ax_u8*)"", .len=0}});
}

void ax_ByteVec_align_byte_vec(struct ax_ByteVec* vec, ax_i64 alignment) {
    while (((vec->len % alignment) != 0)) {
        ax_ByteVec_push_byte(vec, ((ax_u8)(0)));
    }
}

ax_string ax_get_dll_for_symbol(ax_string name) {
    if (((((((((ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"VirtualAlloc", .len=12}) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"VirtualFree", .len=11})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"ExitProcess", .len=11})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"GetLastError", .len=12})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"GetStdHandle", .len=12})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"WriteFile", .len=9})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"CreateFileA", .len=11})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"ReadFile", .len=8})) || ax_str_eq(name, (ax_string){.ptr=(const ax_u8*)"CloseHandle", .len=11}))) {
        return (ax_string){.ptr=(const ax_u8*)"kernel32.dll", .len=12};
    }
    ax_u8* ptr = ((ax_u8*)(name.ptr));
    ax_i64 len = ax_str_len(name);
    if (((((len >= 3) && (((ptr)[0]) == ((ax_u8)('a')))) && (((ptr)[1]) == ((ax_u8)('x')))) && (((ptr)[2]) == ((ax_u8)('_'))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_runtime.dll", .len=14};
    }
    if (((((((len >= 5) && (((ptr)[0]) == ((ax_u8)('_')))) && (((ptr)[1]) == ((ax_u8)('_')))) && (((ptr)[2]) == ((ax_u8)('a')))) && (((ptr)[3]) == ((ax_u8)('x')))) && (((ptr)[4]) == ((ax_u8)('_'))))) {
        return (ax_string){.ptr=(const ax_u8*)"ax_runtime.dll", .len=14};
    }
    return (ax_string){.ptr=(const ax_u8*)"ucrtbase.dll", .len=12};
}

void ax_LinkerStrVec_add_unique_import(struct ax_LinkerStrVec* imports, ax_string name) {
    ax_bool found = AX_FALSE;
    ax_i64 i = ((ax_i64)(0));
    while ((i < imports->len)) {
        if (ax_str_eq(((imports->data)[i]), name)) {
            found = AX_TRUE;
            break;
        }
        i = (i + 1);
    }
    if ((!found)) {
        ax_LinkerStrVec_push_linker_str(imports, name);
    }
}

void ax_write_buf_u32_le(ax_u8* data, ax_i64 off, ax_u32 val) {
    ((data)[off]) = ((ax_u8)((val & ((ax_u32)(0xFF)))));
    ((data)[(off + 1)]) = ((ax_u8)(((val >> ((ax_u32)(8))) & ((ax_u32)(0xFF)))));
    ((data)[(off + 2)]) = ((ax_u8)(((val >> ((ax_u32)(16))) & ((ax_u32)(0xFF)))));
    ((data)[(off + 3)]) = ((ax_u8)(((val >> ((ax_u32)(24))) & ((ax_u32)(0xFF)))));
}

void ax_write_buf_u64_le(ax_u8* data, ax_i64 off, ax_u64 val) {
    ax_write_buf_u32_le(data, off, ((ax_u32)((val & ((ax_u64)(0xFFFFFFFF))))));
    ax_write_buf_u32_le(data, (off + 4), ((ax_u32)(((val >> ((ax_u64)(32))) & ((ax_u64)(0xFFFFFFFF))))));
}

void ax_ByteVec_linker_build_pe_headers(struct ax_ByteVec* out, ax_u32 code_raw_size, ax_u32 idata_raw_size, ax_u32 entry_rva, ax_u32 idata_rva, ax_u32 idata_size, ax_u32 iat_rva, ax_u32 iat_size) {
    ax_ByteVec_push_byte(out, ((ax_u8)(0x4D)));
    ax_ByteVec_push_byte(out, ((ax_u8)(0x5A)));
    ax_i64 i = ((ax_i64)(2));
    while ((i < 0x3C)) {
        ax_ByteVec_push_byte(out, ((ax_u8)(0)));
        i = (i + 1);
    }
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0x80)));
    i = 0x40;
    while ((i < 0x80)) {
        ax_ByteVec_push_byte(out, ((ax_u8)(0)));
        i = (i + 1);
    }
    ax_ByteVec_push_byte(out, ((ax_u8)(0x50)));
    ax_ByteVec_push_byte(out, ((ax_u8)(0x45)));
    ax_ByteVec_push_byte(out, ((ax_u8)(0)));
    ax_ByteVec_push_byte(out, ((ax_u8)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0x8664)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(2)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(240)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0x0023)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0x020B)));
    ax_ByteVec_push_byte(out, ((ax_u8)(1)));
    ax_ByteVec_push_byte(out, ((ax_u8)(0)));
    ax_ByteVec_push_u32_le(out, code_raw_size);
    ax_ByteVec_push_u32_le(out, idata_raw_size);
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out, entry_rva);
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0x1000)));
    ax_ByteVec_push_u64_le(out, ((ax_u64)(0x400000)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0x1000)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0x200)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(6)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(6)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_u32 size_of_image = (((idata_rva + idata_raw_size) + ((ax_u32)(0xFFF))) & ((ax_u32)(0xFFFFF000)));
    ax_ByteVec_push_u32_le(out, size_of_image);
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0x400)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(3)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0x8100)));
    ax_ByteVec_push_u64_le(out, ((ax_u64)(0x100000)));
    ax_ByteVec_push_u64_le(out, ((ax_u64)(0x1000)));
    ax_ByteVec_push_u64_le(out, ((ax_u64)(0x100000)));
    ax_ByteVec_push_u64_le(out, ((ax_u64)(0x1000)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(16)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out, idata_rva);
    ax_ByteVec_push_u32_le(out, idata_size);
    ax_i64 dir_i = ((ax_i64)(2));
    while ((dir_i < 12)) {
        ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
        ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
        dir_i = (dir_i + 1);
    }
    ax_ByteVec_push_u32_le(out, iat_rva);
    ax_ByteVec_push_u32_le(out, iat_size);
    dir_i = 13;
    while ((dir_i < 16)) {
        ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
        ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
        dir_i = (dir_i + 1);
    }
    ax_ByteVec_push_byte(out, ((ax_u8)('.')));
    ax_ByteVec_push_byte(out, ((ax_u8)('t')));
    ax_ByteVec_push_byte(out, ((ax_u8)('e')));
    ax_ByteVec_push_byte(out, ((ax_u8)('x')));
    ax_ByteVec_push_byte(out, ((ax_u8)('t')));
    ax_ByteVec_push_byte(out, ((ax_u8)(0)));
    ax_ByteVec_push_byte(out, ((ax_u8)(0)));
    ax_ByteVec_push_byte(out, ((ax_u8)(0)));
    ax_ByteVec_push_u32_le(out, code_raw_size);
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0x1000)));
    ax_ByteVec_push_u32_le(out, code_raw_size);
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0x400)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0x60000020)));
    ax_ByteVec_push_byte(out, ((ax_u8)('.')));
    ax_ByteVec_push_byte(out, ((ax_u8)('i')));
    ax_ByteVec_push_byte(out, ((ax_u8)('d')));
    ax_ByteVec_push_byte(out, ((ax_u8)('a')));
    ax_ByteVec_push_byte(out, ((ax_u8)('t')));
    ax_ByteVec_push_byte(out, ((ax_u8)('a')));
    ax_ByteVec_push_byte(out, ((ax_u8)(0)));
    ax_ByteVec_push_byte(out, ((ax_u8)(0)));
    ax_ByteVec_push_u32_le(out, idata_raw_size);
    ax_ByteVec_push_u32_le(out, idata_rva);
    ax_ByteVec_push_u32_le(out, idata_raw_size);
    ax_ByteVec_push_u32_le(out, (((ax_u32)(0x400)) + code_raw_size));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(0xC0000040)));
    ax_ByteVec_align_byte_vec(out, ((ax_i64)(0x400)));
}

void ax_ByteVec_linker_build_elf_headers(struct ax_ByteVec* out, ax_u32 code_raw_size, ax_u64 entry_va, ax_u32 dynamic_offset, ax_u32 dynamic_size) {
    ax_ByteVec_push_byte(out, ((ax_u8)(0x7F)));
    ax_ByteVec_push_byte(out, ((ax_u8)('E')));
    ax_ByteVec_push_byte(out, ((ax_u8)('L')));
    ax_ByteVec_push_byte(out, ((ax_u8)('F')));
    ax_ByteVec_push_byte(out, ((ax_u8)(2)));
    ax_ByteVec_push_byte(out, ((ax_u8)(1)));
    ax_ByteVec_push_byte(out, ((ax_u8)(1)));
    ax_ByteVec_push_byte(out, ((ax_u8)(0)));
    ax_i64 i = ((ax_i64)(8));
    while ((i < 16)) {
        ax_ByteVec_push_byte(out, ((ax_u8)(0)));
        i = (i + 1);
    }
    ax_ByteVec_push_u16_le(out, ((ax_u16)(2)));
    ax_ByteVec_push_u16_le(out, ((ax_u16)(62)));
    ax_ByteVec_push_u32_le(out, ((ax_u32)(1)));
    ax_ByteVec_push_u64_le(out, entry_va);
    ax_ByteVec_push_u64_le(out, ((ax_u64)(64)));
    ax_ByteVec_push_u64_le(out, ((ax_u64)(0)));
    if ((dynamic_size > 0)) {
        ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
        ax_ByteVec_push_u16_le(out, ((ax_u16)(64)));
        ax_ByteVec_push_u16_le(out, ((ax_u16)(56)));
        ax_ByteVec_push_u16_le(out, ((ax_u16)(3)));
        ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
        ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
        ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
        ax_ByteVec_push_u32_le(out, ((ax_u32)(3)));
        ax_ByteVec_push_u32_le(out, ((ax_u32)(4)));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(232)));
        ax_ByteVec_push_u64_le(out, (((ax_u64)(0x400000)) + ((ax_u64)(232))));
        ax_ByteVec_push_u64_le(out, (((ax_u64)(0x400000)) + ((ax_u64)(232))));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(28)));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(28)));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(1)));
        ax_ByteVec_push_u32_le(out, ((ax_u32)(1)));
        ax_ByteVec_push_u32_le(out, ((ax_u32)(7)));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(0)));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(0x400000)));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(0x400000)));
        ax_ByteVec_push_u64_le(out, (((ax_u64)(512)) + ((ax_u64)(code_raw_size))));
        ax_ByteVec_push_u64_le(out, (((ax_u64)(512)) + ((ax_u64)(code_raw_size))));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(4096)));
        ax_ByteVec_push_u32_le(out, ((ax_u32)(2)));
        ax_ByteVec_push_u32_le(out, ((ax_u32)(6)));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(dynamic_offset)));
        ax_ByteVec_push_u64_le(out, (((ax_u64)(0x400000)) + ((ax_u64)(dynamic_offset))));
        ax_ByteVec_push_u64_le(out, (((ax_u64)(0x400000)) + ((ax_u64)(dynamic_offset))));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(dynamic_size)));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(dynamic_size)));
        ax_ByteVec_push_u64_le(out, ((ax_u64)(8)));
        ax_u8* interp_str = ((ax_u8*)((ax_string){.ptr=(const ax_u8*)"/lib64/ld-linux-x86-64.so.2", .len=27}.ptr));
        ax_i64 c_i = ((ax_i64)(0));
        while ((c_i < 27)) {
            ax_ByteVec_push_byte(out, ((interp_str)[c_i]));
            c_i = (c_i + 1);
        }
        ax_ByteVec_push_byte(out, ((ax_u8)(0)));
        ax_ByteVec_align_byte_vec(out, ((ax_i64)(512)));
    } else {
        {
            ax_ByteVec_push_u32_le(out, ((ax_u32)(0)));
            ax_ByteVec_push_u16_le(out, ((ax_u16)(64)));
            ax_ByteVec_push_u16_le(out, ((ax_u16)(56)));
            ax_ByteVec_push_u16_le(out, ((ax_u16)(1)));
            ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
            ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
            ax_ByteVec_push_u16_le(out, ((ax_u16)(0)));
            ax_ByteVec_push_u32_le(out, ((ax_u32)(1)));
            ax_ByteVec_push_u32_le(out, ((ax_u32)(5)));
            ax_ByteVec_push_u64_le(out, ((ax_u64)(0)));
            ax_ByteVec_push_u64_le(out, ((ax_u64)(0x400000)));
            ax_ByteVec_push_u64_le(out, ((ax_u64)(0x400000)));
            ax_ByteVec_push_u64_le(out, (((ax_u64)(128)) + ((ax_u64)(code_raw_size))));
            ax_ByteVec_push_u64_le(out, (((ax_u64)(128)) + ((ax_u64)(code_raw_size))));
            ax_ByteVec_push_u64_le(out, ((ax_u64)(0x1000)));
            ax_ByteVec_align_byte_vec(out, ((ax_i64)(128)));
        }
    }
}

void ax_AxiomLinker_axiom_linker_add_input(struct ax_AxiomLinker* self, ax_string file) {
    ax_LinkerStrVec_push_linker_str(&(self->input_files), file);
}

ax_bool ax_AxiomLinker_axiom_linker_link(struct ax_AxiomLinker* self) {
    struct ax_ParsedObjectPtrVec objects = ax_new_parsed_object_ptr_vec();
    ax_string format = (ax_string){.ptr=(const ax_u8*)"coff", .len=4};
    ax_i64 kernel32_iat_offset = ((ax_i64)(0));
    ax_i64 iat_offset = ((ax_i64)(0));
    ax_i64 total_iat_size = ((ax_i64)(0));
    ax_i64 i = ((ax_i64)(0));
    while ((i < self->input_files.len)) {
        ax_string path = ((self->input_files.data)[i]);
        ax_u8* path_nt = ax_x86_str_to_null_terminated(path);
        void* file = fopen(((void*)(path_nt)), (const char*)((ax_string){.ptr=(const ax_u8*)"rb", .len=2}).ptr);
        ax_free(path_nt);
        if ((file == ((void*)(NULL)))) {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: Linker could not open input file: ", .len=41}).ptr);
            puts((const char*)(path).ptr);
            return AX_FALSE;
        }
        fseek(file, 0, 2);
        ax_i64 size = ftell(file);
        rewind(file);
        if ((size <= 4)) {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: Input file too small: ", .len=29}).ptr);
            puts((const char*)(path).ptr);
            fclose(file);
            return AX_FALSE;
        }
        ax_u8* buffer = ((ax_u8*)(ax_alloc(size)));
        fread(((void*)(buffer)), 1, size, file);
        fclose(file);
        struct ax_ParsedObject* obj = ((struct ax_ParsedObject*)(NULL));
        if (((((((buffer)[0]) == ((ax_u8)(0x7F))) && (((buffer)[1]) == ((ax_u8)('E')))) && (((buffer)[2]) == ((ax_u8)('L')))) && (((buffer)[3]) == ((ax_u8)('F'))))) {
            obj = ax_linker_parse_elf(buffer, size);
            format = (ax_string){.ptr=(const ax_u8*)"elf64", .len=5};
        } else if ((ax_read_u16_le(buffer, 0) == ((ax_u16)(0x8664)))) {
            obj = ax_linker_parse_coff(buffer, size);
            format = (ax_string){.ptr=(const ax_u8*)"coff", .len=4};
        } else {
            {
                puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: Unsupported object format in: ", .len=37}).ptr);
                puts((const char*)(path).ptr);
                ax_free(buffer);
                return AX_FALSE;
            }
        }
        ax_free(buffer);
        ax_ParsedObjectPtrVec_push_parsed_object_ptr(&(objects), obj);
        i = (i + 1);
    }
    struct ax_LinkerStrVec func_names = ax_new_linker_str_vec();
    i = 0;
    while ((i < objects.len)) {
        struct ax_ParsedObject* obj = ((objects.data)[i]);
        ax_i64 sym_i = ((ax_i64)(0));
        while ((sym_i < obj->symbols.len)) {
            struct ax_LinkerSymbol* sym = &(((obj->symbols.data)[sym_i]));
            if (sym->defined) {
                ax_LinkerStrVec_push_linker_str(&(func_names), sym->name);
            }
            sym_i = (sym_i + 1);
        }
        i = (i + 1);
    }
    struct ax_LinkerStrVec kernel32_imports = ax_new_linker_str_vec();
    struct ax_LinkerStrVec ax_runtime_imports = ax_new_linker_str_vec();
    struct ax_LinkerStrVec ucrtbase_imports = ax_new_linker_str_vec();
    i = 0;
    while ((i < objects.len)) {
        struct ax_ParsedObject* obj = ((objects.data)[i]);
        ax_i64 r_i = ((ax_i64)(0));
        while ((r_i < obj->relocs.len)) {
            struct ax_ParsedReloc* r = &(((obj->relocs.data)[r_i]));
            if ((r->sym_idx < ((ax_u32)(obj->sym_names.len)))) {
                ax_string target_name = ((obj->sym_names.data)[r->sym_idx]);
                if ((((!ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"", .len=0})) && (!ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"main", .len=4}))) && (!ax_str_eq(target_name, (ax_string){.ptr=(const ax_u8*)"_AX_main_main_v_v", .len=17})))) {
                    ax_bool is_defined = AX_FALSE;
                    ax_i64 search = ((ax_i64)(0));
                    while ((search < func_names.len)) {
                        if (ax_str_eq(((func_names.data)[search]), target_name)) {
                            is_defined = AX_TRUE;
                            break;
                        }
                        search = (search + 1);
                    }
                    if ((!is_defined)) {
                        ax_string dll = ax_get_dll_for_symbol(target_name);
                        if (ax_str_eq(dll, (ax_string){.ptr=(const ax_u8*)"kernel32.dll", .len=12})) {
                            ax_LinkerStrVec_add_unique_import(&(kernel32_imports), target_name);
                        } else if (ax_str_eq(dll, (ax_string){.ptr=(const ax_u8*)"ax_runtime.dll", .len=14})) {
                            ax_LinkerStrVec_add_unique_import(&(ax_runtime_imports), target_name);
                        } else {
                            {
                                ax_LinkerStrVec_add_unique_import(&(ucrtbase_imports), target_name);
                            }
                        }
                    }
                }
            }
            r_i = (r_i + 1);
        }
        i = (i + 1);
    }
    ax_i64 thunk_count = ((kernel32_imports.len + ax_runtime_imports.len) + ucrtbase_imports.len);
    struct ax_ByteVec merged_code = ax_new_byte_vec();
    ax_u64 base_addr = ((ax_u64)(0x401000));
    if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"elf64", .len=5})) {
        if ((thunk_count > 0)) {
            base_addr = ((ax_u64)(0x400200));
        } else {
            {
                base_addr = ((ax_u64)(0x400080));
            }
        }
    }
    struct ax_ParsedRelocVec func_offsets = ax_new_parsed_reloc_vec();
    i = 0;
    while ((i < objects.len)) {
        struct ax_ParsedObject* obj = ((objects.data)[i]);
        obj->va = (base_addr + ((ax_u64)(merged_code.len)));
        ax_i64 sym_i = ((ax_i64)(0));
        while ((sym_i < obj->symbols.len)) {
            struct ax_LinkerSymbol* sym = &(((obj->symbols.data)[sym_i]));
            if (sym->defined) {
                ax_u64 offset_va = sym->offset;
                if ((sym->section == 2)) {
                    offset_va = (((ax_u64)(obj->text.len)) + sym->offset);
                }
                ax_ParsedRelocVec_push_parsed_reloc(&(func_offsets), ((struct ax_ParsedReloc){.offset=((ax_i64)((((ax_u64)(merged_code.len)) + offset_va))), .sym_idx=((ax_u32)(0)), .is_pc=AX_FALSE, .addend=((ax_i64)(0))}));
            }
            sym_i = (sym_i + 1);
        }
        ax_ByteVec_push_bytes(&(merged_code), obj->text.data, obj->text.len);
        ax_ByteVec_push_bytes(&(merged_code), obj->rdata.data, obj->rdata.len);
        i = (i + 1);
    }
    ax_i64 thunks_offset = ((merged_code.len + 15) & (~15));
    ax_u32 thunks_rva = (((ax_u32)(0x1000)) + ((ax_u32)(thunks_offset)));
    ax_i64 thunks_size = (thunk_count * 6);
    ax_i64 idata_rva = ((((((ax_i64)(0x1000)) + thunks_offset) + ((ax_i64)(thunks_size))) + 4095) & (~4095));
    struct ax_ByteVec idata_buf = ax_new_byte_vec();
    struct ax_LinkerStrVec dyn_sym_names = ax_new_linker_str_vec();
    ax_u32* dyn_sym_rvas = ((ax_u32*)(NULL));
    ax_u32* kernel32_hn_rvas = ((ax_u32*)(NULL));
    ax_u32* ax_runtime_hn_rvas = ((ax_u32*)(NULL));
    ax_u32* ucrtbase_hn_rvas = ((ax_u32*)(NULL));
    if ((ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"coff", .len=4}) && (thunk_count > 0))) {
        dyn_sym_rvas = ((ax_u32*)(ax_alloc((thunk_count * 4))));
        ax_bool has_kernel32 = (kernel32_imports.len > 0);
        ax_bool has_ax_runtime = (ax_runtime_imports.len > 0);
        ax_bool has_ucrtbase = (ucrtbase_imports.len > 0);
        ax_i64 K = ((ax_i64)(0));
        if (has_kernel32) {
            K = (K + 1);
        }
        if (has_ax_runtime) {
            K = (K + 1);
        }
        if (has_ucrtbase) {
            K = (K + 1);
        }
        ax_i64 idt_size = ((K + 1) * 20);
        iat_offset = idt_size;
        ax_i64 kernel32_iat_offset = ((ax_i64)(0));
        ax_i64 kernel32_iat_size = ((ax_i64)(0));
        ax_i64 ax_runtime_iat_offset = iat_offset;
        ax_i64 ax_runtime_iat_size = ((ax_i64)(0));
        ax_i64 ucrtbase_iat_offset = iat_offset;
        ax_i64 ucrtbase_iat_size = ((ax_i64)(0));
        ax_i64 cur_iat_offset = idt_size;
        if (has_kernel32) {
            kernel32_iat_offset = cur_iat_offset;
            kernel32_iat_size = ((kernel32_imports.len + 1) * 8);
            cur_iat_offset = (cur_iat_offset + kernel32_iat_size);
        }
        if (has_ax_runtime) {
            ax_runtime_iat_offset = cur_iat_offset;
            ax_runtime_iat_size = ((ax_runtime_imports.len + 1) * 8);
            cur_iat_offset = (cur_iat_offset + ax_runtime_iat_size);
        }
        if (has_ucrtbase) {
            ucrtbase_iat_offset = cur_iat_offset;
            ucrtbase_iat_size = ((ucrtbase_imports.len + 1) * 8);
            cur_iat_offset = (cur_iat_offset + ucrtbase_iat_size);
        }
        total_iat_size = (cur_iat_offset - iat_offset);
        ax_i64 ilt_offset = (iat_offset + total_iat_size);
        ax_i64 cur_ilt_offset = ilt_offset;
        ax_i64 kernel32_ilt_offset = ((ax_i64)(0));
        ax_i64 kernel32_ilt_size = ((ax_i64)(0));
        if (has_kernel32) {
            kernel32_ilt_offset = cur_ilt_offset;
            kernel32_ilt_size = ((kernel32_imports.len + 1) * 8);
            cur_ilt_offset = (cur_ilt_offset + kernel32_ilt_size);
        }
        ax_i64 ax_runtime_ilt_offset = ((ax_i64)(0));
        ax_i64 ax_runtime_ilt_size = ((ax_i64)(0));
        if (has_ax_runtime) {
            ax_runtime_ilt_offset = cur_ilt_offset;
            ax_runtime_ilt_size = ((ax_runtime_imports.len + 1) * 8);
            cur_ilt_offset = (cur_ilt_offset + ax_runtime_ilt_size);
        }
        ax_i64 ucrtbase_ilt_offset = ((ax_i64)(0));
        ax_i64 ucrtbase_ilt_size = ((ax_i64)(0));
        if (has_ucrtbase) {
            ucrtbase_ilt_offset = cur_ilt_offset;
            ucrtbase_ilt_size = ((ucrtbase_imports.len + 1) * 8);
            cur_ilt_offset = (cur_ilt_offset + ucrtbase_ilt_size);
        }
        ax_i64 total_ilt_size = (cur_ilt_offset - ilt_offset);
        ax_i64 dll_names_offset = (ilt_offset + total_ilt_size);
        ax_i64 cur_offset = dll_names_offset;
        ax_i64 kernel32_name_offset = ((ax_i64)(0));
        if (has_kernel32) {
            kernel32_name_offset = cur_offset;
            cur_offset = (cur_offset + 13);
        }
        ax_i64 ax_runtime_name_offset = ((ax_i64)(0));
        if (has_ax_runtime) {
            ax_runtime_name_offset = cur_offset;
            cur_offset = (cur_offset + 15);
        }
        ax_i64 ucrtbase_name_offset = ((ax_i64)(0));
        if (has_ucrtbase) {
            ucrtbase_name_offset = cur_offset;
            cur_offset = (cur_offset + 13);
        }
        ax_i64 hint_name_start_offset = cur_offset;
        ax_i64 zero_i = ((ax_i64)(0));
        while ((zero_i < ((ax_i64)(hint_name_start_offset)))) {
            ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
            zero_i = (zero_i + 1);
        }
        if (has_kernel32) {
            kernel32_hn_rvas = ((ax_u32*)(ax_alloc((kernel32_imports.len * 4))));
            ax_i64 j = ((ax_i64)(0));
            while ((j < kernel32_imports.len)) {
                ((kernel32_hn_rvas)[j]) = ((ax_u32)((((ax_i64)(idata_rva)) + idata_buf.len)));
                ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                ax_string name = ((kernel32_imports.data)[j]);
                ax_i64 len = ax_str_len(name);
                ax_u8* ptr = ((ax_u8*)(name.ptr));
                ax_i64 c_i = ((ax_i64)(0));
                while ((c_i < len)) {
                    ax_ByteVec_push_byte(&(idata_buf), ((ptr)[c_i]));
                    c_i = (c_i + 1);
                }
                ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                if ((((len + 3) % 2) != 0)) {
                    ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                }
                j = (j + 1);
            }
        }
        if (has_ax_runtime) {
            ax_runtime_hn_rvas = ((ax_u32*)(ax_alloc((ax_runtime_imports.len * 4))));
            ax_i64 j = ((ax_i64)(0));
            while ((j < ax_runtime_imports.len)) {
                ((ax_runtime_hn_rvas)[j]) = ((ax_u32)((((ax_i64)(idata_rva)) + idata_buf.len)));
                ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                ax_string name = ((ax_runtime_imports.data)[j]);
                ax_i64 len = ax_str_len(name);
                ax_u8* ptr = ((ax_u8*)(name.ptr));
                ax_i64 c_i = ((ax_i64)(0));
                while ((c_i < len)) {
                    ax_ByteVec_push_byte(&(idata_buf), ((ptr)[c_i]));
                    c_i = (c_i + 1);
                }
                ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                if ((((len + 3) % 2) != 0)) {
                    ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                }
                j = (j + 1);
            }
        }
        if (has_ucrtbase) {
            ucrtbase_hn_rvas = ((ax_u32*)(ax_alloc((ucrtbase_imports.len * 4))));
            ax_i64 j = ((ax_i64)(0));
            while ((j < ucrtbase_imports.len)) {
                ((ucrtbase_hn_rvas)[j]) = ((ax_u32)((((ax_i64)(idata_rva)) + idata_buf.len)));
                ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                ax_string name = ((ucrtbase_imports.data)[j]);
                ax_i64 len = ax_str_len(name);
                ax_u8* ptr = ((ax_u8*)(name.ptr));
                ax_i64 c_i = ((ax_i64)(0));
                while ((c_i < len)) {
                    ax_ByteVec_push_byte(&(idata_buf), ((ptr)[c_i]));
                    c_i = (c_i + 1);
                }
                ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                if ((((len + 3) % 2) != 0)) {
                    ax_ByteVec_push_byte(&(idata_buf), ((ax_u8)(0)));
                }
                j = (j + 1);
            }
        }
        if (has_kernel32) {
            ax_u8* name_ptr = ((ax_u8*)((ax_string){.ptr=(const ax_u8*)"kernel32.dll", .len=12}.ptr));
            ax_i64 c_i = ((ax_i64)(0));
            while ((c_i < 12)) {
                ((idata_buf.data)[(((ax_i64)(kernel32_name_offset)) + c_i)]) = ((name_ptr)[c_i]);
                c_i = (c_i + 1);
            }
            ((idata_buf.data)[(((ax_i64)(kernel32_name_offset)) + 12)]) = ((ax_u8)(0));
        }
        if (has_ax_runtime) {
            ax_u8* name_ptr = ((ax_u8*)((ax_string){.ptr=(const ax_u8*)"ax_runtime.dll", .len=14}.ptr));
            ax_i64 c_i = ((ax_i64)(0));
            while ((c_i < 14)) {
                ((idata_buf.data)[(((ax_i64)(ax_runtime_name_offset)) + c_i)]) = ((name_ptr)[c_i]);
                c_i = (c_i + 1);
            }
            ((idata_buf.data)[(((ax_i64)(ax_runtime_name_offset)) + 14)]) = ((ax_u8)(0));
        }
        if (has_ucrtbase) {
            ax_u8* name_ptr = ((ax_u8*)((ax_string){.ptr=(const ax_u8*)"ucrtbase.dll", .len=12}.ptr));
            ax_i64 c_i = ((ax_i64)(0));
            while ((c_i < 12)) {
                ((idata_buf.data)[(((ax_i64)(ucrtbase_name_offset)) + c_i)]) = ((name_ptr)[c_i]);
                c_i = (c_i + 1);
            }
            ((idata_buf.data)[(((ax_i64)(ucrtbase_name_offset)) + 12)]) = ((ax_u8)(0));
        }
        if (has_kernel32) {
            ax_i64 j = ((ax_i64)(0));
            while ((j < kernel32_imports.len)) {
                ax_u64 val = ((ax_u64)(((kernel32_hn_rvas)[j])));
                ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(kernel32_iat_offset)) + (j * 8)), val);
                ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(kernel32_ilt_offset)) + (j * 8)), val);
                j = (j + 1);
            }
            ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(kernel32_iat_offset)) + (kernel32_imports.len * 8)), ((ax_u64)(0)));
            ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(kernel32_ilt_offset)) + (kernel32_imports.len * 8)), ((ax_u64)(0)));
        }
        if (has_ax_runtime) {
            ax_i64 j = ((ax_i64)(0));
            while ((j < ax_runtime_imports.len)) {
                ax_u64 val = ((ax_u64)(((ax_runtime_hn_rvas)[j])));
                ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(ax_runtime_iat_offset)) + (j * 8)), val);
                ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(ax_runtime_ilt_offset)) + (j * 8)), val);
                j = (j + 1);
            }
            ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(ax_runtime_iat_offset)) + (ax_runtime_imports.len * 8)), ((ax_u64)(0)));
            ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(ax_runtime_ilt_offset)) + (ax_runtime_imports.len * 8)), ((ax_u64)(0)));
        }
        if (has_ucrtbase) {
            ax_i64 j = ((ax_i64)(0));
            while ((j < ucrtbase_imports.len)) {
                ax_u64 val = ((ax_u64)(((ucrtbase_hn_rvas)[j])));
                ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(ucrtbase_iat_offset)) + (j * 8)), val);
                ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(ucrtbase_ilt_offset)) + (j * 8)), val);
                j = (j + 1);
            }
            ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(ucrtbase_iat_offset)) + (ucrtbase_imports.len * 8)), ((ax_u64)(0)));
            ax_write_buf_u64_le(idata_buf.data, (((ax_i64)(ucrtbase_ilt_offset)) + (ucrtbase_imports.len * 8)), ((ax_u64)(0)));
        }
        ax_i64 entry_idx = ((ax_i64)(0));
        if (has_kernel32) {
            ax_u32 ilt_rva = ((ax_u32)((((ax_i64)(idata_rva)) + ((ax_i64)(kernel32_ilt_offset)))));
            ax_u32 name_rva = ((ax_u32)((((ax_i64)(idata_rva)) + ((ax_i64)(kernel32_name_offset)))));
            ax_u32 iat_rva = ((ax_u32)((((ax_i64)(idata_rva)) + ((ax_i64)(kernel32_iat_offset)))));
            ax_i64 off = (entry_idx * 20);
            ax_write_buf_u32_le(idata_buf.data, off, ilt_rva);
            ax_write_buf_u32_le(idata_buf.data, (off + 4), ((ax_u32)(0)));
            ax_write_buf_u32_le(idata_buf.data, (off + 8), ((ax_u32)(0)));
            ax_write_buf_u32_le(idata_buf.data, (off + 12), name_rva);
            ax_write_buf_u32_le(idata_buf.data, (off + 16), iat_rva);
            entry_idx = (entry_idx + 1);
        }
        if (has_ax_runtime) {
            ax_u32 ilt_rva = ((ax_u32)((((ax_i64)(idata_rva)) + ((ax_i64)(ax_runtime_ilt_offset)))));
            ax_u32 name_rva = ((ax_u32)((((ax_i64)(idata_rva)) + ((ax_i64)(ax_runtime_name_offset)))));
            ax_u32 iat_rva = ((ax_u32)((((ax_i64)(idata_rva)) + ((ax_i64)(ax_runtime_iat_offset)))));
            ax_i64 off = (entry_idx * 20);
            ax_write_buf_u32_le(idata_buf.data, off, ilt_rva);
            ax_write_buf_u32_le(idata_buf.data, (off + 4), ((ax_u32)(0)));
            ax_write_buf_u32_le(idata_buf.data, (off + 8), ((ax_u32)(0)));
            ax_write_buf_u32_le(idata_buf.data, (off + 12), name_rva);
            ax_write_buf_u32_le(idata_buf.data, (off + 16), iat_rva);
            entry_idx = (entry_idx + 1);
        }
        if (has_ucrtbase) {
            ax_u32 ilt_rva = ((ax_u32)((((ax_i64)(idata_rva)) + ((ax_i64)(ucrtbase_ilt_offset)))));
            ax_u32 name_rva = ((ax_u32)((((ax_i64)(idata_rva)) + ((ax_i64)(ucrtbase_name_offset)))));
            ax_u32 iat_rva = ((ax_u32)((((ax_i64)(idata_rva)) + ((ax_i64)(ucrtbase_iat_offset)))));
            ax_i64 off = (entry_idx * 20);
            ax_write_buf_u32_le(idata_buf.data, off, ilt_rva);
            ax_write_buf_u32_le(idata_buf.data, (off + 4), ((ax_u32)(0)));
            ax_write_buf_u32_le(idata_buf.data, (off + 8), ((ax_u32)(0)));
            ax_write_buf_u32_le(idata_buf.data, (off + 12), name_rva);
            ax_write_buf_u32_le(idata_buf.data, (off + 16), iat_rva);
            entry_idx = (entry_idx + 1);
        }
        ax_i64 off = (entry_idx * 20);
        ax_write_buf_u32_le(idata_buf.data, off, ((ax_u32)(0)));
        ax_write_buf_u32_le(idata_buf.data, (off + 4), ((ax_u32)(0)));
        ax_write_buf_u32_le(idata_buf.data, (off + 8), ((ax_u32)(0)));
        ax_write_buf_u32_le(idata_buf.data, (off + 12), ((ax_u32)(0)));
        ax_write_buf_u32_le(idata_buf.data, (off + 16), ((ax_u32)(0)));
        while ((merged_code.len < thunks_offset)) {
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0)));
        }
        ax_i64 thunk_i = ((ax_i64)(0));
        ax_i64 k_j = ((ax_i64)(0));
        while ((k_j < kernel32_imports.len)) {
            ax_string sym_name = ((kernel32_imports.data)[k_j]);
            ax_u32 th_rva = (thunks_rva + (((ax_u32)(thunk_i)) * ((ax_u32)(6))));
            ax_u32 iat_rva_sym = ((((ax_u32)(idata_rva)) + ((ax_u32)(kernel32_iat_offset))) + (((ax_u32)(k_j)) * ((ax_u32)(8))));
            ax_u32 disp32 = ((ax_u32)((((ax_i32)(iat_rva_sym)) - ((ax_i32)((th_rva + ((ax_u32)(6))))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0xFF)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0x25)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)((disp32 & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(8))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(16))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(24))) & ((ax_u32)(0xFF))))));
            ax_LinkerStrVec_push_linker_str(&(dyn_sym_names), sym_name);
            ((dyn_sym_rvas)[thunk_i]) = th_rva;
            thunk_i = (thunk_i + 1);
            k_j = (k_j + 1);
        }
        ax_i64 ax_j = ((ax_i64)(0));
        while ((ax_j < ax_runtime_imports.len)) {
            ax_string sym_name = ((ax_runtime_imports.data)[ax_j]);
            ax_u32 th_rva = (thunks_rva + (((ax_u32)(thunk_i)) * ((ax_u32)(6))));
            ax_u32 iat_rva_sym = ((((ax_u32)(idata_rva)) + ((ax_u32)(ax_runtime_iat_offset))) + (((ax_u32)(ax_j)) * ((ax_u32)(8))));
            ax_u32 disp32 = ((ax_u32)((((ax_i32)(iat_rva_sym)) - ((ax_i32)((th_rva + ((ax_u32)(6))))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0xFF)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0x25)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)((disp32 & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(8))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(16))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(24))) & ((ax_u32)(0xFF))))));
            ax_LinkerStrVec_push_linker_str(&(dyn_sym_names), sym_name);
            ((dyn_sym_rvas)[thunk_i]) = th_rva;
            thunk_i = (thunk_i + 1);
            ax_j = (ax_j + 1);
        }
        ax_i64 u_j = ((ax_i64)(0));
        while ((u_j < ucrtbase_imports.len)) {
            ax_string sym_name = ((ucrtbase_imports.data)[u_j]);
            ax_u32 th_rva = (thunks_rva + (((ax_u32)(thunk_i)) * ((ax_u32)(6))));
            ax_u32 iat_rva_sym = ((((ax_u32)(idata_rva)) + ((ax_u32)(ucrtbase_iat_offset))) + (((ax_u32)(u_j)) * ((ax_u32)(8))));
            ax_u32 disp32 = ((ax_u32)((((ax_i32)(iat_rva_sym)) - ((ax_i32)((th_rva + ((ax_u32)(6))))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0xFF)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0x25)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)((disp32 & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(8))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(16))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(24))) & ((ax_u32)(0xFF))))));
            ax_LinkerStrVec_push_linker_str(&(dyn_sym_names), sym_name);
            ((dyn_sym_rvas)[thunk_i]) = th_rva;
            thunk_i = (thunk_i + 1);
            u_j = (u_j + 1);
        }
    }
    struct ax_ByteVec dyn_buf = ax_new_byte_vec();
    ax_i64 offset_dynamic = ((ax_i64)(0));
    ax_i64 size_dynamic = ((ax_i64)(0));
    if ((ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"elf64", .len=5}) && (thunk_count > 0))) {
        dyn_sym_rvas = ((ax_u32*)(ax_alloc((thunk_count * 4))));
        struct ax_LinkerStrVec dyn_imports = ax_new_linker_str_vec();
        struct ax_LinkerStrVec dyn_libs = ax_new_linker_str_vec();
        ax_bool has_ax_lib = AX_FALSE;
        ax_bool has_c_lib = AX_FALSE;
        ax_i64 scan_j = ((ax_i64)(0));
        while ((scan_j < kernel32_imports.len)) {
            ax_LinkerStrVec_push_linker_str(&(dyn_imports), ((kernel32_imports.data)[scan_j]));
            has_ax_lib = AX_TRUE;
            scan_j = (scan_j + 1);
        }
        scan_j = 0;
        while ((scan_j < ax_runtime_imports.len)) {
            ax_LinkerStrVec_push_linker_str(&(dyn_imports), ((ax_runtime_imports.data)[scan_j]));
            has_ax_lib = AX_TRUE;
            scan_j = (scan_j + 1);
        }
        scan_j = 0;
        while ((scan_j < ucrtbase_imports.len)) {
            ax_LinkerStrVec_push_linker_str(&(dyn_imports), ((ucrtbase_imports.data)[scan_j]));
            has_c_lib = AX_TRUE;
            scan_j = (scan_j + 1);
        }
        if (has_ax_lib) {
            ax_LinkerStrVec_push_linker_str(&(dyn_libs), (ax_string){.ptr=(const ax_u8*)"libax_runtime.so", .len=16});
        }
        if (has_c_lib) {
            ax_LinkerStrVec_push_linker_str(&(dyn_libs), (ax_string){.ptr=(const ax_u8*)"libc.so.6", .len=9});
        }
        struct ax_ByteVec dynstr_buf = ax_new_byte_vec();
        ax_ByteVec_push_byte(&(dynstr_buf), ((ax_u8)(0)));
        ax_u64* lib_str_offsets = ((ax_u64*)(ax_alloc((dyn_libs.len * 8))));
        ax_i64 lib_i = ((ax_i64)(0));
        while ((lib_i < dyn_libs.len)) {
            ((lib_str_offsets)[lib_i]) = ((ax_u64)(dynstr_buf.len));
            ax_string name = ((dyn_libs.data)[lib_i]);
            ax_i64 name_len = ax_str_len(name);
            ax_u8* name_ptr = ((ax_u8*)(name.ptr));
            ax_i64 char_i = ((ax_i64)(0));
            while ((char_i < name_len)) {
                ax_ByteVec_push_byte(&(dynstr_buf), ((name_ptr)[char_i]));
                char_i = (char_i + 1);
            }
            ax_ByteVec_push_byte(&(dynstr_buf), ((ax_u8)(0)));
            lib_i = (lib_i + 1);
        }
        ax_u64* sym_str_offsets = ((ax_u64*)(ax_alloc((dyn_imports.len * 8))));
        ax_i64 sym_i = ((ax_i64)(0));
        while ((sym_i < dyn_imports.len)) {
            ((sym_str_offsets)[sym_i]) = ((ax_u64)(dynstr_buf.len));
            ax_string name = ((dyn_imports.data)[sym_i]);
            ax_i64 name_len = ax_str_len(name);
            ax_u8* name_ptr = ((ax_u8*)(name.ptr));
            ax_i64 char_i = ((ax_i64)(0));
            while ((char_i < name_len)) {
                ax_ByteVec_push_byte(&(dynstr_buf), ((name_ptr)[char_i]));
                char_i = (char_i + 1);
            }
            ax_ByteVec_push_byte(&(dynstr_buf), ((ax_u8)(0)));
            sym_i = (sym_i + 1);
        }
        struct ax_ByteVec dynsym_buf = ax_new_byte_vec();
        ax_i64 zero_idx = ((ax_i64)(0));
        while ((zero_idx < 24)) {
            ax_ByteVec_push_byte(&(dynsym_buf), ((ax_u8)(0)));
            zero_idx = (zero_idx + 1);
        }
        sym_i = 0;
        while ((sym_i < dyn_imports.len)) {
            ax_ByteVec_push_u32_le(&(dynsym_buf), ((ax_u32)(((sym_str_offsets)[sym_i]))));
            ax_ByteVec_push_byte(&(dynsym_buf), ((ax_u8)(0x12)));
            ax_ByteVec_push_byte(&(dynsym_buf), ((ax_u8)(0)));
            ax_ByteVec_push_u16_le(&(dynsym_buf), ((ax_u16)(0)));
            ax_ByteVec_push_u64_le(&(dynsym_buf), ((ax_u64)(0)));
            ax_ByteVec_push_u64_le(&(dynsym_buf), ((ax_u64)(0)));
            sym_i = (sym_i + 1);
        }
        struct ax_ByteVec hash_buf = ax_new_byte_vec();
        ax_ByteVec_push_u32_le(&(hash_buf), ((ax_u32)(1)));
        ax_ByteVec_push_u32_le(&(hash_buf), ((ax_u32)((((ax_i64)(1)) + dyn_imports.len))));
        ax_ByteVec_push_u32_le(&(hash_buf), ((ax_u32)(1)));
        ax_ByteVec_push_u32_le(&(hash_buf), ((ax_u32)(0)));
        ax_i64 chain_idx = ((ax_i64)(2));
        while ((chain_idx <= (((ax_i64)(1)) + dyn_imports.len))) {
            if ((chain_idx == (((ax_i64)(1)) + dyn_imports.len))) {
                ax_ByteVec_push_u32_le(&(hash_buf), ((ax_u32)(0)));
            } else {
                {
                    ax_ByteVec_push_u32_le(&(hash_buf), ((ax_u32)(chain_idx)));
                }
            }
            chain_idx = (chain_idx + 1);
        }
        struct ax_ByteVec got_plt_buf = ax_new_byte_vec();
        ax_ByteVec_push_u64_le(&(got_plt_buf), ((ax_u64)(0)));
        ax_ByteVec_push_u64_le(&(got_plt_buf), ((ax_u64)(0)));
        ax_ByteVec_push_u64_le(&(got_plt_buf), ((ax_u64)(0)));
        sym_i = 0;
        while ((sym_i < dyn_imports.len)) {
            ax_ByteVec_push_u64_le(&(got_plt_buf), ((ax_u64)(0)));
            sym_i = (sym_i + 1);
        }
        ax_i64 offset_dynstr = ((ax_i64)(0));
        ax_i64 size_dynstr = dynstr_buf.len;
        ax_i64 offset_dynsym = ((size_dynstr + 7) & (~7));
        ax_i64 size_dynsym = dynsym_buf.len;
        ax_i64 offset_hash = (((offset_dynsym + size_dynsym) + 7) & (~7));
        ax_i64 size_hash = hash_buf.len;
        ax_i64 offset_got_plt = (((offset_hash + size_hash) + 7) & (~7));
        ax_i64 size_got_plt = got_plt_buf.len;
        ax_i64 offset_rela_plt = (((offset_got_plt + size_got_plt) + 7) & (~7));
        ax_i64 size_rela_plt = (dyn_imports.len * 24);
        offset_dynamic = (((offset_rela_plt + size_rela_plt) + 7) & (~7));
        size_dynamic = ((dyn_libs.len + 14) * 16);
        ax_i64 dyn_offset = (((thunks_offset + thunks_size) + 15) & (~15));
        ax_u64 dyn_base_va = (base_addr + ((ax_u64)(dyn_offset)));
        ax_u64 dynstr_va = (dyn_base_va + ((ax_u64)(offset_dynstr)));
        ax_u64 dynsym_va = (dyn_base_va + ((ax_u64)(offset_dynsym)));
        ax_u64 hash_va = (dyn_base_va + ((ax_u64)(offset_hash)));
        ax_u64 got_plt_va = (dyn_base_va + ((ax_u64)(offset_got_plt)));
        ax_u64 rela_plt_va = (dyn_base_va + ((ax_u64)(offset_rela_plt)));
        ax_u64 dynamic_va = (dyn_base_va + ((ax_u64)(offset_dynamic)));
        struct ax_ByteVec rela_plt_buf = ax_new_byte_vec();
        sym_i = 0;
        while ((sym_i < dyn_imports.len)) {
            ax_u64 r_offset = ((got_plt_va + ((ax_u64)(24))) + (((ax_u64)(sym_i)) * ((ax_u64)(8))));
            ax_u64 r_info = (((((ax_u64)(1)) + ((ax_u64)(sym_i))) << ((ax_u64)(32))) | ((ax_u64)(7)));
            ax_ByteVec_push_u64_le(&(rela_plt_buf), r_offset);
            ax_ByteVec_push_u64_le(&(rela_plt_buf), r_info);
            ax_ByteVec_push_u64_le(&(rela_plt_buf), ((ax_u64)(0)));
            sym_i = (sym_i + 1);
        }
        struct ax_ByteVec dynamic_buf = ax_new_byte_vec();
        ax_i64 lib_idx = ((ax_i64)(0));
        while ((lib_idx < dyn_libs.len)) {
            ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(1)));
            ax_ByteVec_push_u64_le(&(dynamic_buf), ((lib_str_offsets)[lib_idx]));
            lib_idx = (lib_idx + 1);
        }
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(2)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(rela_plt_buf.len)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(20)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(7)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(3)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), got_plt_va);
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(4)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), hash_va);
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(5)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), dynstr_va);
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(6)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), dynsym_va);
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(10)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(dynstr_buf.len)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(11)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(24)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(7)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), rela_plt_va);
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(8)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(rela_plt_buf.len)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(9)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(24)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(24)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(1)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(30)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(8)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(0)));
        ax_ByteVec_push_u64_le(&(dynamic_buf), ((ax_u64)(0)));
        ax_ByteVec_push_bytes(&(dyn_buf), dynstr_buf.data, dynstr_buf.len);
        while ((dyn_buf.len < offset_dynsym)) {
            ax_ByteVec_push_byte(&(dyn_buf), ((ax_u8)(0)));
        }
        ax_ByteVec_push_bytes(&(dyn_buf), dynsym_buf.data, dynsym_buf.len);
        while ((dyn_buf.len < offset_hash)) {
            ax_ByteVec_push_byte(&(dyn_buf), ((ax_u8)(0)));
        }
        ax_ByteVec_push_bytes(&(dyn_buf), hash_buf.data, hash_buf.len);
        while ((dyn_buf.len < offset_got_plt)) {
            ax_ByteVec_push_byte(&(dyn_buf), ((ax_u8)(0)));
        }
        ax_ByteVec_push_bytes(&(dyn_buf), got_plt_buf.data, got_plt_buf.len);
        while ((dyn_buf.len < offset_rela_plt)) {
            ax_ByteVec_push_byte(&(dyn_buf), ((ax_u8)(0)));
        }
        ax_ByteVec_push_bytes(&(dyn_buf), rela_plt_buf.data, rela_plt_buf.len);
        while ((dyn_buf.len < offset_dynamic)) {
            ax_ByteVec_push_byte(&(dyn_buf), ((ax_u8)(0)));
        }
        ax_ByteVec_push_bytes(&(dyn_buf), dynamic_buf.data, dynamic_buf.len);
        while ((merged_code.len < thunks_offset)) {
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0)));
        }
        ax_i64 thunk_i = ((ax_i64)(0));
        ax_i64 k_j = ((ax_i64)(0));
        while ((k_j < kernel32_imports.len)) {
            ax_string sym_name = ((kernel32_imports.data)[k_j]);
            ax_u64 th_va = ((base_addr + ((ax_u64)(thunks_offset))) + (((ax_u64)(thunk_i)) * ((ax_u64)(6))));
            ax_u64 got_slot_va = ((got_plt_va + ((ax_u64)(24))) + (((ax_u64)(thunk_i)) * ((ax_u64)(8))));
            ax_u32 disp32 = ((ax_u32)((got_slot_va - (th_va + ((ax_u64)(6))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0xFF)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0x25)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)((disp32 & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(8))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(16))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(24))) & ((ax_u32)(0xFF))))));
            ax_LinkerStrVec_push_linker_str(&(dyn_sym_names), sym_name);
            ((dyn_sym_rvas)[thunk_i]) = ((((ax_u32)(thunks_offset)) + (((ax_u32)(thunk_i)) * ((ax_u32)(6)))) + ((ax_u32)(0x1000)));
            thunk_i = (thunk_i + 1);
            k_j = (k_j + 1);
        }
        ax_i64 ax_j = ((ax_i64)(0));
        while ((ax_j < ax_runtime_imports.len)) {
            ax_string sym_name = ((ax_runtime_imports.data)[ax_j]);
            ax_u64 th_va = ((base_addr + ((ax_u64)(thunks_offset))) + (((ax_u64)(thunk_i)) * ((ax_u64)(6))));
            ax_u64 got_slot_va = ((got_plt_va + ((ax_u64)(24))) + (((ax_u64)(thunk_i)) * ((ax_u64)(8))));
            ax_u32 disp32 = ((ax_u32)((got_slot_va - (th_va + ((ax_u64)(6))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0xFF)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0x25)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)((disp32 & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(8))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(16))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(24))) & ((ax_u32)(0xFF))))));
            ax_LinkerStrVec_push_linker_str(&(dyn_sym_names), sym_name);
            ((dyn_sym_rvas)[thunk_i]) = ((((ax_u32)(thunks_offset)) + (((ax_u32)(thunk_i)) * ((ax_u32)(6)))) + ((ax_u32)(0x1000)));
            thunk_i = (thunk_i + 1);
            ax_j = (ax_j + 1);
        }
        ax_i64 u_j = ((ax_i64)(0));
        while ((u_j < ucrtbase_imports.len)) {
            ax_string sym_name = ((ucrtbase_imports.data)[u_j]);
            ax_u64 th_va = ((base_addr + ((ax_u64)(thunks_offset))) + (((ax_u64)(thunk_i)) * ((ax_u64)(6))));
            ax_u64 got_slot_va = ((got_plt_va + ((ax_u64)(24))) + (((ax_u64)(thunk_i)) * ((ax_u64)(8))));
            ax_u32 disp32 = ((ax_u32)((got_slot_va - (th_va + ((ax_u64)(6))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0xFF)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(0x25)));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)((disp32 & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(8))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(16))) & ((ax_u32)(0xFF))))));
            ax_ByteVec_push_byte(&(merged_code), ((ax_u8)(((disp32 >> ((ax_u32)(24))) & ((ax_u32)(0xFF))))));
            ax_LinkerStrVec_push_linker_str(&(dyn_sym_names), sym_name);
            ((dyn_sym_rvas)[thunk_i]) = ((((ax_u32)(thunks_offset)) + (((ax_u32)(thunk_i)) * ((ax_u32)(6)))) + ((ax_u32)(0x1000)));
            thunk_i = (thunk_i + 1);
            u_j = (u_j + 1);
        }
        if ((dynstr_buf.data != ((ax_u8*)(NULL)))) {
            ax_free(dynstr_buf.data);
        }
        if ((dynsym_buf.data != ((ax_u8*)(NULL)))) {
            ax_free(dynsym_buf.data);
        }
        if ((hash_buf.data != ((ax_u8*)(NULL)))) {
            ax_free(hash_buf.data);
        }
        if ((got_plt_buf.data != ((ax_u8*)(NULL)))) {
            ax_free(got_plt_buf.data);
        }
        if ((rela_plt_buf.data != ((ax_u8*)(NULL)))) {
            ax_free(rela_plt_buf.data);
        }
        if ((dynamic_buf.data != ((ax_u8*)(NULL)))) {
            ax_free(dynamic_buf.data);
        }
        ax_free(((ax_u8*)(lib_str_offsets)));
        ax_free(((ax_u8*)(sym_str_offsets)));
        ax_free(((ax_u8*)(dyn_imports.data)));
        ax_free(((ax_u8*)(dyn_libs.data)));
    }
    i = 0;
    while ((i < objects.len)) {
        struct ax_ParsedObject* obj = ((objects.data)[i]);
        ax_i64 r_i = ((ax_i64)(0));
        while ((r_i < obj->relocs.len)) {
            struct ax_ParsedReloc* r = &(((obj->relocs.data)[r_i]));
            if ((r->sym_idx >= ((ax_u32)(obj->sym_names.len)))) {
                r_i = (r_i + 1);
                continue;
            }
            ax_string target_name = ((obj->sym_names.data)[r->sym_idx]);
            ax_i64 target_offset = (-((ax_i64)(1)));
            ax_i64 search = ((ax_i64)(0));
            while ((search < func_names.len)) {
                if (ax_str_eq(((func_names.data)[search]), target_name)) {
                    target_offset = ((func_offsets.data)[search]).offset;
                    break;
                }
                search = (search + 1);
            }
            ax_u64 target_va = ((ax_u64)(0));
            if ((target_offset != (-((ax_i64)(1))))) {
                target_va = (base_addr + ((ax_u64)(target_offset)));
            } else {
                {
                    ax_i64 dyn_search = ((ax_i64)(0));
                    ax_bool found_dyn = AX_FALSE;
                    while ((dyn_search < dyn_sym_names.len)) {
                        if (ax_str_eq(((dyn_sym_names.data)[dyn_search]), target_name)) {
                            target_va = (base_addr + ((ax_u64)((((dyn_sym_rvas)[dyn_search]) - ((ax_u32)(0x1000))))));
                            found_dyn = AX_TRUE;
                            break;
                        }
                        dyn_search = (dyn_search + 1);
                    }
                    if ((!found_dyn)) {
                        target_va = ((ax_u64)(0x500000));
                    }
                }
            }
            ax_i64 rel_offset = (((ax_i64)((obj->va - base_addr))) + r->offset);
            if (((rel_offset + ((ax_i64)(4))) > merged_code.len)) {
                r_i = (r_i + 1);
                continue;
            }
            if (r->is_pc) {
                ax_u64 pc = (base_addr + ((ax_u64)(rel_offset)));
                ax_u32 val = ((ax_u32)(((((ax_i64)(target_va)) - ((ax_i64)(pc))) + r->addend)));
                ax_write_u32_le(merged_code.data, rel_offset, val);
            } else {
                {
                    if (((rel_offset + ((ax_i64)(8))) <= merged_code.len)) {
                        ax_u64 val = (target_va + ((ax_u64)(r->addend)));
                        ax_write_u64_le(merged_code.data, rel_offset, val);
                    }
                }
            }
            r_i = (r_i + 1);
        }
        i = (i + 1);
    }
    ax_u32 entry_rva = ((ax_u32)(0x1000));
    ax_i64 entry_search = ((ax_i64)(0));
    while ((entry_search < func_names.len)) {
        ax_string fn_name = ((func_names.data)[entry_search]);
        if ((ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"main", .len=4}) || ax_str_eq(fn_name, (ax_string){.ptr=(const ax_u8*)"_AX_main_main_v_v", .len=17}))) {
            entry_rva = ((ax_u32)((((ax_i64)(0x1000)) + ((func_offsets.data)[entry_search]).offset)));
            break;
        }
        entry_search = (entry_search + 1);
    }
    ax_u8* output_path_nt = ax_x86_str_to_null_terminated(self->output_path);
    void* file = fopen(((void*)(output_path_nt)), (const char*)((ax_string){.ptr=(const ax_u8*)"wb", .len=2}).ptr);
    ax_free(output_path_nt);
    if ((file == ((void*)(NULL)))) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: Linker could not write output file", .len=41}).ptr);
        return AX_FALSE;
    }
    if (ax_str_eq(format, (ax_string){.ptr=(const ax_u8*)"coff", .len=4})) {
        ax_i64 orig_code_len = merged_code.len;
        ax_ByteVec_align_byte_vec(&(merged_code), ((ax_i64)(0x200)));
        ax_u32 code_raw_size = ((ax_u32)(merged_code.len));
        ax_u32 idata_raw_size = (((ax_u32)((((ax_i64)(idata_buf.len)) + 0x1FF))) & ((ax_u32)(0xFFFFFE00)));
        ax_ByteVec_align_byte_vec(&(idata_buf), ((ax_i64)(0x200)));
        ax_u32 iat_rva = (((ax_u32)(idata_rva)) + ((ax_u32)(iat_offset)));
        ax_u32 iat_size = ((ax_u32)(total_iat_size));
        struct ax_ByteVec pe_headers = ax_new_byte_vec();
        ax_ByteVec_linker_build_pe_headers(&(pe_headers), code_raw_size, idata_raw_size, entry_rva, ((ax_u32)(idata_rva)), ((ax_u32)(idata_buf.len)), iat_rva, iat_size);
        fwrite(((void*)(pe_headers.data)), 1, pe_headers.len, file);
        fwrite(((void*)(merged_code.data)), 1, merged_code.len, file);
        if ((idata_buf.len > 0)) {
            fwrite(((void*)(idata_buf.data)), 1, idata_buf.len, file);
        }
        ax_free(pe_headers.data);
    } else {
        {
            ax_u64 entry_va = (base_addr + ((ax_u64)((entry_rva - ((ax_u32)(0x1000))))));
            struct ax_ByteVec elf_headers = ax_new_byte_vec();
            ax_u32 dynamic_offset = ((ax_u32)(0));
            ax_u32 dynamic_size = ((ax_u32)(0));
            if ((dyn_buf.len > 0)) {
                dynamic_offset = (((ax_u32)(512)) + ((ax_u32)(offset_dynamic)));
                dynamic_size = ((ax_u32)(size_dynamic));
            }
            ax_ByteVec_linker_build_elf_headers(&(elf_headers), ((ax_u32)((merged_code.len + dyn_buf.len))), entry_va, dynamic_offset, dynamic_size);
            fwrite(((void*)(elf_headers.data)), 1, elf_headers.len, file);
            fwrite(((void*)(merged_code.data)), 1, merged_code.len, file);
            if ((dyn_buf.len > 0)) {
                fwrite(((void*)(dyn_buf.data)), 1, dyn_buf.len, file);
            }
            ax_free(elf_headers.data);
        }
    }
    fclose(file);
    if ((merged_code.data != ((ax_u8*)(NULL)))) {
        ax_free(merged_code.data);
    }
    if ((idata_buf.data != ((ax_u8*)(NULL)))) {
        ax_free(idata_buf.data);
    }
    if ((dyn_buf.data != ((ax_u8*)(NULL)))) {
        ax_free(dyn_buf.data);
    }
    if ((dyn_sym_rvas != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(dyn_sym_rvas)));
    }
    if ((kernel32_hn_rvas != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(kernel32_hn_rvas)));
    }
    if ((ax_runtime_hn_rvas != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(ax_runtime_hn_rvas)));
    }
    if ((ucrtbase_hn_rvas != ((ax_u32*)(NULL)))) {
        ax_free(((ax_u8*)(ucrtbase_hn_rvas)));
    }
    if ((kernel32_imports.data != ((ax_string*)(NULL)))) {
        ax_free(((ax_u8*)(kernel32_imports.data)));
    }
    if ((ax_runtime_imports.data != ((ax_string*)(NULL)))) {
        ax_free(((ax_u8*)(ax_runtime_imports.data)));
    }
    if ((ucrtbase_imports.data != ((ax_string*)(NULL)))) {
        ax_free(((ax_u8*)(ucrtbase_imports.data)));
    }
    if ((dyn_sym_names.data != ((ax_string*)(NULL)))) {
        ax_free(((ax_u8*)(dyn_sym_names.data)));
    }
    i = 0;
    while ((i < objects.len)) {
        struct ax_ParsedObject* obj = ((objects.data)[i]);
        if ((obj->text.data != ((ax_u8*)(NULL)))) {
            ax_free(obj->text.data);
        }
        if ((obj->rdata.data != ((ax_u8*)(NULL)))) {
            ax_free(obj->rdata.data);
        }
        if ((obj->symbols.data != ((struct ax_LinkerSymbol*)(NULL)))) {
            ax_free(((ax_u8*)(obj->symbols.data)));
        }
        if ((obj->sym_names.data != ((ax_string*)(NULL)))) {
            ax_free(((ax_u8*)(obj->sym_names.data)));
        }
        if ((obj->relocs.data != ((struct ax_ParsedReloc*)(NULL)))) {
            ax_free(((ax_u8*)(obj->relocs.data)));
        }
        ax_free(((ax_u8*)(obj)));
        i = (i + 1);
    }
    ax_free(((ax_u8*)(objects.data)));
    ax_free(((ax_u8*)(func_names.data)));
    ax_free(((ax_u8*)(func_offsets.data)));
    return AX_TRUE;
}

static void ax_AirInst_print_inst(struct ax_AirInst inst) {
    ax_u16 op = inst.opcode;
    if ((op == ax_OP_NOP)) {
        return;
    }
    if ((op == ax_OP_RETURN)) {
        if ((inst.src1 != ((ax_u32)(0)))) {
            printf((const char*)((ax_string){.ptr=(const ax_u8*)"    ret %%%d\n", .len=13}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        } else {
            {
                puts((const char*)((ax_string){.ptr=(const ax_u8*)"    ret", .len=7}).ptr);
            }
        }
    } else if ((op == ax_OP_JUMP)) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"    jump block_%d\n", .len=18}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else if ((op == ax_OP_BRANCH)) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"    branch %%%d block_%d block_%d\n", .len=34}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(inst.dest)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    } else {
        {
            ax_string mnemonic = ax_opcode_mnemonic(op);
            ax_bool has_dest = (inst.dest != ((ax_u32)(0)));
            ax_bool has_src1 = (inst.src1 != ((ax_u32)(0)));
            ax_bool has_src2 = (inst.src2 != ((ax_u32)(0)));
            ax_bool is_binary = ax_opcode_is_binary_alu(op);
            if (has_dest) {
                ax_u16 type_id = inst.type_id;
                if ((((((op == ax_OP_CALL) || (op == ax_OP_INDEX)) || (op == ax_OP_DEREF)) || (op == ax_OP_CAST)) || (op == ax_OP_AWAIT))) {
                    type_id = ((ax_u16)(0));
                }
                if ((type_id != ((ax_u16)(0)))) {
                    printf((const char*)((ax_string){.ptr=(const ax_u8*)"    %%%d: t%d = %s", .len=18}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(type_id)), ((ax_i64)(((ax_u8*)(mnemonic.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                } else {
                    {
                        printf((const char*)((ax_string){.ptr=(const ax_u8*)"    %%%d = %s", .len=13}).ptr, ((ax_i64)(inst.dest)), ((ax_i64)(((ax_u8*)(mnemonic.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                    }
                }
            } else {
                {
                    printf((const char*)((ax_string){.ptr=(const ax_u8*)"    %s", .len=6}).ptr, ((ax_i64)(((ax_u8*)(mnemonic.ptr)))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
                }
            }
            if (((is_binary && has_src1) && has_src2)) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)" %%%d, %%%d\n", .len=12}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if ((has_src1 && has_src2)) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)" %%%d, %%%d\n", .len=12}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if (has_src1) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)" %%%d\n", .len=6}).ptr, ((ax_i64)(inst.src1)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else if (has_src2) {
                printf((const char*)((ax_string){.ptr=(const ax_u8*)" %%%d\n", .len=6}).ptr, ((ax_i64)(inst.src2)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
            } else {
                {
                    puts((const char*)((ax_string){.ptr=(const ax_u8*)"", .len=0}).ptr);
                }
            }
        }
    }
}

static void ax_AirFunc_print_block(struct ax_AirFunc f, struct ax_BasicBlock* bb) {
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"  block_%d:", .len=11}).ptr, ((ax_i64)(bb->id)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    if ((bb->is_entry && bb->is_exit)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"  ; entry exit", .len=14}).ptr);
    } else if (bb->is_entry) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"  ; entry", .len=9}).ptr);
    } else if (bb->is_exit) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"  ; exit", .len=8}).ptr);
    } else {
        {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"", .len=0}).ptr);
        }
    }
    ax_i64 i = ((ax_i64)(0));
    while ((i < ((ax_i64)(bb->instrs_len)))) {
        ax_u32 inst_idx = ((f.block_instrs.data)[(((ax_i64)(bb->instrs_start)) + i)]);
        if ((inst_idx < ((ax_u32)(f.insts.len)))) {
            struct ax_AirInst inst = ((f.insts.data)[((ax_i64)(inst_idx))]);
            ax_AirInst_print_inst(inst);
        }
        i = (i + 1);
    }
}

static void ax_AirFunc_print_func(struct ax_AirFunc f) {
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"fn @%d(", .len=7}).ptr, ((ax_i64)(f.name)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    ax_i64 i = ((ax_i64)(0));
    while ((i < f.params.len)) {
        if ((i > 0)) {
            printf((const char*)((ax_string){.ptr=(const ax_u8*)", ", .len=2}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        }
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"t%d", .len=3}).ptr, ((ax_i64)(((f.params.data)[i]))), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        i = (i + 1);
    }
    printf((const char*)((ax_string){.ptr=(const ax_u8*)")", .len=1}).ptr, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    if ((f.ret_type != ((ax_u32)(0)))) {
        printf((const char*)((ax_string){.ptr=(const ax_u8*)" -> t%d", .len=7}).ptr, ((ax_i64)(f.ret_type)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    }
    puts((const char*)((ax_string){.ptr=(const ax_u8*)":", .len=1}).ptr);
    ax_i64 bi = ((ax_i64)(0));
    while ((bi < f.blocks.len)) {
        struct ax_BasicBlock bb = ((f.blocks.data)[bi]);
        ax_AirFunc_print_block(f, &(bb));
        bi = (bi + 1);
    }
}

static void ax_AirModule_print_module(struct ax_AirModule mod) {
    ax_i64 i = ((ax_i64)(0));
    while ((i < mod.funcs.len)) {
        if ((i > 0)) {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"", .len=0}).ptr);
        }
        ax_AirFunc_print_func(((mod.funcs.data)[i]));
        i = (i + 1);
    }
}

static ax_string ax_read_file_content(ax_string path) {
    void* file = fopen(((void*)(path.ptr)), (const char*)((ax_string){.ptr=(const ax_u8*)"rb", .len=2}).ptr);
    if ((file == ((void*)(NULL)))) {
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    fseek(file, 0, 2);
    ax_i64 size = ftell(file);
    rewind(file);
    if ((size <= 0)) {
        fclose(file);
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    ax_u8* buffer = ((ax_u8*)(ax_alloc((size + 1))));
    if ((buffer == ((ax_u8*)(NULL)))) {
        fclose(file);
        return (ax_string){.ptr=(const ax_u8*)"", .len=0};
    }
    ax_i64 bytes_read = fread(((void*)(buffer)), 1, size, file);
    fclose(file);
    ((buffer)[bytes_read]) = ((ax_u8)(0));
    return ((ax_string){.ptr = (const ax_u8*)(buffer), .len = strlen((const char*)(buffer))});
}

static ax_bool ax_match_prefix(ax_string s, ax_i64 i, ax_string prefix) {
    ax_u8* s_ptr = ((ax_u8*)(s.ptr));
    ax_u8* p_ptr = ((ax_u8*)(prefix.ptr));
    ax_i64 s_len = ax_str_len(s);
    ax_i64 p_len = ax_str_len(prefix);
    if (((i + p_len) > s_len)) {
        return AX_FALSE;
    }
    ax_i64 j = ((ax_i64)(0));
    while ((j < p_len)) {
        if ((((s_ptr)[(i + j)]) != ((p_ptr)[j]))) {
            return AX_FALSE;
        }
        j = (j + ((ax_i64)(1)));
    }
    return AX_TRUE;
}

static ax_string ax_strip_imports(ax_string s) {
    ax_u8* s_ptr = ((ax_u8*)(s.ptr));
    ax_i64 s_len = ax_str_len(s);
    ax_u8* out_buf = ((ax_u8*)(ax_alloc((s_len + ((ax_i64)(1))))));
    ax_i64 out_idx = ((ax_i64)(0));
    ax_i64 i = ((ax_i64)(0));
    while ((i < s_len)) {
        ax_bool is_line_start = AX_FALSE;
        if ((i == ((ax_i64)(0)))) {
            is_line_start = AX_TRUE;
        } else if ((((s_ptr)[(i - ((ax_i64)(1)))]) == ((ax_u8)(10)))) {
            is_line_start = AX_TRUE;
        }
        ax_bool skip_line = AX_FALSE;
        if (is_line_start) {
            if (ax_match_prefix(s, i, (ax_string){.ptr=(const ax_u8*)"import std.mem.alloc", .len=20})) {
                skip_line = AX_TRUE;
            } else if (ax_match_prefix(s, i, (ax_string){.ptr=(const ax_u8*)"import std.scheduler", .len=20})) {
                skip_line = AX_TRUE;
            } else if (ax_match_prefix(s, i, (ax_string){.ptr=(const ax_u8*)"import std.runtime", .len=18})) {
                skip_line = AX_TRUE;
            }
        }
        if (skip_line) {
            while (((i < s_len) && (((s_ptr)[i]) != ((ax_u8)(10))))) {
                i = (i + ((ax_i64)(1)));
            }
            if (((i < s_len) && (((s_ptr)[i]) == ((ax_u8)(10))))) {
                i = (i + ((ax_i64)(1)));
            }
        } else {
            {
                ((out_buf)[out_idx]) = ((s_ptr)[i]);
                out_idx = (out_idx + ((ax_i64)(1)));
                i = (i + ((ax_i64)(1)));
            }
        }
    }
    ((out_buf)[out_idx]) = ((ax_u8)(0));
    return ((ax_string){.ptr = (const ax_u8*)(out_buf), .len = strlen((const char*)(out_buf))});
}

ax_i32 ax_main_usr(ax_i32 argc, ax_u8** argv) {
    if ((argc < 2)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"Usage: axc <command> [args]", .len=27}).ptr);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"Commands:", .len=9}).ptr);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"  dump-air <file.ax>", .len=20}).ptr);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"  emit-c <file.ax> [-o <out.c>]", .len=31}).ptr);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"  build <file.ax> [-o <out.exe>] [--target <wasm32|wasm32-unknown-unknown>]", .len=75}).ptr);
        return 1;
    }
    ax_bool is_dump_air = AX_FALSE;
    ax_bool is_emit_c = AX_FALSE;
    ax_bool is_build = AX_FALSE;
    ax_i32 file_arg_idx = 1;
    ax_string output_file = (ax_string){.ptr=(const ax_u8*)"", .len=0};
    ax_bool target_wasm32 = AX_FALSE;
    ax_bool self_link = AX_FALSE;
    ax_bool optimize = AX_FALSE;
    ax_string arg1 = ((ax_string){.ptr = (const ax_u8*)(((argv)[1])), .len = strlen((const char*)(((argv)[1])))});
    if (ax_str_eq(arg1, (ax_string){.ptr=(const ax_u8*)"dump-air", .len=8})) {
        if ((argc < 3)) {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: missing filename for dump-air", .len=36}).ptr);
            return 1;
        }
        is_dump_air = AX_TRUE;
        file_arg_idx = 2;
    } else if (ax_str_eq(arg1, (ax_string){.ptr=(const ax_u8*)"emit-c", .len=6})) {
        if ((argc < 3)) {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: missing filename for emit-c", .len=34}).ptr);
            return 1;
        }
        is_emit_c = AX_TRUE;
        file_arg_idx = 2;
    } else if (ax_str_eq(arg1, (ax_string){.ptr=(const ax_u8*)"build", .len=5})) {
        if ((argc < 3)) {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: missing filename for build", .len=33}).ptr);
            return 1;
        }
        is_build = AX_TRUE;
        file_arg_idx = 2;
    } else {
        {
            is_dump_air = AX_TRUE;
            file_arg_idx = 1;
        }
    }
    ax_i32 opt_idx = (file_arg_idx + 1);
    while ((opt_idx < argc)) {
        ax_string opt = ((ax_string){.ptr = (const ax_u8*)(((argv)[opt_idx])), .len = strlen((const char*)(((argv)[opt_idx])))});
        if ((ax_str_eq(opt, (ax_string){.ptr=(const ax_u8*)"-o", .len=2}) || ax_str_eq(opt, (ax_string){.ptr=(const ax_u8*)"--output", .len=8}))) {
            if (((opt_idx + 1) < argc)) {
                output_file = ((ax_string){.ptr = (const ax_u8*)(((argv)[(opt_idx + 1)])), .len = strlen((const char*)(((argv)[(opt_idx + 1)])))});
                opt_idx = (opt_idx + 1);
            } else {
                {
                    puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: missing value for -o option", .len=34}).ptr);
                    return 1;
                }
            }
        } else if ((ax_str_eq(opt, (ax_string){.ptr=(const ax_u8*)"--target", .len=8}) || ax_str_eq(opt, (ax_string){.ptr=(const ax_u8*)"-target", .len=7}))) {
            if (((opt_idx + 1) < argc)) {
                ax_string target_triple = ((ax_string){.ptr = (const ax_u8*)(((argv)[(opt_idx + 1)])), .len = strlen((const char*)(((argv)[(opt_idx + 1)])))});
                if (((ax_str_eq(target_triple, (ax_string){.ptr=(const ax_u8*)"wasm32", .len=6}) || ax_str_eq(target_triple, (ax_string){.ptr=(const ax_u8*)"wasm32-unknown-unknown", .len=22})) || ax_str_eq(target_triple, (ax_string){.ptr=(const ax_u8*)"wasm", .len=4}))) {
                    target_wasm32 = AX_TRUE;
                }
                opt_idx = (opt_idx + 1);
            } else {
                {
                    puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: missing value for --target option", .len=40}).ptr);
                    return 1;
                }
            }
        } else if ((ax_str_eq(opt, (ax_string){.ptr=(const ax_u8*)"-self-link", .len=10}) || ax_str_eq(opt, (ax_string){.ptr=(const ax_u8*)"--self-link", .len=11}))) {
            self_link = AX_TRUE;
        } else if ((ax_str_eq(opt, (ax_string){.ptr=(const ax_u8*)"-O1", .len=3}) || ax_str_eq(opt, (ax_string){.ptr=(const ax_u8*)"--opt", .len=5}))) {
            optimize = AX_TRUE;
        }
        opt_idx = (opt_idx + 1);
    }
    ax_u8* filename_ptr = ((argv)[file_arg_idx]);
    void* file = fopen(((void*)(filename_ptr)), (const char*)((ax_string){.ptr=(const ax_u8*)"rb", .len=2}).ptr);
    if ((file == ((void*)(NULL)))) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: could not open file", .len=26}).ptr);
        return 1;
    }
    fseek(file, 0, 2);
    ax_i64 size = ftell(file);
    rewind(file);
    if ((size < 0)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: could not get file size", .len=30}).ptr);
        fclose(file);
        return 1;
    }
    ax_u8* buffer = ((ax_u8*)(ax_alloc((size + 1))));
    if ((buffer == ((ax_u8*)(NULL)))) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: memory allocation failed", .len=31}).ptr);
        fclose(file);
        return 1;
    }
    ax_i64 bytes_read = fread(((void*)(buffer)), 1, size, file);
    fclose(file);
    ((buffer)[bytes_read]) = ((ax_u8)(0));
    ax_string src = ((ax_string){.ptr = (const ax_u8*)(buffer), .len = strlen((const char*)(buffer))});
    if (self_link) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Reading alloc.ax...", .len=27}).ptr);
        fflush(((void*)(NULL)));
        ax_string raw_alloc = ax_read_file_content((ax_string){.ptr=(const ax_u8*)"std/mem/alloc.ax", .len=16});
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Reading scheduler.ax...", .len=31}).ptr);
        fflush(((void*)(NULL)));
        ax_string raw_sched = ax_read_file_content((ax_string){.ptr=(const ax_u8*)"std/scheduler.ax", .len=16});
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Reading runtime.ax...", .len=29}).ptr);
        fflush(((void*)(NULL)));
        ax_string raw_rt = ax_read_file_content((ax_string){.ptr=(const ax_u8*)"std/runtime.ax", .len=14});
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stripping imports from alloc.ax...", .len=42}).ptr);
        fflush(((void*)(NULL)));
        ax_string alloc_src = ax_strip_imports(raw_alloc);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stripping imports from scheduler.ax...", .len=46}).ptr);
        fflush(((void*)(NULL)));
        ax_string sched_src = ax_strip_imports(raw_sched);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stripping imports from runtime.ax...", .len=44}).ptr);
        fflush(((void*)(NULL)));
        ax_string rt_src = ax_strip_imports(raw_rt);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stripping imports from src...", .len=37}).ptr);
        fflush(((void*)(NULL)));
        ax_string clean_src = ax_strip_imports(src);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Concatenating buffers...", .len=32}).ptr);
        fflush(((void*)(NULL)));
        ax_i64 len_src = ax_str_len(clean_src);
        ax_i64 len_alloc = ax_str_len(alloc_src);
        ax_i64 len_sched = ax_str_len(sched_src);
        ax_i64 len_rt = ax_str_len(rt_src);
        ax_i64 total_len = ((((((len_alloc + ((ax_i64)(1))) + len_sched) + ((ax_i64)(1))) + len_rt) + ((ax_i64)(1))) + len_src);
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] len_alloc=%d, len_sched=%d, len_rt=%d, len_src=%d, total_len=%d\n", .len=72}).ptr, len_alloc, len_sched, len_rt, len_src, total_len, ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        fflush(((void*)(NULL)));
        ax_u8* new_buf = ((ax_u8*)(ax_alloc((total_len + ((ax_i64)(1))))));
        memcpy(new_buf, ((ax_u8*)(alloc_src.ptr)), len_alloc);
        ((new_buf)[len_alloc]) = ((ax_u8)(10));
        ax_i64 off_sched = (len_alloc + ((ax_i64)(1)));
        memcpy(((ax_u8*)((((ax_i64)(new_buf)) + off_sched))), ((ax_u8*)(sched_src.ptr)), len_sched);
        ((new_buf)[(off_sched + len_sched)]) = ((ax_u8)(10));
        ax_i64 off_rt = ((off_sched + len_sched) + ((ax_i64)(1)));
        memcpy(((ax_u8*)((((ax_i64)(new_buf)) + off_rt))), ((ax_u8*)(rt_src.ptr)), len_rt);
        ((new_buf)[(off_rt + len_rt)]) = ((ax_u8)(10));
        ax_i64 off_src = ((off_rt + len_rt) + ((ax_i64)(1)));
        memcpy(((ax_u8*)((((ax_i64)(new_buf)) + off_src))), ((ax_u8*)(clean_src.ptr)), len_src);
        ((new_buf)[total_len]) = ((ax_u8)(0));
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Cleaning up...", .len=22}).ptr);
        fflush(((void*)(NULL)));
        if ((ax_str_len(raw_alloc) > ((ax_i64)(0)))) {
            ax_free(((ax_u8*)(raw_alloc.ptr)));
        }
        if ((ax_str_len(raw_sched) > ((ax_i64)(0)))) {
            ax_free(((ax_u8*)(raw_sched.ptr)));
        }
        if ((ax_str_len(raw_rt) > ((ax_i64)(0)))) {
            ax_free(((ax_u8*)(raw_rt.ptr)));
        }
        ax_free(((ax_u8*)(alloc_src.ptr)));
        ax_free(((ax_u8*)(sched_src.ptr)));
        ax_free(((ax_u8*)(rt_src.ptr)));
        ax_free(((ax_u8*)(clean_src.ptr)));
        src = ((ax_string){.ptr = (const ax_u8*)(new_buf), .len = strlen((const char*)(new_buf))});
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Concatenation finished successfully!", .len=44}).ptr);
        fflush(((void*)(NULL)));
        void* out_file = fopen(((void*)((ax_string){.ptr=(const ax_u8*)"scratch/self_linked_concatenated.ax", .len=35}.ptr)), (const char*)((ax_string){.ptr=(const ax_u8*)"wb", .len=2}).ptr);
        if ((out_file != ((void*)(NULL)))) {
            fputs((const char*)(src).ptr, out_file);
            fclose(out_file);
        }
    }
    if ((!is_dump_air)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stage 1: Starting compilation...", .len=40}).ptr);
        fflush(((void*)(NULL)));
    }
    struct ax_Lexer lexer = ax_new_lexer(src);
    struct ax_TokenVec tokens = ax_Lexer_tokenize(&(lexer));
    if ((!is_dump_air)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Finished Lexing.", .len=24}).ptr);
        fflush(((void*)(NULL)));
    }
    struct ax_InternPool pool = ax_new_intern_pool();
    struct ax_Parser parser = ax_TokenVec_new_parser(tokens, src, pool);
    ax_Parser_parse_program(&(parser));
    if ((!is_dump_air)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Finished Parsing.", .len=25}).ptr);
        fflush(((void*)(NULL)));
    }
    struct ax_SymbolTable symtable = ax_InternPool_new_symbol_table(&(parser.pool));
    struct ax_NameResolver resolver = ax_AstTree_new_name_resolver(parser.tree, parser.pool, symtable);
    ax_NameResolver_resolve(&(resolver));
    if ((!is_dump_air)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Finished Resolving.", .len=27}).ptr);
        fflush(((void*)(NULL)));
    }
    struct ax_TypeTable typetable = ax_new_type_table();
    struct ax_TypeChecker checker = ax_AstTree_new_type_checker(parser.tree, parser.pool, resolver.symtable, typetable);
    ax_TypeChecker_run_type_checker(&(checker));
    typetable = checker.types;
    if ((!is_dump_air)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Finished Typechecking.", .len=30}).ptr);
        fflush(((void*)(NULL)));
    }
    if ((!is_dump_air)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Running Ownership Checker...", .len=36}).ptr);
        fflush(((void*)(NULL)));
        struct ax_OwnershipChecker ownership_chk = ax_AstTree_new_ownership_checker(parser.tree, parser.pool, resolver.symtable, typetable);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Running Escape Analyser...", .len=34}).ptr);
        fflush(((void*)(NULL)));
        struct ax_EscapeAnalyser escape_an = ax_AstTree_new_escape_analyser(parser.tree, parser.pool, resolver.symtable, typetable);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Running CTGC Injector...", .len=32}).ptr);
        fflush(((void*)(NULL)));
        struct ax_CtgcInjector ctgc_inj = ax_AstTree_new_ctgc_injector(parser.tree, parser.pool, resolver.symtable, typetable);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Running Alias Reuse Optimizer...", .len=40}).ptr);
        fflush(((void*)(NULL)));
        struct ax_AliasReuseOptimizer alias_opt = ax_AstTree_new_alias_reuse_optimizer(parser.tree, parser.pool, resolver.symtable, typetable);
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Running AIR Builder...", .len=30}).ptr);
        fflush(((void*)(NULL)));
    }
    struct ax_AirModuleBuilder air_builder_val = ax_AstTree_new_air_module_builder(parser.tree, resolver.symtable, typetable, parser.pool, checker.node_types);
    struct ax_AirModuleBuilder* air_builder = &(air_builder_val);
    puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] main_air: calling build_module", .len=38}).ptr);
    fflush(((void*)(NULL)));
    ax_AirModuleBuilder_build_module(air_builder);
    puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] main_air: build_module returned", .len=39}).ptr);
    fflush(((void*)(NULL)));
    printf((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] main_air: is_dump_air=%d, is_build=%d, optimize=%d, self_link=%d\n", .len=73}).ptr, ((ax_i64)(is_dump_air)), ((ax_i64)(is_build)), ((ax_i64)(optimize)), ((ax_i64)(self_link)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
    fflush(((void*)(NULL)));
    if ((!is_dump_air)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stage 2: Finished building AIR module.", .len=46}).ptr);
    }
    if (optimize) {
        struct ax_SsaOptimizer optimizer = ax_new_ssa_optimizer();
        ax_SsaOptimizer_run(&(optimizer), &(air_builder->module));
    }
    if (is_dump_air) {
        ax_AirModule_print_module(air_builder->module);
    } else if (is_emit_c) {
        struct ax_CGenerator cgen = ax_AirModule_new_c_generator(air_builder->module, parser.tree, air_builder->pool, resolver.symtable, typetable);
        if ((ax_str_len(output_file) == 0)) {
            output_file = (ax_string){.ptr=(const ax_u8*)"output.c", .len=8};
        }
        ax_bool success = ax_CGenerator_generate(&(cgen), output_file);
        if ((!success)) {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: could not generate C code", .len=32}).ptr);
            return 1;
        }
    } else if (is_build) {
        if (target_wasm32) {
            struct ax_WasmGenerator wasm_gen = ax_AirModule_new_wasm_generator(air_builder->module, parser.tree, air_builder->pool, resolver.symtable, typetable);
            if ((ax_str_len(output_file) == 0)) {
                output_file = (ax_string){.ptr=(const ax_u8*)"output.wat", .len=10};
            }
            ax_bool success = ax_WasmGenerator_generate(&(wasm_gen), output_file);
            if ((!success)) {
                puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: could not generate Wasm Text code", .len=40}).ptr);
                return 1;
            }
            ax_free(buffer);
            if ((tokens.data != ((struct ax_Token*)(NULL)))) {
                ax_free(((ax_u8*)(tokens.data)));
            }
            if ((lexer.tokens.data != ((struct ax_Token*)(NULL)))) {
                ax_free(((ax_u8*)(lexer.tokens.data)));
            }
            if ((lexer.indent_stack.data != ((ax_i32*)(NULL)))) {
                ax_free(((ax_u8*)(lexer.indent_stack.data)));
            }
            if ((lexer.newline_offsets.data != ((ax_i32*)(NULL)))) {
                ax_free(((ax_u8*)(lexer.newline_offsets.data)));
            }
            if ((parser.tree.nodes.data != ((struct ax_AstNode*)(NULL)))) {
                ax_free(((ax_u8*)(parser.tree.nodes.data)));
            }
            if ((parser.tree.extras.data != ((ax_i32*)(NULL)))) {
                ax_free(((ax_u8*)(parser.tree.extras.data)));
            }
            ax_AirModuleBuilder_free_air_module_builder(air_builder);
            ax_TypeChecker_free_type_checker(&(checker));
            ax_TypeTable_free_typetable(&(typetable));
            ax_SymbolTable_free_symtable(&(resolver.symtable));
            ax_InternPool_free_pool(&(parser.pool));
            return 0;
        }
        ax_string temp_obj = (ax_string){.ptr=(const ax_u8*)"axiom_temp.obj", .len=14};
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stage 3: Starting native code generation...", .len=51}).ptr);
        ax_bool success = ax_compile_native_binary(temp_obj, air_builder->module, resolver.symtable, parser.pool, typetable, (ax_string){.ptr=(const ax_u8*)"coff", .len=4});
        if ((!success)) {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: could not generate native binary code", .len=44}).ptr);
            return 1;
        }
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stage 4: Finished native code generation.", .len=49}).ptr);
        if ((ax_str_len(output_file) == 0)) {
            output_file = (ax_string){.ptr=(const ax_u8*)"output.exe", .len=10};
        }
        if (self_link) {
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stage 5: Starting self-linking...", .len=41}).ptr);
            struct ax_AxiomLinker linker = ax_new_axiom_linker();
            ax_AxiomLinker_axiom_linker_add_input(&(linker), temp_obj);
            linker.output_path = output_file;
            ax_bool link_success = ax_AxiomLinker_axiom_linker_link(&(linker));
            puts((const char*)((ax_string){.ptr=(const ax_u8*)"[Debug] Stage 6: Finished self-linking.", .len=39}).ptr);
            ax_u8* temp_obj_nt = ax_str_to_null_terminated(temp_obj);
            remove(((void*)(temp_obj_nt)));
            ax_free(temp_obj_nt);
            if ((!link_success)) {
                puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: Self-linking failed", .len=26}).ptr);
                return 1;
            }
        } else {
            {
                ax_u8* cmd_buf = ((ax_u8*)(ax_alloc(2048)));
                ax_u8* output_file_nt = ax_str_to_null_terminated(output_file);
                ax_u8* temp_obj_nt = ax_str_to_null_terminated(temp_obj);
                snprintf(((void*)(cmd_buf)), 2048, (const char*)((ax_string){.ptr=(const ax_u8*)"gcc -O2 -o %s %s -Iruntime runtime/axalloc/axalloc.c runtime/panic/panic.c runtime/ax_assert.c runtime/ax_collections.c runtime/ax_math.c runtime/ax_print.c runtime/ax_string_ops.c runtime/actor/actor.c runtime/actor/async.c runtime/actor/ioloop.c runtime/actor/isolated.c runtime/actor/msgqueue.c runtime/actor/runtime_init.c runtime/actor/scheduler.c runtime/actor/supervisor.c", .len=379}).ptr, ((void*)(output_file_nt)), ((void*)(temp_obj_nt)));
                ax_i32 link_status = system(((void*)(cmd_buf)));
                ax_free(output_file_nt);
                ax_free(temp_obj_nt);
                ax_free(cmd_buf);
                if ((link_status != 0)) {
                    puts((const char*)((ax_string){.ptr=(const ax_u8*)"Error: Linking failed", .len=21}).ptr);
                    return link_status;
                }
            }
        }
    }
    ax_free(buffer);
    if ((tokens.data != ((struct ax_Token*)(NULL)))) {
        ax_free(((ax_u8*)(tokens.data)));
    }
    if ((lexer.tokens.data != ((struct ax_Token*)(NULL)))) {
        ax_free(((ax_u8*)(lexer.tokens.data)));
    }
    if ((lexer.indent_stack.data != ((ax_i32*)(NULL)))) {
        ax_free(((ax_u8*)(lexer.indent_stack.data)));
    }
    if ((lexer.newline_offsets.data != ((ax_i32*)(NULL)))) {
        ax_free(((ax_u8*)(lexer.newline_offsets.data)));
    }
    if ((parser.tree.nodes.data != ((struct ax_AstNode*)(NULL)))) {
        ax_free(((ax_u8*)(parser.tree.nodes.data)));
    }
    if ((parser.tree.extras.data != ((ax_i32*)(NULL)))) {
        ax_free(((ax_u8*)(parser.tree.extras.data)));
    }
    ax_AirModuleBuilder_free_air_module_builder(air_builder);
    ax_TypeChecker_free_type_checker(&(checker));
    ax_TypeTable_free_typetable(&(typetable));
    ax_SymbolTable_free_symtable(&(resolver.symtable));
    ax_InternPool_free_pool(&(parser.pool));
    if (self_link) {
        ax_free(((ax_u8*)(src.ptr)));
    }
    return 0;
    return 0;
}

/* Entry point wrapper */
ax_i32 ax_main(ax_i32 argc, ax_u8** argv) {
    return ax_main_usr(argc, argv);
}
