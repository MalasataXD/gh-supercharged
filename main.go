package main

import (
	"os"

	"github.com/MalasataXD/gh-supercharged/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
