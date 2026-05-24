package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\minimal.exe")
	if err != nil {
		log.Fatalf("Error opening PE file: %v", err)
	}
	defer file.Close()

	fmt.Printf("Machine: 0x%X\n", file.Machine)
	fmt.Printf("NumberOfSections: %d\n", len(file.Sections))

	fmt.Println("\nSections:")
	for _, sec := range file.Sections {
		fmt.Printf("  Name: %-8s  VirtSize: 0x%-6X  VirtAddr (RVA): 0x%-6X  RawSize: %-6d Offset: 0x%X\n",
			sec.Name, sec.VirtualSize, sec.VirtualAddress, sec.Size, sec.Offset)
	}

	// Read PE Optional Header
	var impRVA, impSize uint32
	switch hdr := file.OptionalHeader.(type) {
	case *pe.OptionalHeader64:
		fmt.Printf("\nPE64 Optional Header:\n")
		fmt.Printf("  AddressOfEntryPoint: 0x%X\n", hdr.AddressOfEntryPoint)
		fmt.Printf("  ImageBase: 0x%X\n", hdr.ImageBase)
		fmt.Printf("  SectionAlignment: 0x%X\n", hdr.SectionAlignment)
		fmt.Printf("  FileAlignment: 0x%X\n", hdr.FileAlignment)
		
		fmt.Printf("\nData Directories:\n")
		// Directory 1 is Import Directory (index 1)
		impRVA = hdr.DataDirectory[1].VirtualAddress
		impSize = hdr.DataDirectory[1].Size
		fmt.Printf("  Import Directory: RVA=0x%X Size=0x%X\n", impRVA, impSize)
	case *pe.OptionalHeader32:
		fmt.Printf("\nPE32 Optional Header:\n")
		impRVA = hdr.DataDirectory[1].VirtualAddress
		impSize = hdr.DataDirectory[1].Size
		fmt.Printf("  Import Directory: RVA=0x%X Size=0x%X\n", impRVA, impSize)
	}
}
