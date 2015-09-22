// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"fmt"

	"github.com/hooklift/govix"
)

func main() {
	host, err := vix.Connect(vix.ConnectConfig{
		Provider: vix.VMWARE_WORKSTATION,
	})

	if err != nil {
		panic(err)
	}

	defer host.Disconnect()

	vm, err := host.OpenVm("/Users/camilo/Documents/Virtual Machines.localized/Ubuntu 64-bit.vmwarevm/Ubuntu 64-bit.vmx", "")
	if err != nil {
		panic(err)
	}

	// err = vm.AddNetworkAdapter(&vix.NetworkAdapter{
	// 	Vdevice:  vix.NETWORK_DEVICE_VMXNET3,
	// 	ConnType: vix.NETWORK_BRIDGED,
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// vm.RemoveNetworkAdapter(&vix.NetworkAdapter{Id: "7"})
	// vm.RemoveNetworkAdapter(&vix.NetworkAdapter{Id: "6"})
	// vm.RemoveNetworkAdapter(&vix.NetworkAdapter{Id: "5"})
	// vm.RemoveNetworkAdapter(&vix.NetworkAdapter{Id: "4"})
	// vm.RemoveNetworkAdapter(&vix.NetworkAdapter{Id: "3"})
	// vm.RemoveNetworkAdapter(&vix.NetworkAdapter{Id: "2"})
	// vm.RemoveNetworkAdapter(&vix.NetworkAdapter{Id: "1"})

	netAdapters, err := vm.NetworkAdapters()
	for _, adapter := range netAdapters {
		fmt.Printf("%#v\n", adapter)
	}

	// toolState, err := vm.ToolsState()
	// if err != nil {
	// 	panic(err)
	// }

	// if toolState != vix.TOOLSSTATE_RUNNING {
	// 	fmt.Printf("VMware Tools is not present!!! %d", toolState)
	// }
}
