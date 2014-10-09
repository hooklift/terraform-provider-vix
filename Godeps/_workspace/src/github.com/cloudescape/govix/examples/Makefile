CGO_CFLAGS:=-I$(CURDIR)/vendor/libvix/include
CGO_LDFLAGS:=-L$(CURDIR)/vendor/libvix -lvixAllProducts -ldl -lpthread

export CGO_CFLAGS CGO_LDFLAGS

bar:
	go build bar.go

foo:
	go build foo.go

.PHONY: bar foo
