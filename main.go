package main

import (
	"fmt"
	"os"

	"github.com/peiman/ckeletin-go/cmd"
)

func run() int {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

// main is intentionally not covered by tests because it's the program's entry point.
// All logic is tested via the run() function and other commands. The main function’s
// sole purpose is to call run() and exit accordingly. Attempting to cover main directly
// would require integration tests or running the built binary separately.
func main() {
	os.Exit(run())
}
