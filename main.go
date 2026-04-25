package main

import (
	"os"

	"github.com/K-tecchan/gh-orz/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
