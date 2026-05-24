package main

import (
	"debug/pe"
	"fmt"
	"log"
	"reflect"
)

func dumpHeader(filePath string) *pe.OptionalHeader64 {
	file, err := pe.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening PE %s: %v", filePath, err)
	}
	defer file.Close()

	hdr, ok := file.OptionalHeader.(*pe.OptionalHeader64)
	if !ok {
		log.Fatalf("Not a 64-bit PE optional header: %s", filePath)
	}
	return hdr
}

func main() {
	h1 := dumpHeader("d:\\projects\\compiler\\Axiom\\test_malloc.exe")
	h2 := dumpHeader("d:\\projects\\compiler\\Axiom\\bin\\axc_stage1.exe")

	val := reflect.ValueOf(*h1)
	val2 := reflect.ValueOf(*h2)
	typ := val.Type()

	fmt.Printf("%-30s %-20s %-20s\n", "Field Name", "test_malloc.exe", "test_va.exe")
	fmt.Println(reflect.ValueOf("--------------------------------------------------------------------------------").String()[:75])

	f1, _ := pe.Open("d:\\projects\\compiler\\Axiom\\test_malloc.exe")
	f2, _ := pe.Open("d:\\projects\\compiler\\Axiom\\scratch\\test_va.exe")
	fmt.Printf("%-30s 0x%-18X 0x%-18X\n", "COFF Characteristics", f1.FileHeader.Characteristics, f2.FileHeader.Characteristics)
	f1.Close()
	f2.Close()

	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		if fieldName == "DataDirectory" {
			continue
		}
		v1 := val.Field(i).Interface()
		v2 := val2.Field(i).Interface()
		
		fmt.Printf("%-30s 0x%-18X 0x%-18X\n", fieldName, v1, v2)
	}
}
