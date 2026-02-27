package main

import (
	"os"

	"nitid/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
