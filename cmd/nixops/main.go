// Command nixops is the CLI and TUI entry point.
package main

import (
	"os"

	"nixops/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], cli.DefaultStdio()))
}
