#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax_Token;
struct ax_TokenVec;
struct ax_IntVec;
struct ax_Lexer;

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
ax_i32 ax_main(ax_i32 argc, ax_u8** argv);


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
    if (ax_str_eq(s, (ax_string){.ptr=(const ax_u8*)"const", .len=5})) {
        return ax_TK_CONST;
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
            self->pos = (self->pos + 1);
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

ax_i32 ax_main(ax_i32 argc, ax_u8** argv) {
    if ((argc < 2)) {
        puts((const char*)((ax_string){.ptr=(const ax_u8*)"Usage: lexer_test <filename>", .len=28}).ptr);
        return 1;
    }
    ax_u8* filename_ptr = ((argv)[1]);
    void* file = fopen(filename_ptr, (const char*)((ax_string){.ptr=(const ax_u8*)"rb", .len=2}).ptr);
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
    ax_i64 bytes_read = fread(buffer, 1, size, file);
    fclose(file);
    ((buffer)[bytes_read]) = ((ax_u8)(0));
    ax_string src = ((ax_string){.ptr = (const ax_u8*)(buffer), .len = strlen((const char*)(buffer))});
    struct ax_Lexer lexer = ax_new_lexer(src);
    struct ax_TokenVec tokens = ax_Lexer_tokenize(&(lexer));
    ax_i32 i = 0;
    while ((i < tokens.len)) {
        struct ax_Token tok = ((tokens.data)[i]);
        printf((const char*)((ax_string){.ptr=(const ax_u8*)"%d,%d,%d\n", .len=9}).ptr, ((ax_i64)(tok.kind)), ((ax_i64)(tok.offset)), ((ax_i64)(tok.len)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)), ((ax_i64)(0)));
        i = (i + 1);
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
    return 0;
    return 0;
}
