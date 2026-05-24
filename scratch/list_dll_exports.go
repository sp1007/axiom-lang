package main

import (
	"debug/pe"
	"fmt"
	"os"
)

func main() {
	dllPath := "ax_runtime.dll"
	if len(os.Args) > 1 {
		dllPath = os.Args[1]
	}
	f, err := pe.Open(dllPath)
	if err != nil {
		fmt.Printf("Failed to open DLL: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var oh *pe.OptionalHeader64
	switch o := f.OptionalHeader.(type) {
	case *pe.OptionalHeader64:
		oh = o
	default:
		fmt.Println("Only PE32+ supported")
		os.Exit(1)
	}

	exportDir := oh.DataDirectory[0]
	if exportDir.VirtualAddress == 0 {
		fmt.Println("No export directory found")
		os.Exit(0)
	}

	// Read the export directory structure
	var expSec *pe.Section
	for _, sec := range f.Sections {
		if exportDir.VirtualAddress >= sec.VirtualAddress && exportDir.VirtualAddress < sec.VirtualAddress+sec.VirtualSize {
			expSec = sec
			break
		}
	}

	if expSec == nil {
		fmt.Println("Export directory not in any section")
		os.Exit(1)
	}

	data, err := expSec.Data()
	if err != nil {
		fmt.Printf("Failed to read section data: %v\n", err)
		os.Exit(1)
	}

	// Offset in section data
	offset := exportDir.VirtualAddress - expSec.VirtualAddress
	if int(offset)+40 > len(data) {
		fmt.Println("Export directory out of bounds")
		os.Exit(1)
	}

	// Parse Export Directory
	// struct IMAGE_EXPORT_DIRECTORY:
	//   0: Characteristics (4 bytes)
	//   4: TimeDateStamp (4 bytes)
	//   8: MajorVersion/MinorVersion (4 bytes)
	//   12: Name RVA (4 bytes)
	//   16: Base (4 bytes)
	//   20: NumberOfFunctions (4 bytes)
	//   24: NumberOfNames (4 bytes)
	//   28: AddressOfFunctions RVA (4 bytes)
	//   32: AddressOfNames RVA (4 bytes)
	//   36: AddressOfNameOrdinals RVA (4 bytes)
	
	rvaToOffset := func(rva uint32) uint32 {
		for _, s := range f.Sections {
			if rva >= s.VirtualAddress && rva < s.VirtualAddress+s.VirtualSize {
				return rva - s.VirtualAddress + s.Offset
			}
		}
		return 0
	}

	readUint32 := func(offset uint32) uint32 {
		off := rvaToOffset(offset)
		fileData, err := os.ReadFile(dllPath)
		if err != nil {
			return 0
		}
		if off+4 <= uint32(len(fileData)) {
			return uint32(fileData[off]) | uint32(fileData[off+1])<<8 | uint32(fileData[off+2])<<16 | uint32(fileData[off+3])<<24
		}
		return 0
	}

	readString := func(rva uint32) string {
		off := rvaToOffset(rva)
		fileData, err := os.ReadFile(dllPath)
		if err != nil {
			return ""
		}
		var s []byte
		for off < uint32(len(fileData)) && fileData[off] != 0 {
			s = append(s, fileData[off])
			off++
		}
		return string(s)
	}

	numberOfNames := readUint32(exportDir.VirtualAddress + 24)
	addressOfNames := readUint32(exportDir.VirtualAddress + 32)

	fmt.Printf("DLL Name: %s\n", readString(readUint32(exportDir.VirtualAddress + 12)))
	fmt.Printf("Number of Names: %d\n", numberOfNames)
	fmt.Println("Exported Symbols:")
	for i := uint32(0); i < numberOfNames; i++ {
		nameRVA := readUint32(addressOfNames + i*4)
		name := readString(nameRVA)
		fmt.Printf("  - %s\n", name)
	}
}
