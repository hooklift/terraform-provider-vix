// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"fmt"

	"github.com/cloudescape/govix"
)

func main() {
	host, err := vix.Connect(vix.ConnectConfig{
		Provider: vix.VMWARE_WORKSTATION,
	})

	if err != nil {
		panic(err)
	}

	defer host.Disconnect()

	fmt.Println("Searching for running vms...")

	urls, err := host.FindItems(vix.FIND_RUNNING_VMS)
	if err != nil {
		panic(err)
	}

	fmt.Println("host.findItems returned!")

	for _, url := range urls {
		vm, _ := host.OpenVm(url, "")
		fmt.Println("Url: " + url)
		vcpus, _ := vm.Vcpus()
		memsize, _ := vm.MemorySize()
		vmxpath, _ := vm.VmxPath()
		teampath, _ := vm.VmTeamPath()
		guestos, _ := vm.GuestOS()
		//features, _ := vm.Features()
		fmt.Printf("vcpus: %d\n", vcpus)
		fmt.Printf("memory: %d\n", memsize)
		fmt.Println("vmx file: " + vmxpath)
		fmt.Println("vmteam file: " + teampath)
		fmt.Println("guest os: " + guestos)
		//fmt.Println("vm features: " + features)
		if err != nil {
			panic(err)
		}
	}
}
