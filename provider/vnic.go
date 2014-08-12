package provider

// import (
// 	"fmt"

// 	"github.com/c4milo/govix"
// )

// // Virtual network adapter configuration
// type VNetworkAdapter struct {
// 	// The id of the network adapter
// 	Id uint
// 	// Virtual network adapter type: "nat", "bridged", "custom" or "hostonly"
// 	Type string
// 	// MAC address for this adapter, it must be within VMware approved range
// 	MACAddress string
// 	// MAC address type: "static", "generated", "vpx"
// 	MACAddressType string
// 	// vSwitch to where this adapter will be attached to
// 	VSwitch string
// 	// Whether the virtual network adapter starts connected when its associated
// 	// virtual machine powers on.
// 	//StartConnected bool
// 	// Network driver: "vmxnet3", "e1000", "vlance", "vmxnet2"
// 	Driver string
// 	// Whether wake-on-lan is enable or not for this virtual network adapter
// 	//WakeOnLan bool
// }

// func (vna *VNetworkAdapter) SetDefaults() {
// 	if vna.Type == "" {
// 		vna.Type = "nat"
// 	}

// 	if vna.Driver == "" {
// 		// Let's follow VMware by setting the same driver they set by default
// 		vna.Driver = "e1000"
// 	}

// 	// Let's defer to Govix setting the MACAddress and MACAddress type
// }

// func (vna *VNetworkAdapter) GetVixNetworkType() (vix.NetworkType, error) {
// 	switch vna.Type {
// 	case "bridged":
// 		return vix.NETWORK_BRIDGED, nil
// 	case "nat":
// 		return vix.NETWORK_NAT, nil
// 	case "hostonly":
// 		return vix.NETWORK_HOSTONLY, nil
// 	case "custom":
// 		return vix.NETWORK_CUSTOM, nil
// 	default:
// 		return "", fmt.Errorf("[ERROR] Invalid virtual network adapter type: %s", vna.Type)
// 	}
// }

// func (vna *VNetworkAdapter) GetVixDriver() (vix.VNetDevice, error) {

// }

// func (vna *VNetworkAdapter) Attach(vm *vix.VM) error {
// 	adapter := &vix.NetworkAdapter{}
// 	adapter.ConnType, err = vna.GetVixNetworkType()
// 	adapter.Id = vna.Id
// 	adapter.MacAddrType = vna.MACAddress
// 	adapter.MacAddress
// 	adapter.StartConnected
// 	adapter.VSwitch
// 	adapter.Vdevice
// 	adapter.WakeOnPcktRcv

// 	if err != nil {
// 		return err
// 	}

// 	return vm.AddNetworkAdapter(adapter)
// }

// func (vna *VNetworkAdapter) Detach(vm *vix.VM) error {
// 	return vm.RemoveNetworkAdapter(adapter)
// }
