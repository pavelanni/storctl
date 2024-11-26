package main

import (
	"os"

	"github.com/pavelanni/labshop/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
