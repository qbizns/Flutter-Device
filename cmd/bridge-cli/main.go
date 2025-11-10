package main

import (
	"os"

	"github.com/Macber-eg/Flutter-Device/cmd/bridge-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
