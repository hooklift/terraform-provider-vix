package terraform_vix

import (
	"github.com/c4milo/govix"
	"github.com/hashicorp/terraform/helper/config"
	"github.com/hashicorp/terraform/terraform"
)

func resource_vix_vm_validation() *config.Validator {
	return &config.Validator{
		Required: []string{
			"image",
		},
		Optional: []string{
			"description",
		},
	}
}

func resource_vix_vm_create(
	s *terraform.ResourceState,
	d *terraform.ResourceDiff,
	meta interface{}) (*terraform.ResourceState, error) {
	//p := meta.(*ResourceProvider)

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

	return nil, nil
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
