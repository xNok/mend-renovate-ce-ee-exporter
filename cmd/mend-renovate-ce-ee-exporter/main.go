package main

import (
	"os"

	"github.com/xnok/mend-renovate-ce-ee-exporter/internal/cli"
)

var version = "devel"

func main() {
	cli.Run(version, os.Args)
}
