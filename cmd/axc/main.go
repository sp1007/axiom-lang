package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "dump-tokens":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: axc dump-tokens <file.ax> [--compact] [--no-text] [--stats]")
			os.Exit(1)
		}
		filename := os.Args[2]
		compact, noText, stats := false, false, false
		for _, arg := range os.Args[3:] {
			switch arg {
			case "--compact":
				compact = true
			case "--no-text":
				noText = true
			case "--stats":
				stats = true
			default:
				fmt.Fprintf(os.Stderr, "axc: unknown flag %q\n", arg)
				os.Exit(1)
			}
		}
		os.Exit(runDumpTokens(filename, compact, noText, stats))

	case "dump-ast":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: axc dump-ast <file.ax>")
			os.Exit(1)
		}
		os.Exit(runDumpAST(os.Args[2]))

	case "check":
		runCheck(os.Args[2:])
		
	case "build":
		fmt.Fprintln(os.Stderr, "axc: build command not yet implemented")
		os.Exit(1)

	case "version":
		fmt.Println("axc 0.0.1-dev (AXIOM compiler)")

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "axc: unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: axc <command> [args]

commands:
  build         Compile an AXIOM source file
  dump-tokens   Tokenize a file and print JSON token stream
  dump-ast      Parse a file and print the AST
  version       Print compiler version
  help          Show this help message

dump-tokens flags:
  --compact     One-line JSON output
  --no-text     Omit token text field
  --stats       Print token count summary`)
}
