package terraform_vix

import (
	"fmt"
	"log"
	"os/user"
	"path/filepath"
	"strconv"

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
			"hardware_version",
			"network_driver",
			"networks.*",
			"sharedfolders",
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

	name := "coreos"
	description := rs.Attributes["description"]
	cpus, err := strconv.ParseUint(rs.Attributes["cpus"], 0, 8)
	memory := rs.Attributes["memory"]
	hwversion, err := strconv.ParseUint(rs.Attributes["hardware_version"], 0, 8)
	netdrv := rs.Attributes["network_driver"]
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

	log.Printf("[DEBUG] name => %s", name)
	log.Printf("[DEBUG] description => %s", description)
	log.Printf("[DEBUG] image => %v", image)
	log.Printf("[DEBUG] cpus => %d", cpus)
	log.Printf("[DEBUG] memory => %s", memory)
	log.Printf("[DEBUG] hwversion => %d", hwversion)
	log.Printf("[DEBUG] netdrv => %s", netdrv)
	log.Printf("[DEBUG] sharedfolders => %t", sharedfolders)

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	// FIXME(c4milo): There is an issue here whenever count is greater than 1
	// please see: https://github.com/hashicorp/terraform/issues/141
	vmPath := filepath.Join(usr.HomeDir, fmt.Sprintf(".terraform/vix/vms/%s", name))
	imagePath := filepath.Join(usr.HomeDir, fmt.Sprintf(".terraform/vix/images"))

	imageConfig := helper.Image{
		URL:          image["url"].(string),
		Checksum:     image["checksum"].(string),
		ChecksumType: image["checksum_type"].(string),
		DownloadPath: imagePath,
	}

	file, err := helper.FetchImage(imageConfig)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = helper.UnpackImage(file, vmPath)
	if err != nil {
		return nil, err
	}

	// Gets VIX instance
	p := meta.(*ResourceProvider)
	client := p.client

	// TODO(c4milo): Lookup VMX file in imagePath
	log.Printf("[INFO] Opening virtual machine from %s", imagePath)

	vm, err := client.OpenVm(imagePath, image["password"].(string))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect()

	memoryInMb, err := humanize.ParseBytes(memory)
	if err != nil {
		log.Printf("[WARN] Unable to set memory size, defaulting to 1g: %s", err)
		memoryInMb = 1024
	} else {
		memoryInMb /= 1024
	}

	log.Printf("[DEBUG] Setting memory size to %d megabytes", memoryInMb)
	vm.SetMemorySize(uint(memoryInMb))

	log.Printf("[DEBUG] Setting vcpus to %d", cpus)
	vm.SetNumberVcpus(uint8(cpus))

	for _, netType := range networks {
		adapter := &vix.NetworkAdapter{
			VSwitch:        vix.VSwitch{},
			StartConnected: true,
		}

		switch netdrv {
		case "e1000":
			adapter.Vdevice = vix.NETWORK_DEVICE_E1000
		case "vmxnet3":
			adapter.Vdevice = vix.NETWORK_DEVICE_VMXNET3
		default:
			adapter.Vdevice = vix.NETWORK_DEVICE_E1000
		}

		switch netType {
		case "hostonly":
			adapter.ConnType = vix.NETWORK_HOSTONLY
		case "bridged":
			adapter.ConnType = vix.NETWORK_BRIDGED
		case "nat":
			adapter.ConnType = vix.NETWORK_NAT
		default:
			adapter.ConnType = vix.NETWORK_CUSTOM

		}

		err = vm.AddNetworkAdapter(adapter)
		if err != nil {
			return nil, err
		}
	}

	// TODO(c4milo): Set hardware version

	log.Println("[INFO] Powering virtual machine on...")
	err = vm.PowerOn(vix.VMPOWEROP_NORMAL)
	if err != nil {
		return rs, err
	}

	// rs.ConnInfo["type"] = "ssh"
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
			"description":      diff.AttrTypeUpdate,
			"image":            diff.AttrTypeCreate,
			"cpus":             diff.AttrTypeUpdate,
			"memory":           diff.AttrTypeUpdate,
			"networks":         diff.AttrTypeUpdate,
			"hardware_version": diff.AttrTypeUpdate,
			"network_driver":   diff.AttrTypeUpdate,
			"sharedfolders":    diff.AttrTypeUpdate,
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
