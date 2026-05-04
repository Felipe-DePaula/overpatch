package main

import (
	"os"

	"github.com/Felipe-DePaula/overpatch/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
