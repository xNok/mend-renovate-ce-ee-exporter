package main

import (
	"fmt"
	"time"

	"github.com/xnok/mend-renovate-ce-ee-exporter/internal/cli"
)

var version = "devel"

func main() {
	fmt.Println(cli.NewApp(version, time.Now()).ToMan())
}
