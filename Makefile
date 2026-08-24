BIN_DIR := bin
GO_BIN := $(BIN_DIR)/ocrogram
HELPER_BIN := $(BIN_DIR)/ocrogram-helper

.PHONY: all go helper clean

all: go helper

go:
	mkdir -p $(BIN_DIR)
	go build -o $(GO_BIN) ./cmd/ocrogram

helper:
	mkdir -p $(BIN_DIR)
	swift build --package-path helper -c release
	cp helper/.build/release/ocrogram-helper $(HELPER_BIN)

clean:
	rm -rf $(BIN_DIR) helper/.build
