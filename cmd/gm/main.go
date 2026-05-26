package main

import (
	"os"

	"github.com/angei24/go-manager/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
