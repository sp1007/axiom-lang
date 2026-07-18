#include <stdint.h>
typedef struct { int64_t lo, hi; } Bounds;
static int64_t clampb(Bounds b, int64_t x){ if(x<b.lo) return b.lo; if(x>b.hi) return b.hi; return x; }
int main(void){ Bounds bd={10,150}; int64_t acc=0; for(int64_t i=0;i<200000000;i++) acc+=clampb(bd,(i%250)-50); return (int)(acc%1000); }
