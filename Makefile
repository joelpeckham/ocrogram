BIN_DIR := bin
GO_BIN := $(BIN_DIR)/ocrogram
HELPER_BIN := $(BIN_DIR)/ocrogram-helper
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# Native arch unless GOARCH is set (release CI cross-compiles amd64 from arm64).
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
NATIVE_GOARCH := amd64
else
NATIVE_GOARCH := arm64
endif

GOARCH ?= $(NATIVE_GOARCH)
ifeq ($(GOARCH),amd64)
SWIFT_ARCH := x86_64
else
SWIFT_ARCH := arm64
endif

SWIFT_TARGET := $(SWIFT_ARCH)-apple-macos14
ARCH_STAMP := $(BIN_DIR)/.arch-$(GOARCH)

GO_SOURCES := $(filter-out %_test.go,$(wildcard cmd/ocrogram/*.go) $(wildcard internal/*/*.go))
SWIFT_SOURCES := $(wildcard helper/Sources/ocrogram-helper/*.swift)

.PHONY: all go helper test clean

# Two independent binaries; -j2 overlaps Go and swiftc.
all:
	$(MAKE) -j2 $(GO_BIN) $(HELPER_BIN)

go: $(GO_BIN)
helper: $(HELPER_BIN)

$(ARCH_STAMP):
	mkdir -p $(BIN_DIR)
	rm -f $(BIN_DIR)/.arch-*
	touch $@

$(GO_BIN): $(GO_SOURCES) go.mod go.sum $(ARCH_STAMP)
	CGO_ENABLED=0 GOARCH=$(GOARCH) go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(GO_BIN) ./cmd/ocrogram

$(HELPER_BIN): $(SWIFT_SOURCES) $(ARCH_STAMP)
	swiftc -O -target $(SWIFT_TARGET) \
		-o $(HELPER_BIN) \
		$(SWIFT_SOURCES) \
		-framework Vision -framework AppKit

test: all
	go test -race ./...

clean:
	rm -rf $(BIN_DIR) helper/.build
