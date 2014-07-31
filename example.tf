/*
terraform apply \
    -var 'key_path=/home/camilo/.ssh/id_pub' \
    -var 'password=test' \
*/

variable "key_path" {
    default = "/Users/camilo/.ssh/id_rsa"
}

# variable "password" {
#     default: ""
# }


provider "vix" {
    product = "fusion"
    verify_ssl = false
}

/*
resource "vix_vswitch" "vmnet10" {
    nat = true
    dhcp = true
    range = "192.168.1.0/24"
    host_access = true
}
*/

resource "vix_vm" "coreos" {
    name = "coreos"
    description = "Terraform VMWARE VIX test"

    image {
        url = "https://github.com/c4milo/dobby-boxes/releases/download/alpha/coreos-alpha-vmware.box"
        checksum = "c791812465f2cda236da1132b9f651cc58d5a7120636e48d82f4cb1546877bbd"
        checksum_type = "sha256"

        # If image is encrypted we need to provide 
        # password = "${var.password}"
    }

    cpus = 2
    memory = "1g"
    networks = [
        #"vmnet10",
        "bridged", 
        #"nat"
    ]

    count = 1
    hardware_version = 10
    network_driver = "vmxnet3"

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