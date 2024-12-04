# Makefile for building httpmon for multiple platforms

# Application name
APP_NAME := httpmon

# Source file
SRC := main.go

# Output directory
BUILD_DIR := build

# Build matrix
OS := linux darwin
ARCH := amd64 arm64

# Default target
.PHONY: all
all: clean build

# Build for all platforms
.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	@for os in $(OS); do \
		for arch in $(ARCH); do \
			echo "Building for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch go build -o $(BUILD_DIR)/$(APP_NAME)-$$os-$$arch $(SRC); \
		done; \
	done

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
