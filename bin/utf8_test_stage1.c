// Generated automatically by AXIOM AirCGen Stage 1
#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

#define r_0 0

// Forward Declarations

// Struct Definitions
// Function Prototypes
ax_i32 ax_main();

// Function Definitions
ax_i32 ax_main() {
    ax_string r_1 = {0};
    ax_char r_3 = {0};
    ax_char r_4 = {0};
    ax_char r_5 = {0};
    ax_string r_6 = {0};
block_0: ;
    r_1 = AX_STR("Xin chào thế giới AXIOM! Tiếng Việt có các ký tự: áàảãạđêôư");
    ax_println_str(r_1);
    r_3 = 196;
    r_4 = 195;
    r_5 = 65;
    r_6 = AX_STR("Đã parse thành công các ký tự UTF-8!");
    ax_println_str(r_6);
    return (ax_i32){0};
}

