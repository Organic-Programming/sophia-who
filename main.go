// Sophia Who? — the primordial holon.
// She creates identity cards (HOLON.md) for other holons.
// "Know thyself."
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "new":
		if err := runNew(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "show":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: who show <uuid>")
			os.Exit(1)
		}
		if err := runShow(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := runList(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "pin":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: who pin <uuid>")
			os.Exit(1)
		}
		if err := runPin(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Printf("sophia-who v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Sophia Who? — the primordial holon identity-maker.
"Know thyself."

Usage:
  who new          Create a new holon identity (interactive)
  who show <uuid>  Display a holon's identity
  who list         List all known holons in the current project
  who pin <uuid>   Capture version/commit/arch for a holon's binary
  who version      Print version
  who help         Print this help`)
}
