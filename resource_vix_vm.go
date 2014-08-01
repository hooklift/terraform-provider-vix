package terraform_vix

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/c4milo/govix"
	"github.com/c4milo/terraform_vix/helper"
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

	name := rs.ID
	description := rs.Attributes["name"]
	image := flatmap.Expand(rs.Attributes, "image").([]interface{})
	cpus, err := strconv.ParseInt(rs.Attributes["cpus"], 0, 16)
	memory := rs.Attributes["memory"]
	hwversion, err := strconv.ParseInt(rs.Attributes["hardware_version"], 0, 8)
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
	log.Printf("[DEBUG] sharedfolders => %s", sharedfolders)

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	// FIXME(c4milo): There is an issue here whenever count is greater than 1
	// please see: https://github.com/hashicorp/terraform/issues/141
	imagePath := filepath.Join(usr.HomeDir, fmt.Sprintf(".terraform/vix/vms/%s", name))

	// Check if there is an image already in imagePath, if not, fetches it and unpacks it.
	finfo, err := os.Stat(imagePath)
	if err != nil {
		if os.IsNotExist(err) {
			imageGzipFile, err := helper.FetchImage(image["url"], image["checksum"], image["checksum_type"])
			if err != nil {
				return nil, err
			}

			err = helper.UnpackImage(imageGzFile, imagePath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Gets VIX instance
	p := meta.(*ResourceProvider)
	client := p.client
	vm, err := client.OpenVm(imagePath, image["password"])
	if err != nil {
		return nil, err
	}
	defer client.Disconnect()

	vm.SetMemorySize(memory)
	vm.SetNumberVcpus(cpus)

	for _, netType := range networks {
		adapter := vix.NetworkAdapter{
			ConnType:       netType,
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

		err = vm.AddNetworkAdapter(adapter)
		if err != nil {
			return nil, err
		}
	}

	// TODO(c4milo): Set hardware version

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
