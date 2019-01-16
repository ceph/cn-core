.PHONY: build tests

COMMIT = $(shell git describe --always --long --dirty)
TARGET_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
VERSION ?= $(TARGET_BRANCH)-$(COMMIT)

# Variables to choose cross-compile target
GOOS:=linux
GOARCH:=amd64
CN_CORE_EXTENSION:=

build: check clean prepare
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -i -ldflags="-X main.version=$(VERSION)" -o cn-core-$(VERSION)-$(GOOS)-$(GOARCH)$(CN_CORE_EXTENSION) main.go
	ln -sf "cn-core-$(VERSION)-$(GOOS)-$(GOARCH)$(CN_CORE_EXTENSION)" cn-core$(CN_CORE_EXTENSION)

check:
ifeq ("$(GOPATH)","")
	@echo "GOPATH variable must be defined"
	@exit 1
endif
ifneq ("$(shell pwd)","$(GOPATH)/src/github.com/ceph/cn-core")
	@echo "You are in $(shell pwd) !"
	@echo "Please go in $(GOPATH)/src/github.com/ceph/cn-core to build"
	@exit 1
endif

prepare:
	dep ensure
	cd cmd; GOOS=linux go test -timeout 1m

linux-%:
	make GOOS=linux GOARCH:=$*

release: linux-amd64

clean:
	rm -f cn-core$(CN_CORE_EXTENSION) cn-core &>/dev/null || true

clean-all: clean
	rm -f cn-* &>/dev/null || true
