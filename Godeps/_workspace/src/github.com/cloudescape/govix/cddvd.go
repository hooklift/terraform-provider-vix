// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vix

import (
	"fmt"
	"strings"

	"github.com/hooklift/govmx"
)

// CD/DVD configuration
type CDDVDDrive struct {
	ID string
	// Either IDE, SCSI or SATA
	Bus vmx.BusType
	// Used only when attaching image files. Ex: ISO images
	// If you just want to attach a raw cdrom device leave it empty
	Filename string
}

// Attaches a CD/DVD drive to the virtual machine.
func (v *VM) AttachCDDVD(drive *CDDVDDrive) error {
	if running, _ := v.IsRunning(); running {
		return &VixError{
			Operation: "vm.AttachCDDVD",
			Code:      200000,
			Text:      "Virtual machine must be powered off in order to attach a CD/DVD drive.",
		}
	}

	// Loads VMX file in memory
	v.vmxfile.Read()
	model := v.vmxfile.model

	device := vmx.Device{}
	if drive.Filename != "" {
		device.Filename = drive.Filename
		device.Type = vmx.CDROM_IMAGE
	} else {
		device.Type = vmx.CDROM_RAW
		device.Autodetect = true
	}

	device.Present = true
	device.StartConnected = true

	if drive.Bus == "" {
		drive.Bus = vmx.IDE
	}

	switch drive.Bus {
	case vmx.IDE:
		model.IDEDevices = append(model.IDEDevices, vmx.IDEDevice{Device: device})
	case vmx.SCSI:
		model.SCSIDevices = append(model.SCSIDevices, vmx.SCSIDevice{Device: device})
	case vmx.SATA:
		model.SATADevices = append(model.SATADevices, vmx.SATADevice{Device: device})
	default:
		return &VixError{
			Operation: "vm.AttachCDDVD",
			Code:      200001,
			Text:      fmt.Sprintf("Unrecognized bus type: %s\n", drive.Bus),
		}
	}

	return v.vmxfile.Write()
}

// Detaches a CD/DVD device from the virtual machine
func (v *VM) DetachCDDVD(drive *CDDVDDrive) error {
	if running, _ := v.IsRunning(); running {
		return &VixError{
			Operation: "vm.DetachCDDVD",
			Code:      200002,
			Text:      "Virtual machine must be powered off in order to detach CD/DVD drive.",
		}
	}

	// Loads VMX file in memory
	err := v.vmxfile.Read()
	if err != nil {
		return err
	}

	model := v.vmxfile.model

	switch drive.Bus {
	case vmx.IDE:
		for i, device := range model.IDEDevices {
			if drive.ID == device.VMXID {
				// This method of removing the element avoids memory leaks
				copy(model.IDEDevices[i:], model.IDEDevices[i+1:])
				model.IDEDevices[len(model.IDEDevices)-1] = vmx.IDEDevice{}
				model.IDEDevices = model.IDEDevices[:len(model.IDEDevices)-1]
			}
		}
	case vmx.SCSI:
		for i, device := range model.SCSIDevices {
			if drive.ID == device.VMXID {
				copy(model.SCSIDevices[i:], model.SCSIDevices[i+1:])
				model.SCSIDevices[len(model.SCSIDevices)-1] = vmx.SCSIDevice{}
				model.SCSIDevices = model.SCSIDevices[:len(model.SCSIDevices)-1]
			}
		}
	case vmx.SATA:
		for i, device := range model.SATADevices {
			if drive.ID == device.VMXID {
				copy(model.SATADevices[i:], model.SATADevices[i+1:])
				model.SATADevices[len(model.SATADevices)-1] = vmx.SATADevice{}
				model.SATADevices = model.SATADevices[:len(model.SATADevices)-1]
			}
		}
	default:
		return &VixError{
			Operation: "vm.DetachCDDVD",
			Code:      200003,
			Text:      fmt.Sprintf("Unrecognized bus type: %s\n", drive.Bus),
		}
	}

	return v.vmxfile.Write()
}

// Returns an unordered slice of currently attached CD/DVD devices on any bus.
func (v *VM) CDDVDs() ([]*CDDVDDrive, error) {
	// Loads VMX file in memory
	err := v.vmxfile.Read()
	if err != nil {
		return nil, err
	}

	model := v.vmxfile.model

	var cddvds []*CDDVDDrive
	model.WalkDevices(func(d vmx.Device) {
		bus := BusTypeFromID(d.VMXID)

		if d.Type == vmx.CDROM_IMAGE || d.Type == vmx.CDROM_RAW {
			cddvds = append(cddvds, &CDDVDDrive{
				ID:       d.VMXID,
				Bus:      bus,
				Filename: d.Filename,
			})
		}
	})
	return cddvds, nil
}

func (v *VM) RemoveAllCDDVDDrives() error {
	drives, err := v.CDDVDs()
	if err != nil {
		return &VixError{
			Operation: "vm.RemoveAllCDDVDDrives",
			Code:      200004,
			Text:      fmt.Sprintf("Error listing CD/DVD Drives: %s\n", err),
		}
	}

	for _, d := range drives {
		err := v.DetachCDDVD(d)
		if err != nil {
			return &VixError{
				Operation: "vm.RemoveAllCDDVDDrives",
				Code:      200004,
				Text:      fmt.Sprintf("Error removing CD/DVD Drive %v, error: %s\n", d, err),
			}
		}
	}

	return nil
}

// Gets BusType from device ID
func BusTypeFromID(ID string) vmx.BusType {
	var bus vmx.BusType
	switch {
	case strings.HasPrefix(ID, string(vmx.IDE)):
		bus = vmx.IDE
	case strings.HasPrefix(ID, string(vmx.SCSI)):
		bus = vmx.SCSI
	case strings.HasPrefix(ID, string(vmx.SATA)):
		bus = vmx.SATA
	}

	return bus
}

// Returns the CD/DVD drive identified by ID
// This function depends entirely on how GoVMX identifies slice's elements
func (v *VM) CDDVD(ID string) (*CDDVDDrive, error) {
	err := v.vmxfile.Read()
	if err != nil {
		return nil, err
	}

	model := v.vmxfile.model
	bus := BusTypeFromID(ID)

	var filename string
	found := model.FindDevice(func(d vmx.Device) bool {
		if ID == d.VMXID {
			filename = d.Filename
		}
		return ID == d.VMXID
	}, bus)

	if !found {
		return nil, nil
	}

	return &CDDVDDrive{Bus: bus, Filename: filename}, nil
}
