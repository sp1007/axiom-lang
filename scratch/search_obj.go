package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func main() {
	data, err := os.ReadFile("d:\\projects\\compiler\\Axiom\\axiom_temp.obj")
	if err != nil {
		log.Fatalf("Failed to read axiom_temp.obj: %v", err)
	}

	target := []byte("Hello, world!")
	idx := bytes.Index(data, target)
	if idx != -1 {
		fmt.Printf("Success! Found 'Hello, world!' in axiom_temp.obj at file offset 0x%X\n", idx)
	} else {
		fmt.Println("Error: 'Hello, world!' not found in axiom_temp.obj!")
	}
}
