package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/j-vizcaino/ir-remotes/pkg/commands"
	"github.com/mixcode/broadlink"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var captureCmd = &cobra.Command{
	Use:   "capture [OPTIONS] COMMAND [COMMAND...]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Capture and save IR control codes.",
	Long:  `Read named IR commands from Broadlink RM and save them to an output file.`,
	Run:   Capture,
}

var remoteName string
var captureTimeout time.Duration
var discoveryTimeout time.Duration

func init() {
	flags := captureCmd.Flags()
	flags.StringVarP(&remoteName,
		"remote-name",
		"n",
		"",
		"Name of the IR remote. (required)")
	captureCmd.MarkFlagRequired("remote-name")

	flags.DurationVar(&captureTimeout,
		"capture-timeout",
		30*time.Second,
		"IR control code capture timeout.")

	flags.DurationVar(&discoveryTimeout, "discovery-timeout", 5*time.Second, "Broadlink device network discovery timeout.")

	rootCmd.AddCommand(captureCmd)
}

func Capture(cmd *cobra.Command, args []string) {
	cmdReg := commands.CommandRegistry{}
	err := cmdReg.LoadFromFile(remoteFile)
	if err != nil && !os.IsNotExist(err) {
		log.WithError(err).WithField("remotes-file", remoteFile).Fatal("Failed to load command registry file")
	}

	bd := findDevice(discoveryTimeout)

	for _, cmdName := range args {
		_, ok := cmdReg[cmdName]
		if ok {
			log.WithField("command", cmdName).Info("Command name already exist in the registry. Skipping capture.")
			continue
		}

		cmd, err := captureIRCode(bd, captureTimeout, cmdName)
		if err != nil {
			log.WithError(err).Fatal("Failed to capture IR command")
		}

		if err := cmdReg.AddCommand(cmdName, cmd); err != nil {
			log.WithError(err).WithField("command", cmdName).Error("Failed to add command to registry")
			continue
		}
	}

	if err := cmdReg.SaveToFile(remoteFile); err != nil {
		log.WithError(err).WithField("remotes-file", remoteFile).Error("Failed to save commands to file")
		os.Exit(1)
	}
}

func findDevice(timeout time.Duration) broadlink.Device {
	log.Info("Looking for Broadlink devices on your network. Please wait...")
	devs, err := broadlink.DiscoverDevices(timeout, 0)
	if err != nil {
		log.WithError(err).Fatal("Failed to discover Broadlink devices")
	}
	if len(devs) == 0 {
		log.Fatal("No Broadlink device found")
	}

	// TODO: add code to pick the right one, in case multiple devices are detected
	d := devs[0]
	log.WithField("address", d.UDPAddr.String()).
		WithField("mac", fmt.Sprintf("%02x", d.MACAddr)).
		Info("Device found!")

	myname, err := os.Hostname() // Your local machine's name.
	if err != nil {
		log.WithError(err).Fatal("Failed to get hostname")
	}
	myid := make([]byte, 15) // Must be 15 bytes long.

	err = d.Auth(myid, myname) // d.ID and d.AESKey will be updated on success.
	if err != nil {
		log.WithError(err).Fatal("Failed to authenticate with device", "addr", d.UDPAddr.String(), "mac", d.MACAddr)
	}
	return d
}

func captureIRCode(device broadlink.Device, timeout time.Duration, cmdName string) (commands.IRCommand, error) {
	// Enter capturing mode.
	if err := device.StartCaptureRemoteControlCode(); err != nil {
		log.WithError(err).Error("Failed to start capture mode")
		return nil, err
	}
	log.Infof("Waiting for IR code. Press '%s' button...", cmdName)

	start := time.Now()
	for time.Now().Sub(start) < timeout {
		remotetype, ircode, err := device.ReadCapturedRemoteControlCode()
		if err != nil {
			if err == broadlink.ErrNotCaptured {
				time.Sleep(time.Second)
				continue
			}
			return nil, err
		}
		if remotetype != broadlink.REMOTE_IR {
			return nil, fmt.Errorf("received unexpected command type %x (expected infra-red type %x)", remotetype, broadlink.REMOTE_IR)
		}
		return ircode, nil
	}
	return nil, fmt.Errorf("timed out waiting for IR control codes")
}