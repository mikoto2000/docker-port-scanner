APP_NAME := docker-port-scanner
BUILD_DIR := ./build
MAIN_PKG := ./cmd/$(APP_NAME)
GOCACHE := $(abspath $(BUILD_DIR)/.cache/go-build)
GOMODCACHE := $(abspath $(BUILD_DIR)/.cache/go-mod)

TARGETS := \
	linux-amd64 \
	linux-arm64 \
	darwin-amd64 \
	darwin-arm64 \
	windows-amd64 \
	windows-arm64

.PHONY: build clean $(TARGETS)

build: $(TARGETS)

linux-amd64:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PKG)

linux-arm64:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_PKG)

darwin-amd64:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PKG)

darwin-arm64:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PKG)

windows-amd64:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PKG)

windows-arm64:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=arm64 GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe $(MAIN_PKG)

clean:
	rm -rf $(BUILD_DIR)
