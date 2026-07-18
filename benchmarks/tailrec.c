#include <stdint.h>
static int64_t sumto(int64_t n, int64_t acc){ if(n==0) return acc; return sumto(n-1, acc+n); }
int main(void){ int64_t r=0; for(int64_t i=0;i<40000;i++) r+=sumto(5000,0); return (int)(r%1000); }
