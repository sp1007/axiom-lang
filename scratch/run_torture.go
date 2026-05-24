package main

import (
	"fmt"
	"os/exec"
	"syscall"
)

func main() {
	cmd := exec.Command("./torture_gen_ref.exe")
	out, err := cmd.CombinedOutput()
	fmt.Printf("Output:\n%s\n", string(out))
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			fmt.Printf("Exited with status: %d (0x%X)\n", ws.ExitStatus(), ws.ExitStatus())
			return
		}
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Success! Exit code: 0")
}
