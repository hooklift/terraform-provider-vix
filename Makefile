#CGO_CFLAGS:=-I$(CURDIR)/vendor/libvix/include -Werror -l$(CURDIR)/vendor/libvix/libvix.dylib
CGO_CFLAGS:=-I$(CURDIR)/vendor/libvix/include -Werror
CGO_LDFLAGS:=-L$(CURDIR)/vendor/libvix -lvixAllProducts -ldl -lpthread

DYLD_LIBRARY_PATH:=$(CURDIR)/vendor/libvix
LD_LIBRARY_PATH:=$(CURDIR)/vendor/libvix

export CGO_CFLAGS CGO_LDFLAGS DYLD_LIBRARY_PATH LD_LIBRARY_PATH

install:
	go install -v

build:
	go build

test:
	go test ./...

clean:
	go clean ./...

.PHONY: build test install clean

