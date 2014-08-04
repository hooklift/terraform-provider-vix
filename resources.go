package terraform_vix

import (
	"log"

	"github.com/hashicorp/terraform/helper/resource"
)

// resourceMap is the mapping of resources we support to their basic
// operations. This makes it easy to implement new resource types.
var resourceMap *resource.Map

func init() {
	// Terraform is already adding the timestamp for us
	log.SetFlags(0)

	resourceMap = &resource.Map{
		Mapping: map[string]resource.Resource{
			"vix_vm": resource.Resource{
				ConfigValidator: resource_vix_vm_validation(),
				Create:          resource_vix_vm_create,
				Destroy:         resource_vix_vm_destroy,
				Diff:            resource_vix_vm_diff,
				Refresh:         resource_vix_vm_refresh,
				Update:          resource_vix_vm_update,
			},

			"vix_vswitch": resource.Resource{
				ConfigValidator: resource_vix_vswitch_validation(),
				Create:          resource_vix_vswitch_create,
				Destroy:         resource_vix_vswitch_destroy,
				Diff:            resource_vix_vswitch_diff,
				Refresh:         resource_vix_vswitch_refresh,
				Update:          resource_vix_vswitch_update,
			},
		},
	}
}
