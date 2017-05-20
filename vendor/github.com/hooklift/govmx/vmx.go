// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vmx

import (
	"bytes"
	"fmt"
	"strings"
)

// Maximum values for a virtual machine
// From http://www.vmware.com/pdf/vsphere5/r55/vsphere-55-configuration-maximums.pdf
const (
	MAX_VCPUS = 64
	// Megabytes
	MAX_MEMORY                   = 1000000 // 1tib
	MAX_IDE_ADAPTERS             = 2
	MAX_IDE_DEVICES_PER_ADAPTER  = 2
	MAX_SATA_ADAPTERS            = 4
	MAX_SATA_DEVICES_PER_ADAPTER = 30
	MAX_SCSI_ADAPTERS            = 4
	MAX_SCSI_DEVICES_PER_ADAPTER = 15
	MAX_VDISKS                   = 60
	// Megabytes
	MAX_VDISK_SIZE                 = 6200000 // 62tib
	MAX_FLOPPY_ADAPTERS            = 1
	MAX_FLOPPY_DEVICES             = 2
	MAX_VNICS                      = 10
	MAX_USB_ADAPTERS               = 1
	MAX_USB_DEVICES                = 20
	MAX_PARALLEL_PORTS             = 3
	MAX_SERIAL_PORTS               = 4
	MAX_REMOTE_CONSOLE_CONNECTIONS = 40
	// Megabytes
	MAX_VIDEO_MEMORY = 512
)

func init() {
	//log.SetFlags(log.Lshortfile)
}

// Marshal traverses the value v recursively.
// If an encountered value implements the Marshaler interface
// and is not a nil pointer, Marshal calls its MarshalVMX method
// to produce VMX.  The nil pointer exception is not strictly necessary
// but mimics a similar, necessary exception in the behavior of
// UnmarshalVMX.
func Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	if err := NewEncoder(&b).Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Takes VMX data and binds it to the Go value pointed by v
func Unmarshal(data []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(data), false).Decode(v)
}

// Parses struct tag
func parseTag(tag string) (string, bool, error) {
	if tag == "" {
		return "", false, nil
	}

	omitempty := false

	// Takes out first colon found
	parts := strings.Split(tag, ":")
	if len(parts) < 2 || parts[1] == "" {
		return "", omitempty, fmt.Errorf("Invalid tag: %s", tag)
	}

	if parts[1] == `""` {
		return "", omitempty, fmt.Errorf("Tag name is missing: %s", tag)
	}

	// Takes out double quotes
	parts2 := strings.Split(parts[1], `"`)
	if len(parts2) < 2 {
		return "", omitempty, fmt.Errorf("Tag name has to be enclosed in double quotes: %s", tag)
	}

	values := strings.Split(parts2[1], ",")
	if len(values) > 1 && values[1] == "omitempty" {
		omitempty = true

	}

	return values[0], omitempty, nil
}
