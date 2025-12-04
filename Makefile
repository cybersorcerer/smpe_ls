.PHONY: all build install clean clean-all test

# Build configuration
BINARY_NAME=smpe_ls
BUILD_DIR=.
INSTALL_DIR=$(HOME)/.local/bin
DATA_INSTALL_DIR=$(HOME)/.local/share/smpe_ls
DATA_DIR=data

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/smpe_ls
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed binary to $(INSTALL_DIR)/$(BINARY_NAME)"
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
	@echo "Clean complete"

clean-all:
	@echo "Cleaning all build artifacts including extension..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@rm -rf client/vscode-smpe/out
	@rm -rf client/vscode-smpe/node_modules
	@echo "Clean all complete"

test:
	@echo "Running tests..."
	@which richgo > /dev/null 2>&1 && richgo test -v ./... || go test -v ./...

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
	@echo "Building Windows binary..."
	GOOS=windows GOARCH=amd64 go build -o client/vscode-smpe/smpe_ls.exe ./cmd/smpe_ls
	@echo "Copying data files..."
	@cp $(DATA_DIR)/smpe.json client/vscode-smpe/
	@echo "Creating VSIX package..."
	cd client/vscode-smpe && npx --yes @vscode/vsce package
	@echo ""
	@echo "VSIX package created in client/vscode-smpe/"
	@echo "Install on Windows with: code --install-extension vscode-smpe-*.vsix"

.PHONY: vscode-deps vscode-compile vscode package-windows
