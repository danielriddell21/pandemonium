package main

import (
	"fmt"
	"os"

	"github.com/danielriddell21/pandemonium/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
