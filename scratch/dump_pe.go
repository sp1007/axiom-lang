package main

import (
	"debug/pe"
	"encoding/binary"
	"fmt"
	"os"
)

// ImportDescriptor represents the raw IMAGE_IMPORT_DESCRIPTOR structure.
type ImportDescriptor struct {
	OriginalFirstThunk uint32
	TimeDateStamp      uint32
	ForwarderChain     uint32
	Name               uint32
	FirstThunk         uint32
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run dump_pe.go <pe-file>")
		os.Exit(1)
	}
	f, err := pe.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Error opening PE: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Printf("File: %s\n", os.Args[1])
	fmt.Printf("NumberOfSections: %d\n", f.FileHeader.NumberOfSections)

	var oh64 *pe.OptionalHeader64
	var oh32 *pe.OptionalHeader32
	var importDir pe.DataDirectory

	switch oh := f.OptionalHeader.(type) {
	case *pe.OptionalHeader64:
		oh64 = oh
		fmt.Printf("AddressOfEntryPoint: 0x%08X (RVA)\n", oh.AddressOfEntryPoint)
		fmt.Println("\nOptional Header Data Directories:")
		dirNames := []string{
			"Export Table", "Import Table", "Resource Table", "Exception Table",
			"Certificate Table", "Base Relocation Table", "Debug Directory", "Architecture Specific",
			"Global Pointer", "Thread Local Storage Table", "Load Config Directory", "Bound Import Table",
			"Import Address Table (IAT)", "Delay Import Descriptor", "CLR Runtime Header", "Reserved",
		}
		for i, dir := range oh.DataDirectory {
			if i < len(dirNames) {
				fmt.Printf("  Dir %2d [%-26s]: RVA=0x%08X, Size=%d\n", i, dirNames[i], dir.VirtualAddress, dir.Size)
			}
		}
		importDir = oh.DataDirectory[1]
	case *pe.OptionalHeader32:
		oh32 = oh
		fmt.Printf("AddressOfEntryPoint: 0x%08X (RVA)\n", oh.AddressOfEntryPoint)
		importDir = oh.DataDirectory[1]
	}

	fmt.Printf("Import Directory: RVA=0x%08X, Size=%d\n", importDir.VirtualAddress, importDir.Size)

	// Dump section details
	fmt.Println("\nSections:")
	for idx, sec := range f.Sections {
		fmt.Printf("  Sec %d: Name: %-8s RVA: 0x%08X-0x%08X (Size: %d) Offset: 0x%08X Characteristics: 0x%08X\n",
			idx+1, sec.Name, sec.VirtualAddress, sec.VirtualAddress+sec.VirtualSize, sec.VirtualSize, sec.Offset, sec.Characteristics)
	}

	// Helper function to read a null-terminated string from PE at RVA
	readString := func(rva uint32) string {
		if rva == 0 {
			return "<nil>"
		}
		// Find section
		for _, sec := range f.Sections {
			if rva >= sec.VirtualAddress && rva < sec.VirtualAddress+sec.VirtualSize {
				offset := sec.Offset + (rva - sec.VirtualAddress)
				
				// Read file bytes
				file, err := os.Open(os.Args[1])
				if err != nil {
					return fmt.Sprintf("<error: %v>", err)
				}
				defer file.Close()
				
				_, _ = file.Seek(int64(offset), 0)
				var buf [256]byte
				_, _ = file.Read(buf[:])
				
				// Extract string
				for i, b := range buf {
					if b == 0 {
						return string(buf[:i])
					}
				}
				return string(buf[:])
			}
		}
		return "<RVA out of bounds>"
	}

	// Helper function to read data bytes from PE at RVA
	readBytes := func(rva uint32, length uint32) []byte {
		for _, sec := range f.Sections {
			if rva >= sec.VirtualAddress && rva < sec.VirtualAddress+sec.VirtualSize {
				offset := sec.Offset + (rva - sec.VirtualAddress)
				file, err := os.Open(os.Args[1])
				if err != nil {
					return nil
				}
				defer file.Close()
				_, _ = file.Seek(int64(offset), 0)
				buf := make([]byte, length)
				_, _ = file.Read(buf)
				return buf
			}
		}
		return nil
	}

	if importDir.VirtualAddress != 0 {
		fmt.Println("\nRaw Import Directory Table (IMAGE_IMPORT_DESCRIPTORs):")
		rva := importDir.VirtualAddress
		for i := 0; ; i++ {
			b := readBytes(rva, 20)
			if len(b) < 20 {
				fmt.Println("  [Truncated read or EOF]")
				break
			}
			
			var desc ImportDescriptor
			desc.OriginalFirstThunk = binary.LittleEndian.Uint32(b[0:4])
			desc.TimeDateStamp = binary.LittleEndian.Uint32(b[4:8])
			desc.ForwarderChain = binary.LittleEndian.Uint32(b[8:12])
			desc.Name = binary.LittleEndian.Uint32(b[12:16])
			desc.FirstThunk = binary.LittleEndian.Uint32(b[16:20])
			
			if desc.OriginalFirstThunk == 0 && desc.Name == 0 && desc.FirstThunk == 0 {
				fmt.Printf("  Descriptor %d: [Terminator (all zeros)]\n", i)
				break
			}
			
			libName := readString(desc.Name)
			fmt.Printf("  Descriptor %d: NameRVA=0x%08X (%s)\n", i, desc.Name, libName)
			fmt.Printf("    OriginalFirstThunk (ILT RVA): 0x%08X\n", desc.OriginalFirstThunk)
			fmt.Printf("    FirstThunk (IAT RVA):         0x%08X\n", desc.FirstThunk)
			fmt.Printf("    TimeDateStamp: 0x%08X, ForwarderChain: 0x%08X\n", desc.TimeDateStamp, desc.ForwarderChain)
			
			// Dump ILT entries
			fmt.Println("    Import Lookup Table (ILT):")
			iltRva := desc.OriginalFirstThunk
			if iltRva == 0 {
				iltRva = desc.FirstThunk // Fallback to IAT if ILT is 0
			}
			for entryIdx := 0; ; entryIdx++ {
				eb := readBytes(iltRva + uint32(entryIdx)*8, 8)
				if len(eb) < 8 {
					break
				}
				val := binary.LittleEndian.Uint64(eb)
				if val == 0 {
					fmt.Printf("      [%d] 0x0000000000000000 (Terminator)\n", entryIdx)
					break
				}
				if val&(1<<63) != 0 {
					// Ordinal
					fmt.Printf("      [%d] Ordinal: %d\n", entryIdx, val&0xFFFF)
				} else {
					// Name
					nameRva := uint32(val & 0xFFFFFFFF)
					hint := uint16(0)
					hBytes := readBytes(nameRva, 2)
					if len(hBytes) >= 2 {
						hint = binary.LittleEndian.Uint16(hBytes)
					}
					symName := readString(nameRva + 2)
					fmt.Printf("      [%d] Hint: %d, Name: %s (RVA: 0x%08X)\n", entryIdx, hint, symName, nameRva)
				}
			}
			
			rva += 20
		}
	}

	_ = oh64
	_ = oh32
}
