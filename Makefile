APP := scaleconnect
PKG := .
BIN_DIR := bin
ADDON_DIR := picooc_scaleconnect
ADDON_BIN := $(ADDON_DIR)/scaleconnect

GO ?= go
GOFLAGS ?=
LDFLAGS ?= -s -w

ADDON_GOOS ?= linux
ADDON_GOARCH ?= $(shell $(GO) env GOARCH)

.PHONY: help tidy fmt test build build-linux-amd64 build-linux-arm64 addon-bin addon-bin-amd64 addon-bin-arm64 addon-image clean

help:
	@printf '%s\n' 'Targets:'
	@printf '  %-18s %s\n' 'tidy' 'Update go.mod/go.sum'
	@printf '  %-18s %s\n' 'fmt' 'Format Go sources'
	@printf '  %-18s %s\n' 'test' 'Run Go tests'
	@printf '  %-18s %s\n' 'build' 'Build local binary into bin/'
	@printf '  %-18s %s\n' 'build-linux-amd64' 'Build Linux amd64 binary into bin/'
	@printf '  %-18s %s\n' 'build-linux-arm64' 'Build Linux arm64 binary into bin/'
	@printf '  %-18s %s\n' 'addon-bin' 'Build static add-on binary for current GOARCH'
	@printf '  %-18s %s\n' 'addon-bin-amd64' 'Build static add-on binary for amd64'
	@printf '  %-18s %s\n' 'addon-bin-arm64' 'Build static add-on binary for arm64'
	@printf '  %-18s %s\n' 'addon-image' 'Build local HA add-on Docker image'
	@printf '  %-18s %s\n' 'clean' 'Remove Makefile build outputs'

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

build:
	mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/$(APP) $(PKG)

build-linux-amd64:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/$(APP)_linux_amd64 $(PKG)

build-linux-arm64:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/$(APP)_linux_arm64 $(PKG)

addon-bin:
	CGO_ENABLED=0 GOOS=$(ADDON_GOOS) GOARCH=$(ADDON_GOARCH) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -trimpath -o $(ADDON_BIN) $(PKG)

addon-bin-amd64:
	$(MAKE) addon-bin ADDON_GOARCH=amd64

addon-bin-arm64:
	$(MAKE) addon-bin ADDON_GOARCH=arm64

addon-image:
	docker build -t picooc-scaleconnect-addon $(ADDON_DIR)

clean:
	rm -f $(BIN_DIR)/$(APP) $(BIN_DIR)/$(APP)_linux_amd64 $(BIN_DIR)/$(APP)_linux_arm64
