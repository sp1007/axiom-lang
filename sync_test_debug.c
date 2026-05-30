#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax__AX_std_Atomic__bool;
struct ax__AX_std_Atomic__i32;
struct ax__AX_std_Mutex__i32;
struct ax__AX_std_MutexGuard__i32;
struct ax__AX_std_Option___AX_std_MutexGuard__i32;
struct ax__AX_std_RwLock__i32;
struct ax__AX_std_ReadGuard__i32;
struct ax__AX_std_WriteGuard__i32;

/* Type definitions */
struct ax__AX_std_Atomic__bool {
    ax_bool value;
};

struct ax__AX_std_Atomic__i32 {
    ax_i32 value;
};

struct ax__AX_std_Mutex__i32 {
    ax_i32 value;
    struct ax__AX_std_Atomic__bool locked;
};

struct ax__AX_std_MutexGuard__i32 {
    struct ax__AX_std_Mutex__i32* mutex;
};

enum ax__AX_std_Option___AX_std_MutexGuard__i32_tag {
    ax__AX_std_Option___AX_std_MutexGuard__i32_Some = 0,
    ax__AX_std_Option___AX_std_MutexGuard__i32_None = 1,
};

struct ax__AX_std_Option___AX_std_MutexGuard__i32 {
    enum ax__AX_std_Option___AX_std_MutexGuard__i32_tag tag;
    union {
        struct ax__AX_std_MutexGuard__i32 Some;
    } data;
};

static inline struct ax__AX_std_Option___AX_std_MutexGuard__i32 ax__AX_std_Option___AX_std_MutexGuard__i32_some(struct ax__AX_std_MutexGuard__i32 value) {
    struct ax__AX_std_Option___AX_std_MutexGuard__i32 _result;
    _result.tag = ax__AX_std_Option___AX_std_MutexGuard__i32_Some;
    _result.data.Some = value;
    return _result;
}

static inline struct ax__AX_std_Option___AX_std_MutexGuard__i32 ax__AX_std_Option___AX_std_MutexGuard__i32_none(void) {
    struct ax__AX_std_Option___AX_std_MutexGuard__i32 _result;
    _result.tag = ax__AX_std_Option___AX_std_MutexGuard__i32_None;
    return _result;
}

struct ax__AX_std_RwLock__i32 {
    ax_i32 value;
    struct ax__AX_std_Atomic__i32 readers;
    struct ax__AX_std_Atomic__bool writer;
};

struct ax__AX_std_ReadGuard__i32 {
    struct ax__AX_std_RwLock__i32* lock;
};

struct ax__AX_std_WriteGuard__i32 {
    struct ax__AX_std_RwLock__i32* lock;
};


/* Function prototypes */
ax_bool ax_sum_layout_is_pointer(void);
ax_i64 syscall(ax_u64 num, ax_u64 a1, ax_u64 a2, ax_u64 a3, ax_u64 a4, ax_u64 a5, ax_u64 a6);
static void ax_test_print_str(ax_string s);
static void ax_test_atomic_operations(void);
static void ax_test_mutex_operations(void);
static void ax_test_rwlock_operations(void);
ax_i32 ax_main_usr(void);
struct ax__AX_std_Atomic__i32 ax_AX_std_atomic_new__i32(ax_i32 value);
ax_i32 ax__AX_std_Atomic__i32_load(struct ax__AX_std_Atomic__i32* self);
void ax__AX_std_Atomic__i32_store(struct ax__AX_std_Atomic__i32* self, ax_i32 value);
ax_i32 ax__AX_std_Atomic__i32_swap(struct ax__AX_std_Atomic__i32* self, ax_i32 value);
ax_bool ax__AX_std_Atomic__i32_compare_and_swap(struct ax__AX_std_Atomic__i32* self, ax_i32 expected, ax_i32 desired);
struct ax__AX_std_Mutex__i32 ax_AX_std_mutex_new__i32(ax_i32 value);
struct ax__AX_std_Atomic__bool ax_AX_std_atomic_new__bool(ax_bool value);
struct ax__AX_std_MutexGuard__i32 ax__AX_std_Mutex__i32_lock(struct ax__AX_std_Mutex__i32* self);
ax_bool ax__AX_std_Atomic__bool_compare_and_swap(struct ax__AX_std_Atomic__bool* self, ax_bool expected, ax_bool desired);
ax_i32 ax__AX_std_MutexGuard__i32_get(struct ax__AX_std_MutexGuard__i32 self);
void ax__AX_std_MutexGuard__i32_set(struct ax__AX_std_MutexGuard__i32 self, ax_i32 value);
void ax__AX_std_MutexGuard__i32_unlock(struct ax__AX_std_MutexGuard__i32 self);
void ax__AX_std_Atomic__bool_store(struct ax__AX_std_Atomic__bool* self, ax_bool value);
struct ax__AX_std_Option___AX_std_MutexGuard__i32 ax__AX_std_Mutex__i32_try_lock(struct ax__AX_std_Mutex__i32* self);
struct ax__AX_std_RwLock__i32 ax_AX_std_rwlock_new__i32(ax_i32 value);
struct ax__AX_std_ReadGuard__i32 ax__AX_std_RwLock__i32_read(struct ax__AX_std_RwLock__i32* self);
ax_bool ax__AX_std_Atomic__bool_load(struct ax__AX_std_Atomic__bool* self);
ax_i32 ax__AX_std_ReadGuard__i32_get(struct ax__AX_std_ReadGuard__i32 self);
void ax__AX_std_ReadGuard__i32_unlock(struct ax__AX_std_ReadGuard__i32 self);
struct ax__AX_std_WriteGuard__i32 ax__AX_std_RwLock__i32_write(struct ax__AX_std_RwLock__i32* self);
ax_i32 ax__AX_std_WriteGuard__i32_get(struct ax__AX_std_WriteGuard__i32 self);
void ax__AX_std_WriteGuard__i32_set(struct ax__AX_std_WriteGuard__i32 self, ax_i32 value);
void ax__AX_std_WriteGuard__i32_unlock(struct ax__AX_std_WriteGuard__i32 self);
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

static void ax_test_atomic_operations(void) {
    struct ax__AX_std_Atomic__i32 at = ax_AX_std_atomic_new__i32(((ax_i32)(10)));
    ax_assert_axiom((ax__AX_std_Atomic__i32_load(&(at)) == 10), AX_STR("(ax__AX_std_Atomic__i32_load(&(at)) == 10)"));
    ax__AX_std_Atomic__i32_store(&(at), ((ax_i32)(20)));
    ax_assert_axiom((ax__AX_std_Atomic__i32_load(&(at)) == 20), AX_STR("(ax__AX_std_Atomic__i32_load(&(at)) == 20)"));
    ax_i32 old = ax__AX_std_Atomic__i32_swap(&(at), ((ax_i32)(30)));
    ax_assert_axiom((old == 20), AX_STR("(old == 20)"));
    ax_assert_axiom((ax__AX_std_Atomic__i32_load(&(at)) == 30), AX_STR("(ax__AX_std_Atomic__i32_load(&(at)) == 30)"));
    ax_bool swapped = ax__AX_std_Atomic__i32_compare_and_swap(&(at), ((ax_i32)(30)), ((ax_i32)(40)));
    ax_assert_axiom((swapped == AX_TRUE), AX_STR("(swapped == AX_TRUE)"));
    ax_assert_axiom((ax__AX_std_Atomic__i32_load(&(at)) == 40), AX_STR("(ax__AX_std_Atomic__i32_load(&(at)) == 40)"));
    ax_bool not_swapped = ax__AX_std_Atomic__i32_compare_and_swap(&(at), ((ax_i32)(30)), ((ax_i32)(50)));
    ax_assert_axiom((not_swapped == AX_FALSE), AX_STR("(not_swapped == AX_FALSE)"));
    ax_assert_axiom((ax__AX_std_Atomic__i32_load(&(at)) == 40), AX_STR("(ax__AX_std_Atomic__i32_load(&(at)) == 40)"));
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_atomic_operations\n", .len=31});
}

static void ax_test_mutex_operations(void) {
    struct ax__AX_std_Mutex__i32 m = ax_AX_std_mutex_new__i32(((ax_i32)(100)));
    struct ax__AX_std_MutexGuard__i32 guard = ax__AX_std_Mutex__i32_lock(&(m));
    ax_assert_axiom((ax__AX_std_MutexGuard__i32_get(guard) == 100), AX_STR("(ax__AX_std_MutexGuard__i32_get(guard) == 100)"));
    ax__AX_std_MutexGuard__i32_set(guard, ((ax_i32)(200)));
    ax_assert_axiom((ax__AX_std_MutexGuard__i32_get(guard) == 200), AX_STR("(ax__AX_std_MutexGuard__i32_get(guard) == 200)"));
    ax__AX_std_MutexGuard__i32_unlock(guard);
    struct ax__AX_std_Option___AX_std_MutexGuard__i32 guard_opt = ax__AX_std_Mutex__i32_try_lock(&(m));
    ax_assert_axiom((guard_opt.tag == ax__AX_std_Option___AX_std_MutexGuard__i32_Some), AX_STR("(guard_opt.tag == ax__AX_std_Option___AX_std_MutexGuard__i32_Some)"));
    struct ax__AX_std_MutexGuard__i32 g = guard_opt.data.Some;
    ax_assert_axiom((ax__AX_std_MutexGuard__i32_get(g) == 200), AX_STR("(ax__AX_std_MutexGuard__i32_get(g) == 200)"));
    ax__AX_std_MutexGuard__i32_unlock(g);
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_mutex_operations\n", .len=30});
}

static void ax_test_rwlock_operations(void) {
    struct ax__AX_std_RwLock__i32 l = ax_AX_std_rwlock_new__i32(((ax_i32)(500)));
    struct ax__AX_std_ReadGuard__i32 r1 = ax__AX_std_RwLock__i32_read(&(l));
    ax_assert_axiom((ax__AX_std_ReadGuard__i32_get(r1) == 500), AX_STR("(ax__AX_std_ReadGuard__i32_get(r1) == 500)"));
    struct ax__AX_std_ReadGuard__i32 r2 = ax__AX_std_RwLock__i32_read(&(l));
    ax_assert_axiom((ax__AX_std_ReadGuard__i32_get(r2) == 500), AX_STR("(ax__AX_std_ReadGuard__i32_get(r2) == 500)"));
    ax__AX_std_ReadGuard__i32_unlock(r1);
    ax__AX_std_ReadGuard__i32_unlock(r2);
    struct ax__AX_std_WriteGuard__i32 w = ax__AX_std_RwLock__i32_write(&(l));
    ax_assert_axiom((ax__AX_std_WriteGuard__i32_get(w) == 500), AX_STR("(ax__AX_std_WriteGuard__i32_get(w) == 500)"));
    ax__AX_std_WriteGuard__i32_set(w, ((ax_i32)(1000)));
    ax_assert_axiom((ax__AX_std_WriteGuard__i32_get(w) == 1000), AX_STR("(ax__AX_std_WriteGuard__i32_get(w) == 1000)"));
    ax__AX_std_WriteGuard__i32_unlock(w);
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_rwlock_operations\n", .len=31});
}

ax_i32 ax_main_usr(void) {
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"Running AXIOM-native synchronization unit tests...\n", .len=51});
    ax_test_atomic_operations();
    ax_test_mutex_operations();
    ax_test_rwlock_operations();
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"All AXIOM-native synchronization tests passed!\n", .len=47});
    return 0;
}

struct ax__AX_std_Atomic__i32 ax_AX_std_atomic_new__i32(ax_i32 value) {
    return ((struct ax__AX_std_Atomic__i32){.value=value});
}

ax_i32 ax__AX_std_Atomic__i32_load(struct ax__AX_std_Atomic__i32* self) {
    return __atomic_load_n(&(self->value), __ATOMIC_SEQ_CST);
}

void ax__AX_std_Atomic__i32_store(struct ax__AX_std_Atomic__i32* self, ax_i32 value) {
    __atomic_store_n(&(self->value), value, __ATOMIC_SEQ_CST);
}

ax_i32 ax__AX_std_Atomic__i32_swap(struct ax__AX_std_Atomic__i32* self, ax_i32 value) {
    return __atomic_exchange_n(&(self->value), value, __ATOMIC_SEQ_CST);
}

ax_bool ax__AX_std_Atomic__i32_compare_and_swap(struct ax__AX_std_Atomic__i32* self, ax_i32 expected, ax_i32 desired) {
    return __sync_bool_compare_and_swap(&(self->value), expected, desired);
}

struct ax__AX_std_Mutex__i32 ax_AX_std_mutex_new__i32(ax_i32 value) {
    return ((struct ax__AX_std_Mutex__i32){.value=value, .locked=ax_AX_std_atomic_new__bool(AX_FALSE)});
}

struct ax__AX_std_Atomic__bool ax_AX_std_atomic_new__bool(ax_bool value) {
    return ((struct ax__AX_std_Atomic__bool){.value=value});
}

struct ax__AX_std_MutexGuard__i32 ax__AX_std_Mutex__i32_lock(struct ax__AX_std_Mutex__i32* self) {
    while ((!ax__AX_std_Atomic__bool_compare_and_swap(&(self->locked), AX_FALSE, AX_TRUE))) {
        ax_i32 dummy = 0;
        (void)dummy;
    }
    return ((struct ax__AX_std_MutexGuard__i32){.mutex=self});
}

ax_bool ax__AX_std_Atomic__bool_compare_and_swap(struct ax__AX_std_Atomic__bool* self, ax_bool expected, ax_bool desired) {
    return __sync_bool_compare_and_swap(&(self->value), expected, desired);
}

ax_i32 ax__AX_std_MutexGuard__i32_get(struct ax__AX_std_MutexGuard__i32 self) {
    return self.mutex->value;
}

void ax__AX_std_MutexGuard__i32_set(struct ax__AX_std_MutexGuard__i32 self, ax_i32 value) {
    self.mutex->value = value;
}

void ax__AX_std_MutexGuard__i32_unlock(struct ax__AX_std_MutexGuard__i32 self) {
    ax__AX_std_Atomic__bool_store(&(self.mutex->locked), AX_FALSE);
}

void ax__AX_std_Atomic__bool_store(struct ax__AX_std_Atomic__bool* self, ax_bool value) {
    __atomic_store_n(&(self->value), value, __ATOMIC_SEQ_CST);
}

struct ax__AX_std_Option___AX_std_MutexGuard__i32 ax__AX_std_Mutex__i32_try_lock(struct ax__AX_std_Mutex__i32* self) {
    if (ax__AX_std_Atomic__bool_compare_and_swap(&(self->locked), AX_FALSE, AX_TRUE)) {
        return ax__AX_std_Option___AX_std_MutexGuard__i32_some(((struct ax__AX_std_MutexGuard__i32){.mutex=self}));
    } else {
        {
            return ax__AX_std_Option___AX_std_MutexGuard__i32_none();
        }
    }
}

struct ax__AX_std_RwLock__i32 ax_AX_std_rwlock_new__i32(ax_i32 value) {
    return ((struct ax__AX_std_RwLock__i32){.value=value, .readers=ax_AX_std_atomic_new__i32(0), .writer=ax_AX_std_atomic_new__bool(AX_FALSE)});
}

struct ax__AX_std_ReadGuard__i32 ax__AX_std_RwLock__i32_read(struct ax__AX_std_RwLock__i32* self) {
    while (ax__AX_std_Atomic__bool_load(&(self->writer))) {
        ax_i32 dummy = 0;
        (void)dummy;
    }
    ax__AX_std_Atomic__i32_store(&(self->readers), (ax__AX_std_Atomic__i32_load(&(self->readers)) + 1));
    return ((struct ax__AX_std_ReadGuard__i32){.lock=self});
}

ax_bool ax__AX_std_Atomic__bool_load(struct ax__AX_std_Atomic__bool* self) {
    return __atomic_load_n(&(self->value), __ATOMIC_SEQ_CST);
}

ax_i32 ax__AX_std_ReadGuard__i32_get(struct ax__AX_std_ReadGuard__i32 self) {
    return self.lock->value;
}

void ax__AX_std_ReadGuard__i32_unlock(struct ax__AX_std_ReadGuard__i32 self) {
    ax_i32 r = ax__AX_std_Atomic__i32_load(&(self.lock->readers));
    ax__AX_std_Atomic__i32_store(&(self.lock->readers), (r - 1));
}

struct ax__AX_std_WriteGuard__i32 ax__AX_std_RwLock__i32_write(struct ax__AX_std_RwLock__i32* self) {
    while ((!ax__AX_std_Atomic__bool_compare_and_swap(&(self->writer), AX_FALSE, AX_TRUE))) {
        ax_i32 dummy = 0;
        (void)dummy;
    }
    while ((ax__AX_std_Atomic__i32_load(&(self->readers)) > 0)) {
        ax_i32 dummy = 0;
        (void)dummy;
    }
    return ((struct ax__AX_std_WriteGuard__i32){.lock=self});
}

ax_i32 ax__AX_std_WriteGuard__i32_get(struct ax__AX_std_WriteGuard__i32 self) {
    return self.lock->value;
}

void ax__AX_std_WriteGuard__i32_set(struct ax__AX_std_WriteGuard__i32 self, ax_i32 value) {
    self.lock->value = value;
}

void ax__AX_std_WriteGuard__i32_unlock(struct ax__AX_std_WriteGuard__i32 self) {
    ax__AX_std_Atomic__bool_store(&(self.lock->writer), AX_FALSE);
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
