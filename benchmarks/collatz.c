/* M6 perf benchmark — C reference for benchmarks/collatz.ax (clang -O2 baseline). */
#include <stdint.h>
int64_t collatz_len(int64_t n0) {
    int64_t n = n0;
    int64_t steps = 0;
    while (n > 1) {
        if (n % 2 == 0) n = n / 2;
        else n = 3 * n + 1;
        steps = steps + 1;
    }
    return steps;
}
int main(void) {
    int64_t total = 0;
    for (int64_t i = 1; i < 300000; i++) total = total + collatz_len(i);
    return (int)(total % 1000);
}
