.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o bin/seckill main.go

# Run the application
run:
	go run main.go

# Run tests
test:
	go test -v ./...

# Run benchmark tests
bench:
	go test -v -bench=. -benchmem ./tests/

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Docker build
docker-build:
	docker build -t go-seckill:latest .

# Docker compose up
docker-up:
	docker-compose up -d

# Docker compose down
docker-down:
	docker-compose down

# Docker compose logs
docker-logs:
	docker-compose logs -f

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

