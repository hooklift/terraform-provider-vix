CGO_CFLAGS:=-I$(CURDIR)/vendor/libvix/include -Werror
CGO_LDFLAGS:=-L$(CURDIR)/vendor/libvix -lvixAllProducts -ldl -lpthread

export CGO_CFLAGS CGO_LDFLAGS

build:
	go build ./...

test:
	go test ./...

.PHONY: build test

