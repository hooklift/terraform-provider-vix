package terraform_vix

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/c4milo/govix"
	"github.com/c4milo/terraform_vix/provider"
	"github.com/hashicorp/terraform/flatmap"
	"github.com/hashicorp/terraform/helper/config"
	"github.com/hashicorp/terraform/helper/diff"
	"github.com/hashicorp/terraform/helper/multierror"
	"github.com/hashicorp/terraform/terraform"
)

func resource_vix_vm_validation() *config.Validator {
	return &config.Validator{
		Required: []string{
			"name",
			"image.*",
			"image.*.url",
			"image.*.checksum",
			"image.*.checksum_type",
			"network_adapter.*.type",
			"shared_folder.*.name",
		},
		Optional: []string{
			"description",
			"image.*.password",
			"cpus",
			"memory",
			"upgrade_vhardware",
			"tools_init_timeout",
			"sharedfolders",
			"sharedfolder.*",
			"network_adapter.*.mac_address",
			"network_adapter.*.vswitch",
			"network_adapter.*.driver",
			"shared_folder.*.enable",
			"shared_folder.*.guest_path",
			"shared_folder.*.host_path",
			"shared_folder.*.readonly",
			"gui",
		},
	}
}

// Maps provider attributes to Terraform's resource state
func vix_to_tf(vm provider.VM, rs *terraform.ResourceState) error {
	rs.Attributes["name"] = vm.Name
	rs.Attributes["description"] = vm.Description
	rs.Attributes["cpus"] = string(vm.CPUs)
	rs.Attributes["memory"] = vm.Memory
	rs.Attributes["tools_init_timeout"] = vm.ToolsInitTimeout.String()
	rs.Attributes["upgrade_vhardware"] = strconv.FormatBool(vm.UpgradeVHardware)
	rs.Attributes["gui"] = strconv.FormatBool(vm.LaunchGUI)
	rs.Attributes["sharedfolders"] = strconv.FormatBool(vm.SharedFolders)

	return nil
}

func net_tf_to_vix(rs *terraform.ResourceState, vm *provider.VM) error {
	var err error
	var errs []error

	if _, ok := rs.Attributes["network_adapter.#"]; !ok {
		return nil
	}

	tf_to_vix_network_type := func(attr string) (vix.NetworkType, error) {
		switch attr {
		case "bridged":
			return vix.NETWORK_BRIDGED, nil
		case "nat":
			return vix.NETWORK_NAT, nil
		case "hostonly":
			return vix.NETWORK_HOSTONLY, nil
		case "custom":
			return vix.NETWORK_CUSTOM, nil
		default:
			return "", fmt.Errorf("[ERROR] Invalid virtual network adapter type: %s", attr)
		}
	}

	tf_to_vix_virtual_device := func(attr string) (vix.VNetDevice, error) {
		switch attr {
		case "vlance":
			return vix.NETWORK_DEVICE_VLANCE, nil
		case "e1000":
			return vix.NETWORK_DEVICE_E1000, nil
		case "vmxnet3":
			return vix.NETWORK_DEVICE_VMXNET3, nil
		default:
			return "", fmt.Errorf("[ERROR] Invalid virtual network device: %s", attr)
		}
	}

	// tf_to_vix_vswitch := func(attr string) (vix.VSwitch, error) {
	// 	return vix.VSwitch{}, nil
	// }

	adapters := flatmap.Expand(
		rs.Attributes, "network_adapter").([]interface{})

	for _, adapter := range adapters {
		adapter := adapter.(map[string]interface{})
		vnic := new(vix.NetworkAdapter)

		if attr, ok := adapter["driver"].(string); ok && attr != "" {
			vnic.Vdevice, err = tf_to_vix_virtual_device(attr)
		}

		if attr, ok := adapter["mac_address"].(string); ok && attr != "" {
			vnic.MacAddress, err = net.ParseMAC(attr)

		}

		if attr, ok := adapter["type"].(string); ok && attr != "" {
			vnic.ConnType, err = tf_to_vix_network_type(attr)
		}

		// if attr, ok := adapter["vswitch"].(string); ok && attr != "" {
		// 	vnic.VSwitch, err = tf_to_vix_vswitch(attr)
		// }

		//log.Printf("[DEBUG] VNIC ==> %#v\n", vnic)
		vm.VNetworkAdapters = append(vm.VNetworkAdapters, vnic)

		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return &multierror.Error{Errors: errs}
	}

	return nil
}

func net_vix_to_tf(vm *provider.VM, rs *terraform.ResourceState) error {

	vix_to_tf_network_type := func(netType vix.NetworkType) string {
		switch netType {
		case vix.NETWORK_CUSTOM:
			return "custom"
		case vix.NETWORK_BRIDGED:
			return "bridged"
		case vix.NETWORK_HOSTONLY:
			return "hostonly"
		case vix.NETWORK_NAT:
			return "nat"
		default:
			return ""
		}
	}

	vix_to_tf_macaddress := func(adapter *vix.NetworkAdapter) string {
		static := adapter.MacAddress.String()
		generated := adapter.GeneratedMacAddress.String()

		if static != "" {
			return static
		}

		return generated
	}

	vix_to_tf_vdevice := func(vdevice vix.VNetDevice) string {
		switch vdevice {
		case vix.NETWORK_DEVICE_E1000:
			return "e1000"
		case vix.NETWORK_DEVICE_VLANCE:
			return "vlance"
		case vix.NETWORK_DEVICE_VMXNET3:
			return "vmxnet3"
		default:
			return ""
		}
	}

	numvnics := len(vm.VNetworkAdapters)
	if numvnics <= 0 {
		return nil
	}

	prefix := "network_adapter"

	rs.Attributes[prefix+".#"] = strconv.Itoa(numvnics)
	for i, adapter := range vm.VNetworkAdapters {
		attr := fmt.Sprintf("%s.%d.", prefix, i)
		rs.Attributes[attr+"type"] = vix_to_tf_network_type(adapter.ConnType)
		rs.Attributes[attr+"mac_address"] = vix_to_tf_macaddress(adapter)
		if adapter.ConnType == vix.NETWORK_CUSTOM {
			rs.Attributes[attr+"vswitch"] = "TODO(c4milo)"
		}
		rs.Attributes[attr+"driver"] = vix_to_tf_vdevice(adapter.Vdevice)
	}

	return nil
}

// Maps Terraform attributes to provider's structs
func tf_to_vix(rs *terraform.ResourceState, vm *provider.VM) error {
	var err error

	vm.Name = rs.Attributes["name"]
	vm.Description = rs.Attributes["description"]

	vcpus, err := strconv.ParseUint(rs.Attributes["cpus"], 0, 8)
	vm.CPUs = uint(vcpus)

	vm.Memory = rs.Attributes["memory"]
	vm.ToolsInitTimeout, err = time.ParseDuration(rs.Attributes["tools_init_timeout"])
	vm.UpgradeVHardware, err = strconv.ParseBool(rs.Attributes["upgrade_vhardware"])
	vm.LaunchGUI, err = strconv.ParseBool(rs.Attributes["gui"])
	vm.SharedFolders, err = strconv.ParseBool(rs.Attributes["sharedfolders"])

	// Maps any defined networks to VIX provider's data types
	net_tf_to_vix(rs, vm)

	if err != nil {
		return err
	}

	// This is nasty but there doesn't seem to be a cleaner way to extract stuff
	// from the TF configuration
	imgconf := flatmap.Expand(rs.Attributes, "image").([]interface{})[0].(map[string]interface{})

	image := provider.Image{
		URL:          imgconf["url"].(string),
		Checksum:     imgconf["checksum"].(string),
		ChecksumType: imgconf["checksum_type"].(string),
		Password:     imgconf["password"].(string),
	}
	vm.Image = image

	return nil
}

func resource_vix_vm_create(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	p := meta.(*ResourceProvider)

	// Merge the diff into the state so that we have all the attributes
	// properly.
	rs := s.MergeDiff(d)

	vm := new(provider.VM)
	vm.Provider = p.Config.Product
	vm.VerifySSL = p.Config.VerifySSL

	// Maps terraform.ResourceState attrbutes to provider.VM
	tf_to_vix(rs, vm)

	id, err := vm.Create()
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Resource ID: %s\n", id)
	rs.ID = id

	return rs, nil
}

func resource_vix_vm_update(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	p := meta.(*ResourceProvider)
	rs := s.MergeDiff(d)

	vm := new(provider.VM)
	vm.Provider = p.Config.Product
	vm.VerifySSL = p.Config.VerifySSL

	// Maps terraform.ResourceState attrbutes to provider.VM
	tf_to_vix(rs, vm)

	err := vm.Update(rs.ID)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func resource_vix_vm_destroy(
	s *terraform.ResourceState,
	meta interface{}) error {
	p := meta.(*ResourceProvider)
	vmxFile := s.ID

	vm := new(provider.VM)
	vm.Provider = p.Config.Product
	vm.VerifySSL = p.Config.VerifySSL

	vm.Image.Password = s.Attributes["password"]

	return vm.Destroy(vmxFile)
}

func resource_vix_vm_diff(
	s *terraform.ResourceState,
	c *terraform.ResourceConfig,
	meta interface{}) (*terraform.ResourceDiff, error) {

	b := &diff.ResourceBuilder{
		// We have to choose whether a change in an attribute triggers a new
		// resource creation or updates the existing resource.
		Attrs: map[string]diff.AttrType{
			"name":               diff.AttrTypeCreate,
			"description":        diff.AttrTypeUpdate,
			"tools_init_timeout": diff.AttrTypeUpdate,
			"image":              diff.AttrTypeCreate,
			"cpus":               diff.AttrTypeUpdate,
			"memory":             diff.AttrTypeUpdate,
			"networks":           diff.AttrTypeUpdate,
			"upgrade_vhardware":  diff.AttrTypeUpdate,
			"sharedfolders":      diff.AttrTypeUpdate,
			"gui":                diff.AttrTypeUpdate,
			"network_adapter":    diff.AttrTypeUpdate,
			"shared_folder":      diff.AttrTypeUpdate,
		},

		ComputedAttrs: []string{
			"ip_address",
		},
	}

	vm := new(provider.VM)
	vm.SetDefaults()

	// Sets defaults in TF raw configuration for minimal configurations so that
	// they show up when running terraform plan.
	if !c.IsSet("description") {
		c.Raw["description"] = vm.Description
	}

	if !c.IsSet("cpus") {
		c.Raw["cpus"] = strconv.Itoa(int(vm.CPUs))
	}

	if !c.IsSet("memory") {
		c.Raw["memory"] = vm.Memory
	}

	if !c.IsSet("tools_init_timeout") {
		c.Raw["tools_init_timeout"] = vm.ToolsInitTimeout.String()
	}

	if !c.IsSet("upgrade_vhardware") {
		c.Raw["upgrade_vhardware"] = strconv.FormatBool(vm.UpgradeVHardware)
	}

	if !c.IsSet("sharedfolders") {
		c.Raw["sharedfolders"] = strconv.FormatBool(vm.SharedFolders)
	}

	if !c.IsSet("gui") {
		c.Raw["gui"] = strconv.FormatBool(vm.LaunchGUI)
	}

	if !c.IsSet("image.0.password") {
		c.Raw["image.0.password"] = ""
	}

	// Sets defaults for network adapters so the plan can
	// show what it is really going to be created upon applying
	if adapters, ok := c.Get("network_adapter"); ok {
		adapters := adapters.([]map[string]interface{})

		for i, _ := range adapters {
			attr := "network_adapter." + strconv.Itoa(i) + "."
			if !c.IsSet(attr + "type") {
				c.Raw[attr+"type"] = "nat"
			}

			if !c.IsSet(attr + "driver") {
				c.Raw[attr+"driver"] = "e1000"
			}

			if !c.IsSet(attr + "mac_address") {
				b.ComputedAttrs = append(b.ComputedAttrs, attr+"mac_address")
			}
		}
	}

	return b.Diff(s, c)
}

func resource_vix_vm_refresh(
	s *terraform.ResourceState,
	meta interface{}) (*terraform.ResourceState, error) {
	p := meta.(*ResourceProvider)

	vmxFile := s.ID

	vm := new(provider.VM)
	vm.Provider = p.Config.Product
	vm.VerifySSL = p.Config.VerifySSL

	running, err := vm.Refresh(vmxFile)
	if err != nil {
		return nil, err
	}

	// This is to let TF know the resource is gone
	if !running {
		return nil, nil
	}

	// Refreshes only what makes sense, for example, we do not refresh settings
	// that modify the behavior of this provider
	s.Attributes["name"] = vm.Name
	s.Attributes["description"] = vm.Description
	s.Attributes["cpus"] = strconv.Itoa(int(vm.CPUs))
	s.Attributes["memory"] = vm.Memory

	err = net_vix_to_tf(vm, s)
	if err != nil {
		return nil, err
	}

	//vix_to_tf(*vm, s)

	return s, nil
}
