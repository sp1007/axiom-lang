package main

import (
	"fmt"
	"syscall"
)

func main() {
	check("kernel32.dll", []string{
		"GetStdHandle",
		"WriteFile",
		"VirtualAlloc",
		"GetLastError",
		"VirtualFree",
		"ExitProcess",
		"GetFileAttributesA",
		"CreateDirectoryA",
		"DeleteFileA",
		"RemoveDirectoryA",
		"MoveFileA",
		"CopyFileA",
		"GetCommandLineW",
		"GetEnvironmentVariableA",
	})

	check("ax_runtime.dll", []string{
		"ax_get_global_state_internal",
		"ax_actor_step",
		"ax_actor_is_running",
		"ax_actor_has_messages",
		"ax_actor_spawn",
		"ax_actor_ref_to_u64",
		"ax_time_now_ns",
		"ax_actor_stop",
		"ax_str_eq",
		"ax_str_parse_i64",
		"ax_str_parse_f64",
		"ax_println_str",
		"ax_sum_layout_is_pointer",
		"ax_panic",
		"ax_str_slice",
	})

	check("ucrtbase.dll", []string{
		"memset",
		"memcpy",
		"strlen",
		"printf",
		"fflush",
	})
}

func check(dllName string, symbols []string) {
	dll, err := syscall.LoadDLL(dllName)
	if err != nil {
		fmt.Printf("Error loading %s: %v\n", dllName, err)
		return
	}
	defer dll.Release()

	for _, sym := range symbols {
		_, err := dll.FindProc(sym)
		if err != nil {
			fmt.Printf("MISSING in %s: %s (error: %v)\n", dllName, sym, err)
		} else {
			fmt.Printf("Found in %s: %s\n", dllName, sym)
		}
	}
}
