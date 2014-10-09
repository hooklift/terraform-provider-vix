// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vix

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// VMware Fusion and Workstation come with predefined virtual switches, tailored
// or the most common uses cases.
//
// This type defines the common networking scenarios available for you to choose from,
// when adding network adapters to your virtual machines.
type NetworkType string

const (
	// Host-only networking provides a network connection between the virtual machine
	// and the host computer, using a virtual Ethernet adapter that is visible to the
	// host operating system. This approach can be very useful if you need to set up an
	// isolated virtual network.
	//
	// If you use host-only networking, your virtual machine and the host-only
	// adapter are connected to a private TCP/IP network. Addresses on this network are
	// provided by the VMware DHCP server.
	//
	// Source: https://www.vmware.com/support/ws55/doc/ws_net_configurations_hostonly.html
	NETWORK_HOSTONLY NetworkType = "hostonly"

	// If you want to connect to the Internet or other TCP/IP network using the host
	// computer's dial-up networking or broadband connection and you are not able
	// to give your virtual machine an IP address on the external network, NAT
	// is often the easiest way to give your virtual machine access to that network.
	//
	// If you use NAT, your virtual machine does not have its own IP address on
	// the external network. Instead, a separate private network is set up
	// on the host computer. Your virtual machine gets an address on that network
	// from the VMware virtual DHCP server. The VMware NAT device passes network
	// data between one or more virtual machines and the external network. It
	// identifies incoming data packets intended for each virtual machine and
	// sends them to the correct destination.
	//
	// The NAT device is connected to the VMnet8 virtual switch. This means that
	// virtual machines connected to the NAT network also use the VMnet8 virtual switch.
	//
	// The NAT device acts as a DNS server for the virtual machines on the NAT
	// network. Actually, the NAT device is a DNS proxy and merely forwards DNS requests
	// from the virtual machines to a DNS server that is known by the host.
	// Responses come back to the NAT device, which then forwards them to the
	// virtual machines.
	//
	// Source: https://www.vmware.com/support/ws55/doc/ws_net_configurations_nat.html
	NETWORK_NAT NetworkType = "nat"

	// If your host computer is on an Ethernet network, this is often the easiest
	// way to give your virtual machine access to that network. Linux and Windows
	// hosts can use bridged networking to connect to both wired and wireless networks.
	//
	// If you use bridged networking, your virtual machine needs to have its own
	// identity on the network. For example, on a TCP/IP network, the virtual
	// machine needs its own IP address. Your network administrator can tell you
	// whether IP addresses are available for your virtual machine and what
	// networking settings you should use in the guest operating system. Generally,
	// your guest operating system may acquire an IP address and other network
	// details automatically from a DHCP server, or you may need to set the IP
	// address and other details manually in the guest operating system.
	//
	// If you use bridged networking, the virtual machine is a full participant
	// in the network. It has access to other machines on the network and can be
	// contacted by other machines on the network as if it were a physical computer
	// on the network.
	//
	// Source: https://www.vmware.com/support/ws55/doc/ws_net_configurations_bridged.html
	NETWORK_BRIDGED NetworkType = "bridged"

	// The virtual networking components provided by VMware Workstation make it
	// possible for you to create sophisticated virtual networks. The virtual networks
	// can be connected to one or more external networks, or they may run entirely
	// on the host computer.
	//
	// If you choose for adding a custom network adapter you also have to provide
	// the virtual switch to which it is going to be attached to.
	NETWORK_CUSTOM NetworkType = "custom"
)

// Represents the drivers supported by VMware
type VNetDevice string

const (
	// AMD PCnet32 network-card, compatible with old OSes
	NETWORK_DEVICE_VLANCE VNetDevice = "vlance"

	// VMXnet network-card, requires VMware Tools
	// NETWORK_DEVICE_VMXNET VNetDevice = "vmxnet"
	// Intel E1000 network-card, most driver compatible
	NETWORK_DEVICE_E1000 VNetDevice = "e1000"

	// VMXNet3 is the fastest network-card, requires VMware Tools
	// see: http://www.vmware.com/pdf/vsp_4_vmxnet3_perf.pdf
	// It also requires the virtual hardware version to be 7 or later
	NETWORK_DEVICE_VMXNET3 VNetDevice = "vmxnet3"
)

// The type of MAC addresses
type MacAddressType string

const (
	// Hard coded by you to a valid MAC address range that
	// is registered to VMware, Inc.
	NETWORK_MACADDRESSTYPE_STATIC MacAddressType = "static"

	// Autocreated by the MUI (will have a 00:0c:29 address)
	NETWORK_MACADDRESSTYPE_GENERATED MacAddressType = "generated"

	// Autocreated by vCenter (will have a 00:50:56 address)
	NETWORK_MACADDRESSTYPE_VPX MacAddressType = "vpx"
)

// A network adapter is usually attached to one virtual switch
type NetworkAdapter struct {
	// The identifier of the network adapter
	Id string

	// This field was made private while we decide whether or not to expose it
	// to users since it could cause some confusion.
	//
	// Whether or not the adapter will be make present to the VM
	present bool

	// bridged, nat, hostonly or custom
	ConnType NetworkType

	// The actual ethernet virtual hardware. e1000 by default
	Vdevice VNetDevice

	// Workstation 6 and higher only. Set to "true" to enable WakeOnLan functions
	// Don't specify unless you really need it. "false" by default
	WakeOnPcktRcv bool

	// Enables applications to seamlessly communicate when using bridged networking,
	// even when the Host computer moves to another network. For example,
	// communication between applications will continue seamlessly when you move
	// your laptop from a wired network to a wireless network.
	LinkStatePropagation bool

	// Address type of the MAC by default it is "generated", but if a MAC address
	// is defined it will be "static"
	MacAddrType MacAddressType

	// Used only when MacAddrType is NETWORK_MACADDRESSTYPE_STATIC
	// It also has to have a value within the MAC address range that
	// is registered to VMware, Inc: 00:50:56:00:00:00 - 00:50:56:3F:FF:FF.
	// Source: http://pubs.vmware.com/vsphere-4-esxi-installable-vcenter/index.jsp?topic=/com.vmware.vsphere.esxi_server_config.doc_41/esx_server_config/advanced_networking/c_setting_up_mac_addresses.html
	MacAddress net.HardwareAddr

	// If ConnType is NETWORK_CUSTOM, this field allows you to choose to which
	// virtual switch you want to plug this virtual adapter to. Ex: vmnet2
	VSwitch VSwitch

	// Whether or not the network adapter will be connected on boot
	StartConnected bool

	// Generated MAC Address by VMware. Not need to set for adding network adapters
	GeneratedMacAddress net.HardwareAddr

	// Generated MAC Address offset. Not need to set for adding network adapters
	GeneratedMacAddressOffset string

	// PCI Slot number generated by VMWare. Not need to set for adding network adapters
	PciSlotNumber string
}

// Adds a network adapter to the virtual machine
//
// The "adapter" parameter is optional. If not specified this function will add,
// by default, a network adapter with NAT support; autogenerated MAC address,
// and e1000 as network device.
//
// Be aware that the adapter will not show up in the VMware Preferences user
// interface immediatelly. Once the virtual machine is started the UI will pick it up.
func (v *VM) AddNetworkAdapter(adapter *NetworkAdapter) error {
	isVmRunning, err := v.IsRunning()
	if err != nil {
		return err
	}

	if isVmRunning {
		return &VixError{
			Code: 100000,
			Text: "The VM has to be powered off in order to change its vmx settings",
		}
	}

	vmxPath, err := v.VmxPath()
	if err != nil {
		return err
	}

	vmx, err := readVmx(vmxPath)
	if err != nil {
		return err
	}

	if adapter == nil {
		adapter = &NetworkAdapter{}
	}

	adapter.present = true

	if adapter.Vdevice == NETWORK_DEVICE_VMXNET3 {
		hwversion, err := strconv.Atoi(vmx["virtualhw.version"])
		if err != nil {
			return err
		}

		if hwversion < 7 {
			return &VixError{
				Operation: "vm.AddNetworkAdapter",
				Code:      100001,
				Text:      fmt.Sprintf("Virtual hardware version needs to be 7 or higher in order to use vmxnet3. Current hardware version: %d", hwversion),
			}
		}
	}

	if adapter.LinkStatePropagation && (adapter.ConnType != NETWORK_BRIDGED) {
		return &VixError{
			Operation: "vm.AddNetworkAdapter",
			Code:      100003,
			Text:      "Link state propagation is only permitted for bridged networks",
		}
	}

	macaddr := adapter.MacAddress.String()

	if macaddr != "" {
		if !strings.HasPrefix(macaddr, "00:50:56") {
			return &VixError{
				Operation: "vm.AddNetworkAdapter",
				Code:      100004,
				Text:      "Static MAC addresses have to start with VMware officially assigned prefix: 00:50:56",
			}
		}
	}

	nextId := v.nextNetworkAdapterId(vmx)
	prefix := fmt.Sprintf("ethernet%d.", nextId)

	vmx[prefix+"present"] = strings.ToUpper(strconv.FormatBool(adapter.present))

	if string(adapter.ConnType) != "" {
		vmx[prefix+"connectionType"] = string(adapter.ConnType)
	} else {
		//default
		vmx[prefix+"connectionType"] = "nat"
	}

	if string(adapter.Vdevice) != "" {
		vmx[prefix+"virtualDev"] = string(adapter.Vdevice)
	} else {
		//default
		vmx[prefix+"virtualDev"] = "e1000"
	}

	vmx[prefix+"wakeOnPcktRcv"] = strconv.FormatBool(adapter.WakeOnPcktRcv)

	if string(adapter.MacAddrType) != "" {
		vmx[prefix+"addressType"] = string(adapter.MacAddrType)
	} else {
		//default
		vmx[prefix+"addressType"] = "generated"
	}

	if macaddr != "" {
		vmx[prefix+"address"] = macaddr
		vmx[prefix+"addressType"] = "static"

		// Makes sure a generated MAC address is removed if it exists
		delete(vmx, "generatedAddress")
		delete(vmx, "generatedAddressOffset")
	}

	if adapter.LinkStatePropagation {
		vmx[prefix+"linkStatePropagation.enable"] = strconv.FormatBool(adapter.LinkStatePropagation)
	}

	if adapter.ConnType == NETWORK_CUSTOM {
		if !ExistVSwitch(adapter.VSwitch.id) {
			return &VixError{
				Operation: "vm.AddNetworkAdapter",
				Code:      100005,
				Text:      "VSwitch " + adapter.VSwitch.id + " does not exist",
			}
		}
		vmx[prefix+"vnet"] = string(adapter.VSwitch.id)
	}

	vmx[prefix+"startConnected"] = strconv.FormatBool(adapter.StartConnected)

	return writeVmx(vmxPath, vmx)
}

// Returns the next available ethernet Id, reusing ids if
// the ethernet adapter has "present" equal to "FALSE"
func (v *VM) nextNetworkAdapterId(vmx map[string]string) int {
	var nextId int = 0
	prefix := "ethernet"

	for key, _ := range vmx {
		if strings.HasPrefix(key, prefix) {
			ethN := strings.Split(key, ".")[0]
			number, _ := strconv.Atoi(strings.Split(ethN, prefix)[1])

			// If ethN is not present, its id is recycled
			if vmx[ethN+".present"] == "FALSE" {
				return number
			}

			if number > nextId {
				nextId = number
			}
		}
	}

	nextId += 1

	return nextId
}

// Returns the total number of network adapters in the VMX file
func (v *VM) totalNetworkAdapters(vmx map[string]string) int {
	var total int = 0
	prefix := "ethernet"

	for key, _ := range vmx {
		if strings.HasPrefix(key, prefix) {
			ethN := strings.Split(key, ".")[0]
			number, _ := strconv.Atoi(strings.Split(ethN, prefix)[1])

			if number > total {
				total = number
			}
		}
	}

	return total
}

// Removes all network adapters
func (v *VM) RemoveAllNetworkAdapters() error {
	vmxPath, err := v.VmxPath()
	if err != nil {
		return err
	}

	vmx, err := readVmx(vmxPath)
	if err != nil {
		return err
	}

	for key, _ := range vmx {
		if strings.HasPrefix(key, "ethernet") {
			delete(vmx, key)
		}
	}

	return writeVmx(vmxPath, vmx)
}

// Removes network adapter from VMX file that matches the ID in "adapter.Id"
func (v *VM) RemoveNetworkAdapter(adapter *NetworkAdapter) error {
	isVmRunning, err := v.IsRunning()
	if err != nil {
		return err
	}

	if isVmRunning {
		return &VixError{
			Operation: "vm.RemoveNetworkAdapter",
			Code:      100000,
			Text:      "The VM has to be powered off in order to change its vmx settings",
		}
	}

	vmxPath, err := v.VmxPath()
	if err != nil {
		return err
	}

	vmx, err := readVmx(vmxPath)
	if err != nil {
		return err
	}

	device := "ethernet" + adapter.Id

	for key, _ := range vmx {
		if strings.HasPrefix(key, device) {
			delete(vmx, key)
		}
	}

	vmx[device+".present"] = "FALSE"

	err = writeVmx(vmxPath, vmx)
	if err != nil {
		return err
	}

	return nil
}

// List current network adapters attached to the virtual
// machine.
func (v *VM) NetworkAdapters() ([]*NetworkAdapter, error) {
	vmxPath, err := v.VmxPath()
	if err != nil {
		return nil, err
	}

	vmx, err := readVmx(vmxPath)
	if err != nil {
		return nil, err
	}

	var adapters []*NetworkAdapter
	// VMX ethernet adapters seem to not be zero based
	for i := 1; i <= v.totalNetworkAdapters(vmx); i++ {
		id := strconv.Itoa(i)
		prefix := "ethernet" + id

		if vmx[prefix+".present"] == "FALSE" {
			continue
		}

		wakeOnPckRcv, _ := strconv.ParseBool(vmx[prefix+".wakeOnPcktRcv"])
		lnkStateProp, _ := strconv.ParseBool(vmx[prefix+".linkStatePropagation.enable"])
		present, _ := strconv.ParseBool(vmx[prefix+".present"])
		startConnected, _ := strconv.ParseBool(vmx[prefix+".startConnected"])
		address, _ := net.ParseMAC(vmx[prefix+".address"])
		genAddress, _ := net.ParseMAC(vmx[prefix+".generatedAddress"])
		vswitch, _ := GetVSwitch(vmx[prefix+".vnet"])

		adapter := &NetworkAdapter{
			Id:                        id,
			present:                   present,
			ConnType:                  NetworkType(vmx[prefix+".connectionType"]),
			Vdevice:                   VNetDevice(vmx[prefix+".virtualDev"]),
			WakeOnPcktRcv:             wakeOnPckRcv,
			LinkStatePropagation:      lnkStateProp,
			MacAddrType:               MacAddressType(vmx[prefix+".addressType"]),
			MacAddress:                address,
			VSwitch:                   vswitch,
			StartConnected:            startConnected,
			GeneratedMacAddress:       genAddress,
			GeneratedMacAddressOffset: vmx[prefix+".generatedAddressOffset"],
			PciSlotNumber:             vmx[prefix+".pciSlotNumber"],
		}

		adapters = append(adapters, adapter)
	}

	return adapters, nil
}

// Returns the VM IP Address as reported by VMWare Tools
func (v *VM) IPAddress() (string, error) {
	return v.ReadVariable(VM_GUEST_VARIABLE, "ip")
}
