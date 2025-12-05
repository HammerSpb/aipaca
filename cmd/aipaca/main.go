package main

import (
	"os"

	"github.com/HammerSpb/aipaca/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
