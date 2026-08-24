BIN_DIR := bin
GO_BIN := $(BIN_DIR)/ocrogram
HELPER_BIN := $(BIN_DIR)/ocrogram-helper
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: all go helper test clean

all: go helper

go:
	mkdir -p $(BIN_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(GO_BIN) ./cmd/ocrogram

helper:
	mkdir -p $(BIN_DIR)
	swift build --package-path helper -c release
	cp helper/.build/release/ocrogram-helper $(HELPER_BIN)

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR) helper/.build
