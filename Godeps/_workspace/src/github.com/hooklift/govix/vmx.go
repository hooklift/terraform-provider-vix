// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vix

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/hooklift/govmx"
)

// Manages VMX file
type VMXFile struct {
	sync.Mutex
	model *vmx.VirtualMachine
	path  string
}

// Reads VMX file from disk and unmarshals it
func (vmxfile *VMXFile) Read() error {
	data, err := ioutil.ReadFile(vmxfile.path)
	if err != nil {
		return err
	}

	model := new(vmx.VirtualMachine)

	err = vmx.Unmarshal(data, model)
	if err != nil {
		return err
	}

	vmxfile.model = model

	return nil
}

// Marshals and writes VMX file to disk
func (vmxfile *VMXFile) Write() error {
	file, err := os.Create(vmxfile.path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := vmx.Marshal(vmxfile.model)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// TODO(c4milo): Legacy function, this is to be removed once we migrate
// the dependant code
func readVmx(path string) (map[string]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	vmx := make(map[string]string)

	for _, line := range strings.Split(string(data), "\n") {
		values := strings.Split(line, "=")
		if len(values) == 2 {
			vmx[strings.TrimSpace(values[0])] = strings.Trim(strings.TrimSpace(values[1]), `"`)
		}
	}

	return vmx, nil
}

// TODO(c4milo): Legacy function, this is to be removed once we migrate
// the dependant code
func writeVmx(path string, vmx map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	keys := make([]string, len(vmx))
	i := 0
	for k := range vmx {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	var buf bytes.Buffer
	for _, key := range keys {
		buf.WriteString(key + " = " + `"` + vmx[key] + `"`)
		buf.WriteString("\n")
	}

	if _, err = io.Copy(f, &buf); err != nil {
		return err
	}

	return nil
}
