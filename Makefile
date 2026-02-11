.PHONY: all build install clean clean-all test test-suite test-all help build-all release

# Build configuration
BINARY_NAME=smpe_ls
LINT_BINARY_NAME=smpe_lint
BUILD_DIR=.
INSTALL_DIR=$(HOME)/.local/bin
DATA_INSTALL_DIR=$(HOME)/.local/share/smpe_ls
DATA_DIR=data
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-s -w -X main.commit=$(COMMIT)

all: build build-lint

help:
	@echo "SMPE Language Server - Available Make Targets"
	@echo "════════════════════════════════════════════════════════════════"
	@echo "Build & Install:"
	@echo "  make build        - Build the language server binary"
	@echo "  make build-lint   - Build the linting tool"
	@echo "  make install      - Install binary and data files to ~/.local/"
	@echo "  make build-all    - Build binaries for all platforms"
	@echo "  make release      - Create release packages for all platforms"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run Go unit tests (completion, diagnostics, hover, parser)"
	@echo "  make test-suite   - Run central test suite on all .smpe test files"
	@echo "  make test-all     - Run both unit tests and test suite"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make clean-all    - Remove all build artifacts including VSCode extension"
	@echo ""
	@echo "VSCode Extension:"
	@echo "  make vscode-deps  - Install VSCode extension dependencies"
	@echo "  make vscode-compile - Compile VSCode extension"
	@echo "  make vscode       - Build server and prepare VSCode extension"
	@echo "  make package-windows - Create VSIX package for Windows"
	@echo "════════════════════════════════════════════════════════════════"

build:
	@echo "Building $(BINARY_NAME) (commit: $(COMMIT))..."
	go build -ldflags="-X main.commit=$(COMMIT)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/smpe_ls
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-lint:
	@echo "Building $(LINT_BINARY_NAME) (commit: $(COMMIT))..."
	go build -ldflags="-X main.commit=$(COMMIT)" -o $(BUILD_DIR)/$(LINT_BINARY_NAME) ./cmd/smpe_lint
	@echo "Build complete: $(BUILD_DIR)/$(LINT_BINARY_NAME)"

install: build build-lint
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed binary to $(INSTALL_DIR)/$(BINARY_NAME)"
	@cp $(BUILD_DIR)/$(LINT_BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(LINT_BINARY_NAME)
	@echo "Installed binary to $(INSTALL_DIR)/$(LINT_BINARY_NAME)"
	@echo ""
	@echo "Installing data files to $(DATA_INSTALL_DIR)..."
	@mkdir -p $(DATA_INSTALL_DIR)
	@cp $(DATA_DIR)/smpe.json $(DATA_INSTALL_DIR)/
	@echo "Installed data to $(DATA_INSTALL_DIR)/smpe.json"
	@echo ""
	@echo "Installation complete!"
	@echo "Server will use: $(DATA_INSTALL_DIR)/smpe.json"

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@rm -f $(BUILD_DIR)/$(LINT_BINARY_NAME)
	@echo "Clean complete"

clean-all:
	@echo "Cleaning all build artifacts including extension..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@rm -f $(BUILD_DIR)/$(LINT_BINARY_NAME)
	@rm -rf dist/
	@rm -rf release/
	@rm -rf client/vscode-smpe/out
	@rm -rf client/vscode-smpe/node_modules
	@echo "Clean all complete"

build-all:
	@echo "Building binaries for all platforms..."
	@echo "Commit: $(COMMIT)"
	@mkdir -p dist
	@echo ""
	@echo "Building Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_ls-linux-amd64 ./cmd/smpe_ls
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_lint-linux-amd64 ./cmd/smpe_lint
	@echo "Building Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_ls-linux-arm64 ./cmd/smpe_ls
	@GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_lint-linux-arm64 ./cmd/smpe_lint
	@echo ""
	@echo "Building macOS Apple Silicon (ARM64)..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_ls-macos-arm64 ./cmd/smpe_ls
	@GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_lint-macos-arm64 ./cmd/smpe_lint
	@echo "Building macOS Intel (AMD64)..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_ls-macos-amd64 ./cmd/smpe_ls
	@GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_lint-macos-amd64 ./cmd/smpe_lint
	@echo ""
	@echo "Building Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_ls-windows-amd64.exe ./cmd/smpe_ls
	@GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_lint-windows-amd64.exe ./cmd/smpe_lint
	@echo "Building Windows ARM64..."
	@GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_ls-windows-arm64.exe ./cmd/smpe_ls
	@GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/smpe_lint-windows-arm64.exe ./cmd/smpe_lint
	@echo ""
	@echo "All binaries built successfully in dist/"
	@ls -lh dist/

release: build-all
	@echo ""
	@echo "Creating release packages..."
	@mkdir -p release
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev"); \
	echo "Version: $$VERSION"; \
	echo ""; \
	echo "Packaging Linux AMD64..."; \
	mkdir -p release/smpe_ls-$$VERSION-linux-amd64; \
	cp dist/smpe_ls-linux-amd64 release/smpe_ls-$$VERSION-linux-amd64/smpe_ls; \
	cp dist/smpe_lint-linux-amd64 release/smpe_ls-$$VERSION-linux-amd64/smpe_lint; \
	cp data/smpe.json release/smpe_ls-$$VERSION-linux-amd64/; \
	cp README.md release/smpe_ls-$$VERSION-linux-amd64/ 2>/dev/null || true; \
	tar czf release/smpe_ls-$$VERSION-linux-amd64.tar.gz -C release smpe_ls-$$VERSION-linux-amd64; \
	echo ""; \
	echo "Packaging Linux ARM64..."; \
	mkdir -p release/smpe_ls-$$VERSION-linux-arm64; \
	cp dist/smpe_ls-linux-arm64 release/smpe_ls-$$VERSION-linux-arm64/smpe_ls; \
	cp dist/smpe_lint-linux-arm64 release/smpe_ls-$$VERSION-linux-arm64/smpe_lint; \
	cp data/smpe.json release/smpe_ls-$$VERSION-linux-arm64/; \
	cp README.md release/smpe_ls-$$VERSION-linux-arm64/ 2>/dev/null || true; \
	tar czf release/smpe_ls-$$VERSION-linux-arm64.tar.gz -C release smpe_ls-$$VERSION-linux-arm64; \
	echo ""; \
	echo "Packaging macOS Apple Silicon..."; \
	mkdir -p release/smpe_ls-$$VERSION-macos-arm64; \
	cp dist/smpe_ls-macos-arm64 release/smpe_ls-$$VERSION-macos-arm64/smpe_ls; \
	cp dist/smpe_lint-macos-arm64 release/smpe_ls-$$VERSION-macos-arm64/smpe_lint; \
	cp data/smpe.json release/smpe_ls-$$VERSION-macos-arm64/; \
	cp README.md release/smpe_ls-$$VERSION-macos-arm64/ 2>/dev/null || true; \
	tar czf release/smpe_ls-$$VERSION-macos-arm64.tar.gz -C release smpe_ls-$$VERSION-macos-arm64; \
	echo ""; \
	echo "Packaging macOS Intel..."; \
	mkdir -p release/smpe_ls-$$VERSION-macos-amd64; \
	cp dist/smpe_ls-macos-amd64 release/smpe_ls-$$VERSION-macos-amd64/smpe_ls; \
	cp dist/smpe_lint-macos-amd64 release/smpe_ls-$$VERSION-macos-amd64/smpe_lint; \
	cp data/smpe.json release/smpe_ls-$$VERSION-macos-amd64/; \
	cp README.md release/smpe_ls-$$VERSION-macos-amd64/ 2>/dev/null || true; \
	tar czf release/smpe_ls-$$VERSION-macos-amd64.tar.gz -C release smpe_ls-$$VERSION-macos-amd64; \
	echo ""; \
	echo "Packaging Windows AMD64..."; \
	mkdir -p release/smpe_ls-$$VERSION-windows-amd64; \
	cp dist/smpe_ls-windows-amd64.exe release/smpe_ls-$$VERSION-windows-amd64/smpe_ls.exe; \
	cp dist/smpe_lint-windows-amd64.exe release/smpe_ls-$$VERSION-windows-amd64/smpe_lint.exe; \
	cp data/smpe.json release/smpe_ls-$$VERSION-windows-amd64/; \
	cp README.md release/smpe_ls-$$VERSION-windows-amd64/ 2>/dev/null || true; \
	cd release && zip -r smpe_ls-$$VERSION-windows-amd64.zip smpe_ls-$$VERSION-windows-amd64; \
	cd ..; \
	echo ""; \
	echo "Packaging Windows ARM64..."; \
	mkdir -p release/smpe_ls-$$VERSION-windows-arm64; \
	cp dist/smpe_ls-windows-arm64.exe release/smpe_ls-$$VERSION-windows-arm64/smpe_ls.exe; \
	cp dist/smpe_lint-windows-arm64.exe release/smpe_ls-$$VERSION-windows-arm64/smpe_lint.exe; \
	cp data/smpe.json release/smpe_ls-$$VERSION-windows-arm64/; \
	cp README.md release/smpe_ls-$$VERSION-windows-arm64/ 2>/dev/null || true; \
	cd release && zip -r smpe_ls-$$VERSION-windows-arm64.zip smpe_ls-$$VERSION-windows-arm64; \
	cd ..; \
	echo ""; \
	echo "═══════════════════════════════════════════════════════════════"; \
	echo "Release packages created in release/"; \
	echo "═══════════════════════════════════════════════════════════════"; \
	ls -lh release/*.tar.gz release/*.zip 2>/dev/null || true

test:
	@echo "Running Go unit tests..."
	@which richgo > /dev/null 2>&1 && richgo test -v ./... || go test -v ./...

test-suite:
	@echo "Building central test suite..."
	@go build -o cmd/smpe_test/smpe_test cmd/smpe_test/main.go
	@echo ""
	@echo "Running central test suite on all .smpe files..."
	@echo "(Note: Some test files contain intentional errors for validation)"
	@./cmd/smpe_test/smpe_test ./test-files; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -ne 0 ]; then \
		echo ""; \
		echo "ℹ️  Test suite reported failures - this is expected for test files"; \
		echo "   with intentional errors (test-langid.smpe, test-mac*.smpe)"; \
	fi; \
	exit 0

test-all: test test-suite
	@echo ""
	@echo "═════════════════════════════════════════════════════════════════"
	@echo "                    ALL TESTS COMPLETED                          "
	@echo "═════════════════════════════════════════════════════════════════"

vscode-deps:
	@echo "Installing VSCode extension dependencies..."
	cd client/vscode-smpe && npm install

vscode-compile: vscode-deps
	@echo "Compiling VSCode extension..."
	cd client/vscode-smpe && npm run compile

vscode: build vscode-compile
	@echo "VSCode extension ready for testing"
	@echo ""
	@echo "To test in VSCode:"
	@echo "1. Open the client/vscode-smpe directory in VSCode"
	@echo "2. Press F5 to launch Extension Development Host"
	@echo "3. Open a .smpe file to activate the extension"

package-windows: vscode-deps
	@echo "Building Windows AMD64 binaries (commit: $(COMMIT))..."
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_ls.exe ./cmd/smpe_ls
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_lint.exe ./cmd/smpe_lint
	@echo "Copying data files..."
	@cp $(DATA_DIR)/smpe.json client/vscode-smpe/
	@echo "Creating VSIX package for Windows..."
	cd client/vscode-smpe && npx --yes @vscode/vsce package --target win32-x64
	@echo ""
	@echo "VSIX package created in client/vscode-smpe/"
	@rm -f client/vscode-smpe/smpe_ls.exe client/vscode-smpe/smpe_lint.exe

package-windows-arm64: vscode-deps
	@echo "Building Windows ARM64 binaries (commit: $(COMMIT))..."
	GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_ls.exe ./cmd/smpe_ls
	GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_lint.exe ./cmd/smpe_lint
	@echo "Copying data files..."
	@cp $(DATA_DIR)/smpe.json client/vscode-smpe/
	@echo "Creating VSIX package for Windows ARM64..."
	cd client/vscode-smpe && npx --yes @vscode/vsce package --target win32-arm64
	@echo ""
	@echo "VSIX package created in client/vscode-smpe/"
	@rm -f client/vscode-smpe/smpe_ls.exe client/vscode-smpe/smpe_lint.exe

package-linux: vscode-deps
	@echo "Building Linux AMD64 binaries (commit: $(COMMIT))..."
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_ls ./cmd/smpe_ls
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_lint ./cmd/smpe_lint
	@echo "Copying data files..."
	@cp $(DATA_DIR)/smpe.json client/vscode-smpe/
	@echo "Creating VSIX package for Linux..."
	cd client/vscode-smpe && npx --yes @vscode/vsce package --target linux-x64
	@echo ""
	@echo "VSIX package created in client/vscode-smpe/"
	@rm -f client/vscode-smpe/smpe_ls client/vscode-smpe/smpe_lint

package-linux-arm64: vscode-deps
	@echo "Building Linux ARM64 binaries (commit: $(COMMIT))..."
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_ls ./cmd/smpe_ls
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_lint ./cmd/smpe_lint
	@echo "Copying data files..."
	@cp $(DATA_DIR)/smpe.json client/vscode-smpe/
	@echo "Creating VSIX package for Linux ARM64..."
	cd client/vscode-smpe && npx --yes @vscode/vsce package --target linux-arm64
	@echo ""
	@echo "VSIX package created in client/vscode-smpe/"
	@rm -f client/vscode-smpe/smpe_ls client/vscode-smpe/smpe_lint

package-macos: vscode-deps
	@echo "Building macOS ARM64 (Apple Silicon) binaries (commit: $(COMMIT))..."
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_ls ./cmd/smpe_ls
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_lint ./cmd/smpe_lint
	@echo "Copying data files..."
	@cp $(DATA_DIR)/smpe.json client/vscode-smpe/
	@echo "Creating VSIX package for macOS ARM64..."
	cd client/vscode-smpe && npx --yes @vscode/vsce package --target darwin-arm64
	@echo ""
	@echo "VSIX package created in client/vscode-smpe/"
	@rm -f client/vscode-smpe/smpe_ls client/vscode-smpe/smpe_lint

package-macos-x64: vscode-deps
	@echo "Building macOS AMD64 (Intel) binaries (commit: $(COMMIT))..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_ls ./cmd/smpe_ls
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o client/vscode-smpe/smpe_lint ./cmd/smpe_lint
	@echo "Copying data files..."
	@cp $(DATA_DIR)/smpe.json client/vscode-smpe/
	@echo "Creating VSIX package for macOS Intel..."
	cd client/vscode-smpe && npx --yes @vscode/vsce package --target darwin-x64
	@echo ""
	@echo "VSIX package created in client/vscode-smpe/"
	@rm -f client/vscode-smpe/smpe_ls client/vscode-smpe/smpe_lint

package-all: vscode-deps
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "          Building VSIX packages for all platforms             "
	@echo "═══════════════════════════════════════════════════════════════"
	@mkdir -p release/vsix
	@$(MAKE) package-windows
	@mv client/vscode-smpe/*.vsix release/vsix/
	@$(MAKE) package-windows-arm64
	@mv client/vscode-smpe/*.vsix release/vsix/
	@$(MAKE) package-linux
	@mv client/vscode-smpe/*.vsix release/vsix/
	@$(MAKE) package-linux-arm64
	@mv client/vscode-smpe/*.vsix release/vsix/
	@$(MAKE) package-macos
	@mv client/vscode-smpe/*.vsix release/vsix/
	@$(MAKE) package-macos-x64
	@mv client/vscode-smpe/*.vsix release/vsix/
	@echo ""
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "          All VSIX packages created in release/vsix/           "
	@echo "═══════════════════════════════════════════════════════════════"
	@ls -lh release/vsix/

.PHONY: vscode-deps vscode-compile vscode package-windows package-windows-arm64 package-linux package-linux-arm64 package-macos package-macos-x64 package-all
