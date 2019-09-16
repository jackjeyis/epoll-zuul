RELEASE_VERSION = $(release_version)

ifeq ("$(RELEASE_VERSION)", "")
	RELEASE_VERSION := "unknow"
endif

ROOT_DIR = $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION_PATH = main
LD_GIT_COMMIT      = -X '$(VERSION_PATH).GitCommit=`git rev-parse --short HEAD`'
LD_BUILD_TIME      = -X '$(VERSION_PATH).BuildTime=`date +%FT%T%z`'
LD_GO_VERSION      = -X '$(VERSION_PATH).GoVersion=`go version`'
LD_GATEWAY_VERSION = -X '$(VERSION_PATH).Version=$(RELEASE_VERSION)'
LD_FLAGS           = -ldflags "$(LD_GIT_COMMIT) $(LD_BUILD_TIME) $(LD_GO_VERSION) $(LD_GATEWAY_VERSION) -w -s"

GOOS = linux
CGO_ENABLED = 0
DIST_DIR = $(ROOT_DIR)/dist

.PHONY: echo_str
echo_str: ; $(info ======== echo string:)
	@echo ======== version path : $(VERSION_PATH) ========

.PHONY: release
release: dist_dir proxy;

.PHONY: release_darwin
release_darwin: darwin dist_dir proxy;

.PHONY: darwin
darwin:
	$(eval GOOS := darwin)

.PHONY: proxy
proxy: ; $(info ======== compiled proxy:)
	env CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -a -installsuffix cgo -o $(DIST_DIR)/proxy	$(LD_FLAGS) $(ROOT_DIR)/cmd/proxy/*.go 

.PHONY: dist_dir
dist_dir: ;	$(info ======== prepare dist dir:)
	mkdir -p $(DIST_DIR)
	@rm -rf $(DIST_DIR)/*

.PHONY: clean
clean: ; $(info ======== clean all:)
	rm -rf $(DIST_DIR)

.PHONY: help
help:
	@echo "build release binary: \n\t\tmake release\n"
	@echo "build Mac OS X release binary: \n\t\tmake release_darwin\n"
	@echo "build docker release with etcd: \n\t\tmake docker\n"
	@echo "\t  add 「with」 params can select what you need:\n"
	@echo "\t  proxy: only compile proxy\n"
	@echo "clean all binary: \n\t\tmake clean\n"

UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Darwin)
	.DEFAULT_GOAL := release_darwin
else
	.DEFAULT_GOAL := release
endif
