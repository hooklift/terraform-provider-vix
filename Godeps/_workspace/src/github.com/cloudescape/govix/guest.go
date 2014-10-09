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
	"unsafe"
)

type Guest struct {
	handle C.VixHandle
}

func (g *Guest) SharedFoldersParentDir() (string, error) {
	var err C.VixError = C.VIX_OK
	var path *C.char

	err = C.get_property(g.handle,
		C.VIX_PROPERTY_GUEST_SHAREDFOLDERS_SHARES_PATH,
		unsafe.Pointer(&path))

	defer C.Vix_FreeBuffer(unsafe.Pointer(path))

	if C.VIX_OK != err {
		return "", &VixError{
			Operation: "guest.SharedFoldersParentDir",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return C.GoString(path), nil
}

// Copies a file or directory from the guest operating system to the local
// system (where the Vix client is running).
//
// Parameters:
//
//   guestpath: The path name of a file on a file system available to the guest.
//   hostpath: The path name of a file on a file system available to the Vix
//   client.
//
func (g *Guest) CopyFileToHost(guestpath, hostpath string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	gpath := C.CString(guestpath)
	hpath := C.CString(hostpath)
	defer C.free(unsafe.Pointer(gpath))
	defer C.free(unsafe.Pointer(hpath))

	jobHandle = C.VixVM_CopyFileFromGuestToHost(g.handle,
		gpath,                // src name
		hpath,                // dest name
		0,                    // options
		C.VIX_INVALID_HANDLE, // propertyListHandle
		nil,                  // callbackProc
		nil)                  // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "guest.CopyFileToHost",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function creates a directory in the guest operating system.
//
// Parameters:
//
//   path: Directory path to be created in the guest OS
//
// Remarks:
//
//   * If the parent directories for the specified path do not exist, this
//     function will create them.
//
//   * If the directory already exists, the error will be set to
//     VIX_E_FILE_ALREADY_EXISTS.
//
//   * Only absolute paths should be used for files in the guest; the resolution
//     of relative paths is not specified.
//
// Since Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) MkDir(path string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	jobHandle = C.VixVM_CreateDirectoryInGuest(g.handle,
		cpath,                // path name
		C.VIX_INVALID_HANDLE, // propertyListHandle
		nil,                  // callbackProc
		nil)                  // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "guest.MkDir",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function creates a temporary file in the guest operating system.
// The user is responsible for removing the file when it is no longer needed.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) MkTemp() (string, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var tempFilePath *C.char

	jobHandle = C.VixVM_CreateTempFileInGuest(g.handle,
		0,                    // options
		C.VIX_INVALID_HANDLE, // propertyListHandle
		nil,                  // callbackProc
		nil)                  // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.get_temp_filepath(jobHandle, tempFilePath)
	defer C.Vix_FreeBuffer(unsafe.Pointer(tempFilePath))

	if C.VIX_OK != err {
		return "", &VixError{
			Operation: "guest.MkTemp",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return C.GoString(tempFilePath), nil
}

// This function deletes a directory in the guest operating system.
// Any files or subdirectories in the specified directory will also be deleted.
//
// Parameters:
//
//   path: Directory path to be deleted in the guest OS
//
// Remarks:
//
//   * Only absolute paths should be used for files in the guest;
//     the resolution of relative paths is not specified.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) RmDir(path string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	jobHandle = C.VixVM_DeleteDirectoryInGuest(g.handle,
		cpath, // path name
		0,     // options
		nil,   // callbackProc
		nil)   // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "guest.RmDir",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function deletes a file in the guest operating system.
//
// Parameters:
//
//   filepath: file path to be deleted in the guest OS
//
// Remarks:
//   * Only absolute paths should be used for files in the guest; the resolution
//     of relative paths is not specified.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) RmFile(filepath string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	fpath := C.CString(filepath)
	defer C.free(unsafe.Pointer(fpath))

	jobHandle = C.VixVM_DeleteFileInGuest(g.handle,
		fpath, // file path name
		nil,   // callbackProc
		nil)   // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "guest.RmFile",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function tests the existence of a directory in the guest operating
// system.
//
// Parameters:
//
//   path: Directory path in the guest OS to be checked.
//
// Remarks:
//
//   * Only absolute paths should be used for files in the guest; the resolution
//     of relative paths is not specified.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) IsDir(path string) (bool, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var result C.int

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	jobHandle = C.VixVM_DirectoryExistsInGuest(g.handle,
		cpath, // dir path name
		nil,   // callbackProc
		nil)   // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.is_file_or_dir(jobHandle, &result)
	if C.VIX_OK != err {
		return false, &VixError{
			Operation: "guest.IsDir",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	if int(result) == C.FALSE {
		return false, nil
	}

	return true, nil
}

// This function tests the existence of a file in the guest operating system.
//
// Parameters:
//
//   filepath: The path to the file to be tested.
//
// Remarks:
//
//   * Only absolute paths should be used for files in the guest; the resolution
//     of relative paths is not specified.
//
//   * If filepath exists as a file system object, but is not a normal file (e.g.
//     it is a directory, device, UNIX domain socket, etc),
//     then VIX_OK is returned, and VIX_PROPERTY_JOB_RESULT_GUEST_OBJECT_EXISTS
//     is set to FALSE.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) IsFile(filepath string) (bool, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var result C.int

	fpath := C.CString(filepath)
	defer C.free(unsafe.Pointer(fpath))

	jobHandle = C.VixVM_FileExistsInGuest(g.handle,
		fpath, // dir path name
		nil,   // callbackProc
		nil)   // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.is_file_or_dir(jobHandle, &result)
	if C.VIX_OK != err {
		return false, &VixError{
			Operation: "guest.IsFile",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	if int(result) == C.FALSE {
		return false, nil
	}

	return true, nil
}

// This function returns information about a file in the guest operating system.
//
// Parameters:
//
//   filepath: The path name of the file in the guest.
//
// Remarks:
//   * Only absolute paths should be used for files in the guest;
//     the resolution of relative paths is not specified.
//
// Since VMware Workstation 6.5
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) FileInfo(filepath string) (*GuestFile, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var fsize *C.int64
	var flags *C.int
	var modtime *C.int64

	fpath := C.CString(filepath)
	defer C.free(unsafe.Pointer(fpath))

	jobHandle = C.VixVM_GetFileInfoInGuest(g.handle,
		fpath, // file path name
		nil,   // callbackProc
		nil)   // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.get_file_info(jobHandle, fsize, flags, modtime)
	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "guest.FileInfo",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return &GuestFile{
		Path:    filepath,
		Size:    int64(*fsize),
		Attrs:   FileAttr(*flags),
		Modtime: int64(*modtime),
	}, nil
}

// This function terminates a process in the guest operating system.
//
// Parameters:
//
//   pid: The ID of the process to be killed.
//
// Remarks:
//
//   * Depending on the behavior of the guest operating system, there may be a
//     short delay after the job completes before the process truly disappears.
//
//   * Because of differences in how various Operating Systems handle process IDs,
//     Vix may return either VIX_E_INVALID_ARG or VIX_E_NO_SUCH_PROCESS
//     for invalid process IDs.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) Kill(pid uint64) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_KillProcessInGuest(g.handle,
		C.uint64(pid), // file path name
		0,             // options
		nil,           // callbackProc
		nil)           // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "guest.Kill",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This struct is used to return File information from the guest OS
type GuestFile struct {
	// Path to the file
	Path string

	// File size as a 64-bit integer. This is 0 for directories.
	Size int64

	// The modification time of the file or directory as a 64-bit integer
	// specifying seconds since the epoch.
	Modtime int64

	// File attributes, the possible attributes are:
	//   * FILE_ATTRIBUTES_DIRECTORY: Set if the pathname identifies a
	//   directory.
	//
	//   * FILE_ATTRIBUTES_SYMLINK: Set if the pathname identifies a symbolic
	//   link file.
	//
	// Either attribute will be combined using the bitwise inclusive
	// OR operator (|).
	Attrs FileAttr
}

// This function lists a directory in the guest operating system.
//
// Parameters:
//
//   dir: The path name of a directory to be listed.
//
// Remarks:
//
//   * Only absolute paths should be used for files in the guest; the resolution
//     of relative paths is not specified.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) Ls(dir string) ([]*GuestFile, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var files []*GuestFile

	guestdir := C.CString(dir)
	defer C.free(unsafe.Pointer(guestdir))

	jobHandle = C.VixVM_ListDirectoryInGuest(g.handle, guestdir, 0, nil, nil)
	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "guest.Ls.ListDirectoryInGuest",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	num := C.VixJob_GetNumProperties(jobHandle, C.VIX_PROPERTY_JOB_RESULT_ITEM_NAME)
	for i := 0; i < int(num); i++ {
		var name *C.char
		var size *C.int64
		var modtime *C.int64
		var attrs *C.int

		gfile := &GuestFile{}

		err = C.get_guest_file(jobHandle, C.int(i), name, size, modtime, attrs)
		if C.VIX_OK != err {
			return nil, &VixError{
				Operation: "guest.Ls.get_guest_file",
				Code:      int(err & 0xFFFF),
				Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
			}
		}

		gfile.Path = C.GoString(name)
		C.Vix_FreeBuffer(unsafe.Pointer(name))

		gfile.Size = int64(*size)
		gfile.Modtime = int64(*modtime)
		gfile.Attrs = FileAttr(*attrs)

		files = append(files, gfile)
	}

	return files, nil
}

// Use to return process information of processes running in a guest OS
type GuestProcess struct {
	// Name of the process
	Name string

	// Process ID
	Pid uint64

	// User running the process
	Owner string

	// Command used to run the process
	Cmdline string

	// Whether is debugged
	IsDebugged bool

	// When did it start running
	StartTime int
}

// This function lists the running processes in the guest
// operating system.
//
// Since Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) Ps() ([]*GuestProcess, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var processes []*GuestProcess

	jobHandle = C.VixVM_ListProcessesInGuest(g.handle, 0, nil, nil)
	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return nil, &VixError{
			Operation: "guest.Ps",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	num := C.VixJob_GetNumProperties(jobHandle, C.VIX_PROPERTY_JOB_RESULT_ITEM_NAME)
	for i := 0; i < int(num); i++ {
		var name *C.char
		var pid *C.uint64
		var owner *C.char
		var cmdline *C.char
		var isDebugged *C.Bool
		var startTime *C.int

		gprocess := &GuestProcess{}

		err = C.get_guest_process(jobHandle, C.int(i),
			name,
			pid,
			owner,
			cmdline,
			isDebugged,
			startTime)

		if C.VIX_OK != err {
			return nil, &VixError{
				Operation: "guest.Ps.get_guest_process",
				Code:      int(err & 0xFFFF),
				Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
			}
		}

		gprocess.Name = C.GoString(name)
		C.Vix_FreeBuffer(unsafe.Pointer(name))

		gprocess.Pid = uint64(*pid)

		gprocess.Owner = C.GoString(owner)
		C.Vix_FreeBuffer(unsafe.Pointer(owner))

		gprocess.Cmdline = C.GoString(cmdline)
		C.Vix_FreeBuffer(unsafe.Pointer(cmdline))

		if *isDebugged == 1 {
			gprocess.IsDebugged = true
		} else {
			gprocess.IsDebugged = false
		}

		gprocess.StartTime = int(*startTime)

		processes = append(processes, gprocess)
	}

	return processes, nil
}

// This function removes any guest operating system authentication
// context created by a previous call to VM.LoginInGuest().
//
// Remarks:
//   * This function has no effect and returns success if VM.LoginInGuest()
//     has not been called.
//   * If you call this function while guest operations are in progress,
//     subsequent operations may fail with a permissions error.
//     It is best to wait for guest operations to complete before logging out.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) Logout() error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	jobHandle = C.VixVM_LogoutFromGuest(g.handle,
		nil, // callbackProc
		nil) // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "guest.Logout",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}

// This function runs a program in the guest operating system.
// The program must be stored on a file system available to the guest
// before calling this function.
//
// Parameters:
//
//   path: The path name of an executable file on the guest operating system.
//   args: A string to be passed as command line arguments to the executable.
//   options: Run options for the program. See the remarks below.
//
// Remarks:
//
//   * This function runs a program in the guest operating system.
//     The program must be stored on a file system available to the guest
//     before calling this function.
//
//   * The current working directory for the program in the guest is not defined.
//     Absolute paths should be used for files in the guest, including
//     command-line arguments.
//
//   * If the program to run in the guest is intended to be visible to the user
//     in the guest, such as an application with a graphical user interface,
//     you must call VM.LoginInGuest() with
//     LOGIN_IN_GUEST_REQUIRE_INTERACTIVE_ENVIRONMENT as the option before calling
//     this function. This will ensure that the program is run within a
//     graphical session that is visible to the user.
//
//   * If the options parameter is RUNPROGRAM_WAIT, this function will block and
//     return only when the program exits in the guest operating system.
//     Alternatively, you can pass RUNPROGRAM_RETURN_IMMEDIATELY as the value of
//     the options parameter, and this function will return as soon as the program
//     starts in the guest.
//
//   * For Windows guest operating systems, when running a program with a
//     graphical user interface, you can pass RUNPROGRAM_ACTIVATE_WINDOW as the
//     value of the options parameter. This option will ensure that the
//     application's window is visible and not minimized on the guest's screen.
//     This can be combined with the RUNPROGRAM_RETURN_IMMEDIATELY flag using
//     the bitwise inclusive OR operator (|). RUNPROGRAM_ACTIVATE_WINDOW
//     has no effect on Linux guest operating systems.
//
//   * On a Linux guest operating system, if you are running a program with a
//     graphical user interface, it must know what X Windows display to use,
//     for example host:0.0, so it can make the program visible on that display.
//     Do this by passing the -display argument to the program, if it supports
//     that argument, or by setting the DISPLAY environment variable on the guest.
//     See documentation on VM.WriteVariable()
//
//   * This functions returns three parameters:
//     PROCESS_ID: the process id; however, if the guest has
//     an older version of Tools (those released with Workstation 6 and earlier)
//     and the RUNPROGRAM_RETURN_IMMEDIATELY flag is used, then the process ID
//     will not be returned from the guest and this property will be 0
//     ELAPSED_TIME: the process elapsed time in seconds;
//     EXIT_CODE: the process exit code.
//     If the option parameter is RUNPROGRAM_RETURN_IMMEDIATELY, the latter two
//     will both be 0.
//
//   * Depending on the behavior of the guest operating system, there may be a
//     short delay after the job completes before the process is visible in the
//     guest operating system. Sometimes you may want to use a goroutine to not
//     block your app while this finishes.
//
// Since VMware Server 1.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) RunProgram(path, args string, options RunProgramOption) (uint64, int, int, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var pid *C.uint64
	var elapsedtime *C.int
	var exitCode *C.int

	cpath := C.CString(path)
	cargs := C.CString(args)
	defer C.free(unsafe.Pointer(cpath))
	defer C.free(unsafe.Pointer(cargs))

	jobHandle = C.VixVM_RunProgramInGuest(g.handle,
		cpath, //guestProgramName
		cargs, //commandLineArgs
		C.VixRunProgramOptions(options), //options
		C.VIX_INVALID_HANDLE,            //propertyListHandle
		nil,                             // callbackProc
		nil)                             // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.get_program_output(jobHandle, pid, elapsedtime, exitCode)

	if C.VIX_OK != err {
		return 0, 0, 0, &VixError{
			Operation: "guest.RunProgram",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return uint64(*pid), int(*elapsedtime), int(*exitCode), nil
}

// This function runs a script in the guest operating system.
//
// Parameters:
//
//   shell: The path to the script interpreter, or NULL to use cmd.exe as
//   the interpreter on Windows.
//
//   script: The text of the script.
//
//   options: Run options for the program. See the remarks below.
//
// Remarks:
//
//   * This function runs the script in the guest operating system.
//
//   * The current working directory for the script executed in the guest is
//     not defined. Absolute paths should be used for files in the guest,
//     including the path to the shell or interpreter, and any files referenced
//     in the script text.
//
//   * If the options parameter is RUNPROGRAM_WAIT, this function will block and
//     return only when the program exits in the guest operating system.
//     Alternatively, you can pass RUNPROGRAM_RETURN_IMMEDIATELY as the value of
//     the options parameter, and this makes the function to return as soon as the
//     program starts in the guest.
//
//   * The following properties will be returned:
//     PROCESS_ID: the process id; however, if the guest has an older version of
//                 Tools (those released with Workstation 6 and earlier) and
//                 the RUNPROGRAM_RETURN_IMMEDIATELY flag is used, then the
//                 process ID will not be returned from the guest and this
//                 property will return 0.
//     ELAPSED_TIME: the process elapsed time;
//     PROGRAM_EXIT_CODE: the process exit code.
//
//   * If the option parameter is RUNPROGRAM_RETURN_IMMEDIATELY, the latter two
//     will both be 0.
//
//   * Depending on the behavior of the guest operating system, there may be a
//     short delay after the function returns before the process is visible in the
//     guest operating system.
//
//   * If the total size of the specified interpreter and the script text is
//     larger than 60536 bytes, then the error VIX_E_ARGUMENT_TOO_BIG is returned.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) RunScript(shell, args string, options RunProgramOption) (uint64, int, int, error) {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK
	var pid *C.uint64
	var elapsedtime *C.int
	var exitCode *C.int

	cshell := C.CString(shell)
	cargs := C.CString(args)
	defer C.free(unsafe.Pointer(cshell))
	defer C.free(unsafe.Pointer(cargs))

	jobHandle = C.VixVM_RunProgramInGuest(g.handle,
		cshell, //guestProgramName
		cargs,  //commandLineArgs
		C.VixRunProgramOptions(options), //options
		C.VIX_INVALID_HANDLE,            //propertyListHandle
		nil,                             // callbackProc
		nil)                             // clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.get_program_output(jobHandle, pid, elapsedtime, exitCode)

	if C.VIX_OK != err {
		return 0, 0, 0, &VixError{
			Operation: "guest.RunScript.get_program_output",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return uint64(*pid), int(*elapsedtime), int(*exitCode), nil
}

// This function renames a file or directory in the guest operating system.
//
// Parameters:
//
//   path1: The path to the file to be renamed.
//   path2: The path to the new file.
//
// Remarks:
//
//   * Only absolute paths should be used for files in the guest; the resolution
//     of relative paths is not specified.
//
//   * On Windows guests, it fails on directory moves when the destination is on a
//     different volume.
//
//   * Because of the differences in how various operating systems handle
//     filenames, Vix may return either VIX_E_INVALID_ARG or
//     VIX_E_FILE_NAME_TOO_LONG for filenames longer than 255 characters.
//
// Since VMware Workstation 6.0
// Minimum Supported Guest OS: Microsoft Windows NT Series, Linux
func (g *Guest) Mv(path1, path2 string) error {
	var jobHandle C.VixHandle = C.VIX_INVALID_HANDLE
	var err C.VixError = C.VIX_OK

	cpath1 := C.CString(path1)
	cpath2 := C.CString(path2)
	defer C.free(unsafe.Pointer(cpath1))
	defer C.free(unsafe.Pointer(cpath2))

	jobHandle = C.VixVM_RenameFileInGuest(g.handle,
		cpath1,               //oldName
		cpath2,               //newName
		0,                    //options
		C.VIX_INVALID_HANDLE, //propertyListHandle
		nil,                  //callbackProc
		nil)                  //clientData

	defer C.Vix_ReleaseHandle(jobHandle)

	err = C.vix_job_wait(jobHandle)
	if C.VIX_OK != err {
		return &VixError{
			Operation: "guest.Mv",
			Code:      int(err & 0xFFFF),
			Text:      C.GoString(C.Vix_GetErrorText(err, nil)),
		}
	}

	return nil
}
