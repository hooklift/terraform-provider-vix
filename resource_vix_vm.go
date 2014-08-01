package terraform_vix

import (
	"log"
	"strconv"

	"github.com/c4milo/govix"
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
			"hardware_version",
			"network_driver",
			"networks.*",
		},
	}
}

func resource_vix_vm_create(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	//p := meta.(*ResourceProvider)
	//client := p.client

	// Merge the diff into the state so that we have all the attributes
	// properly.
	rs := s.MergeDiff(d)

	name := rs.Attributes["name"]
	description := rs.Attributes["name"]
	image := flatmap.Expand(rs.Attributes, "image").([]interface{})
	cpus, err := strconv.ParseInt(rs.Attributes["cpus"], 0, 16)
	memory := rs.Attributes["memory"]
	hwversion, err := strconv.ParseInt(rs.Attributes["hardware_version"], 0, 8)
	netdrv := rs.Attributes["network_driver"]
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

	log.Printf("[DEBUG] Name => %s", name)
	log.Printf("[DEBUG] Description => %s", description)
	log.Printf("[DEBUG] image => %v", image)
	log.Printf("[DEBUG] CPUs => %s", cpus)
	log.Printf("[DEBUG] Memory => %s", memory)
	log.Printf("[DEBUG] hwversion => %s", hwversion)
	log.Printf("[DEBUG] netdrv => %s", netdrv)

	// TODO: Check if there is an image already in ~/.terraform/vix/images/{.Name}
	// usr, err := user.Current()
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(usr.HomeDir)

	// path := ""
	// _, err := os.Stat(filepath.Join(usr.HomeDir, fmt.Sprintf(".terraform/vix/images/%s", ))
	// if err == nil {
	// 	return true, nil
	// }
	// if os.IsNotExist(err) {
	// 	return false, nil
	// }

	// TODO(c4milo): Get image
	// TODO(c4milo): Check image integrity
	// TODO(c4milo): Unpack it in ~/.terraform/vix/images. if it does exist, clone the box into images
	// TODO(c4milo): OpenVM passing vmx path

	// TODO(c4milo): Set memory
	// TODO(c4milo): Set cpus
	// TODO(c4milo): Set networks
	// TODO(c4milo): Set hardware version
	// TODO(c4milo): Set network driver

	return nil, nil
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
			"name":             diff.AttrTypeCreate,
			"description":      diff.AttrTypeUpdate,
			"image":            diff.AttrTypeCreate,
			"cpus":             diff.AttrTypeUpdate,
			"memory":           diff.AttrTypeUpdate,
			"networks":         diff.AttrTypeUpdate,
			"hardware_version": diff.AttrTypeUpdate,
			"network_driver":   diff.AttrTypeUpdate,
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
