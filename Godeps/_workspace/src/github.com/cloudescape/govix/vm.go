// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package vix

/*
#include "vix.h"
#include "helper.h"
*/
import "C"

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"time"
	"unsafe"

	"github.com/hooklift/govmx"
)

type VM struct {
	// Internal VIX handle
	handle  C.VixHandle
	vmxfile *VMXFile
}

func NewVirtualMachine(handle C.VixHandle, vmxpath string) (*VM, error) {
	vmxfile := &VMXFile{
		path: vmxpath,
	}

	// Loads VMX file in memory
	err := vmxfile.Read()
	if err != nil {
		return nil, err
	}

	vm := &VM{
		handle:  handle,
		vmxfile: vmxfile,
	}

	runtime.SetFinalizer(vm, cleanupVM)
	return vm, nil
}

// Returns number of virtual CPUs configured for
// the virtual machine.
func (v *VM) Vcpus() (uint8, error) {
	var err C.VixError = C.VIX_OK
	var vcpus C.int = C.VIX_PROPERTY_NONE

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_NUM_VCPUS,
		unsafe.Pointer(&vcpus))

	if C.VIX_OK != err {
		return 0, &VixError{
			Operation: "vm.Vcpus",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return uint8(vcpus), nil
}

// Returns path to the virtual machine configuration file.
func (v *VM) VmxPath() (string, error) {
	var err C.VixError = C.VIX_OK
	var path *C.char

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_VMX_PATHNAME,
		unsafe.Pointer(&path))

	defer C.Vix_FreeBuffer(unsafe.Pointer(path))

	if C.VIX_OK != err {
		return "", &VixError{
			Operation: "vm.VmxPath",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return C.GoString(path), nil
}

// Returns path to the virtual machine team.
func (v *VM) VmTeamPath() (string, error) {
	var err C.VixError = C.VIX_OK
	var path *C.char

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_VMTEAM_PATHNAME,
		unsafe.Pointer(&path))

	defer C.Vix_FreeBuffer(unsafe.Pointer(path))

	if C.VIX_OK != err {
		return "", &VixError{
			Operation: "vm.VmTeamPath",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return C.GoString(path), nil
}

// Returns memory size of the virtual machine.
func (v *VM) MemorySize() (uint, error) {
	var err C.VixError = C.VIX_OK
	var memsize C.uint = C.VIX_PROPERTY_NONE

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_MEMORY_SIZE,
		unsafe.Pointer(&memsize))

	if C.VIX_OK != err {
		return 0, &VixError{
			Operation: "vm.MemorySize",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return uint(memsize), nil
}

func (v *VM) ReadOnly() (bool, error) {
	var err C.VixError = C.VIX_OK
	var readonly C.Bool = C.VIX_PROPERTY_NONE

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_READ_ONLY,
		unsafe.Pointer(&readonly))

	if C.VIX_OK != err {
		return false, &VixError{
			Operation: "vm.ReadOnly",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	if readonly == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// Returns whether the virtual machine is a member of a team
func (v *VM) InVmTeam() (bool, error) {
	var err C.VixError = C.VIX_OK
	var inTeam C.Bool = C.VIX_PROPERTY_NONE

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_IN_VMTEAM,
		unsafe.Pointer(&inTeam))

	if C.VIX_OK != err {
		return false, &VixError{
			Operation: "vm.InVmTeam",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	if inTeam == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// Returns power state of the virtual machine.
func (v *VM) PowerState() (VMPowerState, error) {
	var err C.VixError = C.VIX_OK
	var state C.VixPowerState = 0x0

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_POWER_STATE,
		unsafe.Pointer(&state))

	if C.VIX_OK != err {
		return VMPowerState(0x0), &VixError{
			Operation: "vm.PowerState",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return VMPowerState(state), nil
}

// Returns state of the VMware Tools suite in the guest.
func (v *VM) ToolsState() (GuestToolsState, error) {
	var err C.VixError = C.VIX_OK
	var state C.VixToolsState = C.VIX_TOOLSSTATE_UNKNOWN

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_TOOLS_STATE,
		unsafe.Pointer(&state))

	if C.VIX_OK != err {
		return TOOLSSTATE_UNKNOWN, &VixError{
			Operation: "vm.ToolsState",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return GuestToolsState(state), nil
}

// Returns whether the virtual machine is running.
func (v *VM) IsRunning() (bool, error) {
	var err C.VixError = C.VIX_OK
	var running C.Bool = C.VIX_PROPERTY_NONE

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_IS_RUNNING,
		unsafe.Pointer(&running))

	if C.VIX_OK != err {
		return false, &VixError{
			Operation: "vm.IsRunning",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	if running == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// Returns the guest os
func (v *VM) GuestOS() (string, error) {
	var err C.VixError = C.VIX_OK
	var os *C.char

	err = C.get_property(v.handle,
		C.VIX_PROPERTY_VM_GUESTOS,
		unsafe.Pointer(&os))

	defer C.Vix_FreeBuffer(unsafe.Pointer(os))

	if C.VIX_OK != err {
		return "", &VixError{
			Operation: "vm.GuestOS",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return C.GoString(os), nil
}

// Returns VM supported features
// func (v *VM) Features() (string, error) {
// 	var err C.VixError = C.VIX_OK
// 	var features *C.char

// 	err = C.get_property(v.handle,
// 		C.VIX_PROPERTY_VM_SUPPORTED_FEATURES,
// 		unsafe.Pointer(&features))

// 	defer C.Vix_FreeBuffer(unsafe.Pointer(features))

// 	if C.VIX_OK != err {
// 		return "", &VixError{
// 			code: int(err & 0xFFFF),
// 			text: C.GoString(C.Vix_GetErrorText(err, nil)),
// 		}
// 	}

// 	return C.GoString(features), nil
// }

// This function enables or disables all shared folders as a feature for a
// virtual machine.
//
// Remarks:
//
//   * This function enables/disables all shared folders as a feature on a
//     virtual machine. In order to access shared folders on a guest, the
//     feature has to be enabled, and in addition, the individual shared folder
//     has to be enabled.
//
//   * It is not necessary to call VM.LoginInGuest() before calling this function.
//
//   * In this release, this function requires the virtual machine to be powered
//     on with VMware Tools installed.
//
//   * Shared folders are not supported for the following guest operating systems:
//     Windows ME, Windows 98, Windows 95, Windows 3.x, and DOS.
//
//   * On Linux virtual machines, calling this function will automatically mount
//     shared folder(s) in the guest.
//
// Since VMware Workstation 6.0, not available on Server 2.0.
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
//
func (v *VM) EnableSharedFolders(enabled bool) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	var share C.Bool = C.FALSE
	if enabled {
		share = C.TRUE
	} else {
		share = C.FALSE
	}

	jobHandle = C.VixVM_EnableSharedFolders(v.handle,
		share,
		0,
		nil,
		nil)

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.EnableSharedFolders",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function mounts a new shared folder in the virtual machine.
//
// Parameters:
//
//   guestpath: Specifies the guest path name of the new shared folder.
//   hostpath: Specifies the host path of the shared folder.
//   flags: The folder options.
//
// Remarks:
//
//   * This function creates a local mount point in the guest file system and
//     mounts a shared folder exported by the host.
//
//   * Shared folders will only be accessible inside the guest operating system
//     if shared folders are enabled for the virtual machine.
//     See the documentation for VM.EnableSharedFolders().
//
//   * The folder options include: SHAREDFOLDER_WRITE_ACCESS - Allow write access.
//
//   * Only absolute paths should be used for files in the guest; the resolution
//     of relative paths is not specified.
//
//   * The hostpath argument must specify a path to a directory that exists on the
//     host, or an error will result.
//
//   * If a shared folder with the same name exists before calling this function,
//     the job handle returned by this function will return VIX_E_ALREADY_EXISTS.
//
//   * It is not necessary to call VM.LoginInGuest() before calling this function.
//
//   * When creating shared folders in a Windows guest, there might be a delay
//     before contents of a shared folder are visible to functions such as
//     Guest.IsFile() and Guest.RunProgram().
//
//   * Shared folders are not supported for the following guest operating
//     systems: Windows ME, Windows 98, Windows 95, Windows 3.x, and DOS.
//
//   * In this release, this function requires the virtual machine to be powered
//     on with VMware Tools installed.
//
//   * To determine in which directory in the guest the shared folder will be,
//     query Guest.SharedFoldersParentDir(). When the virtual machine is powered
//     on and the VMware Tools are running, this property will contain the path to
//     the parent directory of the shared folders for that virtual machine.
//
// Since VMware Workstation 6.0, not available on Server 2.0.
//
func (v *VM) AddSharedFolder(guestpath, hostpath string, flags SharedFolderOption) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	gpath := C.CString(guestpath)
	hpath := C.CString(hostpath)
	defer C.free(unsafe.Pointer(gpath))
	defer C.free(unsafe.Pointer(hpath))

	jobHandle = C.VixVM_AddSharedFolder(v.handle,
		gpath,
		hpath,
		C.VixMsgSharedFolderOptions(flags),
		nil, nil)

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.AddSharedFolder",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function removes a shared folder in the virtual machine.
//
// Parameters:
//
//   guestpath: Specifies the guest pathname of the shared folder to delete.
//
// Remarks:
//
//   * This function removes a shared folder in the virtual machine referenced by
//     the VM object
//
//   * It is not necessary to call VM.LoginInGuest() before calling this function.
//
//   * Shared folders are not supported for the following guest operating
//     systems: Windows ME, Windows 98, Windows 95, Windows 3.x, and DOS.
//
//   * In this release, this function requires the virtual machine to be powered
//     on with VMware Tools installed.
//
//   * Depending on the behavior of the guest operating system, when removing
//     shared folders, there might be a delay before the shared folder is no
//     longer visible to programs running within the guest operating system and
//     to functions such as Guest.IsFile()
//
// Since VMware Workstation 6.0
//
func (v *VM) RemoveSharedFolder(guestpath string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	gpath := C.CString(guestpath)
	defer C.free(unsafe.Pointer(gpath))

	jobHandle = C.VixVM_RemoveSharedFolder(v.handle,
		gpath,
		0,
		nil, nil)

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.RemoveSharedFolder",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function captures the screen of the guest operating system.
//
// Remarks:
//
//   * This function captures the current screen image and returns it as a
//     []byte result.
//
//   * For security reasons, this function requires a successful call to
//     VM.LoginInGuest() must be made.
//
// Since VMware Workstation 6.5
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
//
func (v *VM) Screenshot() ([]byte, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var byte_count C.int
	var screen_bits C.char

	jobHandle = C.VixVM_CaptureScreenImage(v.handle,
		C.VIX_CAPTURESCREENFORMAT_PNG,
		C.VIX_INVALID_HANDLE,
		nil,
		nil)
	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.get_screenshot_bytes(jobHandle, &byte_count, &screen_bits)
	defer C.Vix_FreeBuffer(unsafe.Pointer(&screen_bits))

	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "vm.Screenshot",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return C.GoBytes(unsafe.Pointer(&screen_bits), byte_count), nil
}

// Creates a copy of the virtual machine specified by the current VM instance
//
// Parameters:
//
//   cloneType: Must be either CLONETYPE_FULL or CLONETYPE_LINKED.
//     * CLONETYPE_FULL: Creates a full, independent clone of the virtual machine.
//     * CLONETYPE_LINKED: Creates a linked clone, which is a copy of a virtual
//       machine that shares virtual disks with the parent
//       virtual machine in an ongoing manner.
//       This conserves disk space as long as the parent and
//       clone do not change too much from their original state.
//
//   destVmxFile: The path name of the virtual machine configuration file that will
//     be created for the virtual machine clone produced by this operation.
//     This should be a full absolute path name, with directory names delineated
//     according to host system convention: \ for Windows and / for Linux.
//
// Remarks:
//
//   * The function returns a new VM instance which is a clone of its parent VM.
//
//   * It is not possible to create a full clone of a powered on virtual machine.
//     You must power off or suspend a virtual machine before creating a full
//     clone of that machine.
//
//   * With a suspended virtual machine, requesting a linked clone results in
//     error 3007 VIX_E_VM_IS_RUNNING.
//     Suspended virtual machines retain memory state, so proceeding with a
//     linked clone could cause loss of data.
//
//   * A linked clone must have access to the parent's virtual disks. Without
//     such access, you cannot use a linked clone
//     at all because its file system will likely be incomplete or corrupt.
//
//   * Deleting a virtual machine that is the parent of a linked clone renders
//     the linked clone useless.
//
//   * Because a full clone does not share virtual disks with the parent virtual
//     machine, full clones generally perform better than linked clones.
//     However, full clones take longer to create than linked clones. Creating a
//     full clone can take several minutes if the files involved are large.
//
//   * This function is not supported when using the VMWARE_PLAYER provider.
//
func (v *VM) Clone(cloneType CloneType, destVmxFile string) (*VM, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var clonedHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	dstVmxFile := C.CString(destVmxFile)
	defer C.free(unsafe.Pointer(dstVmxFile))

	jobHandle = C.VixVM_Clone(v.handle,
		C.VIX_INVALID_HANDLE,      // snapshotHandle
		C.VixCloneType(cloneType), // cloneType
		dstVmxFile,                // destConfigPathName
		0,                         // options,
		C.VIX_INVALID_HANDLE,      // propertyListHandle
		nil,                       // callbackProc
		nil)                       // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.get_vix_handle(jobHandle,
		C.VIX_PROPERTY_JOB_RESULT_HANDLE,
		&clonedHandle,
		C.VIX_PROPERTY_NONE)

	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "vm.Clone",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return NewVirtualMachine(clonedHandle, destVmxFile)
}

// Private function to clean up vm handle
func cleanupVM(v *VM) {
	if v.handle != C.VIX_INVALID_HANDLE {
		C.Vix_ReleaseHandle(v.handle)
		v.handle = C.VIX_INVALID_HANDLE
	}
}

// This function saves a copy of the virtual machine state as a snapshot object.
//
// Parameters:
//
//   name: A user-defined name for the snapshot; need not be unique.
//
//   description: A user-defined description for the snapshot.
//
//   options: Flags to specify how the snapshot should be created. Any combination of the
//   following or 0 to exclude memory:
//     * SNAPSHOT_INCLUDE_MEMORY: Captures the full state of a running virtual
//       machine, including the memory.
//
// Remarks:
//
//   * This function creates a child snapshot of the current snapshot.
//
//   * If a virtual machine is suspended, you cannot snapshot it more than once.
//
//   * If a powered-on virtual machine gets a snapshot created with option 0
//     (exclude memory), the power state is not saved, so reverting to the
//     snapshot sets powered-off state.
//
//   * The 'name' and 'description' parameters can be set but not retrieved
//     using the VIX API.
//
//   * VMware Server supports only a single snapshot for each virtual machine.
//     The following considerations apply to VMware Server:
//      * If you call this function a second time for the same virtual machine
//        without first deleting the snapshot,
//        the second call will overwrite the previous snapshot.
//      * A virtual machine imported to VMware Server from another VMware product
//        might have more than one snapshot at the time it is imported. In that
//        case, you can use this function to add a new snapshot to the series.
//
//   * Starting in VMware Workstation 6.5, snapshot operations are allowed on
//     virtual machines that are part of a team.
//     Previously, this operation failed with error code
//     VIX_PROPERTY_VM_IN_VMTEAM. Team members snapshot independently so they can
//     have different and inconsistent snapshot states.
//
//   * This function is not supported when using the VMWARE_PLAYER provider.
//
//   * If the virtual machine is open and powered off in the UI, this function now
//     closes the virtual machine in the UI before creating the snapshot.
//
// Since VMware Workstation 6.0
//
func (v *VM) CreateSnapshot(name, description string, options CreateSnapshotOption) (*Snapshot, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var snapshotHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	sname := C.CString(name)
	sdesc := C.CString(description)
	defer C.free(unsafe.Pointer(sname))
	defer C.free(unsafe.Pointer(sdesc))

	jobHandle = C.VixVM_CreateSnapshot(v.handle,
		sname, // name
		sdesc, // description
		C.VixCreateSnapshotOptions(options), // options
		C.VIX_INVALID_HANDLE,                // propertyListHandle
		nil,                                 // callbackProc
		nil)                                 // clientData

	err = C.get_vix_handle(jobHandle,
		C.VIX_PROPERTY_JOB_RESULT_HANDLE,
		&snapshotHandle,
		C.VIX_PROPERTY_NONE)

	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "vm.CreateSnapshot",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	snapshot := &Snapshot{
		handle: snapshotHandle,
	}

	runtime.SetFinalizer(snapshot, cleanupSnapshot)

	return snapshot, nil
}

// This function deletes all saved states for the snapshot.
//
// Parameters:
//
//   snapshot: A Snapshot instance. Call VM.RootSnapshot() to get a snapshot instance.
//
// Remarks:
//
//   * This function deletes all saved states for the specified snapshot. If the
//     snapshot was based on another snapshot, the base snapshot becomes the new
//     root snapshot.
//
//   * The VMware Server release of the VIX API can manage only a single snapshot
//     for each virtual machine. A virtual machine imported from another VMware
//     product can have more than one snapshot at the time it is imported. In that
//     case, you can delete only a snapshot subsequently added using the VIX API.
//
//   * Starting in VMware Workstation 6.5, snapshot operations are allowed on
//     virtual machines that are part of a team. Previously, this operation
//     failed with error code VIX_PROPERTY_VM_IN_VMTEAM. Team members snapshot
//     independently so they can have different and inconsistent snapshot states.
//
//   * This function is not supported when using the VMWARE_PLAYER provider
//
//   * If the virtual machine is open and powered off in the UI, this function may
//     close the virtual machine in the UI before deleting the snapshot.
//
// Since VMware Server 1.0
//
func (v *VM) RemoveSnapshot(snapshot *Snapshot, options RemoveSnapshotOption) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_RemoveSnapshot(v.handle, //vmHandle
		snapshot.handle,                     //snapshotHandle
		C.VixRemoveSnapshotOptions(options), //options
		nil, //callbackProc
		nil) //clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.RemoveSnapshot",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function permanently deletes a virtual machine from your host system.
//
// Parameters:
//
//   options: For VMware Server 2.0 and ESX, this value must be VMDELETE_DISK_FILES.
//   For all other versions it can be either VMDELETE_KEEP_FILES or VMDELETE_DISK_FILES.
//   When option is VMDELETE_DISK_FILES, it deletes all associated files.
//   When option is VMDELETE_KEEP_FILES, does not delete *.vmdk virtual disk file(s).
//
// Remarks:
//
//   * This function permanently deletes a virtual machine from your host system.
//     You can accomplish the same effect by deleting all virtual machine files
//     using the host file system. This function simplifies the task by deleting
//     all VMware files in a single function call.
//     If a deleteOptions value of VMDELETE_KEEP_FILES is specified, the virtual disk (vmdk) files
//     will not be deleted.
//     This function does not delete other user files in the virtual machine
//     folder.
//
//   * This function is successful only if the virtual machine is powered off or
//     suspended.
//
//   * Deleting a virtual machine that is the parent of a linked clone renders
//     the linked clone useless.
//
//   * If the machine was powered on with GUI enabled, this function is going to
//     return error VIX_E_FILE_ALREADY_LOCKED if the VM tab is open. If you want
//     to force the removal despite this, pass the option VMDELETE_FORCE.
//
// since VMware Server 1.0
//
func (v *VM) Delete(options VmDeleteOption) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	if (options & VMDELETE_FORCE) != 0 {
		vmxPath, err := v.VmxPath()
		if err != nil {
			return err
		}

		err = os.RemoveAll(vmxPath + ".lck")
		if err != nil {
			return err
		}
	}

	jobHandle = C.VixVM_Delete(v.handle, C.VixVMDeleteOptions(options), nil, nil)

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.Delete",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	C.Vix_ReleaseHandle(v.handle)
	return nil
}

// This function returns the handle of the current active snapshot belonging to
// the virtual machine
//
// Remarks:
//
//   * This function returns the handle of the current active snapshot belonging
//     to the virtual machine.
//
//   * This function is not supported when using the VMWARE_PLAYER provider
//
// Since VMware Workstation 6.0
//
func (v *VM) CurrentSnapshot() (*Snapshot, error) {
	var snapshotHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	err = C.VixVM_GetCurrentSnapshot(v.handle, &snapshotHandle)
	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "vm.CurrentSnapshot",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	snapshot := &Snapshot{handle: snapshotHandle}

	runtime.SetFinalizer(snapshot, cleanupSnapshot)

	return snapshot, nil
}

// This function returns a Snapshot object matching the given name
//
// Parameters:
//
//   name: Identifies a snapshot name.
//
// Remarks:
//
//   * When the snapshot name is a duplicate, it returns error 13017
//     VIX_E_SNAPSHOT_NONUNIQUE_NAME.
//
//   * When there are multiple snapshots with the same name, or the same path to
//     that name, you cannot specify a unique name, but you can to use the UI to
//     rename duplicates.
//
//   * You can specify the snapshot name as a path using '/' or '\\' as path
//     separators, including snapshots in the tree above the named snapshot,
//     for example 'a/b/c' or 'x/x'. Do not mix '/' and '\\' in the same path
//     expression.
//
//   * This function is not supported when using the VMWARE_PLAYER provider.
//
// Since VMware Workstation 6.0
//
func (v *VM) SnapshotByName(name string) (*Snapshot, error) {
	var snapshotHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	sname := C.CString(name)
	defer C.free(unsafe.Pointer(sname))

	err = C.VixVM_GetNamedSnapshot(v.handle, sname, &snapshotHandle)
	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "vm.SnapshotByName",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	snapshot := &Snapshot{handle: snapshotHandle}

	runtime.SetFinalizer(snapshot, cleanupSnapshot)

	return snapshot, nil
}

// This function returns the number of top-level (root) snapshots belonging to a
// virtual machine.
//
// Remarks:
//
//   * This function returns the number of top-level (root) snapshots belonging to
//     a virtual machine.
//     A top-level snapshot is one that is not based on any previous snapshot.
//     If the virtual machine has more than one snapshot, the snapshots can be a
//     sequence in which each snapshot is based on the previous one, leaving only
//     a single top-level snapshot.
//     However, if applications create branched trees of snapshots, a single
//     virtual machine can have several top-level snapshots.
//
//   * VMware Server can manage only a single snapshot for each virtual machine.
//     All other snapshots in a sequence are ignored. The return value is always
//     0 or 1.
//
//   * This function is not supported when using the VMWARE_PLAYER provider
//
// Since VMware Workstation 6.0
//
func (v *VM) TotalRootSnapshots() (int, error) {
	var result C.int
	var err C.VixError = C.VIX_OK

	err = C.VixVM_GetNumRootSnapshots(v.handle, &result)
	if C.VIX_OK != err {
		return 0, &VixError{
			Operation: "vm.TotalRootSnapshots",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return int(result), nil
}

// This function returns the number of shared folders mounted in the virtual
// machine.
//
// Remarks:
//
//   * This function returns the number of shared folders mounted in the virtual
//     machine.
//
//   * It is not necessary to call VM.LoginInGuest() before calling this function.
//
//   * Shared folders are not supported for the following guest operating systems:
//     Windows ME, Windows 98, Windows 95, Windows 3.x, and DOS.
//
//   * In this release, this function requires the virtual machine to be powered
//     on with VMware Tools installed.
//
// Since VMware Workstation 6.0
//
func (v *VM) TotalSharedFolders() (int, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var numSharedFolders C.int = 0

	jobHandle = C.VixVM_GetNumSharedFolders(v.handle, nil, nil)
	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.get_num_shared_folders(jobHandle, &numSharedFolders)

	if C.VIX_OK != err {
		return 0, &VixError{
			Operation: "vm.TotalSharedFolders",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return int(numSharedFolders), nil
}

// This function returns a root Snapshot instance
// belonging to the current virtual machine
//
// Parameters:
//
//   index: Identifies a root snapshot. See below for range of values.
//
// Remarks:
//
//   * Snapshots are indexed from 0 to n-1, where n is the number of root
//     snapshots. Use the function VM.TotalRootSnapshots to get the value of n.
//
//   * VMware Server can manage only a single snapshot for each virtual machine.
//     The value of index can only be 0.
//
//   * This function is not supported when using the VMWARE_PLAYER provider
//
// Since VMware Server 1.0
//
func (v *VM) RootSnapshot(index int) (*Snapshot, error) {
	var snapshotHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	err = C.VixVM_GetRootSnapshot(v.handle,
		C.int(index),
		&snapshotHandle)

	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "vm.RootSnapshot",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	snapshot := &Snapshot{handle: snapshotHandle}

	runtime.SetFinalizer(snapshot, cleanupSnapshot)

	return snapshot, nil
}

// This function modifies the state of a shared folder mounted in the virtual
// machine.
//
// Parameters:
//
//   name: Specifies the name of the shared folder.
//   hostpath: Specifies the host path of the shared folder.
//   options: The new flag settings.
//
// Remarks:
//
//   * This function modifies the state flags of an existing shared folder.
//
//   * If the shared folder does not exist before calling
//     this function, the function will return a not found error.
//
//   * Shared folders are not supported for the following guest operating
//     systems: Windows ME, Windows 98, Windows 95, Windows 3.x, and DOS.
//
//   * In this release, this function requires the virtual machine to be powered
//     on with VMware Tools installed.
//
// Since VMware Workstation 6.0
//
func (v *VM) SetSharedFolderState(name, hostpath string, options SharedFolderOption) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	sfname := C.CString(name)
	hpath := C.CString(hostpath)
	defer C.free(unsafe.Pointer(sfname))
	defer C.free(unsafe.Pointer(hpath))

	jobHandle = C.VixVM_SetSharedFolderState(v.handle, //vmHandle
		sfname, //shareName
		hpath,  //hostPathName
		C.VixMsgSharedFolderOptions(options), //flags
		nil, //callbackProc
		nil) //clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.SetSharedFolderState",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	C.Vix_ReleaseHandle(v.handle)
	return nil
}

// This function returns the state of a shared folder mounted in the virtual
// machine.
//
// Parameters:
//
//   index: Identifies the shared folder
//
// Remarks:
//
//   * Shared folders are indexed from 0 to n-1, where n is the number of shared
//     folders. Use the function VM.NumSharedFolders() to get the value of n.
//
//   * Shared folders are not supported for the following guest operating systems:
//     Windows ME, Windows 98, Windows 95, Windows 3.x, and DOS.
//
//   * In this release, this function requires the virtual machine to be powered
//     on with VMware Tools installed.
//
// Since VMware Workstation 6.0
//
func (v *VM) SharedFolderState(index int) (string, string, int, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var folderName *C.char
	var folderHostPath *C.char
	var folderFlags *C.int

	jobHandle = C.VixVM_GetSharedFolderState(v.handle, //vmHandle
		C.int(index), //index
		nil,          //callbackProc
		nil)          //clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.get_shared_folder(jobHandle, folderName, folderHostPath, folderFlags)
	defer C.Vix_FreeBuffer(unsafe.Pointer(folderName))
	defer C.Vix_FreeBuffer(unsafe.Pointer(folderHostPath))

	if C.VIX_OK != err {
		return "", "", 0, &VixError{
			Operation: "vm.SharedFolderState",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	C.Vix_ReleaseHandle(v.handle)
	return C.GoString(folderName), C.GoString(folderHostPath), int(*folderFlags),
		nil
}

// This function pauses a virtual machine. See Remarks section for pause
// behavior when used with different operations.
//
// Remarks:
//
//   * This stops execution of the virtual machine.
//
//   * Functions that invoke guest operations should not be called when the
//     virtual machine is paused.
//
//   * Call VM.Resume() to continue execution of the virtual machine.
//
//   * Calling VM.Reset(), VM.PowerOff(), and VM.Suspend() will successfully
//     work when paused. The pause state is not preserved in a suspended virtual
//     machine; a subsequent VM.PowerOn() will not remember the previous pause
//     state.
//
//   * This function is not supported when using the VMWARE_PLAYER provider
//
// Since VMware Workstation 6.5.
//
func (v *VM) Pause() error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	// Commented code is here to have implementation clues
	// about how to return the resulting
	// snapshot object of the pause operation, if needed.
	//
	// var snapshotHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_Pause(v.handle,
		0,                    // options
		C.VIX_INVALID_HANDLE, // propertyListHandle
		nil,                  // callbackProc
		nil)                  // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	// err = C.get_vix_handle(jobHandle,
	// 	C.VIX_PROPERTY_JOB_RESULT_HANDLE,
	// 	&snapshotHandle,
	// 	C.VIX_PROPERTY_NONE)

	// defer C.Vix_ReleaseHandle(snapshotHandle)
	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.Pause",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function continues execution of a paused virtual machine.
//
// Remarks:
//
//   * This operation continues execution of a virtual machine that was stopped
//     using VM.Pause().
//
//   * Refer to VM.Pause() for pause/unpause behavior with different operations.
//
//   * This function is not supported when using the VMWARE_PLAYER provider
//
// Since VMware Workstation 6.5
//
func (v *VM) Resume() error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	//var snapshotHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_Unpause(v.handle,
		0,                    // options
		C.VIX_INVALID_HANDLE, // propertyListHandle
		nil,                  // callbackProc
		nil)                  // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	// err = C.get_vix_handle(jobHandle,
	// 	C.VIX_PROPERTY_JOB_RESULT_HANDLE,
	// 	&snapshotHandle,
	// 	C.VIX_PROPERTY_NONE)

	// defer C.Vix_ReleaseHandle(snapshotHandle)
	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.Resume",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function powers off a virtual machine.
//
// Parameters:
//
//   options: Set of VMPowerOption flags to consider when powering off the
//   virtual machine.
//
// Remarks:
//   * If you call this function while the virtual machine is suspended or powered
//     off, the operation returns a VIX_E_VM_NOT_RUNNING error.
//     If suspended, the virtual machine remains suspended and is not powered off.
//     If powered off, you can safely ignore the error.
//
//   * If you pass VMPOWEROP_NORMAL as an option, the virtual machine is powered
//     off at the hardware level. Any state of the guest that was not committed
//     to disk will be lost.
//
//   * If you pass VMPOWEROP_FROM_GUEST as an option, the function tries to power
//     off the guest OS, ensuring a clean shutdown of the guest. This option
//     requires that VMware Tools be installed and running in the guest.
//
//   * After VMware Tools begin running in the guest, and VM.WaitForToolsInGuest()
//     returns, there is a short delay before VMPOWEROP_FROM_GUEST becomes
//     available.
//     During this time a job may return error 3009,
//     VIX_E_POWEROP_SCRIPTS_NOT_AVAILABLE.
//     As a workaround, add a short sleep after the WaitForTools call.
//
//   * On a Solaris guest with UFS file system on the root partition, the
//     VMPOWEROP_NORMAL parameter causes an error screen at next power on, which
//     requires user intervention to update the Solaris boot archive by logging
//     into the failsafe boot session from the GRUB menu. Hence, although UFS file
//     systems are supported, VMware recommends using the ZFS file system for
//     Solaris guests.
//
// Since VMware Server 1.0
//
func (v *VM) PowerOff(options VMPowerOption) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_PowerOff(v.handle,
		C.VixVMPowerOpOptions(options), // powerOptions,
		nil, // callbackProc,
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.PowerOff",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// Powers on a virtual machine.
//
// Parameters:
//
//   options: VMPOWEROP_NORMAL or VMPOWEROP_LAUNCH_GUI.
//
// Remarks:
//   * This operation completes when the virtual machine has started to boot.
//     If the VMware Tools have been installed on this guest operating system, you
//     can call VM.WaitForToolsInGuest() to determine when the guest has finished
//     booting.
//
//   * After powering on, you must call VM.WaitForToolsInGuest() before executing
//     guest operations or querying guest properties.
//
//   * In Server 1.0, when you power on a virtual machine, the virtual machine is
//     powered on independent of a console window. If a console window is open,
//     it remains open. Otherwise, the virtual machine is powered on without a
//     console window.
//
//   * To display a virtual machine with a Workstation user interface, the options
//     parameter must have the VMPOWEROP_LAUNCH_GUI flag, and you must be
//     connected to the host with the VMWARE_WORKSTATION provider flag. If there
//     is an existing instance of the Workstation user interface, the virtual
//     machine will power on in a new tab within that instance.
//     Otherwise, a new instance of Workstation will open, and the virtual machine
//     will power on there.
//
//   * To display a virtual machine with a Player user interface, the options
//     parameter must have the VMPOWEROP_LAUNCH_GUI flag, and you must be
//     connected to the host with the VMWARE_PLAYER flag. A new instance of Player
//     will always open, and the virtual machine will power on there.
//
//   * This function can also be used to resume execution of a suspended virtual
//     machine.
//
//   * The VMPOWEROP_LAUNCH_GUI option is not supported for encrypted virtual
//     machines; attempting to power on with this option results in
//     VIX_E_NOT_SUPPORTED.
//
// Since VMware Server 1.0
//
func (v *VM) PowerOn(options VMPowerOption) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_PowerOn(v.handle,
		C.VixVMPowerOpOptions(options), // powerOptions,
		C.VIX_INVALID_HANDLE,
		nil, // callbackProc,
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.PowerOn",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function resets a virtual machine.
//
// Parameters:
//
//   options: Must be VMPOWEROP_NORMAL or VMPOWEROP_FROM_GUEST.
//
// Remarks:
//
//   * If the virtual machine is not powered on when you call this function, it
//     returns an error.
//
//   * If you pass VMPOWEROP_NORMAL as an option, this function is the equivalent
//     of pressing the reset button on a physical machine.
//
//   * If you pass VMPOWEROP_FROM_GUEST as an option, this function tries to reset
//     the guest OS, ensuring a clean shutdown of the guest.
//     This option requires that the VMware Tools be installed and running in the
//     guest.
//
//   * After VMware Tools begin running in the guest, and VM.WaitForToolsInGuest()
//     returns, there is a short delay before VMPOWEROP_FROM_GUEST becomes
//     available. During this time the function may return error 3009,
//     VIX_E_POWEROP_SCRIPTS_NOT_AVAILABLE.
//     As a workaround, add a short sleep after the WaitForTools call.
//
//   * After reset, you must call VM.WaitForToolsInGuest() before executing guest
//     operations or querying guest properties.
//
//   * On a Solaris guest with UFS file system on the root partition, the
//     VMPOWEROP_NORMAL parameter causes an error screen at next power on, which
//     requires user intervention to update the Solaris boot archive by logging
//     into the failsafe boot session from the GRUB menu. Hence, although UFS file
//     systems are supported, VMware recommends using the ZFS file system for
//     Solaris guests.
//
// Since VMware Server 1.0
//
func (v *VM) Reset(options VMPowerOption) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_Reset(v.handle,
		C.VixVMPowerOpOptions(options), // powerOptions,
		nil, // callbackProc,
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.Reset",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function suspends a virtual machine.
//
// Remarks:
//
//   * If the virtual machine is not powered on when you call this function, the
//     function returns the error VIX_E_VM_NOT_RUNNING.
//
//   * Call VM.PowerOn() to resume running a suspended virtual machine.
//
// Since VMware Server 1.0
//
func (v *VM) Suspend() error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_Suspend(v.handle,
		0,   // powerOptions,
		nil, // callbackProc,
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.Suspend",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function reads variables from the virtual machine state.
// This includes the virtual machine configuration,
// environment variables in the guest, and VMware "Guest Variables"
//
// Parameters:
//
//   varType: The type of variable to read. The currently supported values are:
//     * VM_GUEST_VARIABLE: A "Guest Variable". This is a runtime-only
//       value; it is never stored persistently. This is the same guest
//       variable that is exposed through the VMControl APIs, and is a simple
//       way to pass runtime values in and out of the guest.
//
//     * VM_CONFIG_RUNTIME_ONLY: The configuration state of the virtual machine.
//       This is the .vmx file that is stored on the host. You can read this and
//       it will return the persistent data. If you write to this, it will only
//       be a runtime change, so changes will be lost when the VM powers off.
//
//     * GUEST_ENVIRONMENT_VARIABLE: An environment variable in the guest of
//       the VM. On a Windows NT series guest, writing these values is saved
//       persistently so they are immediately visible to every process.
//       On a Linux or Windows 9X guest, writing these values is not persistent
//       so they are only visible to the VMware tools process.
//
//   name: The name of the variable.
//
// Remarks:
//
//   * You must call VM.LoginInGuest() before calling this function to read a
//     GUEST_ENVIRONMENT_VARIABLE value.
//
//   * You do not have to call VM.LoginInGuest() to use this function to read a
//     VM_GUEST_VARIABLE or a VVM_CONFIG_RUNTIME_ONLY value.
//
// Since Workstation 6.0
//
func (v *VM) ReadVariable(varType GuestVarType, name string) (string, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var readValue *C.char

	vname := C.CString(name)
	defer C.free(unsafe.Pointer(vname))

	jobHandle = C.VixVM_ReadVariable(v.handle,
		C.int(varType),
		vname,
		0,   // options
		nil, // callbackProc
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.read_variable(jobHandle, &readValue)
	if C.VIX_OK != err {
		return "", &VixError{
			Operation: "vm.ReadVariable",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}
	defer C.Vix_FreeBuffer(unsafe.Pointer(readValue))

	return C.GoString(readValue), nil
}

// This function writes variables to the virtual machine state.
// This includes the virtual machine configuration, environment variables in
// the guest, and VMware "Guest Variables".
//
// Parameters:
//
//   varType: The type of variable to write. The currently supported values are:
//     * VM_GUEST_VARIABLE: A "Guest Variable". This is a runtime-only value;
//       it is never stored persistently. This is the same guest variable that
//       is exposed through the VMControl APIs, and is a simple way to
//       pass runtime values in and out of the guest.
//
//     * VM_CONFIG_RUNTIME_ONLY: The configuration state of the virtual
//       machine. This is the .vmx file that is stored on the host.
//       You can read this and it will return the persistent data. If you write
//       to this, it will only be a runtime change, so changes will be lost
//       when the VM powers off. Not supported on ESX hosts.
//
//     * GUEST_ENVIRONMENT_VARIABLE: An environment variable in the guest of
//       the VM. On a Windows NT series guest, writing these values is saved
//       persistently so they are immediately visible to every process.
//       On a Linux or Windows 9X guest, writing these values is not persistent
//       so they are only visible to the VMware tools process. Requires root
//       or Administrator privilege.
//
//   name: The name of the variable.
//   value: The value to be written.
//
// Remarks:
//
//   * The VM_CONFIG_RUNTIME_ONLY variable type is not supported on ESX hosts.
//
//   * You must call VM.LoginInGuest() before calling this function to write a
//     GUEST_ENVIRONMENT_VARIABLE value.
//     You do not have to call VM.LoginInGuest() to use this function to write a
//     VM_GUEST_VARIABLE or a VM_CONFIG_RUNTIME_ONLY value.
//
//   * Do not use the slash '/' character in a VM_GUEST_VARIABLE variable name;
//     doing so produces a VIX_E_INVALID_ARG error.
//
//   * Do not use the equal '=' character in the value parameter; doing so
//     produces a VIX_E_INVALID_ARG error.
//
//   * On Linux guests, you must login as root to change environment variables
//     (when variable type is GUEST_ENVIRONMENT_VARIABLE)
//     otherwise it produces a VIX_E_GUEST_USER_PERMISSIONS error.
//
//   * On Windows Vista guests, when variable type is GUEST_ENVIRONMENT_VARIABLE,
//     you must turn off User Account Control (UAC) in Control Panel >
//     User Accounts > User Accounts > Turn User Account on or off,
//     in order for this function to work.
//
// Since VMware Workstation 6.0
//
func (v *VM) WriteVariable(varType GuestVarType, name, value string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	vname := C.CString(name)
	vvalue := C.CString(value)
	defer C.free(unsafe.Pointer(vname))
	defer C.free(unsafe.Pointer(vvalue))

	jobHandle = C.VixVM_WriteVariable(v.handle,
		C.int(varType),
		vname,
		vvalue,
		0,
		nil, // callbackProc
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)

	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.WriteVariable",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// Restores the virtual machine to the state when the specified snapshot was
// created.
//
// Parameters:
//
//   snapshot: A Snapshot instance. Call VVM.GetRootSnapshot() to get a snapshot
//   instance.
//
//   options: Any applicable VMPowerOption. If the virtual machine was powered on
//   when the snapshot was created, then this will determine how the
//   virtual machine is powered back on. To prevent the virtual machine
//   from being powered on regardless of the power state when the
//   flag. VMPOWEROP_SUPPRESS_SNAPSHOT_POWERON is mutually exclusive
//   to all other VMPowerOpOptions.
//
// Remarks:
//
//   * Restores the virtual machine to the state when the specified snapshot was
//     created. This function can power on, power off, or suspend a virtual machine.
//     The resulting power state reflects the power state when the snapshot was
//     created.
//
//   * When you revert a powered on virtual machine and want it to display in the
//     Workstation user interface, options must have the VMPOWEROP_LAUNCH_GUI
//     flag, unless the VMPOWEROP_SUPPRESS_SNAPSHOT_POWERON is used.
//
//   * The ToolsState property of the virtual machine is undefined after the
//     snapshot is reverted.
//
//   * Starting in VMware Workstation 6.5, snapshot operations are allowed on
//     virtual machines that are part of a team.
//     Previously, this operation failed with error code PROPERTY_VM_IN_VMTEAM.
//     Team members snapshot independently so they can have different and
//     inconsistent snapshot states.
//
//   * This function is not supported when using the VMWARE_PLAYER provider
//
//   * If the virtual machine is open and powered off in the UI, this function
//     now closes the virtual machine in the UI before reverting to the snapshot.
//     To refresh this property, you must wait for tools in the guest.
//
//   * After reverting to a snapshot, you must call VM.WaitForToolsInGuest()
//     before executing guest operations or querying guest properties.
//
// Since VMware Server 1.0
//
func (v *VM) RevertToSnapshot(snapshot *Snapshot, options VMPowerOption) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_RevertToSnapshot(v.handle,
		snapshot.handle,
		0,                    // options
		C.VIX_INVALID_HANDLE, // propertyListHandle
		nil,                  // callbackProc
		nil)                  // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)

	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.RevertToSnapshot",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// Upgrades the virtual hardware version of the virtual machine to match the
// version of the VIX library.
// This has no effect if the virtual machine is already at the same version or
// at a newer version than the VIX library.
//
// Remarks:
//   * The virtual machine must be powered off to do this operation.
//   * When the VM is already up-to-date, the function returns without errors.
//   * This function is not supported when using the VMWARE_PLAYER provider.
//
// Since VMware Server 1.0
//
func (v *VM) UpgradeVHardware() error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_UpgradeVirtualHardware(v.handle,
		0,   // options
		nil, // callbackProc
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)

	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.UpgradeVHardware",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function establishes a guest operating system authentication context
// returning an instance of the Guest object
//
// Parameters:
//
//   username: The name of a user account on the guest operating system.
//
//   password: The password of the account identified by username
//
//   options: Must be LOGIN_IN_GUEST_NONE or
//   LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT - directs guest
//   commands invoked after the call to VM.LoginInGuest() to be run from within
//   the session of the user who is interactively logged into the guest operating
//   system.
//
//   See the remarks below for more information about LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT.
//
// Remarks:
//
//   * This function validates the account name and password in the guest OS.
//
//   * You must call this function before calling functions that perform operations
//     on the guest OS, such as those below. Otherwise you do not need to call this
//     function.
//
//   * Logins are supported on Linux and Windows. To log in as a Windows Domain
//     user, specify the 'username' parameter in the form "domain\username".
//
//   * This function does not respect access permissions on Windows 95, Windows 98,
//     and Windows ME, due to limitations of the permissions model in those systems.
//
//   * Other guest operating systems are not supported for login, including Solaris,
//     FreeBSD, and Netware.
//
//   * The option LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT should be used to
//     ensure that the functions Guest.CaptureScreenImage(), and
//     Guest.RunProgramInGuest() work correctly.
//
//   * All guest operations for a particular VM are done using the identity you
//     provide to VM.LoginInGuest(). As a result, guest operations are restricted
//     by file system privileges in the guest OS that apply to the user specified
//     in VM.LoginInGuest(). For example, Guest.RmDir() might fail if the user
//     named in VM.LoginInGuest() does not have access permissions to the directory
//     in the guest OS.
//
//   * VM.LoginInGuest() changes the behavior of Vix functions to use a user account.
//     It does not log a user into a console session on the guest OS. As a result,
//     you might not see the user logged in from within the guest OS. Moreover,
//     operations such as rebooting the guest do not clear the guest credentials.
//
//   * The virtual machine must be powered on before calling this function.
//
//   * VMware Tools must be installed and running on the guest OS before calling
//     this function.
//
//   * You can call VM.WaitForToolsInGuest() to wait for the tools to run.
//
//   * Once VM.LoginInGuest() has succeeded, the user session remains valid until
//     Guest.Logout() is called successfully, VM.LoginInGuest() is called
//     successfully with different user credentials, the virtual machine handle's
//     reference count reaches 0, or the client applications exits.
//
//   * The special login type VIX_CONSOLE_USER_NAME is no longer supported.
//
//   * Calling VM.LoginInGuest() with LOGIN_IN_GUEST_NONE as 'options' can be done
//     at any time when the VMware Tools are running in the guest.
//
//   * Calling VM.LoginInGuest() with the LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT
//     flag can only be done when there is an interactive user logged into the guest OS.
//     Specifically, the "interactive user" refers to the user who has logged into
//     the guest OS through the console (for instance, the user who logged into the Windows
//     log-in screen).
//
//   * The VIX user is the user whose credentials are being provided in the call to
//     VM.LoginInGuest().
//
//   * With LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT, there must be an
//     interactive user logged into the guest when the call to VM.LoginInGuest()
//     is made, and the VIX user must match the interactive user (they must have
//     same account in the guest OS).
//
//   * Using LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT will ensure that the
//     environment in which guest commands are executed is as close as possible to
//     the normal environment in which a user interacts with the guest OS. Without
//     LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT, commands may be run in a more
//     limited environment; however, omitting
//     LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT will ensure that commands can
//     be run regardless of whether an interactive user is present in the guest.
//
//   * On Linux guest operating systems, the
//     LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT flag requires that X11 be
//     installed and running.
//
// Since VMware Server 1.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
//
func (v *VM) LoginInGuest(username, password string, options GuestLoginOption) (*Guest, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	u := C.CString(username)
	p := C.CString(password)
	defer C.free(unsafe.Pointer(u))
	defer C.free(unsafe.Pointer(p))

	jobHandle = C.VixVM_LoginInGuest(v.handle,
		u,              // username
		p,              // password
		C.int(options), // options
		nil,            // callbackProc
		nil)            // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "vm.LoginInGuest",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	guest := &Guest{
		handle: v.handle,
	}
	return guest, nil
}

// Prepares to install VMware Tools on the guest operating system.
//
// Parameters:
//
//   options: May be either INSTALLTOOLS_MOUNT_TOOLS_INSTALLER or
//   INSTALLTOOLS_AUTO_UPGRADE. Either flag can be combined with the
//   INSTALLTOOLS_RETURN_IMMEDIATELY flag using the bitwise inclusive
//   OR operator (|). See remarks for more information.
//
// Remarks:
//
//   * If the option INSTALLTOOLS_MOUNT_TOOLS_INSTALLER is provided, the function
//     prepares an ISO image to install VMware Tools on the guest operating system.
//     If autorun is enabled, as it often is on Windows, installation begins,
//     otherwise you must initiate installation.
//     If VMware Tools is already installed, this function prepares to upgrade it
//     to the version matching the product.
//
//   * If the option VIX_INSTALLTOOLS_AUTO_UPGRADE is provided, the function
//     attempts to automatically upgrade VMware Tools without any user interaction
//     required, and then reboots the virtual machine. This option requires that a
//     version of VMware Tools already be installed. If VMware Tools is not
//     already installed, the function will fail.
//
//   * When the option INSTALLTOOLS_AUTO_UPGRADE is used on virtual machine with a
//     Windows guest operating system, the upgrade process may cause the Windows
//     guest to perform a controlled reset in order to load new device drivers.
//     If you intend to perform additional guest operations after upgrading the
//     VMware Tools, it is recommanded that after this task completes, that the
//     guest be reset using VM.Reset() with the VMPOWEROP_FROM_GUEST flag,
//     followed by calling VM.WaitForToolsInGuest() to ensure that the guest has
//     reached a stable state.
//
//   * If the option INSTALLTOOLS_AUTO_UPGRADE is provided and the newest version
//     of tools is already installed, the function will return successfully.
//     Some older versions of Vix may return VIX_E_TOOLS_INSTALL_ALREADY_UP_TO_DATE.
//
//   * If the INSTALLTOOLS_RETURN_IMMEDIATELY flag is set, this function will
//     return immediately after mounting the VMware Tools ISO image.
//
//   * If the INSTALLTOOLS_RETURN_IMMEDIATELY flag is not set for a WS host,
//     this function will return only after the installation successfully completes
//     or is cancelled.
//
//   * The virtual machine must be powered on to do this operation.
//
//   * If the Workstation installer calls for an ISO file that is not downloaded,
//     this function returns an error, rather than attempting to download the ISO
//     file.
//
// Since VMware Server 1.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
//
func (v *VM) InstallTools(options InstallToolsOption) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_InstallTools(v.handle,
		C.int(options), //options
		nil,            //commandLineArgs
		nil,            //callbackProc
		nil)            //clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.InstallTools",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function signals the job handle when VMware Tools has successfully
// started in the guest operating system.
// VMware Tools is a collection of services that run in the guest.
//
// Parameters:
//
//   timeout: The timeout in seconds. If VMware Tools has not started by
//   this time, the operation completes with an error.
//   If the value of this argument is zero or negative, then this
//   operation will wait indefinitely until the VMware Tools start
//   running in the guest operating system.
//
// Remarks:
//
//   * This function signals the job when VMware Tools has successfully started
//     in the guest operating system.
//     VMware Tools is a collection of services that run in the guest.
//
//   * VMware Tools must be installed and running for some Vix functions to
//     operate correctly.
//     If VMware Tools is not installed in the guest operating system, or if the
//     virtual machine is not powered on, this function reports an error.
//
//   * The ToolsState property of the virtual machine object is undefined until
//     VM.WaitForToolsInGuest() reports that VMware Tools is running.
//
//   * This function should be called after calling any function that resets or
//     reboots the state of the guest operating system, but before calling any
//     functions that require VMware Tools to be running. Doing so assures that
//     VMware Tools are once again up and running. Functions that reset the guest
//     operating system in this way include:
//
//     * VM.PowerOn()
//     * VM.Reset()
//     * VM.RevertToSnapshot()
//
// Since VMware Server 1.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
//
func (v *VM) WaitForToolsInGuest(timeout time.Duration) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_WaitForToolsInGuest(v.handle,
		C.int(timeout.Seconds()), // timeoutInSeconds
		nil, // callbackProc
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "vm.WaitForToolsInGuest",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

func (v *VM) updateVMX(updateFunc func(model *vmx.VirtualMachine) error) error {
	isVmRunning, err := v.IsRunning()
	if err != nil {
		return err
	}

	if isVmRunning {
		return &VixError{
			Operation: "vm.updateVMX",
			Code:      100000,
			Text:      "The VM has to be powered off in order to change its vmx settings",
		}
	}

	err = v.vmxfile.Read()
	if err != nil {
		return &VixError{
			Operation: "vm.updateVMX",
			Code:      300001,
			Text:      fmt.Sprintf("Error reading vmx file: %s", err),
		}
	}

	err = updateFunc(v.vmxfile.model)
	if err != nil {
		return &VixError{
			Operation: "vm.updateVMX",
			Code:      300002,
			Text:      fmt.Sprintf("Error changing vmx value: %s", err),
		}
	}

	err = v.vmxfile.Write()
	if err != nil {
		return &VixError{
			Operation: "vm.updateVMX",
			Code:      300003,
			Text:      fmt.Sprintf("Error writing vmx file: %s", err),
		}
	}

	return nil
}

// Sets memory size in megabytes
//
// VM has to be powered off in order to change
// this parameter
func (v *VM) SetMemorySize(size uint) error {
	if size == 0 {
		size = 4
	}

	// Makes sure memory size is divisible by 4, otherwise VMware is going to
	// silently fail, cancelling vix operations.
	if size%4 != 0 {
		size = uint(math.Floor(float64((size / 4) * 4)))
	}

	return v.updateVMX(func(model *vmx.VirtualMachine) error {
		model.Memsize = size
		return nil
	})
}

// Sets number of virtual cpus assigned to this machine.
//
// VM has to be powered off in order to change
// this parameter
func (v *VM) SetNumberVcpus(vcpus uint) error {
	if vcpus < 1 {
		vcpus = 1
	}

	return v.updateVMX(func(model *vmx.VirtualMachine) error {
		model.NumvCPUs = vcpus
		return nil
	})
}

// Sets virtual machine name
func (v *VM) SetDisplayName(name string) error {
	return v.updateVMX(func(model *vmx.VirtualMachine) error {
		model.DisplayName = name
		return nil
	})
}

// Gets virtual machine name
func (v *VM) DisplayName() (string, error) {
	return v.ReadVariable(VM_CONFIG_RUNTIME_ONLY, "displayname")
}

// Sets annotations for the virtual machine
func (v *VM) SetAnnotation(text string) error {
	return v.updateVMX(func(model *vmx.VirtualMachine) error {
		model.Annotation = text
		return nil
	})
}

// Returns the description or annotations added to the virtual machine
func (v *VM) Annotation() (string, error) {
	return v.ReadVariable(VM_CONFIG_RUNTIME_ONLY, "annotation")
}

func (v *VM) SetVirtualHwVersion(version string) error {
	return v.updateVMX(func(model *vmx.VirtualMachine) error {
		version, err := strconv.ParseInt(version, 10, 32)
		if err != nil {
			return err
		}
		model.Vhardware.Compat = "hosted"
		model.Vhardware.Version = int(version)
		return nil
	})
}
