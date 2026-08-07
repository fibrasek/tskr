package main

import (
	"fmt"
	"os"

	"tskr/internal/cli"
)

func main() {
	if err := cli.Run(os.Args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
