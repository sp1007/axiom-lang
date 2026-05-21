package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	axiomfmt "github.com/axiom-lang/axiom/tools/fmt"
)

func runFmt(args []string) int {
	flags := flag.NewFlagSet("fmt", flag.ExitOnError)
	check := flags.Bool("check", false, "Exit with 1 if any file needs formatting")
	write := flags.Bool("write", true, "Write formatted changes in-place")

	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	var files []string
	if flags.NArg() == 0 {
		// Find all .ax files recursively in the current directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting current directory: %v\n", err)
			return 1
		}
		err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".ax") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error scanning directory: %v\n", err)
			return 1
		}
	} else {
		for _, arg := range flags.Args() {
			fi, err := os.Stat(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error checking path %s: %v\n", arg, err)
				return 1
			}
			if fi.IsDir() {
				// Scan directory recursively
				err = filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && strings.HasSuffix(info.Name(), ".ax") {
						files = append(files, path)
					}
					return nil
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "error scanning directory %s: %v\n", arg, err)
					return 1
				}
			} else {
				files = append(files, arg)
			}
		}
	}

	if len(files) == 0 {
		fmt.Println("No AXIOM (.ax) files found to format.")
		return 0
	}

	formatter := axiomfmt.NewFormatter()
	hasChanges := false
	failed := false

	for _, file := range files {
		if *check {
			needsFmt, err := formatter.Check(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error checking %s: %v\n", file, err)
				failed = true
				continue
			}
			if needsFmt {
				fmt.Printf("[diff] %s needs formatting\n", file)
				hasChanges = true
			}
		} else {
			// Check if it needs formatting before writing (for messaging)
			needsFmt, err := formatter.Check(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error checking %s: %v\n", file, err)
				failed = true
				continue
			}
			if needsFmt {
				fmt.Printf("[write] formatting %s\n", file)
				if *write {
					err = formatter.FormatFile(file)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error writing %s: %v\n", file, err)
						failed = true
					}
				}
				hasChanges = true
			}
		}
	}

	if failed {
		return 1
	}

	if *check && hasChanges {
		fmt.Fprintln(os.Stderr, "error: files need formatting (run 'axc fmt <files>')")
		return 1
	}

	if !hasChanges {
		fmt.Println("All files are clean and canonically formatted.")
	}
	return 0
}
