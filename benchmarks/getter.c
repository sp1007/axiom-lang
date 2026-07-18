/* M6 perf benchmark — C reference for benchmarks/getter.ax (clang -O2 baseline). */
#include <stdint.h>
struct Vec2 { int64_t x, y; };
int64_t getx(struct Vec2 v) { return v.x; }
int main(void) {
    struct Vec2 v = {7, 3};
    int64_t s = 0;
    for (int64_t i = 0; i < 60000000; i++) s = s + getx(v);
    return (int)(s % 1000);
}
