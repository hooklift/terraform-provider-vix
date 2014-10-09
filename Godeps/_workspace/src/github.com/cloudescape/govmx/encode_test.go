// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vmx

import "testing"

func TestParsingTag(t *testing.T) {
	tests := []struct {
		tag       string
		name      string
		omitempty bool
		err       string
	}{
		{"vmx:displayname", "", false, "Tag name has to be enclosed in double quotes: vmx:displayname"},
		{"vmx:", "", false, "Invalid tag: vmx:"},
		{`vmx:""`, "", false, `Tag name is missing: vmx:""`},
		{"vm", "", false, "Invalid tag: vm"},
		{`vmx:"displayname,omitempty`, "displayname", true, ""},
		{`vmx:"displayname,blah"`, "displayname", false, ""},
		{`vmx:"-"`, "-", false, ""},
	}

	for _, tt := range tests {
		name, omitempty, err := parseTag(tt.tag)
		equals(t, tt.name, name)
		equals(t, tt.omitempty, omitempty)
		if err != nil {
			equals(t, tt.err, err.Error())
		} else {
			equals(t, tt.err, "")
		}
	}
}

func TestMarshal(t *testing.T) {
	type VM struct {
		Encoding     string `vmx:".encoding"`
		Annotation   string `vmx:"annotation"`
		Hwversion    uint8  `vmx:"virtualHW.version"`
		HwProdCompat string `vmx:"virtualHW.productCompatibility"`
		Memsize      uint   `vmx:"memsize"`
		Numvcpus     uint   `vmx:"numvcpus"`
		MemHotAdd    bool   `vmx:"mem.hotadd"`
		DisplayName  string `vmx:"displayName"`
		GuestOS      string `vmx:"guestOS"`
		Autoanswer   bool   `vmx:"msg.autoAnswer"`
	}

	vm := new(VM)
	vm.Encoding = "utf-8"
	vm.Annotation = "Test VM"
	vm.Hwversion = 10
	vm.HwProdCompat = "hosted"
	vm.Memsize = 1024
	vm.Numvcpus = 2
	vm.MemHotAdd = false
	vm.DisplayName = "test"
	vm.GuestOS = "other3xlinux-64"
	vm.Autoanswer = true

	data, err := Marshal(vm)
	ok(t, err)
	expected := `.encoding = "utf-8"
annotation = "Test VM"
virtualHW.version = "10"
virtualHW.productCompatibility = "hosted"
memsize = "1024"
numvcpus = "2"
mem.hotadd = "false"
displayName = "test"
guestOS = "other3xlinux-64"
msg.autoAnswer = "true"
`
	equals(t, expected, string(data))
}

func TestMarshalEmbedded(t *testing.T) {
	vm := new(VirtualMachine)
	vm.Encoding = "utf-8"
	vm.Annotation = "Test VM"
	vm.Vhardware = Vhardware{
		Version: 10,
		Compat:  "hosted",
	}
	vm.Memsize = 1024
	vm.NumvCPUs = 2
	vm.MemHotAdd = false
	vm.DisplayName = "test"
	vm.GuestOS = "other3xlinux-64"
	vm.Autoanswer = true

	data, err := Marshal(vm)
	ok(t, err)
	expected := `.encoding = "utf-8"
annotation = "Test VM"
virtualhw.version = "10"
virtualhw.productcompatibility = "hosted"
memsize = "1024"
numvcpus = "2"
displayname = "test"
guestos = "other3xlinux-64"
msg.autoanswer = "true"
sound.present = "false"
usb.present = "false"
rtc.diffFromUTC = "0"
`
	equals(t, expected, string(data))
}

func TestMarshalArray(t *testing.T) {
	vm := new(VirtualMachine)
	vm.Encoding = "utf-8"
	vm.Annotation = "Test VM"
	vm.Vhardware = Vhardware{
		Version: 9,
		Compat:  "hosted",
	}

	vm.Ethernet = []Ethernet{
		{
			StartConnected:       true,
			Present:              true,
			ConnectionType:       "bridged",
			VirtualDev:           "e1000",
			WakeOnPcktRcv:        false,
			AddressType:          "generated",
			LinkStatePropagation: true,
		},
		{
			StartConnected: true,
			Present:        true,
			ConnectionType: "nat",
			VirtualDev:     "e1000",
			WakeOnPcktRcv:  false,
			AddressType:    "generated",
		},
	}

	vm.USBDevices = []USBDevice{
		{
			Present: true,
			Speed:   2,
			Type:    "hub",
			Port:    1,
			Parent:  "1",
		},
		{
			Present: true,
			Type:    "hid",
			Port:    0,
			Parent:  "-1",
		},
	}

	vm.Memsize = 1024
	vm.NumvCPUs = 2
	vm.MemHotAdd = false
	vm.DisplayName = "test"
	vm.GuestOS = "other3xlinux-64"
	vm.Autoanswer = true

	vm.IDEDevices = []IDEDevice{
		{Device: Device{
			Present:  true,
			Type:     "cdrom-image",
			Filename: "coreos.img",
		}},
		{Device: Device{
			Present: true,
			Type:    "cdrom-raw",
		}},
		{Device: Device{
			Present:  true,
			Type:     "cdrom-image",
			Filename: "coreos-alpha.img",
		}},
		{Device: Device{
			Present:  true,
			Type:     "cdrom-image",
			Filename: "coreos-beta.img",
		}},
		{Device: Device{
			Present:  true,
			Type:     "cdrom-image",
			Filename: "coreos-beta2.img",
		}},
	}

	vm.SCSIDevices = []SCSIDevice{
		{Device: Device{
			Present:  true,
			Filename: "",
		}, VirtualDev: "lsilogic"},
		{Device: Device{
			Present:  true,
			Filename: "coreos_production_vmware_image.vmdk",
		}},
	}

	data, err := Marshal(vm)
	ok(t, err)
	expected := `.encoding = "utf-8"
annotation = "Test VM"
virtualhw.version = "9"
virtualhw.productcompatibility = "hosted"
memsize = "1024"
numvcpus = "2"
displayname = "test"
guestos = "other3xlinux-64"
msg.autoanswer = "true"
sound.present = "false"
usb.present = "false"
rtc.diffFromUTC = "0"
ethernet0.startconnected = "true"
ethernet0.present = "true"
ethernet0.connectiontype = "bridged"
ethernet0.virtualdev = "e1000"
ethernet0.addresstype = "generated"
ethernet0.linkstatepropagation.enable = "true"
ethernet1.startconnected = "true"
ethernet1.present = "true"
ethernet1.connectiontype = "nat"
ethernet1.virtualdev = "e1000"
ethernet1.addresstype = "generated"
ide0:0.present = "true"
ide0:0.devicetype = "cdrom-image"
ide0:0.filename = "coreos.img"
ide0:1.present = "true"
ide0:1.devicetype = "cdrom-raw"
ide1:0.present = "true"
ide1:0.devicetype = "cdrom-image"
ide1:0.filename = "coreos-alpha.img"
ide1:1.present = "true"
ide1:1.devicetype = "cdrom-image"
ide1:1.filename = "coreos-beta.img"
scsi0.present = "true"
scsi0.virtualdev = "lsilogic"
scsi0:0.present = "true"
scsi0:0.filename = "coreos_production_vmware_image.vmdk"
usb:0.present = "true"
usb:0.speed = "2"
usb:0.devicetype = "hub"
usb:0.port = "1"
usb:0.parent = "1"
usb:1.present = "true"
usb:1.devicetype = "hid"
usb:1.parent = "-1"
`
	equals(t, expected, string(data))
}
