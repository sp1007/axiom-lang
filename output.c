// Generated automatically by AXIOM AirCGen Stage 1
#define AXIOM_SUMLAYOUT_POINTER
#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

extern ax_i64 syscall(ax_u64 num, ...);

#define ax_Result_ok(v) ({ \
    void* _ptr = ax_alloc(sizeof(v)); \
    *(__typeof__(v)*)_ptr = (v); \
    _ptr; \
})

#define ax_Result_err(v) ({ \
    void* _ptr = ax_alloc(sizeof(v)); \
    *(__typeof__(v)*)_ptr = (v); \
    (void*)((ax_u64)_ptr | 1ULL); \
})

#define ax_Option_some(v) ({ \
    void* _ptr = ax_alloc(sizeof(v)); \
    *(__typeof__(v)*)_ptr = (v); \
    _ptr; \
})

#define ax_Option_none() ((void*)0)

#define r_0 0

// Forward Declarations

// Struct Definitions
// Function Prototypes
ax_i32 ax_main_usr(void);

// Function Definitions
ax_i32 ax_main_usr(void) {
    ax_i32 r_1 = {0};
block_0: ;
    r_1 = 42;
    return r_1;
}


/* Entry point wrapper */
ax_i32 ax_main(void) {
    ax_main_usr();
    return 0;
}
