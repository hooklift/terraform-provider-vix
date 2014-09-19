// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package provider

import "github.com/hashicorp/terraform/helper/schema"

func resourceVIXVSwitch() *schema.Resource {
	return &schema.Resource{
		Create: resourceVIXVSwitchCreate,
		Read:   resourceVIXVSwitchRead,
		Update: resourceVIXVSwitchUpdate,
		Delete: resourceVIXVSwitchDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"nat": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"dhcp": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"host_access": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"range": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceVIXVSwitchCreate(d *schema.ResourceData, meta interface{}) error {
	//config := meta.(*Config)

	return nil
}

func resourceVIXVSwitchRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVIXVSwitchUpdate(d *schema.ResourceData, meta interface{}) error {
	//config := meta.(*Config)

	return nil
}

func resourceVIXVSwitchDelete(d *schema.ResourceData, meta interface{}) error {
	//config := meta.(*Config)

	return nil
}
