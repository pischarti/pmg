package main

import (
	"os"

	"github.com/pischarti/pmg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
