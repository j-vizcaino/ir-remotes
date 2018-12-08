package devices

import (
	"fmt"
	"github.com/mixcode/broadlink"
	"net"
)

type Device struct {
	Name       string `json:"name"`
	UDPAddress string `json:"udpAddress"`
	MACAddress string `json:"macAddress"`
	Type       uint16 `json:"type"`
	TypeName   string `json:"typeName,omitempty"`
}

type DeviceList []Device

func NewFromBroadlink(name string, device broadlink.Device) Device {
	model, _ := device.DeviceName()

	return Device{
		Name:       name,
		UDPAddress: device.UDPAddr.String(),
		MACAddress: net.HardwareAddr(device.MACAddr).String(),
		Type:       device.Type,
		TypeName:   model,
	}
}

func (d *Device) Broadlink() (broadlink.Device, error) {
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

func (dl *DeviceList) AddDevice(name string, device broadlink.Device) error {
	dev := NewFromBroadlink(name, device)

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

type DevicePredicate func(Device) bool

func (dl DeviceList) Find(predicate DevicePredicate) (Device, bool) {
	for _, dev := range dl {
		if predicate(dev) {
			return dev, true
		}
	}
	return Device{}, false
}
