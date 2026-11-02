.PHONY: all build run clean install test build-windows build-linux build-mac build-all help

# Build variables
BINARY_NAME=word-killer
BUILD_DIR=build
MAIN_PATH=./cmd/word-killer

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
# BUILD_TIME is optional; when empty the about page shows the executable's mtime,
# which is the build time anyway. Pass one explicitly when needed:
#   make build BUILD_TIME=2026-11-02T17:20
BUILD_TIME ?=
VERSION_PKG=github.com/word-killer/word-killer/pkg/ui
LDFLAGS=-ldflags "-s -w $(if $(BUILD_TIME),-X $(VERSION_PKG).BuildTime=$(BUILD_TIME),)"

# OS-specific commands
ifeq ($(OS),Windows_NT)
    SHELL = cmd.exe
    .SHELLFLAGS = /c
    RM_BUILD   = if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
    MKDIR_BUILD = if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
    BINARY_EXT = .exe
    RUN_PREFIX =
    COPY_ASSETS = xcopy /E /I /Y /Q data $(BUILD_DIR)\data >nul && copy /Y config.json $(BUILD_DIR) >nul
    # Usage: $(call SETENV,GOOS,linux) $(call SETENV,GOARCH,amd64) go build ...
    SETENV = set $(1)=$(2)&&
else
    RM_BUILD   = rm -rf $(BUILD_DIR)
    MKDIR_BUILD = mkdir -p $(BUILD_DIR)
    BINARY_EXT =
    RUN_PREFIX = ./
    COPY_ASSETS = cp -r data $(BUILD_DIR)/ && cp config.json $(BUILD_DIR)/
    SETENV = $(1)=$(2)
endif

all: build

build:
	@echo Building $(BINARY_NAME)...
	@$(MKDIR_BUILD)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(MAIN_PATH)
	@$(COPY_ASSETS)
	@echo Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)

# Run inside build/ so config.json and data/ resolve via relative paths
run: build
	@echo Running $(BINARY_NAME)...
	cd $(BUILD_DIR) && $(RUN_PREFIX)$(BINARY_NAME)$(BINARY_EXT)

clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@$(RM_BUILD)
	@echo "Clean complete"

install:
	@echo "Installing dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "Dependencies installed"

test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

# Platform-specific builds
build-windows:
	@echo "Building for Windows..."
	@$(MKDIR_BUILD)
	$(call SETENV,GOOS,windows) $(call SETENV,GOARCH,amd64) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)

build-linux:
	@echo "Building for Linux..."
	@$(MKDIR_BUILD)
	$(call SETENV,GOOS,linux) $(call SETENV,GOARCH,amd64) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PATH)

build-mac:
	@echo "Building for macOS..."
	@$(MKDIR_BUILD)
	$(call SETENV,GOOS,darwin) $(call SETENV,GOARCH,amd64) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-mac $(MAIN_PATH)

build-all: build-windows build-linux build-mac
	@echo "All platform builds complete"

help:
	@echo "Available targets:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Build and run the application"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make install      - Install dependencies"
	@echo "  make test         - Run tests"
	@echo "  make build-windows- Build for Windows"
	@echo "  make build-linux  - Build for Linux"
	@echo "  make build-mac    - Build for macOS"
	@echo "  make build-all    - Build for all platforms"
