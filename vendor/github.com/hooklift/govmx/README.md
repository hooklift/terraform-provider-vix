# GoVMX
[![Gitter](https://badges.gitter.im/Join Chat.svg)](https://gitter.im/hooklift/govmx?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![GoDoc](https://godoc.org/github.com/hooklift/govmx?status.svg)](https://godoc.org/github.com/hooklift/govmx)
[![Build Status](https://travis-ci.org/hooklift/govmx.svg?branch=master)](https://travis-ci.org/hooklift/govmx)

Data encoding and decoding library for VMware VMX files.


## Encoding

```go
type VM struct {
	Encoding     string `vmx:".encoding"`
	Annotation   string `vmx:"annotation"`
	Hwversion    uint8  `vmx:"virtualHW.version"`
	HwProdCompat string `vmx:"virtualHW.productCompatibility"`
	Memsize      uint   `vmx:"memsize"`
	Numvcpus     uint   `vmx:"numvcpus"`
	MemHotAdd    bool   `vmx:"mem.hotadd"`
	DisplayName  string `vmx:"displayName"`
	GuestOS      string `vmx:"guestOS"`
	Autoanswer   bool   `vmx:"msg.autoAnswer"`
}

vm := new(VM)
vm.Encoding = "utf-8"
vm.Annotation = "Test VM"
vm.Hwversion = 10
vm.HwProdCompat = "hosted"
vm.Memsize = 1024
vm.Numvcpus = 2
vm.MemHotAdd = false
vm.DisplayName = "test"
vm.GuestOS = "other3xlinux-64"
vm.Autoanswer = true

data, err := Marshal(vm)
```

`data` should be: 

```
.encoding = "utf-8"
annotation = "Test VM"
virtualHW.version = "10"
virtualHW.productCompatibility = "hosted"
memsize = "1024"
numvcpus = "2"
mem.hotadd = "false"
displayName = "test"
guestOS = "other3xlinux-64"
msg.autoAnswer = "true"
```

## Decoding

```go
var data = `.encoding = "utf-8"
annotation = "Test VM"
virtualHW.version = "10"
virtualHW.productCompatibility = "hosted"
memsize = "1024"
numvcpus = "2"
mem.hotadd = "false"
displayName = "test"
guestOS = "other3xlinux-64"
msg.autoAnswer = "true"
`

type VM struct {
	Encoding     string `vmx:".encoding"`
	Annotation   string `vmx:"annotation"`
	Hwversion    uint8  `vmx:"virtualHW.version"`
	HwProdCompat string `vmx:"virtualHW.productCompatibility"`
	Memsize      uint   `vmx:"memsize"`
	Numvcpus     uint   `vmx:"numvcpus"`
	MemHotAdd    bool   `vmx:"mem.hotadd"`
	DisplayName  string `vmx:"displayName"`
	GuestOS      string `vmx:"guestOS"`
	Autoanswer   bool   `vmx:"msg.autoAnswer"`
}

vm := new(VM)
err := Unmarshal([]byte(data), vm)
```
