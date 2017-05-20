// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
#ifndef helpers_h
#define helpers_h 1
#include <stdlib.h>

VixError vix_job_wait(VixHandle jobHandle);

VixError get_vix_handle(
	VixHandle jobHandle,
	VixPropertyID prop1,
	VixHandle* handle,
	VixPropertyID prop2);

VixError alloc_vm_pwd_proplist(
	VixHandle handle,
	VixHandle* resultHandle,
	char* password);

VixError get_screenshot_bytes(
	VixHandle handle,
	int* byte_count,
	char* screen_bits);

VixError get_program_output(
	VixHandle jobHandle,
	uint64* pid,
	int* elapsedTime,
	int* exitCode);

VixError get_shared_folder(
	VixHandle jobHandle,
	char* folderName,
	char* folderHostPath,
	int* folderFlags);

VixError get_file_info(VixHandle jobHandle,
					 int64* fsize,
					 int* flags,
					 int64* modtime);

VixError get_guest_file(VixHandle jobHandle,
						int i,
						char* name,
						int64* size,
						int64* modtime,
						int* flags);

VixError get_guest_process(VixHandle jobHandle,
						int i,
						char* name,
						uint64* pid,
						char* owner,
						char* cmdline,
						Bool* is_debugged,
						int* start_time);

void find_items_callback(
	VixHandle jobHandle,
	VixEventType eventType,
	VixHandle moreEventInfo,
	void* goCallback);

VixError get_num_shared_folders(VixHandle jobHandle, int* numSharedFolders);
VixError read_variable(VixHandle jobHandle, char** readValue);
VixError get_temp_filepath(VixHandle jobHandle, char* tempFilePath);
VixError is_file_or_dir(VixHandle jobHandle, int* result);
VixError get_vm_url(char* url, VixHandle moreEvtInfo);
VixError get_property(VixHandle handle, VixPropertyID id, void* value);

#endif
