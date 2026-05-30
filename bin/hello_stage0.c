#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Function prototypes */
ax_i32 ax_main_usr(void);


ax_i32 ax_main_usr(void) {
    for (ax_i32 i = 0; i < 30; i++) {
        ax_string message = (ax_string){.ptr=(const ax_u8*)"Hello, world! $i", .len=16};
        ax_println_str(message);
    }
    return 0;
}

/* Entry point wrapper */
ax_i32 ax_main(void) {
    return ax_main_usr();
}
