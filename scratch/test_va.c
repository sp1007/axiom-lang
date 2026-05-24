#include <windows.h>
#include <stdio.h>

int main() {
    LPVOID addr = (LPVOID)0x50000000;
    LPVOID p = VirtualAlloc(addr, 4096, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
    if (p == NULL) {
        printf("VirtualAlloc failed. Error code: %lu\n", GetLastError());
        return 1;
    }
    printf("VirtualAlloc succeeded! Pointer: %p\n", p);
    return 0;
}
