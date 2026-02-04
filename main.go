package main

import (
	"fmt"
	"git-ai-commit/cmd"
	"os"
)

func main() {
	if err := cmd.RunWithArgs(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
