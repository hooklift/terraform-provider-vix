// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vmx

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	vm := new(VirtualMachine)
	data, err := ioutil.ReadFile(filepath.Join(".", "fixtures", "b.vmx"))
	ok(t, err)

	err = Unmarshal([]byte(data), vm)
	ok(t, err)
	//fmt.Printf("%+v\n", vm.PowerType)
	assert(t, vm.Vhardware.Version == 9, "vhwversion should be 9")
	assert(t, len(vm.Ethernet) == 3, "there should be 3 ethernet devices")
	assert(t, vm.NumvCPUs == 1, "there should be 1 vcpu")
	// fmt.Printf("%+v\n", vm.IDEDevices)
	// fmt.Printf("%+v\n", vm.SCSIDevices)
	//fmt.Printf("%+v\n", vm.USBDevices)
	assert(t, len(vm.IDEDevices) == 2, fmt.Sprintf("there should be 2 IDE devices, found %d", len(vm.IDEDevices)))
	assert(t, len(vm.SCSIDevices) == 3, fmt.Sprintf("there should be 3 SCSI devices, found %d", len(vm.SCSIDevices)))
	assert(t, len(vm.SATADevices) == 0, fmt.Sprintf("there should be 0 SATA controller, found %d", len(vm.SATADevices)))
	assert(t, len(vm.USBDevices) == 2, fmt.Sprintf("there should be 2 USB devices, found %d", len(vm.USBDevices)))

	data, err = ioutil.ReadFile(filepath.Join(".", "fixtures", "a.vmx"))
	vm2 := new(VirtualMachine)
	err = Unmarshal([]byte(data), vm2)
	//fmt.Printf("%+v\n", vm2.SCSIDevices)
	assert(t, len(vm2.SCSIDevices) == 2, fmt.Sprintf("%d != %d", len(vm2.SCSIDevices), 2))
	assert(t, vm2.SCSIDevices[0].VMXID != "", fmt.Sprintf("VMXID should not be empty: ->%s<-", vm2.SCSIDevices[0].VMXID))
	ok(t, err)
}
