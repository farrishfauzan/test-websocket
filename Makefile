.PHONY: build run test clean install dev

# Build the server
build:
	go build -o server cmd/server/main.go

# Run the server
run: build
	./server

# Run the server in development mode
dev:
	go run cmd/server/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f server

# Install dependencies
install:
	go mod tidy
	go mod download

# Build for different platforms
build-linux:
	GOOS=linux GOARCH=amd64 go build -o server-linux cmd/server/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o server-windows.exe cmd/server/main.go

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o server-mac cmd/server/main.go

# Build all platforms
build-all: build-linux build-windows build-mac

# Run the example client
client:
	go run examples/client.go

# Docker commands
docker-build:
	docker build -t websocket-chat .

docker-run:
	docker run -p 8080:8080 websocket-chat

# Help
help:
	@echo "Available commands:"
	@echo "  build       - Build the server binary"
	@echo "  run         - Build and run the server"
	@echo "  dev         - Run server in development mode"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  install     - Install dependencies"
	@echo "  client      - Run example client"
	@echo "  help        - Show this help message"