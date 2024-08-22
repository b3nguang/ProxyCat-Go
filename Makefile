# Define variables
BINARY_NAME = ProxyCat-Go
MAIN_FILE = main.go

# Default build target
default: build

# Build for the current OS and architecture
build:
	go build -o $(BINARY_NAME) $(MAIN_FILE)

# Build for Linux amd64
build-linux:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o $(BINARY_NAME)-linux-amd64 $(MAIN_FILE)

# Build for Linux ARM64
build-linux-arm:
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o $(BINARY_NAME)-linux-arm64 $(MAIN_FILE)

# Build for macOS amd64
build-macos:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -o $(BINARY_NAME)-macos-amd64 $(MAIN_FILE)

# Build for macOS ARM64 (Apple Silicon)
build-macos-arm:
	CGO_ENABLED=0 GOARCH=arm64 GOOS=darwin go build -o $(BINARY_NAME)-macos-arm64 $(MAIN_FILE)

# Build for Windows amd64
build-windows:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -o $(BINARY_NAME).exe $(MAIN_FILE)

# Build for Windows ARM64
build-windows-arm:
	CGO_ENABLED=0 GOARCH=arm64 GOOS=windows go build -o $(BINARY_NAME)-windows-arm64.exe $(MAIN_FILE)

# Clean up build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-linux-amd64 $(BINARY_NAME)-linux-arm64 $(BINARY_NAME)-macos-amd64 $(BINARY_NAME)-macos-arm64 $(BINARY_NAME).exe $(BINARY_NAME)-windows-arm64.exe

# Show help
help:
	@echo "Usage:"
	@echo "  make build              Build for the current OS and architecture"
	@echo "  make build-linux        Build for Linux amd64"
	@echo "  make build-linux-arm    Build for Linux ARM64"
	@echo "  make build-macos        Build for macOS amd64"
	@echo "  make build-macos-arm    Build for macOS ARM64"
	@echo "  make build-windows      Build for Windows amd64"
	@echo "  make build-windows-arm  Build for Windows ARM64"
	@echo "  make clean              Remove build artifacts"
	@echo "  make help               Display this help message"
