/* M6 perf benchmark — C reference for benchmarks/fib.ax (clang -O2 baseline). */
#include <stdint.h>
int64_t fib(int64_t n) {
    if (n < 2) return n;
    return fib(n - 1) + fib(n - 2);
}
int main(void) {
    int64_t r = fib(42);
    return (int)(r % 1000);
}
