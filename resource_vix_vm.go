package terraform_vix

import (
	"log"

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
	//	client := p.client

	// Merge the diff into the state so that we have all the attributes
	// properly.
	rs := s.MergeDiff(d)

	// TODO(c4milo): Get image if it does not exist in ~/.terraform/vix/boxes
	image, ok := flatmap.Expand(rs.Attributes, "image").([]interface{})
	if ok {
		log.Printf("[DEBUG] Image ==> %v", image)
	}

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
