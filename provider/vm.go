package provider

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/c4milo/govix"
	"github.com/dustin/go-humanize"
)

// Virtual machine configuration
type VM struct {
	// Which VMware VIX service provider to use. ie: fusion, workstation, server, etc
	Provider string
	// Whether to verify SSL or not for remote connections in ESXi
	VerifySSL bool
	// Name of the virtual machine
	Name string
	// Description for the virtual machine, it is created as an annotation in
	// VMware.
	Description string
	// Image to use during the creation of this virtual machine
	Image Image
	// Number of virtual cpus
	CPUs uint
	// Memory size in megabytes.
	Memory string
	// Switches to where this machine is going to be attach to
	VSwitches []string
	// Whether to upgrade the VM virtual hardware
	UpgradeVHardware bool
	// The timeout to wait for VMware Tools to be initialized inside the VM
	ToolsInitTimeout time.Duration
	// Whether to launch the VM with graphical environment
	LaunchGUI bool
	// Whether to enable or disable shared folders for this VM
	SharedFolders bool
	// Network adapters
	VNetworkAdapters []*vix.NetworkAdapter
}

// Creates VIX instance with VMware
func (v *VM) client() (*vix.Host, error) {
	var p vix.Provider

	switch strings.ToLower(v.Provider) {
	case "fusion", "workstation":
		p = vix.VMWARE_WORKSTATION
	case "serverv1":
		p = vix.VMWARE_SERVER
	case "serverv2":
		p = vix.VMWARE_VI_SERVER
	case "player":
		p = vix.VMWARE_PLAYER
	case "workstation_shared":
		p = vix.VMWARE_WORKSTATION_SHARED
	default:
		p = vix.VMWARE_WORKSTATION
	}

	var options vix.HostOption
	if v.VerifySSL {
		options = vix.VERIFY_SSL_CERT
	}

	host, err := vix.Connect(vix.ConnectConfig{
		Provider: p,
		Options:  options,
	})

	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] VIX client configured for product: VMware %s. SSL: %t", v.Provider, v.VerifySSL)

	return host, nil
}

// Sets default values for VM attributes
func (v *VM) SetDefaults() {
	if v.CPUs <= 0 {
		v.CPUs = 1
	}

	if v.Memory == "" {
		v.Memory = "512mib"
	}

	if v.Description == "" {
		v.Description = "Machine was created using Terraform VIX provider"
	}

	if v.ToolsInitTimeout.Seconds() <= 0 {
		v.ToolsInitTimeout = time.Duration(30) * time.Second
	}
}

// Downloads, extracts and opens Gold virtual machine, then it creates a clone
// out of it.
func (v *VM) Create() (string, error) {
	log.Printf("[DEBUG] Creating VM resource...")

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	image := v.Image

	imgPath := filepath.Join(usr.HomeDir, ".terraform/vix/images", image.Checksum)
	if err = image.Download(imgPath); err != nil {
		return "", err
	}
	defer image.file.Close()

	goldPath := filepath.Join(usr.HomeDir, filepath.Join(".terraform/vix/gold", image.Checksum))
	_, err = os.Stat(goldPath)
	goldPathExist := err == nil || err != os.ErrNotExist

	// There is no need to get the error as the slice will be empty anyway
	finfo, _ := ioutil.ReadDir(goldPath)
	goldPathEmpty := len(finfo) == 0

	if !goldPathExist || goldPathEmpty {
		log.Printf("[DEBUG] Gold virtual machine does not exist or is empty: %s", goldPath)
		// TODO(c4milo): Make sure the file is a tgz file before attempting
		// to unpack it.
		_, err = image.Unpack(goldPath)
		if err != nil {
			return "", err
		}
	}

	pattern := filepath.Join(goldPath, "/**/*.vmx")

	log.Printf("[DEBUG] Finding Gold virtual machine vmx file in %s", pattern)
	files, _ := filepath.Glob(pattern)

	if len(files) == 0 {
		return "", fmt.Errorf("[ERROR] vmx file was not found: %s", pattern)
	}

	vmxFile := files[0]
	log.Printf("[DEBUG] Gold virtual machine vmx file found %v", vmxFile)

	// Gets VIX instance
	client, err := v.client()
	if err != nil {
		return "", err
	}
	defer client.Disconnect()

	log.Printf("[INFO] Opening Gold virtual machine from %s", vmxFile)

	vm, err := client.OpenVm(vmxFile, v.Image.Password)
	if err != nil {
		return "", err
	}

	newvmx := filepath.Join(usr.HomeDir, ".terraform/vix/vms",
		image.Checksum, v.Name, v.Name+".vmx")

	if _, err = os.Stat(newvmx); err != os.ErrExist {
		log.Printf("[INFO] Cloning gold vm into %s...", newvmx)
		clonedVM, err := vm.Clone(vix.CLONETYPE_LINKED, newvmx)
		if err != nil {
			return "", err
		}

		// Makes sure the first time the VM is created it has no virtual network adapters
		log.Printf("[DEBUG] Removing all virtual network adapters from cloned VM...")
		err = clonedVM.RemoveAllNetworkAdapters()
		if err != nil {
			return "", err
		}
	} else {
		log.Printf("[INFO] VM Clone %s already exist, moving on.", newvmx)
	}

	if err = v.Update(newvmx); err != nil {
		return "", err
	}

	return newvmx, nil
}

// Opens and updates virtual machine resource
func (v *VM) Update(vmxFile string) error {
	// Sets default values if some attributes were not set or have
	// invalid values
	v.SetDefaults()

	// Gets VIX instance
	client, err := v.client()
	if err != nil {
		return err
	}
	defer client.Disconnect()

	if client.Provider == vix.VMWARE_VI_SERVER ||
		client.Provider == vix.VMWARE_SERVER {
		log.Printf("[INFO] Registering VM in host's inventory...")
		err = client.RegisterVm(vmxFile)
		if err != nil {
			return err
		}
	}

	log.Printf("[INFO] Opening virtual machine from %s", vmxFile)

	vm, err := client.OpenVm(vmxFile, v.Image.Password)
	if err != nil {
		return err
	}

	running, err := vm.IsRunning()
	if err != nil {
		return err
	}

	if running {
		log.Printf("[INFO] Virtual machine seems to be running, we need to " +
			"power it off in order to make changes.")
		err = v.powerOff(vm)
		if err != nil {
			return err
		}
	}

	memoryInMb, err := humanize.ParseBytes(v.Memory)
	if err != nil {
		log.Printf("[WARN] Unable to set memory size, defaulting to 512mib: %s", err)
		memoryInMb = 512
	} else {
		memoryInMb = (memoryInMb / 1024) / 1024
	}

	log.Printf("[DEBUG] Setting memory size to %d megabytes", memoryInMb)
	vm.SetMemorySize(uint(memoryInMb))

	log.Printf("[DEBUG] Setting vcpus to %d", v.CPUs)
	vm.SetNumberVcpus(uint8(v.CPUs))

	log.Printf("[DEBUG] Setting name to %s", v.Name)
	vm.SetDisplayName(v.Name)

	log.Printf("[DEBUG] Setting description to %s", v.Description)
	vm.SetAnnotation(v.Description)

	if v.UpgradeVHardware &&
		client.Provider != vix.VMWARE_PLAYER {

		log.Println("[INFO] Upgrading virtual hardware...")
		err = vm.UpgradeVHardware()
		if err != nil {
			return err
		}
	}

	// log.Printf("[DEBUG] Loading all network adapters in memory...")
	// curAdapters, err := vm.NetworkAdapters()
	// if err != nil {
	// 	return err
	// }

	log.Printf("[DEBUG] Removing all network adapters from vmx file...")
	err = vm.RemoveAllNetworkAdapters()
	if err != nil {
		return err
	}

	log.Println("[DEBUG] Attaching virtual network adapters...")
	for _, adapter := range v.VNetworkAdapters {
		adapter.StartConnected = true
		if adapter.ConnType == vix.NETWORK_BRIDGED {
			adapter.LinkStatePropagation = true
		}

		err := vm.AddNetworkAdapter(adapter)
		if err != nil {
			return err
		}
	}

	log.Println("[INFO] Powering virtual machine on...")
	var options vix.VMPowerOption

	if v.LaunchGUI {
		log.Println("[INFO] Preparing to launch GUI...")
		options |= vix.VMPOWEROP_LAUNCH_GUI
	}

	options |= vix.VMPOWEROP_NORMAL

	err = vm.PowerOn(options)
	if err != nil {
		return err
	}

	log.Println("[INFO] Waiting for VMware Tools to initialize...")
	err = vm.WaitForToolsInGuest(v.ToolsInitTimeout)
	if err != nil {
		log.Println("[WARN] VMware Tools took too long to initialize or is not " +
			"installed.")

		if v.SharedFolders {
			log.Println("[WARN] Enabling shared folders is not possible.")
		}
		return nil
	}

	if v.SharedFolders {
		log.Println("[DEBUG] Enabling shared folders...")

		err = vm.EnableSharedFolders(v.SharedFolders)
		if err != nil {
			return err
		}
	}
	return nil
}

// Powers off a virtual machine attempting a graceful shutdown.
func (v *VM) powerOff(vm *vix.VM) error {
	tstate, err := vm.ToolsState()
	if err != nil {
		return err
	}

	var powerOpts vix.VMPowerOption
	log.Printf("Tools state %d", tstate)

	if (tstate & vix.TOOLSSTATE_RUNNING) != 0 {
		log.Printf("[INFO] VMware Tools is running, attempting a graceful shutdown...")
		// if VMware Tools is running, attempt a graceful shutdown.
		powerOpts |= vix.VMPOWEROP_FROM_GUEST
	} else {
		log.Printf("[INFO] VMware Tools is NOT running, shutting down the " +
			"machine abruptly...")
		powerOpts |= vix.VMPOWEROP_NORMAL
	}

	err = vm.PowerOff(powerOpts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Virtual machine is off.")

	return nil
}

// Destroys a virtual machine resource
func (v *VM) Destroy(vmxFile string) error {
	log.Printf("[DEBUG] Destroying VM resource %s...", vmxFile)

	client, err := v.client()
	if err != nil {
		return err
	}
	defer client.Disconnect()

	vm, err := client.OpenVm(vmxFile, v.Image.Password)
	if err != nil {
		return err
	}

	running, err := vm.IsRunning()
	if err != nil {
		return err
	}

	if running {
		if err = v.powerOff(vm); err != nil {
			return err
		}
	}

	if client.Provider == vix.VMWARE_VI_SERVER ||
		client.Provider == vix.VMWARE_SERVER {
		log.Printf("[INFO] Unregistering VM from host's inventory...")

		err := client.UnregisterVm(vmxFile)
		if err != nil {
			return err
		}
	}

	return vm.Delete(vix.VMDELETE_KEEP_FILES | vix.VMDELETE_FORCE)
}

// Refreshes state with VMware
func (v *VM) Refresh(vmxFile string) (bool, error) {
	log.Printf("[DEBUG] Syncing VM resource %s...", vmxFile)

	client, err := v.client()
	if err != nil {
		return false, err
	}
	defer client.Disconnect()

	vm, err := client.OpenVm(vmxFile, v.Image.Password)
	if err != nil {
		return false, err
	}

	running, err := vm.IsRunning()
	if !running {
		return running, err
	}

	vcpus, err := vm.Vcpus()
	if err != nil {
		return running, err
	}

	memory, err := vm.MemorySize()
	if err != nil {
		return running, err
	}

	// We need to convert memory value to megabytes so humanize can interpret it
	// properly.
	memory = (memory * 1024) * 1024
	v.Memory = strings.ToLower(humanize.IBytes(uint64(memory)))
	v.CPUs = uint(vcpus)
	v.Name, err = vm.DisplayName()
	v.Description, err = vm.Annotation()

	return running, err
}
