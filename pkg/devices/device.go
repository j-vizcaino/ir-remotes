package devices

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/mixcode/broadlink"
)

// DeviceInfo holds the information to access a Broadlink device on the network.
type DeviceInfo struct {
	Name       string `json:"name"`
	UDPAddress string `json:"udpAddress"`
	MACAddress string `json:"macAddress"`
	Type       uint16 `json:"type"`
	TypeName   string `json:"typeName,omitempty"`
	device     *broadlink.Device
}

// DeviceInfoList represents a list of Broadlink device info.
type DeviceInfoList []*DeviceInfo

// NewDeviceInfo creates a structure holding the information from a Broadlink device, as well as a user provided name.
func NewDeviceInfo(name string, device broadlink.Device) *DeviceInfo {
	model, _ := device.DeviceName()

	return &DeviceInfo{
		Name:       name,
		UDPAddress: device.UDPAddr.String(),
		MACAddress: net.HardwareAddr(device.MACAddr).String(),
		Type:       device.Type,
		TypeName:   model,
	}
}

// createDevice prepares a new, uninitliased broadlink.Device from the device information
func (d *DeviceInfo) createDevice() error {
	mac, err := net.ParseMAC(d.MACAddress)
	if err != nil {
		return fmt.Errorf("failed to parse MAC address, %s", err)
	}
	// Parse UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", d.UDPAddress)
	if err != nil {
		return fmt.Errorf("failed to parse UDP address, %s", err)
	}
	d.device = &broadlink.Device{
		Type:    d.Type,
		MACAddr: mac,
		UDPAddr: *udpAddr,
	}
	return nil
}

// GetBroadlinkDevice returns the associated Broadlink device.
// The returned device may not be initialized or even created. Make sure to call InitializeDevice before calling that function.
func (d *DeviceInfo) GetBroadlinkDevice() *broadlink.Device {
	return d.device
}

// InitializeDevice initialize the device by creating a broadlink.Device and authenticating with it.
// Device communication timeout is provided as a parameter.
func (d *DeviceInfo) InitializeDevice(timeout time.Duration) error {
	if d.device == nil {
		if err := d.createDevice(); err != nil {
			return nil
		}
	}

	// Already auth'd
	if d.device.ID != 0 {
		return nil
	}

	hostname, _ := os.Hostname() // Your local machine's name.
	fakeID := make([]byte, 15)   // Must be 15 bytes long.

	d.device.Timeout = timeout

	if err := d.device.Auth(fakeID, hostname); err != nil {
		return fmt.Errorf("failed to authenticate with device %s, addr %s, %s", d.Name, d.UDPAddress, err)
	}
	return nil
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

type DeviceInfoPredicate func(*DeviceInfo) bool

func ByName(name string) DeviceInfoPredicate {
	return func(d *DeviceInfo) bool {
		return d.Name == name
	}
}

func (dl DeviceInfoList) Find(predicate DeviceInfoPredicate) (*DeviceInfo, bool) {
	for _, dev := range dl {
		if predicate(dev) {
			return dev, true
		}
	}
	return nil, false
}

func (dl DeviceInfoList) InitializeDevices(timeout time.Duration) error {
	for _, d := range dl {
		err := d.InitializeDevice(timeout)
		if err != nil {
			return err
		}
	}
	return nil
}
