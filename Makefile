BINARY_NAME := tmux-ctrl
BUILD_DIR := ./build
INSTALL_DIR := $(HOME)/.local/bin
VERSION := $(shell git rev-parse --short=8 HEAD 2>/dev/null || echo "dev")$(shell git diff --quiet 2>/dev/null || echo "-dirty")

.PHONY: build install uninstall clean fmt test

fmt:
	go fmt ./...

check-fmt:
	test -z "$$(gofmt -l . | tee /dev/stderr)"

lint:
	go vet

build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "-X github.com/MunifTanjim/tmux-ctrl/internal/version.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .

install: build uninstall
	mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)

test:
	go test -v ./...
