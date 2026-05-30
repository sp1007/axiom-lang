package main

import (
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("scratch/stage2_preprocessed.ax")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("File size: %d bytes\n", len(data))

	start := 2600
	end := 2800
	for i := start; i < end; i++ {
		b := data[i]
		fmt.Printf("%d: 0x%02X (%q)\n", i, b, string(b))
	}
}

