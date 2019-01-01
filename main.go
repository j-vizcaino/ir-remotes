package main

import (
	"fmt"
	"github.com/j-vizcaino/ir-remotes/cmd"
	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	commit     = "none"
	date       = "unknown"

	cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Print the version of ir-remotes",
		Long:  "Print the version of ir-remotes",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Version: %s\nGit ref: %s\nBuild date: %s\n", version, commit, date)
		},
	}
)

func main() {
	cmd.Root().AddCommand(cmdVersion)
	cmd.Execute()
}
