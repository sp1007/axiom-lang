package main

import (
	"bytes"
	"debug/pe"
	"fmt"
	"log"
	"os"
)

func main() {
	f1, err := os.ReadFile("d:\\projects\\compiler\\Axiom\\print.exe")
	if err != nil {
		log.Fatalf("Error reading print.exe: %v", err)
	}

	f2, err := os.ReadFile("d:\\projects\\compiler\\Axiom\\print_stage1.exe")
	if err != nil {
		log.Fatalf("Error reading print_stage1.exe: %v", err)
	}

	if len(f1) != len(f2) {
		fmt.Printf("File size mismatch: %d vs %d bytes\n", len(f1), len(f2))
	} else {
		fmt.Printf("File sizes are identical: %d bytes\n", len(f1))
	}

	if bytes.Equal(f1, f2) {
		fmt.Println("Files are byte-for-byte identical!")
		return
	}

	// Compare section by section
	pe1, err := pe.Open("d:\\projects\\compiler\\Axiom\\print.exe")
	if err != nil {
		log.Fatalf("Error opening print.exe as PE: %v", err)
	}
	defer pe1.Close()

	pe2, err := pe.Open("d:\\projects\\compiler\\Axiom\\print_stage1.exe")
	if err != nil {
		log.Fatalf("Error opening print_stage1.exe as PE: %v", err)
	}
	defer pe2.Close()

	fmt.Println("\nComparing Sections:")
	for _, sec1 := range pe1.Sections {
		sec2 := findSection(pe2, sec1.Name)
		if sec2 == nil {
			fmt.Printf("Section %s not found in print_stage1.exe\n", sec1.Name)
			continue
		}

		d1, err := sec1.Data()
		if err != nil {
			log.Printf("Error reading data for section %s in print.exe: %v", sec1.Name, err)
			continue
		}
		d2, err := sec2.Data()
		if err != nil {
			log.Printf("Error reading data for section %s in print_stage1.exe: %v", sec2.Name, err)
			continue
		}

		if len(d1) != len(d2) {
			fmt.Printf("  Section %s: size mismatch (%d vs %d)\n", sec1.Name, len(d1), len(d2))
			continue
		}

		if bytes.Equal(d1, d2) {
			fmt.Printf("  Section %s: IDENTICAL\n", sec1.Name)
		} else {
			diffCount := 0
			firstDiff := -1
			for i := 0; i < len(d1); i++ {
				if d1[i] != d2[i] {
					if firstDiff == -1 {
						firstDiff = i
					}
					diffCount++
				}
			}
			fmt.Printf("  Section %s: MISMATCH (%d bytes differ, first diff at local offset 0x%X, RVA 0x%X)\n",
				sec1.Name, diffCount, firstDiff, sec1.VirtualAddress+uint32(firstDiff))

			// Dump first few mismatches
			limit := 10
			count := 0
			for i := firstDiff; i < len(d1) && count < limit; i++ {
				if d1[i] != d2[i] {
					fmt.Printf("    At RVA 0x%X (offset 0x%X): 0x%02X vs 0x%02X\n",
						sec1.VirtualAddress+uint32(i), i, d1[i], d2[i])
					count++
				}
			}
		}
	}
}

func findSection(file *pe.File, name string) *pe.Section {
	for _, sec := range file.Sections {
		if sec.Name == name {
			return sec
		}
	}
	return nil
}
