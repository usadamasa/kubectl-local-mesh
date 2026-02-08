package main

import (
	"fmt"
	"os"

	"github.com/usadamasa/kubectl-localmesh/cmd"
	versionpkg "github.com/usadamasa/kubectl-localmesh/internal/version"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	info := versionpkg.Resolve(version, commit, date)
	cmd.SetVersion(info)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
