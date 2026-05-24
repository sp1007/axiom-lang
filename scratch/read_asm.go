package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	file, err := os.Open("d:\\projects\\compiler\\Axiom\\axiom_temp.asm")
	if err != nil {
		log.Fatalf("failed to open: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		if lineNum >= 6100 && lineNum <= 6180 {
			fmt.Printf("%d: %s\n", lineNum, scanner.Text())
		}
		lineNum++
	}
}
