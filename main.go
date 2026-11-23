// main.go

package main

import (
	"os"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/peiman/ckeletin-go/cmd"
)

func run() int {
	return output.Execute(cmd.RootCmd)
}

// main is intentionally not covered by tests because it's the program's entry point.
// All logic is tested via the run() function and other commands. The main function’s
// sole purpose is to call run() and exit accordingly. Attempting to cover main directly
// would require integration tests or running the built binary separately.
func main() {
	os.Exit(run())
}
