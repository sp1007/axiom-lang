package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func main() {
	data, err := os.ReadFile("d:\\projects\\compiler\\Axiom\\print.exe")
	if err != nil {
		log.Fatalf("Failed to read print.exe: %v", err)
	}

	target := []byte("Hello, world!")
	idx := bytes.Index(data, target)
	if idx != -1 {
		fmt.Printf("Success! Found 'Hello, world!' at file offset 0x%X\n", idx)
		// Check what section it is in
	} else {
		fmt.Println("Error: 'Hello, world!' not found anywhere in print.exe!")
	}
}
