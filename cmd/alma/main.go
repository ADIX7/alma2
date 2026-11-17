package main

import (
	"os"

	"github.com/ADIX7/alma2/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout))
}
