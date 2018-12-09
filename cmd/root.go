package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cmdRoot = &cobra.Command{
	Use:   "ir-remotes",
	Short: "Tool for using Broadlink RM infra-red blasters",
}

var remotesFile string
var devicesFile string

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	flags := cmdRoot.PersistentFlags()
	flags.StringVarP(&remotesFile,
		"remotes-file",
		"f",
		"remotes.json",
		"Filename where remotes IR codes are loaded and saved.")

	flags.StringVarP(&devicesFile,
		"devices-file",
		"d",
		"devices.json",
		"Filename where Broadlink devices information are loaded and saved.")
}

func Execute() {
	if err := cmdRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
