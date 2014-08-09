package terraform_vix

import (
	"log"
	"strconv"
	"time"

	"github.com/c4milo/terraform_vix/provider"
	"github.com/hashicorp/terraform/flatmap"
	"github.com/hashicorp/terraform/helper/config"
	"github.com/hashicorp/terraform/helper/diff"
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
		},
		Optional: []string{
			"description",
			"image.*.password",
			"cpus",
			"memory",
			"upgrade_vhardware",
			"tools_init_timeout",
			"networks.*",
			"sharedfolders",
			"sharedfolder.*",
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

	// networks := make([]string, len(vm.VSwitches))
	// for i, n := range vm.VSwitches {
	// 	networks[i] = n
	// }

	//Converts networks array to a map and merges it with rs.Attributes
	// flatmap.Map(rs.Attributes).Merge(flatmap.Flatten(map[string]interface{}{
	// 	"networks": networks,
	// }))

	// flatmap.Map(rs.Attributes).Merge(flatmap.Flatten(map[string]interface{}{
	// 	"image": vm.Image,
	// }))

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

	if err != nil {
		return err
	}

	// if raw := flatmap.Expand(rs.Attributes, "networks"); raw != nil {
	// 	if networks, ok := raw.([]interface{}); ok {
	// 		for _, n := range networks {
	// 			name, ok := n.(string)
	// 			if !ok {
	// 				continue
	// 			}

	// 			vm.VSwitches = append(vm.VSwitches, name)
	// 		}
	// 	}
	// }

	log.Printf("[DEBUG] image attrs -> %#v", rs.Attributes["image"])

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
	// Merge the diff into the state so that we have all the attributes
	// properly.
	rs := s.MergeDiff(d)

	vm := new(provider.VM)

	// Maps terraform.ResourceState attrbutes to provider.VM
	tf_to_vix(rs, vm)

	p := meta.(*ResourceProvider)
	vm.Provider = p.Config.Product
	vm.VerifySSL = p.Config.VerifySSL

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
	rs := s.MergeDiff(d)

	vm := new(provider.VM)

	// Maps terraform.ResourceState attrbutes to provider.VM
	tf_to_vix(rs, vm)

	p := meta.(*ResourceProvider)
	vm.Provider = p.Config.Product
	vm.VerifySSL = p.Config.VerifySSL

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
		},

		ComputedAttrs: []string{
			"ip_address",
		},
	}

	vm := new(provider.VM)
	vm.SetDefaults()

	// Sets defaults in TF raw configuration so that they show up when running
	// terraform plan on configuration files with only the minimum required
	// attributes.
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

	if !c.IsSet("mage.0.password") {
		c.Raw["image.0.password"] = ""
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

	// Refreshes only what makes sense, for example, it does not refresh
	// settings that modify the behavior of this provider
	s.Attributes["name"] = vm.Name
	s.Attributes["description"] = vm.Description
	s.Attributes["cpus"] = strconv.Itoa(int(vm.CPUs))
	s.Attributes["memory"] = vm.Memory

	//vix_to_tf(*vm, s)

	// log.Println("[DEBUG] New resource state: ")
	// for k, v := range s.Attributes {
	// 	log.Printf("[DEBUG] %s => %s\n", k, v)
	// }

	return s, nil
}
