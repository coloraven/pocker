package main

import (
	"os"

	"github.com/pocker/pocker/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
