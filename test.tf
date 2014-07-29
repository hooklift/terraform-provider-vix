provider "vix" {
    product = "fusion"
}

/*
resource "vix_vswitch" "vmnet10" {
    nat = true
    dhcp = true
    range = "192.168.1.0/24"
    host_access = true
}
*/

resource "vix_vm" "ubuntu" {
    description = "Terraform VMWARE VIX manual test"

    image {
        url = "https://github.com/c4milo/dobby-boxes/releases/download/alpha/coreos-alpha-vmware.box"
        // If image is encrypted we need to provide credentials
        username = "test"
        password = "test"
    }

    cpus = 2
    memory = "1g"
    networks = [
        //"vmnet10", 
        "bridged", 
        //"nat"
    ]

    provisioner "remote-exec" {
        inline = [
            "sudo apt-get -y update",
            "sudo apt-get -y install nginx",
            "sudo service nginx start",
        ]
    }
}

output "address" {
  value = "${vix_vm.ubuntu.ip_address}"
}