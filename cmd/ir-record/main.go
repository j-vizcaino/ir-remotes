package main

import (
	"flag"
	"fmt"
	"github.com/mixcode/broadlink"
	"os"
	"time"

	"github.com/j-vizcaino/ir-remotes/pkg/commands"

	log "github.com/sirupsen/logrus"
)

func init() {
    log.SetFormatter(&log.TextFormatter{
    	DisableTimestamp: true,
	})
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
	myid := make([]byte, 15)   // Must be 15 bytes long.

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

func main() {
	var discoveryTimeout time.Duration
	var captureTimeout time.Duration
	var outputFile string

	flag.DurationVar(&discoveryTimeout, "discovery-timeout", time.Second, "device discovery timeout")
	flag.DurationVar(&captureTimeout, "capture-timeout", 30*time.Second, "infra-red command capture timeout")
	flag.StringVar(&outputFile, "output", "", "output filename where captured codes get saved. If the file already exists, its content is loaded first and the captured codes are added to the content.")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Error("You must specify at least one command name to capture.")
		os.Exit(1)
	}

	if len(outputFile) == 0 {
		log.Error("You must specify an output filename using -output command line option.")
		os.Exit(1)
	}

	cmdReg := commands.CommandRegistry{}
	err := cmdReg.LoadFromFile(outputFile)
	if err != nil && !os.IsNotExist(err) {
		log.WithError(err).WithField("filename", outputFile).Fatal("Failed to load command registry file")
	}

	bd := findDevice(discoveryTimeout)

	for _, cmdName := range flag.Args() {
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

	if err := cmdReg.SaveToFile(outputFile); err != nil {
		log.WithError(err).WithField("filename", outputFile).Error("Failed to save commands to file")
		os.Exit(1)
	}
}
