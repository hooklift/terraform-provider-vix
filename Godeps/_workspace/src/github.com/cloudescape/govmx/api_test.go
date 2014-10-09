package vmx

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestFindDevice(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Join(".", "fixtures", "b.vmx"))
	ok(t, err)

	vm := new(VirtualMachine)

	err = Unmarshal([]byte(data), vm)

	// Test finding all devices on all controllers
	expectedDevices := 5
	devicesFound := 0
	vm.FindDevice(func(d Device) bool {
		devicesFound++
		return false
	})

	assert(t, devicesFound == expectedDevices, "%d != %d", devicesFound, expectedDevices)

	// Test finding one device
	ID := "ide1:0"
	var device Device
	vm.FindDevice(func(d Device) bool {
		if d.VMXID == ID {
			device = d
			return true
		}
		return false
	})

	assert(t, device.VMXID == ID, "%s != %s", device.VMXID, ID)
	assert(t, device.Type == CDROM_IMAGE, "%s != %s", device.Type, CDROM_IMAGE)

	// Test finding devices from a specific controller
	expectedSCSIDevices := 3
	foundSCSIDevices := 0
	vm.FindDevice(func(d Device) bool {
		foundSCSIDevices++
		return false
	}, SCSI)

	assert(t, expectedSCSIDevices == foundSCSIDevices, "%d != %d", expectedSCSIDevices, foundSCSIDevices)
}
