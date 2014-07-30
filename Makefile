CGO_CFLAGS:=-I$(CURDIR)/vendor/libvix/include
CGO_LDFLAGS:=-L$(CURDIR)/vendor/libvix

export CGO_CFLAGS CGO_LDFLAGS

build:
	go build

test:
	go test

.PHONY: build test

