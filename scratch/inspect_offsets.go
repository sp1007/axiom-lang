package main

import (
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("scratch/self_linked_concatenated.ax")
	if err != nil {
		fmt.Printf("Error reading: %v\n", err)
		return
	}
	
	offsets := []int{16952, 21110, 25363}
	for _, offset := range offsets {
		fmt.Printf("\n--- Offset %d ---\n", offset)
		start := offset - 100
		if start < 0 {
			start = 0
		}
		end := offset + 100
		if end > len(data) {
			end = len(data)
		}
		
		// Print snippet with a marker at the exact offset
		before := string(data[start:offset])
		after := string(data[offset:end])
		fmt.Printf("%s>>>HERE<<<%s\n", before, after)
	}
}
