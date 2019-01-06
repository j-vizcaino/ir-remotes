package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cmdCompletion = &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion (Bash or Zsh)",
	}
	cmdCompletionBash = &cobra.Command{
		Use:   "bash",
		Short: "Generate Bash completion",
		Run: func(_ *cobra.Command, _ []string) {
			if err := cmdRoot.GenBashCompletion(os.Stdout); err != nil {
				logrus.WithError(err).Fatal("Failed to generate Bash completion")
			}
		},
	}
	cmdCompletionZsh = &cobra.Command{
		Use:   "zsh",
		Short: "Generate Zsh completion",
		Run: func(_ *cobra.Command, _ []string) {
			if err := cmdRoot.GenZshCompletion(os.Stdout); err != nil {
				logrus.WithError(err).Fatal("Failed to generate Zsh completion")
			}
		},
	}
)

func init() {
	cmdCompletion.AddCommand(cmdCompletionBash, cmdCompletionZsh)
	cmdRoot.AddCommand(cmdCompletion)
}
