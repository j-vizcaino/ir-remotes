package cmd

import (
	"net"
	"os"

	"github.com/mixcode/broadlink"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/j-vizcaino/ir-remotes/pkg/devices"
	"github.com/j-vizcaino/ir-remotes/pkg/utils"

	"github.com/manifoldco/promptui"
)

var (
	cmdDevices = &cobra.Command{
		Use:   "devices COMMAND",
		Short: "Manage Broadlink devices.",
	}

	cmdDevDiscover = &cobra.Command{
		Use:   "discover",
		Short: "Discover Broadlink devices on the network.",
		Long:  "Discover Broadlink devices on the network and save those to a file.",
		Run:   Discover,
	}
)

func init() {
	cmdDevices.AddCommand(cmdDevDiscover)
	cmdRoot.AddCommand(cmdDevices)
}

func getDeviceName() string {
	prompt := promptui.Prompt{
		Label: "Device name",
	}
	result, err := prompt.Run()
	if err != nil {
		log.WithError(err).Fatal("Prompt failed")
	}

	return result
}

func Discover(_ *cobra.Command, _ []string) {
	deviceList := devices.DeviceInfoList{}
	err := utils.LoadFromFile(&deviceList, devicesFile)
	if err != nil && !os.IsNotExist(err) {
		log.WithError(err).WithField("devices-file", devicesFile).Fatal("Failed to load devices from file.")
	}

	log.Info("Looking for Broadlink devices on your network. Please wait...")
	discovered, err := broadlink.DiscoverDevices(discoveryTimeout, 0)
	if err != nil {
		log.WithError(err).Fatal("Failed to discover Broadlink devices")
	}
	if len(discovered) == 0 {
		log.Fatal("No Broadlink device found")
	}

	modified := false
	for _, bd := range discovered {
		model, _ := bd.DeviceName()
		macAddr := net.HardwareAddr(bd.MACAddr).String()
		log.WithField("mac-address", macAddr).
			WithField("udp-address", bd.UDPAddr.String()).
			WithField("model", model).
			Info("Found device.")

		existing, found := deviceList.Find(func(dev *devices.DeviceInfo) bool {
			return dev.MACAddress == macAddr
		})
		if found {
			log.WithField("mac-address", existing.MACAddress).
				WithField("name", existing.Name).
				Info("Device already exist in device list. Skipping.")
			continue
		}

		devName := getDeviceName()
		if err := deviceList.AddDevice(devName, bd); err != nil {
			log.WithError(err).Error("Failed to store device")
		} else {
			modified = true
		}
	}

	if modified {
		if err := utils.SaveToFile(&deviceList, devicesFile); err != nil {
			log.WithError(err).WithField("devices-file", devicesFile).Fatal("Failed to save devices to file.")
		}
		log.WithField("devices-file", devicesFile).Info("Saved devices information to file")
	} else {
		log.Info("No new device found.")
	}
}
