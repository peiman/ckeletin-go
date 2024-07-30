package main

import (
	"fmt"
	"os"

	"github.com/peiman/ckeletin-go/cmd"
)

func run() error {
	return cmd.Execute()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
