// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vix

import (
	//"io/ioutil"
	"net"
	//"strings"
)

// Like a physical switch, a virtual switch lets you
// connect other networking components together.
// Virtual switches are created as needed by the
// VMware Workstation software, up to a total of
// nine switches. You can connect one or more
// virtual machines to a switch.
//
// By default, a few of the switches and the networks
// associated with them are used for special
// named configurations:
//
// The bridged network uses VMnet0.
// The host-only network uses VMnet1.
// The NAT network uses VMnet8.
// The other available networks are simply named VMnet2, VMnet3, VMnet4, and so on.
type VSwitch struct {
	id string

	// Whether or not to provide addresses
	// on this vswitch via DHCP
	DHCP bool

	// Address pool from which to provide IPs to
	// virtual machines plugged to this virtual switch.
	DCHPNetwork net.IPNet

	//Connects host machine to this virtual switch
	VirtualAdapter bool

	// Allow virtual machines on this virtual
	// switch to connect to external
	// networks using NAT
	NAT bool
}

/*
answer VNET_4_DHCP yes
answer VNET_4_VIRTUAL_ADAPTER yes
answer VNET_4_HOSTONLY_NETMASK 255.255.255.0
answer VNET_4_HOSTONLY_SUBNET 192.168.60.0
answer VNET_4_NAT yes
*/

//http://thornelabs.net/2013/10/18/manually-add-and-remove-vmware-fusion-virtual-adapters.html
func AddVSwitch(vswitch VSwitch) (string, error) {
	return "", nil
}

// answer VNET_10_DHCP no
// answer VNET_10_VIRTUAL_ADAPTER no
//http://thornelabs.net/2013/10/18/manually-add-and-remove-vmware-fusion-virtual-adapters.html
func RemoveVSwitch(id string) error {
	return nil
}

func ListVSwitches() ([]*VSwitch, error) {
	return nil, nil
}

func ExistVSwitch(id string) bool {
	return false
}

func GetVSwitch(id string) (VSwitch, error) {
	return VSwitch{}, nil
}

// Source http://kb.vmware.com/selfservice/microsites/search.do?language=en_US&cmd=displayKC&externalId=1026510
func restartVMNetServices() {
	// linux: vmware-networks –stop
	// linux: vmware-networks –start
	// linux: vmware-networks –status
	// linux: /etc/init.d/vmware restart
	//osx: vmnet-cli --stop
	//osx: vmnet-cli --configure
	//osx: vmnet-cli --start
	//osx: vmnet-cli --status
}

// linux: /etc/vmware/networking
// darwin: /Library/Preferences/VMware\ Fusion/networking
// windows: ?
func readNetworkCfg(path string) (map[string]string, error) {
	// data, err := ioutil.ReadFile(path)
	// if err != nil {
	// 	return nil, err
	// }

	// network := make(map[string]string)

	// for _, line := range strings.Split(string(data), "\n") {
	// 	values := strings.Split(line, " ")
	// 	if len(values) == 2 {
	// 		vmx[strings.TrimSpace(values[0])] = strings.Trim(strings.TrimSpace(values[1]), `"`)
	// 	}
	// }

	return nil, nil
}

func writeNetworkCfg(path string, network map[string]string) error {
	return nil
}
