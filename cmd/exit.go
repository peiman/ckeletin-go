package cmd

import "os"

// osExit is a variable that holds the os.Exit function.
// It can be replaced in tests to prevent actual program termination.
var osExit = os.Exit
