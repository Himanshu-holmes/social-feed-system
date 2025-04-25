# Paths to the main files 
SERVICE = ./cmd/post_service/main.go

# Run both services in the background
run:
	@echo "Starting Service ..."
	go run $(SERVICE) > posts.log 2>&1 &

# Run both in foreground (useful for debugging)
run-foreground:
	@echo "Running Service  in foreground..."
	go run $(SERVICE)

# Stop all running Go services
stop:
	@echo "Killing Go services..."
	@pkill -f $(SERVICE) || true
	

# Build binaries
build:
	go build -o bin/service $(SERVICE)


# Clean binaries
clean:
	rm -rf bin/



