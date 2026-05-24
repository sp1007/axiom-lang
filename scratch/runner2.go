package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	cmd := exec.Command("./test_malloc_native.exe")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			fmt.Printf("Process exited with code: %d\n", ws.ExitStatus())
			if ws.Signaled() {
				fmt.Printf("Process was signaled: %v\n", ws.Signal())
			}
		} else {
			fmt.Printf("Command failed to run: %v\n", err)
		}
		os.Exit(1)
	}
	fmt.Println("Command succeeded!")
}
