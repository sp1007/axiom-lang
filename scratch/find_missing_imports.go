package main

import (
	"debug/pe"
	"fmt"
	"os"
	"syscall"
)

func main() {
	target := "bin/axc_stage2.exe"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}
	f, err := pe.Open(target)
	if err != nil {
		fmt.Printf("Failed to open PE %s: %v\n", target, err)
		os.Exit(1)
	}
	defer f.Close()

	importedSymbols, err := f.ImportedSymbols()
	if err != nil {
		fmt.Printf("Failed to get imported symbols: %v\n", err)
		os.Exit(1)
	}

	libs := make(map[string]syscall.Handle)
	getLib := func(name string) syscall.Handle {
		if h, ok := libs[name]; ok {
			return h
		}
		h, err := syscall.LoadLibrary(name)
		if err != nil {
			fmt.Printf("Failed to LoadLibrary %s: %v\n", name, err)
			libs[name] = 0
			return 0
		}
		libs[name] = h
		return h
	}

	defer func() {
		for _, h := range libs {
			if h != 0 {
				syscall.FreeLibrary(h)
			}
		}
	}()

	fmt.Println("Testing imported symbols resolution:")
	missingCount := 0
	for _, sym := range importedSymbols {
		// sym is in format "FuncName:DllName.dll"
		var funcName, dllName string
		for i := 0; i < len(sym); i++ {
			if sym[i] == ':' {
				funcName = sym[:i]
				dllName = sym[i+1:]
				break
			}
		}
		if funcName == "" || dllName == "" {
			fmt.Printf("Invalid import format: %s\n", sym)
			continue
		}

		h := getLib(dllName)
		if h == 0 {
			missingCount++
			continue
		}

		proc, err := syscall.GetProcAddress(h, funcName)
		if err != nil {
			fmt.Printf("  [MISSING] Symbol '%s' NOT found in %s (err: %v)\n", funcName, dllName, err)
			missingCount++
		} else {
			_ = proc
			// fmt.Printf("  [OK] %s in %s\n", funcName, dllName)
		}
	}

	if missingCount == 0 {
		fmt.Println("All imported symbols successfully resolved on this machine!")
	} else {
		fmt.Printf("Found %d missing symbols!\n", missingCount)
	}
}
