package main

import (
	"debug/pe"
	"fmt"
	"os"
)

func main() {
	f, err := pe.Open("ax_runtime.dll")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var oh64 *pe.OptionalHeader64
	switch oh := f.OptionalHeader.(type) {
	case *pe.OptionalHeader64:
		oh64 = oh
	default:
		fmt.Println("Not PE32+")
		os.Exit(1)
	}

	exportDir := oh64.DataDirectory[0]
	fmt.Printf("Export Directory: RVA=0x%08X, Size=%d\n", exportDir.VirtualAddress, exportDir.Size)
	if exportDir.VirtualAddress == 0 {
		fmt.Println("No exports")
		os.Exit(0)
	}

	// Dump section containing export directory
	for _, sec := range f.Sections {
		if exportDir.VirtualAddress >= sec.VirtualAddress && exportDir.VirtualAddress < sec.VirtualAddress+sec.VirtualSize {
			fmt.Printf("Exports in section: %s\n", sec.Name)
		}
	}
}
