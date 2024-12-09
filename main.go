package main

import (
	"os"

	"github.com/pavelanni/storctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
