#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax__AX_std_Option__i64;
struct ax__AX_std_HashMap__string__i64;
struct ax__AX_std_HashMap__i64__bool;
struct ax__AX_std_HashSet__i64;
struct ax__AX_std_Option__bool;

/* Type definitions */
enum ax__AX_std_Option__i64_tag {
    ax__AX_std_Option__i64_Some = 0,
    ax__AX_std_Option__i64_None = 1,
};

struct ax__AX_std_Option__i64 {
    enum ax__AX_std_Option__i64_tag tag;
    union {
        ax_i64 Some;
    } data;
};

static inline struct ax__AX_std_Option__i64 ax__AX_std_Option__i64_some(ax_i64 value) {
    struct ax__AX_std_Option__i64 _result;
    _result.tag = ax__AX_std_Option__i64_Some;
    _result.data.Some = value;
    return _result;
}

static inline struct ax__AX_std_Option__i64 ax__AX_std_Option__i64_none(void) {
    struct ax__AX_std_Option__i64 _result;
    _result.tag = ax__AX_std_Option__i64_None;
    return _result;
}

struct ax__AX_std_HashMap__string__i64 {
    ax_string* keys;
    ax_i64* values;
    ax_u64* hashes;
    ax_bool* occupied;
    ax_i64 size;
    ax_i64 cap;
};

struct ax__AX_std_HashMap__i64__bool {
    ax_i64* keys;
    ax_bool* values;
    ax_u64* hashes;
    ax_bool* occupied;
    ax_i64 size;
    ax_i64 cap;
};

struct ax__AX_std_HashSet__i64 {
    struct ax__AX_std_HashMap__i64__bool map;
};

enum ax__AX_std_Option__bool_tag {
    ax__AX_std_Option__bool_Some = 0,
    ax__AX_std_Option__bool_None = 1,
};

struct ax__AX_std_Option__bool {
    enum ax__AX_std_Option__bool_tag tag;
    union {
        ax_bool Some;
    } data;
};

static inline struct ax__AX_std_Option__bool ax__AX_std_Option__bool_some(ax_bool value) {
    struct ax__AX_std_Option__bool _result;
    _result.tag = ax__AX_std_Option__bool_Some;
    _result.data.Some = value;
    return _result;
}

static inline struct ax__AX_std_Option__bool ax__AX_std_Option__bool_none(void) {
    struct ax__AX_std_Option__bool _result;
    _result.tag = ax__AX_std_Option__bool_None;
    return _result;
}


/* Function prototypes */
ax_i64 syscall(ax_u64 num, ax_u64 a1, ax_u64 a2, ax_u64 a3, ax_u64 a4, ax_u64 a5, ax_u64 a6);
static void ax_test_print_str(ax_string s);
static void ax_test_vec(void);
static void ax_test_hashmap(void);
static void ax_test_hashset(void);
ax_i32 ax_main_usr(void);
ax_vec ax_AX_std_new_vec__i64(void);
void ax__AX_std_Vec__i64_push(ax_vec* self, ax_i64 item);
struct ax__AX_std_Option__i64 ax__AX_std_Vec__i64_get(ax_vec self, ax_i64 index);
struct ax__AX_std_Option__i64 ax__AX_std_Vec__i64_pop(ax_vec* self);
void ax__AX_std_Vec__i64_clear(ax_vec* self);
void ax__AX_std_Vec__i64_destroy(ax_vec* self);
struct ax__AX_std_HashMap__string__i64 ax_AX_std_new_hashmap__string__i64(void);
ax_i64 ax__AX_std_HashMap__string__i64_len(struct ax__AX_std_HashMap__string__i64 self);
void ax__AX_std_HashMap__string__i64_insert(struct ax__AX_std_HashMap__string__i64* self, ax_string key, ax_i64 value);
static ax_u64 ax_AX_std_hash_key__string(ax_string key);
struct ax__AX_std_Option__i64 ax__AX_std_HashMap__string__i64_get(struct ax__AX_std_HashMap__string__i64 self, ax_string key);
ax_bool ax__AX_std_HashMap__string__i64_remove(struct ax__AX_std_HashMap__string__i64* self, ax_string key);
void ax__AX_std_HashMap__string__i64_destroy(struct ax__AX_std_HashMap__string__i64* self);
struct ax__AX_std_HashSet__i64 ax_AX_std_new_hashset__i64(void);
struct ax__AX_std_HashMap__i64__bool ax_AX_std_new_hashmap__i64__bool(void);
ax_i64 ax__AX_std_HashSet__i64_len(struct ax__AX_std_HashSet__i64 self);
ax_i64 ax__AX_std_HashMap__i64__bool_len(struct ax__AX_std_HashMap__i64__bool self);
ax_bool ax__AX_std_HashSet__i64_insert(struct ax__AX_std_HashSet__i64* self, ax_i64 item);
ax_bool ax__AX_std_HashSet__i64_contains(struct ax__AX_std_HashSet__i64 self, ax_i64 item);
struct ax__AX_std_Option__bool ax__AX_std_HashMap__i64__bool_get(struct ax__AX_std_HashMap__i64__bool self, ax_i64 key);
static ax_u64 ax_AX_std_hash_key__i64(ax_i64 key);
void ax__AX_std_HashMap__i64__bool_insert(struct ax__AX_std_HashMap__i64__bool* self, ax_i64 key, ax_bool value);
ax_bool ax__AX_std_HashSet__i64_remove(struct ax__AX_std_HashSet__i64* self, ax_i64 item);
ax_bool ax__AX_std_HashMap__i64__bool_remove(struct ax__AX_std_HashMap__i64__bool* self, ax_i64 key);
void ax__AX_std_HashSet__i64_destroy(struct ax__AX_std_HashSet__i64* self);
void ax__AX_std_HashMap__i64__bool_destroy(struct ax__AX_std_HashMap__i64__bool* self);
ax_i64 ax_std_string_len(ax_string p0);
ax_bool ax_std_string_starts_with(ax_string p0, ax_string p1);


static void ax_test_print_str(ax_string s) {
    if (1) {
        void* h = GetStdHandle(((ax_u32)(0xFFFFFFF5)));
        ax_u32 written = ((ax_u32)(0));
        WriteFile(h, ((void*)(s.ptr)), ((ax_u32)(ax_str_len(s))), ((void*)(&(written))), ((void*)(NULL)));
    } else {
        {
            syscall(((ax_u64)(1)), ((ax_u64)(1)), ((ax_u64)(((ax_u8*)(s.ptr)))), ((ax_u64)(ax_str_len(s))), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
        }
    }
}

static void ax_test_vec(void) {
    ax_vec v = ax_AX_std_new_vec__i64();
    ax_assert_axiom((v.len == 0), AX_STR("(v.len == 0)"));
    ax_assert_axiom((v.cap == 0), AX_STR("(v.cap == 0)"));
    ax__AX_std_Vec__i64_push(&(v), ((ax_i64)(10)));
    ax__AX_std_Vec__i64_push(&(v), ((ax_i64)(20)));
    ax__AX_std_Vec__i64_push(&(v), ((ax_i64)(30)));
    ax_assert_axiom((v.len == 3), AX_STR("(v.len == 3)"));
    ax_assert_axiom((v.cap >= 3), AX_STR("(v.cap >= 3)"));
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_Vec__i64_get(v, 0);
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_i64 val = (_discrim).data.Some;
            ax_assert_axiom((val == 10), AX_STR("(val == 10)"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_Vec__i64_get(v, 2);
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_i64 val = (_discrim).data.Some;
            ax_assert_axiom((val == 30), AX_STR("(val == 30)"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_Vec__i64_get(v, 3);
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_TRUE, AX_STR("AX_TRUE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_Vec__i64_pop(&(v));
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_i64 val = (_discrim).data.Some;
            ax_assert_axiom((val == 30), AX_STR("(val == 30)"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    ax_assert_axiom((v.len == 2), AX_STR("(v.len == 2)"));
    ax__AX_std_Vec__i64_clear(&(v));
    ax_assert_axiom((v.len == 0), AX_STR("(v.len == 0)"));
    ax__AX_std_Vec__i64_destroy(&(v));
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_vec\n", .len=17});
}

static void ax_test_hashmap(void) {
    struct ax__AX_std_HashMap__string__i64 m = ax_AX_std_new_hashmap__string__i64();
    ax_assert_axiom((ax__AX_std_HashMap__string__i64_len(m) == 0), AX_STR("(ax__AX_std_HashMap__string__i64_len(m) == 0)"));
    ax__AX_std_HashMap__string__i64_insert(&(m), (ax_string){.ptr=(const ax_u8*)"apple", .len=5}, ((ax_i64)(100)));
    ax__AX_std_HashMap__string__i64_insert(&(m), (ax_string){.ptr=(const ax_u8*)"banana", .len=6}, ((ax_i64)(200)));
    ax__AX_std_HashMap__string__i64_insert(&(m), (ax_string){.ptr=(const ax_u8*)"orange", .len=6}, ((ax_i64)(300)));
    ax_assert_axiom((ax__AX_std_HashMap__string__i64_len(m) == 3), AX_STR("(ax__AX_std_HashMap__string__i64_len(m) == 3)"));
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_HashMap__string__i64_get(m, (ax_string){.ptr=(const ax_u8*)"apple", .len=5});
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_i64 val = (_discrim).data.Some;
            ax_assert_axiom((val == 100), AX_STR("(val == 100)"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_HashMap__string__i64_get(m, (ax_string){.ptr=(const ax_u8*)"banana", .len=6});
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_i64 val = (_discrim).data.Some;
            ax_assert_axiom((val == 200), AX_STR("(val == 200)"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_HashMap__string__i64_get(m, (ax_string){.ptr=(const ax_u8*)"orange", .len=6});
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_i64 val = (_discrim).data.Some;
            ax_assert_axiom((val == 300), AX_STR("(val == 300)"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_HashMap__string__i64_get(m, (ax_string){.ptr=(const ax_u8*)"pear", .len=4});
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_TRUE, AX_STR("AX_TRUE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    ax_i64 k = ((ax_i64)(0));
    while ((k < 50)) {
        ax_u8* key_buf = ((ax_u8*)(ax_alloc(2)));
        (((ax_u8*)(key_buf))[0]) = (((ax_u8)('A')) + ((ax_u8)((k % ((ax_i64)(26))))));
        (((ax_u8*)(key_buf))[1]) = ((ax_u8)(0));
        ax_string key_str = ((ax_string){.ptr = (const ax_u8*)(key_buf), .len = strlen((const char*)(key_buf))});
        ax__AX_std_HashMap__string__i64_insert(&(m), key_str, k);
        k = (k + 1);
    }
    k = 0;
    while ((k < 50)) {
        ax_u8* key_buf = ((ax_u8*)(ax_alloc(2)));
        (((ax_u8*)(key_buf))[0]) = (((ax_u8)('A')) + ((ax_u8)((k % ((ax_i64)(26))))));
        (((ax_u8*)(key_buf))[1]) = ((ax_u8)(0));
        ax_string key_str = ((ax_string){.ptr = (const ax_u8*)(key_buf), .len = strlen((const char*)(key_buf))});
        {
            struct ax__AX_std_Option__i64 _discrim = ax__AX_std_HashMap__string__i64_get(m, key_str);
            switch (_discrim.tag) {
            case ax__AX_std_Option__i64_Some: {
                ax_i64 val = (_discrim).data.Some;
                ax_assert_axiom(AX_TRUE, AX_STR("AX_TRUE"));
                break;
            }
            case ax__AX_std_Option__i64_None: {
                ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
                break;
            }
                default: {
                    /* unreachable: exhaustiveness checked by type checker */
                    __builtin_unreachable();
                }
            }
        }
        ax_free(key_buf);
        k = (k + 1);
    }
    ax_bool ok1 = ax__AX_std_HashMap__string__i64_remove(&(m), (ax_string){.ptr=(const ax_u8*)"apple", .len=5});
    ax_assert_axiom((ok1 == AX_TRUE), AX_STR("(ok1 == AX_TRUE)"));
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_HashMap__string__i64_get(m, (ax_string){.ptr=(const ax_u8*)"apple", .len=5});
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_TRUE, AX_STR("AX_TRUE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    ax_bool ok2 = ax__AX_std_HashMap__string__i64_remove(&(m), (ax_string){.ptr=(const ax_u8*)"banana", .len=6});
    ax_assert_axiom((ok2 == AX_TRUE), AX_STR("(ok2 == AX_TRUE)"));
    {
        struct ax__AX_std_Option__i64 _discrim = ax__AX_std_HashMap__string__i64_get(m, (ax_string){.ptr=(const ax_u8*)"banana", .len=6});
        switch (_discrim.tag) {
        case ax__AX_std_Option__i64_Some: {
            ax_assert_axiom(AX_FALSE, AX_STR("AX_FALSE"));
            break;
        }
        case ax__AX_std_Option__i64_None: {
            ax_assert_axiom(AX_TRUE, AX_STR("AX_TRUE"));
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
    ax_bool ok3 = ax__AX_std_HashMap__string__i64_remove(&(m), (ax_string){.ptr=(const ax_u8*)"pear", .len=4});
    ax_assert_axiom((ok3 == AX_FALSE), AX_STR("(ok3 == AX_FALSE)"));
    ax__AX_std_HashMap__string__i64_destroy(&(m));
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_hashmap\n", .len=21});
}

static void ax_test_hashset(void) {
    struct ax__AX_std_HashSet__i64 s = ax_AX_std_new_hashset__i64();
    ax_assert_axiom((ax__AX_std_HashSet__i64_len(s) == 0), AX_STR("(ax__AX_std_HashSet__i64_len(s) == 0)"));
    ax_bool ins1 = ax__AX_std_HashSet__i64_insert(&(s), ((ax_i64)(10)));
    ax_assert_axiom((ins1 == AX_TRUE), AX_STR("(ins1 == AX_TRUE)"));
    ax_bool ins2 = ax__AX_std_HashSet__i64_insert(&(s), ((ax_i64)(20)));
    ax_assert_axiom((ins2 == AX_TRUE), AX_STR("(ins2 == AX_TRUE)"));
    ax_bool ins3 = ax__AX_std_HashSet__i64_insert(&(s), ((ax_i64)(10)));
    ax_assert_axiom((ins3 == AX_FALSE), AX_STR("(ins3 == AX_FALSE)"));
    ax_assert_axiom((ax__AX_std_HashSet__i64_len(s) == 2), AX_STR("(ax__AX_std_HashSet__i64_len(s) == 2)"));
    ax_assert_axiom((ax__AX_std_HashSet__i64_contains(s, ((ax_i64)(10))) == AX_TRUE), AX_STR("(ax__AX_std_HashSet__i64_contains(s, ((ax_i64)(10))) == AX_TRUE)"));
    ax_assert_axiom((ax__AX_std_HashSet__i64_contains(s, ((ax_i64)(20))) == AX_TRUE), AX_STR("(ax__AX_std_HashSet__i64_contains(s, ((ax_i64)(20))) == AX_TRUE)"));
    ax_assert_axiom((ax__AX_std_HashSet__i64_contains(s, ((ax_i64)(30))) == AX_FALSE), AX_STR("(ax__AX_std_HashSet__i64_contains(s, ((ax_i64)(30))) == AX_FALSE)"));
    ax_bool rem1 = ax__AX_std_HashSet__i64_remove(&(s), ((ax_i64)(10)));
    ax_assert_axiom((rem1 == AX_TRUE), AX_STR("(rem1 == AX_TRUE)"));
    ax_assert_axiom((ax__AX_std_HashSet__i64_contains(s, ((ax_i64)(10))) == AX_FALSE), AX_STR("(ax__AX_std_HashSet__i64_contains(s, ((ax_i64)(10))) == AX_FALSE)"));
    ax_assert_axiom((ax__AX_std_HashSet__i64_len(s) == 1), AX_STR("(ax__AX_std_HashSet__i64_len(s) == 1)"));
    ax_bool rem2 = ax__AX_std_HashSet__i64_remove(&(s), ((ax_i64)(30)));
    ax_assert_axiom((rem2 == AX_FALSE), AX_STR("(rem2 == AX_FALSE)"));
    ax__AX_std_HashSet__i64_destroy(&(s));
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_hashset\n", .len=21});
}

ax_i32 ax_main_usr(void) {
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"Running AXIOM-native collections unit tests (Robin Hood Hashing)...\n", .len=68});
    ax_test_vec();
    ax_test_hashmap();
    ax_test_hashset();
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"All AXIOM-native collections tests passed!\n", .len=43});
    return 0;
}

ax_vec ax_AX_std_new_vec__i64(void) {
    return ((ax_vec){.data=((ax_i64*)(NULL)), .len=((ax_i64)(0)), .cap=((ax_i64)(0))});
}

void ax__AX_std_Vec__i64_push(ax_vec* self, ax_i64 item) {
    if ((self->len == self->cap)) {
        ax_i64 new_cap = ((ax_i64)(16));
        if ((self->cap != 0)) {
            new_cap = (self->cap * 2);
        }
        ax_i64* new_data = ((ax_i64*)(ax_alloc((new_cap * 8))));
        if ((self->data != ((ax_i64*)(NULL)))) {
            memcpy(((ax_u8*)(new_data)), ((ax_u8*)(self->data)), (self->len * 8));
            ax_free(((ax_u8*)(self->data)));
        }
        self->data = new_data;
        self->cap = new_cap;
    }
    (((ax_i64*)(self->data))[self->len]) = item;
    self->len = (self->len + 1);
}

struct ax__AX_std_Option__i64 ax__AX_std_Vec__i64_get(ax_vec self, ax_i64 index) {
    if (((index < 0) || (index >= self.len))) {
        return ax__AX_std_Option__i64_none();
    }
    return ax__AX_std_Option__i64_some((((ax_i64*)(self.data))[index]));
}

struct ax__AX_std_Option__i64 ax__AX_std_Vec__i64_pop(ax_vec* self) {
    if ((self->len == 0)) {
        return ax__AX_std_Option__i64_none();
    }
    self->len = (self->len - 1);
    return ax__AX_std_Option__i64_some((((ax_i64*)(self->data))[self->len]));
}

void ax__AX_std_Vec__i64_clear(ax_vec* self) {
    self->len = 0;
}

void ax__AX_std_Vec__i64_destroy(ax_vec* self) {
    if ((self->data != ((ax_i64*)(NULL)))) {
        ax_free(((ax_u8*)(self->data)));
        self->data = ((ax_i64*)(NULL));
    }
    self->len = 0;
    self->cap = 0;
}

struct ax__AX_std_HashMap__string__i64 ax_AX_std_new_hashmap__string__i64(void) {
    return ((struct ax__AX_std_HashMap__string__i64){.keys=((ax_string*)(NULL)), .values=((ax_i64*)(NULL)), .hashes=((ax_u64*)(NULL)), .occupied=((ax_bool*)(NULL)), .size=((ax_i64)(0)), .cap=((ax_i64)(0))});
}

ax_i64 ax__AX_std_HashMap__string__i64_len(struct ax__AX_std_HashMap__string__i64 self) {
    return self.size;
}

void ax__AX_std_HashMap__string__i64_insert(struct ax__AX_std_HashMap__string__i64* self, ax_string key, ax_i64 value) {
    if ((self->cap == 0)) {
        self->cap = 16;
        self->keys = ((ax_string*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_string)))))));
        self->values = ((ax_i64*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_i64)))))));
        self->hashes = ((ax_u64*)(ax_alloc((self->cap * 8))));
        self->occupied = ((ax_bool*)(ax_alloc((self->cap * 1))));
        ax_i64 i = ((ax_i64)(0));
        while ((i < self->cap)) {
            (((ax_bool*)(self->occupied))[i]) = AX_FALSE;
            i = (i + 1);
        }
    }
    if (((self->size * 2) >= self->cap)) {
        ax_i64 old_cap = self->cap;
        ax_string* old_keys = self->keys;
        ax_i64* old_values = self->values;
        ax_u64* old_hashes = self->hashes;
        ax_bool* old_occupied = self->occupied;
        self->cap = (old_cap * 2);
        self->keys = ((ax_string*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_string)))))));
        self->values = ((ax_i64*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_i64)))))));
        self->hashes = ((ax_u64*)(ax_alloc((self->cap * 8))));
        self->occupied = ((ax_bool*)(ax_alloc((self->cap * 1))));
        ax_i64 i = ((ax_i64)(0));
        while ((i < self->cap)) {
            (((ax_bool*)(self->occupied))[i]) = AX_FALSE;
            i = (i + 1);
        }
        self->size = 0;
        ax_i64 j = ((ax_i64)(0));
        while ((j < old_cap)) {
            if ((((ax_bool*)(old_occupied))[j])) {
                ax__AX_std_HashMap__string__i64_insert(self, (((ax_string*)(old_keys))[j]), (((ax_i64*)(old_values))[j]));
            }
            j = (j + 1);
        }
        if ((old_keys != ((ax_string*)(NULL)))) {
            ax_free(((ax_u8*)(old_keys)));
            ax_free(((ax_u8*)(old_values)));
            ax_free(((ax_u8*)(old_hashes)));
            ax_free(((ax_u8*)(old_occupied)));
        }
    }
    ax_u64 curr_h = ax_AX_std_hash_key__string(key);
    ax_string curr_key = key;
    ax_i64 curr_value = value;
    ax_i64 idx = ((ax_i64)((curr_h % ((ax_u64)(self->cap)))));
    ax_i64 current_dib = ((ax_i64)(0));
    ax_bool loop = AX_TRUE;
    while (loop) {
        if ((!(((ax_bool*)(self->occupied))[idx]))) {
            (((ax_string*)(self->keys))[idx]) = curr_key;
            (((ax_i64*)(self->values))[idx]) = curr_value;
            (((ax_u64*)(self->hashes))[idx]) = curr_h;
            (((ax_bool*)(self->occupied))[idx]) = AX_TRUE;
            self->size = (self->size + 1);
            loop = AX_FALSE;
        } else if ((((((ax_u64*)(self->hashes))[idx]) == curr_h) && ax_str_eq((((ax_string*)(self->keys))[idx]), curr_key))) {
            (((ax_i64*)(self->values))[idx]) = curr_value;
            loop = AX_FALSE;
        } else {
            {
                ax_u64 resident_h = (((ax_u64*)(self->hashes))[idx]);
                ax_i64 resident_dib = (((idx - ((ax_i64)((resident_h % ((ax_u64)(self->cap)))))) + self->cap) % self->cap);
                if ((current_dib > resident_dib)) {
                    ax_string tmp_key = (((ax_string*)(self->keys))[idx]);
                    ax_i64 tmp_value = (((ax_i64*)(self->values))[idx]);
                    ax_u64 tmp_h = resident_h;
                    (((ax_string*)(self->keys))[idx]) = curr_key;
                    (((ax_i64*)(self->values))[idx]) = curr_value;
                    (((ax_u64*)(self->hashes))[idx]) = curr_h;
                    curr_key = tmp_key;
                    curr_value = tmp_value;
                    curr_h = tmp_h;
                    current_dib = resident_dib;
                }
                idx = ((idx + 1) % self->cap);
                current_dib = (current_dib + 1);
            }
        }
    }
}

static ax_u64 ax_AX_std_hash_key__string(ax_string key) {
    ax_u8* ptr = ((ax_u8*)(&(key)));
    ax_u64 size = sizeof(ax_string);
    if ((size == 16)) {
        ax_string s = (*((ax_string*)(((ax_string*)(ptr)))));
        ax_u64 h = ((ax_u64)(14695981039346656037));
        ax_i64 i = ((ax_i64)(0));
        while ((i < s.len)) {
            h = (h ^ ((ax_u64)((((ax_u8*)(s.ptr))[i]))));
            h = (h * ((ax_u64)(1099511628211)));
            i = (i + 1);
        }
        return h;
    } else {
        {
            ax_u64 h = ((ax_u64)(14695981039346656037));
            ax_i64 i = ((ax_i64)(0));
            while ((i < size)) {
                h = (h ^ ((ax_u64)((((ax_u8*)(ptr))[i]))));
                h = (h * ((ax_u64)(1099511628211)));
                i = (i + 1);
            }
            return h;
        }
    }
}

struct ax__AX_std_Option__i64 ax__AX_std_HashMap__string__i64_get(struct ax__AX_std_HashMap__string__i64 self, ax_string key) {
    if ((self.cap == 0)) {
        return ax__AX_std_Option__i64_none();
    }
    ax_u64 h = ax_AX_std_hash_key__string(key);
    ax_i64 idx = ((ax_i64)((h % ((ax_u64)(self.cap)))));
    ax_i64 current_dib = ((ax_i64)(0));
    ax_bool loop = AX_TRUE;
    while (loop) {
        if ((!(((ax_bool*)(self.occupied))[idx]))) {
            return ax__AX_std_Option__i64_none();
        } else if ((((((ax_u64*)(self.hashes))[idx]) == h) && ax_str_eq((((ax_string*)(self.keys))[idx]), key))) {
            return ax__AX_std_Option__i64_some((((ax_i64*)(self.values))[idx]));
        } else {
            {
                ax_u64 resident_h = (((ax_u64*)(self.hashes))[idx]);
                ax_i64 resident_dib = (((idx - ((ax_i64)((resident_h % ((ax_u64)(self.cap)))))) + self.cap) % self.cap);
                if ((current_dib > resident_dib)) {
                    return ax__AX_std_Option__i64_none();
                }
                idx = ((idx + 1) % self.cap);
                current_dib = (current_dib + 1);
                if ((current_dib >= self.cap)) {
                    return ax__AX_std_Option__i64_none();
                }
            }
        }
    }
}

ax_bool ax__AX_std_HashMap__string__i64_remove(struct ax__AX_std_HashMap__string__i64* self, ax_string key) {
    if ((self->cap == 0)) {
        return AX_FALSE;
    }
    ax_u64 h = ax_AX_std_hash_key__string(key);
    ax_i64 idx = ((ax_i64)((h % ((ax_u64)(self->cap)))));
    ax_i64 current_dib = ((ax_i64)(0));
    ax_bool found = AX_FALSE;
    ax_bool loop = AX_TRUE;
    while (loop) {
        if ((!(((ax_bool*)(self->occupied))[idx]))) {
            return AX_FALSE;
        } else if ((((((ax_u64*)(self->hashes))[idx]) == h) && ax_str_eq((((ax_string*)(self->keys))[idx]), key))) {
            found = AX_TRUE;
            loop = AX_FALSE;
        } else {
            {
                ax_u64 resident_h = (((ax_u64*)(self->hashes))[idx]);
                ax_i64 resident_dib = (((idx - ((ax_i64)((resident_h % ((ax_u64)(self->cap)))))) + self->cap) % self->cap);
                if ((current_dib > resident_dib)) {
                    return AX_FALSE;
                }
                idx = ((idx + 1) % self->cap);
                current_dib = (current_dib + 1);
                if ((current_dib >= self->cap)) {
                    return AX_FALSE;
                }
            }
        }
    }
    if ((!found)) {
        return AX_FALSE;
    }
    self->size = (self->size - 1);
    ax_i64 curr = idx;
    ax_i64 next = ((idx + 1) % self->cap);
    ax_bool shift_loop = AX_TRUE;
    while (shift_loop) {
        if ((!(((ax_bool*)(self->occupied))[next]))) {
            (((ax_bool*)(self->occupied))[curr]) = AX_FALSE;
            shift_loop = AX_FALSE;
        } else {
            {
                ax_u64 next_h = (((ax_u64*)(self->hashes))[next]);
                ax_i64 next_dib = (((next - ((ax_i64)((next_h % ((ax_u64)(self->cap)))))) + self->cap) % self->cap);
                if ((next_dib == ((ax_i64)(0)))) {
                    (((ax_bool*)(self->occupied))[curr]) = AX_FALSE;
                    shift_loop = AX_FALSE;
                } else {
                    {
                        (((ax_string*)(self->keys))[curr]) = (((ax_string*)(self->keys))[next]);
                        (((ax_i64*)(self->values))[curr]) = (((ax_i64*)(self->values))[next]);
                        (((ax_u64*)(self->hashes))[curr]) = next_h;
                        (((ax_bool*)(self->occupied))[curr]) = AX_TRUE;
                        curr = next;
                        next = ((next + 1) % self->cap);
                    }
                }
            }
        }
    }
    return AX_TRUE;
}

void ax__AX_std_HashMap__string__i64_destroy(struct ax__AX_std_HashMap__string__i64* self) {
    if ((self->keys != ((ax_string*)(NULL)))) {
        ax_free(((ax_u8*)(self->keys)));
        ax_free(((ax_u8*)(self->values)));
        ax_free(((ax_u8*)(self->hashes)));
        ax_free(((ax_u8*)(self->occupied)));
        self->keys = ((ax_string*)(NULL));
        self->values = ((ax_i64*)(NULL));
        self->hashes = ((ax_u64*)(NULL));
        self->occupied = ((ax_bool*)(NULL));
    }
    self->size = 0;
    self->cap = 0;
}

struct ax__AX_std_HashSet__i64 ax_AX_std_new_hashset__i64(void) {
    return ((struct ax__AX_std_HashSet__i64){.map=ax_AX_std_new_hashmap__i64__bool()});
}

struct ax__AX_std_HashMap__i64__bool ax_AX_std_new_hashmap__i64__bool(void) {
    return ((struct ax__AX_std_HashMap__i64__bool){.keys=((ax_i64*)(NULL)), .values=((ax_bool*)(NULL)), .hashes=((ax_u64*)(NULL)), .occupied=((ax_bool*)(NULL)), .size=((ax_i64)(0)), .cap=((ax_i64)(0))});
}

ax_i64 ax__AX_std_HashSet__i64_len(struct ax__AX_std_HashSet__i64 self) {
    return ax__AX_std_HashMap__i64__bool_len(self.map);
}

ax_i64 ax__AX_std_HashMap__i64__bool_len(struct ax__AX_std_HashMap__i64__bool self) {
    return self.size;
}

ax_bool ax__AX_std_HashSet__i64_insert(struct ax__AX_std_HashSet__i64* self, ax_i64 item) {
    ax_bool exists = ax__AX_std_HashSet__i64_contains(*(self), item);
    ax__AX_std_HashMap__i64__bool_insert(&(self->map), item, AX_TRUE);
    return (!exists);
}

ax_bool ax__AX_std_HashSet__i64_contains(struct ax__AX_std_HashSet__i64 self, ax_i64 item) {
    {
        struct ax__AX_std_Option__bool _discrim = ax__AX_std_HashMap__i64__bool_get(self.map, item);
        switch (_discrim.tag) {
        case ax__AX_std_Option__bool_Some: {
            return AX_TRUE;
            break;
        }
        case ax__AX_std_Option__bool_None: {
            return AX_FALSE;
            break;
        }
            default: {
                /* unreachable: exhaustiveness checked by type checker */
                __builtin_unreachable();
            }
        }
    }
}

struct ax__AX_std_Option__bool ax__AX_std_HashMap__i64__bool_get(struct ax__AX_std_HashMap__i64__bool self, ax_i64 key) {
    if ((self.cap == 0)) {
        return ax__AX_std_Option__bool_none();
    }
    ax_u64 h = ax_AX_std_hash_key__i64(key);
    ax_i64 idx = ((ax_i64)((h % ((ax_u64)(self.cap)))));
    ax_i64 current_dib = ((ax_i64)(0));
    ax_bool loop = AX_TRUE;
    while (loop) {
        if ((!(((ax_bool*)(self.occupied))[idx]))) {
            return ax__AX_std_Option__bool_none();
        } else if ((((((ax_u64*)(self.hashes))[idx]) == h) && ((((ax_i64*)(self.keys))[idx]) == key))) {
            return ax__AX_std_Option__bool_some((((ax_bool*)(self.values))[idx]));
        } else {
            {
                ax_u64 resident_h = (((ax_u64*)(self.hashes))[idx]);
                ax_i64 resident_dib = (((idx - ((ax_i64)((resident_h % ((ax_u64)(self.cap)))))) + self.cap) % self.cap);
                if ((current_dib > resident_dib)) {
                    return ax__AX_std_Option__bool_none();
                }
                idx = ((idx + 1) % self.cap);
                current_dib = (current_dib + 1);
                if ((current_dib >= self.cap)) {
                    return ax__AX_std_Option__bool_none();
                }
            }
        }
    }
}

static ax_u64 ax_AX_std_hash_key__i64(ax_i64 key) {
    ax_u8* ptr = ((ax_u8*)(&(key)));
    ax_u64 size = sizeof(ax_i64);
    if ((size == 16)) {
        ax_string s = (*((ax_string*)(((ax_string*)(ptr)))));
        ax_u64 h = ((ax_u64)(14695981039346656037));
        ax_i64 i = ((ax_i64)(0));
        while ((i < s.len)) {
            h = (h ^ ((ax_u64)((((ax_u8*)(s.ptr))[i]))));
            h = (h * ((ax_u64)(1099511628211)));
            i = (i + 1);
        }
        return h;
    } else {
        {
            ax_u64 h = ((ax_u64)(14695981039346656037));
            ax_i64 i = ((ax_i64)(0));
            while ((i < size)) {
                h = (h ^ ((ax_u64)((((ax_u8*)(ptr))[i]))));
                h = (h * ((ax_u64)(1099511628211)));
                i = (i + 1);
            }
            return h;
        }
    }
}

void ax__AX_std_HashMap__i64__bool_insert(struct ax__AX_std_HashMap__i64__bool* self, ax_i64 key, ax_bool value) {
    if ((self->cap == 0)) {
        self->cap = 16;
        self->keys = ((ax_i64*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_i64)))))));
        self->values = ((ax_bool*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_bool)))))));
        self->hashes = ((ax_u64*)(ax_alloc((self->cap * 8))));
        self->occupied = ((ax_bool*)(ax_alloc((self->cap * 1))));
        ax_i64 i = ((ax_i64)(0));
        while ((i < self->cap)) {
            (((ax_bool*)(self->occupied))[i]) = AX_FALSE;
            i = (i + 1);
        }
    }
    if (((self->size * 2) >= self->cap)) {
        ax_i64 old_cap = self->cap;
        ax_i64* old_keys = self->keys;
        ax_bool* old_values = self->values;
        ax_u64* old_hashes = self->hashes;
        ax_bool* old_occupied = self->occupied;
        self->cap = (old_cap * 2);
        self->keys = ((ax_i64*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_i64)))))));
        self->values = ((ax_bool*)(ax_alloc((self->cap * ((ax_i64)(sizeof(ax_bool)))))));
        self->hashes = ((ax_u64*)(ax_alloc((self->cap * 8))));
        self->occupied = ((ax_bool*)(ax_alloc((self->cap * 1))));
        ax_i64 i = ((ax_i64)(0));
        while ((i < self->cap)) {
            (((ax_bool*)(self->occupied))[i]) = AX_FALSE;
            i = (i + 1);
        }
        self->size = 0;
        ax_i64 j = ((ax_i64)(0));
        while ((j < old_cap)) {
            if ((((ax_bool*)(old_occupied))[j])) {
                ax__AX_std_HashMap__i64__bool_insert(self, (((ax_i64*)(old_keys))[j]), (((ax_bool*)(old_values))[j]));
            }
            j = (j + 1);
        }
        if ((old_keys != ((ax_i64*)(NULL)))) {
            ax_free(((ax_u8*)(old_keys)));
            ax_free(((ax_u8*)(old_values)));
            ax_free(((ax_u8*)(old_hashes)));
            ax_free(((ax_u8*)(old_occupied)));
        }
    }
    ax_u64 curr_h = ax_AX_std_hash_key__i64(key);
    ax_i64 curr_key = key;
    ax_bool curr_value = value;
    ax_i64 idx = ((ax_i64)((curr_h % ((ax_u64)(self->cap)))));
    ax_i64 current_dib = ((ax_i64)(0));
    ax_bool loop = AX_TRUE;
    while (loop) {
        if ((!(((ax_bool*)(self->occupied))[idx]))) {
            (((ax_i64*)(self->keys))[idx]) = curr_key;
            (((ax_bool*)(self->values))[idx]) = curr_value;
            (((ax_u64*)(self->hashes))[idx]) = curr_h;
            (((ax_bool*)(self->occupied))[idx]) = AX_TRUE;
            self->size = (self->size + 1);
            loop = AX_FALSE;
        } else if ((((((ax_u64*)(self->hashes))[idx]) == curr_h) && ((((ax_i64*)(self->keys))[idx]) == curr_key))) {
            (((ax_bool*)(self->values))[idx]) = curr_value;
            loop = AX_FALSE;
        } else {
            {
                ax_u64 resident_h = (((ax_u64*)(self->hashes))[idx]);
                ax_i64 resident_dib = (((idx - ((ax_i64)((resident_h % ((ax_u64)(self->cap)))))) + self->cap) % self->cap);
                if ((current_dib > resident_dib)) {
                    ax_i64 tmp_key = (((ax_i64*)(self->keys))[idx]);
                    ax_bool tmp_value = (((ax_bool*)(self->values))[idx]);
                    ax_u64 tmp_h = resident_h;
                    (((ax_i64*)(self->keys))[idx]) = curr_key;
                    (((ax_bool*)(self->values))[idx]) = curr_value;
                    (((ax_u64*)(self->hashes))[idx]) = curr_h;
                    curr_key = tmp_key;
                    curr_value = tmp_value;
                    curr_h = tmp_h;
                    current_dib = resident_dib;
                }
                idx = ((idx + 1) % self->cap);
                current_dib = (current_dib + 1);
            }
        }
    }
}

ax_bool ax__AX_std_HashSet__i64_remove(struct ax__AX_std_HashSet__i64* self, ax_i64 item) {
    return ax__AX_std_HashMap__i64__bool_remove(&(self->map), item);
}

ax_bool ax__AX_std_HashMap__i64__bool_remove(struct ax__AX_std_HashMap__i64__bool* self, ax_i64 key) {
    if ((self->cap == 0)) {
        return AX_FALSE;
    }
    ax_u64 h = ax_AX_std_hash_key__i64(key);
    ax_i64 idx = ((ax_i64)((h % ((ax_u64)(self->cap)))));
    ax_i64 current_dib = ((ax_i64)(0));
    ax_bool found = AX_FALSE;
    ax_bool loop = AX_TRUE;
    while (loop) {
        if ((!(((ax_bool*)(self->occupied))[idx]))) {
            return AX_FALSE;
        } else if ((((((ax_u64*)(self->hashes))[idx]) == h) && ((((ax_i64*)(self->keys))[idx]) == key))) {
            found = AX_TRUE;
            loop = AX_FALSE;
        } else {
            {
                ax_u64 resident_h = (((ax_u64*)(self->hashes))[idx]);
                ax_i64 resident_dib = (((idx - ((ax_i64)((resident_h % ((ax_u64)(self->cap)))))) + self->cap) % self->cap);
                if ((current_dib > resident_dib)) {
                    return AX_FALSE;
                }
                idx = ((idx + 1) % self->cap);
                current_dib = (current_dib + 1);
                if ((current_dib >= self->cap)) {
                    return AX_FALSE;
                }
            }
        }
    }
    if ((!found)) {
        return AX_FALSE;
    }
    self->size = (self->size - 1);
    ax_i64 curr = idx;
    ax_i64 next = ((idx + 1) % self->cap);
    ax_bool shift_loop = AX_TRUE;
    while (shift_loop) {
        if ((!(((ax_bool*)(self->occupied))[next]))) {
            (((ax_bool*)(self->occupied))[curr]) = AX_FALSE;
            shift_loop = AX_FALSE;
        } else {
            {
                ax_u64 next_h = (((ax_u64*)(self->hashes))[next]);
                ax_i64 next_dib = (((next - ((ax_i64)((next_h % ((ax_u64)(self->cap)))))) + self->cap) % self->cap);
                if ((next_dib == ((ax_i64)(0)))) {
                    (((ax_bool*)(self->occupied))[curr]) = AX_FALSE;
                    shift_loop = AX_FALSE;
                } else {
                    {
                        (((ax_i64*)(self->keys))[curr]) = (((ax_i64*)(self->keys))[next]);
                        (((ax_bool*)(self->values))[curr]) = (((ax_bool*)(self->values))[next]);
                        (((ax_u64*)(self->hashes))[curr]) = next_h;
                        (((ax_bool*)(self->occupied))[curr]) = AX_TRUE;
                        curr = next;
                        next = ((next + 1) % self->cap);
                    }
                }
            }
        }
    }
    return AX_TRUE;
}

void ax__AX_std_HashSet__i64_destroy(struct ax__AX_std_HashSet__i64* self) {
    ax__AX_std_HashMap__i64__bool_destroy(&(self->map));
}

void ax__AX_std_HashMap__i64__bool_destroy(struct ax__AX_std_HashMap__i64__bool* self) {
    if ((self->keys != ((ax_i64*)(NULL)))) {
        ax_free(((ax_u8*)(self->keys)));
        ax_free(((ax_u8*)(self->values)));
        ax_free(((ax_u8*)(self->hashes)));
        ax_free(((ax_u8*)(self->occupied)));
        self->keys = ((ax_i64*)(NULL));
        self->values = ((ax_bool*)(NULL));
        self->hashes = ((ax_u64*)(NULL));
        self->occupied = ((ax_bool*)(NULL));
    }
    self->size = 0;
    self->cap = 0;
}

/* Entry point wrapper */
ax_i32 ax_main(void) {
    return ax_main_usr();
}

// Linker stubs for standalone bootstrap tests
ax_i32 ax_ax_driver_load_module(void* mod, struct ax_SymbolTable* st, void* tt) { return 0; }
ax_bool ax_std_string_starts_with(ax_string s, ax_string prefix) { return 0; }
ax_bool ax_std_string_ends_with(ax_string s, ax_string suffix) { return 0; }
ax_bool ax_std_string_contains(ax_string s, ax_string sub) { return 0; }
ax_i64 ax_std_string_char_count(ax_string s) { return 0; }
ax_string ax_std_string_trim(ax_string s) { return s; }
ax_string ax_std_string_to_upper(ax_string s) { return s; }
ax_string ax_std_string_to_lower(ax_string s) { return s; }
