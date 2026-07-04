#include <stdio.h>
#include <windows.h>
typedef int (*fn2)(int,int);
int main(void){
    HMODULE h = LoadLibraryA("axmath.dll");
    if(!h){ printf("LoadLibrary fail err=%lu\n",(unsigned long)GetLastError()); return 10; }
    fn2 add = (fn2)GetProcAddress(h,"ax_add");
    fn2 mul = (fn2)GetProcAddress(h,"ax_mul");
    if(!add||!mul){ printf("GetProcAddress fail add=%p mul=%p\n",(void*)add,(void*)mul); return 11; }
    int r1=add(40,2), r2=mul(6,7);
    printf("ax_add(40,2)=%d  ax_mul(6,7)=%d\n", r1, r2);
    return (r1==42 && r2==42) ? 0 : 12;
}
