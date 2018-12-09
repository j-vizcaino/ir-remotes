package devices

import (
	"fmt"
	"net"
	"os"

	"github.com/mixcode/broadlink"
)

type DeviceInfo struct {
	Name       string `json:"name"`
	UDPAddress string `json:"udpAddress"`
	MACAddress string `json:"macAddress"`
	Type       uint16 `json:"type"`
	TypeName   string `json:"typeName,omitempty"`
}

type DeviceInfoList []DeviceInfo

func NewDeviceInfo(name string, device broadlink.Device) DeviceInfo {
	model, _ := device.DeviceName()

	return DeviceInfo{
		Name:       name,
		UDPAddress: device.UDPAddr.String(),
		MACAddress: net.HardwareAddr(device.MACAddr).String(),
		Type:       device.Type,
		TypeName:   model,
	}
}

func (d *DeviceInfo) Broadlink() (broadlink.Device, error) {
	mac, err := net.ParseMAC(d.MACAddress)
	if err != nil {
		return broadlink.Device{}, fmt.Errorf("failed to parse MAC address, %s", err)
	}
	// Parse UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", d.UDPAddress)
	if err != nil {
		return broadlink.Device{}, fmt.Errorf("failed to parse UDP address, %s", err)
	}
	return broadlink.Device{
		Type:    d.Type,
		MACAddr: mac,
		UDPAddr: *udpAddr,
	}, nil
}

func (dl *DeviceInfoList) AddDevice(name string, device broadlink.Device) error {
	dev := NewDeviceInfo(name, device)

	for idx, d := range *dl {
		// Same MAC address -> update existing record
		if d.MACAddress == dev.MACAddress {
			(*dl)[idx] = dev
			return nil
		}
		if d.Name == dev.Name {
			return fmt.Errorf("device %s already exists but MAC address does not match (existing=%s, new=%s)", name, d.MACAddress, dev.MACAddress)
		}
	}
	*dl = append(*dl, dev)
	return nil
}

type DeviceInfoPredicate func(DeviceInfo) bool

func (dl DeviceInfoList) Find(predicate DeviceInfoPredicate) (DeviceInfo, bool) {
	for _, dev := range dl {
		if predicate(dev) {
			return dev, true
		}
	}
	return DeviceInfo{}, false
}

func (dl DeviceInfoList) Initialize() ([]broadlink.Device, error) {
	out := make([]broadlink.Device, 0, len(dl))
	myname, _ := os.Hostname() // Your local machine's name.
	myid := make([]byte, 15) // Must be 15 bytes long.

	for _, d := range dl {
		bd, err := d.Broadlink()
		if err != nil {
			return nil, err
		}

		if err := bd.Auth(myid, myname); err != nil {
			return nil, fmt.Errorf("failed to authenticate with device %s, addr %s, %s", d.Name, d.UDPAddress, err)
		}
		out = append(out, bd)
	}
	return out, nil
}
