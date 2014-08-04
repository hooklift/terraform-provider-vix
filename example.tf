/*
terraform apply \
    -var 'key_path=/home/camilo/.ssh/id_rsa' \
    -var 'password=test' \
*/

variable "key_path" {
    default = "/Users/camilo/.ssh/id_rsa"
}

variable "password" {
     default: ""
}

provider "vix" {
    product = "fusion"
    verify_ssl = false
}

resource "vix_vswitch" "vmnet10" {
    name = "vmnet10"
    nat = true
    dhcp = true
    range = "192.168.1.0/24"
    host_access = true
}

resource "vix_vm" "coreos" {
    name = "core01"
    description = "Terraform VMWARE VIX test"

    image {
        url = "https://github.com/c4milo/dobby-boxes/releases/download/stable/coreos-stable-vmware.box"
        checksum = "0a1dcd15da093f37965bbc8c12dd6085defe9e4883465be9e7e7e3b08e91d3dc"
        checksum_type = "sha256"

        # If image is encrypted we need to provide a password
        password = "${var.password}"
    }

    cpus = 1
    memory = "1g"
    networks = [
        "vmnet10",
        "bridged",
        # "nat"
    ]

    count = 1
    hardware_version = 10
    network_driver = "vmxnet3"

    # Whether to enable or disable shared folders for this VM
    sharedfolders = true

    sharedfolder {
        name = "Dev1"
        guest_path = "/home/camilo/dev"
        host_path = "/Users/camilo/Development"
        readonly = false
    }

    sharedfolder {
        name = "Dev2"
        guest_path = "/home/camilo/dev2"
        host_path = "/Users/camilo/Dropbox/Development"
        readonly = false
    }

    connection {
        # The default username for our Box image
        user = "c4milo"

        # The path to your keyfile
        key_file = "${var.key_path}"
    }

    provisioner "remote-exec" {
        inline = [
            "sudo apt-get -y update",
            "sudo apt-get -y install nginx",
            "sudo service nginx start",
        ]
    }
}

output "address" {
  value = "${vix_vm.coreos.ip_address}"
}