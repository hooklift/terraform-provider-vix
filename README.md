# VMware VIX provider for Terraform
[![Build Status](https://travis-ci.org/cloudescape/terraform-provider-vix.svg?branch=master)](https://travis-ci.org/cloudescape/terraform-provider-vix)

Allows you to define infrastructure for VMware Fusion, Workstation, Server and Player.

## Mailing list
**Google groups:** https://groups.google.com/group/terraform-provider-vix

## Requirements
* VMware Fusion or Workstation installed
* **Govix:** The library used to interface with VMware
* **Godep:** Dependency manager
* Terraform

The exact list of dependencies can be found in the `Godeps` file. To install dependencies run: `godep get`


## Development workflow
Make sure you exported `DYLD_LIBRARY_PATH` or `LD_LIBRARY_PATH` with a path pointing to `vendor/libvix`

1. Make changes in your local fork of terraform-provider-vix
3. Test your changes running `TF_LOG=1 terraform plan` or `TF_LOG=1 terraform apply` inside a directory that contains *.tf files declaring VIX resources.

## Provider configurations

### Minimal configuration
```
resource "vix_vm" "core02" {
	name = "core02"
	gui = true
	image {
        url = "https://github.com/c4milo/dobby-boxes/releases/download/stable/coreos-stable-vmware.box"
        checksum = "545fec52ef3f35eee6e906fae8665abbad62d2007c7655ffa2ff4133ea3038b8"
        checksum_type = "sha256"
    }
}
```

### A more advanced configuration
```
variable "password" {
     default: ""
}

provider "vix" {
    # valid options are: "fusion", "workstation", "serverv1", "serverv2", "player"
    # and "workstation_shared"
    product = "fusion"
    verify_ssl = false

    # clone_type can be "full" or "linked". Advantages of one over the other are described here:
    # https://www.vmware.com/support/ws5/doc/ws_clone_typeofclone.html
    clone_type = "linked"
}

resource "vix_vswitch" "vmnet10" {
    name = "vmnet10"
    nat = true
    dhcp = true
    range = "192.168.1.0/24"
    host_access = true
}

resource "vix_vm" "core01" {
    name = "core01"
    description = "Terraform VMWARE VIX test"

	# The provider will download, verify, decompress and untar the image. 
	# Ideally you will provide images that have VMware Tools installed already,
	# otherwise the provider will be considerably limited for what it can do.
    image {
        url = "https://github.com/c4milo/dobby-boxes/releases/download/stable/coreos-stable-vmware.box"
        checksum = "545fec52ef3f35eee6e906fae8665abbad62d2007c7655ffa2ff4133ea3038b8"
        checksum_type = "sha256"

        # If image is encrypted we need to provide a password
        password = "${var.password}"
    }

    cpus = 1
    # Memory sizes must be provided using IEC sizes such as: kib, ki, mib, mi, gib or gi.
    memory = "1.0gib"
    count = 1
    upgrade_vhardware = false
    tools_init_timeout = 30s

    # Be aware that GUI does not work if VM is encrypted
    gui = true

    # Whether to enable or disable all shared folders for this VM
    sharedfolders = true

    # Advanced configuration
    network_adapter {
        # type can be either "custom", "nat", "bridged" or "hostonly"
	    type = "custom"
	    mac_address = "00:50:56:aa:bb:cc"

	    # mac address type can be "static", "generated" or "vpx"
	    mac_address_type = "static"

	    # vswitch is only required when network type is "custom"
	    vswitch = ${vix_vswitch.vmnet10.name}
	    start_connected = true
	    driver = "vmxnet3"
	    wake_on_lan = false
    }

    # Minimal required
    network_adapter {
	    type = "bridged"
    }
    
    network_adapter {
        type = "nat"
        mac_address = "00:50:56:aa:bb:cc"
        mac_address_type = "static"
    }

    network_adapter {
        type = "hostonly"
    }

    shared_folder {
        enable = false
        name = "Dev1"
        guest_path = "/home/camilo/dev"
        host_path = "/Users/camilo/Development"
        readonly = false
    }
}
```


## Known issues
* When launching multiple VM resources, make sure all of them have the same GUI setting, otherwise a race condition will kick in and `terraform apply` will fail. This issue is being tracked here https://github.com/c4milo/terraform-provider-vix/issues/10

* Terraform `count` attribute does not work with `vix_vm` resources as Terraform does not provide a way to get the resource index, causing the provider to fail. This issue is being tracked here https://github.com/hashicorp/terraform/issues/141


