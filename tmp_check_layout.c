#include <windows.h>
#include <stdio.h>

typedef int (*fn_t)(void);

int main(void) {
    const char* paths[] = { "ax_runtime.dll", "bin\\ax_runtime.dll" };
    for (int i = 0; i < 2; i++) {
        HMODULE h = LoadLibraryA(paths[i]);
        if (!h) { printf("%s: load failed (%lu)\n", paths[i], GetLastError()); continue; }
        fn_t f = (fn_t)GetProcAddress(h, "ax_sum_layout_is_pointer");
        if (!f) { printf("%s: symbol not found\n", paths[i]); FreeLibrary(h); continue; }
        printf("%s: ax_sum_layout_is_pointer() = %d\n", paths[i], f());
        FreeLibrary(h);
    }
    return 0;
}
