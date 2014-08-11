# VMware VIX provider for Terraform
Allows you to define infrastructure for VMware Fusion, Workstation, Server and Player.

## Mailing list
**Google groups:** https://groups.google.com/group/terraform_vix

## Requirements
* VMware Fusion or Workstation installed
* **Govix:** The library used to interface with VMware
* **Godep:** Dependency manager
* **Terraform:** We are using our own fork for now, which is located at https://github.com/c4milo/terraform. Clone it and change `c4milo` for `hashicorp`

The exact list of dependencies can be found in the `Godeps` file. To install dependencies run: `godep get`


## Development workflow
1. Make changes in your local fork of terraform_vix
2. Compile Terraform from `$GOPATH/src/github.com/hashicorp/terraform`, running `make dev`
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

	# Images are required to be packaged using tar and gzip (`*.tar.gz`), the provider will download, verify,
	# decompress and untar the image. Ideally you will provide images that have VMware Tools installed already, 
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
	    mac_address = "00:00:00:00:00"

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

    shared_folder {
        enable = false
        name = "Dev1"
        guest_path = "/home/camilo/dev"
        host_path = "/Users/camilo/Development"
        readonly = false
    }
}
```

## To be aware of
Since Terraform at its current state does not load yet external plugins, we are building the plugin along with Terraform binary. The following were the changes made in Terraform to allow this:

* https://github.com/c4milo/terraform/commit/658f44dec9035ee9cef0b85f63d0c33c83acceea#diff-d41d8cd98f00b204e9800998ecf8427e

* https://github.com/c4milo/terraform/commit/3b3095e9834ea5da86f791da6f563439b74783cd#diff-d41d8cd98f00b204e9800998ecf8427e

* https://github.com/c4milo/terraform/commit/c4144afc8559050aa83a68a7f7d9518ada37cbbb#diff-d41d8cd98f00b204e9800998ecf8427e


## Known issues
* When launching multiple VM resources, make sure all of them have the same GUI setting, otherwise a race condition will kick in and `terraform apply` will fail. This issue is being tracked here https://github.com/c4milo/terraform_vix/issues/10

* Terraform `count` attribute does not work with `vix_vm` resources as Terraform does not provide a way to get the resource index, causing the provider to fail. This issue is being tracked here https://github.com/hashicorp/terraform/issues/141
