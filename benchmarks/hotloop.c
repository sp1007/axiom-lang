/* M6 perf benchmark — C reference for benchmarks/hotloop.ax (clang -O2 baseline). */
#include <stdint.h>
int64_t sq(int64_t x) { return x * x; }
int main(void) {
    int64_t s = 0;
    for (int64_t i = 0; i < 60000000; i++) s = s + sq(i);
    return (int)(s % 1000);
}
