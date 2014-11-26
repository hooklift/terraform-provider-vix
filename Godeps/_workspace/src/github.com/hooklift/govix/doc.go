/*
Govix is a set of Go bindings for VMware VIX.

The VIX API allows you to automate virtual machine operations on most
current VMware platform products, especially hosted VMware products such as:
vmware workstation, player, fusion and server.

vSphere API, starting from 5.0, merged VIX API in the so-called
GuestOperationsManager managed object. So, we encourage you to use
GovSphere for vSphere instead.


Features

This API supports:

  * Adding, removing and listing virtual networks adapters attached to a VM
  * Adding and removing virtual CPUs and memory from a VM
  * Managing virtual switches
  * Managing virtual machines life cycle: power on, power off, reset, pause and resume.
  * Adding and removing shared folders
  * Taking screenshots from a running VM
  * Cloning VMs
  * Creating and removing Snapshots as well as restoring a VM from a Snapshot
  * Upgrading virtual hardware
  * Guest management: login, logout, install vmware tools, etc.

Dynamic library loading

In order for Go to find libvix when running your compiled binary, a govix
path has to be added to the LD_LIBRARY_PATH environment variable.

Example:

  * OSX: export DYLD_LIBRARY_PATH=${GOPATH}/src/github.com/cloudescape/govix
  * Linux: export LD_LIBRARY_PATH=${GOPATH}/src/github.com/cloudescape/govix
  * Windows: append the path to the PATH environment variable

Be aware that the previous example assumes $GOPATH only has a path set.


Debugging

In order to enable VIX debugging in VMware, you have set the following setting:

  * OSX: echo "vix.debugLevel = \"9\"" >> ~/Library/Preferences/VMware\ Fusion/config
  * Linux: ?
  * Windows: ?


Logs

For logging, the following are the paths in each operating system:

  * OSX: `~/Library/Logs/VMware/*.log`
  * Linux: `/tmp/vmware-<username>/vix-<pid>.log`
  * Windows: `%TEMP%\vmware-<username>\vix-<pid>.log`


Multithreading

The Vix library is intended for use by multi-threaded clients. Vix shared
objects are managed by the Vix library to avoid conflicts between threads.
Clients need only be responsible for protecting user-defined shared data.


VMware VIX EULA

As noted in the End User License Agreement, the VIX API allows you to build and distribute your own applications. To facilitate this, the following files are designated as redistributable for the purpose of that agreement:

  * VixAllProducts.lib
  * VixAllProductsd.lib
  * VixAllProductsDyn.lib
  * vix.lib and vix.dll
  * vixCOM.dll
  * gvmomi-vix-1.13.1.dll
  * libvixAllProducts.so
  * libvix.so
  * libgvmomi-vix-1.13.1.so.0
  * vixwrapper-config.txt
  * manifest.txt
  * compiled perl modules resulting from building the contents of vix-perl.tar.gz or vix-perl.zip

Redistribution of the open source libraries included with the VIX API is governed by their respective open source license agreements.

http://blogs.vmware.com/vix/2010/05/redistibutable-vix-api-client-libraries.html
*/
package vix
