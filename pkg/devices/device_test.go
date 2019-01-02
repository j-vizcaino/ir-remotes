package devices

import (
	"github.com/mixcode/broadlink"
	"net"
	"testing"

	. "github.com/onsi/gomega"
)

func TestDeviceList_AddDevice(t *testing.T) {
	g := NewGomegaWithT(t)

	devList := DeviceInfoList{}
	g.Expect(devList).To(HaveLen(0))

	dev := broadlink.Device{
		MACAddr: []byte{0, 1, 2, 3, 4, 5},
		UDPAddr: net.UDPAddr{IP: net.ParseIP("1.1.1.1"), Port: 80},
		Type:    1234,
	}

	g.Expect(devList.AddDevice("foo", dev)).To(Succeed())
	g.Expect(devList).To(HaveLen(1))
	g.Expect(devList[0]).To(Equal(NewDeviceInfo("foo", dev)))

	// Update name, based on MAC address colision
	g.Expect(devList.AddDevice("bar", dev)).To(Succeed())
	g.Expect(devList).To(HaveLen(1))
	g.Expect(devList[0]).To(Equal(NewDeviceInfo("bar", dev)))

	// Add new device
	dev.MACAddr = []byte{5, 4, 3, 2, 1, 0}
	g.Expect(devList.AddDevice("boo", dev)).To(Succeed())
	g.Expect(devList).To(HaveLen(2))

	// Add conflict: will not replace existing bar since MAC address does not match
	g.Expect(devList.AddDevice("bar", dev)).To(HaveOccurred())
	g.Expect(devList).To(HaveLen(2))
}

func TestDeviceList_Find(t *testing.T) {
	g := NewGomegaWithT(t)

	devList := DeviceInfoList{}
	dev := broadlink.Device{
		MACAddr: []byte{0, 1, 2, 3, 4, 5},
		UDPAddr: net.UDPAddr{IP: net.ParseIP("1.1.1.1"), Port: 80},
		Type:    1234,
	}

	foo := NewDeviceInfo("foo", dev)
	g.Expect(devList.AddDevice("foo", dev)).To(Succeed())
	dev.MACAddr = []byte{5, 4, 3, 2, 1, 0}
	boo := NewDeviceInfo("boo", dev)
	g.Expect(devList.AddDevice("boo", dev)).To(Succeed())

	res, found := devList.Find(func(d DeviceInfo) bool { return d.Name == "boo" })
	g.Expect(found).To(BeTrue())
	g.Expect(res).To(Equal(boo))

	res, found = devList.Find(func(d DeviceInfo) bool { return d.MACAddress == foo.MACAddress })
	g.Expect(found).To(BeTrue())
	g.Expect(res).To(Equal(foo))

	res, found = devList.Find(func(d DeviceInfo) bool { return d.Name == "missing" })
	g.Expect(found).To(BeFalse())
	g.Expect(res).To(Equal(DeviceInfo{}))
}
