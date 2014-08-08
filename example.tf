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

/*
resource "vix_vswitch" "vmnet10" {
    name = "vmnet10"
    nat = true
    dhcp = true
    range = "192.168.1.0/24"
    host_access = true
}

resource "vix_vnic" "custom" {
    type = "custom"
    mac_address = ""
    mac_address_type = "static"
    vswitch = ${vix_vswitch.vmnet10}
    start_connected = true
    driver = "vmxnet3"
    wake_on_lan = false
}

resource "vix_sharedfolder" "myfolder" {
    enable = false
    name = "Dev1"
    guest_path = "/home/camilo/dev"
    host_path = "/Users/camilo/Development"
    readonly = false 
}
*/

resource "vix_vm" "coreos" {
    name = "core01"
    description = "Terraform VMWARE VIX test"

    image {
        url = "https://github.com/c4milo/dobby-boxes/releases/download/stable/coreos-stable-vmware.box"
        checksum = "545fec52ef3f35eee6e906fae8665abbad62d2007c7655ffa2ff4133ea3038b8"
        checksum_type = "sha256"

        # If image is encrypted we need to provide a password
        password = "${var.password}"
    }

    cpus = 1
    memory = "1gib"
    /*networks = [
        "${vix_vnic.custom}",
        "${vix_vnic.bridged}",
    ]*/

    count = 1
    upgrade_vhardware = false
    tools_init_timeout = 30s

    # Be aware that GUI does not work if VM is encrypted
    gui = true

    # Whether to enable or disable shared folders for this VM
    sharedfolders = true

    /*connection {
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
    }*/
}
