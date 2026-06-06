// mini_link.c — call AXIOM linker directly from axiom_temp.obj
// Usage: mini_linker.exe <input.obj> <output.exe>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <windows.h>

// AXIOM string: {ptr, len}
typedef struct { const char* ptr; int64_t len; } ax_str;

// LinkerStrVec: {data, len, capacity}
typedef struct { ax_str* data; int64_t len; int64_t capacity; } LinkerStrVec;

// AxiomLinker: {input_files: LinkerStrVec, output_path: ax_str} = 40 bytes
typedef struct { LinkerStrVec input_files; ax_str output_path; } AxiomLinker;

// VEH to catch crash address
static LONG CALLBACK crash_handler(EXCEPTION_POINTERS* ep) {
    printf("!!! EXCEPTION 0x%08lX at RIP=0x%llx, addr=0x%llx\n",
        ep->ExceptionRecord->ExceptionCode,
        (unsigned long long)ep->ContextRecord->Rip,
        ep->ExceptionRecord->NumberParameters >= 2
            ? (unsigned long long)ep->ExceptionRecord->ExceptionInformation[1] : 0);
    fflush(stdout);
    return EXCEPTION_CONTINUE_SEARCH;
}

// ax_runtime_init(): must be called before any AXIOM heap/alloc functions
extern void ax_runtime_init(void);
// debug: check global state directly
extern void* ax_get_global_state_internal(void);
extern void* ax_alloc(int64_t size);
// test: call ax_str_eq to verify import works
extern int8_t ax_str_eq(ax_str* a, ax_str* b);

// ax_new_axiom_linker(): returns AxiomLinker* (heap-allocated)
extern AxiomLinker* ax_new_axiom_linker(void);

// ax_AxiomLinker_axiom_linker_add_input(self: ptr, file: ptr[str])
// RCX=self, RDX=&file (str passed by pointer since sizeof(str)==16)
extern void ax_AxiomLinker_axiom_linker_add_input(AxiomLinker* self, ax_str* file);

// ax_AxiomLinker_axiom_linker_link(self: ptr) -> bool
extern int8_t ax_AxiomLinker_axiom_linker_link(AxiomLinker* self);

int main(int argc, char** argv) {
    if (argc < 3) {
        fprintf(stderr, "Usage: mini_linker.exe <input.obj> <output.exe>\n");
        return 1;
    }
    const char* obj_path = argv[1];
    const char* out_path = argv[2];

    ax_str s1 = {"hello", 5};
    ax_str s2 = {"hello", 5};
    ax_str s3 = {"world", 5};
    printf("Testing ax_str_eq before init...\n"); fflush(stdout);
    int eq1 = ax_str_eq(&s1, &s2);
    int eq2 = ax_str_eq(&s1, &s3);
    printf("ax_str_eq(hello,hello)=%d ax_str_eq(hello,world)=%d\n", eq1, eq2); fflush(stdout);

    printf("Calling ax_runtime_init...\n"); fflush(stdout);
    ax_runtime_init();
    printf("ax_runtime_init done\n"); fflush(stdout);
    void* gs = ax_get_global_state_internal();
    printf("ax_get_global_state_internal = %p\n", gs); fflush(stdout);
    void* test_alloc = ax_alloc(64);
    printf("ax_alloc(64) = %p\n", test_alloc); fflush(stdout);
    void* big_alloc = ax_alloc(1790542);
    printf("ax_alloc(1790542) = %p\n", big_alloc); fflush(stdout);

    printf("Calling ax_new_axiom_linker...\n"); fflush(stdout);
    AxiomLinker* linker = ax_new_axiom_linker();
    printf("ax_new_axiom_linker done: linker=%p\n", (void*)linker); fflush(stdout);

    ax_str obj_str;
    obj_str.ptr = obj_path;
    obj_str.len = (int64_t)strlen(obj_path);
    printf("Calling axiom_linker_add_input with '%s'...\n", obj_path); fflush(stdout);
    ax_AxiomLinker_axiom_linker_add_input(linker, &obj_str);
    printf("axiom_linker_add_input done\n"); fflush(stdout);

    linker->output_path.ptr = out_path;
    linker->output_path.len = (int64_t)strlen(out_path);
    printf("output_path set to '%s'\n", out_path); fflush(stdout);

    // Commit entire reserved stack (2MB) to avoid guard page issues
    // (axiom_temp.obj functions lack .pdata, so guard page exceptions crash)
    {
        HANDLE thread = GetCurrentThread();
        ULONG guarantee = 0;
        SetThreadStackGuarantee(&guarantee);
        printf("SetThreadStackGuarantee: old=%lu\n", guarantee); fflush(stdout);

        // Actually commit all stack pages by VirtualAlloc-touching them
        ULONG_PTR stackLow, stackHigh;
        GetCurrentThreadStackLimits(&stackLow, &stackHigh);
        printf("Stack: lo=0x%llx hi=0x%llx size=%lluKB\n",
            (unsigned long long)stackLow, (unsigned long long)stackHigh,
            (unsigned long long)((stackHigh - stackLow) / 1024)); fflush(stdout);
        if (stackLow > 0 && stackHigh > stackLow) {
            // Commit all stack pages
            VirtualAlloc((LPVOID)stackLow, stackHigh - stackLow,
                MEM_COMMIT, PAGE_READWRITE);
            printf("VirtualAlloc committed %llu KB of stack\n",
                (unsigned long long)((stackHigh - stackLow)/1024)); fflush(stdout);
        }
    }
    // AXIOM codegen bug: native codegen writes through str.ptr (string literal in .rdata)
    // instead of writing to str.ptr (the field itself). Make .rdata writable to allow this.
    {
        HMODULE hmod = GetModuleHandleW(NULL);
        IMAGE_DOS_HEADER* dos = (IMAGE_DOS_HEADER*)hmod;
        IMAGE_NT_HEADERS* nt = (IMAGE_NT_HEADERS*)((char*)hmod + dos->e_lfanew);
        IMAGE_SECTION_HEADER* sect = (IMAGE_SECTION_HEADER*)
            ((char*)&nt->OptionalHeader + nt->FileHeader.SizeOfOptionalHeader);
        int nsections = nt->FileHeader.NumberOfSections;
        for (int s = 0; s < nsections; s++) {
            char name[9] = {0}; memcpy(name, sect[s].Name, 8);
            void* va = (char*)hmod + sect[s].VirtualAddress;
            DWORD sz = sect[s].Misc.VirtualSize;
            // Make all sections writable (text too, for potential JIT-like fixups)
            DWORD old;
            VirtualProtect(va, sz, PAGE_EXECUTE_READWRITE, &old);
            printf("Made %s writable (0x%llx, 0x%lx bytes)\n", name, (unsigned long long)va, sz);
        }
        fflush(stdout);
    }
    AddVectoredExceptionHandler(1, crash_handler);
    printf("Calling axiom_linker_link...\n"); fflush(stdout);
    int8_t ok = ax_AxiomLinker_axiom_linker_link(linker);
    printf("axiom_linker_link done: ok=%d\n", ok); fflush(stdout);
    if (ok) {
        printf("Link OK: %s\n", out_path);
        return 0;
    } else {
        fprintf(stderr, "Link FAILED\n");
        return 1;
    }
}
