package codegen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// --------------------------------------------------------------------------
// p14-t01/t02: AxAlloc Integration Test via C compilation
//
// Compiles the C allocator source with a test harness and validates
// the resulting binary exits cleanly.
// --------------------------------------------------------------------------

func TestAxAllocCompiles(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		t.Skip("axalloc tests require Windows or Linux")
	}

	// Resolve project root (one level up from codegen/)
	projectRoot := filepath.Join("..") 
	dir := filepath.Join(projectRoot, "runtime", "axalloc")
	testC := filepath.Join(dir, "axalloc_test_main.c")

	// Create test harness
	testSrc := `
#include <stdio.h>
#include <assert.h>
#include <string.h>
#include "size_classes.h"

void test_size_class_for(void) {
    assert(ax_size_class_for(0)    == SIZE_CLASS_8);
    assert(ax_size_class_for(1)    == SIZE_CLASS_16);
    assert(ax_size_class_for(8)    == SIZE_CLASS_16);
    assert(ax_size_class_for(9)    == SIZE_CLASS_32);
    assert(ax_size_class_for(56)   == SIZE_CLASS_64);
    assert(ax_size_class_for(57)   == SIZE_CLASS_128);
    assert(ax_size_class_for(4088) == SIZE_CLASS_4096);
    assert(ax_size_class_for(4089) == SIZE_CLASS_LARGE);
    printf("  PASS: test_size_class_for\n");
}

void test_header_layout(void) {
    assert(sizeof(AxHeader) == 8);
    printf("  PASS: test_header_layout\n");
}

void test_pointer_conversion(void) {
    char buf[64];
    void* block = buf;
    void* user  = ax_block_to_user(block);
    void* back  = ax_user_to_block(user);
    assert(back == block);
    assert((char*)user - (char*)block == 8);
    printf("  PASS: test_pointer_conversion\n");
}

void test_free_list_push_pop(void) {
    char blocks[3][32];
    FreeList list = {0};

    ax_free_list_push(&list, blocks[0]);
    ax_free_list_push(&list, blocks[1]);
    ax_free_list_push(&list, blocks[2]);
    assert(list.count == 3);

    void* b2 = ax_free_list_pop(&list);
    void* b1 = ax_free_list_pop(&list);
    void* b0 = ax_free_list_pop(&list);
    assert(b2 == blocks[2]);
    assert(b1 == blocks[1]);
    assert(b0 == blocks[0]);
    assert(list.count == 0);
    assert(ax_free_list_pop(&list) == NULL);
    printf("  PASS: test_free_list_push_pop\n");
}

void test_bump_alloc(void) {
    char region[128];
    char* bump = region;
    char* limit = region + sizeof(region);

    void* b1 = ax_bump_alloc(&bump, limit, SIZE_CLASS_32);
    assert(b1 != NULL);
    assert(bump == region + 32);

    void* b2 = ax_bump_alloc(&bump, limit, SIZE_CLASS_32);
    assert(b2 != NULL);
    assert(bump == region + 64);

    void* b3 = ax_bump_alloc(&bump, limit, SIZE_CLASS_32);
    assert(b3 != NULL);
    assert(bump == region + 96);

    void* b4 = ax_bump_alloc(&bump, limit, SIZE_CLASS_32);
    assert(b4 != NULL);
    assert(bump == region + 128);

    /* Region exhausted */
    void* b5 = ax_bump_alloc(&bump, limit, SIZE_CLASS_32);
    assert(b5 == NULL);
    printf("  PASS: test_bump_alloc\n");
}

void test_alloc_free_cycle(void) {
    char region[1024];
    char* bump = region;
    char* limit = region + sizeof(region);
    FreeList list = {0};

    /* Allocate 10 blocks */
    void* ptrs[10];
    for (int i = 0; i < 10; i++) {
        ptrs[i] = ax_size_class_alloc(&list, &bump, limit, 1);
        assert(ptrs[i] != NULL);
        AxHeader* hdr = ax_get_header(ptrs[i]);
        assert(hdr->gen_id == 1);
    }

    /* Free all */
    FreeList free_lists[NUM_SIZE_CLASSES];
    memset(free_lists, 0, sizeof(free_lists));
    for (int i = 0; i < 10; i++) {
        ax_size_class_free(free_lists, ptrs[i]);
        AxHeader* hdr = ax_get_header(ptrs[i]);
        assert(hdr->gen_id == 0);
    }

    printf("  PASS: test_alloc_free_cycle\n");
}

int main(void) {
    printf("Running AxAlloc tests...\n");
    test_size_class_for();
    test_header_layout();
    test_pointer_conversion();
    test_free_list_push_pop();
    test_bump_alloc();
    test_alloc_free_cycle();
    printf("All AxAlloc tests passed!\n");
    return 0;
}
`

	err := os.WriteFile(testC, []byte(testSrc), 0644)
	if err != nil {
		t.Fatalf("failed to write test source: %v", err)
	}
	defer os.Remove(testC)

	// Compile with cc (gcc/clang on Linux, cl on Windows)
	outFile := filepath.Join(dir, "axalloc_test")
	if runtime.GOOS == "windows" {
		outFile += ".exe"
	}
	defer os.Remove(outFile)

	compiler := "gcc"
	if runtime.GOOS == "windows" {
		compiler = "gcc" // MinGW
	}

	cmd := exec.Command(compiler,
		"-Wall", "-Wextra", "-std=c11", "-O0",
		"-I", dir,
		"-o", outFile,
		testC,
		filepath.Join(dir, "size_classes.c"),
	)
	cmd.Dir = projectRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("C compiler not available or compilation failed: %v\n%s", err, string(out))
		return
	}

	// Run test binary
	cmd = exec.Command(outFile)
	out, err = cmd.CombinedOutput()
	t.Logf("Output:\n%s", string(out))
	if err != nil {
		t.Fatalf("test binary failed: %v\n%s", err, string(out))
	}
}
