package main

import (
	"flag"
	"fmt"
	"github.com/mixcode/broadlink"
	"os"
	"time"

	"github.com/j-vizcaino/ir-remotes/pkg/commands"

	"github.com/manifoldco/promptui"
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
	log.WithField("device", d).Info("Device found!")
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

func captureIRCode(device broadlink.Device, timeout time.Duration) (commands.IRCommand, error) {
	// Enter capturing mode.
	if err := device.StartCaptureRemoteControlCode(); err != nil {
		log.WithError(err).Error("Failed to start capture mode")
		return nil, err
	}
	log.Info("Waiting for IR command...")

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

func loadCommandRegistry(filename string) (commands.CommandRegistry, error) {
	ret := commands.CommandRegistry{}
	fd, err := os.Open(filename)
	if os.IsNotExist(err) {
		return ret, nil
	}
	defer fd.Close()

	err = ret.Load(fd)
	return ret, err
}

func saveCommandRegistry(filename string, cmdReg commands.CommandRegistry) error {
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	return cmdReg.Save(fd)
}

func askContinue() (bool, error) {
	prompt := promptui.Select{
		Label: "Select next action",
		Items: []string{"Add another command", "Save and exit"},
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return idx == 0, nil
}

func main() {
	var discoveryTimeout time.Duration
	var captureTimeout time.Duration

	flag.DurationVar(&discoveryTimeout, "discovery-timeout", time.Second, "device discovery timeout")
	flag.DurationVar(&captureTimeout, "capture-timeout", 30*time.Second, "infra-red command capture timeout")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Error("You must specify the command registry output filename.")
		os.Exit(1)
	}

	outFile := flag.Arg(0)
	cmdReg, err := loadCommandRegistry(outFile)
	if err != nil {
		log.WithError(err).WithField("filename", outFile).Fatal("Failed to load command registry file")
	}

	bd := findDevice(discoveryTimeout)

	keepGoing := true

	for keepGoing {
		cmd, err := captureIRCode(bd, captureTimeout)
		if err != nil {
			log.WithError(err).Fatal("Failed to capture IR command")
		}
		prompt := promptui.Prompt{
			Label:    "Name of the command",
		}
		result, err := prompt.Run()
		if err != nil {
			log.WithError(err).Error("Prompt failed")
			break
		}

		if err := cmdReg.AddCommand(result, cmd); err != nil {
			log.WithError(err).Error("Failed to add command to registry")
			continue
		}

		keepGoing, err = askContinue()
		if err != nil {
			log.WithError(err).Error("Prompt failed")
			break
		}
	}

	if err := saveCommandRegistry(outFile, cmdReg); err != nil {
		log.WithError(err).WithField("filename", outFile).Error("Failed to save commands to file")
		os.Exit(1)
	}
}
