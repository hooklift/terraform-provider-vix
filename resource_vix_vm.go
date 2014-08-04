package terraform_vix

import (
	"fmt"
	"log"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"github.com/c4milo/govix"
	"github.com/c4milo/terraform_vix/helper"
	"github.com/dustin/go-humanize"
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
			"network_driver",
			"networks.*",
			"sharedfolders",
			"sharedfolder.*",
			"gui",
		},
	}
}

func resource_vix_vm_create(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	// Merge the diff into the state so that we have all the attributes
	// properly.
	rs := s.MergeDiff(d)

	name := rs.Attributes["name"]
	description := rs.Attributes["description"]
	cpus, err := strconv.ParseUint(rs.Attributes["cpus"], 0, 8)
	memory := rs.Attributes["memory"]
	toolsInitTimeout, err := time.ParseDuration(rs.Attributes["tools_init_timeout"])
	upgradehw, err := strconv.ParseBool(rs.Attributes["upgrade_vhardware"])
	//netdrv := rs.Attributes["network_driver"]
	launchGUI, err := strconv.ParseBool(rs.Attributes["gui"])
	sharedfolders, err := strconv.ParseBool(rs.Attributes["sharedfolders"])
	var networks []string

	if err != nil {
		return nil, err
	}

	if raw := flatmap.Expand(rs.Attributes, "networks"); raw != nil {
		if nets, ok := raw.([]interface{}); ok {
			for _, net := range nets {
				str, ok := net.(string)
				if !ok {
					continue
				}

				networks = append(networks, str)
			}
		}
	}

	// This is nasty but there doesn't seem to be a cleaner way to extract stuff
	// from the TF configuration
	image := flatmap.Expand(rs.Attributes, "image").([]interface{})[0].(map[string]interface{})

	log.Printf("[DEBUG] networks => %v", networks)

	if len(networks) == 0 {
		networks = append(networks, "bridged")
	}

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	// FIXME(c4milo): There is an issue here whenever count is greater than 1
	// please see: https://github.com/hashicorp/terraform/issues/141
	vmPath := filepath.Join(usr.HomeDir, fmt.Sprintf(".terraform/vix/vms/%s", name))

	imageConfig := helper.FetchConfig{
		URL:          image["url"].(string),
		Checksum:     image["checksum"].(string),
		ChecksumType: image["checksum_type"].(string),
		DownloadPath: vmPath,
	}

	vmPath, err = helper.FetchFile(imageConfig)
	if err != nil {
		return nil, err
	}

	// FIXME(c4milo): This has an edge case when a resource with the same
	// name is declared with a different image box, it will return multiple
	// vmx files.
	pattern := filepath.Join(vmPath, "/**/*.vmx")

	log.Printf("[DEBUG] Finding VMX file in %s", pattern)
	files, _ := filepath.Glob(pattern)

	log.Printf("[DEBUG] VMX files found %v", files)

	if len(files) == 0 {
		return nil, fmt.Errorf("[ERROR] VMX file was not found: %s", pattern)
	}

	// Gets VIX instance
	p := meta.(*ResourceProvider)
	client := p.client

	if ((client.Provider & vix.VMWARE_VI_SERVER) == 0) ||
		((client.Provider & vix.VMWARE_SERVER) == 0) {
		log.Printf("[INFO] Registering VM in host's inventory...")
		err = client.RegisterVm(files[0])
		if err != nil {
			return nil, err
		}
	}

	log.Printf("[INFO] Opening virtual machine from %s", files[0])

	vm, err := client.OpenVm(files[0], image["password"].(string))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect()

	memoryInMb, err := humanize.ParseBytes(memory)
	if err != nil {
		log.Printf("[WARN] Unable to set memory size, defaulting to 1g: %s", err)
		memoryInMb = 1024
	} else {
		memoryInMb = (memoryInMb / 1024) / 1024
	}

	log.Printf("[DEBUG] Setting memory size to %d megabytes", memoryInMb)
	vm.SetMemorySize(uint(memoryInMb))

	log.Printf("[DEBUG] Setting vcpus to %d", cpus)
	vm.SetNumberVcpus(uint8(cpus))

	log.Printf("[DEBUG] Setting annotation to %s", description)
	vm.SetAnnotation(description)

	// for _, netType := range networks {
	// 	adapter := &vix.NetworkAdapter{
	// 		//VSwitch:        vix.VSwitch{},
	// 		StartConnected: true,
	// 	}

	// 	switch netdrv {
	// 	case "e1000":
	// 		adapter.Vdevice = vix.NETWORK_DEVICE_E1000
	// 	case "vmxnet3":
	// 		adapter.Vdevice = vix.NETWORK_DEVICE_VMXNET3
	// 	default:
	// 		adapter.Vdevice = vix.NETWORK_DEVICE_E1000
	// 	}

	// 	switch netType {
	// 	case "hostonly":
	// 		adapter.ConnType = vix.NETWORK_HOSTONLY
	// 	case "bridged":
	// 		adapter.ConnType = vix.NETWORK_BRIDGED
	// 	case "nat":
	// 		adapter.ConnType = vix.NETWORK_NAT
	// 	default:
	// 		adapter.ConnType = vix.NETWORK_CUSTOM

	// 	}

	// 	err = vm.AddNetworkAdapter(adapter)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	running, err := vm.IsRunning()
	if err != nil {
		return nil, err
	}

	if !running {
		if upgradehw &&
			((client.Provider & vix.VMWARE_PLAYER) == 0) {

			log.Println("[INFO] Upgrading virtual hardware...")
			err = vm.UpgradeVHardware()
			if err != nil {
				return nil, err
			}
		}

		log.Println("[INFO] Powering virtual machine on...")
		var options vix.VMPowerOption

		if launchGUI {
			log.Println("[INFO] Preparing to launch GUI...")
			options |= vix.VMPOWEROP_LAUNCH_GUI
		}

		options |= vix.VMPOWEROP_NORMAL

		err = vm.PowerOn(options)
		if err != nil {
			return rs, err
		}

		log.Println("[INFO] Waiting for VMware Tools to initialize...")
		err = vm.WaitForToolsInGuest(toolsInitTimeout)
		if err != nil {
			log.Println("[WARN] VMware Tools initialization timed out.")
			if sharedfolders {
				log.Println("[WARN] Enabling shared folders is not possible.")
			}
			return rs, nil
		}

		if sharedfolders {
			log.Println("[DEBUG] Enabling shared folders...")

			err = vm.EnableSharedFolders(sharedfolders)
			if err != nil {
				return nil, err
			}
		}
	} else {
		log.Println("[INFO] Virtual machine is already powered on")
	}

	//rs.ConnInfo["type"] = "ssh"
	// rs.ConnInfo["host"] = ?

	return rs, nil
}

func resource_vix_vm_update(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	//p := meta.(*ResourceProvider)

	return nil, nil
}

func resource_vix_vm_destroy(
	s *terraform.ResourceState,
	meta interface{}) error {
	// p := meta.(*ResourceProvider)
	// client := p.client

	return nil
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
			"network_driver":     diff.AttrTypeUpdate,
			"sharedfolders":      diff.AttrTypeUpdate,
			"gui":                diff.AttrTypeUpdate,
		},

		ComputedAttrs: []string{
			"ip_address",
		},
	}

	return b.Diff(s, c)
}

func resource_vix_vm_refresh(
	s *terraform.ResourceState,
	meta interface{}) (*terraform.ResourceState, error) {

	return nil, nil
}

func resource_vix_vm_update_state(
	s *terraform.ResourceState,
	vm *vix.VM) (*terraform.ResourceState, error) {

	return nil, nil
}
