// main.go
package main

import (
	"github.com/peiman/ckeletin-go/cmd"
	"github.com/peiman/ckeletin-go/internal/logger"
)

func main() {
	logger.Init()
	cmd.Execute()
}
