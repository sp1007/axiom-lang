// Generated automatically by AXIOM AirCGen Stage 1
#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

#define r_0 0

// Forward Declarations

// Struct Definitions
// Function Prototypes
void ax_my_actor_handler(void* p_0, void* p_1, ax_u32 p_2);
ax_i32 ax_main();

// Function Definitions
void ax_my_actor_handler(void* p_0, void* p_1, ax_u32 p_2) {
    void* r_1 = {0};
    void* r_2 = {0};
    ax_u32 r_3 = {0};
    ax_i32 r_4 = {0};
    ax_i32 r_5 = {0};
    ax_i32 r_6 = {0};
    ax_i32 r_7 = {0};
    r_1 = p_0;
    r_2 = p_1;
    r_3 = p_2;
block_0: ;
    r_1 = r_1;
    r_2 = r_2;
    r_3 = r_3;
    r_4 = 72;
    r_5 = putchar(r_4);
    r_6 = 10;
    r_7 = putchar(r_6);
    return;
}

ax_i32 ax_main() {
    void* r_1 = {0};
    ax_i32 r_2 = {0};
    ax_i32 r_3 = {0};
    ax_i32 r_4 = {0};
    ax_i32 r_5 = {0};
    ax_i32 r_6 = {0};
block_0: ;
    r_1 = (void*)ax_actor_spawn((AxHandlerFn)r_0, NULL, 0);
    r_2 = 83;
    r_3 = putchar(r_2);
    r_4 = 10;
    r_5 = putchar(r_4);
    r_6 = 0;
    return r_6;
}

