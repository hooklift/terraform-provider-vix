package provider

import (
	"github.com/c4milo/govix"
	"github.com/hashicorp/terraform/helper/config"
	"github.com/hashicorp/terraform/helper/diff"
	"github.com/hashicorp/terraform/terraform"
)

func resource_vix_vswitch_validation() *config.Validator {
	return &config.Validator{
		Required: []string{
			"name",
		},

		Optional: []string{
			"nat",
			"dhcp",
			"range",
			"host_access",
		},
	}
}

func resource_vix_vswitch_create(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	//p := meta.(*ResourceProvider)

	return nil, nil
}

func resource_vix_vswitch_update(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	//p := meta.(*ResourceProvider)

	return nil, nil
}

func resource_vix_vswitch_destroy(
	s *terraform.ResourceState,
	meta interface{}) error {
	// p := meta.(*ResourceProvider)
	// client := p.client

	return nil
}

func resource_vix_vswitch_diff(
	s *terraform.ResourceState,
	c *terraform.ResourceConfig,
	meta interface{}) (*terraform.ResourceDiff, error) {

	b := &diff.ResourceBuilder{
		// We have to choose whether a change in an attribute triggers a new
		// resource creation or updates the existing resource.
		Attrs: map[string]diff.AttrType{
			"nat":         diff.AttrTypeUpdate,
			"dhcp":        diff.AttrTypeUpdate,
			"range":       diff.AttrTypeUpdate,
			"host_access": diff.AttrTypeUpdate,
		},
	}

	return b.Diff(s, c)
}

func resource_vix_vswitch_refresh(
	s *terraform.ResourceState,
	meta interface{}) (*terraform.ResourceState, error) {

	return nil, nil
}

func resource_vix_vswitch_update_state(
	s *terraform.ResourceState,
	vswitch *vix.VSwitch) (*terraform.ResourceState, error) {

	return nil, nil
}
