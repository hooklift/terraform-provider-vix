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
	"runtime"
	"unsafe"
)

// VixPowerState
//
// These are the possible values reported for VIX_PROPERTY_VM_POWER_STATE
// property. These values are bitwise flags. The actual value returned for may
// be a bitwise OR of one more of these flags, along with other reserved values
// not documented here. They represent runtime information about the state of
// the virtual machine. To test the value of the property, use the
// Vix.GetProperties() function.
//
// Since VMware Server 1.0.
type VMPowerState int

const (
	// Indicates that VM.PowerOff() has been called, but the operation itself
	// has not completed.
	POWERSTATE_POWERING_OFF VMPowerState = C.VIX_POWERSTATE_POWERING_OFF

	// Indicates that the virtual machine is not running.
	POWERSTATE_POWERED_OFF VMPowerState = C.VIX_POWERSTATE_POWERED_OFF

	// Indicates that VM.PowerOn() has been called, but the operation itself
	// has not completed.
	POWERSTATE_POWERING_ON VMPowerState = C.VIX_POWERSTATE_POWERING_ON

	// Indicates that the virtual machine is running.
	POWERSTATE_POWERED_ON VMPowerState = C.VIX_POWERSTATE_POWERED_ON

	// Indicates that VM.Suspend() has been called, but the operation itself
	// has not completed.
	POWERSTATE_SUSPENDING VMPowerState = C.VIX_POWERSTATE_SUSPENDING

	// Indicates that the virtual machine is suspended. Use VM.PowerOn() to
	// resume the virtual machine.
	POWERSTATE_SUSPENDED VMPowerState = C.VIX_POWERSTATE_SUSPENDED

	// Indicates that the virtual machine is running and the VMware Tools
	// suite is active. See also the VixToolsState property.
	POWERSTATE_TOOLS_RUNNING VMPowerState = C.VIX_POWERSTATE_TOOLS_RUNNING

	// Indicates that VM.Reset() has been called, but the operation itself
	// has not completed.
	POWERSTATE_RESETTING VMPowerState = C.VIX_POWERSTATE_RESETTING

	// Indicates that a virtual machine state change is blocked, waiting for
	// user interaction.
	POWERSTATE_BLOCKED_ON_MSG VMPowerState = C.VIX_POWERSTATE_BLOCKED_ON_MSG
)

// VixFindItemType
//
// These are the types of searches you can do with Host.FindItems().
//
// Since VMware Server 1.0.
type SearchType int

const (
	// Finds all virtual machines currently running on the host.
	FIND_RUNNING_VMS SearchType = C.VIX_FIND_RUNNING_VMS

	// Finds all virtual machines registered on the host.
	// This search applies only to platform products that maintain a virtual
	// machine registry,
	// such as ESX/ESXi and VMware Server, but not Workstation or Player.
	FIND_REGISTERED_VMS SearchType = C.VIX_FIND_REGISTERED_VMS
)

// VixToolsState
//
// These are the possible values reported for VIX_PROPERTY_VM_TOOLS_STATE.
// They represent runtime information about the VMware Tools suite in the guest
// operating system.
// To test the value of the property, use the Vix.GetProperties() function.
//
// Since VMware Server 1.0.
type GuestToolsState int

const (
	// Indicates that Vix is unable to determine the VMware Tools status.
	TOOLSSTATE_UNKNOWN GuestToolsState = C.VIX_TOOLSSTATE_UNKNOWN

	// Indicates that VMware Tools is running in the guest operating system.
	TOOLSSTATE_RUNNING GuestToolsState = C.VIX_TOOLSSTATE_RUNNING

	// Indicates that VMware Tools is not installed in the guest operating system.
	TOOLSSTATE_NOT_INSTALLED GuestToolsState = C.VIX_TOOLSSTATE_NOT_INSTALLED
)

// Service Provider
type Provider int

const (
	// vCenter Server, ESX/ESXi hosts, and VMware Server 2.0
	VMWARE_VI_SERVER Provider = C.VIX_SERVICEPROVIDER_VMWARE_VI_SERVER

	// VMware Workstation
	VMWARE_WORKSTATION Provider = C.VIX_SERVICEPROVIDER_VMWARE_WORKSTATION

	// VMware Workstation (shared mode)
	VMWARE_WORKSTATION_SHARED Provider = C.VIX_SERVICEPROVIDER_VMWARE_WORKSTATION_SHARED

	// With VMware Player
	VMWARE_PLAYER Provider = C.VIX_SERVICEPROVIDER_VMWARE_PLAYER

	// VMware Server 1.0.x
	VMWARE_SERVER Provider = C.VIX_SERVICEPROVIDER_VMWARE_SERVER
)

type EventType int

const (
	JOB_COMPLETED EventType = C.VIX_EVENTTYPE_JOB_COMPLETED
	JOB_PROGRESS  EventType = C.VIX_EVENTTYPE_JOB_PROGRESS
	FIND_ITEM     EventType = C.VIX_EVENTTYPE_FIND_ITEM
)

type HostOption int

const (
	HOST_OPTIONS_NONE = 0x0
	VERIFY_SSL_CERT   = C.VIX_HOSTOPTION_VERIFY_SSL_CERT
)

type GuestLoginOption int

const (
	LOGIN_IN_GUEST_NONE                            GuestLoginOption = 0x0
	LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT GuestLoginOption = C.VIX_LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT
)

type GuestVarType int

const (
	VM_GUEST_VARIABLE          GuestVarType = C.VIX_VM_GUEST_VARIABLE
	VM_CONFIG_RUNTIME_ONLY     GuestVarType = C.VIX_VM_CONFIG_RUNTIME_ONLY
	GUEST_ENVIRONMENT_VARIABLE GuestVarType = C.VIX_GUEST_ENVIRONMENT_VARIABLE
)

type SharedFolderOption int

const (
	SHAREDFOLDER_WRITE_ACCESS = C.VIX_SHAREDFOLDER_WRITE_ACCESS
)

type CloneType int

const (
	CLONETYPE_FULL   CloneType = C.VIX_CLONETYPE_FULL
	CLONETYPE_LINKED CloneType = C.VIX_CLONETYPE_LINKED
)

type CreateSnapshotOption int

const (
	SNAPSHOT_INCLUDE_MEMORY CreateSnapshotOption = C.VIX_SNAPSHOT_INCLUDE_MEMORY
)

type VmDeleteOption int

const (
	VMDELETE_KEEP_FILES VmDeleteOption = 0x0
	VMDELETE_FORCE      VmDeleteOption = 0x3
	VMDELETE_DISK_FILES VmDeleteOption = C.VIX_VMDELETE_DISK_FILES
)

type VMPowerOption int

const (
	VMPOWEROP_NORMAL                    VMPowerOption = C.VIX_VMPOWEROP_NORMAL
	VMPOWEROP_FROM_GUEST                VMPowerOption = C.VIX_VMPOWEROP_FROM_GUEST
	VMPOWEROP_SUPPRESS_SNAPSHOT_POWERON VMPowerOption = C.VIX_VMPOWEROP_SUPPRESS_SNAPSHOT_POWERON
	VMPOWEROP_LAUNCH_GUI                VMPowerOption = C.VIX_VMPOWEROP_LAUNCH_GUI
	VMPOWEROP_START_VM_PAUSED           VMPowerOption = C.VIX_VMPOWEROP_START_VM_PAUSED
)

type RunProgramOption int

const (
	RUNPROGRAM_WAIT               RunProgramOption = 0x0
	RUNPROGRAM_RETURN_IMMEDIATELY RunProgramOption = C.VIX_RUNPROGRAM_RETURN_IMMEDIATELY
	RUNPROGRAM_ACTIVATE_WINDOW    RunProgramOption = C.VIX_RUNPROGRAM_ACTIVATE_WINDOW
)

type InstallToolsOption int

const (
	INSTALLTOOLS_MOUNT_TOOLS_INSTALLER InstallToolsOption = C.VIX_INSTALLTOOLS_MOUNT_TOOLS_INSTALLER
	INSTALLTOOLS_AUTO_UPGRADE          InstallToolsOption = C.VIX_INSTALLTOOLS_AUTO_UPGRADE
	INSTALLTOOLS_RETURN_IMMEDIATELY    InstallToolsOption = C.VIX_INSTALLTOOLS_RETURN_IMMEDIATELY
)

type RemoveSnapshotOption int

const (
	SNAPSHOT_REMOVE_NONE     RemoveSnapshotOption = 0x0
	SNAPSHOT_REMOVE_CHILDREN RemoveSnapshotOption = C.VIX_SNAPSHOT_REMOVE_CHILDREN
)

type FileAttr int

const (
	FILE_ATTRIBUTES_DIRECTORY FileAttr = C.VIX_FILE_ATTRIBUTES_DIRECTORY
	FILE_ATTRIBUTES_SYMLINK   FileAttr = C.VIX_FILE_ATTRIBUTES_SYMLINK
)

// Configuration struct for the Connect function
type ConnectConfig struct {
	// Hostname varies by product platform. With vCenter Server, ESX/ESXi hosts,
	// VMware Workstation (shared mode) and VMware Server 2.0,
	// use a URL of the form "https://<hostName>:<port>/sdk"
	// where <hostName> is either the DNS name or IP address.
	// If missing, <port> may default to 443 (see Remarks below).
	//
	// In VIX API 1.10 and later, you can omit "https://" and "/sdk" specifying
	// just the DNS name or IP address.
	// Credentials are required even for connections made locally.
	// With Workstation, use nil to connect to the local host.
	// With VMware Server 1.0.x, use the DNS name or IP address for remote
	// connections, or the same as Workstation for local connections.
	Hostname string

	// TCP/IP port on the remote host.
	// With VMware Workstation and VMware Player, let it empty for localhost.
	// With ESX/ESXi hosts, VMware Workstation (shared mode) and VMware Server 2.0
	// you specify port number within the hostName parameter, so this parameter is
	// ignored (see Connect Remarks below).
	Port uint

	// Username for authentication on the remote machine.
	// With VMware Workstation, VMware Player, and VMware Server 1.0.x,
	// let it empty to authenticate as the current user on localhost.
	// With vCenter Server, ESX/ESXi hosts, VMware Workstation (shared mode)
	// and VMware Server 2.0, you must use a valid login.
	Username string

	// Password for authentication on the remote machine.
	// With VMware Workstation, VMware Player, and VMware Server 1.0.x,
	// let it empty to authenticate as the current user on localhost.
	// With ESX/ESXi, VMware Workstation (shared mode) and VMware Server 2.0, you
	// must use a valid password.
	Password string

	// Provider is the VMware product you would like to connect to:
	// * With vCenter Server, ESX/ESXi hosts, and VMware Server 2.0, use
	//   VMWARE_VI_SERVER.
	// * With VMware Workstation, use VMWARE_WORKSTATION.
	// * With VMware Workstation (shared mode), use VMWARE_WORKSTATION_SHARED.
	// * With VMware Player, use VMWARE_PLAYER.
	// * With VMware Server 1.0.x, use VMWARE_SERVER.
	Provider Provider

	// Bitwised option parameter.
	Options HostOption
}

// Connects to a Provider
//
// Parameters:
//
//  config: See type ConnectConfig documentation for details
//
// Remarks:
//   * To specify the local host (where the API client runs) with VMware
//     Workstation and VMware Player, pass empty values for the hostname, port,
//     login, and password parameters or just don't set them.
//
//   * With vCenter Server, ESX/ESXi hosts, and VMware Server 2.0, the URL for
//     the hostname argument may specify the port.
//     Otherwise a HTTPS connection is attempted on port 443. HTTPS is strongly
//     recommended.
//     Port numbers are set during installation of Server 2.0. The installer's
//     default HTTP and HTTPS values are 8222 and 8333 for Server on Windows, or
//     (if not already in use) 80 and 443 for Server on Linux, and 902 for the
//     automation socket, authd. If connecting to a virtual machine through a
//     firewall, port 902 and the communicating port must be opened to allow
//     guest operations.
//
//   * If a VMware ESX host is being managed by a VMware VCenter Server, you
//     should call VixHost_Connect with the hostname or IP address of the VCenter
//     server, not the ESX host.
//     Connecting directly to an ESX host while bypassing its VCenter Server can
//     cause state inconsistency.
//
//   * On Windows, this function should not be called multiple times with
//     different service providers in the same process; doing so will result in
//     a VIX_E_WRAPPER_MULTIPLE_SERVICEPROVIDERS error.
//     A single client process can connect to multiple hosts as long as it
//     connects using the same service provider type.
//
//   * To enable SSL certificate verification, set the value of the options
//     parameter to include the bit flag specified by VERIFY_SSL_CERT.
//     This option can also be set in the VMware config file by assigning
//     vix.enableSslCertificateCheck as TRUE or FALSE.
//
//     The vix.sslCertificateFile config option specifies the path to a file
//     containing CA certificates in PEM format.
//
//     The vix.sslCertificateDirectory config option can specify a directory
//     containing files that each contain a CA certificate.
//     Upon encountering a SSL validation error, the host handle is not created
//     with a resulting error code of E_NET_HTTP_SSL_SECURITY.
//
//   * With VMware vCenter Server and ESX/ESXi 4.0 hosts, an existing VI API
//     session can be used instead of the username/password pair to authenticate
//     when connecting. To use an existing VI API session, a VI "clone ticket"
//     is required; call the VI API AcquireCloneTicket() method of the
//     SessionManager object to get this ticket.
//     Using the ticket string returned by this method, call vix.Connect()
//     with "" as the 'username' and the ticket as the 'password'.
//
// Since VMware Server 1.0
func Connect(config ConnectConfig) (*Host, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var hostHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	chostname := C.CString(config.Hostname)
	cusername := C.CString(config.Username)
	cpassword := C.CString(config.Password)
	defer C.free(unsafe.Pointer(chostname))
	defer C.free(unsafe.Pointer(cusername))
	defer C.free(unsafe.Pointer(cpassword))

	jobHandle = C.VixHost_Connect(C.VIX_API_VERSION,
		C.VixServiceProvider(config.Provider),
		chostname,
		C.int(config.Port),
		cusername,
		cpassword,
		C.VixHostOptions(config.Options),
		C.VIX_INVALID_HANDLE, // propertyListHandle
		nil,                  // callbackProc
		nil)                  // clientData

	err = C.get_vix_handle(jobHandle,
		C.VIX_PROPERTY_JOB_RESULT_HANDLE,
		&hostHandle,
		C.VIX_PROPERTY_NONE)

	defer C.Vix_ReleaseHandle(jobHandle)

	if C.VIX_OK != err {
		return nil, &Error{
			Operation: "vix.Connect",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	host := &Host{
		handle:   hostHandle,
		Provider: config.Provider,
	}

	runtime.SetFinalizer(host, cleanupHost)

	return host, nil
}

// Private function to clean up host handle
func cleanupHost(host *Host) {
	if host.handle != C.VIX_INVALID_HANDLE {
		host.Disconnect()
	}
}

// GoVix Error
type Error struct {
	// The GoVix operation involved at the time of the error
	Operation string
	// Error code
	Code int
	// Description of the erro
	Text string
}

// Returns a description of the error along with its code and operation
func (e *Error) Error() string {
	return fmt.Sprintf("VIX Error: %s, code: %d, operation: %s", e.Text, e.Code, e.Operation)
}
