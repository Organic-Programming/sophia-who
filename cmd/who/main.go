// Command who is the CLI entry point for Sophia Who?,
// the holon identity manager.
package main

import (
	"fmt"
	"os"

	"sophia-who/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "new":
		err = cli.RunNew()
	case "show":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: who show <uuid>")
			os.Exit(1)
		}
		err = cli.RunShow(os.Args[2])
	case "list":
		err = cli.RunList()
	case "pin":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: who pin <uuid>")
			os.Exit(1)
		}
		err = cli.RunPin(os.Args[2])
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Sophia Who? — holon identity manager

Usage:
  who new         create a new holon identity (interactive)
  who show <uuid> display a holon's identity
  who list        list all known holons (local + cached)
  who pin <uuid>  capture version/commit/arch for a holon's binary`)
}
