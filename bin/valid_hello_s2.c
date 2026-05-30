// Generated automatically by AXIOM AirCGen Stage 1
#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

#define r_0 0

// Forward Declarations

// Struct Definitions
// Function Prototypes
ax_i32 ax_main(void);

// Function Definitions
ax_i32 ax_main(void) {
    ax_i32 r_1 = {0};
    ax_i32 r_2 = {0};
    ax_i32 r_3 = {0};
    ax_bool r_4 = {0};
    ax_string r_5 = {0};
    ax_string r_6 = {0};
    ax_i32 r_8 = {0};
    ax_i32 r_9 = {0};
    ax_i32 r_10 = {0};
block_0: ;
    r_1 = 0;
    r_2 = 30;
    r_3 = r_1;
    goto block_1;
block_1: ;
    r_4 = r_3 < r_2;
    if (r_4) goto block_2; else goto block_3;
block_2: ;
    r_5 = AX_STR("Hello, world!");
    r_6 = r_5;
    ax_println_str(r_6);
    r_8 = 1;
    r_9 = r_3 + r_8;
    r_3 = r_9;
    goto block_1;
block_3: ;
    r_10 = 0;
    return r_10;
}

