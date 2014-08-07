package terraform_vix

import (
	"github.com/c4milo/govix"
	"github.com/hashicorp/terraform/helper/config"
	"github.com/hashicorp/terraform/helper/diff"
	"github.com/hashicorp/terraform/terraform"
)

func resource_vix_sharedfolder_validation() *config.Validator {
	return &config.Validator{
		Required: []string{
			"host_path",
			"guest_path",
		},
		Optional: []string{
			"enable",
			"readonly",
		},
	}
}

func resource_vix_sharedfolder_create(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	//p := meta.(*ResourceProvider)

	return nil, nil
}

func resource_vix_sharedfolder_update(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	//p := meta.(*ResourceProvider)

	return nil, nil
}

func resource_vix_sharedfolder_destroy(
	s *terraform.ResourceState,
	meta interface{}) error {
	// p := meta.(*ResourceProvider)
	// client := p.client

	return nil
}

func resource_vix_sharedfolder_diff(
	s *terraform.ResourceState,
	c *terraform.ResourceConfig,
	meta interface{}) (*terraform.ResourceDiff, error) {

	b := &diff.ResourceBuilder{
		// We have to choose whether a change in an attribute triggers a new
		// resource creation or updates the existing resource.
		Attrs: map[string]diff.AttrType{
			"host_path":  diff.AttrTypeUpdate,
			"guest_path": diff.AttrTypeUpdate,
			"readonly":   diff.AttrTypeUpdate,
		},
	}

	return b.Diff(s, c)
}

func resource_vix_sharedfolder_refresh(
	s *terraform.ResourceState,
	meta interface{}) (*terraform.ResourceState, error) {

	return nil, nil
}

func resource_vix_sharedfolder_update_state(
	s *terraform.ResourceState,
	vswitch *vix.VSwitch) (*terraform.ResourceState, error) {

	return nil, nil
}
