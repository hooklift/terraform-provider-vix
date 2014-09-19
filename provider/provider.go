// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package provider

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
)

type Config struct {
	Product   string
	VerifySSL bool
}

func init() {
	// Terraform is already adding the timestamp for us
	log.SetFlags(log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid-%d-", os.Getpid()))
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"product": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"verify_ssl": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"vix_vm":      resourceVIXVM(),
			"vix_vswitch": resourceVIXVSwitch(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &Config{
		Product:   d.Get("product").(string),
		VerifySSL: d.Get("verify_ssl").(bool),
	}
	if config.Product == "" {
		log.Printf("[INFO] No product was configured, using 'workstation' by default.")
		config.Product = "workstation"
	}
	return config, nil
}
