// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package provider

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform/helper/resource"
)

// resourceMap is the mapping of resources we support to their basic
// operations. This makes it easy to implement new resource types.
var resourceMap *resource.Map

func init() {
	// Terraform is already adding the timestamp for us
	log.SetFlags(log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid-%d-", os.Getpid()))

	resourceMap = &resource.Map{
		Mapping: map[string]resource.Resource{
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
