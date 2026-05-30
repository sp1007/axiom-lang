// Generated automatically by AXIOM AirCGen Stage 1
#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

extern ax_i64 syscall(ax_u64 num, ...);

#define ax_cast_to_void_ptr(v) ({ union { __typeof__(v) _val; void* _ptr; } _u; _u._val = (v); _u._ptr; })
#define ax_Result_ok(v) ax_cast_to_void_ptr(v)
#define ax_Result_err(v) ax_cast_to_void_ptr(v)
#define ax_Option_some(v) ax_cast_to_void_ptr(v)
#define ax_Option_none() ((void*)0)

#define r_0 0

// Forward Declarations

// Struct Definitions
// Function Prototypes
void* ax_divide(ax_i32 p_0, ax_i32 p_1);
ax_i32 ax_main_usr(void);
ax_bool ax__AX_std_Option__i32__AX_std_is_some__i32(void* p_0);
ax_bool ax__AX_std_Result__i32__i32__AX_std_is_ok__i32__i32(void* p_0);

// Function Definitions
void* ax_divide(ax_i32 p_0, ax_i32 p_1) {
    ax_i32 r_1 = {0};
    ax_i32 r_2 = {0};
    void* r_3 = {0};
    r_1 = p_0;
    r_2 = p_1;
block_0: ;
    r_1 = r_1;
    r_2 = r_2;
    r_3 = ax_Result_ok(r_1);
    return r_3;
}

ax_i32 ax_main_usr(void) {
    ax_i32 r_1 = {0};
    ax_i32 r_2 = {0};
    void* r_3 = {0};
    void* r_4 = {0};
block_0: ;
    r_1 = 10;
    r_2 = 2;
    r_3 = ax_divide(r_1, r_2);
    r_4 = r_3;
    return (ax_i32){0};
}

ax_bool ax__AX_std_Option__i32__AX_std_is_some__i32(void* p_0) {
    void* r_1 = {0};
    r_1 = p_0;
block_0: ;
    r_1 = r_1;
    return (ax_bool){0};
}

ax_bool ax__AX_std_Result__i32__i32__AX_std_is_ok__i32__i32(void* p_0) {
    void* r_1 = {0};
    r_1 = p_0;
block_0: ;
    r_1 = r_1;
    return (ax_bool){0};
}


/* Entry point wrapper */
ax_i32 ax_main(void) {
    return ax_main_usr();
}
