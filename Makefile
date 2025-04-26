# Default port if not specified in environment
GRPC_PORT ?= 6001

# Paths to the main files 
SERVICE = ./cmd/timeline_service/main.go

# Default build output path
OUTPUT_DIR = bin



#proto compilation
pcompile:
	 protoc --go_out=. --go_opt=paths=source_relative     --go-grpc_out=. --go-grpc_opt=paths=source_relative     proto/timeline.proto


# run graphql server
gql:
	go run ./server.go
# Run the service in the background
run:
	@echo "Starting Service ..."
	go run $(SERVICE) > posts.log 2>&1 &

# Run the service in the foreground
run-foreground:
	@echo "Running Service in foreground..."
	go run $(SERVICE)

# Stop all running Go services
stop:
	@echo "Killing service on port ${GRPC_PORT}..."
	@lsof -ti :${GRPC_PORT} | xargs kill -9 || echo "No service running on port ${GRPC_PORT}"

# Build binaries for the current OS
build:
	@echo "Building binary for current OS..."
	go build -o $(OUTPUT_DIR)/service $(SERVICE)

# Build binaries for specific OS and architecture
build-cross:
	@echo "Building binaries for different platforms..."
	GOOS=linux   GOARCH=amd64 go build -o $(OUTPUT_DIR)/service-linux-amd64 $(SERVICE)
	GOOS=windows GOARCH=amd64 go build -o $(OUTPUT_DIR)/service-windows-amd64.exe $(SERVICE)
	GOOS=darwin  GOARCH=amd64 go build -o $(OUTPUT_DIR)/service-darwin-amd64 $(SERVICE)
	GOOS=darwin  GOARCH=arm64 go build -o $(OUTPUT_DIR)/service-darwin-arm64 $(SERVICE)

# Clean binaries
clean:
	@echo "Cleaning binaries..."
	rm -rf $(OUTPUT_DIR)/
