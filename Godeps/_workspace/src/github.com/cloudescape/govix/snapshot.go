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
	"runtime"
	"unsafe"
)

type Snapshot struct {
	// Internal VIX handle
	handle C.VixHandle
}

// Returns user defined name for the snapshot.
func (s *Snapshot) Name() (string, error) {
	var err C.VixError = C.VIX_OK
	var name *C.char

	err = C.get_property(s.handle,
		C.VIX_PROPERTY_SNAPSHOT_DISPLAYNAME,
		unsafe.Pointer(&name))

	defer C.Vix_FreeBuffer(unsafe.Pointer(name))

	if C.VIX_OK != err {
		return "", &VixError{
			Operation: "snapshot.Name",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return C.GoString(name), nil
}

// Returns user defined description for the snapshot.
func (s *Snapshot) Description() (string, error) {
	var err C.VixError = C.VIX_OK
	var desc *C.char

	err = C.get_property(s.handle,
		C.VIX_PROPERTY_SNAPSHOT_DESCRIPTION,
		unsafe.Pointer(&desc))

	defer C.Vix_FreeBuffer(unsafe.Pointer(desc))

	if C.VIX_OK != err {
		return "", &VixError{
			Operation: "snapshot.Description",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return C.GoString(desc), nil
}

// This function returns the child snapshot corresponding to the index parameter
//
// Parameters:
//
//  index: Index into the list of snapshots.
//
// Remarks:
//
//   * Snapshots are indexed from 0 to n-1, where n is the number of child
//     snapshots. Use the function Snapshot.NumChildren() to get the value of n.
//   * This function is not supported when using the VMWARE_PLAYER provider.
//
// Since VMware Workstation 6.0
func (s *Snapshot) Child(index int) (*Snapshot, error) {
	var snapshotHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	err = C.VixSnapshot_GetChild(s.handle,
		C.int(index),    //index
		&snapshotHandle) //(output) A handle to the child snapshot.

	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "snapshot.Child",
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

// This function returns the number of child snapshots of a specified snapshot.
//
// Remarks:
//
//   * This function is not supported when using the VMWARE_PLAYER provider.
//
// Since VMware Workstation 6.0.
func (s *Snapshot) NumChildren() (int, error) {
	var err C.VixError = C.VIX_OK
	var numChildren *C.int

	err = C.VixSnapshot_GetNumChildren(s.handle, numChildren)

	if C.VIX_OK != err {
		return 0, &VixError{
			Operation: "snapshot.NumChildren",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return int(*numChildren), nil
}

// This function returns the parent of a snapshot.
//
// Remarks:
//
//   * This function is not supported when using the VMWARE_PLAYER provider
//
// Since VMware Workstation 6.0
func (s *Snapshot) Parent() (*Snapshot, error) {
	var snapshotHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	err = C.VixSnapshot_GetParent(s.handle,
		&snapshotHandle) //(output) A handle to the child snapshot.

	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "snapshot.Parent",
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

// Private function to clean up snapshot handle
func cleanupSnapshot(s *Snapshot) {
	if s.handle != C.VIX_INVALID_HANDLE {
		C.Vix_ReleaseHandle(s.handle)
		s.handle = C.VIX_INVALID_HANDLE
	}
}
