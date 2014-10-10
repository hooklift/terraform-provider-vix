// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vix

/*
#include "vix.h"
#include "helper.h"
*/
import "C"

import "unsafe"

type Host struct {
	Provider Provider
	handle   C.VixHandle
}

// Destroys the state for a particular host instance
//
// Call this function to disconnect the host. After you call this function the
// Host object is no longer valid and you should not longer use it.
// Similarly, you should not use any other object instances obtained from the
// Host object while it was connected.
//
// Since VMware Server 1.0
func (h *Host) Disconnect() {
	if h.handle != C.VIX_INVALID_HANDLE {
		C.VixHost_Disconnect(h.handle)
		h.handle = C.VIX_INVALID_HANDLE
	}
}

//export go_callback_char
func go_callback_char(callbackPtr unsafe.Pointer, item *C.char) {
	callback := *(*func(*C.char))(callbackPtr)
	callback(item)
}

// This function finds Vix objects. For example, when used to find all
// running virtual machines, Host.FindItems() returns a series of virtual
// machine file path names.
func (h *Host) FindItems(options SearchType) ([]string, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var items []string

	callback := func(item *C.char) {
		items = append(items, C.GoString(item))
	}

	jobHandle = C.VixHost_FindItems(h.handle,
		C.VixFindItemType(options), //searchType
		C.VIX_INVALID_HANDLE,       //searchCriteria
		-1,                         //timeout
		(*C.VixEventProc)(C.find_items_callback), //callbackProc
		unsafe.Pointer(&callback))                //clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "host.FindItems",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return items, nil
}

// This function opens a virtual machine on the host
// and returns a VM instance.
//
// Parameters:
//
//   VmxFile: The path name of the virtual machine configuration file on the
//   local host.
//
//   Password: If VM is encrypted, this is the password for VIX to be able to
//   open it.
//
// Remarks:
//
//   * This function opens a virtual machine on the host instance
//     The virtual machine is identified by vmxFile, which is a path name to the
//     configuration file (.VMX file) for that virtual machine.
//
//   * The format of the path name depends on the host operating system.
//     For example, a path name for a Windows host requires backslash as a
//     directory separator, whereas a Linux host requires a forward slash. If the
//     path name includes backslash characters, you need to precede each one with
//     an escape character. For VMware Server 2.x, the path contains a preceeding
//     data store, for example [storage1] vm/vm.vmx.
//
//   * For VMware Server hosts, a virtual machine must be registered before you
//     can open it. You can register a virtual machine by opening it with the
//     VMware Server Console, through the vmware-cmd command with the register
//     parameter, or with Host.RegisterVM().
//
//   * For vSphere, the virtual machine opened may not be the one desired if more
//     than one Datacenter contains VmxFile.
//
//   * To open an encrypted virtual machine, pass its correspondent password.
//
// Since VMware Workstation 7.0
func (h *Host) OpenVm(vmxFile, password string) (*VM, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var propertyHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var vmHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	defer C.Vix_ReleaseHandle(propertyHandle)
	defer C.Vix_ReleaseHandle(jobHandle)

	if password != "" {
		cpassword := C.CString(password)
		defer C.free(unsafe.Pointer(cpassword))

		err = C.alloc_vm_pwd_proplist(h.handle,
			&propertyHandle,
			cpassword)

		if C.VIX_OK != err {
			return nil, &VixError{
				Operation: "host.OpenVM",
				Code:      int(err & 0xFFFF),
				Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
			}
		}
	}

	cVmxFile := C.CString(vmxFile)
	defer C.free(unsafe.Pointer(cVmxFile))

	jobHandle = C.VixHost_OpenVM(h.handle,
		cVmxFile,
		C.VIX_VMOPEN_NORMAL,
		propertyHandle,
		nil, // callbackProc
		nil) // clientData

	err = C.get_vix_handle(jobHandle,
		C.VIX_PROPERTY_JOB_RESULT_HANDLE,
		&vmHandle,
		C.VIX_PROPERTY_NONE)

	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "host.OpenVM.get_vix_handle",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return NewVirtualMachine(vmHandle, vmxFile)
}

// This function adds a virtual machine to the host's inventory.
//
// Parameters:
//
//   vmxFile: The path name of the .vmx file on the host.
//
// Remarks:
//
//   * This function registers the virtual machine identified by vmxFile, which
//     is a storage path to the configuration file (.vmx) for that virtual machine.
//     You can register a virtual machine regardless of its power state.
//
//   * The format of the path name depends on the host operating system.
//     If the path name includes backslash characters, you need to precede each
//     one with an escape character. Path to storage [standard] or [storage1] may
//     vary.
//
//   * For VMware Server 1.x, supply the full path name instead of storage path,
//     and specify provider VMWARE_SERVER to connect.
//
//   * This function has no effect on Workstation or Player, which lack a virtual
//     machine inventory.
//
//   * It is not a Vix error to register an already-registered virtual machine,
//     although the VMware Server UI shows an error icon in the Task pane.
//     Trying to register a non-existent virtual machine results in error 2000,
//     VIX_E_NOT_FOUND.
//
// Since VMware Server 1.0
func (h *Host) RegisterVm(vmxFile string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	cVmxFile := C.CString(vmxFile)
	defer C.free(unsafe.Pointer(cVmxFile))

	jobHandle = C.VixHost_RegisterVM(h.handle,
		cVmxFile,
		nil, // callbackProc
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "host.RegisterVM",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function removes a virtual machine from the host's inventory.
//
// Parameters:
//
//   VmxFile: The path name of the .vmx file on the host.
//
// Remarks:
//
//   * This function unregisters the virtual machine identified by vmxFile,
//     which is a storage path to the configuration file (.vmx) for that virtual
//     machine. A virtual machine must be powered off to unregister it.
//   * The format of the storage path depends on the host operating system.
//     If the storage path includes backslash characters, you need to precede each
//     one with an escape character. Path to storage [standard] or [storage1] may
//     vary.
//   * For VMware Server 1.x, supply the full path name instead of storage path,
//     and specify VMWARE_SERVER provider to connect.
//   * This function has no effect on Workstation or Player, which lack a virtual
//     machine inventory.
//   * It is not a Vix error to unregister an already-unregistered virtual machine,
//     nor is it a Vix error to unregister a non-existent virtual machine.
//
// Since VMware Server 1.0
func (h *Host) UnregisterVm(vmxFile string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	cVmxFile := C.CString(vmxFile)
	defer C.free(unsafe.Pointer(cVmxFile))

	jobHandle = C.VixHost_UnregisterVM(h.handle,
		cVmxFile,
		nil, // callbackProc
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "host.UnregisterVM",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// Copies a file or directory from the local system (where the Vix client is
// running) to the guest operating system.
//
// Parameters:
//
//   src: The path name of a file on a file system available to the Vix client.
//   guest: Guest instance where the file is going to be copied to
//   dest: The path name of a file on a file system available to the guest.
//
// Remarks:
//
//   * The virtual machine must be running while the file is copied from the Vix
//     client machine to the guest operating system.
//
//   * Existing files of the same name are overwritten, and folder contents are
//     merged.
//
//   * The copy operation requires VMware Tools to be installed and running in
//     the guest operating system.
//
//   * You must call VM.LoginInGuest() before calling this function in order
//     to get a Guest instance.
//
//   * The format of the file name depends on the guest or local operating system.
//     For example, a path name for a Microsoft Windows guest or host requires
//     backslash as a directory separator, whereas a Linux guest or host requires
//     a forward slash. If the path name includes backslash characters,
//     you need to precede each one with an escape character.
//
//   * Only absolute paths should be used for files in the guest; the resolution
//     of relative paths is not specified.
//
//   * If any file fails to be copied, Vix aborts the operation, does not attempt
//     to copy the remaining files, and returns an error.
//
//   * In order to copy a file to a mapped network drive in a Windows guest
//     operating system, it is necessary to call VixVM_LoginInGuest() with the
//     LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT flag set.
//     Using the interactive session option incurs an overhead in file transfer
//     speed.
//
// Since VMware Server 1.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
//
func (h *Host) CopyFileToGuest(src string, guest *Guest, dest string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	csrc := C.CString(src)
	cdest := C.CString(dest)
	defer C.free(unsafe.Pointer(csrc))
	defer C.free(unsafe.Pointer(cdest))

	jobHandle = C.VixVM_CopyFileFromHostToGuest(
		guest.handle,         // VM handle
		csrc,                 // src name
		cdest,                // dest name
		C.int(0),             // options
		C.VIX_INVALID_HANDLE, // propertyListHandle
		nil,                  // callbackProc
		nil)                  // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "host.CopyFileToGuest",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}
