package main

import (
	"fmt"
	"os"

	"github.com/peiman/ckeletin-go/cmd"
)

var runFunc = defaultRun
var exitFunc = os.Exit

func defaultRun() error {
	return cmd.Execute()
}

func main() {
	if err := runFunc(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		exitFunc(1)
	}
}
