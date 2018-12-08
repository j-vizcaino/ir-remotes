package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ir-remotes",
	Short: "Tool for using Broadlink RM infra-red blasters",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var remoteFile string

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&remoteFile,
		"remotes-file",
		"f",
		"remotes.json",
		"Filename where remotes IR codes are loaded and saved.")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
