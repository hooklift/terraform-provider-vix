// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package provider

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	govix "github.com/cloudescape/govix"
	"github.com/cloudescape/terraform-provider-vix/provider/vix"

	"github.com/hashicorp/terraform/helper/multierror"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVixVm() *schema.Resource {
	return &schema.Resource{
		Create: resourceVixVmCreate,
		Read:   resourceVixVmRead,
		Update: resourceVixVmUpdate,
		Delete: resourceVixVmDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"cpus": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"memory": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"upgrade_vhardware": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},

			"tools_init_timeout": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"sharedfolders": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},

			"gui": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},

			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"image": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"checksum": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"checksum_type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"password": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"network_adapter": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"mac_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"mac_address_type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"vswitch": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"driver": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"shared_folder": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"enable": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
						"guest_path": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"host_path": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"readonly": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func net_tf_to_vix(d *schema.ResourceData, vm *vix.VM) error {
	tf_to_vix_virtual_device := func(attr string) (govix.VNetDevice, error) {
		switch attr {
		case "vlance":
			return govix.NETWORK_DEVICE_VLANCE, nil
		case "e1000":
			return govix.NETWORK_DEVICE_E1000, nil
		case "vmxnet3":
			return govix.NETWORK_DEVICE_VMXNET3, nil
		default:
			return "", fmt.Errorf("[ERROR] Invalid virtual network device: %s", attr)
		}
	}

	tf_to_vix_network_type := func(attr string) (govix.NetworkType, error) {
		switch attr {
		case "bridged":
			return govix.NETWORK_BRIDGED, nil
		case "nat":
			return govix.NETWORK_NAT, nil
		case "hostonly":
			return govix.NETWORK_HOSTONLY, nil
		case "custom":
			return govix.NETWORK_CUSTOM, nil
		default:
			return "", fmt.Errorf("[ERROR] Invalid virtual network adapter type: %s", attr)
		}
	}

	// tf_to_vix_vswitch := func(attr string) (vix.VSwitch, error) {
	// 	return vix.VSwitch{}, nil
	// }

	var err error
	var errs []error
	adaptersCount := d.Get("network_adapter.#").(int)
	vm.VNetworkAdapters = make([]*govix.NetworkAdapter, 0, adaptersCount)

	for i := 0; i < adaptersCount; i++ {
		prefix := fmt.Sprintf("network_adapter.%d.", i)
		adapter := new(govix.NetworkAdapter)

		if attr, ok := d.Get(prefix + "driver").(string); ok && attr != "" {
			adapter.Vdevice, err = tf_to_vix_virtual_device(attr)
		}

		if attr, ok := d.Get(prefix + "mac_address").(string); ok && attr != "" {
			// Only set a MAC address if it is declared as static
			// otherwise leave Govix to assign or continue using the generated
			// one.
			if addrtype, ok := d.Get(prefix + "mac_address_type").(string); ok && addrtype == "static" {
				adapter.MacAddress, err = net.ParseMAC(attr)
			}
		}

		if attr, ok := d.Get(prefix + "type").(string); ok && attr != "" {
			adapter.ConnType, err = tf_to_vix_network_type(attr)
		}

		// if attr, ok := adapter["vswitch"].(string); ok && attr != "" {
		// 	vnic.VSwitch, err = tf_to_vix_vswitch(attr)
		// }
		if err != nil {
			errs = append(errs, err)
		}

		log.Printf("[DEBUG] Network adapter: %+v\n", adapter)
		vm.VNetworkAdapters = append(vm.VNetworkAdapters, adapter)
	}

	if len(errs) > 0 {
		return &multierror.Error{Errors: errs}
	}

	return nil
}

// Maps Terraform attributes to provider's structs
func tf_to_vix(d *schema.ResourceData, vm *vix.VM) error {
	var err error

	vm.Name = d.Get("name").(string)
	vm.Description = d.Get("description").(string)
	vm.CPUs = uint(d.Get("cpus").(int))

	vm.Memory = d.Get("memory").(string)
	vm.UpgradeVHardware = d.Get("upgrade_vhardware").(bool)
	vm.LaunchGUI = d.Get("gui").(bool)
	vm.SharedFolders = d.Get("sharedfolders").(bool)

	vm.ToolsInitTimeout, err = time.ParseDuration(d.Get("tools_init_timeout").(string))

	// Maps any defined networks to VIX provider's data types
	net_tf_to_vix(d, vm)

	if err != nil {
		return err
	}

	if i := d.Get("image.#").(int); i > 0 {
		prefix := "image.0."
		vm.Image = vix.Image{
			URL:          d.Get(prefix + "url").(string),
			Checksum:     d.Get(prefix + "checksum").(string),
			ChecksumType: d.Get(prefix + "checksum_type").(string),
			Password:     d.Get(prefix + "password").(string),
		}
	}

	return nil
}

func net_vix_to_tf(vm *vix.VM, d *schema.ResourceData) error {

	vix_to_tf_network_type := func(netType govix.NetworkType) string {
		switch netType {
		case govix.NETWORK_CUSTOM:
			return "custom"
		case govix.NETWORK_BRIDGED:
			return "bridged"
		case govix.NETWORK_HOSTONLY:
			return "hostonly"
		case govix.NETWORK_NAT:
			return "nat"
		default:
			return ""
		}
	}

	vix_to_tf_macaddress := func(adapter *govix.NetworkAdapter) string {
		static := adapter.MacAddress.String()
		generated := adapter.GeneratedMacAddress.String()

		if static != "" {
			return static
		}

		return generated
	}

	vix_to_tf_vdevice := func(vdevice govix.VNetDevice) string {
		switch vdevice {
		case govix.NETWORK_DEVICE_E1000:
			return "e1000"
		case govix.NETWORK_DEVICE_VLANCE:
			return "vlance"
		case govix.NETWORK_DEVICE_VMXNET3:
			return "vmxnet3"
		default:
			return ""
		}
	}

	numvnics := len(vm.VNetworkAdapters)
	if numvnics <= 0 {
		return nil
	}

	prefix := "network_adapter"

	d.Set(prefix+".#", strconv.Itoa(numvnics))
	for i, adapter := range vm.VNetworkAdapters {
		attr := fmt.Sprintf("%s.%d.", prefix, i)
		d.Set(attr+"type", vix_to_tf_network_type(adapter.ConnType))
		d.Set(attr+"mac_address", vix_to_tf_macaddress(adapter))
		if adapter.ConnType == govix.NETWORK_CUSTOM {
			d.Set(attr+"vswitch", "TODO(c4milo)")
		}
		d.Set(attr+"driver", vix_to_tf_vdevice(adapter.Vdevice))
	}

	return nil
}

func resourceVixVmCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	vm := new(vix.VM)
	vm.Provider = config.Product
	vm.VerifySSL = config.VerifySSL

	if err := tf_to_vix(d, vm); err != nil {
		return err
	}

	id, err := vm.Create()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Resource ID: %s\n", id)
	d.SetId(id)

	// Initialize the connection info
	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": vm.IPAddress,
	})

	return resourceVixVmRead(d, meta)
}

func resourceVixVmUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	vm := new(vix.VM)
	vm.Provider = config.Product
	vm.VerifySSL = config.VerifySSL

	// Maps terraform.ResourceState attrbutes to vix.VM
	tf_to_vix(d, vm)

	err := vm.Update(d.Id())
	if err != nil {
		return err
	}

	return resourceVixVmRead(d, meta)
}

func resourceVixVmDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	vmxFile := d.Id()

	vm := new(vix.VM)
	vm.Provider = config.Product
	vm.VerifySSL = config.VerifySSL

	if password := d.Get("password"); password != nil {
		vm.Image.Password = d.Get("password").(string)
	}

	return vm.Destroy(vmxFile)
}

func resourceVixVmRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	vmxFile := d.Id()

	vm := new(vix.VM)
	vm.Provider = config.Product
	vm.VerifySSL = config.VerifySSL

	running, err := vm.Refresh(vmxFile)
	if err != nil {
		return err
	}

	// This is to let TF know the resource is gone
	if !running {
		return nil
	}

	// Refreshes only what makes sense, for example, we do not refresh settings
	// that modify the behavior of this provider
	d.Set("name", vm.Name)
	d.Set("description", vm.Description)
	d.Set("cpus", vm.CPUs)
	d.Set("memory", vm.Memory)
	d.Set("ip_address", vm.IPAddress)

	err = net_vix_to_tf(vm, d)
	if err != nil {
		return err
	}

	//vix_to_tf(*vm, s)

	return nil
}
