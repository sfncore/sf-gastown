.PHONY: build install clean test generate

BINARY := gt
BUILD_DIR := .
INSTALL_DIR := $(HOME)/.local/bin

# Get version info for ldflags
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X github.com/sfncore/sf-gastown/internal/cmd.Version=$(VERSION) \
           -X github.com/sfncore/sf-gastown/internal/cmd.Commit=$(COMMIT) \
           -X github.com/sfncore/sf-gastown/internal/cmd.BuildTime=$(BUILD_TIME) \
           -X github.com/sfncore/sf-gastown/internal/cmd.BuiltProperly=1

generate:
	go generate ./...

build: generate
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/gt
ifeq ($(shell uname),Darwin)
	@codesign -s - -f $(BUILD_DIR)/$(BINARY) 2>/dev/null || true
	@echo "Signed $(BINARY) for macOS"
endif

install: build
	@mkdir -p $(INSTALL_DIR)
	@rm -f $(INSTALL_DIR)/$(BINARY)
	@cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed $(BINARY) to $(INSTALL_DIR)/$(BINARY)"
	@# Auto-configure gastown-src for 'gt stabilize'
	@# Only configure if GT_ROOT is set (indicates a Gas Town workspace exists)
	@if [ -n "$$GT_ROOT" ] && [ -d "$$GT_ROOT" ]; then \
		cd "$$GT_ROOT" && $(INSTALL_DIR)/$(BINARY) config gastown-src "$(CURDIR)" 2>/dev/null && \
		echo "Configured gastown-src=$(CURDIR)"; \
	elif [ -d "$(HOME)/gt/mayor" ]; then \
		cd "$(HOME)/gt" && $(INSTALL_DIR)/$(BINARY) config gastown-src "$(CURDIR)" 2>/dev/null && \
		echo "Configured gastown-src=$(CURDIR)"; \
	else \
		echo "Note: Run 'gt config gastown-src $(CURDIR)' from your Gas Town workspace to enable 'gt stabilize'"; \
	fi

clean:
	rm -f $(BUILD_DIR)/$(BINARY)

test:
	go test ./...
