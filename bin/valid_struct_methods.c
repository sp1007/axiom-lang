#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax_Counter;

/* Type definitions */
struct ax_Counter {
    ax_i32 value;
};

/* Function prototypes */
ax_i32 ax_main(void);


ax_i32 ax_main(void) {
    ax_i32 v = 42;
    return 0;
}
