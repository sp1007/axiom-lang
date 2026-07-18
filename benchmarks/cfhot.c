#include <stdint.h>
static int64_t classify(int64_t x) {
    if (x < 0) return 0 - x;
    if (x > 100) return 100;
    return x;
}
int main(void) {
    int64_t acc = 0;
    for (int64_t i = 0; i < 200000000; i++)
        acc = acc + classify((i % 250) - 50);
    return (int)(acc % 1000);
}
