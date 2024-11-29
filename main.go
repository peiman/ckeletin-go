// main.go
package main

import (
	"os"

	"github.com/peiman/ckeletin-go/cmd"
)

var osExit = os.Exit // Mockable variable for os.Exit

func main() {
	cmd.Execute()
	osExit(0) // Use osExit for testability
}
