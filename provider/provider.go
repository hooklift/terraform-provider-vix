package provider

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"product": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
			},
			"verify_ssl": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"vix_vm": resourceVixVm(),
			//"vix_vswitch": resourceVixVSwitch(),
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
