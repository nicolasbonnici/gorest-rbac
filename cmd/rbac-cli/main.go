package main

import (
	"fmt"
	"os"

	"github.com/nicolasbonnici/gorest-rbac/cmd/rbac-cli/commands"
)

// Version information (set during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set version info
	commands.SetVersion(version, commit, date)

	// Execute root command
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
