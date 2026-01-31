package main

import (
	"os"

	"github.com/BuildGuard-Test-Lab/pbom/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
